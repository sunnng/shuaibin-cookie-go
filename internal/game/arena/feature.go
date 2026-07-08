package arena

import (
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

type Feature struct {
	Lobby      LobbyFeature
	Opponent   OpponentFeature
	TeamSelect TeamSelectFeature
	Dialog     DialogFeature
	Settlement SettlementFeature
	Pagination PaginationFeature
}

type LobbyFeature struct {
	Feature          screen.Feature
	CloseBtn         screen.Region
	MedalTicketOCR   screen.Region
	TrophyOCR        screen.Region
	RefreshOCR       screen.Region
	FreeRefreshOCR   screen.Region
	FreeRefreshTap   screen.Point
	BuyTicketBtn     screen.Point
	BuyTicketSlider  action.Swipe
	BuyTicketConfirm screen.Point
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
	Feature     screen.Feature
	StartBattle screen.Point
}

type DialogFeature struct {
	MissingTopping DialogDef
	DeployMore     DialogDef
}

type DialogDef struct {
	Feature screen.Feature
	Confirm screen.Point
}

type SettlementFeature struct {
	Feature      screen.Feature
	ResultOCR    screen.Rect
	LeaveFeature screen.Feature
	LeaveBtn     screen.Rect
}

type PaginationFeature struct {
	SwipeLeft action.Swipe
}

func DefaultFeature() *Feature {
	return &Feature{
		// Placeholder: fill from Lua 竞技场_特征库.lua
	}
}
