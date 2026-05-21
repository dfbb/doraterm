import { base64ToString } from "@/util/util";
import { describe, expect, it, vi } from "vitest";
import { DefaultMockFilesystem } from "./mockfilesystem";

const { showPreviewContextMenu } = vi.hoisted(() => ({
    showPreviewContextMenu: vi.fn(),
}));

vi.mock("../preview-contextmenu", () => ({
    showPreviewContextMenu,
}));

describe("makeMockDoraEnv", () => {
    it("uses the preview context menu by default", async () => {
        const { makeMockDoraEnv } = await import("./mockdoraenv");
        const env = makeMockDoraEnv();
        const menu = [{ label: "Open" }];
        const event = { stopPropagation: vi.fn() } as any;

        env.showContextMenu(menu, event);

        expect(showPreviewContextMenu).toHaveBeenCalledWith(menu, event);
    });

    it("provides a populated mock filesystem rooted at /Users/mike", () => {
        expect(DefaultMockFilesystem.homePath).toBe("/Users/mike");
        expect(DefaultMockFilesystem.fileCount).toBeGreaterThanOrEqual(100);
        expect(DefaultMockFilesystem.directoryCount).toBeGreaterThanOrEqual(10);
    });

    it("implements file info, read, list, and join commands", async () => {
        const { makeMockDoraEnv } = await import("./mockdoraenv");
        const env = makeMockDoraEnv();

        const bashrcInfo = await env.rpc.FileInfoCommand(null as any, {
            info: { path: "dsh://local//Users/mike/.bashrc" },
        });
        expect(bashrcInfo.path).toBe("/Users/mike/.bashrc");
        expect(bashrcInfo.mimetype).toBe("text/plain");

        const bashrcData = await env.rpc.FileReadCommand(null as any, {
            info: { path: "dsh://local//Users/mike/.bashrc" },
        });
        expect(base64ToString(bashrcData.data64)).toContain('alias gs="git status -sb"');

        const visibleHomeEntries = await env.rpc.FileListCommand(null as any, {
            path: "/Users/mike",
        });
        expect(visibleHomeEntries.some((entry) => entry.name === ".bashrc")).toBe(false);
        expect(visibleHomeEntries.some((entry) => entry.name === "doraterm")).toBe(true);

        const allHomeEntries = await env.rpc.FileListCommand(null as any, {
            path: "/Users/mike",
            opts: { all: true },
        });
        expect(allHomeEntries.some((entry) => entry.name === ".bashrc")).toBe(true);

        const dirRead = await env.rpc.FileReadCommand(null as any, {
            info: { path: "/Users/mike/doraterm" },
        });
        expect(dirRead.entries.some((entry) => entry.name === "docs" && entry.isdir)).toBe(true);

        const joined = await env.rpc.FileJoinCommand(null as any, [
            "dsh://local//Users/mike/Documents",
            "../doraterm/docs",
            "preview-notes.md",
        ]);
        expect(joined.path).toBe("/Users/mike/doraterm/docs/preview-notes.md");
        expect(joined.mimetype).toBe("text/markdown");
    });

    it("implements file list and read stream commands", async () => {
        const { makeMockDoraEnv } = await import("./mockdoraenv");
        const env = makeMockDoraEnv();

        const listPackets: CommandRemoteListEntriesRtnData[] = [];
        for await (const packet of env.rpc.FileListStreamCommand(null as any, {
            path: "/Users/mike",
            opts: { all: true, limit: 4 },
        })) {
            listPackets.push(packet);
        }
        expect(listPackets).toHaveLength(1);
        expect(listPackets[0].fileinfo).toHaveLength(4);
    });

});
