package live

import (
    "math/big"
    "encoding/json"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/core/vm"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/eth/tracers"
    "github.com/ethereum/go-ethereum/core/tracing"

    "github.com/ethereum/go-ethereum/eth/tracers/live/dep_tracer"
)

func init() {
    tracers.LiveDirectory.Register("dep", NewDep)
}

type Dep struct {
    handler               *dep_tracer.DepHandler
    writingBlock          bool
    transacting           bool
    selfdestructProtector bool
    blockNumber           *big.Int
    time                  uint64
}

func NewDep(cfg json.RawMessage) (*tracing.Hooks, error) {
    t := &Dep{
        writingBlock:          false,
        transacting:           false,
        selfdestructProtector: false,
        handler:               dep_tracer.NewDepHandler(cfg),
        blockNumber:           nil,
        time:                  0,
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
    t.writingBlock = true

    t.blockNumber = ev.Block.Number()
    t.time = ev.Block.Time()
}

func (t *Dep) OnBlockEnd(err error) {
    if !t.writingBlock {
        return
    }
    t.writingBlock = false
}

func (t *Dep) OnTxStart(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
    if !t.writingBlock {
        return
    }
    t.transacting = true

    create := tx.To() == nil
    var addr common.Address
    var code []byte
    if create {
        addr = crypto.CreateAddress(from, tx.Nonce())
        code = nil
    } else {
        addr = *tx.To()
        code = vm.StateDB.GetCode(addr)
    }

    isSelfdestruct6780 := vm.ChainConfig.IsCancun(t.blockNumber, t.time)
    isRandom := vm.ChainConfig.IsLondon(t.blockNumber)

    t.handler.StartTransactionRecording(create, addr, tx.Data(), vm.BlockNumber, vm.Time, from, tx.Hash(), code, isSelfdestruct6780, isRandom, vm.StateDB)
}

func (t *Dep) OnTxEnd(receipt *types.Receipt, err error) {
    if !t.transacting {
        return
    }
    t.transacting = false

    t.handler.EndTransactionRecording()
}

func (t *Dep) OnOpcode(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
    if !t.transacting {
        return
    }

    o := vm.OpCode(op)
    hasError := err != nil
    isInvalid := cost == 0 && o != vm.STOP && o != vm.RETURN && o != vm.REVERT

    if !hasError && !isInvalid && o == vm.SELFDESTRUCT {
        t.selfdestructProtector = true
    }

    t.handler.HandleOpcode(
        scope.StackData(), scope.MemoryData(), scope.Address(),
        pc, op, isInvalid, hasError,
    )
}

func (t *Dep) OnEnter(depth int, typ byte, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
    if !t.transacting {
        return
    }
    if t.selfdestructProtector {
        return
    }

    t.handler.HandleEnter(to, input)
}

func (t *Dep) OnFault(pc uint64, op byte, gas, cost uint64, _ tracing.OpContext, depth int, err error) {
    if !t.transacting {
        return
    }

    t.handler.HandleFault(op)
}

func (t *Dep) OnExit(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
    if !t.transacting {
        return
    }
    if t.selfdestructProtector {
        t.selfdestructProtector = false
        return
    }

    t.handler.HandleExit(output, err != nil)
}
