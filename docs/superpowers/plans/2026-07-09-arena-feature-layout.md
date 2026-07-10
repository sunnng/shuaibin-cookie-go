# Arena Feature 布局 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按 `2026-07-09-arena-feature-layout-design.md` 重组 arena `feature.go`，并同步 `page.go` 访问路径与开发手册约定。

**Architecture:** 顶层按真实页面（Lobby / TeamSelect / Settlement）+ `Dialogs`；页内 `Identify` / `Actions` / `Reads`（及 Lobby 的 Opponent / Gestures）。不改状态机与 `page` 方法签名。

**Tech Stack:** Go，现有 `internal/game/arena`、`docs/开发手册.md`

## Global Constraints

- 基准分辨率 1600×900
- `feature` 只存常量；行为在 `page.go`
- 不填真实坐标/比色串
- 不改 `task` / `statemachine` / `session` / `route` 流程与 `page` 小接口
- Dialogs 挂模块顶层

---

### Task 1: 重组 feature.go + 更新 page.go 路径

**Files:**
- Modify: `internal/game/arena/feature.go`
- Modify: `internal/game/arena/page.go`
- Test: `internal/game/arena/page_test.go`（已有 `DefaultFeature()` 冒烟）

**Interfaces:**
- Produces: `Feature{Lobby, TeamSelect, Settlement, Dialogs}`；`LobbyFeature{Identify, Actions, Reads, Opponent, Gestures}` 等（见下方完整类型）

- [ ] **Step 1: 重写 `feature.go` 为下列完整内容**

```go
package arena

import (
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

type Feature struct {
	Lobby      LobbyFeature
	TeamSelect TeamSelectFeature
	Settlement SettlementFeature
	Dialogs    DialogsFeature
}

type LobbyFeature struct {
	Identify screen.Feature
	Actions  LobbyActions
	Reads    LobbyReads
	Opponent OpponentFeature
	Gestures LobbyGestures
}

type LobbyActions struct {
	Close            screen.Region
	FreeRefresh      screen.Point
	BuyTicket        screen.Point
	BuyTicketSlider  action.Swipe
	BuyTicketConfirm screen.Point
}

type LobbyReads struct {
	MedalTicket screen.Region
	Trophy      screen.Region
	Refresh     screen.Region
	FreeRefresh screen.Region
}

type LobbyGestures struct {
	SwipeLeft action.Swipe
}

type OpponentFeature struct {
	FindDef      screen.FindDef
	BaseSite     screen.Point
	TrophyRect   screen.Region
	ResultOffset screen.Point
	ResultColors ResultColors
	NumberOCR    screen.OCRCfg
}

type ResultColors struct {
	Win  string
	Draw string
	Lose string
}

type TeamSelectFeature struct {
	Identify screen.Feature
	Actions  TeamSelectActions
}

type TeamSelectActions struct {
	StartBattle screen.Point
}

type SettlementFeature struct {
	Identify screen.Feature
	Actions  SettlementActions
	Reads    SettlementReads
}

type SettlementActions struct {
	LeaveIdentify screen.Feature
	Leave         screen.Rect
}

type SettlementReads struct {
	Result screen.Rect
}

type DialogsFeature struct {
	MissingTopping DialogDef
	DeployMore     DialogDef
}

type DialogDef struct {
	Identify screen.Feature
	Confirm  screen.Point
}

func DefaultFeature() *Feature {
	return &Feature{}
}
```

- [ ] **Step 2: 更新 `page.go` 中所有 feature 字段访问**

替换为：
- `p.feature.Lobby.Reads.MedalTicket`
- `p.feature.Lobby.Reads.Trophy`
- `p.feature.Lobby.Reads.FreeRefresh`
- `p.feature.Lobby.Reads.Refresh`
- `p.feature.Lobby.Actions.FreeRefresh`
- `p.feature.Lobby.Actions.BuyTicket`
- `p.feature.Lobby.Actions.BuyTicketSlider`
- `p.feature.Lobby.Actions.BuyTicketConfirm`
- `p.feature.Lobby.Gestures.SwipeLeft`
- 注释改为 `feature.Lobby.Opponent`

- [ ] **Step 3: 跑测试与编译**

Run: `go test ./internal/game/arena/...`
Expected: PASS

Run: `go build ./...`
Expected: success

- [ ] **Step 4: Commit**（仅当用户要求时）

---

### Task 2: 开发手册增加 Feature 约定

**Files:**
- Modify: `docs/开发手册.md`（§6.1 建包之后插入新小节）

- [ ] **Step 1: 在 §6.1 与 §6.2 之间插入「模块 Feature 约定」**

内容要点：
- 顶层 = 真实页面 + 可选 Dialogs
- 页内 Identify / Actions / Reads；可选 Opponent、Gestures、Items
- 无则省略；feature 只存常量
- 访问路径示例；指向 `docs/superpowers/specs/2026-07-09-arena-feature-layout-design.md`

- [ ] **Step 2: Commit**（仅当用户要求时）

---

## Spec coverage

| Spec 要求 | Task |
|-----------|------|
| 通用规范 Identify/Actions/Reads/Dialogs | Task 1 + 2 |
| Arena 字段映射 | Task 1 |
| page.go 路径更新 | Task 1 |
| 开发手册约定 | Task 2 |
| 不改状态机/接口 | 遵守 Global Constraints |
