# Arena Feature 布局设计

> 日期：2026-07-09  
> 状态：已定稿  
> 范围：模块 `feature.go` 的组织规范；以 `arena` 为样板，供后续游戏模块复用。

---

## 1. 背景与目标

### 1.1 问题

当前 `internal/game/arena/feature.go` 顶层混杂了真正的页面（`Lobby` / `TeamSelect` / `Settlement`）与非页面概念（`Opponent`、`Pagination`、`Dialog`）。填特征、对照 `page.go` 时边界不清。

### 1.2 目标

1. **顶层按真实页面组织**，页内再按用途分组（按钮 / OCR 等），便于「打开某页就知道有哪些控件」。
2. **`Dialogs` 挂模块顶层**，供 `page` 与 `guard` 共用，不绑死某一页。
3. **沉淀为后续模块通用约定**（矿山、广场等照抄骨架）。
4. **不改业务状态机与 `page` 对外接口**；本轮只重组常量归属与访问路径。

### 1.3 非目标

- 不在本轮填入真实坐标 / 比色串（仍后续从 Lua 特征库迁移）。
- 不改造 `statemachine` / `task` / `session` / `route` 流程。
- 不引入按「所有按钮一堆、所有 OCR 一堆」的顶层分类。

---

## 2. 通用规范（所有游戏模块）

### 2.1 原则

1. 顶层 = **真实页面** + 可选 **`Dialogs`**；不按控件类型（Button / OCR）做顶层拆分。
2. 页内固定槽位（**有则写，无则省略**，不硬凑空结构）：
   - `Identify` — 页面识别（多点比色等）
   - `Actions` — 可点击 / 可滑动的操作目标
   - `Reads` — OCR / 读数区域
   - 可选扩展：`Opponent`、`Gestures`、`Items` 等页面特有子结构
3. `feature` **只存常量**；识别与点击逻辑在 `page.go`。
4. 基准分辨率 **1600×900**；触控坐标经 `action` 缩放。
5. 访问路径：`feature.<Page>.Actions.<Name>` / `feature.<Page>.Reads.<Name>` / `feature.Dialogs.<Name>`。

### 2.2 模块骨架

```text
模块 Feature
├── <PageA>
│   ├── Identify
│   ├── Actions
│   ├── Reads          # 可省略
│   └── …可选子结构
├── <PageB>
│   └── …
└── Dialogs            # 可选
    └── …
```

### 2.3 与分层的关系

```text
feature.go  → 数据（坐标 / 比色 / OCR 区）
page.go     → 行为（IsX / TapY / ReadZ）
route.go    → 跨页导航
statemachine→ 业务流程
```

`page` 对外小接口保持稳定；状态机只依赖 `page` / `route`，不直接读 `Feature` 字段。

---

## 3. Arena 字段映射

### 3.1 顶层

```text
Feature
  Lobby
  TeamSelect
  Settlement
  Dialogs
```

原顶层 `Opponent`、`Pagination` 并入 `Lobby`；原 `Dialog` 改名为 `Dialogs`。

### 3.2 Lobby（大厅）

| 槽位 | 字段 | 来源（现状） | 用途 |
|------|------|--------------|------|
| Identify | `Identify` | `Lobby.Feature` | `IsLobby` |
| Actions | `Close` | `CloseBtn` | 关大厅 |
| Actions | `FreeRefresh` | `FreeRefreshTap` | 点免费刷新 |
| Actions | `BuyTicket` | `BuyTicketBtn` | 买票入口 |
| Actions | `BuyTicketSlider` | `BuyTicketSlider` | 买票滑条 |
| Actions | `BuyTicketConfirm` | `BuyTicketConfirm` | 买票确认 |
| Reads | `MedalTicket` | `MedalTicketOCR` | 勋章 + 门票 |
| Reads | `Trophy` | `TrophyOCR` | 我的奖杯 |
| Reads | `Refresh` | `RefreshOCR` | 刷新倒计时 |
| Reads | `FreeRefresh` | `FreeRefreshOCR` | 「免费刷新」文案 |
| Opponent | （整块） | 原 `OpponentFeature` | 列表项模板 |
| Gestures | `SwipeLeft` | 原 `Pagination.SwipeLeft` | 翻页 |

```go
type LobbyFeature struct {
	Identify screen.Feature
	Actions  LobbyActions
	Reads    LobbyReads
	Opponent OpponentFeature
	Gestures LobbyGestures
}
```

### 3.3 TeamSelect（备战）

| 槽位 | 字段 | 来源 | 用途 |
|------|------|------|------|
| Identify | `Identify` | `TeamSelect.Feature` | 是否在备战页 |
| Actions | `StartBattle` | `StartBattle` | 开战 |

无独立 OCR 时不建空的 `Reads`。

### 3.4 Settlement（结算）

| 槽位 | 字段 | 来源 | 用途 |
|------|------|------|------|
| Identify | `Identify` | `Settlement.Feature` | 结算页 |
| Actions | `Leave` | `LeaveBtn` | 离开 |
| Actions | （可选识别）`LeaveIdentify` | `LeaveFeature` | 离开按钮可见性 |
| Reads | `Result` | `ResultOCR` | 胜 / 平 / 负 |

### 3.5 Dialogs（顶层浮层）

| 字段 | 来源 | 用途 |
|------|------|------|
| `MissingTopping` | 原 `Dialog.MissingTopping` | 缺配料 |
| `DeployMore` | 原 `Dialog.DeployMore` | 再部署 |

每个为 `DialogDef{ Identify, Confirm }`（原 `Feature` 字段改名为 `Identify`，与页面槽位一致）。

### 3.6 page.go 访问路径对照

| 旧 | 新 |
|----|----|
| `feature.Lobby.MedalTicketOCR` | `feature.Lobby.Reads.MedalTicket` |
| `feature.Lobby.FreeRefreshTap` | `feature.Lobby.Actions.FreeRefresh` |
| `feature.Pagination.SwipeLeft` | `feature.Lobby.Gestures.SwipeLeft` |
| `feature.Opponent.*` | `feature.Lobby.Opponent.*` |
| `feature.Dialog.MissingTopping` | `feature.Dialogs.MissingTopping` |

---

## 4. 落地范围

### 4.1 修改

| 路径 | 动作 |
|------|------|
| `internal/game/arena/feature.go` | 按本文重组类型；`DefaultFeature()` 可仍为空占位 |
| `internal/game/arena/page.go` | 仅改字段访问路径，方法签名不变 |
| `docs/开发手册.md` | 增加「模块 Feature 约定」小节 |

### 4.2 不修改

- `task.go` / `statemachine.go` / `session.go` / `route.go` 的流程与接口
- `page` 小接口（`IsLobby`、`BuyTicket` 等）
- `platform/screen` 基础类型（除非实现时出现命名冲突，再单独处理）
- 真实特征数值填写

### 4.3 验收

- `go test ./internal/game/arena/...`
- `go build ./...`
- 若有直接引用 `Feature` 字段的测试，同步改路径

### 4.4 后续模块

新建模块时按第 2 节骨架写 `feature.go`；以本 spec 的 arena 映射为范例，不引入按钮/OCR 顶层分类。

---

## 5. 决策记录

| 决策 | 选择 | 理由 |
|------|------|------|
| 分类目标 | 可读性 + `page` 对齐 + 跨模块规范 | 用户确认 C |
| 弹窗位置 | 模块顶层 `Dialogs` | 跨页共用、便于 guard 注册（用户确认 B） |
| 页内分组 | `Identify` / `Actions` / `Reads` + 可选扩展 | 方案 2；避免顶层按控件类型拆 |
| 对手列表 / 翻页 | 并入 `Lobby` | 非独立页面 |

---

## 6. 未决事项

无。实现阶段按第 4 节执行即可。
