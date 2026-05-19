// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows && !(linux && (mips || mips64))

package doraattach

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/dfbb/doraterm/pkg/util/utilfn"
	"github.com/dfbb/doraterm/pkg/dorabase"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshrpc/dshclient"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

const renderDebounce = 16 * time.Millisecond

type pendingEvent struct {
	at   time.Time
	data []byte
}

type eventBuffer struct {
	mu      sync.Mutex
	pending []pendingEvent
	flushed bool
}

func makeEventBuffer() *eventBuffer {
	return &eventBuffer{}
}

func (b *eventBuffer) flush(cutoff time.Time, w io.Writer) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ev := range b.pending {
		if ev.at.After(cutoff) {
			if _, err := w.Write(ev.data); err != nil {
				return err
			}
		}
	}
	b.pending = nil
	b.flushed = true
	return nil
}

func (b *eventBuffer) write(at time.Time, data []byte, w io.Writer) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.flushed {
		b.pending = append(b.pending, pendingEvent{at: at, data: data})
		return nil
	}
	_, err := w.Write(data)
	return err
}

func (b *eventBuffer) unflush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushed = false
	b.pending = nil
}

type ViewportRenderer struct {
	mu     sync.Mutex
	vp     *Viewport
	w      io.Writer
	tickCh chan struct{}
}

func newViewportRenderer(vp *Viewport, w io.Writer) *ViewportRenderer {
	return &ViewportRenderer{vp: vp, w: w, tickCh: make(chan struct{}, 1)}
}

func (vr *ViewportRenderer) Render() {
	vr.mu.Lock()
	defer vr.mu.Unlock()
	vr.vp.Render(vr.w)
}

func (vr *ViewportRenderer) requestRender() {
	select {
	case vr.tickCh <- struct{}{}:
	default:
	}
}

func (vr *ViewportRenderer) runRenderLoop(ctx context.Context) {
	timer := time.NewTimer(renderDebounce)
	timer.Stop()
	timerArmed := false
	for {
		select {
		case <-ctx.Done():
			if timerArmed {
				timer.Stop()
			}
			return
		case <-vr.tickCh:
			if !timerArmed {
				timer.Reset(renderDebounce)
				timerArmed = true
			}
		case <-timer.C:
			timerArmed = false
			select {
			case <-vr.tickCh:
			default:
			}
			vr.Render()
		}
	}
}

func StreamOutput(ctx context.Context, rpcClient *dshutil.DshRpc, blockId string, vr *ViewportRenderer, resyncCh <-chan struct{}) error {
	vp := vr.vp
	buf := makeEventBuffer()
	blockRef := doraobj.MakeORef(doraobj.OType_Block, blockId).String()

	go vr.runRenderLoop(ctx)

	rpcClient.EventListener.On(dps.Event_BlockFile, func(ev *dps.DoraEvent) {
		var fed dps.WSFileEventData
		if err := utilfn.ReUnmarshal(&fed, ev.Data); err != nil {
			return
		}
		if fed.ZoneId != blockId || fed.FileName != dorabase.BlockFile_Term {
			return
		}
		if fed.FileOp != dps.FileOp_Append {
			return
		}
		data, err := base64.StdEncoding.DecodeString(fed.Data64)
		if err != nil {
			return
		}
		if err := buf.write(time.Now(), data, vp); err != nil {
			return
		}
		vr.requestRender()
	})

	subReq := dps.SubscriptionRequest{
		Event:  dps.Event_BlockFile,
		Scopes: []string{blockRef},
	}
	if err := dshclient.EventSubCommand(rpcClient, subReq, nil); err != nil {
		return fmt.Errorf("subscribing to blockfile events: %w", err)
	}

	if err := loadSnapshotAndFlush(rpcClient, blockId, vp, buf); err != nil {
		return err
	}
	vr.Render()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-resyncCh:
			buf.unflush()
			vp.Reset()
			if err := loadSnapshotAndFlush(rpcClient, blockId, vp, buf); err != nil {
				return err
			}
			vr.Render()
		}
	}
}

func loadSnapshotAndFlush(rpcClient *dshutil.DshRpc, blockId string, vp *Viewport, buf *eventBuffer) error {
	if err := readSnapshot(rpcClient, blockId, vp); err != nil {
		return fmt.Errorf("reading snapshot: %w", err)
	}
	cutoff := time.Now()
	if err := buf.flush(cutoff, vp); err != nil {
		return err
	}
	return nil
}

func readSnapshot(rpcClient *dshutil.DshRpc, blockId string, w io.Writer) error {
	broker := rpcClient.StreamBroker
	if broker == nil {
		return fmt.Errorf("stream broker not available")
	}

	readerRouteId, err := dshclient.ControlGetRouteIdCommand(rpcClient, &dshrpc.RpcOpts{Route: dshutil.ControlRoute})
	if err != nil {
		return fmt.Errorf("getting route id: %w", err)
	}
	if readerRouteId == "" {
		return fmt.Errorf("no route to receive data")
	}

	reader, streamMeta := broker.CreateStreamReader(readerRouteId, "", 64*1024)
	defer reader.Close()

	data := dshrpc.CommandDoraFileReadStreamData{
		ZoneId:     blockId,
		Name:       dorabase.BlockFile_Term,
		StreamMeta: *streamMeta,
	}

	_, err = dshclient.DoraFileReadStreamCommand(rpcClient, data, nil)
	if err != nil {
		return fmt.Errorf("starting stream read: %w", err)
	}

	_, err = io.Copy(w, reader)
	if err != nil {
		return fmt.Errorf("reading stream: %w", err)
	}
	return nil
}
