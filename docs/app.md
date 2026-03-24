# Desktop App

Mister Morph ships a desktop App that wraps the existing Console backend and UI into a single local experience.

## User Quick Start

Download a release asset from the [GitHub Releases](https://github.com/quailyquaily/mistermorph/releases) page:

- macOS `arm64`: `mistermorph-desktop-darwin-arm64.dmg`
- Linux `amd64`: `mistermorph-desktop-linux-amd64.AppImage`
- Windows `amd64`: `mistermorph-desktop-windows-amd64.zip`

Then:

1. launch the App
2. complete the setup flow inside the App
3. let the App start and host the local Console for you

You do not need to run `mistermorph console serve` manually when using the App.

## Current Shape

The desktop code lives under `desktop/wails` and is built with `wailsdesktop production`.

- the Wails process owns the native window and Go bindings
- a child process runs `mistermorph console serve`
- the App routes WebView traffic to that local child process

The wrapper stays intentionally thin: lifecycle, process hosting, restart flow, and local proxying only.

## Architecture

```text
+--------------------------------------------------------------+
| Desktop App Process (Wails)                                  |
|                                                              |
|  +----------------------+        +------------------------+  |
|  | WebView (UI)         | <----> | Reverse Proxy Handler  |  |
|  | path: /console/*     |  HTTP  | in desktop/wails host  |  |
|  +----------------------+        +------------------------+  |
|             ^                              |                 |
|             | JS bridge                    v                 |
|  +----------------------+        +------------------------+  |
|  | App binding          |        | Child Process Manager  |  |
|  | RestartApp()         |        | start/stop/wait/health |  |
|  +----------------------+        +------------------------+  |
+----------------------------------------------|---------------+
                                               |
                                               | spawn
                                               v
                           +----------------------------------+
                           | Child: `mistermorph console serve` |
                           | listen: 127.0.0.1:<random>       |
                           | base path: /console              |
                           | allow-empty-password: enabled    |
                           +----------------------------------+
```

## Startup and Restart

Startup sequence:

```text
desktop main
  -> resolve backend binary path
  -> reserve random loopback port
  -> spawn child: mistermorph console serve
  -> poll GET /health until ready
  -> open native window
  -> proxy requests to the child process
```

First run:

```text
incomplete config
  -> console backend starts with allow-empty-password
  -> frontend routes to /setup
  -> user saves agent settings + identity/soul
  -> frontend calls App.RestartApp()
```

`App.RestartApp()` starts a new copy of the current desktop executable and then quits the old one.

## Paths and Configuration

- Console frontend assets are embedded in the bundled `mistermorph` backend by default.
- You can override static assets with:
  - `console.static_dir`
  - `--console-static-dir /abs/path/to/dist`
- `--config <path>` passed to the desktop App is forwarded to the child `console serve` process.

## Local Build and Run

On Ubuntu or Debian, install desktop build dependencies first:

```bash
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev
```

Build the backend binary used by the desktop wrapper:

```bash
./scripts/build-backend.sh --output ./bin/mistermorph
```

Build a local desktop release binary:

```bash
./scripts/build-desktop.sh --release
```

Run from source:

```bash
go run -tags 'wailsdesktop production' ./desktop/wails
```

Build only the desktop wrapper directly:

```bash
go build -tags 'wailsdesktop production' -o ./bin/mistermorph-desktop ./desktop/wails
```

For local debug builds with DevTools, use:

```bash
./scripts/build-desktop.sh
```

## Release Packaging

Tagged releases currently publish:

- macOS `arm64`: `mistermorph-desktop-darwin-arm64.dmg`
- Linux `amd64`: `mistermorph-desktop-linux-amd64.AppImage`
- Windows `amd64`: `mistermorph-desktop-windows-amd64.zip`

The packaged desktop App includes a sibling `mistermorph` backend binary so the wrapper can start `console serve` locally without a first-run download.

That bundled backend is intentionally built with `CGO_ENABLED=0`. Keep it that way unless there is a deliberate packaging plan to change the constraint.

## Known Gaps

- No notarization or codesign flow yet for the macOS DMG.
- Windows ships as a zip bundle, not an installer.
- No dedicated UI yet for detailed backend startup failures.
- The desktop wrapper still reuses the CLI backend through child-process orchestration rather than an in-process console module.
