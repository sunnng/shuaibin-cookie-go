# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an **AutoGo** Android automation project written in Go. AutoGo is a framework that compiles Go scripts into Android native binaries or APKs. The project targets Android only (`.vscode/settings.json` sets `AutoGo.targetPlatform: "android"`).

The root Go module (`module app`) is the user script/project. It depends on the local AutoGo SDK via a `replace` directive in `go.mod`:

```
replace github.com/Dasongzi1366/AutoGo => ./AutoGo
```

## Build & Run Workflow

Real deployment is driven by the **AutoGo JetBrains plugin**:

- Install the plugin in GoLand / IntelliJ IDEA.
- Connect an Android device or emulator (`adb devices` must list it).
- Use the plugin's run action (default shortcut **F7**) to compile and run the script on the device.
- Use the plugin's UI to package the script as an APK or binary.

The `resources/META-INF/Android.toml` file controls APK packaging options such as `appPackage`, `appName`, `autoRun`, and `showFloatingBall`.

For local development on non-Android hosts, the project uses `//go:build android` tags:

- Files that call AutoGo C/C++ bindings (OpenCV, screen capture, OCR) are Android-only.
- `internal/screen/factory.go` and `internal/action/factory.go` provide no-op stubs on other platforms so the project still type-checks and tests can run.

## Project Structure

```
.
├── main.go                      # AutoGo entry point: load config, wire detector/executor/states, run machine
├── config.json                  # User preferences (tick interval, modules, recovery limits)
├── go.mod                       # Module "app", replaces AutoGo to ./AutoGo
├── internal/
│   ├── action/                  # Touch/gesture/navigate abstractions + coordinate scaling
│   ├── bot/                     # Config, Context, State interface, Registry, Machine (state loop + watchdog)
│   ├── bot/states/              # Concrete State implementations (home, battle, unknown)
│   ├── screen/                  # Detector interface + Android wrappers for color/image/OCR
│   └── utils/                   # Small logging helpers
├── AutoGo/                      # Local AutoGo SDK (stub API surface)
├── resources/                   # Android packaging assets
└── docs/                        # AutoGo documentation export and design specs
```

## Key Architectural Notes

- **State machine**: `internal/bot/machine.go` runs a loop of `Detect → Act → Transition`. It also handles watchdog timeouts and recovery when no state is detected.
- **Interfaces for testability**: `screen.Detector` and `action.Executor` are interfaces. Unit tests in `internal/bot` mock them so the state machine can be tested without Android runtime.
- **Coordinate scaling**: `internal/action/coord.go` scales 1600×900 base coordinates to the actual device resolution and bounds the result.
- **Config**: `internal/bot/config.go` defines defaults; `config.json` overrides user preferences. Stable UI constants belong in state code, not JSON.
- **Cross-platform compilation**: `go build ./...` and `go test ./...` work on Windows because platform-specific AutoGo calls are guarded by build tags or live in stub factories.

## Common Commands

```bash
# Run all unit tests (works on Windows)
go test ./...

# Type-check / build all packages (works on Windows)
go build ./...

# Format all Go files
gofmt -w .

# Tidy modules after adding/removing dependencies
go mod tidy
```

Do not expect real Android behavior from local tests; they validate the state machine, coordinate math, and wiring only.

## Documentation

- Full AutoGo API reference: `docs/autogo-doc文档2026.6.6.md`.
- Design doc: `docs/superpowers/specs/2026-07-08-autogo-game-bot-design.md`.
- Implementation plan: `docs/superpowers/plans/2026-07-08-autogo-game-bot.md`.
- Progress ledger: `.superpowers/sdd/progress.md`.

When adding a new game state, create a file under `internal/bot/states/`, implement `bot.State`, register it in `main.go`, and place it before the `unknown` fallback in the registry order.
