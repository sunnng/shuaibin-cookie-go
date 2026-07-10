# 王国竞技场 · 识别层补全设计（第一段）

> 日期：2026-07-10
> 承接：[2026-07-08 模块设计](./2026-07-08-arena-module-design.md)
> 范围：本段只补 **识别层**（`arena.Page` 的识别/读数/选对手）。后续段再处理 `teamSelect`、`RunBattle`、`route`。

---

## 1. 目标与范围

### 1.1 目标

让 `arena.Page` 的识别/读数/选对手方法在 `feature` 填值后能正确工作，且本地 mock 可测。本段交付后，`page` 的接口与逻辑定型，后续段只替换 `teamSelect` / `RunBattle` 等内部实现。

### 1.2 本段改动清单

| 文件 | 改动 |
|---|---|
| `internal/game/arena/page.go` | 实现 9 个方法（§5） |
| `internal/game/arena/feature.go` | **只补字段结构、不填数值**（§4） |
| `internal/platform/screen/detector.go` | 接口新增 `FindMultiColorsAll`（§3） |
| `internal/platform/screen/color.go` | android 实现 `FindMultiColorsAll` |
| `internal/platform/screen/factory.go` | stub 返回 nil |
| `internal/game/arena/page_test.go`（新增/补全） | mock 单元测试（§8） |

### 1.3 不在本段

`teamSelect` 真实选将、`RunBattle` 战斗流程、`route.Enter/Leave` 完善、`feature` 数值填写。其中 `feature` 数值由用户用取色工具填，本段不代填。

### 1.4 完成标准

`page` 识别方法在 mock `Detector/Executor` 下单元测试通过。本段**不断言真机识别结果**；真机核对在 `feature` 填值后按 §9 进行。

---

## 2. 找色 API 对齐

`internal/platform/screen/color.go` 已包装 `images` 绑定。本段用到：

```
images.FindColor          (x1,y1,x2,y2, colorStr, sim, dir, displayId) → (x,y)   单色·找首个
images.FindMultiColorsAll (x1,y1,x2,y2, colors,   sim, dir, displayId) → []Point 多点·找全部
images.CmpColor           (x, y, colorStr, sim, displayId)             → bool    单点比色
images.DetectsMultiColors (colors, sim, displayId)                     → bool    多点判定
```

颜色串格式：

- 单色：`"FFFFFF|CCCCCC-101010"`（`|` 候选色，`-` 偏色容差）。
- 多点：`"ffccff-151515,635,978,ffab2d-101010,6,29,24b1ff-101010"`，首项为基准点颜色，后续每项为「相对基准点的偏移 x,y + 颜色 + 偏色」；找色返回基准点坐标。

`FindColor` 与 `FindMultiColorsAll` 参数集相同（region + 颜色串 + sim + dir），所以 `feature` 只需一套 `{Region, Colors, Sim, Dir}` 即可兼容两者——颜色串写单色还是多点、返回首个还是全部，是代码侧选择，feature 字段不变。

`dir`：0 左上起、1 右上起、2 左下起、3 右下起。本段默认 `0`（从上到下）。

---

## 3. Detector 接口扩展

```go
// internal/platform/screen/detector.go
type Detector interface {
    // ...existing...
    FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point
}
```

- **android**（`color.go`）：调 `images.FindMultiColorsAll(region.X1, region.Y1, region.X2, region.Y2, colors, sim, dir, d.displayId)`，把返回的 `[]images.Point` 逐点映成 `[]screen.Point`。
- **stub**（`factory.go`）：`return nil`。
- `displayId` 由 wrapper 注入，不进入 `Detector` 接口签名（与现有 `FindColor` 一致）。

---

## 4. feature 结构（仅结构，数值由用户取色）

### 4.1 通用描述结构（新增于 `feature.go`）

```go
// 对齐 images.FindColor / FindMultiColors[All] 的全部静态参数
type ColorFind struct {
    Region screen.Region // x1,y1,x2,y2；右下 0,0 = 屏幕最大
    Colors string        // 单色串或多点串（取色工具直接拷）
    Sim    float32       // 0.1–1.0
    Dir    int           // 0–3
}

// 对齐 images.CmpColor（单点比色，用于战绩判定）
type ColorCmp struct {
    Point screen.Point // 相对锚点的偏移点
    Color string       // 单色串
    Sim   float32
}
```

### 4.2 OpponentFeature（复用现有字段 + 重语义 + 补齐）

```go
type OpponentFeature struct {
    Anchor       ColorFind     // 找卡锚点：Region=搜索区, Colors=锚点(单/多点串), Sim, Dir
    TrophyRect   screen.Region // ★重语义：相对锚点的奖杯 OCR 偏移矩形 (dx1,dy1,dx2,dy2)
    ResultOffset screen.Point  // 复用：相对锚点的战绩标记点偏移
    ResultColors ResultColors  // 复用：已战颜色 {Win,Draw,Lose}
    ResultSim    float32       // 战绩 CmpColor 相似度（新增）
    ClickOffset  screen.Point  // 相对锚点的点击偏移；锚点本身可点则 (0,0)（新增）
    NumberOCR    screen.OCRCfg // 保留
    // BaseSite → 弃用（锚点是 FindMultiColorsAll 找出来的，不写死）
    // FindDef  → 弃用（能力由 Anchor.ColorFind 取代）
}
```

### 4.3 LobbyFeature（现有结构不变，识别用 `DetectsMultiColors`）

`Lobby.Identify` 为 `screen.Feature{Colors, Sim}`，`IsLobby` 用 `DetectsMultiColors(Colors, Sim)`（不需要 region/dir）。`Lobby.Actions` / `Lobby.Reads` / `Lobby.Gestures` 保持现有字段。

### 4.4 取色表（交付给用户，数值由取色工具填）

| 字段 | 取什么 |
|---|---|
| `Lobby.Identify.Colors` / `.Sim` | 大厅多点比色串 + 相似度 |
| `Lobby.Reads.MedalTicket` | "勋章 门票"两数字 OCR 区域 |
| `Lobby.Reads.Trophy` | 自己奖杯数 OCR 区域 |
| `Lobby.Reads.Refresh` | 刷新倒计时文本 OCR 区域 |
| `Lobby.Reads.FreeRefresh` | "免费刷新"文本 OCR 区域 |
| `Lobby.Actions.FreeRefresh` | 免费刷新按钮点击点 |
| `Lobby.Gestures.SwipeLeft` | 翻页滑动 From/To/DurationMs |
| `Opponent.Anchor.Region` | 对手列表搜索区 (x1,y1,x2,y2) |
| `Opponent.Anchor.Colors` | 锚点颜色串（单色或多点，多点更稳） |
| `Opponent.Anchor.Sim` / `.Dir` | 相似度 / 方向（默认 0） |
| `Opponent.TrophyRect` | 相对锚点的奖杯 OCR 偏移矩形 |
| `Opponent.ResultOffset` + `ResultColors` + `ResultSim` | 相对锚点的战绩点 + 已战色 + 相似度 |
| `Opponent.ClickOffset` | 相对锚点的点击偏移（锚点可点则 0,0） |

代码里只写 `detector.FindMultiColorsAll(Anchor.Region, Anchor.Colors, Anchor.Sim, Anchor.Dir)` 这类引用，**零颜色/坐标硬编码**；UI 变化只改 feature。

---

## 5. page 方法逻辑

`offsetRegion(rel Region, a Point) Region`、`offsetPoint(rel Point, a Point) Point`、`readInt(s string) (int, bool)` 为文件内私有辅助：前两者把"相对锚点的偏移"换算成绝对坐标；`readInt` 即 `strconv.Atoi(strings.TrimSpace(s))` 的包装，失败返回 `(0, false)`。

### 5.1 判定类

```go
func (p *Page) IsLobby() bool {
    id := p.feature.Lobby.Identify
    return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

func (p *Page) IsFreeRefresh() bool {
    return strings.TrimSpace(p.detector.OCRText(p.feature.Lobby.Reads.FreeRefresh)) == "免费刷新"
}
```

`WaitLobby(timeout)` 保持现有实现（轮询 `IsLobby`）。

### 5.2 读数类

```go
func (p *Page) ReadMedalAndTicket() (int, int, bool) {
    parts := strings.Fields(p.detector.OCRText(p.feature.Lobby.Reads.MedalTicket))
    if len(parts) < 2 { return 0, 0, false }
    medal, e1 := strconv.Atoi(parts[0])
    ticket, e2 := strconv.Atoi(parts[1])
    if e1 != nil || e2 != nil { return 0, 0, false }
    return medal, ticket, true
}

func (p *Page) ReadTrophyCount() (int, bool) {
    n, err := strconv.Atoi(strings.TrimSpace(p.detector.OCRText(p.feature.Lobby.Reads.Trophy)))
    return n, err == nil
}

func (p *Page) ReadRefreshCountdown() (time.Duration, bool) {
    text := strings.TrimSpace(p.detector.OCRText(p.feature.Lobby.Reads.Refresh))
    return parseCountdown(text)
}
```

`parseCountdown(s string) (time.Duration, bool)`（文件内私有）：

- 用正则抓取数字，支持 `"5分30秒"` → 330s、`"30秒"` → 30s、`"5分"` → 300s。
- 也兼容 `"05:30"` 冒号格式 → 330s。
- 抓不到数字或全 0 → `(0, false)`。

### 5.3 动作类（现有逻辑保留）

`TapFreeRefresh` 点 `Lobby.Actions.FreeRefresh` 后 `Sleep`；`SwipePageLeft` 用 `Lobby.Gestures.SwipeLeft` 调 `executor.Swipe` 后 `Sleep`。

### 5.4 FindFirstValidOpponent（核心）

```go
func (p *Page) FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo {
    op := p.feature.Lobby.Opponent
    anchors := p.detector.FindMultiColorsAll(op.Anchor.Region, op.Anchor.Colors, op.Anchor.Sim, op.Anchor.Dir)
    lo, hi := myTrophy-cfg.TrophyDiff, myTrophy+cfg.TrophyDiff

    for _, a := range anchors {
        trophy, ok := readInt(p.detector.OCRText(offsetRegion(op.TrophyRect, a)))
        if !ok {
            continue // OCR 失败：跳过该锚点，不漏后面的卡
        }
        battled := battledAt(p.detector, offsetPoint(op.ResultOffset, a), op)
        if battled {
            continue
        }
        if trophy >= lo && trophy <= hi {
            return &OpponentInfo{Site: offsetPoint(op.ClickOffset, a), Trophies: trophy, IsBattled: false}
        }
    }
    return nil
}
```

`battledAt`：对 `ResultColors.Win/Draw/Lose` 逐个 `CmpColor(ResultOffset 偏移点, color, ResultSim)`。任一命中 → 已战；**全部未命中或比色异常 → 视为已战（保守，宁可漏打不重复打）**（§7）。

返回的锚点已按 `Dir=0` 排序（从上到下），第一个命中区间的就是视觉最靠上的有效对手。

---

## 6. 选对手策略

`config.Arena.TrophyDiff`（int）采用**区间半宽**语义：只打奖杯在 `[myTrophy - diff, myTrophy + diff]` 内的对手。

- `diff = 0`：`lo == hi == myTrophy`，只打奖杯完全等于自己的对手。
- 注意：`TrophyDiff` 默认 `0` 在此语义下表示"严格相等"，**不是"不过滤"**。需在 `config` 结构注释中写明，避免误解。

---

## 7. 错误处理

| 情况 | 处理 |
|---|---|
| OCR 奖杯失败 | `continue` 跳过该锚点，不 Fatal、不漏后面的卡 |
| `FindMultiColorsAll` 返回空 | `return nil`，由 `selectOpponent` 状态决定滑屏/刷新/leave |
| 战绩比色全部未命中或异常 | 视为**已战**（保守，省票；宁漏勿重） |
| `FindFirstValidOpponent` 整体 | 不返回 error，不 Fatal；失败都降级为跳过/返 nil |

读数类方法返回 `(…, ok bool)`，解析失败 `ok=false`，由调用方（状态机 `sync`）决定是否重试，本层不 Fatal。

---

## 8. 测试策略（mock Detector，纯本地）

在 `internal/game/arena/page_test.go` 用 mock `screen.Detector` / `action.Executor`：

- mock 预置：`FindMultiColorsAll` 返回锚点切片；`OCRText` 按 region 返回文本；`CmpColor` 按 (point,color) 返回 bool。
- table-driven 用例：
  1. 区间内命中：返回第一个奖杯在区间的锚点，`Site == anchor+ClickOffset`。
  2. 已战跳过：战绩点命中 Win/Draw/Lose 的卡被跳过。
  3. 区间外跳过：奖杯不在 `[myTrophy±diff]` 的卡被跳过。
  4. 无锚点：`FindMultiColorsAll` 返空 → `nil`。
  5. OCR 失败跳过：奖杯 OCR 非数字 → 跳过该锚点继续。
  6. 战绩异常当已战：战绩色全未命中 → 跳过。
  7. 多锚点取序：返回第一个区间内锚点（验证 dir 顺序）。
  8. `parseCountdown`：`"5分30秒"→330s`、`"30秒"→30s`、`"5分"→300s`、`"05:30"→330s`、`""→(0,false)`、`"abc"→(0,false)`。
  9. `IsFreeRefresh`：`"免费刷新"` 真 / 其它文本假。
  10. `ReadMedalAndTicket`：两数字解析、单/非数字 false。

驱动用 `page` 直接调用（不经过状态机），断言返回值与 `ok`。

---

## 9. 真机验证清单（feature 填值后，由用户上机核对）

- [ ] 进竞技场大厅：`IsLobby()==true`；离开大厅：`IsLobby()==false`。
- [ ] `ReadMedalAndTicket` 读到的勋章/票与界面一致。
- [ ] `ReadTrophyCount` 与界面自己奖杯一致。
- [ ] `ReadRefreshCountdown`：`"5分30秒"` 类文本解析为正确 `Duration`。
- [ ] `IsFreeRefresh` 在免费刷新可用时为 true。
- [ ] `FindMultiColorsAll(Anchor…)` 返回的锚点数量 == 当前页对手数。
- [ ] `FindFirstValidOpponent` 选中的卡：奖杯在区间、未战、点击位准确。

任一不符：只调 `feature` 数值（坐标/颜色/sim/dir），不改 `page.go` 逻辑。

---

## 10. 后续衔接

- 下一段（战斗全流程）：实现 `teamSelect` 真实选将、`RunBattle`、`settlement` 独立状态、弹窗。届时 `FindFirstValidOpponent` 返回的 `Site` 会被 `teamSelect` 消费，本段保证 `Site` 是"可点击的挑战点"。
- 再下一段（入口/离场）：`route.Enter/Leave` 的 OCR 入口与回主页。
- 本段不修改 `task.go` 的 `page` 私有接口（`FindFirstValidOpponent` 签名不变），状态机无需改动。

---

## 11. 风险与未决

- `FindMultiColorsAll` 返回顺序依赖 `dir` 与底层实现；若实际顺序不稳定，`selectOpponent` 取到的"第一个"可能不是视觉最上。取色时固定 `dir=0` 并观察，必要时在代码侧按 `a.Y` 再排序。
- 相对偏移取色误差：`TrophyRect` / `ResultOffset` 可能要微调 1–3 px。
- 奖杯 OCR 字体/相似度：`NumberOCR` 当前为占位配置，若 OCR 不稳需补 OCR 参数。
- 锚点单色不稳：用多点串（`Anchor.Colors` 写多点格式即可，字段不变）。
