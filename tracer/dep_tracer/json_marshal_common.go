package dep_tracer

import (
    "fmt"
    "math/big"
    "github.com/holiman/uint256"
)


type BigInt struct {
    value big.Int
}

func (b BigInt) MarshalJSON() ([]byte, error) {
    return []byte(b.value.String()), nil
}

func (b *BigInt) UnmarshalJSON(p []byte) error {
    var z big.Int
    _, ok := z.SetString(string(p), 10)
    if !ok {
        return fmt.Errorf("not a valid big integer: %s", p)
    }
    b.value = z
    return nil
}

type DataStart struct {
    IsCreate  bool           `json:"is_create"`
    Address   Address `json:"address"`
    Input     []byte         `json:"input"`
    Block     *big.Int       `json:"block"`
    Timestamp uint64         `json:"timestamp"`
    Origin    Address `json:"origin"`
    TxHash    Hash    `json:"tx_hash"`
    Code      []byte         `json:"code"`
}

type DataError struct {
    Reverted  bool `json:"reverted"`
}

type DataPush struct {
    Pc   uint64 `json:"pc"`
    Size uint64 `json:"size"`
}

type DataDup struct {
    Size int `json:"size"`
}

type DataSwap struct {
    Size int64  `json:"size"`
}

type DataPop struct {}

type DataMLoad struct {
    Offset uint64 `json:"offset"`
}

type DataMStore struct {
    Offset uint64 `json:"offset"`
}

type DataMStore8 struct {
    Offset uint64 `json:"offset"`
}

type DataMCopy struct {
    ToOffset   uint64 `json:"to_offset"`
    FromOffset uint64 `json:"from_offset"`
    Size       uint64 `json:"size"`
}

type DataConstant struct {
    Op    uint8   `json:"op"`
    Value uint256.Int `json:"value"`
}

type DataConstant20 struct {
    Op    uint8   `json:"op"`
    Value uint256.Int `json:"value"`
}

type DataSLoad struct {
    Slot  uint256.Int `json:"slot"`
    Value uint256.Int `json:"value"`
}

type DataSStore struct {
    Slot  uint256.Int `json:"slot"`
    Value uint256.Int `json:"value"`
}

type DataTLoad struct {
    Slot  uint256.Int `json:"slot"`
}

type DataTStore struct {
    Slot  uint256.Int `json:"slot"`
}

type DataOne struct {
    Op    uint8   `json:"op"`
    Value uint256.Int `json:"value"`
}

type DataTwo struct {
    Op    uint8   `json:"op"`
    Value uint256.Int `json:"value"`
}

type DataThree struct {
    Op    uint8   `json:"op"`
    Value uint256.Int `json:"value"`
}

type DataByte struct {
    Offset uint256.Int `json:"offset"`
}

type DataKeccak struct {
    Result Hash `json:"result"`
    Offset uint64     `json:"offset"`
    Size   uint64     `json:"size"`
}

type DataCodeSize struct {
    CodeSize uint64 `json:"code_size"`
}

type DataExtCodeSize struct {
    Address  Address `json:"address"`
    CodeSize uint256.Int    `json:"code_size"`
    Code     []byte         `json:"code"`
}

type DataExtCodeHash struct {
    Address  Address `json:"address"`
    Hash     Hash    `json:"hash"`
    Code     []byte         `json:"code"`
}

type DataCalldataSize struct {
    CalldataSize uint64 `json:"calldata_size"`
}

type DataReturndataSize struct {
    ReturndataSize uint64 `json:"returndata_size"`
}

type DataCodeCopy struct {
    MemoryOffset uint64 `json:"memory_offset"`
    CodeOffset   uint64 `json:"code_offset"`
    Length       uint64 `json:"length"`
}

type DataExtCodeCopy struct {
    Address      Address `json:"address"`
    MemoryOffset uint64         `json:"memory_offset"`
    CodeOffset   uint64         `json:"code_offset"`
    Length       uint64         `json:"length"`
    Code         []byte         `json:"code"`
}

type DataCalldataCopy struct {
    MemoryOffset uint64 `json:"memory_offset"`
    DataOffset   uint64 `json:"data_offset"`
    Size         uint64 `json:"size"`
}

type DataReturndataCopy struct {
    MemoryOffset uint64 `json:"memory_offset"`
    DataOffset   uint64 `json:"data_offset"`
    Size         uint64 `json:"size"`
}

type DataCalldataLoad struct {
    Offset uint64 `json:"offset"`
}

type DataLog struct {
    Offset    uint64 `json:"offset"`
    Size      uint64 `json:"size"`
    TopicsNum int    `json:"size"`
}

type DataReturn struct {
    Offset uint64   `json:"offset"`
    Size   uint64   `json:"size"`
    Result []byte `json:"result"`
}

type DataStop struct {}

type DataSelfdestruct struct {}

type DataSelfdestruct6780 struct {}

type DataRevert struct {
    Offset uint64 `json:"offset"`
    Size   uint64 `json:"size"`
}

type DataEmpty struct {
    N int `json:"n"`
}

type DataBalance struct {
    Balance uint256.Int `json:"balance"`
}

type DataSelfBalance struct {
    Balance uint256.Int `json:"balance"`
}

type DataBlockHash struct {
    Hash Hash `json:"hash"`
}

type DataBlobHash struct {
    Hash Hash `json:"hash"`
}

type DataCreateStart struct {
    Address Address `json:"address"`
    Offset  uint64     `json:"offset"`
    Size    uint64     `json:"size"`
    Data    []byte   `json:"data"`
}

type DataCreateEnd struct {
    Address Address `json:"address"`
}

type DataCreate2Start struct {
    Address Address `json:"address"`
    Offset  uint64     `json:"offset"`
    Size    uint64     `json:"size"`
    Data    []byte   `json:"data"`
}

type DataCreate2End struct {
    Address Address `json:"address"`
}

type DataCallStart struct {
    N           int            `json:"n"`
    Address     Address `json:"address"`
    CodeAddress Address `json:"code_address"`
    InOffset    uint64         `json:"in_offset"`
    InSize      uint64         `json:"in_size"`
    Code        []byte         `json:"code"`
}

type DataCallEnd struct {
    Success      bool   `json:"success"`
    ReturnOffset uint64 `json:"return_offset"`
    ReturnSize   uint64 `json:"return_size"`
}

type DataPrecompileEcRecover struct {
    Result []byte `json:"result"`
}

type DataPrecompileSha256 struct {
    Result []byte `json:"result"`
}

type DataPrecompileRipemd160 struct {
    Result []byte `json:"result"`
}

type DataPrecompileIdentity struct {
    Result []byte `json:"result"`
}

type DataPrecompileModExp struct {
    Result []byte `json:"result"`
    BSize  uint64   `json:"bsize"`
    ESize  uint64   `json:"esize"`
    MSize  uint64   `json:"msize"`
}

type DataPrecompileEcAdd struct {
    Result []byte `json:"result"`
}

type DataPrecompileEcMul struct {
    Result []byte `json:"result"`
}

type DataPrecompileEcPairing struct {
    Result []byte `json:"result"`
}

type DataPrecompileBlake2F struct {
    Result []byte `json:"result"`
}

type DataPointEvaluation struct {
    Result []byte `json:"result"`
}
