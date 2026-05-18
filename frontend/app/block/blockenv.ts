// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import {
    ConnConfigKeyAtomFnType,
    MetaKeyAtomFnType,
    SettingsKeyAtomFnType,
    DoraEnv,
    DoraEnvSubset,
} from "@/app/waveenv/waveenv";

export type BlockEnv = DoraEnvSubset<{
    getSettingsKeyAtom: SettingsKeyAtomFnType<
        | "app:focusfollowscursor"
        | "app:showoverlayblocknums"
        | "term:showsplitbuttons"
        | "window:magnifiedblockblurprimarypx"
        | "window:magnifiedblockopacity"
    >;
    showContextMenu: DoraEnv["showContextMenu"];
    atoms: {
        modalOpen: DoraEnv["atoms"]["modalOpen"];
        controlShiftDelayAtom: DoraEnv["atoms"]["controlShiftDelayAtom"];
    };
    electron: {
        openExternal: DoraEnv["electron"]["openExternal"];
    };
    rpc: {
        ActivityCommand: DoraEnv["rpc"]["ActivityCommand"];
    };
    wos: DoraEnv["wos"];
    getConnStatusAtom: DoraEnv["getConnStatusAtom"];
    getLocalHostDisplayNameAtom: DoraEnv["getLocalHostDisplayNameAtom"];
    getConnConfigKeyAtom: ConnConfigKeyAtomFnType<"conn:wshenabled">;
    getBlockMetaKeyAtom: MetaKeyAtomFnType<
        | "frame:text"
        | "frame:activebordercolor"
        | "frame:bordercolor"
        | "view"
        | "connection"
        | "icon:color"
        | "frame:title"
        | "frame:icon"
    >;
    getTabMetaKeyAtom: MetaKeyAtomFnType<"bg:activebordercolor" | "bg:bordercolor" | "tab:background">;
    getConfigBackgroundAtom: DoraEnv["getConfigBackgroundAtom"];
}>;
