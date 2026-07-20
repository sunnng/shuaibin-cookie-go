package action

// Executor 执行触控动作。动作层不提供 error 返回：唯一可失败点是取显示信息
// （通常意味着设备掉线），实现侧记录日志并跳过该次动作；调用方无需也无法
// 对单次点击失败做有意义的恢复，识别层会在下一轮自然发现页面异常。
type Executor interface {
	Tap(p Point)
	LongTap(p Point, ms int)
	Swipe(from, to Point, ms int)
	Back()
	Home()
	Sleep(ms int)
}

// Swipe is a placeholder for a swipe gesture definition.
type Swipe struct {
	From       Point
	To         Point
	DurationMs int
}
