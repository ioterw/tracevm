package dep_tracer

import (
    "encoding/binary"

    "github.com/ethereum/go-ethereum/common"
)

func SetupDB(kvEngine, kvRoot string, toLog *LoggerDefinition, writer OutputWriter) *SimpleDB {
    protected := []ProtectedDefinition{}
    protected = append(protected, CryptoProtectedDefinition())

    if toLog == nil {
        toLog = NewLoggerDefinition()
        // toLog.AddOpcdesShort([]uint8{OPSLoad, OPSStore}) // it can be reverted
        // toLog.AddOpcdesFull([]uint8{OPSLoad, OPSStore})
        toLog.FinalSlotsShort = true
        toLog.FinalSlotsFull = true
        // toLog.CodesShort = true
        // toLog.CodesFull = true
        // toLog.ReturnDataShort = true
        toLog.ReturnDataFull  = true
        toLog.LogsFull = true
        // toLog.LogsShort  = true
        toLog.SolViewFinalSlots = true
    }

    return SimpleDBNew(
        protected, *toLog,
        kvEngine, kvRoot,
        writer,
    )
}

func TransactionStart(db *SimpleDB, data DataStart) *TransactionDB {
    var state *TransactionDB
    if data.IsCreate {
        state = TransactionDBCreate(db, data.Address, common.Address{}, data.Input)
    } else {
        state = TransactionDBCall(db, data.Address, data.Address, data.Input)
    }

    db.logger.EnterContext(data.Block, data.Timestamp, data.Origin, data.TxHash)
    db.logger.SetContractAddress(state.Address(), state.AddressVersion(), state.CodeAddress(), state.CodeHash(), state.InitcodeHash())

    return state
}

func TransactionFinish(state *TransactionDB) {
    state.Commit()
}

func (data DataStart) Handle(db *SimpleDB, state *TransactionDB) {
    panic("DataStart shouldn't be called")
}

func (data DataError) Handle(db *SimpleDB, state *TransactionDB) {
    if data.Reverted {
        state.Revert([]DEPByte{})
    } else {
        state.Return([]DEPByte{}, []byte{})
    }
}

func (data DataPush) Handle(db *SimpleDB, state *TransactionDB) {
    if data.Size == 0 {
        state.Stack().PushN([]DEPByte{})
        return
    }
    code := state.Code()
    val := OverflowSliceDEPBytes(code, data.Pc+1, data.Size)
    state.Stack().PushN(val)
}

func (data DataDup) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Dup(data.Size)
}

func (data DataSwap) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Swap(int(data.Size))
}

func (data DataPop) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop()
}

func (data DataMLoad) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // offset
    val := state.Memory().Load(data.Offset, 32)
    state.Stack().PushN(val)
}

func (data DataMStore) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // offset
    value := state.Stack().Pop() 
    state.Memory().Set32(data.Offset, value)
}

func (data DataMStore8) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // offset
    value := state.Stack().Pop()
    valueByte := value[31]
    state.Memory().Set(data.Offset, valueByte)
}

func (data DataMCopy) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // toOffset
    state.Stack().Pop() // fromOffset
    state.Stack().Pop() // size

    d := state.Memory().Load(data.FromOffset, data.Size)
    state.Memory().SetN(data.ToOffset, d)
}

func (data DataConstant) Handle(db *SimpleDB, state *TransactionDB) {
    valBin := data.Value.Bytes32()
    val := state.ConstantNewWithShorts(data.Op, valBin[:])
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataConstant20) Handle(db *SimpleDB, state *TransactionDB) {
    valBin := data.Value.Bytes20()
    val := state.ConstantNewWithShorts(data.Op, valBin[:])
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataSLoad) Handle(db *SimpleDB, state *TransactionDB) {
    slot := state.Stack().Pop()
    slotFormula := state.FormulaDepWithShorts(slot[:])

    value := state.GetSlot(&data.Slot)
    valueBin := data.Value.Bytes32()
    valueFormula := state.FormulaDepWithShorts(value[:])

    val := state.FormulaNewWithShorts(OPSLoad, valueBin[:], []common.Hash{valueFormula.hash, slotFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataSStore) Handle(db *SimpleDB, state *TransactionDB) {
    slot := state.Stack().Pop()
    slotFormula := state.FormulaDepWithShorts(slot[:])

    value := state.Stack().Pop()
    valueBin := data.Value.Bytes32()
    valueFormula := state.FormulaDepWithShorts(value[:])

    val := state.FormulaNewWithShorts(OPSStore, valueBin[:], []common.Hash{valueFormula.hash, slotFormula.hash})
    state.SetSlot(&data.Slot, FormulaDEPBytes(val))
}

func (data DataTLoad) Handle(db *SimpleDB, state *TransactionDB) {
    value := state.GetTransient(&data.Slot)
    state.Stack().PushN(value)
}

func (data DataTStore) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // slot
    value := state.Stack().Pop()
    state.SetTransient(&data.Slot, value[:])
}

func (data DataOne) Handle(db *SimpleDB, state *TransactionDB) { // aNum
    a := state.Stack().Pop()
    aFormula := state.FormulaDepWithShorts(a[:])

    valBin := data.Value.Bytes32()
    val := state.FormulaNewWithShorts(data.Op, valBin[:], []common.Hash{aFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataTwo) Handle(db *SimpleDB, state *TransactionDB) {
    a := state.Stack().Pop()
    aFormula := state.FormulaDepWithShorts(a[:])

    b := state.Stack().Pop()
    bFormula := state.FormulaDepWithShorts(b[:])

    valBin := data.Value.Bytes32()
    val := state.FormulaNewWithShorts(data.Op, valBin[:], []common.Hash{aFormula.hash, bFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataThree) Handle(db *SimpleDB, state *TransactionDB) {
    a := state.Stack().Pop()
    aFormula := state.FormulaDepWithShorts(a[:])

    b := state.Stack().Pop()
    bFormula := state.FormulaDepWithShorts(b[:])

    c := state.Stack().Pop()
    cFormula := state.FormulaDepWithShorts(c[:])

    valBin := data.Value.Bytes32()
    val := state.FormulaNewWithShorts(data.Op, valBin[:], []common.Hash{aFormula.hash, bFormula.hash, cFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataByte) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // offset
    value := state.Stack().Pop()

    offset := data.Offset

    offset64, overflow := offset.Uint64WithOverflow()
    if overflow || offset64 >= 32 {
        offset64 = 32
    }
    val := OverflowSliceDEPBytes(value[:], offset64, 1)

    state.Stack().PushN(val)
}

func (data DataKeccak) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // offset
    state.Stack().Pop() // size

    d := state.Memory().Load(data.Offset, data.Size)
    dataFormula := state.FormulaDepWithShorts(d)
    val := state.FormulaNewWithShorts(OPKeccak, data.Result[:], []common.Hash{dataFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataCodeSize) Handle(db *SimpleDB, state *TransactionDB) {
    codeFormula := state.FormulaDepWithShorts(state.Code())

    codeSizeBin := []byte{}
    codeSizeBin = binary.BigEndian.AppendUint64(codeSizeBin, data.CodeSize)

    val := state.FormulaNewWithShorts(OPSize, codeSizeBin, []common.Hash{codeFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataExtCodeSize) Handle(db *SimpleDB, state *TransactionDB) {
    addr := state.Stack().Pop()
    addrFormula := state.FormulaDepWithShorts(addr[32-20:])

    addrBin := data.Address
    codeFormula := state.FormulaDepWithShorts(state.GetCode(addrBin))

    codeSizeBin := data.CodeSize.Bytes32()

    val := state.FormulaNewWithShorts(OPCodeSize, codeSizeBin[:], []common.Hash{codeFormula.hash, addrFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataExtCodeHash) Handle(db *SimpleDB, state *TransactionDB) {
    addr := state.Stack().Pop()
    addrFormula := state.FormulaDepWithShorts(addr[32-20:])

    addrBin := data.Address
    codeFormula := state.FormulaDepWithShorts(state.GetCode(addrBin))

    hashBin := data.Hash

    val := state.FormulaNewWithShorts(OPCodeKeccak, hashBin[:], []common.Hash{codeFormula.hash, addrFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataCalldataSize) Handle(db *SimpleDB, state *TransactionDB) {
    sizeBin := []byte{}
    sizeBin = binary.BigEndian.AppendUint64(sizeBin, data.CalldataSize)

    dataFormula := state.FormulaDepWithShorts(state.Calldata())

    val := state.FormulaNewWithShorts(OPSize, sizeBin, []common.Hash{dataFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataReturndataSize) Handle(db *SimpleDB, state *TransactionDB) {
    sizeBin := []byte{}
    sizeBin = binary.BigEndian.AppendUint64(sizeBin, data.ReturndataSize)

    dataFormula := state.FormulaDepWithShorts(state.returndata)

    val := state.FormulaNewWithShorts(OPSize, sizeBin, []common.Hash{dataFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataCodeCopy) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // memOffset
    state.Stack().Pop() // codeOffset
    state.Stack().Pop() // length

    val := OverflowSliceDEPBytes(state.Code(), data.CodeOffset, data.Length)
    state.Memory().SetN(data.MemoryOffset, val)
}

func (data DataExtCodeCopy) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // addr
    state.Stack().Pop() // memOffset
    state.Stack().Pop() // codeOffset
    state.Stack().Pop() // length

    val := OverflowSliceDEPBytes(state.GetCode(data.Address), data.CodeOffset, data.Length)
    state.Memory().SetN(data.MemoryOffset, val)
}

func (data DataCalldataCopy) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // destOffset
    state.Stack().Pop() // offset
    state.Stack().Pop() // size

    d := OverflowSliceDEPBytes(state.Calldata(), data.DataOffset, data.Size)
    state.Memory().SetN(data.MemoryOffset, d)
}

func (data DataReturndataCopy) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // destOffset
    state.Stack().Pop() // offset
    state.Stack().Pop() // size

    d := OverflowSliceDEPBytes(state.returndata, data.DataOffset, data.Size)
    state.Memory().SetN(data.MemoryOffset, d)
}

func (data DataCalldataLoad) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // offset
    val := OverflowSliceDEPBytes(state.Calldata(), data.Offset, 32)
    state.Stack().PushN(val)
}

func (data DataLog) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // offset
    state.Stack().Pop() // size
    d := state.Memory().Load(data.Offset, data.Size)
    dataFormula := state.FormulaDepWithShorts(d[:])
    topicFormulas := make([]Formula, 0)
    for i := 0; i < data.TopicsNum; i++ {
        topic := state.Stack().Pop()
        topicFormula := state.FormulaDepWithShorts(topic[:])
        topicFormulas = append(topicFormulas, topicFormula)
    }
    state.AddLog(dataFormula, topicFormulas)
}

func (data DataReturn) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // offset
    state.Stack().Pop() // size

    val := state.Memory().Load(data.Offset, data.Size)
    state.Return(val, data.Result)
}

func (data DataStop) Handle(db *SimpleDB, state *TransactionDB) {
    state.Return([]DEPByte{}, []byte{})
}

func (data DataSelfdestruct) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // beneficiary

    state.Selfdestruct()
}

func (data DataSelfdestruct6780) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // beneficiary

    if state.Created(state.Address()) {
        state.Selfdestruct()
    } else {
        state.Return([]DEPByte{}, []byte{})
    }
}

func (data DataRevert) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // offset
    state.Stack().Pop() // size

    val := state.Memory().Load(data.Offset, data.Size)
    state.Revert(val)
}

func (data DataEmpty) Handle(db *SimpleDB, state *TransactionDB) {
    for i := 0; i < data.N; i ++ {
        state.Stack().Pop()
    }
}

func (data DataBalance) Handle(db *SimpleDB, state *TransactionDB) {
    balanceBin := data.Balance.Bytes32()
    balance := state.ConstantNewWithShorts(OPConstant, balanceBin[:])

    addr := state.Stack().Pop()
    addrFormula := state.FormulaDepWithShorts(addr[32-20:])

    val := state.FormulaNewWithShorts(OPBalance, balanceBin[:], []common.Hash{balance.hash, addrFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataSelfBalance) Handle(db *SimpleDB, state *TransactionDB) {
    balanceBin := data.Balance.Bytes32()
    balance := state.ConstantNewWithShorts(OPConstant, balanceBin[:])

    addrBin := state.Address()
    addr := state.ConstantNewWithShorts(OPConstant, addrBin[:])

    val := state.FormulaNewWithShorts(OPBalance, balanceBin[:], []common.Hash{balance.hash, addr.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataBlockHash) Handle(db *SimpleDB, state *TransactionDB) {
    hashBin := data.Hash
    hash := state.ConstantNewWithShorts(OPConstant, hashBin[:])

    blockNumber := state.Stack().Pop()
    blockNumberFormula := state.FormulaDepWithShorts(blockNumber[:])

    val := state.FormulaNewWithShorts(OPBlockHash, hashBin[:], []common.Hash{hash.hash, blockNumberFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataBlobHash) Handle(db *SimpleDB, state *TransactionDB) {
    hashBin := data.Hash
    hash := state.ConstantNewWithShorts(OPConstant, hashBin[:])

    blockNumber := state.Stack().Pop()
    blockNumberFormula := state.FormulaDepWithShorts(blockNumber[:])

    val := state.FormulaNewWithShorts(OPBlobHash, hashBin[:], []common.Hash{hash.hash, blockNumberFormula.hash})
    state.Stack().PushN(FormulaDEPBytes(val))
}

func (data DataCreateStart) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // value
    state.Stack().Pop() // offset
    state.Stack().Pop() // size

    initcode := state.Memory().Load(data.Offset, data.Size)

    state.Create(data.Address, common.Address{}, initcode, data.Data)
    db.logger.SetContractAddress(state.Address(), state.AddressVersion(), state.CodeAddress(), state.CodeHash(), state.InitcodeHash())
}

func (data DataCreateEnd) Handle(db *SimpleDB, state *TransactionDB) {
    addrBin := data.Address
    addr := state.ConstantNewWithShorts(OPCreateAddr, addrBin[:])
    state.Stack().PushN(FormulaDEPBytes(addr))
    db.logger.SetContractAddress(state.Address(), state.AddressVersion(), state.CodeAddress(), state.CodeHash(), state.InitcodeHash())
}

func (data DataCreate2Start) Handle(db *SimpleDB, state *TransactionDB) {
    state.Stack().Pop() // value
    state.Stack().Pop() // offset
    state.Stack().Pop() // size
    state.Stack().Pop() // salt

    initcode := state.Memory().Load(data.Offset, data.Size)

    state.Create(data.Address, common.Address{}, initcode, data.Data)
    db.logger.SetContractAddress(state.Address(), state.AddressVersion(), state.CodeAddress(), state.CodeHash(), state.InitcodeHash())
}

func (data DataCreate2End) Handle(db *SimpleDB, state *TransactionDB) {
    addrBin := data.Address
    addr := state.ConstantNewWithShorts(OPCreate2Addr, addrBin[:])
    state.Stack().PushN(FormulaDEPBytes(addr))
    db.logger.SetContractAddress(state.Address(), state.AddressVersion(), state.CodeAddress(), state.CodeHash(), state.InitcodeHash())
}

func (data DataCallStart) Handle(db *SimpleDB, state *TransactionDB) {
    for i := 0; i < data.N; i++ {
        state.Stack().Pop()
    }

    calldata := state.Memory().Load(data.InOffset, data.InSize)

    state.Call(data.Address, data.CodeAddress, calldata)
    db.logger.SetContractAddress(state.Address(), state.AddressVersion(), state.CodeAddress(), state.CodeHash(), state.InitcodeHash())
}

func (data DataCallEnd) Handle(db *SimpleDB, state *TransactionDB) {
    // success bool, retOffset, retSize uint64
    d := OverflowSliceDEPBytes(state.returndata, 0, data.ReturnSize)
    state.Memory().SetN(data.ReturnOffset, d)

    var valBin []byte
    if data.Success {
        valBin = []byte{1}
    } else {
        valBin = []byte{0}
    }
    val := state.ConstantNewWithShorts(OPCallResult, valBin)
    state.Stack().PushN(FormulaDEPBytes(val))
    db.logger.SetContractAddress(state.Address(), state.AddressVersion(), state.CodeAddress(), state.CodeHash(), state.InitcodeHash())
}

func (data DataPrecompileEcRecover) Handle(db *SimpleDB, state *TransactionDB) { // 01
    if len(data.Result) < 1 {
        state.Return([]DEPByte{}, []byte{})
        return
    }

    d    := state.Calldata()
    hash := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 0, 32))
    v    := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 32, 32))
    r    := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 64, 32))
    s    := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 96, 32))

    zeroesNum := 32 - 20 // returns zeroes on the left in the sake of something unclear

    formula := state.FormulaNewWithShorts(OPEcRecover, data.Result[zeroesNum:], []common.Hash{hash.hash, v.hash, r.hash, s.hash})
    args := []common.Hash{}
    h := ConstantInitZero.hash
    for i := 0; i < zeroesNum; i++ {
        args = append(args, h)
    }
    args = append(args, formula.hash)
    val := state.FormulaNewWithShorts(OPConcat, data.Result, args)

    state.Return(FormulaDEPBytes(val), data.Result)
}

func (data DataPrecompileSha256) Handle(db *SimpleDB, state *TransactionDB) { // 02
    d := state.Calldata()

    dataFormula := state.FormulaDepWithShorts(d)
    val := state.FormulaNewWithShorts(OPSha256, data.Result, []common.Hash{dataFormula.hash})

    state.Return(FormulaDEPBytes(val), data.Result)
}

func (data DataPrecompileRipemd160) Handle(db *SimpleDB, state *TransactionDB) { // 03
    zeroesNum := 32 - 20 // returns zeroes on the left in the sake of something unclear

    d := state.Calldata()

    dataFormula := state.FormulaDepWithShorts(d)
    formula := state.FormulaNewWithShorts(OPRipemd160, data.Result[zeroesNum:], []common.Hash{dataFormula.hash})
    args := []common.Hash{}
    h := ConstantInitZero.hash
    for i := 0; i < zeroesNum; i++ {
        args = append(args, h)
    }
    args = append(args, formula.hash)
    val := state.FormulaNewWithShorts(OPConcat, data.Result, args)

    state.Return(FormulaDEPBytes(val), data.Result)
}

func (data DataPrecompileIdentity) Handle(db *SimpleDB, state *TransactionDB) { // 04
    d := state.Calldata()
    state.Return(d, data.Result)
}

func (data DataPrecompileModExp) Handle(db *SimpleDB, state *TransactionDB) { // 05 errors not handled
    // res []byte, bSizeNum, eSizeNum, mSizeNum uint64
    if len(data.Result) < 1 {
        state.Return([]DEPByte{}, []byte{})
        return
    }

    d := state.Calldata()
    i := uint64(96)
    b := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, i, data.BSize))
    i += data.BSize
    e := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, i, data.ESize))
    i += data.ESize
    m := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, i, data.MSize))

    val := state.FormulaNewWithShorts(OPModExp, data.Result, []common.Hash{b.hash, e.hash, m.hash})
    state.Return(FormulaDEPBytes(val), data.Result)
}

func (data DataPrecompileEcAdd) Handle(db *SimpleDB, state *TransactionDB) { // 06
    if len(data.Result) < 1 {
        state.Return([]DEPByte{}, []byte{})
        return
    }

    resX := data.Result[0:32]
    resY := data.Result[32:64]

    d  := state.Calldata()
    x1 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 0,  32))
    y1 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 32, 32))
    x2 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 64, 32))
    y2 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 96, 32))

    valX := state.FormulaNewWithShorts(OPEcAddX, resX, []common.Hash{x1.hash, y1.hash, x2.hash, y2.hash})
    valY := state.FormulaNewWithShorts(OPEcAddY, resY, []common.Hash{x1.hash, y1.hash, x2.hash, y2.hash})

    val := FormulaDEPBytes(valX)
    val = append(val, FormulaDEPBytes(valY)...)
    state.Return(val, data.Result)
}

func (data DataPrecompileEcMul) Handle(db *SimpleDB, state *TransactionDB) { // 07
    if len(data.Result) < 1 {
        state.Return([]DEPByte{}, []byte{})
        return
    }

    resX := data.Result[0:32]
    resY := data.Result[32:64]

    d  := state.Calldata()
    x1 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 0,  32))
    y1 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 32, 32))
    s  := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 64, 32))

    valX := state.FormulaNewWithShorts(OPEcMulX, resX, []common.Hash{x1.hash, y1.hash, s.hash})
    valY := state.FormulaNewWithShorts(OPEcMulY, resY, []common.Hash{x1.hash, y1.hash, s.hash})

    val := FormulaDEPBytes(valX)
    val = append(val, FormulaDEPBytes(valY)...)
    state.Return(val, data.Result)
}

func (data DataPrecompileEcPairing) Handle(db *SimpleDB, state *TransactionDB) { // 08
    if len(data.Result) < 1 {
        state.Return([]DEPByte{}, []byte{})
        return
    }

    d := state.Calldata()
    args := []common.Hash{};
    for i := uint64(0); i < uint64(len(d)); i += 192 {
        x1 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, i,     32))
        y1 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, i+32,  32))
        x2 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, i+64,  32))
        y2 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, i+96,  32))
        x3 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, i+128, 32))
        y3 := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, i+160, 32))
        args = append(args, x1.hash, y1.hash, x2.hash, y2.hash, x3.hash, y3.hash)
    }
    val := state.FormulaNewWithShorts(OPEcPairing, data.Result, args)
    state.Return(FormulaDEPBytes(val), data.Result)
}

func (data DataPrecompileBlake2F) Handle(db *SimpleDB, state *TransactionDB) { // 09
    if len(data.Result) < 1 {
        state.Return([]DEPByte{}, []byte{})
        return
    }
    
    d      := state.Calldata()
    rounds := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 0,   4))
    h      := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 4,   64))
    m      := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 68,  128))
    t      := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 196, 16))
    f      := state.FormulaDepWithShorts(OverflowSliceDEPBytes(d, 212, 1))

    val := state.FormulaNewWithShorts(OPBlake2F, data.Result, []common.Hash{rounds.hash, h.hash, m.hash, t.hash, f.hash})
    state.Return(FormulaDEPBytes(val), data.Result)
}

func (data DataPointEvaluation) Handle(db *SimpleDB, state *TransactionDB) { // 0A
    d := state.Calldata()

    dataFormula := state.FormulaDepWithShorts(d)
    val := state.FormulaNewWithShorts(OPPointEvaluation, data.Result, []common.Hash{dataFormula.hash})

    state.Return(FormulaDEPBytes(val), data.Result)
}
