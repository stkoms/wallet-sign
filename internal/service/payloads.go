package service

import (
	"wallet-sign/internal/chain/types"

	"github.com/filecoin-project/go-address"
)

type Payload struct {
	Type            string              `json:"type"`
	FromAddr        address.Address     `json:"from_address"`
	ToAddr          address.Address     `json:"to_address"`
	Amount          types.FIL           `json:"amount"`
	MinerID         address.Address     `json:"miner_id"`
	NewOwner        address.Address     `json:"new_owner"`
	FromOwner       address.Address     `json:"from_owner"`
	NewWorker       address.Address     `json:"new_worker"`
	NewControlAddrs []address.Address   `json:"new_control_addrs"`
	Items           []BatchTransferItem `json:"items"`
}

type TransferPayload struct {
	From   address.Address `json:"from"`
	To     address.Address `json:"to"`
	Amount types.FIL       `json:"amount"`
}

type MinerWithdrawPayload struct {
	MinerID address.Address `json:"miner_id"`
	Amount  types.FIL       `json:"amount"`
}

type MarketWithdrawPayload struct {
	Address address.Address `json:"address"`
	Amount  types.FIL       `json:"amount"`
}

type BatchTransferItem struct {
	From   address.Address `json:"from"`
	To     address.Address `json:"to"`
	Amount types.FIL       `json:"amount"`
}

type BatchTransferPayload struct {
	Items []BatchTransferItem `json:"items"`
}

type MinerChangeOwnerPayload struct {
	MinerID   address.Address `json:"miner_id"`
	NewOwner  address.Address `json:"new_owner"`
	FromOwner address.Address `json:"from_owner"`
}

type MinerChangeWorkerPayload struct {
	MinerID         address.Address   `json:"miner_id"`
	NewWorker       address.Address   `json:"new_worker"`
	NewControlAddrs []address.Address `json:"new_control_addrs"`
}

type MinerConfirmWorkerPayload struct {
	MinerID   address.Address `json:"miner_id"`
	NewWorker address.Address `json:"new_worker"`
}
