package action

type Executor interface {
	Tap(p Point) error
	LongTap(p Point, ms int) error
	Swipe(from, to Point, ms int) error
	Back() error
	Home() error
	Sleep(ms int)
}
