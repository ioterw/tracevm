package dep_tracer

import (
    "github.com/ethereum/go-ethereum/common"
)

type ProtectedDefinition struct {
    ops  map[uint8]bool
    name string
}

func NewProtectedDefinition(name string, ops []uint8) ProtectedDefinition {
    res := ProtectedDefinition{}
    res.name = name
    res.ops = make(map[uint8]bool)
    for _, op := range ops {
        res.ops[op] = true
    }
    return res
}

func (po *ProtectedDefinition) IsProtected(op uint8) bool {
    _, ok := po.ops[op]
    return ok
}

func CryptoProtectedDefinition() ProtectedDefinition {
    opcodes := []uint8{OPSLoad, OPSStore, OPKeccak, OPCodeKeccak}
    opcodes = append(opcodes, OPEcRecover, OPSha256, OPRipemd160, OPModExp, OPEcAddX)
    opcodes = append(opcodes, OPEcAddY, OPEcMulX, OPEcMulY, OPEcPairing, OPBlake2F)
    return NewProtectedDefinition("crypto", opcodes)
}


type HashAndProtected struct {
    hash            common.Hash
    protected       bool
    sourceHash      common.Hash
    sourceProtected bool
}

func HashAndProtectedFromBin(val []byte) HashAndProtected {
    res := HashAndProtected{}
    i := 0
    copy(res.hash[:], val[i:i+32])
    i += 32
    res.protected = val[i] != 0
    i += 1
    if len(val) > i {
        copy(res.sourceHash[:], val[i:i+32])
        i += 32
        res.sourceProtected = val[i] != 0
    }
    return res
}

func (hp *HashAndProtected) Bin() []byte {
    res := hp.hash[:]
    if hp.protected {
        res = append(res, 1)
    } else {
        res = append(res, 0)        
    }
    if hp.sourceHash != (common.Hash{}) {
        res = append(res, hp.sourceHash[:]...)
        if hp.sourceProtected {
            res = append(res, 1)
        } else {
            res = append(res, 0)        
        }
    }
    return res
}


type Shorterner struct {
    simpleDB          *SimpleDB
    formulasMappingDB DB
    formulasMapping   map[common.Hash]HashAndProtected
    protected         ProtectedDefinition
}

func NewShorterner(simpleDB *SimpleDB, kvEngine, kvRoot string, single_instances map[string]DB, def ProtectedDefinition) *Shorterner {
    s := new(Shorterner)
    var ok bool
    formulasMappingName := "global." + def.name + ".formula_mappings"
    if s.formulasMappingDB, ok = single_instances[formulasMappingName]; !ok {
        s.formulasMappingDB = NewDB(kvEngine, kvRoot, formulasMappingName)
        single_instances[formulasMappingName] = s.formulasMappingDB
    }
    s.simpleDB = simpleDB
    s.protected = def
    s.Reset()
    return s
}

func (s *Shorterner) Reset() {
    s.formulasMapping = make(map[common.Hash]HashAndProtected)
    initHash := ConstantInitZero.hash
    zeroHash := ConstantZero.hash
    if s.protected.IsProtected(OPInitZero) {
        s.formulasMapping[initHash] = HashAndProtected{initHash, true, common.Hash{}, false}
    } else {
        s.formulasMapping[initHash] = HashAndProtected{zeroHash, false, common.Hash{}, false}
    }
    s.formulasMapping[zeroHash] = HashAndProtected{zeroHash, false, common.Hash{}, false}
}

func (s *Shorterner) LoadChildHash(parentHash common.Hash) HashAndProtected {
    if child, ok := s.formulasMapping[parentHash]; ok {
        return child
    }
    val := s.formulasMappingDB.Get(parentHash[:], false)
    child := HashAndProtectedFromBin(val)
    s.formulasMapping[parentHash] = child
    return child
}

func (s *Shorterner) SaveChildHash(parentHash common.Hash) {
    child := s.formulasMapping[parentHash]
    s.formulasMappingDB.Set(parentHash[:], child.Bin())
}

func (s *Shorterner) Shortern(parentFormula Formula) {
    parentHash := parentFormula.hash

    formulaOps := []Formula{}
    protected  := s.protected.IsProtected(parentFormula.opcode)

    isSource        := OpcodeIsAddressable(parentFormula.opcode)
    sourceHash      := common.Hash{}
    sourceProtected := false

    for i, hash := range parentFormula.operands {
        child := s.LoadChildHash(hash)
        if child.sourceHash != (common.Hash{}) {
            child = HashAndProtected{child.sourceHash, child.sourceProtected, common.Hash{}, false}
        }
        if isSource {
            if i == 0 {
                sourceHash = child.hash
                sourceProtected = child.protected
                protected = protected || child.protected
            }
        } else {
            protected = protected || child.protected
        }
        formula := s.simpleDB.GetFormula(child.hash)
        formulaOps = append(formulaOps, formula)
    }

    if protected && parentFormula.opcode == OPConcat {
        ops := []common.Hash{}
        constData := []byte{}
        for _, op := range formulaOps {
            if op.opcode == OPConstant {
                constData = append(constData, op.result...)
            } else {
                if len(constData) > 0 {
                    constFormula := s.simpleDB.ConstantNew(OPConstant, constData)
                    ops = append(ops, constFormula.hash)
                    constData = []byte{}
                }
                ops = append(ops, op.hash)
            }
        }
        if len(constData) > 0 {
            constFormula := s.simpleDB.ConstantNew(OPConstant, constData)
            ops = append(ops, constFormula.hash)
        }
        childFormula := s.simpleDB.FormulaNew(OPConcat, parentFormula.result, ops)
        s.formulasMapping[parentHash] = HashAndProtected{childFormula.hash, true, common.Hash{}, false}
        return
    }

    if protected {
        ops := []common.Hash{}
        for _, op := range formulaOps {
            ops = append(ops, op.hash)
        }
        childFormula := s.simpleDB.FormulaNew(parentFormula.opcode, parentFormula.result, ops)
        if isSource {
            s.formulasMapping[parentHash] = HashAndProtected{childFormula.hash, true, sourceHash, sourceProtected}
        } else {
            s.formulasMapping[parentHash] = HashAndProtected{childFormula.hash, true, common.Hash{}, false}
        }
        return
    }

    childFormula := s.simpleDB.ConstantNew(OPConstant, parentFormula.result)
    s.formulasMapping[parentHash] = HashAndProtected{childFormula.hash, false, common.Hash{}, false}
}
