package cli

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/builtin"
	"github.com/urfave/cli/v2"

	"wallet-sign/internal/chain/types"
	"wallet-sign/internal/service"
)

// SendCmd 转账命令
// 用于在账户之间发送 FIL 代币
var SendCmd = &cli.Command{
	Name:      "send",
	Usage:     "在账户之间转账",
	ArgsUsage: "[目标地址] [金额]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "from",
			Usage: "指定发送方账户地址",
		},
		&cli.StringFlag{
			Name:  "gas-premium",
			Usage: "指定 Gas 溢价（单位：AttoFIL）",
			Value: "0",
		},
		&cli.StringFlag{
			Name:  "gas-feecap",
			Usage: "指定 Gas 费用上限（单位：AttoFIL）",
			Value: "0",
		},
		&cli.Int64Flag{
			Name:  "gas-limit",
			Usage: "指定 Gas 限制",
			Value: 0,
		},
		&cli.Uint64Flag{
			Name:  "method",
			Usage: "指定调用的方法编号",
			Value: uint64(builtin.MethodSend),
		},
		&cli.Uint64Flag{
			Name:  "nonce",
			Usage: "指定交易 nonce 值",
			Value: 0,
		},
	},
	Action: func(cctx *cli.Context) error {
		// 解析发送方地址
		fromAddr, err := address.NewFromString(cctx.String("from"))
		if err != nil {
			return err
		}

		// 解析接收方地址
		toAddr, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		// 解析转账金额
		val, err := types.ParseFIL(cctx.Args().Get(1))
		if err != nil {
			return fmt.Errorf("failed to parse amount: %w", err)
		}

		client, err := service.NewClient()
		if err != nil {
			return err
		}

		// 创建转账审批请求
		data := &service.Payload{
			Type:     service.RequestTypeTransfer,
			FromAddr: fromAddr,
			ToAddr:   toAddr,
			Amount:   val,
		}
		client.Ex.Execute(data)

		return nil
	},
}
