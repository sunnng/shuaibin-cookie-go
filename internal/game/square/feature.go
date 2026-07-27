package square

import (
	"app/internal/platform/screen"
)

// Feature 布谷鸟广场全部 UI 常量（1600×900 基准）。
//
// 颜色串/坐标原样搬自 Lua 广场_特征库.lua；EntryBtn 搬自 Lua 王国特征库
// （通用_王国/特征库.lua 的 squareBtn）。注意分隔符差异：Lua cmpColorExT
// 用 "x|y|color" 分段，Go images.DetectsMultiColors 用 "x,y,color"，已逐点
// 机械转换，坐标与颜色值未动。
type Feature struct {
	Home   HomeFeature
	Dialog DialogFeature
	// EntryBtn 王国首页 → 布谷鸟广场 的入口按钮（Lua kingdom home.squareBtn）。
	// Go 的 kingdom.Feature 暂未携带广场入口，且本任务不改 common 包，暂存于此。
	EntryBtn screen.Region
}

type HomeFeature struct {
	Identify screen.Feature
	Actions  HomeActions
}

type HomeActions struct {
	BackBtn screen.Region // 广场页返回键；点击后弹出「离开广场」弹窗
}

type DialogFeature struct {
	Identify screen.Feature
	Actions  DialogActions
	Reads    DialogReads
}

type DialogActions struct {
	CancelBtn        screen.Region // 关闭弹窗（右上 X），留在广场继续挂机
	LeaveBtn         screen.Region // 确认离开 → 回王国主城
	ConfirmRewardBtn screen.Region // 一次领回奖励
	TapUntilRect     screen.Region // 领奖后反复点此区域，直至弹窗重新出现（Lua tapUtilDialog）
}

type DialogReads struct {
	RewardNow   screen.Region // 目前可获得奖励 OCR
	RewardTotal screen.Region // 累计获得奖励 OCR
	DailyMax    screen.Region // 每日上限提示 OCR（IsFinish 未配置时兜底）
	IsFinish    screen.Region // 满额判定 OCR（"最大"/"已领取…奖励"）
}

func DefaultFeature() *Feature {
	return &Feature{
		Home: HomeFeature{
			Identify: screen.Feature{
				Colors: "1531,211,ffe314-101010,1533,68,36a3e3-101010,86,541,f9c16a-101010,59,298,fbe7ab-101010,59,96,ef345c-101010,67,110,95eb0e-101010",
				Sim:    0.9,
			},
			Actions: HomeActions{
				BackBtn: screen.Region{X1: 1530, Y1: 39, X2: 1543, Y2: 58},
			},
		},
		Dialog: DialogFeature{
			Identify: screen.Feature{
				Colors: "581,814,0ca5db-101010,407,819,87433b-101010,393,68,dd9387-101010,449,386,f5f3e3-101010,483,414,cd294e-101010,517,438,adf308-101010",
				Sim:    0.9,
			},
			Actions: DialogActions{
				CancelBtn:        screen.Region{X1: 1214, Y1: 52, X2: 1231, Y2: 70},
				LeaveBtn:         screen.Region{X1: 619, Y1: 790, X2: 669, Y2: 817},
				ConfirmRewardBtn: screen.Region{X1: 930, Y1: 800, X2: 983, Y2: 817},
				TapUntilRect:     screen.Region{X1: 722, Y1: 686, X2: 886, Y2: 725},
			},
			Reads: DialogReads{
				RewardNow:   screen.Region{X1: 794, Y1: 342, X2: 860, Y2: 374},
				RewardTotal: screen.Region{X1: 748, Y1: 382, X2: 820, Y2: 414},
				DailyMax:    screen.Region{X1: 720, Y1: 406, X2: 879, Y2: 483},
				IsFinish:    screen.Region{X1: 754, Y1: 431, X2: 848, Y2: 462},
			},
		},
		EntryBtn: screen.Region{X1: 589, Y1: 811, X2: 616, Y2: 830},
	}
}
