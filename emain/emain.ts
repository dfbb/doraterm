// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import { RpcApi } from "@/app/store/dshclientapi";
import { waveEventSubscribeSingle } from "@/app/store/wps";
import * as electron from "electron";
import { globalEvents } from "emain/emain-events";
import { sprintf } from "sprintf-js";
import * as services from "../frontend/app/store/services";
import { initElectronWshrpc, shutdownWshrpc } from "../frontend/app/store/dshrpcutil-base";
import { fireAndForget, sleep } from "../frontend/util/util";
import { AuthKey, configureAuthKeyRequestInjection } from "./authkey";
import { configureRemotePasswordInjection, setRemotePassword } from "./remoteauth";
import {
    getActivityState,
    getAndClearTermCommandsDurable,
    getAndClearTermCommandsRun,
    getForceQuit,
    getGlobalIsRelaunching,
    getUserConfirmedQuit,
    setForceQuit,
    setGlobalIsQuitting,
    setGlobalIsStarting,
    setUserConfirmedQuit,
    setWasActive,
    setWasInFg,
} from "./emain-activity";
import { initIpcHandlers } from "./emain-ipc";
import { log } from "./emain-log";
import { initMenuEventSubscriptions, makeAndSetAppMenu, makeDockTaskbar } from "./emain-menu";
import {
    checkIfRunningUnderARM64Translation,
    getElectronAppBasePath,
    getElectronAppUnpackedBasePath,
    getDoraConfigDir,
    getDoraDataDir,
    getRemoteState,
    isDev,
    unameArch,
    unamePlatform,
} from "./emain-platform";
import { ensureHotSpareTab, setMaxTabCacheSize } from "./emain-tabview";
import { getIsDoraSrvDead, getDoraSrvProc, getDoraSrvReady, runDoraSrv } from "./emain-dorasrv";
import {
    createBrowserWindow,
    createNewDoraWindow,
    focusedDoraWindow,
    getAllDoraWindows,
    getQuakeWindow,
    getDoraWindowById,
    getDoraWindowByWorkspaceId,
    initGlobalHotkeyEventSubscription,
    registerGlobalHotkey,
    relaunchBrowserWindows,
    DoraBrowserWindow,
} from "./emain-window";
import { ElectronDshClient, initElectronDshClient } from "./emain-dsh";
import { getLaunchSettings } from "./launchsettings";
import { configureAutoUpdater, updater } from "./updater";

const electronApp = electron.app;

let confirmQuit = true;

const waveDataDir = getDoraDataDir();
const doraConfigDir = getDoraConfigDir();

electron.nativeTheme.themeSource = "dark";

console.log = log;
console.log(
    sprintf(
        "doraterm-app starting, data_dir=%s, config_dir=%s electronpath=%s gopath=%s arch=%s/%s electron=%s",
        waveDataDir,
        doraConfigDir,
        getElectronAppBasePath(),
        getElectronAppUnpackedBasePath(),
        unamePlatform,
        unameArch,
        process.versions.electron
    )
);
if (isDev) {
    console.log("doraterm-app DORATERM_DEV set");
}

function handleWSEvent(evtMsg: WSEventType) {
    fireAndForget(async () => {
        console.log("handleWSEvent", evtMsg?.eventtype);
        if (evtMsg.eventtype == "electron:newwindow") {
            console.log("electron:newwindow", evtMsg.data);
            const windowId: string = evtMsg.data;
            const windowData: DoraWindow = (await services.ObjectService.GetObject("window:" + windowId)) as DoraWindow;
            if (windowData == null) {
                return;
            }
            const fullConfig = await RpcApi.GetFullConfigCommand(ElectronDshClient);
            const newWin = await createBrowserWindow(windowData, fullConfig, {
                unamePlatform,
                isPrimaryStartupWindow: false,
            });
            newWin.show();
        } else if (evtMsg.eventtype == "electron:closewindow") {
            console.log("electron:closewindow", evtMsg.data);
            if (evtMsg.data === undefined) return;
            const ww = getDoraWindowById(evtMsg.data);
            if (ww != null) {
                ww.destroy(); // bypass the "are you sure?" dialog
            }
        } else if (evtMsg.eventtype == "electron:updateactivetab") {
            const activeTabUpdate: { workspaceid: string; newactivetabid: string } = evtMsg.data;
            console.log("electron:updateactivetab", activeTabUpdate);
            const ww = getDoraWindowByWorkspaceId(activeTabUpdate.workspaceid);
            if (ww == null) {
                return;
            }
            await ww.setActiveTab(activeTabUpdate.newactivetabid, false);
        } else {
            console.log("unhandled electron ws eventtype", evtMsg.eventtype);
        }
    });
}

async function initElectronControlEventSubscription(): Promise<void> {
    waveEventSubscribeSingle({
        eventType: "electron:control",
        handler: (event) => {
            const evtMsg = event?.data as WSEventType;
            if (evtMsg == null) return;
            handleWSEvent(evtMsg);
        },
    });
    try {
        await RpcApi.EventSubCommand(ElectronDshClient, {
            event: "electron:control",
            scopes: [],
            allscopes: true,
        });
    } catch (e) {
        console.log("error acknowledging electron:control subscription", e);
    }
}

async function alignAllWindowsActiveTab() {
    for (const ww of getAllDoraWindows()) {
        try {
            const workspace = await services.WorkspaceService.GetWorkspace(ww.workspaceId);
            if (workspace?.activetabid == null) continue;
            if (ww.activeTabView?.waveTabId == workspace.activetabid) continue;
            await ww.setActiveTab(workspace.activetabid, false);
        } catch (e) {
            console.log("error aligning window", ww.waveWindowId, e);
        }
    }
}

function hideWindowWithCatch(window: DoraBrowserWindow) {
    if (window == null) {
        return;
    }
    try {
        if (window.isDestroyed()) {
            return;
        }
        window.hide();
    } catch (e) {
        console.log("error hiding window", e);
    }
}

electronApp.on("window-all-closed", () => {
    if (getGlobalIsRelaunching()) {
        return;
    }
    if (unamePlatform !== "darwin") {
        setUserConfirmedQuit(true);
        electronApp.quit();
    }
});
electronApp.on("before-quit", (e) => {
    const allWindows = getAllDoraWindows();
    if (
        confirmQuit &&
        !getForceQuit() &&
        !getUserConfirmedQuit() &&
        allWindows.length > 0 &&
        !getIsDoraSrvDead() &&
        !process.env.DORATERM_NOCONFIRMQUIT
    ) {
        e.preventDefault();
        const choice = electron.dialog.showMessageBoxSync(null, {
            type: "question",
            buttons: ["Cancel", "Quit"],
            title: "Confirm Quit",
            message: "Are you sure you want to quit Wave Terminal?",
            defaultId: 0,
            cancelId: 0,
        });
        if (choice === 0) {
            return;
        }
        setUserConfirmedQuit(true);
        electronApp.quit();
        return;
    }
    setGlobalIsQuitting(true);
    updater?.stop();
    if (unamePlatform == "win32") {
        // win32 doesn't have a SIGINT, so we just let electron die, which
        // ends up killing dorasrv via closing it's stdin.
        return;
    }
    getDoraSrvProc()?.kill("SIGINT");
    shutdownWshrpc();
    if (getForceQuit()) {
        return;
    }
    e.preventDefault();
    for (const window of allWindows) {
        hideWindowWithCatch(window);
    }
    if (getIsDoraSrvDead()) {
        console.log("dorasrv is dead, quitting immediately");
        setForceQuit(true);
        electronApp.quit();
        return;
    }
    setTimeout(() => {
        console.log("waiting for dorasrv to exit...");
        setForceQuit(true);
        electronApp.quit();
    }, 3000);
});
process.on("SIGINT", () => {
    console.log("Caught SIGINT, shutting down");
    setUserConfirmedQuit(true);
    electronApp.quit();
});
process.on("SIGHUP", () => {
    console.log("Caught SIGHUP, shutting down");
    setUserConfirmedQuit(true);
    electronApp.quit();
});
process.on("SIGTERM", () => {
    console.log("Caught SIGTERM, shutting down");
    setUserConfirmedQuit(true);
    electronApp.quit();
});
let caughtException = false;
process.on("uncaughtException", (error) => {
    if (caughtException) {
        return;
    }

    // Check if the error is related to QUIC protocol, if so, ignore (can happen with the updater)
    if (error?.message?.includes("net::ERR_QUIC_PROTOCOL_ERROR")) {
        console.log("Ignoring QUIC protocol error:", error.message);
        console.log("Stack Trace:", error.stack);
        return;
    }

    caughtException = true;
    console.log("Uncaught Exception, shutting down: ", error);
    console.log("Stack Trace:", error.stack);
    // Optionally, handle cleanup or exit the app
    setUserConfirmedQuit(true);
    electronApp.quit();
});

let lastDoraWindowCount = 0;
globalEvents.on("windows-updated", () => {
    const wwCount = getAllDoraWindows().length;
    if (wwCount == lastDoraWindowCount) {
        return;
    }
    lastDoraWindowCount = wwCount;
    console.log("windows-updated", wwCount);
    makeAndSetAppMenu();
});

async function appMain() {
    // Set disableHardwareAcceleration as early as possible, if required.
    const launchSettings = getLaunchSettings();
    if (launchSettings?.["window:disablehardwareacceleration"]) {
        console.log("disabling hardware acceleration, per launch settings");
        electronApp.disableHardwareAcceleration();
    }
    const startTs = Date.now();
    const instanceLock = electronApp.requestSingleInstanceLock();
    if (!instanceLock) {
        console.log("doraterm-app could not get single-instance-lock, shutting down");
        setUserConfirmedQuit(true);
        electronApp.quit();
        return;
    }
    electronApp.on("second-instance", (_event, argv, workingDirectory) => {
        console.log("second-instance event, argv:", argv, "workingDirectory:", workingDirectory);
        fireAndForget(createNewDoraWindow);
    });
    try {
        await runDoraSrv(handleWSEvent);
    } catch (e) {
        console.log(e.toString());
    }
    const ready = await getDoraSrvReady();
    console.log("dorasrv ready signal received", ready, Date.now() - startTs, "ms");
    await electronApp.whenReady();
    const remote = getRemoteState();
    if (remote.isRemote) {
        setRemotePassword(remote.password!);
        configureRemotePasswordInjection(electron.session.defaultSession);
    } else {
        configureAuthKeyRequestInjection(electron.session.defaultSession);
    }
    initIpcHandlers();

    await sleep(10); // wait a bit for dorasrv to be ready
    try {
        initElectronDshClient();
        const dshOpts = remote.isRemote ? { remotePassword: remote.password! } : { authKey: AuthKey };
        initElectronWshrpc(ElectronDshClient, dshOpts);
        initMenuEventSubscriptions();
        await initElectronControlEventSubscription();
    } catch (e) {
        console.log("error initializing wshrpc", e);
    }
    const fullConfig = await RpcApi.GetFullConfigCommand(ElectronDshClient);
    checkIfRunningUnderARM64Translation(fullConfig);
    if (fullConfig?.settings?.["app:confirmquit"] != null) {
        confirmQuit = fullConfig.settings["app:confirmquit"];
    }
    ensureHotSpareTab(fullConfig);
    await relaunchBrowserWindows();
    await alignAllWindowsActiveTab();

    makeAndSetAppMenu();
    if (!remote.isRemote) {
        makeDockTaskbar();
        await configureAutoUpdater();
    }
    setGlobalIsStarting(false);
    if (fullConfig?.settings?.["window:maxtabcachesize"] != null) {
        setMaxTabCacheSize(fullConfig.settings["window:maxtabcachesize"]);
    }

    electronApp.on("activate", () => {
        const allWindows = getAllDoraWindows();
        const anyVisible = allWindows.some((w) => !w.isDestroyed() && w.isVisible());
        if (anyVisible) {
            return;
        }
        const qw = getQuakeWindow();
        if (qw != null && !qw.isDestroyed()) {
            qw.show();
            qw.focus();
            return;
        }
        if (allWindows.length === 0) {
            fireAndForget(createNewDoraWindow);
        }
    });
    if (!remote.isRemote) {
        const rawGlobalHotKey = launchSettings?.["app:globalhotkey"];
        if (rawGlobalHotKey) {
            registerGlobalHotkey(rawGlobalHotKey);
        }
        initGlobalHotkeyEventSubscription();
    }
}

appMain().catch((e) => {
    console.log("appMain error", e);
    setUserConfirmedQuit(true);
    electronApp.quit();
});
