# AutoGo Game Bot Template

基于 AutoGo 框架的 Android 游戏自动化脚本模板。目标设备分辨率 **1600×900 / 240 dpi**，以图色识别 + OCR 判断页面，通过点击、滑动等方式完成自动化操作。

## 项目结构

```
.
├── main.go                      # AutoGo 入口：加载配置、装配识别器/执行器/状态、启动状态机
├── config.json                  # 用户偏好（循环间隔、模块开关、恢复阈值）
├── internal/
│   ├── action/                  # 点击、长按、滑动、返回、Home、等待 + 坐标换算
│   ├── bot/                     # 配置、上下文、状态接口、注册表、状态机（含看门狗与恢复）
│   ├── bot/states/              # 示例状态：home、battle、unknown
│   ├── screen/                  # 屏幕识别接口与 AutoGo 封装（颜色、模板、OCR）
│   └── utils/                   # 日志封装
├── AutoGo/                      # 本地 AutoGo SDK（stub）
├── resources/                   # APK 打包资源
└── docs/                        # 设计文档与 AutoGo API 参考
```

## 快速开始

1. 使用 AutoGo JetBrains 插件打开本项目。
2. 连接 Android 设备或模拟器，确保 `adb devices` 已列出。
3. 根据游戏界面修改 `internal/bot/states/` 中的颜色、坐标和识别逻辑。
4. 按 **F7** 运行脚本；通过插件菜单打包为 APK。

## 配置

`config.json` 中可调整的项：

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
    "arena": false
  }
}
```

稳定的 UI 常量（按钮坐标、颜色字符串、模板文件名）建议直接写在对应 state 文件里；`config.json` 只放用户偏好和开关。

## 状态机

核心循环位于 `internal/bot/machine.go`：

1. `Detect(ctx)` 判断当前是否处于该状态。
2. `Act(ctx)` 执行该状态应做的操作。
3. `Next(ctx)` 决定下一个状态。
4. `Recover(ctx)` 在识别丢失或看门狗超时时被调用，负责安全返回主页/上一级。

状态注册顺序影响识别优先级，建议把 `unknown` 放在最后作为兜底。

## 添加新状态

1. 在 `internal/bot/states/` 下新建文件，实现 `bot.State` 接口。
2. 在 `main.go` 中实例化并注册到 `bot.Registry`。
3. 在 `config.json` 中增加对应模块开关（如果需要）。

示例骨架：

```go
type Shop struct{}

func (s *Shop) Name() string                 { return "shop" }
func (s *Shop) Detect(ctx *bot.Context) bool { return ctx.Detector.MatchColor(100, 100, "FF0000", 0.9) }
func (s *Shop) Act(ctx *bot.Context) error   {
    _ = ctx.Executor.Tap(action.Point{X: 100, Y: 100})
    return nil
}
func (s *Shop) Next(ctx *bot.Context) bot.State { return nil }
func (s *Shop) Recover(ctx *bot.Context) error  {
    _ = ctx.Executor.Back()
    return nil
}
```

## 本地开发

项目使用 `//go:build android` 限制真正调用 AutoGo C/C++ 绑定的文件。Windows 等非 Android 环境仍可编译和测试：

```bash
go test ./...
go build ./...
```

本地测试不会执行真实点击或截图，仅验证状态机、坐标换算和代码结构。

## 识别方式

- **颜色识别**：`ctx.Detector.MatchColor(x, y, color, sim)`、`FindColor`。
- **多点找色**：`ctx.Detector.MatchMultiColor("x1,y1,color1|x2,y2,color2", sim)`。
- **模板匹配**：`ctx.Detector.MatchImage(region, templateBytes, sim)`。
- **OCR**：`ctx.Detector.OCRText(region)`。

## 恢复策略

- 每个状态优先使用自己的 `Recover()` 返回安全页面。
- 失败后执行通用序列：Back ×3 → Home。
- 连续恢复失败超过 `maxRecoveryAttempts` 后进入低功耗等待，每 `lowPowerWaitSec` 秒再试一次。

## 依赖

- `github.com/Dasongzi1366/AutoGo`（本地 `replace` 到 `./AutoGo`）
- 标准库：`image`、`encoding/json`、`time` 等

## 文档索引

- 设计文档：`docs/superpowers/specs/2026-07-08-autogo-game-bot-design.md`
- 实现计划：`docs/superpowers/plans/2026-07-08-autogo-game-bot.md`
- 进度台账：`.superpowers/sdd/progress.md`
- AutoGo API 参考：`docs/autogo-doc文档2026.6.6.md`
