package cli

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/urfave/cli/v2"

	"wallet-sign/internal/chain/types"
	"wallet-sign/internal/service"
)

// MarketWithdrawCmd 存储市场提现命令
// 用于从 Filecoin 存储市场中提取资金
var MarketWithdrawCmd = &cli.Command{
	Name:      "market-withdraw",
	Usage:     "Withdraw funds from the storage market",
	ArgsUsage: "[address] [amount]",
	Action: func(cctx *cli.Context) error {
		// 检查参数数量
		if cctx.NArg() < 2 {
			return fmt.Errorf("address and amount are required")
		}

		// 解析地址参数
		addr, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		// 解析金额参数
		amount, err := types.ParseFIL(cctx.Args().Get(1))
		if err != nil {
			return fmt.Errorf("failed to parse amount: %w", err)
		}

		// 创建审批客户端
		client, err := service.NewClient()
		if err != nil {
			return err
		}

		// 创建矿工提现请求
		data := &service.Payload{
			Type:    service.RequestTypeMarketWithdraw,
			MinerID: addr,
			Amount:  amount,
		}
		client.Ex.Execute(data)

		return nil
	},
}
