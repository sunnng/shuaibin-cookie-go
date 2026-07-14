package arena

import (
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
