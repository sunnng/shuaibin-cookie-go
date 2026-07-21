# UI 框架采用函数组件模型（immediate-mode 上的 React 心智）

框架的 UI 构造单元是**函数组件**：每帧执行的函数，唯一输入为一个 Props 结构体（数据 + 回调字段，如 `ButtonProps{Label, Kind, OnClick}`），组件自由组合嵌套。组件状态由框架托管，以「组件嵌套路径 + 显式键」寻址（`ui.State[T](ctx, key, initial)`），同一组件多处使用各自隔离。`Form` 组件消费字段描述符列表完成自动排版与绑定回写（ADR-0002 的"描述符自动渲染"由此组件承担）；自定义 section 即应用自写的组件组合。v1 组件清单以迁移对等为界：Button（变体）、Input、NumberInput、MultilineInput、Checkbox、Dropdown、Form、Tabs/TabItem、Collapsible、Image、Text + 布局件 Row/Column/ScrollArea——覆盖现有 `widgets_android.go` 全部能力，多一个不做。灵动岛与配置面板自身也用该组件 kit 构建（自举验证）。

## Considered Options

- **保留式组件树 + diff**：被否。与 immediate-mode 本质冲突——imgui 没有"更新已存在控件"的概念，模拟需自绘或厚适配层，工作量与风险爆炸。
- **仅组件库、调用方自管状态**：被否。输入框草稿、折叠态等每个应用重复造，且与"React 组件式"的目标不符。
- **调用位次 hooks**：被否。要求"不得条件调用"的隐性规则，immediate-mode 下极易错位；显式键规则简单且条件渲染安全。
- **函数选项 / imgui 返回值风格**：被否。前者每组件一套选项机制、约束弱；后者对表单等复杂组件表达不下，且最不 React。
- **大而全组件库（Table/Modal/Toast…一次到位）**：被否。无消费者的 API 未经场景检验必返工（YAGNI）。

## Consequences

- 组件 API = 每组件一个 Props 结构体；新增可选字段向后兼容。
- 状态规则：同一组件实例内键唯一；条件渲染自由，无 hooks 式调用约束。
- 框架自身 UI（岛、面板）是组件 kit 的第一消费者，API 缺陷在框架内部先暴露。
- 术语遵守 CONTEXT.md「组件」「组件状态」条目。
