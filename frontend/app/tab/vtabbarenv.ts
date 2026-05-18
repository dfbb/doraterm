// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import { SettingsKeyAtomFnType, DoraEnv, DoraEnvSubset } from "@/app/waveenv/waveenv";

export type VTabBarEnv = DoraEnvSubset<{
    electron: {
        createTab: DoraEnv["electron"]["createTab"];
        closeTab: DoraEnv["electron"]["closeTab"];
        setActiveTab: DoraEnv["electron"]["setActiveTab"];
        deleteWorkspace: DoraEnv["electron"]["deleteWorkspace"];
        createWorkspace: DoraEnv["electron"]["createWorkspace"];
        switchWorkspace: DoraEnv["electron"]["switchWorkspace"];
        installAppUpdate: DoraEnv["electron"]["installAppUpdate"];
    };
    rpc: {
        UpdateWorkspaceTabIdsCommand: DoraEnv["rpc"]["UpdateWorkspaceTabIdsCommand"];
        UpdateTabNameCommand: DoraEnv["rpc"]["UpdateTabNameCommand"];
        ActivityCommand: DoraEnv["rpc"]["ActivityCommand"];
        SetConfigCommand: DoraEnv["rpc"]["SetConfigCommand"];
        SetMetaCommand: DoraEnv["rpc"]["SetMetaCommand"];
    };
    atoms: {
        staticTabId: DoraEnv["atoms"]["staticTabId"];
        fullConfigAtom: DoraEnv["atoms"]["fullConfigAtom"];
        reinitVersion: DoraEnv["atoms"]["reinitVersion"];
        documentHasFocus: DoraEnv["atoms"]["documentHasFocus"];
        workspace: DoraEnv["atoms"]["workspace"];
        updaterStatusAtom: DoraEnv["atoms"]["updaterStatusAtom"];
        isFullScreen: DoraEnv["atoms"]["isFullScreen"];
    };
    services: {
        workspace: DoraEnv["services"]["workspace"];
    };
    wos: DoraEnv["wos"];
    showContextMenu: DoraEnv["showContextMenu"];
    getSettingsKeyAtom: SettingsKeyAtomFnType<"tab:confirmclose" | "app:tabbar" | "app:hideaibutton">;
    mockSetDoraObj: DoraEnv["mockSetDoraObj"];
    isWindows: DoraEnv["isWindows"];
    isMacOS: DoraEnv["isMacOS"];
}>;
