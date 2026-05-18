// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dshutil

import (
	"context"
	"fmt"
	"log"

	"github.com/dfbb/doraterm/pkg/baseds"
	"github.com/dfbb/doraterm/pkg/util/shellutil"
	"github.com/dfbb/doraterm/pkg/util/utilfn"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dstore"
)

type DshRouterControlImpl struct {
	Router *DshRouter
}

func (impl *DshRouterControlImpl) DshServerImpl() {}

func (impl *DshRouterControlImpl) RouteAnnounceCommand(ctx context.Context) error {
	source := GetRpcSourceFromContext(ctx)
	if source == "" {
		return fmt.Errorf("no source in routeannounce")
	}
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return fmt.Errorf("no response handler in context")
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return fmt.Errorf("no ingress link found")
	}
	return impl.Router.bindRoute(linkId, source, false)
}

func (impl *DshRouterControlImpl) RouteUnannounceCommand(ctx context.Context) error {
	source := GetRpcSourceFromContext(ctx)
	if source == "" {
		return fmt.Errorf("no source in routeunannounce")
	}
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return fmt.Errorf("no response handler in context")
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return fmt.Errorf("no ingress link found")
	}
	return impl.Router.unbindRoute(linkId, source)
}

func (impl *DshRouterControlImpl) ControlGetRouteIdCommand(ctx context.Context) (string, error) {
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return "", nil
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return "", nil
	}
	lm := impl.Router.getLinkMeta(linkId)
	if lm == nil {
		return "", nil
	}
	return lm.sourceRouteId, nil
}

func (impl *DshRouterControlImpl) SetPeerInfoCommand(ctx context.Context, peerInfo string) error {
	source := GetRpcSourceFromContext(ctx)
	linkId := impl.Router.GetLinkIdForRoute(source)
	if linkId == baseds.NoLinkId {
		return fmt.Errorf("no link found for source route %q", source)
	}
	lm := impl.Router.getLinkMeta(linkId)
	if lm == nil {
		return fmt.Errorf("no link meta found for linkId %d", linkId)
	}
	if proxy, ok := lm.client.(*DshRpcProxy); ok {
		proxy.SetPeerInfo(peerInfo)
		return nil
	}
	return fmt.Errorf("setpeerinfo only valid for proxy connections")
}

func (impl *DshRouterControlImpl) AuthenticateCommand(ctx context.Context, data string) (dshrpc.CommandAuthenticateRtnData, error) {
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no response handler in context")
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no ingress link found")
	}

	newCtx, err := ValidateAndExtractRpcContextFromToken(data)
	if err != nil {
		log.Printf("wshrouter authenticate error linkid=%d: %v", linkId, err)
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("error validating token: %w", err)
	}
	routeId, err := validateRpcContextFromAuth(newCtx)
	if err != nil {
		return dshrpc.CommandAuthenticateRtnData{}, err
	}

	rtnData := dshrpc.CommandAuthenticateRtnData{RouteId: routeId}
	if newCtx.IsRouter {
		log.Printf("wshrouter authenticate success linkid=%d (router)", linkId)
		impl.Router.trustLink(linkId, LinkKind_Router)
	} else {
		log.Printf("wshrouter authenticate success linkid=%d routeid=%q", linkId, routeId)
		impl.Router.trustLink(linkId, LinkKind_Leaf)
		impl.Router.bindRoute(linkId, routeId, true)
	}

	return rtnData, nil
}

func extractTokenData(token string) (dshrpc.CommandAuthenticateRtnData, error) {
	entry := shellutil.GetAndRemoveTokenSwapEntry(token)
	if entry == nil {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no token entry found")
	}
	_, err := validateRpcContextFromAuth(entry.RpcContext)
	if err != nil {
		return dshrpc.CommandAuthenticateRtnData{}, err
	}
	if entry.RpcContext.IsRouter {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("cannot auth router via token")
	}
	routeId := entry.RpcContext.GenerateRouteId()
	if routeId == "" {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no routeid")
	}
	return dshrpc.CommandAuthenticateRtnData{
		RouteId:        routeId,
		Env:            entry.Env,
		InitScriptText: entry.ScriptText,
		RpcContext:     entry.RpcContext,
	}, nil
}

func (impl *DshRouterControlImpl) AuthenticateTokenVerifyCommand(ctx context.Context, data dshrpc.CommandAuthenticateTokenData) (dshrpc.CommandAuthenticateRtnData, error) {
	if !impl.Router.IsRootRouter() {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("authenticatetokenverify can only be called on root router")
	}
	if data.Token == "" {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no token in authenticatetoken message")
	}

	rtnData, err := extractTokenData(data.Token)
	if err != nil {
		log.Printf("wshrouter authenticate-token-verify error: %v", err)
		return dshrpc.CommandAuthenticateRtnData{}, err
	}

	log.Printf("wshrouter authenticate-token-verify success routeid=%q", rtnData.RouteId)
	return rtnData, nil
}

func (impl *DshRouterControlImpl) AuthenticateTokenCommand(ctx context.Context, data dshrpc.CommandAuthenticateTokenData) (dshrpc.CommandAuthenticateRtnData, error) {
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no response handler in context")
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no ingress link found")
	}

	if data.Token == "" {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no token in authenticatetoken message")
	}

	var rtnData dshrpc.CommandAuthenticateRtnData
	var err error

	if impl.Router.IsRootRouter() {
		rtnData, err = extractTokenData(data.Token)
		if err != nil {
			log.Printf("wshrouter authenticate-token error linkid=%d: %v", linkId, err)
			return dshrpc.CommandAuthenticateRtnData{}, err
		}
	} else {
		wshRpc := GetDshRpcFromContext(ctx)
		if wshRpc == nil {
			return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no wshrpc in context")
		}
		respData, err := wshRpc.SendRpcRequest(dshrpc.Command_AuthenticateTokenVerify, data, &dshrpc.RpcOpts{Route: ControlRootRoute})
		if err != nil {
			log.Printf("wshrouter authenticate-token error linkid=%d: failed to verify token: %v", linkId, err)
			return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("failed to verify token: %w", err)
		}
		err = utilfn.ReUnmarshal(&rtnData, respData)
		if err != nil {
			return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	if rtnData.RpcContext == nil {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no rpccontext in token response")
	}
	if rtnData.RouteId == "" {
		return dshrpc.CommandAuthenticateRtnData{}, fmt.Errorf("no routeid in token response")
	}
	log.Printf("wshrouter authenticate-token success linkid=%d routeid=%q", linkId, rtnData.RouteId)
	impl.Router.trustLink(linkId, LinkKind_Leaf)
	impl.Router.bindRoute(linkId, rtnData.RouteId, true)

	return rtnData, nil
}

func (impl *DshRouterControlImpl) AuthenticateJobManagerVerifyCommand(ctx context.Context, data dshrpc.CommandAuthenticateJobManagerData) error {
	if !impl.Router.IsRootRouter() {
		return fmt.Errorf("authenticatejobmanagerverify can only be called on root router")
	}

	if data.JobId == "" {
		return fmt.Errorf("no jobid in authenticatejobmanager message")
	}
	if data.JobAuthToken == "" {
		return fmt.Errorf("no jobauthtoken in authenticatejobmanager message")
	}

	job, err := dstore.DBMustGet[*doraobj.Job](ctx, data.JobId)
	if err != nil {
		log.Printf("wshrouter authenticate-jobmanager-verify error jobid=%q: failed to get job: %v", data.JobId, err)
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job.JobAuthToken != data.JobAuthToken {
		log.Printf("wshrouter authenticate-jobmanager-verify error jobid=%q: invalid jobauthtoken", data.JobId)
		return fmt.Errorf("invalid jobauthtoken")
	}

	log.Printf("wshrouter authenticate-jobmanager-verify success jobid=%q", data.JobId)
	return nil
}

func (impl *DshRouterControlImpl) AuthenticateJobManagerCommand(ctx context.Context, data dshrpc.CommandAuthenticateJobManagerData) error {
	handler := GetRpcResponseHandlerFromContext(ctx)
	if handler == nil {
		return fmt.Errorf("no response handler in context")
	}
	linkId := handler.GetIngressLinkId()
	if linkId == baseds.NoLinkId {
		return fmt.Errorf("no ingress link found")
	}

	if data.JobId == "" {
		return fmt.Errorf("no jobid in authenticatejobmanager message")
	}
	if data.JobAuthToken == "" {
		return fmt.Errorf("no jobauthtoken in authenticatejobmanager message")
	}

	if impl.Router.IsRootRouter() {
		job, err := dstore.DBMustGet[*doraobj.Job](ctx, data.JobId)
		if err != nil {
			log.Printf("wshrouter authenticate-jobmanager error linkid=%d jobid=%q: failed to get job: %v", linkId, data.JobId, err)
			return fmt.Errorf("failed to get job: %w", err)
		}

		if job.JobAuthToken != data.JobAuthToken {
			log.Printf("wshrouter authenticate-jobmanager error linkid=%d jobid=%q: invalid jobauthtoken", linkId, data.JobId)
			return fmt.Errorf("invalid jobauthtoken")
		}
	} else {
		wshRpc := GetDshRpcFromContext(ctx)
		if wshRpc == nil {
			return fmt.Errorf("no wshrpc in context")
		}
		_, err := wshRpc.SendRpcRequest(dshrpc.Command_AuthenticateJobManagerVerify, data, &dshrpc.RpcOpts{Route: ControlRootRoute})
		if err != nil {
			log.Printf("wshrouter authenticate-jobmanager error linkid=%d jobid=%q: failed to verify job auth token: %v", linkId, data.JobId, err)
			return fmt.Errorf("failed to verify job auth token: %w", err)
		}
	}

	routeId := MakeJobRouteId(data.JobId)
	log.Printf("wshrouter authenticate-jobmanager success linkid=%d jobid=%q routeid=%q", linkId, data.JobId, routeId)
	impl.Router.trustLink(linkId, LinkKind_Leaf)
	impl.Router.bindRoute(linkId, routeId, true)

	return nil
}

func validateRpcContextFromAuth(newCtx *dshrpc.RpcContext) (string, error) {
	if newCtx == nil {
		return "", fmt.Errorf("no context found in jwt token")
	}
	if newCtx.IsRouter && newCtx.RouteId != "" {
		return "", fmt.Errorf("invalid context, router cannot have a routeid")
	}
	if newCtx.IsRouter && newCtx.ProcRoute {
		return "", fmt.Errorf("invalid context, router cannot have a proc-route")
	}
	if !newCtx.IsRouter && newCtx.RouteId == "" && !newCtx.ProcRoute {
		return "", fmt.Errorf("invalid context, must have a routeid")
	}
	if newCtx.IsRouter {
		return "", nil
	}
	return newCtx.GenerateRouteId(), nil
}
