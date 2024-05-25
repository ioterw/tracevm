package dep_tracer

import (
    "encoding/binary"

    "github.com/ethereum/go-ethereum/common"
)

type DEPByte struct {
    pos     uint64
    formula common.Hash
}

func DEPByteFromBin(val []byte) DEPByte {
    res := DEPByte{}
    i := 0
    res.pos = binary.BigEndian.Uint64(val[i:i+8])
    i += 8
    res.formula = *(*common.Hash)(val[i:])
    return res
}

func (b *DEPByte) Bin() []byte {
    res := []byte{}
    res = binary.BigEndian.AppendUint64(res, b.pos)
    res = append(res, b.formula[:]...)
    return res
}

func InitDEPBytes(size uint64) []DEPByte {
    res := []DEPByte{}
    b := DEPByte{0, ConstantInitZero.hash}
    for i := uint64(0); i < size; i++ {
        res = append(res, b)
    }
    return res
}

func FormulaDEPBytes(formula Formula) []DEPByte {
    res := []DEPByte{}
    size := uint64(len(formula.result))
    for i := uint64(0); i < size; i++ {
        res = append(res, DEPByte{i, formula.hash})
    }
    return res
}

func FormulaSliceDEPBytes(formula Formula, offset, size uint64) []DEPByte {
    res := []DEPByte{}
    formulaSize := uint64(len(formula.result))
    if offset + size > formulaSize {
        panic("formula size overflow")
    }
    for i := uint64(0); i < size; i++ {
        res = append(res, DEPByte{offset+i, formula.hash})
    }
    return res
}

func OverflowSliceDEPBytes(input []DEPByte, offset, size uint64) []DEPByte {
    l := uint64(len(input))
    if offset >= l {
        return InitDEPBytes(size)
    }
    inputBytes := size
    extraBytes := uint64(0)
    if offset + size > l {
        extraBytes = offset + size - l
        inputBytes = size - extraBytes
    }
    res := input[offset:offset+inputBytes]
    res = append(res, InitDEPBytes(extraBytes)...)
    return res
}

func CopyDEPBytes(val []DEPByte) []DEPByte {
    res := make([]DEPByte, len(val))
    copy(res, val)
    return res
}
