package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/filecoin-project/go-address"
	"github.com/urfave/cli/v2"

	"wallet-sign/internal/chain/types"
	"wallet-sign/internal/rpc"
	"wallet-sign/internal/service"
	"wallet-sign/internal/ui/tablewriter"
	"wallet-sign/internal/vapi"
)

// ActorCmd 矿工管理命令
// 提供矿工信息查询、所有者变更、Worker 变更等功能
var ActorCmd = &cli.Command{
	Name:  "actor",
	Usage: "矿工管理",
	Subcommands: []*cli.Command{
		setOwner,      // 设置所有者
		setWorker,     // 提议更改 Worker
		confirmWorker, // 确认更改 Worker
		WithdrawCmd,   // 提现命令
		infoCmd,       // 查看矿工信息
	},
}

// infoCmd 查看矿工信息命令
// 显示矿工的 Owner、Worker、Control 地址及其余额和用途
var infoCmd = &cli.Command{
	Name:      "info",
	Usage:     "查看矿工信息",
	ArgsUsage: "[矿工ID]",
	Action: func(cctx *cli.Context) error {
		// 创建 Lotus API 客户端
		api := rpc.NewLotusApi()
		ctx := cctx.Context
		node := vapi.NewNode(ctx, api)

		// 获取矿工ID参数
		mid := cctx.Args().First()
		if mid == "" {
			return errors.New("请输入minerid")
		}

		// 解析矿工地址
		maddr, err := address.NewFromString(mid)
		if err != nil {
			return err
		}

		// 获取矿工信息
		mi, err := node.StateMinerInfo(maddr)
		if err != nil {
			return err
		}

		tw := tablewriter.New(
			tablewriter.Col("name"),
			tablewriter.Col("ID"),
			tablewriter.Col("key"),
			tablewriter.Col("use"),
			tablewriter.Col("balance"),
		)

		commit := map[address.Address]struct{}{}
		precommit := map[address.Address]struct{}{}
		terminate := map[address.Address]struct{}{}
		dealPublish := map[address.Address]struct{}{}
		post := map[address.Address]struct{}{}

		for _, ca := range mi.ControlAddresses {
			post[ca] = struct{}{}
		}

		printKey := func(name string, a address.Address) {
			var actor *types.Actor
			if actor, err = node.StateGetActor(a); err != nil {
				fmt.Printf("%s\t%s: error getting actor: %s\n", name, a, err)
				return
			}
			b := actor.Balance

			k := a
			if keyAddr, err := node.StateAccountKey(a); err == nil {
				k = keyAddr
			}
			kstr := k.String()

			bstr := types.FIL(b).String()
			switch {
			case b.LessThan(types.FromFil(10)):
				bstr = color.RedString(bstr)
			case b.LessThan(types.FromFil(50)):
				bstr = color.YellowString(bstr)
			default:
				bstr = color.GreenString(bstr)
			}

			var uses []string
			if a == mi.Worker {
				uses = append(uses, color.YellowString("other"))
			}
			if _, ok := post[a]; ok {
				uses = append(uses, color.GreenString("post"))
			}
			if _, ok := precommit[a]; ok {
				uses = append(uses, color.CyanString("precommit"))
			}
			if _, ok := commit[a]; ok {
				uses = append(uses, color.BlueString("commit"))
			}
			if _, ok := terminate[a]; ok {
				uses = append(uses, color.YellowString("terminate"))
			}
			if _, ok := dealPublish[a]; ok {
				uses = append(uses, color.MagentaString("deals"))
			}

			tw.Write(map[string]interface{}{
				"name":    name,
				"ID":      a,
				"key":     kstr,
				"use":     strings.Join(uses, " "),
				"balance": bstr,
			})
		}

		printKey("owner", mi.Owner)
		printKey("worker", mi.Worker)
		for i, ca := range mi.ControlAddresses {
			printKey(fmt.Sprintf("control-%d", i), ca)
		}

		return tw.Flush(os.Stdout)
	},
}
var setOwner = &cli.Command{
	Name:      "set-owner",
	Usage:     "Set owner address (this command should be invoked twice, first with the old owner as the senderAddress, and then with the new owner)",
	ArgsUsage: "[newOwnerAddress senderAddress]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "really-do-it",
			Usage: "Actually send transaction performing the action",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "minerid",
			Usage: "minerID",
		},
		&cli.Uint64Flag{
			Name:  "nonce",
			Usage: "specify the nonce to use",
			Value: 0,
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Bool("really-do-it") {
			fmt.Println("Pass --really-do-it to actually execute this action")
			return nil
		}

		if cctx.NArg() != 2 {
			return errors.New("参数数量错误")
		}

		minerid, err := address.NewFromString(cctx.String("minerid"))
		if minerid.String() != " " {
			return errors.New("minerid不能为空")
		}

		na, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		fa, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		client, err := service.NewClient()
		if err != nil {
			return err
		}

		data := &service.Payload{
			Type:      service.RequestTypeMinerChangeOwner,
			MinerID:   minerid,
			NewOwner:  na,
			FromOwner: fa,
		}
		client.Ex.Execute(data)

		return nil
	},
}

var setWorker = &cli.Command{
	Name:      "propose-change-worker",
	Usage:     "Propose a worker address change",
	ArgsUsage: "[address]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "really-do-it",
			Usage: "Actually send transaction performing the action",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "minerid",
			Usage: "minerID",
		},
		&cli.Uint64Flag{
			Name:  "nonce",
			Usage: "specify the nonce to use",
			Value: 0,
		},
	},
	Action: func(cctx *cli.Context) error {
		miner, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		na, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}
		if !cctx.Bool("really-do-it") {
			fmt.Println(cctx.App.Writer, "Pass --really-do-it to actually execute this action")
			return nil
		}
		client, err := service.NewClient()
		if err != nil {
			return err
		}

		data := &service.Payload{
			Type:            service.RequestTypeMinerChangeWorker,
			MinerID:         miner,
			NewWorker:       na,
			NewControlAddrs: nil,
		}
		client.Ex.Execute(data)

		return nil
	},
}

var confirmWorker = &cli.Command{
	Name:      "confirm-change-worker",
	Usage:     "Propose a worker address change",
	ArgsUsage: "[address]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "really-do-it",
			Usage: "Actually send transaction performing the action",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "minerid",
			Usage: "minerID",
		},
		&cli.Uint64Flag{
			Name:  "nonce",
			Usage: "specify the nonce to use",
			Value: 0,
		},
	},
	Action: func(cctx *cli.Context) error {
		miner, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		na, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		if !cctx.Bool("really-do-it") {
			fmt.Println("Pass --really-do-it to actually execute this action")
			return nil
		}

		client, err := service.NewClient()
		if err != nil {
			return err
		}

		data := &service.Payload{
			Type:      service.RequestTypeMinerConfirmWorker,
			MinerID:   miner,
			NewWorker: na,
		}
		client.Ex.Execute(data)

		return nil
	},
}
