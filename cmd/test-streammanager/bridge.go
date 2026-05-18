// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/dfbb/doraterm/pkg/dshrpc"
)

// WriterBridge - used by the writer broker
// Sends data to the pipe, receives acks from the pipe
type WriterBridge struct {
	pipe *DeliveryPipe
}

func (b *WriterBridge) StreamDataCommand(data dshrpc.CommandStreamData, opts *dshrpc.RpcOpts) error {
	b.pipe.EnqueueData(data)
	return nil
}

func (b *WriterBridge) StreamDataAckCommand(ack dshrpc.CommandStreamAckData, opts *dshrpc.RpcOpts) error {
	return fmt.Errorf("writer bridge should not send acks")
}

// ReaderBridge - used by the reader broker
// Sends acks to the pipe, receives data from the pipe
type ReaderBridge struct {
	pipe *DeliveryPipe
}

func (b *ReaderBridge) StreamDataCommand(data dshrpc.CommandStreamData, opts *dshrpc.RpcOpts) error {
	return fmt.Errorf("reader bridge should not send data")
}

func (b *ReaderBridge) StreamDataAckCommand(ack dshrpc.CommandStreamAckData, opts *dshrpc.RpcOpts) error {
	b.pipe.EnqueueAck(ack)
	return nil
}
