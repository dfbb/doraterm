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

type WshServer struct{}

func (*WshServer) WshServerImpl() {}

var WshServerImpl = WshServer{}

var waveSrvClient_Singleton *dshutil.WshRpc
var waveSrvClient_Once = &sync.Once{}
var waveSrvClient_RouteId string

func GetBareRpcClient() *dshutil.WshRpc {
	waveSrvClient_Once.Do(func() {
		waveSrvClient_Singleton = dshutil.MakeWshRpc(dshrpc.RpcContext{}, &WshServerImpl, "bare-client")
		waveSrvClient_RouteId = fmt.Sprintf("bare:%s", uuid.New().String())
		// we can safely ignore the error from RegisterTrustedLeaf since the route is valid
		dshutil.DefaultRouter.RegisterTrustedLeaf(waveSrvClient_Singleton, waveSrvClient_RouteId)
		dps.Broker.SetClient(dshutil.DefaultRouter)
	})
	return waveSrvClient_Singleton
}

func GetBareRpcClientRouteId() string {
	GetBareRpcClient()
	return waveSrvClient_RouteId
}
