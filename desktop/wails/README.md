# MisterMorph Desktop (Wails v3)

This directory contains the Wails desktop host for `mistermorph`.

## Current MVP wiring

- Runs a local `mistermorph console serve` subprocess on a random loopback port.
- Prefers a sibling or configured `mistermorph` backend binary and can auto-download one as fallback.
- If no `mistermorph` backend binary is found, desktop host tries to download a matching release binary first.
- Proxies the Wails WebView traffic to the local console server at root path `/`.
- Exposes a Go binding `App.RestartApp()` for setup-complete restart.

## Dev prerequisites

- Go (same version as repository)
- Wails v3 desktop dependencies installed for your OS
- Built and staged console assets for the backend binary

On Ubuntu/Debian, install the Linux desktop build deps first:

```bash
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
```

Build console assets first:

```bash
pnpm --dir web/console build
./scripts/stage-console-assets.sh
go build -o ./bin/mistermorph ./cmd/mistermorph
```

Run desktop app from source:

```bash
go run -tags 'wailsdesktop production' ./desktop/wails
```

Build desktop binary:

```bash
go build -tags 'wailsdesktop production' -o ./bin/mistermorph-desktop ./desktop/wails
```

For local Linux builds with DevTools enabled, use [`scripts/build-desktop.sh`](/home/lyric/Codework/arch/mistermorph/scripts/build-desktop.sh). It automatically switches Linux debug builds to `wailsdesktop dev devtools`, because Wails v3 alpha does not currently support `linux + production + devtools`.

## Config file forwarding

If you start desktop app with `--config <path>`, that path is forwarded to the child `console serve` subprocess.

## Backend binary discovery/download

Backend binary candidate order:

1. `MISTERMORPH_DESKTOP_BACKEND_BIN`
2. `./bin/mistermorph` (or `.exe` on Windows)
3. sibling paths near desktop executable
4. `PATH` lookup (`mistermorph`)
5. download from GitHub releases (enabled by default)

Optional envs:

- `MISTERMORPH_DESKTOP_BACKEND_AUTO_DOWNLOAD=true|false` (default `true`)
- `MISTERMORPH_DESKTOP_BACKEND_VERSION=latest|vX.Y.Z` (default `latest`)
- `MISTERMORPH_DESKTOP_BACKEND_CACHE_DIR=/abs/path` (default: user cache dir under `mistermorph/desktop/backend`)
- `MISTERMORPH_DESKTOP_WEBVIEW_GPU_POLICY=ondemand|always|never` (Linux only, default `ondemand`)

## Release packaging

Tag releases now build desktop release assets in GitHub Actions:

- macOS: `mistermorph-desktop-darwin-arm64.dmg`
- Linux: `mistermorph-desktop-linux-amd64.AppImage`
- Windows: `mistermorph-desktop-windows-amd64.zip`

The macOS DMG and Linux AppImage bundle a sibling `mistermorph` backend binary so the packaged app can launch `console serve` without a first-run download.
The Windows release bundle now includes both `MisterMorph.exe` and `mistermorph.exe`; keep them in the same directory after unzip.
The Windows release workflow also generates a `.ico` and Windows `.syso` resource on the runner so the published desktop executable carries the app icon.

If you want the same Windows executable icon in a local Windows build, run:

```bash
./scripts/generate-desktop-windows-resources.sh
```
