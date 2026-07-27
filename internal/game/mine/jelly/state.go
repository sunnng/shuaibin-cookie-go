package jelly

import (
	"time"

	"app/internal/store"
)

// waitUntilKey 对应 Lua mine_jelly_session.waitUntil（冷却等待截止戳，Unix 秒）。
const waitUntilKey = "mine_jelly_wait_until"

// State 解除洋菜冻会话：完成/无操作后的冷却等待持久化。
type State struct {
	store *store.Store
}

func NewState(store *store.Store) *State {
	return &State{store: store}
}

// EnterWait 记录等待截止时间（Lua Session.enterWait）。
func (s *State) EnterWait(d time.Duration) {
	_ = s.store.Set(waitUntilKey, time.Now().Add(d).Unix())
}

// CheckReady 冷却是否到期（Lua Session.checkReady）。
// ready=true 表示可运行；否则 remain 为剩余等待时长。
func (s *State) CheckReady() (bool, time.Duration) {
	ts, ok := s.store.GetInt64(waitUntilKey)
	if !ok {
		return true, 0
	}
	remain := time.Until(time.Unix(ts, 0))
	if remain <= 0 {
		return true, 0
	}
	return false, remain
}

// RestoreProgress 冷却剩余时长（0 表示无等待或已到期），供 idle provider。
func (s *State) RestoreProgress() time.Duration {
	_, remain := s.CheckReady()
	return remain
}

// Clear 清理会话。
func (s *State) Clear() {
	_ = s.store.Delete(waitUntilKey)
}
