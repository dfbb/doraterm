// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package tsgen

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"github.com/dfbb/doraterm/pkg/baseds"
	"github.com/dfbb/doraterm/pkg/blockcontroller"
	"github.com/dfbb/doraterm/pkg/userinput"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dconfig"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dshrpc"
)

var waveEventRType = reflect.TypeOf(dps.DoraEvent{})

var DoraEventDataTypes = map[string]reflect.Type{
	dps.Event_BlockClose:          reflect.TypeOf(""),
	dps.Event_ConnChange:          reflect.TypeOf(dshrpc.ConnStatus{}),
	dps.Event_SysInfo:             reflect.TypeOf(dshrpc.TimeSeriesData{}),
	dps.Event_ControllerStatus:    reflect.TypeOf((*blockcontroller.BlockControllerRuntimeStatus)(nil)),
	dps.Event_DoraObjUpdate:       reflect.TypeOf(doraobj.DoraObjUpdate{}),
	dps.Event_BlockFile:           reflect.TypeOf((*dps.WSFileEventData)(nil)),
	dps.Event_Config:              reflect.TypeOf(dconfig.WatcherUpdate{}),
	dps.Event_UserInput:           reflect.TypeOf((*userinput.UserInputRequest)(nil)),
	dps.Event_RouteDown:           nil,
	dps.Event_RouteUp:             nil,
	dps.Event_WorkspaceUpdate:     nil,
	dps.Event_BlockJobStatus:      reflect.TypeOf(dshrpc.BlockJobStatusData{}),
	dps.Event_Badge:               reflect.TypeOf(baseds.BadgeEvent{}),
}

func getDoraEventDataTSType(eventName string, tsTypesMap map[reflect.Type]string) string {
	rtype, found := DoraEventDataTypes[eventName]
	if !found {
		return "any"
	}
	if rtype == nil {
		return "null"
	}
	tsType, _ := TypeToTSType(rtype, tsTypesMap)
	if tsType == "" {
		return "any"
	}
	return tsType
}

func GenerateDoraEventTypes(tsTypesMap map[reflect.Type]string) string {
	for _, rtype := range DoraEventDataTypes {
		GenerateTSType(rtype, tsTypesMap)
	}
	// suppress default struct generation, this type is custom generated
	tsTypesMap[waveEventRType] = ""

	var buf bytes.Buffer
	buf.WriteString("// dps.DoraEvent\n")
	buf.WriteString("type DoraEventName =\n")
	for _, eventName := range dps.AllEvents {
		buf.WriteString(fmt.Sprintf("    | %s\n", strconv.Quote(eventName)))
	}
	buf.WriteString(";\n\n")
	buf.WriteString("type DoraEvent = {\n")
	buf.WriteString("    event: DoraEventName;\n")
	buf.WriteString("    scopes?: string[];\n")
	buf.WriteString("    sender?: string;\n")
	buf.WriteString("    persist?: number;\n")
	buf.WriteString("    data?: unknown;\n")
	buf.WriteString("} & (\n")
	for idx, eventName := range dps.AllEvents {
		if idx > 0 {
			buf.WriteString(" | \n")
		}
		buf.WriteString(fmt.Sprintf("    { event: %s; data?: %s; }", strconv.Quote(eventName), getDoraEventDataTSType(eventName, tsTypesMap)))
	}
	buf.WriteString("\n);\n")
	return buf.String()
}
