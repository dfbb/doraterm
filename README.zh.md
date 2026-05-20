# DoraTerm

[English](./README.md)

DoraTerm 是 [WaveTerm](https://github.com/wavetermdev/waveterm) 的一个专注于终端体验的衍生版本。WaveTerm 是一个由 [Command Line Inc.](https://www.commandline.dev) 开源的、集成了 AI 能力的终端。

## 用户指南

DoraTerm 在 WaveTerm 的基础上引入了两个旗舰功能：**Remote**（在任何地方用客户端直连你内网中的 DoraTerm）和 **dsh attach**（通过 SSH 回到你的 DoraTerm，attach 到任意正在运行的终端）。

### Remote — 在任何地方连接到你的本地 DoraTerm

#### 使用场景

在家里或办公室的机器上跑着一个 DoraTerm。当你在咖啡馆、在另一个大陆、或在公司防火墙后面时，用笔记本启动第二个 DoraTerm 客户端，让它通过公网连接 *回到* 家里的实例。你看到的窗口、tab、block、实时终端输出和坐在原机器前一模一样。所有状态都保存在家里的机器上；远程客户端只是一个查看器/控制器。

#### 工作原理

DoraTerm 的后端进程（`dorasrv`）内置了一个反向代理入口，监听 TCP 端口 `31577`（可配置）。任何用 `--remote-host <url>` 启动的 DoraTerm 客户端都会跳过本地 `dorasrv` 进程，转而把所有 HTTP 和 WebSocket 流量转发给远程入口。每个请求都用密码 header 守护，因此监听器可以安全地暴露到公网。

#### 主机端（家里的机器）配置

1. 打开 DoraTerm 设置（`Cmd+,` / `Ctrl+,`）或直接编辑 `~/.config/doraterm/settings.json`，添加：

   ```json
   {
     "remote:password": "选一个足够长的随机密码",
     "remote:listenport": 31577,
     "remote:bindaddr": "127.0.0.1"
   }
   ```

   - `remote:password` — 共享密钥。**必填**。不填的话远程入口会保持关闭。
   - `remote:listenport` — 默认 `31577`。
   - `remote:bindaddr` — 默认 `127.0.0.1`。除非有特殊需要，建议保持环回地址（下面的 Cloudflare Tunnel 比直接暴露公网端口更安全）。

2. 重启 DoraTerm。监听器启动后，日志会出现 `Server [remote-entry] listening on 127.0.0.1:31577`。

#### 用 Cloudflare Tunnel 暴露主机（推荐方式）

Cloudflare Tunnel 不需要在路由器上开任何端口，就能免费拿到一个公网 HTTPS 入口。它还会自动处理 TLS，密码在传输过程中保持加密。

1. **在主机上安装 cloudflared：**

   ```bash
   # macOS
   brew install cloudflared

   # Linux (Debian/Ubuntu)
   curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
   sudo dpkg -i cloudflared.deb
   ```

2. **认证并创建 tunnel：**

   ```bash
   cloudflared tunnel login                  # 打开浏览器，选择你的 Cloudflare 域名
   cloudflared tunnel create dora            # 创建一个名为 "dora" 的 tunnel
   ```

3. **给 tunnel 绑定域名。** 把 `dora.tests.com` 换成你拥有的、托管在 Cloudflare 上的子域名：

   ```bash
   cloudflared tunnel route dns dora dora.tests.com
   ```

4. **配置 tunnel** — 编辑 `~/.cloudflared/config.yml`：

   ```yaml
   tunnel: dora
   credentials-file: /Users/yourname/.cloudflared/<tunnel-id>.json

   ingress:
     - hostname: dora.tests.com
       service: http://127.0.0.1:31577
     - service: http_status:404
   ```

5. **运行 tunnel：**

   ```bash
   cloudflared tunnel run dora
   ```

   要以后台服务运行：`sudo cloudflared service install`（详见 `cloudflared service --help`）。

如果只是想快速测试，可以用 Quick Tunnel（不需要域名，每次重启 URL 都会变）：

```bash
cloudflared tunnel --url http://127.0.0.1:31577
```

#### 远程客户端连接

1. **在远程机器上安装 DoraTerm**（你带着出门的那台笔记本）。

2. **把密码写到客户端的本地配置里**（客户端机器上的 `~/.config/doraterm/settings.json`）。客户端只从这里读取密码，其它配置全部来自主机。

   ```json
   {
     "remote:password": "选一个足够长的随机密码"
   }
   ```

3. **用 `--remote-host` 启动 DoraTerm：**

   ```bash
   # macOS
   /Applications/Dora.app/Contents/MacOS/Dora --remote-host https://dora.tests.com

   # Windows
   "C:\Program Files\Dora\Dora.exe" --remote-host https://dora.tests.com

   # Linux
   dora --remote-host https://dora.tests.com
   ```

   在同一局域网内、不走 tunnel 时，也可以直接用 `host:port` 格式：

   ```bash
   dora --remote-host 192.168.1.50:31577
   ```

远程客户端使用独立的 `userData` 目录（`doraterm-remote-<hash>`），所以不会干扰同一台机器上的本地 DoraTerm 实例。可以并存。

#### 安全提示

- 远程入口使用常数时间密码比较，并拒绝无密码/密码错误的请求。
- 在把入口暴露到公网时务必走 HTTPS（Cloudflare Tunnel、nginx、Caddy 等）。明文 HTTP 会让密码在网络上裸奔。
- 选一个足够长、熵足够高的密码。入口没有速率限制，弱密码会被暴力破解。

---

### dsh attach — SSH 回去 attach 任意正在运行的终端

#### 使用场景

你在 DoraTerm 的某个 block 里启动了一个长跑的编译、一个 `claude-code` 会话、或一个想盯着的服务。后来你换到了另一台机器，SSH 回开发机，希望直接 *跟* 那个正在运行的 shell *交互* — 不是 `tail -f` 看日志，而是像回到 GUI 一样直接在里面打字。这就是 `dsh attach` 干的事。

心智模型和 `tmux attach` 或 `screen -r` 一样，只是面向 DoraTerm 的 block。block 在 GUI 里继续跑；你的 SSH 终端成为同一个 PTY 的另一个视图 *兼* 输入源。

#### 配置

`dsh` 随每个 DoraTerm 安装包一起发布。在 DoraTerm GUI 内的任意终端 block 里，`dsh` 已经在 PATH 中 — 试试 `dsh version`。要在 *外部* 终端（例如 SSH 会话）里使用 `dsh attach`，需要手动把它放到系统 PATH 上。

**macOS：**

```bash
sudo ln -s "/Applications/Dora.app/Contents/Resources/app.asar.unpacked/dist/bin/dsh-*-darwin.$(uname -m | sed 's/x86_64/x64/')" /usr/local/bin/dsh
```

或者把 bin 目录加进 shell rc：

```bash
echo 'export PATH="/Applications/Dora.app/Contents/Resources/app.asar.unpacked/dist/bin:$PATH"' >> ~/.zshrc
# 给带版本号的二进制起别名
echo 'alias dsh="dsh-*-darwin.$(uname -m | sed s/x86_64/x64/)"' >> ~/.zshrc
```

最简单也最可移植的方式 — 直接复制对应平台的二进制：

```bash
sudo cp /Applications/Dora.app/Contents/Resources/app.asar.unpacked/dist/bin/dsh-*-darwin.arm64 /usr/local/bin/dsh   # Apple Silicon
sudo cp /Applications/Dora.app/Contents/Resources/app.asar.unpacked/dist/bin/dsh-*-darwin.x64   /usr/local/bin/dsh   # Intel
sudo chmod +x /usr/local/bin/dsh
```

**Linux：**

```bash
# 把路径换成 Dora 解压的实际位置（典型：/opt/Dora/resources/app.asar.unpacked/dist/bin 或 AppImage 解压目录）
sudo cp /opt/Dora/resources/app.asar.unpacked/dist/bin/dsh-*-linux.x64 /usr/local/bin/dsh
sudo chmod +x /usr/local/bin/dsh
```

**Windows（以管理员身份打开 PowerShell）：**

```powershell
# 把对应平台的二进制复制到已经在 PATH 上的目录
Copy-Item "C:\Program Files\Dora\resources\app.asar.unpacked\dist\bin\dsh-*-windows.x64.exe" "C:\Windows\dsh.exe"
```

安装完成后验证：

```bash
which dsh && dsh version
```

#### 使用方法

连接到你的主机（SSH、控制台、任意方式），然后：

```bash
# 交互式选择器 — 列出 workspace → tab → block
dsh attach

# 或者直接通过 block ID attach（GUI block 标题栏里显示）
dsh attach <blockid>
```

你会看到选中的 block 的当前屏幕内容被重绘到 SSH 终端里。打字 — 按键被转发到远程 PTY。GUI 与此同步更新。

#### attach 状态下的快捷键

| 键               | 动作                                              |
|------------------|---------------------------------------------------|
| `Ctrl+A` `d`     | 分离（block 继续运行）                            |
| `Ctrl+A` `r`     | 强制完全重绘                                      |
| `Ctrl+A` `s`     | 重新同步 — 用新快照重建模拟器状态                 |
| `Ctrl+A` `k`     | 向上平移视口（适用于 Ctrl+Up 被占用的终端）       |
| `Ctrl+A` `j`     | 向下平移视口                                      |
| `Ctrl+Arrow`     | 任意方向平移视口                                  |
| `Ctrl+A` `Ctrl+A` | 向远程 shell 发送一个真正的 `Ctrl+A`             |

#### 视口平移

DoraTerm 的 block 有固定的终端尺寸（由 GUI 窗口决定）。如果你从一个更小的终端 attach 进来（例如用 80 列的 SSH 会话 attach 到一个 220 列、跑着 Claude Code 的 block），本地看到的就是远程终端的一个 *窗口* — 用 `Ctrl+Arrow` 平移它，远程 PTY 不会被 resize。

---

## 与 WaveTerm 的关系

本项目是 WaveTerm 的 fork，独立开发和维护。我们感谢 Command Line Inc. 和所有 WaveTerm 贡献者，是他们打下了 DoraTerm 的基础。

DoraTerm 专注于终端体验：

- **精简**：移除了 WaveTerm 中非终端相关的功能，保持应用轻量、聚焦核心终端用例。
- **Attach**：新增 attach 到已有终端会话的能力，方便重新连接到正在运行的进程。
- **Remote**：增强了远程终端能力，让跨机器工作无缝衔接。

这些改动反映了我们的观点：一个终端应当把"终端"这件事做到极致，同时通过 WaveTerm 继承下来的扩展机制保持与其它工具的集成开放性。

## 致谢

WaveTerm 由 Command Line Inc. 创建并维护，开源社区也有贡献。没有他们的工作就没有 DoraTerm。感谢他们在 Apache 2.0 许可下构建并开源了 WaveTerm。

原项目地址：[github.com/wavetermdev/waveterm](https://github.com/wavetermdev/waveterm)。

## 许可证

DoraTerm 采用 Apache License, Version 2.0，与原 WaveTerm 相同。完整许可证文本见 [LICENSE](./LICENSE)。

原 WaveTerm 代码库的版权声明保留在 [NOTICE](./NOTICE) 中。
