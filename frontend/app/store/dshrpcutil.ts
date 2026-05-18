// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import { setWpsRpcClient, wpsReconnectHandler } from "@/app/store/wps";
import { TabClient } from "@/app/store/tabrpcclient";
import { DshRouter } from "@/app/store/wshrouter";
import { getWSServerEndpoint } from "@/util/endpoints";
import { addWSReconnectHandler, globalWS, initGlobalWS, WSControl } from "./ws";
import { DefaultRouter, setDefaultRouter } from "./dshrpcutil-base";

let TabRpcClient: TabClient;

function initWshrpc(routeId: string): WSControl {
    const router = new DshRouter(new UpstreamDshRpcProxy());
    setDefaultRouter(router);
    const handleFn = (event: WSEventType) => {
        DefaultRouter.recvRpcMessage(event.data);
    };
    initGlobalWS(getWSServerEndpoint(), routeId, handleFn);
    globalWS.connectNow("connectWshrpc");
    TabRpcClient = new TabClient(routeId);
    setWpsRpcClient(TabRpcClient);
    DefaultRouter.registerRoute(TabRpcClient.routeId, TabRpcClient);
    addWSReconnectHandler(() => {
        DefaultRouter.reannounceRoutes();
    });
    addWSReconnectHandler(wpsReconnectHandler);
    return globalWS;
}

class UpstreamDshRpcProxy implements AbstractDshClient {
    recvRpcMessage(msg: RpcMessage): void {
        const wsMsg: WSRpcCommand = { wscommand: "rpc", message: msg };
        globalWS?.pushMessage(wsMsg);
    }
}

export { DefaultRouter, initWshrpc, TabRpcClient };
export { initElectronWshrpc, sendRpcCommand, sendRpcResponse, shutdownWshrpc } from "./dshrpcutil-base";
