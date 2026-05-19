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

var doraSrvClient_Singleton *dshutil.DshRpc
var doraSrvClient_Once = &sync.Once{}

// returns the dorasrv main rpc client singleton
func GetMainRpcClient() *dshutil.DshRpc {
	doraSrvClient_Once.Do(func() {
		doraSrvClient_Singleton = dshutil.MakeDshRpc(dshrpc.RpcContext{}, &DshServerImpl, "main-client")
	})
	return doraSrvClient_Singleton
}
