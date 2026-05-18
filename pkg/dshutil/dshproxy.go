// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshutil

import (
	"fmt"
	"sync"

	"github.com/dfbb/doraterm/pkg/baseds"
	"github.com/dfbb/doraterm/pkg/panichandler"
	"github.com/dfbb/doraterm/pkg/dshrpc"
)

type DshRpcProxy struct {
	Lock         *sync.Mutex
	RpcContext   *dshrpc.RpcContext
	ToRemoteCh   chan []byte
	FromRemoteCh chan baseds.RpcInputChType
	PeerInfo     string
}

func MakeRpcProxy(peerInfo string) *DshRpcProxy {
	return MakeRpcProxyWithSize(peerInfo, DefaultInputChSize, DefaultOutputChSize)
}

func MakeRpcProxyWithSize(peerInfo string, inputChSize int, outputChSize int) *DshRpcProxy {
	return &DshRpcProxy{
		Lock:         &sync.Mutex{},
		ToRemoteCh:   make(chan []byte, inputChSize),
		FromRemoteCh: make(chan baseds.RpcInputChType, outputChSize),
		PeerInfo:     peerInfo,
	}
}

func (p *DshRpcProxy) GetPeerInfo() string {
	return p.PeerInfo
}

func (p *DshRpcProxy) SetPeerInfo(peerInfo string) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.PeerInfo = peerInfo
}

func (p *DshRpcProxy) SendRpcMessage(msg []byte, ingressLinkId baseds.LinkId, debugStr string) bool {
	defer func() {
		panicCtx := "DshRpcProxy.SendRpcMessage"
		if debugStr != "" {
			panicCtx = fmt.Sprintf("%s:%s", panicCtx, debugStr)
		}
		panichandler.PanicHandler(panicCtx, recover())
	}()
	select {
	case p.ToRemoteCh <- msg:
		return true
	default:
		return false
	}
}

func (p *DshRpcProxy) RecvRpcMessage() ([]byte, bool) {
	inputVal, more := <-p.FromRemoteCh
	return inputVal.MsgBytes, more
}
