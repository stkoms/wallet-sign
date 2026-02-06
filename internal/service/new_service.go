package service

import (
	"wallet-sign/internal/config"
	"wallet-sign/internal/repository"
)

type NewService struct {
	Ex *Executor
}

func NewClient() (*NewService, error) {
	// 加载配置文件
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	// 打开数据库连接
	store, err := repository.OpenStore(cfg.DBDSN)
	if err != nil {
		return nil, err
	}
	// 创建执行器
	executor := NewExecutor(store)

	return &NewService{Ex: executor}, nil
}
