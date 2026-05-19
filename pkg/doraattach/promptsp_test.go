// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package doraattach

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestPromptSP_RenderFiltersPercent(t *testing.T) {
	promptCursor := "\x1b[1;16H"
	promptSP := "\x1b[1m\x1b[7m%\x1b[27m\x1b[1m\x1b[0m" + strings.Repeat(" ", 138) + "\r \r"

	vp := newViewport(24, 139, 80, 24)
	vp.Write([]byte(promptCursor))
	vp.Write([]byte(promptSP))

	var out strings.Builder
	vp.Render(&out)

	if strings.Contains(out.String(), "%") {
		t.Errorf("render output contains %% (PROMPT_SP bold+inverse %% should be filtered)")
	}
}

func TestPromptSP_RealData(t *testing.T) {
	snapshot, err := os.ReadFile("/tmp/snapshot.bin")
	if err != nil {
		t.Skipf("no snapshot.bin: %v", err)
	}
	var events [][]byte
	for i := 0; ; i++ {
		data, err := os.ReadFile(fmt.Sprintf("/tmp/event_%d.bin", i))
		if err != nil {
			break
		}
		events = append(events, data)
	}

	vp := newViewport(24, 139, 80, 24)
	if _, err := vp.Write(snapshot); err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	for i, ev := range events {
		if _, err := vp.Write(ev); err != nil {
			t.Fatalf("event %d: %v", i, err)
		}
		var out strings.Builder
		vp.Render(&out)
		if strings.Contains(out.String(), "%") {
			t.Errorf("event %d render has %%", i)
		}
	}
}
