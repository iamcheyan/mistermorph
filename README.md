# Mister Morph

Desktop app, CLI, and reusable Go runtime for running local or channel-connected agents.

Other languages: [简体中文](docs/zh-CN/README.md) | [日本語](docs/ja-JP/README.md)

If you just want to try Mister Morph, start with the desktop App from the [GitHub Releases](https://github.com/quailyquaily/mistermorph/releases) page. It bundles the Console UI, starts the local backend for you, and walks you through first-run setup.

## Why Mister Morph

- 🖥️ App-first onboarding: the desktop App removes the old multi-terminal setup path, but the CLI is still there when you want it.
- 🧩 Reusable Go core: run Mister Morph as a desktop App, CLI, Console backend, or embed it into another Go project.
- 🔀 One backend, multiple entrypoints: desktop App, Console server, CLI, and channel runtimes all build on the same core runtime.
- 🛠️ Practical extension model: built-in tools, `SKILL.md` skills, and Go embedding cover local use, automation, and integration.
- 🔒 Security-minded by design: auth profiles, outbound policy controls, approvals, and redaction are built into the runtime model.

## Quick Start

### Desktop App (recommended)

1. Download a release asset from the [GitHub Releases](https://github.com/quailyquaily/mistermorph/releases) page:
   - macOS: `mistermorph-desktop-darwin-arm64.dmg`
   - Linux: `mistermorph-desktop-linux-amd64.AppImage`
   - Windows: `mistermorph-desktop-windows-amd64.zip`
2. Launch the App.
3. Complete the setup flow inside the App.
4. Use the Console UI. You do not need to run `mistermorph console serve` manually.

Build, packaging, and platform notes: [docs/app.md](docs/app.md)

### CLI

Install a CLI binary:

```bash
curl -fsSL -o /tmp/install-mistermorph.sh https://raw.githubusercontent.com/quailyquaily/mistermorph/refs/heads/master/scripts/install-release.sh
sudo bash /tmp/install-mistermorph.sh
```

Or install from source:

```bash
go install github.com/quailyquaily/mistermorph/cmd/mistermorph@latest
```

Bootstrap a workspace, set an API key, and run one task:

```bash
mistermorph install
export MISTER_MORPH_LLM_API_KEY="YOUR_API_KEY"
mistermorph run --task "Hello!"
```

If `config.yaml` does not exist yet, `mistermorph install` launches the setup wizard and writes the initial workspace files for you.

CLI modes and configuration details: [docs/modes.md](docs/modes.md), [docs/configuration.md](docs/configuration.md)

## What Mister Morph Includes

- A desktop App for local use with first-run setup and an embedded Console UI.
- A CLI for one-shot tasks, scripting, automation, and server modes.
- A local Console server for browser-based setup, runtime management, and monitoring.
- Channel runtimes for Telegram, Slack, LINE, and Lark.
- A reusable Go integration layer for embedding Mister Morph into other projects.
- Built-in tools and a `SKILL.md`-based skills system.
- Security controls for auth profiles, outbound policies, approvals, and redaction.

## Documentation

Start here:

- [Desktop App](docs/app.md)
- [Modes](docs/modes.md)
- [Configuration](docs/configuration.md)
- [Troubleshoots](docs/troubleshoots.md)

Reference:

- [Console](docs/console.md)
- [Tools](docs/tools.md)
- [Skills](docs/skills.md)
- [Security](docs/security.md)
- [Integration](docs/integration.md)
- [Architecture](docs/arch.md)

Channel setup:

- [Telegram](docs/telegram.md)
- [Slack](docs/slack.md)
- [LINE](docs/line.md)
- [Lark](docs/lark.md)

Full docs index: [docs/README.md](docs/README.md)

## Development

Useful local commands:

```bash
./scripts/build-backend.sh --output ./bin/mistermorph
./scripts/build-desktop.sh --release
go test ./...
```

The Console frontend uses `pnpm` under `web/console/`. See [docs/console.md](docs/console.md) and [docs/app.md](docs/app.md) for local build details.

## Configuration Template

The canonical config template lives at [assets/config/config.example.yaml](assets/config/config.example.yaml).
Environment variables use the `MISTER_MORPH_` prefix. Full configuration notes and common flags are in [docs/configuration.md](docs/configuration.md).

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=quailyquaily/mistermorph&type=date&legend=top-left)](https://www.star-history.com/#quailyquaily/mistermorph&type=date&legend=top-left)
