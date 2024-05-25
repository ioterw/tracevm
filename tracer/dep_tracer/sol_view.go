package dep_tracer

import (
    "encoding/hex"
)

type solViewLine struct {
    Type uint8
    Data []byte
}

type SolView []solViewLine

func (s *SolView) Print(writer OutputWriter) {
    for i, line := range *s {
        switch line.Type {
        case 'c':
            writer.Println("#", i, "constant", hex.EncodeToString(line.Data))
        case 'o':
            writer.Println("#", i, "offset  ", hex.EncodeToString(line.Data))
        case 'm':
            writer.Print("# ", i, " mapping  ", hex.EncodeToString(line.Data))
            if len(line.Data) == 0 {
                writer.Print("(possibly array)")
            }
            writer.Println()
        default:
            panic("unknown type")
        }
    }
}

func (s *SolView) Serialize() [][]byte {
    sig := []byte{}
    for _, line := range *s {
        sig = append(sig, line.Type)
    }
    res := [][]byte{sig}
    for _, line := range *s {
        res = append(res, line.Data)
    }
    return res
}

func SolViewNew(s *SimpleDB, formula Formula) SolView {
    allZero := func(s []byte) bool {
        for _, v := range s {
            if v != 0 {
                return false
            }
        }
        return true
    }

    res := SolView{}
    switch (formula.opcode) {
    case OPKeccak:
        keccakValueFormula := s.GetFormula(formula.operands[0])
        l := len(keccakValueFormula.result)
        if l >= 32 {
            slotFormula := s.FormulaSlice(keccakValueFormula, uint64(l-32), 32)
            res = append(res, SolViewNew(s, slotFormula)...)

            val := s.FormulaSlice(keccakValueFormula, 0, uint64(l-32)).result
            // mapping (or array when key length is zero)
            res = append(res, solViewLine{'m', val})
        } else {
            val := formula.result
            // constant
            res = append(res, solViewLine{'c', val})
        }
    case OPAdd:
        op0 := s.GetFormula(formula.operands[0])
        op1 := s.GetFormula(formula.operands[1])
        if op0.opcode != OPKeccak && op1.opcode != OPKeccak {
            val := formula.result
            // constant
            res = append(res, solViewLine{'c', val})
        } else {
            if op0.opcode != OPKeccak {
                op0, op1 = op1, op0
            }
            slotFormula := op0
            res = append(res, SolViewNew(s, slotFormula)...)

            val := op1.result
            if !allZero(val) {
                // offset
                res = append(res, solViewLine{'o', val})
            }
        }
    case OPConstant:
        val := formula.result
        // constant
        res = append(res, solViewLine{'c', val})
    default:
        val := formula.result
        // constant
        res = append(res, solViewLine{'c', val})
    }
    return res
}
