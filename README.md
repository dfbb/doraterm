# DoraTerm

[简体中文](./README.zh.md)

DoraTerm is a terminal-focused derivative of [WaveTerm](https://github.com/wavetermdev/waveterm), an open-source, AI-integrated terminal originally created by [Command Line Inc.](https://www.commandline.dev)

## User Guide

DoraTerm adds two flagship features on top of WaveTerm: **Remote** (connect a client anywhere to your in-network DoraTerm) and **dsh attach** (SSH back to your DoraTerm and attach to running terminals).

### Remote — Connect to Your Local DoraTerm From Anywhere

#### Use case

Run DoraTerm on a machine at home or in your office. From a laptop in a coffee shop, on a different continent, or behind a corporate firewall, launch a second DoraTerm client and have it connect *into* your home instance over the public internet. You see the same windows, tabs, blocks, and live terminal output as if you were sitting in front of the original machine. All state lives on the home machine; the remote client is purely a viewer/controller.

#### How it works

DoraTerm's backend process (`dorasrv`) ships with an embedded reverse-proxy entry point that listens on TCP port `31577` (configurable). Any DoraTerm client launched with `--remote-host <url>` skips spawning its own `dorasrv` and instead forwards all HTTP and WebSocket traffic to the remote entry. A password header guards every request so the listener can be safely exposed to the public internet.

#### Setup on the host (your home machine)

1. Open DoraTerm settings (`Cmd+,` / `Ctrl+,`) or edit `~/.config/doraterm/settings.json` directly. Add:

   ```json
   {
     "remote:password": "pick-a-long-random-password",
     "remote:listenport": 31577,
     "remote:bindaddr": "127.0.0.1"
   }
   ```

   - `remote:password` — shared secret. Required. Without it the remote entry stays disabled.
   - `remote:listenport` — defaults to `31577`.
   - `remote:bindaddr` — defaults to `127.0.0.1`. Keep this loopback unless you have a reason to bind a public interface (the Cloudflare Tunnel below is safer than exposing a public port directly).

2. Restart DoraTerm. The log will show `Server [remote-entry] listening on 127.0.0.1:31577` once the listener is up.

#### Expose the host via Cloudflare Tunnel (recommended)

Cloudflare Tunnel gives you a free, public HTTPS endpoint without opening any inbound ports on your network. It also handles TLS, so your password stays encrypted in transit.

1. **Install cloudflared** on the host:

   ```bash
   # macOS
   brew install cloudflared

   # Linux (Debian/Ubuntu)
   curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
   sudo dpkg -i cloudflared.deb
   ```

2. **Authenticate and create a tunnel:**

   ```bash
   cloudflared tunnel login                  # opens browser, pick your Cloudflare domain
   cloudflared tunnel create dora            # creates a tunnel named "dora"
   ```

3. **Route a hostname through the tunnel.** Replace `dora.tests.com` with a subdomain on a domain you own and have on Cloudflare:

   ```bash
   cloudflared tunnel route dns dora dora.tests.com
   ```

4. **Configure the tunnel** at `~/.cloudflared/config.yml`:

   ```yaml
   tunnel: dora
   credentials-file: /Users/yourname/.cloudflared/<tunnel-id>.json

   ingress:
     - hostname: dora.tests.com
       service: http://127.0.0.1:31577
     - service: http_status:404
   ```

5. **Run the tunnel:**

   ```bash
   cloudflared tunnel run dora
   ```

   To run it as a background service: `sudo cloudflared service install` (see `cloudflared service --help`).

You can also use the quick-and-dirty Quick Tunnel for testing (no domain required, URL changes every restart):

```bash
cloudflared tunnel --url http://127.0.0.1:31577
```

#### Connect from the remote client

1. **Install DoraTerm** on the remote machine (the laptop you're traveling with).

2. **Put the password in the client's local settings** (`~/.config/doraterm/settings.json` on the client machine). The client reads only the password from this file; everything else comes from the host.

   ```json
   {
     "remote:password": "pick-a-long-random-password"
   }
   ```

3. **Launch DoraTerm with `--remote-host`:**

   ```bash
   # macOS
   /Applications/Dora.app/Contents/MacOS/Dora --remote-host https://dora.tests.com

   # Windows
   "C:\Program Files\Dora\Dora.exe" --remote-host https://dora.tests.com

   # Linux
   dora --remote-host https://dora.tests.com
   ```

   You can also use bare `host:port` form when connecting on the same LAN without a tunnel:

   ```bash
   dora --remote-host 192.168.1.50:31577
   ```

The remote client uses a separate `userData` directory (`doraterm-remote-<hash>`), so it will not interfere with a local DoraTerm instance running on the same machine. You can keep both side-by-side.

#### Security notes

- The remote entry uses a constant-time password comparison and rejects requests with no/wrong password.
- Always use HTTPS (Cloudflare Tunnel, nginx, Caddy, etc.) when exposing the entry to the public internet. Plain HTTP exposes your password to the network.
- Pick a long, high-entropy password. The entry has no rate-limiting; a weak password can be brute-forced.

---

### dsh attach — SSH In and Attach to Any Running Terminal

#### Use case

You start a long-running build, a `claude-code` session, or a server you'd like to babysit inside a DoraTerm block. Later, you're on another machine, ssh into your dev box, and want to *interact* with that exact running shell — not just `tail -f` its log, but type into it as if you were back at the GUI. `dsh attach` does that.

This is the same mental model as `tmux attach` or `screen -r`, but for DoraTerm blocks. The block keeps running in the GUI; your SSH terminal becomes an additional view *and* input source for the same PTY.

#### Setup

`dsh` ships with every DoraTerm install. Inside any terminal block in the DoraTerm GUI, `dsh` is already on the PATH — try `dsh version`. To use `dsh attach` from an *external* terminal (e.g. an SSH session), put it on the system PATH manually:

**macOS:**

```bash
sudo ln -s "/Applications/Dora.app/Contents/Resources/app.asar.unpacked/dist/bin/dsh-*-darwin.$(uname -m | sed 's/x86_64/x64/')" /usr/local/bin/dsh
```

Or just add the bin directory to your shell rc:

```bash
echo 'export PATH="/Applications/Dora.app/Contents/Resources/app.asar.unpacked/dist/bin:$PATH"' >> ~/.zshrc
# then alias the versioned binary
echo 'alias dsh="dsh-*-darwin.$(uname -m | sed s/x86_64/x64/)"' >> ~/.zshrc
```

The simplest portable option — copy the right binary directly:

```bash
sudo cp /Applications/Dora.app/Contents/Resources/app.asar.unpacked/dist/bin/dsh-*-darwin.arm64 /usr/local/bin/dsh   # Apple Silicon
sudo cp /Applications/Dora.app/Contents/Resources/app.asar.unpacked/dist/bin/dsh-*-darwin.x64   /usr/local/bin/dsh   # Intel
sudo chmod +x /usr/local/bin/dsh
```

**Linux:**

```bash
# Replace path with wherever Dora unpacked (typical: /opt/Dora/resources/app.asar.unpacked/dist/bin or your AppImage extract dir)
sudo cp /opt/Dora/resources/app.asar.unpacked/dist/bin/dsh-*-linux.x64 /usr/local/bin/dsh
sudo chmod +x /usr/local/bin/dsh
```

**Windows (PowerShell as Administrator):**

```powershell
# Copy the matching binary into a directory already on PATH
Copy-Item "C:\Program Files\Dora\resources\app.asar.unpacked\dist\bin\dsh-*-windows.x64.exe" "C:\Windows\dsh.exe"
```

Verify after installation:

```bash
which dsh && dsh version
```

#### Usage

Connect to your host (over SSH, console, anything), then:

```bash
# Interactive picker — lists workspaces → tabs → blocks
dsh attach

# Or attach directly by block ID (shown in the GUI block header)
dsh attach <blockid>
```

You'll see the current screen contents of the chosen block redrawn into your SSH terminal. Type — and the keystrokes are forwarded to the remote PTY. The GUI continues to update simultaneously.

#### Key bindings while attached

| Key            | Action                                                  |
|----------------|---------------------------------------------------------|
| `Ctrl+A` `d`   | Detach (block keeps running)                            |
| `Ctrl+A` `r`   | Force full redraw                                       |
| `Ctrl+A` `s`   | Resync — rebuild emulator state from a fresh snapshot   |
| `Ctrl+A` `k`   | Pan viewport up (for terminals where Ctrl+Up is taken)  |
| `Ctrl+A` `j`   | Pan viewport down                                       |
| `Ctrl+Arrow`   | Pan viewport in any direction                           |
| `Ctrl+A` `Ctrl+A` | Send a literal `Ctrl+A` to the remote shell          |

#### Viewport panning

DoraTerm blocks have a fixed terminal size (set by the GUI window). If you attach from a smaller terminal (e.g. an 80-column SSH session attached to a 220-column block running Claude Code), the local view is a *window* onto the remote terminal — use `Ctrl+Arrow` to pan it around without resizing the remote PTY.

---

## Relationship to WaveTerm

This project is a fork of WaveTerm, developed and maintained independently. We are grateful to Command Line Inc. and all WaveTerm contributors for building the foundation that DoraTerm is built upon.

DoraTerm focuses specifically on the terminal experience:

- **Trimmed**: Non-terminal features present in WaveTerm have been removed to keep the application lean and focused on core terminal use cases.
- **Attach**: Added the ability to attach to existing terminal sessions, making it easy to reconnect to running processes.
- **Remote**: Enhanced remote terminal capabilities for seamless work across machines.

These changes reflect our opinionated view that a terminal should excel at being a terminal, while remaining open to integration with other tools through the existing extension mechanisms inherited from WaveTerm.

## Acknowledgments

WaveTerm was created and is maintained by Command Line Inc., with contributions from the open-source community. DoraTerm would not exist without their work. We thank them for building and open-sourcing WaveTerm under the Apache 2.0 license.

For the original project, visit [github.com/wavetermdev/waveterm](https://github.com/wavetermdev/waveterm).

## License

DoraTerm is licensed under the Apache License, Version 2.0, as is the original WaveTerm. See [LICENSE](./LICENSE) for the full license text.

Copyright notices for the original WaveTerm codebase are retained in [NOTICE](./NOTICE).
