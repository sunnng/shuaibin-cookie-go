//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/images"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// ImageProps 图像组件属性，定义在本文件（android），因组件依赖纹理。
//
// 尺寸语义：
//   - Width > 0 且 Height > 0：按基准分辨率尺寸经 ctx.S 缩放后的指定值绘制。
//   - Width <= 0 或 Height <= 0：等比缩放至当前可用区域，只缩不放大。
type ImageProps struct {
	Path          string
	Width, Height float64
	OnClick       func()
}

type cachedTex struct {
	tex  *imgui.Texture
	w, h float32
}

// Image 图像组件（ADR-0003）。纹理按 Path 全局缓存一次（经 ctx.resource），
// 加载失败时本帧不绘制且后续不再重试。支持点击回调。
//
// 尺寸语义见 ImageProps。等比缩放模式将图片按可用区域的最小边等比缩放，
// 不会放大超过原图尺寸。
func Image(ctx *Ctx, p ImageProps) {
	entryRaw := ctx.resource("tex:"+p.Path, func() any {
		imgData := images.ReadFromPath(p.Path)
		if imgData == nil {
			return nil
		}
		b := imgData.Bounds()
		return &cachedTex{
			tex: imgui.CreateTextureNrgba(imgData),
			w:   float32(b.Dx()),
			h:   float32(b.Dy()),
		}
	})
	if entryRaw == nil {
		return
	}
	entry := entryRaw.(*cachedTex)

	var drawW, drawH float32
	if p.Width > 0 && p.Height > 0 {
		drawW = float32(ctx.S(p.Width))
		drawH = float32(ctx.S(p.Height))
	} else {
		avail := imgui.ContentRegionAvail()
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
		if scale > 1 {
			scale = 1
		}
		drawW = entry.w * scale
		drawH = entry.h * scale
	}

	imgui.Image(entry.tex.ID, imgui.NewVec2(drawW, drawH))

	if p.OnClick != nil && imgui.IsItemClicked() {
		p.OnClick()
	}
}

// EnableSlidingScroll 启用当前窗口的触屏滑动滚动。
// speed 为滑动速度系数；<=0 时按默认值 1.2 处理。
func EnableSlidingScroll(speed float32) {
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

// Column 纵向作用域：为内容建立组件状态命名空间（Push/Pop），
// 布局本身是 imgui 默认纵向流。
func Column(ctx *Ctx, id string, content func()) {
	ctx.Push(id)
	defer ctx.Pop()
	if content != nil {
		content()
	}
}

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
