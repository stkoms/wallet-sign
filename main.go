package main

import (
	"os"
	cli2 "wallet-sign/cli"
	appcfg "wallet-sign/internal/config"
	"wallet-sign/internal/repository"
	"wallet-sign/lib/signlog"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
)

// logger 全局日志记录器
var log = logging.Logger("wallet-sign")

// main 程序入口函数
// 初始化 CLI 应用并启动命令行界面
func main() {
	// 设置全局日志级别为 INFO
	signlog.SetupLogLevels()

	// 加载 TOML 配置文件
	if err := appcfg.Load(); err != nil {
		log.Fatal(err)
		return
	}

	// 初始化加密密钥
	repository.InitEncryptionKey()

	// 加载配置
	cfg, err := appcfg.LoadConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	// 打开数据库连接并自动迁移表结构
	if _, err := repository.OpenStore(cfg.DBDSN); err != nil {
		return
	}

	// 创建 CLI 应用实例
	app := &cli.App{
		Name:    "lotus-sign",
		Usage:   "Lotus-sign 钱包签名工具，支持转账、提现、修改worker地址",
		Version: "1.0.0",

		Commands: cli2.All(),
	}

	// 运行 CLI 应用
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
