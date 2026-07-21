package ui

// StatusSource 任务状态文本来源（ADR-0002 窄接口）：应用的 status.Reporter
// 天然实现它；框架不拥有该机制，游戏代码因此无需 import 框架。
type StatusSource interface {
	Text() string
}

// NavEntry 面板左栏导航条目（ADR-0002 导航全描述符化）。Render 为绘制层
// 函数（Phase 2 起使用）；任务列表页是框架提供的可复用组件，应用把它作为
// 一个条目挂载。
type NavEntry struct {
	ID     string
	Title  string
	Render func(*Ctx)
}

// ShellOptions Shell 的全部外部输入：描述符、持久化路径、生命周期钩子。
// 零值字段取默认：Store→NewStore、Theme→DefaultTheme、BaseWidth/Height→1600×900。
type ShellOptions struct {
	Title      string
	Tasks      []Task
	Nav        []NavEntry
	Store      *Store
	Controller *ScriptController
	// Status 任务状态来源；非 nil 且脚本运行中时，灵动岛展示该文本。
	Status StatusSource
	Theme  Theme
	// ConfigPath UI 配置持久化路径（启动脚本时落盘）；空则不落盘。
	ConfigPath string
	// DataStorePath 业务 KV 路径；清除缓存时一并删除。
	DataStorePath    string
	OpenPanelOnStart bool
	BaseWidth        int
	BaseHeight       int
}
