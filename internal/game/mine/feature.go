// Package mine 承载「未知的地底矿山」的共享部分：矿山首页识别/导航特征、
// 矿山首页 Page 与跨子任务复用的 Route。四个子任务（勘查 survey / 开采
// mining / 战斗 battle / 解除洋菜冻 jelly）各自为政，只共享本包。
//
// 特征坐标一律 1600×900 基准，颜色串原样搬自 Lua 矿山_特征库.lua，
// 只做格式转换：Lua 的 "x|y|color,x|y|color" → Go DetectsMultiColors 的
// "x,y,color,x,y,color"；Lua findMultiColorT 的 {x1,y1,x2,y2,ref,"dx|dy|color…",dir,sim}
// → ColorFind{Region, "ref,dx,dy,color,…", Sim, Dir}。
package mine

import (
	"app/internal/platform/screen"
)

// ColorFind 对齐 images.FindMultiColors[All] 的全部静态参数（与 arena.ColorFind 同形）。
// Colors 首段为参考颜色，其后为 dx,dy,color 偏移三元组；返回首个还是全部由代码侧选择。
type ColorFind struct {
	Region screen.Region // x1,y1,x2,y2；右下 0,0 = 屏幕最大
	Colors string        // 参考颜色+偏移多点串（取色工具直接拷贝，| 已转 ,）
	Sim    float32       // 0.1–1.0
	Dir    int           // 0 左上起 / 1 右上 / 2 左下 / 3 右下；本段默认 0
}

// Feature 矿山共享特征：矿山首页 + 王国活动页上的矿山入口。
type Feature struct {
	Home  HomeFeature
	Entry EntryFeature
}

// HomeFeature 矿山首页（四个子任务入口所在页）。
type HomeFeature struct {
	Identify              screen.Feature // 矿山首页多点比色
	CompletedTaskIdentify screen.Feature // 首页「存在已完成开采任务」角标
	Actions               HomeActions
}

// HomeActions 首页入口与返回；点击时在区域内随机取点。
type HomeActions struct {
	VentureBtn screen.Region // 勘查入口
	MiningBtn  screen.Region // 开采入口
	BattleBtn  screen.Region // 战斗入口
	JellyBtn   screen.Region // 解除洋菜冻入口（Lua 定义在 mineVenture 特征库）
	BackBtn    screen.Region // 返回王国首页
}

// EntryFeature 王国活动页 → 未知的地底矿山入口。
// Go 的 common/kingdom 包没有矿山入口按钮（Lua KingdomPage.tapMineBtn），
// 因此入口坐标放在本包，由 Route.KingdomHomeToMineHome 使用。
type EntryFeature struct {
	MineBtn screen.Region // 王国活动页上的矿山入口
}

func DefaultFeature() *Feature {
	return &Feature{
		Home: HomeFeature{
			Identify: screen.Feature{
				Colors: "1531,49,34a1e3,1379,60,f7e5cb,1291,58,e79d56,57,235,b3001c,73,256,f7e1b3,66,655,df994c,282,796,64677f,97,804,f37f0a,432,761,832427,1277,773,fffdf3,1298,791,db71c3",
				Sim:    0.9,
			},
			CompletedTaskIdentify: screen.Feature{
				Colors: "1565,747,ffffff-101010,1570,751,ff0000-101010,1558,757,000000-101010,1483,790,2e9ddf-101010,1535,767,ef6693-101010,1511,758,ffebff-101010,1454,840,efa150-101010",
				Sim:    0.9,
			},
			Actions: HomeActions{
				VentureBtn: screen.Region{X1: 1271, Y1: 799, X2: 1307, Y2: 831},
				MiningBtn:  screen.Region{X1: 1457, Y1: 791, X2: 1499, Y2: 822},
				BattleBtn:  screen.Region{X1: 1505, Y1: 651, X2: 1528, Y2: 675},
				JellyBtn:   screen.Region{X1: 419, Y1: 788, X2: 435, Y2: 808},
				BackBtn:    screen.Region{X1: 1546, Y1: 39, X2: 1559, Y2: 53},
			},
		},
		Entry: EntryFeature{
			MineBtn: screen.Region{X1: 1228, Y1: 578, X2: 1253, Y2: 601},
		},
	}
}
