package square

import (
	"testing"
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// routeEnv 用脚本化 detector/executor 驱动真实 *Page + *Route（学 arena route_test）。
// taps 命中指定区域时翻转对应页面匹配。
type routeEnv struct {
	det  *pageDet
	exec *routeTapExec
	page *Page
	rt   *Route

	sqColors  string
	dlgColors string
	kgColors  string
}

type routeTapExec struct {
	det  *pageDet
	env  *routeEnv
	taps []action.Point
	f    *Feature
	kf   *kingdom.Feature
}

func (e *routeTapExec) Tap(p action.Point) {
	e.taps = append(e.taps, p)
	in := func(r screen.Region) bool {
		return p.X >= r.X1 && p.X <= r.X2 && p.Y >= r.Y1 && p.Y <= r.Y2
	}
	env := e.env
	switch {
	case in(e.f.EntryBtn):
		// 王国首页 → 广场
		e.det.match[env.kgColors] = false
		e.det.match[env.sqColors] = true
	case in(e.f.Home.Actions.BackBtn):
		// 广场页 → 离开弹窗
		e.det.match[env.sqColors] = false
		e.det.match[env.dlgColors] = true
	case in(e.f.Dialog.Actions.LeaveBtn):
		// 弹窗 → 王国主城
		e.det.match[env.dlgColors] = false
		e.det.match[env.kgColors] = true
	case in(e.f.Dialog.Actions.CancelBtn):
		// 弹窗 → 广场页
		e.det.match[env.dlgColors] = false
		e.det.match[env.sqColors] = true
	}
}

func (e *routeTapExec) LongTap(p action.Point, ms int)      {}
func (e *routeTapExec) Swipe(from, to action.Point, ms int) {}
func (e *routeTapExec) Back()                               {}
func (e *routeTapExec) Home()                               {}
func (e *routeTapExec) Sleep(ms int)                        {}

func newRouteEnv() *routeEnv {
	f := DefaultFeature()
	kf := kingdom.DefaultFeature()
	det := &pageDet{match: map[string]bool{}, ocr: map[screen.Region]string{}}
	env := &routeEnv{
		det:       det,
		sqColors:  f.Home.Identify.Colors,
		dlgColors: f.Dialog.Identify.Colors,
		kgColors:  kf.Home.Identify.Colors,
	}
	exec := &routeTapExec{det: det, env: env, f: f, kf: kf}
	env.exec = exec
	env.page = NewPage(det, exec, f)
	env.rt = NewRoute(env.page, kingdom.NewPage(det, exec, kf))
	return env
}

func (env *routeEnv) atKingdomHome() {
	env.det.match[env.kgColors] = true
	env.det.match[env.sqColors] = false
	env.det.match[env.dlgColors] = false
}

func TestRouteEnter(t *testing.T) {
	env := newRouteEnv()
	env.atKingdomHome()
	if !env.rt.Enter() {
		t.Fatal("Enter from kingdom home should succeed")
	}
	if !env.det.match[env.sqColors] {
		t.Fatal("should end on square page")
	}

	// 已在广场：直接 true，不再点。
	env.exec.taps = nil
	if !env.rt.Enter() {
		t.Fatal("Enter on square page should be a no-op true")
	}
	if len(env.exec.taps) != 0 {
		t.Fatalf("want 0 taps, got %d", len(env.exec.taps))
	}

	// 不在主城也不在广场：false。
	env.det.match[env.sqColors] = false
	if env.rt.Enter() {
		t.Fatal("Enter from unknown page should fail")
	}
}

func TestRouteEnsureSquare(t *testing.T) {
	env := newRouteEnv()
	env.atKingdomHome()
	if !env.rt.EnsureSquare() {
		t.Fatal("EnsureSquare from kingdom home should enter square")
	}
	if !env.det.match[env.sqColors] {
		t.Fatal("should end on square page")
	}

	// 弹窗内也算广场上下文。
	env.det.match[env.sqColors] = false
	env.det.match[env.dlgColors] = true
	env.det.match[env.kgColors] = false
	if !env.rt.EnsureSquare() {
		t.Fatal("leave dialog is square context")
	}

	// 未知界面：false。
	env.det.match[env.dlgColors] = false
	if env.rt.EnsureSquare() {
		t.Fatal("unknown page should fail")
	}
}

func TestRouteOpenLeaveDialog(t *testing.T) {
	env := newRouteEnv()
	env.atKingdomHome()
	if !env.rt.OpenLeaveDialog() {
		t.Fatal("OpenLeaveDialog should navigate and open the dialog")
	}
	if !env.det.match[env.dlgColors] {
		t.Fatal("should end in leave dialog")
	}

	// 已在弹窗：直接 true，不再点。
	env.exec.taps = nil
	if !env.rt.OpenLeaveDialog() {
		t.Fatal("already in dialog → true")
	}
	if len(env.exec.taps) != 0 {
		t.Fatalf("want 0 taps, got %d", len(env.exec.taps))
	}
}

func TestRouteLeaveToKingdom(t *testing.T) {
	env := newRouteEnv()
	env.atKingdomHome()

	// 从广场页离开：先点返回开弹窗，再点离开。
	env.rt.Enter()
	if !env.rt.LeaveToKingdom(time.Second) {
		t.Fatal("LeaveToKingdom from square should succeed")
	}
	if !env.det.match[env.kgColors] {
		t.Fatal("should end on kingdom home")
	}

	// 从弹窗离开：直接点离开。
	env.rt.OpenLeaveDialog()
	if !env.rt.LeaveToKingdom(time.Second) {
		t.Fatal("LeaveToKingdom from dialog should succeed")
	}
	if !env.det.match[env.kgColors] {
		t.Fatal("should end on kingdom home")
	}

	// 已在主城：直接 true。
	if !env.rt.LeaveToKingdom(time.Second) {
		t.Fatal("already home → true")
	}
}

func TestRouteLeave(t *testing.T) {
	env := newRouteEnv()
	env.atKingdomHome()

	// 卡在广场页：开弹窗后回主城。
	env.rt.Enter()
	if !env.rt.Leave() {
		t.Fatal("Leave from square page should succeed")
	}
	if !env.det.match[env.kgColors] {
		t.Fatal("should end on kingdom home")
	}

	// 卡在弹窗：直接回主城。
	env.rt.OpenLeaveDialog()
	if !env.rt.Leave() {
		t.Fatal("Leave from dialog should succeed")
	}
	if !env.det.match[env.kgColors] {
		t.Fatal("should end on kingdom home")
	}

	// 不在广场上下文：LeaveToKingdom 兜底（主城已 true）。
	if !env.rt.Leave() {
		t.Fatal("already home → true")
	}
}
