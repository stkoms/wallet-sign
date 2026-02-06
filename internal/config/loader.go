package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

const (
	defaultConfigPath = "configs/config.toml" // 默认配置文件路径
	legacyConfigPath  = "config.toml"         // 旧版配置文件路径
)

// Load 加载配置文件
// 自动查找并解析 TOML 格式的配置文件
func Load() error {
	path := ResolveConfigPath()
	if path == "" {
		return nil
	}
	_, err := toml.DecodeFile(path, &LotusConfig)
	return err
}

// ResolveConfigPath 解析配置文件路径
// 按优先级查找配置文件：先查找默认路径，再查找旧版路径
func ResolveConfigPath() string {
	if fileExists(defaultConfigPath) {
		return defaultConfigPath
	}
	if fileExists(legacyConfigPath) {
		return legacyConfigPath
	}
	return ""
}

// fileExists 检查文件是否存在
// 返回 true 表示文件存在且不是目录
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
