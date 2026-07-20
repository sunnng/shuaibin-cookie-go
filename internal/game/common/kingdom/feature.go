package kingdom

import (
	"app/internal/platform/screen"
)

// Feature 描述王国内互斥页面状态 + 首页上的导航点击。
//
// 页面同级（同一时刻通常只命中一个识别）：
//
//	Home      — 王国首页（多点比色）
//	Adventure — 冒险页（OCR：区域内出现「冒险」；竞技场入口在此页）
//	Event     — 活动页（OCR：区域内出现「王国活动」）
//
// 按钮类字段一律用 Region{x1,y1,x2,y2}，点击时在区域内随机取点。
type Feature struct {
	Home      PageSlot
	Adventure OCRPage
	Event     OCRPage
	Actions   NavActions
}

// PageSlot 某一王国页的身份识别（多点比色）。
type PageSlot struct {
	Identify screen.Feature // Colors + Sim → MatchMultiColor；Colors=="" 视为未配置
}

// OCRPage 用 OCR 判定是否在某页：Region 内出现 Keyword 即命中。
type OCRPage struct {
	Region  screen.Region
	Keyword string // 空则对应 Is* 视为未配置
}

// NavActions 首页入口与回首页；值为可点矩形，脚本随机点区域内一点。
type NavActions struct {
	AdventureBtn screen.Region // 首页 → 冒险
	EventBtn     screen.Region // 首页 → 活动
	BackHome     screen.Region // 冒险/活动等 → 回首页；全 0=未配置
}

func DefaultFeature() *Feature {
	return &Feature{
		Home: PageSlot{
			Identify: screen.Feature{
				// 多点比色：x,y,color,x,y,color,...（取色工具绝对坐标格式）
				Colors: "1327,825,d5e7e7-101010,266,806,a157e3-101010,99,818,f9ed78-101010,58,325,b1001b-101010,80,554,6c2d28-101010",
				Sim:    0.9,
			},
		},
		Adventure: OCRPage{
			Region:  screen.Region{X1: 89, Y1: 28, X2: 181, Y2: 76},
			Keyword: "冒险",
		},
		Event: OCRPage{
			Region:  screen.Region{X1: 681, Y1: 171, X2: 915, Y2: 242},
			Keyword: "王国活动",
		},
		Actions: NavActions{
			AdventureBtn: screen.Region{X1: 1335, Y1: 814, X2: 1375, Y2: 841},
			EventBtn:     screen.Region{X1: 246, Y1: 802, X2: 271, Y2: 832},
		},
	}
}
