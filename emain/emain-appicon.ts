// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import * as electron from "electron";
import * as path from "node:path";
import { getElectronAppBasePath, isDev, unamePlatform } from "./emain-platform";

function getRemoteLogoPath(filename: string): string {
    // In dev mode Vite serves public/ directly from the project root.
    // In production the renderer's public/ is built into dist/frontend/.
    if (isDev) {
        return path.join(getElectronAppBasePath(), "public", "logos", filename);
    }
    return path.join(getElectronAppBasePath(), "dist", "frontend", "logos", filename);
}

let remoteIcon: electron.NativeImage | null = null;

function getRemoteIcon(): electron.NativeImage | null {
    if (remoteIcon != null) return remoteIcon;
    const icon512 = getRemoteLogoPath("dora-logo-remote-512.png");
    const icon256 = getRemoteLogoPath("dora-logo-remote-256.png");
    // Try 512 first (higher resolution for Retina), fall back to 256
    for (const p of [icon512, icon256]) {
        const img = electron.nativeImage.createFromPath(p);
        if (!img.isEmpty()) {
            remoteIcon = img;
            return remoteIcon;
        }
    }
    console.warn("[remote-icon] could not load remote icon from", icon512);
    return null;
}

// applyRemoteDockIcon sets the macOS Dock icon to the yellow remote variant.
// Call once after app.whenReady().
export function applyRemoteDockIcon(): void {
    if (unamePlatform !== "darwin") return;
    const icon = getRemoteIcon();
    if (icon == null) return;
    electron.app.dock.setIcon(icon);
}

// applyRemoteWindowIcon sets the taskbar/window icon to the yellow remote
// variant on Windows and Linux.  Call from DoraBrowserWindow constructor.
export function applyRemoteWindowIcon(win: electron.BaseWindow): void {
    if (unamePlatform !== "win32" && unamePlatform !== "linux") return;
    const icon = getRemoteIcon();
    if (icon == null) return;
    win.setIcon(icon);
}
