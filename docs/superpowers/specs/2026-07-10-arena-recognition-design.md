# 竞技场识别层补全设计

> 日期：2026-07-10
> 承接：[2026-07-08 王国竞技场模块设计](2026-07-08-arena-module-design.md)
> 范围：补全 `internal/game/arena` 的**识别层**（`page.go`）+ 对齐 AutoGo 找色 API 的 `feature.go` 结构 + `screen.Detector` 接口扩展。
> 状态：待用户审 → 写实施计划（writing-plans）

---

## 1. 背景与目标

竞技场六件套骨架已落地，但识别相关方法仍是占位：`page.go` 的 `FindFirstValidOpponent` / `ReadRefreshCountdown` 直接返回零值，`feature.go` 的坐标/颜色多为零值，状态机因此无法在真机上"看见"界面。

本设计是「补全竞技场」四段拆分中的**第一段（识别层）**，目标是让 `page.go` 的所有识别/读数/对手扫描方法在逻辑上完整、可测，并把全部静态识别参数（颜色串、相似度、搜索区域、方向）收进 `feature.go`，代码侧零硬编码。

四段拆分（每段独立 spec → plan → 实现）：

| 段 | 内容 | 依赖真机 |
|---|---|---|
| **1 · 识别层（本设计）** | `page.go` 识别/读数/选对手 + `feature` 结构 + `Detector` 扩展 | 逻辑本地完成；识别正确性待用户填 feature 后上机核对 |
| 2 · 状态机闭环 | `teamSelect` / 独立 `settlement` / `Leave` / `TapToLobby` | 否 |
| 3 · 战斗全流程 | `RunBattle` + 选将 + 弹窗 + 结算读胜负 | 是 |
| 4 · 入口/离场 | `route` 的 OCR 入口、回主页 | 是 |

### 1.1 分工

- **用户**：用取色工具按本设计第 6 节的取色表填 `feature.go` 的数值。
- **本设计覆盖**：`Detector` 接口扩展、`feature.go` 字段结构、`page.go` 逻辑、单元测试。

---

## 2. 范围

### 2.1 改动的文件

| 文件 | 改动性质 |
|---|---|
| `internal/platform/screen/detector.go` | 接口新增 `FindMultiColorsAll` |
| `internal/platform/screen/color.go` | Android 实现 `FindMultiColorsAll` |
| `internal/platform/screen/factory.go` | stub 实现 `FindMultiColorsAll`（返回 nil） |
| `internal/game/arena/feature.go` | 新增通用描述结构；`OpponentFeature` 字段重语义 + 补齐；**不填数值** |
| `internal/game/arena/page.go` | 实现识别/读数/选对手方法 |
| `internal/game/arena/page_test.go` | 新增/补全识别层单元测试 |

### 2.2 不在本段范围

`teamSelect`（选将）、`RunBattle`（战斗）、独立 `settlement` 状态、`route.Enter/Leave` 的 OCR 入口与回主页——留给第 2-4 段。本段只让 `page.go` 能"看见"。

### 2.3 完成标准（Done Criteria）

1. `Detector` 接口及 Android/stub 实现通过 `go build ./...` 与 `go build -tags android ./...`。
2. `feature.go` 字段结构对齐 AutoGo 找色 API，所有静态参数入 feature，page 代码无颜色/坐标硬编码。
3. `page.go` 识别方法在 mock `Detector/Executor` 下单元测试通过（`go test ./internal/game/arena/...`）。
4. 取色表（第 6 节）交付；真机识别正确性由用户填 feature 后按第 8 节验证清单上机核对——**本段不在本地断言真机识别**。

---

## 3. 平台接口扩展

### 3.1 对齐的 AutoGo API

```
images.FindColor          (x1,y1,x2,y2, colorStr, sim, dir, displayId) → (x,y)    单色·找首个
images.FindMultiColors    (x1,y1,x2,y2, colors,   sim, dir, displayId) → (x,y)    多点·找首个
images.FindMultiColorsAll (x1,y1,x2,y2, colors,   sim, dir, displayId) → []Point  多点·找全部
images.CmpColor           (x, y, colorStr, sim, displayId)             → bool     单点比色
images.DetectsMultiColors (colors, sim, displayId)                     → bool     多点判定
```

`x2/y2 = 0` 表示屏幕最大宽/高；`dir ∈ 0..3`（左上/右上/左下/右下起）。

颜色串格式：

- 单色：`"FFFFFF|CCCCCC-101010"`（`|` 候选色，`-` 偏色容差）。
- 多点：`"ffccff-151515,635,978,ffab2d-101010,6,29,24b1ff-101010"`——首项是基准点颜色，后续每项是"相对基准点的偏移 x,y + 颜色 + 偏色"；找色返回基准点坐标。该格式天然编码"锚点 + 相对偏移"。

### 3.2 接口变更

`internal/platform/screen/detector.go` 新增：

```go
type Detector interface {
    // ...existing...
    // FindMultiColorsAll 返回区域内所有匹配多点颜色序列的基准点坐标；无匹配返回 nil。
    FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point
}
```

- `color.go`（android）：调用 `images.FindMultiColorsAll(region.X1, region.Y1, region.X2, region.Y2, colors, sim, dir, d.displayId)`，将 `[]images.Point` 映射为 `[]screen.Point`。
- `factory.go`（`!android` stub）：返回 `nil`。
- `Point` 复用 `screen.Point`（已有 `X/Y int`）。

> 选型说明：找所有对手卡用 `FindMultiColorsAll`（一次返回全部锚点），而非循环 `FindColor`。`FindColor` 与 `FindMultiColorsAll` 参数集相同，feature 用一套 `{Region, Colors, Sim, Dir}` 兼容两者；区别仅在"颜色串写单色还是多点、返回一个还是全部"。

---

## 4. `feature.go` 结构（静态参数全入）

### 4.1 通用描述结构（新增于 `feature.go`，导出供 `page.go` 使用）

```go
// ColorFind 对齐 images.FindColor / FindMultiColors[All] 的全部静态参数。
type ColorFind struct {
    Region screen.Region // x1,y1,x2,y2；右下 0,0 = 屏幕最大
    Colors string        // 单色串或多点串（取色工具直接拷贝）
    Sim    float32       // 0.1-1.0
    Dir    int           // 0-3
}

// ColorCmp 对齐 images.CmpColor（单点比色，用于战绩判定）。
type ColorCmp struct {
    Point screen.Point // 相对锚点的偏移点
    Color string       // 单色串
    Sim   float32
}
```

### 4.2 `OpponentFeature` 字段（重语义 + 补齐）

```go
type OpponentFeature struct {
    Anchor       ColorFind     // 找卡锚点：Region=搜索区, Colors=锚点(单色或多点串), Sim, Dir
    TrophyRect   screen.Region // 相对锚点的奖杯 OCR 偏移矩形 (dx1,dy1,dx2,dy2)
    ResultOffset screen.Point  // 相对锚点的战绩标记点偏移
    ResultColors ResultColors  // 已战颜色 {Win,Draw,Lose}（单色串）
    ResultSim    float32       // 战绩 CmpColor 相似度
    ClickOffset  screen.Point  // 相对锚点的点击偏移；锚点本身可点则 (0,0)
    NumberOCR    screen.OCRCfg // 保留：数字 OCR 配置
}
```

旧字段处理：

- `BaseSite`（原 `screen.Point`）：找卡方案下锚点是 `FindMultiColorsAll` 找出来的，不再写死基点——**删除**。
- `FindDef`（原 `screen.FindDef`，空 struct）：能力被 `Anchor`（`ColorFind`）取代——**删除**。
- `TrophyRect`：**重语义**为相对锚点的偏移矩形（原未注明绝对/相对）。
- `ResultOffset` / `ResultColors` / `NumberOCR`：**保留并复用**。

### 4.3 `LobbyFeature` 调整

- `Lobby.Identify`（`screen.Feature`：Colors+Sim）用于 `DetectsMultiColors`，判定大厅（无需 region/dir）——保留。
- `Lobby.Reads`（MedalTicket/Trophy/Refresh/FreeRefresh，OCR 区域）、`Lobby.Actions`（FreeRefresh/Close/BuyTicket*）、`Lobby.Gestures.SwipeLeft`——保留，本段用到。
- 其余（TeamSelect/Settlement/Dialogs）保留结构，本段不用。

---

## 5. `page.go` 方法逻辑

未列出的方法（`WaitLobby` / `ReadMedalAndTicket` / `ReadTrophyCount` / `TapFreeRefresh` / `SwipePageLeft`）保持现有实现。

### 5.1 工具函数（包内私有）

```go
// offsetRegion 将相对锚点的偏移矩形换算为绝对区域。
func offsetRegion(rel screen.Region, anchor screen.Point) screen.Region {
    return screen.Region{
        X1: anchor.X + rel.X1, Y1: anchor.Y + rel.Y1,
        X2: anchor.X + rel.X2, Y2: anchor.Y + rel.Y2,
    }
}

// offsetPoint 将相对锚点的偏移点换算为绝对点。
func offsetPoint(rel screen.Point, anchor screen.Point) screen.Point {
    return screen.Point{X: anchor.X + rel.X, Y: anchor.Y + rel.Y}
}

// readInt 从 OCR 文本中解析首个整数；失败返回 false。
func readInt(text string) (int, bool) {
    n, err := strconv.Atoi(strings.TrimSpace(text))
    return n, err == nil
}

// parseCountdown 解析刷新倒计时：支持 "X分Y秒" / "Y秒" / "mm:ss" / 纯秒数。
func parseCountdown(text string) (time.Duration, bool) { /* 见 5.4 */ }
```

### 5.2 `IsLobby`

```go
func (p *Page) IsLobby() bool {
    id := p.feature.Lobby.Identify
    return p.detector.MatchMultiColor(id.Colors, id.Sim)
}
```

### 5.3 `IsFreeRefresh`

```go
func (p *Page) IsFreeRefresh() bool {
    return strings.TrimSpace(p.detector.OCRText(p.feature.Lobby.Reads.FreeRefresh)) == "免费刷新"
}
```

> OCR 偶发噪点时，可把精确匹配放宽为 `strings.Contains(..., "免费刷新")`，由真机验证决定。

### 5.4 `ReadRefreshCountdown`

```go
func (p *Page) ReadRefreshCountdown() (time.Duration, bool) {
    text := strings.TrimSpace(p.detector.OCRText(p.feature.Lobby.Reads.Refresh))
    return parseCountdown(text)
}
```

`parseCountdown` 规则：

1. 含 `"分"`：抓 `(\d+)分(?:(\d+)秒)?` → `分*60 + 秒`（秒缺省为 0）。
2. 否则含 `"秒"`：抓 `(\d+)秒` → 秒数。
3. 否则匹配 `^(\d{1,2}):(\d{2})$`（冒号 `mm:ss`）→ `mm*60 + ss`。
4. 否则纯数字 `^\d+$` → 视作秒数。
5. 均不匹配 → 返回 `(0, false)`。

> 实际文本格式以上机为准；解析器用正则表实现，新增格式只需加规则。

### 5.5 `FindFirstValidOpponent`（核心）

策略：奖杯在 `[myTrophy - TrophyDiff, myTrophy + TrophyDiff]` 区间内、且未战的对手，取首张（按 `Dir` 顺序）。

```go
func (p *Page) FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo {
    op := p.feature.Lobby.Opponent
    lo := myTrophy - cfg.TrophyDiff
    hi := myTrophy + cfg.TrophyDiff

    anchors := p.detector.FindMultiColorsAll(op.Anchor.Region, op.Anchor.Colors, op.Anchor.Sim, op.Anchor.Dir)
    for _, anchor := range anchors {
        // 奖杯
        trophy, okNum := readInt(p.detector.OCRText(offsetRegion(op.TrophyRect, anchor)))
        if !okNum {
            continue // OCR 失败：跳过该卡，不视为无卡
        }
        // 已战判定：战绩点命中 Win/Draw/Lose 任一色
        rp := offsetPoint(op.ResultOffset, anchor)
        battled := p.detector.MatchColor(rp.X, rp.Y, op.ResultColors.Win, op.ResultSim) ||
            p.detector.MatchColor(rp.X, rp.Y, op.ResultColors.Draw, op.ResultSim) ||
            p.detector.MatchColor(rp.X, rp.Y, op.ResultColors.Lose, op.ResultSim)
        if battled {
            continue
        }
        // 区间筛选
        if trophy < lo || trophy > hi {
            continue
        }
        return &OpponentInfo{
            Site:      offsetPoint(op.ClickOffset, anchor),
            Trophies:  trophy,
            IsBattled: false,
        }
    }
    return nil // 当前屏无合适对手；翻页/刷新由 statemachine 驱动
}
```

不变量：

- `FindMultiColorsAll` 已按 `Dir` 排序返回锚点，遍历顺序即选卡优先级；本方法只负责"当前屏"，翻页由 `selectOpponent` 状态的 `SwipePageLeft` + 再次调用本方法实现（保持现状）。
- OCR 失败 / 已战 / 不在区间 → `continue`，不误返回 nil 漏掉后面的卡。

### 5.6 边界：`TrophyDiff = 0`

区间退化为 `[myTrophy, myTrophy]`，即**只打奖杯完全相等的对手**。需要"不过滤奖杯"时不属于本（B）策略，留作后续配置项（如新增 `TrophyDiff < 0` 表示不过滤），本段不实现。

---

## 6. 取色表（交付用户填 `feature.go` 数值）

| 字段 | 取什么 |
|---|---|
| `Lobby.Identify.Colors` / `.Sim` | 大厅多点比色串 + 相似度（`DetectsMultiColors`） |
| `Lobby.Reads.MedalTicket` | "勋章 门票"两数字的 OCR 区域 |
| `Lobby.Reads.Trophy` | 自己奖杯数的 OCR 区域 |
| `Lobby.Reads.Refresh` | 刷新倒计时文本的 OCR 区域 |
| `Lobby.Reads.FreeRefresh` | "免费刷新"文本的 OCR 区域 |
| `Lobby.Actions.FreeRefresh` | 免费刷新按钮点击点 |
| `Lobby.Gestures.SwipeLeft` | 翻页滑动 From/To/DurationMs |
| `Opponent.Anchor.Region` | 对手列表搜索区 (x1,y1,x2,y2) |
| `Opponent.Anchor.Colors` | 锚点颜色串（单色或多点；多点更稳） |
| `Opponent.Anchor.Sim` / `.Dir` | 相似度 / 方向 0-3 |
| `Opponent.TrophyRect` | 相对锚点的奖杯 OCR 偏移矩形 (dx1,dy1,dx2,dy2) |
| `Opponent.ResultOffset` | 相对锚点的战绩标记点偏移 (dx,dy) |
| `Opponent.ResultColors.Win/Draw/Lose` | 已战标记的三个颜色串 |
| `Opponent.ResultSim` | 战绩比色相似度 |
| `Opponent.ClickOffset` | 相对锚点的点击偏移；锚点可点则 (0,0) |

---

## 7. 错误处理

| 失败点 | 处理 |
|---|---|
| `FindMultiColorsAll` 返回空 | `FindFirstValidOpponent` 返回 nil（statemachine 走翻页/刷新/leave） |
| 单卡 OCR 奖杯失败 | `continue` 跳过该卡，不影响其余卡 |
| 战绩比色失败 | 视为未战（`MatchColor` 返回 false 即未命中） |
| 倒计时文本不识 | `parseCountdown` 返回 `(0,false)`，statemachine 收到 `ok=false` 后按"未读到倒计时"处理（不持久化刷新时间） |
| 区间 `TrophyDiff=0` | 仅匹配同奖杯（见 5.6） |

识别层所有方法均不返回 `error`，统一用 `(v, ok bool)` 表达"读到/没读到"，由 statemachine 决策。

---

## 8. 测试

### 8.1 mock 扩展

在 `page_test.go` 中扩展 mock `Detector`，预置：`FindMultiColorsAll` 的锚点列表、`OCRText` 按区域返回文本、`MatchColor` 按点/色返回命中、`MatchMultiColor` 返回布尔。

### 8.2 用例

| 用例 | 验证 |
|---|---|
| `TestReadRefreshCountdown` | `"5分30秒"→330s`、`"30秒"→30s`、`"05:30"→330s`、`"90"→90s`、`"??"→(0,false)` |
| `TestFindFirstValidOpponent_InRange` | 区间内未战卡被选中，`Site = anchor + ClickOffset` |
| `TestFindFirstValidOpponent_BattledSkipped` | 已战卡被跳过，选中下一张 |
| `TestFindFirstValidOpponent_OutOfRange` | 区间外卡被跳过 |
| `TestFindFirstValidOpponent_NoAnchor` | 无锚点返回 nil |
| `TestIsFreeRefresh` | 文本精确匹配 |

### 8.3 真机验证清单（用户填 feature 后上机）

- [ ] `IsLobby` 在大厅 true、其它页 false
- [ ] `ReadMedalAndTicket` / `ReadTrophyCount` 读数与界面一致
- [ ] `ReadRefreshCountdown` 解析与实际倒计时一致
- [ ] `FindMultiColorsAll` 锚点全部检出（数量=当前屏对手数）
- [ ] `FindFirstValidOpponent` 区间筛选与战绩跳过符合预期

---

## 9. 后续

本段落地后，进入第 2 段（状态机闭环：`teamSelect` / 独立 `settlement` / `Leave` / `TapToLobby`），另起 spec。
