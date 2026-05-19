// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshclient

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

type DshServer struct{}

func (*DshServer) DshServerImpl() {}

var DshServerImpl = DshServer{}

var doraSrvClient_Singleton *dshutil.DshRpc
var doraSrvClient_Once = &sync.Once{}
var doraSrvClient_RouteId string

func GetBareRpcClient() *dshutil.DshRpc {
	doraSrvClient_Once.Do(func() {
		doraSrvClient_Singleton = dshutil.MakeDshRpc(dshrpc.RpcContext{}, &DshServerImpl, "bare-client")
		doraSrvClient_RouteId = fmt.Sprintf("bare:%s", uuid.New().String())
		// we can safely ignore the error from RegisterTrustedLeaf since the route is valid
		dshutil.DefaultRouter.RegisterTrustedLeaf(doraSrvClient_Singleton, doraSrvClient_RouteId)
		dps.Broker.SetClient(dshutil.DefaultRouter)
	})
	return doraSrvClient_Singleton
}

func GetBareRpcClientRouteId() string {
	GetBareRpcClient()
	return doraSrvClient_RouteId
}
