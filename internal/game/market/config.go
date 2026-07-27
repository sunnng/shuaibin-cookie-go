package market

// Config 海滩交易所任务配置（Lua config.lua USER.seasideMarket）。
// 由编排者从面板/配置文件填充后传入 NewTask；默认值见 DefaultConfig。
type Config struct {
	Enabled          bool     // 是否启用（Lua 默认 false）
	Items            []string // 购买清单（Stock 键名）；空 = 全部已配置商品
	RestockBufferSec int      // 补货缓冲秒数：下一次运行 = 补货倒计时 + 该缓冲
}

// DefaultConfig 与 Lua config.lua USER.seasideMarket 默认值一致。
func DefaultConfig() *Config {
	return &Config{
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
	}
}

const defaultBufferSec = 30

// BufferSec 补货缓冲秒数；nil 或负值回退 30（对齐 Lua bufferSec()）。
func (c *Config) BufferSec() int {
	if c == nil || c.RestockBufferSec < 0 {
		return defaultBufferSec
	}
	return c.RestockBufferSec
}
