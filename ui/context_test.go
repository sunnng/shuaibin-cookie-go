package ui

import "testing"

func TestStatePersistsAcrossFrames(t *testing.T) {
	c := NewCtx(NewStore(), 1)
	c.Push("form")
	p1 := State(c, "draft", "init")
	*p1 = "edited"
	c.Pop()

	// 模拟下一帧：同一路径同一键拿到同一份状态
	c.Push("form")
	p2 := State(c, "draft", "init")
	c.Pop()
	if p2 != p1 || *p2 != "edited" {
		t.Fatalf("state should persist across frames, got %q", *p2)
	}
}

func TestStateIsolatedByComponentPath(t *testing.T) {
	c := NewCtx(NewStore(), 1)

	c.Push("panel")
	c.Push("formA")
	a := State(c, "draft", 0)
	*a = 1
	c.Pop()
	c.Push("formB")
	b := State(c, "draft", 0)
	*b = 2
	c.Pop()
	c.Pop()

	if *a != 1 || *b != 2 {
		t.Fatalf("same component twice must be isolated: a=%d b=%d", *a, *b)
	}

	// 键的类型不同属调用方错误；同键同型返回同一指针
	c.Push("panel")
	c.Push("formA")
	a2 := State(c, "draft", 99)
	c.Pop()
	c.Pop()
	if a2 != a {
		t.Fatal("same path+key must return same pointer")
	}
}

func TestStateConditionalRenderingSafe(t *testing.T) {
	c := NewCtx(NewStore(), 1)
	// 第一帧渲染了可选区块
	c.Push("root")
	State(c, "a", 1)
	c.Push("optional")
	State(c, "b", 2)
	c.Pop()
	c.Pop()
	// 第二帧不渲染可选区块，已有状态不受影响
	c.Push("root")
	if got := *State(c, "a", 0); got != 1 {
		t.Fatalf("a=%d want 1", got)
	}
	c.Pop()
}

func TestCtxScaleAndPopSafety(t *testing.T) {
	c := NewCtx(NewStore(), 1.5)
	if got := c.S(100); got != 150 {
		t.Fatalf("S(100)=%v want 150", got)
	}
	c.Pop() // 空路径 Pop 不得 panic
	if got := NewCtx(NewStore(), 0).Scale; got != 1 {
		t.Fatalf("scale<=0 should normalize to 1, got %v", got)
	}
}
