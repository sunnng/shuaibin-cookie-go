//go:build android && cgo

package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/images"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// ==================== 标签栏 配置结构体 ====================
// UITabStyle 标签栏整体样式配置结构体
// 统一管理父子标签按钮颜色、尺寸、字体、边框、圆角等外观参数
type UITabStyle struct {
	FontScale       float32 // 标签文字字体缩放比例
	TextColor       string  // 标签未激活状态文字颜色(十六进制色值)
	BgColor         string  // 标签未激活状态背景颜色(十六进制色值)
	HoveredColor    string  // 标签鼠标悬停状态背景颜色(十六进制色值)
	ActiveColor     string  // 标签选中激活状态背景颜色(十六进制色值)
	ActiveTextColor string  // 标签选中激活状态文字颜色(十六进制色值)
	BorderColor     string  // 标签按钮边框、底部分割线颜色(十六进制色值)
	Height          float32 // 标签按钮高度
	MinWidth        float32 // 标签按钮最小宽度
	PaddingX        float32 // 标签文字左右内边距宽度
	BorderSize      float32 // 标签按钮边框粗细，0 为关闭边框
	BorderRounding  float32 // 标签按钮圆角半径，0 为直角
}

// 父子标签默认样式
var (
	uiTabActiveMap = map[string]int{}
	baseTabStyle   = UITabStyle{
		FontScale:       0.9,       // 标签文字字体缩放比例
		TextColor:       "#ffffff", // 未激活文字：沉稳的深蓝灰
		BgColor:         "#ffffff", // 未激活背景：极淡的灰白
		HoveredColor:    "#686868", // 悬停背景：稍深一点的灰，给出反馈
		ActiveColor:     "#686868", // 激活背景：优雅的靛蓝色
		ActiveTextColor: "#ffffff", // 激活文字：白色
		BorderColor:     "#cbd5e1", // 边框/分割线：浅灰蓝，不抢眼
		Height:          48,
		MinWidth:        80,
		PaddingX:        16,
		BorderSize:      1,
		BorderRounding:  8, // 8px 圆角更现代
	}
	// 当前生效样式
	curTabStyle = baseTabStyle
)

// HexToVec4 将十六进制颜色字符串转换为 imgui.Vec4
// 支持 "#RRGGBB" 和 "#RRGGBBAA" 格式
func HexToVec4(hex string) imgui.Vec4 {
	hex = strings.TrimSpace(hex)
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 && len(hex) != 8 {
		return imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}
	}

	var r, g, b, a uint8 = 255, 255, 255, 255

	if len(hex) == 6 {
		fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	} else {
		fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
	}

	return imgui.Vec4{
		X: float32(r) / 255.0,
		Y: float32(g) / 255.0,
		Z: float32(b) / 255.0,
		W: float32(a) / 255.0,
	}
}

// UI_创建左侧父标签页 生成多标签切换控件
// id 控件唯一ID
// titles 父标签名称
// pages 要渲染的子标签
func UI_创建左侧父标签页(id string, titles []string, pages []func()) {
	if len(titles) == 0 || len(pages) == 0 || len(titles) != len(pages) {
		return
	}
	activeIdx, ok := uiTabActiveMap[id]
	if !ok || activeIdx < 0 || activeIdx >= len(titles) {
		activeIdx = 0
		uiTabActiveMap[id] = 0
	}
	style := curTabStyle
	leftWidth := float32(150)

	// 左侧父标签区域
	imgui.PushStyleColorVec4(imgui.ColChildBg, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.07})
	imgui.PushStyleColorVec4(imgui.ColBorder, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.16})
	imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, 10)

	imgui.BeginChildStrV(
		id+"_left",
		imgui.Vec2{X: leftWidth, Y: 0},
		imgui.ChildFlagsBorders,
		imgui.WindowFlagsNone, //imgui.WindowFlagsNoScrollbar, //imgui.WindowFlagsNone
	)
	imgui.Dummy(imgui.Vec2{X: 0, Y: 0})
	for i, title := range titles {
		active := i == activeIdx
		colBg := imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.06}
		colHovered := imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.14}
		colActive := imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.20}
		colText := HexToVec4(style.TextColor)
		colBorder := imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.14}
		label := "  " + title

		if active {
			colBg = imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.30}
			colHovered = imgui.Vec4{X: 0.78, Y: 0.90, Z: 1, W: 0.36}
			colActive = imgui.Vec4{X: 0.62, Y: 0.80, Z: 1, W: 0.42}
			colText = HexToVec4(style.ActiveTextColor)
			colBorder = imgui.Vec4{X: 0.92, Y: 0.97, Z: 1, W: 0.38}
			label = "△" + title
		}

		imgui.PushStyleColorVec4(imgui.ColButton, colBg)
		imgui.PushStyleColorVec4(imgui.ColButtonHovered, colHovered)
		imgui.PushStyleColorVec4(imgui.ColButtonActive, colActive)
		imgui.PushStyleColorVec4(imgui.ColText, colText)
		imgui.PushStyleColorVec4(imgui.ColBorder, colBorder)
		imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
		imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)

		btnW := imgui.ContentRegionAvail().X
		if btnW < 40 {
			btnW = 40
		}
		labelSz := measureLabelSize(label)
		btnH := style.Height + 5
		if need := labelSz.Y + 20; need > btnH {
			btnH = need
		}
		// 左栏过窄时用换行不安全，至少保证高度够；宽度不足则略增 child 内可用感（文本仍可能紧）
		if imgui.ButtonV(label+"##"+id+"_parent_"+fmt.Sprint(i), imgui.Vec2{X: btnW + 5, Y: btnH}) {
			uiTabActiveMap[id] = i
		}

		imgui.PopStyleVarV(2)
		imgui.PopStyleColorV(5)
		imgui.Dummy(imgui.Vec2{X: 0, Y: 5})

	}

	// 让左侧也支持手指上下滑动
	UI_EnableSlidingScroll(1.5)
	imgui.EndChild()
	imgui.PopStyleVar()
	imgui.PopStyleColorV(2)

	imgui.SameLine()
	// 右侧内容区域
	imgui.PushStyleColorVec4(imgui.ColChildBg, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.045})
	imgui.PushStyleColorVec4(imgui.ColBorder, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.12})
	imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, 10)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 12, Y: 12})

	imgui.BeginChildStrV(id+"_right", imgui.Vec2{X: 0, Y: -1}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	if pages[activeIdx] != nil {
		pages[activeIdx]()
	}
	imgui.EndChild()

	imgui.PopStyleVarV(2)
	imgui.PopStyleColorV(2)
}

// UI_创建标签栏 生成多标签切换控件
// id 控件唯一ID
// titles 子标签名称
// pages 要渲染的控件
func UI_创建标签栏(id string, titles []string, pages []func()) {
	if len(titles) == 0 || len(pages) == 0 || len(titles) != len(pages) {
		return
	}

	activeIdx, ok := uiTabActiveMap[id]
	if !ok || activeIdx < 0 || activeIdx >= len(titles) {
		activeIdx = 0
		uiTabActiveMap[id] = 0
	}

	style := curTabStyle

	const (
		tabPadX = float32(14)
		tabPadY = float32(10)
	)
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: tabPadX, Y: tabPadY})
	imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.SetWindowFontScale(style.FontScale)
	tabH := imgui.FrameHeight()
	if style.Height > tabH {
		tabH = style.Height
	}
	tabBarHeight := tabH + 20

	imgui.PushStyleColorVec4(imgui.ColChildBg, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.06})
	imgui.PushStyleColorVec4(imgui.ColBorder, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.14})
	imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, 10)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 8, Y: 10})

	if imgui.BeginChildStrV(
		id+"_glass_tab_bar",
		imgui.Vec2{X: 0, Y: tabBarHeight},
		imgui.ChildFlagsBorders,
		imgui.WindowFlagsNone,
	) {
		for i, title := range titles {
			active := i == activeIdx

			colBg := imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.08}
			colHover := imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.16}
			colActive := imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.22}
			colText := HexToVec4(style.TextColor)
			colBorder := imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.12}

			if active {
				colBg = imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.32}
				colHover = imgui.Vec4{X: 0.78, Y: 0.90, Z: 1, W: 0.40}
				colActive = imgui.Vec4{X: 0.62, Y: 0.80, Z: 1, W: 0.46}
				colText = HexToVec4(style.ActiveTextColor)
				colBorder = imgui.Vec4{X: 0.92, Y: 0.97, Z: 1, W: 0.36}
			}

			imgui.PushStyleColorVec4(imgui.ColButton, colBg)
			imgui.PushStyleColorVec4(imgui.ColButtonHovered, colHover)
			imgui.PushStyleColorVec4(imgui.ColButtonActive, colActive)
			imgui.PushStyleColorVec4(imgui.ColText, colText)
			imgui.PushStyleColorVec4(imgui.ColBorder, colBorder)

			imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
			imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)

			labelSz := measureLabelSize(title)
			tabW := labelSz.X + style.PaddingX + tabPadX*2
			if tabW < style.MinWidth {
				tabW = style.MinWidth
			}
			thisH := tabH
			if need := labelSz.Y + tabPadY*2; need > thisH {
				thisH = need
			}

			if imgui.ButtonV(title+"##"+id+"_child_"+fmt.Sprint(i), imgui.Vec2{
				X: tabW,
				Y: thisH,
			}) {
				uiTabActiveMap[id] = i
			}

			imgui.PopStyleVarV(2)
			imgui.PopStyleColorV(5)

			if i != len(titles)-1 {
				imgui.SameLine()
			}
		}

		imgui.EndChild()
	}

	imgui.PopStyleVarV(2)
	imgui.PopStyleColorV(2)
	imgui.SetWindowFontScale(1.0)
	imgui.PopStyleVarV(2) // FramePadding + ButtonTextAlign

	imgui.Spacing()

	pageHeight := imgui.ContentRegionAvail().Y
	if pageHeight < 1 {
		pageHeight = 1
	}

	if imgui.BeginChildStrV(
		id+"_page_scroll",
		imgui.Vec2{X: 0, Y: pageHeight},
		imgui.ChildFlagsNone,
		imgui.WindowFlagsNone, //imgui.WindowFlagsNoScrollbar
	) {
		UI_EnableSlidingScroll(1.5)
		if pages[activeIdx] != nil {
			pages[activeIdx]()
		}
		imgui.EndChild()
	}

}

// 允许ui滑动
func UI_EnableSlidingScroll(speed float32) {
	if speed <= 0 {
		speed = 1.2
	}
	if !imgui.IsWindowHovered() {
		return
	}
	if !imgui.IsMouseDragging(imgui.MouseButtonLeft) {
		return
	}

	delta := imgui.MouseDragDelta()
	if delta.Y == 0 {
		return
	}

	newScrollY := imgui.ScrollY() - delta.Y*speed
	if newScrollY < 0 {
		newScrollY = 0
	}

	maxY := imgui.ScrollMaxY()
	if maxY > 0 && newScrollY > maxY {
		newScrollY = maxY
	}

	imgui.SetScrollYFloat(newScrollY)
	imgui.ResetMouseDragDelta()
}

// UI_创建按钮 绘制一个液态玻璃风格按钮
// width/height: -1=填满剩余, -2=自适应, >0=固定尺寸
// 可选参数顺序: width, height, fontSize, textColor(hex)
func UI_创建按钮(id string, showName string, callback func(), opts ...interface{}) {
	width := float32(-2)
	height := float32(-2)
	fontSize := float32(0) // 0=不改 WindowFontScale；>0 为缩放因子
	textColor := "#ffffff"

	if len(opts) > 0 {
		if v, ok := opts[0].(float32); ok {
			width = v
		}
	}
	if len(opts) > 1 {
		if v, ok := opts[1].(float32); ok {
			height = v
		}
	}
	if len(opts) > 2 {
		if v, ok := opts[2].(float32); ok {
			fontSize = v
		}
	}
	if len(opts) > 3 {
		if v, ok := opts[3].(string); ok {
			textColor = v
		}
	}

	scaled := false
	if fontSize > 0 {
		imgui.SetWindowFontScale(fontSize)
		scaled = true
	}

	const (
		padX = float32(16)
		padY = float32(12)
	)
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)

	// 用 CJK 保底测宽，避免中文在按钮内横向/纵向溢出
	avail := imgui.ContentRegionAvail()
	availW, availH := avail.X, avail.Y
	fitW, fitH := fitButtonSize(showName, padX, padY)
	minH := fitH
	var btnW, btnH float32
	switch {
	case width == -1:
		btnW = availW
	case width == -2:
		btnW = fitW
		if btnW < 64 {
			btnW = 64
		}
	default:
		btnW = width
		if btnW < fitW {
			btnW = fitW // 固定宽度不足以放下文字时抬高，防止裁切
		}
	}
	switch {
	case height == -1:
		btnH = availH
	case height == -2:
		btnH = fitH
	default:
		btnH = height
	}
	if btnH > 0 && btnH < minH {
		btnH = minH
	}

	imgui.PushStyleColorVec4(imgui.ColButton, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.10})
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.20})
	imgui.PushStyleColorVec4(imgui.ColButtonActive, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.35})
	imgui.PushStyleColorVec4(imgui.ColBorder, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.18})
	imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))

	if imgui.ButtonV(showName+"##"+id, imgui.Vec2{X: btnW, Y: btnH}) {
		if callback != nil {
			callback()
		}
	}

	imgui.PopStyleColorV(5)
	imgui.PopStyleVarV(4)
	if scaled {
		imgui.SetWindowFontScale(1.0)
	}
}

// ==================== 倒计时按钮状态 ====================
type countdownState struct {
	seconds   float64
	lastTime  time.Time
	running   bool
	triggered bool
}

var countdownMap = map[string]*countdownState{}

// UI_创建倒计时按钮 带倒计时显示的液态玻璃按钮
// initSeconds: 初始倒计时秒数
// callback: 点击或倒计时结束时触发
func UI_创建倒计时按钮(initSeconds float64, id string, showName string, callback func(), opts ...interface{}) {
	state, ok := countdownMap[id]
	if !ok {
		state = &countdownState{seconds: initSeconds, lastTime: time.Now(), running: true}
		countdownMap[id] = state
	}

	if state.running {
		now := time.Now()
		state.seconds -= now.Sub(state.lastTime).Seconds()
		state.lastTime = now

		if state.seconds <= 0 && !state.triggered {
			state.seconds = 0
			state.running = false
			state.triggered = true
			if callback != nil {
				go callback()
			}
		}
	}

	label := fmt.Sprintf("%s (%.0fs)", showName, state.seconds)
	if state.triggered {
		label = showName + " (已触发)"
	}

	UI_创建按钮(id, label, func() {
		if !state.triggered {
			state.running = false
			state.triggered = true
			if callback != nil {
				go callback()
			}
		}
	}, opts...)
}

// UI_倒计时加时 给指定倒计时按钮增加秒数
func UI_倒计时加时(id string, addSeconds float64) {
	state, ok := countdownMap[id]
	if !ok {
		return
	}
	state.seconds += addSeconds
	state.running = true
	state.triggered = false
}

// UI_创建复选框 文字在左,勾选框贴右边
// 指针
// key唯一id
// showName显示的昵称
// 可选参数顺序: width, height, fontSize, textColor
//
//	width:  -1 填满剩余宽度(框贴窗口右边) / -2 自适应(框紧跟文字) / >0 指定总宽度(框贴该宽度右边)
//	height: -2 自适应 / >0 指定方框边长(通过 paddingY 实现)
//	fontSize: 绝对字号, 0 = 默认
//	textColor: 文字色 hex, 如 "#e65100"
func UI_创建复选框(store *Store, key string, showName string, args ...interface{}) {
	// 内置样式
	var (
		bgColor     string  = "#ffffff14"
		checkColor  string  = "#72b3ff"
		borderColor string  = "#ffffff28"
		paddingX    float32 = 5
		paddingY    float32 = 4
	)
	// 可传参数(默认值)
	var (
		width     float32 = -1
		height    float32 = -2
		fontSize  float32 = 0
		textColor string  = "#ffffff"
	)

	// 解析可选参数: width, height, fontSize, textColor
	idx := 0
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			width = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			height = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			fontSize = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(string); ok {
			textColor = v
		}
		idx++
	}

	if !store.HasKey(key) {
		store.SetBool(key, false)
	}
	checked := store.GetBool(key)

	// 字号: 把绝对字号换算成 scale
	scale := float32(1.0)
	if fontSize > 0 {
		if base := imgui.CalcTextSize("A").Y; base > 0 { // scale=1 时的字号
			scale = fontSize / base
		}
	}

	imgui.SetWindowFontScale(scale)

	// 用 CalcTextSize 获取当前字号下的文字高度(等价于字号)
	fontH := imgui.CalcTextSize("A").Y

	// height>0 时通过 paddingY 控制方框边长 (边长 = 字号 + paddingY*2)
	if height > 0 {
		if p := (height - fontH) / 2; p > 0 {
			paddingY = p
		} else {
			paddingY = 0
		}
	}

	// 记录行起点
	startX := imgui.CursorPosX()

	// 文字在左
	imgui.AlignTextToFramePadding()
	imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))
	imgui.Text(showName)
	imgui.PopStyleColorV(1)
	imgui.SameLine()

	// 方框边长 = 字号 + paddingY*2
	boxSize := fontH + paddingY*2

	// 计算方框目标 X, 贴右边
	avail := imgui.ContentRegionAvail()
	availW := avail.X
	cur := imgui.CursorPosX()
	var targetX float32
	switch {
	case width == -2: // 紧跟文字
		targetX = cur
	case width == -1: // 贴窗口右边
		targetX = cur + availW - boxSize
	default: // 指定总宽度
		targetX = startX + width - boxSize
	}
	if targetX < cur {
		targetX = cur
	}
	imgui.SetCursorPosX(targetX)

	// 勾选框样式
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)
	imgui.PushStyleColorVec4(imgui.ColFrameBg, HexToVec4(bgColor))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.22})
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.35})
	imgui.PushStyleColorVec4(imgui.ColCheckMark, HexToVec4(checkColor))
	imgui.PushStyleColorVec4(imgui.ColBorder, HexToVec4(borderColor))

	if imgui.Checkbox("##"+key, &checked) {
		store.SetBool(key, checked)
	}

	imgui.PopStyleColorV(5)
	imgui.PopStyleVarV(3)
	imgui.SetWindowFontScale(1.0)
}

// UI_创建输入框 文字在左，输入框在右占满剩余宽度
// hint string占用文字
// 可选参数顺序: width, height, fontSize, textColor
//
//	width:  -1 填满剩余宽度 / -2 自适应 / >0 指定输入框宽度
//	height: -2 自适应 / >0 指定高度
//	fontSize: 绝对字号(scale), 0=默认
//	textColor: 文字色 hex, 如 "#e65100"
func UI_创建输入框(store *Store, key string, showName string, hint string, args ...interface{}) {
	// 内置样式
	const (
		bgColor     = "#ffffff14"
		hintColor   = "#ffffff55"
		borderColor = "#ffffff28"
		paddingX    = float32(8)
		paddingY    = float32(5)
	)

	var (
		width     float32 = -1
		height    float32 = -2
		fontSize  float32 = 0
		textColor string  = "#ffffff"
	)

	idx := 0
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			width = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			height = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			fontSize = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(string); ok {
			textColor = v
		}
		idx++
	}
	_ = height

	scale := float32(1.0)
	if fontSize > 0 {
		if base := imgui.CalcTextSize("A").Y; base > 0 {
			scale = fontSize / base
		}
	}
	imgui.SetWindowFontScale(scale)
	defer imgui.SetWindowFontScale(1.0)

	imgui.PushStyleColorVec4(imgui.ColFrameBg, HexToVec4(bgColor))
	imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))
	imgui.PushStyleColorVec4(imgui.ColTextDisabled, HexToVec4(hintColor))
	imgui.PushStyleColorVec4(imgui.ColBorder, HexToVec4(borderColor))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.1})
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.2})
	defer imgui.PopStyleColorV(6)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
	defer imgui.PopStyleVarV(3)

	startX := imgui.CursorPosX()
	if showName != "" {
		imgui.AlignTextToFramePadding()
		imgui.Text(showName)
		imgui.SameLine()
	}

	avail := imgui.ContentRegionAvail()
	cur := imgui.CursorPosX()
	var inputW float32
	switch {
	case width == -1:
		inputW = avail.X
	case width == -2:
		inputW = imgui.CalcTextSize(hint).X + paddingX*2 + 20
	default:
		inputW = startX + width - cur
	}
	if inputW < 40 {
		inputW = 40
	}

	imgui.SetNextItemWidth(inputW)

	// 从 store 取值，渲染后写回
	if !store.HasKey(key) {
		store.SetString(key, "")
	}
	val := store.GetString(key)
	if imgui.InputTextWithHint("##"+key, hint, &val, imgui.InputTextFlags(0), nil) {
		store.SetString(key, val)
	}
}

// UI_创建数字输入框 [-] [数字] [+] 布局
// 可选参数顺序: width, fontSize, textColor, step, min, max
//
//	width:  -1 填满 / -2 自适应 / >0 指定总宽度
//	fontSize: 0=默认
//	textColor: hex 如 "#e65100"
//	step: 每次加减步长, 默认 1
//	min/max: 限制范围, 不传则不限制
func UI_创建数字输入框(store *Store, key string, showName string, hint string, args ...interface{}) {

	const (
		bgColor     = "#ffffff14"
		borderColor = "#ffffff28"
		btnColor    = "#ffffff20"
		paddingX    = float32(8)
		paddingY    = float32(5)
	)

	var (
		width     float32 = -1
		fontSize  float32 = 0
		textColor string  = "#ffffff"
		step      float64 = 1
		minVal    float64 = -1e18
		maxVal    float64 = 1e18
	)

	idx := 0
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			width = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			fontSize = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(string); ok {
			textColor = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float64); ok {
			step = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float64); ok {
			minVal = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float64); ok {
			maxVal = v
		}
		idx++
	}

	// key 不存在时初始化默认值 0, 保证 ToJSON 一定包含该 key
	if !store.HasKey(key) {
		store.SetFloat(key, 0)
	}

	scale := float32(1.0)
	if fontSize > 0 {
		if base := imgui.CalcTextSize("A").Y; base > 0 {
			scale = fontSize / base
		}
	}
	imgui.SetWindowFontScale(scale)
	defer imgui.SetWindowFontScale(1.0)

	// 左侧标签
	startX := imgui.CursorPosX()
	if showName != "" {
		imgui.AlignTextToFramePadding()
		imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))
		imgui.Text(showName)
		imgui.PopStyleColorV(1)
		imgui.SameLine()
	}

	// 计算总可用宽度
	avail := imgui.ContentRegionAvail()
	cur := imgui.CursorPosX()
	var totalW float32
	switch {
	case width == -1:
		totalW = avail.X
	case width == -2:
		totalW = 160
	default:
		totalW = startX + width - cur
	}
	if totalW < 80 {
		totalW = 80
	}

	// 按钮宽 = 高 = paddingY*2 + 字号
	fontH := imgui.CalcTextSize("A").Y
	btnW := fontH + paddingY*2     // 加上横向 padding
	inputW := totalW - btnW*2 - 32 // 8 = 两个 SameLine 间距
	val := store.GetFloat(key)

	// 样式
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
	imgui.PushStyleColorVec4(imgui.ColFrameBg, HexToVec4(bgColor))
	imgui.PushStyleColorVec4(imgui.ColBorder, HexToVec4(borderColor))
	imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))
	imgui.PushStyleColorVec4(imgui.ColButton, HexToVec4(btnColor))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.18})
	imgui.PushStyleColorVec4(imgui.ColButtonActive, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.3})
	defer imgui.PopStyleColorV(6)
	defer imgui.PopStyleVarV(3)

	// [-] 按钮
	if imgui.ButtonV("-##sub_"+key, imgui.Vec2{X: btnW, Y: btnW}) {
		val = math.Max(minVal, val-step)
		store.SetFloat(key, val)
	}
	imgui.SameLine()

	// 中间输入框
	imgui.SetNextItemWidth(inputW)
	str := strconv.FormatFloat(val, 'f', -1, 64)
	if imgui.InputTextWithHint("##"+key, hint, &str, imgui.InputTextFlagsCharsDecimal, nil) {

		if n, err := strconv.ParseFloat(str, 64); err == nil {
			n = math.Max(minVal, math.Min(maxVal, n))
			store.SetFloat(key, n)
		}
	}
	imgui.SameLine()

	// [+] 按钮
	if imgui.ButtonV("+##add_"+key, imgui.Vec2{X: btnW, Y: btnW}) {
		val = math.Min(maxVal, val+step)
		store.SetFloat(key, val)
	}
}

// UI_创建多行输入框
// 可选参数顺序: width, height, fontSize, textColor
//
//	width:  -1 填满 / -2 自适应 / >0 指定宽度
//	height: >0 指定高度, 默认 80
//	fontSize: 0=默认
//	textColor: hex 如 "#e65100"
func UI_创建多行输入框(store *Store, key string, showName string, hint string, args ...interface{}) {
	const (
		bgColor     = "#ffffff14"
		hintColor   = "#ffffff55"
		borderColor = "#ffffff28"
		paddingX    = float32(8)
		paddingY    = float32(5)
	)

	var (
		width     float32 = -1
		height    float32 = 80
		fontSize  float32 = 0
		textColor string  = "#ffffff"
	)

	idx := 0
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			width = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			height = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			fontSize = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(string); ok {
			textColor = v
		}
		idx++
	}

	scale := float32(1.0)
	if fontSize > 0 {
		if base := imgui.CalcTextSize("A").Y; base > 0 {
			scale = fontSize / base
		}
	}
	imgui.SetWindowFontScale(scale)
	defer imgui.SetWindowFontScale(1.0)

	// 左侧标签(独占一行，输入框在下方)
	if showName != "" {
		imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))
		imgui.Text(showName)
		imgui.PopStyleColorV(1)
	}

	// 计算宽度
	avail := imgui.ContentRegionAvail()
	var inputW float32
	switch {
	case width == -1:
		inputW = avail.X
	case width == -2:
		inputW = imgui.CalcTextSize(hint).X + paddingX*2 + 20
	default:
		inputW = width
	}
	if inputW < 40 {
		inputW = 40
	}

	imgui.PushStyleColorVec4(imgui.ColFrameBg, HexToVec4(bgColor))
	imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))
	imgui.PushStyleColorVec4(imgui.ColTextDisabled, HexToVec4(hintColor))
	imgui.PushStyleColorVec4(imgui.ColBorder, HexToVec4(borderColor))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.1})
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.2})
	defer imgui.PopStyleColorV(6)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
	defer imgui.PopStyleVarV(3)

	// 第一次使用时初始化默认值, 保证 ToJSON 一定包含该 key
	if !store.HasKey(key) {
		store.SetString(key, hint)
	}

	val := store.GetString(key)
	if imgui.InputTextMultiline("##"+key, &val, imgui.Vec2{X: inputW, Y: height}, imgui.InputTextFlags(0), nil) {
		store.SetString(key, val)
	}
}

// UI_创建下拉框 文字在左，下拉框在右占满剩余宽度
// []string, arg
// 可选参数顺序: width, height, fontSize, textColor
//
//	width:  -1 填满剩余 / -2 自适应 / >0 指定宽度
//	height: -2 自适应 / >0 指定高度(通过 paddingY 控制)
//	fontSize: 0=默认
//	textColor: hex 如 "#e65100"
func UI_创建下拉框(store *Store, key string, showName string, options []string, args ...interface{}) {
	const (
		bgColor     = "#ffffff14"
		borderColor = "#ffffff28"
		paddingX    = float32(8)
		paddingY    = float32(3)
		maxVisible  = 5
		edgeMargin  = float32(6)
	)

	var (
		width     float32 = -1
		height    float32 = -2
		fontSize  float32 = 0
		textColor string  = "#ffffff"
	)

	idx := 0
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			width = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			height = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			fontSize = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(string); ok {
			textColor = v
		}
		idx++
	}

	if len(options) == 0 {
		return
	}

	// key 不存在时初始化默认值 0 (默认选中第一项), 保证 ToJSON 一定包含该 key
	if !store.HasKey(key) {
		store.SetFloat(key, 0)
	}

	scale := float32(1.0)
	if fontSize > 0 {
		if base := imgui.CalcTextSize("A").Y; base > 0 {
			scale = fontSize / base
		}
	}

	imgui.SetWindowFontScale(scale)
	defer imgui.SetWindowFontScale(1.0)

	py := paddingY
	if height > 0 {
		fontH := imgui.CalcTextSize("A").Y
		if p := (height - fontH) / 2; p > 0 {
			py = p
		} else {
			py = 0
		}
	}

	startX := imgui.CursorPosX()
	if showName != "" {
		imgui.AlignTextToFramePadding()
		imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))
		imgui.Text(showName)
		imgui.PopStyleColorV(1)
		imgui.SameLine()
	}

	avail := imgui.ContentRegionAvail()
	cur := imgui.CursorPosX()

	var comboW float32
	switch {
	case width == -1:
		comboW = avail.X
	case width == -2:
		var maxW float32
		for _, opt := range options {
			if w := imgui.CalcTextSize(opt).X; w > maxW {
				maxW = w
			}
		}
		comboW = maxW + paddingX*2 + 30
	default:
		comboW = startX + width - cur
	}

	if comboW < 60 {
		comboW = 60
	}

	selected := int32(store.GetFloat(key))
	if selected < 0 || int(selected) >= len(options) {
		selected = 0
	}

	imgui.PushStyleColorVec4(imgui.ColButton, HexToVec4(bgColor))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.15})
	imgui.PushStyleColorVec4(imgui.ColButtonActive, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.25})
	imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))
	imgui.PushStyleColorVec4(imgui.ColBorder, HexToVec4(borderColor))
	imgui.PushStyleColorVec4(imgui.ColPopupBg, imgui.Vec4{X: 0.1, Y: 0.1, Z: 0.12, W: 0.97})
	imgui.PushStyleColorVec4(imgui.ColHeaderHovered, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.2})
	imgui.PushStyleColorVec4(imgui.ColHeader, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.12})
	defer imgui.PopStyleColorV(8)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: py})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
	imgui.PushStyleVarFloat(imgui.StyleVarPopupRounding, 6)
	imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: 0, Y: 0.5}) // 文字靠左
	defer imgui.PopStyleVarV(5)                                                     // 原来是 4，这里改成 5

	buttonH := float32(0)
	if height > 0 {
		buttonH = height
	}

	arrowPad := "   " // 给箭头预留的视觉空间
	buttonID := options[selected] + arrowPad + "##DROPDOWN_BUTTON_" + key
	popupID := "DROPDOWN_POPUP_" + key

	clicked := imgui.ButtonV(buttonID, imgui.Vec2{X: comboW, Y: buttonH})

	itemMin := imgui.ItemRectMin()
	itemMax := imgui.ItemRectMax()
	itemSize := imgui.ItemRectSize()

	if clicked {
		imgui.OpenPopupStr(popupID)
	}

	itemH := imgui.FrameHeight()
	winPadY := imgui.CurrentStyle().WindowPadding().Y

	visibleCount := len(options)
	if visibleCount > maxVisible {
		visibleCount = maxVisible
	}

	popupH := itemH*float32(visibleCount) + winPadY*2

	_, screenHeight, displayScale, _ := device.GetDisplayInfo(0)

	screenH := float32(screenHeight)

	// 如果你打印发现 itemMax.Y 明显是逻辑坐标，而 screenHeight 是物理像素，

	if displayScale <= 0 {
		displayScale = 1
	}

	screenTop := edgeMargin
	screenBottom := screenH - edgeMargin

	spaceAbove := itemMin.Y - screenTop
	spaceBelow := screenBottom - itemMax.Y

	if spaceAbove < 0 {
		spaceAbove = 0
	}
	if spaceBelow < 0 {
		spaceBelow = 0
	}

	openUpward := spaceBelow < popupH && spaceAbove > spaceBelow

	// ===== 插入开始：绘制展开箭头标识 =====
	{
		dl := imgui.WindowDrawList()
		fontH := imgui.CalcTextSize("A").Y
		half := fontH * 0.28
		cx := itemMax.X - paddingX - half
		cy := (itemMin.Y + itemMax.Y) * 0.5
		col := imgui.ColorConvertFloat4ToU32(HexToVec4(textColor))

		if openUpward {
			// 向上的三角
			dl.AddTriangleFilled(
				imgui.Vec2{X: cx - half, Y: cy + half*0.5},
				imgui.Vec2{X: cx + half, Y: cy + half*0.5},
				imgui.Vec2{X: cx, Y: cy - half*0.5},
				col,
			)
		} else {
			// 向下的三角
			dl.AddTriangleFilled(
				imgui.Vec2{X: cx - half, Y: cy - half*0.5},
				imgui.Vec2{X: cx + half, Y: cy - half*0.5},
				imgui.Vec2{X: cx, Y: cy + half*0.5},
				col,
			)
		}
	}

	finalPopupH := popupH

	if openUpward {
		if finalPopupH > spaceAbove {
			finalPopupH = spaceAbove
		}
	} else {
		if finalPopupH > spaceBelow {
			finalPopupH = spaceBelow
		}
	}

	minPopupH := itemH + winPadY*2
	if finalPopupH < minPopupH {
		finalPopupH = minPopupH
	}

	popupPos := imgui.Vec2{
		X: itemMin.X,
		Y: itemMax.Y,
	}

	if openUpward {
		popupPos.Y = itemMin.Y - finalPopupH
	}

	imgui.SetNextWindowPosV(popupPos, imgui.CondAlways, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: itemSize.X, Y: finalPopupH}, imgui.CondAlways)

	if imgui.BeginPopup(popupID) {
		for i, opt := range options {
			isSelected := int32(i) == selected

			if imgui.SelectableBoolV(opt+"##"+key+strconv.Itoa(i), isSelected, 0, imgui.Vec2{X: itemSize.X, Y: 0}) {
				store.SetFloat(key, float64(i))
				imgui.CloseCurrentPopup()
			}

			if isSelected {
				imgui.SetItemDefaultFocus()
			}
		}

		imgui.EndPopup()
	}
}

// ==================== 创建图像 ====================
// 纹理缓存，避免每帧重复上传
type cachedTex struct {
	tex  *imgui.Texture
	w, h float32
}

var textureCache = map[string]*cachedTex{}

// width/height:
//
//	-3 等比缩放至可用区域内(推荐大图用)
//	-1 填满剩余(会拉伸)
//	-2 自适应原图/按比例
//	>0 指定值
func UI_创建图像(ID string, showName string, path string, callback func(), args ...interface{}) {
	var (
		width  float32 = -3 // 默认就用等比缩放，大图友好
		height float32 = -3
	)
	idx := 0
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			width = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			height = v
		}
		idx++
	}

	entry, ok := textureCache[path]
	if !ok {
		imgData := images.ReadFromPath(path)
		if imgData == nil {
			return
		}
		b := imgData.Bounds()
		entry = &cachedTex{
			tex: imgui.CreateTextureNrgba(imgData),
			w:   float32(b.Dx()),
			h:   float32(b.Dy()),
		}
		textureCache[path] = entry
	}

	avail := imgui.ContentRegionAvail()
	var drawW, drawH float32

	// 等比缩放模式：任一维度为 -3 就走整体等比
	if width == -3 || height == -3 {
		// 可用宽高，避免为 0
		availW := avail.X
		availH := avail.Y
		if availW < 1 {
			availW = entry.w
		}
		if availH < 1 {
			availH = entry.h
		}

		scaleW := availW / entry.w
		scaleH := availH / entry.h
		scale := scaleW
		if scaleH < scale {
			scale = scaleH
		}
		// 只缩不放大，避免小图被拉糊（按需删掉这两行）
		if scale > 1 {
			scale = 1
		}
		drawW = entry.w * scale
		drawH = entry.h * scale
	} else {
		switch {
		case width == -1:
			drawW = avail.X
		case width == -2:
			drawW = entry.w
		default:
			drawW = width
		}
		switch {
		case height == -1:
			drawH = avail.Y
		case height == -2:
			drawH = drawW * entry.h / entry.w
		default:
			drawH = height
		}
	}

	imgui.Image(entry.tex.ID, imgui.NewVec2(drawW, drawH))

	if callback != nil && imgui.IsItemClicked() {
		callback()
	}

	if showName != "" {
		dl := imgui.WindowDrawList()
		pos := imgui.ItemRectMin()
		textPos := imgui.Vec2{X: pos.X + 4, Y: pos.Y + 4}
		shadow := imgui.ColorConvertFloat4ToU32(imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0.6})
		white := imgui.ColorConvertFloat4ToU32(imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
		dl.AddTextVec2(imgui.Vec2{X: textPos.X + 1, Y: textPos.Y + 1}, shadow, showName)
		dl.AddTextVec2(textPos, white, showName)
	}
}

// 如果项目里没有这两个 map，加在包级别
var (
	uiCollapseOpenMap = map[string]bool{}
	uiCollapseInitMap = map[string]bool{}
)

// UI_创建折叠 绘制一个无指针折叠控件
// defaultOpen: 默认是否展开，仅第一次初始化时生效
// 可选参数顺序: width, height
func UI_创建折叠(ID string, title string, defaultOpen bool, content func(), args ...interface{}) bool {
	const (
		bgColor     = "#ffffff14"
		borderColor = "#ffffff28"
		textColor   = "#ffffff"
		paddingX    = float32(8)
		paddingY    = float32(6)
	)

	var (
		width  float32 = -1
		height float32 = -2
	)
	idx := 0
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			width = v
		}
		idx++
	}
	if idx < len(args) {
		if v, ok := args[idx].(float32); ok {
			height = v
		}
		idx++
	}

	// 初始化状态，仅第一次生效
	if !uiCollapseInitMap[ID] {
		uiCollapseOpenMap[ID] = defaultOpen
		uiCollapseInitMap[ID] = true
	}
	open := uiCollapseOpenMap[ID]

	avail := imgui.ContentRegionAvail()
	var btnW float32
	switch {
	case width == -1:
		btnW = avail.X
	case width == -2:
		btnW = 0
	default:
		btnW = width
	}
	var btnH float32
	if height > 0 {
		btnH = height
	}

	imgui.PushStyleColorVec4(imgui.ColButton, HexToVec4(bgColor))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.15})
	imgui.PushStyleColorVec4(imgui.ColButtonActive, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.25})
	imgui.PushStyleColorVec4(imgui.ColText, HexToVec4(textColor))
	imgui.PushStyleColorVec4(imgui.ColBorder, HexToVec4(borderColor))
	defer imgui.PopStyleColorV(5)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
	imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: 0, Y: 0.5})
	defer imgui.PopStyleVarV(4)
	arrow := "▼"
	if open {
		arrow = "▲"
	}
	clicked := imgui.ButtonV(title+arrow+"##COLLAPSE_"+ID, imgui.Vec2{X: btnW, Y: btnH})

	{
		dl := imgui.WindowDrawList()
		itemMin := imgui.ItemRectMin()
		itemMax := imgui.ItemRectMax()
		fontH := imgui.CalcTextSize("A").Y
		half := fontH * 0.28
		cx := itemMax.X - paddingX - half
		cy := (itemMin.Y + itemMax.Y) * 0.5
		col := imgui.ColorConvertFloat4ToU32(HexToVec4(textColor))
		if open {
			dl.AddTriangleFilled(
				imgui.Vec2{X: cx - half, Y: cy + half*0.5},
				imgui.Vec2{X: cx + half, Y: cy + half*0.5},
				imgui.Vec2{X: cx, Y: cy - half*0.5},
				col,
			)
		} else {
			dl.AddTriangleFilled(
				imgui.Vec2{X: cx - half, Y: cy - half*0.5},
				imgui.Vec2{X: cx + half, Y: cy - half*0.5},
				imgui.Vec2{X: cx, Y: cy + half*0.5},
				col,
			)
		}
	}

	if clicked {
		open = !open
		uiCollapseOpenMap[ID] = open
	}

	if open && content != nil {
		imgui.Spacing()
		content()
		imgui.Spacing()
	}

	return open
}
