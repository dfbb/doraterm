// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/dfbb/doraterm/pkg/baseds"
	"github.com/dfbb/doraterm/pkg/panichandler"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dshrpc"
)

const (
	DefaultRoute     = "dorasrv"
	ElectronRoute    = "electron"
	ControlRoute     = "$control"      // control plane route
	ControlRootRoute = "$control:root" // control plane route to root router

	ControlPrefix = "$"

	RoutePrefix_Conn       = "conn:"
	RoutePrefix_Controller = "controller:"
	RoutePrefix_Proc       = "proc:"
	RoutePrefix_Tab        = "tab:"
	RoutePrefix_FeBlock    = "feblock:"
	RoutePrefix_Builder    = "builder:"
	RoutePrefix_Link       = "link:"
	RoutePrefix_Job        = "job:"
	RoutePrefix_Bare       = "bare:"
)

const RouterInputChQueueSize = 100

var BacklogLogThresholds = map[int]bool{1: true, 5: true, 10: true, 20: true, 30: true, 40: true, 50: true, 100: true, 200: true, 500: true, 1000: true}

// this works like a network switch

// TODO maybe move the wps integration here instead of in wshserver

type routeInfo struct {
	RpcId         string
	SourceRouteId string
	DestRouteId   string
}

const LinkKind_Leaf = "leaf"
const LinkKind_Router = "router"

type linkMeta struct {
	linkId        baseds.LinkId
	trusted       bool
	linkKind      string
	sourceRouteId string
	client        AbstractRpcClient
}

func (lm *linkMeta) Name() string {
	return fmt.Sprintf("%d#[%s]", lm.linkId, lm.client.GetPeerInfo())
}

type rpcRoutingInfo struct {
	rpcId        string
	sourceLinkId baseds.LinkId
	destRouteId  string
}

type messageWrap struct {
	msgBytes []byte
	debugStr string
}

type backlogMessageWrap struct {
	msgBytes      []byte
	ingressLinkId baseds.LinkId
	debugStr      string
}

type DshRouter struct {
	lock           *sync.Mutex
	isRootRouter   bool
	nextLinkId     baseds.LinkId
	upstreamLinkId baseds.LinkId
	inputCh        chan baseds.RpcInputChType
	rpcMap         map[string]rpcRoutingInfo // rpcid => routeinfo
	routeMap       map[string]baseds.LinkId  // routeid => linkid
	linkMap        map[baseds.LinkId]*linkMeta

	upstreamBufLock     sync.Mutex
	upstreamBufCond     *sync.Cond
	upstreamBuf         []messageWrap
	upstreamLoopStarted bool

	linkBacklogCond      *sync.Cond
	linkMsgBacklog       map[baseds.LinkId][]backlogMessageWrap
	backlogHighWaterMark map[baseds.LinkId]int

	controlRpc *DshRpc
}

func MakeConnectionRouteId(connId string) string {
	return "conn:" + connId
}

func MakeControllerRouteId(blockId string) string {
	return "controller:" + blockId
}

func MakeProcRouteId(procId string) string {
	return "proc:" + procId
}

func MakeRandomProcRouteId() string {
	return MakeProcRouteId(uuid.New().String())
}

func MakeTabRouteId(tabId string) string {
	return "tab:" + tabId
}

func MakeFeBlockRouteId(blockId string) string {
	return "feblock:" + blockId
}

func MakeBuilderRouteId(builderId string) string {
	return "builder:" + builderId
}

func MakeJobRouteId(jobId string) string {
	return "job:" + jobId
}

func MakeLinkRouteId(linkId baseds.LinkId) string {
	return fmt.Sprintf("%s%d", RoutePrefix_Link, linkId)
}

var DefaultRouter *DshRouter

func NewDshRouter() *DshRouter {
	rtn := &DshRouter{
		lock:                 &sync.Mutex{},
		nextLinkId:           0,
		upstreamLinkId:       baseds.NoLinkId,
		inputCh:              make(chan baseds.RpcInputChType, RouterInputChQueueSize),
		rpcMap:               make(map[string]rpcRoutingInfo),
		linkMap:              make(map[baseds.LinkId]*linkMeta),
		routeMap:             make(map[string]baseds.LinkId),
		linkMsgBacklog:       make(map[baseds.LinkId][]backlogMessageWrap),
		backlogHighWaterMark: make(map[baseds.LinkId]int),
	}
	rtn.upstreamBufCond = sync.NewCond(&rtn.upstreamBufLock)
	rtn.linkBacklogCond = sync.NewCond(rtn.lock)
	rtn.registerControlPlane()
	go rtn.runServer()
	go rtn.processBacklog()
	return rtn
}

func (router *DshRouter) IsRootRouter() bool {
	router.lock.Lock()
	defer router.lock.Unlock()
	return router.isRootRouter
}

func (router *DshRouter) GetControlRpc() *DshRpc {
	return router.controlRpc
}

func (router *DshRouter) SetAsRootRouter() {
	router.lock.Lock()
	defer router.lock.Unlock()
	router.isRootRouter = true

	// also bind $control:root to the control RPC
	linkId := router.routeMap[ControlRoute]
	if linkId != baseds.NoLinkId {
		router.routeMap[ControlRootRoute] = linkId
		log.Printf("dshrouter registered control:root route linkid=%d", linkId)
	}
}

func noRouteErr(routeId string) error {
	if routeId == "" {
		return errors.New("no default route")
	}
	return fmt.Errorf("no route for %q", routeId)
}

func (router *DshRouter) SendEvent(routeId string, event dps.DoraEvent) {
	defer func() {
		panichandler.PanicHandler("DshRouter.SendEvent", recover())
	}()
	lm := router.getLinkForRoute(routeId)
	if lm == nil {
		return
	}
	msg := RpcMessage{
		Command: dshrpc.Command_EventRecv,
		Route:   routeId,
		Data:    event,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		// nothing to do
		return
	}
	router.sendRpcMessageToLink(lm.linkId, lm.client, msgBytes, baseds.NoLinkId, "eventrecv")
}

func (router *DshRouter) handleNoRoute(msg RpcMessage, ingressLinkId baseds.LinkId) {
	lm := router.getLinkMeta(ingressLinkId)
	if lm == nil {
		return
	}
	nrErr := noRouteErr(msg.Route)
	if msg.ReqId == "" {
		if msg.Command == dshrpc.Command_Message {
			// to prevent infinite loops
			return
		}
		// no response needed, but send message back to source
		respMsg := RpcMessage{
			Command: dshrpc.Command_Message,
			Route:   msg.Source,
			Source:  ControlRoute,
			Data:    dshrpc.CommandMessageData{Message: nrErr.Error()},
		}
		respBytes, _ := json.Marshal(respMsg)
		router.sendRpcMessageToLink(lm.linkId, lm.client, respBytes, baseds.NoLinkId, "no-route-err")
		return
	}
	// send error response
	response := RpcMessage{
		ResId: msg.ReqId,
		Error: nrErr.Error(),
	}
	respBytes, _ := json.Marshal(response)
	router.sendRoutedMessage(respBytes, msg.Source, msg.Command, baseds.NoLinkId)
}

func (router *DshRouter) registerRouteInfo(rpcId string, sourceLinkId baseds.LinkId, destRouteId string) {
	if rpcId == "" {
		return
	}
	router.lock.Lock()
	defer router.lock.Unlock()
	router.rpcMap[rpcId] = rpcRoutingInfo{
		rpcId:        rpcId,
		sourceLinkId: sourceLinkId,
		destRouteId:  destRouteId,
	}
}

func (router *DshRouter) unregisterRouteInfo(rpcId string) {
	router.lock.Lock()
	defer router.lock.Unlock()
	delete(router.rpcMap, rpcId)
}

func (router *DshRouter) getRouteInfo(rpcId string) *rpcRoutingInfo {
	router.lock.Lock()
	defer router.lock.Unlock()
	rtn, ok := router.rpcMap[rpcId]
	if !ok {
		return nil
	}
	return &rtn
}

// returns true if message was sent, false if failed
func (router *DshRouter) sendRoutedMessage(msgBytes []byte, routeId string, commandName string, ingressLinkId baseds.LinkId) bool {
	if strings.HasPrefix(routeId, RoutePrefix_Link) {
		linkIdStr := strings.TrimPrefix(routeId, RoutePrefix_Link)
		linkIdInt, err := strconv.ParseInt(linkIdStr, 10, 32)
		if err == nil {
			return router.sendMessageToLink(msgBytes, baseds.LinkId(linkIdInt), ingressLinkId)
		}
	}
	lm := router.getLinkForRoute(routeId)
	if lm != nil {
		router.sendRpcMessageToLink(lm.linkId, lm.client, msgBytes, ingressLinkId, "route")
		return true
	}
	upstreamLinkId, upstream := router.getUpstreamClient()
	if upstream != nil {
		router.sendRpcMessageToLink(upstreamLinkId, upstream, msgBytes, ingressLinkId, "route-upstream")
		return true
	}
	if commandName != "" {
		log.Printf("[router] no rpc for route id %q command:%s\n", routeId, commandName)
	} else {
		log.Printf("[router] no rpc for route id %q\n", routeId)
	}
	return false
}

func (router *DshRouter) sendMessageToLink(msgBytes []byte, linkId baseds.LinkId, ingressLinkId baseds.LinkId) bool {
	lm := router.getLinkMeta(linkId)
	if lm == nil {
		return false
	}
	router.sendRpcMessageToLink(lm.linkId, lm.client, msgBytes, ingressLinkId, "link")
	return true
}

func (router *DshRouter) addToBacklog_withlock(linkId baseds.LinkId, msgBytes []byte, ingressLinkId baseds.LinkId, debugStr string) {
	mapWasEmpty := len(router.linkMsgBacklog) == 0
	backlog := router.linkMsgBacklog[linkId]
	backlog = append(backlog, backlogMessageWrap{msgBytes: msgBytes, ingressLinkId: ingressLinkId, debugStr: debugStr})
	router.linkMsgBacklog[linkId] = backlog

	newLen := len(backlog)
	highWater := router.backlogHighWaterMark[linkId]

	if BacklogLogThresholds[newLen] && highWater < newLen {
		log.Printf("[router] backlog for linkid=%d reached %d messages\n", linkId, newLen)
	}

	if newLen > highWater {
		router.backlogHighWaterMark[linkId] = newLen
	}

	if mapWasEmpty {
		router.linkBacklogCond.Signal()
	}
}

func (router *DshRouter) sendRpcMessageToLink(linkId baseds.LinkId, client AbstractRpcClient, msgBytes []byte, ingressLinkId baseds.LinkId, debugStr string) {
	router.lock.Lock()
	defer router.lock.Unlock()
	sent := false
	backlog := router.linkMsgBacklog[linkId]
	if len(backlog) == 0 {
		sent = client.SendRpcMessage(msgBytes, ingressLinkId, debugStr)
	}
	if !sent {
		router.addToBacklog_withlock(linkId, msgBytes, ingressLinkId, debugStr)
	}
}

func (router *DshRouter) runServer() {
	for input := range router.inputCh {
		msgBytes := input.MsgBytes
		var msg RpcMessage
		err := json.Unmarshal(msgBytes, &msg)
		if err != nil {
			fmt.Println("error unmarshalling message: ", err)
			continue
		}
		routeId := msg.Route
		if msg.Command != "" {
			// new comand, setup new rpc
			ok := router.sendRoutedMessage(msgBytes, routeId, msg.Command, input.IngressLinkId)
			if !ok {
				router.handleNoRoute(msg, input.IngressLinkId)
				continue
			}
			router.registerRouteInfo(msg.ReqId, input.IngressLinkId, routeId)
			continue
		}
		// look at reqid or resid to route correctly
		if msg.ReqId != "" {
			routeInfo := router.getRouteInfo(msg.ReqId)
			if routeInfo == nil {
				// no route info, nothing to do
				continue
			}
			// no need to check the return value here (noop if failed)
			router.sendRoutedMessage(msgBytes, routeInfo.destRouteId, "", input.IngressLinkId)
			continue
		} else if msg.ResId != "" {
			routeInfo := router.getRouteInfo(msg.ResId)
			if routeInfo == nil {
				// no route info, nothing to do
				continue
			}
			router.sendMessageToLink(msgBytes, routeInfo.sourceLinkId, input.IngressLinkId)
			if !msg.Cont {
				router.unregisterRouteInfo(msg.ResId)
			}
			continue
		} else {
			// this is a bad message (no command, reqid, or resid)
			continue
		}
	}
}

func (router *DshRouter) WaitForRegister(ctx context.Context, routeId string) error {
	for {
		if router.getLinkForRoute(routeId) != nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Millisecond):
			continue
		}
	}
}

// this will never block, can be called while holding router.Lock
func (router *DshRouter) queueUpstreamMessage(msgBytes []byte, debugStr string) {
	_, upstream := router.getUpstreamClient()
	if upstream == nil {
		return
	}
	router.upstreamBufLock.Lock()
	defer router.upstreamBufLock.Unlock()
	router.upstreamBuf = append(router.upstreamBuf, messageWrap{msgBytes: msgBytes, debugStr: debugStr})
	if !router.upstreamLoopStarted {
		router.upstreamLoopStarted = true
		go router.runUpstreamBufferLoop()
	}
	router.upstreamBufCond.Signal()
}

func (router *DshRouter) runUpstreamBufferLoop() {
	defer func() {
		panichandler.PanicHandler("DshRouter:runUpstreamBufferLoop", recover())
	}()
	for {
		router.upstreamBufLock.Lock()
		for len(router.upstreamBuf) == 0 {
			router.upstreamBufCond.Wait()
		}
		msg := router.upstreamBuf[0]
		router.upstreamBuf = router.upstreamBuf[1:]
		router.upstreamBufLock.Unlock()

		upstreamLinkId, upstream := router.getUpstreamClient()
		if upstream != nil {
			router.sendRpcMessageToLink(upstreamLinkId, upstream, msg.msgBytes, baseds.NoLinkId, msg.debugStr)
		}
	}
}

func (router *DshRouter) drainLinkBacklog_withLock(linkId baseds.LinkId, lm *linkMeta, backlog []backlogMessageWrap) []backlogMessageWrap {
	for len(backlog) > 0 {
		msg := backlog[0]
		sent := lm.client.SendRpcMessage(msg.msgBytes, msg.ingressLinkId, msg.debugStr)
		if !sent {
			return backlog
		}
		backlog = backlog[1:]
	}
	return backlog
}

func (router *DshRouter) processOneBacklogRound() {
	router.lock.Lock()
	defer router.lock.Unlock()
	for linkId, backlog := range router.linkMsgBacklog {
		lm := router.linkMap[linkId]
		if lm == nil {
			highWater := router.backlogHighWaterMark[linkId]
			if highWater > 0 {
				log.Printf("[router] backlog for linkid=%d cleared, link gone (highwater mark was %d messages)\n", linkId, highWater)
			}
			delete(router.linkMsgBacklog, linkId)
			delete(router.backlogHighWaterMark, linkId)
			continue
		}
		newBacklog := router.drainLinkBacklog_withLock(linkId, lm, backlog)
		if len(newBacklog) == 0 {
			highWater := router.backlogHighWaterMark[linkId]
			if highWater > 0 {
				log.Printf("[router] backlog for linkid=%d cleared (highwater mark was %d messages)\n", linkId, highWater)
			}
			delete(router.linkMsgBacklog, linkId)
			delete(router.backlogHighWaterMark, linkId)
			continue
		}
		router.linkMsgBacklog[linkId] = newBacklog
	}
}

func (router *DshRouter) processBacklog() {
	defer func() {
		panichandler.PanicHandler("DshRouter:processBacklog", recover())
	}()
	for {
		router.lock.Lock()
		for len(router.linkMsgBacklog) == 0 {
			router.linkBacklogCond.Wait()
		}
		router.lock.Unlock()
		router.processOneBacklogRound()
		time.Sleep(50 * time.Millisecond)
	}
}

func (router *DshRouter) RegisterUntrustedLink(client AbstractRpcClient) baseds.LinkId {
	router.lock.Lock()
	defer router.lock.Unlock()
	router.nextLinkId++
	linkId := router.nextLinkId
	lm := &linkMeta{
		linkId:  linkId,
		trusted: false,
		client:  client,
	}
	log.Printf("dshrouter register link %s", lm.Name())
	router.linkMap[linkId] = lm
	go router.runLinkClientRecvLoop(linkId, client)
	return linkId
}

func (router *DshRouter) trustLink(linkId baseds.LinkId, linkKind string) {
	router.lock.Lock()
	defer router.lock.Unlock()
	lm := router.linkMap[linkId]
	if lm == nil {
		return
	}
	log.Printf("dshrouter trust link %s kind=%s", lm.Name(), linkKind)
	lm.trusted = true
	lm.linkKind = linkKind
}

func (router *DshRouter) runLinkClientRecvLoop(linkId baseds.LinkId, client AbstractRpcClient) {
	defer func() {
		panichandler.PanicHandler("DshRouter:runLinkClientRecvLoop", recover())
	}()
	exitReason := "unknown"
	lmForLog := router.getLinkMeta(linkId)
	linkName := fmt.Sprintf("%d", linkId)
	if lmForLog != nil {
		linkName = lmForLog.Name()
	}
	log.Printf("link recvloop start for %s", linkName)
	defer log.Printf("link recvloop done for %s (%s)", linkName, exitReason)
	for {
		msgBytes, ok := client.RecvRpcMessage()
		if !ok {
			exitReason = "recv-eof"
			break
		}
		var rpcMsg RpcMessage
		err := json.Unmarshal(msgBytes, &rpcMsg)
		if err != nil {
			continue
		}
		lm := router.getLinkMeta(linkId)
		if lm == nil {
			exitReason = "link-gone"
			break
		}
		if rpcMsg.IsRpcRequest() {
			if lm.sourceRouteId != "" {
				rpcMsg.Source = lm.sourceRouteId
			}
			if rpcMsg.Route == "" {
				rpcMsg.Route = DefaultRoute
			}
			msgBytes, err = json.Marshal(rpcMsg)
			if err != nil {
				continue
			}
			// allow control routes even for untrusted links (for authentication)
			isControlRoute := rpcMsg.Route == ControlRoute || rpcMsg.Route == ControlRootRoute
			if !lm.trusted {
				if !isControlRoute {
					sendControlUnauthenticatedErrorResponse(rpcMsg, *lm, router)
					continue
				}
				log.Printf("dshrouter control-msg route=%s link=%s command=%s source=%s", rpcMsg.Route, lm.Name(), rpcMsg.Command, rpcMsg.Source)
			}
		} else {
			// non-request messages (responses)
			if !lm.trusted {
				// allow responses to RPCs we initiated
				if rpcMsg.ResId == "" || router.getRouteInfo(rpcMsg.ResId) == nil {
					continue
				}
			}
		}
		router.inputCh <- baseds.RpcInputChType{MsgBytes: msgBytes, IngressLinkId: linkId}
	}
}

// synchronized, returns a copy
func (router *DshRouter) getLinkMeta(linkId baseds.LinkId) *linkMeta {
	if linkId == baseds.NoLinkId {
		return nil
	}
	router.lock.Lock()
	defer router.lock.Unlock()
	lm := router.linkMap[linkId]
	if lm == nil {
		return nil
	}
	lmCopy := *lm
	return &lmCopy
}

// synchronized, returns a copy
func (router *DshRouter) getLinkForRoute(routeId string) *linkMeta {
	if routeId == "" {
		return nil
	}
	router.lock.Lock()
	defer router.lock.Unlock()
	linkId := router.routeMap[routeId]
	if linkId == baseds.NoLinkId {
		return nil
	}
	lm := router.linkMap[linkId]
	if lm == nil {
		return nil
	}
	lmCopy := *lm
	return &lmCopy
}

func (router *DshRouter) GetLinkIdForRoute(routeId string) baseds.LinkId {
	lm := router.getLinkForRoute(routeId)
	if lm == nil {
		return baseds.NoLinkId
	}
	return lm.linkId
}

// only for leaves
func (router *DshRouter) RegisterTrustedLeaf(rpc AbstractRpcClient, routeId string) (baseds.LinkId, error) {
	if !isBindableRouteId(routeId) {
		return 0, fmt.Errorf("invalid routeid %q", routeId)
	}
	linkId := router.RegisterUntrustedLink(rpc)
	router.trustLink(linkId, LinkKind_Leaf)
	router.bindRoute(linkId, routeId, true)
	return linkId, nil
}

// only for routers
func (router *DshRouter) RegisterTrustedRouter(rpc AbstractRpcClient) baseds.LinkId {
	linkId := router.RegisterUntrustedLink(rpc)
	router.trustLink(linkId, LinkKind_Router)
	return linkId
}

func (router *DshRouter) RegisterUpstream(rpc AbstractRpcClient) baseds.LinkId {
	if router.IsRootRouter() {
		panic("cannot register upstream for root router")
	}
	linkId := router.RegisterUntrustedLink(rpc)
	router.trustLink(linkId, LinkKind_Router)
	router.lock.Lock()
	defer router.lock.Unlock()
	router.upstreamLinkId = linkId
	return linkId
}

func (router *DshRouter) registerControlPlane() {
	controlImpl := &DshRouterControlImpl{Router: router}
	controlRpcCtx := dshrpc.RpcContext{RouteId: ControlRoute}
	router.controlRpc = MakeDshRpc(controlRpcCtx, controlImpl, "control")

	linkId := router.RegisterUntrustedLink(router.controlRpc)
	router.trustLink(linkId, LinkKind_Leaf)

	router.lock.Lock()
	defer router.lock.Unlock()
	lm := router.linkMap[linkId]
	if lm != nil {
		lm.sourceRouteId = ControlRoute
		router.routeMap[ControlRoute] = linkId
		log.Printf("dshrouter registered control route %q linkid=%d", ControlRoute, linkId)
	}
}

func (router *DshRouter) announceUpstream(routeId string) {
	msg := RpcMessage{
		Command: dshrpc.Command_RouteAnnounce,
		Route:   ControlRoute,
		Source:  routeId,
	}
	msgBytes, _ := json.Marshal(msg)
	router.queueUpstreamMessage(msgBytes, "upstream-announce")
}

func (router *DshRouter) unannounceUpstream(routeId string) {
	msg := RpcMessage{
		Command: dshrpc.Command_RouteUnannounce,
		Route:   ControlRoute,
		Source:  routeId,
	}
	msgBytes, _ := json.Marshal(msg)
	router.queueUpstreamMessage(msgBytes, "upstream-unannounce")
}

func (router *DshRouter) getRoutesForLink(linkId baseds.LinkId) []string {
	router.lock.Lock()
	defer router.lock.Unlock()
	var routes []string
	for routeId, mappedLinkId := range router.routeMap {
		if mappedLinkId == linkId {
			routes = append(routes, routeId)
		}
	}
	return routes
}

func (router *DshRouter) UnregisterLink(linkId baseds.LinkId) {
	routes := router.getRoutesForLink(linkId)
	for _, routeId := range routes {
		router.unbindRoute(linkId, routeId)
	}
	router.lock.Lock()
	defer router.lock.Unlock()
	lm := router.linkMap[linkId]
	if lm != nil {
		log.Printf("dshrouter unregister link %s", lm.Name())
	}
	delete(router.linkMap, linkId)
	if router.upstreamLinkId == linkId {
		router.upstreamLinkId = baseds.NoLinkId
	}
}

func isBindableRouteId(routeId string) bool {
	if routeId == "" || strings.HasPrefix(routeId, ControlPrefix) || strings.HasPrefix(routeId, RoutePrefix_Link) {
		return false
	}
	return true
}

func (router *DshRouter) unbindRouteLocally(linkId baseds.LinkId, routeId string) error {
	if linkId == baseds.NoLinkId {
		return fmt.Errorf("cannot unbind %q to NoLinkId", routeId)
	}
	router.lock.Lock()
	defer router.lock.Unlock()
	if router.routeMap[routeId] == linkId {
		delete(router.routeMap, routeId)
	}
	return nil
}

func (router *DshRouter) unbindRoute(linkId baseds.LinkId, routeId string) error {
	err := router.unbindRouteLocally(linkId, routeId)
	if err != nil {
		return err
	}
	lm := router.getLinkMeta(linkId)
	if lm != nil {
		log.Printf("dshrouter unbind route %q from %s", routeId, lm.Name())
	}
	router.unannounceUpstream(routeId)
	if router.IsRootRouter() {
		router.unsubscribeFromBroker(routeId)
	}
	return nil
}

func (router *DshRouter) bindRouteLocally(linkId baseds.LinkId, routeId string, isSourceRoute bool) error {
	if linkId == baseds.NoLinkId {
		return fmt.Errorf("cannot bindroute %q to NoLinkId", routeId)
	}
	if !isBindableRouteId(routeId) {
		return fmt.Errorf("router cannot register %q route (invalid routeid)", routeId)
	}
	router.lock.Lock()
	defer router.lock.Unlock()
	lm := router.linkMap[linkId]
	if lm == nil {
		return fmt.Errorf("cannot bind route %q, no link with id %d found", routeId, linkId)
	}
	if !lm.trusted {
		return fmt.Errorf("cannot bind route %q, link %d is not trusted", routeId, linkId)
	}
	if isSourceRoute {
		if lm.linkKind != LinkKind_Leaf {
			return fmt.Errorf("cannot bind source route %q to link %d (link is not a leaf)", routeId, linkId)
		}
		if lm.sourceRouteId != "" && lm.sourceRouteId != routeId {
			return fmt.Errorf("cannot bind source route %q to link %d (link already has source route %q)", routeId, linkId, lm.sourceRouteId)
		}
		lm.sourceRouteId = routeId
	} else {
		if lm.linkKind != LinkKind_Router {
			return fmt.Errorf("cannot bind route %q to link %d (link is not a router)", routeId, linkId)
		}
	}
	router.routeMap[routeId] = linkId
	return nil
}

func (router *DshRouter) bindRoute(linkId baseds.LinkId, routeId string, isSourceRoute bool) error {
	err := router.bindRouteLocally(linkId, routeId, isSourceRoute)
	if err != nil {
		return err
	}
	lm := router.getLinkMeta(linkId)
	if lm != nil {
		log.Printf("dshrouter bind route %q to %s", routeId, lm.Name())
	}
	// don't announce control routes upstream (they are local only)
	if !strings.HasPrefix(routeId, ControlPrefix) {
		router.announceUpstream(routeId)
	}
	if router.IsRootRouter() {
		router.publishRouteToBroker(routeId)
	}
	return nil
}

func (router *DshRouter) getUpstreamClient() (baseds.LinkId, AbstractRpcClient) {
	router.lock.Lock()
	defer router.lock.Unlock()
	if router.upstreamLinkId == baseds.NoLinkId {
		return baseds.NoLinkId, nil
	}
	lm := router.linkMap[router.upstreamLinkId]
	if lm == nil {
		return baseds.NoLinkId, nil
	}
	return router.upstreamLinkId, lm.client
}

func (router *DshRouter) publishRouteToBroker(routeId string) {
	defer func() {
		panichandler.PanicHandler("DshRouter:publishRouteToBroker", recover())
	}()
	dps.Broker.Publish(dps.DoraEvent{Event: dps.Event_RouteUp, Scopes: []string{routeId}})
}

func (router *DshRouter) unsubscribeFromBroker(routeId string) {
	defer func() {
		panichandler.PanicHandler("DshRouter:unregisterRoute:routedown", recover())
	}()
	dps.Broker.UnsubscribeAll(routeId)
	dps.Broker.Publish(dps.DoraEvent{Event: dps.Event_RouteDown, Scopes: []string{routeId}})
}

func sendControlUnauthenticatedErrorResponse(cmdMsg RpcMessage, linkMeta linkMeta, router *DshRouter) {
	if cmdMsg.ReqId == "" {
		return
	}
	rtnMsg := RpcMessage{
		Source: ControlRoute,
		ResId:  cmdMsg.ReqId,
		Error:  fmt.Sprintf("link is unauthenticated (%s), cannot call %q", linkMeta.Name(), cmdMsg.Command),
	}
	rtnBytes, _ := json.Marshal(rtnMsg)
	router.sendRpcMessageToLink(linkMeta.linkId, linkMeta.client, rtnBytes, baseds.NoLinkId, "unauthenticated")
}
