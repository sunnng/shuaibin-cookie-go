# 矿山分类 UI 骨架 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 任务面板新增「矿山」分类与四个仅含启用开关的占位任务，不注册调度器。

**Architecture:** 仿 `ui_arena.go`：`config` 增加四个 `Enabled` 结构；`mineTaskDescriptors` 返回四个 `ui.Task`（`Category: "矿山"`）；`main` 拼进 `Tasks`；`buildRuntime` 不动。

**Tech Stack:** Go、现有 `ui` 描述符框架、`internal/config`

**Spec:** `docs/superpowers/specs/2026-07-27-mine-category-ui-skeleton-design.md`

## Global Constraints

- 不调用 `sched.Build` 注册矿山任务
- 每个任务仅一个 `ui.Bool`「启用」；Summary 固定 `"未实现"`；默认 `Enabled: false`
- ID / EnabledKey / config JSON 键以 spec 表为准，勿改名
- 不新建 `internal/game/mine*`；不改分类色条
- 仅 gofmt 本次修改的文件；未经用户要求不 commit

---

### Task 1: Config 四个 Enabled 结构

**Files:**
- Modify: `internal/config/static.go`
- Modify: `internal/config/config_test.go`

**Interfaces:**
- Produces: `OreVeinMining` / `MineVenture` / `MineBattle` / `MeltedAgarCubes`（各 `Enabled bool`）；`TaskConfig` 增加对应字段与 JSON 标签；`DefaultConfig` 四者 `Enabled: false`

- [ ] **Step 1: 写失败测试** — `TestDefaultConfig` 断言四任务 `Enabled == false`
- [ ] **Step 2: 实现 static.go 结构与 DefaultConfig**
- [ ] **Step 3: `go test ./internal/config/ -count=1` 通过**

---

### Task 2: UI 描述符 + Seed/Apply 单测

**Files:**
- Create: `ui_mine.go`
- Create: `ui_mine_test.go`

**Interfaces:**
- Produces: `func mineTaskDescriptors(cfg *config.Config) []ui.Task`（长度 4，顺序：开采→勘查→战斗→洋菜冻）

- [ ] **Step 1: 写 `TestMineDescriptorsSeedApplyRoundTrip` 与 `TestMineDescriptorsSummaryAndCategory`**
- [ ] **Step 2: 实现 `mineTaskDescriptors`**
- [ ] **Step 3: `go test . -run Mine -count=1` 通过**

---

### Task 3: main 挂载

**Files:**
- Modify: `main.go`（`tasks` 切片拼接）

- [ ] **Step 1: `tasks := append([]ui.Task{arenaTaskDescriptor(cfg)}, mineTaskDescriptors(cfg)...)`**
- [ ] **Step 2: `go test ./... -count=1` 与 `go build ./...` 通过**
- [ ] **Step 3: 确认 `buildRuntime` 无矿山 `sched.Build`**

---

## Spec coverage

| Spec 项 | Task |
|---|---|
| 四个任务命名与键 | 1–2 |
| Category 矿山 / Summary 未实现 | 2 |
| main 挂载、不调度 | 3 |
| 验收 go test | 3 |
