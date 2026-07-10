# 王国竞技场 · 识别层补全 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补全 `internal/game/arena` 的识别层，让 `Page` 的识别/读数/选对手方法在 `feature` 填值后能正确工作，且本地 mock 可测；并把全部静态识别参数收进 `feature.go`。

**Architecture:** `screen.Detector` 新增 `FindMultiColorsAll`（android 调 `images.FindMultiColorsAll`，stub 返 nil）；`feature.go` 引入 `ColorFind` 描述结构并把 `OpponentFeature` 改为「锚点 + 相对偏移」模型；`page.go` 用纯函数辅助（`offsetRegion/offsetPoint/readInt/parseCountdown/battledAt`）+ 9 个方法实现识别，零颜色/坐标硬编码；测试用 mock `Detector/Executor` 表驱动。

**Tech Stack:** Go 1.x，标准库 `testing`/`strconv`/`strings`/`regexp`/`time`；AutoGo `images` 绑定（`FindMultiColorsAll`/`CmpColor`/`DetectsMultiColors`）；跨平台 build tag（`android && cgo` vs `!android || !cgo`）。

## Global Constraints

- 本段**只补识别层**：不动 `teamSelect`/`RunBattle`/`route`/状态机。
- `feature.go` **只补字段结构、不填数值**；数值由用户用取色工具按取色表填（Task 6）。
- 代码侧**零颜色/坐标硬编码**；所有静态参数入 `feature`，UI 变化只改 `feature`。
- 跨平台必须同时通过 `go build ./...` 与 `go build -tags android ./...`。
- 本段**不在本地断言真机识别结果**；真机核对在 Task 7（上机清单）。
- `config.Arena.TrophyDiff` 语义为**区间半宽**：`diff=0` = 严格相等（不是"不过滤"）。
- 坐标域：`screen.Point/Region` 给 `Detector`；`action.Point` 给 `Executor`。`OpponentInfo.Site` 是 **`action.Point`**（page.go:14），计算后必须转换。

---

## 对 spec 的有意偏离（执行前请确认）

本计划在落到真实代码时发现 spec 两处需调整，已在对应任务里采用下列处理；如你不同意，执行前提出，我按你的意见改回：

1. **不引入 `ColorCmp` 类型**（spec §4.1 列了它）。原因：`battledAt` 直接对 `ResultColors.Win/Draw/Lose` 三色各调一次 `Detector.MatchColor(x,y,color,sim)`，没有消费 `ColorCmp` 结构的地方——引入即死代码（违反 YAGNI）。只新增 `ColorFind`（`FindMultiColorsAll` 真正消费）。
2. **`battledAt` 语义改为「命中任一胜负色 = 已战(true)；三色都不命中 = 未战(false)」**。spec §5.4/§7 原文是「全部未命中或异常 → 视为已战（保守）」。问题：`Detector.MatchColor` 只返 `bool`、无 `error` 通道，代码层**无法区分"正常未命中（未战卡的中性态）"与"比色异常"**；若按 spec 把"全未命中"当已战，则**所有未战卡都会被跳过 → 永远选不到对手**，功能失效。故采用：命中任一胜负色才跳过；都不命中视为未战、进入奖杯区间判断。用户原话"异常当已战"以代码注释保留——待 `MatchColor` 未来提供 error 通道再实现。

---

## File Structure

| 文件 | 责任 | 改动 |
|---|---|---|
| `internal/platform/screen/detector.go` | `Detector` 接口 | 新增 `FindMultiColorsAll` 方法签名 |
| `internal/platform/screen/color.go` | Android 实现（`android && cgo`） | 实现 `FindMultiColorsAll`（包装 `images.FindMultiColorsAll`） |
| `internal/platform/screen/factory.go` | stub 实现（`!android \|\| !cgo`） | 实现 `FindMultiColorsAll`（返 nil） |
| `internal/game/arena/feature.go` | 静态识别参数结构 | 新增 `ColorFind`；改造 `OpponentFeature`（删 `BaseSite`/`FindDef`，加 `Anchor`/`ResultSim`/`ClickOffset`） |
| `internal/game/arena/page.go` | 页面识别/读数/选对手 | 实现 9 方法 + 5 私有辅助 |
| `internal/config/static.go` | `config.Arena` 默认值 | 给 `TrophyDiff` 字段加语义注释 |
| `internal/game/arena/page_test.go` | mock 单元测试（新建） | mock + 表驱动用例 |

依赖顺序：Task 1（接口）→ Task 2（feature 结构）→ Task 3（纯函数辅助）→ Task 4（读数/判定小方法）→ Task 5（`FindFirstValidOpponent`）→ Task 6（用户填 feature，手动）→ Task 7（真机核对，手动）。

---

### Task 1: Detector 接口扩展（接口 + android + stub 必须同 task，否则任一侧编译断）

**Files:**
- Modify: `internal/platform/screen/detector.go:31-38`（接口块）
- Modify: `internal/platform/screen/color.go`（追加方法到 `AndroidDetector`）
- Modify: `internal/platform/screen/factory.go`（追加方法到 `stubDetector`）
- Test: `internal/platform/screen/factory_test.go`（新建，验证 stub 不 panic 且返 nil）

**Interfaces:**
- Consumes: `images.FindMultiColorsAll(x1,y1,x2,y2 int, colors string, sim float32, dir, displayId int) []images.Point`；`images.Point{X,Y int}`
- Produces: `screen.Detector.FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point`（后续 Task 5 调用）

- [ ] **Step 1: 写 stub 测试（红）**

新建 `internal/platform/screen/factory_test.go`：

```go
//go:build !android || !cgo

package screen

import "testing"

func TestStubFindMultiColorsAllReturnsNil(t *testing.T) {
	d := NewDetector(0)
	got := d.FindMultiColorsAll(Region{0, 0, 100, 100}, "ffffff", 0.9, 0)
	if got != nil {
		t.Fatalf("stub FindMultiColorsAll should return nil, got %v", got)
	}
}
```

- [ ] **Step 2: 跑测试确认红**

Run: `go test ./internal/platform/screen/ -run TestStubFindMultiColorsAllReturnsNil -v`
Expected: 编译失败 —— `d.FindMultiColorsAll undefined`（接口尚未有该方法）。

- [ ] **Step 3: 在 `Detector` 接口加方法签名**

`internal/platform/screen/detector.go`，把接口块改为：

```go
type Detector interface {
	Capture() *image.NRGBA
	MatchColor(x, y int, color string, sim float32) bool
	FindColor(region Region, color string, sim float32, dir int) (Point, bool)
	FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point
	MatchMultiColor(colors string, sim float32) bool
	MatchImage(region Region, template []byte, sim float32) (Point, bool)
	OCRText(region Region) string
}
```

- [ ] **Step 4: Android 实现 `FindMultiColorsAll`**

在 `internal/platform/screen/color.go` 末尾追加（文件首行 `//go:build android && cgo` 保持不变，`images` 已 import）：

```go
func (d *AndroidDetector) FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point {
	pts := images.FindMultiColorsAll(region.X1, region.Y1, region.X2, region.Y2, colors, sim, dir, d.displayId)
	out := make([]Point, len(pts))
	for i, p := range pts {
		out[i] = Point{X: p.X, Y: p.Y}
	}
	return out
}
```

- [ ] **Step 5: stub 实现 `FindMultiColorsAll`**

在 `internal/platform/screen/factory.go` 末尾追加（`stubDetector` 已 import 无新增）：

```go
func (s *stubDetector) FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point {
	return nil
}
```

- [ ] **Step 6: 跑测试确认绿 + 双侧 build**

Run:
- `go test ./internal/platform/screen/ -run TestStubFindMultiColorsAllReturnsNil -v` → Expected: PASS
- `go build ./...` → Expected: 无输出（成功）
- `go build -tags android ./...` → Expected: 无输出（成功；若本机无 cgo/android 工具链导致该命令本身不可用，记录错误但不得因此改接口——以 `go build ./...` 为准，Android 侧由 CI/真机插件验证）

- [ ] **Step 7: Commit**

```bash
git add internal/platform/screen/detector.go internal/platform/screen/color.go internal/platform/screen/factory.go internal/platform/screen/factory_test.go
git commit -m "feat(screen): add FindMultiColorsAll to Detector (android+stub)"
```

---

### Task 2: feature.go 结构（ColorFind + OpponentFeature 改造，不填数值）

**Files:**
- Modify: `internal/game/arena/feature.go:42-49`（`OpponentFeature`）
- Modify: `internal/game/arena/feature.go`（`ColorFind` 新类型，追加在 `ResultColors` 之前）
- Modify: `internal/config/static.go:7`（`TrophyDiff` 注释）

**Interfaces:**
- Consumes: `screen.Region`、`screen.Point`、`screen.OCRCfg`（已存在）
- Produces: `arena.ColorFind{Region, Colors, Sim, Dir}`；`arena.OpponentFeature{Anchor, TrophyRect, ResultOffset, ResultColors, ResultSim, ClickOffset, NumberOCR}`（Task 3/5 使用）

- [ ] **Step 1: 删除 `OpponentFeature` 旧字段并替换为新结构**

把 `internal/game/arena/feature.go` 中：

```go
type OpponentFeature struct {
	FindDef      screen.FindDef
	BaseSite     screen.Point
	TrophyRect   screen.Region
	ResultOffset screen.Point
	ResultColors ResultColors
	NumberOCR    screen.OCRCfg
}
```

替换为：

```go
type OpponentFeature struct {
	Anchor       ColorFind     // 找卡锚点：Region=搜索区, Colors=锚点颜色串(单/多点), Sim, Dir
	TrophyRect   screen.Region // 相对锚点的奖杯 OCR 偏移矩形 (dx1,dy1,dx2,dy2)
	ResultOffset screen.Point  // 相对锚点的战绩标记点偏移
	ResultColors ResultColors  // 已战颜色 {Win,Draw,Lose}
	ResultSim    float32       // 战绩 MatchColor 相似度
	ClickOffset  screen.Point  // 相对锚点的点击偏移；锚点本身可点则 (0,0)
	NumberOCR    screen.OCRCfg // 奖杯数字 OCR 配置（占位，留待 OCR 不稳时补参）
}
```

- [ ] **Step 2: 新增 `ColorFind` 类型**

在 `internal/game/arena/feature.go` 的 `type ResultColors struct { ... }` 之前插入：

```go
// ColorFind 对齐 images.FindColor / FindMultiColors[All] 的全部静态参数。
// Colors 写单色串即"找首个/单点"，写多点串即"多点匹配"；返回首个还是全部由代码侧选择，字段不变。
type ColorFind struct {
	Region screen.Region // x1,y1,x2,y2；右下 0,0 = 屏幕最大
	Colors string        // 单色串或多点串（取色工具直接拷贝）
	Sim    float32       // 0.1–1.0
	Dir    int           // 0 左上起 / 1 右上 / 2 左下 / 3 右下；本段默认 0
}
```

- [ ] **Step 3: 给 `config.Arena.TrophyDiff` 加语义注释**

先 `Read internal/config/static.go` 确认 `type Arena struct` 字段位置（约第 3-9 行），在 `TrophyDiff int \`json:"trophyDiff"\`` 行**上方**插入注释：

```go
	// TrophyDiff：选对手奖杯区间半宽。只打奖杯在 [myTrophy-TrophyDiff, myTrophy+TrophyDiff] 内的对手。
	// 注意：0 表示"严格相等"（只打奖杯完全等于自己的对手），不是"不过滤"。
	TrophyDiff   int  `json:"trophyDiff"`
```

（保持其它字段不动；若字段顺序与现状不同，仅在该字段上方加注释即可。）

- [ ] **Step 4: 编译确认**

Run: `go build ./...`
Expected: 无输出（成功）。`DefaultFeature()` 仍返回 `&Feature{}`，零值 `ColorFind` 合法。

- [ ] **Step 5: Commit**

```bash
git add internal/game/arena/feature.go internal/config/static.go
git commit -m "refactor(arena): reshape OpponentFeature around anchor+offset; add ColorFind"
```

---

### Task 3: page 私有纯函数辅助（offsetRegion / offsetPoint / readInt / parseCountdown）+ 单测

**Files:**
- Modify: `internal/game/arena/page.go`（追加文件级私有函数；import 加 `regexp`）
- Test: `internal/game/arena/page_test.go`（新建；本 task 先放纯函数测试，Task 4/5 续加 mock 测试）

**Interfaces:**
- Consumes: `screen.Region`、`screen.Point`
- Produces:
  - `func offsetRegion(rel screen.Region, a screen.Point) screen.Region`
  - `func offsetPoint(rel screen.Point, a screen.Point) screen.Point`
  - `func readInt(s string) (int, bool)`
  - `func parseCountdown(s string) (time.Duration, bool)`
  （Task 4/5 调用）

- [ ] **Step 1: 写纯函数测试（红）**

新建 `internal/game/arena/page_test.go`：

```go
package arena

import (
	"testing"
	"time"

	"app/internal/platform/screen"
)

func TestOffsetRegion(t *testing.T) {
	rel := screen.Region{X1: -10, Y1: -20, X2: 10, Y2: 20}
	a := screen.Point{X: 100, Y: 200}
	got := offsetRegion(rel, a)
	want := screen.Region{X1: 90, Y1: 180, X2: 110, Y2: 220}
	if got != want {
		t.Fatalf("offsetRegion = %+v, want %+v", got, want)
	}
}

func TestOffsetPoint(t *testing.T) {
	got := offsetPoint(screen.Point{X: 30, Y: -5}, screen.Point{X: 100, Y: 200})
	want := screen.Point{X: 130, Y: 195}
	if got != want {
		t.Fatalf("offsetPoint = %+v, want %+v", got, want)
	}
}

func TestReadInt(t *testing.T) {
	cases := []struct {
		in     string
		wantN  int
		wantOK bool
	}{
		{"1050", 1050, true},
		{"  99 ", 99, true},
		{"abc", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		n, ok := readInt(c.in)
		if n != c.wantN || ok != c.wantOK {
			t.Errorf("readInt(%q) = (%d,%v), want (%d,%v)", c.in, n, ok, c.wantN, c.wantOK)
		}
	}
}

func TestParseCountdown(t *testing.T) {
	cases := []struct {
		in     string
		want   time.Duration
		wantOK bool
	}{
		{"5分30秒", 330 * time.Second, true},
		{"30秒", 30 * time.Second, true},
		{"5分", 300 * time.Second, true},
		{"05:30", 330 * time.Second, true},
		{"", 0, false},
		{"abc", 0, false},
		{"0秒", 0, false},
	}
	for _, c := range cases {
		got, ok := parseCountdown(c.in)
		if got != c.want || ok != c.wantOK {
			t.Errorf("parseCountdown(%q) = (%v,%v), want (%v,%v)", c.in, got, ok, c.want, c.wantOK)
		}
	}
}
```

- [ ] **Step 2: 跑测试确认红**

Run: `go test ./internal/game/arena/ -run 'TestOffset|TestReadInt|TestParseCountdown' -v`
Expected: 编译失败 —— `offsetRegion/offsetPoint/readInt/parseCountdown` 未定义。

- [ ] **Step 3: 实现纯函数**

在 `internal/game/arena/page.go` 顶部 import 加入 `"regexp"`（现有 import：`strconv`/`strings`/`time`/`config`/`action`/`screen`），并在文件末尾（`TapToLobby` 之后）追加：

```go
func offsetRegion(rel screen.Region, a screen.Point) screen.Region {
	return screen.Region{X1: a.X + rel.X1, Y1: a.Y + rel.Y1, X2: a.X + rel.X2, Y2: a.Y + rel.Y2}
}

func offsetPoint(rel screen.Point, a screen.Point) screen.Point {
	return screen.Point{X: a.X + rel.X, Y: a.Y + rel.Y}
}

func readInt(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	return n, err == nil
}

var (
	reColon = regexp.MustCompile(`^\s*(\d{1,3}):(\d{1,2})\s*$`)
	reMin   = regexp.MustCompile(`(\d+)\s*分`)
	reSec   = regexp.MustCompile(`(\d+)\s*秒`)
)

// parseCountdown 解析刷新倒计时。支持 "5分30秒"/"30秒"/"5分"/"05:30"。
// 抓不到数字或合计为 0 → (0, false)。
func parseCountdown(s string) (time.Duration, bool) {
	if m := reColon.FindStringSubmatch(s); m != nil {
		min, _ := strconv.Atoi(m[1])
		sec, _ := strconv.Atoi(m[2])
		d := time.Duration(min)*time.Minute + time.Duration(sec)*time.Second
		if d == 0 {
			return 0, false
		}
		return d, true
	}
	total := 0
	if m := reMin.FindStringSubmatch(s); m != nil {
		v, _ := strconv.Atoi(m[1])
		total += v * 60
	}
	if m := reSec.FindStringSubmatch(s); m != nil {
		v, _ := strconv.Atoi(m[1])
		total += v
	}
	if total == 0 {
		return 0, false
	}
	return time.Duration(total) * time.Second, true
}
```

- [ ] **Step 4: 跑测试确认绿**

Run: `go test ./internal/game/arena/ -run 'TestOffset|TestReadInt|TestParseCountdown' -v`
Expected: PASS（4 个用例全过）。

- [ ] **Step 5: Commit**

```bash
git add internal/game/arena/page.go internal/game/arena/page_test.go
git commit -m "feat(arena): add offset/readInt/parseCountdown helpers with unit tests"
```

---

### Task 4: 读数/判定/动作小方法（IsLobby 改用 feature、ReadRefreshCountdown 实现、其余补测试）

**Files:**
- Modify: `internal/game/arena/page.go:30-32`（`IsLobby` 改用 feature）
- Modify: `internal/game/arena/page.go:87-93`（`ReadRefreshCountdown` 调 `parseCountdown`）
- Test: `internal/game/arena/page_test.go`（追加 mock + 小方法测试）

**Interfaces:**
- Consumes: `screen.Detector{MatchMultiColor, OCRText}`、`action.Executor{Tap, Swipe, Sleep}`、`Feature{Lobby.Identify/Reads/Actions/Gestures}`、Task 3 的 `parseCountdown`
- Produces: `IsLobby/IsFreeRefresh/ReadMedalAndTicket/ReadTrophyCount/ReadRefreshCountdown/TapFreeRefresh/SwipePageLeft` 行为定型（Task 5 不依赖，状态机后续段依赖）

- [ ] **Step 1: 在 `page_test.go` 追加 mock（文件顶部 `import` 之后）**

```go
// ---- mock Detector ----
type mockDetector struct {
	matchMulti bool
	ocrByKey   map[string]string // key: "x1,y1,x2,y2"
	matchByKey map[string]bool   // key: "x,y,color"
	anchors    []screen.Point
}

func (m *mockDetector) Capture() *image.NRGBA { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool {
	return m.matchByKey[fmt.Sprintf("%d,%d,%s", x, y, color)]
}
func (m *mockDetector) FindColor(r screen.Region, c string, s float32, d int) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) FindMultiColorsAll(r screen.Region, c string, s float32, d int) []screen.Point {
	return m.anchors
}
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool { return m.matchMulti }
func (m *mockDetector) MatchImage(r screen.Region, t []byte, s float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(r screen.Region) string {
	return m.ocrByKey[fmt.Sprintf("%d,%d,%d,%d", r.X1, r.Y1, r.X2, r.Y2)]
}

// ---- mock Executor ----
type mockExecutor struct {
	taps   []action.Point
	swipes [][2]action.Point
	sleeps []int
}

func (e *mockExecutor) Tap(p action.Point) error            { e.taps = append(e.taps, p); return nil }
func (e *mockExecutor) LongTap(p action.Point, ms int) error { return nil }
func (e *mockExecutor) Swipe(f, t action.Point, ms int) error {
	e.swipes = append(e.swipes, [2]action.Point{f, t}); return nil
}
func (e *mockExecutor) Back() error { return nil }
func (e *mockExecutor) Home() error { return nil }
func (e *mockExecutor) Sleep(ms int) { e.sleeps = append(e.sleeps, ms) }
```

并在 `page_test.go` 的 `import` 块追加：`"fmt"`、`"image"`、`"app/internal/platform/action"`（`screen`/`testing`/`time` 已在 Task 3 引入）。

- [ ] **Step 2: 跑确认仍编译（mock 暂未使用，Go 允许未用类型但 `fmt`/`image` 一旦 import 必须用——本 step 先不写用到它们的测试会报 unused import）**

为避免 unused import，本 step 与 Step 3 一起做：先不要单独编译；直接进入 Step 3 写测试（测试会用到 `fmt.Sprintf` 和 `image.NRGBA`）。

- [ ] **Step 3: 写小方法测试（红，针对 `IsLobby` 改动与 `ReadRefreshCountdown`）**

追加到 `page_test.go`：

```go
func newLobbyFeature() *Feature {
	f := DefaultFeature()
	f.Lobby.Identify = screen.Feature{Colors: "lobby", Sim: 0.95}
	f.Lobby.Reads.MedalTicket = screen.Region{1, 1, 100, 30}
	f.Lobby.Reads.Trophy = screen.Region{1, 40, 100, 70}
	f.Lobby.Reads.Refresh = screen.Region{1, 80, 100, 110}
	f.Lobby.Reads.FreeRefresh = screen.Region{1, 120, 100, 150}
	f.Lobby.Actions.FreeRefresh = screen.Point{X: 300, Y: 800}
	f.Lobby.Gestures.SwipeLeft = action.Swipe{
		From: action.Point{X: 1400, Y: 450}, To: action.Point{X: 200, Y: 450}, DurationMs: 300}
	return f
}

func TestIsLobby(t *testing.T) {
	d := &mockDetector{matchMulti: true}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	if !p.IsLobby() {
		t.Fatal("IsLobby should be true when MatchMultiColor returns true")
	}
	d.matchMulti = false
	if p.IsLobby() {
		t.Fatal("IsLobby should be false when MatchMultiColor returns false")
	}
}

func TestIsFreeRefresh(t *testing.T) {
	d := &mockDetector{ocrByKey: map[string]string{"1,120,100,150": "免费刷新"}}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	if !p.IsFreeRefresh() {
		t.Fatal("IsFreeRefresh should be true for '免费刷新'")
	}
	d.ocrByKey["1,120,100,150"] = "  刷新  "
	if p.IsFreeRefresh() {
		t.Fatal("IsFreeRefresh should be false for non-exact text")
	}
}

func TestReadMedalAndTicket(t *testing.T) {
	d := &mockDetector{ocrByKey: map[string]string{"1,1,100,30": "123 45"}}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	m, tk, ok := p.ReadMedalAndTicket()
	if !ok || m != 123 || tk != 45 {
		t.Fatalf("ReadMedalAndTicket = (%d,%d,%v), want (123,45,true)", m, tk, ok)
	}
	d.ocrByKey["1,1,100,30"] = "onlyone"
	if _, _, ok := p.ReadMedalAndTicket(); ok {
		t.Fatal("single token should be ok=false")
	}
	d.ocrByKey["1,1,100,30"] = "abc def"
	if _, _, ok := p.ReadMedalAndTicket(); ok {
		t.Fatal("non-numeric should be ok=false")
	}
}

func TestReadTrophyCount(t *testing.T) {
	d := &mockDetector{ocrByKey: map[string]string{"1,40,100,70": "  1024 "}}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	n, ok := p.ReadTrophyCount()
	if !ok || n != 1024 {
		t.Fatalf("ReadTrophyCount = (%d,%v), want (1024,true)", n, ok)
	}
	d.ocrByKey["1,40,100,70"] = "x"
	if _, ok := p.ReadTrophyCount(); ok {
		t.Fatal("non-numeric trophy should be ok=false")
	}
}

func TestReadRefreshCountdown(t *testing.T) {
	d := &mockDetector{ocrByKey: map[string]string{"1,80,100,110": "5分30秒"}}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	got, ok := p.ReadRefreshCountdown()
	if !ok || got != 330*time.Second {
		t.Fatalf("ReadRefreshCountdown = (%v,%v), want (330s,true)", got, ok)
	}
}

func TestTapFreeRefresh(t *testing.T) {
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, newLobbyFeature())
	p.TapFreeRefresh()
	if len(e.taps) != 1 || e.taps[0] != (action.Point{X: 300, Y: 800}) {
		t.Fatalf("TapFreeRefresh taps = %v, want [(300,800)]", e.taps)
	}
	if len(e.sleeps) == 0 {
		t.Fatal("TapFreeRefresh should sleep after tap")
	}
}

func TestSwipePageLeft(t *testing.T) {
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, newLobbyFeature())
	p.SwipePageLeft()
	if len(e.swipes) != 1 {
		t.Fatalf("SwipePageLeft swipes = %v, want 1 swipe", e.swipes)
	}
	if e.swipes[0][0] != (action.Point{X: 1400, Y: 450}) || e.swipes[0][1] != (action.Point{X: 200, Y: 450}) {
		t.Fatalf("SwipePageLeft swipe endpoints = %v", e.swipes[0])
	}
}
```

- [ ] **Step 4: 跑确认红（`IsLobby` 仍用 `MatchMultiColor("",0.95)` 占位也能过 matchMulti，但 `ReadRefreshCountdown` 占位返 `(0,false)` 会红）**

Run: `go test ./internal/game/arena/ -run 'TestIsLobby|TestIsFreeRefresh|TestReadMedalAndTicket|TestReadTrophyCount|TestReadRefreshCountdown|TestTapFreeRefresh|TestSwipePageLeft' -v`
Expected: `TestReadRefreshCountdown` FAIL（占位返 0,false）；其余因现状已实现应 PASS（`IsLobby` 占位凑巧过——下一步改为引用 feature，行为不变但消除硬编码）。

- [ ] **Step 5: 改 `IsLobby` 为引用 feature**

把 `internal/game/arena/page.go:30-32`：

```go
func (p *Page) IsLobby() bool {
	return p.detector.MatchMultiColor("", 0.95) // use feature
}
```

改为：

```go
func (p *Page) IsLobby() bool {
	id := p.feature.Lobby.Identify
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}
```

- [ ] **Step 6: 实现 `ReadRefreshCountdown`**

把 `internal/game/arena/page.go:87-93`：

```go
func (p *Page) ReadRefreshCountdown() (time.Duration, bool) {
	text := p.detector.OCRText(p.feature.Lobby.Reads.Refresh)
	// Parse text like "5分30秒" or "30秒"
	// Placeholder
	_ = text
	return 0, false
}
```

改为：

```go
func (p *Page) ReadRefreshCountdown() (time.Duration, bool) {
	text := strings.TrimSpace(p.detector.OCRText(p.feature.Lobby.Reads.Refresh))
	return parseCountdown(text)
}
```

- [ ] **Step 7: 跑全部 page 测试确认绿**

Run: `go test ./internal/game/arena/ -v`
Expected: 全部 PASS（含 Task 3 的纯函数测试）。

- [ ] **Step 8: Commit**

```bash
git add internal/game/arena/page.go internal/game/arena/page_test.go
git commit -m "feat(arena): wire IsLobby to feature; implement ReadRefreshCountdown; add page mock tests"
```

---

### Task 5: FindFirstValidOpponent + battledAt（核心，锚点扫描 + 区间 + 已战跳过）

**Files:**
- Modify: `internal/game/arena/page.go:65-68`（`FindFirstValidOpponent` 占位 → 实现；追加 `battledAt`）
- Test: `internal/game/arena/page_test.go`（追加对手扫描用例）

**Interfaces:**
- Consumes: `Detector.FindMultiColorsAll/OCRText/MatchColor`；`config.Arena{TrophyDiff}`；`OpponentFeature{Anchor,TrophyRect,ResultOffset,ResultColors,ResultSim,ClickOffset}`；Task 3 的 `offsetRegion/offsetPoint/readInt`；`action.Point`
- Produces: `func (p *Page) FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo` 与 `func battledAt(det screen.Detector, pt screen.Point, op OpponentFeature) bool`（状态机 `selectOpponent` 后续段消费 `OpponentInfo.Site`）

> 语义（**偏离点 2**）：`battledAt` 命中任一 `Win/Draw/Lose` 色 → 已战（跳过）；三色都不命中 → 未战（进入区间判断）。`MatchColor` 无 error 通道，无法表达"比色异常"，故"异常当已战"以注释保留，不实现。

- [ ] **Step 1: 写对手扫描测试（红）**

追加到 `page_test.go`：

```go
func newOpponentFeature() *Feature {
	f := DefaultFeature()
	op := &f.Lobby.Opponent
	op.Anchor = ColorFind{Region: screen.Region{0, 0, 1600, 900}, Colors: "anchor", Sim: 0.95, Dir: 0}
	op.TrophyRect = screen.Region{X1: -10, Y1: -20, X2: 10, Y2: 20} // 相对锚点偏移
	op.ResultOffset = screen.Point{X: 30, Y: 0}
	op.ResultColors = ResultColors{Win: "w", Draw: "d", Lose: "l"}
	op.ResultSim = 0.9
	op.ClickOffset = screen.Point{X: 5, Y: 5}
	return f
}

func key(r screen.Region) string { return fmt.Sprintf("%d,%d,%d,%d", r.X1, r.Y1, r.X2, r.Y2) }

func TestFindFirstValidOpponent(t *testing.T) {
	cfg := &config.Arena{TrophyDiff: 50} // 区间 [950,1050] 当 myTrophy=1000

	t.Run("in-range hit returns first with offset Site", func(t *testing.T) {
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}},
			ocrByKey: map[string]string{
				key(screen.Region{90, 180, 110, 220}): "1050", // anchor1 奖杯，区间内
			},
			matchByKey: map[string]bool{}, // 三色都不命中 → 未战
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(cfg, 1000)
		if got == nil {
			t.Fatal("want a valid opponent, got nil")
		}
		if got.Site != (action.Point{X: 105, Y: 205}) { // anchor(100,200)+ClickOffset(5,5)
			t.Errorf("Site = %+v, want (105,205)", got.Site)
		}
		if got.Trophies != 1050 || got.IsBattled {
			t.Errorf("Trophies/IsBattled = %d/%v, want 1050/false", got.Trophies, got.IsBattled)
		}
	})

	t.Run("battled card is skipped", func(t *testing.T) {
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}},
			ocrByKey: map[string]string{
				key(screen.Region{90, 180, 110, 220}): "1000", // anchor1 区间内但已战
				key(screen.Region{90, 380, 110, 420}): "1020", // anchor2 区间内未战
			},
			matchByKey: map[string]bool{
				"130,200,w": true, // anchor1 ResultOffset(100+30,200) 命中 Win → 已战
			},
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(cfg, 1000)
		if got == nil || got.Trophies != 1020 {
			t.Fatalf("want anchor2 (1020), got %+v", got)
		}
	})

	t.Run("out-of-range is skipped", func(t *testing.T) {
		d := &mockDetector{
			anchors:  []screen.Point{{X: 100, Y: 200}},
			ocrByKey: map[string]string{key(screen.Region{90, 180, 110, 220}): "2000"}, // >1050
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		if got := p.FindFirstValidOpponent(cfg, 1000); got != nil {
			t.Fatalf("out-of-range should yield nil, got %+v", got)
		}
	})

	t.Run("no anchors returns nil", func(t *testing.T) {
		d := &mockDetector{anchors: nil}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		if got := p.FindFirstValidOpponent(cfg, 1000); got != nil {
			t.Fatalf("no anchors should yield nil, got %+v", got)
		}
	})

	t.Run("OCR failure skips that anchor", func(t *testing.T) {
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}},
			ocrByKey: map[string]string{
				key(screen.Region{90, 180, 110, 220}): "abc",  // OCR 失败 → 跳过
				key(screen.Region{90, 380, 110, 420}): "1010", // 区间内未战
			},
			matchByKey: map[string]bool{},
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(cfg, 1000)
		if got == nil || got.Trophies != 1010 {
			t.Fatalf("want anchor2 (1010), got %+v", got)
		}
	})

	t.Run("first in-range wins by anchor order", func(t *testing.T) {
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}}, // dir=0 已按 Y 排
			ocrByKey: map[string]string{
				key(screen.Region{90, 180, 110, 220}): "960",  // 区间内（第一个）
				key(screen.Region{90, 380, 110, 420}): "1040", // 也在区间内但应被前者抢先
			},
			matchByKey: map[string]bool{},
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(cfg, 1000)
		if got == nil || got.Trophies != 960 {
			t.Fatalf("want first anchor (960), got %+v", got)
		}
	})

	t.Run("zero diff means strict equality", func(t *testing.T) {
		eq := &config.Arena{TrophyDiff: 0} // 只接受奖杯==myTrophy
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}},
			ocrByKey: map[string]string{
				key(screen.Region{90, 180, 110, 220}): "999",  // 不等 → 跳
				key(screen.Region{90, 380, 110, 420}): "1000", // 严格相等 → 中
			},
			matchByKey: map[string]bool{},
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(eq, 1000)
		if got == nil || got.Trophies != 1000 {
			t.Fatalf("diff=0 should only match exact trophy, got %+v", got)
		}
	})
}
```

并在 `page_test.go` 的 import 追加 `"app/internal/config"`（`action`/`fmt`/`image`/`screen`/`testing`/`time` 已引入）。

- [ ] **Step 2: 跑确认红**

Run: `go test ./internal/game/arena/ -run TestFindFirstValidOpponent -v`
Expected: FAIL —— `FindFirstValidOpponent` 当前 `return nil`，所有"want 非 nil"子用例失败。

- [ ] **Step 3: 实现 `battledAt` 与 `FindFirstValidOpponent`**

在 `internal/game/arena/page.go` 把占位：

```go
func (p *Page) FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo {
	// Placeholder: implement opponent scanning using feature.Lobby.Opponent
	return nil
}
```

替换为：

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
		if battledAt(p.detector, offsetPoint(op.ResultOffset, a), op) {
			continue // 已战：跳过
		}
		if trophy < lo || trophy > hi {
			continue // 奖杯不在区间：跳过
		}
		return &OpponentInfo{
			Site:      action.Point{X: a.X + op.ClickOffset.X, Y: a.Y + op.ClickOffset.Y},
			Trophies:  trophy,
			IsBattled: false,
		}
	}
	return nil
}
```

并在文件末尾（`parseCountdown` 之后）追加：

```go
// battledAt 判断结果点 pt 是否显示已战标记色。
// 命中 Win/Draw/Lose 任一 → 已战(true)。三色都不命中 → 未战(false)。
// 注：Detector.MatchColor 只返 bool、无 error 通道，无法区分"未战中性态"与"比色异常"，
// 因此"异常当已战"暂不实现；待 MatchColor 提供 error 通道再补保守分支。
func battledAt(det screen.Detector, pt screen.Point, op OpponentFeature) bool {
	if det.MatchColor(pt.X, pt.Y, op.ResultColors.Win, op.ResultSim) {
		return true
	}
	if det.MatchColor(pt.X, pt.Y, op.ResultColors.Draw, op.ResultSim) {
		return true
	}
	if det.MatchColor(pt.X, pt.Y, op.ResultColors.Lose, op.ResultSim) {
		return true
	}
	return false
}
```

（`action` 已在 page.go import，无需新增。）

- [ ] **Step 4: 跑确认绿 + 全量**

Run:
- `go test ./internal/game/arena/ -run TestFindFirstValidOpponent -v` → Expected: 7 子用例全 PASS
- `go test ./...` → Expected: 全包 PASS
- `go vet ./...` → Expected: 无输出
- `gofmt -l internal/game/arena internal/platform/screen internal/config` → Expected: 无输出（无未格式化文件；如有，跑 `gofmt -w .` 再提交）

- [ ] **Step 5: Commit**

```bash
git add internal/game/arena/page.go internal/game/arena/page_test.go
git commit -m "feat(arena): implement FindFirstValidOpponent with trophy-range and battled-skip"
```

---

### Task 6: 用户按取色表填 feature.go（手动节点，无代码）

**Files:**
- Modify: `internal/game/arena/feature.go` 的 `DefaultFeature()`（由用户填写；本计划不代填）

**取色表（交付给用户，按 AutoGo 取色工具拷值）：**

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
| `Opponent.Anchor.Colors` | 锚点颜色串（单色不稳就用多点格式，字段不变） |
| `Opponent.Anchor.Sim` / `.Dir` | 相似度 / 方向（默认 0） |
| `Opponent.TrophyRect` | **相对锚点**的奖杯 OCR 偏移矩形 (dx1,dy1,dx2,dy2) |
| `Opponent.ResultOffset` + `ResultColors` + `ResultSim` | 相对锚点的战绩点 + 已战三色 + 相似度 |
| `Opponent.ClickOffset` | 相对锚点的点击偏移（锚点可点则 0,0） |

- [ ] **Step 1: 在 `DefaultFeature()` 内按上表填入所有字段数值**

- [ ] **Step 2: 编译确认**

Run: `go build ./...`
Expected: 无输出（成功）。`feature.go` 只是数值变化，不应影响编译。

- [ ] **Step 3: Commit（由用户决定时机）**

```bash
git add internal/game/arena/feature.go
git commit -m "feat(arena): fill arena feature values from color picker"
```

---

### Task 7: 真机核对（手动节点，spec §9 清单；识别不对只调 feature，不改 page.go）

**Files:** 无代码改动（仅在识别不符时调 `feature.go` 数值）。

- [ ] **Step 1: 上机逐项核对（用 AutoGo 插件 F7 跑，观察日志/返回值）**

- [ ] 进竞技场大厅：`IsLobby()==true`；离开大厅：`IsLobby()==false`。
- [ ] `ReadMedalAndTicket` 读到的勋章/票与界面一致。
- [ ] `ReadTrophyCount` 与界面自己奖杯一致。
- [ ] `ReadRefreshCountdown`：`"5分30秒"` 类文本解析为正确 `Duration`。
- [ ] `IsFreeRefresh` 在免费刷新可用时为 true。
- [ ] `FindMultiColorsAll(Anchor…)` 返回的锚点数量 == 当前页对手数；顺序按 `Dir=0` 从上到下。
- [ ] `FindFirstValidOpponent` 选中的卡：奖杯在区间、未战、点击位准确。

- [ ] **Step 2: 任一不符 → 只调 `feature.go` 数值（坐标/颜色/sim/dir），不改 `page.go` 逻辑**

微调提示：相对偏移误差通常 1–3 px；锚点单色不稳 → `Anchor.Colors` 改多点串；OCR 不稳 → 后续补 `NumberOCR` 参数。

- [ ] **Step 3: 通过后，本段（识别层）收尾；进入下一段（状态机闭环 / 战斗全流程）另行 spec → plan**

---

## 验收总览（本段完成的判定）

- `go build ./...` 与 `go build -tags android ./...` 均通过。
- `go test ./...` 全绿；`go vet ./...` 无输出；`gofmt -l` 无输出。
- `feature.go` 数值已填（Task 6），真机清单勾选完成（Task 7）。
- 代码侧无颜色/坐标硬编码（除测试 mock 与 feature 数值）。
