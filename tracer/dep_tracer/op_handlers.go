package dep_tracer

import (
    "math/big"
    "github.com/holiman/uint256"

    "github.com/ethereum/go-ethereum/params"
    "github.com/ethereum/go-ethereum/core/tracing"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/core/vm"
)

const (
    DIRECTION_CALL   = iota
    DIRECTION_RETURN = iota
    DIRECTION_NONE   = iota
)

type OPHandler interface {
    Register(map[byte]OPHandler)
    Before(
        db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int,
        stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64,
        pc uint64, op byte, scope tracing.OpContext,
   	) int
    After(
        db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int,
        stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64,
        pc uint64, op byte, scope tracing.OpContext,
   	)
    Exit(
        db *SimpleDB, state *TransactionDB,
        success bool,
    )
}

func NewOPHandlers() map[byte]OPHandler {
    handlers := map[byte]OPHandler{}

    new(StopHandler).Register(handlers)
    new(AddHandler).Register(handlers)
    new(MulHandler).Register(handlers)
    new(SubHandler).Register(handlers)
    new(DivHandler).Register(handlers)
    new(SDivHandler).Register(handlers)
    new(ModHandler).Register(handlers)
    new(SModHandler).Register(handlers)
    new(AddModHandler).Register(handlers)
    new(MulModHandler).Register(handlers)
    new(ExpHandler).Register(handlers)
    new(SignExtendHandler).Register(handlers)

    new(LTHandler).Register(handlers)
    new(GTHandler).Register(handlers)
    new(SLTHandler).Register(handlers)
    new(SGTHandler).Register(handlers)
    new(EQHandler).Register(handlers)
    new(IsZeroHandler).Register(handlers)
    new(AndHandler).Register(handlers)
    new(OrHandler).Register(handlers)
    new(XorHandler).Register(handlers)
    new(NotHandler).Register(handlers)
    new(ByteHandler).Register(handlers)
    new(SHLHandler).Register(handlers)
    new(SHRHandler).Register(handlers)
    new(SARHandler).Register(handlers)

    new(KeccakHandler).Register(handlers)
    new(AddressHandler).Register(handlers)
    new(BalanceHandler).Register(handlers)
    new(OriginHandler).Register(handlers)
    new(CallerHandler).Register(handlers)
    new(CallValueHandler).Register(handlers)
    new(CallDataLoadHandler).Register(handlers)
    new(CallDataSizeHandler).Register(handlers)
    new(CallDataCopyHandler).Register(handlers)
    new(CodeSizeHandler).Register(handlers)
    new(CodeCopyHandler).Register(handlers)
    new(GasPriceHandler).Register(handlers)
    new(ExtCodeSizeHandler).Register(handlers)
    new(ExtCodeCopyHandler).Register(handlers)
    new(ReturnDataSizeHandler).Register(handlers)
    new(ReturnDataCopyHandler).Register(handlers)
    new(ExtCodeHashHandler).Register(handlers)

    new(BlockHashHandler).Register(handlers)
    new(CoinbaseHandler).Register(handlers)
    new(TimestampHandler).Register(handlers)
    new(NumberHandler).Register(handlers)
    new(PrevrandaoOrDifficultyHandler).Register(handlers)
    new(GasLimitHandler).Register(handlers)
    new(ChainIDHandler).Register(handlers)
    new(SelfBalanceHandler).Register(handlers)
    new(BaseFeeHandler).Register(handlers)
    new(BlobHashHandler).Register(handlers)
    new(BlobBaseFeeHandler).Register(handlers)

    new(PopHandler).Register(handlers)
    new(MLoadHandler).Register(handlers)
    new(MStoreHandler).Register(handlers)
    new(SLoadHandler).Register(handlers)
    new(SStoreHandler).Register(handlers)
    new(JumpHandler).Register(handlers)
    new(JumpIHandler).Register(handlers)
    new(PCHandler).Register(handlers)
    new(MSizeHandler).Register(handlers)
    new(GasHandler).Register(handlers)
    new(JumpDestHandler).Register(handlers)
    new(TLoadHandler).Register(handlers)
    new(TStoreHandler).Register(handlers)
    new(MCopyHandler).Register(handlers)

    new(PushHandler).Register(handlers)
    new(DupHandler).Register(handlers)
    new(SwapHandler).Register(handlers)
    new(LogHandler).Register(handlers)

    new(CreateHandler).Register(handlers)
    new(CallHandler).Register(handlers)
    new(CallCodeHandler).Register(handlers)
    new(ReturnHandler).Register(handlers)
    new(DelegateCallHandler).Register(handlers)
    new(Create2Handler).Register(handlers)
    new(StaticCallHandler).Register(handlers)
    new(RevertHandler).Register(handlers)
    new(SelfdestructHandler).Register(handlers)

    return handlers
}


type PushHandler struct {
    data DataPush
}
func (oh *PushHandler) Register(handlers map[byte]OPHandler) {
    for i := vm.PUSH0; i <= vm.PUSH32; i++ {
        handlers[byte(i)] = oh
    }
}
func (oh *PushHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataPush {
        Pc: pc,
        Size: uint64(op) - uint64(vm.PUSH0),
    }

    return DIRECTION_NONE
}
func (oh *PushHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *PushHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type DupHandler struct {
    data DataDup
}
func (oh *DupHandler) Register(handlers map[byte]OPHandler) {
    for i := vm.DUP1; i <= vm.DUP16; i++ {
        handlers[byte(i)] = oh
    }
}
func (oh *DupHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataDup {
        Size: 1 + int(op) - int(vm.DUP1),
    }

    return DIRECTION_NONE
}
func (oh *DupHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *DupHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SwapHandler struct {
    data DataSwap
}
func (oh *SwapHandler) Register(handlers map[byte]OPHandler) {
    for i := vm.SWAP1; i <= vm.SWAP16; i++ {
        handlers[byte(i)] = oh
    }
}
func (oh *SwapHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataSwap {
        Size: 2 + int64(op) - int64(vm.SWAP1),
    }

    return DIRECTION_NONE
}
func (oh *SwapHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *SwapHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type MStoreHandler struct {
    data DataMStore
}
func (oh *MStoreHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.MSTORE)] = oh
}
func (oh *MStoreHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataMStore {
        Offset: stack[stackSize-1].Uint64(),
    }

    return DIRECTION_NONE
}
func (oh *MStoreHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *MStoreHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type MLoadHandler struct {
    data DataMLoad
}
func (oh *MLoadHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.MLOAD)] = oh
}
func (oh *MLoadHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataMLoad {
        Offset: stack[stackSize-1].Uint64(),
    }
    return DIRECTION_NONE
}
func (oh *MLoadHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *MLoadHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type GasHandler struct {}
func (oh *GasHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.GAS)] = oh
}
func (oh *GasHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *GasHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPGas,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *GasHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type CallValueHandler struct {}
func (oh *CallValueHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CALLVALUE)] = oh
}
func (oh *CallValueHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *CallValueHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPCallValue,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *CallValueHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type AddressHandler struct {}
func (oh *AddressHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.ADDRESS)] = oh
}
func (oh *AddressHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *AddressHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant20 {
        Op: OPAddress,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *AddressHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type IsZeroHandler struct {}
func (oh *IsZeroHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.ISZERO)] = oh
}
func (oh *IsZeroHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *IsZeroHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataOne {
        Op: OPIsZero,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *IsZeroHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type NotHandler struct {}
func (oh *NotHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.NOT)] = oh
}
func (oh *NotHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *NotHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataOne {
        Op: OPNot,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *NotHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ByteHandler struct {
    data DataByte
}
func (oh *ByteHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.BYTE)] = oh
}
func (oh *ByteHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataByte {
        Offset: stack[stackSize-1],
    }
    return DIRECTION_NONE
}
func (oh *ByteHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *ByteHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type JumpHandler struct {
    data DataEmpty
}
func (oh *JumpHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.JUMP)] = oh
}
func (oh *JumpHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataEmpty {
        N: 1,
    }
    
    return DIRECTION_NONE
}
func (oh *JumpHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *JumpHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type JumpIHandler struct {
    data DataEmpty
}
func (oh *JumpIHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.JUMPI)] = oh
}
func (oh *JumpIHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataEmpty {
        N: 2,
    }
    
    return DIRECTION_NONE
}
func (oh *JumpIHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *JumpIHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type JumpDestHandler struct {
    data DataEmpty
}
func (oh *JumpDestHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.JUMPDEST)] = oh
}
func (oh *JumpDestHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataEmpty {
        N: 0,
    }
    
    return DIRECTION_NONE
}
func (oh *JumpDestHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *JumpDestHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type PopHandler struct {
    data DataPop
}
func (oh *PopHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.POP)] = oh
}
func (oh *PopHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataPop {}
    
    return DIRECTION_NONE
}
func (oh *PopHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *PopHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type CodeCopyHandler struct {
    data DataCodeCopy
}
func (oh *CodeCopyHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CODECOPY)] = oh
}
func (oh *CodeCopyHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    uint64CodeOffset, overflow := stack[stackSize-2].Uint64WithOverflow()
    if overflow {
        uint64CodeOffset = 0xffffffffffffffff
    }
    oh.data = DataCodeCopy {
        MemoryOffset: stack[stackSize-1].Uint64(),
        CodeOffset  : uint64CodeOffset,
        Length      : stack[stackSize-3].Uint64(),
    }

    return DIRECTION_NONE
}
func (oh *CodeCopyHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *CodeCopyHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ExtCodeSizeHandler struct {
    data DataExtCodeSize
}
func (oh *ExtCodeSizeHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.EXTCODESIZE)] = oh
}
func (oh *ExtCodeSizeHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataExtCodeSize {
        Address: stack[stackSize-1].Bytes20(),
    }

    return DIRECTION_NONE
}
func (oh *ExtCodeSizeHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.CodeSize = stack[stackSize-1]
    oh.data.Handle(db, state)
}
func (oh *ExtCodeSizeHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type RevertHandler struct {}
func (oh *RevertHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.REVERT)] = oh
}
func (oh *RevertHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    DataRevert {
        Offset: stack[stackSize-1].Uint64(),
        Size: stack[stackSize-2].Uint64(),
    }.Handle(db, state)

    return DIRECTION_RETURN
}
func (oh *RevertHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *RevertHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ReturnHandler struct {}
func (oh *ReturnHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.RETURN)] = oh
}
func (oh *ReturnHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    offset := stack[stackSize-1].Uint64()
    size := stack[stackSize-2].Uint64()

    destOffset := offset + size
    extraZeros := uint64(0)
    memory := scope.MemoryData()
    if destOffset > uint64(len(memory)) {
        extraZeros = destOffset - uint64(len(memory))
        destOffset = uint64(len(memory))
    }
    result := memory[offset:destOffset]
    if extraZeros > 0 {
        result = append(result, make([]byte, extraZeros)...)
    }
    
    DataReturn {
        Offset: offset,
        Size: size,
        Result: result,
    }.Handle(db, state)

    return DIRECTION_RETURN
}
func (oh *ReturnHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *ReturnHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type StopHandler struct {}
func (oh *StopHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.STOP)] = oh
}
func (oh *StopHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    DataStop {}.Handle(db, state)

    return DIRECTION_RETURN
}
func (oh *StopHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *StopHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SLoadHandler struct {
    data DataSLoad
}
func (oh *SLoadHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SLOAD)] = oh
}
func (oh *SLoadHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataSLoad {
        Slot: stack[stackSize-1],
    }
    return DIRECTION_NONE
}
func (oh *SLoadHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Value = stack[stackSize-1]
    oh.data.Handle(db, state)
}
func (oh *SLoadHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SStoreHandler struct {
    data DataSStore
}
func (oh *SStoreHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SSTORE)] = oh
}
func (oh *SStoreHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataSStore {
        Slot: stack[stackSize-1],
        Value: stack[stackSize-2],
    }
    return DIRECTION_NONE
}
func (oh *SStoreHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *SStoreHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type TLoadHandler struct {
    data DataTLoad
}
func (oh *TLoadHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.TLOAD)] = oh
}
func (oh *TLoadHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataTLoad {
        Slot: stack[stackSize-1],
    }
    return DIRECTION_NONE
}
func (oh *TLoadHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *TLoadHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type TStoreHandler struct {
    data DataTStore
}
func (oh *TStoreHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.TSTORE)] = oh
}
func (oh *TStoreHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataTStore {
        Slot: stack[stackSize-1],
    }
    return DIRECTION_NONE
}
func (oh *TStoreHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *TStoreHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type AddHandler struct {}
func (oh *AddHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.ADD)] = oh
}
func (oh *AddHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *AddHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPAdd,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *AddHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type DivHandler struct {}
func (oh *DivHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.DIV)] = oh
}
func (oh *DivHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *DivHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPDiv,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *DivHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SDivHandler struct {}
func (oh *SDivHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SDIV)] = oh
}
func (oh *SDivHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SDivHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPSDiv,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SDivHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ModHandler struct {}
func (oh *ModHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.MOD)] = oh
}
func (oh *ModHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *ModHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPMod,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *ModHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SModHandler struct {}
func (oh *SModHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SMOD)] = oh
}
func (oh *SModHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SModHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPSMod,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SModHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type AddModHandler struct {}
func (oh *AddModHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.ADDMOD)] = oh
}
func (oh *AddModHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *AddModHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataThree {
        Op: OPAddMod,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *AddModHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type MulModHandler struct {}
func (oh *MulModHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.MULMOD)] = oh
}
func (oh *MulModHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *MulModHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataThree {
        Op: OPMulMod,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *MulModHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ExpHandler struct {}
func (oh *ExpHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.EXP)] = oh
}
func (oh *ExpHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *ExpHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPExp,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *ExpHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SignExtendHandler struct {}
func (oh *SignExtendHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SIGNEXTEND)] = oh
}
func (oh *SignExtendHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SignExtendHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPSignExtend,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SignExtendHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type MulHandler struct {}
func (oh *MulHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.MUL)] = oh
}
func (oh *MulHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *MulHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPMul,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *MulHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SubHandler struct {}
func (oh *SubHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SUB)] = oh
}
func (oh *SubHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SubHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPSub,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SubHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SHLHandler struct {}
func (oh *SHLHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SHL)] = oh
}
func (oh *SHLHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SHLHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPShl,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SHLHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SHRHandler struct {}
func (oh *SHRHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SHR)] = oh
}
func (oh *SHRHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SHRHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPShr,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SHRHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SARHandler struct {}
func (oh *SARHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SAR)] = oh
}
func (oh *SARHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SARHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPSar,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SARHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type AndHandler struct {}
func (oh *AndHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.AND)] = oh
}
func (oh *AndHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *AndHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPAnd,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *AndHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type OrHandler struct {}
func (oh *OrHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.OR)] = oh
}
func (oh *OrHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *OrHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPOr,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *OrHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type XorHandler struct {}
func (oh *XorHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.XOR)] = oh
}
func (oh *XorHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *XorHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPXor,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *XorHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type GTHandler struct {}
func (oh *GTHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.GT)] = oh
}
func (oh *GTHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *GTHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPGt,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *GTHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type EQHandler struct {}
func (oh *EQHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.EQ)] = oh
}
func (oh *EQHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *EQHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPEq,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *EQHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type LTHandler struct {}
func (oh *LTHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.LT)] = oh
}
func (oh *LTHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *LTHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPLt,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *LTHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SLTHandler struct {}
func (oh *SLTHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SLT)] = oh
}
func (oh *SLTHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SLTHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPSlt,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SLTHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SGTHandler struct {}
func (oh *SGTHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SGT)] = oh
}
func (oh *SGTHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SGTHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataTwo {
        Op: OPSgt,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SGTHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type KeccakHandler struct {
    data DataKeccak
}
func (oh *KeccakHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.KECCAK256)] = oh
}
func (oh *KeccakHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataKeccak {
        Offset: stack[stackSize-1].Uint64(),
        Size: stack[stackSize-2].Uint64(),
    }

    return DIRECTION_NONE
}
func (oh *KeccakHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Result = stack[stackSize-1].Bytes32()
    oh.data.Handle(db, state)
}
func (oh *KeccakHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type CallDataSizeHandler struct {
    data DataCalldataSize
}
func (oh *CallDataSizeHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CALLDATASIZE)] = oh
}
func (oh *CallDataSizeHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataCalldataSize {}

    return DIRECTION_NONE
}
func (oh *CallDataSizeHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.CalldataSize = stack[stackSize-1].Uint64()
    oh.data.Handle(db, state)
}
func (oh *CallDataSizeHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type CallDataCopyHandler struct {
    data DataCalldataCopy
}
func (oh *CallDataCopyHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CALLDATACOPY)] = oh
}
func (oh *CallDataCopyHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    dataOffset64, overflow := stack[stackSize-2].Uint64WithOverflow()
    if overflow {
        dataOffset64 = 0xffffffffffffffff
    }
    oh.data = DataCalldataCopy {
        MemoryOffset: stack[stackSize-1].Uint64(),
        DataOffset: dataOffset64,
        Size: stack[stackSize-3].Uint64(),
    }

    return DIRECTION_NONE
}
func (oh *CallDataCopyHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *CallDataCopyHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type CallDataLoadHandler struct {
    data DataCalldataLoad
}
func (oh *CallDataLoadHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CALLDATALOAD)] = oh
}
func (oh *CallDataLoadHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    offset64, overflow := stack[stackSize-1].Uint64WithOverflow()
    if overflow {
        offset64 = 0xffffffffffffffff
    }
    oh.data = DataCalldataLoad {
        Offset: offset64,
    }

    return DIRECTION_NONE
}
func (oh *CallDataLoadHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *CallDataLoadHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ReturnDataSizeHandler struct {}
func (oh *ReturnDataSizeHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.RETURNDATASIZE)] = oh
}
func (oh *ReturnDataSizeHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {    
    return DIRECTION_NONE
}
func (oh *ReturnDataSizeHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataReturndataSize {
        ReturndataSize: stack[stackSize-1].Uint64(),
    }.Handle(db, state)
}
func (oh *ReturnDataSizeHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type BalanceHandler struct {}
func (oh *BalanceHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.BALANCE)] = oh
}
func (oh *BalanceHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *BalanceHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataBalance {
        Balance: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *BalanceHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ExtCodeCopyHandler struct {
    data DataExtCodeCopy
}
func (oh *ExtCodeCopyHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.EXTCODECOPY)] = oh
}
func (oh *ExtCodeCopyHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    uint64CodeOffset, overflow := stack[stackSize-3].Uint64WithOverflow()
    if overflow {
        uint64CodeOffset = 0xffffffffffffffff
    }
    oh.data = DataExtCodeCopy {
        Address: stack[stackSize-1].Bytes20(),
        MemoryOffset: stack[stackSize-2].Uint64(),
        CodeOffset: uint64CodeOffset,
        Length: stack[stackSize-4].Uint64(),
    }

    return DIRECTION_NONE
}
func (oh *ExtCodeCopyHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *ExtCodeCopyHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ReturnDataCopyHandler struct {
    data DataReturndataCopy
}
func (oh *ReturnDataCopyHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.RETURNDATACOPY)] = oh
}
func (oh *ReturnDataCopyHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataReturndataCopy {
        MemoryOffset: stack[stackSize-1].Uint64(),
        DataOffset: stack[stackSize-2].Uint64(),
        Size: stack[stackSize-3].Uint64(),
    }

    return DIRECTION_NONE
}
func (oh *ReturnDataCopyHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *ReturnDataCopyHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type OriginHandler struct {}
func (oh *OriginHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.ORIGIN)] = oh
}
func (oh *OriginHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *OriginHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant20 {
        Op: OPOrigin,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *OriginHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type CallerHandler struct {}
func (oh *CallerHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CALLER)] = oh
}
func (oh *CallerHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *CallerHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant20 {
        Op: OPCaller,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *CallerHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type CodeSizeHandler struct {}
func (oh *CodeSizeHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CODESIZE)] = oh
}
func (oh *CodeSizeHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *CodeSizeHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataCodeSize {
        CodeSize: stack[stackSize-1].Uint64(),
    }.Handle(db, state)
}
func (oh *CodeSizeHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type GasPriceHandler struct {}
func (oh *GasPriceHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.GASPRICE)] = oh
}
func (oh *GasPriceHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *GasPriceHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *GasPriceHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ExtCodeHashHandler struct {
    data DataExtCodeHash
}
func (oh *ExtCodeHashHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.EXTCODEHASH)] = oh
}
func (oh *ExtCodeHashHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataExtCodeHash {
        Address: stack[stackSize-1].Bytes20(),
    }

    return DIRECTION_NONE
}
func (oh *ExtCodeHashHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Hash = stack[stackSize-1].Bytes32()
    oh.data.Handle(db, state)
}
func (oh *ExtCodeHashHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type BlockHashHandler struct {}
func (oh *BlockHashHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.BLOCKHASH)] = oh
}
func (oh *BlockHashHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *BlockHashHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataBlockHash {
        Hash: stack[stackSize-1].Bytes32(),
    }.Handle(db, state)
}
func (oh *BlockHashHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type CoinbaseHandler struct {}
func (oh *CoinbaseHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.COINBASE)] = oh
}
func (oh *CoinbaseHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *CoinbaseHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant20 {
        Op: OPCoinbase,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *CoinbaseHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type TimestampHandler struct {}
func (oh *TimestampHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.TIMESTAMP)] = oh
}
func (oh *TimestampHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *TimestampHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPTimestamp,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *TimestampHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type NumberHandler struct {}
func (oh *NumberHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.NUMBER)] = oh
}
func (oh *NumberHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *NumberHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPNumber,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *NumberHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type GasLimitHandler struct {}
func (oh *GasLimitHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.GASLIMIT)] = oh
}
func (oh *GasLimitHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *GasLimitHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPGasLimit,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *GasLimitHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type ChainIDHandler struct {}
func (oh *ChainIDHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CHAINID)] = oh
}
func (oh *ChainIDHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *ChainIDHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPChainID,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *ChainIDHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SelfBalanceHandler struct {}
func (oh *SelfBalanceHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SELFBALANCE)] = oh
}
func (oh *SelfBalanceHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *SelfBalanceHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataSelfBalance {
        Balance: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *SelfBalanceHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type BaseFeeHandler struct {}
func (oh *BaseFeeHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.BASEFEE)] = oh
}
func (oh *BaseFeeHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *BaseFeeHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPBaseFee,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *BaseFeeHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type PCHandler struct {}
func (oh *PCHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.PC)] = oh
}
func (oh *PCHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *PCHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPPc,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *PCHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type MSizeHandler struct {}
func (oh *MSizeHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.MSIZE)] = oh
}
func (oh *MSizeHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *MSizeHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPMsize,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *MSizeHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type LogHandler struct {
    data DataLog
}
func (oh *LogHandler) Register(handlers map[byte]OPHandler) {
    for i := vm.LOG0; i <= vm.LOG4; i++ {
        handlers[byte(i)] = oh
    }
}
func (oh *LogHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataLog {
        Offset: stack[stackSize-1].Uint64(),
        Size: stack[stackSize-2].Uint64(),
        TopicsNum: int(op) - int(vm.LOG0),
    }
    return DIRECTION_NONE
}
func (oh *LogHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *LogHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type CallHandler struct {
    DataEnd DataCallEnd
}
func (oh *CallHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CALL)] = oh
}
func (oh *CallHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    DataCallStart {
        N: 7,
        Address: stack[stackSize-2].Bytes20(),
        CodeAddress: stack[stackSize-2].Bytes20(),
        InOffset: stack[stackSize-4].Uint64(),
        InSize: stack[stackSize-5].Uint64(),
    }.Handle(db, state)

    oh.DataEnd = DataCallEnd {
        ReturnOffset: stack[stackSize-6].Uint64(),
        ReturnSize: stack[stackSize-7].Uint64(),
    }

    return DIRECTION_CALL
}
func (oh *CallHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *CallHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {
    oh.DataEnd.Success = success
    oh.DataEnd.Handle(db, state)
}


type CallCodeHandler struct {
    DataEnd DataCallEnd
}
func (oh *CallCodeHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CALLCODE)] = oh
}
func (oh *CallCodeHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    DataCallStart {
        N: 7,
        Address: scope.Address(),
        CodeAddress: stack[stackSize-2].Bytes20(),
        InOffset: stack[stackSize-4].Uint64(),
        InSize: stack[stackSize-5].Uint64(),
    }.Handle(db, state)

    oh.DataEnd = DataCallEnd {
        ReturnOffset: stack[stackSize-6].Uint64(),
        ReturnSize: stack[stackSize-7].Uint64(),
    }

    return DIRECTION_CALL
}
func (oh *CallCodeHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *CallCodeHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {
    oh.DataEnd.Success = success
    oh.DataEnd.Handle(db, state)
}


type DelegateCallHandler struct {
    DataEnd DataCallEnd
}
func (oh *DelegateCallHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.DELEGATECALL)] = oh
}
func (oh *DelegateCallHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    DataCallStart {
        N: 6,
        Address: scope.Address(),
        CodeAddress: stack[stackSize-2].Bytes20(),
        InOffset: stack[stackSize-3].Uint64(),
        InSize: stack[stackSize-4].Uint64(),
    }.Handle(db, state)

    oh.DataEnd = DataCallEnd {
        ReturnOffset: stack[stackSize-5].Uint64(),
        ReturnSize: stack[stackSize-6].Uint64(),
    }

    return DIRECTION_CALL
}
func (oh *DelegateCallHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *DelegateCallHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {
    oh.DataEnd.Success = success
    oh.DataEnd.Handle(db, state)
}


type StaticCallHandler struct {
    DataEnd DataCallEnd
}
func (oh *StaticCallHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.STATICCALL)] = oh
}
func (oh *StaticCallHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    DataCallStart {
        N: 6,
        Address: stack[stackSize-2].Bytes20(),
        CodeAddress: stack[stackSize-2].Bytes20(),
        InOffset: stack[stackSize-3].Uint64(),
        InSize: stack[stackSize-4].Uint64(),
    }.Handle(db, state)

    oh.DataEnd = DataCallEnd {
        ReturnOffset: stack[stackSize-5].Uint64(),
        ReturnSize: stack[stackSize-6].Uint64(),
    }

    return DIRECTION_CALL
}
func (oh *StaticCallHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *StaticCallHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {
    oh.DataEnd.Success = success
    oh.DataEnd.Handle(db, state)
}


type CreateHandler struct {
    DataEnd DataCreateEnd
}
func (oh *CreateHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CREATE)] = oh
}
func (oh *CreateHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    addr := crypto.CreateAddress(scope.Address(), stateDB.GetNonce(scope.Address())) // 1 is nonce
    
    offset := stack[stackSize-2].Uint64()
    size := stack[stackSize-3].Uint64()

    destOffset := offset + size
    extraZeros := uint64(0)
    memory := scope.MemoryData()
    if destOffset > uint64(len(memory)) {
        extraZeros = destOffset - uint64(len(memory))
        destOffset = uint64(len(memory))
    }
    result := memory[offset:destOffset]
    if extraZeros > 0 {
        result = append(result, make([]byte, extraZeros)...)
    }

    DataCreateStart {
        Address: addr,
        Offset: offset,
        Size: size,
        Data: result,
    }.Handle(db, state)

    oh.DataEnd = DataCreateEnd {
        Address: addr,
    }

    return DIRECTION_CALL
}
func (oh *CreateHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *CreateHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {
    oh.DataEnd.Handle(db, state)
}


type Create2Handler struct {
    DataEnd DataCreate2End
}
func (oh *Create2Handler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.CREATE2)] = oh
}
func (oh *Create2Handler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    offset := stack[stackSize-2].Uint64()
    size := stack[stackSize-3].Uint64()

    destOffset := offset + size
    extraZeros := uint64(0)
    memory := scope.MemoryData()
    if destOffset > uint64(len(memory)) {
        extraZeros = destOffset - uint64(len(memory))
        destOffset = uint64(len(memory))
    }
    result := memory[offset:destOffset]
    if extraZeros > 0 {
        result = append(result, make([]byte, extraZeros)...)
    }

    salt := stack[stackSize-4]
    addr := crypto.CreateAddress2(scope.Address(), salt.Bytes32(), crypto.Keccak256Hash(result).Bytes())
    
    DataCreate2Start {
        Address: addr,
        Offset: offset,
        Size: size,
        Data: result,
    }.Handle(db, state)

    oh.DataEnd = DataCreate2End {
        Address: addr,
    }

    return DIRECTION_CALL
}
func (oh *Create2Handler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *Create2Handler) Exit(db *SimpleDB, state *TransactionDB, success bool) {
    oh.DataEnd.Handle(db, state)
}


type MCopyHandler struct {
    data DataMCopy
}
func (oh *MCopyHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.MCOPY)] = oh
}
func (oh *MCopyHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    oh.data = DataMCopy {
        ToOffset: stack[stackSize-1].Uint64(),
        FromOffset: stack[stackSize-2].Uint64(),
        Size: stack[stackSize-3].Uint64(),
    }

    return DIRECTION_NONE
}
func (oh *MCopyHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    oh.data.Handle(db, state)
}
func (oh *MCopyHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type BlobBaseFeeHandler struct {}
func (oh *BlobBaseFeeHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.BLOBBASEFEE)] = oh
}
func (oh *BlobBaseFeeHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *BlobBaseFeeHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataConstant {
        Op: OPBlobBaseFee,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *BlobBaseFeeHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type BlobHashHandler struct {}
func (oh *BlobHashHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.BLOBHASH)] = oh
}
func (oh *BlobHashHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *BlobHashHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    DataBlobHash {
        Hash: stack[stackSize-1].Bytes32(),
    }.Handle(db, state)
}
func (oh *BlobHashHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type PrevrandaoOrDifficultyHandler struct {
    // difficulty || random - since london
}
func (oh *PrevrandaoOrDifficultyHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.PREVRANDAO)] = oh
}
func (oh *PrevrandaoOrDifficultyHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    return DIRECTION_NONE
}
func (oh *PrevrandaoOrDifficultyHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {
    var operand uint8
    if chainConfig.IsLondon(number) {
        operand = OPRandom
    } else {
        operand = OPDifficulty
    }
    DataConstant {
        Op: operand,
        Value: stack[stackSize-1],
    }.Handle(db, state)
}
func (oh *PrevrandaoOrDifficultyHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}


type SelfdestructHandler struct {
    // selfdestruct || selfdestruct6780 - since cancun
}
func (oh *SelfdestructHandler) Register(handlers map[byte]OPHandler) {
    handlers[byte(vm.SELFDESTRUCT)] = oh
}
func (oh *SelfdestructHandler) Before(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) int {
    if chainConfig.IsCancun(number, time) {
        DataSelfdestruct6780 {}.Handle(db, state)
    } else {
        DataSelfdestruct {}.Handle(db, state)
    }

    return DIRECTION_RETURN
}
func (oh *SelfdestructHandler) After(db *SimpleDB, state *TransactionDB, stack []uint256.Int, stackSize int, stateDB tracing.StateDB, chainConfig *params.ChainConfig, number *big.Int, time uint64, pc uint64, op byte, scope tracing.OpContext) {}
func (oh *SelfdestructHandler) Exit(db *SimpleDB, state *TransactionDB, success bool) {}
