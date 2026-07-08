# 王国竞技场模块设计文档

> 日期：2026-07-08  
> 参考项目：`D:\BaiduSyncdisk\项目\项目_帅斌饼干助手\帅斌饼干\脚本\game\常规_王国竞技场`  
> 目标：基于 AutoGo Go 项目，按现有 Lua 项目架构范式重新设计并接入「王国竞技场」模块。

---

## 1. 背景与目标

### 1.1 当前问题

当前 Go 项目只有一个极简的主状态机（`Home` / `Battle` / `Unknown`），没有体现 Lua 项目中成熟的「调度器 + 任务级状态机」分层。随着业务模块增多，职责会越来越散乱：

- 全局状态机和任务内部流程混在一起。
- 没有统一的任务注册、就绪检查、空闲等待机制。
- 弹窗处理、配置加载、持久化等基础能力没有独立抽象。

### 1.2 设计目标

1. **复刻 Lua 项目架构**：把 `core.scheduler`、`core.state-machine`、`core.guard`、`lib.store`、`lib.user-config` 等概念映射到 Go。
2. **接入王国竞技场**：作为第一个完整业务模块，验证新架构的可行性。
3. **保持可测试性**：所有平台相关操作通过 interface 抽象，本地可跑单元测试。
4. **不引入过度设计**：先支持竞技场一个模块，其他模块后续按同样模式扩展。

---

## 2. 整体架构重设计

### 2.1 Lua → Go 包映射

| Lua 模块 | Go 包 | 说明 |
|---|---|---|
| `core.scheduler` | `internal/scheduler` | 任务轮询 |
| `core.state-machine` | `internal/statemachine` | 任务内部状态机 |
| `core.guard` | `internal/guard` | 全局弹窗守卫 |
| `core.runtime` | `internal/runtime` | 大循环入口 |
| `lib.store` | `internal/store` | 本地持久化 |
| `lib.user-config` | `internal/config` | 用户配置加载与默认值 |
| `lib.logger` | `internal/logger` | 分级日志 |
| `lib.status-hud` | `internal/hud` | 状态浮层 |
| `lib.color` / `lib.ocr` | `internal/platform/screen` | 屏幕识别 |
| `lib.touch` | `internal/platform/action` | 触控执行 |
| `lib.dialog` | `internal/dialog` | 通用弹窗处理 |
| `game.task-builder` | `internal/scheduler/builder.go` | 任务注册模板 |
| `game/通用_王国` | `internal/game/common/kingdom` | 王国通用页面与导航 |
| `game/常规_王国竞技场` | `internal/game/arena` | 竞技场模块 |

### 2.2 项目结构

```
.
├── main.go
├── config.json
├── internal/
│   ├── config/
│   │   ├── static.go          # 静态常量：屏幕尺寸、OCR、RUNTIME、USER 默认值
│   │   └── user.go            # 用户配置加载/合并/保存
│   ├── store/
│   │   └── store.go           # data/store.json 键值读写
│   ├── logger/
│   │   └── logger.go          # 分级日志
│   ├── hud/
│   │   └── hud.go             # 状态浮层
│   ├── guard/
│   │   └── guard.go           # 弹窗守卫 + 分片 sleep
│   ├── statemachine/
│   │   └── machine.go         # KEEP/RETRY/DONE/Next/Fatal 语义
│   ├── scheduler/
│   │   ├── scheduler.go       # 任务轮询
│   │   └── builder.go         # TaskBuilder
│   ├── runtime/
│   │   └── runtime.go         # 大循环：guard → scheduler → idle
│   ├── dialog/
│   │   └── dialog.go          # 通用弹窗对象
│   ├── platform/
│   │   ├── screen/            # 屏幕识别
│   │   └── action/            # 动作执行
│   └── game/
│       ├── common/
│       │   └── kingdom/       # 王国首页、冒险页、导航
│       └── arena/             # 竞技场模块
│           ├── feature.go
│           ├── page.go
│           ├── route.go
│           ├── session.go
│           ├── task.go
│           └── statemachine.go
```

### 2.3 核心变化

1. **取消 `bot.Machine` 的全局主状态机垄断**：全局只保留 `Runtime` 大循环和 `Scheduler` 任务调度；每个任务内部再使用 `statemachine.Machine`。
2. **`Context` 拆分为全局运行时 + 任务上下文**：
   - 全局 `RuntimeContext`：`Config`、`Store`、`HUD`、`Logger`、`Guard`、`Detector`、`Executor`。
   - 任务级 `sm.Ctx`：任务自己的中间状态，如 `nextOcrPollAt`、`lastFloor`、`selectedOpponent`。
3. **业务模块固定 5+1 文件**：`feature.go` / `page.go` / `route.go` / `session.go` / `task.go` / `statemachine.go`。

---

## 3. 核心抽象设计

### 3.1 `internal/statemachine`

复刻 Lua 的返回语义，使用结构体 ADT：

```go
package statemachine

type Result interface{ result() }

type Keep struct{}                // 保持当前状态，不计重试
type Retry struct{}               // 保持当前状态，重试+1
type Done struct{}                // 正常结束
type Next string                  // 切换到指定状态
type Fatal struct{ Err error }    // 致命错误，终止状态机

func (Keep) result() {}
func (Retry) result() {}
func (Done) result() {}
func (Next) result() {}
func (Fatal) result() {}

type Handler func(sm *Machine) Result

type Machine struct {
    Current string
    Ctx     any
    // 内部计数与配置
}

func New() *Machine
func (m *Machine) Init(firstState string, opts Options)
func (m *Machine) Run(handlers map[string]Handler, runOpts RunOptions) error

type Options struct {
    MaxRetry      int
    MaxError      int
    Timeout       time.Duration
    RetryInterval time.Duration
}

type RunOptions struct {
    Interval time.Duration
    Guard    func() bool
    Label    string
}
```

### 3.2 `internal/scheduler`

```go
package scheduler

type Task struct {
    Name      string
    Condition func() bool
    Action    func() error
}

type Scheduler struct {
    tasks         []Task
    idleProviders map[string]func() (time.Duration, string)
}

func New() *Scheduler
func (s *Scheduler) Add(name string, condition func() bool, action func() error)
func (s *Scheduler) AddIdleProvider(name string, provider func() (time.Duration, string))
func (s *Scheduler) Run(stopOnError bool) (hasWork bool, err error)
func (s *Scheduler) MaxIdleWait() (time.Duration, string)
func (s *Scheduler) Clear()
```

### 3.3 `internal/scheduler` 的 `TaskBuilder`

```go
package scheduler

type TaskOpts struct {
    Name string

    ConfigKey    string
    CheckEnabled func() bool

    CanResume  func() bool
    CheckReady func() (bool, time.Duration)

    WaitHUD      func(remain time.Duration) string
    OnNotReady   func(remain time.Duration)

    Precondition       func() bool
    OnPreconditionFail func()

    Prepare func() error
    Action  func() error
}

func (s *Scheduler) Build(opts TaskOpts)
```

### 3.4 `internal/runtime`

```go
package runtime

type Runtime struct {
    scheduler *scheduler.Scheduler
    guard     *guard.Guard
    hud       *hud.HUD
    logger    *logger.Logger
    register  func()
}

func New(opts Options) *Runtime
func (r *Runtime) Register(fn func())
func (r *Runtime) Run() error
```

---

## 4. `arena` 模块详细设计

### 4.1 文件职责

| 文件 | 对应 Lua | 职责 |
|---|---|---|
| `feature.go` | `竞技场_特征库.lua` | 颜色、坐标、OCR 区域常量 |
| `page.go` | `竞技场_页面.lua` | 页面识别与点击 |
| `route.go` | `竞技场_路由.lua` | 进入/离开竞技场导航 |
| `session.go` | `竞技场_会话.lua` | 内存会话 + 刷新时间持久化 |
| `task.go` | `竞技场_任务.lua` | 任务入口，创建并运行内部状态机 |
| `statemachine.go` | 新增 | 内部状态机 handlers |

### 4.2 `feature.go`

```go
package arena

import "app/internal/platform/screen"

type Feature struct {
    Lobby      LobbyFeature
    Opponent   OpponentFeature
    TeamSelect TeamSelectFeature
    Dialog     DialogFeature
    Settlement SettlementFeature
    Pagination PaginationFeature
}

type LobbyFeature struct {
    Feature          screen.Feature
    CloseBtn         screen.Rect
    MedalTicketOCR   screen.Rect
    TrophyOCR        screen.Rect
    RefreshOCR       screen.Rect
    FreeRefreshOCR   screen.Rect
    FreeRefreshTap   screen.Point
    BuyTicketBtn     screen.Point
    BuyTicketSlider  screen.Swipe
    BuyTicketConfirm screen.Point
}

// ...
```

### 4.3 `page.go`

```go
package arena

import (
    "time"
    "app/internal/platform/action"
    "app/internal/platform/screen"
)

type Page struct {
    detector screen.Detector
    executor action.Executor
    feature  *Feature
}

func NewPage(det screen.Detector, exec action.Executor, f *Feature) *Page

func (p *Page) IsLobby() bool
func (p *Page) WaitLobby(timeout time.Duration) bool
func (p *Page) ReadMedalAndTicket() (medal, ticket int, ok bool)
func (p *Page) ReadTrophyCount() (int, bool)
func (p *Page) FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo
func (p *Page) SwipePageLeft()
func (p *Page) IsFreeRefresh() bool
func (p *Page) TapFreeRefresh()
func (p *Page) ReadRefreshCountdown() (time.Duration, bool)
func (p *Page) BuyTicket()
func (p *Page) RunBattle() (result string, ok bool)
func (p *Page) TapToLobby() bool
```

### 4.4 `route.go`

```go
package arena

type Route struct {
    page        *Page
    kingdomPage *kingdom.Page
}

func NewRoute(page *Page, kingdomPage *kingdom.Page) *Route

func (r *Route) Enter() bool
func (r *Route) Leave() bool
```

### 4.5 `session.go`

```go
package arena

import (
    "time"
    "app/internal/store"
)

type Session struct {
    store *store.Store

    Wins     int
    Draws    int
    Losses   int
    BuyCount int
    Tickets  int
    Medals   int
    Trophies int
}

func NewSession(store *store.Store) *Session

func (s *Session) TotalBattles() int
func (s *Session) IsReachMaxBattles(cfg *config.Arena) bool
func (s *Session) Describe() string
func (s *Session) SetNextFreeRefreshAt(at time.Time)
func (s *Session) TimeUntilRefresh() time.Duration
func (s *Session) ClearNextFreeRefresh()
func (s *Session) Reset()
```

只持久化 `nextFreeRefreshAt`，其它字段在每次脚本启动时归零。

### 4.6 `task.go` + `statemachine.go`

```go
package arena

type Task struct {
    cfg     *config.Arena
    page    *Page
    route   *Route
    session *Session
    sm      *statemachine.Machine
}

func NewTask(cfg *config.Arena, page *Page, route *Route, session *Session) *Task

func (t *Task) Run() error {
    t.sm.Init("detect", statemachine.Options{
        MaxRetry:      3,
        MaxError:      3,
        Timeout:       30 * time.Minute,
        RetryInterval: 1 * time.Second,
    })
    t.sm.Ctx = &arenaCtx{task: t}
    return t.sm.Run(t.handlers(), statemachine.RunOptions{
        Interval: 500 * time.Millisecond,
        Guard:    guard.Check,
        Label:    "王国竞技场",
    })
}
```

内部状态：

```
detect ──► navigate ──► sync ──► check ──► selectOpponent ──► teamSelect ──► battle ──► settlement
                                          ▲                    │
                                          └────────────────────┘
```

退出条件：

- `check`：达到战斗上限 / 无门票且不买票 → `leave`
- `selectOpponent`：翻页+刷新后无合适敌人且不可刷新 → `leave`
- `settlement` 后：满足退出条件 → `leave`

`leave` 状态负责调用 `route.Leave()` 返回王国首页，然后返回 `Done`。

---

## 5. 配置设计

### 5.1 `config.json`

```json
{
  "tickIntervalMs": 800,
  "maxStateDurationSec": 45,
  "maxUnknownRetries": 5,
  "maxRecoveryAttempts": 3,
  "lowPowerWaitSec": 30,
  "modules": {
    "collectResources": true,
    "farmLevels": false,
    "arena": {
      "enabled": true,
      "maxBattles": null,
      "autoBuyCount": 0,
      "trophyDiff": 0
    }
  }
}
```

### 5.2 Go 结构

```go
package config

type Arena struct {
    Enabled      bool `json:"enabled"`
    MaxBattles   *int `json:"maxBattles"`   // nil 表示不限
    AutoBuyCount int  `json:"autoBuyCount"`
    TrophyDiff   int  `json:"trophyDiff"`
}
```

---

## 6. 错误处理与恢复

### 6.1 三层错误处理

| 层级 | 职责 | 处理方式 |
|---|---|---|
| 页面/动作层 | 截图失败、点击失败、OCR 失败 | 返回 `(ok bool)` 或 `error`，不 panic |
| 状态机层 | handler 报错、RETRY 超限、timeout | `errors` 计数，超限后返回 fatal error |
| 调度/运行时层 | 任务 `Run()` 返回 error | 记录日志，根据 `stopOnError` 决定是否停止 |

### 6.2 状态机恢复语义

- `Retry`：同状态重试，最多 `MaxRetry + 1` 次执行。
- handler 报错：异常重试，最多 `MaxError + 1` 次。
- `Timeout`：单任务运行超过阈值后强制失败。
- `Fatal`：直接退出当前任务。

### 6.3 竞技场恢复

- 各状态失败优先返回 `Retry`。
- 达到重试上限后返回 `Fatal`，任务结束。
- 任务结束后由 `Scheduler` 根据 `Session` 的统计和刷新时间决定下次是否再执行。
- 全局 `Guard` 在长等待期间自动处理弹窗。

---

## 7. 测试策略

### 7.1 单元测试

| 被测对象 | 测试内容 |
|---|---|
| `internal/statemachine` | mock handler 验证 `Keep/Retry/Done/Next/Fatal`、重试超限、timeout |
| `internal/scheduler` | mock task 验证 condition/action 调用、idle provider 计算 |
| `internal/store` | 文件读写、Get/Set/Incr |
| `internal/game/arena/session` | 战斗统计、刷新时间持久化 |
| `internal/game/arena/page` | mock Detector/Executor 验证识别与点击逻辑 |

### 7.2 集成测试

- 真机/模拟器手动运行竞技场任务，观察状态切换。
- 使用 AutoGo 插件单步调试。

### 7.3 Mock 策略

- `screen.Detector` 和 `action.Executor` 保持 interface。
- `arena.Page` 接收 interface，可独立测试。
- `arena.Task` 在测试中注入 mock `Page` / `Route` / `Session`，验证状态机流转。

---

## 8. 实施顺序

建议按以下顺序实现，每步都可编译和跑测试：

1. 搭建新包结构：`config`、`store`、`logger`、`hud`、`guard`、`statemachine`、`scheduler`、`runtime`。
2. 实现 `statemachine` 和 `scheduler` 并补充单元测试。
3. 实现 `platform/screen` 和 `platform/action`（保留现有接口，迁移到新路径）。
4. 实现 `game/common/kingdom` 基础页面与导航。
5. 实现 `game/arena` 模块：
   - `feature.go` → `page.go` → `route.go` → `session.go` → `statemachine.go` → `task.go`。
6. 在 `main.go` 中初始化 `Runtime`，注册竞技场任务。
7. 跑完整 `go test ./...` 和 `go build ./...`。

---

## 9. 后续扩展

- 其它业务模块（矿山、广场、海滩交易所等）按 `game/arena` 的 5+1 文件模式添加。
- 通用弹窗、网络异常弹窗等逐步注册到 `Guard`。
- `HUD`、`Logger` 可根据需要接入 AutoGo 的 toast/日志能力。

---

## 10. 未决事项

无。本设计已覆盖：架构重设计、核心抽象、竞技场模块、配置、错误处理、测试策略、实施顺序。
