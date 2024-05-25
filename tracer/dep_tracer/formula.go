package dep_tracer

import (
    "crypto/sha256"
    "encoding/binary"

    "github.com/ethereum/go-ethereum/common"
)

type Formula struct {
    opcode   uint8
    result   []byte
    operands []common.Hash
    hash     common.Hash
}

func ConstantNew(opcode uint8, result []byte) Formula {
    res := Formula{opcode, result, []common.Hash{}, common.Hash{}}
    res.init_hash()
    return res
}

func FormulaNew(opcode uint8, result []byte, operands []common.Hash) Formula {
    res := Formula{opcode, result, operands, common.Hash{}}
    res.init_hash()
    return res
}

func FormulaBin(val []byte) Formula {
    f := Formula{}
    i := uint64(0)
    f.opcode = val[0]
    i += 1
    resultSize := binary.BigEndian.Uint64(val[i:i+8])
    i += 8
    f.result = val[i:i+resultSize]
    i += resultSize
    operandsSize := binary.BigEndian.Uint64(val[i:i+8])
    i += 8
    for j := uint64(0); j < operandsSize; j++ {
        f.operands = append(f.operands, *(*common.Hash)(val[i:i+32]))
        i += 32
    }
    f.init_hash()
    return f
}

func (f *Formula) Bin() []byte {
    res := []byte{}
    res = append(res, f.opcode)
    res = binary.BigEndian.AppendUint64(res, uint64(len(f.result)))
    res = append(res, f.result...)
    res = binary.BigEndian.AppendUint64(res, uint64(len(f.operands)))
    for _, operand := range f.operands {
        res = append(res, operand[:]...)
    }
    return res
}

func (f *Formula) IsConstant() bool {
    return OpcodeIsConstant(f.opcode)
}

func(f *Formula) init_hash() common.Hash {
    if f.hash != (common.Hash{}) {
        return f.hash
    }
    val := []byte{f.opcode}
    if f.IsConstant() {
        val = append(val, f.result...)
    } else {
        for _, o := range f.operands {
            val = append(val, o[:]...)
        }
    }
    f.hash = sha256.Sum256(val)
    return f.hash
}
