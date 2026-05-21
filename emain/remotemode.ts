// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import * as dns from "node:dns/promises";
import * as fs from "node:fs";
import * as net from "node:net";
import * as path from "node:path";

const DEFAULT_PORT = 31577;

// baseUrl is the full origin, e.g. "http://192.168.1.1:31577" or "https://dora.example.com"
// displayHost is the original hostname the user passed (for UI/log); host/baseUrl may be
// rewritten to an IP after DNS resolution to bypass Host-header–based interception
// (e.g. Tencent Cloud DNSPod redirects HTTP traffic when the Host header matches certain
// unregistered/reserved domains, even though the IP itself is reachable).
export type RemoteTarget = { host: string; port: number; baseUrl: string; displayHost: string };

export type RemoteModeState = {
    isRemote: boolean;
    target: RemoteTarget | null;
    password: string | null;
    safeSuffix: string | null; // e.g. "host_example_com-31577"
};

function parseRemoteHostArg(argv: string[]): RemoteTarget | null {
    let value: string | null = null;
    for (let i = 0; i < argv.length; i++) {
        const a = argv[i];
        if (a === "--remote-host") {
            value = argv[i + 1] ?? null;
            break;
        }
        if (a.startsWith("--remote-host=")) {
            value = a.slice("--remote-host=".length);
            break;
        }
    }
    if (!value) return null;
    // Accept full https:// or http:// URLs (e.g. Cloudflare Tunnel)
    if (value.startsWith("https://") || value.startsWith("http://")) {
        const url = new URL(value);
        const port = url.port ? parseInt(url.port, 10) : url.protocol === "https:" ? 443 : 80;
        const baseUrl = url.origin;
        return { host: url.hostname, port, baseUrl, displayHost: url.hostname };
    }
    const idx = value.lastIndexOf(":");
    if (idx < 0) {
        return { host: value, port: DEFAULT_PORT, baseUrl: `http://${value}:${DEFAULT_PORT}`, displayHost: value };
    }
    const host = value.slice(0, idx);
    const port = parseInt(value.slice(idx + 1), 10);
    if (!host || isNaN(port) || port <= 0 || port > 65535) {
        throw new Error(`invalid --remote-host value: ${value}`);
    }
    return { host, port, baseUrl: `http://${host}:${port}`, displayHost: host };
}

// Resolve a plain-HTTP hostname target to an IP-based target. This rewrites both
// the TCP connect host AND the HTTP Host header (via baseUrl) so middleboxes that
// filter on Host (e.g. cloud DNSPod webblock) can't intercept the request.
// HTTPS is intentionally skipped because TLS SNI and cert validation need the hostname.
export async function resolveTargetIP(target: RemoteTarget): Promise<RemoteTarget> {
    if (target.baseUrl.startsWith("https://")) return target;
    if (net.isIP(target.host)) return target;
    try {
        const { address } = await dns.lookup(target.host);
        const hostForUrl = net.isIPv6(address) ? `[${address}]` : address;
        return {
            ...target,
            host: address,
            baseUrl: `http://${hostForUrl}:${target.port}`,
        };
    } catch (e) {
        console.warn(
            `[remote-mode] DNS lookup failed for ${target.host}; using hostname directly:`,
            (e as Error).message
        );
        return target;
    }
}

function readPasswordFromSettings(settingsDir: string): string | null {
    const p = path.join(settingsDir, "settings.json");
    if (!fs.existsSync(p)) return null;
    try {
        const raw = fs.readFileSync(p, "utf-8");
        const parsed = JSON.parse(raw);
        const v = parsed["remote:password"];
        return typeof v === "string" && v.length > 0 ? v : null;
    } catch {
        return null;
    }
}

function safeSuffixFor(target: RemoteTarget): string {
    return `${target.host}-${target.port}`.replace(/[^a-z0-9-]/gi, "_");
}

// Resolve once at startup. Callers pass the local config dir explicitly
// (not derived from getDoraConfigDir, because that may already account for
// the remote-mode userData path).
export function resolveRemoteMode(argv: string[], localConfigDir: string): RemoteModeState {
    const target = parseRemoteHostArg(argv);
    if (target == null) {
        return { isRemote: false, target: null, password: null, safeSuffix: null };
    }
    const password = readPasswordFromSettings(localConfigDir);
    return {
        isRemote: true,
        target,
        password,
        safeSuffix: safeSuffixFor(target),
    };
}
