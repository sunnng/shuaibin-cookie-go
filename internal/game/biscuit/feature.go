package biscuit

import (
	"app/internal/platform/screen"
)

// Feature 洗脆饼词条任务的界面特征。坐标与颜色串均为 1600×900 基准，
// 原样搬自 Lua game/功能_洗脆饼/task.lua。
type Feature struct {
	Reads   ReadsFeature
	Actions ActionsFeature
	Dialogs DialogsFeature
}

type ReadsFeature struct {
	Effects screen.Region // 副词条 OCR 区域（脆饼固定 4 条）
}

type ActionsFeature struct {
	Reroll screen.Region // 洗炼按钮；区域内随机点
}

type DialogsFeature struct {
	ResetConfirm DialogDef // 确认重置弹窗
	SameConfirm  DialogDef // 确认相同脆饼弹窗
}

type DialogDef struct {
	Identify      screen.Feature
	DontShowAgain screen.Region // 今日不再显示；区域内随机点
	Confirm       screen.Region // 确认按钮；区域内随机点
}

func DefaultFeature() *Feature {
	return &Feature{
		Reads: ReadsFeature{
			Effects: screen.Region{X1: 427, Y1: 390, X2: 1162, Y2: 760},
		},
		Actions: ActionsFeature{
			Reroll: screen.Region{X1: 914, Y1: 815, X2: 961, Y2: 851},
		},
		Dialogs: DialogsFeature{
			ResetConfirm: DialogDef{
				Identify: screen.Feature{
					Colors: "1026|627|7ace0e-101010,745|629|0ca6df-101010,863|257|363d5f-101010,782|466|505050-101010,785|419|505050-101010",
					Sim:    0.9,
				},
				DontShowAgain: screen.Region{X1: 874, Y1: 727, X2: 887, Y2: 740},
				Confirm:       screen.Region{X1: 932, Y1: 624, X2: 977, Y2: 643},
			},
			SameConfirm: DialogDef{
				Identify: screen.Feature{
					Colors: "1041|635|7ace0e-101010,711|632|0ca6df-101010,815|263|f70b05-101010,972|257|363d5f-101010,802|248|ffffff-101010,836|440|505050-101010",
					Sim:    0.9,
				},
				DontShowAgain: screen.Region{X1: 876, Y1: 725, X2: 885, Y2: 739},
				Confirm:       screen.Region{X1: 942, Y1: 626, X2: 971, Y2: 641},
			},
		},
	}
}
