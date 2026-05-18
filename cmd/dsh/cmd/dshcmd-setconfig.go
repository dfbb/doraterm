// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshrpc/dshclient"
)

var setConfigCmd = &cobra.Command{
	Use:     "setconfig",
	Short:   "set config",
	Args:    cobra.MinimumNArgs(1),
	RunE:    setConfigRun,
	PreRunE: preRunSetupRpcClient,
}

func init() {
	rootCmd.AddCommand(setConfigCmd)
}

func setConfigRun(cmd *cobra.Command, args []string) (rtnErr error) {
	defer func() {
		sendActivity("setconfig", rtnErr == nil)
	}()

	metaSetsStrs := args[:]
	meta, err := parseMetaSets(metaSetsStrs)
	if err != nil {
		return err
	}
	commandData := dshrpc.MetaSettingsType{MetaMapType: meta}
	err = dshclient.SetConfigCommand(RpcClient, commandData, &dshrpc.RpcOpts{Timeout: 2000})
	if err != nil {
		return fmt.Errorf("setting config: %w", err)
	}
	WriteStdout("config set\n")
	return nil
}
