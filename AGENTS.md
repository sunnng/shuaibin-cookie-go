# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

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
- `internal/platform/screen/factory.go` and `internal/platform/action/factory.go` provide no-op stubs on other platforms so the project still type-checks and tests can run.

## Project Structure

```
.
├── main.go                      # AutoGo entry point: UI panel → wire detector/executor → runtime
├── config.json                  # User preferences (modules.arena)
├── go.mod                       # Module "app", replaces AutoGo to ./AutoGo
├── internal/
│   ├── config/                  # User config loader + static defaults
│   ├── game/                    # Game modules (arena, common/kingdom)
│   ├── guard/                   # Popup/interrupt guard with priority traps
│   ├── logger/                  # Level-based logging
│   ├── platform/
│   │   ├── action/              # Touch/gesture/navigate abstractions + coordinate scaling
│   │   └── screen/              # Detector interface + Android wrappers for color/image/OCR
│   ├── runtime/                 # Main runtime loop: guard → scheduler → idle wait
│   ├── scheduler/               # Task scheduler + TaskOpts registration helper
│   ├── statemachine/            # Reusable task-level state machine
│   ├── store/                   # JSON file-backed key-value store
│   └── ui/                      # ImGui config panel (android) / stub (desktop)
├── AutoGo/                      # Local AutoGo SDK (stub API surface)
├── resources/                   # Android packaging assets
└── docs/                        # AutoGo documentation export and design specs
```

## Key Architectural Notes

- **Runtime loop**: `internal/runtime/runtime.go` runs a forever loop of `guard.Check → scheduler.Run → idle wait`.
- **Task scheduling**: `internal/scheduler/scheduler.go` polls registered tasks by condition. `internal/scheduler/builder.go` provides `TaskOpts`; when `CheckReady` is set, `Build` also registers an idle provider so runtime can wait on `remain`.
- **Task-level state machines**: each game module (e.g. `internal/game/arena`) owns an `internal/statemachine.Machine` for its internal flow (`detect → navigate → sync → ...`).
- **Interfaces for testability**: `platform/screen.Detector` and `platform/action.Executor` are interfaces. Unit tests mock them so the scheduler, runtime, guard, and state machines can be tested without Android runtime.
- **Cross-platform compilation**: `go build ./...` and `go test ./...` work on Windows because platform-specific AutoGo calls are guarded by build tags or live in stub factories.
- **Build-tag split**: packages `platform/screen` and `platform/action` each have a `factory.go` with `//go:build !android` that returns stubs, and a `factory_android.go` with `//go:build android` that returns real `AndroidDetector` / `AndroidExecutor` implementations.
- **Coordinate scaling**: `internal/platform/action/coord.go` scales 1600×900 base coordinates to the actual device resolution and bounds the result.
- **Config**: `internal/config/static.go` defines defaults and `internal/config/user.go` loads `config.json`. Missing `config.json` falls back to `DefaultConfig()`. Stable UI constants belong in module code, not JSON.
- **Guard**: `internal/guard/guard.go` registers popup traps by priority and provides segmented sleep that checks traps periodically.
- **Store**: `internal/store/store.go` persists small JSON-backed state (e.g. arena next free refresh time).

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
- Design doc: `docs/superpowers/specs/2026-07-08-arena-module-design.md`.
- Implementation plan: `docs/superpowers/plans/2026-07-08-arena-module.md`.
- Progress ledger: `.superpowers/sdd/progress.md`.

When adding a new game module, create a package under `internal/game/`, follow the `feature.go` / `page.go` / `route.go` / `session.go` / `task.go` / `statemachine.go` convention, register its task with `sched.Build(scheduler.TaskOpts{...})` in `main.go` before `rt.Run()`, and place it before any fallback task in scheduler order.
