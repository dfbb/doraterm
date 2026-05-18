// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshserver

// this file contains the implementation of the wsh server methods

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/skratchdot/open-golang/open"
	"github.com/dfbb/doraterm/pkg/baseds"
	"github.com/dfbb/doraterm/pkg/blockcontroller"
	"github.com/dfbb/doraterm/pkg/blocklogger"
	"github.com/dfbb/doraterm/pkg/filestore"
	"github.com/dfbb/doraterm/pkg/genconn"
	"github.com/dfbb/doraterm/pkg/jobcontroller"
	"github.com/dfbb/doraterm/pkg/panichandler"
	"github.com/dfbb/doraterm/pkg/remote/fileshare/wshfs"
	"github.com/dfbb/doraterm/pkg/secretstore"
	"github.com/dfbb/doraterm/pkg/telemetry"
	"github.com/dfbb/doraterm/pkg/telemetry/telemetrydata"
	"github.com/dfbb/doraterm/pkg/util/envutil"
	"github.com/dfbb/doraterm/pkg/dorabase"
	"github.com/dfbb/doraterm/pkg/dorajwt"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/wcloud"
	"github.com/dfbb/doraterm/pkg/dconfig"
	"github.com/dfbb/doraterm/pkg/dcore"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshutil"
	"github.com/dfbb/doraterm/pkg/dstore"
)


type WshServer struct{}

func (*WshServer) WshServerImpl() {}

var WshServerImpl = WshServer{}

func (ws *WshServer) GetJwtPublicKeyCommand(ctx context.Context) (string, error) {
	return dorajwt.GetPublicKeyBase64(), nil
}

func (ws *WshServer) TestCommand(ctx context.Context, data string) error {
	defer func() {
		panichandler.PanicHandler("TestCommand", recover())
	}()
	rpcSource := dshutil.GetRpcSourceFromContext(ctx)
	log.Printf("TEST src:%s | %s\n", rpcSource, data)
	return nil
}

func (ws *WshServer) TestMultiArgCommand(ctx context.Context, arg1 string, arg2 int, arg3 bool) (string, error) {
	defer func() {
		panichandler.PanicHandler("TestMultiArgCommand", recover())
	}()
	rpcSource := dshutil.GetRpcSourceFromContext(ctx)
	rtn := fmt.Sprintf("src:%s arg1:%q arg2:%d arg3:%t", rpcSource, arg1, arg2, arg3)
	log.Printf("TESTMULTI %s\n", rtn)
	return rtn, nil
}

// for testing
func (ws *WshServer) MessageCommand(ctx context.Context, data dshrpc.CommandMessageData) error {
	log.Printf("MESSAGE: %s\n", data.Message)
	return nil
}

// for testing
func (ws *WshServer) StreamTestCommand(ctx context.Context) chan dshrpc.RespOrErrorUnion[int] {
	rtn := make(chan dshrpc.RespOrErrorUnion[int])
	go func() {
		defer func() {
			panichandler.PanicHandler("StreamTestCommand", recover())
		}()
		for i := 1; i <= 5; i++ {
			rtn <- dshrpc.RespOrErrorUnion[int]{Response: i}
			time.Sleep(1 * time.Second)
		}
		close(rtn)
	}()
	return rtn
}

func MakePlotData(ctx context.Context, blockId string) error {
	block, err := dstore.DBMustGet[*doraobj.Block](ctx, blockId)
	if err != nil {
		return err
	}
	viewName := block.Meta.GetString(doraobj.MetaKey_View, "")
	if viewName != "cpuplot" && viewName != "sysinfo" {
		return fmt.Errorf("invalid view type: %s", viewName)
	}
	return filestore.WFS.MakeFile(ctx, blockId, "cpuplotdata", nil, dshrpc.FileOpts{})
}

func SavePlotData(ctx context.Context, blockId string, history string) error {
	block, err := dstore.DBMustGet[*doraobj.Block](ctx, blockId)
	if err != nil {
		return err
	}
	viewName := block.Meta.GetString(doraobj.MetaKey_View, "")
	if viewName != "cpuplot" && viewName != "sysinfo" {
		return fmt.Errorf("invalid view type: %s", viewName)
	}
	// todo: interpret the data being passed
	// for now, this is just to throw an error if the block was closed
	historyBytes, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("unable to serialize plot data: %v", err)
	}
	// ignore MakeFile error (already exists is ok)
	return filestore.WFS.WriteFile(ctx, blockId, "cpuplotdata", historyBytes)
}

func (ws *WshServer) GetMetaCommand(ctx context.Context, data dshrpc.CommandGetMetaData) (doraobj.MetaMapType, error) {
	obj, err := dstore.DBGetORef(ctx, data.ORef)
	if err != nil {
		return nil, fmt.Errorf("error getting object: %w", err)
	}
	if obj == nil {
		return nil, fmt.Errorf("object not found: %s", data.ORef)
	}
	return doraobj.GetMeta(obj), nil
}

func (ws *WshServer) UpdateTabNameCommand(ctx context.Context, tabId string, newName string) error {
	oref := doraobj.ORef{OType: doraobj.OType_Tab, OID: tabId}
	err := dstore.UpdateTabName(ctx, tabId, newName)
	if err != nil {
		return fmt.Errorf("error updating tab name: %w", err)
	}
	dcore.SendWaveObjUpdate(oref)
	return nil
}

func (ws *WshServer) UpdateWorkspaceTabIdsCommand(ctx context.Context, workspaceId string, tabIds []string) error {
	oref := doraobj.ORef{OType: doraobj.OType_Workspace, OID: workspaceId}
	err := dcore.UpdateWorkspaceTabIds(ctx, workspaceId, tabIds)
	if err != nil {
		return fmt.Errorf("error updating workspace tab ids: %w", err)
	}
	dcore.SendWaveObjUpdate(oref)
	return nil
}

func (ws *WshServer) SetMetaCommand(ctx context.Context, data dshrpc.CommandSetMetaData) error {
	log.Printf("SetMetaCommand: %s | %v\n", data.ORef, data.Meta)
	oref := data.ORef
	err := dstore.UpdateObjectMeta(ctx, oref, data.Meta, false)
	if err != nil {
		return fmt.Errorf("error updating object meta: %w", err)
	}
	dcore.SendWaveObjUpdate(oref)
	return nil
}

func (ws *WshServer) GetRTInfoCommand(ctx context.Context, data dshrpc.CommandGetRTInfoData) (*doraobj.ObjRTInfo, error) {
	return dstore.GetRTInfo(data.ORef), nil
}

func (ws *WshServer) SetRTInfoCommand(ctx context.Context, data dshrpc.CommandSetRTInfoData) error {
	if data.Delete {
		dstore.DeleteRTInfo(data.ORef)
		return nil
	}
	dstore.SetRTInfo(data.ORef, data.Data)
	return nil
}

func (ws *WshServer) ResolveIdsCommand(ctx context.Context, data dshrpc.CommandResolveIdsData) (dshrpc.CommandResolveIdsRtnData, error) {
	rtn := dshrpc.CommandResolveIdsRtnData{}
	rtn.ResolvedIds = make(map[string]doraobj.ORef)
	var firstErr error
	for _, simpleId := range data.Ids {
		oref, err := resolveSimpleId(ctx, data, simpleId)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if oref == nil {
			continue
		}
		rtn.ResolvedIds[simpleId] = *oref
	}
	if firstErr != nil && len(data.Ids) == 1 {
		return rtn, firstErr
	}
	return rtn, nil
}

func (ws *WshServer) CreateBlockCommand(ctx context.Context, data dshrpc.CommandCreateBlockData) (*doraobj.ORef, error) {
	ctx = doraobj.ContextWithUpdates(ctx)
	tabId := data.TabId
	blockData, err := dcore.CreateBlock(ctx, tabId, data.BlockDef, data.RtOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating block: %w", err)
	}
	var layoutAction *doraobj.LayoutActionData
	if data.TargetBlockId != "" {
		switch data.TargetAction {
		case "replace":
			layoutAction = &doraobj.LayoutActionData{
				ActionType:    dcore.LayoutActionDataType_Replace,
				TargetBlockId: data.TargetBlockId,
				BlockId:       blockData.OID,
				Focused:       data.Focused,
			}
			err = dcore.DeleteBlock(ctx, data.TargetBlockId, false)
			if err != nil {
				return nil, fmt.Errorf("error deleting block (trying to do block replace): %w", err)
			}
		case "splitright":
			layoutAction = &doraobj.LayoutActionData{
				ActionType:    dcore.LayoutActionDataType_SplitHorizontal,
				BlockId:       blockData.OID,
				TargetBlockId: data.TargetBlockId,
				Position:      "after",
				Focused:       data.Focused,
			}
		case "splitleft":
			layoutAction = &doraobj.LayoutActionData{
				ActionType:    dcore.LayoutActionDataType_SplitHorizontal,
				BlockId:       blockData.OID,
				TargetBlockId: data.TargetBlockId,
				Position:      "before",
				Focused:       data.Focused,
			}
		case "splitup":
			layoutAction = &doraobj.LayoutActionData{
				ActionType:    dcore.LayoutActionDataType_SplitVertical,
				BlockId:       blockData.OID,
				TargetBlockId: data.TargetBlockId,
				Position:      "before",
				Focused:       data.Focused,
			}
		case "splitdown":
			layoutAction = &doraobj.LayoutActionData{
				ActionType:    dcore.LayoutActionDataType_SplitVertical,
				BlockId:       blockData.OID,
				TargetBlockId: data.TargetBlockId,
				Position:      "after",
				Focused:       data.Focused,
			}
		default:
			return nil, fmt.Errorf("invalid target action: %s", data.TargetAction)
		}
	} else {
		layoutAction = &doraobj.LayoutActionData{
			ActionType: dcore.LayoutActionDataType_Insert,
			BlockId:    blockData.OID,
			Magnified:  data.Magnified,
			Ephemeral:  data.Ephemeral,
			Focused:    data.Focused,
		}
	}
	err = dcore.QueueLayoutActionForTab(ctx, tabId, *layoutAction)
	if err != nil {
		return nil, fmt.Errorf("error queuing layout action: %w", err)
	}
	updates := doraobj.ContextGetUpdatesRtn(ctx)
	dps.Broker.SendUpdateEvents(updates)
	return &doraobj.ORef{OType: doraobj.OType_Block, OID: blockData.OID}, nil
}

func (ws *WshServer) CreateSubBlockCommand(ctx context.Context, data dshrpc.CommandCreateSubBlockData) (*doraobj.ORef, error) {
	parentBlockId := data.ParentBlockId
	blockData, err := dcore.CreateSubBlock(ctx, parentBlockId, data.BlockDef)
	if err != nil {
		return nil, fmt.Errorf("error creating block: %w", err)
	}
	blockRef := &doraobj.ORef{OType: doraobj.OType_Block, OID: blockData.OID}
	return blockRef, nil
}

func (ws *WshServer) ControllerDestroyCommand(ctx context.Context, blockId string) error {
	blockcontroller.DestroyBlockController(blockId)
	return nil
}

func (ws *WshServer) ControllerResyncCommand(ctx context.Context, data dshrpc.CommandControllerResyncData) error {
	ctx = genconn.ContextWithConnData(ctx, data.BlockId)
	ctx = termCtxWithLogBlockId(ctx, data.BlockId)
	return blockcontroller.ResyncController(ctx, data.TabId, data.BlockId, data.RtOpts, data.ForceRestart)
}

func (ws *WshServer) ControllerInputCommand(ctx context.Context, data dshrpc.CommandBlockInputData) error {
	inputUnion := &blockcontroller.BlockInputUnion{
		SigName:  data.SigName,
		TermSize: data.TermSize,
	}
	if len(data.InputData64) > 0 {
		inputBuf := make([]byte, base64.StdEncoding.DecodedLen(len(data.InputData64)))
		nw, err := base64.StdEncoding.Decode(inputBuf, []byte(data.InputData64))
		if err != nil {
			return fmt.Errorf("error decoding input data: %w", err)
		}
		inputUnion.InputData = inputBuf[:nw]
	}
	return blockcontroller.SendInput(data.BlockId, inputUnion)
}

func (ws *WshServer) ControllerAppendOutputCommand(ctx context.Context, data dshrpc.CommandControllerAppendOutputData) error {
	outputBuf := make([]byte, base64.StdEncoding.DecodedLen(len(data.Data64)))
	nw, err := base64.StdEncoding.Decode(outputBuf, []byte(data.Data64))
	if err != nil {
		return fmt.Errorf("error decoding output data: %w", err)
	}
	err = blockcontroller.HandleAppendBlockFile(data.BlockId, dorabase.BlockFile_Term, outputBuf[:nw])
	if err != nil {
		return fmt.Errorf("error appending to block file: %w", err)
	}
	return nil
}

func (ws *WshServer) FileCreateCommand(ctx context.Context, data dshrpc.FileData) error {
	data.Data64 = ""
	err := wshfs.PutFile(ctx, data)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	return nil
}

func (ws *WshServer) FileMkdirCommand(ctx context.Context, data dshrpc.FileData) error {
	return wshfs.Mkdir(ctx, data.Info.Path)
}

func (ws *WshServer) FileDeleteCommand(ctx context.Context, data dshrpc.CommandDeleteFileData) error {
	return wshfs.Delete(ctx, data)
}

func (ws *WshServer) FileInfoCommand(ctx context.Context, data dshrpc.FileData) (*dshrpc.FileInfo, error) {
	return wshfs.Stat(ctx, data.Info.Path)
}

func (ws *WshServer) FileListCommand(ctx context.Context, data dshrpc.FileListData) ([]*dshrpc.FileInfo, error) {
	return wshfs.ListEntries(ctx, data.Path, data.Opts)
}

func (ws *WshServer) FileListStreamCommand(ctx context.Context, data dshrpc.FileListData) <-chan dshrpc.RespOrErrorUnion[dshrpc.CommandRemoteListEntriesRtnData] {
	return wshfs.ListEntriesStream(ctx, data.Path, data.Opts)
}

func (ws *WshServer) FileWriteCommand(ctx context.Context, data dshrpc.FileData) error {
	return wshfs.PutFile(ctx, data)
}

func (ws *WshServer) FileReadCommand(ctx context.Context, data dshrpc.FileData) (*dshrpc.FileData, error) {
	return wshfs.Read(ctx, data)
}

func (ws *WshServer) FileStreamCommand(ctx context.Context, data dshrpc.CommandFileStreamData) (*dshrpc.FileInfo, error) {
	return wshfs.FileStream(ctx, data)
}

func (ws *WshServer) FileCopyCommand(ctx context.Context, data dshrpc.CommandFileCopyData) error {
	return wshfs.Copy(ctx, data)
}

func (ws *WshServer) FileMoveCommand(ctx context.Context, data dshrpc.CommandFileCopyData) error {
	return wshfs.Move(ctx, data)
}

func (ws *WshServer) FileAppendCommand(ctx context.Context, data dshrpc.FileData) error {
	return wshfs.Append(ctx, data)
}

func (ws *WshServer) FileJoinCommand(ctx context.Context, paths []string) (*dshrpc.FileInfo, error) {
	if len(paths) < 2 {
		if len(paths) == 0 {
			return nil, fmt.Errorf("no paths provided")
		}
		return wshfs.Stat(ctx, paths[0])
	}
	return wshfs.Join(ctx, paths[0], paths[1:]...)
}


func (ws *WshServer) GetTempDirCommand(ctx context.Context, data dshrpc.CommandGetTempDirData) (string, error) {
	tempDir := os.TempDir()
	if data.FileName != "" {
		// Reduce to a simple file name to avoid absolute paths or traversal
		name := filepath.Base(data.FileName)
		// Normalize/trim any stray separators and whitespace
		name = strings.Trim(name, `/\`+" ")
		if name == "" || name == "." {
			return tempDir, nil
		}
		return filepath.Join(tempDir, name), nil
	}
	return tempDir, nil
}

func (ws *WshServer) WriteTempFileCommand(ctx context.Context, data dshrpc.CommandWriteTempFileData) (string, error) {
	if data.FileName == "" {
		return "", fmt.Errorf("filename is required")
	}
	name := filepath.Base(data.FileName)
	if name == "" || name == "." || name == ".." {
		return "", fmt.Errorf("invalid filename")
	}
	tempDir, err := os.MkdirTemp("", "waveterm-")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(data.Data64)
	if err != nil {
		return "", fmt.Errorf("error decoding base64 data: %w", err)
	}
	tempPath := filepath.Join(tempDir, name)
	err = os.WriteFile(tempPath, decoded, 0600)
	if err != nil {
		return "", fmt.Errorf("error writing temp file: %w", err)
	}
	return tempPath, nil
}

func (ws *WshServer) DeleteSubBlockCommand(ctx context.Context, data dshrpc.CommandDeleteBlockData) error {
	if data.BlockId == "" {
		return fmt.Errorf("blockid is required")
	}
	err := dcore.DeleteBlock(ctx, data.BlockId, false)
	if err != nil {
		return fmt.Errorf("error deleting block: %w", err)
	}
	return nil
}

func (ws *WshServer) DeleteBlockCommand(ctx context.Context, data dshrpc.CommandDeleteBlockData) error {
	if data.BlockId == "" {
		return fmt.Errorf("blockid is required")
	}
	ctx = doraobj.ContextWithUpdates(ctx)
	tabId, err := dstore.DBFindTabForBlockId(ctx, data.BlockId)
	if err != nil {
		return fmt.Errorf("error finding tab for block: %w", err)
	}
	if tabId == "" {
		return fmt.Errorf("no tab found for block")
	}
	err = dcore.DeleteBlock(ctx, data.BlockId, true)
	if err != nil {
		return fmt.Errorf("error deleting block: %w", err)
	}
	dcore.QueueLayoutActionForTab(ctx, tabId, doraobj.LayoutActionData{
		ActionType: dcore.LayoutActionDataType_Remove,
		BlockId:    data.BlockId,
	})
	updates := doraobj.ContextGetUpdatesRtn(ctx)
	dps.Broker.SendUpdateEvents(updates)
	return nil
}

func (ws *WshServer) WaitForRouteCommand(ctx context.Context, data dshrpc.CommandWaitForRouteData) (bool, error) {
	waitCtx, cancelFn := context.WithTimeout(ctx, time.Duration(data.WaitMs)*time.Millisecond)
	defer cancelFn()
	err := dshutil.DefaultRouter.WaitForRegister(waitCtx, data.RouteId)
	return err == nil, nil
}

func (ws *WshServer) EventRecvCommand(ctx context.Context, data dps.WaveEvent) error {
	return nil
}

func (ws *WshServer) EventPublishCommand(ctx context.Context, data dps.WaveEvent) error {
	rpcSource := dshutil.GetRpcSourceFromContext(ctx)
	if rpcSource == "" {
		return fmt.Errorf("no rpc source set")
	}
	if data.Sender == "" {
		data.Sender = rpcSource
	}
	dps.Broker.Publish(data)
	return nil
}

func (ws *WshServer) EventSubCommand(ctx context.Context, data dps.SubscriptionRequest) error {
	rpcSource := dshutil.GetRpcSourceFromContext(ctx)
	if rpcSource == "" {
		return fmt.Errorf("no rpc source set")
	}
	dps.Broker.Subscribe(rpcSource, data)
	return nil
}

func (ws *WshServer) EventUnsubCommand(ctx context.Context, data string) error {
	rpcSource := dshutil.GetRpcSourceFromContext(ctx)
	if rpcSource == "" {
		return fmt.Errorf("no rpc source set")
	}
	dps.Broker.Unsubscribe(rpcSource, data)
	return nil
}

func (ws *WshServer) EventUnsubAllCommand(ctx context.Context) error {
	rpcSource := dshutil.GetRpcSourceFromContext(ctx)
	if rpcSource == "" {
		return fmt.Errorf("no rpc source set")
	}
	dps.Broker.UnsubscribeAll(rpcSource)
	return nil
}

func (ws *WshServer) EventReadHistoryCommand(ctx context.Context, data dshrpc.CommandEventReadHistoryData) ([]*dps.WaveEvent, error) {
	events := dps.Broker.ReadEventHistory(data.Event, data.Scope, data.MaxItems)
	return events, nil
}

func (ws *WshServer) SetConfigCommand(ctx context.Context, data dshrpc.MetaSettingsType) error {
	return dconfig.SetBaseConfigValue(data.MetaMapType)
}

func (ws *WshServer) func (ws *WshServer) GetFullConfigCommand(ctx context.Context) (dconfig.FullConfigType, error) {
	watcher := dconfig.GetWatcher()
	return watcher.GetFullConfig(), nil
}

func (ws *WshServer) func (ws *WshServer) func (ws *WshServer) func termCtxWithLogBlockId(ctx context.Context, logBlockId string) context.Context {
	if logBlockId == "" {
		return ctx
	}
	block, err := dstore.DBMustGet[*doraobj.Block](ctx, logBlockId)
	if err != nil {
		return ctx
	}
	connDebug := block.Meta.GetString(doraobj.MetaKey_TermConnDebug, "")
	if connDebug == "" {
		return ctx
	}
	return blocklogger.ContextWithLogBlockId(ctx, logBlockId, connDebug == "debug")
}


func waveFileToWaveFileInfo(wf *filestore.WaveFile) *dshrpc.WaveFileInfo {
	return &dshrpc.WaveFileInfo{
		ZoneId:    wf.ZoneId,
		Name:      wf.Name,
		Opts:      wf.Opts,
		CreatedTs: wf.CreatedTs,
		Size:      wf.Size,
		ModTs:     wf.ModTs,
		Meta:      wf.Meta,
	}
}

func (ws *WshServer) BlockInfoCommand(ctx context.Context, blockId string) (*dshrpc.BlockInfoData, error) {
	blockData, err := dstore.DBMustGet[*doraobj.Block](ctx, blockId)
	if err != nil {
		return nil, fmt.Errorf("error getting block: %w", err)
	}
	tabId, err := dstore.DBFindTabForBlockId(ctx, blockId)
	if err != nil {
		return nil, fmt.Errorf("error finding tab for block: %w", err)
	}
	workspaceId, err := dstore.DBFindWorkspaceForTabId(ctx, tabId)
	if err != nil {
		return nil, fmt.Errorf("error finding window for tab: %w", err)
	}
	fileList, err := filestore.WFS.ListFiles(ctx, blockId)
	if err != nil {
		return nil, fmt.Errorf("error listing blockfiles: %w", err)
	}
	var fileInfoList []*dshrpc.WaveFileInfo
	for _, wf := range fileList {
		fileInfoList = append(fileInfoList, waveFileToWaveFileInfo(wf))
	}
	return &dshrpc.BlockInfoData{
		BlockId:     blockId,
		TabId:       tabId,
		WorkspaceId: workspaceId,
		Block:       blockData,
		Files:       fileInfoList,
	}, nil
}

func (ws *WshServer) DebugTermCommand(ctx context.Context, data dshrpc.CommandDebugTermData) (*dshrpc.CommandDebugTermRtnData, error) {
	if data.BlockId == "" {
		return nil, fmt.Errorf("blockid is required")
	}
	if data.Size <= 0 {
		return nil, fmt.Errorf("size must be greater than 0")
	}
	waveFile, err := filestore.WFS.Stat(ctx, data.BlockId, dorabase.BlockFile_Term)
	if err == fs.ErrNotExist {
		return &dshrpc.CommandDebugTermRtnData{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error statting term file: %w", err)
	}
	readSize := data.Size
	dataLength := waveFile.DataLength()
	if readSize > dataLength {
		readSize = dataLength
	}
	readOffset := waveFile.Size - readSize
	readOffset, readData, err := filestore.WFS.ReadAt(ctx, data.BlockId, dorabase.BlockFile_Term, readOffset, readSize)
	if err != nil {
		return nil, fmt.Errorf("error reading term file: %w", err)
	}
	return &dshrpc.CommandDebugTermRtnData{
		Offset: readOffset,
		Data64: base64.StdEncoding.EncodeToString(readData),
	}, nil
}

func (ws *WshServer) WaveInfoCommand(ctx context.Context) (*dshrpc.WaveInfoData, error) {
	return &dshrpc.WaveInfoData{
		Version:   dorabase.WaveVersion,
		ClientId:  dstore.GetClientId(),
		BuildTime: dorabase.BuildTime,
		ConfigDir: dorabase.GetWaveConfigDir(),
		DataDir:   dorabase.GetWaveDataDir(),
	}, nil
}

func (ws *WshServer) MacOSVersionCommand(ctx context.Context) (string, error) {
	return dorabase.ClientMacOSVersion(), nil
}

// BlocksListCommand returns every block visible in the requested
// scope (current workspace by default).
func (ws *WshServer) BlocksListCommand(
	ctx context.Context,
	req dshrpc.BlocksListRequest) ([]dshrpc.BlocksListEntry, error) {
	var results []dshrpc.BlocksListEntry

	// Resolve the set of workspaces to inspect
	var workspaceIDs []string
	if req.WorkspaceId != "" {
		workspaceIDs = []string{req.WorkspaceId}
	} else if req.WindowId != "" {
		win, err := dcore.GetWindow(ctx, req.WindowId)
		if err != nil {
			return nil, err
		}
		workspaceIDs = []string{win.WorkspaceId}
	} else {
		// "current" == first workspace in client focus list
		client, err := dstore.DBGetSingleton[*doraobj.Client](ctx)
		if err != nil {
			return nil, err
		}
		if len(client.WindowIds) == 0 {
			return nil, fmt.Errorf("no active window")
		}
		win, err := dcore.GetWindow(ctx, client.WindowIds[0])
		if err != nil {
			return nil, err
		}
		workspaceIDs = []string{win.WorkspaceId}
	}

	for _, wsID := range workspaceIDs {
		wsData, err := dcore.GetWorkspace(ctx, wsID)
		if err != nil {
			return nil, err
		}

		windowId, err := dstore.DBFindWindowForWorkspaceId(ctx, wsID)
		if err != nil {
			log.Printf("error finding window for workspace %s: %v", wsID, err)
		}

		for _, tabID := range wsData.TabIds {
			tab, err := dstore.DBMustGet[*doraobj.Tab](ctx, tabID)
			if err != nil {
				return nil, err
			}
			for _, blkID := range tab.BlockIds {
				blk, err := dstore.DBMustGet[*doraobj.Block](ctx, blkID)
				if err != nil {
					return nil, err
				}
				results = append(results, dshrpc.BlocksListEntry{
					WindowId:    windowId,
					WorkspaceId: wsID,
					TabId:       tabID,
					BlockId:     blkID,
					Meta:        blk.Meta,
				})
			}
		}
	}
	return results, nil
}

func (ws *WshServer) WorkspaceListCommand(ctx context.Context) ([]dshrpc.WorkspaceInfoData, error) {
	workspaceList, err := dcore.ListWorkspaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing workspaces: %w", err)
	}
	var rtn []dshrpc.WorkspaceInfoData
	for _, workspaceEntry := range workspaceList {
		workspaceData, err := dcore.GetWorkspace(ctx, workspaceEntry.WorkspaceId)
		if err != nil {
			return nil, fmt.Errorf("error getting workspace: %w", err)
		}
		rtn = append(rtn, dshrpc.WorkspaceInfoData{
			WindowId:      workspaceEntry.WindowId,
			WorkspaceData: workspaceData,
		})
	}
	return rtn, nil
}


func (ws WshServer) SendTelemetryCommand(ctx context.Context) error {
	return wcloud.SendAllTelemetry(dstore.GetClientId())
}

var wshActivityRe = regexp.MustCompile(`^[a-z:#]+$`)

func (ws *WshServer) WshActivityCommand(ctx context.Context, data map[string]int) error {
	if len(data) == 0 {
		return nil
	}
	props := telemetrydata.TEventProps{}
	for key, value := range data {
		if len(key) > 20 {
			delete(data, key)
		}
		if !wshActivityRe.MatchString(key) {
			delete(data, key)
		}
		if value != 1 {
			delete(data, key)
		}
		if strings.HasSuffix(key, "#error") {
			props.WshCmd = strings.TrimSuffix(key, "#error")
			props.WshErrorCount = 1
		} else {
			props.WshCmd = key
		}
	}
	activityUpdate := dshrpc.ActivityUpdate{
		WshCmds: data,
	}
	telemetry.GoUpdateActivityWrap(activityUpdate, "wsh-activity")
	telemetry.GoRecordTEventWrap(&telemetrydata.TEvent{
		Event: telemetry.WshRunEventName,
		Props: props,
	})
	return nil
}

func (ws *WshServer) ActivityCommand(ctx context.Context, activity dshrpc.ActivityUpdate) error {
	telemetry.GoUpdateActivityWrap(activity, "wshrpc-activity")
	return nil
}

func (ws *WshServer) GetVarCommand(ctx context.Context, data dshrpc.CommandVarData) (*dshrpc.CommandVarResponseData, error) {
	_, fileData, err := filestore.WFS.ReadFile(ctx, data.ZoneId, data.FileName)
	if err == fs.ErrNotExist {
		return &dshrpc.CommandVarResponseData{Key: data.Key, Exists: false}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading blockfile: %w", err)
	}
	envMap := envutil.EnvToMap(string(fileData))
	value, ok := envMap[data.Key]
	return &dshrpc.CommandVarResponseData{Key: data.Key, Exists: ok, Val: value}, nil
}

func (ws *WshServer) GetAllVarsCommand(ctx context.Context, data dshrpc.CommandVarData) ([]dshrpc.CommandVarResponseData, error) {
	_, fileData, err := filestore.WFS.ReadFile(ctx, data.ZoneId, data.FileName)
	if err == fs.ErrNotExist {
		return []dshrpc.CommandVarResponseData{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading blockfile: %w", err)
	}
	envMap := envutil.EnvToMap(string(fileData))
	keys := make([]string, 0, len(envMap))
	for k := range envMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]dshrpc.CommandVarResponseData, 0, len(keys))
	for _, k := range keys {
		result = append(result, dshrpc.CommandVarResponseData{
			Key:    k,
			Val:    envMap[k],
			Exists: true,
		})
	}
	return result, nil
}

func (ws *WshServer) SetVarCommand(ctx context.Context, data dshrpc.CommandVarData) error {
	_, fileData, err := filestore.WFS.ReadFile(ctx, data.ZoneId, data.FileName)
	if err == fs.ErrNotExist {
		fileData = []byte{}
		err = filestore.WFS.MakeFile(ctx, data.ZoneId, data.FileName, nil, dshrpc.FileOpts{})
		if err != nil {
			return fmt.Errorf("error creating blockfile: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error reading blockfile: %w", err)
	}
	envMap := envutil.EnvToMap(string(fileData))
	if data.Remove {
		delete(envMap, data.Key)
	} else {
		envMap[data.Key] = data.Val
	}
	envStr := envutil.MapToEnv(envMap)
	return filestore.WFS.WriteFile(ctx, data.ZoneId, data.FileName, []byte(envStr))
}

func (ws *WshServer) PathCommand(ctx context.Context, data dshrpc.PathCommandData) (string, error) {
	pathType := data.PathType
	openInternal := data.Open
	openExternal := data.OpenExternal
	var path string
	switch pathType {
	case "config":
		path = dorabase.GetWaveConfigDir()
	case "data":
		path = dorabase.GetWaveDataDir()
	case "log":
		path = filepath.Join(dorabase.GetWaveDataDir(), "waveapp.log")
	}

	if openInternal && openExternal {
		return "", fmt.Errorf("open and openExternal cannot both be true")
	}

	if openInternal {
		_, err := ws.CreateBlockCommand(ctx, dshrpc.CommandCreateBlockData{
			TabId: data.TabId,
			BlockDef: &doraobj.BlockDef{Meta: map[string]any{
				doraobj.MetaKey_View: "preview",
				doraobj.MetaKey_File: path,
			}},
			Ephemeral: true,
			Focused:   true,
		})

		if err != nil {
			return path, fmt.Errorf("error opening path: %w", err)
		}
	} else if openExternal {
		err := open.Run(path)
		if err != nil {
			return path, fmt.Errorf("error opening path: %w", err)
		}
	}
	return path, nil
}

func (ws *WshServer) GetTabCommand(ctx context.Context, tabId string) (*doraobj.Tab, error) {
	tab, err := dstore.DBGet[*doraobj.Tab](ctx, tabId)
	if err != nil {
		return nil, fmt.Errorf("error getting tab: %w", err)
	}
	return tab, nil
}

func (ws *WshServer) GetAllBadgesCommand(ctx context.Context) ([]baseds.BadgeEvent, error) {
	return dcore.GetAllBadges(), nil
}

func (ws *WshServer) GetSecretsCommand(ctx context.Context, names []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, name := range names {
		value, exists, err := secretstore.GetSecret(name)
		if err != nil {
			return nil, fmt.Errorf("error getting secret %q: %w", name, err)
		}
		if exists {
			result[name] = value
		}
	}
	return result, nil
}

func (ws *WshServer) GetSecretsNamesCommand(ctx context.Context) ([]string, error) {
	names, err := secretstore.GetSecretNames()
	if err != nil {
		return nil, fmt.Errorf("error getting secret names: %w", err)
	}
	return names, nil
}

func (ws *WshServer) SetSecretsCommand(ctx context.Context, secrets map[string]*string) error {
	for name, value := range secrets {
		if value == nil {
			err := secretstore.DeleteSecret(name)
			if err != nil {
				return fmt.Errorf("error deleting secret %q: %w", name, err)
			}
		} else {
			err := secretstore.SetSecret(name, *value)
			if err != nil {
				return fmt.Errorf("error setting secret %q: %w", name, err)
			}
		}
	}
	return nil
}

func (ws *WshServer) GetSecretsLinuxStorageBackendCommand(ctx context.Context) (string, error) {
	backend, err := secretstore.GetLinuxStorageBackend()
	if err != nil {
		return "", fmt.Errorf("error getting linux storage backend: %w", err)
	}
	return backend, nil
}

func (ws *WshServer) JobCmdExitedCommand(ctx context.Context, data dshrpc.CommandJobCmdExitedData) error {
	return jobcontroller.HandleCmdJobExited(ctx, data.JobId, data)
}

func (ws *WshServer) JobControllerListCommand(ctx context.Context) ([]*doraobj.Job, error) {
	return dstore.DBGetAllObjsByType[*doraobj.Job](ctx, doraobj.OType_Job)
}

func (ws *WshServer) JobControllerDeleteJobCommand(ctx context.Context, jobId string) error {
	return jobcontroller.DeleteJob(ctx, jobId)
}

func (ws *WshServer) JobControllerStartJobCommand(ctx context.Context, data dshrpc.CommandJobControllerStartJobData) (string, error) {
	params := jobcontroller.StartJobParams{
		ConnName: data.ConnName,
		JobKind:  data.JobKind,
		Cmd:      data.Cmd,
		Args:     data.Args,
		Env:      data.Env,
		TermSize: data.TermSize,
	}
	return jobcontroller.StartJob(ctx, params)
}

func (ws *WshServer) JobControllerExitJobCommand(ctx context.Context, jobId string) error {
	return jobcontroller.TerminateJobManager(ctx, jobId)
}

func (ws *WshServer) JobControllerDisconnectJobCommand(ctx context.Context, jobId string) error {
	return jobcontroller.DisconnectJob(ctx, jobId)
}

func (ws *WshServer) JobControllerReconnectJobCommand(ctx context.Context, jobId string) error {
	return jobcontroller.ReconnectJob(ctx, jobId, nil)
}

func (ws *WshServer) JobControllerReconnectJobsForConnCommand(ctx context.Context, connName string) error {
	return jobcontroller.ReconnectJobsForConn(ctx, connName)
}

func (ws *WshServer) JobControllerConnectedJobsCommand(ctx context.Context) ([]string, error) {
	return jobcontroller.GetConnectedJobIds(), nil
}

func (ws *WshServer) JobControllerAttachJobCommand(ctx context.Context, data dshrpc.CommandJobControllerAttachJobData) error {
	return jobcontroller.AttachJobToBlock(ctx, data.JobId, data.BlockId)
}

func (ws *WshServer) JobControllerDetachJobCommand(ctx context.Context, jobId string) error {
	return jobcontroller.DetachJobFromBlock(ctx, jobId, true)
}

func (ws *WshServer) BlockJobStatusCommand(ctx context.Context, blockId string) (*dshrpc.BlockJobStatusData, error) {
	return jobcontroller.GetBlockJobStatus(ctx, blockId)
}
