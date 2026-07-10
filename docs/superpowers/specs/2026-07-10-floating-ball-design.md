# ImGui 悬浮球集成设计

> 日期：2026-07-10  
> 参考：`c:\Users\1\Downloads\imgui悬浮球.go`  
> 范围：把悬浮球接入现有 `internal/ui` 框架，并与 runtime 的开始/暂停/停止/退出打通。

---

## 1. 目标与范围

### 1.1 目标

- 启动即显示可拖动、可吸附边缘的悬浮球。
- 展开菜单提供：设置（开配置面板）、开始/停止、暂停/恢复、关闭（退出进程）、Logo（收起）。
- 配置面板与悬浮球共存于同一 `imgui.Run` 循环；脚本在 goroutine 中运行，不阻塞 UI。
- 面板「运行脚本」与球「开始」均可启动脚本。

### 1.2 本段改动清单

| 文件 | 改动 |
|------|------|
| `internal/ui/controller.go`（新增） | 脚本控制接口 + 状态枚举（平台无关） |
| `internal/ui/controller_test.go`（新增） | 状态机单测 |
| `internal/ui/ball_android.go`（新增） | 悬浮球绘制/拖动/吸附/展开（android+cgo） |
| `internal/ui/shell_android.go`（新增） | UI 壳：每帧画球 + 按需画面板 |
| `internal/ui/panel_android.go` | 从独占 `RunPanel` 改为可被壳调用的面板渲染 |
| `internal/ui/panel_stub.go` | 非 Android 仍直接 `OnRun`，无球 |
| `internal/runtime/runtime.go` | 增加 `Pause` / `Resume` |
| `internal/runtime/runtime_test.go` | Pause/Resume 单测 |
| `main.go` | 实现 Controller，接线壳与 bot |

### 1.3 不在本段

- 运行中改配置热更新（改完需 Stop 后再 Start 才生效）。
- 隐藏球后再次唤起。
- 更换示例视觉主题（颜色/图标基本照搬参考实现）。
- AutoGo 系统自带悬浮球（`Android.toml` 的 `showFloatingBall`）行为变更。

### 1.4 完成标准

- Android：启动可见球；可开面板；面板或球可 Start；Pause/Resume/Stop/Exit 行为符合 §3。
- 桌面：`go test ./...`、`go build ./...` 通过；stub 不依赖 imgui。

---

## 2. 架构

### 2.1 原则

单一 `imgui.Run` 循环驱动全部 UI；脚本控制通过 `Controller` 与 UI 解耦。

```
main
  ├─ ui.Store + SeedFromConfig
  ├─ BotController (实现 ui.Controller)
  └─ ui.RunShell(ShellOptions)
        imgui.Run 每帧:
          ├─ FloatingBall.Draw()          // 始终
          └─ if panelOpen: drawPanel()    // 按需
```

### 2.2 组件职责

| 组件 | 职责 |
|------|------|
| `FloatingBall` | 位置、拖动、吸附、展开动画、按钮命中；通过回调通知壳 |
| `RunShell` | 生命周期：init imgui、每帧调度球/面板、处理面板开关与保存 |
| `Controller` | `Start` / `Pause` / `Resume` / `Stop` / `Exit` + `State()` |
| `Panel` | 现有灰色配置面板内容；「运行脚本」调用 `Controller.Start` |
| `Runtime` | 主循环；支持 Pause 挂起、Stop 退出 |

### 2.3 生命周期

```
启动 → RunShell
  ├─ 始终绘制悬浮球
  ├─ 设置按钮 → panelOpen = true
  ├─ 面板「运行脚本」或球「开始」→ Controller.Start（goroutine 跑 runtime）
  ├─ 球「暂停/恢复/停止」→ Pause / Resume / Stop
  └─ 球「关闭」→ Exit（Stop + os.Exit(0)）
```

启动时是否默认打开面板：由 `ShellOptions.OpenPanelOnStart` 控制，默认 `true`（保持现有「先看配置再跑」体验）。

---

## 3. 按钮与状态语义

### 3.1 Controller 状态

```go
type ScriptState int

const (
    StateIdle ScriptState = iota
    StateRunning
    StatePaused
)
```

转换：

- `Idle` --Start--> `Running`
- `Running` --Pause--> `Paused`
- `Paused` --Resume--> `Running`
- `Running|Paused` --Stop--> `Idle`
- 任意 --Exit--> 停 runtime 后 `os.Exit(0)`

重复 `Start`（已在 Running/Paused）：忽略。  
`Pause` 在非 Running：忽略。  
`Resume` 在非 Paused：忽略。

### 3.2 UI 入口

| 入口 | 行为 |
|------|------|
| 面板「运行脚本」 | 保存 Store → `ApplyToConfig` → `Start`；可关面板，球保留 |
| 球「开始」 | Idle 时用当前 Store 配置 `Start`；Running 时该按钮显示为「停止」并调用 `Stop` |
| 球「暂停」 | Running → `Pause`；Paused 时显示恢复并 `Resume`；Idle 时无效 |
| 球「停止」 | 见「开始」在 Running 时的切换（与示例一致：开始/停止同一按钮位） |
| 球「设置」 | `panelOpen = true`（运行中也可开；首版不热更新配置） |
| 球「关闭」 | `Exit` |
| 球 Logo | 收起展开菜单 |

展开菜单按钮顺序（右贴边时向左展开，与示例一致）：

0. Logo（收起）  
1. 关闭  
2. 暂停/恢复  
3. 开始/停止  
4. 设置（相对示例新增，再向外侧一格）

### 3.3 Runtime Pause 语义

- `Pause`：主循环在**每轮开头**阻塞，直到 `Resume` 或 `Stop`。
- 正在执行的单次 `scheduler.Run` / 任务 step **不中途打断**；当前 step 结束后才进入挂起点。
- `Stop`：沿用现有 `stopCh`；Pause 等待中也必须能被 Stop 唤醒。

---

## 4. API 草图

### 4.1 Controller（`internal/ui/controller.go`）

```go
type Controller interface {
    State() ScriptState
    Start()
    Pause()
    Resume()
    Stop()
    Exit()
}
```

`main` 提供实现：持有 `*config.Config`、`*ui.Store`、当前 `*runtime.Runtime`、mutex；`Start` 内 `ApplyToConfig` 后组装 detector/executor/scheduler 并 `go rt.Run()`。

### 4.2 Shell（`internal/ui/shell_android.go`）

```go
type ShellOptions struct {
    Title            string
    ConfigPath       string
    CountdownSec     float64
    Store            *Store
    Render           func(store *Store)
    Controller       Controller
    OpenPanelOnStart bool // default true
}
```

`RunShell` 替代对外主入口；保留 `RunPanel` 作为薄封装（内部转调 `RunShell`）或直接改 `main` 调 `RunShell`。推荐 `main` 改调 `RunShell`，`RunPanel` 可删除或保留兼容。

### 4.3 FloatingBall 回调

```go
type BallCallbacks struct {
    OnSettings func()
    OnStartStop func() // 根据 State 调 Start 或 Stop
    OnPauseResume func()
    OnClose func() // Exit
}
```

球内部只负责 UI；状态颜色读 `Controller.State()`（Running=绿，Paused=黄，Idle=蓝）。

---

## 5. 与现有代码的衔接

### 5.1 面板

- 现有 `DefaultCookiePanel` / widgets 不变。
- 「运行脚本」不再 `select{}` 式堵死：只关面板并 `Controller.Start()`。
- 关闭面板标题栏 X：仅 `panelOpen=false`，**不**退出进程（退出只走球「关闭」）。

### 5.2 main.go

```go
ui.RunShell(ui.ShellOptions{
    Title: "Superbin Cookie",
    Store: uiStore,
    Controller: botCtrl,
    OpenPanelOnStart: true,
})
```

`runBot` 逻辑迁入 `BotController.Start`（或由其调用），避免在 imgui 回调里同步阻塞。

### 5.3 非 Android stub

`RunShell` / `RunPanel`：直接 `Controller.Start()`（或 `OnRun`），无 imgui、无球。

---

## 6. 错误处理

- runtime `Run` 返回错误：打日志，状态置 `Idle`，球恢复蓝色。
- 配置 Load 失败：不阻止壳启动（与现面板一致）。
- `Exit`：先 `Stop`，短等（如 500ms）再 `os.Exit(0)`，降低写 store 竞态。
- `Start` 组装失败：打日志，保持 `Idle`。

---

## 7. 测试计划

| 用例 | 包 | 断言 |
|------|-----|------|
| Pause 后不再调度 | `runtime` | Pause 期间 mock task 调用次数不增加；Resume 后继续 |
| Stop 可打断 Pause 等待 | `runtime` | Pause 中 Stop → `Run` 返回 |
| 状态转换 | `ui` Controller 测试替身或可测实现 | Idle→Running→Paused→Running→Idle |
| 重复 Start | 同上 | 第二次 Start 不新建 runtime |
| stub 编译 | 全仓 | `go test ./...`、`go build ./...` |

悬浮球命中/动画：Android 真机手测，不做桌面单测。

---

## 8. 真机手测清单

1. 启动：球在右侧中间；默认面板打开。  
2. 拖动球 → 松手吸附左/右边缘。  
3. 点球展开 → 设置打开面板；Logo 收起。  
4. 面板「运行脚本」→ 球变绿，任务开始跑。  
5. 暂停 → 球变黄，任务不再推进；恢复 → 继续。  
6. 停止 → 球变蓝；再点开始可重新跑。  
7. 关闭 → 进程退出。  
8. 运行中开设置改选项 → 需停止再开始后新配置生效。

---

## 9. 决策记录

| 决策 | 选择 |
|------|------|
| 出现时机 | 启动即有球 + 运行期保留 |
| 关闭按钮 | 退出整个脚本进程 |
| 打开面板 | 展开菜单「设置」按钮 |
| 启停语义 | 开始/暂停/停止对齐示例 |
| 面板「运行脚本」 | 保留，与球「开始」双入口 |
| 集成方式 | UI 壳统一驱动（方案 1） |
| Pause 粒度 | 轮次边界挂起，不打断当前 step |
