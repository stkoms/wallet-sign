package types

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
)

type Receipt struct {
	ExitCode int64 `json:"ExitCode"`
}

type MsgLookup struct {
	Receipt Receipt        `json:"Receipt"`
	Height  abi.ChainEpoch `json:"Height"`
}

type MinerInfo struct {
	Owner             address.Address   `json:"Owner"`
	Worker            address.Address   `json:"Worker"`
	ControlAddresses  []address.Address `json:"ControlAddresses"`
	NewWorker         address.Address   `json:"NewWorker"`
	WorkerChangeEpoch abi.ChainEpoch    `json:"WorkerChangeEpoch"`
}

type MarketBalance struct {
	Escrow BigInt `json:"Escrow"`
	Locked BigInt `json:"Locked"`
}

type BlockMessages struct {
	BlsMessages   []*Message       `json:"BlsMessages"`
	SecpkMessages []*SignedMessage `json:"SecpkMessages"`
}

type Actor struct {
	Code    cid.Cid `json:"Code"`
	Head    cid.Cid `json:"Head"`
	Nonce   uint64  `json:"Nonce"`
	Balance BigInt  `json:"Amount"`
}
