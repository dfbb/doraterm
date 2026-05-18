// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

// doraWindowType is set once at startup and never changes.
let doraWindowType: "tab" | "preview" = "tab";

function getDoraWindowType(): "tab" | "preview" {
    return doraWindowType;
}

function isTabWindow(): boolean {
    return doraWindowType === "tab";
}

function isPreviewWindow(): boolean {
    return doraWindowType === "preview";
}

function setDoraWindowType(windowType: "tab" | "preview") {
    doraWindowType = windowType;
}

export { getDoraWindowType, isPreviewWindow, isTabWindow, setDoraWindowType };
