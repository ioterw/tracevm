package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct {
    uint8_t data[20];
} Address;

typedef struct {
    uint8_t data[32];
} Hash;

typedef struct {
    uint8_t *data;
    int      size;
} SizedArray;

typedef struct {
    Hash *data;
    int   size;
} Stack;

typedef uint64_t (*get_nonce_function) (Address address);
inline uint64_t get_nonce_bridge(get_nonce_function f, Address address) {
    return f(address);
}

typedef SizedArray (*get_code_function) (Address address);
inline SizedArray get_code_bridge(get_code_function f, Address address) {
    return f(address);
}

typedef uint64_t (*log_data_function) (char *data);
inline void log_data_bridge(log_data_function f, char *data) {
    f(data);
}
*/
import "C"

import (
    "unsafe"
    "math/big"
    "github.com/holiman/uint256"

    "dep_tracer/dep_tracer"
)

func main() {}

var (
    getNoncePointer       C.get_nonce_function        = nil
    getCodePointer        C.get_code_function         = nil
    cDepHandler           *dep_tracer.DepHandler      = nil
    cTracing              bool                        = false
)


//export RegisterGetNonce
func RegisterGetNonce(pointer C.get_nonce_function) {
   getNoncePointer = pointer
}
//export RegisterGetCode
func RegisterGetCode(pointer C.get_code_function) {
   getCodePointer = pointer
}

func packAddress(addr dep_tracer.Address) C.Address {
    res := C.Address{}
    for i := 0; i < 20; i ++ {
        res.data[i] = C.uchar(addr[i])
    }
    return res
}
func unpackAddress(addr C.Address) dep_tracer.Address {
    res := dep_tracer.Address {}
    for i := 0; i < 20; i ++ {
        res[i] = byte(addr.data[i])
    }
    return res
}
func unpackHash(hash C.Hash) dep_tracer.Hash {
    res := dep_tracer.Hash {}
    for i := 0; i < 32; i ++ {
        res[i] = byte(hash.data[i])
    }
    return res
}
func unpackSizedArray(data C.SizedArray) []byte {
    return C.GoBytes(unsafe.Pointer(data.data), data.size)
}
func unpackStack(data C.Stack) []uint256.Int {
    res := []uint256.Int{}
    for i := 0; i < int(data.size); i++ {
        a := (*C.Hash)(unsafe.Pointer(uintptr(unsafe.Pointer(data.data)) + uintptr(i) * unsafe.Sizeof(C.Hash{})))
        hashNum := unpackHash(*a)
        num := uint256.Int{}
        num.SetBytes32(hashNum[:])
        res = append(res, num)
    }
    return res
}

type StateDBC struct {}
func (s StateDBC) GetNonce(addr [20]byte) uint64 {
    res := C.get_nonce_bridge(getNoncePointer, packAddress(addr))
    return uint64(res)
}
func (s StateDBC) GetCode(addr [20]byte) []byte {
    res := C.get_code_bridge(getCodePointer, packAddress(addr))
    return unpackSizedArray(res)
}

type CallbackWriterCB struct {
    pointer C.log_data_function
}
func (cw CallbackWriterCB) Write(data []byte) {
    cdata := C.CString(string(data))
    C.log_data_bridge(cw.pointer, cdata)
    C.free(unsafe.Pointer(cdata))
}

//export InitDep
func InitDep(cfg *C.char, pointer C.log_data_function) {
    if cDepHandler != nil {
        panic("InitDep called twice")
    }
    if pointer == nil {
        cDepHandler = dep_tracer.NewDepHandler([]byte(C.GoString(cfg)), nil)
    } else {
        cDepHandler = dep_tracer.NewDepHandler([]byte(C.GoString(cfg)), CallbackWriterCB{pointer})        
    }
}

//export StartTransactionRecording
func StartTransactionRecording(
    isCreate bool, addr C.Address, input C.SizedArray, block uint64,
    timestamp uint64, origin C.Address, txHash C.Hash,
    code C.SizedArray, isSelfdestruct6780, isRandom bool,
) {
    cTracing = true
    block0 := new(big.Int)
    block0.SetUint64(block)
    cDepHandler.StartTransactionRecording(
        isCreate, unpackAddress(addr), unpackSizedArray(input), block0,
        timestamp, unpackAddress(origin), unpackHash(txHash),
        unpackSizedArray(code), isSelfdestruct6780, isRandom, StateDBC{},
    )
}

//export EndTransactionRecording
func EndTransactionRecording() {
    if !cTracing {
        return
    }
    cTracing = false
    cDepHandler.EndTransactionRecording()
}

//export HandleOpcode
func HandleOpcode(
    stack C.Stack, memory C.SizedArray, addr C.Address,
    pc uint64, op byte, isInvalid bool, hasError bool,
) {
    if !cTracing {
        return
    }
    cDepHandler.HandleOpcode(
        unpackStack(stack), unpackSizedArray(memory), unpackAddress(addr),
        pc, op, isInvalid, hasError,
    )
}

//export HandleEnter
func HandleEnter(to C.Address, input C.SizedArray) {
    if !cTracing {
        return
    }
    cDepHandler.HandleEnter(
        unpackAddress(to), unpackSizedArray(input),
    )
}

//export HandleFault
func HandleFault(op byte) {
    if !cTracing {
        return
    }
    cDepHandler.HandleFault(op)
}

//export HandleExit
func HandleExit(output C.SizedArray, hasError bool) {
    if !cTracing {
        return
    }
    cDepHandler.HandleExit(unpackSizedArray(output), hasError)
}
