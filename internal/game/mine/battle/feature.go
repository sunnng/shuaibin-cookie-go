// Package battle 矿山战斗子任务（Lua 模块_矿山战斗）。
// 流程：导航进战斗页 → 扫描快转按钮/战斗卡 → 灵魂石匹配则快转 → 结算 →
// 翻页至末页 → 回城。每次运行记录战斗时间控制检测频率。
package battle

import (
	"app/internal/game/mine"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// 灵魂石品类（Lua SOUL_STONE_CATEGORIES 顺序）。
const (
	CategoryEpic    = "史诗"
	CategoryLegend  = "传奇"
	CategoryAncient = "上古"
	CategoryBeast   = "野兽"
)

// Feature 矿山战斗页识别与动作（对应 Lua 矿山_特征库.battle）。
type Feature struct {
	Identify          screen.Feature
	BackBtn           screen.Region
	QuickBattleBtn    mine.ColorFind // 快转按钮
	QuickBattleDialog QuickBattleDialogFeature
	SettleBtn         screen.Region // 快转后结算页点击完成区域
	SoulStoneRegion   screen.Region // 灵魂石识别区域（参考；各灵魂石特征自带区域）
	SoulStones        []SoulStoneCategory
	BattleCard        mine.ColorFind // 战斗卡（findAll）
	PageSwipe         action.Swipe   // 翻页滑动（向上）
	LastPage          mine.ColorFind // 末页特征
}

// QuickBattleDialogFeature 快转弹窗。
type QuickBattleDialogFeature struct {
	Identify      screen.Feature
	ConfirmBtn    screen.Region
	CancelBtn     screen.Region
	ClockCountOCR screen.Region // 快转发条数量 OCR（使用/持有，如 1/12,611）
}

// SoulStoneCategory 一类灵魂石的特征表，键为灵魂石名。
type SoulStoneCategory struct {
	Name   string
	Stones map[string]mine.ColorFind
}

func soulStone(colors string) mine.ColorFind {
	return mine.ColorFind{
		Region: screen.Region{X1: 271, Y1: 618, X2: 346, Y2: 694},
		Colors: colors,
		Sim:    0.9,
	}
}

func DefaultFeature() *Feature {
	return &Feature{
		Identify: screen.Feature{
			Colors: "161,84,df958b-101010,730,93,ffffff-101010,821,120,ffffff-101010,1417,78,87433b-101010,1429,799,87433b-101010,102,814,7e763a-101010",
			Sim:    0.9,
		},
		BackBtn: screen.Region{X1: 1411, Y1: 103, X2: 1422, Y2: 114},
		QuickBattleBtn: mine.ColorFind{
			Region: screen.Region{X1: 543, Y1: 735, X2: 634, Y2: 819},
			Colors: "d3af16-101010,-14,-9,18070a-101010,18,-13,050401-101010,7,27,ffffff-101010,-18,39,463b00-101010,15,-28,493900-101010,-26,-11,ffcf00-101010",
			Sim:    0.9,
		},
		QuickBattleDialog: QuickBattleDialogFeature{
			Identify: screen.Feature{
				Colors: "800,177,d1af19-101010,1111,173,349fdf-101010,523,180,696f9b-101010,876,688,7acd0e-101010,594,768,685408-101010",
				Sim:    0.9,
			},
			ConfirmBtn:    screen.Region{X1: 785, Y1: 667, X2: 818, Y2: 687},
			CancelBtn:     screen.Region{X1: 1081, Y1: 167, X2: 1096, Y2: 183},
			ClockCountOCR: screen.Region{X1: 756, Y1: 408, X2: 973, Y2: 451},
		},
		SettleBtn:       screen.Region{X1: 745, Y1: 821, X2: 858, Y2: 876},
		SoulStoneRegion: screen.Region{X1: 271, Y1: 617, X2: 348, Y2: 694},
		SoulStones: []SoulStoneCategory{
			{Name: CategoryEpic, Stones: map[string]mine.ColorFind{
				"浓缩奶油": soulStone("bba983-101010,3,-9,9b7370-101010,5,-15,adc1cb-101010,-13,-9,9d59c7-101010,16,12,f9d5a7-101010,-8,11,d789c3-101010,-8,5,dfb187-101010"),
				"牡蛎":   soulStone("cdb9cb-101010,-8,-8,9ba9d3-101010,11,-12,787fab-101010,-12,-11,855da7-101010,-5,12,eda3b7-101010,17,9,f9dfa7-101010"),
				"雪酪":   soulStone("cdebfb-101010,3,-14,afc7e7-101010,-13,-14,8564a3-101010,17,0,fbeff3-101010,14,9,fbd7a7-101010,10,-11,cfe3fb-101010,-10,8,d787c3-101010"),
				"辣椒素":  soulStone("f9c36a-101010,-9,-4,46527c-101010,8,-7,df6a46-101010,12,-6,64565c-101010,8,10,ffe793-101010,-11,10,f77f7a-101010,-15,-3,7751a7-101010"),
				"闪耀之星": soulStone("9dd7f3-101010,-14,-5,af6bdf-101010,-18,-6,8b58b3-101010,-15,10,c9d5fb-101010,10,7,fdeff3-101010,9,-4,e37faf-101010"),
				"绯红珊瑚": soulStone("e97d8f-101010,4,-8,7f497e-101010,16,-10,ed78b3-101010,-10,-10,8b5eaf-101010,12,3,e18593-101010,18,8,f9d7a7-101010,-8,10,dd8bc7-101010"),
				"妖精王":  soulStone("f3e5fb-101010,-10,-13,a9c3e3-101010,-15,-3,7677b3-101010,-16,8,f3e7fb-101010,5,-12,89a7bb-101010,6,5,fbefff-101010"),
				"星辰":   soulStone("997470-101010,-16,-8,737ddf-101010,7,-6,9587d7-101010,-4,-3,d5e1ef-101010,-11,7,fbf1fb-101010,1,11,d5b1d3-101010,-16,11,d585c7-101010,12,-10,ab6197-101010"),
			}},
			{Name: CategoryLegend, Stones: map[string]mine.ColorFind{
				"雷神武将": soulStone("fddfbb-101010,3,-9,b1c5af-101010,9,-14,496283-101010,-4,-17,6e9bb3-101010,-12,-12,59a3e3-101010,15,10,ebe3f7-101010,-4,13,9b7fd3-101010"),
				"冰霜女王": soulStone("8799e3-101010,-6,1,8bafe7-101010,-13,-5,b1c9e7-101010,11,-5,83838f-101010,-14,14,c3cbef-101010,10,15,b9bffb-101010"),
				"海妖精":  soulStone("a1dbeb-101010,7,0,afddf3-101010,11,-11,d1ddcf-101010,1,-11,4c93d7-101010,-14,-3,719dd3-101010,17,9,b1bdf3-101010,3,-23,7fa9e7-101010"),
				"风箭手":  soulStone("cfe1c7-101010,-3,-8,7aa160-101010,-16,1,7da164-101010,9,-1,d3d3ab-101010,-16,15,c3c7ef-101010,-5,-17,89aff3-101010"),
			}},
			{Name: CategoryAncient, Stones: map[string]mine.ColorFind{
				"莓果": soulStone("c5568f-101010,12,-9,41897a-101010,-4,-8,4f8f9f-101010,-4,10,ef6ffb-101010,-2,14,9b74d7-101010,16,12,c7a3df-101010"),
			}},
			{Name: CategoryBeast, Stones: map[string]mine.ColorFind{}},
		},
		BattleCard: mine.ColorFind{
			Region: screen.Region{X1: 138, Y1: 61, X2: 1463, Y2: 827},
			Colors: "333333-101010,0,-14,6685bb-101010,-16,0,223858-101010,13,1,4269ab-101010,-1,9,999793-101010,-9,1,999997-101010,7,4,ebebe3-101010",
			Sim:    0.9,
		},
		PageSwipe: action.Swipe{
			From:       action.Point{X: 588, Y: 646},
			To:         action.Point{X: 587, Y: 150},
			DurationMs: 500,
		},
		LastPage: mine.ColorFind{
			Region: screen.Region{X1: 599, Y1: 549, X2: 1437, Y2: 715},
			Colors: "552e2b-101010,767,8,552e2b-101010,774,102,552e2b-101010,412,36,552e2b-101010,220,31,552e2b-101010,707,35,552e2b-101010,71,81,552e2b-101010,188,36,552e2b-101010,21,91,552e2b-101010",
			Sim:    0.9,
		},
	}
}
