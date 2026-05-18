// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshutil

import (
	"github.com/dfbb/doraterm/pkg/dshrpc"
)

type DshRpcStreamClientAdapter struct {
	rpc *DshRpc
}

func (a *DshRpcStreamClientAdapter) StreamDataAckCommand(data dshrpc.CommandStreamAckData, opts *dshrpc.RpcOpts) error {
	return a.rpc.SendCommand("streamdataack", data, opts)
}

func (a *DshRpcStreamClientAdapter) StreamDataCommand(data dshrpc.CommandStreamData, opts *dshrpc.RpcOpts) error {
	return a.rpc.SendCommand("streamdata", data, opts)
}

func AdaptDshRpc(rpc *DshRpc) *DshRpcStreamClientAdapter {
	return &DshRpcStreamClientAdapter{rpc: rpc}
}
