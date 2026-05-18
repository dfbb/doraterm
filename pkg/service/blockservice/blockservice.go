// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package blockservice

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/dfbb/doraterm/pkg/blockcontroller"
	"github.com/dfbb/doraterm/pkg/filestore"
	"github.com/dfbb/doraterm/pkg/tsgen/tsgenmeta"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dcore"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dstore"
)

type BlockService struct{}

const DefaultTimeout = 2 * time.Second

var BlockServiceInstance = &BlockService{}

func (bs *BlockService) SendCommand_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		Desc:     "send command to block",
		ArgNames: []string{"blockid", "cmd"},
	}
}

func (bs *BlockService) GetControllerStatus(ctx context.Context, blockId string) (*blockcontroller.BlockControllerRuntimeStatus, error) {
	return blockcontroller.GetBlockControllerRuntimeStatus(blockId), nil
}

func (*BlockService) SaveTerminalState_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		Desc:     "save the terminal state to a blockfile",
		ArgNames: []string{"ctx", "blockId", "state", "stateType", "ptyOffset", "termSize"},
	}
}

func (bs *BlockService) SaveTerminalState(ctx context.Context, blockId string, state string, stateType string, ptyOffset int64, termSize doraobj.TermSize) error {
	_, err := dstore.DBMustGet[*doraobj.Block](ctx, blockId)
	if err != nil {
		return err
	}
	if stateType != "full" && stateType != "preview" {
		return fmt.Errorf("invalid state type: %q", stateType)
	}
	// ignore MakeFile error (already exists is ok)
	filestore.WFS.MakeFile(ctx, blockId, "cache:term:"+stateType, nil, dshrpc.FileOpts{})
	err = filestore.WFS.WriteFile(ctx, blockId, "cache:term:"+stateType, []byte(state))
	if err != nil {
		return fmt.Errorf("cannot save terminal state: %w", err)
	}
	fileMeta := dshrpc.FileMeta{
		"ptyoffset": ptyOffset,
		"termsize":  termSize,
	}
	err = filestore.WFS.WriteMeta(ctx, blockId, "cache:term:"+stateType, fileMeta, true)
	if err != nil {
		return fmt.Errorf("cannot save terminal state meta: %w", err)
	}
	return nil
}

func (*BlockService) CleanupOrphanedBlocks_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		Desc:     "queue a layout action to cleanup orphaned blocks in the tab",
		ArgNames: []string{"ctx", "tabId"},
	}
}

func (bs *BlockService) CleanupOrphanedBlocks(ctx context.Context, tabId string) (doraobj.UpdatesRtnType, error) {
	ctx = doraobj.ContextWithUpdates(ctx)
	layoutAction := doraobj.LayoutActionData{
		ActionType: dcore.LayoutActionDataType_CleanupOrphaned,
		ActionId:   uuid.NewString(),
	}
	err := dcore.QueueLayoutActionForTab(ctx, tabId, layoutAction)
	if err != nil {
		return nil, fmt.Errorf("error queuing cleanup layout action: %w", err)
	}
	return doraobj.ContextGetUpdatesRtn(ctx), nil
}
