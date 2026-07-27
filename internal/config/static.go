package config

type Arena struct {
	Enabled      bool `json:"enabled"`
	MaxBattles   *int `json:"maxBattles"`
	AutoBuyCount int  `json:"autoBuyCount"`
	// TrophyDiff：选对手奖杯区间半宽。只打奖杯在 [myTrophy-TrophyDiff, myTrophy+TrophyDiff] 内的对手。
	// 注意：0 表示"严格相等"（只打奖杯完全等于自己的对手），不是"不过滤"。
	TrophyDiff int `json:"trophyDiff"`
}

// OreVeinMining 矿山开采配置（映射 mining.Config；包内对空矿卡列表有默认兜底）。
type OreVeinMining struct {
	Enabled     bool     `json:"enabled"`
	IntervalSec int      `json:"intervalSec"` // 全忙后再次调度间隔秒
	OreCards    []string `json:"oreCards"`    // 选卡优先级（特征库 oreVeinCards 键名）
}

// MineVenture 矿山勘查配置（映射 survey.Config）。
type MineVenture struct {
	Enabled     bool `json:"enabled"`
	TargetFloor int  `json:"targetFloor"` // 目标层数
	FarGap      int  `json:"farGap"`      // 近距阈值：|目标-当前|<=farGap 时原地轮询
	OCRPollSec  int  `json:"ocrPollSec"`  // 近距轮询 OCR 间隔秒
	FarWaitSec  int  `json:"farWaitSec"`  // 远距回城等待秒
}

// MineBattle 矿山战斗配置（映射 battle.Config）。
type MineBattle struct {
	Enabled     bool     `json:"enabled"`
	IntervalSec int      `json:"intervalSec"` // 战斗检测间隔秒
	SoulStones  []string `json:"soulStones"`  // 目标灵魂石名称（特征库键名，中文）
}

// MeltedAgarCubes 解除洋菜冻配置（映射 jelly.Config）。
type MeltedAgarCubes struct {
	Enabled     bool `json:"enabled"`
	IntervalSec int  `json:"intervalSec"` // 完成后冷却间隔秒
}

// Square 布谷鸟广场配置（映射 square.Config，包内对 <=0 有兜底）。
type Square struct {
	Enabled          bool `json:"enabled"`
	DailyCap         int  `json:"dailyCap"`         // 每日奖励领取上限
	CheckIntervalSec int  `json:"checkIntervalSec"` // 一轮奖励结算所需的有效停留秒数（下限 60）
	ChunkSec         int  `json:"chunkSec"`         // 单次挂机等待的睡眠粒度秒数
}

// Starlight 梦幻繁星岛配置。
type Starlight struct {
	Enabled bool `json:"enabled"`
}

// SeasideMarket 海滩交易所配置（映射 market.Config）。
type SeasideMarket struct {
	Enabled          bool     `json:"enabled"`
	Items            []string `json:"items"`            // 购买清单（Stock 键名）
	RestockBufferSec int      `json:"restockBufferSec"` // 补货缓冲秒数
}

// BiscuitTarget 脆饼槽位目标规则（映射 biscuit.TargetRule）。
type BiscuitTarget struct {
	Enabled    bool    `json:"enabled"`
	Name       string  `json:"name"`
	MinPercent float64 `json:"minPercent"`
}

// BiscuitSumRule 脆饼总和规则（映射 biscuit.SumRule）。
type BiscuitSumRule struct {
	Enabled bool    `json:"enabled"`
	Name    string  `json:"name"`
	Count   int     `json:"count"`
	MinSum  float64 `json:"minSum"`
}

// Biscuit 洗脆饼词条配置（映射 biscuit.Config）。
type Biscuit struct {
	Enabled  bool             `json:"enabled"`
	MaxRolls int              `json:"maxRolls"` // 洗炼上限次数
	Targets  []BiscuitTarget  `json:"targets"`  // 固定 4 条槽位规则
	SumRules []BiscuitSumRule `json:"sumRules"` // 总和规则（可选）
}

type TaskConfig struct {
	Arena           Arena           `json:"arena"`
	OreVeinMining   OreVeinMining   `json:"oreVeinMining"`
	MineVenture     MineVenture     `json:"mineVenture"`
	MineBattle      MineBattle      `json:"mineBattle"`
	MeltedAgarCubes MeltedAgarCubes `json:"meltedAgarCubes"`
	Square          Square          `json:"square"`
	Starlight       Starlight       `json:"starlight"`
	SeasideMarket   SeasideMarket   `json:"seasideMarket"`
	Biscuit         Biscuit         `json:"biscuit"`
}

type Config struct {
	Tasks TaskConfig `json:"tasks"`
}

// DefaultConfig 默认值对齐 Lua config.lua 各 USER 段；internal/config 不 import
// 游戏包，main 包在 buildRuntime 里做字段映射（各游戏包 Config 对 <=0 有兜底，
// 这里统一给正数默认值）。
func DefaultConfig() *Config {
	return &Config{
		Tasks: TaskConfig{
			Arena: Arena{
				Enabled:      true,
				AutoBuyCount: 0,
				TrophyDiff:   0,
			},
			OreVeinMining: OreVeinMining{
				Enabled:     false,
				IntervalSec: 1200,
				OreCards: []string{
					"butterAmber", "amberFossil", "sugarOre",
					"purpleFossil", "emeraldFossil", "flourStone",
				},
			},
			MineVenture: MineVenture{
				Enabled:     true,
				TargetFloor: 6,
				FarGap:      2,
				OCRPollSec:  60,
				FarWaitSec:  600,
			},
			MineBattle: MineBattle{
				Enabled:     false,
				IntervalSec: 21600,
				SoulStones:  []string{"妖精王", "莓果", "雷神武将"},
			},
			MeltedAgarCubes: MeltedAgarCubes{
				Enabled:     false,
				IntervalSec: 3600,
			},
			Square: Square{
				Enabled:          true,
				DailyCap:         240,
				CheckIntervalSec: 60,
				ChunkSec:         10,
			},
			Starlight: Starlight{Enabled: false},
			SeasideMarket: SeasideMarket{
				Enabled: false,
				Items: []string{
					"灿烂的光之碎片",
					"十分钟加速券",
					"商品1_金紫",
					"商品2_蓝盒",
					"商品3_罗盘",
					"商品4_绿书",
					"商品5_卷轴",
				},
				RestockBufferSec: 30,
			},
			Biscuit: Biscuit{
				Enabled:  false,
				MaxRolls: 500,
				Targets: []BiscuitTarget{
					{Enabled: true, Name: "冷却时间", MinPercent: 5},
					{Enabled: true, Name: "会心", MinPercent: 6},
					{Enabled: false, Name: "", MinPercent: 0},
					{Enabled: false, Name: "", MinPercent: 0},
				},
				SumRules: []BiscuitSumRule{
					{Enabled: true, Name: "攻击力", Count: 2, MinSum: 11},
				},
			},
		},
	}
}
