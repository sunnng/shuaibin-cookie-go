# 术语表约束代码标识符

CONTEXT.md 的术语不仅约束文档、讨论与 UI 文案，也约束代码标识符。存量违规一次性收敛：`ui.Module*` → Task 系（`BuiltinTasks` / `TaskCategory` / `FindTask`，`modules.go` → `tasks.go`）、`config.ModuleConfig` → `TaskConfig` 且 `config.json` 键 `"modules"` → `"tasks"`、`SessionController` / `SessionHooks` → `ScriptController` / `ScriptHooks`、`arena.Session` → `arena.State`（`session.go` → `state.go`，与 `production.State` 对齐）。文档（CLAUDE.md、AGENTS.md、docs/开发手册.md）同步更新。

## Considered Options

- **只管言语不管代码**：被否。代码里 `BuiltinModules`、`Session` 与术语表的 _Avoid_ 直接冲突，混用词表正是术语表要消灭的东西。
- **渐进收敛**：被否。代码量小、测试全，且生产任务尚未铺开，一次改完最便宜。
- **JSON 键保留 `"modules"`**：被否。`config.json` 只是默认种子（设备真源是 ui.json），统一优于兼容。

## 豁免规则

_Avoid_ 中标注「实现 / 实现细节 / 实现用词」的词（runtime、Store、Shell）可继续作机制名出现在代码中；旧称与易混淆词（Module、Session）全面禁绝。

## Consequences

新增标识符前先查 CONTEXT.md 的 _Avoid_ 列表；术语表新增/修订时同步检查代码违规。
