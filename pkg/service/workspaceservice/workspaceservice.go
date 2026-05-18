// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"context"
	"fmt"
	"time"

	"github.com/dfbb/doraterm/pkg/blockcontroller"
	"github.com/dfbb/doraterm/pkg/panichandler"
	"github.com/dfbb/doraterm/pkg/tsgen/tsgenmeta"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dcore"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dstore"
)

const DefaultTimeout = 2 * time.Second

type WorkspaceService struct{}

func (svc *WorkspaceService) CreateWorkspace_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames:   []string{"ctx", "name", "icon", "color", "applyDefaults"},
		ReturnDesc: "workspaceId",
	}
}

func (svc *WorkspaceService) CreateWorkspace(ctx context.Context, name string, icon string, color string, applyDefaults bool) (string, error) {
	newWS, err := dcore.CreateWorkspace(ctx, name, icon, color, applyDefaults, false)
	if err != nil {
		return "", fmt.Errorf("error creating workspace: %w", err)
	}
	return newWS.OID, nil
}

func (svc *WorkspaceService) UpdateWorkspace_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"ctx", "workspaceId", "name", "icon", "color", "applyDefaults"},
	}
}

func (svc *WorkspaceService) UpdateWorkspace(ctx context.Context, workspaceId string, name string, icon string, color string, applyDefaults bool) (doraobj.UpdatesRtnType, error) {
	ctx = doraobj.ContextWithUpdates(ctx)
	_, updated, err := dcore.UpdateWorkspace(ctx, workspaceId, name, icon, color, applyDefaults)
	if err != nil {
		return nil, fmt.Errorf("error updating workspace: %w", err)
	}
	if !updated {
		return nil, nil
	}

	dps.Broker.Publish(dps.WaveEvent{
		Event: dps.Event_WorkspaceUpdate,
	})

	updates := doraobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:UpdateWorkspace:SendUpdateEvents", recover())
		}()
		dps.Broker.SendUpdateEvents(updates)
	}()
	return updates, nil
}

func (svc *WorkspaceService) GetWorkspace_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames:   []string{"workspaceId"},
		ReturnDesc: "workspace",
	}
}

func (svc *WorkspaceService) GetWorkspace(workspaceId string) (*doraobj.Workspace, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ws, err := dstore.DBGet[*doraobj.Workspace](ctx, workspaceId)
	if err != nil {
		return nil, fmt.Errorf("error getting workspace: %w", err)
	}
	return ws, nil
}

func (svc *WorkspaceService) DeleteWorkspace_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"workspaceId"},
	}
}

func (svc *WorkspaceService) DeleteWorkspace(workspaceId string) (doraobj.UpdatesRtnType, string, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ctx = doraobj.ContextWithUpdates(ctx)
	deleted, claimableWorkspace, err := dcore.DeleteWorkspace(ctx, workspaceId, true)
	if claimableWorkspace != "" {
		return nil, claimableWorkspace, nil
	}
	if err != nil {
		return nil, claimableWorkspace, fmt.Errorf("error deleting workspace: %w", err)
	}
	if !deleted {
		return nil, claimableWorkspace, nil
	}
	updates := doraobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:DeleteWorkspace:SendUpdateEvents", recover())
		}()
		dps.Broker.SendUpdateEvents(updates)
	}()
	return updates, claimableWorkspace, nil
}

func (svc *WorkspaceService) ListWorkspaces() (doraobj.WorkspaceList, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	return dcore.ListWorkspaces(ctx)
}

func (svc *WorkspaceService) CreateTab_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames:   []string{"workspaceId", "tabName", "activateTab"},
		ReturnDesc: "tabId",
	}
}

func (svc *WorkspaceService) GetColors_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ReturnDesc: "colors",
	}
}

func (svc *WorkspaceService) GetColors() []string {
	return dcore.WorkspaceColors[:]
}

func (svc *WorkspaceService) GetIcons_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ReturnDesc: "icons",
	}
}

func (svc *WorkspaceService) GetIcons() []string {
	return dcore.WorkspaceIcons[:]
}

func (svc *WorkspaceService) CreateTab(workspaceId string, tabName string, activateTab bool) (string, doraobj.UpdatesRtnType, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ctx = doraobj.ContextWithUpdates(ctx)
	tabId, err := dcore.CreateTab(ctx, workspaceId, tabName, activateTab, false)
	if err != nil {
		return "", nil, fmt.Errorf("error creating tab: %w", err)
	}
	updates := doraobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:CreateTab:SendUpdateEvents", recover())
		}()
		dps.Broker.SendUpdateEvents(updates)
	}()
	return tabId, updates, nil
}

func (svc *WorkspaceService) SetActiveTab_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames: []string{"workspaceId", "tabId"},
	}
}

func (svc *WorkspaceService) SetActiveTab(workspaceId string, tabId string) (doraobj.UpdatesRtnType, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()
	ctx = doraobj.ContextWithUpdates(ctx)
	err := dcore.SetActiveTab(ctx, workspaceId, tabId)
	if err != nil {
		return nil, fmt.Errorf("error setting active tab: %w", err)
	}
	// check all blocks in tab and start controllers (if necessary)
	tab, err := dstore.DBMustGet[*doraobj.Tab](ctx, tabId)
	if err != nil {
		return nil, fmt.Errorf("error getting tab: %w", err)
	}
	blockORefs := tab.GetBlockORefs()
	blocks, err := dstore.DBSelectORefs(ctx, blockORefs)
	if err != nil {
		return nil, fmt.Errorf("error getting tab blocks: %w", err)
	}
	updates := doraobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:SetActiveTab:SendUpdateEvents", recover())
		}()
		dps.Broker.SendUpdateEvents(updates)
	}()
	var extraUpdates doraobj.UpdatesRtnType
	extraUpdates = append(extraUpdates, updates...)
	extraUpdates = append(extraUpdates, doraobj.MakeUpdate(tab))
	extraUpdates = append(extraUpdates, doraobj.MakeUpdates(blocks)...)
	return extraUpdates, nil
}

type CloseTabRtnType struct {
	CloseWindow    bool   `json:"closewindow,omitempty"`
	NewActiveTabId string `json:"newactivetabid,omitempty"`
}

func (svc *WorkspaceService) CloseTab_Meta() tsgenmeta.MethodMeta {
	return tsgenmeta.MethodMeta{
		ArgNames:   []string{"ctx", "workspaceId", "tabId", "fromElectron"},
		ReturnDesc: "CloseTabRtn",
	}
}

// returns the new active tabid
func (svc *WorkspaceService) CloseTab(ctx context.Context, workspaceId string, tabId string, fromElectron bool) (*CloseTabRtnType, doraobj.UpdatesRtnType, error) {
	ctx = doraobj.ContextWithUpdates(ctx)
	tab, err := dstore.DBGet[*doraobj.Tab](ctx, tabId)
	if err == nil && tab != nil {
		go func() {
			for _, blockId := range tab.BlockIds {
				blockcontroller.DestroyBlockController(blockId)
			}
		}()
	}
	newActiveTabId, err := dcore.DeleteTab(ctx, workspaceId, tabId, true)
	if err != nil {
		return nil, nil, fmt.Errorf("error closing tab: %w", err)
	}
	rtn := &CloseTabRtnType{}
	if newActiveTabId == "" {
		rtn.CloseWindow = true
	} else {
		rtn.NewActiveTabId = newActiveTabId
	}
	updates := doraobj.ContextGetUpdatesRtn(ctx)
	go func() {
		defer func() {
			panichandler.PanicHandler("WorkspaceService:CloseTab:SendUpdateEvents", recover())
		}()
		dps.Broker.SendUpdateEvents(updates)
	}()
	return rtn, updates, nil
}
