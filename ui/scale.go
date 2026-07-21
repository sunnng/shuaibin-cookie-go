package ui

// 基准分辨率（CONTEXT.md）：UI 布局常量与游戏特征常量同写于 1600×900。
const (
	DefaultBaseWidth  = 1600
	DefaultBaseHeight = 900
)

// ComputeScale 计算 UI 布局缩放系数：宽度驱动 displayW/baseW，
// 启动时算一次，帧内经 Ctx.S 统一换算。无效输入回退 1。
// displayH/baseH 为策略调整预留，当前不参与计算。
func ComputeScale(displayW, displayH, baseW, baseH int) float64 {
	if displayW <= 0 || baseW <= 0 {
		return 1
	}
	return float64(displayW) / float64(baseW)
}
