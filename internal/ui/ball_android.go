//go:build android && cgo

package ui

import (
	"math"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// BallCallbacks 由调用方（shell）提供，悬浮球点击按钮时回调。
// 任一字段为 nil 时点击该按钮无副作用。
type BallCallbacks struct {
	OnSettings    func()
	OnStartStop   func()
	OnPauseResume func()
	OnClose       func()
}

// FloatingBall 是一个可拖动、可展开菜单的悬浮球。
// 业务运行状态由调用方通过 Draw 的 state 参数传入，球内部只保留 UI 状态。
type FloatingBall struct {
	X      float32
	Y      float32
	Radius float32

	ScreenWidth  int
	ScreenHeight int

	IsDragging  bool
	DragStartX  float32
	DragStartY  float32
	DragOffsetX float32
	DragOffsetY float32

	IsExpanded bool
	ExpandAnim float32

	IsOnRightSide bool

	IsAnimating  bool
	AnimProgress float32
	AnimStartX   float32
	AnimTargetX  float32
}

// NewFloatingBall 创建默认悬浮球：半径 40，初始贴右侧。
func NewFloatingBall() *FloatingBall {
	return &FloatingBall{Radius: 40, IsOnRightSide: true}
}

// 球体主色（依据脚本状态）。
func ballMainColor(state ScriptState) imgui.Vec4 {
	switch state {
	case StateRunning:
		return imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.3, W: 0.95}
	case StatePaused:
		return imgui.Vec4{X: 1.0, Y: 0.85, Z: 0.2, W: 0.95}
	default:
		return imgui.Vec4{X: 0.2, Y: 0.6, Z: 1.0, W: 0.95}
	}
}

var (
	ballShadow  = imgui.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 0.3}
	btnClose    = imgui.Vec4{X: 0.6, Y: 0.6, Z: 0.6, W: 1.0}
	btnPause    = imgui.Vec4{X: 1.0, Y: 0.7, Z: 0.2, W: 1.0}
	btnResume   = imgui.Vec4{X: 0.3, Y: 0.7, Z: 1.0, W: 1.0}
	btnStart    = imgui.Vec4{X: 0.3, Y: 0.8, Z: 0.4, W: 1.0}
	btnStop     = imgui.Vec4{X: 0.9, Y: 0.3, Z: 0.3, W: 1.0}
	btnSettings = imgui.Vec4{X: 0.4, Y: 0.5, Z: 0.9, W: 1.0}
)

// Draw 每帧调用：推进动画并绘制悬浮球与展开菜单。
func (ball *FloatingBall) Draw(cb BallCallbacks, state ScriptState) {
	if ball.ScreenWidth == 0 {
		w, h, _, _ := device.GetDisplayInfo(0)
		if w == 0 {
			w, h = 1080, 1920
		}
		ball.ScreenWidth = w
		ball.ScreenHeight = h
		ball.X = float32(w) - ball.Radius - 10
		ball.Y = float32(h) / 2
	}

	ball.updateAnimations()
	ball.drawInteractionWindow(cb, state)
}

func (ball *FloatingBall) drawInteractionWindow(cb BallCallbacks, state ScriptState) {
	var windowX, windowY, windowW, windowH float32

	if ball.IsExpanded || ball.ExpandAnim > 0 {
		buttonSpacing := float32(90) * easeOutCubic(ball.ExpandAnim)
		if ball.IsOnRightSide {
			windowX = ball.X - buttonSpacing*4 - 50
		} else {
			windowX = ball.X - 50
		}
		windowW = buttonSpacing*4 + 100
		windowY = ball.Y - 50
		windowH = 100
	} else {
		windowX = ball.X - ball.Radius - 10
		windowY = ball.Y - ball.Radius - 10
		windowW = ball.Radius*2 + 20
		windowH = ball.Radius*2 + 20
	}

	imgui.SetNextWindowPosV(imgui.Vec2{X: windowX, Y: windowY}, imgui.CondAlways, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: windowW, Y: windowH}, imgui.CondAlways)

	imgui.PushStyleColorVec4(imgui.ColWindowBg, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{})

	flags := imgui.WindowFlagsNoTitleBar | imgui.WindowFlagsNoResize |
		imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoBackground |
		imgui.WindowFlagsNoSavedSettings | imgui.WindowFlagsNoMove

	imgui.BeginV("##FloatBall", nil, flags)

	drawList := imgui.WindowDrawList()

	ball.handleDragging()

	if ball.IsExpanded || ball.ExpandAnim > 0 {
		ball.drawExpandedMenu(drawList, cb, state)
	} else {
		ball.drawSmallBall(drawList, state)
	}

	imgui.End()

	imgui.PopStyleVar()
	imgui.PopStyleVar()
	imgui.PopStyleColor()
}

func (ball *FloatingBall) drawSmallBall(drawList *imgui.DrawList, state ScriptState) {
	pos := imgui.Vec2{X: ball.X, Y: ball.Y}

	shadowOffset := float32(3)
	drawList.AddCircleFilled(
		imgui.Vec2{X: pos.X + shadowOffset, Y: pos.Y + shadowOffset},
		ball.Radius,
		imgui.ColorU32Vec4(ballShadow),
	)

	mainColor := ballMainColor(state)
	for i := 3; i > 0; i-- {
		glowRadius := ball.Radius + float32(i)*4
		alpha := 0.1 / float32(i)
		glowColor := imgui.Vec4{
			X: mainColor.X,
			Y: mainColor.Y,
			Z: mainColor.Z,
			W: alpha,
		}
		drawList.AddCircleFilled(pos, glowRadius, imgui.ColorU32Vec4(glowColor))
	}

	drawList.AddCircleFilled(pos, ball.Radius, imgui.ColorU32Vec4(mainColor))

	highlightPos := imgui.Vec2{X: pos.X - 10, Y: pos.Y - 10}
	drawList.AddCircleFilled(highlightPos, ball.Radius*0.3,
		imgui.ColorU32Vec4(imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.5}))
}

func (ball *FloatingBall) drawExpandedMenu(drawList *imgui.DrawList, cb BallCallbacks, state ScriptState) {
	anim := easeOutCubic(ball.ExpandAnim)

	buttonRadius := float32(35)
	spacing := float32(90) * anim

	mainColor := ballMainColor(state)
	pauseColor := btnPause
	if state == StatePaused {
		pauseColor = btnResume
	}
	startColor := btnStart
	if state != StateIdle {
		startColor = btnStop
	}

	type btnSpec struct {
		pos   imgui.Vec2
		color imgui.Vec4
		index int
	}

	var positions []imgui.Vec2
	if ball.IsOnRightSide {
		positions = []imgui.Vec2{
			{X: ball.X, Y: ball.Y},             // 0 Logo
			{X: ball.X - spacing, Y: ball.Y},   // 1 关闭
			{X: ball.X - spacing*2, Y: ball.Y}, // 2 暂停/恢复
			{X: ball.X - spacing*3, Y: ball.Y}, // 3 开始/停止
			{X: ball.X - spacing*4, Y: ball.Y}, // 4 设置
		}
	} else {
		positions = []imgui.Vec2{
			{X: ball.X, Y: ball.Y},
			{X: ball.X + spacing, Y: ball.Y},
			{X: ball.X + spacing*2, Y: ball.Y},
			{X: ball.X + spacing*3, Y: ball.Y},
			{X: ball.X + spacing*4, Y: ball.Y},
		}
	}

	buttons := []btnSpec{
		{positions[0], mainColor, 0},
		{positions[1], btnClose, 1},
		{positions[2], pauseColor, 2},
		{positions[3], startColor, 3},
		{positions[4], btnSettings, 4},
	}

	for _, btn := range buttons {
		drawList.AddCircleFilled(
			imgui.Vec2{X: btn.pos.X + 3, Y: btn.pos.Y + 3},
			buttonRadius,
			imgui.ColorU32Vec4(ballShadow),
		)
		drawList.AddCircleFilled(btn.pos, buttonRadius, imgui.ColorU32Vec4(btn.color))

		if btn.index == 0 {
			highlightPos := imgui.Vec2{X: btn.pos.X - 8, Y: btn.pos.Y - 8}
			drawList.AddCircleFilled(highlightPos, buttonRadius*0.25,
				imgui.ColorU32Vec4(imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.3}))
		}

		ball.drawButtonIcon(drawList, btn.pos, buttonRadius, btn.index, state)

		if ball.IsExpanded && anim > 0.9 {
			ball.checkButtonClick(btn.pos, buttonRadius, btn.index, cb, state)
		}
	}
}

func (ball *FloatingBall) handleDragging() {
	mousePos := imgui.MousePos()

	dx := mousePos.X - ball.X
	dy := mousePos.Y - ball.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if imgui.IsMouseDown(imgui.MouseButtonLeft) {
		if !ball.IsDragging && !ball.IsExpanded && distance < ball.Radius {
			ball.IsDragging = true
			ball.DragStartX = mousePos.X
			ball.DragStartY = mousePos.Y
			ball.DragOffsetX = mousePos.X - ball.X
			ball.DragOffsetY = mousePos.Y - ball.Y
		}

		if ball.IsDragging {
			ball.X = mousePos.X - ball.DragOffsetX
			ball.Y = mousePos.Y - ball.DragOffsetY

			if ball.X < ball.Radius {
				ball.X = ball.Radius
			}
			if ball.X > float32(ball.ScreenWidth)-ball.Radius {
				ball.X = float32(ball.ScreenWidth) - ball.Radius
			}
			if ball.Y < ball.Radius {
				ball.Y = ball.Radius
			}
			if ball.Y > float32(ball.ScreenHeight)-ball.Radius {
				ball.Y = float32(ball.ScreenHeight) - ball.Radius
			}
		}
	} else if imgui.IsMouseReleased(imgui.MouseButtonLeft) {
		if ball.IsDragging {
			dragDistance := float32(math.Sqrt(
				float64((mousePos.X-ball.DragStartX)*(mousePos.X-ball.DragStartX) +
					(mousePos.Y-ball.DragStartY)*(mousePos.Y-ball.DragStartY)),
			))

			if dragDistance < 10 && !ball.IsExpanded {
				ball.IsExpanded = true
			} else if dragDistance >= 10 {
				ball.startAutoAlign()
			}

			ball.IsDragging = false
		}
	}
}

func (ball *FloatingBall) startAutoAlign() {
	ball.IsAnimating = true
	ball.AnimProgress = 0
	ball.AnimStartX = ball.X

	if ball.X > float32(ball.ScreenWidth)/2 {
		ball.IsOnRightSide = true
		ball.AnimTargetX = float32(ball.ScreenWidth) - ball.Radius - 10
	} else {
		ball.IsOnRightSide = false
		ball.AnimTargetX = ball.Radius + 10
	}
}

func (ball *FloatingBall) updateAnimations() {
	if ball.IsAnimating {
		ball.AnimProgress += 0.05
		if ball.AnimProgress >= 1.0 {
			ball.AnimProgress = 1.0
			ball.IsAnimating = false
			ball.X = ball.AnimTargetX
		} else {
			t := easeOutBounce(ball.AnimProgress)
			ball.X = ball.AnimStartX + (ball.AnimTargetX-ball.AnimStartX)*t
		}
	}

	if ball.IsExpanded {
		if ball.ExpandAnim < 1.0 {
			ball.ExpandAnim += 0.08
			if ball.ExpandAnim > 1.0 {
				ball.ExpandAnim = 1.0
			}
		}
	} else {
		if ball.ExpandAnim > 0.0 {
			ball.ExpandAnim -= 0.08
			if ball.ExpandAnim < 0.0 {
				ball.ExpandAnim = 0.0
			}
		}
	}
}

func (ball *FloatingBall) checkButtonClick(pos imgui.Vec2, radius float32, buttonIndex int, cb BallCallbacks, state ScriptState) {
	if !imgui.IsMouseReleased(imgui.MouseButtonLeft) {
		return
	}

	mousePos := imgui.MousePos()
	dx := mousePos.X - pos.X
	dy := mousePos.Y - pos.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if distance >= radius {
		return
	}

	switch buttonIndex {
	case 0: // Logo - 收起菜单
		ball.IsExpanded = false
	case 1: // 关闭
		ball.IsExpanded = false
		if cb.OnClose != nil {
			cb.OnClose()
		}
	case 2: // 暂停/恢复
		ball.IsExpanded = false
		if cb.OnPauseResume != nil {
			cb.OnPauseResume()
		}
	case 3: // 开始/停止
		ball.IsExpanded = false
		if cb.OnStartStop != nil {
			cb.OnStartStop()
		}
	case 4: // 设置
		ball.IsExpanded = false
		if cb.OnSettings != nil {
			cb.OnSettings()
		}
	}
}

func easeOutCubic(t float32) float32 {
	t = t - 1
	return t*t*t + 1
}

func easeOutBounce(t float32) float32 {
	if t < 1/2.75 {
		return 7.5625 * t * t
	} else if t < 2/2.75 {
		t -= 1.5 / 2.75
		return 7.5625*t*t + 0.75
	} else if t < 2.5/2.75 {
		t -= 2.25 / 2.75
		return 7.5625*t*t + 0.9375
	}
	t -= 2.625 / 2.75
	return 7.5625*t*t + 0.984375
}

func (ball *FloatingBall) drawButtonIcon(drawList *imgui.DrawList, pos imgui.Vec2, radius float32, buttonIndex int, state ScriptState) {
	iconColor := imgui.ColorU32Vec4(imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.9})
	iconSize := radius * 0.4

	switch buttonIndex {
	case 0: // Logo - 小圆点
		drawList.AddCircleFilled(pos, iconSize*0.5, iconColor)

	case 1: // 关闭 - X形
		thickness := float32(3)
		offset := iconSize * 0.7
		drawList.AddLineV(
			imgui.Vec2{X: pos.X - offset, Y: pos.Y - offset},
			imgui.Vec2{X: pos.X + offset, Y: pos.Y + offset},
			iconColor, thickness,
		)
		drawList.AddLineV(
			imgui.Vec2{X: pos.X + offset, Y: pos.Y - offset},
			imgui.Vec2{X: pos.X - offset, Y: pos.Y + offset},
			iconColor, thickness,
		)

	case 2: // 暂停/恢复
		if state == StatePaused {
			drawResumeIcon(drawList, pos, iconSize, iconColor)
		} else {
			drawPauseIcon(drawList, pos, iconSize, iconColor)
		}

	case 3: // 开始/停止
		if state == StateIdle {
			drawPlayIcon(drawList, pos, iconSize, iconColor)
		} else {
			drawList.AddRectFilled(
				imgui.Vec2{X: pos.X - iconSize, Y: pos.Y - iconSize},
				imgui.Vec2{X: pos.X + iconSize, Y: pos.Y + iconSize},
				iconColor,
			)
		}

	case 4: // 设置 - 齿轮近似（外圆环 + 内圆 + 中心点）
		gearOuter := iconSize * 1.1
		gearInner := iconSize * 0.55
		ringColor := iconColor
		// 外环
		segments := 12
		for i := 0; i < segments; i++ {
			a1 := float32(i) / float32(segments) * 2 * float32(math.Pi)
			a2 := float32(i+1) / float32(segments) * 2 * float32(math.Pi)
			drawList.AddLineV(
				imgui.Vec2{X: pos.X + gearOuter*float32(math.Cos(float64(a1))), Y: pos.Y + gearOuter*float32(math.Sin(float64(a1)))},
				imgui.Vec2{X: pos.X + gearOuter*float32(math.Cos(float64(a2))), Y: pos.Y + gearOuter*float32(math.Sin(float64(a2)))},
				ringColor, 3,
			)
		}
		// 内圆
		drawList.AddCircleFilled(pos, gearInner, ringColor)
		// 中心镂空点
		drawList.AddCircleFilled(pos, gearInner*0.4,
			imgui.ColorU32Vec4(btnSettings))
	}
}

func drawPlayIcon(drawList *imgui.DrawList, pos imgui.Vec2, size float32, color uint32) {
	p1 := imgui.Vec2{X: pos.X - size*0.5, Y: pos.Y - size}
	p2 := imgui.Vec2{X: pos.X - size*0.5, Y: pos.Y + size}
	p3 := imgui.Vec2{X: pos.X + size, Y: pos.Y}
	drawList.AddTriangleFilled(p1, p2, p3, color)
}

func drawPauseIcon(drawList *imgui.DrawList, pos imgui.Vec2, size float32, color uint32) {
	barWidth := size * 0.6
	barHeight := size * 2
	gap := size * 0.5

	drawList.AddRectFilled(
		imgui.Vec2{X: pos.X - gap - barWidth, Y: pos.Y - barHeight/2},
		imgui.Vec2{X: pos.X - gap, Y: pos.Y + barHeight/2},
		color,
	)
	drawList.AddRectFilled(
		imgui.Vec2{X: pos.X + gap, Y: pos.Y - barHeight/2},
		imgui.Vec2{X: pos.X + gap + barWidth, Y: pos.Y + barHeight/2},
		color,
	)
}

func drawResumeIcon(drawList *imgui.DrawList, pos imgui.Vec2, size float32, color uint32) {
	radius := size * 0.9
	thickness := float32(3)

	segments := 20
	startAngle := float32(0.5)
	endAngle := float32(6.0)

	for i := 0; i < segments; i++ {
		angle1 := startAngle + (endAngle-startAngle)*float32(i)/float32(segments)
		angle2 := startAngle + (endAngle-startAngle)*float32(i+1)/float32(segments)

		p1 := imgui.Vec2{
			X: pos.X + radius*float32(math.Cos(float64(angle1))),
			Y: pos.Y + radius*float32(math.Sin(float64(angle1))),
		}
		p2 := imgui.Vec2{
			X: pos.X + radius*float32(math.Cos(float64(angle2))),
			Y: pos.Y + radius*float32(math.Sin(float64(angle2))),
		}

		drawList.AddLineV(p1, p2, color, thickness)
	}

	arrowSize := size * 0.4
	arrowAngle := endAngle
	arrowX := pos.X + radius*float32(math.Cos(float64(arrowAngle)))
	arrowY := pos.Y + radius*float32(math.Sin(float64(arrowAngle)))

	perpAngle := arrowAngle + 1.57
	p1 := imgui.Vec2{
		X: arrowX + arrowSize*float32(math.Cos(float64(perpAngle))),
		Y: arrowY + arrowSize*float32(math.Sin(float64(perpAngle))),
	}
	p2 := imgui.Vec2{
		X: arrowX - arrowSize*0.5*float32(math.Cos(float64(perpAngle))),
		Y: arrowY - arrowSize*0.5*float32(math.Sin(float64(perpAngle))),
	}
	p3 := imgui.Vec2{
		X: arrowX + arrowSize*float32(math.Cos(float64(arrowAngle))),
		Y: arrowY + arrowSize*float32(math.Sin(float64(arrowAngle))),
	}

	drawList.AddTriangleFilled(p1, p2, p3, color)
}
