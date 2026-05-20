// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import * as electron from "electron";
import { getWebServerEndpoint, getWSServerEndpoint } from "../frontend/util/endpoints";

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
    // Match both HTTP and WebSocket connections. Chromium's webRequest fires for
    // WebSocket connections with the ws:// scheme, not http://, so we need
    // separate patterns to cover both.
    const httpEndpoint = getWebServerEndpoint();
    const wsEndpoint = getWSServerEndpoint();
    const filter: electron.WebRequestFilter = {
        urls: [`${httpEndpoint}/*`, `${wsEndpoint}/*`],
    };
    session.webRequest.onBeforeSendHeaders(filter, (details, callback) => {
        details.requestHeaders[RemotePasswordHeader] = injectedPassword!;
        callback({ requestHeaders: details.requestHeaders });
    });
}
