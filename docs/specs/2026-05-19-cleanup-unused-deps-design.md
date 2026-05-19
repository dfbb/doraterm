# 清理未使用依赖设计文档

**日期**: 2026-05-19  
**状态**: 待实施

## 背景

项目经过大规模精简（删除了 AI 模块、WSL、telemetry、VDOM 等功能），遗留了两类需清理的内容：

1. **项目目录废弃文件**：`git status` 里已标记删除但未 commit 的文件（`aiprompts/`、`assets/waveterm-*` 等）
2. **未使用的 npm 依赖**：功能删除后，对应的 npm 包仍留在 `package.json` 中

## 目标

- 提交已删除的废弃文件
- 自动审计并移除不再被代码引用的 npm 包
- 每次移除一个包后构建验证，失败则回退，保证仓库始终可构建

## 整体流程

```
Phase 1: 提交已删除文件
  git add -A → git commit "chore: remove stale files"

Phase 2: 扫描未使用依赖
  npx depcheck --json → 解析 unusedDependencies + unusedDevDependencies

Phase 3: 逐包自动清理（主循环）
  for each pkg in candidates:
    1. 备份 package.json
    2. npm uninstall <pkg>
    3. npm run build:dev（最多等待 5 分钟）
    4. 构建成功 → 记录"已删除"，继续
    5. 构建失败/超时 → 恢复 package.json + npm install，记录"跳过"

Phase 4: 输出最终报告
  已删除: [...], 跳过: [...], 节省磁盘: N MB
```

## 脚本

**位置**: `scripts/cleanup-unused-deps.mjs`（Node.js ESM，无第三方依赖）

**运行方式**:
```bash
node scripts/cleanup-unused-deps.mjs           # 正常运行
node scripts/cleanup-unused-deps.mjs --dry-run # 只打印候选列表，不删除
```

## 自动过滤规则（depcheck 误报）

以下类型的包由脚本自动跳过，不进入候选列表：

| 类型 | 原因 |
|------|------|
| `@types/*` | TypeScript 类型包，depcheck 常误报为 unused |
| `electron-builder` 相关 | 在 `electron-builder.config.cjs` 里引用 |
| `postinstall.cjs` 里引用的包 | depcheck 不扫描 `.cjs` 脚本 |

## Rollback 机制

```
备份 = 读取当前 package.json 内容
执行 npm uninstall <pkg>
运行构建（超时 5 分钟）
if (构建失败 || 超时) {
  写回备份的 package.json
  执行 npm install   // 重建 node_modules
  标记该包为"跳过"
}
```

## 边界处理

| 场景 | 处理方式 |
|------|----------|
| Phase 1 commit 时有非删除的修改 | 提示用户确认，防止误提交 |
| depcheck 候选列表为空 | 打印"无未使用依赖"并退出 |
| 构建超时（> 5 分钟） | 视为失败，执行 rollback |
| 用户 Ctrl+C 中断 | 捕获 SIGINT，先完成当前包的 rollback 再退出 |
| `--dry-run` 模式 | 只运行 depcheck 并打印过滤后的候选列表，不做任何修改 |

## 实时输出格式

```
Phase 1: Committing staged deletions...
  ✓ Committed (42 files deleted)

Phase 2: Scanning for unused dependencies...
  Found 23 candidates (after filtering)

Phase 3: Testing removals...
  [1/23] some-package        ... ✓ removed
  [2/23] another-pkg         ... ✗ skipped (build failed)
  [3/23] old-ai-library      ... ✓ removed
  ...

Done in 18m 32s
Removed: 15 packages | Skipped: 8 packages
Disk saved: ~45 MB
```

## 成功标准

- Phase 1 commit 成功
- 脚本运行全程无崩溃，每次 rollback 后 `package.json` 与 `node_modules` 保持一致
- 最终 `npm run build:dev` 成功（可手动验证一次）
- `package-lock.json` 已更新并可提交
