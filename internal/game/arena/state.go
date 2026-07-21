package arena

import (
	"fmt"
	"time"

	"app/internal/config"
	"app/internal/store"
)

const nextFreeRefreshKey = "arena_next_free_refresh_at"

type State struct {
	store *store.Store

	Wins     int
	Draws    int
	Losses   int
	BuyCount int
	Tickets  int
	Medals   int
	Trophies int
}

func NewState(store *store.Store) *State {
	return &State{store: store}
}

func (s *State) TotalBattles() int {
	return s.Wins + s.Draws + s.Losses
}

// StatusText 生成给灵动岛展示的一行状态：战斗次数（有上限时带上限）与胜率。
// 0 场时不显示胜率。
func (s *State) StatusText(cfg *config.Arena) string {
	total := s.TotalBattles()
	text := "竞技场 "
	if cfg != nil && cfg.MaxBattles != nil && *cfg.MaxBattles > 0 {
		text += fmt.Sprintf("%d/%d", total, *cfg.MaxBattles)
	} else {
		text += fmt.Sprintf("%d 场", total)
	}
	if total > 0 {
		text += fmt.Sprintf(" · 胜率 %d%%", s.Wins*100/total)
	}
	return text
}

func (s *State) IsReachMaxBattles(cfg *config.Arena) bool {
	if cfg.MaxBattles == nil || *cfg.MaxBattles <= 0 {
		return false
	}
	return s.TotalBattles() >= *cfg.MaxBattles
}

func (s *State) SetNextFreeRefreshAt(at time.Time) {
	_ = s.store.Set(nextFreeRefreshKey, at.Unix())
}

func (s *State) NextFreeRefreshAt() time.Time {
	ts, ok := s.store.GetInt64(nextFreeRefreshKey)
	if !ok {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

func (s *State) TimeUntilRefresh() time.Duration {
	at := s.NextFreeRefreshAt()
	if at.IsZero() {
		return 0
	}
	remain := time.Until(at)
	if remain < 0 {
		return 0
	}
	return remain
}

func (s *State) ClearNextFreeRefresh() {
	_ = s.store.Delete(nextFreeRefreshKey)
}

func (s *State) Reset() {
	s.Wins = 0
	s.Draws = 0
	s.Losses = 0
	s.BuyCount = 0
	s.Tickets = 0
	s.Medals = 0
	s.Trophies = 0
}
