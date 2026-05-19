// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/dfbb/doraterm/pkg/doraattach"
)

var attachCmd = &cobra.Command{
	Use:                   "attach [blockid]",
	Short:                 "attach to a Dora Terminal block from an external terminal",
	Long:                  "Attach to a running term block in Dora Terminal. Press Ctrl+A D to detach.",
	Args:                  cobra.MaximumNArgs(1),
	RunE:                  attachRun,
	DisableFlagsInUseLine: true,
}

func init() {
	rootCmd.AddCommand(attachCmd)
}

func attachRun(cmd *cobra.Command, args []string) error {
	rpcClient, _, err := doraattach.Connect()
	if err != nil {
		return err
	}

	var blockId string
	if len(args) == 1 {
		blockId = args[0]
	} else {
		blockId, err = doraattach.SelectBlock(rpcClient)
		if err != nil {
			return err
		}
	}

	return doraattach.Attach(rpcClient, blockId)
}
