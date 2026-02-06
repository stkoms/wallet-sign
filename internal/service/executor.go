package service

import (
	"context"
	"fmt"
	"wallet-sign/internal/repository"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	builtintypes "github.com/filecoin-project/go-state-types/builtin"
	markettypes "github.com/filecoin-project/go-state-types/builtin/v9/market"
	minertypes "github.com/filecoin-project/go-state-types/builtin/v9/miner"
	logging "github.com/ipfs/go-log/v2"

	"wallet-sign/internal/chain/actors"
	"wallet-sign/internal/chain/types"
	"wallet-sign/internal/rpc"
	"wallet-sign/internal/vapi"
	"wallet-sign/internal/wallet"
)

var log = logging.Logger("executor")

type Executor struct {
	store *repository.Store
	node  *vapi.Node
}

func NewExecutor(store *repository.Store) *Executor {
	log.Info("NewExecutor: creating new executor instance")
	client := rpc.NewLotusApi()
	node := vapi.NewNode(contextBackground(), client)
	return &Executor{store: store, node: node}
}

func (e *Executor) Execute(req *Payload) {

	err := e.executeRequest(req)
	if err != nil {
		log.Errorf("Execute: failed to execute request #%d: %v", req.Type, err)
	}

	log.Infof("Execute: request #%d completed successfully", req.Type)

}

func (e *Executor) executeRequest(req *Payload) error {
	switch req.Type {
	case RequestTypeTransfer:
		var payload TransferPayload
		payload.From = req.FromAddr
		payload.To = req.ToAddr
		payload.Amount = req.Amount

		return e.transfer(payload)
	case RequestTypeMinerWithdraw:
		var payload MinerWithdrawPayload
		payload.MinerID = req.MinerID
		payload.Amount = req.Amount
		return e.minerWithdraw(payload)
	case RequestTypeMarketWithdraw:
		var payload MarketWithdrawPayload
		payload.Address = req.MinerID
		payload.Amount = req.Amount
		return e.marketWithdraw(payload)
	case RequestTypeBatchTransfer:
		var payload BatchTransferPayload
		payload.Items = req.Items
		return e.batchTransfer(payload)

	default:
		return fmt.Errorf("unsupported request type: %s", req.Type)
	}
}

func (e *Executor) transfer(p TransferPayload) error {

	fromAddr := p.From
	toAddr := p.To
	val := p.Amount
	nonce, err := e.node.MpoolGetNonce(p.From)
	if err != nil {
		log.Errorf("transfer: failed to get nonce for %s: %v", fromAddr, err)
		return err
	}

	msg := &types.Message{
		Version:    0,
		To:         toAddr,
		From:       fromAddr,
		Nonce:      nonce,
		Value:      types.BigInt(val),
		GasLimit:   0,
		GasFeeCap:  abi.NewTokenAmount(0),
		GasPremium: abi.NewTokenAmount(0),
		Method:     builtintypes.MethodSend,
	}
	if err := wallet.SetGas(e.node, msg); err != nil {
		log.Errorf("transfer: failed to set gas: %v", err)
		return err
	}

	hasKey, err := wallet.WalletHas(e.store, fromAddr)
	if err != nil {
		log.Errorf("transfer: failed to check key for %s: %v", fromAddr, err)
		return err
	}
	if !hasKey {
		log.Errorf("transfer: wallet does not have key for %s", fromAddr)
		return fmt.Errorf("wallet does not have key for %s", fromAddr)
	}

	sig, err := wallet.WalletSign(e.store, fromAddr, msg.Cid().Bytes())
	if err != nil {
		log.Errorf("transfer: failed to sign message: %v", err)
		return err
	}
	signed := &types.SignedMessage{Message: *msg, Signature: *sig}

	log.Infof("transfer: pushing message to mempool")
	msgCid, err := e.node.MpoolPush(signed)
	if err != nil {
		log.Errorf("transfer: failed to push message: %v", err)
		return err
	}

	log.Infof("transfer: waiting for message %s", msgCid)
	_, err = e.node.StateWaitMsg(msgCid)
	if err != nil {
		log.Errorf("transfer: message %s failed: %v", msgCid, err)
		return err
	}

	log.Infof("transfer: completed successfully, msgCid=%s", msgCid)
	return err
}

func (e *Executor) minerWithdraw(p MinerWithdrawPayload) error {

	minerAddr := p.MinerID
	val := p.Amount

	minerInfo, err := e.node.StateMinerInfo(minerAddr)
	if err != nil {
		log.Errorf("minerWithdraw: failed to get miner info for %s: %v", p.MinerID, err)
		return err
	}
	ownerAddr := minerInfo.Owner
	if ownerAddr.Protocol() == address.ID {
		ownerAddr, err = e.node.StateAccountKey(ownerAddr)
		if err != nil {
			log.Errorf("minerWithdraw: failed to get account key for owner: %v", err)
			return err
		}
	}

	available, err := e.node.StateMinerAvailableBalance(minerAddr)
	if err != nil {
		log.Errorf("minerWithdraw: failed to get available balance: %v", err)
		return err
	}
	if types.BigCmp(types.BigInt(val), available) == 1 {
		log.Errorf("minerWithdraw: requested %s > available %s", val, types.FIL(available))
		return fmt.Errorf("requested %s > available %s", val, types.FIL(available))
	}

	params, err := actors.SerializeParams(&minertypes.WithdrawBalanceParams{
		AmountRequested: types.BigInt(val),
	})
	if err != nil {
		log.Errorf("minerWithdraw: failed to serialize params: %v", err)
		return err
	}

	nonce, err := e.node.MpoolGetNonce(ownerAddr)
	if err != nil {
		log.Errorf("minerWithdraw: failed to get nonce: %v", err)
		return err
	}
	msg := &types.Message{
		Version:    0,
		To:         minerAddr,
		From:       ownerAddr,
		Nonce:      nonce,
		Value:      abi.NewTokenAmount(0),
		GasLimit:   0,
		GasFeeCap:  abi.NewTokenAmount(0),
		GasPremium: abi.NewTokenAmount(0),
		Method:     builtintypes.MethodsMiner.WithdrawBalance,
		Params:     params,
	}
	if err := wallet.SetGas(e.node, msg); err != nil {
		log.Errorf("minerWithdraw: failed to set gas: %v", err)
		return err
	}

	hasKey, err := wallet.WalletHas(e.store, ownerAddr)
	if err != nil {
		log.Errorf("minerWithdraw: failed to check key: %v", err)
		return err
	}
	if !hasKey {
		log.Errorf("minerWithdraw: wallet does not have key for %s", ownerAddr)
		return fmt.Errorf("wallet does not have key for %s", ownerAddr)
	}

	sig, err := wallet.WalletSign(e.store, ownerAddr, msg.Cid().Bytes())
	if err != nil {
		log.Errorf("minerWithdraw: failed to sign: %v", err)
		return err
	}
	signed := &types.SignedMessage{Message: *msg, Signature: *sig}

	msgCid, err := e.node.MpoolPush(signed)
	if err != nil {
		log.Errorf("minerWithdraw: failed to push message: %v", err)
		return err
	}

	_, err = e.node.StateWaitMsg(msgCid)
	if err != nil {
		log.Errorf("minerWithdraw: message %s failed: %v", msgCid, err)
		return err
	}

	log.Infof("minerWithdraw: completed successfully, msgCid=%s", msgCid)
	return err
}

func (e *Executor) marketWithdraw(p MarketWithdrawPayload) error {
	idAddr := p.Address
	if p.Address.Protocol() != address.ID {
		_, err := e.node.StateLookupID(p.Address)
		if err != nil {
			log.Errorf("marketWithdraw: failed to lookup ID for %s: %v", p.Address, err)
			return err
		}
	}
	signAddr := p.Address
	if p.Address.Protocol() == address.ID {
		_, err := e.node.StateAccountKey(p.Address)
		if err != nil {
			log.Errorf("marketWithdraw: failed to get account key for %s: %v", p.Address, err)
			return err
		}
	}

	bal, err := e.node.StateMarketBalance(idAddr)
	if err != nil {
		log.Errorf("marketWithdraw: failed to get market balance: %v", err)
		return err
	}
	available := types.BigSub(bal.Escrow, bal.Locked)
	if types.BigCmp(types.BigInt(p.Amount), available) == 1 {
		log.Errorf("marketWithdraw: requested %s > available %s", p.Amount, types.FIL(available))
		return fmt.Errorf("requested %s > available %s", p.Amount, types.FIL(available))
	}

	params, err := actors.SerializeParams(&markettypes.WithdrawBalanceParams{
		ProviderOrClientAddress: idAddr,
		Amount:                  types.BigInt(p.Amount),
	})
	if err != nil {
		log.Errorf("marketWithdraw: failed to serialize params: %v", err)
		return err
	}

	nonce, err := e.node.MpoolGetNonce(signAddr)
	if err != nil {
		log.Errorf("marketWithdraw: failed to get nonce: %v", err)
		return err
	}
	msg := &types.Message{
		Version:    0,
		To:         builtintypes.StorageMarketActorAddr,
		From:       signAddr,
		Nonce:      nonce,
		Value:      abi.NewTokenAmount(0),
		GasLimit:   0,
		GasFeeCap:  abi.NewTokenAmount(0),
		GasPremium: abi.NewTokenAmount(0),
		Method:     builtintypes.MethodsMarket.WithdrawBalance,
		Params:     params,
	}
	if err := wallet.SetGas(e.node, msg); err != nil {
		log.Errorf("marketWithdraw: failed to set gas: %v", err)
		return err
	}

	hasKey, err := wallet.WalletHas(e.store, signAddr)
	if err != nil {
		log.Errorf("marketWithdraw: failed to check key: %v", err)
		return err
	}
	if !hasKey {
		log.Errorf("marketWithdraw: wallet does not have key for %s", signAddr)
		return fmt.Errorf("wallet does not have key for %s", signAddr)
	}

	sig, err := wallet.WalletSign(e.store, signAddr, msg.Cid().Bytes())
	if err != nil {
		log.Errorf("marketWithdraw: failed to sign: %v", err)
		return err
	}
	signed := &types.SignedMessage{Message: *msg, Signature: *sig}

	msgCid, err := e.node.MpoolPush(signed)
	if err != nil {
		log.Errorf("marketWithdraw: failed to push message: %v", err)
		return err
	}

	_, err = e.node.StateWaitMsg(msgCid)
	if err != nil {
		log.Errorf("marketWithdraw: message %s failed: %v", msgCid, err)
		return err
	}

	log.Infof("marketWithdraw: completed successfully, msgCid=%s", msgCid)
	return err
}

func (e *Executor) batchTransfer(p BatchTransferPayload) error {

	for idx, item := range p.Items {
		log.Infof("batchTransfer: processing item %d/%d", idx+1, len(p.Items))
		data := TransferPayload{
			From:   item.From,
			To:     item.To,
			Amount: item.Amount,
		}
		if err := e.transfer(data); err != nil {
			log.Errorf("batchTransfer: item %d failed: %v", idx+1, err)
			return err
		}
	}

	log.Infof("batchTransfer: completed all %d items", len(p.Items))
	return nil
}

func (e *Executor) changeMinerOwner(p MinerChangeOwnerPayload) error {
	newAddrID, err := e.node.StateLookupID(p.NewOwner)
	if err != nil {
		log.Errorf("changeMinerOwner: failed to lookup new owner ID: %v", err)
		return err
	}
	fromAddrID, err := e.node.StateLookupID(p.FromOwner)
	if err != nil {
		log.Errorf("changeMinerOwner: failed to lookup from owner ID: %v", err)
		return err
	}

	minerInfo, err := e.node.StateMinerInfo(p.MinerID)
	if err != nil {
		log.Errorf("changeMinerOwner: failed to get miner info: %v", err)
		return err
	}
	if fromAddrID != minerInfo.Owner && fromAddrID != newAddrID {
		log.Errorf("changeMinerOwner: from address must be old owner or new owner")
		return fmt.Errorf("from address must be old owner or new owner")
	}

	params, err := actors.SerializeParams(&newAddrID)
	if err != nil {
		log.Errorf("changeMinerOwner: failed to serialize params: %v", err)
		return err
	}
	nonce, err := e.node.MpoolGetNonce(p.FromOwner)
	if err != nil {
		log.Errorf("changeMinerOwner: failed to get nonce: %v", err)
		return err
	}
	msg := &types.Message{
		From:   p.FromOwner,
		To:     p.MinerID,
		Method: builtintypes.MethodsMiner.ChangeOwnerAddress,
		Value:  types.NewInt(0),
		Nonce:  nonce,
		Params: params,
	}
	if err := wallet.SetGas(e.node, msg); err != nil {
		log.Errorf("changeMinerOwner: failed to set gas: %v", err)
		return err
	}

	hasKey, err := wallet.WalletHas(e.store, p.FromOwner)
	if err != nil {
		log.Errorf("changeMinerOwner: failed to check key: %v", err)
		return err
	}
	if !hasKey {
		log.Errorf("changeMinerOwner: wallet does not have key for %s", p.FromOwner)
		return fmt.Errorf("wallet does not have key for %s", p.FromOwner)
	}

	sig, err := wallet.WalletSign(e.store, p.FromOwner, msg.Cid().Bytes())
	if err != nil {
		log.Errorf("changeMinerOwner: failed to sign: %v", err)
		return err
	}
	signed := &types.SignedMessage{Message: *msg, Signature: *sig}

	msgCid, err := e.node.MpoolPush(signed)
	if err != nil {
		log.Errorf("changeMinerOwner: failed to push message: %v", err)
		return err
	}

	_, err = e.node.StateWaitMsg(msgCid)
	if err != nil {
		log.Errorf("changeMinerOwner: message %s failed: %v", msgCid, err)
		return err
	}

	log.Infof("changeMinerOwner: completed successfully, msgCid=%s", msgCid)
	return err
}

func (e *Executor) changeMinerWorker(p MinerChangeWorkerPayload) error {
	minerInfo, err := e.node.StateMinerInfo(p.MinerID)
	if err != nil {
		log.Errorf("changeMinerWorker: failed to get miner info: %v", err)
		return err
	}

	newWorker := minerInfo.Worker
	if p.NewWorker != address.Undef {
		newWorker = p.NewWorker
	}

	controlAddrs := minerInfo.ControlAddresses
	if len(p.NewControlAddrs) > 0 {
		controlAddrs = make([]address.Address, 0, len(p.NewControlAddrs))
		for _, a := range p.NewControlAddrs {
			addr := a
			controlAddrs = append(controlAddrs, addr)
		}
	}

	params, err := actors.SerializeParams(&minertypes.ChangeWorkerAddressParams{
		NewWorker:       newWorker,
		NewControlAddrs: controlAddrs,
	})
	if err != nil {
		log.Errorf("changeMinerWorker: failed to serialize params: %v", err)
		return err
	}

	nonce, err := e.node.MpoolGetNonce(minerInfo.Owner)
	if err != nil {
		log.Errorf("changeMinerWorker: failed to get nonce: %v", err)
		return err
	}
	owner, err := e.node.StateAccountKey(minerInfo.Owner)
	if err != nil {
		log.Errorf("changeMinerWorker: failed to get owner account key: %v", err)
		return err
	}
	msg := &types.Message{
		To:     p.MinerID,
		From:   owner,
		Value:  types.NewInt(0),
		Nonce:  nonce,
		Method: builtintypes.MethodsMiner.ChangeWorkerAddress,
		Params: params,
	}
	if err = wallet.SetGas(e.node, msg); err != nil {
		log.Errorf("changeMinerWorker: failed to set gas: %v", err)
		return err
	}

	hasKey, err := wallet.WalletHas(e.store, msg.From)
	if err != nil {
		log.Errorf("changeMinerWorker: failed to check key: %v", err)
		return err
	}
	if !hasKey {
		log.Errorf("changeMinerWorker: wallet does not have key for %s", msg.From)
		return fmt.Errorf("wallet does not have key for %s", msg.From)
	}

	sig, err := wallet.WalletSign(e.store, msg.From, msg.Cid().Bytes())
	if err != nil {
		log.Errorf("changeMinerWorker: failed to sign: %v", err)
		return err
	}
	signed := &types.SignedMessage{Message: *msg, Signature: *sig}

	msgCid, err := e.node.MpoolPush(signed)
	if err != nil {
		log.Errorf("changeMinerWorker: failed to push message: %v", err)
		return err
	}

	_, err = e.node.StateWaitMsg(msgCid)
	if err != nil {
		log.Errorf("changeMinerWorker: message %s failed: %v", msgCid, err)
		return err
	}

	log.Infof("changeMinerWorker: completed successfully, msgCid=%s", msgCid)
	return err
}

func (e *Executor) confirmMinerWorker(reqID uint, p MinerConfirmWorkerPayload) error {

	minerInfo, err := e.node.StateMinerInfo(p.MinerID)
	if err != nil {
		log.Errorf("confirmMinerWorker: failed to get miner info: %v", err)
		return err
	}

	newAddr, err := e.node.StateLookupID(p.NewWorker)
	if err != nil {
		log.Errorf("confirmMinerWorker: failed to lookup new worker ID: %v", err)
		return err
	}
	if minerInfo.NewWorker.Empty() || minerInfo.NewWorker != newAddr {
		log.Errorf("confirmMinerWorker: no matching worker change proposed")
		return fmt.Errorf("no matching worker change proposed")
	}

	head, err := e.node.ChainHead()
	if err != nil {
		log.Errorf("confirmMinerWorker: failed to get chain head: %v", err)
		return err
	}
	if head.Height() < minerInfo.WorkerChangeEpoch {
		log.Errorf("confirmMinerWorker: cannot confirm until epoch %d, current height %d", minerInfo.WorkerChangeEpoch, head.Height())
		return fmt.Errorf("cannot confirm until %d, current height %d", minerInfo.WorkerChangeEpoch, head.Height())
	}

	log.Infof("confirmMinerWorker: ready to confirm worker change at epoch %d", head.Height())

	nonce, err := e.node.MpoolGetNonce(minerInfo.Owner)
	if err != nil {
		log.Errorf("confirmMinerWorker: failed to get nonce: %v", err)
		return err
	}
	owner, err := e.node.StateAccountKey(minerInfo.Owner)
	if err != nil {
		log.Errorf("confirmMinerWorker: failed to get owner account key: %v", err)
		return err
	}
	msg := &types.Message{
		To:     p.MinerID,
		From:   owner,
		Value:  types.NewInt(0),
		Nonce:  nonce,
		Method: builtintypes.MethodsMiner.ConfirmChangeWorkerAddress,
	}
	if err := wallet.SetGas(e.node, msg); err != nil {
		log.Errorf("confirmMinerWorker: failed to set gas: %v", err)
		return err
	}

	hasKey, err := wallet.WalletHas(e.store, msg.From)
	if err != nil {
		log.Errorf("confirmMinerWorker: failed to check key: %v", err)
		return err
	}
	if !hasKey {
		log.Errorf("confirmMinerWorker: wallet does not have key for %s", msg.From)
		return fmt.Errorf("wallet does not have key for %s", msg.From)
	}

	log.Infof("confirmMinerWorker: signing message for %s", msg.From)
	sig, err := wallet.WalletSign(e.store, msg.From, msg.Cid().Bytes())
	if err != nil {
		log.Errorf("confirmMinerWorker: failed to sign: %v", err)
		return err
	}
	signed := &types.SignedMessage{Message: *msg, Signature: *sig}

	log.Infof("confirmMinerWorker: pushing message to mempool")
	msgCid, err := e.node.MpoolPush(signed)
	if err != nil {
		log.Errorf("confirmMinerWorker: failed to push message: %v", err)
		return err
	}

	log.Infof("confirmMinerWorker: waiting for message %s", msgCid)
	_, err = e.node.StateWaitMsg(msgCid)
	if err != nil {
		log.Errorf("confirmMinerWorker: message %s failed: %v", msgCid, err)
		return err
	}

	log.Infof("confirmMinerWorker: completed successfully, msgCid=%s", msgCid)
	return err
}

func contextBackground() context.Context {
	return context.Background()
}
