package live

import (
    "encoding/json"
    "encoding/hex"
    "math/big"
    "reflect"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/tracing"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/params"
    "github.com/ethereum/go-ethereum/core/vm"
    "github.com/ethereum/go-ethereum/eth/tracers"
    "github.com/ethereum/go-ethereum/crypto"

    "github.com/ethereum/go-ethereum/eth/tracers/live/dep_tracer"
)

func init() {
    tracers.LiveDirectory.Register("dep", newDep)
}

type Dep struct {
    writingBlock bool
    db                    *dep_tracer.SimpleDB
    state                 *dep_tracer.TransactionDB
    prevOPHandler         dep_tracer.OPHandler
    opHandlers            map[byte]dep_tracer.OPHandler
    pcHandlers            map[common.Address]dep_tracer.PrecompileHandler
    retHandlers           []dep_tracer.OPHandler
    chainConfig           *params.ChainConfig
    stateDB               tracing.StateDB
    returnHandled         bool
    returnAddress         common.Address
    returnInput           []byte
    blockNumber           *big.Int
    time                  uint64
    selfdestructProtector bool
}

func newDep(cfg json.RawMessage) (*tracing.Hooks, error) {
    type depTracerConfig struct {
        KV struct {
            Engine string `json:"engine"`
            Root   string `json:"root"`
        } `json:"kv"`
        Logger *dep_tracer.LoggerDefinition `json:"logger,omitempty"`
        Output string `json:"output"`
    }

    var config depTracerConfig
    if cfg != nil {
        if err := json.Unmarshal(cfg, &config); err != nil {
            return nil, fmt.Errorf("failed to parse config: %v", err)
        }
    }
    if config.KV.Engine == "" {
        panic("kv engine is not set")
    }
    if config.KV.Root == "" {
        panic("kv root (path) is not set")
    }
    var writer dep_tracer.OutputWriter
    if config.Output == "" {
        writer = dep_tracer.NewStdoutWriter()
    } else {
        writer = dep_tracer.NewFileWriter(config.Output)
    }

    db := dep_tracer.SetupDB(
        config.KV.Engine,
        config.KV.Root,
        config.Logger,
        writer,
    )

    t := &Dep{
        writingBlock:          false,
        db:                    db,
        state:                 nil,
        prevOPHandler:         nil,
        opHandlers:            dep_tracer.NewOPHandlers(),
        pcHandlers:            dep_tracer.NewPrecompileHandlers(),
        retHandlers:           []dep_tracer.OPHandler{},
        chainConfig:           nil,
        stateDB:               nil,
        returnHandled:         false,
        returnAddress:         common.Address{},
        returnInput:           nil,
        blockNumber:           nil,
        time:                  0,
        selfdestructProtector: false,
    }
    return &tracing.Hooks{
        OnBlockStart: t.OnBlockStart,
        OnBlockEnd: t.OnBlockEnd,
        OnTxStart: t.OnTxStart,
        OnTxEnd: t.OnTxEnd,
        OnOpcode: t.OnOpcode,
        OnEnter: t.OnEnter,
        OnFault: t.OnFault,
        OnExit: t.OnExit,
    }, nil
}

func (t *Dep) OnBlockStart(ev tracing.BlockEvent) {
    if t.writingBlock {
        panic("OnBlockStart called during writingBlock state")
    }
    // fmt.Println("OnBlockStart")
    t.writingBlock = true
    t.blockNumber = ev.Block.Number()
    t.time = ev.Block.Time()
}

func (t *Dep) OnBlockEnd(err error) {
    if !t.writingBlock {
        return
    }
    // fmt.Println("OnBlockEnd")
    t.writingBlock = false
}

func (t *Dep) OnTxStart(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
    if !t.writingBlock {
        return
    }
    // fmt.Println("OnTxStart")

    create := tx.To() == nil
    var addr common.Address
    if create {
        addr = crypto.CreateAddress(from, tx.Nonce())
    } else {
        addr = *tx.To()
    }

    startData := dep_tracer.DataStart {
        IsCreate: create,
        Address: addr,
        Input: tx.Data(),
        Block: vm.BlockNumber,
        Timestamp: vm.Time,
        Origin: from,
        TxHash: tx.Hash(),
    }
    t.state = dep_tracer.TransactionStart(t.db, startData)
    t.chainConfig = vm.ChainConfig
    t.stateDB = vm.StateDB
}

func (t *Dep) OnTxEnd(receipt *types.Receipt, err error) {
    if t.state == nil {
        return
    }
    // fmt.Println("OnTxEnd")

    dep_tracer.TransactionFinish(t.state)
    t.state = nil
}

func (t *Dep) OnOpcode(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
    if t.state == nil {
        return
    }

    // normal execution

    stack := scope.StackData()
    stackSize := len(stack)

    if t.prevOPHandler != nil {
        t.prevOPHandler.After(t.db, t.state, stack, stackSize, t.stateDB, t.chainConfig, t.blockNumber, t.time, pc, op, scope)
        t.prevOPHandler = nil
    }

    // fmt.Println("OnOpcode", vm.OpCode(op).String(), err, gas, cost)

    // t.state.DebugPrintState()

    if err != nil {
        dep_tracer.DataError{
            Reverted: true,
        }.Handle(t.db, t.state)
        t.returnHandled = true
        return
    }

    o := vm.OpCode(op)
    invalid := cost == 0 && o != vm.STOP && o != vm.RETURN && o != vm.REVERT
    if invalid {
        return
    }
    if o == vm.SELFDESTRUCT {
        t.selfdestructProtector = true
    }

    if opHandler, ok := t.opHandlers[op]; ok {
        direction := opHandler.Before(t.db, t.state, stack, stackSize, t.stateDB, t.chainConfig, t.blockNumber, t.time, pc, op, scope)
        switch direction {
        case dep_tracer.DIRECTION_NONE:
            t.prevOPHandler = opHandler
        case dep_tracer.DIRECTION_RETURN:
            t.returnHandled = true
        case dep_tracer.DIRECTION_CALL:
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
            opHandlerCopy := clone(opHandler).(dep_tracer.OPHandler)
            t.retHandlers = append(t.retHandlers, opHandlerCopy)
        default:
            panic("Unknown direction")
        }
        return
    }

    panic(fmt.Sprintf("Unknown opcode %s", vm.OpCode(op).String()))
}

func (t *Dep) OnEnter(depth int, typ byte, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
    if t.state == nil {
        return
    }
    if t.selfdestructProtector {
        return
    }
    // fmt.Println("OnEnter")
    t.returnAddress = to
    t.returnInput = input
}

func (t *Dep) OnFault(pc uint64, op byte, gas, cost uint64, _ tracing.OpContext, depth int, err error) {
    if t.state == nil {
        return
    }
    // fmt.Println("OnFault", vm.OpCode(op).String(), err)

    o := vm.OpCode(op)
    if o == vm.REVERT {
        return
    }

    dep_tracer.DataError{
        Reverted: true,
    }.Handle(t.db, t.state)
    t.returnHandled = true
}

func (t *Dep) OnExit(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
    if t.state == nil {
        return
    }
    if t.selfdestructProtector {
        t.selfdestructProtector = false
        return
    }
    // fmt.Println("OnExit", err)

    if !t.returnHandled {
        if len(output) > 0 {
            if ph, ok := t.pcHandlers[t.returnAddress]; ok {
                ph.Execute(t.db, t.state, t.returnInput, output)
            } else {
                panic(fmt.Sprintf("Unknown precompile %", hex.EncodeToString(t.returnAddress[:])))
            }
        } else {
            dep_tracer.DataError{
                Reverted: err != nil,
            }.Handle(t.db, t.state)
        }
    }
    t.returnHandled = false

    if len(t.retHandlers) == 0 {
        return
    }

    newSize := len(t.retHandlers) - 1
    handler := t.retHandlers[newSize]
    t.retHandlers = t.retHandlers[:newSize]
    handler.Exit(t.db, t.state, err == nil)
}
