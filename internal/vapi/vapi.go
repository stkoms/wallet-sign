package vapi

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"

	"wallet-sign/internal/chain/types"
	"wallet-sign/internal/rpc"
)

var log = logging.Logger("vapi")

const (
	// GasLimitOverestimation Gas 限制过度估算系数
	// 用于在估算的基础上增加 50% 的 Gas 限制以确保交易成功
	GasLimitOverestimation = 1.5
)

// Node Lotus API 客户端节点
// 封装了 RPC 客户端，提供与 Filecoin 网络交互的方法
type Node struct {
	*rpc.Client
	ctx context.Context
}

// NewNode 创建新的节点实例
// 使用给定的上下文和 RPC 客户端初始化节点
func NewNode(ctx context.Context, rpc *rpc.Client) *Node {
	log.Debugf("NewNode: creating new node instance")
	node := &Node{rpc, ctx}
	log.Debugf("NewNode: node instance created successfully")
	return node
}

// StateWaitMsg 等待消息被打包到区块中并返回消息查找结果
// 轮询链直到给定 CID 的消息被确认（3 个 tipset 的确认度）
// 如果消息执行失败（非零退出码）或 RPC 调用失败，返回错误
func (vapi Node) StateWaitMsg(msgCid cid.Cid) (*types.MsgLookup, error) {
	log.Debugf("StateWaitMsg: waiting for message with CID: %s", msgCid)
	var msgLookup types.MsgLookup
	err := vapi.Call(vapi.ctx, "StateWaitMsg", []interface{}{msgCid, 3}, &msgLookup)
	if err != nil {
		log.Errorf("StateWaitMsg: failed to wait for message: %v", err)
		return nil, fmt.Errorf("failed to wait for message: %w", err)
	}

	if msgLookup.Receipt.ExitCode != 0 {
		log.Errorf("StateWaitMsg: message execution failed with exit code: %d", msgLookup.Receipt.ExitCode)
		return nil, fmt.Errorf("withdrawal failed with exit code: %d", msgLookup.Receipt.ExitCode)
	}

	log.Debugf("StateWaitMsg: message confirmed successfully, exit code: %d", msgLookup.Receipt.ExitCode)
	return &msgLookup, nil
}

// MpoolPush 将已签名的消息推送到内存池并返回其 CID
// 消息将被广播到网络并最终被打包到区块中
// 成功时返回消息 CID，失败时返回错误
func (vapi Node) MpoolPush(signedMsg *types.SignedMessage) (cid.Cid, error) {
	log.Debugf("MpoolPush: pushing signed message to mempool")
	var msgCid cid.Cid
	err := vapi.Call(vapi.ctx, "MpoolPush", []interface{}{signedMsg}, &msgCid)
	if err != nil {
		log.Errorf("MpoolPush: failed to push message: %v", err)
		return cid.Undef, fmt.Errorf("failed to push message: %w", err)
	}

	log.Debugf("MpoolPush: message pushed successfully, CID: %s", msgCid)
	fmt.Printf("Message CID: %s\n", msgCid)
	fmt.Println("Waiting for message confirmation...")
	return msgCid, nil
}

// GasEstimateGasLimit 估算消息的 Gas 参数
// 更新消息的 GasLimit、GasFeeCap 和 GasPremium 字段为估算值
// 成功时返回更新后的消息，失败时返回错误
func (vapi Node) GasEstimateGasLimit(msg *types.Message) (*types.Message, error) {
	log.Debugf("GasEstimateGasLimit: estimating gas for message")
	var gasEstimate types.Message
	err := vapi.Call(vapi.ctx, "GasEstimateMessageGas", []interface{}{msg, nil, nil}, &gasEstimate)
	if err != nil {
		log.Errorf("GasEstimateGasLimit: failed to estimate gas: %v", err)
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	msg.GasLimit = int64(float64(gasEstimate.GasLimit) * GasLimitOverestimation)
	msg.GasFeeCap = gasEstimate.GasFeeCap
	msg.GasPremium = gasEstimate.GasPremium

	log.Debugf("GasEstimateGasLimit: gas estimated successfully, GasLimit: %d, GasFeeCap: %s, GasPremium: %s", msg.GasLimit, msg.GasFeeCap, msg.GasPremium)
	return msg, nil
}

// StateMinerInfo 从链状态中检索矿工信息
// 返回矿工的所有者、Worker 和控制地址等信息
func (vapi Node) StateMinerInfo(minerAddr address.Address) (*types.MinerInfo, error) {
	log.Debugf("StateMinerInfo: getting miner info for address: %s", minerAddr)
	var minerInfo types.MinerInfo
	err := vapi.Call(vapi.ctx, "StateMinerInfo", []interface{}{minerAddr, nil}, &minerInfo)
	if err != nil {
		log.Errorf("StateMinerInfo: failed to get miner info: %v", err)
		return nil, fmt.Errorf("failed to get miner info: %w", err)
	}
	log.Debugf("StateMinerInfo: miner info retrieved successfully for address: %s, owner: %s, worker: %s", minerAddr, minerInfo.Owner, minerInfo.Worker)
	return &minerInfo, nil
}

// StateMinerAvailableBalance 检索矿工的可用余额
// 返回 types.BigInt 类型的余额，查询失败时返回错误
func (vapi Node) StateMinerAvailableBalance(minerAddr address.Address) (types.BigInt, error) {
	log.Debugf("StateMinerAvailableBalance: getting available balance for miner: %s", minerAddr)
	var availableBalance types.BigInt
	err := vapi.Call(vapi.ctx, "StateMinerAvailableBalance", []interface{}{minerAddr, nil}, &availableBalance)
	if err != nil {
		log.Errorf("StateMinerAvailableBalance: failed to get available balance: %v", err)
		return types.BigInt{}, fmt.Errorf("failed to get available balance: %w", err)
	}
	log.Debugf("StateMinerAvailableBalance: available balance retrieved successfully for miner: %s, balance: %s", minerAddr, availableBalance)
	return availableBalance, nil
}

// MpoolGetNonce 从内存池中检索给定地址的下一个 nonce
// 用于构造需要从该地址发送的新消息
func (vapi Node) MpoolGetNonce(address address.Address) (uint64, error) {
	log.Debugf("MpoolGetNonce: getting nonce for address: %s", address)
	var nonce uint64
	err := vapi.Call(vapi.ctx, "MpoolGetNonce", []interface{}{address}, &nonce)
	if err != nil {
		log.Errorf("MpoolGetNonce: failed to get nonce: %v", err)
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}
	log.Debugf("MpoolGetNonce: nonce retrieved successfully for address: %s, nonce: %d", address, nonce)
	return nonce, nil
}

// WalletBalance 检索钱包地址的余额
// 返回 attoFIL 单位的余额字符串，查询失败时返回错误
func (vapi Node) WalletBalance(address address.Address) (types.BigInt, error) {
	log.Debugf("WalletBalance: getting balance for address: %s", address)
	var balance types.BigInt
	err := vapi.Call(vapi.ctx, "WalletBalance", []interface{}{address}, &balance)
	if err != nil {
		log.Errorf("WalletBalance: failed to get balance: %v", err)
		return types.BigInt{}, fmt.Errorf("failed to get balance: %w", err)
	}
	log.Debugf("WalletBalance: balance retrieved successfully for address: %s, balance: %s", address, balance)
	return balance, nil
}

// StateAccountKey 将 ID 地址 (f0) 解析为对应的密钥地址 (f1/f3)
// 当你有 actor 的 ID 地址但需要密钥地址来签名消息时很有用
func (vapi Node) StateAccountKey(addr address.Address) (address.Address, error) {
	log.Debugf("StateAccountKey: resolving account key for address: %s", addr)
	var keyAddr address.Address
	err := vapi.Call(vapi.ctx, "StateAccountKey", []interface{}{addr, nil}, &keyAddr)
	if err != nil {
		log.Errorf("StateAccountKey: failed to resolve account key: %v", err)
		return address.Undef, fmt.Errorf("failed to resolve account key: %w", err)
	}
	log.Debugf("StateAccountKey: account key resolved successfully, input: %s, key address: %s", addr, keyAddr)
	return keyAddr, nil
}

// StateMarketBalance 检索地址的市场余额
// 返回存储市场中的托管（可用）余额和锁定余额
func (vapi Node) StateMarketBalance(addr address.Address) (*types.MarketBalance, error) {
	log.Debugf("StateMarketBalance: getting market balance for address: %s", addr)
	var balance types.MarketBalance
	err := vapi.Call(vapi.ctx, "StateMarketBalance", []interface{}{addr, nil}, &balance)
	if err != nil {
		log.Errorf("StateMarketBalance: failed to get market balance: %v", err)
		return nil, fmt.Errorf("failed to get market balance: %w", err)
	}
	log.Debugf("StateMarketBalance: market balance retrieved successfully for address: %s, escrow: %s, locked: %s", addr, balance.Escrow, balance.Locked)
	return &balance, nil
}

// StateLookupID 查找给定地址 (f1/f3) 的 ID 地址 (f0)
// 这是 StateAccountKey 的反向操作 - 将密钥地址转换为 ID 地址
func (vapi Node) StateLookupID(addr address.Address) (address.Address, error) {
	log.Debugf("StateLookupID: looking up ID address for: %s", addr)
	var idAddr address.Address
	err := vapi.Call(vapi.ctx, "StateLookupID", []interface{}{addr, nil}, &idAddr)
	if err != nil {
		log.Errorf("StateLookupID: failed to lookup ID address: %v", err)
		return address.Undef, fmt.Errorf("failed to lookup ID address: %w", err)
	}
	log.Debugf("StateLookupID: ID address lookup successful, input: %s, ID address: %s", addr, idAddr)
	return idAddr, nil
}

// StateGetActor 检索给定地址的 actor 状态
// 返回 actor 的信息，包括代码、头部、nonce 和余额
func (vapi Node) StateGetActor(addr address.Address) (*types.Actor, error) {
	log.Debugf("StateGetActor: getting actor state for address: %s", addr)
	var actor types.Actor
	err := vapi.Call(vapi.ctx, "StateGetActor", []interface{}{addr, nil}, &actor)
	if err != nil {
		log.Errorf("StateGetActor: failed to get actor: %v", err)
		return nil, fmt.Errorf("failed to get actor: %w", err)
	}
	log.Debugf("StateGetActor: actor state retrieved successfully for address: %s, nonce: %d, balance: %s", addr, actor.Nonce, actor.Balance)
	return &actor, nil
}

// GasEstimateGasPremium 估算消息在 nblocksincl 个区块内被打包所需的 Gas 溢价
// nblocksincl: 目标打包区块数（越低 = 溢价越高）
// sender: 发送者地址
// gaslimit: 消息的 Gas 限制
func (vapi Node) GasEstimateGasPremium(nblocksincl uint64, sender address.Address, gaslimit int64) (types.BigInt, error) {
	log.Debugf("GasEstimateGasPremium: estimating gas premium for sender: %s, nblocksincl: %d, gaslimit: %d", sender, nblocksincl, gaslimit)
	var premium types.BigInt
	err := vapi.Call(vapi.ctx, "GasEstimateGasPremium", []interface{}{nblocksincl, sender, gaslimit, nil}, &premium)
	if err != nil {
		log.Errorf("GasEstimateGasPremium: failed to estimate gas premium: %v", err)
		return types.BigInt{}, fmt.Errorf("failed to estimate gas premium: %w", err)
	}
	log.Debugf("GasEstimateGasPremium: gas premium estimated successfully, premium: %s", premium)
	return premium, nil
}

// ChainHead 返回链的当前头部
func (vapi Node) ChainHead() (*types.TipSet, error) {
	log.Debugf("ChainHead: getting current chain head")
	var tipset types.TipSet
	err := vapi.Call(vapi.ctx, "ChainHead", []interface{}{}, &tipset)
	if err != nil {
		log.Errorf("ChainHead: failed to get chain head: %v", err)
		return nil, fmt.Errorf("failed to get chain head: %w", err)
	}
	log.Debugf("ChainHead: chain head retrieved successfully, height: %d", tipset.Height())
	return &tipset, nil
}

// GasEstimateFeeCap 估算消息的费用上限
func (vapi Node) GasEstimateFeeCap(msg *types.Message, maxqueueblks int64) (types.BigInt, error) {
	log.Debugf("GasEstimateFeeCap: estimating fee cap, maxqueueblks: %d", maxqueueblks)
	var feecap types.BigInt
	err := vapi.Call(vapi.ctx, "GasEstimateFeeCap", []interface{}{msg, maxqueueblks, nil}, &feecap)
	if err != nil {
		log.Errorf("GasEstimateFeeCap: failed to estimate fee cap: %v", err)
		return types.BigInt{}, fmt.Errorf("failed to estimate fee cap: %w", err)
	}
	log.Debugf("GasEstimateFeeCap: fee cap estimated successfully, feecap: %s", feecap)
	return feecap, nil
}
