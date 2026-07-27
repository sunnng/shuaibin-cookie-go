# 矿山分类 UI 骨架设计

> 日期：2026-07-27  
> 范围：任务面板新增「矿山」分类与四个占位任务；**不**实现游戏流程、**不**注册调度器。

---

## 1. 背景与目标

任务列表目前只有「王国竞技场」（分类 `daily` → chip「日常」）。需要在面板露出矿山相关入口，便于后续逐个补逻辑，同时保持「未实现任务不进调度」的约定。

**目标：**

1. 任务页自动出现分类 chip「矿山」。
2. 分类下挂四个可开关的占位任务（仅「启用」字段）。
3. 启用态经 `ui.json` / `config.json` 可持久化，与竞技场同一套 Seed / ApplyAll。
4. `buildRuntime` **不**对这四个任务调用 `sched.Build`——勾选启用也不会跑。

**非目标：** `internal/game/mine*` 业务包、特征取色、状态机流程、状态上报。

---

## 2. 任务清单与命名

分类字符串：`Category: "矿山"`（中文；`categoryLabel` 原样返回；色条走 `candyCategoryColor` 默认灰）。

| 面板标题 | 英文名（文档/注释） | ID | EnabledKey | config 字段 |
|---|---|---|---|---|
| 矿山开采 | Ore Vein Mining | `ore_vein_mining` | `ore_vein_mining_enabled` | `tasks.oreVeinMining.enabled` |
| 矿山勘查 | Mine Venture | `mine_venture` | `mine_venture_enabled` | `tasks.mineVenture.enabled` |
| 矿山战斗 | Mine Battle | `mine_battle` | `mine_battle_enabled` | `tasks.mineBattle.enabled` |
| 解除洋菜冻 | Melted Agar Cubes | `melted_agar_cubes` | `melted_agar_cubes_enabled` | `tasks.meltedAgarCubes.enabled` |

每个任务：

- `Fields`：仅一个 `ui.Bool(<EnabledKey>, "启用", get, set)`。
- `Summary`：固定返回 `"未实现"`。
- 默认 `Enabled: false`。

---

## 3. 配置

在 `internal/config/static.go` 增加四个仅含 `Enabled` 的结构，挂到 `TaskConfig`：

```go
type OreVeinMining struct {
	Enabled bool `json:"enabled"`
}
// MineVenture / MineBattle / MeltedAgarCubes 同形

type TaskConfig struct {
	Arena            Arena            `json:"arena"`
	OreVeinMining    OreVeinMining    `json:"oreVeinMining"`
	MineVenture      MineVenture      `json:"mineVenture"`
	MineBattle       MineBattle       `json:"mineBattle"`
	MeltedAgarCubes  MeltedAgarCubes  `json:"meltedAgarCubes"`
}
```

`DefaultConfig()` 中四个 `Enabled` 均为 `false`。`config.json` 可不预先写入（缺键走默认）；可选在仓库 `config.json` 补空对象以示意，非必须。

---

## 4. UI 接线

- 新增 `ui_mine.go`：`mineTaskDescriptors(cfg *config.Config) []ui.Task`，返回上述四个描述符（闭包绑定对应 config 指针）。
- `main.go`：`tasks := append([]ui.Task{arenaTaskDescriptor(cfg)}, mineTaskDescriptors(cfg)...)`。
- **不**改 `buildRuntime` 的 `sched.Build` 列表。
- 单测：`ui_mine_test.go` 覆盖 Seed → 改值 → ApplyAll 回写四任务 `Enabled`，以及 Summary 文案。

面板预期：`[全部][日常][矿山]`；点「矿山」见四行，点进详情只有启用开关。

---

## 5. 后续扩展（本文不实现）

补某个矿山任务时：建 `internal/game/...` 包 → 填特征与状态机 → `sched.Build` → 按需加 Number/Text 字段。本骨架的 ID / EnabledKey / config 键保持稳定，避免设备 `ui.json` 键迁移。

---

## 6. 验收

1. `go test ./...` 通过（含 config 默认值与 mine 描述符往返）。
2. 真机/模拟器打开任务面板：可见「矿山」chip 与四条任务；开关可持久化。
3. 全部勾选启用后点开始：运行时行为与仅竞技场时一致（矿山任务不执行）。
