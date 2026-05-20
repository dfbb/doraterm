// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import * as electron from "electron";
import { getWebServerEndpoint } from "../frontend/util/endpoints";

export const RemotePasswordHeader = "X-Remote-Password";

let injectedPassword: string | null = null;

export function setRemotePassword(p: string): void {
    injectedPassword = p;
}

export function getRemotePassword(): string | null {
    return injectedPassword;
}

export function configureRemotePasswordInjection(session: electron.Session): void {
    if (!injectedPassword) return;
    // Chromium's webRequest API only supports http/https schemes in URL patterns.
    // WebSocket upgrade requests start as HTTP requests, so the http(s) pattern
    // catches them before the upgrade to ws/wss completes.
    const endpoint = getWebServerEndpoint();
    const filter: electron.WebRequestFilter = {
        urls: [`${endpoint}/*`],
    };
    session.webRequest.onBeforeSendHeaders(filter, (details, callback) => {
        details.requestHeaders[RemotePasswordHeader] = injectedPassword!;
        callback({ requestHeaders: details.requestHeaders });
    });
}
