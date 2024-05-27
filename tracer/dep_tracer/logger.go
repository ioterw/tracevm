package dep_tracer

import (
    "strings"
    "strconv"
    "math/big"
    "encoding/hex"
    "encoding/json"
    "github.com/holiman/uint256"

    "github.com/ethereum/go-ethereum/common"
)

type LoggerDefinition struct {
    OpcodesShort      []string       `json:"opcodes_short"`
    OpcodesFull       []string       `json:"opcodes"`
    opcodesShort      map[uint64]bool
    opcodesFull       map[uint64]bool

    FinalSlotsShort   bool           `json:"final_slots_short"`
    FinalSlotsFull    bool           `json:"final_slots"`
    CodesShort        bool           `json:"codes_short"`
    CodesFull         bool           `json:"codes"`
    ReturnDataShort   bool           `json:"return_data_short"`
    ReturnDataFull    bool           `json:"return_data"`
    LogsShort         bool           `json:"logs_short"`
    LogsFull          bool           `json:"logs"`
    SolViewFinalSlots bool           `json:"sol_view"`
}

func NewLoggerDefinition(ld *LoggerDefinition) *LoggerDefinition {
    if ld == nil {
        ld.OpcodesShort      = []string{}
        ld.OpcodesFull       = []string{}
        ld.FinalSlotsShort   = true
        ld.FinalSlotsFull    = true
        ld.CodesShort        = false
        ld.CodesFull         = false
        ld.ReturnDataShort   = false
        ld.ReturnDataFull    = true
        ld.LogsShort         = false
        ld.LogsFull          = true
        ld.SolViewFinalSlots = true
    }
    ld.opcodesShort      = map[uint64]bool{}
    ld.opcodesFull       = map[uint64]bool{}
    for _, op := range ld.OpcodesShort {
        val, err := strconv.ParseUint(op, 16, 16)
        if err != nil {
            panic(err)
        }
        ld.opcodesShort[val] = true
    }
    for _, op := range ld.OpcodesFull {
        val, err := strconv.ParseUint(op, 16, 16)
        if err != nil {
            panic(err)
        }
        ld.opcodesFull[val] = true
    }
    return ld
}

func (ld *LoggerDefinition) OpcodeFull(opcode uint8) bool {
    _, ok := ld.opcodesFull[uint64(opcode)]
    return ok
}

func (ld *LoggerDefinition) OpcodeShort(opcode uint8) bool {
    _, ok := ld.opcodesShort[uint64(opcode)]
    return ok
}


type LoggerContext struct {
    block          *big.Int
    timestamp      uint64
    origin         common.Address
    txHash         common.Hash
    address        common.Address
    addressVersion uint64
    codeAddress    common.Address
    codeHash       common.Hash
    initcodeHash   common.Hash
}

type Logger struct {
    simpleDB *SimpleDB
    toLog    LoggerDefinition
    context  LoggerContext
    writer   OutputWriter
}

func NewLogger(simpleDB *SimpleDB, toLog LoggerDefinition, writer OutputWriter) Logger {
    l := Logger{}
    l.simpleDB = simpleDB
    l.toLog = toLog
    l.writer = writer
    return l
}

func (l *Logger) EnterContext(block *big.Int, timestamp uint64, origin common.Address, txHash common.Hash) {
    l.context.block     = block
    l.context.timestamp = timestamp
    l.context.origin    = origin
    l.context.txHash    = txHash
}

func (l *Logger) SetContractAddress(address common.Address, addressVersion uint64, codeAddress common.Address, codeHash, initcodeHash common.Hash) {
    l.context.address        = address
    l.context.addressVersion = addressVersion
    l.context.codeAddress    = codeAddress
    l.context.codeHash       = codeHash
    l.context.initcodeHash   = initcodeHash
}

func (l *Logger) LogLog(log Log) {
    eventType := "log"
    fullEnabled := l.toLog.LogsFull
    shortEnabled := l.toLog.LogsShort
    l.logFormulasWithShorts(eventType, log.addr, log.addrVersion, log.codeAddr, append([]Formula{log.data}, log.topics...), fullEnabled, shortEnabled)
}

func (l *Logger) LogReturnData(addr common.Address, addrVersion uint64, codeAddress common.Address, val []DEPByte) {
    eventType := "return"
    fullEnabled := l.toLog.ReturnDataFull
    shortEnabled := l.toLog.ReturnDataShort
    l.logFormulasWithShorts(eventType, addr, addrVersion, codeAddress, []Formula{l.simpleDB.FormulaDepWithShorts(val)}, fullEnabled, shortEnabled)
}

func (l *Logger) LogFinalCode(addr common.Address, addrVersion uint64, codeAddress common.Address, val []DEPByte) {
    eventType := "final_code"
    fullEnabled := l.toLog.CodesFull
    shortEnabled := l.toLog.CodesShort
    l.logFormulasWithShorts(eventType, addr, addrVersion, codeAddress, []Formula{l.simpleDB.FormulaDepWithShorts(val)}, fullEnabled, shortEnabled)
}

func (l *Logger) LogFinalSlot(addr common.Address, addrVersion uint64, codeAddress common.Address, val []DEPByte, slot *uint256.Int) {

    eventType := "final_slot"
    fullEnabled := l.toLog.FinalSlotsFull
    shortEnabled := l.toLog.FinalSlotsShort
    l.logFormulasWithShorts(eventType, addr, addrVersion, codeAddress, []Formula{l.simpleDB.FormulaDepWithShorts(val)}, fullEnabled, shortEnabled)
}

func (l *Logger) LogOpcode(formula Formula) {
    eventType := "opcode"
    fullEnabled := l.toLog.OpcodeFull(formula.opcode)
    shortEnabled := l.toLog.OpcodeShort(formula.opcode)
    l.logFormulasWithShorts(eventType, l.context.address, l.context.addressVersion, l.context.codeAddress, []Formula{formula}, fullEnabled, shortEnabled)
}

func (l *Logger) logFormulasWithShorts(eventType string, addr common.Address, addrVersion uint64, codeAddr common.Address, formulas []Formula, fullEnabled, shortEnabled bool) {
    outputFormulas := make(map[string][]Formula)
    if fullEnabled {
        outputFormulas["full"] = formulas
    }
    if shortEnabled {
        for _, short := range l.simpleDB.shorts {
            shortFormulas := []Formula{}
            for _, formula := range formulas {
                shortHash := short.LoadChildHash(formula.hash).hash
                shortFormula := l.simpleDB.GetFormula(shortHash)
                shortFormulas = append(shortFormulas, shortFormula)
            }
            outputFormulas[short.protected.name] = shortFormulas
        }
    }
    if len(outputFormulas) > 0 {
        l.logFormulas(eventType, addr, addrVersion, codeAddr, outputFormulas)
    }
}

func sstoreSolidity(s *SimpleDB, formula Formula) {
    if formula.opcode != OPSStore && formula.opcode != OPSLoad {
        panic("Trying to solidify strage opcode")
    }
    solView := SolViewNew(s, s.GetFormula(formula.operands[1]))
    s.writer.Println("## SOLIDITY")
    solView.Print(s.writer)
    s.writer.Println("# -         ", hex.EncodeToString(s.GetFormula(formula.operands[0]).result))
}

func (l *Logger) logFormulas(
    eventType string,
    addr common.Address, addrVersion uint64,
    codeAddr common.Address,
    outputFormulas map[string][]Formula,
) { 
    type MessageJSON struct {
        EventType      string `json:"event_type"`
        ShortTypes     map[string][]string `json:"short_types"`
        Block          string `json:"block"`
        TxHash         string `json:"txhash"`
        Timestamp      uint64 `json:"timestamp"`
        Origin         string `json:"origin"`
        Address        string `json:"address"`
        AddressVersion uint64 `json:"address_version"`
        CodeAddress    string `json:"code_address"`
        CodeHash       string `json:"code_hash"`
        InitcodeHash   string `json:"initcode_hash"`
    }

    outputHashes := make(map[string][]string)
    for shortType, formulas := range outputFormulas {
        formulaHashes := []string{}
        for _, v := range formulas {
            h := v.hash
            formulaHashes = append(formulaHashes, hex.EncodeToString(h[:]))
        }
        outputHashes[shortType] = formulaHashes
    }

    message := MessageJSON{}
    message.EventType      = eventType
    message.ShortTypes     = outputHashes
    message.Block          = l.context.block.String()
    message.TxHash         = hex.EncodeToString(l.context.txHash[:])
    message.Timestamp      = l.context.timestamp
    message.Origin         = hex.EncodeToString(l.context.origin[:])
    message.Address        = hex.EncodeToString(addr[:])
    message.AddressVersion = addrVersion
    message.CodeAddress    = hex.EncodeToString(codeAddr[:])
    message.CodeHash       = hex.EncodeToString(l.context.codeHash[:])
    message.InitcodeHash   = hex.EncodeToString(l.context.initcodeHash[:])

    encodedMessage, err := json.MarshalIndent(message, "", "  ")
    if err != nil {
        panic(err)
    }

    l.writer.Println("## INFO")
    l.writer.Println(string(encodedMessage))
    if eventType == "final_slot" && l.toLog.SolViewFinalSlots {
        cryptoFormula := outputFormulas["crypto"][0]
        sstoreSolidity(l.simpleDB, cryptoFormula)
    }
    for shortType, formulas := range outputFormulas {
        if shortType == "full" {
            continue
        }
        for _, formula := range formulas {
            l.writer.Println("##", strings.ToUpper(shortType))
            l.simpleDB.Print(formula)
        }
    }
    if formulas, ok := outputFormulas["full"]; ok {
        for _, formula := range formulas {
            l.writer.Println("## FULL")
            l.simpleDB.Print(formula)
        }
    }
    l.writer.Println()
}
