// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import { SettingsKeyAtomFnType, DoraEnv, DoraEnvSubset } from "@/app/waveenv/waveenv";

export type TabBarEnv = DoraEnvSubset<{
    electron: {
        createTab: DoraEnv["electron"]["createTab"];
        closeTab: DoraEnv["electron"]["closeTab"];
        setActiveTab: DoraEnv["electron"]["setActiveTab"];
        showWorkspaceAppMenu: DoraEnv["electron"]["showWorkspaceAppMenu"];
        installAppUpdate: DoraEnv["electron"]["installAppUpdate"];
    };
    rpc: {
        ActivityCommand: DoraEnv["rpc"]["ActivityCommand"];
        SetConfigCommand: DoraEnv["rpc"]["SetConfigCommand"];
        SetMetaCommand: DoraEnv["rpc"]["SetMetaCommand"];
        UpdateTabNameCommand: DoraEnv["rpc"]["UpdateTabNameCommand"];
        UpdateWorkspaceTabIdsCommand: DoraEnv["rpc"]["UpdateWorkspaceTabIdsCommand"];
    };
    atoms: {
        fullConfigAtom: DoraEnv["atoms"]["fullConfigAtom"];
        hasConfigErrors: DoraEnv["atoms"]["hasConfigErrors"];
        staticTabId: DoraEnv["atoms"]["staticTabId"];
        isFullScreen: DoraEnv["atoms"]["isFullScreen"];
        zoomFactorAtom: DoraEnv["atoms"]["zoomFactorAtom"];
        reinitVersion: DoraEnv["atoms"]["reinitVersion"];
        updaterStatusAtom: DoraEnv["atoms"]["updaterStatusAtom"];
    };
    wos: DoraEnv["wos"];
    getSettingsKeyAtom: SettingsKeyAtomFnType<"app:hideaibutton" | "app:tabbar" | "tab:confirmclose" | "window:showmenubar">;
    showContextMenu: DoraEnv["showContextMenu"];
    mockSetDoraObj: DoraEnv["mockSetDoraObj"];
    isWindows: DoraEnv["isWindows"];
    isMacOS: DoraEnv["isMacOS"];
}>;
