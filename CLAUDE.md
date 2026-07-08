# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an **AutoGo** Android automation project written in Go. AutoGo is a framework that compiles Go scripts into Android native binaries or APKs. The project currently targets Android only (`.vscode/settings.json` sets `AutoGo.targetPlatform: "android"`).

The root Go module (`module app`) is the user script/project. It depends on the local AutoGo SDK via a `replace` directive in `go.mod`.

## Build & Run Workflow

This project is **not built with standard `go build`**. The AutoGo SDK under `./AutoGo` contains stub declarations; the real runtime implementations are injected by the AutoGo runtime when the script is deployed to a device.

Development is driven by the **AutoGo JetBrains plugin**:

- Install the plugin in GoLand / IntelliJ IDEA.
- Connect an Android device or emulator (`adb devices` must list it).
- Use the plugin's run action (default shortcut **F7**) to compile and run the script on the device.
- Use the plugin's UI to package the script as an APK or binary.

The `resources/META-INF/Android.toml` file controls APK packaging options such as `appPackage`, `appName`, `autoRun`, and `showFloatingBall`.

## Module Structure

```
.
├── go.mod              # Root module "app", replaces AutoGo to ./AutoGo
├── AutoGo/             # Local AutoGo SDK (stub API surface)
│   ├── app/            # Application / Intent APIs
│   ├── device/         # Device info, battery, volume, display
│   ├── files/          # File system helpers
│   ├── images/         # Screenshot, color matching, image ops
│   ├── motion/         # Touch, gestures, key events
│   ├── uiacc/          # Android Accessibility node queries/actions
│   ├── utils/          # Logging, shell, dialogs, conversions
│   ├── opencv/         # OpenCV bindings
│   ├── imgui/          # Immediate-mode UI drawing
│   ├── console/        # Floating console window
│   ├── hud/            # On-screen HUD
│   ├── vdisplay/       # Virtual display
│   ├── plugin/         # External Android plugin interop
│   ├── rhino/          # JS engine integration
│   ├── ppocr/          # PaddleOCR
│   ├── yolo/           # YOLO object detection
│   └── ...
├── resources/          # Android packaging assets
│   ├── META-INF/Android.toml
│   ├── libs/           # Native libraries per ABI
│   ├── assets/
│   └── ui/index.html
└── docs/               # AutoGo documentation export
```

## Key Architectural Notes

- **Stub SDK**: Files under `AutoGo/` define the public API surface with no-op implementations. They exist so the project compiles in the IDE and provides autocomplete. Do not expect unit tests that exercise real Android behavior to pass locally.
- **No `main.go` / no tests**: The project root currently has no Go source files and no `*_test.go` files. The plugin supplies the entry point and runtime when deploying.
- **Target platform**: Android. Keep Android in mind when modifying code; iOS APIs are documented but not active in this workspace.

## Common Commands

Because the SDK is stubbed, standard Go commands have limited usefulness here:

- `go test ./...` — reports no tests because none exist.
- `go build ./...` — may fail for packages that depend on CGo or platform-specific symbols not present in this checkout; rely on the AutoGo plugin for real builds.
- `go mod tidy` — safe to run when adding/removing dependencies.

## Documentation

The full AutoGo API reference is in `docs/autogo-doc文档2026.6.6.md`. Consult it for function signatures, parameter formats (e.g., color strings like `"FFFFFF|CCCCCC-101010"`), and platform-specific behavior.
