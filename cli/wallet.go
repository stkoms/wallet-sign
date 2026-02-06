package cli

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	appcfg "wallet-sign/internal/config"

	"github.com/filecoin-project/go-address"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"wallet-sign/internal/chain/types"
	"wallet-sign/internal/repository"
	"wallet-sign/internal/rpc"
	"wallet-sign/internal/ui/tablewriter"
	"wallet-sign/internal/vapi"
	"wallet-sign/internal/wallet"
)

type ctxKey string

const (
	CtxConfig ctxKey = "config"
)

// WalletCmd 钱包管理命令
// 提供密钥生成、导入、导出、列表、余额查询等功能
var WalletCmd = &cli.Command{
	Name:  "wallet",
	Usage: "钱包管理",

	Before: func(c *cli.Context) error {
		// 1. 加载配置
		cfg, err := appcfg.LoadConfig()
		if err != nil {
			return err
		}

		// 3. 注入到 Context（关键）
		c.Context = context.WithValue(c.Context, CtxConfig, cfg)

		return nil
	},

	Subcommands: []*cli.Command{
		walletNew,
		walletList,
		walletExport,
		walletImport,
		walletBalance,
		walletDelete,
	},
}

// walletNew 生成新密钥命令
// 支持 BLS 和 secp256k1 两种密钥类型
var walletNew = &cli.Command{
	Name:      "new",
	Usage:     "生成指定类型的新密钥",
	ArgsUsage: "[bls|secp256k1 (默认 secp256k1)]",
	Action: func(cctx *cli.Context) error {
		cfg := cctx.Context.Value(CtxConfig).(*appcfg.Config)
		// 打开数据库连接
		store, err := repository.OpenStore(cfg.DBDSN)
		if err != nil {
			return err
		}

		// 获取密钥类型参数
		t := cctx.Args().First()
		if t == "" {
			t = "secp256k1" // 默认使用 secp256k1
		}

		// 生成新密钥
		ki, addr, err := wallet.WalletNew(types.KeyType(t))
		if err != nil {
			return err
		}

		// 如果配置了数据库，则保存到数据库
		if err := storeWalletKeyIfConfigured(store, addr, ki); err != nil {
			return err
		}

		// 输出新生成的地址
		fmt.Println(addr)

		return nil
	},
}

// walletExport 导出密钥命令
// 将指定地址的私钥导出为十六进制格式
var walletExport = &cli.Command{
	Name:      "export",
	Usage:     "导出密钥",
	ArgsUsage: "[地址]",
	Action: func(cctx *cli.Context) error {
		// 检查是否提供了地址参数
		if !cctx.Args().Present() {
			return fmt.Errorf("must specify key to export")
		}

		// 解析地址
		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		// 获取数据库连接
		dsn := os.Getenv("DB_DSN")
		if dsn == "" {
			return fmt.Errorf("DB_DSN environment variable not set")
		}

		store, err := repository.OpenStore(dsn)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		// 从数据库获取密钥
		walletKey, err := store.GetWalletKey(addr.String())
		if err != nil {
			return fmt.Errorf("failed to get key: %w", err)
		}

		// 直接输出解密后的密钥数据（已经是 JSON 格式）
		fmt.Println(hex.EncodeToString(walletKey.EncryptedKey))
		return nil
	},
}

// walletImport 导入密钥命令（完整版）
// 支持多种格式：hex-lotus、json-lotus、gfc-json
var walletImport = &cli.Command{
	Name:      "import",
	Usage:     "导入密钥",
	ArgsUsage: "[<路径> (可选，如果省略则从标准输入读取)]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "format",
			Usage: "指定密钥输入格式",
			Value: "hex-lotus",
		},
		&cli.BoolFlag{
			Name:  "as-default",
			Usage: "将导入的密钥设置为默认密钥",
		},
	},
	Action: func(cctx *cli.Context) error {
		cfg := cctx.Context.Value(CtxConfig).(*appcfg.Config)
		// 打开数据库连接
		store, err := repository.OpenStore(cfg.DBDSN)
		if err != nil {
			return err
		}

		var inpdata []byte
		// 从标准输入或文件读取密钥数据
		if !cctx.Args().Present() || cctx.Args().First() == "-" {
			reader := bufio.NewReader(os.Stdin)
			indata, err := reader.ReadBytes('\n')
			if err != nil {
				return err
			}
			inpdata = indata
		} else {
			fdata, err := os.ReadFile(cctx.Args().First())
			if err != nil {
				return err
			}
			inpdata = fdata
		}

		var ki types.KeyInfo
		// 根据指定格式解析密钥
		switch cctx.String("format") {
		case "hex-lotus":
			// Lotus 十六进制格式
			data, err := hex.DecodeString(strings.TrimSpace(string(inpdata)))
			if err != nil {
				return err
			}
			if err := json.Unmarshal(data, &ki); err != nil {
				return err
			}
		case "json-lotus":
			// Lotus JSON 格式
			if err := json.Unmarshal(inpdata, &ki); err != nil {
				return err
			}
		case "gfc-json":
			// Go-Filecoin JSON 格式
			var f struct {
				KeyInfo []struct {
					PrivateKey []byte
					SigType    int
				}
			}
			if err := json.Unmarshal(inpdata, &f); err != nil {
				return xerrors.Errorf("failed to parse go-filecoin key: %s", err)
			}

			gk := f.KeyInfo[0]
			ki.PrivateKey = gk.PrivateKey
			switch gk.SigType {
			case 1:
				ki.Type = types.KTSecp256k1
			case 2:
				ki.Type = types.KTBLS
			default:
				return fmt.Errorf("unrecognized key type: %d", gk.SigType)
			}
		default:
			return fmt.Errorf("unrecognized format: %s", cctx.String("format"))
		}

		// 导入密钥
		addr, err := wallet.WalletImport(&ki)
		if err != nil {
			return err
		}

		// 如果配置了数据库，则保存到数据库
		if err := storeWalletKeyIfConfigured(store, addr, &ki); err != nil {
			return err
		}

		fmt.Printf("imported key %s successfully!\n", addr)

		return nil
	},
}

// walletList 列出钱包地址命令
// 显示所有钱包地址及其余额、Nonce 等信息
var walletList = &cli.Command{
	Name:  "list",
	Usage: "列出钱包地址",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {

		cfg := cctx.Context.Value(CtxConfig).(*appcfg.Config)
		// 打开数据库连接
		store, err := repository.OpenStore(cfg.DBDSN)
		if err != nil {
			return err
		}

		// 创建 Lotus API 客户端
		api := rpc.NewLotusApi()
		ctx := cctx.Context
		node := vapi.NewNode(ctx, api)

		// 获取本地钱包实例
		addrs, err := store.GetAllWalletAddresses()
		// 创建表格输出
		tw := tablewriter.New(
			tablewriter.Col("Address"),
			tablewriter.Col("ID"),
			tablewriter.Col("Amount"),
			tablewriter.Col("Market(Avail)"),
			tablewriter.Col("Market(Locked)"),
			tablewriter.Col("Nonce"),
			tablewriter.Col("Default"),
			tablewriter.NewLineCol("Error"))

		// 遍历所有地址，获取详细信息
		for _, addr := range addrs {
			Addr, err := address.NewFromString(addr.Address)
			if err != nil {
				return err
			}

			a, err := node.StateGetActor(Addr)
			if err != nil {
				if !strings.Contains(err.Error(), "actor not found") {
					tw.Write(map[string]interface{}{
						"Address": addr.Address,
						"Error":   err,
					})
					continue
				}

				// 如果 actor 不存在，使用零余额
				a = &types.Actor{
					Balance: types.NewInt(0),
				}
			}

			// 构建行数据
			row := map[string]interface{}{
				"Address": addr.Address,
				"Amount":  types.FIL(a.Balance),
				"Nonce":   a.Nonce,
			}

			// 如果需要显示 ID
			if cctx.Bool("id") {
				id, err := node.StateLookupID(Addr)
				if err != nil {
					row["ID"] = "n/a"
				} else {
					row["ID"] = id
				}
			}

			// 如果需要显示市场余额
			if cctx.Bool("market") {
				mbal, err := node.StateMarketBalance(Addr)
				if err == nil {
					row["Market(Avail)"] = types.FIL(types.BigSub(mbal.Escrow, mbal.Locked))
					row["Market(Locked)"] = types.FIL(mbal.Locked)
				}
			}

			tw.Write(row)
		}

		return tw.Flush(os.Stdout)
	},
}

// walletBalance 查询余额命令
// 查询指定地址的账户余额
var walletBalance = &cli.Command{
	Name:      "balance",
	Usage:     "查询地址余额",
	ArgsUsage: "[地址]",
	Action: func(cctx *cli.Context) error {
		// 创建 Lotus API 客户端
		client := rpc.NewLotusApi()
		ctx := cctx.Context
		node := vapi.NewNode(ctx, client)

		// 解析地址参数
		addr, err := address.NewFromString(os.Args[3])
		if err != nil {
			fmt.Printf("Invalid address: %v\n", err)
			os.Exit(1)
		}

		// 查询地址余额
		balance, err := node.WalletBalance(addr)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		// 输出地址和余额信息
		fmt.Printf("Address: %s\n", addr)
		fmt.Printf("Amount: %s\n", types.FIL(balance))

		// 转换为 FIL 单位并显示
		if err == nil {
			fmt.Printf("Amount: %s\n", types.FIL(balance))
		}

		return nil
	},
}

// storeWalletKeyIfConfigured 如果配置了数据库则保存钱包密钥
// 检查环境变量 DB_DSN，如果设置了则将密钥保存到数据库
// 参数：
//   - addr: 钱包地址
//   - ki: 密钥信息
//
// 返回：错误信息（如果有）
func storeWalletKeyIfConfigured(store *repository.Store, addr address.Address, ki *types.KeyInfo) error {
	// 检查是否配置了数据库连接

	// 保存密钥到数据库
	return store.SaveWalletKey(addr.String(), *ki)
}

// walletDelete 删除钱包密钥命令
var walletDelete = &cli.Command{
	Name:      "del",
	Usage:     "删除钱包密钥",
	ArgsUsage: "[地址]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "force",
			Usage: "强制删除，不需要确认",
		},
	},
	Action: func(cctx *cli.Context) error {
		// 检查是否提供了地址参数
		if !cctx.Args().Present() {
			return fmt.Errorf("请指定要删除的钱包地址")
		}

		// 解析地址
		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return fmt.Errorf("无效的地址: %w", err)
		}

		cfg := cctx.Context.Value(CtxConfig).(*appcfg.Config)
		// 打开数据库连接
		store, err := repository.OpenStore(cfg.DBDSN)
		if err != nil {
			return err
		}

		// 如果没有 --force 标志，请求确认
		if !cctx.Bool("force") {
			fmt.Printf("确定要删除钱包 %s 吗？此操作不可恢复！\n", addr)
			fmt.Print("输入 'yes' 确认: ")
			reader := bufio.NewReader(os.Stdin)
			confirm, _ := reader.ReadString('\n')
			if strings.TrimSpace(confirm) != "yes" {
				fmt.Println("已取消删除操作")
				return nil
			}
		}

		// 删除密钥
		if err := store.DeleteWalletKey(addr.String()); err != nil {
			return fmt.Errorf("删除钱包失败: %w", err)
		}

		fmt.Printf("已成功删除钱包 %s\n", addr)
		return nil
	},
}
