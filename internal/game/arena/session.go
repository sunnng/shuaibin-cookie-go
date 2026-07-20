package arena

import (
	"fmt"
	"time"

	"app/internal/config"
	"app/internal/store"
)

const nextFreeRefreshKey = "arena_next_free_refresh_at"

type Session struct {
	store *store.Store

	Wins     int
	Draws    int
	Losses   int
	BuyCount int
	Tickets  int
	Medals   int
	Trophies int
}

func NewSession(store *store.Store) *Session {
	return &Session{store: store}
}

func (s *Session) TotalBattles() int {
	return s.Wins + s.Draws + s.Losses
}

// StatusText 生成给灵动岛展示的一行状态：战斗次数（有上限时带上限）与胜率。
// 0 场时不显示胜率。
func (s *Session) StatusText(cfg *config.Arena) string {
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

func (s *Session) IsReachMaxBattles(cfg *config.Arena) bool {
	if cfg.MaxBattles == nil || *cfg.MaxBattles <= 0 {
		return false
	}
	return s.TotalBattles() >= *cfg.MaxBattles
}

func (s *Session) SetNextFreeRefreshAt(at time.Time) {
	_ = s.store.Set(nextFreeRefreshKey, at.Unix())
}

func (s *Session) NextFreeRefreshAt() time.Time {
	ts, ok := s.store.GetInt64(nextFreeRefreshKey)
	if !ok {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

func (s *Session) TimeUntilRefresh() time.Duration {
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

func (s *Session) ClearNextFreeRefresh() {
	_ = s.store.Delete(nextFreeRefreshKey)
}

func (s *Session) Reset() {
	s.Wins = 0
	s.Draws = 0
	s.Losses = 0
	s.BuyCount = 0
	s.Tickets = 0
	s.Medals = 0
	s.Trophies = 0
}
