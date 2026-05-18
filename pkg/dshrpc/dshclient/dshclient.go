// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

// Generated Code. DO NOT EDIT.

package dshclient

import (
	"github.com/dfbb/doraterm/pkg/aiusechat/uctypes"
	"github.com/dfbb/doraterm/pkg/baseds"
	"github.com/dfbb/doraterm/pkg/telemetry/telemetrydata"
	"github.com/dfbb/doraterm/pkg/vdom"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dconfig"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

// command "activity", dshserver.ActivityCommand
func ActivityCommand(w *dshutil.DshRpc, data dshrpc.ActivityUpdate, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "activity", data, opts)
	return err
}

// command "aisendmessage", dshserver.AiSendMessageCommand
func AiSendMessageCommand(w *dshutil.DshRpc, data dshrpc.AiMessageData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "aisendmessage", data, opts)
	return err
}

// command "authenticate", dshserver.AuthenticateCommand
func AuthenticateCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) (dshrpc.CommandAuthenticateRtnData, error) {
	resp, err := sendRpcRequestCallHelper[dshrpc.CommandAuthenticateRtnData](w, "authenticate", data, opts)
	return resp, err
}

// command "authenticatejobmanager", dshserver.AuthenticateJobManagerCommand
func AuthenticateJobManagerCommand(w *dshutil.DshRpc, data dshrpc.CommandAuthenticateJobManagerData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "authenticatejobmanager", data, opts)
	return err
}

// command "authenticatejobmanagerverify", dshserver.AuthenticateJobManagerVerifyCommand
func AuthenticateJobManagerVerifyCommand(w *dshutil.DshRpc, data dshrpc.CommandAuthenticateJobManagerData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "authenticatejobmanagerverify", data, opts)
	return err
}

// command "authenticatetojobmanager", dshserver.AuthenticateToJobManagerCommand
func AuthenticateToJobManagerCommand(w *dshutil.DshRpc, data dshrpc.CommandAuthenticateToJobData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "authenticatetojobmanager", data, opts)
	return err
}

// command "authenticatetoken", dshserver.AuthenticateTokenCommand
func AuthenticateTokenCommand(w *dshutil.DshRpc, data dshrpc.CommandAuthenticateTokenData, opts *dshrpc.RpcOpts) (dshrpc.CommandAuthenticateRtnData, error) {
	resp, err := sendRpcRequestCallHelper[dshrpc.CommandAuthenticateRtnData](w, "authenticatetoken", data, opts)
	return resp, err
}

// command "authenticatetokenverify", dshserver.AuthenticateTokenVerifyCommand
func AuthenticateTokenVerifyCommand(w *dshutil.DshRpc, data dshrpc.CommandAuthenticateTokenData, opts *dshrpc.RpcOpts) (dshrpc.CommandAuthenticateRtnData, error) {
	resp, err := sendRpcRequestCallHelper[dshrpc.CommandAuthenticateRtnData](w, "authenticatetokenverify", data, opts)
	return resp, err
}

// command "badgewatchpid", dshserver.BadgeWatchPidCommand
func BadgeWatchPidCommand(w *dshutil.DshRpc, data dshrpc.CommandBadgeWatchPidData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "badgewatchpid", data, opts)
	return err
}

// command "blockinfo", dshserver.BlockInfoCommand
func BlockInfoCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) (*dshrpc.BlockInfoData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.BlockInfoData](w, "blockinfo", data, opts)
	return resp, err
}

// command "blockjobstatus", dshserver.BlockJobStatusCommand
func BlockJobStatusCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) (*dshrpc.BlockJobStatusData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.BlockJobStatusData](w, "blockjobstatus", data, opts)
	return resp, err
}

// command "blockslist", dshserver.BlocksListCommand
func BlocksListCommand(w *dshutil.DshRpc, data dshrpc.BlocksListRequest, opts *dshrpc.RpcOpts) ([]dshrpc.BlocksListEntry, error) {
	resp, err := sendRpcRequestCallHelper[[]dshrpc.BlocksListEntry](w, "blockslist", data, opts)
	return resp, err
}

// command "captureblockscreenshot", dshserver.CaptureBlockScreenshotCommand
func CaptureBlockScreenshotCommand(w *dshutil.DshRpc, data dshrpc.CommandCaptureBlockScreenshotData, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "captureblockscreenshot", data, opts)
	return resp, err
}

// command "checkgoversion", dshserver.CheckGoVersionCommand
func CheckGoVersionCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (*dshrpc.CommandCheckGoVersionRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandCheckGoVersionRtnData](w, "checkgoversion", nil, opts)
	return resp, err
}

// command "connconnect", dshserver.ConnConnectCommand
func ConnConnectCommand(w *dshutil.DshRpc, data dshrpc.ConnRequest, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "connconnect", data, opts)
	return err
}

// command "conndisconnect", dshserver.ConnDisconnectCommand
func ConnDisconnectCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "conndisconnect", data, opts)
	return err
}

// command "connensure", dshserver.ConnEnsureCommand
func ConnEnsureCommand(w *dshutil.DshRpc, data dshrpc.ConnExtData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "connensure", data, opts)
	return err
}

// command "connlist", dshserver.ConnListCommand
func ConnListCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "connlist", nil, opts)
	return resp, err
}

// command "connreinstallwsh", dshserver.ConnReinstallWshCommand
func ConnReinstallWshCommand(w *dshutil.DshRpc, data dshrpc.ConnExtData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "connreinstallwsh", data, opts)
	return err
}

// command "connserverinit", dshserver.ConnServerInitCommand
func ConnServerInitCommand(w *dshutil.DshRpc, data dshrpc.CommandConnServerInitData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "connserverinit", data, opts)
	return err
}

// command "connstatus", dshserver.ConnStatusCommand
func ConnStatusCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]dshrpc.ConnStatus, error) {
	resp, err := sendRpcRequestCallHelper[[]dshrpc.ConnStatus](w, "connstatus", nil, opts)
	return resp, err
}

// command "connupdatewsh", dshserver.ConnUpdateWshCommand
func ConnUpdateWshCommand(w *dshutil.DshRpc, data dshrpc.RemoteInfo, opts *dshrpc.RpcOpts) (bool, error) {
	resp, err := sendRpcRequestCallHelper[bool](w, "connupdatewsh", data, opts)
	return resp, err
}

// command "controlgetrouteid", dshserver.ControlGetRouteIdCommand
func ControlGetRouteIdCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "controlgetrouteid", nil, opts)
	return resp, err
}

// command "controllerappendoutput", dshserver.ControllerAppendOutputCommand
func ControllerAppendOutputCommand(w *dshutil.DshRpc, data dshrpc.CommandControllerAppendOutputData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "controllerappendoutput", data, opts)
	return err
}

// command "controllerdestroy", dshserver.ControllerDestroyCommand
func ControllerDestroyCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "controllerdestroy", data, opts)
	return err
}

// command "controllerinput", dshserver.ControllerInputCommand
func ControllerInputCommand(w *dshutil.DshRpc, data dshrpc.CommandBlockInputData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "controllerinput", data, opts)
	return err
}

// command "controllerresync", dshserver.ControllerResyncCommand
func ControllerResyncCommand(w *dshutil.DshRpc, data dshrpc.CommandControllerResyncData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "controllerresync", data, opts)
	return err
}

// command "createblock", dshserver.CreateBlockCommand
func CreateBlockCommand(w *dshutil.DshRpc, data dshrpc.CommandCreateBlockData, opts *dshrpc.RpcOpts) (doraobj.ORef, error) {
	resp, err := sendRpcRequestCallHelper[doraobj.ORef](w, "createblock", data, opts)
	return resp, err
}

// command "createsubblock", dshserver.CreateSubBlockCommand
func CreateSubBlockCommand(w *dshutil.DshRpc, data dshrpc.CommandCreateSubBlockData, opts *dshrpc.RpcOpts) (doraobj.ORef, error) {
	resp, err := sendRpcRequestCallHelper[doraobj.ORef](w, "createsubblock", data, opts)
	return resp, err
}

// command "debugterm", dshserver.DebugTermCommand
func DebugTermCommand(w *dshutil.DshRpc, data dshrpc.CommandDebugTermData, opts *dshrpc.RpcOpts) (*dshrpc.CommandDebugTermRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandDebugTermRtnData](w, "debugterm", data, opts)
	return resp, err
}

// command "deleteappfile", dshserver.DeleteAppFileCommand
func DeleteAppFileCommand(w *dshutil.DshRpc, data dshrpc.CommandDeleteAppFileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "deleteappfile", data, opts)
	return err
}

// command "deleteblock", dshserver.DeleteBlockCommand
func DeleteBlockCommand(w *dshutil.DshRpc, data dshrpc.CommandDeleteBlockData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "deleteblock", data, opts)
	return err
}

// command "deletebuilder", dshserver.DeleteBuilderCommand
func DeleteBuilderCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "deletebuilder", data, opts)
	return err
}

// command "deletesubblock", dshserver.DeleteSubBlockCommand
func DeleteSubBlockCommand(w *dshutil.DshRpc, data dshrpc.CommandDeleteBlockData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "deletesubblock", data, opts)
	return err
}

// command "dismisswshfail", dshserver.DismissWshFailCommand
func DismissWshFailCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "dismisswshfail", data, opts)
	return err
}

// command "dispose", dshserver.DisposeCommand
func DisposeCommand(w *dshutil.DshRpc, data dshrpc.CommandDisposeData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "dispose", data, opts)
	return err
}

// command "disposesuggestions", dshserver.DisposeSuggestionsCommand
func DisposeSuggestionsCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "disposesuggestions", data, opts)
	return err
}

// command "electrondecrypt", dshserver.ElectronDecryptCommand
func ElectronDecryptCommand(w *dshutil.DshRpc, data dshrpc.CommandElectronDecryptData, opts *dshrpc.RpcOpts) (*dshrpc.CommandElectronDecryptRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandElectronDecryptRtnData](w, "electrondecrypt", data, opts)
	return resp, err
}

// command "electronencrypt", dshserver.ElectronEncryptCommand
func ElectronEncryptCommand(w *dshutil.DshRpc, data dshrpc.CommandElectronEncryptData, opts *dshrpc.RpcOpts) (*dshrpc.CommandElectronEncryptRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandElectronEncryptRtnData](w, "electronencrypt", data, opts)
	return resp, err
}

// command "electronsystembell", dshserver.ElectronSystemBellCommand
func ElectronSystemBellCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "electronsystembell", nil, opts)
	return err
}

// command "eventpublish", dshserver.EventPublishCommand
func EventPublishCommand(w *dshutil.DshRpc, data dps.DoraEvent, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventpublish", data, opts)
	return err
}

// command "eventreadhistory", dshserver.EventReadHistoryCommand
func EventReadHistoryCommand(w *dshutil.DshRpc, data dshrpc.CommandEventReadHistoryData, opts *dshrpc.RpcOpts) ([]*dps.DoraEvent, error) {
	resp, err := sendRpcRequestCallHelper[[]*dps.DoraEvent](w, "eventreadhistory", data, opts)
	return resp, err
}

// command "eventrecv", dshserver.EventRecvCommand
func EventRecvCommand(w *dshutil.DshRpc, data dps.DoraEvent, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventrecv", data, opts)
	return err
}

// command "eventsub", dshserver.EventSubCommand
func EventSubCommand(w *dshutil.DshRpc, data dps.SubscriptionRequest, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventsub", data, opts)
	return err
}

// command "eventunsub", dshserver.EventUnsubCommand
func EventUnsubCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventunsub", data, opts)
	return err
}

// command "eventunsuball", dshserver.EventUnsubAllCommand
func EventUnsubAllCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "eventunsuball", nil, opts)
	return err
}

// command "fetchsuggestions", dshserver.FetchSuggestionsCommand
func FetchSuggestionsCommand(w *dshutil.DshRpc, data dshrpc.FetchSuggestionsData, opts *dshrpc.RpcOpts) (*dshrpc.FetchSuggestionsResponse, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.FetchSuggestionsResponse](w, "fetchsuggestions", data, opts)
	return resp, err
}

// command "fileappend", dshserver.FileAppendCommand
func FileAppendCommand(w *dshutil.DshRpc, data dshrpc.FileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "fileappend", data, opts)
	return err
}

// command "filecopy", dshserver.FileCopyCommand
func FileCopyCommand(w *dshutil.DshRpc, data dshrpc.CommandFileCopyData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filecopy", data, opts)
	return err
}

// command "filecreate", dshserver.FileCreateCommand
func FileCreateCommand(w *dshutil.DshRpc, data dshrpc.FileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filecreate", data, opts)
	return err
}

// command "filedelete", dshserver.FileDeleteCommand
func FileDeleteCommand(w *dshutil.DshRpc, data dshrpc.CommandDeleteFileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filedelete", data, opts)
	return err
}

// command "fileinfo", dshserver.FileInfoCommand
func FileInfoCommand(w *dshutil.DshRpc, data dshrpc.FileData, opts *dshrpc.RpcOpts) (*dshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.FileInfo](w, "fileinfo", data, opts)
	return resp, err
}

// command "filejoin", dshserver.FileJoinCommand
func FileJoinCommand(w *dshutil.DshRpc, data []string, opts *dshrpc.RpcOpts) (*dshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.FileInfo](w, "filejoin", data, opts)
	return resp, err
}

// command "filelist", dshserver.FileListCommand
func FileListCommand(w *dshutil.DshRpc, data dshrpc.FileListData, opts *dshrpc.RpcOpts) ([]*dshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[[]*dshrpc.FileInfo](w, "filelist", data, opts)
	return resp, err
}

// command "fileliststream", dshserver.FileListStreamCommand
func FileListStreamCommand(w *dshutil.DshRpc, data dshrpc.FileListData, opts *dshrpc.RpcOpts) chan dshrpc.RespOrErrorUnion[dshrpc.CommandRemoteListEntriesRtnData] {
	return sendRpcRequestResponseStreamHelper[dshrpc.CommandRemoteListEntriesRtnData](w, "fileliststream", data, opts)
}

// command "filemkdir", dshserver.FileMkdirCommand
func FileMkdirCommand(w *dshutil.DshRpc, data dshrpc.FileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filemkdir", data, opts)
	return err
}

// command "filemove", dshserver.FileMoveCommand
func FileMoveCommand(w *dshutil.DshRpc, data dshrpc.CommandFileCopyData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filemove", data, opts)
	return err
}

// command "fileread", dshserver.FileReadCommand
func FileReadCommand(w *dshutil.DshRpc, data dshrpc.FileData, opts *dshrpc.RpcOpts) (*dshrpc.FileData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.FileData](w, "fileread", data, opts)
	return resp, err
}

// command "filerestorebackup", dshserver.FileRestoreBackupCommand
func FileRestoreBackupCommand(w *dshutil.DshRpc, data dshrpc.CommandFileRestoreBackupData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filerestorebackup", data, opts)
	return err
}

// command "filestream", dshserver.FileStreamCommand
func FileStreamCommand(w *dshutil.DshRpc, data dshrpc.CommandFileStreamData, opts *dshrpc.RpcOpts) (*dshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.FileInfo](w, "filestream", data, opts)
	return resp, err
}

// command "filewrite", dshserver.FileWriteCommand
func FileWriteCommand(w *dshutil.DshRpc, data dshrpc.FileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "filewrite", data, opts)
	return err
}

// command "findgitbash", dshserver.FindGitBashCommand
func FindGitBashCommand(w *dshutil.DshRpc, data bool, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "findgitbash", data, opts)
	return resp, err
}

// command "focuswindow", dshserver.FocusWindowCommand
func FocusWindowCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "focuswindow", data, opts)
	return err
}

// command "getallbadges", dshserver.GetAllBadgesCommand
func GetAllBadgesCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]baseds.BadgeEvent, error) {
	resp, err := sendRpcRequestCallHelper[[]baseds.BadgeEvent](w, "getallbadges", nil, opts)
	return resp, err
}

// command "getallvars", dshserver.GetAllVarsCommand
func GetAllVarsCommand(w *dshutil.DshRpc, data dshrpc.CommandVarData, opts *dshrpc.RpcOpts) ([]dshrpc.CommandVarResponseData, error) {
	resp, err := sendRpcRequestCallHelper[[]dshrpc.CommandVarResponseData](w, "getallvars", data, opts)
	return resp, err
}

// command "getbuilderoutput", dshserver.GetBuilderOutputCommand
func GetBuilderOutputCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "getbuilderoutput", data, opts)
	return resp, err
}

// command "getbuilderstatus", dshserver.GetBuilderStatusCommand
func GetBuilderStatusCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) (*dshrpc.BuilderStatusData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.BuilderStatusData](w, "getbuilderstatus", data, opts)
	return resp, err
}

// command "getfocusedblockdata", dshserver.GetFocusedBlockDataCommand
func GetFocusedBlockDataCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (*dshrpc.FocusedBlockData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.FocusedBlockData](w, "getfocusedblockdata", nil, opts)
	return resp, err
}

// command "getfullconfig", dshserver.GetFullConfigCommand
func GetFullConfigCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (dconfig.FullConfigType, error) {
	resp, err := sendRpcRequestCallHelper[dconfig.FullConfigType](w, "getfullconfig", nil, opts)
	return resp, err
}

// command "getjwtpublickey", dshserver.GetJwtPublicKeyCommand
func GetJwtPublicKeyCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "getjwtpublickey", nil, opts)
	return resp, err
}

// command "getmeta", dshserver.GetMetaCommand
func GetMetaCommand(w *dshutil.DshRpc, data dshrpc.CommandGetMetaData, opts *dshrpc.RpcOpts) (doraobj.MetaMapType, error) {
	resp, err := sendRpcRequestCallHelper[doraobj.MetaMapType](w, "getmeta", data, opts)
	return resp, err
}

// command "getrtinfo", dshserver.GetRTInfoCommand
func GetRTInfoCommand(w *dshutil.DshRpc, data dshrpc.CommandGetRTInfoData, opts *dshrpc.RpcOpts) (*doraobj.ObjRTInfo, error) {
	resp, err := sendRpcRequestCallHelper[*doraobj.ObjRTInfo](w, "getrtinfo", data, opts)
	return resp, err
}

// command "getsecrets", dshserver.GetSecretsCommand
func GetSecretsCommand(w *dshutil.DshRpc, data []string, opts *dshrpc.RpcOpts) (map[string]string, error) {
	resp, err := sendRpcRequestCallHelper[map[string]string](w, "getsecrets", data, opts)
	return resp, err
}

// command "getsecretslinuxstoragebackend", dshserver.GetSecretsLinuxStorageBackendCommand
func GetSecretsLinuxStorageBackendCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "getsecretslinuxstoragebackend", nil, opts)
	return resp, err
}

// command "getsecretsnames", dshserver.GetSecretsNamesCommand
func GetSecretsNamesCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "getsecretsnames", nil, opts)
	return resp, err
}

// command "gettab", dshserver.GetTabCommand
func GetTabCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) (*doraobj.Tab, error) {
	resp, err := sendRpcRequestCallHelper[*doraobj.Tab](w, "gettab", data, opts)
	return resp, err
}

// command "gettempdir", dshserver.GetTempDirCommand
func GetTempDirCommand(w *dshutil.DshRpc, data dshrpc.CommandGetTempDirData, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "gettempdir", data, opts)
	return resp, err
}

// command "getupdatechannel", dshserver.GetUpdateChannelCommand
func GetUpdateChannelCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "getupdatechannel", nil, opts)
	return resp, err
}

// command "getvar", dshserver.GetVarCommand
func GetVarCommand(w *dshutil.DshRpc, data dshrpc.CommandVarData, opts *dshrpc.RpcOpts) (*dshrpc.CommandVarResponseData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandVarResponseData](w, "getvar", data, opts)
	return resp, err
}

// command "getwaveaichat", dshserver.GetDoraAIChatCommand
func GetDoraAIChatCommand(w *dshutil.DshRpc, data dshrpc.CommandGetDoraAIChatData, opts *dshrpc.RpcOpts) (*uctypes.UIChat, error) {
	resp, err := sendRpcRequestCallHelper[*uctypes.UIChat](w, "getwaveaichat", data, opts)
	return resp, err
}

// command "getwaveaimodeconfig", dshserver.GetDoraAIModeConfigCommand
func GetDoraAIModeConfigCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (dconfig.AIModeConfigUpdate, error) {
	resp, err := sendRpcRequestCallHelper[dconfig.AIModeConfigUpdate](w, "getwaveaimodeconfig", nil, opts)
	return resp, err
}

// command "getwaveairatelimit", dshserver.GetDoraAIRateLimitCommand
func GetDoraAIRateLimitCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (*uctypes.RateLimitInfo, error) {
	resp, err := sendRpcRequestCallHelper[*uctypes.RateLimitInfo](w, "getwaveairatelimit", nil, opts)
	return resp, err
}

// command "jobcmdexited", dshserver.JobCmdExitedCommand
func JobCmdExitedCommand(w *dshutil.DshRpc, data dshrpc.CommandJobCmdExitedData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobcmdexited", data, opts)
	return err
}

// command "jobcontrollerattachjob", dshserver.JobControllerAttachJobCommand
func JobControllerAttachJobCommand(w *dshutil.DshRpc, data dshrpc.CommandJobControllerAttachJobData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobcontrollerattachjob", data, opts)
	return err
}

// command "jobcontrollerconnectedjobs", dshserver.JobControllerConnectedJobsCommand
func JobControllerConnectedJobsCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "jobcontrollerconnectedjobs", nil, opts)
	return resp, err
}

// command "jobcontrollerdeletejob", dshserver.JobControllerDeleteJobCommand
func JobControllerDeleteJobCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobcontrollerdeletejob", data, opts)
	return err
}

// command "jobcontrollerdetachjob", dshserver.JobControllerDetachJobCommand
func JobControllerDetachJobCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobcontrollerdetachjob", data, opts)
	return err
}

// command "jobcontrollerdisconnectjob", dshserver.JobControllerDisconnectJobCommand
func JobControllerDisconnectJobCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobcontrollerdisconnectjob", data, opts)
	return err
}

// command "jobcontrollerexitjob", dshserver.JobControllerExitJobCommand
func JobControllerExitJobCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobcontrollerexitjob", data, opts)
	return err
}

// command "jobcontrollergetalljobmanagerstatus", dshserver.JobControllerGetAllJobManagerStatusCommand
func JobControllerGetAllJobManagerStatusCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]*dshrpc.JobManagerStatusUpdate, error) {
	resp, err := sendRpcRequestCallHelper[[]*dshrpc.JobManagerStatusUpdate](w, "jobcontrollergetalljobmanagerstatus", nil, opts)
	return resp, err
}

// command "jobcontrollerlist", dshserver.JobControllerListCommand
func JobControllerListCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]*doraobj.Job, error) {
	resp, err := sendRpcRequestCallHelper[[]*doraobj.Job](w, "jobcontrollerlist", nil, opts)
	return resp, err
}

// command "jobcontrollerreconnectjob", dshserver.JobControllerReconnectJobCommand
func JobControllerReconnectJobCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobcontrollerreconnectjob", data, opts)
	return err
}

// command "jobcontrollerreconnectjobsforconn", dshserver.JobControllerReconnectJobsForConnCommand
func JobControllerReconnectJobsForConnCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobcontrollerreconnectjobsforconn", data, opts)
	return err
}

// command "jobcontrollerstartjob", dshserver.JobControllerStartJobCommand
func JobControllerStartJobCommand(w *dshutil.DshRpc, data dshrpc.CommandJobControllerStartJobData, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "jobcontrollerstartjob", data, opts)
	return resp, err
}

// command "jobinput", dshserver.JobInputCommand
func JobInputCommand(w *dshutil.DshRpc, data dshrpc.CommandJobInputData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobinput", data, opts)
	return err
}

// command "jobprepareconnect", dshserver.JobPrepareConnectCommand
func JobPrepareConnectCommand(w *dshutil.DshRpc, data dshrpc.CommandJobPrepareConnectData, opts *dshrpc.RpcOpts) (*dshrpc.CommandJobConnectRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandJobConnectRtnData](w, "jobprepareconnect", data, opts)
	return resp, err
}

// command "jobstartstream", dshserver.JobStartStreamCommand
func JobStartStreamCommand(w *dshutil.DshRpc, data dshrpc.CommandJobStartStreamData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "jobstartstream", data, opts)
	return err
}

// command "listallappfiles", dshserver.ListAllAppFilesCommand
func ListAllAppFilesCommand(w *dshutil.DshRpc, data dshrpc.CommandListAllAppFilesData, opts *dshrpc.RpcOpts) (*dshrpc.CommandListAllAppFilesRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandListAllAppFilesRtnData](w, "listallappfiles", data, opts)
	return resp, err
}

// command "listallapps", dshserver.ListAllAppsCommand
func ListAllAppsCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]dshrpc.AppInfo, error) {
	resp, err := sendRpcRequestCallHelper[[]dshrpc.AppInfo](w, "listallapps", nil, opts)
	return resp, err
}

// command "listalleditableapps", dshserver.ListAllEditableAppsCommand
func ListAllEditableAppsCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]dshrpc.AppInfo, error) {
	resp, err := sendRpcRequestCallHelper[[]dshrpc.AppInfo](w, "listalleditableapps", nil, opts)
	return resp, err
}

// command "macosversion", dshserver.MacOSVersionCommand
func MacOSVersionCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "macosversion", nil, opts)
	return resp, err
}

// command "makedraftfromlocal", dshserver.MakeDraftFromLocalCommand
func MakeDraftFromLocalCommand(w *dshutil.DshRpc, data dshrpc.CommandMakeDraftFromLocalData, opts *dshrpc.RpcOpts) (*dshrpc.CommandMakeDraftFromLocalRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandMakeDraftFromLocalRtnData](w, "makedraftfromlocal", data, opts)
	return resp, err
}

// command "message", dshserver.MessageCommand
func MessageCommand(w *dshutil.DshRpc, data dshrpc.CommandMessageData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "message", data, opts)
	return err
}

// command "networkonline", dshserver.NetworkOnlineCommand
func NetworkOnlineCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (bool, error) {
	resp, err := sendRpcRequestCallHelper[bool](w, "networkonline", nil, opts)
	return resp, err
}

// command "notify", dshserver.NotifyCommand
func NotifyCommand(w *dshutil.DshRpc, data dshrpc.DoraNotificationOptions, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "notify", data, opts)
	return err
}

// command "notifysystemresume", dshserver.NotifySystemResumeCommand
func NotifySystemResumeCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "notifysystemresume", nil, opts)
	return err
}

// command "path", dshserver.PathCommand
func PathCommand(w *dshutil.DshRpc, data dshrpc.PathCommandData, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "path", data, opts)
	return resp, err
}

// command "publishapp", dshserver.PublishAppCommand
func PublishAppCommand(w *dshutil.DshRpc, data dshrpc.CommandPublishAppData, opts *dshrpc.RpcOpts) (*dshrpc.CommandPublishAppRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandPublishAppRtnData](w, "publishapp", data, opts)
	return resp, err
}

// command "readappfile", dshserver.ReadAppFileCommand
func ReadAppFileCommand(w *dshutil.DshRpc, data dshrpc.CommandReadAppFileData, opts *dshrpc.RpcOpts) (*dshrpc.CommandReadAppFileRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandReadAppFileRtnData](w, "readappfile", data, opts)
	return resp, err
}

// command "recordtevent", dshserver.RecordTEventCommand
func RecordTEventCommand(w *dshutil.DshRpc, data telemetrydata.TEvent, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "recordtevent", data, opts)
	return err
}

// command "remotedisconnectfromjobmanager", dshserver.RemoteDisconnectFromJobManagerCommand
func RemoteDisconnectFromJobManagerCommand(w *dshutil.DshRpc, data dshrpc.CommandRemoteDisconnectFromJobManagerData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotedisconnectfromjobmanager", data, opts)
	return err
}

// command "remotefilecopy", dshserver.RemoteFileCopyCommand
func RemoteFileCopyCommand(w *dshutil.DshRpc, data dshrpc.CommandFileCopyData, opts *dshrpc.RpcOpts) (bool, error) {
	resp, err := sendRpcRequestCallHelper[bool](w, "remotefilecopy", data, opts)
	return resp, err
}

// command "remotefiledelete", dshserver.RemoteFileDeleteCommand
func RemoteFileDeleteCommand(w *dshutil.DshRpc, data dshrpc.CommandDeleteFileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotefiledelete", data, opts)
	return err
}

// command "remotefileinfo", dshserver.RemoteFileInfoCommand
func RemoteFileInfoCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) (*dshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.FileInfo](w, "remotefileinfo", data, opts)
	return resp, err
}

// command "remotefilejoin", dshserver.RemoteFileJoinCommand
func RemoteFileJoinCommand(w *dshutil.DshRpc, data []string, opts *dshrpc.RpcOpts) (*dshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.FileInfo](w, "remotefilejoin", data, opts)
	return resp, err
}

// command "remotefilemove", dshserver.RemoteFileMoveCommand
func RemoteFileMoveCommand(w *dshutil.DshRpc, data dshrpc.CommandFileCopyData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotefilemove", data, opts)
	return err
}

// command "remotefilemultiinfo", dshserver.RemoteFileMultiInfoCommand
func RemoteFileMultiInfoCommand(w *dshutil.DshRpc, data dshrpc.CommandRemoteFileMultiInfoData, opts *dshrpc.RpcOpts) (map[string]dshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[map[string]dshrpc.FileInfo](w, "remotefilemultiinfo", data, opts)
	return resp, err
}

// command "remotefilestream", dshserver.RemoteFileStreamCommand
func RemoteFileStreamCommand(w *dshutil.DshRpc, data dshrpc.CommandRemoteFileStreamData, opts *dshrpc.RpcOpts) (*dshrpc.FileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.FileInfo](w, "remotefilestream", data, opts)
	return resp, err
}

// command "remotefiletouch", dshserver.RemoteFileTouchCommand
func RemoteFileTouchCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotefiletouch", data, opts)
	return err
}

// command "remotegetinfo", dshserver.RemoteGetInfoCommand
func RemoteGetInfoCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (dshrpc.RemoteInfo, error) {
	resp, err := sendRpcRequestCallHelper[dshrpc.RemoteInfo](w, "remotegetinfo", nil, opts)
	return resp, err
}

// command "remoteinstallrcfiles", dshserver.RemoteInstallRcFilesCommand
func RemoteInstallRcFilesCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remoteinstallrcfiles", nil, opts)
	return err
}

// command "remotelistentries", dshserver.RemoteListEntriesCommand
func RemoteListEntriesCommand(w *dshutil.DshRpc, data dshrpc.CommandRemoteListEntriesData, opts *dshrpc.RpcOpts) chan dshrpc.RespOrErrorUnion[dshrpc.CommandRemoteListEntriesRtnData] {
	return sendRpcRequestResponseStreamHelper[dshrpc.CommandRemoteListEntriesRtnData](w, "remotelistentries", data, opts)
}

// command "remotemkdir", dshserver.RemoteMkdirCommand
func RemoteMkdirCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotemkdir", data, opts)
	return err
}

// command "remoteprocesslist", dshserver.RemoteProcessListCommand
func RemoteProcessListCommand(w *dshutil.DshRpc, data dshrpc.CommandRemoteProcessListData, opts *dshrpc.RpcOpts) (*dshrpc.ProcessListResponse, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.ProcessListResponse](w, "remoteprocesslist", data, opts)
	return resp, err
}

// command "remoteprocesssignal", dshserver.RemoteProcessSignalCommand
func RemoteProcessSignalCommand(w *dshutil.DshRpc, data dshrpc.CommandRemoteProcessSignalData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remoteprocesssignal", data, opts)
	return err
}

// command "remotereconnecttojobmanager", dshserver.RemoteReconnectToJobManagerCommand
func RemoteReconnectToJobManagerCommand(w *dshutil.DshRpc, data dshrpc.CommandRemoteReconnectToJobManagerData, opts *dshrpc.RpcOpts) (*dshrpc.CommandRemoteReconnectToJobManagerRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandRemoteReconnectToJobManagerRtnData](w, "remotereconnecttojobmanager", data, opts)
	return resp, err
}

// command "remotestartjob", dshserver.RemoteStartJobCommand
func RemoteStartJobCommand(w *dshutil.DshRpc, data dshrpc.CommandRemoteStartJobData, opts *dshrpc.RpcOpts) (*dshrpc.CommandStartJobRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandStartJobRtnData](w, "remotestartjob", data, opts)
	return resp, err
}

// command "remotestreamcpudata", dshserver.RemoteStreamCpuDataCommand
func RemoteStreamCpuDataCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) chan dshrpc.RespOrErrorUnion[dshrpc.TimeSeriesData] {
	return sendRpcRequestResponseStreamHelper[dshrpc.TimeSeriesData](w, "remotestreamcpudata", nil, opts)
}

// command "remoteterminatejobmanager", dshserver.RemoteTerminateJobManagerCommand
func RemoteTerminateJobManagerCommand(w *dshutil.DshRpc, data dshrpc.CommandRemoteTerminateJobManagerData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remoteterminatejobmanager", data, opts)
	return err
}

// command "remotewritefile", dshserver.RemoteWriteFileCommand
func RemoteWriteFileCommand(w *dshutil.DshRpc, data dshrpc.FileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "remotewritefile", data, opts)
	return err
}

// command "renameappfile", dshserver.RenameAppFileCommand
func RenameAppFileCommand(w *dshutil.DshRpc, data dshrpc.CommandRenameAppFileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "renameappfile", data, opts)
	return err
}

// command "resolveids", dshserver.ResolveIdsCommand
func ResolveIdsCommand(w *dshutil.DshRpc, data dshrpc.CommandResolveIdsData, opts *dshrpc.RpcOpts) (dshrpc.CommandResolveIdsRtnData, error) {
	resp, err := sendRpcRequestCallHelper[dshrpc.CommandResolveIdsRtnData](w, "resolveids", data, opts)
	return resp, err
}

// command "restartbuilderandwait", dshserver.RestartBuilderAndWaitCommand
func RestartBuilderAndWaitCommand(w *dshutil.DshRpc, data dshrpc.CommandRestartBuilderAndWaitData, opts *dshrpc.RpcOpts) (*dshrpc.RestartBuilderAndWaitResult, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.RestartBuilderAndWaitResult](w, "restartbuilderandwait", data, opts)
	return resp, err
}

// command "routeannounce", dshserver.RouteAnnounceCommand
func RouteAnnounceCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "routeannounce", nil, opts)
	return err
}

// command "routeunannounce", dshserver.RouteUnannounceCommand
func RouteUnannounceCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "routeunannounce", nil, opts)
	return err
}

// command "sendtelemetry", dshserver.SendTelemetryCommand
func SendTelemetryCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "sendtelemetry", nil, opts)
	return err
}

// command "setblockfocus", dshserver.SetBlockFocusCommand
func SetBlockFocusCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setblockfocus", data, opts)
	return err
}

// command "setconfig", dshserver.SetConfigCommand
func SetConfigCommand(w *dshutil.DshRpc, data dshrpc.MetaSettingsType, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setconfig", data, opts)
	return err
}

// command "setconnectionsconfig", dshserver.SetConnectionsConfigCommand
func SetConnectionsConfigCommand(w *dshutil.DshRpc, data dshrpc.ConnConfigRequest, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setconnectionsconfig", data, opts)
	return err
}

// command "setmeta", dshserver.SetMetaCommand
func SetMetaCommand(w *dshutil.DshRpc, data dshrpc.CommandSetMetaData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setmeta", data, opts)
	return err
}

// command "setpeerinfo", dshserver.SetPeerInfoCommand
func SetPeerInfoCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setpeerinfo", data, opts)
	return err
}

// command "setrtinfo", dshserver.SetRTInfoCommand
func SetRTInfoCommand(w *dshutil.DshRpc, data dshrpc.CommandSetRTInfoData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setrtinfo", data, opts)
	return err
}

// command "setsecrets", dshserver.SetSecretsCommand
func SetSecretsCommand(w *dshutil.DshRpc, data map[string]*string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setsecrets", data, opts)
	return err
}

// command "setvar", dshserver.SetVarCommand
func SetVarCommand(w *dshutil.DshRpc, data dshrpc.CommandVarData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "setvar", data, opts)
	return err
}

// command "startbuilder", dshserver.StartBuilderCommand
func StartBuilderCommand(w *dshutil.DshRpc, data dshrpc.CommandStartBuilderData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "startbuilder", data, opts)
	return err
}

// command "startjob", dshserver.StartJobCommand
func StartJobCommand(w *dshutil.DshRpc, data dshrpc.CommandStartJobData, opts *dshrpc.RpcOpts) (*dshrpc.CommandStartJobRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandStartJobRtnData](w, "startjob", data, opts)
	return resp, err
}

// command "stopbuilder", dshserver.StopBuilderCommand
func StopBuilderCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "stopbuilder", data, opts)
	return err
}

// command "streamcpudata", dshserver.StreamCpuDataCommand
func StreamCpuDataCommand(w *dshutil.DshRpc, data dshrpc.CpuDataRequest, opts *dshrpc.RpcOpts) chan dshrpc.RespOrErrorUnion[dshrpc.TimeSeriesData] {
	return sendRpcRequestResponseStreamHelper[dshrpc.TimeSeriesData](w, "streamcpudata", data, opts)
}

// command "streamdata", dshserver.StreamDataCommand
func StreamDataCommand(w *dshutil.DshRpc, data dshrpc.CommandStreamData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "streamdata", data, opts)
	return err
}

// command "streamdataack", dshserver.StreamDataAckCommand
func StreamDataAckCommand(w *dshutil.DshRpc, data dshrpc.CommandStreamAckData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "streamdataack", data, opts)
	return err
}

// command "streamtest", dshserver.StreamTestCommand
func StreamTestCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) chan dshrpc.RespOrErrorUnion[int] {
	return sendRpcRequestResponseStreamHelper[int](w, "streamtest", nil, opts)
}

// command "termgetscrollbacklines", dshserver.TermGetScrollbackLinesCommand
func TermGetScrollbackLinesCommand(w *dshutil.DshRpc, data dshrpc.CommandTermGetScrollbackLinesData, opts *dshrpc.RpcOpts) (*dshrpc.CommandTermGetScrollbackLinesRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandTermGetScrollbackLinesRtnData](w, "termgetscrollbacklines", data, opts)
	return resp, err
}

// command "test", dshserver.TestCommand
func TestCommand(w *dshutil.DshRpc, data string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "test", data, opts)
	return err
}

// command "testmultiarg", dshserver.TestMultiArgCommand
func TestMultiArgCommand(w *dshutil.DshRpc, arg1 string, arg2 int, arg3 bool, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "testmultiarg", dshrpc.MultiArg{Args: []any{arg1, arg2, arg3}}, opts)
	return resp, err
}

// command "updatetabname", dshserver.UpdateTabNameCommand
func UpdateTabNameCommand(w *dshutil.DshRpc, arg1 string, arg2 string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "updatetabname", dshrpc.MultiArg{Args: []any{arg1, arg2}}, opts)
	return err
}

// command "updateworkspacetabids", dshserver.UpdateWorkspaceTabIdsCommand
func UpdateWorkspaceTabIdsCommand(w *dshutil.DshRpc, arg1 string, arg2 []string, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "updateworkspacetabids", dshrpc.MultiArg{Args: []any{arg1, arg2}}, opts)
	return err
}

// command "vdomasyncinitiation", dshserver.VDomAsyncInitiationCommand
func VDomAsyncInitiationCommand(w *dshutil.DshRpc, data vdom.VDomAsyncInitiationRequest, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "vdomasyncinitiation", data, opts)
	return err
}

// command "vdomcreatecontext", dshserver.VDomCreateContextCommand
func VDomCreateContextCommand(w *dshutil.DshRpc, data vdom.VDomCreateContext, opts *dshrpc.RpcOpts) (*doraobj.ORef, error) {
	resp, err := sendRpcRequestCallHelper[*doraobj.ORef](w, "vdomcreatecontext", data, opts)
	return resp, err
}

// command "vdomrender", dshserver.VDomRenderCommand
func VDomRenderCommand(w *dshutil.DshRpc, data vdom.VDomFrontendUpdate, opts *dshrpc.RpcOpts) chan dshrpc.RespOrErrorUnion[*vdom.VDomBackendUpdate] {
	return sendRpcRequestResponseStreamHelper[*vdom.VDomBackendUpdate](w, "vdomrender", data, opts)
}

// command "vdomurlrequest", dshserver.VDomUrlRequestCommand
func VDomUrlRequestCommand(w *dshutil.DshRpc, data dshrpc.VDomUrlRequestData, opts *dshrpc.RpcOpts) chan dshrpc.RespOrErrorUnion[dshrpc.VDomUrlRequestResponse] {
	return sendRpcRequestResponseStreamHelper[dshrpc.VDomUrlRequestResponse](w, "vdomurlrequest", data, opts)
}

// command "waitforroute", dshserver.WaitForRouteCommand
func WaitForRouteCommand(w *dshutil.DshRpc, data dshrpc.CommandWaitForRouteData, opts *dshrpc.RpcOpts) (bool, error) {
	resp, err := sendRpcRequestCallHelper[bool](w, "waitforroute", data, opts)
	return resp, err
}

// command "waveaiaddcontext", dshserver.DoraAIAddContextCommand
func DoraAIAddContextCommand(w *dshutil.DshRpc, data dshrpc.CommandDoraAIAddContextData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "waveaiaddcontext", data, opts)
	return err
}

// command "waveaienabletelemetry", dshserver.DoraAIEnableTelemetryCommand
func DoraAIEnableTelemetryCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "waveaienabletelemetry", nil, opts)
	return err
}

// command "waveaigettooldiff", dshserver.DoraAIGetToolDiffCommand
func DoraAIGetToolDiffCommand(w *dshutil.DshRpc, data dshrpc.CommandDoraAIGetToolDiffData, opts *dshrpc.RpcOpts) (*dshrpc.CommandDoraAIGetToolDiffRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandDoraAIGetToolDiffRtnData](w, "waveaigettooldiff", data, opts)
	return resp, err
}

// command "waveaitoolapprove", dshserver.DoraAIToolApproveCommand
func DoraAIToolApproveCommand(w *dshutil.DshRpc, data dshrpc.CommandDoraAIToolApproveData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "waveaitoolapprove", data, opts)
	return err
}

// command "wavefilereadstream", dshserver.DoraFileReadStreamCommand
func DoraFileReadStreamCommand(w *dshutil.DshRpc, data dshrpc.CommandDoraFileReadStreamData, opts *dshrpc.RpcOpts) (*dshrpc.DoraFileInfo, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.DoraFileInfo](w, "wavefilereadstream", data, opts)
	return resp, err
}

// command "waveinfo", dshserver.DoraInfoCommand
func DoraInfoCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (*dshrpc.DoraInfoData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.DoraInfoData](w, "waveinfo", nil, opts)
	return resp, err
}

// command "webselector", dshserver.WebSelectorCommand
func WebSelectorCommand(w *dshutil.DshRpc, data dshrpc.CommandWebSelectorData, opts *dshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "webselector", data, opts)
	return resp, err
}

// command "workspacelist", dshserver.WorkspaceListCommand
func WorkspaceListCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]dshrpc.WorkspaceInfoData, error) {
	resp, err := sendRpcRequestCallHelper[[]dshrpc.WorkspaceInfoData](w, "workspacelist", nil, opts)
	return resp, err
}

// command "writeappfile", dshserver.WriteAppFileCommand
func WriteAppFileCommand(w *dshutil.DshRpc, data dshrpc.CommandWriteAppFileData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "writeappfile", data, opts)
	return err
}

// command "writeappgofile", dshserver.WriteAppGoFileCommand
func WriteAppGoFileCommand(w *dshutil.DshRpc, data dshrpc.CommandWriteAppGoFileData, opts *dshrpc.RpcOpts) (*dshrpc.CommandWriteAppGoFileRtnData, error) {
	resp, err := sendRpcRequestCallHelper[*dshrpc.CommandWriteAppGoFileRtnData](w, "writeappgofile", data, opts)
	return resp, err
}

// command "writeappsecretbindings", dshserver.WriteAppSecretBindingsCommand
func WriteAppSecretBindingsCommand(w *dshutil.DshRpc, data dshrpc.CommandWriteAppSecretBindingsData, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "writeappsecretbindings", data, opts)
	return err
}

// command "writetempfile", dshserver.WriteTempFileCommand
func WriteTempFileCommand(w *dshutil.DshRpc, data dshrpc.CommandWriteTempFileData, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "writetempfile", data, opts)
	return resp, err
}

// command "wshactivity", dshserver.DshActivityCommand
func DshActivityCommand(w *dshutil.DshRpc, data map[string]int, opts *dshrpc.RpcOpts) error {
	_, err := sendRpcRequestCallHelper[any](w, "wshactivity", data, opts)
	return err
}

// command "wsldefaultdistro", dshserver.WslDefaultDistroCommand
func WslDefaultDistroCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) (string, error) {
	resp, err := sendRpcRequestCallHelper[string](w, "wsldefaultdistro", nil, opts)
	return resp, err
}

// command "wsllist", dshserver.WslListCommand
func WslListCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]string, error) {
	resp, err := sendRpcRequestCallHelper[[]string](w, "wsllist", nil, opts)
	return resp, err
}

// command "wslstatus", dshserver.WslStatusCommand
func WslStatusCommand(w *dshutil.DshRpc, opts *dshrpc.RpcOpts) ([]dshrpc.ConnStatus, error) {
	resp, err := sendRpcRequestCallHelper[[]dshrpc.ConnStatus](w, "wslstatus", nil, opts)
	return resp, err
}


