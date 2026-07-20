# 帅宾 Cookie（AutoGo）

Android 游戏自动化脚本（Go + AutoGo）。当前业务：王国竞技场。

> **权威说明以 [`docs/开发手册.md`](docs/开发手册.md)、[`CLAUDE.md`](CLAUDE.md) / [`AGENTS.md`](AGENTS.md) 为准。**  
> 本文档仅作快速入口；旧版「全局 bot 状态机 / `internal/bot`」架构已删除。

## 架构一句话

`main` → `ui.RunShell`（灵动岛 + 面板）→ `SessionController` → `buildRuntime` →  
`runtime` 循环：`guard.Check` → `scheduler.Run` → idle；页面流程在各模块任务级 `statemachine` 中。

## 快速开始

1. 用 AutoGo JetBrains 插件打开本项目，连接设备（`adb devices`）。
2. 按 **F7** 运行；面板改配置后点运行脚本。
3. 本地校验（Windows 可用 stub 路径）：

```bash
go test ./...
go build ./...
```

## 关键路径

| 路径 | 说明 |
|------|------|
| `main.go` | 组合根：配置、UI、模块装配 |
| `internal/game/arena/` | 竞技场模块 |
| `internal/status/` | 任务状态上报通道（灵动岛显示战斗次数/胜率） |
| `internal/ui/` | ImGui 壳（灵动岛悬浮窗 + 配置面板）、SessionController、配置 binding |
| `internal/runtime/` | 主循环 Pause/Resume/Stop |
| `config.json` | 仓库默认配置（面板状态另存 `/sdcard/shuaibin-cookie/ui.json`） |

## 文档

- [开发手册](docs/开发手册.md)
- [AutoGo API 参考](docs/autogo-doc文档2026.6.6.md)
- 设计稿与计划：`docs/superpowers/`（部分为历史稿，以手册为准）
