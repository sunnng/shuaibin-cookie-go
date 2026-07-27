package market

import (
	"strings"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// Feature 海滩交易所（常规_海滩交易所）全部 UI 常量，1600×900 基准。
// 颜色串/坐标原样搬运自 Lua 交易所_坐标库.lua；入口按钮来自 Lua 通用_王国/特征库 event。
type Feature struct {
	Page     PageFeature
	Entry    EntryFeature
	Tab      TabFeature
	List     ListFeature
	Slot     SlotOffsets
	Dialog   DialogFeature   // 购买确认弹窗
	Shortage ShortageFeature // 道具不足弹窗
	Stock    map[string]ColorFind
}

type PageFeature struct {
	Identify   screen.Feature // 交易所页身份（多点比色）
	CloseBtn   screen.Region  // 关闭交易所；区域内随机点
	RefreshBtn screen.Region  // 刷新按钮
	RefreshOcr screen.Region  // 刷新状态 OCR 区（"免费刷新" 或 "h:m:s" 倒计时）
	Arrow      ColorFind      // 列表右箭头（FindColor，Colors 支持 "|" 备选色）
}

// EntryFeature 王国活动页上的交易所入口。
type EntryFeature struct {
	Btn screen.Region // 活动页「海滩交易所」按钮；区域内随机点
}

type TabFeature struct {
	ItemExchange screen.Region // 「道具交易所」Tab 点击区
}

type ListFeature struct {
	Swipe     action.Swipe // 翻到下一页
	MaxSwipes int
}

// SlotOffsets 货架格子的相对几何：以商品锚点为基准换算购买按钮/售罄 OCR 矩形。
type SlotOffsets struct {
	BuyBtnOffsetY int // 购买按钮中心相对锚点的 Y 偏移
	BuyBtnHalfW   int
	BuyBtnHalfH   int
	CrateHalfW    int // 售罄 OCR 矩形半宽
	CrateHalfH    int
	CrateOffsetY  int // 售罄矩形中心相对锚点的 Y 偏移
}

// DialogFeature 购买确认弹窗。
type DialogFeature struct {
	Identify screen.Feature
	Confirm  screen.Region // 确认按钮；区域内随机点
	Cancel   screen.Region // 取消/关闭按钮
}

// ShortageFeature 「以下道具不足」弹窗。
type ShortageFeature struct {
	Identify screen.Feature
	Ocr      screen.Region // 弹窗文案 OCR 区（当前未用于判定，留作排障）
	Cancel   screen.Region
}

// ColorFind 对齐 images.FindColor / FindMultiColorsAll 的静态参数。
// Colors 作 FindColor 参数时为 "color|color" 备选串；作 FindMultiColorsAll
// 参数时为 "baseColor,dx,dy,color,..." 偏移串（取色工具串经 goColors 转换）。
type ColorFind struct {
	Region screen.Region // x1,y1,x2,y2；右下 0,0 = 屏幕最大
	Colors string
	Sim    float32 // 0.1–1.0
	Dir    int     // 0 左上起 / 1 右上 / 2 左下 / 3 右下
}

// goColors 把取色工具的 "|" 分隔串转成 AutoGo Go SDK 的 "," 分隔格式：
// 多点比色 "x|y|color,x|y|color" → "x,y,color,x,y,color"；
// 多点找色 offsets "dx|dy|color|..." → "dx,dy,color,..."。
func goColors(s string) string {
	return strings.ReplaceAll(s, "|", ",")
}

// stockFeature 组装商品货架特征（Lua findMultiColorT：
// {x1,y1,x2,y2, baseColor, "dx|dy|color|...", dir, sim}）。
func stockFeature(baseColor, offsets string) ColorFind {
	return ColorFind{
		Region: screen.Region{X1: 3, Y1: 602, X2: 1587, Y2: 707},
		Colors: baseColor + "," + goColors(offsets),
		Sim:    0.9,
		Dir:    0,
	}
}

func DefaultFeature() *Feature {
	return &Feature{
		Page: PageFeature{
			Identify: screen.Feature{
				Colors: goColors("37|723|f51b67-101010,189|95|14633c-101010,477|128|b2d155-101010,448|285|f5365e-101010,1409|828|895b3d-101010,1508|320|57432a-101010"),
				Sim:    0.9,
			},
			CloseBtn:   screen.Region{X1: 1530, Y1: 14, X2: 1584, Y2: 77},
			RefreshBtn: screen.Region{X1: 1419, Y1: 458, X2: 1459, Y2: 480},
			RefreshOcr: screen.Region{X1: 1298, Y1: 447, X2: 1557, Y2: 488},
			Arrow: ColorFind{
				Region: screen.Region{X1: 1524, Y1: 616, X2: 1577, Y2: 684},
				Colors: "000000-101010|000000-101010|000000-101010|030303-101010|030303-101010|12110d-101010",
				Sim:    0.9,
				Dir:    0,
			},
		},
		Entry: EntryFeature{
			Btn: screen.Region{X1: 574, Y1: 582, X2: 593, Y2: 604},
		},
		Tab: TabFeature{
			ItemExchange: screen.Region{X1: 559, Y1: 831, X2: 643, Y2: 862},
		},
		List: ListFeature{
			Swipe: action.Swipe{
				From:       action.Point{X: 1500, Y: 650},
				To:         action.Point{X: 100, Y: 650},
				DurationMs: 1200,
			},
			MaxSwipes: 20,
		},
		Slot: SlotOffsets{
			BuyBtnOffsetY: 110,
			BuyBtnHalfW:   105,
			BuyBtnHalfH:   24,
			CrateHalfW:    90,
			CrateHalfH:    65,
			CrateOffsetY:  -20,
		},
		Dialog: DialogFeature{
			Identify: screen.Feature{
				Colors: goColors("120|127|126345-101010,38|720|7f0e37-101010,339|241|2e5825-101010,478|128|5a692a-101010,1152|237|36a6e8-101010,455|220|686f9d-101010,1471|829|44351e-101010"),
				Sim:    0.9,
			},
			Confirm: screen.Region{X1: 776, Y1: 621, X2: 829, Y2: 646},
			Cancel:  screen.Region{X1: 1143, Y1: 211, X2: 1159, Y2: 229},
		},
		Shortage: ShortageFeature{
			Identify: screen.Feature{
				Colors: goColors("1559|854|261010-101010,27|717|710b2d-101010,131|222|1a3722-101010,351|273|7f7f55-101010,475|129|58682a-101010,517|246|68709d-101010,1092|265|36a5e5-101010"),
				Sim:    0.9,
			},
			Ocr:    screen.Region{X1: 715, Y1: 331, X2: 885, Y2: 376},
			Cancel: screen.Region{X1: 1086, Y1: 231, X2: 1102, Y2: 253},
		},
		Stock: map[string]ColorFind{
			"灿烂的光之碎片": stockFeature("a9e4ff-101010", "-7|-1|fffff3-101010|-34|-1|cfefb9-101010|14|1|cf97f6-101010|8|20|ef84a9-101010|-3|27|fad4a2-101010|-14|29|ee7fe3-101010|38|16|320f5d-101010|-24|-19|31105a-101010"),
			"十分钟加速券":  stockFeature("ffffff-101010", "0|-1|ffffff-101010|-4|12|2bd0e9-101010|-1|-25|6786bd-101010|-19|24|4168ad-101010|-40|0|6496c9-101010|37|-12|f9ffff-101010|36|17|c8e9f6-101010"),
			"商品1_金紫":  stockFeature("b5a7f9-101010", "-65|10|80824f-101010|5|5|ffffff-101010|0|10|ada2fa-101010|10|-5|edcffc-101010"),
			"商品2_蓝盒":  stockFeature("99bdff-101010", "0|-5|403f5e-101010|-15|0|2b2183-101010|5|0|201d53-101010|0|20|222255-101010"),
			"商品3_罗盘":  stockFeature("c6c0fa-101010", "-10|-10|8974e8-101010|10|-10|cec0f8-101010|0|10|fefefe-101010|15|0|ffffff-101010"),
			"商品4_绿书":  stockFeature("827d45-101010", "-20|0|837e47-101010|70|0|6b4eff-101010|80|0|6b4eff-101010|0|-20|837c46-101010|0|20|857c41-101010"),
			"商品5_卷轴":  stockFeature("896dfc-101010", "-10|-10|fb1efe-101010|10|-10|ffffff-101010|0|10|c5b5ff-101010|15|0|ffffff-101010"),
		},
	}
}
