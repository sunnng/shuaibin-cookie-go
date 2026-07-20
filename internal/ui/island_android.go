//go:build android && cgo

package ui

import (
	"math"
	"time"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// IslandCallbacks 由调用方（shell）提供，灵动岛按钮点击时回调。
// 任一字段为 nil 时点击该按钮无副作用。
type IslandCallbacks struct {
	OnSettings    func()
	OnStartStop   func()
	OnPauseResume func()
	OnClose       func()
}

// FloatingIsland 苹果刘海（灵动岛）样式悬浮窗：顶部居中的黑色胶囊，
// 点击展开为圆角卡片，提供 开始/停止、暂停/继续、设置、关闭 四个操作。
// 固定在顶部居中，不可拖动（与真实刘海一致）；展开后再点卡片空白处或卡片外任意处收起，
// 展开超过 islandAutoCollapse 无操作也会自动收起，避免长时间遮挡游戏画面干扰识别。
// 业务运行状态由调用方通过 Draw 的 state 参数传入，岛内部只保留 UI 状态。
type FloatingIsland struct {
	ScreenWidth  int
	ScreenHeight int

	IsExpanded bool
	ExpandAnim float32

	expandedAt time.Time
}

// NewFloatingIsland 创建灵动岛：初始为收起的胶囊。
func NewFloatingIsland() *FloatingIsland {
	return &FloatingIsland{}
}

// 尺寸基准：1600 宽（与平台层 1600×900 基准坐标一致），按屏幕宽等比缩放并夹取。
const (
	islandRefWidth   = float32(1600)
	islandTopMargin  = float32(14)
	islandPillHeight = float32(60)
	islandCardWidth  = float32(560)
	islandCardHeight = float32(190)
	islandCardRadius = float32(34)
	islandBtnRadius  = float32(34)
	islandBtnSpacing = float32(120)
)

// islandAutoCollapse 展开卡片的无操作自动收起时长。
const islandAutoCollapse = 6 * time.Second

// 展开卡片按钮图标编号。
const (
	islandIconStartStop = iota
	islandIconPauseResume
	islandIconSettings
	islandIconClose
)

var (
	islandBg     = imgui.Vec4{X: 0.02, Y: 0.02, Z: 0.03, W: 0.96}
	islandBorder = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.10}
	islandText   = imgui.Vec4{X: 0.92, Y: 0.93, Z: 0.96, W: 1.0}

	// 苹果风格按钮：iOS 控制中心式半透明灰圆底 + 单色图标；
	// 彩色只留给破坏性操作（系统红）与状态点。
	islandBtnBg  = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.12}
	islandGlyph  = imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.95}
	islandDanger = imgui.Vec4{X: 1.0, Y: 0.27, Z: 0.23, W: 1.0}
)

// 状态点颜色（依据脚本状态）。
func islandStateColor(state ScriptState) imgui.Vec4 {
	switch state {
	case StateRunning:
		return imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.3, W: 1.0}
	case StatePaused:
		return imgui.Vec4{X: 1.0, Y: 0.85, Z: 0.2, W: 1.0}
	default:
		return imgui.Vec4{X: 0.2, Y: 0.6, Z: 1.0, W: 1.0}
	}
}

func islandStateLabel(state ScriptState) string {
	switch state {
	case StateRunning:
		return "运行中"
	case StatePaused:
		return "已暂停"
	default:
		return "空闲"
	}
}

// Draw 每帧调用：推进动画并绘制灵动岛。
// taskStatus 为任务侧上报的状态文本（如"竞技场 3/10 · 胜率 67%"），非空时替代默认状态文案。
func (isl *FloatingIsland) Draw(cb IslandCallbacks, state ScriptState, taskStatus string) {
	if isl.ScreenWidth == 0 {
		w, h, _, _ := device.GetDisplayInfo(0)
		if w == 0 {
			w, h = 1600, 900
		}
		isl.ScreenWidth = w
		isl.ScreenHeight = h
	}

	// 展开超时自动收起：卡片遮挡面积较大，长时间盖在游戏上会干扰截图识别。
	if isl.IsExpanded && !isl.expandedAt.IsZero() && time.Since(isl.expandedAt) > islandAutoCollapse {
		isl.IsExpanded = false
	}

	isl.updateAnimation()
	isl.drawWindow(cb, state, islandLabel(state, taskStatus))
}

// islandLabel 有任务状态时优先展示任务状态，否则回退到默认状态文案。
func islandLabel(state ScriptState, taskStatus string) string {
	if taskStatus != "" {
		return taskStatus
	}
	return islandStateText(state)
}

func (isl *FloatingIsland) scale() float32 {
	s := float32(isl.ScreenWidth) / islandRefWidth
	if s < 0.8 {
		s = 0.8
	}
	if s > 1.6 {
		s = 1.6
	}
	return s
}

type islandButton struct {
	pos    imgui.Vec2
	radius float32
	icon   int
}

type islandLayout struct {
	x, y, w, h, radius float32
	scale              float32
	anim               float32
	buttons            []islandButton
}

// layout 计算本帧胶囊/卡片矩形（展开动画对宽高与圆角做插值）与按钮排布。
func (isl *FloatingIsland) layout(label string) islandLayout {
	s := isl.scale()
	anim := easeOutCubic(isl.ExpandAnim)

	pillW := isl.pillWidth(label, s)
	pillH := islandPillHeight * s
	cardW := islandCardWidth * s
	cardH := islandCardHeight * s

	l := islandLayout{
		w:      pillW + (cardW-pillW)*anim,
		h:      pillH + (cardH-pillH)*anim,
		radius: pillH/2 + (islandCardRadius*s-pillH/2)*anim,
		scale:  s,
		anim:   anim,
	}
	l.x = (float32(isl.ScreenWidth) - l.w) / 2
	l.y = islandTopMargin * s

	centerX := l.x + l.w/2
	btnY := l.y + cardH - 62*s
	for i := 0; i < 4; i++ {
		l.buttons = append(l.buttons, islandButton{
			pos:    imgui.Vec2{X: centerX + (float32(i)-1.5)*islandBtnSpacing*s, Y: btnY},
			radius: islandBtnRadius * s,
			icon:   i,
		})
	}
	return l
}

// pillWidth 胶囊宽度随状态文字自适应（灵动岛风格：内容多宽胶囊就多宽）。
func (isl *FloatingIsland) pillWidth(label string, s float32) float32 {
	textW := measureLabelSize(label).X
	dotD := float32(18) * s
	gap := float32(10) * s
	pad := float32(24) * s
	return pad*2 + dotD + gap + textW
}

func islandStateText(state ScriptState) string {
	return "帅宾 Cookie · " + islandStateLabel(state)
}

func (isl *FloatingIsland) drawWindow(cb IslandCallbacks, state ScriptState, label string) {
	l := isl.layout(label)

	imgui.SetNextWindowPosV(imgui.Vec2{X: l.x, Y: l.y}, imgui.CondAlways, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: l.w, Y: l.h}, imgui.CondAlways)

	imgui.PushStyleColorVec4(imgui.ColWindowBg, imgui.Vec4{})
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{})

	flags := imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoResize |
		imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoBackground |
		imgui.WindowFlagsNoSavedSettings | imgui.WindowFlagsNoMove

	imgui.BeginV("##FloatIsland", nil, flags)

	drawList := imgui.WindowDrawList()

	isl.handleInput(l, cb)

	pMin := imgui.Vec2{X: l.x, Y: l.y}
	pMax := imgui.Vec2{X: l.x + l.w, Y: l.y + l.h}
	drawList.AddRectFilledV(pMin, pMax, imgui.ColorU32Vec4(islandBg), l.radius, imgui.DrawFlagsRoundCornersAll)
	drawList.AddRectV(pMin, pMax, imgui.ColorU32Vec4(islandBorder), l.radius, imgui.DrawFlagsRoundCornersAll, 1)

	if l.anim < 0.85 {
		isl.drawPillContent(drawList, l, state, label)
	} else {
		isl.drawCardContent(drawList, l, state, label)
	}

	imgui.End()

	imgui.PopStyleVar()
	imgui.PopStyleVar()
	imgui.PopStyleColor()
}

// drawPillContent 收起态：状态点 + 状态文案（默认"帅宾 Cookie · 状态"，可被任务状态替代），整体水平居中。
func (isl *FloatingIsland) drawPillContent(drawList *imgui.DrawList, l islandLayout, state ScriptState, label string) {
	textSz := measureLabelSize(label)
	dotD := float32(18) * l.scale
	gap := float32(10) * l.scale
	groupW := dotD + gap + textSz.X

	cy := l.y + l.h/2
	dotX := l.x + (l.w-groupW)/2 + dotD/2
	drawList.AddCircleFilled(imgui.Vec2{X: dotX, Y: cy}, dotD/2, imgui.ColorU32Vec4(islandStateColor(state)))
	drawList.AddTextVec2V(
		imgui.Vec2{X: dotX + dotD/2 + gap, Y: cy - textSz.Y/2},
		imgui.ColorU32Vec4(islandText),
		label,
	)
}

// drawCardContent 展开态：顶部状态行 + 一排四个苹果风格图标按钮。
func (isl *FloatingIsland) drawCardContent(drawList *imgui.DrawList, l islandLayout, state ScriptState, label string) {
	textSz := measureLabelSize(label)
	dotD := float32(14) * l.scale
	gap := float32(8) * l.scale
	groupW := dotD + gap + textSz.X

	rowY := l.y + 44*l.scale
	dotX := l.x + (l.w-groupW)/2 + dotD/2
	drawList.AddCircleFilled(imgui.Vec2{X: dotX, Y: rowY}, dotD/2, imgui.ColorU32Vec4(islandStateColor(state)))
	drawList.AddTextVec2V(
		imgui.Vec2{X: dotX + dotD/2 + gap, Y: rowY - textSz.Y/2},
		imgui.ColorU32Vec4(islandText),
		label,
	)

	for _, b := range l.buttons {
		drawList.AddCircleFilled(b.pos, b.radius, imgui.ColorU32Vec4(islandBtnBg))
		drawIslandIcon(drawList, b.pos, b.radius, b.icon, state)
	}
}

// handleInput 命中检测：收起点胶囊展开；展开点按钮执行并收起；再点卡片空白处或卡片外收起。
func (isl *FloatingIsland) handleInput(l islandLayout, cb IslandCallbacks) {
	if !imgui.IsMouseReleased(imgui.MouseButtonLeft) {
		return
	}
	// 动画进行中不响应，避免展开/收起过程中的误触。
	if isl.ExpandAnim > 0.05 && isl.ExpandAnim < 0.95 {
		return
	}

	m := imgui.MousePos()
	if !isl.IsExpanded {
		if pointInLayout(m, l) {
			isl.IsExpanded = true
			isl.expandedAt = time.Now()
		}
		return
	}

	for _, b := range l.buttons {
		dx := m.X - b.pos.X
		dy := m.Y - b.pos.Y
		if float32(math.Sqrt(float64(dx*dx+dy*dy))) >= b.radius {
			continue
		}
		isl.IsExpanded = false
		switch b.icon {
		case islandIconStartStop:
			if cb.OnStartStop != nil {
				cb.OnStartStop()
			}
		case islandIconPauseResume:
			if cb.OnPauseResume != nil {
				cb.OnPauseResume()
			}
		case islandIconSettings:
			if cb.OnSettings != nil {
				cb.OnSettings()
			}
		case islandIconClose:
			if cb.OnClose != nil {
				cb.OnClose()
			}
		}
		return
	}

	// 展开后再点灵动岛（卡片空白处）或卡片外任意处：收起。
	isl.IsExpanded = false
}

func pointInLayout(m imgui.Vec2, l islandLayout) bool {
	return m.X >= l.x && m.X <= l.x+l.w && m.Y >= l.y && m.Y <= l.y+l.h
}

func (isl *FloatingIsland) updateAnimation() {
	if isl.IsExpanded {
		if isl.ExpandAnim < 1.0 {
			isl.ExpandAnim += 0.08
			if isl.ExpandAnim > 1.0 {
				isl.ExpandAnim = 1.0
			}
		}
	} else {
		if isl.ExpandAnim > 0.0 {
			isl.ExpandAnim -= 0.08
			if isl.ExpandAnim < 0.0 {
				isl.ExpandAnim = 0.0
			}
		}
	}
}

// drawIslandIcon 苹果 SF Symbols 风格图标：单色细线条（白色，关闭为系统红），
// 笔画粗细随按钮半径等比；按钮底为统一的半透明灰圆，不用彩色填充底。
func drawIslandIcon(drawList *imgui.DrawList, pos imgui.Vec2, radius float32, icon int, state ScriptState) {
	thickness := radius * 0.07
	if thickness < 1.5 {
		thickness = 1.5
	}
	size := radius * 0.42 // 图标半尺寸

	glyph := islandGlyph
	if icon == islandIconClose {
		glyph = islandDanger
	}
	col := imgui.ColorU32Vec4(glyph)

	switch icon {
	case islandIconStartStop:
		if state == StateIdle {
			drawPlayIcon(drawList, pos, size, col)
		} else {
			// stop：圆角方块
			half := size * 0.85
			drawList.AddRectFilledV(
				imgui.Vec2{X: pos.X - half, Y: pos.Y - half},
				imgui.Vec2{X: pos.X + half, Y: pos.Y + half},
				col, half*0.55, imgui.DrawFlagsRoundCornersAll,
			)
		}

	case islandIconPauseResume:
		if state == StatePaused {
			// 苹果媒体控制惯例：暂停态显示播放键（点击即继续）
			drawPlayIcon(drawList, pos, size, col)
		} else {
			drawPauseIcon(drawList, pos, size, col)
		}

	case islandIconSettings:
		// gearshape 近似：外环 + 8 齿 + 中心孔，全描边
		ring := size * 0.95
		drawList.AddCircleV(pos, ring, col, 0, thickness)
		for i := 0; i < 8; i++ {
			a := float32(i) * float32(math.Pi) / 4
			ca := float32(math.Cos(float64(a)))
			sa := float32(math.Sin(float64(a)))
			drawList.AddLineV(
				imgui.Vec2{X: pos.X + ring*ca, Y: pos.Y + ring*sa},
				imgui.Vec2{X: pos.X + ring*1.35*ca, Y: pos.Y + ring*1.35*sa},
				col, thickness,
			)
		}
		drawList.AddCircleV(pos, ring*0.45, col, 0, thickness)

	case islandIconClose:
		// xmark：两条细斜线
		off := size * 0.7
		drawList.AddLineV(
			imgui.Vec2{X: pos.X - off, Y: pos.Y - off},
			imgui.Vec2{X: pos.X + off, Y: pos.Y + off},
			col, thickness,
		)
		drawList.AddLineV(
			imgui.Vec2{X: pos.X + off, Y: pos.Y - off},
			imgui.Vec2{X: pos.X - off, Y: pos.Y + off},
			col, thickness,
		)
	}
}

func drawPlayIcon(drawList *imgui.DrawList, pos imgui.Vec2, size float32, color uint32) {
	p1 := imgui.Vec2{X: pos.X - size*0.5, Y: pos.Y - size}
	p2 := imgui.Vec2{X: pos.X - size*0.5, Y: pos.Y + size}
	p3 := imgui.Vec2{X: pos.X + size, Y: pos.Y}
	drawList.AddTriangleFilled(p1, p2, p3, color)
}

// drawPauseIcon 两根圆角竖条（iOS 风格）。
func drawPauseIcon(drawList *imgui.DrawList, pos imgui.Vec2, size float32, color uint32) {
	barW := size * 0.42
	barH := size * 1.7
	gap := size * 0.36
	rounding := barW / 2
	drawList.AddRectFilledV(
		imgui.Vec2{X: pos.X - gap - barW, Y: pos.Y - barH/2},
		imgui.Vec2{X: pos.X - gap, Y: pos.Y + barH/2},
		color, rounding, imgui.DrawFlagsRoundCornersAll,
	)
	drawList.AddRectFilledV(
		imgui.Vec2{X: pos.X + gap, Y: pos.Y - barH/2},
		imgui.Vec2{X: pos.X + gap + barW, Y: pos.Y + barH/2},
		color, rounding, imgui.DrawFlagsRoundCornersAll,
	)
}

func easeOutCubic(t float32) float32 {
	t = t - 1
	return t*t*t + 1
}
