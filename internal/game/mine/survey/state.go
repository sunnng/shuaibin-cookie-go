package survey

import (
	"time"

	"app/internal/store"
)

// farWaitUntilKey 对应 Lua mine_venture_session.farWaitUntil（远距等待截止戳，Unix 秒）。
const farWaitUntilKey = "mine_survey_far_wait_until"

// State 勘查会话：远距等待截止时间持久化（跨轮次由调度 CheckReady 检查）。
type State struct {
	store *store.Store
}

func NewState(store *store.Store) *State {
	return &State{store: store}
}

// EnterFarWait 进入远距等待（Lua Session.enterFarWait）。
func (s *State) EnterFarWait(d time.Duration) {
	_ = s.store.Set(farWaitUntilKey, time.Now().Add(d).Unix())
}

// CheckFarWait 远距等待是否已到期（Lua Session.checkFarWait）。
// ready=true 表示可运行；否则 remain 为剩余等待时长。
func (s *State) CheckFarWait() (bool, time.Duration) {
	ts, ok := s.store.GetInt64(farWaitUntilKey)
	if !ok {
		return true, 0
	}
	remain := time.Until(time.Unix(ts, 0))
	if remain <= 0 {
		return true, 0
	}
	return false, remain
}

// RestoreProgress 远距等待剩余时长（0 表示无等待或已到期），供 idle provider。
func (s *State) RestoreProgress() time.Duration {
	_, remain := s.CheckFarWait()
	return remain
}

// Clear 清理会话。
func (s *State) Clear() {
	_ = s.store.Delete(farWaitUntilKey)
}
