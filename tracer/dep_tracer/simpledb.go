package dep_tracer

import (
    "fmt"
    "strconv"
    "strings"
    "crypto/sha256"
    "encoding/hex"
    "encoding/binary"
    "github.com/holiman/uint256"

    "github.com/ethereum/go-ethereum/common"
)

type CommitFormula struct {
    formula  Formula
    commited bool
}

type SimpleDB struct {
    formulasWithShorts map[common.Hash]*CommitFormula
    formulas           map[common.Hash]*CommitFormula
    formulasDB         DB
    slotsDB            DB
    codesDB            DB
    codeHashesDB       DB
    versionsDB         DB
    shorts             []*Shorterner
    logger             Logger
    writer             OutputWriter
}

func CodeHash(code []byte) common.Hash {
    return sha256.Sum256(code)
}

func storeLocation(addr common.Address, version uint64, slot *uint256.Int, pos uint8) []byte {
    res := addr[:]
    res = binary.BigEndian.AppendUint64(res, version)
    slotBytes := slot.Bytes32()
    res = append(res, slotBytes[:]...)
    res = append(res, pos)
    return res
}

func codeHashLocation(addr common.Address, version uint64) []byte {
    res := addr[:]
    res = binary.BigEndian.AppendUint64(res, version)
    return res
}

func codeLocation(addr common.Address, version uint64, pos uint64) []byte {
    res := addr[:]
    res = binary.BigEndian.AppendUint64(res, version)
    res = binary.BigEndian.AppendUint64(res, pos)
    return res
}

func SimpleDBNew(
    protectedDifinitions []ProtectedDefinition,
    toLog LoggerDefinition,
    kvEngine, kvRoot string,
    writer OutputWriter,
) *SimpleDB {
    s := new(SimpleDB)
    var ok bool
    
    single_instances := make(map[string]DB)
    formulasName := "global.formulas"
    if s.formulasDB, ok = single_instances[formulasName]; !ok {
        s.formulasDB = NewDB(kvEngine, kvRoot, formulasName)
        s.saveFormula(ConstantInitZero)
        s.saveFormula(ConstantZero)
        single_instances[formulasName] = s.formulasDB
    }
    s.formulasWithShorts = make(map[common.Hash]*CommitFormula)
    s.formulas           = make(map[common.Hash]*CommitFormula)

    s.shorts = make([]*Shorterner, 0)
    for _, def := range protectedDifinitions {
        s.shorts = append(s.shorts, NewShorterner(s, kvEngine, kvRoot, single_instances, def))
    }
    
    s.slotsDB      = NewDB(kvEngine, kvRoot, "slots")
    s.codesDB      = NewDB(kvEngine, kvRoot, "codes")
    s.codeHashesDB = NewDB(kvEngine, kvRoot, "code_hashes")
    s.versionsDB   = NewDB(kvEngine, kvRoot, "versions")

    s.logger = NewLogger(s, toLog, writer)
    s.writer = writer

    return s
}

func (s *SimpleDB) CommitDEPBytes(data []DEPByte) {
    prevFormula := common.Hash{}
    for _, b := range data {
        curFormula := b.formula
        if curFormula != prevFormula {
            s.CommitFormula(b.formula)
            prevFormula = curFormula
        }
    }
}

func (s *SimpleDB) CommitDEPBytesWithShorts(data []DEPByte) {
    prevFormula := common.Hash{}
    for _, b := range data {
        curFormula := b.formula
        if curFormula != prevFormula {
            s.CommitFormulaWithShorts(b.formula)
            prevFormula = curFormula
        }
    }
}

func (s *SimpleDB) commitFormulaInternal(hash common.Hash, ignoreExistance bool) {
    cf, ok := s.formulas[hash]
    if !ok {
        if !ignoreExistance {
            panic("CommitFormula missing hash")
        }
        return
    }
    if cf.commited {
        return
    }
    s.saveFormula(cf.formula)
    cf.commited = true
    for _, operandHash := range cf.formula.operands {
        s.commitFormulaInternal(operandHash, true)
    }
}

func (s *SimpleDB) CommitFormula(hash common.Hash) {
    s.commitFormulaInternal(hash, false)
}

func (s *SimpleDB) commitFormulaWithShortsInternal(hash common.Hash, ignoreExistance bool) {
    cf, ok := s.formulasWithShorts[hash]
    if !ok {
        if !ignoreExistance {
            panic("CommitFormulaWithShorts missing hash")
        }
        return
    }
    if cf.commited {
        return
    }
    s.saveFormula(cf.formula)
    cf.commited = true
    for _, short := range s.shorts {
        shortHash := short.formulasMapping[hash].hash
        cf1, ok := s.formulas[shortHash]
        if !ok {
            panic("short formula is not found")
        }
        if cf1.commited {
            continue
        }
        s.saveFormula(cf1.formula)
        cf1.commited = true
        short.SaveChildHash(hash)
    }
    for _, operandHash := range cf.formula.operands {
        s.commitFormulaWithShortsInternal(operandHash, true)
    }
}

func (s *SimpleDB) CommitFormulaWithShorts(hash common.Hash) {
    s.commitFormulaWithShortsInternal(hash, false)
}

func (s *SimpleDB) ResetFormulas() {
    s.formulasWithShorts = make(map[common.Hash]*CommitFormula)
    s.formulas           = make(map[common.Hash]*CommitFormula)
    for _, short := range s.shorts {
        short.Reset()
    }
    // s.DebugPrintAllFormulas()
}

func (s *SimpleDB) GetFormulaWithShorts(hash common.Hash) Formula {
    if cf, ok := s.formulasWithShorts[hash]; ok {
        return cf.formula
    }
    formula := s.GetFormula(hash)
    s.formulasWithShorts[hash] = &CommitFormula{formula, true}
    for _, short := range s.shorts {
        shortHash := short.LoadChildHash(hash).hash
        s.GetFormula(shortHash)
    }
    return formula
}

func (s *SimpleDB) GetFormula(hash common.Hash) Formula {
    if cf, ok := s.formulas[hash]; ok {
        return cf.formula
    }
    formula := s.loadFormula(hash)
    s.formulas[hash] = &CommitFormula{formula, true}
    return formula
}

func (s *SimpleDB) ConstantNewWithShorts(opcode uint8, result []byte) Formula {
    res := s.ConstantNew(opcode, result)
    if _, ok := s.formulasWithShorts[res.hash]; ok {
        return res
    }
    s.formulasWithShorts[res.hash] = &CommitFormula{res, false}
    for _, short := range s.shorts {
        short.Shortern(res)
    }
    s.logger.LogOpcode(res)
    return res
}

func (s *SimpleDB) ConstantNew(opcode uint8, result []byte) Formula {
    res := ConstantNew(opcode, result)
    h := res.hash
    if _, ok := s.formulas[h]; ok {
        return res
    }
    s.formulas[h] = &CommitFormula{res, false}
    return res
}

func (s *SimpleDB) FormulaNewWithShorts(opcode uint8, result []byte, operands []common.Hash) Formula {
    res := s.FormulaNew(opcode, result, operands)
    if _, ok := s.formulasWithShorts[res.hash]; ok {
        return res
    }
    s.formulasWithShorts[res.hash] = &CommitFormula{res, false}
    for _, short := range s.shorts {
        short.Shortern(res)
    }
    s.logger.LogOpcode(res)
    return res
}

func (s *SimpleDB) FormulaNew(opcode uint8, result []byte, operands []common.Hash) Formula {
    res := FormulaNew(opcode, result, operands)
    h := res.hash
    if _, ok := s.formulas[h]; ok {
        return res
    }
    s.formulas[h] = &CommitFormula{res, false}
    return res
}

func (s *SimpleDB) FormulaSlice(formula Formula, offset, size uint64) Formula {
    totalSize := uint64(len(formula.result))

    offset1 := offset + size
    if offset > totalSize || offset1 > totalSize {
        panic(fmt.Errorf("Out of bounds"))
    }

    if size == 0 {
        return s.FormulaNew(OPConcat, []byte{}, []common.Hash{})
    }

    switch(formula.opcode) {
    case OPConcat:
        byte_parts := []byte{}
        hash_parts := []common.Hash{}
        i := uint64(0)
        for _, formulaHash := range formula.operands {
            formula1 := s.GetFormula(formulaHash)
            j := i + uint64(len(formula1.result))
            if j >= offset {
                if i >= offset && j <= offset1 {
                    // slice none
                    byte_parts = append(byte_parts, formula1.result...)
                    hash_parts = append(hash_parts, formula1.hash)
                } else if i < offset && j > offset1 {
                    // slice left right
                    le := offset - i
                    si := offset1 - offset
                    if si > 0 {
                        fo := s.FormulaSlice(formula1, le, si)
                        byte_parts = append(byte_parts, fo.result...)
                        hash_parts = append(hash_parts, fo.hash)
                    }
                } else if i < offset && j <= offset1 {
                    // slice left
                    le := offset - i
                    si := j - offset
                    if si > 0 {
                        fo := s.FormulaSlice(formula1, le, si)
                        byte_parts = append(byte_parts, fo.result...)
                        hash_parts = append(hash_parts, fo.hash)
                    }
                } else if i >= offset && j > offset1 {
                    le := uint64(0)
                    si := offset1 - i
                    if si > 0 {
                        fo := s.FormulaSlice(formula1, le, si)
                        byte_parts = append(byte_parts, fo.result...)
                        hash_parts = append(hash_parts, fo.hash)
                    }
                } else {
                    panic("Some strange range happened")
                }
            }
            if offset1 <= j {
                break
            }
            i = j
        }
        if len(hash_parts) == 1 {
            return s.GetFormula(hash_parts[0])
        }
        return s.FormulaNew(OPConcat, byte_parts, hash_parts)
    case OPSlice:
        // modify slice
        prevOffsetOpVal := s.GetFormula(formula.operands[1]).result
        prevOffset := binary.BigEndian.Uint64(prevOffsetOpVal)

        offsetOpVal := binary.BigEndian.AppendUint64([]byte{}, prevOffset + offset)
        offsetOp := s.ConstantNew(OPConstant, offsetOpVal)

        sizeOpVal := binary.BigEndian.AppendUint64([]byte{}, size)
        sizeOp := s.ConstantNew(OPConstant, sizeOpVal)

        return s.FormulaNew(OPSlice, formula.result[offset:offset1], []common.Hash{formula.operands[0], offsetOp.hash, sizeOp.hash})
    default:
        // make slice
        offsetOpVal := binary.BigEndian.AppendUint64([]byte{}, offset)
        offsetOp := s.ConstantNew(OPConstant, offsetOpVal)

        sizeOpVal := binary.BigEndian.AppendUint64([]byte{}, size)
        sizeOp := s.ConstantNew(OPConstant, sizeOpVal)

        return s.FormulaNew(OPSlice, formula.result[offset:offset1], []common.Hash{formula.hash, offsetOp.hash, sizeOp.hash})
    }
}

func (s *SimpleDB) FormulaDep(val []DEPByte) Formula {
    if len(val) == 0 {
        return s.FormulaNew(OPConcat, []byte{}, []common.Hash{})
    }

    valBin := make([]byte, 0)
    byteRanges := make([][]DEPByte, 0)
    byteRange := make([]DEPByte, 0)
    for _, b := range val {
        if len(byteRange) == 0 {
            byteRange = append(byteRange, b)
            continue
        }
        prevB := byteRange[len(byteRange)-1]
        if prevB.formula == b.formula && prevB.pos + 1 == b.pos {
            byteRange = append(byteRange, b)
            continue
        }
        byteRanges = append(byteRanges, byteRange)
        byteRange = []DEPByte{b}
    }
    if len(byteRange) > 0 {
        byteRanges = append(byteRanges, byteRange)
    }

    res := []common.Hash{}
    for _, byteRange := range byteRanges {
        rangeFirst := byteRange[0].pos
        rangeSize := byteRange[len(byteRange)-1].pos - rangeFirst + 1
        formulaHash := byteRange[0].formula
        formula := s.GetFormula(formulaHash)
        size := uint64(len(formula.result))
        if rangeFirst == 0 && rangeSize == size {
            res = append(res, formulaHash)
            valBin = append(valBin, formula.result...)
        } else {
            offsetOpVal := []byte{}
            offsetOpVal = binary.BigEndian.AppendUint64(offsetOpVal, rangeFirst)
            offsetOp := s.ConstantNew(OPConstant, offsetOpVal)
            sizeOpVal := []byte{}
            sizeOpVal = binary.BigEndian.AppendUint64(sizeOpVal, rangeSize)
            sizeOp := s.ConstantNew(OPConstant, sizeOpVal)
            valBinSlice := formula.result[rangeFirst:rangeFirst+rangeSize]
            valBin = append(valBin, valBinSlice...)
            slice := s.FormulaNew(OPSlice, valBinSlice, []common.Hash{formulaHash, offsetOp.hash, sizeOp.hash})
            res = append(res, slice.hash)
        }
    }

    if len(res) == 1 {
        return s.GetFormula(res[0])
    }
    return s.FormulaNew(OPConcat, valBin, res)
}

func (s *SimpleDB) FormulaDepWithShorts(val []DEPByte) Formula {
    if len(val) == 0 {
        return s.FormulaNewWithShorts(OPConcat, []byte{}, []common.Hash{})
    }

    valBin := make([]byte, 0)
    byteRanges := make([][]DEPByte, 0)
    byteRange := make([]DEPByte, 0)
    for _, b := range val {
        if len(byteRange) == 0 {
            byteRange = append(byteRange, b)
            continue
        }
        prevB := byteRange[len(byteRange)-1]
        if prevB.formula == b.formula && prevB.pos + 1 == b.pos {
            byteRange = append(byteRange, b)
            continue
        }
        byteRanges = append(byteRanges, byteRange)
        byteRange = []DEPByte{b}
    }
    if len(byteRange) > 0 {
        byteRanges = append(byteRanges, byteRange)
    }

    res := []common.Hash{}
    for _, byteRange := range byteRanges {
        rangeFirst := byteRange[0].pos
        rangeSize := byteRange[len(byteRange)-1].pos - rangeFirst + 1
        formulaHash := byteRange[0].formula
        formula := s.GetFormulaWithShorts(formulaHash)
        size := uint64(len(formula.result))
        if rangeFirst == 0 && rangeSize == size {
            res = append(res, formulaHash)
            valBin = append(valBin, formula.result...)
        } else {
            offsetOpVal := []byte{}
            offsetOpVal = binary.BigEndian.AppendUint64(offsetOpVal, rangeFirst)
            offsetOp := s.ConstantNewWithShorts(OPConstant, offsetOpVal)
            sizeOpVal := []byte{}
            sizeOpVal = binary.BigEndian.AppendUint64(sizeOpVal, rangeSize)
            sizeOp := s.ConstantNewWithShorts(OPConstant, sizeOpVal)
            valBinSlice := formula.result[rangeFirst:rangeFirst+rangeSize]
            valBin = append(valBin, valBinSlice...)
            slice := s.FormulaNewWithShorts(OPSlice, valBinSlice, []common.Hash{formulaHash, offsetOp.hash, sizeOp.hash})
            res = append(res, slice.hash)
        }
    }

    if len(res) == 1 {
        return s.GetFormulaWithShorts(res[0])
    }
    return s.FormulaNewWithShorts(OPConcat, valBin, res)
}

func (s *SimpleDB) saveFormula(f Formula) {
    hash := f.hash
    s.formulasDB.Set(hash[:], f.Bin())
}

func (s *SimpleDB) loadFormula(h common.Hash) Formula {
    val := s.formulasDB.Get(h[:], false)
    return FormulaBin(val)
}

func (s *SimpleDB) GetAddressVersion(addr common.Address) uint64 {
    val := s.versionsDB.Get(addr[:], true)
    if val == nil {
        return 0
    }
    return binary.BigEndian.Uint64(val)
}

func (s *SimpleDB) IncreaseAddressVersion(addr common.Address) uint64 {
    version := s.GetAddressVersion(addr) + 1
    versionBin := []byte{}
    versionBin = binary.BigEndian.AppendUint64(versionBin, version)
    s.versionsDB.Set(addr[:], versionBin)
    return version
}

func (s *SimpleDB) GetSlot(addr common.Address, slot *uint256.Int) []DEPByte {
    version := s.GetAddressVersion(addr)
    res := make([]DEPByte, 0)
    for i := uint8(0); i < 32; i++ {
        location := storeLocation(addr, version, slot, i)
        val := s.slotsDB.Get(location, true)
        if val != nil {
            res = append(res, DEPByteFromBin(val))
        }
    }
    if len(res) == 0 {
        return InitDEPBytes(32)
    }
    if len(res) == 32 {
        return res
    }
    panic("Invalid number of results for slot")
}

func (s *SimpleDB) SetSlot(addr common.Address, slot *uint256.Int, val []DEPByte) {
    if len(val) != 32 {
        panic("Invalid number of arguments for slot")
    }
    version := s.GetAddressVersion(addr)
    for i, b := range val {
        location := storeLocation(addr, version, slot, uint8(i))
        s.slotsDB.Set(location, b.Bin())
    }
}

func (s *SimpleDB) GetCode(addr common.Address) (common.Hash, common.Hash, []DEPByte) {
    version := s.GetAddressVersion(addr)

    location := codeHashLocation(addr, version)
    codeHashData := s.codeHashesDB.Get(location, true)
    codeHash := common.Hash{}
    initcodeHash := common.Hash{}
    if codeHashData != nil {
        copy(codeHash[:],     codeHashData[:32])
        copy(initcodeHash[:], codeHashData[32:])
    }

    res := make([]DEPByte, 0)
    for i := uint64(0);; i++ {
        location := codeLocation(addr, version, i)
        val := s.codesDB.Get(location, true)
        if val == nil {
            break
        }
        res = append(res, DEPByteFromBin(val))
    }
    return codeHash, initcodeHash, res
}

func (s *SimpleDB) SetCode(addr common.Address, val []DEPByte, codeHash, initcodeHash common.Hash) {
    version := s.GetAddressVersion(addr)

    location := codeHashLocation(addr, version)
    s.codeHashesDB.Set(location, append(codeHash[:], initcodeHash[:]...))

    var i int
    var b DEPByte
    for i = 0; i < len(val); i++ {
        b = val[i]
        location := codeLocation(addr, version, uint64(i))
        s.codesDB.Set(location, b.Bin())
    }
    for ;; i++ {
        location := codeLocation(addr, version, uint64(i))
        val := s.codesDB.Get(location, true)
        if val == nil {
            break
        }
        s.codesDB.Delete(location)
    }
}

func (s *SimpleDB) Print(f Formula) {
    var fun func(f1 *Formula, offset int) string
    fun = func(f1 *Formula, offset int) string {
        res := ""
        if f1.IsConstant() {
            res += strings.Repeat("    ", offset) + OpcodeToString[f1.opcode] + "(0x" + hex.EncodeToString(f1.result) + ")\n"
            return res
        }
        if len(f1.operands) < 1 {
            res += strings.Repeat("    ", offset) + OpcodeToString[f1.opcode] + "()\n"
            return res
        }
        h0 := common.Hash{}
        repeated := 0
        res += strings.Repeat("    ", offset) + OpcodeToString[f1.opcode] + "( # 0x" + hex.EncodeToString(f1.result) + "\n"
        for i, h1 := range f1.operands {
            offset += 1
            if (h0 == h1 && i > 0) {
                repeated += 1
            } else {
                if repeated > 0 {
                    res = res[:len(res)-1] + " * " + strconv.Itoa(repeated + 1) + "\n"
                    repeated = 0
                }
                f2 := s.GetFormula(h1)
                res += fun(&f2, offset)
            }
            offset -= 1
            h0 = h1
        }
        if repeated > 0 {
            res = res[:len(res)-1] + " * " + strconv.Itoa(repeated + 1) + "\n"
            repeated = 0
        }
        res += strings.Repeat("    ", offset) + ")\n"
        return res
    }
    s.writer.Print(fun(&f, 0))
}

func (s *SimpleDB) FullPrint(f Formula) {
    s.Print(f)
    for _, short := range s.shorts {
        hash := short.LoadChildHash(f.hash).hash
        formula := s.GetFormula(hash)
        s.Print(formula)
    }
}

func (s *SimpleDB) DebugPrintAllFormulas() {
    a := s.formulasDB.DumpAllDebug()
    i := 0
    for _, data := range a {
        s.writer.Print(i, " >> ")
        s.Print(FormulaBin(data))
        i += 1
    }
    for _, data := range s.formulas {
        s.writer.Print(i, " >> ")
        s.Print(data.formula)
        i += 1
    }
}

func (s *SimpleDB) PrintData(data []DEPByte) {
    formula := s.FormulaDep(data)
    s.Print(formula)
}

func (s *SimpleDB) FullPrintData(data []DEPByte) {
    formula := s.FormulaDepWithShorts(data)
    s.FullPrint(formula)
}
