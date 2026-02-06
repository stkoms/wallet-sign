package cli

import (
	"fmt"
	"wallet-sign/internal/chain/types"
	"wallet-sign/internal/service"

	"github.com/filecoin-project/go-address"
	"github.com/urfave/cli/v2"
)

// WithdrawCmd 矿工提现命令
// 用于从矿工账户中提取可用余额
var WithdrawCmd = &cli.Command{
	Name:      "withdraw",
	Usage:     "Send funds between accounts",
	ArgsUsage: "[targetAddress] [amount]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "minerId",
			Usage: "miner id",
		},
	},
	Action: func(cctx *cli.Context) error {
		// 解析矿工 ID
		miner, err := address.NewFromString(cctx.String("minerId"))
		if err != nil {
			return err
		}

		// 解析提现金额
		val, err := types.ParseFIL(cctx.Args().Get(1))
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
			Type:    service.RequestTypeMinerWithdraw,
			MinerID: miner,
			Amount:  val,
		}
		client.Ex.Execute(data)

		// 等待审批完成
		return nil
	},
}
