// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshclient

import (
	"context"
	"errors"

	"github.com/dfbb/doraterm/pkg/panichandler"
	"github.com/dfbb/doraterm/pkg/util/utilfn"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

func sendRpcRequestCallHelper[T any](w *dshutil.DshRpc, command string, data interface{}, opts *dshrpc.RpcOpts) (T, error) {
	if opts == nil {
		opts = &dshrpc.RpcOpts{}
	}
	var respData T
	if w == nil {
		return respData, errors.New("nil wshrpc passed to wshclient")
	}
	if opts.NoResponse {
		err := w.SendCommand(command, data, opts)
		if err != nil {
			return respData, err
		}
		return respData, nil
	}
	resp, err := w.SendRpcRequest(command, data, opts)
	if err != nil {
		return respData, err
	}
	err = utilfn.ReUnmarshal(&respData, resp)
	if err != nil {
		return respData, err
	}
	return respData, nil
}

func rtnErr[T any](ch chan dshrpc.RespOrErrorUnion[T], err error) {
	go func() {
		defer func() {
			panichandler.PanicHandler("wshclientutil:rtnErr", recover())
		}()
		ch <- dshrpc.RespOrErrorUnion[T]{Error: err}
		close(ch)
	}()
}

func sendRpcRequestResponseStreamHelper[T any](w *dshutil.DshRpc, command string, data interface{}, opts *dshrpc.RpcOpts) chan dshrpc.RespOrErrorUnion[T] {
	if opts == nil {
		opts = &dshrpc.RpcOpts{}
	}
	respChan := make(chan dshrpc.RespOrErrorUnion[T], 32)
	if w == nil {
		rtnErr(respChan, errors.New("nil wshrpc passed to wshclient"))
		return respChan
	}
	reqHandler, err := w.SendComplexRequest(command, data, opts)
	if err != nil {
		rtnErr(respChan, err)
		return respChan
	}
	opts.StreamCancelFn = func(ctx context.Context) error {
		return reqHandler.SendCancel(ctx)
	}
	go func() {
		defer func() {
			panichandler.PanicHandler("sendRpcRequestResponseStreamHelper", recover())
		}()
		defer close(respChan)
		for {
			if reqHandler.ResponseDone() {
				break
			}
			resp, err := reqHandler.NextResponse()
			if err != nil {
				respChan <- dshrpc.RespOrErrorUnion[T]{Error: err}
				break
			}
			var respData T
			err = utilfn.ReUnmarshal(&respData, resp)
			if err != nil {
				respChan <- dshrpc.RespOrErrorUnion[T]{Error: err}
				break
			}
			respChan <- dshrpc.RespOrErrorUnion[T]{Response: respData}
		}
	}()
	return respChan
}
