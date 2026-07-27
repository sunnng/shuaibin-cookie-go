//go:build android && cgo

package ui

import (
	"math"
	"time"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// floatingIsland 灵动岛样式悬浮窗：深色胶囊，点击展开为圆角卡片，
// 提供 开始/停止、暂停/继续、设置、关闭 四个操作（图标 + 文字标签）。
// 默认顶部居中；收起态按住胶囊可拖动（位移超阈值视为拖动，否则视为点按），
// 位置持久化到 ui.json（island_pos_x/y），重启后恢复——游戏顶栏有货币等关键数据，
// 用户可自行把岛挪到不遮挡的位置。展开后再点卡片空白处或卡片外任意处收起，
// 展开超过 islandAutoCollapse 无操作也会自动收起，避免长时间遮挡游戏画面干扰识别。
// 业务运行状态由 Shell 提供，岛内部只保留 UI 状态。
// 视觉为糖果积木语言（docs/ui-redesign/design-system.md §4.1）：深色底 +
// 3px 状态色描边 + 纸面圆角方块按钮 + 墨色图标；岛不走 Theme，缩放为本岛局部策略。
type floatingIsland struct {
	ScreenWidth  int
	ScreenHeight int

	IsExpanded bool
	ExpandAnim float32

	expandedAt time.Time

	// 拖动：customPos 为 false 时默认顶部居中；拖动后置位并持久化。
	customPos          bool
	posX, posY         float32
	dragging           bool // 按下命中胶囊后的跟踪中（事后区分点按/拖动）
	dragMoved          bool // 本次跟踪已越过拖动阈值
	dragOffX, dragOffY float32
}

// newFloatingIsland 创建灵动岛：初始为收起的胶囊。
func newFloatingIsland() *floatingIsland {
	return &floatingIsland{}
}

// 尺寸基准：1600 宽（与平台层 1600×900 基准坐标一致），按屏幕宽等比缩放并夹取。
const (
	islandRefWidth   = float32(1600)
	islandTopMargin  = float32(14)
	islandPillHeight = float32(64)
	islandPillMinW   = float32(240)
	islandCardWidth  = float32(600)
	islandCardHeight = float32(224)
	islandCardRadius = float32(28)
	islandBtnSize    = float32(64)  // 纸面方块边长
	islandBtnHit     = float32(88)  // 热区（触控目标）
	islandBtnSpacing = float32(100) // 按钮中心距（64 方块 + 36 间隙）
	islandBtnRadius  = float32(18)  // 方块圆角
)

// islandSizeFactor 整体缩小系数（用户反馈默认尺寸遮挡游戏顶栏过多）：
// 几何经 scale() 乘它，字号经 drawIslandText/measureIslandText 乘它，两边同系数保持比例。
const islandSizeFactor = float32(0.75)

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
	// #17130C 97% 不透明：深色底压得住任何游戏画面。
	islandBg      = imgui.Vec4{X: 0.090, Y: 0.075, Z: 0.047, W: 0.97}
	islandText    = imgui.Vec4{X: 0.929, Y: 0.914, Z: 0.886, W: 1.0} // #EDE9E2
	islandSubText = imgui.Vec4{X: 0.729, Y: 0.698, Z: 0.651, W: 1.0} // #B9B2A6 按钮文字标签
)

// islandStateColor 状态点 / 胶囊描边共用：空闲蓝 / 运行绿 / 暂停橙（糖果色）。
func islandStateColor(state ScriptState) imgui.Vec4 {
	return candyStateColor(state)
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
// state 取自 shell.ScriptState()；label 优先使用 shell.StatusText()，为空时回退为
// shell.Title() + " · " + 状态标签。
// 命中回调直连 Shell：开始/停止 -> shell.StartStop()；暂停/继续 -> shell.PauseResume()；
// 设置 -> shell.OpenPanel()；关闭 -> shell.Exit()。
// ctx 仅用于签名一致（由 RunShell 注入），本岛使用自有缩放策略，不经 ctx.S。
func (isl *floatingIsland) Draw(ctx *Ctx, shell *Shell) {
	_ = ctx // ctx 仅签名一致，岛使用局部缩放

	state := shell.ScriptState()
	label := shell.StatusText()
	if label == "" {
		label = shell.Title() + " · " + islandStateLabel(state)
	}

	if isl.ScreenWidth == 0 {
		w, h, _, _ := device.GetDisplayInfo(0)
		if w == 0 {
			w, h = 1600, 900
		}
		isl.ScreenWidth = w
		isl.ScreenHeight = h
		// 恢复上次拖动持久化的位置（RunShell 首帧 LoadConfig 先于首次 Draw）。
		if st := shell.Store(); st != nil && st.HasKey(islandPosXKey) && st.HasKey(islandPosYKey) {
			isl.posX = float32(st.GetFloat(islandPosXKey))
			isl.posY = float32(st.GetFloat(islandPosYKey))
			isl.customPos = true
		}
	}

	// 展开超时自动收起：卡片遮挡面积较大，长时间盖在游戏上会干扰截图识别。
	if isl.IsExpanded && !isl.expandedAt.IsZero() && time.Since(isl.expandedAt) > islandAutoCollapse {
		isl.IsExpanded = false
	}

	isl.updateAnimation()
	isl.drawWindow(shell, state, label)
}

func (isl *floatingIsland) scale() float32 {
	s := float32(isl.ScreenWidth) / islandRefWidth
	if s < 0.8 {
		s = 0.8
	}
	if s > 1.6 {
		s = 1.6
	}
	return s * islandSizeFactor
}

type islandButton struct {
	pos     imgui.Vec2 // 方块中心
	size    float32    // 方块边长
	hit     float32    // 热区边长
	icon    int
	label   string
	enabled bool
}

type islandLayout struct {
	x, y, w, h, radius float32
	scale              float32
	anim               float32
	buttons            []islandButton
}

// layout 计算本帧胶囊/卡片矩形（展开动画对宽高与圆角做插值）与按钮排布。
func (isl *floatingIsland) layout(label string, state ScriptState) islandLayout {
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
	// 拖动过的自定义位置（夹取到屏幕内，展开卡片同样整体保持可见）；
	// 否则默认顶部居中。
	if isl.customPos {
		l.x, l.y = clampIslandPos(isl.posX, isl.posY, l.w, l.h,
			float32(isl.ScreenWidth), float32(isl.ScreenHeight))
	} else {
		l.x = (float32(isl.ScreenWidth) - l.w) / 2
		l.y = islandTopMargin * s
	}

	// 按钮方块居中排布，底部预留文字标签行。
	labelH := measureIslandText("设置").Y
	boxCY := l.y + cardH - 20*s - labelH - 8*s - islandBtnSize*s/2
	centerX := l.x + l.w/2
	labels := islandButtonLabels(state)
	for i := 0; i < 4; i++ {
		l.buttons = append(l.buttons, islandButton{
			pos:     imgui.Vec2{X: centerX + (float32(i)-1.5)*islandBtnSpacing*s, Y: boxCY},
			size:    islandBtnSize * s,
			hit:     islandBtnHit * s,
			icon:    i,
			label:   labels[i],
			enabled: i != islandIconPauseResume || state != StateIdle, // 空闲时暂停钮禁用
		})
	}
	return l
}

// islandButtonLabels 展开卡片的按钮文字标签（随状态切换）。
func islandButtonLabels(state ScriptState) [4]string {
	labels := [4]string{"开始", "暂停", "设置", "关闭"}
	if state != StateIdle {
		labels[0] = "停止"
	}
	if state == StatePaused {
		labels[1] = "继续"
	}
	return labels
}

// pillWidth 胶囊宽度随状态文字自适应（灵动岛风格：内容多宽胶囊就多宽），
// 下限 islandPillMinW，避免短文案时胶囊过窄。
func (isl *floatingIsland) pillWidth(label string, s float32) float32 {
	textW := measureIslandText(label).X
	dotD := float32(18) * s
	gap := float32(10) * s
	pad := float32(24) * s
	w := pad*2 + dotD + gap + textW
	if min := islandPillMinW * s; w < min {
		w = min
	}
	return w
}

func (isl *floatingIsland) drawWindow(shell *Shell, state ScriptState, label string) {
	l := isl.layout(label, state)

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

	isl.handleInput(l, shell)

	pMin := imgui.Vec2{X: l.x, Y: l.y}
	pMax := imgui.Vec2{X: l.x + l.w, Y: l.y + l.h}
	drawList.AddRectFilledV(pMin, pMax, imgui.ColorU32Vec4(islandBg), l.radius, imgui.DrawFlagsRoundCornersAll)
	// 3px 状态色描边：与状态点同色，远远一瞥即知状态。
	borderW := 3 * l.scale
	if borderW < 2 {
		borderW = 2
	}
	drawList.AddRectV(pMin, pMax, imgui.ColorU32Vec4(islandStateColor(state)), l.radius, imgui.DrawFlagsRoundCornersAll, borderW)

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

// drawPillContent 收起态：状态点 + 状态文案，整体水平居中。
func (isl *floatingIsland) drawPillContent(drawList *imgui.DrawList, l islandLayout, state ScriptState, label string) {
	textSz := measureIslandText(label)
	dotD := float32(18) * l.scale
	gap := float32(10) * l.scale
	groupW := dotD + gap + textSz.X

	cy := l.y + l.h/2
	dotX := l.x + (l.w-groupW)/2 + dotD/2
	drawList.AddCircleFilled(imgui.Vec2{X: dotX, Y: cy}, dotD/2, imgui.ColorU32Vec4(islandStateColor(state)))
	drawIslandText(
		drawList,
		imgui.Vec2{X: dotX + dotD/2 + gap, Y: cy - textSz.Y/2},
		islandText,
		label,
	)
}

// measureIslandText 测量灵动岛文字尺寸：在 measureLabelSize（含 CJK 保底）基础上
// 乘 islandSizeFactor，与 drawIslandText 的缩放字号一致。
func measureIslandText(label string) imgui.Vec2 {
	sz := measureLabelSize(label)
	sz.X *= islandSizeFactor
	sz.Y *= islandSizeFactor
	return sz
}

// drawIslandText 按 islandSizeFactor 缩小字号绘制灵动岛文字（几何同系数缩小，保持比例）。
func drawIslandText(drawList *imgui.DrawList, pos imgui.Vec2, col imgui.Vec4, text string) {
	drawList.AddTextFontPtr(
		imgui.CurrentFont(), imgui.FontSize()*islandSizeFactor, pos,
		imgui.ColorU32Vec4(col), text,
	)
}

// drawCardContent 展开态：顶部状态行 + 一排四个纸面方块按钮（墨色图标 + 文字标签）。
func (isl *floatingIsland) drawCardContent(drawList *imgui.DrawList, l islandLayout, state ScriptState, label string) {
	textSz := measureIslandText(label)
	dotD := float32(14) * l.scale
	gap := float32(8) * l.scale
	groupW := dotD + gap + textSz.X

	rowY := l.y + 44*l.scale
	dotX := l.x + (l.w-groupW)/2 + dotD/2
	drawList.AddCircleFilled(imgui.Vec2{X: dotX, Y: rowY}, dotD/2, imgui.ColorU32Vec4(islandStateColor(state)))
	drawIslandText(
		drawList,
		imgui.Vec2{X: dotX + dotD/2 + gap, Y: rowY - textSz.Y/2},
		islandText,
		label,
	)

	for _, b := range l.buttons {
		half := b.size / 2
		pMin := imgui.Vec2{X: b.pos.X - half, Y: b.pos.Y - half}
		pMax := imgui.Vec2{X: b.pos.X + half, Y: b.pos.Y + half}
		box := candyPaper
		if b.icon == islandIconClose {
			box = candyRed // 破坏性操作：红底
		}
		if !b.enabled {
			box.W *= 0.35
		}
		radius := islandBtnRadius * l.scale
		drawList.AddRectFilledV(pMin, pMax, imgui.ColorU32Vec4(box), radius, imgui.DrawFlagsRoundCornersAll)
		drawIslandIcon(drawList, b.pos, half, b.icon, state, b.enabled)

		// 图标下文字标签（解决纯图标新用户不可发现的痛点）。
		labSz := measureIslandText(b.label)
		labCol := islandSubText
		if !b.enabled {
			labCol.W *= 0.35
		}
		drawIslandText(
			drawList,
			imgui.Vec2{X: b.pos.X - labSz.X/2, Y: pMax.Y + 8*l.scale},
			labCol,
			b.label,
		)
	}
}

// islandDragThreshold 按住胶囊位移超过该像素数即判定为拖动（否则松手视为点按展开）。
const islandDragThreshold = float32(12)

// handleInput 命中检测：收起态按住胶囊可拖动（松手持久化位置；未越过拖动阈值
// 视为点按展开）；展开态点按钮（热区 88×88 方块）执行并收起，再点卡片空白处或
// 卡片外任意处收起。禁用的按钮（空闲时的暂停）不响应。
func (isl *floatingIsland) handleInput(l islandLayout, shell *Shell) {
	// 动画进行中不响应，避免展开/收起过程中的误触。
	if isl.ExpandAnim > 0.05 && isl.ExpandAnim < 0.95 {
		return
	}
	m := imgui.MousePos()

	if !isl.IsExpanded {
		// 按下命中胶囊：开始跟踪，松手时按位移区分点按/拖动。
		if imgui.IsMouseClickedBoolV(imgui.MouseButtonLeft, false) && pointInLayout(m, l) {
			isl.dragging = true
			isl.dragMoved = false
			isl.dragOffX = m.X - l.x
			isl.dragOffY = m.Y - l.y
		}
		if isl.dragging && imgui.IsMouseDown(imgui.MouseButtonLeft) &&
			(isl.dragMoved || imgui.IsMouseDraggingV(imgui.MouseButtonLeft, islandDragThreshold)) {
			isl.dragMoved = true
			isl.posX, isl.posY = clampIslandPos(
				m.X-isl.dragOffX, m.Y-isl.dragOffY,
				l.w, l.h, float32(isl.ScreenWidth), float32(isl.ScreenHeight))
			isl.customPos = true
		}
		if isl.dragging && imgui.IsMouseReleased(imgui.MouseButtonLeft) {
			if isl.dragMoved {
				isl.persistPos(shell)
			} else if pointInLayout(m, l) {
				isl.IsExpanded = true
				isl.expandedAt = time.Now()
			}
			isl.dragging = false
		}
		return
	}

	if !imgui.IsMouseReleased(imgui.MouseButtonLeft) {
		return
	}

	for _, b := range l.buttons {
		if !b.enabled {
			continue
		}
		hh := b.hit / 2
		if m.X < b.pos.X-hh || m.X > b.pos.X+hh || m.Y < b.pos.Y-hh || m.Y > b.pos.Y+hh {
			continue
		}
		isl.IsExpanded = false
		switch b.icon {
		case islandIconStartStop:
			shell.StartStop()
		case islandIconPauseResume:
			shell.PauseResume()
		case islandIconSettings:
			shell.OpenPanel()
		case islandIconClose:
			shell.Exit()
		}
		return
	}

	// 展开后再点灵动岛（卡片空白处）或卡片外任意处：收起。
	isl.IsExpanded = false
}

// persistPos 拖动结束后把位置写进 Store 并立即落盘（ConfigPath 非空时），
// 保证不启动脚本直接退出也不丢位置。
func (isl *floatingIsland) persistPos(shell *Shell) {
	st := shell.Store()
	if st == nil {
		return
	}
	st.SetFloat(islandPosXKey, float64(isl.posX))
	st.SetFloat(islandPosYKey, float64(isl.posY))
	if p := shell.ConfigPath(); p != "" {
		if err := st.SaveConfig(p); err != nil {
			LogErrorf("[UI] save island pos: %v", err)
		}
	}
}

func pointInLayout(m imgui.Vec2, l islandLayout) bool {
	return m.X >= l.x && m.X <= l.x+l.w && m.Y >= l.y && m.Y <= l.y+l.h
}

func (isl *floatingIsland) updateAnimation() {
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

// drawIslandIcon 积木风格图标：纸面方块上的墨色粗线图标；停止（运行中时）
// 用 candy-red 强调破坏性，关闭按钮为红底纸色图标。笔画粗细随按钮尺寸等比。
func drawIslandIcon(drawList *imgui.DrawList, pos imgui.Vec2, half float32, icon int, state ScriptState, enabled bool) {
	thickness := half * 0.14
	if thickness < 2 {
		thickness = 2
	}
	size := half * 0.42 // 图标半尺寸

	glyph := candyInk
	switch {
	case icon == islandIconClose:
		glyph = candyPaper // 红底上的纸色 ×
	case icon == islandIconStartStop && state != StateIdle:
		glyph = candyRed // 停止图标红色强调
	}
	if !enabled {
		glyph.W *= 0.35
	}
	col := imgui.ColorU32Vec4(glyph)

	switch icon {
	case islandIconStartStop:
		if state == StateIdle {
			drawPlayIcon(drawList, pos, size, col)
		} else {
			// stop：圆角方块
			h := size * 0.85
			drawList.AddRectFilledV(
				imgui.Vec2{X: pos.X - h, Y: pos.Y - h},
				imgui.Vec2{X: pos.X + h, Y: pos.Y + h},
				col, h*0.55, imgui.DrawFlagsRoundCornersAll,
			)
		}

	case islandIconPauseResume:
		if state == StatePaused {
			// 媒体控制惯例：暂停态显示播放键（点击即继续）
			drawPlayIcon(drawList, pos, size, col)
		} else {
			drawPauseIcon(drawList, pos, size, col)
		}

	case islandIconSettings:
		// 齿轮近似：外环 + 8 齿 + 中心孔，全描边
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
		// xmark：两条粗斜线
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

// drawPauseIcon 两根圆角竖条。
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
