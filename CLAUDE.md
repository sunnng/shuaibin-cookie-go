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

For local development on non-Android hosts, the project uses build tags:

- Files that call AutoGo C/C++ bindings (imgui, OpenCV, screen capture, OCR) carry `//go:build android && cgo` — they are only compilable with the plugin's NDK toolchain, never on a plain Windows/Mac Go install.
- Stub factories carry `//go:build !android || !cgo`: `internal/platform/screen/factory.go`, `internal/platform/action/factory.go`, and `internal/ui/panel_stub.go` provide no-op implementations so the project type-checks and tests run anywhere.

## Project Structure

```
.
├── main.go                      # Entry: config → ui.RunShell (island + panel) → SessionController builds runtime on Start
├── config.json                  # User preferences (modules.arena)
├── go.mod                       # Module "app", replaces AutoGo to ./AutoGo
├── internal/
│   ├── config/                  # Config type + LoadConfig + DefaultConfig
│   ├── game/                    # Game modules (arena; common/kingdom shared pages)
│   ├── guard/                   # Popup/interrupt guard with priority traps
│   ├── logger/                  # Level-based logging
│   ├── platform/
│   │   ├── action/              # Touch/gesture/navigate abstractions + coordinate scaling
│   │   └── screen/              # Detector interface + Android wrappers for color/image/OCR
│   ├── runtime/                 # Main runtime loop: guard → scheduler → idle wait; Pause/Resume/Stop
│   ├── scheduler/               # Task scheduler + TaskOpts registration helper
│   ├── statemachine/            # Reusable task-level state machine
│   ├── status/                  # One-line task-status reporter: task goroutine writes, island pill reads
│   ├── store/                   # JSON file-backed key-value store (device path: ui.DefaultStorePath)
│   └── ui/                      # ImGui shell: Dynamic-Island capsule, config panel, SessionController, config binding (android&&cgo) / stub (desktop)
├── AutoGo/                      # Local AutoGo SDK (Go API surface; imgui/opencv packages use cgo)
├── assets/tpl/                  # Template images for MatchImage
├── build/                       # Plugin build artifacts (APKs) — gitignored
├── resources/                   # Android packaging assets
└── docs/                        # AutoGo API reference, dev manual, design specs/plans
```

## Key Architectural Notes

- **Runtime loop**: `internal/runtime/runtime.go` runs a forever loop of `guard.Check → scheduler.Run → idle wait`. Idle sleep uses `scheduler.MinIdleWait()` (earliest `CheckReady` remain across tasks) capped by `IdleDelay`; after work it sleeps `StepDelay`. Supports `Pause`/`Resume`/`Stop`, driven by the UI controller.
- **Task scheduling**: `internal/scheduler/scheduler.go` polls registered tasks in registration order. `internal/scheduler/builder.go` provides `TaskOpts`; when `CheckReady` is set, `Build` also registers an idle provider so runtime can wait on `remain`. Register all tasks **before** `rt.Run()` — later `Build` calls are never scheduled.
- **Task-level state machines**: each game module (e.g. `internal/game/arena`) owns an `internal/statemachine.Machine` for its internal flow (`detect → navigate → sync → check → selectOpponent → teamSelect → battle → leave`). Handlers return `Keep` / `Retry` (bounded by MaxRetry) / `Next("state")` / `Done` / `Fatal{Err}`; a task can hook `RunOptions.Guard = guard.Check` so popups interrupt its ticks. Do not reintroduce a global bot state machine — page flow lives in each module.
- **UI shell & config binding**: on Android, `ui.RunShell` runs the ImGui loop (Dynamic-Island style capsule always drawn + QQ-blue themed config panel with a custom gradient title bar; theme lives in `theme_android.go` — `ApplyQQBlueTheme`). `SessionController` (Idle/Running/Paused) starts the bot in a goroutine and rebuilds the runtime on every Start. Config flows one way: `config.json` → `SeedFromConfig` (fills only keys missing from the device store at `/sdcard/shuaibin-cookie/ui.json`) → panel widgets mutate the `ui.Store` → on Start, `ApplyToConfig` writes the store back into `config.Config` → `buildRuntime(cfg, statusReporter)`. Adding a user setting means four edits that must use the **same key**: a `KeyXxx` const in `internal/ui/binding.go`, `SeedFromConfig`, `ApplyToConfig`, and the widget in `cookie_panel_android.go` — otherwise panel changes silently don't take effect. Panel is list+detail with categories over a compile-time `BuiltinModules()` table — do **not** reintroduce a global `RegisterModule` registry. The overlay is captured by `CaptureScreen` along with the game and cannot be excluded from it, so the policy is "don't recognize while occluded": opening the panel auto-pauses a running script and closing it auto-resumes (`autoPaused` in `shell_android.go`), the expanded island card auto-collapses after ~6s idle, and recognition regions in feature files must avoid the island's top-center strip (1600-base y≤80; expanded card y≤210, x≈520–1080). Task status reporting: modules push a one-line status (arena: battles/win-rate, from `Session.StatusText`) through `internal/status.Reporter`; `main.go` shares one instance between `ShellOptions.Status` and `Task.SetStatusReporter`, and the island pill shows it in place of the default label while Running. See `docs/开发手册.md` §6.4.
- **Interfaces for testability**: `platform/screen.Detector` and `platform/action.Executor` are interfaces. Modules define narrow consumer-side interfaces (arena's `page` / `route` in `task.go`) so tests inject mocks via unexported constructors (`newTask`) and drive machines with short intervals (`runWithOptions`). Unit tests mock them so the scheduler, runtime, guard, and state machines can be tested without Android runtime.
- **Cross-platform compilation**: `go build ./...` and `go test ./...` work on Windows because platform-specific AutoGo calls are guarded by build tags or live in stub factories.
- **Build-tag split**: packages `platform/screen` and `platform/action` each have a `factory.go` with `//go:build !android || !cgo` that returns stubs, and a `factory_android.go` with `//go:build android && cgo` that returns real `AndroidDetector` / `AndroidExecutor` implementations. The real implementations are split across files (e.g., `color.go`, `image.go`, `ocr.go`, `tap.go`, `navigate.go`) and are only compiled for Android. When extending the platform layer: extend the interface first, then the Android impl, then the stub — business code depends only on the interface.
- **Coordinate scaling**: `internal/platform/action/coord.go` scales 1600×900 base coordinates to the actual device resolution and bounds the result. `AndroidExecutor.Tap`/`Swipe` call `action.SafeTap` using the live display size.
- **Config**: `internal/config/static.go` defines defaults and `internal/config/user.go` loads `config.json`. Missing `config.json` falls back to `DefaultConfig()`. Stable UI constants belong in module code, not JSON.
- **Guard**: `internal/guard/guard.go` registers popup traps (typed `screen.Feature`) by priority; runtime and task state machines poll `guard.Check` periodically.
- **Store**: `internal/store/store.go` persists small JSON-backed state (e.g. arena next free refresh time).

## Common Commands

```bash
# Run all unit tests (works on Windows)
go test ./...

# Run tests verbosely
go test ./... -v

# Run a single test
go test ./internal/statemachine -run TestMachineNextTransition

# Run tests for one package
go test ./internal/game/arena/...

# Type-check / build all packages (works on Windows)
go build ./...

# The real Android files (`//go:build android && cgo`) can only be compiled with the
# AutoGo plugin's NDK toolchain — there is no working local type-check for them.
# `go build -tags android ./...` FAILS on Windows (stdlib "GOOS redeclared"), expected.
# Closest local sanity check (still takes the stub path since cgo is off):
GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...

# Format files you touched (see caveat below — do NOT blanket-run on this checkout)
gofmt -w <file>

# Check formatting without writing
gofmt -l .

# Run basic static analysis
go vet ./...

# Tidy modules after adding/removing dependencies
go mod tidy
```

Do not expect real Android behavior from local tests; they validate the state machine, coordinate math, and wiring only.

**gofmt caveat on this Windows host**: the checkout uses `core.autocrlf=true`, so working-copy files are CRLF and `gofmt -l` flags most of the tree. Some committed files are also genuinely unformatted. Only gofmt files you actually modify — a blanket `gofmt -w .` produces a huge spurious diff.

## Documentation

- **Dev manual (authoritative, Chinese)**: `docs/开发手册.md` — detailed conventions: module creation, UI binding, platform layer, testing, pitfalls (§13), pre-PR checklist (§15).
- Full AutoGo API reference: `docs/autogo-doc文档2026.6.6.md`.
- Design specs: `docs/superpowers/specs/` (arena module, feature layout, recognition, floating ball, architecture simplify, ...).
- Implementation plans: `docs/superpowers/plans/`.
- Progress ledger: `.superpowers/sdd/progress.md` (task-based SDD workflow: briefs, reports, per-commit review diffs).
- **README.md** is a quick entry point consistent with this file; `docs/开发手册.md` remains the authoritative deep reference.

## Adding a Game Module

1. Create a package under `internal/game/`, following the convention: `feature.go` (UI constants only — regions/points/color strings at 1600×900 base; organized per page with `Identify`/`Actions`/`Reads` slots, see `internal/game/arena/feature.go`), `page.go`, `route.go`, `session.go`, `task.go`, `statemachine.go`. Shared navigation goes under `internal/game/common/`.
2. Add config fields in `internal/config/static.go` (+ `DefaultConfig`, `config_test.go`). **Do not add config/UI for unimplemented modules.**
3. Wire the UI: key const + `SeedFromConfig` + `ApplyToConfig` in `internal/ui/binding.go`, widget in `cookie_panel_android.go` (same key in all four places).
4. Register with `sched.Build(scheduler.TaskOpts{...})` in `main.go`'s `buildRuntime`, before `rt.Run()`. Scheduling order = `Build` call order; put urgent tasks first, fallbacks last. Optional: to show task status on the island, call `task.SetStatusReporter(statusReporter)` in `buildRuntime` and push one-line text from a loop node (see "Task status reporting" in the UI-shell note above).
5. Cover at least: ready/not-ready, main-path state transitions, key failure branches (no tickets, max reached).

## Working Notes

- Task names and many log messages are Chinese (e.g. `"王国竞技场"`); keep the `[Module]` log prefix convention (`[Runtime]`, `[Arena]`, `[Guard]`).
- `.worktrees/` holds git worktrees for feature branches and is gitignored.
- `AGENTS.md` mirrors this file for Codex — keep the two in sync when editing either.
