// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import * as electron from "electron";
import * as child_process from "node:child_process";
import * as readline from "readline";
import { WebServerEndpointVarName, WSServerEndpointVarName } from "../frontend/util/endpoints";
import { AuthKey, DoraAuthKeyEnv } from "./authkey";
import { setForceQuit, setUserConfirmedQuit } from "./emain-activity";
import {
    getElectronAppResourcesPath,
    getElectronAppUnpackedBasePath,
    getDoraConfigDir,
    getDoraDataDir,
    getDoraSrvCwd,
    getDoraSrvPath,
    getXdgCurrentDesktop,
    DoraConfigHomeVarName,
    DoraDataHomeVarName,
} from "./emain-platform";
import {
    getElectronExecPath,
    DoraAppElectronExecPath,
    DoraAppPathVarName,
    DoraAppResourcesPathVarName,
} from "./emain-util";
import { updater } from "./updater";

let isDoraSrvDead = false;
let waveSrvProc: child_process.ChildProcessWithoutNullStreams | null = null;
let DoraVersion = "unknown"; // set by DORASRV-ESTART
let DoraBuildTime = 0; // set by DORASRV-ESTART

export function getDoraVersion(): { version: string; buildTime: number } {
    return { version: DoraVersion, buildTime: DoraBuildTime };
}

let waveSrvReadyResolve = (value: boolean) => {};
const waveSrvReady: Promise<boolean> = new Promise((resolve, _) => {
    waveSrvReadyResolve = resolve;
});

export function getDoraSrvReady(): Promise<boolean> {
    return waveSrvReady;
}

export function getDoraSrvProc(): child_process.ChildProcessWithoutNullStreams | null {
    return waveSrvProc;
}

export function getIsDoraSrvDead(): boolean {
    return isDoraSrvDead;
}

export function runDoraSrv(handleWSEvent: (evtMsg: WSEventType) => void): Promise<boolean> {
    let pResolve: (value: boolean) => void;
    let pReject: (reason?: any) => void;
    const rtnPromise = new Promise<boolean>((argResolve, argReject) => {
        pResolve = argResolve;
        pReject = argReject;
    });
    const envCopy = { ...process.env };
    const xdgCurrentDesktop = getXdgCurrentDesktop();
    if (xdgCurrentDesktop != null) {
        envCopy["XDG_CURRENT_DESKTOP"] = xdgCurrentDesktop;
    }
    envCopy[DoraAppPathVarName] = getElectronAppUnpackedBasePath();
    envCopy[DoraAppResourcesPathVarName] = getElectronAppResourcesPath();
    envCopy[DoraAppElectronExecPath] = getElectronExecPath();
    envCopy[DoraAuthKeyEnv] = AuthKey;
    envCopy[DoraDataHomeVarName] = getDoraDataDir();
    envCopy[DoraConfigHomeVarName] = getDoraConfigDir();
    const waveSrvCmd = getDoraSrvPath();
    console.log("trying to run local server", waveSrvCmd);
    const proc = child_process.spawn(getDoraSrvPath(), {
        cwd: getDoraSrvCwd(),
        env: envCopy,
    });
    proc.on("exit", (e) => {
        if (updater?.status == "installing") {
            return;
        }
        console.log("dorasrv exited, shutting down");
        setForceQuit(true);
        isDoraSrvDead = true;
        electron.app.quit();
    });
    proc.on("spawn", (e) => {
        console.log("spawned dorasrv");
        waveSrvProc = proc;
        pResolve(true);
    });
    proc.on("error", (e) => {
        console.log("error running dorasrv", e);
        pReject(e);
    });
    const rlStdout = readline.createInterface({
        input: proc.stdout,
        terminal: false,
    });
    rlStdout.on("line", (line) => {
        console.log(line);
    });
    const rlStderr = readline.createInterface({
        input: proc.stderr,
        terminal: false,
    });
    rlStderr.on("line", (line) => {
        if (line.includes("DORASRV-ESTART")) {
            const startParams = /ws:([a-z0-9.:]+) web:([a-z0-9.:]+) version:([a-z0-9.-]+) buildtime:(\d+)/gm.exec(
                line
            );
            if (startParams == null) {
                console.log("error parsing DORASRV-ESTART line", line);
                setUserConfirmedQuit(true);
                electron.app.quit();
                return;
            }
            process.env[WSServerEndpointVarName] = startParams[1];
            process.env[WebServerEndpointVarName] = startParams[2];
            DoraVersion = startParams[3];
            DoraBuildTime = parseInt(startParams[4]);
            waveSrvReadyResolve(true);
            return;
        }
        if (line.startsWith("DORASRV-EVENT:")) {
            const evtJson = line.slice("DORASRV-EVENT:".length);
            try {
                const evtMsg: WSEventType = JSON.parse(evtJson);
                handleWSEvent(evtMsg);
            } catch (e) {
                console.log("error handling DORASRV-EVENT", e);
            }
            return;
        }
        console.log(line);
    });
    return rtnPromise;
}
