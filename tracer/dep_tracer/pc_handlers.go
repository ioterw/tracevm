package dep_tracer

import (
    "math/big"
)

type PrecompileHandler interface {
    Register(map[Address]PrecompileHandler)
    Execute(db *SimpleDB, state *TransactionDB, input, output []byte)
}

func NewPrecompileHandlers() map[Address]PrecompileHandler {
    handlers := map[Address]PrecompileHandler{}

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

func addressify(data []byte) Address {
    val := append(make([]byte, 20-len(data)), data...)
    var res Address
    copy(res[:], val)
    return res
}


type ECRecoverHandler struct {}
func (ph *ECRecoverHandler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0x1})] = ph
}
func (ph *ECRecoverHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileEcRecover {
        Result: output,
    }.Handle(db, state)
}


type SHA256Handler struct {}
func (ph *SHA256Handler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0x2})] = ph
}
func (ph *SHA256Handler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileSha256 {
        Result: output,
    }.Handle(db, state)
}


type Ripemd160Handler struct {}
func (ph *Ripemd160Handler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0x3})] = ph
}
func (ph *Ripemd160Handler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileSha256 {
        Result: output,
    }.Handle(db, state)
}


type IdentityHandler struct {}
func (ph *IdentityHandler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0x4})] = ph
}
func (ph *IdentityHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileIdentity {
        Result: output,
    }.Handle(db, state)
}


type ModExpHandler struct {}
func (ph *ModExpHandler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0x5})] = ph
}
func (ph *ModExpHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    rightPadBytes := func(slice []byte, l int) []byte {
        if l <= len(slice) {
            return slice
        }

        padded := make([]byte, l)
        copy(padded, slice)

        return padded
    }
    getData := func(data []byte, start uint64, size uint64) []byte {
        length := uint64(len(data))
        if start > length {
            start = length
        }
        end := start + size
        if end > length {
            end = length
        }
        return rightPadBytes(data[start:end], int(size))
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
func (ph *EcAddHandler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0x6})] = ph
}
func (ph *EcAddHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileEcAdd {
        Result: output,
    }.Handle(db, state)
}


type EcMulHandler struct {}
func (ph *EcMulHandler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0x7})] = ph
}
func (ph *EcMulHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileEcMul {
        Result: output,
    }.Handle(db, state)
}


type EcPairingHandler struct {}
func (ph *EcPairingHandler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0x8})] = ph
}
func (ph *EcPairingHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileEcPairing {
        Result: output,
    }.Handle(db, state)
}


type Blake2FHandler struct {}
func (ph *Blake2FHandler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0x9})] = ph
}
func (ph *Blake2FHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPrecompileBlake2F {
        Result: output,
    }.Handle(db, state)
}


type PointEvaluationHandler struct {}
func (ph *PointEvaluationHandler) Register(handlers map[Address]PrecompileHandler) {
    handlers[addressify([]byte{0xA})] = ph
}
func (ph *PointEvaluationHandler) Execute(db *SimpleDB, state *TransactionDB, input, output []byte) {
    DataPointEvaluation {
        Result: output,
    }.Handle(db, state)
}
