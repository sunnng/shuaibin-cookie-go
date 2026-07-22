package ui

// Shell 框架运行时实例（CONTEXT.md 术语见 ADR-0002）：持有全部 UI 状态
// （面板可见性、最小化、自动暂停标记），无包级可变状态；桌面测试可起多个
// 实例互不干扰。绘制层（Phase 2）以它为状态后端。
// Shell 只在 UI goroutine 使用，非并发安全。
type Shell struct {
	opts ShellOptions

	store *Store
	ctrl  *ScriptController

	panelOpen  bool
	minimized  bool
	autoPaused bool
}

// NewShell 构造 Shell 并归一默认值。OpenPanelOnStart 时面板初始打开
// （此时脚本必为空闲，不触发自动暂停）。
func NewShell(opts ShellOptions) *Shell {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.Controller == nil {
		opts.Controller = NewScriptController(ScriptHooks{})
	}
	if opts.Theme == (Theme{}) {
		opts.Theme = DefaultTheme()
	}
	if opts.BaseWidth <= 0 {
		opts.BaseWidth = DefaultBaseWidth
	}
	if opts.BaseHeight <= 0 {
		opts.BaseHeight = DefaultBaseHeight
	}
	return &Shell{
		opts:      opts,
		store:     opts.Store,
		ctrl:      opts.Controller,
		panelOpen: opts.OpenPanelOnStart,
	}
}

func (s *Shell) Store() *Store    { return s.store }
func (s *Shell) Theme() Theme     { return s.opts.Theme }
func (s *Shell) Tasks() []Task    { return s.opts.Tasks }
func (s *Shell) Nav() []NavEntry  { return s.opts.Nav }
func (s *Shell) PanelOpen() bool  { return s.panelOpen }
func (s *Shell) Minimized() bool  { return s.minimized }
func (s *Shell) AutoPaused() bool { return s.autoPaused }

func (s *Shell) Title() string         { return s.opts.Title }
func (s *Shell) ConfigPath() string    { return s.opts.ConfigPath }
func (s *Shell) DataStorePath() string { return s.opts.DataStorePath }

// BaseSize 返回基准分辨率（NewShell 已归一默认值 1600×900）。
func (s *Shell) BaseSize() (w, h int) { return s.opts.BaseWidth, s.opts.BaseHeight }

// ScriptState 代理脚本的当前生命周期状态。
func (s *Shell) ScriptState() ScriptState { return s.ctrl.State() }

// Exit 停止脚本并触发退出钩子（灵动岛「关闭」的语义）。
func (s *Shell) Exit() { s.ctrl.Exit() }

func (s *Shell) ToggleMinimized() { s.minimized = !s.minimized }

// Seed 用任务字段默认值与面板偏好默认值填充 Store（仅填缺失键）。
// 应用通常在首帧（LoadConfig 之后）调用。
func (s *Shell) Seed() {
	SeedAll(s.store, s.opts.Tasks)
	navIDs := make([]string, 0, len(s.opts.Nav))
	for _, n := range s.opts.Nav {
		navIDs = append(navIDs, n.ID)
	}
	SeedPanelDefaults(s.store, s.opts.Tasks, navIDs)
}

// Apply 把 Store 中的配置写回应用配置。应用在 ScriptHooks.OnStart 内调用。
func (s *Shell) Apply() {
	ApplyAll(s.store, s.opts.Tasks)
}

// OpenPanel 打开配置面板；脚本运行中时自动暂停（遮挡策略：面板遮挡画面
// 期间不识别）。重开总是展开完整面板（复位最小化，与旧 internal/ui 行为一致）。
func (s *Shell) OpenPanel() {
	s.panelOpen = true
	s.minimized = false
	if s.ctrl.State() == StateRunning {
		s.ctrl.Pause()
		s.autoPaused = true
	}
}

// ClosePanel 关闭配置面板；若为自动暂停则恢复运行。手动暂停/恢复后
// （PauseResume 已清除 autoPaused）关闭面板不触碰控制器。
func (s *Shell) ClosePanel() {
	s.panelOpen = false
	if s.autoPaused {
		s.ctrl.Resume()
		s.autoPaused = false
	}
}

// PauseResume 手动暂停/继续；清除自动暂停标记（手动操作优先于遮挡策略）。
func (s *Shell) PauseResume() {
	switch s.ctrl.State() {
	case StateRunning:
		s.ctrl.Pause()
	case StatePaused:
		s.ctrl.Resume()
	}
	s.autoPaused = false
}

// StartStop 启动或停止脚本。启动时先把配置落盘（ConfigPath 非空），
// 若面板仍打开则启动后立即自动暂停。
func (s *Shell) StartStop() error {
	if s.ctrl.State() == StateIdle {
		if s.opts.ConfigPath != "" {
			if err := s.store.SaveConfig(s.opts.ConfigPath); err != nil {
				return err
			}
		}
		s.ctrl.Start()
		if s.panelOpen && s.ctrl.State() == StateRunning {
			s.ctrl.Pause()
			s.autoPaused = true
		}
		return nil
	}
	s.autoPaused = false
	s.ctrl.Stop()
	return nil
}

// StatusText 运行中且配置了状态来源时返回任务状态文本，否则空串
// （UI 回退默认文案）。
func (s *Shell) StatusText() string {
	if s.opts.Status == nil || s.ctrl.State() != StateRunning {
		return ""
	}
	return s.opts.Status.Text()
}
