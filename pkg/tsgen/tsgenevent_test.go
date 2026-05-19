// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package tsgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dshrpc"
)

func TestGenerateDoraEventTypes(t *testing.T) {
	tsTypesMap := make(map[reflect.Type]string)
	doraEventTypeDecl := GenerateDoraEventTypes(tsTypesMap)

	if !strings.Contains(doraEventTypeDecl, `type DoraEventName = "blockclose"`) {
		t.Fatalf("expected DoraEventName declaration, got:\n%s", doraEventTypeDecl)
	}
	if !strings.Contains(doraEventTypeDecl, `{ event: "block:jobstatus"; data?: BlockJobStatusData; }`) {
		t.Fatalf("expected typed block:jobstatus event, got:\n%s", doraEventTypeDecl)
	}
	if !strings.Contains(doraEventTypeDecl, `{ event: "route:up"; data?: null; }`) {
		t.Fatalf("expected null for known no-data event, got:\n%s", doraEventTypeDecl)
	}
	if got := getDoraEventDataTSType("unmapped:event", tsTypesMap); got != "any" {
		t.Fatalf("expected any for unmapped event fallback, got: %q", got)
	}
	if _, found := tsTypesMap[reflect.TypeOf(dps.DoraEvent{})]; !found {
		t.Fatalf("expected DoraEvent type to be seeded in tsTypesMap")
	}
	if _, found := tsTypesMap[reflect.TypeOf(dshrpc.BlockJobStatusData{})]; !found {
		t.Fatalf("expected mapped data types to be generated into tsTypesMap")
	}
}
