// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package gogen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dfbb/doraterm/pkg/dshrpc"
)

func TestGetWshMethodDataParamsAndExpr_MultiArg(t *testing.T) {
	methodDecl := &dshrpc.WshRpcMethodDecl{
		CommandDataTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(0),
		},
	}
	params, expr := getWshMethodDataParamsAndExpr(methodDecl)
	if params != ", arg1 string, arg2 int" {
		t.Fatalf("unexpected params: %q", params)
	}
	if expr != "dshrpc.MultiArg{Args: []any{arg1, arg2}}" {
		t.Fatalf("unexpected expr: %q", expr)
	}
}

func TestGenMethodCall_MultiArg(t *testing.T) {
	methodDecl := &dshrpc.WshRpcMethodDecl{
		Command:          "test",
		CommandType:      dshrpc.RpcType_Call,
		MethodName:       "TestCommand",
		CommandDataTypes: []reflect.Type{reflect.TypeOf(""), reflect.TypeOf(0)},
	}
	var sb strings.Builder
	GenMethod_Call(&sb, methodDecl)
	out := sb.String()
	if !strings.Contains(out, "func TestCommand(w *dshutil.WshRpc, arg1 string, arg2 int, opts *dshrpc.RpcOpts) error {") {
		t.Fatalf("generated method missing multi-arg signature:\n%s", out)
	}
	if !strings.Contains(out, "sendRpcRequestCallHelper[any](w, \"test\", dshrpc.MultiArg{Args: []any{arg1, arg2}}, opts)") {
		t.Fatalf("generated method missing MultiArg payload:\n%s", out)
	}
}
