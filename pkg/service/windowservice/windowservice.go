// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package windowservice

import (
	"context"
	"fmt"
	"time"

	"github.com/dfbb/doraterm/pkg/panichandler"
	"github.com/dfbb/doraterm/pkg/tsgen/tsgenmeta"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dcore"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dstore"
)

const DefaultTimeout = 2 * time.Second

type WindowService struct{}

func (svc *WindowService) GetWindow_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"windowId"},
	}
}

func (svc *WindowService) GetWindow(windowId string) (*doraobj.Window, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	window, err := dstore.DBGet[*doraobj.Window](ctx, windowId)
	if err != nil {
		return nil, fmt.Errorf("error getting window: %w", err)
	}
	return window, nil
}

func (svc *WindowService) CreateWindow_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"ctx", "winSize", "workspaceId"},
	}
}

func (svc *WindowService) CreateWindow(ctx context.Context, winSize *doraobj.WinSize, workspaceId string) (*doraobj.Window, error) {
	window, err := dcore.CreateWindow(ctx, winSize, workspaceId)
	if err != nil {
		return nil, fmt.Errorf("error creating window: %w", err)
	}
	return window, nil
}

func (svc *WindowService) SetWindowPosAndSize_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		Desc:     "set window position and size",
		ArgNames: []string{"ctx", "windowId", "pos", "size"},
	}
}

func (ws *WindowService) SetWindowPosAndSize(ctx context.Context, windowId string, pos *doraobj.Point, size *doraobj.WinSize) (doraobj.UpdatesRtnType, error) {
	if pos == nil && size == nil {
		return nil, nil
	}
	ctx = doraobj.ContextWithUpdates(ctx)
	win, err := dstore.DBMustGet[*doraobj.Window](ctx, windowId)
	if err != nil {
		return nil, err
	}
	if pos != nil {
		win.Pos = *pos
	}
	if size != nil {
		win.WinSize = *size
	}
	win.IsNew = false
	err = dstore.DBUpdate(ctx, win)
	if err != nil {
		return nil, err
	}
	return doraobj.ContextGetUpdatesRtn(ctx), nil
}

func (svc *WindowService) SwitchWorkspace_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"ctx", "windowId", "workspaceId"},
	}
}

func (svc *WindowService) SwitchWorkspace(ctx context.Context, windowId string, workspaceId string) (*doraobj.Workspace, error) {
	ctx = doraobj.ContextWithUpdates(ctx)
	ws, err := dcore.SwitchWorkspace(ctx, windowId, workspaceId)

	updates := doraobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WindowService:SwitchWorkspace:SendUpdateEvents", recover())
		}()
		dps.Broker.SendUpdateEvents(updates)
	}()
	return ws, err
}

func (svc *WindowService) CloseWindow_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"ctx", "windowId", "fromElectron"},
	}
}

func (svc *WindowService) CloseWindow(ctx context.Context, windowId string, fromElectron bool) error {
	ctx = doraobj.ContextWithUpdates(ctx)
	return dcore.CloseWindow(ctx, windowId, fromElectron)
}
