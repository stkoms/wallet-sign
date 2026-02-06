package config

import (
	"os"
	"path/filepath"
)

// LotusConfig 全局配置实例（从 TOML 文件加载）
var LotusConfig struct {
	Lotus    *Lotus    // Lotus 节点配置
	Security *Security // 安全配置
	Database *Database // 数据库配置
}

// Security 安全相关配置
type Security struct {
	Seed string // 加密种子
}

// Database 数据库配置
type Database struct {
	Path string // SQLite 数据库路径
}

// Config 应用程序运行时配置
type Config struct {
	DBDSN string // SQLite 数据库路径
}

// Lotus 节点连接配置
type Lotus struct {
	Host  string // Lotus 节点地址
	Token string // API 访问令牌
}

// LoadConfig 加载配置
// 优先使用配置文件，否则使用默认值
func LoadConfig() (*Config, error) {
	// 获取数据库路径
	var dbPath string
	if LotusConfig.Database != nil && LotusConfig.Database.Path != "" {
		dbPath = expandPath(LotusConfig.Database.Path)
	} else {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			dbPath = filepath.Join(homeDir, ".lotus-sign", "wallet.db")
		}
	}

	return &Config{
		DBDSN: dbPath,
	}, nil
}

// expandPath 展开路径中的 ~ 为用户主目录
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[1:])
		}
	}
	return path
}
