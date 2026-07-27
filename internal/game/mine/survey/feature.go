// Package survey 矿山勘查子任务（Lua 模块_矿山勘查）。
// 流程：导航进勘查域 → setup 启动勘查 → OCR 读层数 → 达标结算 /
// 远距回城等待（farWait，持久化截止时间）/ 近距轮询（polling）。
package survey

import (
	"app/internal/platform/screen"
)

// Feature 勘查域各阶段识别与动作（对应 Lua 矿山_特征库.mineVenture）。
type Feature struct {
	Setup   SetupFeature
	Ready   ReadyFeature
	Running RunningFeature
	Settle  SettleFeature
	Dialogs DialogsFeature
	BackBtn screen.Region // 勘查域 → 矿山首页
}

// SetupFeature 布阵页（自动选择饼干 + 开始）。
type SetupFeature struct {
	Identify      screen.Feature
	StartBtn      screen.Region
	AutoSelectBtn screen.Region
}

// ReadyFeature 就绪页（饼干已选好，可开始）。
type ReadyFeature struct {
	Identify screen.Feature
	StartBtn screen.Region
}

// RunningFeature 勘查进行中（可停止，可读当前层数）。
type RunningFeature struct {
	Identify screen.Feature
	StopBtn  screen.Region
	FloorOCR screen.Region // 当前层数 OCR 区域
}

// SettleFeature 结算页。
type SettleFeature struct {
	Identify  screen.Feature
	FinishBtn screen.Region
}

// DialogsFeature setup/stop 流程内顺序出现的弹窗（流程内联处理，非 Guard trap）。
type DialogsFeature struct {
	Info          DialogDef // 勘查说明弹窗
	ConfirmCookie DialogDef // 确认饼干弹窗
	Stop          DialogDef // 停止确认弹窗
}

type DialogDef struct {
	Identify screen.Feature
	Confirm  screen.Region
}

func DefaultFeature() *Feature {
	return &Feature{
		Setup: SetupFeature{
			Identify: screen.Feature{
				Colors: "79,178,fff49f-101010,85,184,4a496b-101010,401,660,08a6de-101010,491,667,ffd300-101010,558,747,a56142-101010,1184,304,8b4000-101010,1196,291,ce7931-101010,1322,833,a1a1a1-101010,1315,839,ffffff-101010",
				Sim:    0.9,
			},
			StartBtn:      screen.Region{X1: 1217, Y1: 815, X2: 1301, Y2: 846},
			AutoSelectBtn: screen.Region{X1: 545, Y1: 643, X2: 630, Y2: 669},
		},
		Ready: ReadyFeature{
			Identify: screen.Feature{
				Colors: "1337,837,7ace0e-101010,657,657,ffd200-101010,398,659,0ca6df-101010,682,751,a8623f-101010,80,179,fbe788-101010,1217,292,cc4f55-101010",
				Sim:    0.95,
			},
			StartBtn: screen.Region{X1: 1217, Y1: 815, X2: 1301, Y2: 846},
		},
		Running: RunningFeature{
			Identify: screen.Feature{
				Colors: "78,177,fdf7a1-101010,219,51,ffffff-101010,573,833,0ca6df-101010,1191,292,d38235-101010,1290,684,d4b89a-101010",
				Sim:    0.95,
			},
			StopBtn:  screen.Region{X1: 379, Y1: 816, X2: 515, Y2: 850},
			FloorOCR: screen.Region{X1: 222, Y1: 138, X2: 686, Y2: 203},
		},
		Settle: SettleFeature{
			Identify: screen.Feature{
				Colors: "233,612,d88c16-101010,1576,795,da9019-101010,913,185,ff2323-101010,903,208,9d9d9d-101010,901,196,deac51-101010",
				Sim:    0.95,
			},
			FinishBtn: screen.Region{X1: 708, Y1: 815, X2: 888, Y2: 860},
		},
		Dialogs: DialogsFeature{
			Info: DialogDef{
				Identify: screen.Feature{
					Colors: "898,617,7bcf10-101010,864,263,ffffff-101010,867,264,020101-101010,704,368,efebe7-101010,710,368,4269ad-101010,1110,264,39a6e7-101010",
					Sim:    0.9,
				},
				Confirm: screen.Region{X1: 760, Y1: 604, X2: 849, Y2: 629},
			},
			ConfirmCookie: DialogDef{
				Identify: screen.Feature{
					Colors: "1005,630,7bcf10-101010,730,630,08a6de-101010,830,248,393c63-101010,890,420,505050-101010,893,419,f7ebde-101010,919,506,8c8c8c-101010",
					Sim:    0.9,
				},
				Confirm: screen.Region{X1: 916, Y1: 618, X2: 977, Y2: 645},
			},
			Stop: DialogDef{
				Identify: screen.Feature{
					Colors: "1062,629,f45a1e-101010,745,626,0ca6df-101010,892,246,363d5f-101010,75,197,553a23-101010,556,833,021821-101010",
					Sim:    0.95,
				},
				Confirm: screen.Region{X1: 905, Y1: 608, X2: 1004, Y2: 650},
			},
		},
		BackBtn: screen.Region{X1: 1546, Y1: 39, X2: 1559, Y2: 53},
	}
}
