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

var waveSrvClient_Singleton *dshutil.DshRpc
var waveSrvClient_Once = &sync.Once{}

// returns the dorasrv main rpc client singleton
func GetMainRpcClient() *dshutil.DshRpc {
	waveSrvClient_Once.Do(func() {
		waveSrvClient_Singleton = dshutil.MakeDshRpc(dshrpc.RpcContext{}, &DshServerImpl, "main-client")
	})
	return waveSrvClient_Singleton
}
