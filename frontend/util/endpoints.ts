// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import { isPreviewWindow } from "@/app/store/windowtype";
import { getEnv } from "./getenv";
import { lazy } from "./util";

export const WebServerEndpointVarName = "DORA_SERVER_WEB_ENDPOINT";
export const WSServerEndpointVarName = "DORA_SERVER_WS_ENDPOINT";

export const getWebServerEndpoint = lazy(() => {
    if (isPreviewWindow()) return null;
    const v = getEnv(WebServerEndpointVarName);
    if (v.startsWith("http://") || v.startsWith("https://")) return v;
    return `http://${v}`;
});

export const getWSServerEndpoint = lazy(() => {
    if (isPreviewWindow()) return null;
    const v = getEnv(WSServerEndpointVarName);
    if (v.startsWith("https://")) return `wss://${v.slice("https://".length)}`;
    if (v.startsWith("http://")) return `ws://${v.slice("http://".length)}`;
    return `ws://${v}`;
});
