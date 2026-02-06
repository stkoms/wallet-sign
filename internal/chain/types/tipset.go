package types

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
)

type TipSet struct {
	CidsField   []cid.Cid      `json:"Cids"`
	HeightField abi.ChainEpoch `json:"Height"`
}

func (ts *TipSet) Cids() []cid.Cid {
	if ts == nil {
		return nil
	}
	return ts.CidsField
}

func (ts *TipSet) Height() abi.ChainEpoch {
	if ts == nil {
		return 0
	}
	return ts.HeightField
}
