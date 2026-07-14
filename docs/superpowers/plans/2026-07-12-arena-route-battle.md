# 竞技场路由 + 最小战斗 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按 [2026-07-12-arena-route-battle-design.md](../specs/2026-07-12-arena-route-battle-design.md) 落地段1 进出路由与段3 最小战斗闭环；段2 为取色清单（已在 spec，无代码强制）。

**Architecture:** 平台层新增 `FindOCRText`；`kingdom` 按 Identify/Actions 实现页识别；`arena.route` 补全 Enter/Leave；`teamSelect`/`RunBattle` 去掉硬闸与假胜。数值仍空，靠 mock 测。

**Tech Stack:** Go、AutoGo ppocr/images、现有 `screen.Detector` / `action.Executor`、arena 任务级 statemachine。

## Global Constraints

- `DefaultFeature` 不代填比色/坐标数值
- 不接 Guard / Dialogs
- 不新增独立 settlement SM 态
- `(0,0)` Point = 未配置；Identify.Colors=="" → Is* false
- 基准分辨率 1600×900
- TDD：先写失败测试再实现
- 仅 gofmt 自己改过的文件

---

## File map

| 文件 | 职责 |
|------|------|
| `internal/platform/screen/detector.go` | `FindOCRText` 接口 |
| `internal/platform/screen/ocr.go` | android 实现 |
| `internal/platform/screen/factory.go` | stub |
| `internal/game/common/kingdom/*` | feature 槽位 + page 实现 + tests |
| `internal/game/arena/feature.go` | `EntryFeature` |
| `internal/game/arena/page.go` | `TapEntry` / `TapToLobby` / 段3 RunBattle |
| `internal/game/arena/route.go` | Enter/Leave |
| `internal/game/arena/statemachine.go` | teamSelect 真实逻辑 |
| mocks in `*_test.go` / guard/runtime | 补 `FindOCRText` |

---

### Task 1: FindOCRText on Detector

**Files:** `detector.go`, `ocr.go`, `factory.go`, `factory_test.go`, update all Detector mocks

- [ ] 接口加 `FindOCRText(region Region, keyword string) (Point, bool)`
- [ ] stub 返回 `(Point{}, false)`；空 keyword 亦 false
- [ ] android：截屏 PPOCR，Label contains keyword → 中心点 + region 偏移
- [ ] 单测 stub；更新 guard/runtime/arena page mocks
- [ ] `go test ./internal/platform/screen/ ./internal/guard/ ./internal/runtime/`

### Task 2: kingdom feature + page

**Files:** `kingdom/feature.go`, `kingdom/page.go`, `kingdom/page_test.go` (new)

- [ ] 重写 Feature 为 Home/Adventure PageSlot + Actions
- [ ] Is* 用 MatchMultiColor；空 Colors → false
- [ ] TapAdventureBtn 点 Actions.AdventureBtn
- [ ] mock 单测
- [ ] `go test ./internal/game/common/kingdom/`

### Task 3: arena Entry + TapEntry + TapToLobby + route

**Files:** `arena/feature.go`, `page.go`, `route.go`, `route_test.go` (new), page/task mocks if needed

- [ ] Feature 加 Entry；Keyword 空默认「王国竞技场」
- [ ] TapEntry / TapToLobby 按 spec
- [ ] Enter/Leave 按 spec
- [ ] route 单测（注入 mock page/kingdom 或通过接口——保持 Route 用具体类型则测真实 Page+mock Detector）
- [ ] `go test ./internal/game/arena/`
- [ ] 保持 teamSelect 硬闸至 Task 4

### Task 4: teamSelect + RunBattle

**Files:** `statemachine.go`, `page.go`, `task_test.go`, `page_test.go`

- [ ] 实现 teamSelect；Identify.Colors=="" 跳过等待
- [ ] RunBattle 真实逻辑；禁止固定「胜利」
- [ ] 更新/替换硬闸测试
- [ ] `go test ./...`

### Task 5: 文档勾一点

- [ ] 开发手册若有 kingdom 结构描述可补一句（可选、最小）

---

**段2：** 无实现任务；用户按 spec §2 取色。
