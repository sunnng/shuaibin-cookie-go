package battle

import (
	"time"

	"app/internal/store"
)

// lastBattleAtKey 对应 Lua mine_battle_session.lastBattleAt（上次战斗时间戳，Unix 秒）。
const lastBattleAtKey = "mine_battle_last_at"

// State 战斗会话：记录上次战斗时间，控制战斗检测频率。
type State struct {
	store *store.Store
}

func NewState(store *store.Store) *State {
	return &State{store: store}
}

// RecordBattle 记录本次战斗开始时间（Lua Session.recordBattle）。
func (s *State) RecordBattle() {
	_ = s.store.Set(lastBattleAtKey, time.Now().Unix())
}

// GetTimeUntilNext 距离下次战斗还剩多久；0 表示可运行（Lua Session.getTimeUntilNext）。
func (s *State) GetTimeUntilNext(interval time.Duration) time.Duration {
	ts, ok := s.store.GetInt64(lastBattleAtKey)
	if !ok {
		return 0
	}
	remain := time.Until(time.Unix(ts, 0).Add(interval))
	if remain < 0 {
		return 0
	}
	return remain
}

// CheckReady 冷却是否到期（供调度 CheckReady 接线）。
func (s *State) CheckReady(interval time.Duration) (bool, time.Duration) {
	remain := s.GetTimeUntilNext(interval)
	return remain <= 0, remain
}

// Clear 清理会话。
func (s *State) Clear() {
	_ = s.store.Delete(lastBattleAtKey)
}
