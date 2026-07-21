# 脚本 UI 框架 Phase 2（android 绘制层） Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `app/ui` 框架补齐 android 绘制层：主题应用、文本测量、组件渲染函数（Button/Checkbox/NumberInput/TextInput/Dropdown/Tabs/Collapsible/Image/Form/布局件）、灵动岛、描述符驱动面板、RunShell——使框架在设备上自举渲染完整 UI。

**Architecture:** 依据 ADR-0002/0003。android 文件（`//go:build android && cgo`）只做薄绘制：视觉算法从 `internal/ui` 既有代码移植（逐段给出源行号与变换规则），所有可测逻辑（Form 的 Store↔Props 桥接、Ctx 扩展、Shell 取值器）放无标签文件配单测。组件是每帧执行的函数，输入为 Phase 1 的 Props 结构体；样式全部读 `ctx.Theme`，禁止新的硬编码色值与包级可变状态。`internal/ui` 继续服役（Phase 3 才切换删除）。

**Tech Stack:** Go 1.25.0、`github.com/Dasongzi1366/AutoGo` 的 `imgui`/`device` 包（仅 android 文件可 import）。

## Global Constraints

- 新代码全部在仓库根 `ui/`。**不得改动 `internal/ui` 的任何文件**（它是移植源，只读）。
- android 文件命名 `*_android.go`，首行 `//go:build android && cgo`；只有这些文件可 import `github.com/Dasongzi1366/AutoGo/imgui` 与 `device`。
- 无标签文件不得 import imgui/device；`Ctx.states` 是 `map[string]any`，android 代码可在其中存放 imgui 类型。
- **禁止新增包级可变状态**（旧代码的 `countdownMap`、`panelMinimized`、`settingsStatus` 等全部实例化/状态化）；纹理缓存经 `Ctx.resource`。
- 样式令牌一律取 `ctx.theme()`；尺寸经 `ctx.S()`（基准 1600×900，宽度驱动）。
- 旧的中文组件名 `UI_创建*` 不移植命名，新 API 为英文 Props 驱动（ADR-0003）。
- 每个任务的本地验证（android 文件本地不可类型检查，靠评审+移植忠实性把关）：
  1. `gofmt -l ui/` 无输出；2. `go build ./...`；3. `go test ./...` 全绿；
  4. `GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`（走 stub 路径，验证标签切分正确）。
- 提交信息：`feat(ui): <中文摘要>`，结尾 `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`。
- 分支：`git checkout -b feat/script-ui-framework-phase2`（Task 1 的提交步执行）。

## 移植源行号索引（internal/ui，只读）

| 源 | 文件:行 |
|---|---|
| 文本测量 measureLabelSize/fitButtonSize/isCJKRune | `internal/ui/text_measure_android.go:14-78` |
| HexToVec4 | `internal/ui/widgets_android.go:58-80` |
| 标签栏 UI_创建标签栏 | `internal/ui/widgets_android.go:183-300` |
| EnableSlidingScroll | `internal/ui/widgets_android.go:302-330` |
| 液态玻璃按钮 | `internal/ui/widgets_android.go:335-426` |
| 复选框 | `internal/ui/widgets_android.go:500-620` |
| 输入框 | `internal/ui/widgets_android.go:629-728` |
| 数字输入框（步进） | `internal/ui/widgets_android.go:737-881` |
| 多行输入框 | `internal/ui/widgets_android.go:889-985` |
| 下拉框（自绘弹层） | `internal/ui/widgets_android.go:994-1250` |
| 图像+纹理缓存 | `internal/ui/widgets_android.go:1251-1366` |
| 折叠 | `internal/ui/widgets_android.go:1370-1475` |
| 灵动岛全部 | `internal/ui/island_android.go:13-445` |
| 面板窗壳/标题栏/脚本条 | `internal/ui/panel_android.go:16-250` |
| 列表勾选/胶囊/系统页 | `internal/ui/cookie_panel_android.go:142-168,243-271,343-385` |

---

### Task 1: Ctx 与 Shell 的绘制层配套扩展（无标签，TDD）

**Files:**
- Modify: `ui/context.go`（新增字段与方法）
- Modify: `ui/shell.go`（新增取值器）
- Test: `ui/context_test.go`、`ui/shell_test.go`（追加）

**Interfaces:**
- Consumes: 既有 `Ctx`、`Shell`、`Theme`、`DefaultTheme()`、`ScriptController.Exit()`
- Produces（后续全部任务依赖）:
  - `Ctx` 新增导出字段 `Theme Theme`、`Shell *Shell`
  - `(c *Ctx) theme() Theme`（零值回退 `DefaultTheme()`）
  - `(c *Ctx) resource(key string, create func() any) any`（写一次缓存，仅在 miss 时调用 create；android 任务用它存纹理）
  - `Shell` 新方法：`Title() string`、`ScriptState() ScriptState`、`ConfigPath() string`、`DataStorePath() string`、`BaseSize() (w, h int)`、`Exit()`（调控制器的 Exit）

- [ ] **Step 1: 写失败测试**

追加到 `ui/context_test.go`:

```go
func TestCtxThemeFallback(t *testing.T) {
	c := NewCtx(NewStore(), 1)
	if c.theme() != DefaultTheme() {
		t.Fatal("zero Theme should fall back to DefaultTheme")
	}
	custom := DefaultTheme()
	custom.Rounding = 99
	c.Theme = custom
	if c.theme().Rounding != 99 {
		t.Fatal("explicit Theme should win")
	}
}

func TestCtxResourceCreatedOnce(t *testing.T) {
	c := NewCtx(NewStore(), 1)
	calls := 0
	create := func() any { calls++; return calls }
	v1 := c.resource("tex:a", create)
	v2 := c.resource("tex:a", create)
	if v1 != v2 || calls != 1 {
		t.Fatalf("resource should create once: v1=%v v2=%v calls=%d", v1, v2, calls)
	}
	c.Push("form")
	if v := c.resource("tex:a", create); v != v1 {
		t.Fatal("resource is path-independent (global cache)")
	}
	c.Pop()
}
```

追加到 `ui/shell_test.go`:

```go
func TestShellGetters(t *testing.T) {
	s := NewShell(ShellOptions{
		Title: "测试脚本", ConfigPath: "/data/ui.json", DataStorePath: "/data/store.json",
		BaseWidth: 1280, BaseHeight: 720,
	})
	if s.Title() != "测试脚本" || s.ConfigPath() != "/data/ui.json" || s.DataStorePath() != "/data/store.json" {
		t.Fatal("getters")
	}
	if w, h := s.BaseSize(); w != 1280 || h != 720 {
		t.Fatalf("BaseSize=%d,%d", w, h)
	}
	if s.ScriptState() != StateIdle {
		t.Fatal("ScriptState should proxy controller")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run "TestCtxTheme|TestCtxResource|TestShellGetters" -v`
Expected: FAIL（undefined / no field）

- [ ] **Step 3: 写实现**

`ui/context.go`：在 `Ctx` 结构体加字段（紧随 `Scale` 后）、在文件末尾加两个方法：

```go
	// Theme 本帧主题（RunShell 在启动时注入）；零值时 theme() 回退 DefaultTheme。
	Theme Theme
	// Shell 持有本帧所属的 Shell 实例（RunShell 注入）；面板页面组件经它
	// 访问任务表、路径与控制器。手工构造的 Ctx 可留空。
	Shell *Shell
```

```go
// theme 返回生效主题：Ctx.Theme 为零值时回退默认主题。
func (c *Ctx) theme() Theme {
	if c.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return c.Theme
}

// resource 返回 key 对应的缓存资源，仅在缺失时调用 create 创建一次。
// 供 android 绘制层缓存纹理等后端资源；与组件路径无关（全局缓存）。
func (c *Ctx) resource(key string, create func() any) any {
	full := "res\x00" + key
	if v, ok := c.states[full]; ok {
		return v
	}
	v := create()
	c.states[full] = v
	return v
}
```

`ui/shell.go`：在取值器区块追加：

```go
func (s *Shell) Title() string         { return s.opts.Title }
func (s *Shell) ConfigPath() string    { return s.opts.ConfigPath }
func (s *Shell) DataStorePath() string { return s.opts.DataStorePath }

// BaseSize 返回基准分辨率（NewShell 已归一默认值 1600×900）。
func (s *Shell) BaseSize() (w, h int) { return s.opts.BaseWidth, s.opts.BaseHeight }

// ScriptState 代理脚本的当前生命周期状态。
func (s *Shell) ScriptState() ScriptState { return s.ctrl.State() }

// Exit 停止脚本并触发退出钩子（灵动岛「关闭」的语义）。
func (s *Shell) Exit() { s.ctrl.Exit() }
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -v`
Expected: PASS（全部）

- [ ] **Step 5: 提交**

```bash
git checkout -b feat/script-ui-framework-phase2
git add ui/context.go ui/context_test.go ui/shell.go ui/shell_test.go
git commit -m "feat(ui): Ctx 主题/资源缓存与 Shell 取值器（绘制层配套）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 2: 主题应用层（android）

**Files:**
- Create: `ui/theme_apply_android.go`

**Interfaces:**
- Consumes: `Theme`、`Color`（Phase 1）
- Produces: `toVec4(c Color) imgui.Vec4`、`ApplyTheme(t Theme)`（Task 12 RunShell 调用；组件任务用 `toVec4`）

- [ ] **Step 1: 写实现**（无本地可测逻辑，移植忠实性靠评审）

`ui/theme_apply_android.go`：

```go
//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// toVec4 把框架 Color 换算为 imgui.Vec4。
func toVec4(c Color) imgui.Vec4 {
	return imgui.Vec4{X: c.R, Y: c.G, Z: c.B, W: c.A}
}

// ApplyTheme 把主题令牌推入当前 ImGui 样式（在 imgui.Init 之后调用一次；
// 颜色/圆角在整个 RunShell 生命周期内保持）。
func ApplyTheme(t Theme) {
	imgui.PushStyleColorVec4(imgui.ColWindowBg, toVec4(t.WindowBg))
	imgui.PushStyleColorVec4(imgui.ColChildBg, toVec4(t.ChildBg))
	imgui.PushStyleColorVec4(imgui.ColPopupBg, toVec4(t.PopupBg))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(t.Border))
	imgui.PushStyleColorVec4(imgui.ColFrameBg, toVec4(t.FrameBg))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, toVec4(t.FrameHover))
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, toVec4(t.FrameActive))
	imgui.PushStyleColorVec4(imgui.ColButton, toVec4(t.Button))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, toVec4(t.ButtonHover))
	imgui.PushStyleColorVec4(imgui.ColButtonActive, toVec4(t.ButtonActive))
	imgui.PushStyleColorVec4(imgui.ColHeader, toVec4(t.Header))
	imgui.PushStyleColorVec4(imgui.ColHeaderHovered, toVec4(t.HeaderHover))
	imgui.PushStyleColorVec4(imgui.ColHeaderActive, toVec4(t.HeaderActive))
	imgui.PushStyleColorVec4(imgui.ColText, toVec4(t.Text))
	imgui.PushStyleColorVec4(imgui.ColTextDisabled, toVec4(t.TextDisabled))
	imgui.PushStyleColorVec4(imgui.ColTitleBg, toVec4(t.TitleBg))
	imgui.PushStyleColorVec4(imgui.ColTitleBgActive, toVec4(t.TitleBgActive))
	imgui.PushStyleColorVec4(imgui.ColCheckMark, toVec4(t.Accent))
	imgui.PushStyleColorVec4(imgui.ColSliderGrab, toVec4(t.Accent))
	imgui.PushStyleColorVec4(imgui.ColSliderGrabActive, toVec4(t.Accent))

	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarGrabRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarPopupRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarScrollbarRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 1)
	imgui.PushStyleVarFloat(imgui.StyleVarChildBorderSize, 1)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
}
```

（与 `internal/ui/theme_android.go:58-91` 的推栈清单一一对应，仅数据源从 QQBlueTheme 换成 Theme。）

- [ ] **Step 2: 验证**

Run: `gofmt -l ui/`（无输出）、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过（android 文件不参与本地编译，标签切分正确即可）

- [ ] **Step 3: 提交**

```bash
git add ui/theme_apply_android.go
git commit -m "feat(ui): 主题应用层（Theme→imgui 样式）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 3: 文本测量移植（android）

**Files:**
- Create: `ui/text_measure_android.go`

**Interfaces:**
- Produces: `measureLabelSize(label string) imgui.Vec2`、`fitButtonSize(label string, padX, padY float32) (w, h float32)`（全部组件任务依赖）、`isCJKRune(r rune) bool`

- [ ] **Step 1: 写实现**

逐字复制 `internal/ui/text_measure_android.go:1-78`（含 `//go:build android && cgo` 头与包名——同为 `package ui`，零改动）。这是 CJK 字宽保底的生产 workaround，注释全部保留。

- [ ] **Step 2: 验证**

Run: `gofmt -l ui/`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 3: 提交**

```bash
git add ui/text_measure_android.go
git commit -m "feat(ui): CJK 文本测量移植

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 4: Button 组件（android）

**Files:**
- Create: `ui/components_button_android.go`

**Interfaces:**
- Consumes: `Ctx.theme()`、`toVec4`、`measureLabelSize`/`fitButtonSize`、`ButtonProps`/`ButtonKind`（Phase 1）
- Produces: `Button(ctx *Ctx, p ButtonProps)`

- [ ] **Step 1: 写实现**

`ui/components_button_android.go`（液态玻璃视觉移植自 `internal/ui/widgets_android.go:335-426`，改为 Props 驱动 + 主题令牌 + 三种 Kind）:

```go
//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// Button 按钮组件（ADR-0003）。尺寸语义：Width/Height 为基准分辨率尺寸，
// 经 ctx.S 换算；<=0 表示按文字自适应。Kind 决定配色：
// Primary=主题 Accent 底白字；Secondary=液态玻璃（半透明白）；Danger=系统红底白字。
func Button(ctx *Ctx, p ButtonProps) {
	th := ctx.theme()
	const padX, padY = float32(16), float32(12)

	w, h := float32(p.Width), float32(p.Height)
	if w > 0 {
		w = ctx.S(p.Width)
	}
	if h > 0 {
		h = ctx.S(p.Height)
	}
	fitW, fitH := fitButtonSize(p.Label, padX, padY)
	if w <= 0 {
		w = fitW
		if w < 64 {
			w = 64
		}
	} else if w < fitW {
		w = fitW
	}
	if h <= 0 {
		h = fitH
	} else if h < fitH {
		h = fitH
	}

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)

	var bg, bgHover, bgActive, text imgui.Vec4
	switch p.Kind {
	case ButtonPrimary:
		bg, bgHover, bgActive = toVec4(th.Accent), toVec4(th.Accent), toVec4(th.TitleBottom)
		text = toVec4(th.White)
	case ButtonDanger:
		bg = imgui.Vec4{X: 0.85, Y: 0.16, Z: 0.22, W: 1}
		bgHover = imgui.Vec4{X: 0.95, Y: 0.29, Z: 0.33, W: 1}
		bgActive = imgui.Vec4{X: 0.75, Y: 0.10, Z: 0.16, W: 1}
		text = toVec4(th.White)
	default: // ButtonSecondary：液态玻璃
		bg = imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.10}
		bgHover = imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.20}
		bgActive = imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.35}
		text = toVec4(th.Text)
	}
	imgui.PushStyleColorVec4(imgui.ColButton, bg)
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, bgHover)
	imgui.PushStyleColorVec4(imgui.ColButtonActive, bgActive)
	imgui.PushStyleColorVec4(imgui.ColBorder, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.18})
	imgui.PushStyleColorVec4(imgui.ColText, text)

	if p.Disabled {
		imgui.BeginDisabled()
	}
	if imgui.ButtonV(p.Label+"##btn", imgui.Vec2{X: w, Y: h}) && p.OnClick != nil {
		p.OnClick()
	}
	if p.Disabled {
		imgui.EndDisabled()
	}

	imgui.PopStyleColorV(5)
	imgui.PopStyleVarV(4)
}
```

注意：`##btn` 裸 ID 在同窗口多按钮时会撞 ID——imgui 以 `Label##id` 整体判同，Label 不同即不撞；同 Label 多按钮的调用方应自行用 `ctx.Push` 隔离（组件状态同理）。评审时确认此注释写入 godoc。

- [ ] **Step 2: 验证**

Run: `gofmt -l ui/`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 3: 提交**

```bash
git add ui/components_button_android.go
git commit -m "feat(ui): Button 组件（三种变体，液态玻璃移植）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 5: 输入组件（Checkbox/NumberInput/TextInput，android）

**Files:**
- Modify: `ui/props.go`（`NumberInputProps` 增 `Hint string`；`InputProps` 增 `Multiline bool`）
- Create: `ui/components_input_android.go`

**Interfaces:**
- Consumes: `Ctx.theme()`、`toVec4`、`measureLabelSize`、`CheckboxProps`/`NumberInputProps`/`InputProps`
- Produces: `Checkbox(ctx *Ctx, p CheckboxProps)`、`NumberInput(ctx *Ctx, p NumberInputProps)`、`TextInput(ctx *Ctx, p InputProps)`（Form 任务依赖三者）

- [ ] **Step 1: props 增补（无标签）**

`ui/props.go`：`NumberInputProps` 加 `Hint string`（占位提示，如 "0=不限"）；`InputProps` 加 `Multiline bool`。跑 `go test ./ui/` 确认全绿。

- [ ] **Step 2: 写实现（移植 + 变换）**

创建 `ui/components_input_android.go`，`//go:build android && cgo`。三个函数分别移植：

1. `Checkbox(ctx, p)` — 源 `internal/ui/widgets_android.go:500-620`。变换：
   - 签名 `(store, key, showName, args...)` → `(ctx, p)`；勾选值读 `p.Checked`，切换时调 `p.OnChange(!p.Checked)`（不再写 store）。
   - 宽度语义只保留「自适应」（label 在左、勾选框紧跟其后，SameLine 布局）。
   - 硬编码色 → 主题令牌：白底 → `th.FrameBg`、描边 `#9cc3e5` → `th.Border`、勾 → `th.Accent`、文字 `#1f3a52` → `th.Text`（用 `toVec4`）。
2. `NumberInput(ctx, p)` — 源 `internal/ui/widgets_android.go:737-881`（[-] [输入] [+] 步进）。变换：
   - 值读 `p.Value`；步进/编辑后经 `p.Clamp(v)` 调 `p.OnChange(v)`。
   - `step/min/max` 取 `p.Step/p.Min/p.Max`（`p.Step<=0` 视为 1）；占位提示用 `p.Hint`；标签 `p.Label` 在左，控件占剩余宽度（`p.Width>0` 时固定宽并经 `ctx.S`）。
   - 输入框编辑中的草稿用组件状态：`buf := State(ctx, "numbuf", "")` 模式参照源文件的编辑态处理，改动提交时回调。
   - 颜色换主题令牌（FrameBg/Border/Accent/Text）。
3. `TextInput(ctx, p)` — 源 `internal/ui/widgets_android.go:629-728`（单行）与 `889-985`（多行）。变换：
   - `p.Multiline` 为 true 走多行路径，否则单行。
   - 受控缓冲：`buf := State(ctx, "inbuf", p.Value)`；imgui 返回编辑发生时 `p.OnChange(*buf)`（v1 不回写外部 p.Value 变更，godoc 注明）。
   - 标签 `p.Label`、提示 `p.Hint`、宽度 `p.Width`（>0 经 `ctx.S`，否则占满剩余）。
   - 颜色换主题令牌。

每个函数 godoc 注明 Props 字段语义。完成后 `gofmt -w ui/components_input_android.go`。

- [ ] **Step 3: 验证**

Run: `gofmt -l ui/`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 4: 提交**

```bash
git add ui/props.go ui/components_input_android.go
git commit -m "feat(ui): 输入组件（复选框/步进数字/单行与多行文本）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 6: Dropdown/Tabs/Collapsible（android）

**Files:**
- Create: `ui/components_select_android.go`

**Interfaces:**
- Consumes: `DropdownProps`/`TabsProps`/`CollapsibleProps`、`Ctx.theme()`、`toVec4`、`fitButtonSize`
- Produces: `Dropdown(ctx *Ctx, p DropdownProps)`、`Tabs(ctx *Ctx, p TabsProps)`、`Collapsible(ctx *Ctx, p CollapsibleProps, content func())`

- [ ] **Step 1: 写实现（移植 + 变换）**

创建 `ui/components_select_android.go`，`//go:build android && cgo`：

1. `Dropdown(ctx, p)` — 源 `internal/ui/widgets_android.go:994-1250`（自绘弹层）。变换：选中项 `p.Selected`、选项 `p.Options`，选择后 `p.OnChange(i)`；标签 `p.Label` 在左；弹层展开态用组件状态 `State(ctx, "open", false)`；颜色换主题令牌。
2. `Tabs(ctx, p)` — 源 `internal/ui/widgets_android.go:183-300`（顶部标签栏）。变换：不再接收 `pages []func()`——受控组件：`p.Items` 渲染为标签按钮行，激活项 `p.Selected` 用 Accent 底白字，点击 `p.OnChange(i)`；内容区由调用方按 Selected 自行 switch（godoc 给出示例）。颜色换主题令牌。
3. `Collapsible(ctx, p)` — 源 `internal/ui/widgets_android.go:1370-1475`。变换：展开态 `open := State(ctx, "open", p.Open)`；点击切换并调 `p.OnToggle(*open)`；`open` 为真时调用 `content()`。颜色换主题令牌。

- [ ] **Step 2: 验证**

Run: `gofmt -l ui/`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 3: 提交**

```bash
git add ui/components_select_android.go
git commit -m "feat(ui): 选择类组件（下拉/标签页/折叠）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 7: Image 与布局件（android）

**Files:**
- Create: `ui/components_media_android.go`

**Interfaces:**
- Consumes: `Ctx.resource`（Task 1）、`Ctx.Push/Pop/S`
- Produces:
  - `ImageProps struct{ Path string; Width, Height float64; OnClick func() }`（android 文件内定义，ADR-0003 允许的例外：依赖纹理）
  - `Image(ctx *Ctx, p ImageProps)`
  - `Row(ctx *Ctx, items ...func())`（首个直接绘制，其余 SameLine 后绘制）
  - `Column(ctx *Ctx, id string, content func())`（Push/Pop 作用域包装，提供组件状态命名空间）
  - `ScrollArea(ctx *Ctx, id string, height float64, content func())`（BeginChild/EndChild + 滑动滚动）
  - `EnableSlidingScroll(speed float32)`

- [ ] **Step 1: 写实现（移植 + 变换）**

创建 `ui/components_media_android.go`，`//go:build android && cgo`：

1. `ImageProps` + `Image` — 源 `internal/ui/widgets_android.go:1251-1366`。变换：包级纹理缓存 map → `ctx.resource("tex:"+p.Path, func() any { return imgui.CreateTextureNrgba(...) })`（加载逻辑照源文件）；`Width/Height>0` 经 `ctx.S`；`OnClick` 非空时用 ImageButton 语义（参照源文件的 callback 处理）。
2. `EnableSlidingScroll` — 逐字移植 `internal/ui/widgets_android.go:302-330`。
3. `Row` — 新代码：

```go
// Row 水平排布：首个元素原位绘制，其余 SameLine 衔接。
func Row(ctx *Ctx, items ...func()) {
	for i, item := range items {
		if i > 0 {
			imgui.SameLine()
		}
		if item != nil {
			item()
		}
	}
}
```

4. `Column` — 新代码：

```go
// Column 纵向作用域：为内容建立组件状态命名空间（Push/Pop），
// 布局本身是 imgui 默认纵向流。
func Column(ctx *Ctx, id string, content func()) {
	ctx.Push(id)
	defer ctx.Pop()
	if content != nil {
		content()
	}
}
```

5. `ScrollArea` — 新代码：

```go
// ScrollArea 固定高度滚动区（height 为基准分辨率尺寸，<=0 占满剩余），
// 内含触屏滑动滚动。
func ScrollArea(ctx *Ctx, id string, height float64, content func()) {
	h := float32(0)
	if height > 0 {
		h = float32(ctx.S(height))
	}
	ctx.Push(id)
	defer ctx.Pop()
	if imgui.BeginChildStrV(id, imgui.Vec2{X: 0, Y: h}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone) {
		EnableSlidingScroll(1)
		if content != nil {
			content()
		}
	}
	imgui.EndChild()
}
```

- [ ] **Step 2: 验证**

Run: `gofmt -l ui/`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 3: 提交**

```bash
git add ui/components_media_android.go
git commit -m "feat(ui): Image 组件与布局件（Row/Column/ScrollArea）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 8: Form 组件（无标签桥接 + android 渲染，TDD）

**Files:**
- Create: `ui/form.go`（无标签桥接）
- Test: `ui/form_test.go`
- Create: `ui/form_android.go`（渲染）

**Interfaces:**
- Consumes: `Field`、`Store`（Phase 1）、`Checkbox/NumberInput/TextInput`（Task 5）、`FormProps`（Phase 1）
- Produces:
  - `FormFieldValue(s *Store, f Field) any`（按 `f.Widget()` 返回 bool/float64/string）
  - `FormFieldChanged(s *Store, f Field, v any)`（按 widget 写回 store；number 收 float64）
  - `Form(ctx *Ctx, p FormProps)`（android：两列栅格自动渲染）

- [ ] **Step 1: 写失败测试**

`ui/form_test.go`:

```go
package ui

import "testing"

func TestFormFieldValueAndChanged(t *testing.T) {
	bf := Bool("b", "开关", func() bool { return false }, func(bool) {})
	nf := Number("n", "数量", 0, 99, 1, func() int { return 0 }, func(int) {})
	tf := Text("t", "文本", func() string { return "" }, func(string) {})

	s := NewStore()
	s.SetBool("b", true)
	s.SetFloat("n", 7)
	s.SetString("t", "hello")

	if v := FormFieldValue(s, bf); v != true {
		t.Fatalf("bool value=%v (%T)", v, v)
	}
	if v := FormFieldValue(s, nf); v != float64(7) {
		t.Fatalf("number value=%v (%T)", v, v)
	}
	if v := FormFieldValue(s, tf); v != "hello" {
		t.Fatalf("text value=%v (%T)", v, v)
	}

	FormFieldChanged(s, bf, false)
	FormFieldChanged(s, nf, 8.6)
	FormFieldChanged(s, tf, "world")
	if s.GetBool("b") || s.GetFloat("n") != 8.6 || s.GetString("t") != "world" {
		t.Fatalf("changed: b=%v n=%v t=%v", s.GetBool("b"), s.GetFloat("n"), s.GetString("t"))
	}

	// 类型不符的 v 安全忽略，nil store 不 panic
	FormFieldChanged(s, bf, "oops")
	if s.GetBool("b") != false {
		t.Fatal("mismatched type should be ignored")
	}
	FormFieldChanged(nil, bf, true)
	if v := FormFieldValue(nil, bf); v != false {
		t.Fatalf("nil store bool zero value, got %v", v)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run TestFormField -v`
Expected: FAIL（undefined: FormFieldValue / FormFieldChanged）

- [ ] **Step 3: 写实现**

`ui/form.go`:

```go
package ui

// FormFieldValue 按字段控件类型从 store 读当前值：
// checkbox→bool、number→float64、text→string。供 Form 渲染与测试使用。
func FormFieldValue(s *Store, f Field) any {
	if s == nil {
		switch f.Widget() {
		case WidgetCheckbox:
			return false
		case WidgetNumberInput:
			return float64(0)
		default:
			return ""
		}
	}
	switch f.Widget() {
	case WidgetCheckbox:
		return s.GetBool(f.Key())
	case WidgetNumberInput:
		return s.GetFloat(f.Key())
	default:
		return s.GetString(f.Key())
	}
}

// FormFieldChanged 把控件新值写回 store：checkbox 收 bool、number 收 float64、
// text 收 string；类型不符安全忽略，nil store 不 panic。
func FormFieldChanged(s *Store, f Field, v any) {
	if s == nil {
		return
	}
	switch f.Widget() {
	case WidgetCheckbox:
		if b, ok := v.(bool); ok {
			s.SetBool(f.Key(), b)
		}
	case WidgetNumberInput:
		if n, ok := v.(float64); ok {
			s.SetFloat(f.Key(), n)
		}
	default:
		if str, ok := v.(string); ok {
			s.SetString(f.Key(), str)
		}
	}
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -run TestFormField -v`
Expected: PASS

- [ ] **Step 5: 写 android 渲染**

`ui/form_android.go`（两列栅格移植自 `internal/ui/cookie_panel_android.go:273-295` 的 arena 表单布局，泛化为描述符驱动）:

```go
//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// Form 表单组件（ADR-0003）：按 Fields 自动排版为两列栅格（标签列固定宽 =
// 最长标签 + 12 间距，控件列占剩余宽度），值直连 Store 读写。
// 任务详情页默认渲染器；自定义 section 也可在 RenderDetail 内复用。
func Form(ctx *Ctx, p FormProps) {
	if p.Store == nil {
		return
	}
	const gapM = float32(12)
	const gapS = float32(8)

	labelW := float32(0)
	for _, f := range p.Fields {
		if w := measureLabelSize(f.Label()).X; w > labelW {
			labelW = w
		}
	}
	controlX := imgui.CursorPosX() + labelW + gapM

	for i, f := range p.Fields {
		ctx.Push(f.Key())
		imgui.AlignTextToFramePadding()
		imgui.Text(f.Label())
		imgui.SameLineV(0, 0)
		imgui.SetCursorPosX(controlX)

		switch f.Widget() {
		case WidgetCheckbox:
			Checkbox(ctx, CheckboxProps{
				Checked: FormFieldValue(p.Store, f).(bool),
				OnChange: func(v bool) {
					FormFieldChanged(p.Store, f, v)
				},
			})
		case WidgetNumberInput:
			c := f.Constraints()
			NumberInput(ctx, NumberInputProps{
				Value: FormFieldValue(p.Store, f).(float64),
				Min:   c.Min, Max: c.Max, Step: c.Step,
				OnChange: func(v float64) {
					FormFieldChanged(p.Store, f, v)
				},
			})
		default:
			TextInput(ctx, InputProps{
				Value: FormFieldValue(p.Store, f).(string),
				OnChange: func(v string) {
					FormFieldChanged(p.Store, f, v)
				},
			})
		}
		ctx.Pop()
		if i < len(p.Fields)-1 {
			imgui.Dummy(imgui.Vec2{X: 0, Y: gapS})
		}
	}
}
```

- [ ] **Step 6: 验证**

Run: `gofmt -l ui/`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 7: 提交**

```bash
git add ui/form.go ui/form_test.go ui/form_android.go
git commit -m "feat(ui): Form 组件（描述符自动渲染 + Store 桥接）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 9: 灵动岛（android，移植 + Shell 重接）

**Files:**
- Create: `ui/island_android.go`

**Interfaces:**
- Consumes: `Shell`（`ScriptState/StatusText/StartStop/PauseResume/OpenPanel/Exit/Title`）、`measureLabelSize`
- Produces: `floatingIsland`（未导出）、`newFloatingIsland()`、`(isl *floatingIsland) Draw(ctx *Ctx, shell *Shell)`（Task 12 RunShell 依赖）

- [ ] **Step 1: 写实现（移植 + 变换）**

逐段移植 `internal/ui/island_android.go:13-445` 到 `ui/island_android.go`，变换如下：

1. 类型 `FloatingIsland` → 未导出 `floatingIsland`；`NewFloatingIsland` → `newFloatingIsland`；`IslandCallbacks` 类型删除。
2. `Draw(cb, state, taskStatus)` → `Draw(ctx *Ctx, shell *Shell)`：
   - `state` = `shell.ScriptState()`；label = `shell.StatusText()`，空则 `shell.Title() + " · " + islandStateLabel(state)`（替代旧 `islandStateText` 的硬编码「帅宾 Cookie」）。
   - 命中回调直连 Shell：startStop → `shell.StartStop()`（忽略 error）；pauseResume → `shell.PauseResume()`；settings → `shell.OpenPanel()`；close → `shell.Exit()`。
3. 包级颜色 var（`islandBg` 等）保留为包级**不可变**声明（imgui.Vec4 字面量，无写入——允许；灵动岛是深色浮层，不走面板主题）。
4. 其余（layout/动画/图标绘制/自动收起/`easeOutCubic`）逐字移植，包括基于 `device.GetDisplayInfo` 的屏宽缓存与 clamp [0.8,1.6] 的岛局部缩放（岛的特殊策略，godoc 注明不经 `ctx.S`）。
5. 所有图标/图标编号常量保持原名但首字母小写化以适配未导出类型（`islandIconStartStop` 等已是小写，沿用）。

- [ ] **Step 2: 验证**

Run: `gofmt -l ui/`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 3: 提交**

```bash
git add ui/island_android.go
git commit -m "feat(ui): 灵动岛移植（回调直连 Shell）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 10: 面板窗壳与脚本条（android）

**Files:**
- Create: `ui/panel_android.go`

**Interfaces:**
- Consumes: `Shell`（含 Task 1 取值器）、`Ctx`、`toVec4`、`fitButtonSize`、`measureLabelSize`、`Button`（Task 4）、`CountEnabled`
- Produces: `drawPanel(ctx *Ctx, shell *Shell)`（Task 12 RunShell 依赖；内部含标题栏、脚本条、导航分派）

- [ ] **Step 1: 写实现（移植 + 变换）**

创建 `ui/panel_android.go`，`//go:build android && cgo`。移植 `internal/ui/panel_android.go:16-250`，变换：

1. `drawConfigPanel` → `drawPanel(ctx, shell)`：
   - 窗口尺寸/位置算法（0.72 宽 cap 980/640、0.78 高 cap 700/480、`device.GetDisplayInfo`）逐字保留。
   - `panelMinimized` 包级变量 → `shell.Minimized()`/`shell.ToggleMinimized()`。
   - 窗口 `open` 标志：本地 `open := true`，`imgui.BeginV(title, &open, flags)` 后若 `!open` 调 `shell.ClosePanel()`。
   - 标题栏渐变/窗控按钮逐字移植（`internal/ui/panel_android.go:69-141`），颜色 `QQBlueTitleTop/Bottom/White` → `th.TitleTop/TitleBottom/White`；缩小钮 → `shell.ToggleMinimized()`；关闭钮 → `shell.ClosePanel()`（关闭时若 minimized 同时复位：先 `if shell.Minimized() { shell.ToggleMinimized() }`）。
   - 标题 `shell.Title()`。
2. `renderScriptBar(opts, open)` → `renderScriptBar(ctx, shell)`：
   - 状态文案/颜色（`internal/ui/panel_android.go:228-250`）逐字移植。
   - 启用计数 `CountEnabled(shell.Store(), shell.Tasks())`。
   - 暂停/继续钮 → `shell.PauseResume()`；主钮（开始/停止）→ `if err := shell.StartStop(); err == nil && wasIdle { shell.ClosePanel() }`（wasIdle = 点击前 `shell.ScriptState() == StateIdle`；复刻旧面板「开始后关面板」语义，而 `StartStop` 的自动暂停会被 `ClosePanel` 配对恢复）。
   - 主按钮样式 Accent 底白字用 `th.Accent/th.White`。
3. 内容区（`renderCookiePanel` 的导航分派逻辑泛化）：

```go
// renderPanelContent 左轨导航 + 当前条目内容。导航条目来自 shell.Nav()；
// 任务列表页/系统页是应用挂载的普通条目（框架提供 TaskListPage/SystemPage）。
func renderPanelContent(ctx *Ctx, shell *Shell) {
	store := shell.Store()
	nav := shell.Nav()
	if len(nav) == 0 {
		return
	}
	th := ctx.theme()

	avail := imgui.ContentRegionAvail()
	railW := float32(ctx.S(120))

	imgui.PushStyleColorVec4(imgui.ColChildBg, toVec4(th.RailBg))
	imgui.BeginChildStrV("panel_rail", imgui.Vec2{X: railW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	current := store.GetString(KeyPanelNav)
	for _, entry := range nav {
		railButton(ctx, store, entry.ID, entry.Title, current == entry.ID)
	}
	imgui.EndChild()
	imgui.PopStyleColor()

	imgui.SameLine()
	imgui.BeginChildStrV("panel_body", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	for _, entry := range nav {
		if entry.ID == current && entry.Render != nil {
			ctx.Push("nav:" + entry.ID)
			entry.Render(ctx)
			ctx.Pop()
		}
	}
	imgui.EndChild()
}
```

   `railButton` 移植 `internal/ui/cookie_panel_android.go:72-93`：选中态 `th.Accent` 底 + `th.White` 字，点击 `store.SetString(KeyPanelNav, id)`。

- [ ] **Step 2: 验证**

Run: `gofmt -l ui/`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 3: 提交**

```bash
git add ui/panel_android.go
git commit -m "feat(ui): 面板窗壳/标题栏/脚本条与描述符导航

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 11: 任务列表页与系统页（android）

**Files:**
- Create: `ui/taskpage_android.go`
- Create: `ui/system_android.go`

**Interfaces:**
- Consumes: `Task`/`Categories`/`FilterByCategory`/`FindTask`（Phase 1）、`Form`（Task 8）、`measureLabelSize`、`ctx.theme()`、`ClearPanelCache`
- Produces:
  - `TaskListPage() func(*Ctx)`（应用挂载为 NavEntry.Render；读 `ctx.Shell` 的任务表）
  - `SystemPage() func(*Ctx)`（保存配置/清缓存标准页）

- [ ] **Step 1: 写实现**

`ui/taskpage_android.go`，`//go:build android && cgo`：

```go
// TaskListPage 框架可复用的任务列表页（ADR-0002）：分类 chips + 任务列表
// + 详情（RenderDetail 逃生门或 Form 自动渲染）。应用把它作为一个导航
// 条目挂载：ui.NavEntry{ID: "tasks", Title: "任务", Render: ui.TaskListPage()}。
func TaskListPage() func(*Ctx) {
	return func(ctx *Ctx) {
		shell := ctx.Shell
		if shell == nil {
			return
		}
		store := shell.Store()
		tasks := shell.Tasks()
		th := ctx.theme()

		avail := imgui.ContentRegionAvail()
		listW := float32(ctx.S(260))
		if avail.X < float32(ctx.S(520)) {
			listW = avail.X * 0.42
			if listW < float32(ctx.S(160)) {
				listW = float32(ctx.S(160))
			}
		}

		imgui.BeginChildStrV("panel_list", imgui.Vec2{X: listW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderCatChips(ctx, store, tasks)
		imgui.Separator()
		renderTaskRows(ctx, store, tasks)
		imgui.EndChild()

		imgui.SameLine()
		imgui.BeginChildStrV("panel_detail", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderTaskDetail(ctx, store, tasks, th)
		imgui.EndChild()
	}
}
```

配套移植（同文件，未导出）：

1. `renderCatChips(ctx, store, tasks)` — 源 `internal/ui/cookie_panel_android.go:170-207`。chips = `全部`（id `PanelCatAll`）+ `Categories(tasks)` 动态推导；当前分类读 `store.GetString(KeyPanelCat)`；选中样式 Accent/White。
2. `renderTaskRows(ctx, store, tasks)` — 源 `internal/ui/cookie_panel_android.go:95-140`：`FilterByCategory` 过滤；行 = 手绘启用勾选（`drawListCheck` 源 `:142-168`，颜色换 `th.White/th.Border/th.Accent`）+ 标题 + `task.Summary(store)` 摘要（`th.TextDisabled`，Summary 为 nil 跳过）；点击写 `KeyPanelSelected`；空列表提示「（该分类暂无任务）」。
3. `renderTaskDetail(ctx, store, tasks, th)` — 源 `internal/ui/cookie_panel_android.go:209-241`：`FindTask`（找不到回退首个并回写选中）；头部标题 + 启用胶囊 + 分类胶囊（`drawEnabledPill`/`drawPill` 源 `:243-271`）；然后：
   - `task.RenderDetail != nil` → `ctx.Push("detail:"+task.ID)`、调用、`Pop`
   - 否则 → `Form(ctx, FormProps{Store: store, Fields: task.Fields})`

`ui/system_android.go`，`//go:build android && cgo`：

```go
// SystemPage 标准系统页：配置持久化（保存/清缓存）。应用挂载为导航条目。
// 反馈文案经组件状态持有（替代旧包级 settingsStatus）。
func SystemPage() func(*Ctx) {
	return func(ctx *Ctx) {
		shell := ctx.Shell
		if shell == nil {
			return
		}
		store := shell.Store()
		status := State(ctx, "sysStatus", "")

		imgui.Text("系统")
		imgui.Dummy(imgui.Vec2{X: 0, Y: 4})
		imgui.TextDisabled("配置持久化")
		imgui.Separator()
		imgui.Dummy(imgui.Vec2{X: 0, Y: 4})
		imgui.TextDisabled("配置文件  " + shell.ConfigPath())
		imgui.Dummy(imgui.Vec2{X: 0, Y: 8})

		Row(ctx,
			func() {
				Button(ctx, ButtonProps{Label: "保存配置", Kind: ButtonSecondary, OnClick: func() {
					if err := store.SaveConfig(shell.ConfigPath()); err != nil {
						*status = fmt.Sprintf("保存失败：%v", err)
						return
					}
					*status = "配置已保存"
				}})
			},
			func() {
				Button(ctx, ButtonProps{Label: "清除缓存", Kind: ButtonSecondary, OnClick: func() {
					if shell.ScriptState() != StateIdle {
						_ = shell.StartStop() // 停止脚本
					}
					if err := ClearPanelCache(store, shell.ConfigPath(), shell.DataStorePath(), func(*Store) { shell.Seed() }); err != nil {
						*status = fmt.Sprintf("清除失败：%v", err)
						return
					}
					*status = "缓存已清除，默认配置已恢复"
				}})
			},
		)

		if *status != "" {
			imgui.Spacing()
			imgui.TextWrapped(*status)
		}
	}
}
```

- [ ] **Step 2: 验证**

Run: `gofmt -l ui/`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 3: 提交**

```bash
git add ui/taskpage_android.go ui/system_android.go
git commit -m "feat(ui): 任务列表页与系统页组件

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 12: RunShell 与桌面 stub + 全量验证

**Files:**
- Create: `ui/shell_run_android.go`
- Create: `ui/run_stub.go`（`//go:build !android || !cgo`）
- Modify: `CLAUDE.md`、`AGENTS.md`（ui/ 行描述更新）

**Interfaces:**
- Consumes: 全部前述任务
- Produces: `RunShell(opts ShellOptions)`（android：完整壳；stub：仅 Seed + 启动控制器，与旧 stub 行为一致）

- [ ] **Step 1: 写实现**

`ui/shell_run_android.go`：

```go
//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// RunShell 启动框架 UI 主循环（阻塞）：灵动岛始终绘制，配置面板按
// Shell 状态开关。首帧加载持久化配置并 Seed。
func RunShell(opts ShellOptions) {
	shell := NewShell(opts)

	_ = imgui.Init()
	ApplyTheme(shell.Theme())

	w, h, _, _ := device.GetDisplayInfo(0)
	bw, bh := shell.BaseSize()
	ctx := NewCtx(shell.Store(), ComputeScale(w, h, bw, bh))
	ctx.Theme = shell.Theme()
	ctx.Shell = shell

	island := newFloatingIsland()
	loaded := false

	imgui.Run(func() {
		if !loaded {
			loaded = true
			if shell.ConfigPath() != "" {
				_ = shell.Store().LoadConfig(shell.ConfigPath())
			}
			shell.Seed()
		}

		island.Draw(ctx, shell)
		if shell.PanelOpen() {
			drawPanel(ctx, shell)
		}
	})
	select {}
}
```

`ui/run_stub.go`：

```go
//go:build !android || !cgo

package ui

// RunShell 桌面/非 cgo stub：无 UI，Seed 后直接启动脚本（与旧 internal/ui
// stub 行为一致），保证 go build/test 在任何平台可用。
func RunShell(opts ShellOptions) {
	shell := NewShell(opts)
	shell.Seed()
	if opts.Controller != nil {
		opts.Controller.Start()
	}
}
```

- [ ] **Step 2: 验证（本计划的收口门禁）**

Run: `gofmt -l ui/`（无输出）、`go vet ./ui/...`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 3: 文档同步**

`CLAUDE.md` 与 `AGENTS.md` 中 ui/ 行改为：

```
├── ui/                        # 脚本 UI 框架（ADR-0002/0003）：纯逻辑无标签可测 + android 绘制层（组件/灵动岛/面板/RunShell）
```

- [ ] **Step 4: 提交**

```bash
git add ui/shell_run_android.go ui/run_stub.go CLAUDE.md AGENTS.md
git commit -m "feat(ui): RunShell 与桌面 stub（Phase 2 完成）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Self-Review 记录

- **Spec 覆盖**：ADR-0003 组件清单（Button/Input/NumberInput/MultilineInput/Checkbox/Dropdown/Form/Tabs/Collapsible/Image/Text+布局件）→ Task 4-8 全覆盖（Text = imgui.Text 直用 + Form/页面内封装，不单设函数）；TabItem 不单独提供（Tabs 受控 + 调用方 switch 内容，较 ADR 措辞简化，记录在案）；倒计时按钮不移植（无消费者，YAGNI）。ADR-0002 的岛入框架、任务列表页可复用组件、系统页应用挂载、主题应用、缩放 → Task 9-12。
- **已知行为差（有意）**：① TextInput/NumberInput v1 不回写外部 Value 变更（godoc 注明）；② 灵动岛保留局部 clamp 缩放，不经 ctx.S；③ 面板「开始」保持旧语义（开始后关面板），与岛上启动（面板开着则自动暂停）不同路径、均有测试/旧代码依据。
- **本地不可验声明**：Task 2-7、9-11 的 android 文件无法本地类型检查（仅插件 NDK 工具链可编译）；质量门 = 移植忠实性评审 + gofmt + stub 构建 + 无标签测试。Task 12 后建议在设备上经插件构建验证一次。
- **类型一致性**：`resource`/`theme()`（T1）↔ Task 2-7 消费；`FormFieldValue/FormFieldChanged`（T8）↔ Form android 渲染一致；`TaskListPage()/SystemPage()` 返回 `func(*Ctx)` 与 `NavEntry.Render` 签名一致；`drawPanel/floatingIsland.Draw`（T9/T10）↔ RunShell（T12）一致。
