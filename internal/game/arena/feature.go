package arena

import (
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

type Feature struct {
	Lobby      LobbyFeature
	TeamSelect TeamSelectFeature
	Settlement SettlementFeature
	Dialogs    DialogsFeature
}

type LobbyFeature struct {
	Identify screen.Feature
	Actions  LobbyActions
	Reads    LobbyReads
	Opponent OpponentFeature
	Gestures LobbyGestures
}

type LobbyActions struct {
	Close            screen.Region
	FreeRefresh      screen.Point
	BuyTicket        screen.Point
	BuyTicketSlider  action.Swipe
	BuyTicketConfirm screen.Point
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
	FindDef      screen.FindDef
	BaseSite     screen.Point
	TrophyRect   screen.Region
	ResultOffset screen.Point
	ResultColors ResultColors
	NumberOCR    screen.OCRCfg
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
	StartBattle screen.Point
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
	Confirm  screen.Point
}

func DefaultFeature() *Feature {
	return &Feature{}
}
