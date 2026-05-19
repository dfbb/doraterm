// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package jobmanager

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/dfbb/doraterm/pkg/baseds"
	"github.com/dfbb/doraterm/pkg/panichandler"
	"github.com/dfbb/doraterm/pkg/utilds"
	"github.com/dfbb/doraterm/pkg/dorabase"
	"github.com/dfbb/doraterm/pkg/dorajwt"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshrpc/dshclient"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

const JobAccessTokenLabel = "Dora-JobAccessToken"
const JobManagerStartLabel = "Dora-JobManagerStart"
const JobInputQueueTimeout = 100 * time.Millisecond
const JobInputQueueSize = 1000

var DshCmdJobManager JobManager

type JobManager struct {
	ClientId              string
	JobId                 string
	Cmd                   *JobCmd
	JwtPublicKey          []byte
	JobAuthToken          string
	StreamManager         *StreamManager
	InputQueue            *utilds.QuickReorderQueue[dshrpc.CommandJobInputData]
	lock                  sync.Mutex
	attachedClient        *MainServerConn
	connectedStreamClient *MainServerConn
	pendingStreamMeta     *dshrpc.StreamMeta
}

func SetupJobManager(clientId string, jobId string, publicKeyBytes []byte, jobAuthToken string, readyFile *os.File) error {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return fmt.Errorf("job manager only supported on unix systems, not %s", runtime.GOOS)
	}
	DshCmdJobManager.ClientId = clientId
	DshCmdJobManager.JobId = jobId
	DshCmdJobManager.JwtPublicKey = publicKeyBytes
	DshCmdJobManager.JobAuthToken = jobAuthToken
	DshCmdJobManager.StreamManager = MakeStreamManager()
	DshCmdJobManager.InputQueue = utilds.MakeQuickReorderQueue[dshrpc.CommandJobInputData](JobInputQueueSize, JobInputQueueTimeout)
	err := dorajwt.SetPublicKey(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to set public key: %w", err)
	}
	err = MakeJobDomainSocket(clientId, jobId)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			panichandler.PanicHandler("JobManager:processInputQueue", recover())
		}()
		DshCmdJobManager.processInputQueue()
	}()

	fmt.Fprintf(readyFile, JobManagerStartLabel+"\n")
	readyFile.Close()

	err = daemonize(clientId, jobId)
	if err != nil {
		return fmt.Errorf("failed to daemonize: %w", err)
	}

	go func() {
		defer func() {
			panichandler.PanicHandler("JobManager:keepalive", recover())
		}()
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			log.Printf("keepalive: job manager active\n")
		}
	}()

	return nil
}

func (jm *JobManager) processInputQueue() {
	for data := range jm.InputQueue.C() {
		jm.lock.Lock()
		cmd := jm.Cmd
		jm.lock.Unlock()

		if cmd == nil {
			log.Printf("processInputQueue: skipping input, job not started\n")
			continue
		}

		err := cmd.HandleInput(data)
		if err != nil {
			log.Printf("processInputQueue: error handling input: %v\n", err)
		}
	}
}

func (jm *JobManager) GetCmd() *JobCmd {
	jm.lock.Lock()
	defer jm.lock.Unlock()
	return jm.Cmd
}

func (jm *JobManager) sendJobExited() {
	jm.lock.Lock()
	attachedClient := jm.attachedClient
	cmd := jm.Cmd
	jm.lock.Unlock()

	if attachedClient == nil {
		log.Printf("sendJobExited: no attached client, exit notification not sent\n")
		return
	}
	if attachedClient.DshRpc == nil {
		log.Printf("sendJobExited: no wsh rpc connection, exit notification not sent\n")
		return
	}
	if cmd == nil {
		log.Printf("sendJobExited: no cmd, exit notification not sent\n")
		return
	}

	exited, exitData := cmd.GetExitInfo()
	if !exited || exitData == nil {
		log.Printf("sendJobExited: process not exited yet\n")
		return
	}

	exitCodeStr := "nil"
	if exitData.ExitCode != nil {
		exitCodeStr = fmt.Sprintf("%d", *exitData.ExitCode)
	}
	log.Printf("sendJobExited: sending exit notification to main server exitcode=%s signal=%s\n", exitCodeStr, exitData.ExitSignal)
	err := dshclient.JobCmdExitedCommand(attachedClient.DshRpc, *exitData, nil)
	if err != nil {
		log.Printf("sendJobExited: error sending exit notification: %v\n", err)
	}
}

func (jm *JobManager) GetJobAuthInfo() (string, string) {
	jm.lock.Lock()
	defer jm.lock.Unlock()
	return jm.JobId, jm.JobAuthToken
}

func (jm *JobManager) IsJobStarted() bool {
	jm.lock.Lock()
	defer jm.lock.Unlock()
	return jm.Cmd != nil
}

func (jm *JobManager) connectToStreamHelper_withlock(mainServerConn *MainServerConn, streamMeta dshrpc.StreamMeta, seq int64) (int64, error) {
	rwndSize := int(streamMeta.RWnd)
	if rwndSize < 0 {
		return 0, fmt.Errorf("invalid rwnd size: %d", rwndSize)
	}

	if jm.connectedStreamClient != nil {
		log.Printf("connectToStreamHelper: disconnecting existing client\n")
		oldStreamId := jm.StreamManager.GetStreamId()
		jm.StreamManager.ClientDisconnected()
		if oldStreamId != "" {
			mainServerConn.DshRpc.StreamBroker.DetachStreamWriter(oldStreamId)
			log.Printf("connectToStreamHelper: detached old stream id=%s\n", oldStreamId)
		}
		jm.connectedStreamClient = nil
	}
	dataSender := &routedDataSender{
		wshRpc: mainServerConn.DshRpc,
		route:  streamMeta.ReaderRouteId,
	}
	serverSeq, err := jm.StreamManager.ClientConnected(
		streamMeta.Id,
		dataSender,
		rwndSize,
		seq,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to connect client: %w", err)
	}
	jm.connectedStreamClient = mainServerConn
	return serverSeq, nil
}

func (jm *JobManager) disconnectFromStreamHelper(mainServerConn *MainServerConn) {
	jm.lock.Lock()
	defer jm.lock.Unlock()
	if jm.connectedStreamClient == nil || jm.connectedStreamClient != mainServerConn {
		return
	}
	jm.StreamManager.ClientDisconnected()
	jm.connectedStreamClient = nil
}

func (jm *JobManager) SetAttachedClient(msc *MainServerConn) {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	if jm.attachedClient != nil {
		log.Printf("SetAttachedClient: kicking out existing client\n")
		jm.attachedClient.Close()
	}
	jm.attachedClient = msc
}

func (jm *JobManager) StartJob(msc *MainServerConn, data dshrpc.CommandStartJobData) (*dshrpc.CommandStartJobRtnData, error) {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	if jm.Cmd != nil {
		log.Printf("StartJob: job already started")
		return nil, fmt.Errorf("job already started")
	}

	cmdDef := CmdDef{
		Cmd:      data.Cmd,
		Args:     data.Args,
		Env:      data.Env,
		TermSize: data.TermSize,
	}
	log.Printf("StartJob: creating job cmd for jobid=%s", jm.JobId)
	jobCmd, err := MakeJobCmd(jm.JobId, cmdDef)
	if err != nil {
		log.Printf("StartJob: failed to make job cmd: %v", err)
		return nil, fmt.Errorf("failed to start job: %w", err)
	}
	jm.Cmd = jobCmd
	log.Printf("StartJob: job cmd created successfully")

	if data.StreamMeta != nil {
		serverSeq, err := jm.connectToStreamHelper_withlock(msc, *data.StreamMeta, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to connect stream: %w", err)
		}
		err = msc.DshRpc.StreamBroker.AttachStreamWriter(data.StreamMeta, jm.StreamManager)
		if err != nil {
			return nil, fmt.Errorf("failed to attach stream writer: %w", err)
		}
		log.Printf("StartJob: connected stream streamid=%s serverSeq=%d\n", data.StreamMeta.Id, serverSeq)
	}

	cmd, cmdPty := jobCmd.GetCmd()
	if cmdPty != nil {
		log.Printf("StartJob: attaching pty reader to stream manager")
		err = jm.StreamManager.AttachReader(cmdPty)
		if err != nil {
			log.Printf("StartJob: failed to attach reader: %v", err)
			return nil, fmt.Errorf("failed to attach reader to stream manager: %w", err)
		}
		log.Printf("StartJob: pty reader attached successfully")
	} else {
		log.Printf("StartJob: no pty to attach")
	}

	if cmd == nil || cmd.Process == nil {
		log.Printf("StartJob: cmd or process is nil")
		return nil, fmt.Errorf("cmd or process is nil")
	}
	cmdPid := cmd.Process.Pid
	cmdProc, err := process.NewProcess(int32(cmdPid))
	if err != nil {
		log.Printf("StartJob: failed to get cmd process: %v", err)
		return nil, fmt.Errorf("failed to get cmd process: %w", err)
	}
	cmdStartTs, err := cmdProc.CreateTime()
	if err != nil {
		log.Printf("StartJob: failed to get cmd start time: %v", err)
		return nil, fmt.Errorf("failed to get cmd start time: %w", err)
	}

	jobManagerPid := os.Getpid()
	jobManagerProc, err := process.NewProcess(int32(jobManagerPid))
	if err != nil {
		log.Printf("StartJob: failed to get job manager process: %v", err)
		return nil, fmt.Errorf("failed to get job manager process: %w", err)
	}
	jobManagerStartTs, err := jobManagerProc.CreateTime()
	if err != nil {
		log.Printf("StartJob: failed to get job manager start time: %v", err)
		return nil, fmt.Errorf("failed to get job manager start time: %w", err)
	}

	log.Printf("StartJob: job started successfully cmdPid=%d cmdStartTs=%d jobManagerPid=%d jobManagerStartTs=%d", cmdPid, cmdStartTs, jobManagerPid, jobManagerStartTs)
	return &dshrpc.CommandStartJobRtnData{
		CmdPid:            cmdPid,
		CmdStartTs:        cmdStartTs,
		JobManagerPid:     jobManagerPid,
		JobManagerStartTs: jobManagerStartTs,
	}, nil
}

func (jm *JobManager) PrepareConnect(msc *MainServerConn, data dshrpc.CommandJobPrepareConnectData) (*dshrpc.CommandJobConnectRtnData, error) {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	if jm.Cmd == nil {
		return nil, fmt.Errorf("job not started")
	}

	err := jm.Cmd.SetTermSize(data.TermSize)
	if err != nil {
		log.Printf("PrepareConnect: failed to set term size: %v\n", err)
	}

	rtnData := &dshrpc.CommandJobConnectRtnData{}
	streamDone, streamError := jm.StreamManager.GetStreamDoneInfo()

	if streamDone {
		log.Printf("PrepareConnect: stream already done, skipping connection streamError=%q\n", streamError)
		rtnData.Seq = data.Seq
		rtnData.StreamDone = true
		rtnData.StreamError = streamError
	} else {
		corkedStreamMeta := data.StreamMeta
		corkedStreamMeta.RWnd = 0
		serverSeq, err := jm.connectToStreamHelper_withlock(msc, corkedStreamMeta, data.Seq)
		if err != nil {
			return nil, err
		}
		jm.pendingStreamMeta = &data.StreamMeta
		rtnData.Seq = serverSeq
		rtnData.StreamDone = false
	}

	hasExited, exitData := jm.Cmd.GetExitInfo()
	if hasExited && exitData != nil {
		rtnData.HasExited = true
		rtnData.ExitCode = exitData.ExitCode
		rtnData.ExitSignal = exitData.ExitSignal
		rtnData.ExitErr = exitData.ExitErr
	}

	log.Printf("PrepareConnect: streamid=%s clientSeq=%d serverSeq=%d streamDone=%v streamError=%q hasExited=%v\n", data.StreamMeta.Id, data.Seq, rtnData.Seq, rtnData.StreamDone, rtnData.StreamError, hasExited)
	return rtnData, nil
}

func (jm *JobManager) StartStream(msc *MainServerConn) error {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	if jm.Cmd == nil {
		return fmt.Errorf("job not started")
	}
	if jm.pendingStreamMeta == nil {
		return fmt.Errorf("no pending stream (call PrepareConnect first)")
	}

	err := msc.DshRpc.StreamBroker.AttachStreamWriter(jm.pendingStreamMeta, jm.StreamManager)
	if err != nil {
		return fmt.Errorf("failed to attach stream writer: %w", err)
	}

	err = jm.StreamManager.SetRwndSize(int(jm.pendingStreamMeta.RWnd))
	if err != nil {
		return fmt.Errorf("failed to set rwnd size: %w", err)
	}

	log.Printf("StartStream: streamid=%s rwnd=%d streaming started\n", jm.pendingStreamMeta.Id, jm.pendingStreamMeta.RWnd)
	jm.pendingStreamMeta = nil
	return nil
}

func MakeJobDomainSocket(clientId string, jobId string) error {
	socketDir := filepath.Join("/tmp", fmt.Sprintf("doraterm-%d", os.Getuid()))
	err := os.MkdirAll(socketDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	socketPath := dorabase.GetRemoteJobSocketPath(jobId)

	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on domain socket: %w", err)
	}

	go func() {
		defer func() {
			panichandler.PanicHandler("MakeJobDomainSocket:accept", recover())
			listener.Close()
			os.Remove(socketPath)
		}()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("error accepting connection: %v\n", err)
				return
			}
			go handleJobDomainSocketClient(conn)
		}
	}()

	return nil
}

func handleJobDomainSocketClient(conn net.Conn) {
	inputCh := make(chan baseds.RpcInputChType, dshutil.DefaultInputChSize)
	outputCh := make(chan []byte, dshutil.DefaultOutputChSize)

	serverImpl := &MainServerConn{
		Conn:    conn,
		inputCh: inputCh,
	}
	rpcCtx := dshrpc.RpcContext{}
	wshRpc := dshutil.MakeDshRpcWithChannels(inputCh, outputCh, rpcCtx, serverImpl, "job-domain")
	serverImpl.DshRpc = wshRpc
	defer DshCmdJobManager.disconnectFromStreamHelper(serverImpl)

	go func() {
		defer func() {
			panichandler.PanicHandler("handleJobDomainSocketClient:AdaptOutputChToStream", recover())
		}()
		defer serverImpl.Close()
		writeErr := dshutil.AdaptOutputChToStream(outputCh, conn)
		if writeErr != nil {
			log.Printf("error writing to domain socket: %v\n", writeErr)
		}
	}()

	go func() {
		defer func() {
			panichandler.PanicHandler("handleJobDomainSocketClient:AdaptStreamToMsgCh", recover())
		}()
		defer serverImpl.Close()
		dshutil.AdaptStreamToMsgCh(conn, inputCh, nil)
	}()

	_ = wshRpc
}
