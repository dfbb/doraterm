// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build windows || (linux && (mips || mips64))

package doraattach

import (
	"fmt"

	"github.com/dfbb/doraterm/pkg/dshutil"
)

func ResolveDataDir() (string, error) {
	return "", fmt.Errorf("dsh attach is not supported on this platform")
}

func Connect() (*dshutil.DshRpc, string, error) {
	return nil, "", fmt.Errorf("dsh attach is not supported on this platform")
}

func SelectBlock(rpcClient *dshutil.DshRpc) (string, error) {
	return "", fmt.Errorf("dsh attach is not supported on this platform")
}

func Attach(rpcClient *dshutil.DshRpc, blockId string) error {
	return fmt.Errorf("dsh attach is not supported on this platform")
}
