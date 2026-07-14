package arena

import (
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

type Feature struct {
	Entry      EntryFeature
	Lobby      LobbyFeature
	TeamSelect TeamSelectFeature
	Settlement SettlementFeature
	Dialogs    DialogsFeature
}

// EntryFeature locates the arena entrance on the adventure page via OCR.
type EntryFeature struct {
	Region  screen.Region // OCR search region
	Keyword string        // empty → treated as "王国竞技场" in TapEntry
}

type LobbyFeature struct {
	Identify screen.Feature
	Actions  LobbyActions
	Reads    LobbyReads
	Opponent OpponentFeature
	Gestures LobbyGestures
}

type LobbyActions struct {
	Close            screen.Region // 关闭大厅；区域内随机点
	FreeRefresh      screen.Region
	BuyTicket        screen.Region
	BuyTicketSlider  action.Swipe
	BuyTicketConfirm screen.Region
}

type LobbyReads struct {
	MedalTicket screen.Region
	Trophy      screen.Region
	Refresh     screen.Region
	FreeRefresh screen.Region
}

type LobbyGestures struct {
	SwipeLeft action.Swipe
}

type OpponentFeature struct {
	Anchor       ColorFind     // 找卡锚点：Region=搜索区, Colors=锚点颜色串(单/多点), Sim, Dir
	TrophyRect   screen.Region // 相对锚点的奖杯 OCR 偏移矩形 (dx1,dy1,dx2,dy2)
	ResultOffset screen.Point  // 相对锚点的战绩标记点偏移
	ResultColors ResultColors  // 已战颜色 {Win,Draw,Lose}
	ResultSim    float32       // 战绩 MatchColor 相似度
	ClickOffset  screen.Point  // 相对锚点的点击偏移；锚点本身可点则 (0,0)
}

// ColorFind 对齐 images.FindColor / FindMultiColors[All] 的全部静态参数。
// Colors 写单色串即"找首个/单点"，写多点串即"多点匹配"；返回首个还是全部由代码侧选择，字段不变。
type ColorFind struct {
	Region screen.Region // x1,y1,x2,y2；右下 0,0 = 屏幕最大
	Colors string        // 单色串或多点串（取色工具直接拷贝）
	Sim    float32       // 0.1–1.0
	Dir    int           // 0 左上起 / 1 右上 / 2 左下 / 3 右下；本段默认 0
}

type ResultColors struct {
	Win  string
	Draw string
	Lose string
}

type TeamSelectFeature struct {
	Identify screen.Feature
	Actions  TeamSelectActions
}

type TeamSelectActions struct {
	StartBattle screen.Region // 开战按钮区域；随机点
}

type SettlementFeature struct {
	Identify screen.Feature
	Actions  SettlementActions
	Reads    SettlementReads
}

type SettlementActions struct {
	LeaveIdentify screen.Feature
	Leave         screen.Rect
}

type SettlementReads struct {
	Result screen.Rect
}

type DialogsFeature struct {
	MissingTopping DialogDef
	DeployMore     DialogDef
}

type DialogDef struct {
	Identify screen.Feature
	Confirm  screen.Region // 确认按钮区域；随机点
}

func DefaultFeature() *Feature {
	return &Feature{}
}
