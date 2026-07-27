// Package jelly 解除洋菜冻子任务（Lua 模块_解除洋菜冻）。
// 流程：导航进解除洋菜冻页 → 可全部领取则领取结算 → OCR 找「配置」按钮 →
// 配置洋菜冻（可选择则选择）→ 无配置则 OCR 剩余时间 → 回城并按剩余时间/冷却间隔等待。
// Lua 洋菜冻_坐标库.lua 为空文件，无额外坐标需迁移。
package jelly

import (
	"app/internal/platform/screen"
)

// Feature 解除洋菜冻页与配置页识别（对应 Lua 矿山_特征库.jelly）。
type Feature struct {
	Identify         screen.Feature
	BackBtn          screen.Region
	ClaimAllIdentify screen.Feature // 可全部领取
	ClaimAllBtn      screen.Region  // 全部领取按钮
	SettleBtn        screen.Region  // 领取后结算点击区域
	OCRRegion        screen.Region  // 「配置」按钮与剩余时间 OCR 区域
	Config           ConfigFeature  // 配置洋菜冻界面
}

// ConfigFeature 配置洋菜冻界面。
type ConfigFeature struct {
	Identify          screen.Feature
	BackBtn           screen.Region
	ChooseBtn         screen.Region  // 选择按钮
	CanChooseIdentify screen.Feature // 可选择
}

func DefaultFeature() *Feature {
	return &Feature{
		Identify: screen.Feature{
			Colors: "246,106,df958b-101010,254,789,87433b-101010,717,114,ffffff-101010,766,145,ffffff-101010,751,148,190c0b-101010,1348,802,622620-101010",
			Sim:    0.9,
		},
		BackBtn: screen.Region{X1: 1330, Y1: 125, X2: 1338, Y2: 141},
		ClaimAllIdentify: screen.Feature{
			Colors: "1298,759,7acd10-101010",
			Sim:    0.9,
		},
		ClaimAllBtn: screen.Region{X1: 1179, Y1: 733, X2: 1219, Y2: 763},
		SettleBtn:   screen.Region{X1: 699, Y1: 758, X2: 907, Y2: 806},
		OCRRegion:   screen.Region{X1: 274, Y1: 586, X2: 1328, Y2: 669},
		Config: ConfigFeature{
			Identify: screen.Feature{
				Colors: "712,158,ffffff-101010,885,185,ffffff-101010,717,116,7f7f7e-101010,488,149,df958b-101010,243,107,6f4a44-101010,473,723,87433b-101010,243,745,43211d-101010",
				Sim:    0.9,
			},
			BackBtn:   screen.Region{X1: 1083, Y1: 170, X2: 1098, Y2: 184},
			ChooseBtn: screen.Region{X1: 973, Y1: 700, X2: 995, Y2: 719},
			CanChooseIdentify: screen.Feature{
				Colors: "1052,731,7acd0e-101010,902,697,93d73e-101010",
				Sim:    0.9,
			},
		},
	}
}
