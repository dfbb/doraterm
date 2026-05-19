// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import { RpcApi } from "@/app/store/dshclientapi";
import { adaptFromElectronKeyEvent, checkKeyPressed } from "@/util/keyutil";
import { CHORD_TIMEOUT } from "@/util/sharedconst";
import { Rectangle, shell, WebContentsView } from "electron";
import { createNewDoraWindow, getDoraWindowById } from "emain/emain-window";
import path from "path";
import { configureAuthKeyRequestInjection } from "./authkey";
import { setWasActive } from "./emain-activity";
import { getElectronAppBasePath, getRemoteState, isDevVite, unamePlatform } from "./emain-platform";
import { configureRemotePasswordInjection } from "./remoteauth";
import {
    decreaseZoomLevel,
    handleCtrlShiftFocus,
    handleCtrlShiftState,
    increaseZoomLevel,
    resetZoomLevel,
    shFrameNavHandler,
    shNavHandler,
} from "./emain-util";
import { ElectronDshClient } from "./emain-dsh";

function handleWindowsMenuAccelerators(
    doraEvent: DoraKeyboardEvent,
    tabView: DoraTabView,
    fullConfig: FullConfigType
): boolean {
    const doraWindow = getDoraWindowById(tabView.doraWindowId);

    if (checkKeyPressed(doraEvent, "Ctrl:Shift:n")) {
        createNewDoraWindow();
        return true;
    }

    if (checkKeyPressed(doraEvent, "Ctrl:Shift:r")) {
        tabView.webContents.reloadIgnoringCache();
        return true;
    }

    if (checkKeyPressed(doraEvent, "Ctrl:v")) {
        const ctrlVPaste = fullConfig?.settings?.["app:ctrlvpaste"];
        const shouldPaste = ctrlVPaste ?? true;
        if (!shouldPaste) {
            return false;
        }
        tabView.webContents.paste();
        return true;
    }

    if (checkKeyPressed(doraEvent, "Ctrl:0")) {
        resetZoomLevel(tabView.webContents);
        return true;
    }

    if (checkKeyPressed(doraEvent, "Ctrl:=") || checkKeyPressed(doraEvent, "Ctrl:Shift:=")) {
        increaseZoomLevel(tabView.webContents);
        return true;
    }

    if (checkKeyPressed(doraEvent, "Ctrl:-") || checkKeyPressed(doraEvent, "Ctrl:Shift:-")) {
        decreaseZoomLevel(tabView.webContents);
        return true;
    }

    if (checkKeyPressed(doraEvent, "F11")) {
        if (doraWindow) {
            doraWindow.setFullScreen(!doraWindow.isFullScreen());
        }
        return true;
    }

    for (let i = 1; i <= 9; i++) {
        if (checkKeyPressed(doraEvent, `Alt:Ctrl:${i}`)) {
            const workspaceNum = i - 1;
            RpcApi.WorkspaceListCommand(ElectronDshClient).then((workspaceList) => {
                if (workspaceList && workspaceNum < workspaceList.length) {
                    const workspace = workspaceList[workspaceNum];
                    if (doraWindow) {
                        doraWindow.switchWorkspace(workspace.workspacedata.oid);
                    }
                }
            });
            return true;
        }
    }

    if (checkKeyPressed(doraEvent, "Alt:Shift:i")) {
        tabView.webContents.toggleDevTools();
        return true;
    }

    return false;
}

function computeBgColor(fullConfig: FullConfigType): string {
    const settings = fullConfig?.settings;
    const isTransparent = settings?.["window:transparent"] ?? false;
    const isBlur = !isTransparent && (settings?.["window:blur"] ?? false);
    if (isTransparent) {
        return "#00000000";
    } else if (isBlur) {
        return "#00000000";
    } else {
        return "#222222";
    }
}

const wcIdToDoraTabMap = new Map<number, DoraTabView>();

export function getDoraTabViewByWebContentsId(webContentsId: number): DoraTabView {
    if (webContentsId == null) {
        return null;
    }
    return wcIdToDoraTabMap.get(webContentsId);
}

export class DoraTabView extends WebContentsView {
    doraWindowId: string; // this will be set for any tabviews that are initialized. (unset for the hot spare)
    isActiveTab: boolean;
    private _doraTabId: string; // always set, DoraTabViews are unique per tab
    lastUsedTs: number; // ts milliseconds
    createdTs: number; // ts milliseconds
    initPromise: Promise<void>;
    initResolve: () => void;
    savedInitOpts: DoraInitOpts;
    doraReadyPromise: Promise<void>;
    doraReadyResolve: () => void;
    isInitialized: boolean = false;
    isDoraReady: boolean = false;
    isDestroyed: boolean = false;
    keyboardChordMode: boolean = false;
    resetChordModeTimeout: NodeJS.Timeout = null;

    constructor(fullConfig: FullConfigType) {
        console.log("createBareTabView");
        super({
            webPreferences: {
                preload: path.join(getElectronAppBasePath(), "preload", "index.cjs"),
                webviewTag: true,
            },
        });
        this.createdTs = Date.now();
        this.savedInitOpts = null;
        this.initPromise = new Promise((resolve, _) => {
            this.initResolve = resolve;
        });
        this.initPromise.then(() => {
            this.isInitialized = true;
            console.log("tabview init", Date.now() - this.createdTs + "ms");
        });
        this.doraReadyPromise = new Promise((resolve, _) => {
            this.doraReadyResolve = resolve;
        });
        this.doraReadyPromise.then(() => {
            this.isDoraReady = true;
        });
        const wcId = this.webContents.id;
        wcIdToDoraTabMap.set(wcId, this);
        if (isDevVite) {
            this.webContents.loadURL(`${process.env.ELECTRON_RENDERER_URL}/index.html`);
        } else {
            this.webContents.loadFile(path.join(getElectronAppBasePath(), "frontend", "index.html"));
        }
        this.webContents.on("destroyed", () => {
            wcIdToDoraTabMap.delete(wcId);
            removeDoraTabView(this.doraTabId);
            this.isDestroyed = true;
        });
        this.setBackgroundColor(computeBgColor(fullConfig));
    }

    get doraTabId(): string {
        return this._doraTabId;
    }

    set doraTabId(doraTabId: string) {
        this._doraTabId = doraTabId;
    }

    setKeyboardChordMode(mode: boolean) {
        this.keyboardChordMode = mode;
        if (mode) {
            if (this.resetChordModeTimeout) {
                clearTimeout(this.resetChordModeTimeout);
            }
            this.resetChordModeTimeout = setTimeout(() => {
                this.keyboardChordMode = false;
            }, CHORD_TIMEOUT);
        } else {
            if (this.resetChordModeTimeout) {
                clearTimeout(this.resetChordModeTimeout);
                this.resetChordModeTimeout = null;
            }
        }
    }

    positionTabOnScreen(winBounds: Rectangle) {
        const curBounds = this.getBounds();
        if (
            curBounds.width == winBounds.width &&
            curBounds.height == winBounds.height &&
            curBounds.x == 0 &&
            curBounds.y == 0
        ) {
            return;
        }
        this.setBounds({ x: 0, y: 0, width: winBounds.width, height: winBounds.height });
    }

    positionTabOffScreen(winBounds: Rectangle) {
        this.setBounds({
            x: -15000,
            y: -15000,
            width: winBounds.width,
            height: winBounds.height,
        });
    }

    isOnScreen() {
        const bounds = this.getBounds();
        return bounds.x == 0 && bounds.y == 0;
    }

    destroy() {
        console.log("destroy tab", this.doraTabId);
        removeDoraTabView(this.doraTabId);
        if (!this.isDestroyed) {
            this.webContents?.close();
        }
        this.isDestroyed = true;
    }
}

let MaxCacheSize = 10;
const wcvCache = new Map<string, DoraTabView>();

export function setMaxTabCacheSize(size: number) {
    console.log("setMaxTabCacheSize", size);
    MaxCacheSize = size;
}

export function getDoraTabView(doraTabId: string): DoraTabView | undefined {
    const rtn = wcvCache.get(doraTabId);
    if (rtn) {
        rtn.lastUsedTs = Date.now();
    }
    return rtn;
}

function tryEvictEntry(doraTabId: string): boolean {
    const tabView = wcvCache.get(doraTabId);
    if (!tabView) {
        return false;
    }
    if (tabView.isActiveTab) {
        return false;
    }
    const lastUsedDiff = Date.now() - tabView.lastUsedTs;
    if (lastUsedDiff < 1000) {
        return false;
    }
    const ww = getDoraWindowById(tabView.doraWindowId);
    if (!ww) {
        // this shouldn't happen, but if it does, just destroy the tabview
        console.log("[error] DoraWindow not found for DoraTabView", tabView.doraTabId);
        tabView.destroy();
        return true;
    } else {
        // will trigger a destroy on the tabview
        ww.removeTabView(tabView.doraTabId, false);
        return true;
    }
}

function checkAndEvictCache(): void {
    if (wcvCache.size <= MaxCacheSize) {
        return;
    }
    const sorted = Array.from(wcvCache.values()).sort((a, b) => {
        // Prioritize entries which are active
        if (a.isActiveTab && !b.isActiveTab) {
            return -1;
        }
        // Otherwise, sort by lastUsedTs
        return a.lastUsedTs - b.lastUsedTs;
    });
    for (let i = 0; i < sorted.length - MaxCacheSize; i++) {
        tryEvictEntry(sorted[i].doraTabId);
    }
}

export function clearTabCache() {
    const wcVals = Array.from(wcvCache.values());
    for (let i = 0; i < wcVals.length; i++) {
        const tabView = wcVals[i];
        tryEvictEntry(tabView.doraTabId);
    }
}

// returns [tabview, initialized]
export async function getOrCreateWebViewForTab(doraWindowId: string, tabId: string): Promise<[DoraTabView, boolean]> {
    let tabView = getDoraTabView(tabId);
    if (tabView) {
        return [tabView, true];
    }
    const fullConfig = await RpcApi.GetFullConfigCommand(ElectronDshClient);
    tabView = getSpareTab(fullConfig);
    tabView.doraWindowId = doraWindowId;
    tabView.lastUsedTs = Date.now();
    setDoraTabView(tabId, tabView);
    tabView.doraTabId = tabId;
    tabView.webContents.on("will-navigate", shNavHandler);
    tabView.webContents.on("will-frame-navigate", shFrameNavHandler);
    tabView.webContents.on("did-attach-webview", (event, wc) => {
        wc.setWindowOpenHandler((details) => {
            if (wc == null || wc.isDestroyed() || tabView.webContents == null || tabView.webContents.isDestroyed()) {
                return { action: "deny" };
            }
            tabView.webContents.send("webview-new-window", wc.id, details);
            return { action: "deny" };
        });
    });
    tabView.webContents.on("before-input-event", (e, input) => {
        const doraEvent = adaptFromElectronKeyEvent(input);
        // console.log("WIN bie", tabView.doraTabId.substring(0, 8), doraEvent.type, doraEvent.code);
        handleCtrlShiftState(tabView.webContents, doraEvent);
        setWasActive(true);
        if (input.type == "keyDown" && tabView.keyboardChordMode) {
            e.preventDefault();
            tabView.setKeyboardChordMode(false);
            tabView.webContents.send("reinject-key", doraEvent);
            return;
        }

        if (unamePlatform === "win32" && input.type == "keyDown") {
            if (handleWindowsMenuAccelerators(doraEvent, tabView, fullConfig)) {
                e.preventDefault();
                return;
            }
        }
    });
    tabView.webContents.setWindowOpenHandler(({ url, frameName }) => {
        if (url.startsWith("http://") || url.startsWith("https://") || url.startsWith("file://")) {
            console.log("openExternal fallback", url);
            shell.openExternal(url);
        }
        console.log("window-open denied", url);
        return { action: "deny" };
    });
    tabView.webContents.on("blur", () => {
        handleCtrlShiftFocus(tabView.webContents, false);
    });
    const remote = getRemoteState();
    if (remote.isRemote) {
        configureRemotePasswordInjection(tabView.webContents.session);
    } else {
        configureAuthKeyRequestInjection(tabView.webContents.session);
    }
    return [tabView, false];
}

export function setDoraTabView(doraTabId: string, wcv: DoraTabView): void {
    if (doraTabId == null) {
        return;
    }
    wcvCache.set(doraTabId, wcv);
    checkAndEvictCache();
}

function removeDoraTabView(doraTabId: string): void {
    if (doraTabId == null) {
        return;
    }
    wcvCache.delete(doraTabId);
}

let HotSpareTab: DoraTabView = null;

export function ensureHotSpareTab(fullConfig: FullConfigType) {
    console.log("ensureHotSpareTab");
    if (HotSpareTab == null) {
        HotSpareTab = new DoraTabView(fullConfig);
    }
}

export function getSpareTab(fullConfig: FullConfigType): DoraTabView {
    setTimeout(() => ensureHotSpareTab(fullConfig), 500);
    if (HotSpareTab != null) {
        const rtn = HotSpareTab;
        HotSpareTab = null;
        console.log("getSpareTab: returning hotspare");
        return rtn;
    } else {
        console.log("getSpareTab: creating new tab");
        return new DoraTabView(fullConfig);
    }
}
