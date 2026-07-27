package ui

// 灵动岛位置持久化键（存 ui.json，由 android 绘制层读写）。
const (
	islandPosXKey = "island_pos_x"
	islandPosYKey = "island_pos_y"
)

// clampIslandPos 把灵动岛窗口位置夹取到屏幕内，保证窗口完整可见；
// 窗口比屏幕还大时贴左上角。
func clampIslandPos(x, y, w, h, screenW, screenH float32) (float32, float32) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+w > screenW {
		x = screenW - w
	}
	if y+h > screenH {
		y = screenH - h
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y
}
