package dep_tracer

import (
    "github.com/ethereum/go-ethereum/common"
)

var (
    ConstantInitZero = ConstantNew(OPInitZero, []byte{0})
    ConstantZero     = ConstantNew(OPConstant, []byte{0})
)

type StateDB interface {
    GetNonce(addr common.Address) uint64
}
