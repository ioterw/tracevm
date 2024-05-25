package dep_tracer

import (
    "fmt"
    "encoding/hex"
    "github.com/holiman/uint256"

    "github.com/ethereum/go-ethereum/common"
)

type Log struct {
    addr        common.Address
    addrVersion uint64
    codeAddr    common.Address
    data        Formula
    topics      []Formula
}


type TransactionState struct {
    overlayDB *OverlayDB
    stacked   *Stacked
    logs      []Log
}

func transactionStateNew(simpleDB *SimpleDB, isCreate bool, addr, codeAddr common.Address) *TransactionState {
    ts := new(TransactionState)
    ts.overlayDB = OverlayDBNew(simpleDB)
    addrVersion := ts.overlayDB.GetAddressVersion(addr)
    ts.stacked = StackedNew(isCreate, addr, addrVersion, codeAddr, []DEPByte{}, []DEPByte{}, common.Hash{}, common.Hash{})
    ts.logs = make([]Log, 0)
    return ts
}

func (ts *TransactionState) Copy() *TransactionState {
    res := new(TransactionState)
    res.overlayDB = ts.overlayDB.Copy()
    res.stacked = ts.stacked.Copy()
    res.logs = make([]Log, len(ts.logs))
    copy(res.logs, ts.logs)
    return res
}

func (ts *TransactionState) AddLog(addr common.Address, addrVersion uint64, codeAddr common.Address, data Formula, topics []Formula) {
    log := Log{}
    log.addr = addr
    log.addrVersion = addrVersion
    log.codeAddr = codeAddr
    log.data = data
    log.topics = make([]Formula, len(topics))
    copy(log.topics, topics)
    ts.logs = append(ts.logs, log)
}

func (ts *TransactionState) CommitLogs() {
    for _, log := range ts.logs {
        ts.overlayDB.simpleDB.CommitFormulaWithShorts(log.data.hash)
        for _, topic := range log.topics {
            ts.overlayDB.simpleDB.CommitFormulaWithShorts(topic.hash)
        }
    }
}

func (ts *TransactionState) PrintLogs() {
    fmt.Println("-- LOGS --")
    for _, log := range ts.logs {
        fmt.Println("[", hex.EncodeToString(log.addr[:]), "]")
        fmt.Println("[ data ]")
        ts.overlayDB.FullPrint(log.data)
        for i, topic := range log.topics {
            fmt.Println("[ topic", i, "]")
            ts.overlayDB.FullPrint(topic)
        }
    }
}

type TransactionDB struct {
    simpleDB   *SimpleDB
    states     []*TransactionState
    returndata []DEPByte
}

func TransactionDBCall(simpleDB *SimpleDB, addr, codeAddr common.Address, calldataBin []byte) *TransactionDB {
    t := new(TransactionDB)
    t.simpleDB = simpleDB
    t.states = []*TransactionState{transactionStateNew(simpleDB, false, addr, codeAddr)}

    calldata := FormulaDEPBytes(simpleDB.ConstantNewWithShorts(OPCallData, calldataBin))
    t.Call(addr, addr, calldata)

    return t
}

func TransactionDBCreate(simpleDB *SimpleDB, addr, codeAddr common.Address, initcodeBin []byte) *TransactionDB {
    t := new(TransactionDB)
    t.simpleDB = simpleDB
    t.states = []*TransactionState{transactionStateNew(simpleDB, true, addr, codeAddr)}

    initcode := FormulaDEPBytes(simpleDB.ConstantNewWithShorts(OPInitCode, initcodeBin))
    t.Create(addr, codeAddr, initcode, initcodeBin)

    return t
}

func (t *TransactionDB) Commit() {
    if !t.IsCreate() {
        t.simpleDB.logger.LogReturnData(t.Address(), t.AddressVersion(), t.CodeAddress(), t.returndata)
    }
    for _, log := range t.curState().logs {
        t.simpleDB.logger.LogLog(log)
    }
    
    t.curState().overlayDB.Commit()
    t.curState().CommitLogs()
    t.simpleDB.CommitDEPBytesWithShorts(t.returndata)

    t.simpleDB.ResetFormulas()
}

func (t *TransactionDB) Call(addr, codeAddr common.Address, calldata []DEPByte) {
    t.dupState()
    t.returndata = make([]DEPByte, 0)
    addrVersion := t.GetAddressVersion(addr)
    t.curState().stacked.Push(false, addr, addrVersion, codeAddr, calldata, t.GetCode(codeAddr), t.GetCodeHash(codeAddr), t.GetInitcodeHash(codeAddr))
}

func (t *TransactionDB) Create(addr, codeAddr common.Address, initcode []DEPByte, initcodeBin []byte) {
    t.dupState()
    t.returndata = make([]DEPByte, 0)
    codeHash := CodeHash(initcodeBin)
    addrVersion := t.GetAddressVersion(addr)
    t.curState().stacked.Push(true, addr, addrVersion, codeAddr, make([]DEPByte, 0), initcode, codeHash, codeHash)
}

func (t *TransactionDB) Revert(returndata []DEPByte) {
    t.popState()
    t.returndata = CopyDEPBytes(returndata)
}

func (t *TransactionDB) Return(returndata []DEPByte, returndataBytes []byte) {
    if t.IsCreate() {
        t.SetCode(returndata, returndataBytes, t.CodeHash())
        returndata = []DEPByte{}
    }
    t.returndata = CopyDEPBytes(returndata)
    t.setState(t.popState())
    t.curState().stacked.Pop()
}

func (t *TransactionDB) Selfdestruct() {
    t.curState().overlayDB.Destruct(t.Address())
    t.Return([]DEPByte{}, []byte{})
}

func (t *TransactionDB) Created(addr common.Address) bool {
    return t.curState().overlayDB.Created(addr)
}

func (t *TransactionDB) curState() *TransactionState {
    return t.states[len(t.states)-1]
}

func (t *TransactionDB) setState(state *TransactionState) {
    t.states[len(t.states)-1] = state
}

func (t *TransactionDB) dupState() {
    state := t.curState().Copy()
    t.states = append(t.states, state)
}

func (t *TransactionDB) popState() *TransactionState {
    res := t.curState()
    t.states = t.states[:len(t.states)-1]
    return res
}

func (t *TransactionDB) GetAddressVersion(addr common.Address) uint64 {
    return t.curState().overlayDB.GetAddressVersion(addr)
}

func (t *TransactionDB) GetSlot(slot *uint256.Int) []DEPByte {
    return t.curState().overlayDB.GetSlot(t.Address(), slot).data
}

func (t *TransactionDB) SetSlot(slot *uint256.Int, val []DEPByte) {
    t.curState().overlayDB.SetSlot(t.Address(), t.CodeAddress(), slot, val)
}

func (t *TransactionDB) GetTransient(slot *uint256.Int) []DEPByte {
    return t.curState().overlayDB.GetTransient(t.Address(), slot)
}

func (t *TransactionDB) SetTransient(slot *uint256.Int, val []DEPByte) {
    t.curState().overlayDB.SetTransient(t.Address(), slot, val)
}

func (t *TransactionDB) GetCode(addr common.Address) []DEPByte {
    return t.curState().overlayDB.GetCode(addr).data
}

func (t *TransactionDB) GetCodeHash(addr common.Address) common.Hash {
    return t.curState().overlayDB.GetCode(addr).codeHash
}

func (t *TransactionDB) GetInitcodeHash(addr common.Address) common.Hash {
    return t.curState().overlayDB.GetCode(addr).initcodeHash
}

func (t *TransactionDB) SetCode(val []DEPByte, valBytes []byte, initcodeHash common.Hash) {
    t.curState().overlayDB.SetCode(t.Address(), t.CodeAddress(), val, valBytes, initcodeHash)
}

func (t *TransactionDB) Address() common.Address {
    return t.curState().stacked.Cur().addr
}

func (t *TransactionDB) AddressVersion() uint64 {
    return t.curState().stacked.Cur().addrVersion
}

func (t *TransactionDB) CodeAddress() common.Address {
    return t.curState().stacked.Cur().codeAddr
}

func (t *TransactionDB) CodeHash() common.Hash {
    return t.curState().stacked.Cur().codeHash
}

func (t *TransactionDB) InitcodeHash() common.Hash {
    return t.curState().stacked.Cur().initcodeHash
}

func (t *TransactionDB) IsCreate() bool {
    return t.curState().stacked.Cur().isCreate
}

func (t *TransactionDB) Calldata() []DEPByte {
    return CopyDEPBytes(t.curState().stacked.Cur().calldata)
}

func (t *TransactionDB) Code() []DEPByte {
    return CopyDEPBytes(t.curState().stacked.Cur().code)
}

func (t *TransactionDB) Stack() *Stack {
    return t.curState().stacked.Cur().stack
}

func (t *TransactionDB) Memory() *Memory {
    return t.curState().stacked.Cur().memory
}

func (t *TransactionDB) AddLog(data Formula, topics []Formula) {
    t.curState().AddLog(t.Address(), t.AddressVersion(), t.CodeAddress(), data, topics)
}

func (t *TransactionDB) Returndata() []DEPByte {
    return CopyDEPBytes(t.returndata)
}

func (t *TransactionDB) ConstantNewWithShorts(opcode uint8, result []byte) Formula {
    return t.simpleDB.ConstantNewWithShorts(opcode, result)
}

func (t *TransactionDB) FormulaNewWithShorts(opcode uint8, result []byte, operands []common.Hash) Formula {
    return t.simpleDB.FormulaNewWithShorts(opcode, result, operands)
}

func (t *TransactionDB) FormulaDepWithShorts(val []DEPByte) Formula {
    return t.simpleDB.FormulaDepWithShorts(val)
}

func (t *TransactionDB) Print(f Formula) {
    t.simpleDB.Print(f)
}

func (t *TransactionDB) FullPrint(f Formula) {
    t.simpleDB.FullPrint(f)
}

func (t *TransactionDB) PrintData(data []DEPByte) {
    t.simpleDB.PrintData(data)
}

func (t *TransactionDB) FullPrintData(data []DEPByte) {
    t.simpleDB.FullPrintData(data)
}

func (t *TransactionDB) PrintResults() {
    t.curState().overlayDB.PrintCommit()
    t.curState().PrintLogs()
    fmt.Println("-- RETURN --")
    t.simpleDB.FullPrintData(t.returndata)
}

func (t *TransactionDB) DebugPrintState() {
    for i, el := range t.curState().stacked.elements {
        if i == 0 {
            continue
        }
        fmt.Println("================================= ", i, " =================================")
        fmt.Println("> ", i," | address ", el.addr, " (",el.codeAddr,"/",el.addrVersion,") | create ", el.isCreate, " | ")

        fmt.Print("CALLDATA  : "); t.PrintData(el.calldata)
        fmt.Print("CODE      : "); t.PrintData(el.code)
        fmt.Print("MEMORY    : "); t.PrintData(el.memory.Data)
        fmt.Println("STACK     : ")
        for j, d := range el.stack.Data {
            fmt.Print("== ", j, ": "); t.PrintData(d[:])
        }
        fmt.Println("--------------------------------- ", i, " ---------------------------------")
    }
    fmt.Println("=======================================================================")
    fmt.Print("RETURNDATA: "); t.PrintData(t.returndata)
    fmt.Println("-----------------------------------------------------------------------")
}
