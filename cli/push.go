package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"wallet-sign/internal/chain/types"
	"wallet-sign/internal/rpc"
	"wallet-sign/internal/vapi"
)

// MpoolPushCmd 内存池推送命令
// 用于将已签名的消息推送到 Filecoin 内存池中
var MpoolPushCmd = &cli.Command{
	Name:  "push",
	Usage: "replace a message in the mempool",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "msg",
			Usage: "十六进制编码的已签名消息",
		},
	},
	Action: func(cctx *cli.Context) error {
		// 创建 Lotus API 客户端
		api := rpc.NewLotusApi()
		ctx := cctx.Context
		node := vapi.NewNode(ctx, api)

		// 获取消息参数
		msg := cctx.String("msg")
		// 解码十六进制消息
		buf, err := hex.DecodeString(msg)
		if err != nil {
			return err
		}
		var sig = new(types.SignedMessage)

		// 反序列化 CBOR 格式的签名消息
		if err := sig.UnmarshalCBOR(bytes.NewReader(buf)); err != nil {
			return err
		}

		// 推送消息到内存池
		msgCid, err := node.MpoolPush(sig)
		if err != nil {
			return xerrors.Errorf("failed to push new message to mempool: %w", err)
		}

		fmt.Println("new message cid: ", msgCid)
		return nil
	},
}
