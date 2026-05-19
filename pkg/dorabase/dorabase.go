// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dorabase

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dfbb/doraterm/pkg/util/utilfn"
)

// set by main-server.go
var DoraVersion = "0.0.0"
var BuildTime = "0"

const (
	DoraConfigHomeEnvVar = "DORATERM_CONFIG_HOME"
	DoraDataHomeEnvVar = "DORATERM_DATA_HOME"
	DoraAppPathVarName = "DORATERM_APP_PATH"
	DoraAppResourcesPathVarName = "DORATERM_RESOURCES_PATH"
	DoraAppElectronExecPathVarName = "DORATERM_ELECTRONEXECPATH"
	DoraDevVarName = "DORATERM_DEV"
	DoraDevViteVarName = "DORATERM_DEV_VITE"
	DoraDshForceUpdateVarName = "DORATERM_DSHFORCEUPDATE"
	DoraNoConfirmQuitVarName = "DORATERM_NOCONFIRMQUIT"

	DoraJwtTokenVarName = "DORATERM_JWT"
	DoraSwapTokenVarName = "DORATERM_SWAPTOKEN"
)

const (
	BlockFile_Term  = "term"            // used for main pty output
	BlockFile_Cache = "cache:term:full" // for cached block
	BlockFile_Env   = "env"
)

const NeedJwtConst = "NEED-JWT"

var ConfigHome_VarCache string          // caches DORATERM_CONFIG_HOME
var DataHome_VarCache string            // caches DORATERM_DATA_HOME
var AppPath_VarCache string             // caches DORATERM_APP_PATH
var AppResourcesPath_VarCache string    // caches DORATERM_RESOURCES_PATH
var AppElectronExecPath_VarCache string // caches DORATERM_ELECTRONEXECPATH
var Dev_VarCache string                 // caches DORATERM_DEV

const DoraLockFile = "dora.lock"
const DomainSocketBaseName = "dora.sock"
const RemoteDomainSocketBaseName = "dora-remote.sock"
const DoraDBDir = "db"
const ConfigDir = "config"
const RemoteDoraHomeDirName = ".doraterm"
const RemoteDshBinDirName = "bin"
const RemoteFullDshBinPath = "~/.doraterm/bin/dsh"
const RemoteFullDomainSocketPath = "~/.doraterm/dora-remote.sock"

const AppPathBinDir = "bin"

var baseLock = &sync.Mutex{}
var ensureDirCache = map[string]bool{}

var waveCachesDirOnce = &sync.Once{}
var waveCachesDir string

var SupportedDshBinaries = map[string]bool{
	"darwin-x64":    true,
	"darwin-arm64":  true,
	"linux-x64":     true,
	"linux-arm64":   true,
	"windows-x64":   true,
	"windows-arm64": true,
}

type FDLock interface {
	Close() error
}

func CacheAndRemoveEnvVars() error {
	ConfigHome_VarCache = os.Getenv(DoraConfigHomeEnvVar)
	if ConfigHome_VarCache == "" {
		return fmt.Errorf(DoraConfigHomeEnvVar + " not set")
	}
	os.Unsetenv(DoraConfigHomeEnvVar)
	DataHome_VarCache = os.Getenv(DoraDataHomeEnvVar)
	if DataHome_VarCache == "" {
		return fmt.Errorf("%s not set", DoraDataHomeEnvVar)
	}
	os.Unsetenv(DoraDataHomeEnvVar)
	AppPath_VarCache = os.Getenv(DoraAppPathVarName)
	os.Unsetenv(DoraAppPathVarName)
	AppResourcesPath_VarCache = os.Getenv(DoraAppResourcesPathVarName)
	os.Unsetenv(DoraAppResourcesPathVarName)
	AppElectronExecPath_VarCache = os.Getenv(DoraAppElectronExecPathVarName)
	os.Unsetenv(DoraAppElectronExecPathVarName)
	Dev_VarCache = os.Getenv(DoraDevVarName)
	os.Unsetenv(DoraDevVarName)
	os.Unsetenv(DoraDevViteVarName)
	os.Unsetenv(DoraNoConfirmQuitVarName)
	return nil
}

func IsDevMode() bool {
	return Dev_VarCache != ""
}

func GetDoraAppPath() string {
	return AppPath_VarCache
}

func GetDoraAppResourcesPath() string {
	return AppResourcesPath_VarCache
}

func GetDoraDataDir() string {
	return DataHome_VarCache
}

func GetDoraConfigDir() string {
	return ConfigHome_VarCache
}

func GetDoraAppBinPath() string {
	return filepath.Join(GetDoraAppPath(), AppPathBinDir)
}

func GetDoraAppElectronExecPath() string {
	return AppElectronExecPath_VarCache
}

func GetHomeDir() string {
	homeVar, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return homeVar
}

func ExpandHomeDir(pathStr string) (string, error) {
	if pathStr != "~" && !strings.HasPrefix(pathStr, "~/") && (!strings.HasPrefix(pathStr, `~\`) || runtime.GOOS != "windows") {
		return filepath.Clean(pathStr), nil
	}
	homeDir := GetHomeDir()
	if pathStr == "~" {
		return homeDir, nil
	}
	expandedPath := filepath.Clean(filepath.Join(homeDir, pathStr[2:]))
	absPath, err := filepath.Abs(filepath.Join(homeDir, expandedPath))
	if err != nil || !strings.HasPrefix(absPath, homeDir) {
		return "", fmt.Errorf("potential path traversal detected for path %s", pathStr)
	}
	return expandedPath, nil
}

func ExpandHomeDirSafe(pathStr string) string {
	path, _ := ExpandHomeDir(pathStr)
	return path
}

func ReplaceHomeDir(pathStr string) string {
	homeDir := GetHomeDir()
	if pathStr == homeDir {
		return "~"
	}
	if strings.HasPrefix(pathStr, homeDir+"/") {
		return "~" + pathStr[len(homeDir):]
	}
	return pathStr
}

func GetDomainSocketName() string {
	return filepath.Join(GetDoraDataDir(), DomainSocketBaseName)
}

// returns a Unix-style path for the remote socket (using fmt.Sprintf instead of filepath.Join
// because this path is for a remote Unix system, not the local OS which might be Windows)
func GetPersistentRemoteSockName(clientId string) string {
	return fmt.Sprintf("~/.doraterm/client/%s/doraterm.sock", clientId)
}

func EnsureDoraDataDir() error {
	return CacheEnsureDir(GetDoraDataDir(), "dorahome", 0700, "wave home directory")
}

func EnsureDoraDBDir() error {
	return CacheEnsureDir(filepath.Join(GetDoraDataDir(), DoraDBDir), "doradb", 0700, "wave db directory")
}

func EnsureDoraConfigDir() error {
	return CacheEnsureDir(GetDoraConfigDir(), "doraconfig", 0700, "wave config directory")
}

func EnsureDoraPresetsDir() error {
	return CacheEnsureDir(filepath.Join(GetDoraConfigDir(), "presets"), "dorapresets", 0700, "wave presets directory")
}

func resolveDoraCachesDir() string {
	var cacheDir string
	appBundle := "doraterm"
	if IsDevMode() {
		appBundle = "doraterm-dev"
	}

	switch runtime.GOOS {
	case "darwin":
		homeDir := GetHomeDir()
		cacheDir = filepath.Join(homeDir, "Library", "Caches", appBundle)
	case "linux":
		xdgCache := os.Getenv("XDG_CACHE_HOME")
		if xdgCache != "" {
			cacheDir = filepath.Join(xdgCache, appBundle)
		} else {
			homeDir := GetHomeDir()
			cacheDir = filepath.Join(homeDir, ".cache", appBundle)
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			cacheDir = filepath.Join(localAppData, appBundle, "Cache")
		}
	}

	if cacheDir == "" {
		tmpDir := os.TempDir()
		cacheDir = filepath.Join(tmpDir, appBundle)
	}

	return cacheDir
}

func GetDoraCachesDir() string {
	waveCachesDirOnce.Do(func() {
		waveCachesDir = resolveDoraCachesDir()
	})
	return waveCachesDir
}

func EnsureDoraCachesDir() error {
	return CacheEnsureDir(GetDoraCachesDir(), "doracaches", 0700, "wave caches directory")
}

func CacheEnsureDir(dirName string, cacheKey string, perm os.FileMode, dirDesc string) error {
	baseLock.Lock()
	ok := ensureDirCache[cacheKey]
	baseLock.Unlock()
	if ok {
		return nil
	}
	err := TryMkdirs(dirName, perm, dirDesc)
	if err != nil {
		return err
	}
	baseLock.Lock()
	ensureDirCache[cacheKey] = true
	baseLock.Unlock()
	return nil
}

func TryMkdirs(dirName string, perm os.FileMode, dirDesc string) error {
	info, err := os.Stat(dirName)
	if errors.Is(err, fs.ErrNotExist) {
		err = os.MkdirAll(dirName, perm)
		if err != nil {
			return fmt.Errorf("cannot make %s %q: %w", dirDesc, dirName, err)
		}
		info, err = os.Stat(dirName)
	}
	if err != nil {
		return fmt.Errorf("error trying to stat %s: %w", dirDesc, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s %q must be a directory", dirDesc, dirName)
	}
	return nil
}

func listValidLangs(ctx context.Context) []string {
	out, err := exec.CommandContext(ctx, "locale", "-a").CombinedOutput()
	if err != nil {
		log.Printf("error running 'locale -a': %s\n", err)
		return []string{}
	}
	// don't bother with CRLF line endings
	// this command doesn't work on windows
	return strings.Split(string(out), "\n")
}

var osLangOnce = &sync.Once{}
var osLang string

func determineLang() string {
	defaultLang := "en_US.UTF-8"
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFn()
	if runtime.GOOS == "darwin" {
		out, err := exec.CommandContext(ctx, "defaults", "read", "-g", "AppleLocale").CombinedOutput()
		if err != nil {
			log.Printf("error executing 'defaults read -g AppleLocale', will use default 'en_US.UTF-8': %v\n", err)
			return defaultLang
		}
		strOut := string(out)
		truncOut := strings.Split(strOut, "@")[0]
		preferredLang := strings.TrimSpace(truncOut) + ".UTF-8"
		validLangs := listValidLangs(ctx)

		if !utilfn.ContainsStr(validLangs, preferredLang) {
			log.Printf("unable to use desired lang %s, will use default 'en_US.UTF-8'\n", preferredLang)
			return defaultLang
		}

		return preferredLang
	} else {
		// this is specifically to get the dorasrv LANG so dorashell
		// on a remote uses the same LANG
		return os.Getenv("LANG")
	}
}

func DetermineLang() string {
	osLangOnce.Do(func() {
		osLang = determineLang()
	})
	return osLang
}

func DetermineLocale() string {
	truncated := strings.Split(DetermineLang(), ".")[0]
	if truncated == "" {
		return "C"
	}
	return strings.Replace(truncated, "_", "-", -1)
}

func ClientArch() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

func ClientPackageType() string {
	if os.Getenv("SNAP") != "" {
		return "snap"
	}
	if os.Getenv("APPIMAGE") != "" {
		return "appimage"
	}
	return ""
}

var macOSVersionOnce = &sync.Once{}
var cachedMacOSVersion string

var macOSVersionRegex = regexp.MustCompile(`^(\d+\.\d+(?:\.\d+)?)`)

func internalMacOSVersion() string {
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFn()
	out, err := exec.CommandContext(ctx, "sw_vers", "-productVersion").Output()
	if err != nil {
		return ""
	}
	versionStr := strings.TrimSpace(string(out))
	m := macOSVersionRegex.FindStringSubmatch(versionStr)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func ClientMacOSVersion() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	macOSVersionOnce.Do(func() {
		cachedMacOSVersion = internalMacOSVersion()
	})
	return cachedMacOSVersion
}

var releaseRegex = regexp.MustCompile(`^(\d+\.\d+\.\d+)`)
var osReleaseOnce = &sync.Once{}
var osRelease string

func unameKernelRelease() string {
	if runtime.GOOS == "windows" {
		return "-"
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFn()
	out, err := exec.CommandContext(ctx, "uname", "-r").CombinedOutput()
	if err != nil {
		log.Printf("error executing uname -r: %v\n", err)
		return "-"
	}
	releaseStr := strings.TrimSpace(string(out))
	m := releaseRegex.FindStringSubmatch(releaseStr)
	if len(m) < 2 {
		log.Printf("invalid uname -r output: [%s]\n", releaseStr)
		return "-"
	}
	return m[1]
}

func UnameKernelRelease() string {
	osReleaseOnce.Do(func() {
		osRelease = unameKernelRelease()
	})
	return osRelease
}

var systemSummaryOnce = &sync.Once{}
var systemSummary string

func GetSystemSummary() string {
	systemSummaryOnce.Do(func() {
		ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelFn()
		systemSummary = getSystemSummary(ctx)
	})
	return systemSummary
}

func ValidateDshSupportedArch(os string, arch string) error {
	if SupportedDshBinaries[fmt.Sprintf("%s-%s", os, arch)] {
		return nil
	}
	return fmt.Errorf("unsupported dsh platform: %s-%s", os, arch)
}

func getSystemSummary(ctx context.Context) string {
	osName := runtime.GOOS

	switch osName {
	case "darwin":
		out, _ := exec.CommandContext(ctx, "sw_vers", "-productVersion").Output()
		return fmt.Sprintf("macOS %s (%s)", strings.TrimSpace(string(out)), runtime.GOARCH)
	case "linux":
		// Read /etc/os-release directly (standard location since 2012)
		data, err := os.ReadFile("/etc/os-release")
		var prettyName string
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					prettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
					break
				}
			}
		}
		if prettyName == "" {
			prettyName = "Linux"
		} else if !strings.Contains(strings.ToLower(prettyName), "linux") {
			prettyName = "Linux " + prettyName
		}
		return fmt.Sprintf("%s (%s)", prettyName, runtime.GOARCH)
	case "windows":
		var details string
		out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", "(Get-CimInstance Win32_OperatingSystem).Caption").Output()
		if err == nil && len(out) > 0 {
			details = strings.TrimSpace(string(out))
		} else {
			details = "Windows"
		}
		return fmt.Sprintf("%s (%s)", details, runtime.GOARCH)
	default:
		return fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
	}
}

// job socket path on remote machine
func GetRemoteJobSocketPath(jobId string) string {
	socketDir := filepath.Join("/tmp", fmt.Sprintf("doraterm-%d", os.Getuid()))
	return filepath.Join(socketDir, fmt.Sprintf("%s.sock", jobId))
}

// job file path on remote machine
func GetRemoteJobFilePath(jobId string, extension string) string {
	jobDir := GetRemoteJobLogDir()
	return filepath.Join(jobDir, fmt.Sprintf("%s.%s", jobId, extension))
}

// job file dir on remote machines
func GetRemoteJobLogDir() string {
	homeDir := GetHomeDir()
	jobDir := filepath.Join(homeDir, ".doraterm", "jobs")
	return jobDir
}
