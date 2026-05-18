// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshutil

import (
	"github.com/dfbb/doraterm/pkg/dshrpc"
)

type WshRpcStreamClientAdapter struct {
	rpc *WshRpc
}

func (a *WshRpcStreamClientAdapter) StreamDataAckCommand(data dshrpc.CommandStreamAckData, opts *dshrpc.RpcOpts) error {
	return a.rpc.SendCommand("streamdataack", data, opts)
}

func (a *WshRpcStreamClientAdapter) StreamDataCommand(data dshrpc.CommandStreamData, opts *dshrpc.RpcOpts) error {
	return a.rpc.SendCommand("streamdata", data, opts)
}

func AdaptWshRpc(rpc *WshRpc) *WshRpcStreamClientAdapter {
	return &WshRpcStreamClientAdapter{rpc: rpc}
}
