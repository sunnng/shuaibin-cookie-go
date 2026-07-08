# AutoGo 游戏脚本模板设计文档

## 背景

基于 AutoGo 框架开发一款可在 Android 设备上以 APK 形式运行的游戏自动化脚本。目标游戏为「模拟经营 + 卡牌战斗」双核玩法（类似《部落冲突》建造生产 + 角色养成对战）。项目从 0 开始，需要一个可扩展、不过度设计的生产级模板。

默认设备分辨率：**1600×900，240 dpi**。

## 目标

1. 提供一套清晰的状态机模板，便于后续接入具体游戏状态（主页、战斗、商店、养成等）。
2. 支持颜色识别、模板匹配、OCR 三种识别方式，并按场景灵活选择。
3. 封装点击、滑动、返回、Home 等基础操作，统一基于基准分辨率做坐标换算。
4. 内置看门狗与恢复机制，能够处理游戏页面繁多、返回路径不同的情况。
5. 遵循 Go 最佳实践，项目结构清晰，便于单元测试与维护。

## 非目标

- 不实现具体游戏的所有业务逻辑（本次只交付模板 + 2~3 个示例状态）。
- 不引入复杂的事件总线、插件系统或远程控制。
- 不针对 iOS 做适配（当前项目 `AutoGo.targetPlatform` 为 `android`）。

## 架构概览

采用「状态机 + 屏幕识别 + 动作执行」三层结构：

1. **StateMachine**（`internal/bot/machine.go`）
   - 维护当前状态，周期性执行 `Detect → Act → Transition`。
   - 内置看门狗与异常恢复。

2. **ScreenDetector**（`internal/screen/`）
   - 统一封装 color / image / OCR 识别。
   - 返回 ScreenID 与可信度。

3. **ActionExecutor**（`internal/action/`）
   - 封装点击、长按、滑动、等待、返回、Home 等操作。
   - 所有坐标基于 1600×900 计算，支持比例换算到其他分辨率。

4. **Config**（`internal/bot/config.go` + `config.json`）
   - 稳定 UI 元素（按钮坐标、颜色、模板路径）写入 Go 常量/struct。
   - 用户偏好（循环次数、开关、延迟）放入 `config.json`。

5. **Recovery**
   - 每个 State 自带 `Recover()` 方法，定义从该状态安全返回主页/上一级的路径。
   - 全局兜底：Back 多次 → Home。
   - Recovery 失败后进入低功耗等待状态，每 30s 重试一次。

## 项目结构

遵循 AutoGo 约定，`main.go` 位于项目根目录。

```
.
├── main.go                      # AutoGo 入口
├── internal/
│   ├── bot/
│   │   ├── bot.go               # Bot 实例与生命周期
│   │   ├── config.go            # 配置加载与常量
│   │   ├── state.go             # 状态接口与注册表
│   │   ├── machine.go           # 状态机调度、看门狗、恢复
│   │   └── states/
│   │       ├── home.go          # 主页状态示例
│   │       ├── battle.go        # 战斗状态示例
│   │       └── unknown.go       # 未知/恢复状态
│   ├── screen/
│   │   ├── detector.go          # 屏幕识别统一入口
│   │   ├── color.go             # 颜色/多点找色
│   │   ├── image.go             # 模板匹配
│   │   └── ocr.go               # OCR 读取数字/文本
│   ├── action/
│   │   ├── tap.go               # 点击、长按、滑动
│   │   ├── navigate.go          # 返回、Home、等待
│   │   └── coord.go             # 坐标换算（1600×900 基准）
│   └── utils/
│       └── log.go               # 日志封装
├── assets/
│   └── tpl/                     # 模板图片（打包进 APK）
│       ├── home_btn.png
│       └── battle_start.png
├── config.json                  # 用户偏好（循环次数、延迟等）
└── go.mod
```

## 数据流

1. **主循环**（`machine.go`）
   - 每 500~1000ms 执行一次 tick。
   - 当前 state's `Detect()` 返回是否命中当前屏幕。
   - 若命中：执行 `Act()`，再根据结果决定 `Next()` 返回的下一状态。
   - 若未命中：尝试用 `ScreenDetector` 识别全局已知屏幕，若识别到则切换；否则进入 unknown 计数。

2. **State 接口**

   ```go
   type State interface {
       Name() string
       Detect(ctx *Context) bool
       Act(ctx *Context) error
       Next(ctx *Context) State
       Recover(ctx *Context) error
   }
   ```

3. **Context**
   - Config、Logger、上一次状态、重试计数、状态进入时间、屏幕截图缓存（同一次 tick 内复用）。

4. **看门狗**
   - 记录进入某状态的时间。
   - 超过 `maxStateDuration` 未切换 → 触发 recovery。
   - 连续 N 次未识别到任何已知屏幕 → 触发 recovery。

## 错误处理与恢复

1. **错误分类**
   - `ErrScreenNotMatched`：未识别到目标屏幕，交给状态机重试/全局探测。
   - `ErrActionFailed`：点击/滑动失败，记录日志后重试一次。
   - `ErrStuck`：看门狗触发，强制 recovery。

2. **恢复流程**
   - 优先调用当前状态的 `Recover()` 方法返回主页/上一级。
   - 若失败或未实现，执行通用返回序列：Back 3 次 → Home。
   - 在 `config.go` 中维护「最深页面 → 主页」的点击序列映射，按当前识别到的页面选择最短路径。
   - Recovery 失败后进入低功耗等待状态，每 30s 重试一次，便于人工介入或游戏自动跳转。

3. **防御性设计**
   - 所有点击前检查坐标是否在 1600×900 范围内。
   - 截图失败时跳过本次 tick，不 panic。
   - OCR/找图失败返回零值，调用方决定是否重试。

## 测试策略

1. **单元测试**
   - `internal/action/coord.go`：测试 1600×900 基准到其他分辨率的坐标换算。
   - `internal/bot/machine.go`：用 mock State 验证状态切换、看门狗触发、恢复流程。

2. **集成测试**
   - 在真机/模拟器上手动运行，验证每个状态的 Detect/Act/Recover。
   - 使用 AutoGo 插件的「快速调试」功能迭代。

3. **Mock 层**
   - `screen.Detector` 和 `action.Executor` 定义 interface，便于单元测试时 mock，避免依赖 Android runtime。

4. **日志**
   - 每个 tick 输出「当前状态 → 识别结果 → 执行动作 → 下一状态」，方便复盘。

## 依赖

- `github.com/Dasongzi1366/AutoGo`（通过 `replace` 指向本地 `./AutoGo`）。
- 标准库：`image`、`encoding/json`、`log`、`time` 等。

## 后续扩展点

1. 在 `internal/bot/states/` 下新增具体游戏状态文件。
2. 在 `internal/bot/config.go` 中添加新的 UI 常量。
3. 在 `assets/tpl/` 中添加新的模板图片。
4. 通过 `config.json` 开启/关闭不同功能模块。
