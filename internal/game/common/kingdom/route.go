package kingdom

import "time"

type Route struct {
	page *Page
}

func NewRoute(page *Page) *Route { return &Route{page: page} }

func (r *Route) KingdomHomeToAdventure() bool {
	if r.page.IsAdventurePage() {
		return true
	}
	if !r.page.IsKingdomHome() {
		return false
	}
	r.page.TapAdventureBtn()
	return r.page.WaitAdventure(30 * time.Second)
}

func (r *Route) AdventureToKingdomHome() bool {
	// Placeholder: press back until home
	return true
}
