// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import type { AllServiceImpls } from "@/app/store/services";
import { RpcApiType } from "@/app/store/wshclientapi";
import { Atom, PrimitiveAtom } from "jotai";
import React from "react";

export type MetaKeyAtomFnType<Keys extends keyof MetaType = keyof MetaType> = <T extends Keys>(
    id: string,
    key: T
) => Atom<MetaType[T]>;

export type ConnConfigKeyAtomFnType<Keys extends keyof ConnKeywords = keyof ConnKeywords> = <T extends Keys>(
    connName: string,
    key: T
) => Atom<ConnKeywords[T]>;

export type SettingsKeyAtomFnType<Keys extends keyof SettingsType = keyof SettingsType> = <T extends Keys>(
    key: T
) => Atom<SettingsType[T]>;

type OmitNever<T> = {
    [K in keyof T as [T[K]] extends [never] ? never : K]: T[K];
};

type Subset<T, U> = OmitNever<{
    [K in keyof T]: K extends keyof U ? T[K] : never;
}>;

type ComplexDoraEnvKeys = {
    rpc: DoraEnv["rpc"];
    electron: DoraEnv["electron"];
    atoms: DoraEnv["atoms"];
    wos: DoraEnv["wos"];
    services: DoraEnv["services"];
};

type DoraEnvMockFields = {
    isMock: DoraEnv["isMock"];
    mockSetDoraObj: DoraEnv["mockSetDoraObj"];
    mockModels: DoraEnv["mockModels"];
};

export type DoraEnvSubset<T> = DoraEnvMockFields &
    OmitNever<{
        [K in keyof T]: K extends keyof ComplexDoraEnvKeys
            ? Subset<T[K], ComplexDoraEnvKeys[K]>
            : K extends keyof DoraEnv
              ? T[K]
              : never;
    }>;

// default implementation for production is in ./waveenvimpl.ts
export type DoraEnv = {
    isMock: boolean;
    electron: ElectronApi;
    rpc: RpcApiType;
    platform: NodeJS.Platform;
    isDev: () => boolean;
    isWindows: () => boolean;
    isMacOS: () => boolean;
    atoms: GlobalAtomsType;
    createBlock: (blockDef: BlockDef, magnified?: boolean, ephemeral?: boolean) => Promise<string>;
    services: typeof AllServiceImpls;
    callBackendService: (service: string, method: string, args: any[], noUIContext?: boolean) => Promise<any>;
    showContextMenu: (menu: ContextMenuItem[], e: React.MouseEvent) => void;
    getConnStatusAtom: (conn: string) => PrimitiveAtom<ConnStatus>;
    getLocalHostDisplayNameAtom: () => Atom<string>;
    wos: {
        getDoraObjectAtom: <T extends DoraObj>(oref: string) => Atom<T>;
        getDoraObjectLoadingAtom: (oref: string) => Atom<boolean>;
        isDoraObjectNullAtom: (oref: string) => Atom<boolean>;
        useDoraObjectValue: <T extends DoraObj>(oref: string) => [T, boolean];
    };
    getSettingsKeyAtom: SettingsKeyAtomFnType;
    getBlockMetaKeyAtom: MetaKeyAtomFnType;
    getTabMetaKeyAtom: MetaKeyAtomFnType;
    getConnConfigKeyAtom: ConnConfigKeyAtomFnType;
    getConfigBackgroundAtom: (bgKey: string | null) => Atom<BackgroundConfigType>;

    // the mock fields are only usable in the preview server (may be be null or throw errors in production)
    mockSetDoraObj: <T extends DoraObj>(oref: string, obj: T) => void;
    mockModels: Map<any, any>;
};

export const DoraEnvContext = React.createContext<DoraEnv>(null);

type EnvContract<T> = {
    [K in keyof T]?: T[K] extends (...args: any[]) => any ? T[K] : T[K] extends object ? EnvContract<T[K]> : T[K];
};

export function useDoraEnv<T extends EnvContract<DoraEnv> = DoraEnv>(): T {
    return React.useContext(DoraEnvContext) as T;
}
