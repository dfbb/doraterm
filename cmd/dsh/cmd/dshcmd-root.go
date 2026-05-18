// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/dfbb/doraterm/pkg/util/shellutil"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshrpc/wshclient"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

var (
	rootCmd = &cobra.Command{
		Use:          "wsh",
		Short:        "CLI tool to control Wave Terminal",
		Long:         `wsh is a small utility that lets you do cool things with Wave Terminal, right from the command line`,
		SilenceUsage: true,
	}
)

var WrappedStdin io.Reader = os.Stdin
var WrappedStdout io.Writer = &WrappedWriter{dest: os.Stdout}
var WrappedStderr io.Writer = &WrappedWriter{dest: os.Stderr}
var RpcClient *dshutil.DshRpc
var RpcContext dshrpc.RpcContext
var RpcClientRouteId string
var UsingTermWshMode bool
var blockArg string
var DshExitCode int

type WrappedWriter struct {
	dest io.Writer
}

func (w *WrappedWriter) Write(p []byte) (n int, err error) {
	if !UsingTermWshMode {
		return w.dest.Write(p)
	}
	count := 0
	for _, b := range p {
		if b == '\n' {
			count++
		}
	}
	if count == 0 {
		return w.dest.Write(p)
	}
	buf := make([]byte, len(p)+count) // Each '\n' adds one extra byte for '\r'
	writeIdx := 0
	for _, b := range p {
		if b == '\n' {
			buf[writeIdx] = '\r'
			buf[writeIdx+1] = '\n'
			writeIdx += 2
		} else {
			buf[writeIdx] = b
			writeIdx++
		}
	}
	return w.dest.Write(buf)
}

func WriteStderr(fmtStr string, args ...interface{}) {
	WrappedStderr.Write([]byte(fmt.Sprintf(fmtStr, args...)))
}

func WriteStdout(fmtStr string, args ...interface{}) {
	WrappedStdout.Write([]byte(fmt.Sprintf(fmtStr, args...)))
}

func OutputHelpMessage(cmd *cobra.Command) {
	cmd.SetOutput(WrappedStderr)
	cmd.Help()
	WriteStderr("\n")
}

func preRunSetupRpcClient(cmd *cobra.Command, args []string) error {
	jwtToken := os.Getenv(dshutil.DoraJwtTokenVarName)
	if jwtToken == "" {
		return fmt.Errorf("wsh must be run inside a Wave-managed SSH session (DORATERM_JWT not found)")
	}
	err := setupRpcClient(nil, jwtToken)
	if err != nil {
		return err
	}
	return nil
}

func getIsTty() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	return false
}

type RunEFnType = func(*cobra.Command, []string) error

func activityWrap(activityStr string, origRunE RunEFnType) RunEFnType {
	return func(cmd *cobra.Command, args []string) (rtnErr error) {
		defer func() {
			sendActivity(activityStr, rtnErr == nil)
		}()
		return origRunE(cmd, args)
	}
}

func resolveBlockArg() (*doraobj.ORef, error) {
	oref := blockArg
	if oref == "" {
		oref = "this"
	}
	fullORef, err := resolveSimpleId(oref)
	if err != nil {
		return nil, fmt.Errorf("resolving blockid: %w", err)
	}
	return fullORef, nil
}

func setupRpcClientWithToken(swapTokenStr string) (dshrpc.CommandAuthenticateRtnData, error) {
	var rtn dshrpc.CommandAuthenticateRtnData
	token, err := shellutil.UnpackSwapToken(swapTokenStr)
	if err != nil {
		return rtn, fmt.Errorf("error unpacking token: %w", err)
	}
	if token.RpcContext == nil {
		return rtn, fmt.Errorf("no rpccontext in token")
	}
	if token.RpcContext.SockName == "" {
		return rtn, fmt.Errorf("no sockname in token")
	}
	RpcContext = *token.RpcContext
	RpcClient, err = dshutil.SetupDomainSocketRpcClient(token.RpcContext.SockName, nil, "wshcmd")
	if err != nil {
		return rtn, fmt.Errorf("error setting up domain socket rpc client: %w", err)
	}
	rtn, err = dshclient.AuthenticateTokenCommand(RpcClient, dshrpc.CommandAuthenticateTokenData{Token: token.Token}, &dshrpc.RpcOpts{Route: dshutil.ControlRoute})
	if err != nil {
		return rtn, err
	}
	RpcClientRouteId = rtn.RouteId
	return rtn, nil
}

// returns the wrapped stdin and a new rpc client (that wraps the stdin input and stdout output)
func setupRpcClient(serverImpl dshutil.ServerImpl, jwtToken string) error {
	rpcCtx, err := dshutil.ExtractUnverifiedRpcContext(jwtToken)
	if err != nil {
		return fmt.Errorf("error extracting rpc context from %s: %v", dshutil.DoraJwtTokenVarName, err)
	}
	RpcContext = *rpcCtx
	sockName, err := dshutil.ExtractUnverifiedSocketName(jwtToken)
	if err != nil {
		return fmt.Errorf("error extracting socket name from %s: %v", dshutil.DoraJwtTokenVarName, err)
	}
	RpcClient, err = dshutil.SetupDomainSocketRpcClient(sockName, serverImpl, "wshcmd")
	if err != nil {
		return fmt.Errorf("error setting up domain socket rpc client: %v", err)
	}
	authRtn, err := dshclient.AuthenticateCommand(RpcClient, jwtToken, &dshrpc.RpcOpts{Route: dshutil.ControlRoute})
	if err != nil {
		return fmt.Errorf("error authenticating: %v", err)
	}
	RpcClientRouteId = authRtn.RouteId
	blockId := os.Getenv("DORATERM_BLOCKID")
	if blockId != "" {
		peerInfo := fmt.Sprintf("domain:block:%s", blockId)
		dshclient.SetPeerInfoCommand(RpcClient, peerInfo, &dshrpc.RpcOpts{Route: dshutil.ControlRoute})
	}
	// note we don't modify WrappedStdin here (just use os.Stdin)
	return nil
}

func isFullORef(orefStr string) bool {
	_, err := doraobj.ParseORef(orefStr)
	return err == nil
}

func resolveSimpleId(id string) (*doraobj.ORef, error) {
	if isFullORef(id) {
		orefObj, err := doraobj.ParseORef(id)
		if err != nil {
			return nil, fmt.Errorf("error parsing full ORef: %v", err)
		}
		return &orefObj, nil
	}
	blockId := os.Getenv("DORATERM_BLOCKID")
	if blockId == "" {
		return nil, fmt.Errorf("no DORATERM_BLOCKID env var set")
	}
	rtnData, err := dshclient.ResolveIdsCommand(RpcClient, dshrpc.CommandResolveIdsData{
		BlockId: blockId,
		Ids:     []string{id},
	}, &dshrpc.RpcOpts{Timeout: 2000})
	if err != nil {
		return nil, fmt.Errorf("error resolving ids: %v", err)
	}
	oref, ok := rtnData.ResolvedIds[id]
	if !ok {
		return nil, fmt.Errorf("id not found: %q", id)
	}
	return &oref, nil
}

func getTabIdFromEnv() string {
	return os.Getenv("DORATERM_TABID")
}

// this will send wsh activity to the client running on *your* local machine (it does not contact any wave cloud infrastructure)
// if you've turned off telemetry in your local client, this data never gets sent to us
// no parameters or timestamps are sent, as you can see below, it just sends the name of the command (and if there was an error)
// (e.g. "wsh ai ..." would send "ai")
// this helps us understand which commands are actually being used so we know where to concentrate our effort
func sendActivity(wshCmdName string, success bool) {
	if RpcClient == nil || wshCmdName == "" {
		return
	}
	dataMap := make(map[string]int)
	dataMap[wshCmdName] = 1
	if !success {
		dataMap[wshCmdName+"#"+"error"] = 1
	}
	dshclient.DshActivityCommand(RpcClient, dataMap, nil)
}

// Execute executes the root command.
func Execute() {
	defer func() {
		r := recover()
		if r != nil {
			WriteStderr("[panic] %v\n", r)
			debug.PrintStack()
			dshutil.DoShutdown("", 1, true)
		} else {
			dshutil.DoShutdown("", DshExitCode, false)
		}
	}()
	rootCmd.PersistentFlags().StringVarP(&blockArg, "block", "b", "", "for commands which require a block id")
	err := rootCmd.Execute()
	if err != nil {
		dshutil.DoShutdown("", 1, true)
		return
	}
}
