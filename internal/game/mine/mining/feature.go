// Package mining 矿山开采子任务（Lua 模块_矿山开采）。
// 流程：导航进矿山首页 → 预检 → 开采页扫描（收奖励/填空闲栏位/启动开采）→
// 选矿卡（按优先级扫卡列表）→ 自动选饼干开始 → 全忙回城记录 busy 等待。
package mining

import (
	"app/internal/game/mine"
	"app/internal/platform/screen"
)

// 矿卡键名（同 Lua 矿山_特征库.oreVeinCards）。
const (
	CardFlourStone    = "flourStone"    // 白
	CardSugarOre      = "sugarOre"      // 蓝
	CardButterAmber   = "butterAmber"   // 黄
	CardAmberFossil   = "amberFossil"   // 黄
	CardPurpleFossil  = "purpleFossil"  // 紫
	CardEmeraldFossil = "emeraldFossil" // 绿
)

// DefaultCardPriority 默认选卡优先级（Lua DEFAULT_CARD_PRIORITY）。
var DefaultCardPriority = []string{
	CardButterAmber,
	CardAmberFossil,
	CardSugarOre,
	CardPurpleFossil,
	CardEmeraldFossil,
	CardFlourStone,
}

// Feature 开采相关识别与动作（对应 Lua 矿山_特征库.mining + oreVeinCards）。
type Feature struct {
	Page                screen.Feature // 开采首页识别
	NoMineCardOCR       screen.Region  // 「没有可选择的矿脉卡」提示 OCR 区域
	FreeLocation        mine.ColorFind // 空闲栏位
	FreePlus            mine.ColorFind // 空闲栏位（+ 样式）
	CanChooseNum        screen.Region  // 可选数量 OCR（cur/max）
	MultiSelectBtn      screen.Region  // 「选择多个」按钮
	MultiSelectOCR      screen.Region  // 「选择多个」OCR 区域
	CardListStartOCR    screen.Region  // 卡列表左缘 OCR（判是否到尽头）
	CardListEndOCR      screen.Region  // 卡列表右缘 OCR
	CompletedTask       mine.ColorFind // 已完成开采任务槽位
	RewardPage          RewardPageFeature
	CardSelect          CardSelectFeature
	StartableCard       mine.ColorFind // 可开始开采的矿卡
	SetupIdentify       screen.Feature // 准备开采页（选饼干）
	SetupReadyIdentify  screen.Feature // 准备就绪（可点开始）
	Dialogs             DialogsFeature
	AutoSelectCookieBtn screen.Region
	ConfirmStartBtn     screen.Region
	BackBtn             screen.Region
	OreVeinCards        map[string]mine.ColorFind // 矿脉卡特征，键名同 Card* 常量
}

// RewardPageFeature 获得开采奖励页。
type RewardPageFeature struct {
	TitleText  string        // OCR 标题文本
	TitleOCR   screen.Region // 标题 OCR 区域
	ConfirmBtn screen.Region
}

// CardSelectFeature 矿脉卡选择页。
type CardSelectFeature struct {
	ConfirmBtn   screen.Region
	BackBtn      screen.Region
	SelectedMark mine.ColorFind // 已选矿卡角标（Lua 为 nil：未取色，Colors 留空）
}

// DialogsFeature 开始开采后顺序未知的两个饼干弹窗（流程内联处理，非 Guard trap）。
type DialogsFeature struct {
	ConfirmCookie      DialogDef // 确认饼干弹窗
	CookieCountWarning DialogDef // 饼干数量不足警告弹窗
}

type DialogDef struct {
	Identify         screen.Feature
	Confirm          screen.Region
	TodayNotAskAgain screen.Region // 今日不再询问
}

func DefaultFeature() *Feature {
	return &Feature{
		Page: screen.Feature{
			Colors: "219,48,ffffff-101010,176,57,ffffff-101010,903,65,a8623f-101010,1318,64,09c4ff-101010,1316,52,0589ff-101010,1547,68,36a6e6-101010,417,891,4b1d00-101010,1571,882,2c0d00-101010",
			Sim:    0.9,
		},
		NoMineCardOCR: screen.Region{X1: 662, Y1: 503, X2: 915, Y2: 542},
		FreeLocation: mine.ColorFind{
			Region: screen.Region{X1: 27, Y1: 165, X2: 1580, Y2: 359},
			Colors: "c67654-101010,-4,-26,f5ece4-101010,103,46,2f1e1b-101010,47,106,2f1e1b-101010,-50,110,2f1e1b-101010,-95,-58,392520-101010,90,-57,392520-101010,-68,120,804a40-101010",
			Sim:    0.9,
		},
		FreePlus: mine.ColorFind{
			Region: screen.Region{X1: 1320, Y1: 60, X2: 1575, Y2: 220},
			Colors: "f5c079-101010,-31,-12,f9bd76-101010,-3,-44,f7b873-101010,28,-10,f3aa69-101010,-1,33,f1a365-101010,51,-51,2e1d1d-101010,-60,54,2d1b1b-101010",
			Sim:    0.9,
		},
		CanChooseNum:     screen.Region{X1: 769, Y1: 740, X2: 841, Y2: 785},
		MultiSelectBtn:   screen.Region{X1: 129, Y1: 828, X2: 265, Y2: 875},
		MultiSelectOCR:   screen.Region{X1: 129, Y1: 828, X2: 265, Y2: 875},
		CardListStartOCR: screen.Region{X1: 95, Y1: 452, X2: 390, Y2: 532},
		CardListEndOCR:   screen.Region{X1: 1210, Y1: 452, X2: 1505, Y2: 532},
		CompletedTask: mine.ColorFind{
			Region: screen.Region{X1: 34, Y1: 101, X2: 1570, Y2: 251},
			Colors: "9bd400-101010,-10,0,97d501-101010,-17,4,befd00-101010,-7,5,97d400-101010,-69,77,333333-101010,-76,68,d7e1f0-101010,-75,79,ffffff-101010,-63,87,4169ac-101010,-24,79,90f90a-101010,36,73,c9ff3a-101010",
			Sim:    0.9,
		},
		RewardPage: RewardPageFeature{
			TitleText:  "获得开采奖励",
			TitleOCR:   screen.Region{X1: 284, Y1: 204, X2: 891, Y2: 312},
			ConfirmBtn: screen.Region{X1: 678, Y1: 762, X2: 926, Y2: 799},
		},
		CardSelect: CardSelectFeature{
			ConfirmBtn: screen.Region{X1: 955, Y1: 742, X2: 1015, Y2: 769},
			BackBtn:    screen.Region{X1: 1516, Y1: 14, X2: 1584, Y2: 77},
		},
		StartableCard: mine.ColorFind{
			Region: screen.Region{X1: 15, Y1: 93, X2: 1588, Y2: 182},
			Colors: "ffffff-101010,-11,-14,ffffff-101010,12,-15,fd7430-101010,12,15,ef1909-101010,1,20,ed1608-101010,-16,13,ef1808-101010,-17,15,230a0b-101010,14,-16,000000-101010",
			Sim:    0.9,
		},
		SetupIdentify: screen.Feature{
			Colors: "1487,827,a0a0a0-101010,1244,806,ffffff-101010,1277,570,ffffff-101010,1525,431,ffd200-101010,1516,499,0ca6df-101010",
			Sim:    0.9,
		},
		SetupReadyIdentify: screen.Feature{
			Colors: "1411,824,ffffff-101010,1388,808,7ace0e-101010,1247,803,ffffff-101010",
			Sim:    0.9,
		},
		Dialogs: DialogsFeature{
			ConfirmCookie: DialogDef{
				Identify: screen.Feature{
					Colors: "468,223,6a719e-101010,1126,223,363d5f-101010,1131,683,afa09c-101010,474,664,dbcfc6-101010,571,638,0ca6df-101010,1069,644,7ace0e-101010,814,419,505050-101010,794,498,8c8c8c-101010,793,360,505050-101010",
					Sim:    0.9,
				},
				Confirm:          screen.Region{X1: 932, Y1: 619, X2: 972, Y2: 642},
				TodayNotAskAgain: screen.Region{X1: 871, Y1: 724, X2: 887, Y2: 740},
			},
			CookieCountWarning: DialogDef{
				Identify: screen.Feature{
					Colors: "682,414,505050-101010,897,425,505050-101010,864,471,505050-101010,706,470,505050-101010,512,631,3db8e5-101010,1093,643,7ace0e-101010,1129,227,363d5f-101010,472,222,6a719e-101010",
					Sim:    0.9,
				},
				Confirm:          screen.Region{X1: 930, Y1: 623, X2: 959, Y2: 646},
				TodayNotAskAgain: screen.Region{X1: 878, Y1: 728, X2: 886, Y2: 740},
			},
		},
		AutoSelectCookieBtn: screen.Region{X1: 1424, Y1: 412, X2: 1497, Y2: 432},
		ConfirmStartBtn:     screen.Region{X1: 1363, Y1: 816, X2: 1451, Y2: 835},
		BackBtn:             screen.Region{X1: 1516, Y1: 14, X2: 1584, Y2: 77},
		OreVeinCards: map[string]mine.ColorFind{
			CardFlourStone: {
				Region: screen.Region{X1: 3, Y1: 602, X2: 1587, Y2: 707},
				Colors: "dfc3a6-101010,-10,-11,dfc4a8-101010,-15,-15,502828-101010,13,13,3d1c1b-101010,-17,6,c9ae8f-101010,-22,13,4f2828-101010",
				Sim:    0.9,
			},
			CardSugarOre: {
				Region: screen.Region{X1: 3, Y1: 602, X2: 1587, Y2: 707},
				Colors: "8d48e7-101010,0,3,545883-101010,5,-11,aea2fb-101010,-4,-21,fafdfd-101010,5,-23,2085f7-101010,19,-11,46b5fe-101010,-12,2,a4dffd-101010",
				Sim:    0.9,
			},
			CardButterAmber: {
				Region: screen.Region{X1: 15, Y1: 542, X2: 1599, Y2: 656},
				Colors: "f9fec5-101010,2,-16,e7c028-101010,-18,2,dc6712-101010,17,2,cc4f0b-101010,-10,14,f1fb92-101010,12,14,5f1310-101010,-10,-11,e6fee8-101010",
				Sim:    0.9,
			},
			CardAmberFossil: {
				Region: screen.Region{X1: 3, Y1: 602, X2: 1587, Y2: 707},
				Colors: "feb96b-101010,4,-8,a64000-101010,1,10,c49a67-101010,12,-11,76000d-101010,-12,-12,efdea8-101010,17,9,835437-101010,-23,11,5a2614-101010",
				Sim:    0.9,
			},
			CardPurpleFossil: {
				Region: screen.Region{X1: 3, Y1: 602, X2: 1587, Y2: 707},
				Colors: "ea9aff-101010,4,-8,8000bc-101010,9,-13,a3069e-101010,2,10,6f5ca6-101010,-11,13,bfaee9-101010,16,10,4f3e76-101010,-22,-2,372d53-101010",
				Sim:    0.9,
			},
			CardEmeraldFossil: {
				Region: screen.Region{X1: 3, Y1: 602, X2: 1587, Y2: 707},
				Colors: "9ff9c5-101010,4,-8,1c9145-101010,5,-17,16572f-101010,-18,4,91a98d-101010,0,9,89a188-101010,17,9,536c5b-101010,12,17,07220d-101010",
				Sim:    0.9,
			},
		},
	}
}
