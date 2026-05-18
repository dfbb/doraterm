// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"runtime"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/dfbb/doraterm/pkg/authkey"
	"github.com/dfbb/doraterm/pkg/blockcontroller"
	"github.com/dfbb/doraterm/pkg/blocklogger"
	"github.com/dfbb/doraterm/pkg/filestore"
	"github.com/dfbb/doraterm/pkg/jobcontroller"
	"github.com/dfbb/doraterm/pkg/panichandler"
	"github.com/dfbb/doraterm/pkg/remote/fileshare/wshfs"
	"github.com/dfbb/doraterm/pkg/secretstore"
	"github.com/dfbb/doraterm/pkg/service"
	"github.com/dfbb/doraterm/pkg/telemetry"
	"github.com/dfbb/doraterm/pkg/telemetry/telemetrydata"
	"github.com/dfbb/doraterm/pkg/util/envutil"
	"github.com/dfbb/doraterm/pkg/util/shellutil"
	"github.com/dfbb/doraterm/pkg/util/sigutil"
	"github.com/dfbb/doraterm/pkg/util/utilfn"
	"github.com/dfbb/doraterm/pkg/dorabase"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/wcloud"
	"github.com/dfbb/doraterm/pkg/dconfig"
	"github.com/dfbb/doraterm/pkg/dcore"
	"github.com/dfbb/doraterm/pkg/web"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshrpc/wshclient"
	"github.com/dfbb/doraterm/pkg/dshrpc/wshlocal"
	"github.com/dfbb/doraterm/pkg/dshrpc/wshserver"
	"github.com/dfbb/doraterm/pkg/dshutil"
	"github.com/dfbb/doraterm/pkg/dstore"

	"net/http"
	_ "net/http/pprof"
)

// these are set at build time
var WaveVersion = "0.0.0"
var BuildTime = "0"

const InitialTelemetryWait = 10 * time.Second
const TelemetryTick = 2 * time.Minute
const TelemetryInterval = 4 * time.Hour
const TelemetryInitialCountsWait = 5 * time.Second
const TelemetryCountsInterval = 1 * time.Hour
const InitialDiagnosticWait = 5 * time.Minute
const DiagnosticTick = 10 * time.Minute

var shutdownOnce sync.Once

func init() {
	envFilePath := os.Getenv("WAVETERM_ENVFILE")
	if envFilePath != "" {
		log.Printf("applying env file: %s\n", envFilePath)
		_ = godotenv.Load(envFilePath)
	}
}

func doShutdown(reason string) {
	shutdownOnce.Do(func() {
		log.Printf("shutting down: %s\n", reason)
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()
		go blockcontroller.StopAllBlockControllersForShutdown()
		shutdownActivityUpdate()
		sendTelemetryWrapper()
		// TODO deal with flush in progress
		clearTempFiles()
		filestore.WFS.FlushCache(ctx)
		watcher := dconfig.GetWatcher()
		if watcher != nil {
			watcher.Close()
		}
		time.Sleep(500 * time.Millisecond)
		log.Printf("shutdown complete\n")
		os.Exit(0)
	})
}

// watch stdin, kill server if stdin is closed
func stdinReadWatch() {
	defer func() {
		panichandler.PanicHandler("stdinReadWatch", recover())
	}()
	buf := make([]byte, 1024)
	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			doShutdown(fmt.Sprintf("stdin closed/error (%v)", err))
			break
		}
	}
}

func startConfigWatcher() {
	watcher := dconfig.GetWatcher()
	if watcher != nil {
		watcher.Start()
	}
}

func telemetryLoop() {
	defer func() {
		panichandler.PanicHandler("telemetryLoop", recover())
	}()
	var nextSend int64
	time.Sleep(InitialTelemetryWait)
	for {
		if time.Now().Unix() > nextSend {
			nextSend = time.Now().Add(TelemetryInterval).Unix()
			sendTelemetryWrapper()
		}
		time.Sleep(TelemetryTick)
	}
}

func diagnosticLoop() {
	defer func() {
		panichandler.PanicHandler("diagnosticLoop", recover())
	}()
	if os.Getenv("WAVETERM_NOPING") != "" {
		log.Printf("WAVETERM_NOPING set, disabling diagnostic ping\n")
		return
	}
	var lastSentDate string
	time.Sleep(InitialDiagnosticWait)
	for {
		currentDate := time.Now().Format("2006-01-02")
		if lastSentDate == "" || lastSentDate != currentDate {
			if sendDiagnosticPing() {
				lastSentDate = currentDate
			}
		}
		time.Sleep(DiagnosticTick)
	}
}

func sendDiagnosticPing() bool {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	rpcClient := dshclient.GetBareRpcClient()
	isOnline, err := dshclient.NetworkOnlineCommand(rpcClient, &dshrpc.RpcOpts{Route: "electron", Timeout: 2000})
	if err != nil || !isOnline {
		return false
	}
	clientId := dstore.GetClientId()
	usageTelemetry := telemetry.IsTelemetryEnabled()
	wcloud.SendDiagnosticPing(ctx, clientId, usageTelemetry)
	return true
}

func setupTelemetryConfigHandler() {
	watcher := dconfig.GetWatcher()
	if watcher == nil {
		return
	}
	currentConfig := watcher.GetFullConfig()
	currentTelemetryEnabled := currentConfig.Settings.TelemetryEnabled

	watcher.RegisterUpdateHandler(func(newConfig dconfig.FullConfigType) {
		newTelemetryEnabled := newConfig.Settings.TelemetryEnabled
		if newTelemetryEnabled != currentTelemetryEnabled {
			currentTelemetryEnabled = newTelemetryEnabled
			dcore.GoSendNoTelemetryUpdate(newTelemetryEnabled)
		}
	})
}

func panicTelemetryHandler(panicName string) {
	activity := dshrpc.ActivityUpdate{NumPanics: 1}
	err := telemetry.UpdateActivity(context.Background(), activity)
	if err != nil {
		log.Printf("error updating activity (panicTelemetryHandler): %v\n", err)
	}
	telemetry.RecordTEvent(context.Background(), telemetrydata.MakeTEvent("debug:panic", telemetrydata.TEventProps{
		PanicType: panicName,
	}))
}

func sendTelemetryWrapper() {
	defer func() {
		panichandler.PanicHandler("sendTelemetryWrapper", recover())
	}()
	ctx, cancelFn := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFn()
	beforeSendActivityUpdate(ctx)
	clientId := dstore.GetClientId()
	err := wcloud.SendAllTelemetry(clientId)
	if err != nil {
		log.Printf("[error] sending telemetry: %v\n", err)
	}
}

func updateTelemetryCounts(lastCounts telemetrydata.TEventProps) telemetrydata.TEventProps {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()
	var props telemetrydata.TEventProps
	props.CountBlocks, _ = dstore.DBGetCount[*doraobj.Block](ctx)
	props.CountTabs, _ = dstore.DBGetCount[*doraobj.Tab](ctx)
	props.CountWindows, _ = dstore.DBGetCount[*doraobj.Window](ctx)
	props.CountWorkspaces, _, _ = dstore.DBGetWSCounts(ctx)
	props.CountJobs = jobcontroller.GetNumJobsRunning()
	props.CountJobsConnected = jobcontroller.GetNumJobsConnected()
	props.CountViews, _ = dstore.DBGetBlockViewCounts(ctx)

	fullConfig := dconfig.GetWatcher().GetFullConfig()
	customWidgets := fullConfig.CountCustomWidgets()
	customAIPresets := fullConfig.CountCustomAIPresets()
	customSettings := dconfig.CountCustomSettings()
	customAIModes := fullConfig.CountCustomAIModes()

	props.UserSet = &telemetrydata.TEventUserProps{
		SettingsCustomWidgets:   customWidgets,
		SettingsCustomAIPresets: customAIPresets,
		SettingsCustomSettings:  customSettings,
		SettingsCustomAIModes:   customAIModes,
	}

	secretsCount, err := secretstore.CountSecrets()
	if err == nil {
		props.UserSet.SettingsSecretsCount = secretsCount
	}

	if utilfn.CompareAsMarshaledJson(props, lastCounts) {
		return lastCounts
	}
	tevent := telemetrydata.MakeTEvent("app:counts", props)
	err = telemetry.RecordTEvent(ctx, tevent)
	if err != nil {
		log.Printf("error recording counts tevent: %v\n", err)
	}
	return props
}

func updateTelemetryCountsLoop() {
	defer func() {
		panichandler.PanicHandler("updateTelemetryCountsLoop", recover())
	}()
	var nextSend int64
	var lastCounts telemetrydata.TEventProps
	time.Sleep(TelemetryInitialCountsWait)
	for {
		if time.Now().Unix() > nextSend {
			nextSend = time.Now().Add(TelemetryCountsInterval).Unix()
			lastCounts = updateTelemetryCounts(lastCounts)
		}
		time.Sleep(TelemetryTick)
	}
}

func beforeSendActivityUpdate(ctx context.Context) {
	activity := dshrpc.ActivityUpdate{}
	activity.NumTabs, _ = dstore.DBGetCount[*doraobj.Tab](ctx)
	activity.NumBlocks, _ = dstore.DBGetCount[*doraobj.Block](ctx)
	activity.Blocks, _ = dstore.DBGetBlockViewCounts(ctx)
	activity.NumWindows, _ = dstore.DBGetCount[*doraobj.Window](ctx)
	activity.NumWSNamed, activity.NumWS, _ = dstore.DBGetWSCounts(ctx)
	err := telemetry.UpdateActivity(ctx, activity)
	if err != nil {
		log.Printf("error updating before activity: %v\n", err)
	}
}

func startupActivityUpdate(firstLaunch bool) {
	defer func() {
		panichandler.PanicHandler("startupActivityUpdate", recover())
	}()
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()
	activity := dshrpc.ActivityUpdate{Startup: 1}
	err := telemetry.UpdateActivity(ctx, activity) // set at least one record into activity (don't use go routine wrap here)
	if err != nil {
		log.Printf("error updating startup activity: %v\n", err)
	}
	autoUpdateChannel := telemetry.AutoUpdateChannel()
	autoUpdateEnabled := telemetry.IsAutoUpdateEnabled()
	shellType, shellVersion, shellErr := shellutil.DetectShellTypeAndVersion()
	if shellErr != nil {
		shellType = "error"
		shellVersion = ""
	}
	userSetOnce := &telemetrydata.TEventUserProps{
		ClientInitialVersion: "v" + WaveVersion,
	}
	tosTs := telemetry.GetTosAgreedTs()
	var cohortTime time.Time
	if tosTs > 0 {
		cohortTime = time.UnixMilli(tosTs)
	} else {
		cohortTime = time.Now()
	}
	cohortMonth := cohortTime.Format("2006-01")
	year, week := cohortTime.ISOWeek()
	cohortISOWeek := fmt.Sprintf("%04d-W%02d", year, week)
	userSetOnce.CohortMonth = cohortMonth
	userSetOnce.CohortISOWeek = cohortISOWeek
	fullConfig := dconfig.GetWatcher().GetFullConfig()
	props := telemetrydata.TEventProps{
		UserSet: &telemetrydata.TEventUserProps{
			ClientVersion:       "v" + dorabase.WaveVersion,
			ClientBuildTime:     dorabase.BuildTime,
			ClientArch:          dorabase.ClientArch(),
			ClientOSRelease:     dorabase.UnameKernelRelease(),
			ClientIsDev:         dorabase.IsDevMode(),
			ClientPackageType:   dorabase.ClientPackageType(),
			ClientMacOSVersion:  dorabase.ClientMacOSVersion(),
			AutoUpdateChannel:   autoUpdateChannel,
			AutoUpdateEnabled:   autoUpdateEnabled,
			LocalShellType:      shellType,
			LocalShellVersion:   shellVersion,
			SettingsTransparent: fullConfig.Settings.WindowTransparent,
		},
		UserSetOnce: userSetOnce,
	}
	if firstLaunch {
		props.AppFirstLaunch = true
	}
	tevent := telemetrydata.MakeTEvent("app:startup", props)
	err = telemetry.RecordTEvent(ctx, tevent)
	if err != nil {
		log.Printf("error recording startup event: %v\n", err)
	}
}

func shutdownActivityUpdate() {
	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFn()
	activity := dshrpc.ActivityUpdate{Shutdown: 1}
	err := telemetry.UpdateActivity(ctx, activity) // do NOT use the go routine wrap here (this needs to be synchronous)
	if err != nil {
		log.Printf("error updating shutdown activity: %v\n", err)
	}
	err = telemetry.TruncateActivityTEventForShutdown(ctx)
	if err != nil {
		log.Printf("error truncating activity t-event for shutdown: %v\n", err)
	}
	tevent := telemetrydata.MakeTEvent("app:shutdown", telemetrydata.TEventProps{})
	err = telemetry.RecordTEvent(ctx, tevent)
	if err != nil {
		log.Printf("error recording shutdown event: %v\n", err)
	}
}

func createMainWshClient() {
	rpc := dshserver.GetMainRpcClient()
	dshutil.DefaultRouter.RegisterTrustedLeaf(rpc, dshutil.DefaultRoute)
	dps.Broker.SetClient(dshutil.DefaultRouter)
	localInitialEnv := envutil.PruneInitialEnv(envutil.SliceToMap(os.Environ()))
	sockName := dorabase.GetDomainSocketName()
	localImpl := dshlocal.MakeLocalRpcServerImpl(nil, dshutil.DefaultRouter, dshclient.GetBareRpcClient(), localInitialEnv, sockName)
	localConnWsh := dshutil.MakeWshRpc(dshrpc.RpcContext{Conn: dshrpc.LocalConnName}, localImpl, "conn:local")
	dshutil.DefaultRouter.RegisterTrustedLeaf(localConnWsh, dshutil.MakeConnectionRouteId(dshrpc.LocalConnName))
	wshfs.RpcClient = localConnWsh
	wshfs.RpcClientRouteId = dshutil.MakeConnectionRouteId(dshrpc.LocalConnName)
}

func grabAndRemoveEnvVars() error {
	err := authkey.SetAuthKeyFromEnv()
	if err != nil {
		return fmt.Errorf("setting auth key: %v", err)
	}
	err = dorabase.CacheAndRemoveEnvVars()
	if err != nil {
		return err
	}
	err = wcloud.CacheAndRemoveEnvVars()
	if err != nil {
		return err
	}

	// Remove WAVETERM env vars that leak from prod => dev
	os.Unsetenv("WAVETERM_CLIENTID")
	os.Unsetenv("WAVETERM_WORKSPACEID")
	os.Unsetenv("WAVETERM_TABID")
	os.Unsetenv("WAVETERM_BLOCKID")
	os.Unsetenv("WAVETERM_CONN")
	os.Unsetenv("WAVETERM_JWT")
	os.Unsetenv("WAVETERM_VERSION")

	return nil
}

func clearTempFiles() error {
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFn()
	client, err := dstore.DBGetSingleton[*doraobj.Client](ctx)
	if err != nil {
		return fmt.Errorf("error getting client: %v", err)
	}
	filestore.WFS.DeleteZone(ctx, client.TempOID)
	return nil
}

func maybeStartPprofServer() {
	settings := dconfig.GetWatcher().GetFullConfig().Settings
	if settings.DebugPprofMemProfileRate != nil {
		runtime.MemProfileRate = *settings.DebugPprofMemProfileRate
		log.Printf("set runtime.MemProfileRate to %d\n", runtime.MemProfileRate)
	}
	if settings.DebugPprofPort == nil {
		return
	}
	pprofPort := *settings.DebugPprofPort
	if pprofPort < 1 || pprofPort > 65535 {
		log.Printf("[error] debug:pprofport must be between 1 and 65535, got %d\n", pprofPort)
		return
	}
	go func() {
		addr := fmt.Sprintf("localhost:%d", pprofPort)
		log.Printf("starting pprof server on %s\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("[error] pprof server failed: %v\n", err)
		}
	}()
}

func main() {
	log.SetFlags(0) // disable timestamp since electron's winston logger already wraps with timestamp
	log.SetPrefix("[wavesrv] ")
	dorabase.WaveVersion = WaveVersion
	dorabase.BuildTime = BuildTime
	dshutil.DefaultRouter = dshutil.NewWshRouter()
	dshutil.DefaultRouter.SetAsRootRouter()

	err := grabAndRemoveEnvVars()
	if err != nil {
		log.Printf("[error] %v\n", err)
		return
	}
	err = service.ValidateServiceMap()
	if err != nil {
		log.Printf("error validating service map: %v\n", err)
		return
	}
	err = dorabase.EnsureWaveDataDir()
	if err != nil {
		log.Printf("error ensuring wave home dir: %v\n", err)
		return
	}
	err = dorabase.EnsureWaveDBDir()
	if err != nil {
		log.Printf("error ensuring wave db dir: %v\n", err)
		return
	}
	err = dorabase.EnsureWaveConfigDir()
	if err != nil {
		log.Printf("error ensuring wave config dir: %v\n", err)
		return
	}

	// TODO: rather than ensure this dir exists, we should let the editor recursively create parent dirs on save
	err = dorabase.EnsureWavePresetsDir()
	if err != nil {
		log.Printf("error ensuring wave presets dir: %v\n", err)
		return
	}
	err = dorabase.EnsureWaveCachesDir()
	if err != nil {
		log.Printf("error ensuring wave caches dir: %v\n", err)
		return
	}
	waveLock, err := dorabase.AcquireWaveLock()
	if err != nil {
		log.Printf("error acquiring wave lock (another instance of Wave is likely running): %v\n", err)
		return
	}
	defer func() {
		err = waveLock.Close()
		if err != nil {
			log.Printf("error releasing wave lock: %v\n", err)
		}
	}()
	log.Printf("wave version: %s (%s)\n", WaveVersion, BuildTime)
	log.Printf("wave data dir: %s\n", dorabase.GetWaveDataDir())
	log.Printf("wave config dir: %s\n", dorabase.GetWaveConfigDir())
	err = filestore.InitFilestore()
	if err != nil {
		log.Printf("error initializing filestore: %v\n", err)
		return
	}
	err = dstore.InitWStore()
	if err != nil {
		log.Printf("error initializing wstore: %v\n", err)
		return
	}
	panichandler.PanicTelemetryHandler = panicTelemetryHandler
	go func() {
		defer func() {
			panichandler.PanicHandler("InitCustomShellStartupFiles", recover())
		}()
		err := shellutil.InitCustomShellStartupFiles()
		if err != nil {
			log.Printf("error initializing wsh and shell-integration files: %v\n", err)
		}
	}()
	firstLaunch, err := dcore.EnsureInitialData()
	if err != nil {
		log.Printf("error ensuring initial data: %v\n", err)
		return
	}
	if firstLaunch {
		log.Printf("first launch detected")
	}
	err = clearTempFiles()
	if err != nil {
		log.Printf("error clearing temp files: %v\n", err)
		return
	}
	err = dcore.InitMainServer()
	if err != nil {
		log.Printf("error initializing mainserver: %v\n", err)
		return
	}

	err = shellutil.FixupWaveZshHistory()
	if err != nil {
		log.Printf("error fixing up wave zsh history: %v\n", err)
	}
	createMainWshClient()
	sigutil.InstallShutdownSignalHandlers(doShutdown)
	sigutil.InstallSIGUSR1Handler()
	dconfig.MigratePresetsBackgrounds()
	startConfigWatcher()
	maybeStartPprofServer()
	go stdinReadWatch()
	go telemetryLoop()
	go diagnosticLoop()
	setupTelemetryConfigHandler()
	go updateTelemetryCountsLoop()
	go startupActivityUpdate(firstLaunch) // must be after startConfigWatcher()
	blocklogger.InitBlockLogger()
	jobcontroller.InitJobController()
	blockcontroller.InitBlockController()
	err = dcore.InitBadgeStore()
	if err != nil {
		log.Printf("error initializing badge store: %v\n", err)
		return
	}
	go func() {
		defer func() {
			panichandler.PanicHandler("GetSystemSummary", recover())
		}()
		dorabase.GetSystemSummary()
	}()

	webListener, err := web.MakeTCPListener("web")
	if err != nil {
		log.Printf("error creating web listener: %v\n", err)
		return
	}
	wsListener, err := web.MakeTCPListener("websocket")
	if err != nil {
		log.Printf("error creating websocket listener: %v\n", err)
		return
	}
	go web.RunWebSocketServer(wsListener)
	unixListener, err := web.MakeUnixListener()
	if err != nil {
		log.Printf("error creating unix listener: %v\n", err)
		return
	}
	go func() {
		if BuildTime == "" {
			BuildTime = "0"
		}
		// use fmt instead of log here to make sure it goes directly to stderr
		fmt.Fprintf(os.Stderr, "WAVESRV-ESTART ws:%s web:%s version:%s buildtime:%s\n", wsListener.Addr(), webListener.Addr(), WaveVersion, BuildTime)
	}()
	go dshutil.RunWshRpcOverListener(unixListener, nil)
	web.RunWebServer(webListener) // blocking
	runtime.KeepAlive(waveLock)
}
