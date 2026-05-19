# Cleanup Unused Dependencies Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Commit stale deleted files and automatically remove unused npm packages via a build-verified loop.

**Architecture:** A single Node.js ESM script (`scripts/cleanup-unused-deps.mjs`) runs in four phases: commit staged deletions, scan with depcheck, test-remove each candidate one at a time (rollback on build failure), then print a summary. No new runtime dependencies are added.

**Tech Stack:** Node.js ESM, `npx depcheck`, `npm uninstall`, `electron-vite` (via `npm run build:dev`)

---

## File Structure

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `scripts/cleanup-unused-deps.mjs` | All four phases of the cleanup automation |

---

### Task 1: Create the cleanup script

**Files:**
- Create: `scripts/cleanup-unused-deps.mjs`

- [ ] **Step 1: Create `scripts/cleanup-unused-deps.mjs` with the full implementation**

```javascript
#!/usr/bin/env node
import { execSync, spawnSync } from "child_process";
import { readFileSync, writeFileSync } from "fs";
import { resolve, dirname } from "path";
import { fileURLToPath } from "url";
import { createInterface } from "readline";

const __filename = fileURLToPath(import.meta.url);
const ROOT = resolve(dirname(__filename), "..");
const PKG_PATH = resolve(ROOT, "package.json");
const DRY_RUN = process.argv.includes("--dry-run");
const BUILD_TIMEOUT_MS = 5 * 60 * 1000;

// depcheck can't see packages referenced only in these non-JS-import contexts
const SKIP_PATTERNS = [/^@types\//];

// Rollback state for SIGINT handler
let currentBackup = null;
process.on("SIGINT", () => {
    if (currentBackup) {
        console.log("\n  Interrupted. Rolling back current package...");
        writeFileSync(PKG_PATH, currentBackup);
        execSync("npm install --silent", { cwd: ROOT, stdio: "inherit" });
    }
    process.exit(1);
});

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
                if (content.includes(p)) referenced.add(p);
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

async function confirmProceed(msg) {
    const rl = createInterface({ input: process.stdin, output: process.stdout });
    return new Promise((resolve) => {
        rl.question(msg + " [Enter to continue, Ctrl+C to abort] ", () => {
            rl.close();
            resolve();
        });
    });
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
        if (hasNonDelete && !DRY_RUN) {
            await confirmProceed(`  ⚠  git status has non-deletion changes (${lines.length} total).`);
        }
        if (DRY_RUN) {
            console.log(`  [dry-run] Would commit ${lines.length} changes`);
        } else {
            run("git add -A");
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

    for (let i = 0; i < candidates.length; i++) {
        const pkg = candidates[i];
        const label = `[${i + 1}/${candidates.length}] ${pkg}`;
        process.stdout.write(`  ${label.padEnd(55)} ... `);

        currentBackup = readFileSync(PKG_PATH, "utf8");

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
```

- [ ] **Step 2: Verify the file was created**

```bash
ls -la scripts/cleanup-unused-deps.mjs
```

Expected: file exists, non-zero size.

- [ ] **Step 3: Commit the script**

```bash
git add scripts/cleanup-unused-deps.mjs
git commit -m "chore: add cleanup-unused-deps script"
```

---

### Task 2: Dry-run preview

**Files:**
- Run: `scripts/cleanup-unused-deps.mjs --dry-run`

- [ ] **Step 1: Run dry-run from project root**

```bash
node scripts/cleanup-unused-deps.mjs --dry-run
```

Expected output (shape, not exact values):
```
Phase 1: Committing staged deletions...
  [dry-run] Would commit N changes

Phase 2: Scanning for unused dependencies...
  Found N candidates (M auto-filtered)

[dry-run] Candidates:
  - some-package
  - another-package
  ...
```

- [ ] **Step 2: Sanity-check the candidate list**

Scan the printed list. If any obviously-needed package appears (e.g. `electron`, `vite`, `tailwindcss`, `react`), it means depcheck missed a usage. In that case, add the package name to `SKIP_PATTERNS` or the `configFiles` list in `getConfigReferencedPackages()` before proceeding.

If the list looks reasonable (old AI libs, unused utilities, etc.), continue.

---

### Task 3: Run the full cleanup

**Files:**
- Run: `scripts/cleanup-unused-deps.mjs`
- Modified: `package.json`, `package-lock.json` (by npm)

- [ ] **Step 1: Run the cleanup (this will take a while — one build per candidate)**

```bash
node scripts/cleanup-unused-deps.mjs
```

Expected output (shape):
```
Phase 1: Committing staged deletions...
  ✓ Committed (N files)

Phase 2: Scanning for unused dependencies...
  Found N candidates (M auto-filtered)

Phase 3: Testing removals...
  [1/N] some-package                                     ... ✓ removed
  [2/N] another-package                                  ... ✗ skipped (build failed or timed out)
  ...

────────────────────────────────────────────────────────────
Removed : N packages
  pkg-a, pkg-b, ...
Skipped : N packages
  pkg-c, pkg-d, ...
```

- [ ] **Step 2: Verify the project still builds after all removals**

```bash
npm run build:dev
```

Expected: exits 0 with no errors. If this fails (shouldn't happen since every removal was build-verified), run `git restore package.json && npm install` to recover.

---

### Task 4: Commit cleanup results

**Files:**
- Commit: `package.json`, `package-lock.json`

- [ ] **Step 1: Check what changed**

```bash
git diff --stat
```

Expected: `package.json` and `package-lock.json` modified, showing removed packages.

- [ ] **Step 2: Commit**

```bash
git add package.json package-lock.json
git commit -m "chore: remove unused npm dependencies"
```

- [ ] **Step 3: Verify clean state**

```bash
git status
```

Expected: `nothing to commit, working tree clean`
