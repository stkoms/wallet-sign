package types

import (
	"math/big"

	big2 "github.com/filecoin-project/go-state-types/big"
)

// FilecoinPrecision is the number of attoFIL in 1 FIL.
const FilecoinPrecision = uint64(1_000_000_000_000_000_000)

type BigInt = big2.Int

var EmptyInt = BigInt{}

func NewInt(i uint64) BigInt {
	return BigInt{Int: big.NewInt(0).SetUint64(i)}
}

func FromFil(i uint64) BigInt {
	return BigMul(NewInt(i), NewInt(FilecoinPrecision))
}

func BigMul(a, b BigInt) BigInt {
	return big2.Mul(a, b)
}

func BigSub(a, b BigInt) BigInt {
	return big2.Sub(a, b)
}

func BigCmp(a, b BigInt) int {
	return big2.Cmp(a, b)
}
