// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package doraattach

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDataDir_EnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DORATERM_DATA_HOME", tmp)
	got, err := ResolveDataDir()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != tmp {
		t.Errorf("want %q, got %q", tmp, got)
	}
}

func TestResolveDataDir_FallbackProd(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DORATERM_DATA_HOME", "")
	prod := filepath.Join(home, ".doraterm")
	if err := os.MkdirAll(prod, 0700); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveDataDir()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != prod {
		t.Errorf("want %q, got %q", prod, got)
	}
}

func TestResolveDataDir_FallbackDev(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DORATERM_DATA_HOME", "")
	dev := filepath.Join(home, ".doraterm-dev")
	if err := os.MkdirAll(dev, 0700); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveDataDir()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != dev {
		t.Errorf("want %q, got %q", dev, got)
	}
}

func TestResolveDataDir_NoneFound(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DORATERM_DATA_HOME", "")
	if _, err := ResolveDataDir(); err == nil {
		t.Fatal("expected error when no data dir exists")
	}
}
