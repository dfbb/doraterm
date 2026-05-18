// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/dfbb/doraterm/pkg/gogen"
	"github.com/dfbb/doraterm/pkg/util/utilfn"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dconfig"
	"github.com/dfbb/doraterm/pkg/dshrpc"
)

const DshClientFileName = "pkg/dshrpc/dshclient/dshclient.go"
const DoraObjMetaConstsFileName = "pkg/doraobj/metaconsts.go"
const SettingsMetaConstsFileName = "pkg/dconfig/metaconsts.go"

func GenerateDshClient() error {
	fmt.Fprintf(os.Stderr, "generating dshclient file to %s\n", DshClientFileName)
	var buf strings.Builder
	gogen.GenerateBoilerplate(&buf, "dshclient", []string{
		"github.com/dfbb/doraterm/pkg/baseds",
		"github.com/dfbb/doraterm/pkg/doraobj",
		"github.com/dfbb/doraterm/pkg/dconfig",
		"github.com/dfbb/doraterm/pkg/dps",
		"github.com/dfbb/doraterm/pkg/dshrpc",
		"github.com/dfbb/doraterm/pkg/dshutil",
	})
	wshDeclMap := dshrpc.GenerateDshCommandDeclMap()
	for _, key := range utilfn.GetOrderedMapKeys(wshDeclMap) {
		methodDecl := wshDeclMap[key]
		if methodDecl.CommandType == dshrpc.RpcType_ResponseStream {
			gogen.GenMethod_ResponseStream(&buf, methodDecl)
		} else if methodDecl.CommandType == dshrpc.RpcType_Call {
			gogen.GenMethod_Call(&buf, methodDecl)
		} else {
			panic("unsupported command type " + methodDecl.CommandType)
		}
	}
	buf.WriteString("\n")
	written, err := utilfn.WriteFileIfDifferent(DshClientFileName, []byte(buf.String()))
	if !written {
		fmt.Fprintf(os.Stderr, "no changes to %s\n", DshClientFileName)
	}
	return err
}

func GenerateDoraObjMetaConsts() error {
	fmt.Fprintf(os.Stderr, "generating doraobj meta consts file to %s\n", DoraObjMetaConstsFileName)
	var buf strings.Builder
	gogen.GenerateBoilerplate(&buf, "doraobj", []string{})
	gogen.GenerateMetaMapConsts(&buf, "MetaKey_", reflect.TypeOf(doraobj.MetaTSType{}), false)
	buf.WriteString("\n")
	written, err := utilfn.WriteFileIfDifferent(DoraObjMetaConstsFileName, []byte(buf.String()))
	if !written {
		fmt.Fprintf(os.Stderr, "no changes to %s\n", DoraObjMetaConstsFileName)
	}
	return err
}

func GenerateSettingsMetaConsts() error {
	fmt.Fprintf(os.Stderr, "generating settings meta consts file to %s\n", SettingsMetaConstsFileName)
	var buf strings.Builder
	gogen.GenerateBoilerplate(&buf, "wconfig", []string{})
	gogen.GenerateMetaMapConsts(&buf, "ConfigKey_", reflect.TypeOf(dconfig.SettingsType{}), false)
	buf.WriteString("\n")
	written, err := utilfn.WriteFileIfDifferent(SettingsMetaConstsFileName, []byte(buf.String()))
	if !written {
		fmt.Fprintf(os.Stderr, "no changes to %s\n", SettingsMetaConstsFileName)
	}
	return err
}

func main() {
	err := GenerateDshClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating dshclient: %v\n", err)
		return
	}
	err = GenerateDoraObjMetaConsts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating doraobj meta consts: %v\n", err)
		return
	}
	err = GenerateSettingsMetaConsts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating settings meta consts: %v\n", err)
		return
	}
}
