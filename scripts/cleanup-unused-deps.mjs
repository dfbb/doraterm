#!/usr/bin/env node
import { execSync, spawnSync } from "child_process";
import { readFileSync, writeFileSync } from "fs";
import { resolve, dirname } from "path";
import { fileURLToPath } from "url";


const __filename = fileURLToPath(import.meta.url);
const ROOT = resolve(dirname(__filename), "..");
const PKG_PATH = resolve(ROOT, "package.json");
const DRY_RUN = process.argv.includes("--dry-run");
const FORCE = process.argv.includes("--force");
const BUILD_TIMEOUT_MS = 5 * 60 * 1000;

// depcheck can't see packages referenced only in these non-JS-import contexts
const SKIP_PATTERNS = [/^@types\//, /^tailwindcss$/];

// Rollback state for SIGINT handler
let currentBackup = null;
let rollbackInProgress = false;
process.on("SIGINT", () => {
    if (rollbackInProgress) {
        console.log("\n  Rollback interrupted! Run 'npm install' to recover.");
        process.exit(1);
    }
    if (currentBackup) {
        rollbackInProgress = true;
        console.log("\n  Interrupted. Rolling back current package...");
        try {
            writeFileSync(PKG_PATH, currentBackup);
            execSync("npm install --silent", { cwd: ROOT, stdio: "inherit" });
        } catch {
            console.log("  Rollback failed! Run 'npm install' to recover.");
        }
    }
    process.exit(1);
});

function escapeRegex(str) {
    return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function getConfigReferencedPackages() {
    const pkg = JSON.parse(readFileSync(PKG_PATH, "utf8"));
    const allPkgs = new Set([
        ...Object.keys(pkg.dependencies ?? {}),
        ...Object.keys(pkg.devDependencies ?? {}),
    ]);
    const referenced = new Set();
    const configFiles = [
        "postinstall.cjs",
        "electron-builder.config.cjs",
        "version.cjs",
        "prettier.config.cjs",
        "eslint.config.js",
    ];
    for (const file of configFiles) {
        try {
            const content = readFileSync(resolve(ROOT, file), "utf8");
            for (const p of allPkgs) {
                const escaped = escapeRegex(p);
                const pattern = new RegExp(`(?<![a-z0-9@._/-])${escaped}(?![a-z0-9@._/-])`);
                if (pattern.test(content)) referenced.add(p);
            }
        } catch {}
    }
    return referenced;
}

function shouldSkip(pkg, configReferenced) {
    return SKIP_PATTERNS.some((p) => p.test(pkg)) || configReferenced.has(pkg);
}

function run(cmd) {
    return execSync(cmd, { cwd: ROOT, encoding: "utf8", stdio: ["pipe", "pipe", "pipe"] });
}

function runDepcheck() {
    try {
        const out = run("npx --yes depcheck --json");
        return JSON.parse(out);
    } catch (e) {
        if (e.stdout) return JSON.parse(e.stdout);
        throw e;
    }
}

function runBuild() {
    const result = spawnSync("npm", ["run", "build:dev"], {
        cwd: ROOT,
        stdio: "inherit",
        timeout: BUILD_TIMEOUT_MS,
    });
    return result.status === 0 && !result.error;
}

function isValidPackageName(pkg) {
    return /^@?[a-z0-9][a-z0-9._/-]*$/.test(pkg);
}

async function main() {
    // Phase 1: Commit staged deletions
    const status = run("git status --short").trim();
    if (status) {
        console.log("Phase 1: Committing staged deletions...");
        const lines = status.split("\n").filter(Boolean);
        const hasNonDelete = lines.some((l) => {
            const code = l.slice(0, 2).trim();
            return code !== "D" && code !== "??";
        });
        if (hasNonDelete && !FORCE) {
            console.error("  Error: Non-deletion changes found in working tree.");
            console.error("  Commit or stash your changes first, or use --force to override.");
            process.exit(1);
        }
        if (DRY_RUN) {
            console.log(`  [dry-run] Would commit ${lines.length} changes`);
        } else {
            run("git add -u");
            run('git commit -m "chore: remove stale files"');
            console.log(`  ✓ Committed (${lines.length} files)`);
        }
    } else {
        console.log("Phase 1: No staged changes.");
    }

    // Phase 2: Scan
    console.log("\nPhase 2: Scanning for unused dependencies...");
    const depcheckResult = runDepcheck();
    const configReferenced = getConfigReferencedPackages();
    const raw = [
        ...(depcheckResult.dependencies ?? []),
        ...(depcheckResult.devDependencies ?? []),
    ];
    const candidates = raw.filter((p) => !shouldSkip(p, configReferenced));

    if (candidates.length === 0) {
        console.log("  No unused dependencies found after filtering.");
        return;
    }
    console.log(`  Found ${candidates.length} candidates (${raw.length - candidates.length} auto-filtered)`);

    if (DRY_RUN) {
        console.log("\n[dry-run] Candidates:");
        candidates.forEach((p) => console.log(`  - ${p}`));
        return;
    }

    // Phase 3: Test removals
    console.log("\nPhase 3: Testing removals...");
    const removed = [];
    const skipped = [];

    const MaxPkgLen = Math.max(...candidates.map((p) => p.length));
    for (let i = 0; i < candidates.length; i++) {
        const pkg = candidates[i];
        const label = `[${i + 1}/${candidates.length}] ${pkg}`;
        process.stdout.write(`  ${label.padEnd(MaxPkgLen + 15)} ... `);

        currentBackup = readFileSync(PKG_PATH, "utf8");

        if (!isValidPackageName(pkg)) {
            process.stdout.write("✗ skipped (invalid package name)\n");
            currentBackup = null;
            skipped.push(pkg);
            continue;
        }

        try {
            run(`npm uninstall ${pkg}`);
        } catch {
            process.stdout.write("✗ skipped (uninstall failed)\n");
            currentBackup = null;
            skipped.push(pkg);
            continue;
        }

        const ok = runBuild();
        if (ok) {
            process.stdout.write("✓ removed\n");
            removed.push(pkg);
        } else {
            process.stdout.write("✗ skipped (build failed or timed out)\n");
            writeFileSync(PKG_PATH, currentBackup);
            run("npm install --silent");
            skipped.push(pkg);
        }
        currentBackup = null;
    }

    // Phase 4: Report
    console.log(`\n${"─".repeat(60)}`);
    console.log(`Removed : ${removed.length} packages`);
    if (removed.length) console.log(`  ${removed.join(", ")}`);
    console.log(`Skipped : ${skipped.length} packages`);
    if (skipped.length) console.log(`  ${skipped.join(", ")}`);
}

main().catch((e) => {
    console.error(e);
    process.exit(1);
});
