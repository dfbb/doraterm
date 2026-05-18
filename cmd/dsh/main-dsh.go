// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/dfbb/doraterm/cmd/dsh/cmd"
	"github.com/dfbb/doraterm/pkg/dorabase"
)

// set by main-server.go
var DoraVersion = "0.0.0"
var BuildTime = "0"

func main() {
	dorabase.DoraVersion = DoraVersion
	dorabase.BuildTime = BuildTime
	cmd.Execute()
}
