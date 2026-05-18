// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshserver

import (
	"sync"

	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

const (
	DefaultOutputChSize = 32
	DefaultInputChSize  = 32
)

var waveSrvClient_Singleton *dshutil.WshRpc
var waveSrvClient_Once = &sync.Once{}

// returns the wavesrv main rpc client singleton
func GetMainRpcClient() *dshutil.WshRpc {
	waveSrvClient_Once.Do(func() {
		waveSrvClient_Singleton = dshutil.MakeWshRpc(dshrpc.RpcContext{}, &WshServerImpl, "main-client")
	})
	return waveSrvClient_Singleton
}
