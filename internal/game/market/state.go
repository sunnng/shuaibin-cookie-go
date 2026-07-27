package market

import (
	"fmt"
	"time"

	"app/internal/logger"
	"app/internal/store"
)

// nextRunAtKey 对齐 Lua 交易所_会话.lua 的 seaside_market_session.nextRunAt。
const nextRunAtKey = "market_next_run_at"

// defaultRestock 补货读数非法时的兜底等待（Lua DEFAULT_RESTOCK_SEC = 6h）。
const defaultRestock = 6 * time.Hour

// State 交易所会话：补货调度持久化 + 本轮扫货统计 + 启动首轮强制。
type State struct {
	store *store.Store

	// 启动首轮强制（Lua startupBypassPending/Active，进程级语义）：
	// 脚本每次启动后第一次 CheckReady 直接就绪，任务内 ConsumeStartupBypass
	// 取走该标记后，即使页面显示补货倒计时也强制扫一轮货。
	bypassPending bool
	bypassActive  bool

	Purchased int // 本轮购入
	SoldOut   int // 售罄跳过
	Shortage  int // 道具不足取消
	Failed    int // 购买失败
}

func NewState(store *store.Store) *State {
	return &State{store: store, bypassPending: true}
}

// CheckReady 调度就绪检查（对齐 Lua Session.checkReady）。
// 返回 (true, 0) 可运行；否则 (false, remain) 距补货到期剩余时长。
// 每次脚本启动的首次调用直接就绪并武装首轮强制标记（有副作用）。
func (s *State) CheckReady() (bool, time.Duration) {
	if s.bypassPending {
		s.bypassPending = false
		s.bypassActive = true
		logger.Infof("[Market] 本次脚本启动首轮强制执行，忽略补货等待")
		return true, 0
	}
	if remain := s.TimeUntilRestock(); remain > 0 {
		logger.Debugf("[Market] 补货等待中，剩余 %v", remain)
		return false, remain
	}
	return true, 0
}

// RestoreProgress 供 idle provider 查询距下次运行的剩余时长（Lua Session.restoreProgress）。
// 纯查询：不消费启动首轮标记，无补货记录或已到期返回 0。
func (s *State) RestoreProgress() time.Duration {
	return s.TimeUntilRestock()
}

// ConsumeStartupBypass 取走首轮强制标记，仅一次有效（Lua Session.consumeStartupBypass）。
func (s *State) ConsumeStartupBypass() bool {
	if s.bypassActive {
		s.bypassActive = false
		return true
	}
	return false
}

// ScheduleAfterRestock 按页面读到的补货倒计时写下次运行时间
// （Lua Session.scheduleAfterRestock：restock + buffer；非法读数兜底 6h）。
func (s *State) ScheduleAfterRestock(restock time.Duration, bufferSec int) {
	if restock < 0 {
		restock = defaultRestock
	}
	wait := restock + time.Duration(bufferSec)*time.Second
	next := time.Now().Add(wait)
	_ = s.store.Set(nextRunAtKey, next.Unix())
	logger.Infof("[Market] 下次补货调度 %v 后（到期戳 %d）", wait, next.Unix())
}

// NextRunAt 下次运行到期时间；无记录返回零值。
func (s *State) NextRunAt() time.Time {
	ts, ok := s.store.GetInt64(nextRunAtKey)
	if !ok {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

// TimeUntilRestock 距补货到期剩余时长；无记录或已到期返回 0。
func (s *State) TimeUntilRestock() time.Duration {
	at := s.NextRunAt()
	if at.IsZero() {
		return 0
	}
	remain := time.Until(at)
	if remain < 0 {
		return 0
	}
	return remain
}

// Clear 清除补货调度记录（Lua Session.clear）。
func (s *State) Clear() {
	_ = s.store.Delete(nextRunAtKey)
	logger.Infof("[Market] 会话已清理")
}

// StatusText 灵动岛一行状态：本轮扫货统计，全零时显示未购入。
func (s *State) StatusText() string {
	text := fmt.Sprintf("交易所 购%d", s.Purchased)
	if s.SoldOut > 0 {
		text += fmt.Sprintf(" · 售罄%d", s.SoldOut)
	}
	if s.Shortage > 0 {
		text += fmt.Sprintf(" · 不足%d", s.Shortage)
	}
	if s.Failed > 0 {
		text += fmt.Sprintf(" · 失败%d", s.Failed)
	}
	return text
}

// Reset 清零本轮统计（不动持久化的补货调度）。
func (s *State) Reset() {
	s.Purchased = 0
	s.SoldOut = 0
	s.Shortage = 0
	s.Failed = 0
}
