// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import type { MetaKeyAtomFnType, DoraEnv, DoraEnvSubset } from "@/app/doraenv/doraenv";

export type DoraConfigEnv = DoraEnvSubset<{
    electron: {
        getConfigDir: DoraEnv["electron"]["getConfigDir"];
        getPlatform: DoraEnv["electron"]["getPlatform"];
    };
    rpc: {
        FileInfoCommand: DoraEnv["rpc"]["FileInfoCommand"];
        FileReadCommand: DoraEnv["rpc"]["FileReadCommand"];
        FileWriteCommand: DoraEnv["rpc"]["FileWriteCommand"];
        SetMetaCommand: DoraEnv["rpc"]["SetMetaCommand"];
        GetSecretsLinuxStorageBackendCommand: DoraEnv["rpc"]["GetSecretsLinuxStorageBackendCommand"];
        GetSecretsNamesCommand: DoraEnv["rpc"]["GetSecretsNamesCommand"];
        GetSecretsCommand: DoraEnv["rpc"]["GetSecretsCommand"];
        SetSecretsCommand: DoraEnv["rpc"]["SetSecretsCommand"];
    };
    atoms: {
        fullConfigAtom: DoraEnv["atoms"]["fullConfigAtom"];
    };
    getBlockMetaKeyAtom: MetaKeyAtomFnType<"file">;
    isWindows: DoraEnv["isWindows"];
}>;
