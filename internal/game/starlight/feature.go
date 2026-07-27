package starlight

import (
	"app/internal/platform/screen"
)

// Feature 描述梦幻繁星岛的页面识别与交互坐标，全部基于 1600×900 基准。
// 颜色串/坐标原样搬运自 Lua 坐标库（繁星岛_坐标库.lua）与王国特征库
// （通用_王国/特征库.lua 的 event 段），已由竖线格式转为 Go 的逗号格式。
type Feature struct {
	Home    HomeFeature     // 繁星岛首页
	Manual  ManualFeature   // 航海手册页
	Vanilla VanillaFeature  // 纯香草小岛页
	Task    TaskPageFeature // 任务页
	Event   EventFeature    // 王国活动页（导航中转）
}

type HomeFeature struct {
	Identify screen.Feature
	Actions  HomeActions
}

type HomeActions struct {
	ManualBtn screen.Region // 航海手册按钮
	TaskBtn   screen.Region // 任务按钮
	BackBtn   screen.Region // 返回王国按钮
}

type ManualFeature struct {
	Identify screen.Feature
	Actions  ManualActions
}

type ManualActions struct {
	LoginIslandBtn screen.Region // 登陆回忆小岛按钮
}

type VanillaFeature struct {
	Identify screen.Feature
	Actions  VanillaActions
}

type VanillaActions struct {
	BackBtn screen.Region // 返回繁星岛首页
}

type TaskPageFeature struct {
	Identify     screen.Feature
	Actions      TaskPageActions
	Claim        ColorFind    // 可领奖按钮多点找色
	DismissPoint screen.Point // 领奖弹窗关闭点击点（屏幕中央空白处）
}

type TaskPageActions struct {
	BackBtn screen.Region // 返回繁星岛首页
}

// EventFeature 王国活动页：路由从王国首页点事件按钮后在此页找繁星岛入口。
// 特征取自 Lua 通用_王国/特征库.lua 的 event 段（common/kingdom 的 Go 特征
// 只保留了 OCR 判定，不含多点比色串，故在本包内重复一份）。
type EventFeature struct {
	Identify screen.Feature
	Actions  EventActions
}

type EventActions struct {
	StarlightBtn screen.Region // 梦幻繁星岛入口按钮
}

// ColorFind 对齐 FindMultiColorsAll 的静态参数（与 arena.ColorFind 同形）。
// Colors 为 Go 逗号格式："基准色,dx,dy,颜色,dx,dy,颜色,..."。
type ColorFind struct {
	Region screen.Region // x1,y1,x2,y2；右下 0,0 = 屏幕最大
	Colors string        // 多点找色串（由 Lua 竖线格式转换而来）
	Sim    float32       // 0.1–1.0
	Dir    int           // 0 左上起 / 1 右上 / 2 左下 / 3 右下
}

func DefaultFeature() *Feature {
	return &Feature{
		Home: HomeFeature{
			Identify: screen.Feature{
				Colors: "265,777,1a953c-101010,72,246,fbe7ab-101010,573,63,5cebaf-101010,1526,55,36a1e3-101010,1516,204,263e7a-101010,1462,801,b12d38-101010",
				Sim:    0.9,
			},
			Actions: HomeActions{
				ManualBtn: screen.Region{X1: 1431, Y1: 800, X2: 1453, Y2: 822},
				TaskBtn:   screen.Region{X1: 1526, Y1: 188, X2: 1538, Y2: 201},
				BackBtn:   screen.Region{X1: 1536, Y1: 51, X2: 1548, Y2: 58},
			},
		},
		Manual: ManualFeature{
			Identify: screen.Feature{
				Colors: "67,83,dba12c-101010,1532,68,36a5e6-101010,367,775,7ace0e-101010,81,824,288ead-101010,1454,831,2d91af-101010",
				Sim:    0.9,
			},
			Actions: ManualActions{
				LoginIslandBtn: screen.Region{X1: 428, Y1: 761, X2: 464, Y2: 770},
			},
		},
		Vanilla: VanillaFeature{
			Identify: screen.Feature{
				Colors: "29,36,ffffff-101010,463,55,ffffff-101010,1532,66,36a3e3-101010,1513,294,fff3a7-101010,1463,779,28a946-101010",
				Sim:    0.9,
			},
			Actions: VanillaActions{
				BackBtn: screen.Region{X1: 1536, Y1: 51, X2: 1548, Y2: 58},
			},
		},
		Task: TaskPageFeature{
			Identify: screen.Feature{
				Colors: "342,68,95979f-101010,1421,73,34a1e3-101010,886,81,ffd960-101010,104,796,4f0411-101010,1463,802,58161b-101010",
				Sim:    0.9,
			},
			Actions: TaskPageActions{
				BackBtn: screen.Region{X1: 1392, Y1: 67, X2: 1407, Y2: 82},
			},
			Claim: ColorFind{
				Region: screen.Region{X1: 216, Y1: 171, X2: 1364, Y2: 618},
				Colors: "e12d52-101010,-48,-39,93db04-101010,48,34,00a100-101010,16,-36,f7f3df-101010,-24,27,f1c1ab-101010,24,0,fd6e8b-101010,-18,-26,560308-101010",
				Sim:    0.9,
				Dir:    0,
			},
			DismissPoint: screen.Point{X: 800, Y: 450},
		},
		Event: EventFeature{
			Identify: screen.Feature{
				Colors: "1311,843,261d06-101010,722,820,2d1f00-101010,235,825,2d2718-101010,69,805,252625-101010,71,332,2e2a1f-101010,795,140,b59756-101010,1290,65,2a1c0f-101010,1551,69,36a3e3-101010",
				Sim:    0.9,
			},
			Actions: EventActions{
				StarlightBtn: screen.Region{X1: 1225, Y1: 366, X2: 1252, Y2: 387},
			},
		},
	}
}
