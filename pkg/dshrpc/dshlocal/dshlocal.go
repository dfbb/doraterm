// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshlocal

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

	"github.com/dfbb/doraterm/pkg/baseds"
	"github.com/dfbb/doraterm/pkg/panichandler"
	"github.com/dfbb/doraterm/pkg/util/unixutil"
	"github.com/dfbb/doraterm/pkg/dorabase"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshrpc/wshclient"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

type JobManagerConnection struct {
	JobId     string
	Conn      net.Conn
	WshRpc    *dshutil.WshRpc
	CleanupFn func()
}

type ServerImpl struct {
	LogWriter     io.Writer
	Router        *dshutil.WshRouter
	RpcClient     *dshutil.WshRpc
	IsLocal       bool
	InitialEnv    map[string]string
	JobManagerMap map[string]*JobManagerConnection
	SockName      string
	Lock          sync.Mutex
}

func MakeLocalRpcServerImpl(logWriter io.Writer, router *dshutil.WshRouter, rpcClient *dshutil.WshRpc, initialEnv map[string]string, sockName string) *ServerImpl {
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

func (impl *ServerImpl) MessageCommand(ctx context.Context, data dshrpc.CommandMessageData) error {
	impl.Log("[message] %q\n", data.Message)
	return nil
}

func (impl *ServerImpl) StreamTestCommand(ctx context.Context) chan dshrpc.RespOrErrorUnion[int] {
	ch := make(chan dshrpc.RespOrErrorUnion[int], 16)
	go func() {
		defer close(ch)
		idx := 0
		for {
			ch <- dshrpc.RespOrErrorUnion[int]{Response: idx}
			idx++
			if idx == 1000 {
				break
			}
		}
	}()
	return ch
}

func (*ServerImpl) RemoteGetInfoCommand(ctx context.Context) (dshrpc.RemoteInfo, error) {
	return dshutil.GetInfo(), nil
}

func (*ServerImpl) RemoteInstallRcFilesCommand(ctx context.Context) error {
	return dshutil.InstallRcFiles()
}

func (impl *ServerImpl) getWshPath() (string, error) {
	return filepath.Join(dorabase.GetDoraDataDir(), "bin", "wsh"), nil
}

func (impl *ServerImpl) BadgeWatchPidCommand(ctx context.Context, data dshrpc.CommandBadgeWatchPidData) error {
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
			event := dps.WaveEvent{
				Event:  dps.Event_Badge,
				Scopes: []string{orefStr},
				Data: baseds.BadgeEvent{
					ORef:      orefStr,
					ClearById: data.BadgeId,
				},
			}
			dshclient.EventPublishCommand(impl.RpcClient, event, nil)
			log.Printf("BadgeWatchPidCommand: pid %d gone, cleared badge %s for oref %s\n", data.Pid, data.BadgeId, orefStr)
			return
		}
	}()
	return nil
}
