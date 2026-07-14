# 王国竞技场 · 路由 + 最小战斗闭环设计

> 日期：2026-07-12  
> 承接：[识别层设计](./2026-07-10-arena-recognition-design.md)、[模块总设计](./2026-07-08-arena-module-design.md)  
> 决策：整包规划、竖切分段；`DefaultFeature` 数值由用户取色后填；战斗为最小闭环（不接 Guard 弹窗）

---

## 0. 目标与约束

### 0.1 目标

在识别层已定型的前提下，补全：

1. **进出路由**：王国首页 → 冒险 → 竞技场大厅 → 回王国首页  
2. **取色清单**：字段路径与验收顺序（数值不代填）  
3. **最小战斗闭环**：点对手 →（可选选将页）开战 → 等结算 → 读胜负 → 回大厅 → sync  

### 0.2 约束（已确认）

| 项 | 决定 |
|---|---|
| 实施切分 | 竖切：段1 路由 → 段2 取色验识别 → 段3 战斗 |
| Feature 数值 | 只补结构 + 清单；`DefaultFeature` 可保持空 |
| 战斗深度 | 最小闭环；**不**接 `DialogsFeature` / Guard 注册 |
| 结算 | **不**新增独立 `settlement` SM 态；逻辑在 `RunBattle` 内 |
| 未知页恢复 | 不在本设计范围；非首页/非冒险/非大厅 → 现有 detect Fatal |

### 0.3 不在范围

- Guard 弹窗接线、复杂异常恢复  
- 假胜占位（段3 必须移除 `teamSelect` 硬闸与 `RunBattle` 固定胜利）  
- 改调度/UI binding（除非发现阻塞）

---

## 1. 段1 · 进出路由

### 1.1 改动清单

| 文件 | 改动 |
|---|---|
| `internal/game/common/kingdom/feature.go` | 按 Identify/Actions 槽位整理（见 §1.2） |
| `internal/game/common/kingdom/page.go` | 实现真实识别/点击（读 feature，无 placeholder） |
| `internal/game/arena/feature.go` | 增加入口 `Entry`（OCR 区 + 关键字）；确认 `Lobby.Actions.Close` 语义 |
| `internal/platform/screen/detector.go` + android/stub | 新增 `FindOCRText(region, keyword) (Point, bool)` |
| `internal/game/arena/page.go` | `TapEntry()`、`TapToLobby()` |
| `internal/game/arena/route.go` | 补全 `Enter` / `Leave` |
| `*_test.go` | kingdom + route + FindOCRText stub/mock 单测 |

### 1.2 `kingdom.Feature` 结构

```go
type Feature struct {
    Home      PageSlot // Identify: 王国首页多点比色
    Adventure PageSlot // Identify: 冒险页多点比色
    Actions   HomeActions
}

type PageSlot struct {
    Identify screen.Feature // Colors + Sim → MatchMultiColor
}

type HomeActions struct {
    AdventureBtn screen.Point // 点「冒险」
    BackHome     screen.Point // 从冒险（等）回首页；未配置则 Leave 仅依赖大厅 Close + 有限次返回键策略见 §1.5
}
```

`DefaultFeature()` 返回零值结构，不填数值。

### 1.3 竞技场入口 feature

```go
// 挂在 arena.Feature 上，与 Lobby 平级
type EntryFeature struct {
    Region  screen.Region // 冒险页内搜索「王国竞技场」的 OCR 区
    Keyword string        // 默认逻辑：空则视为 "王国竞技场"
}

type Feature struct {
    Entry      EntryFeature
    Lobby      LobbyFeature
    TeamSelect TeamSelectFeature
    Settlement SettlementFeature
    Dialogs    DialogsFeature
}
```

### 1.4 `FindOCRText`

```go
// Detector 新增
FindOCRText(region Region, keyword string) (Point, bool)
```

- **语义**：在 `region` 截屏后 PPOCR；若某条 `Result.Label` **包含** `keyword`，返回该条中心点的**屏幕绝对坐标**（截屏相对坐标 + region 左上角），否则 `(0,false)`。  
- **android**：复用现有 `OCRText` 的截屏/引擎路径；优先复用引擎实例若后续优化，本段可不强制。  
- **stub**：返回 `(0,false)`。  
- **空 keyword**：直接 `(0,false)`，避免误点。

### 1.5 `route.Enter` / `Leave` / Page 方法

**Enter：**

```
若 page.IsLobby() → true
若 !kingdom.IsAdventurePage():
  若 !kingdom.IsKingdomHome() → false
  TapAdventureBtn → WaitAdventure(30s)；失败 → false
TapEntry（FindOCRText + Tap 中心）→ WaitLobby(30s)
```

**Leave：**

```
若 kingdom.IsKingdomHome() → true
若 page.IsLobby():
  Tap Close（Lobby.Actions.Close 中心）→ 短等
  若仍 Lobby：有限次（如 3）重试 Close
若 !IsKingdomHome() 且 BackHome 已配置（非 0,0）:
  Tap BackHome，轮询至首页或超时（如 15s）
否则：若已不在 Lobby 且在冒险页但无 BackHome → false（要求取色补 BackHome）
超时 → false
```

约定：`screen.Point{0,0}` 视为「未配置」（游戏可点区域一般不在真原点；与现有空 feature 一致）。若真机按钮恰在 (0,0)，取色时改用 1px 偏移——写入取色注意事项。

**`TapToLobby() bool`（保留现有名字，语义定稿）：**

- 若 `IsLobby()` → true  
- 否则点 `Lobby.Actions.Close` 中心，短睡后轮询 `IsLobby`，有限次重试（建议 3）或总超时（建议 10s）  
- 仍失败 → false  

段3 的 `RunBattle` **先**用 `Settlement.Actions.Leave` 离开结算；若已是大厅则成功；若仍非大厅再调用 `TapToLobby` 兜底。

### 1.6 空 feature 行为

- Identify.Colors == "" → `Is*` 为 false（不 Match 空串当成功）  
- Enter 在找不到 OCR / WaitLobby 失败 → false  
- 建议：Colors/Region 未配置时 `logger.Warnf` 一次，避免静默

### 1.7 段1 完成标准

- mock 单测：已在大厅 Enter；首页→冒险→OCR 入口→大厅；Leave 回首页；非首页 Enter 失败  
- 无真机数值断言  
- 硬闸 `teamSelect` **保持**（避免假胜）

---

## 2. 段2 · 取色清单与验收

数值由用户用取色工具填入 `DefaultFeature`（或本地覆盖）；本段以清单 + 真机手册为主，代码仅允许补「空配置 Warn」。

### 2.1 取色总表（1600×900）

#### 路由 / kingdom

| 字段 | 取什么 |
|---|---|
| `kingdom.Home.Identify.Colors/Sim` | 王国首页多点比色 |
| `kingdom.Adventure.Identify.Colors/Sim` | 冒险页多点比色 |
| `kingdom.Actions.AdventureBtn` | 「冒险」按钮点击坐标 |
| `kingdom.Actions.BackHome` | 从冒险回首页（若需要） |
| `arena.Entry.Region` | 冒险页内「王国竞技场」OCR 搜索区 |
| `arena.Entry.Keyword` | 默认可用 `"王国竞技场"`（可写进 DefaultFeature 仅关键字、区仍空） |
| `arena.Lobby.Actions.Close` | 大厅关闭（Region 中心点击；现为 Region） |

#### 大厅识别（已有，recognition spec）

| 字段 | 取什么 |
|---|---|
| `Lobby.Identify` | 大厅多点比色 |
| `Lobby.Reads.*` | 勋章门票 / 奖杯 / 刷新倒计时 / 免费刷新 文案区 |
| `Lobby.Actions.FreeRefresh` 等 | 免费刷新、买票滑动与确认 |
| `Lobby.Opponent.*` | 锚点 ColorFind、奖杯偏移、战绩色、点击偏移 |
| `Lobby.Gestures.SwipeLeft` | 翻页滑动 |

#### 战斗（段3 填）

| 字段 | 取什么 |
|---|---|
| `TeamSelect.Identify` | 选将页多点比色；**Colors 空 = 无选将页，跳过等待** |
| `TeamSelect.Actions.StartBattle` | 开战按钮 |
| `Settlement.Identify` | 结算页多点比色 |
| `Settlement.Reads.Result` | 胜负 OCR 区（Rect→当 Region 用） |
| `Settlement.Actions.LeaveIdentify` | 可选：离开钮可见性 |
| `Settlement.Actions.Leave` | 离开点击区 |

### 2.2 真机验收顺序

1. 只填 §路由 → 任务应能进大厅，并在 leave（或选将硬闸前）可关回首页  
2. 填大厅识别 → sync / 选对手 / 刷新逻辑可测  
3. 填战斗 → 打完一场回大厅  

### 2.3 段2 完成标准

- 本 spec 清单与代码字段路径一致（实现段1/3 时若改名须改本表）  
- 用户按序验收；仓库可不提交具体颜色串  

---

## 3. 段3 · 最小战斗闭环

### 3.1 状态机

保持：

```
detect → navigate → sync → check → selectOpponent → teamSelect → battle → sync
                                                              ↘ leave
```

变更：

- 移除 `teamSelect` 的 `Fatal("选将/战斗尚未实现")`  
- `battle` 调用真实 `RunBattle`；`ok==false` → Fatal  

不新增 `settlement` 状态。

### 3.2 `teamSelect` 逻辑

```
若 selectedOpponent == nil → Fatal
Tap(selectedOpponent.Site)
若 TeamSelect.Identify.Colors != "":
  等待 Identify 命中或超时 → 超时则 Retry/Fatal（与 MaxRetry 对齐）
否则: 跳过等待（无选将页）
Tap(StartBattle)
→ Next("battle")
```

### 3.3 `RunBattle() (result string, ok bool)`

常量建议：`battleTimeout = 3 * time.Minute`（可包内 const）。

```
deadline := now + battleTimeout
循环直至 Settlement.Identify 命中或超时:
  sleep 短间隔
超时 → ("", false)
OCR Settlement.Reads.Result → 规范化为 "胜利"|"平局"|"失败"（含 contains 匹配）
无法解析 → ("", false) 或记一次 Retry 由上层处理；定稿：返回 ("", false)
点 Leave（Leave 矩形中心；若 LeaveIdentify 有配置则先等到可见）
直到 IsLobby 或子超时（如 30s）:
  失败 → ("", false)
返回 (result, true)
```

`battle` handler：按 result 增加 Wins/Draws/Losses；Tickets--；`Next("sync")`。

### 3.4 测试

- 替换硬闸测试为：teamSelect 点击 + 无选将页跳过 + RunBattle 三结果 + 超时失败  
- route 测试保持  

### 3.5 段3 完成标准

- 填完战斗取色后真机一场闭环  
- 未填 Settlement/TeamSelect 时：明确失败（false/Fatal），**禁止**返回固定「胜利」  

---

## 4. 实施顺序与依赖

```
段1 FindOCRText + kingdom + route + TapToLobby/Close
  → 段2 用户取色（可与段1 文档并行）
  → 段3 teamSelect + RunBattle + 测例
```

每段独立 PR/提交均可；段3 依赖段1 的离开能力。

---

## 5. 风险与注意事项

| 风险 | 缓解 |
|---|---|
| PPOCR 中心点在裁剪图坐标系 | FindOCRText 必须加 region 偏移 |
| (0,0) 表示未配置 | 写入取色说明；真机勿用原点 |
| 无选将页 | Identify.Colors=="" 跳过 |
| OCR 入口不稳定 | Enter 可有限次重试 FindOCRText（≤2），仍失败再 false |
| 引擎每次 New | 已知债；本设计不强制修，可另开优化 |

---

## 6. 文档同步（实现时）

- `docs/开发手册.md`：模块文件约定若 kingdom 槽位变更则补一句  
- 不改 README 架构摘要除非入口行为变化  

---

## 7. 修订记录

| 日期 | 说明 |
|---|---|
| 2026-07-12 | 初稿：路由 + 取色清单 + 最小战斗；用户确认方案 1 / 数值 A / 战斗 A |
