package mining

import (
	"time"

	"app/internal/store"
)

// allBusyUntilKey 对应 Lua mine_mining_session.allBusyUntil（全部栏位 busy 截止戳，Unix 秒）。
const allBusyUntilKey = "mine_mining_all_busy_until"

// State 开采会话：所有栏位 busy 的等待截止时间持久化（跨轮次由调度 CheckReady 检查）。
type State struct {
	store *store.Store
}

func NewState(store *store.Store) *State {
	return &State{store: store}
}

// EnterBusyWait 进入 busy 等待（开采页确认没有已完成/空闲/可启动栏位时调用）。
func (s *State) EnterBusyWait(d time.Duration) {
	_ = s.store.Set(allBusyUntilKey, time.Now().Add(d).Unix())
}

// CheckReady busy 等待是否已到期（Lua Session.checkReady）。
// ready=true 表示可运行；否则 remain 为剩余等待时长。
func (s *State) CheckReady() (bool, time.Duration) {
	ts, ok := s.store.GetInt64(allBusyUntilKey)
	if !ok {
		return true, 0
	}
	remain := time.Until(time.Unix(ts, 0))
	if remain <= 0 {
		return true, 0
	}
	return false, remain
}

// RestoreProgress busy 剩余时长（0 表示无等待或已到期），供 idle provider。
func (s *State) RestoreProgress() time.Duration {
	_, remain := s.CheckReady()
	return remain
}

// Clear 清理会话。
func (s *State) Clear() {
	_ = s.store.Delete(allBusyUntilKey)
}
