# 架构保守精简设计

> 日期：2026-07-09  
> 范围：方案 A（保留分层，清理空壳与未接线能力）

## 目标

在保留 `runtime` / `scheduler` / `statemachine` / `guard` 分层的前提下，删除死代码与空壳包，打通 `CheckReady` → idle 等待，收紧 config/UI 与 `TaskOpts`，使架构与当前单模块现实对齐。

## 已删除

- `internal/dialog`（占位未接线）
- `internal/utils`（无生产引用）
- `internal/hud`（仅转发 logger）
- `internal/game/common/kingdom/route.go`（无引用）
- `runtime.Register` 延迟注册回调
- 无消费者的 config 字段与 UI 死开关（`collectResources` / `farmLevels` / 旧 bot 字段）

## 行为修正

`scheduler.Build` 在提供 `CheckReady` 时自动 `AddIdleProvider`，使 runtime 的 `MaxIdleWait` 能按任务 `remain`（封顶 `IdleDelay`）等待。

## 保留

- 平台接口 + build tag 工厂
- 任务级状态机与 guard 机制
- `game/arena` 包结构与 `ui/widgets` 通用控件库
