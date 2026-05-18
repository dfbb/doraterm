// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/dfbb/doraterm/pkg/remote"
	"github.com/dfbb/doraterm/pkg/remote/conncontroller"
	"github.com/dfbb/doraterm/pkg/shellexec"
	"github.com/dfbb/doraterm/pkg/userinput"
	"github.com/dfbb/doraterm/pkg/util/shellutil"
	"github.com/dfbb/doraterm/pkg/dorabase"
	"github.com/dfbb/doraterm/pkg/dorajwt"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dconfig"
	"github.com/dfbb/doraterm/pkg/dshrpc/wshserver"
	"github.com/dfbb/doraterm/pkg/dshutil"
	"github.com/dfbb/doraterm/pkg/dstore"
)

func setupDoraEnvVars() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	isDev := os.Getenv("DORATERM_DEV") != ""
	devSuffix := ""
	if isDev {
		devSuffix = "-dev"
	}

	configHome := os.Getenv("DORATERM_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(homeDir, ".config", "doraterm"+devSuffix)
		os.Setenv("DORATERM_CONFIG_HOME", configHome)
	}
	log.Printf("Using config directory: %s", configHome)

	dataHome := os.Getenv("DORATERM_DATA_HOME")
	if dataHome == "" {
		if runtime.GOOS == "darwin" {
			dataHome = filepath.Join(homeDir, "Library", "Application Support", "doraterm"+devSuffix)
			os.Setenv("DORATERM_DATA_HOME", dataHome)
		} else {
			return fmt.Errorf("DORATERM_DATA_HOME must be set on non-macOS systems")
		}
	}
	log.Printf("Using data directory: %s", dataHome)

	return nil
}

func initTestHarness(autoAccept bool) error {
	log.Printf("Initializing test harness...")

	err := setupDoraEnvVars()
	if err != nil {
		return fmt.Errorf("failed to setup wave env vars: %w", err)
	}

	err = dorabase.CacheAndRemoveEnvVars()
	if err != nil {
		return fmt.Errorf("failed to cache env vars: %w", err)
	}

	dshutil.DefaultRouter = dshutil.NewDshRouter()
	dshutil.DefaultRouter.SetAsRootRouter()

	dstore.SetClientId("test-client-" + fmt.Sprintf("%d", time.Now().Unix()))

	userinput.SetUserInputProvider(&CLIProvider{AutoAccept: autoAccept})

	keyPair, err := dorajwt.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate JWT key pair: %w", err)
	}

	err = dorajwt.SetPrivateKey(keyPair.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to set JWT private key: %w", err)
	}

	err = dorajwt.SetPublicKey(keyPair.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to set JWT public key: %w", err)
	}

	rpc := dshserver.GetMainRpcClient()
	dshutil.DefaultRouter.RegisterTrustedLeaf(rpc, dshutil.DefaultRoute)

	dconfig.GetWatcher().Start()

	log.Printf("Test harness initialized")
	return nil
}

func testBasicConnect(connName string, timeout time.Duration) error {
	opts, err := remote.ParseOpts(connName)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	log.Printf("Connecting to %s...", opts.String())

	conn := conncontroller.GetConn(opts)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = conn.Connect(ctx, &dconfig.ConnKeywords{})
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	status := conn.DeriveConnStatus()
	log.Printf("✓ Connected!")
	log.Printf("  Status: %s", status.Status)
	log.Printf("  DshEnabled: %v", status.DshEnabled)
	log.Printf("  Connection: %s", status.Connection)
	if status.DshVersion != "" {
		log.Printf("  DshVersion: %s", status.DshVersion)
	}
	if status.DshError != "" {
		log.Printf("  DshError: %s", status.DshError)
	}
	if status.NoWshReason != "" {
		log.Printf("  NoWshReason: %s", status.NoWshReason)
	}

	return nil
}

func testShellWithCommand(connName string, cmd string, timeout time.Duration) error {
	opts, err := remote.ParseOpts(connName)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	log.Printf("Connecting to %s...", opts.String())

	conn := conncontroller.GetConn(opts)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = conn.Connect(ctx, &dconfig.ConnKeywords{})
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	log.Printf("✓ Connected! Starting shell...")

	termSize := doraobj.TermSize{Rows: 24, Cols: 80}
	shellProc, err := shellexec.StartRemoteShellProcNoWsh(ctx, termSize, "", shellexec.CommandOptsType{}, conn)
	if err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}
	defer shellProc.Close()

	log.Printf("✓ Shell started! Executing: %s", cmd)

	_, err = shellProc.Cmd.Write([]byte(cmd + "\n"))
	if err != nil {
		return fmt.Errorf("failed to write command: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	buf := make([]byte, 8192)
	n, err := shellProc.Cmd.Read(buf)
	if err != nil {
		log.Printf("Warning: read error (may be expected): %v", err)
	}

	if n > 0 {
		log.Printf("\n--- Output ---\n%s\n--- End Output ---", string(buf[:n]))
	} else {
		log.Printf("No output received (timeout or no data)")
	}

	return nil
}

func testWshExec(connName string, cmd string, timeout time.Duration) error {
	opts, err := remote.ParseOpts(connName)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	log.Printf("Connecting to %s with wsh enabled...", opts.String())

	conn := conncontroller.GetConn(opts)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	wshEnabled := true
	err = conn.Connect(ctx, &dconfig.ConnKeywords{
		ConnDshEnabled: &wshEnabled,
	})
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	status := conn.DeriveConnStatus()
	log.Printf("✓ Connected! (wsh enabled: %v)", status.DshEnabled)
	if status.DshVersion != "" {
		log.Printf("  wsh version: %s", status.DshVersion)
	}
	if !status.DshEnabled {
		log.Printf("  WARNING: wsh not enabled - reason: %s", status.NoWshReason)
	}

	log.Printf("Starting wsh-enabled shell...")

	swapToken := &shellutil.TokenSwapEntry{
		Token: uuid.New().String(),
		Env:   make(map[string]string),
		Exp:   time.Now().Add(5 * time.Minute),
	}
	swapToken.Env["TERM_PROGRAM"] = "doraterm"
	swapToken.Env["WAVETERM"] = "1"
	swapToken.Env["DORATERM_VERSION"] = dorabase.DoraVersion
	swapToken.Env["DORATERM_CONN"] = connName

	cmdOpts := shellexec.CommandOptsType{
		SwapToken: swapToken,
	}

	termSize := doraobj.TermSize{Rows: 24, Cols: 80}
	shellProc, err := shellexec.StartRemoteShellProc(ctx, ctx, termSize, "", cmdOpts, conn)
	if err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}
	defer shellProc.Close()

	log.Printf("✓ Shell started! Executing: %s", cmd)

	_, err = shellProc.Cmd.Write([]byte(cmd + "\n"))
	if err != nil {
		return fmt.Errorf("failed to write command: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	buf := make([]byte, 8192)
	n, err := shellProc.Cmd.Read(buf)
	if err != nil {
		log.Printf("Warning: read error (may be expected): %v", err)
	}

	if n > 0 {
		log.Printf("\n--- Output ---\n%s\n--- End Output ---", string(buf[:n]))
	} else {
		log.Printf("No output received (timeout or no data)")
	}

	return nil
}

func testInteractiveShell(connName string, timeout time.Duration) error {
	opts, err := remote.ParseOpts(connName)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	log.Printf("Connecting to %s...", opts.String())

	conn := conncontroller.GetConn(opts)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = conn.Connect(ctx, &dconfig.ConnKeywords{})
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	log.Printf("✓ Connected! Starting interactive shell...")
	log.Printf("Note: This is a simple test - output may be mixed with prompts")
	log.Printf("Type commands and press Enter. Type 'exit' to quit.\n")

	termSize := doraobj.TermSize{Rows: 24, Cols: 80}
	shellProc, err := shellexec.StartRemoteShellProcNoWsh(ctx, termSize, "", shellexec.CommandOptsType{}, conn)
	if err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}
	defer shellProc.Close()

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := shellProc.Cmd.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				fmt.Print(string(buf[:n]))
			}
		}
	}()

	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				shellProc.Cmd.Write(buf[:n])
			}
		}
	}()

	shellProc.Wait()
	log.Printf("\nShell exited")

	return nil
}
