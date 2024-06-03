package dep_tracer

import (
    "fmt"
    "strings"
    "reflect"
    "math/big"
    "encoding/hex"
    "encoding/json"
    "github.com/holiman/uint256"
    
    "github.com/ethereum/go-ethereum/core/vm"
)

type DepHandler struct {
    // flow variables
    returnHandled bool
    activated     bool

    // dep_tracer variables
    db            *SimpleDB
    state         *TransactionDB
    prevOPHandler OPHandler
    opHandlers    map[byte]OPHandler
    pcHandlers    map[Address]PrecompileHandler
    retHandlers   []OPHandler

    // input variables
    stateDB            StateDB
    returnAddress      Address
    returnInput        []byte
    isRandom           bool
    isSelfdestruct6780 bool
}

func NewDepHandler(cfg json.RawMessage) *DepHandler {
    type depTracerConfig struct {
        KV struct {
            Engine  string `json:"engine"`
            Root    string `json:"root"`
        } `json:"kv"`
        Logger      *LoggerDefinition `json:"logger,omitempty"`
        Output      string            `json:"output"`
        PastUnknown bool              `json:"past_unknown"`
    }

    var config depTracerConfig
    if cfg != nil {
        if err := json.Unmarshal(cfg, &config); err != nil {
            panic(fmt.Errorf("failed to parse config: %v", err))
        }
    }
    if config.KV.Engine == "" {
        panic("kv engine is not set")
    }
    if config.KV.Engine != "memory" && config.KV.Engine != "amnesia" {
        if config.KV.Root == "" {
            panic("kv root (path) is not set")
        }
    }
    var writer OutputWriter
    if config.Output == "" {
        writer = NewStdoutWriter()
    } else if strings.HasPrefix(config.Output, "http://") {
        writer = NewHttpWriter(config.Output)
    } else {
        writer = NewFileWriter(config.Output)
    }

    db := SetupDB(
        config.KV.Engine,
        config.KV.Root,
        config.Logger,
        config.PastUnknown,
        writer,
    )

    return &DepHandler{
        returnHandled: false,
        activated:     false,

        db:            db,
        state:         nil,
        prevOPHandler: nil,
        opHandlers:    NewOPHandlers(),
        pcHandlers:    NewPrecompileHandlers(),
        retHandlers:   []OPHandler{},

        stateDB:       nil,
        returnAddress: Address{},
        returnInput:   nil,
    }
}

type StateDB interface {
    GetNonce(addr [20]byte) uint64
    GetCode(addr [20]byte) []byte
}

func (handler *DepHandler) StartTransactionRecording(
    isCreate bool, addr [20]byte, input []byte, block *big.Int,
    timestamp uint64, origin [20]byte, txHash [32]byte,
    code []byte, isSelfdestruct6780, isRandom bool, stateDB StateDB,
) {
    if handler.activated {
        panic("StartTransactionRecording activated twice")
    }
    handler.activated = true;

    startData := DataStart {
        IsCreate: isCreate,
        Address: addr,
        Input: input,
        Block: block,
        Timestamp: timestamp,
        Origin: origin,
        TxHash: txHash,
        Code: code,
    }
    handler.state = TransactionStart(handler.db, startData)
    handler.isSelfdestruct6780 = isSelfdestruct6780
    handler.isRandom = isRandom
    handler.stateDB = stateDB
}

func (handler *DepHandler) EndTransactionRecording() {
    if !handler.activated {
        panic("EndTransactionRecording is not activated")
    }
    handler.activated = false;

    TransactionFinish(handler.state)
    handler.state = nil
}

func (handler *DepHandler) HandleOpcode(
    stack []uint256.Int, memory []byte, addr [20]byte,
    pc uint64, op byte, isInvalid bool, hasError bool,
) {
    if !handler.activated {
        panic("HandleOpcode is not activated")
    }

    stackSize := len(stack)

    if handler.prevOPHandler != nil {
        handler.prevOPHandler.After(
            handler.db, handler.state, stack, stackSize, handler.stateDB,
            handler.isSelfdestruct6780, handler.isRandom, pc, op, addr, memory,
        )
        handler.prevOPHandler = nil
    }

    if hasError {
        DataError {
            Reverted: true,
        }.Handle(handler.db, handler.state)
        handler.returnHandled = true
        return
    }

    if isInvalid {
        return
    }

    if opHandler, ok := handler.opHandlers[op]; ok {
        direction := opHandler.Before(handler.db, handler.state, stack, stackSize, handler.stateDB, handler.isSelfdestruct6780, handler.isRandom, pc, op, addr, memory)
        switch direction {
        case DIRECTION_NONE:
            handler.prevOPHandler = opHandler
        case DIRECTION_RETURN:
            handler.returnHandled = true
        case DIRECTION_CALL:
            // https://gist.github.com/penglongli/a0698546f1026731c81c0d327a3d6b32
            clone := func(inter any) any {
                nInter := reflect.New(reflect.TypeOf(inter).Elem())

                val := reflect.ValueOf(inter).Elem()
                nVal := nInter.Elem()
                for i := 0; i < val.NumField(); i++ {
                    nvField := nVal.Field(i)
                    nvField.Set(val.Field(i))
                }

                return nInter.Interface()
            }
            opHandlerCopy := clone(opHandler).(OPHandler)
            handler.retHandlers = append(handler.retHandlers, opHandlerCopy)
        default:
            panic("Unknown direction")
        }
        return
    }
}

func (handler *DepHandler) HandleEnter(to [20]byte, input []byte) {
    if !handler.activated {
        panic("HandleEnter is not activated")
    }

    handler.returnAddress = to
    handler.returnInput = input
}

func (handler *DepHandler) HandleFault(op byte) {
    if !handler.activated {
        panic("HandleFault is not activated")
    }

    o := vm.OpCode(op)
    if o == vm.REVERT {
        return
    }

    DataError {
        Reverted: true,
    }.Handle(handler.db, handler.state)
    handler.returnHandled = true
}

func (handler *DepHandler) HandleExit(output []byte, hasError bool) {
    if !handler.activated {
        panic("HandleExit is not activated")
    }

    if !handler.returnHandled {
        if len(output) > 0 {
            if ph, ok := handler.pcHandlers[handler.returnAddress]; ok {
                ph.Execute(handler.db, handler.state, handler.returnInput, output)
            } else {
                panic(fmt.Sprintf("Unknown precompile %", hex.EncodeToString(handler.returnAddress[:])))
            }
        } else {
            DataError {
                Reverted: hasError,
            }.Handle(handler.db, handler.state)
        }
    }
    handler.returnHandled = false

    if len(handler.retHandlers) == 0 {
        return
    }

    newSize := len(handler.retHandlers) - 1
    h := handler.retHandlers[newSize]
    handler.retHandlers = handler.retHandlers[:newSize]
    h.Exit(handler.db, handler.state, !hasError)
}

