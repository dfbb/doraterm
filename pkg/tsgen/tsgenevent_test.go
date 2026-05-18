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
	waveEventTypeDecl := GenerateDoraEventTypes(tsTypesMap)

	if !strings.Contains(waveEventTypeDecl, `type DoraEventName = "blockclose"`) {
		t.Fatalf("expected DoraEventName declaration, got:\n%s", waveEventTypeDecl)
	}
	if !strings.Contains(waveEventTypeDecl, `{ event: "block:jobstatus"; data?: BlockJobStatusData; }`) {
		t.Fatalf("expected typed block:jobstatus event, got:\n%s", waveEventTypeDecl)
	}
	if !strings.Contains(waveEventTypeDecl, `{ event: "route:up"; data?: null; }`) {
		t.Fatalf("expected null for known no-data event, got:\n%s", waveEventTypeDecl)
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
