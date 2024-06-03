package dep_tracer

// OpCode is an EVM opcode
type OpCode byte

// 0x0 range - arithmetic ops.
const (
    STOP       OpCode = 0x0
    ADD        OpCode = 0x1
    MUL        OpCode = 0x2
    SUB        OpCode = 0x3
    DIV        OpCode = 0x4
    SDIV       OpCode = 0x5
    MOD        OpCode = 0x6
    SMOD       OpCode = 0x7
    ADDMOD     OpCode = 0x8
    MULMOD     OpCode = 0x9
    EXP        OpCode = 0xa
    SIGNEXTEND OpCode = 0xb
)

// 0x10 range - comparison ops.
const (
    LT     OpCode = 0x10
    GT     OpCode = 0x11
    SLT    OpCode = 0x12
    SGT    OpCode = 0x13
    EQ     OpCode = 0x14
    ISZERO OpCode = 0x15
    AND    OpCode = 0x16
    OR     OpCode = 0x17
    XOR    OpCode = 0x18
    NOT    OpCode = 0x19
    BYTE   OpCode = 0x1a
    SHL    OpCode = 0x1b
    SHR    OpCode = 0x1c
    SAR    OpCode = 0x1d
)

// 0x20 range - crypto.
const (
    KECCAK256 OpCode = 0x20
)

// 0x30 range - closure state.
const (
    ADDRESS        OpCode = 0x30
    BALANCE        OpCode = 0x31
    ORIGIN         OpCode = 0x32
    CALLER         OpCode = 0x33
    CALLVALUE      OpCode = 0x34
    CALLDATALOAD   OpCode = 0x35
    CALLDATASIZE   OpCode = 0x36
    CALLDATACOPY   OpCode = 0x37
    CODESIZE       OpCode = 0x38
    CODECOPY       OpCode = 0x39
    GASPRICE       OpCode = 0x3a
    EXTCODESIZE    OpCode = 0x3b
    EXTCODECOPY    OpCode = 0x3c
    RETURNDATASIZE OpCode = 0x3d
    RETURNDATACOPY OpCode = 0x3e
    EXTCODEHASH    OpCode = 0x3f
)

// 0x40 range - block operations.
const (
    BLOCKHASH   OpCode = 0x40
    COINBASE    OpCode = 0x41
    TIMESTAMP   OpCode = 0x42
    NUMBER      OpCode = 0x43
    DIFFICULTY  OpCode = 0x44
    RANDOM      OpCode = 0x44 // Same as DIFFICULTY
    PREVRANDAO  OpCode = 0x44 // Same as DIFFICULTY
    GASLIMIT    OpCode = 0x45
    CHAINID     OpCode = 0x46
    SELFBALANCE OpCode = 0x47
    BASEFEE     OpCode = 0x48
    BLOBHASH    OpCode = 0x49
    BLOBBASEFEE OpCode = 0x4a
)

// 0x50 range - 'storage' and execution.
const (
    POP      OpCode = 0x50
    MLOAD    OpCode = 0x51
    MSTORE   OpCode = 0x52
    MSTORE8  OpCode = 0x53
    SLOAD    OpCode = 0x54
    SSTORE   OpCode = 0x55
    JUMP     OpCode = 0x56
    JUMPI    OpCode = 0x57
    PC       OpCode = 0x58
    MSIZE    OpCode = 0x59
    GAS      OpCode = 0x5a
    JUMPDEST OpCode = 0x5b
    TLOAD    OpCode = 0x5c
    TSTORE   OpCode = 0x5d
    MCOPY    OpCode = 0x5e
    PUSH0    OpCode = 0x5f
)

// 0x60 range - pushes.
const (
    PUSH1 OpCode = 0x60 + iota
    PUSH2
    PUSH3
    PUSH4
    PUSH5
    PUSH6
    PUSH7
    PUSH8
    PUSH9
    PUSH10
    PUSH11
    PUSH12
    PUSH13
    PUSH14
    PUSH15
    PUSH16
    PUSH17
    PUSH18
    PUSH19
    PUSH20
    PUSH21
    PUSH22
    PUSH23
    PUSH24
    PUSH25
    PUSH26
    PUSH27
    PUSH28
    PUSH29
    PUSH30
    PUSH31
    PUSH32
)

// 0x80 range - dups.
const (
    DUP1 = 0x80 + iota
    DUP2
    DUP3
    DUP4
    DUP5
    DUP6
    DUP7
    DUP8
    DUP9
    DUP10
    DUP11
    DUP12
    DUP13
    DUP14
    DUP15
    DUP16
)

// 0x90 range - swaps.
const (
    SWAP1 = 0x90 + iota
    SWAP2
    SWAP3
    SWAP4
    SWAP5
    SWAP6
    SWAP7
    SWAP8
    SWAP9
    SWAP10
    SWAP11
    SWAP12
    SWAP13
    SWAP14
    SWAP15
    SWAP16
)

// 0xa0 range - logging ops.
const (
    LOG0 OpCode = 0xa0 + iota
    LOG1
    LOG2
    LOG3
    LOG4
)

// 0xf0 range - closures.
const (
    CREATE       OpCode = 0xf0
    CALL         OpCode = 0xf1
    CALLCODE     OpCode = 0xf2
    RETURN       OpCode = 0xf3
    DELEGATECALL OpCode = 0xf4
    CREATE2      OpCode = 0xf5

    STATICCALL   OpCode = 0xfa
    REVERT       OpCode = 0xfd
    INVALID      OpCode = 0xfe
    SELFDESTRUCT OpCode = 0xff
)
