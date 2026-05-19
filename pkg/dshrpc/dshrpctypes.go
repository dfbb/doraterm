// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

// types and methods for dsh rpc calls
package dshrpc

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/dfbb/doraterm/pkg/baseds"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dconfig"
	"github.com/dfbb/doraterm/pkg/dps"
)

type RespOrErrorUnion[T any] struct {
	Response T
	Error    error
}

type MultiArg struct {
	Args []any `json:"args"`
}

// Instructions for adding a new RPC call
// * methods must end with Command
// * methods must take context as their first parameter
// * methods may take additional typed parameters, and may return either just an error, or one return value plus an error
// * after modifying DshRpcInterface, run `task generate` to regnerate bindings

type DshRpcInterface interface {
	AuthenticateCommand(ctx context.Context, data string) (CommandAuthenticateRtnData, error)
	AuthenticateTokenCommand(ctx context.Context, data CommandAuthenticateTokenData) (CommandAuthenticateRtnData, error)
	AuthenticateTokenVerifyCommand(ctx context.Context, data CommandAuthenticateTokenData) (CommandAuthenticateRtnData, error) // (special) validates token without binding, root router only
	AuthenticateJobManagerCommand(ctx context.Context, data CommandAuthenticateJobManagerData) error
	AuthenticateJobManagerVerifyCommand(ctx context.Context, data CommandAuthenticateJobManagerData) error // (special) validates job auth token without binding, root router only
	DisposeCommand(ctx context.Context, data CommandDisposeData) error
	RouteAnnounceCommand(ctx context.Context) error               // (special) announces a new route to the main router
	RouteUnannounceCommand(ctx context.Context) error             // (special) unannounces a route to the main router
	ControlGetRouteIdCommand(ctx context.Context) (string, error) // (special) gets the route for the link that we're on
	SetPeerInfoCommand(ctx context.Context, peerInfo string) error
	GetJwtPublicKeyCommand(ctx context.Context) (string, error) // (special) gets the public JWT signing key

	MessageCommand(ctx context.Context, data CommandMessageData) error
	GetMetaCommand(ctx context.Context, data CommandGetMetaData) (doraobj.MetaMapType, error)
	SetMetaCommand(ctx context.Context, data CommandSetMetaData) error
	ControllerInputCommand(ctx context.Context, data CommandBlockInputData) error
	ControllerDestroyCommand(ctx context.Context, blockId string) error
	ControllerResyncCommand(ctx context.Context, data CommandControllerResyncData) error
	ControllerAppendOutputCommand(ctx context.Context, data CommandControllerAppendOutputData) error
	ResolveIdsCommand(ctx context.Context, data CommandResolveIdsData) (CommandResolveIdsRtnData, error)
	CreateBlockCommand(ctx context.Context, data CommandCreateBlockData) (doraobj.ORef, error)
	CreateSubBlockCommand(ctx context.Context, data CommandCreateSubBlockData) (doraobj.ORef, error)
	DeleteBlockCommand(ctx context.Context, data CommandDeleteBlockData) error
	DeleteSubBlockCommand(ctx context.Context, data CommandDeleteBlockData) error
	WaitForRouteCommand(ctx context.Context, data CommandWaitForRouteData) (bool, error)

	EventPublishCommand(ctx context.Context, data dps.DoraEvent) error
	EventSubCommand(ctx context.Context, data dps.SubscriptionRequest) error
	EventUnsubCommand(ctx context.Context, data string) error
	EventUnsubAllCommand(ctx context.Context) error
	EventReadHistoryCommand(ctx context.Context, data CommandEventReadHistoryData) ([]*dps.DoraEvent, error)

	GetTempDirCommand(ctx context.Context, data CommandGetTempDirData) (string, error)
	WriteTempFileCommand(ctx context.Context, data CommandWriteTempFileData) (string, error)
	StreamTestCommand(ctx context.Context) chan RespOrErrorUnion[int]
	StreamCpuDataCommand(ctx context.Context, request CpuDataRequest) chan RespOrErrorUnion[TimeSeriesData]
	TestCommand(ctx context.Context, data string) error
	TestMultiArgCommand(ctx context.Context, arg1 string, arg2 int, arg3 bool) (string, error)
	SetConfigCommand(ctx context.Context, data MetaSettingsType) error
	GetFullConfigCommand(ctx context.Context) (dconfig.FullConfigType, error)
	BlockInfoCommand(ctx context.Context, blockId string) (*BlockInfoData, error)
	DebugTermCommand(ctx context.Context, data CommandDebugTermData) (*CommandDebugTermRtnData, error)
	BlocksListCommand(ctx context.Context, data BlocksListRequest) ([]BlocksListEntry, error)
	DoraInfoCommand(ctx context.Context) (*DoraInfoData, error)
	MacOSVersionCommand(ctx context.Context) (string, error)
	GetVarCommand(ctx context.Context, data CommandVarData) (*CommandVarResponseData, error)
	GetAllVarsCommand(ctx context.Context, data CommandVarData) ([]CommandVarResponseData, error)
	SetVarCommand(ctx context.Context, data CommandVarData) error
	PathCommand(ctx context.Context, data PathCommandData) (string, error)
	GetTabCommand(ctx context.Context, tabId string) (*doraobj.Tab, error)
	UpdateTabNameCommand(ctx context.Context, tabId string, newName string) error
	UpdateWorkspaceTabIdsCommand(ctx context.Context, workspaceId string, tabIds []string) error
	GetAllBadgesCommand(ctx context.Context) ([]baseds.BadgeEvent, error)

	// eventrecv is special, it's handled internally by DshRpc with EventListener
	EventRecvCommand(ctx context.Context, data dps.DoraEvent) error

	// remotes
	DshRpcRemoteFileInterface
	RemoteStreamCpuDataCommand(ctx context.Context) chan RespOrErrorUnion[TimeSeriesData]
	RemoteGetInfoCommand(ctx context.Context) (RemoteInfo, error)
	RemoteInstallRcFilesCommand(ctx context.Context) error
	RemoteStartJobCommand(ctx context.Context, data CommandRemoteStartJobData) (*CommandStartJobRtnData, error)
	RemoteReconnectToJobManagerCommand(ctx context.Context, data CommandRemoteReconnectToJobManagerData) (*CommandRemoteReconnectToJobManagerRtnData, error)
	RemoteDisconnectFromJobManagerCommand(ctx context.Context, data CommandRemoteDisconnectFromJobManagerData) error
	RemoteTerminateJobManagerCommand(ctx context.Context, data CommandRemoteTerminateJobManagerData) error
	BadgeWatchPidCommand(ctx context.Context, data CommandBadgeWatchPidData) error
	RemoteProcessListCommand(ctx context.Context, data CommandRemoteProcessListData) (*ProcessListResponse, error)
	RemoteProcessSignalCommand(ctx context.Context, data CommandRemoteProcessSignalData) error

	// emain
	WebSelectorCommand(ctx context.Context, data CommandWebSelectorData) ([]string, error)
	NotifyCommand(ctx context.Context, notificationOptions DoraNotificationOptions) error
	FocusWindowCommand(ctx context.Context, windowId string) error
	ElectronEncryptCommand(ctx context.Context, data CommandElectronEncryptData) (*CommandElectronEncryptRtnData, error)
	ElectronDecryptCommand(ctx context.Context, data CommandElectronDecryptData) (*CommandElectronDecryptRtnData, error)
	NetworkOnlineCommand(ctx context.Context) (bool, error)
	ElectronSystemBellCommand(ctx context.Context) error

	WorkspaceListCommand(ctx context.Context) ([]WorkspaceInfoData, error)
	GetUpdateChannelCommand(ctx context.Context) (string, error)

	// screenshot
	CaptureBlockScreenshotCommand(ctx context.Context, data CommandCaptureBlockScreenshotData) (string, error)

	// block focus
	SetBlockFocusCommand(ctx context.Context, blockId string) error
	GetFocusedBlockDataCommand(ctx context.Context) (*FocusedBlockData, error)

	// rtinfo
	GetRTInfoCommand(ctx context.Context, data CommandGetRTInfoData) (*doraobj.ObjRTInfo, error)
	SetRTInfoCommand(ctx context.Context, data CommandSetRTInfoData) error

	// terminal
	TermGetScrollbackLinesCommand(ctx context.Context, data CommandTermGetScrollbackLinesData) (*CommandTermGetScrollbackLinesRtnData, error)

	// file
	DshRpcFileInterface
	DoraFileReadStreamCommand(ctx context.Context, data CommandDoraFileReadStreamData) (*DoraFileInfo, error)

	// streams
	StreamDataCommand(ctx context.Context, data CommandStreamData) error
	StreamDataAckCommand(ctx context.Context, data CommandStreamAckData) error

	// jobs
	AuthenticateToJobManagerCommand(ctx context.Context, data CommandAuthenticateToJobData) error
	StartJobCommand(ctx context.Context, data CommandStartJobData) (*CommandStartJobRtnData, error)
	JobPrepareConnectCommand(ctx context.Context, data CommandJobPrepareConnectData) (*CommandJobConnectRtnData, error)
	JobStartStreamCommand(ctx context.Context, data CommandJobStartStreamData) error
	JobInputCommand(ctx context.Context, data CommandJobInputData) error
	JobCmdExitedCommand(ctx context.Context, data CommandJobCmdExitedData) error // this is sent FROM the job manager => main server

	// job controller
	JobControllerDeleteJobCommand(ctx context.Context, jobId string) error
	JobControllerListCommand(ctx context.Context) ([]*doraobj.Job, error)
	JobControllerStartJobCommand(ctx context.Context, data CommandJobControllerStartJobData) (string, error)
	JobControllerExitJobCommand(ctx context.Context, jobId string) error
	JobControllerDisconnectJobCommand(ctx context.Context, jobId string) error
	JobControllerReconnectJobCommand(ctx context.Context, jobId string) error
	JobControllerReconnectJobsForConnCommand(ctx context.Context, connName string) error
	JobControllerConnectedJobsCommand(ctx context.Context) ([]string, error)
	JobControllerAttachJobCommand(ctx context.Context, data CommandJobControllerAttachJobData) error
	JobControllerDetachJobCommand(ctx context.Context, jobId string) error
	JobControllerGetAllJobManagerStatusCommand(ctx context.Context) ([]*JobManagerStatusUpdate, error)
	BlockJobStatusCommand(ctx context.Context, blockId string) (*BlockJobStatusData, error)
}

// for frontend
type DshServerCommandMeta struct {
	CommandType string `json:"commandtype"`
}

type RpcOpts struct {
	Timeout    int64  `json:"timeout,omitempty"`
	NoResponse bool   `json:"noresponse,omitempty"`
	Route      string `json:"route,omitempty"`

	StreamCancelFn func(context.Context) error `json:"-"` // this is an *output* parameter, set by the handler
}

type RpcContext struct {
	SockName  string `json:"sockname,omitempty"`  // the domain socket name
	RouteId   string `json:"routeid"`             // the routeid from the jwt
	ProcRoute bool   `json:"procroute,omitempty"` // use a random procid for route
	BlockId   string `json:"blockid,omitempty"`   // blockid for this rpc
	Conn      string `json:"conn,omitempty"`      // the conn name
	IsRouter  bool   `json:"isrouter,omitempty"`  // if this is for a sub-router
}

func (rc RpcContext) GenerateRouteId() string {
	if rc.RouteId != "" {
		return rc.RouteId
	}
	return "proc:" + uuid.New().String()
}

type CommandAuthenticateRtnData struct {
	RouteId string `json:"routeid"`

	// these fields are only set when doing a token swap
	Env            map[string]string `json:"env,omitempty"`
	InitScriptText string            `json:"initscripttext,omitempty"`
	RpcContext     *RpcContext       `json:"rpccontext,omitempty"`
}

type CommandAuthenticateTokenData struct {
	Token string `json:"token"`
}

type CommandDisposeData struct {
	RouteId string `json:"routeid"`
	// auth token travels in the packet directly
}

type CommandMessageData struct {
	Message string `json:"message"`
}

type CommandGetMetaData struct {
	ORef doraobj.ORef `json:"oref"`
}

type CommandSetMetaData struct {
	ORef doraobj.ORef        `json:"oref"`
	Meta doraobj.MetaMapType `json:"meta"`
}

type CommandResolveIdsData struct {
	BlockId string   `json:"blockid"`
	Ids     []string `json:"ids"`
}

type CommandResolveIdsRtnData struct {
	ResolvedIds map[string]doraobj.ORef `json:"resolvedids"`
}

type CommandCreateBlockData struct {
	TabId         string               `json:"tabid"`
	BlockDef      *doraobj.BlockDef    `json:"blockdef"`
	RtOpts        *doraobj.RuntimeOpts `json:"rtopts,omitempty"`
	Magnified     bool                 `json:"magnified,omitempty"`
	Ephemeral     bool                 `json:"ephemeral,omitempty"`
	Focused       bool                 `json:"focused,omitempty"`
	TargetBlockId string               `json:"targetblockid,omitempty"`
	TargetAction  string               `json:"targetaction,omitempty"` // "replace", "splitright", "splitdown", "splitleft", "splitup"
}

type CommandCreateSubBlockData struct {
	ParentBlockId string            `json:"parentblockid"`
	BlockDef      *doraobj.BlockDef `json:"blockdef"`
}

type CommandControllerResyncData struct {
	ForceRestart bool                 `json:"forcerestart,omitempty"`
	TabId        string               `json:"tabid"`
	BlockId      string               `json:"blockid"`
	RtOpts       *doraobj.RuntimeOpts `json:"rtopts,omitempty"`
}

type CommandControllerAppendOutputData struct {
	BlockId string `json:"blockid"`
	Data64  string `json:"data64"`
}

type CommandBlockInputData struct {
	BlockId     string            `json:"blockid"`
	InputData64 string            `json:"inputdata64,omitempty"`
	SigName     string            `json:"signame,omitempty"`
	TermSize    *doraobj.TermSize `json:"termsize,omitempty"`
}

type CommandJobInputData struct {
	JobId          string            `json:"jobid"`
	InputSessionId string            `json:"inputsessionid,omitempty"`
	SeqNum         int               `json:"seqnum,omitempty"`
	InputData64    string            `json:"inputdata64,omitempty"`
	SigName        string            `json:"signame,omitempty"`
	TermSize       *doraobj.TermSize `json:"termsize,omitempty"`
}

type CommandWaitForRouteData struct {
	RouteId string `json:"routeid"`
	WaitMs  int    `json:"waitms"`
}

type CommandDeleteBlockData struct {
	BlockId string `json:"blockid"`
}

type CommandEventReadHistoryData struct {
	Event    string `json:"event"`
	Scope    string `json:"scope"`
	MaxItems int    `json:"maxitems"`
}


type CpuDataRequest struct {
	Id    string `json:"id"`
	Count int    `json:"count"`
}

type CpuDataType struct {
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

type CommandGetTempDirData struct {
	FileName string `json:"filename,omitempty"`
}

type CommandWriteTempFileData struct {
	FileName string `json:"filename"`
	Data64   string `json:"data64"`
}

type RemoteInfo struct {
	ClientArch    string `json:"clientarch"`
	ClientOs      string `json:"clientos"`
	ClientVersion string `json:"clientversion"`
	Shell         string `json:"shell"`
	HomeDir       string `json:"homedir"`
}

const (
	TimeSeries_Cpu = "cpu"
)

type TimeSeriesData struct {
	Ts     int64              `json:"ts"`
	Values map[string]float64 `json:"values"`
}

type MetaSettingsType struct {
	doraobj.MetaMapType
}

func (m *MetaSettingsType) UnmarshalJSON(data []byte) error {
	var metaMap doraobj.MetaMapType
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&metaMap); err != nil {
		return err
	}
	*m = MetaSettingsType{MetaMapType: metaMap}
	return nil
}

func (m MetaSettingsType) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.MetaMapType)
}

type ConnStatus struct {
	Status                        string `json:"status"`
	ConnHealthStatus              string `json:"connhealthstatus,omitempty"`
	DshEnabled                    bool   `json:"wshenabled"`
	Connection                    string `json:"connection"`
	Connected                     bool   `json:"connected"`
	HasConnected                  bool   `json:"hasconnected"` // true if it has *ever* connected successfully
	ActiveConnNum                 int    `json:"activeconnnum"`
	Error                         string `json:"error,omitempty"`
	DshError                      string `json:"wsherror,omitempty"`
	NoWshReason                   string `json:"nowshreason,omitempty"`
	DshVersion                    string `json:"wshversion,omitempty"`
	LastActivityBeforeStalledTime int64  `json:"lastactivitybeforestalledtime,omitempty"`
	KeepAliveSentTime             int64  `json:"keepalivesenttime,omitempty"`
}

type WebSelectorOpts struct {
	All   bool `json:"all,omitempty"`
	Inner bool `json:"inner,omitempty"`
}

type CommandWebSelectorData struct {
	WorkspaceId string           `json:"workspaceid"`
	BlockId     string           `json:"blockid"`
	TabId       string           `json:"tabid"`
	Selector    string           `json:"selector"`
	Opts        *WebSelectorOpts `json:"opts,omitempty"`
}

type BlockInfoData struct {
	BlockId     string          `json:"blockid"`
	TabId       string          `json:"tabid"`
	WorkspaceId string          `json:"workspaceid"`
	Block       *doraobj.Block  `json:"block"`
	Files       []*DoraFileInfo `json:"files"`
}

type DoraNotificationOptions struct {
	Title  string `json:"title,omitempty"`
	Body   string `json:"body,omitempty"`
	Silent bool   `json:"silent,omitempty"`
}

type DoraInfoData struct {
	Version   string `json:"version"`
	ClientId  string `json:"clientid"`
	BuildTime string `json:"buildtime"`
	ConfigDir string `json:"configdir"`
	DataDir   string `json:"datadir"`
}

type WorkspaceInfoData struct {
	WindowId      string             `json:"windowid"`
	WorkspaceData *doraobj.Workspace `json:"workspacedata"`
}

type BlocksListRequest struct {
	WindowId    string `json:"windowid,omitempty"`
	WorkspaceId string `json:"workspaceid,omitempty"`
}

type BlocksListEntry struct {
	WindowId    string              `json:"windowid"`
	WorkspaceId string              `json:"workspaceid"`
	TabId       string              `json:"tabid"`
	BlockId     string              `json:"blockid"`
	Meta        doraobj.MetaMapType `json:"meta"`
}

type CommandCaptureBlockScreenshotData struct {
	BlockId string `json:"blockid"`
}

type CommandVarData struct {
	Key      string `json:"key"`
	Val      string `json:"val,omitempty"`
	Remove   bool   `json:"remove,omitempty"`
	ZoneId   string `json:"zoneid"`
	FileName string `json:"filename"`
}

type CommandVarResponseData struct {
	Key    string `json:"key"`
	Val    string `json:"val"`
	Exists bool   `json:"exists"`
}

type CommandDebugTermData struct {
	BlockId string `json:"blockid"`
	Size    int64  `json:"size"`
}

type CommandDebugTermRtnData struct {
	Offset int64  `json:"offset"`
	Data64 string `json:"data64"`
}

type PathCommandData struct {
	PathType     string `json:"pathtype"`
	Open         bool   `json:"open"`
	OpenExternal bool   `json:"openexternal"`
	TabId        string `json:"tabid"`
}

type ActivityDisplayType struct {
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	DPR      float64 `json:"dpr"`
	Internal bool    `json:"internal,omitempty"`
}

type ActivityUpdate struct {
	FgMinutes           int                   `json:"fgminutes,omitempty"`
	ActiveMinutes       int                   `json:"activeminutes,omitempty"`
	OpenMinutes         int                   `json:"openminutes,omitempty"`
	DoraAIFgMinutes     int                   `json:"doraaifgminutes,omitempty"`
	DoraAIActiveMinutes int                   `json:"doraaiactiveminutes,omitempty"`
	NumTabs             int                   `json:"numtabs,omitempty"`
	NewTab              int                   `json:"newtab,omitempty"`
	NumBlocks           int                   `json:"numblocks,omitempty"`
	NumWindows          int                   `json:"numwindows,omitempty"`
	NumWS               int                   `json:"numws,omitempty"`
	NumWSNamed          int                   `json:"numwsnamed,omitempty"`
	NumSSHConn          int                   `json:"numsshconn,omitempty"`
	NumWSLConn          int                   `json:"numwslconn,omitempty"`
	NumMagnify          int                   `json:"nummagnify,omitempty"`
	TermCommandsRun     int                   `json:"termcommandsrun,omitempty"`
	NumPanics           int                   `json:"numpanics,omitempty"`
	NumAIReqs           int                   `json:"numaireqs,omitempty"`
	Startup             int                   `json:"startup,omitempty"`
	Shutdown            int                   `json:"shutdown,omitempty"`
	SetTabTheme         int                   `json:"settabtheme,omitempty"`
	BuildTime           string                `json:"buildtime,omitempty"`
	Displays            []ActivityDisplayType `json:"displays,omitempty"`
	Renderers           map[string]int        `json:"renderers,omitempty"`
	Blocks              map[string]int        `json:"blocks,omitempty"`
	DshCmds             map[string]int        `json:"wshcmds,omitempty"`
	Conn                map[string]int        `json:"conn,omitempty"`
}

type CommandGetRTInfoData struct {
	ORef doraobj.ORef `json:"oref"`
}

type CommandSetRTInfoData struct {
	ORef   doraobj.ORef   `json:"oref"`
	Data   map[string]any `json:"data" tstype:"ObjRTInfo"`
	Delete bool           `json:"delete,omitempty"`
}

type CommandTermGetScrollbackLinesData struct {
	LineStart   int  `json:"linestart"`
	LineEnd     int  `json:"lineend"`
	LastCommand bool `json:"lastcommand"`
}

type CommandTermGetScrollbackLinesRtnData struct {
	TotalLines  int      `json:"totallines"`
	LineStart   int      `json:"linestart"`
	Lines       []string `json:"lines"`
	LastUpdated int64    `json:"lastupdated"`
}

type CommandTermUpdateAttachedJobData struct {
	BlockId string `json:"blockid"`
	JobId   string `json:"jobid,omitempty"`
}

type CommandElectronEncryptData struct {
	PlainText string `json:"plaintext"`
}

type CommandElectronEncryptRtnData struct {
	CipherText     string `json:"ciphertext"`
	StorageBackend string `json:"storagebackend"` // only returned for linux
}

type CommandElectronDecryptData struct {
	CipherText string `json:"ciphertext"`
}

type CommandElectronDecryptRtnData struct {
	PlainText      string `json:"plaintext"`
	StorageBackend string `json:"storagebackend"` // only returned for linux
}

type CommandStreamData struct {
	Id     string `json:"id"`  // streamid
	Seq    int64  `json:"seq"` // start offset (bytes)
	Data64 string `json:"data64,omitempty"`
	Eof    bool   `json:"eof,omitempty"`   // can be set with data or without
	Error  string `json:"error,omitempty"` // stream terminated with error
}

type CommandStreamAckData struct {
	Id     string `json:"id"`               // streamid
	Seq    int64  `json:"seq"`              // next expected byte
	RWnd   int64  `json:"rwnd"`             // receive window size
	Fin    bool   `json:"fin,omitempty"`    // observed end-of-stream (eof or error)
	Delay  int64  `json:"delay,omitempty"`  // ack delay in microseconds (from when data was received to when we sent out ack -- monotonic clock)
	Cancel bool   `json:"cancel,omitempty"` // used to cancel the stream
	Error  string `json:"error,omitempty"`  // reason for cancel (may only be set if cancel is true)
}

type StreamMeta struct {
	Id            string `json:"id"`   // streamid
	RWnd          int64  `json:"rwnd"` // initial receive window size
	ReaderRouteId string `json:"readerrouteid"`
	WriterRouteId string `json:"writerrouteid"`
}

type CommandAuthenticateToJobData struct {
	JobAccessToken string `json:"jobaccesstoken"`
}

type CommandAuthenticateJobManagerData struct {
	JobId        string `json:"jobid"`
	JobAuthToken string `json:"jobauthtoken"`
}

type CommandStartJobData struct {
	Cmd        string            `json:"cmd"`
	Args       []string          `json:"args"`
	Env        map[string]string `json:"env"`
	TermSize   doraobj.TermSize  `json:"termsize"`
	StreamMeta *StreamMeta       `json:"streammeta,omitempty"`
}

type CommandRemoteStartJobData struct {
	Cmd                string            `json:"cmd"`
	Args               []string          `json:"args"`
	Env                map[string]string `json:"env"`
	TermSize           doraobj.TermSize  `json:"termsize"`
	StreamMeta         *StreamMeta       `json:"streammeta,omitempty"`
	JobAuthToken       string            `json:"jobauthtoken"`
	JobId              string            `json:"jobid"`
	MainServerJwtToken string            `json:"mainserverjwttoken"`
	ClientId           string            `json:"clientid"`
	PublicKeyBase64    string            `json:"publickeybase64"`
}

type CommandRemoteReconnectToJobManagerData struct {
	JobId              string `json:"jobid"`
	JobAuthToken       string `json:"jobauthtoken"`
	MainServerJwtToken string `json:"mainserverjwttoken"`
	JobManagerPid      int    `json:"jobmanagerpid"`
	JobManagerStartTs  int64  `json:"jobmanagerstartts"`
}

type CommandRemoteReconnectToJobManagerRtnData struct {
	Success        bool   `json:"success"`
	JobManagerGone bool   `json:"jobmanagergone"`
	Error          string `json:"error,omitempty"`
}

type CommandRemoteDisconnectFromJobManagerData struct {
	JobId string `json:"jobid"`
}

type CommandRemoteTerminateJobManagerData struct {
	JobId             string `json:"jobid"`
	JobManagerPid     int    `json:"jobmanagerpid"`
	JobManagerStartTs int64  `json:"jobmanagerstartts"`
}

type CommandStartJobRtnData struct {
	CmdPid            int   `json:"cmdpid"`
	CmdStartTs        int64 `json:"cmdstartts"`
	JobManagerPid     int   `json:"jobmanagerpid"`
	JobManagerStartTs int64 `json:"jobmanagerstartts"`
}

type CommandJobPrepareConnectData struct {
	StreamMeta StreamMeta       `json:"streammeta"`
	Seq        int64            `json:"seq"`
	TermSize   doraobj.TermSize `json:"termsize"`
}

type CommandJobStartStreamData struct {
}

type CommandJobConnectRtnData struct {
	Seq         int64  `json:"seq"`
	StreamDone  bool   `json:"streamdone,omitempty"`
	StreamError string `json:"streamerror,omitempty"`
	HasExited   bool   `json:"hasexited,omitempty"`
	ExitCode    *int   `json:"exitcode,omitempty"`
	ExitSignal  string `json:"exitsignal,omitempty"`
	ExitErr     string `json:"exiterr,omitempty"`
}

type CommandJobCmdExitedData struct {
	JobId      string `json:"jobid"`
	ExitCode   *int   `json:"exitcode,omitempty"`
	ExitSignal string `json:"exitsignal,omitempty"`
	ExitErr    string `json:"exiterr,omitempty"`
	ExitTs     int64  `json:"exitts,omitempty"`
}

type CommandJobControllerStartJobData struct {
	ConnName string            `json:"connname"`
	JobKind  string            `json:"jobkind"`
	Cmd      string            `json:"cmd"`
	Args     []string          `json:"args"`
	Env      map[string]string `json:"env"`
	TermSize *doraobj.TermSize `json:"termsize,omitempty"`
}

type CommandJobControllerAttachJobData struct {
	JobId   string `json:"jobid"`
	BlockId string `json:"blockid"`
}

type JobManagerStatusUpdate struct {
	JobId            string `json:"jobid"`
	JobManagerStatus string `json:"jobmanagerstatus"`
}

type CommandDoraFileReadStreamData struct {
	ZoneId     string     `json:"zoneid"`
	Name       string     `json:"name"`
	StreamMeta StreamMeta `json:"streammeta"`
}

// see blockstore.go (DoraFile)
type DoraFileInfo struct {
	ZoneId    string   `json:"zoneid"`
	Name      string   `json:"name"`
	Opts      FileOpts `json:"opts"`
	CreatedTs int64    `json:"createdts"`
	Size      int64    `json:"size"`
	ModTs     int64    `json:"modts"`
	Meta      FileMeta `json:"meta"`
}

type CommandBadgeWatchPidData struct {
	Pid     int          `json:"pid"`
	ORef    doraobj.ORef `json:"oref"`
	BadgeId string       `json:"badgeid"`
}

type BlockJobStatusData struct {
	BlockId       string `json:"blockid"`
	JobId         string `json:"jobid"`
	Status        string `json:"status,omitempty" tstype:"null | \"init\" | \"connected\" | \"disconnected\" | \"done\""`
	VersionTs     int64  `json:"versionts"`
	DoneReason    string `json:"donereason,omitempty"`
	StartupError  string `json:"startuperror,omitempty"`
	CmdExitTs     int64  `json:"cmdexitts,omitempty"`
	CmdExitCode   *int   `json:"cmdexitcode,omitempty"`
	CmdExitSignal string `json:"cmdexitsignal,omitempty"`
}

type FocusedBlockData struct {
	BlockId                    string              `json:"blockid"`
	ViewType                   string              `json:"viewtype"`
	Controller                 string              `json:"controller"`
	ConnName                   string              `json:"connname"`
	BlockMeta                  doraobj.MetaMapType `json:"blockmeta"`
	TermJobStatus              *BlockJobStatusData `json:"termjobstatus,omitempty"`
	ConnStatus                 *ConnStatus         `json:"connstatus,omitempty"`
	TermShellIntegrationStatus string              `json:"termshellintegrationstatus,omitempty"`
	TermLastCommand            string              `json:"termlastcommand,omitempty"`
}

// ProcessInfo holds per-process information for the process viewer.
// Mem, MemPct, Cpu, and NumThreads are set to -1 when the data is unavailable
// (e.g. permission denied reading another user's process on macOS).
type ProcessInfo struct {
	Pid        int32   `json:"pid"`
	Ppid       int32   `json:"ppid,omitempty"`
	Command    string  `json:"command,omitempty"`
	Status     string  `json:"status,omitempty"`
	User       string  `json:"user,omitempty"`
	Mem        int64   `json:"mem"`        // resident set size in bytes; -1 if unavailable
	MemPct     float64 `json:"mempct"`     // memory percent; -1 if unavailable
	Cpu        float64 `json:"cpu"`        // cpu percent; -1 if unavailable
	NumThreads int32   `json:"numthreads"` // -1 if unavailable
	Gone       bool    `json:"gone,omitempty"`
}

type ProcessSummary struct {
	Total    int     `json:"total"`
	Load1    float64 `json:"load1,omitempty"`
	Load5    float64 `json:"load5,omitempty"`
	Load15   float64 `json:"load15,omitempty"`
	MemTotal uint64  `json:"memtotal,omitempty"`
	MemUsed  uint64  `json:"memused,omitempty"`
	MemFree  uint64  `json:"memfree,omitempty"`
	NumCPU   int     `json:"numcpu,omitempty"`
	CpuSum   float64 `json:"cpusum,omitempty"`
}

type ProcessListResponse struct {
	Processes     []ProcessInfo  `json:"processes"`
	Summary       ProcessSummary `json:"summary"`
	Ts            int64          `json:"ts"`
	HasCPU        bool           `json:"hascpu,omitempty"`
	Platform      string         `json:"platform,omitempty"`
	TotalCount    int            `json:"totalcount,omitempty"`
	FilteredCount int            `json:"filteredcount,omitempty"`
}

type CommandRemoteProcessListData struct {
	WidgetId   string `json:"widgetid,omitempty"`
	SortBy     string `json:"sortby,omitempty"`
	SortDesc   bool   `json:"sortdesc,omitempty"`
	Start      int    `json:"start,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	TextSearch string `json:"textsearch,omitempty"`
	// LastPidOrder, when set, ignores SortBy/SortDesc/TextSearch and returns processes in the order
	// they were returned in the previous request for this WidgetId (with Gone=true for dead pids).
	LastPidOrder bool `json:"lastpidorder,omitempty"`
	// KeepAlive, when set, overrides all other fields and simply keeps the backend cache alive (returns nil).
	KeepAlive bool `json:"keepalive,omitempty"`
}

type CommandRemoteProcessSignalData struct {
	Pid    int32  `json:"pid"`
	Signal string `json:"signal"`
}
