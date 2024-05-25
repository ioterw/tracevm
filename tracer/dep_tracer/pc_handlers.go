package dep_tracer

import (
    "math/big"

    "github.com/ethereum/go-ethereum/common"
)

type PrecompileHandler interface {
    Register(map[common.Address]PrecompileHandler)
    Execute(db *SimpleDB, state *TransactionDB, input, output []byte)
}

func NewPrecompileHandlers() map[common.Address]PrecompileHandler {
    handlers := map[common.Address]PrecompileHandler{}

    new(ECRecoverHandler).Register(handlers)
    new(SHA256Handler).Register(handlers)
    new(Ripemd160Handler).Register(handlers)
    new(IdentityHandler).Register(handlers)
    new(ModExpHandler).Register(handlers)
    new(EcAddHandler).Register(handlers)
    new(EcMulHandler).Register(handlers)
    new(EcPairingHandler).Register(handlers)
    new(Blake2FHandler).Register(handlers)
    new(PointEvaluationHandler).Register(handlers)

    return handlers
}


type ECRecoverHandler struct {}
func (ph *ECRecoverHandler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0x1})] = ph
}
func (ph *ECRecoverHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileEcRecover {
        Result: output,
    }.Handle(db, state)
}


type SHA256Handler struct {}
func (ph *SHA256Handler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0x2})] = ph
}
func (ph *SHA256Handler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileSha256 {
        Result: output,
    }.Handle(db, state)
}


type Ripemd160Handler struct {}
func (ph *Ripemd160Handler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0x3})] = ph
}
func (ph *Ripemd160Handler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileSha256 {
        Result: output,
    }.Handle(db, state)
}


type IdentityHandler struct {}
func (ph *IdentityHandler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0x4})] = ph
}
func (ph *IdentityHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileIdentity {
        Result: output,
    }.Handle(db, state)
}


type ModExpHandler struct {}
func (ph *ModExpHandler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0x5})] = ph
}
func (ph *ModExpHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
     getData := func(data []byte, start uint64, size uint64) []byte {
        length := uint64(len(data))
        if start > length {
            start = length
        }
        end := start + size
        if end > length {
            end = length
        }
        return common.RightPadBytes(data[start:end], int(size))
    }
    var (
        baseLen = new(big.Int).SetBytes(getData(input, 0, 32)).Uint64()
        expLen  = new(big.Int).SetBytes(getData(input, 32, 32)).Uint64()
        modLen  = new(big.Int).SetBytes(getData(input, 64, 32)).Uint64()
    )

    DataPrecompileModExp {
        Result: output,
        BSize: baseLen,
        ESize: expLen,
        MSize: modLen,
    }.Handle(db, state)
}


type EcAddHandler struct {}
func (ph *EcAddHandler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0x6})] = ph
}
func (ph *EcAddHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileEcAdd {
        Result: output,
    }.Handle(db, state)
}


type EcMulHandler struct {}
func (ph *EcMulHandler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0x7})] = ph
}
func (ph *EcMulHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileEcMul {
        Result: output,
    }.Handle(db, state)
}


type EcPairingHandler struct {}
func (ph *EcPairingHandler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0x8})] = ph
}
func (ph *EcPairingHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileEcPairing {
        Result: output,
    }.Handle(db, state)
}


type Blake2FHandler struct {}
func (ph *Blake2FHandler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0x9})] = ph
}
func (ph *Blake2FHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileBlake2F {
        Result: output,
    }.Handle(db, state)
}


type PointEvaluationHandler struct {}
func (ph *PointEvaluationHandler) Register(handlers map[common.Address]PrecompileHandler) {
    handlers[common.BytesToAddress([]byte{0xA})] = ph
}
func (ph *PointEvaluationHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPointEvaluation {
        Result: output,
    }.Handle(db, state)
}
