// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package wshlocal

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wavetermdev/waveterm/pkg/baseds"
	"github.com/wavetermdev/waveterm/pkg/panichandler"
	"github.com/wavetermdev/waveterm/pkg/util/unixutil"
	"github.com/wavetermdev/waveterm/pkg/wavebase"
	"github.com/wavetermdev/waveterm/pkg/wps"
	"github.com/wavetermdev/waveterm/pkg/wshrpc"
	"github.com/wavetermdev/waveterm/pkg/wshrpc/wshclient"
	"github.com/wavetermdev/waveterm/pkg/wshutil"
)

type JobManagerConnection struct {
	JobId     string
	Conn      net.Conn
	WshRpc    *wshutil.WshRpc
	CleanupFn func()
}

type ServerImpl struct {
	LogWriter     io.Writer
	Router        *wshutil.WshRouter
	RpcClient     *wshutil.WshRpc
	IsLocal       bool
	InitialEnv    map[string]string
	JobManagerMap map[string]*JobManagerConnection
	SockName      string
	Lock          sync.Mutex
}

func MakeLocalRpcServerImpl(logWriter io.Writer, router *wshutil.WshRouter, rpcClient *wshutil.WshRpc, initialEnv map[string]string, sockName string) *ServerImpl {
	return &ServerImpl{
		LogWriter:     logWriter,
		Router:        router,
		RpcClient:     rpcClient,
		IsLocal:       true,
		InitialEnv:    initialEnv,
		JobManagerMap: make(map[string]*JobManagerConnection),
		SockName:      sockName,
	}
}

func (*ServerImpl) WshServerImpl() {}

func (impl *ServerImpl) Log(format string, args ...interface{}) {
	if impl.LogWriter != nil {
		fmt.Fprintf(impl.LogWriter, format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (impl *ServerImpl) MessageCommand(ctx context.Context, data wshrpc.CommandMessageData) error {
	impl.Log("[message] %q\n", data.Message)
	return nil
}

func (impl *ServerImpl) StreamTestCommand(ctx context.Context) chan wshrpc.RespOrErrorUnion[int] {
	ch := make(chan wshrpc.RespOrErrorUnion[int], 16)
	go func() {
		defer close(ch)
		idx := 0
		for {
			ch <- wshrpc.RespOrErrorUnion[int]{Response: idx}
			idx++
			if idx == 1000 {
				break
			}
		}
	}()
	return ch
}

func (*ServerImpl) RemoteGetInfoCommand(ctx context.Context) (wshrpc.RemoteInfo, error) {
	return wshutil.GetInfo(), nil
}

func (*ServerImpl) RemoteInstallRcFilesCommand(ctx context.Context) error {
	return wshutil.InstallRcFiles()
}

func (impl *ServerImpl) getWshPath() (string, error) {
	return filepath.Join(wavebase.GetWaveDataDir(), "bin", "wsh"), nil
}

func (impl *ServerImpl) BadgeWatchPidCommand(ctx context.Context, data wshrpc.CommandBadgeWatchPidData) error {
	if data.Pid <= 0 {
		return fmt.Errorf("invalid pid: %d", data.Pid)
	}
	if data.ORef.IsEmpty() {
		return fmt.Errorf("oref is required")
	}
	if data.BadgeId == "" {
		return fmt.Errorf("badgeid is required")
	}
	go func() {
		defer func() {
			panichandler.PanicHandler("BadgeWatchPidCommand", recover())
		}()
		for {
			time.Sleep(time.Second)
			if unixutil.IsPidRunning(data.Pid) {
				continue
			}
			orefStr := data.ORef.String()
			event := wps.WaveEvent{
				Event:  wps.Event_Badge,
				Scopes: []string{orefStr},
				Data: baseds.BadgeEvent{
					ORef:      orefStr,
					ClearById: data.BadgeId,
				},
			}
			wshclient.EventPublishCommand(impl.RpcClient, event, nil)
			log.Printf("BadgeWatchPidCommand: pid %d gone, cleared badge %s for oref %s\n", data.Pid, data.BadgeId, orefStr)
			return
		}
	}()
	return nil
}
