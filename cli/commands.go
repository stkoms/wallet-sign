package cli

import "github.com/urfave/cli/v2"

// All 返回所有可用的 CLI 命令列表
// 包括数据库管理、服务启动、转账、钱包管理等功能
func All() []*cli.Command {
	return []*cli.Command{
		SendCmd,           // 发送转账交易
		MpoolPushCmd,      // 推送消息到内存池
		WalletCmd,         // 钱包管理命令
		ActorCmd,          // 矿工相关操作
		WithdrawCmd,       // 矿工提现命令
		MarketWithdrawCmd, // 市场提现命令
	}
}
