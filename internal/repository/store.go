package repository

import (
	"os"
	"path/filepath"
	"wallet-sign/internal/models"

	logging "github.com/ipfs/go-log/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var log = logging.Logger("repository")

// Store 数据存储结构
// 封装了 GORM 数据库连接，提供数据访问功能
type Store struct {
	DB *gorm.DB // GORM 数据库实例
}

// OpenStore 打开数据库存储
// 使用 SQLite 数据库，自动创建数据库文件
// 参数：
//   - dbPath: SQLite 数据库文件路径
//
// 返回：Store 实例或错误
func OpenStore(dbPath string) (*Store, error) {
	log.Debug("OpenStore: opening SQLite database connection")

	// 如果路径为空，使用默认路径
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Errorf("OpenStore: failed to get home directory: %v", err)
			return nil, err
		}
		dbPath = filepath.Join(homeDir, ".lotus-sign", "wallet.db")
	}

	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Errorf("OpenStore: failed to create directory %s: %v", dir, err)
		return nil, err
	}

	gormLogger := logger.Default.LogMode(logger.Silent)

	// 打开 SQLite 数据库连接
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Errorf("OpenStore: failed to open database: %v", err)
		return nil, err
	}

	// 自动迁移所有数据表
	if err = db.AutoMigrate(
		&models.WalletKey{},
	); err != nil {
		log.Errorf("OpenStore: auto migration failed: %v", err)
		return nil, err
	}

	log.Debugf("OpenStore: SQLite database opened successfully at %s", dbPath)
	return &Store{DB: db}, nil
}
