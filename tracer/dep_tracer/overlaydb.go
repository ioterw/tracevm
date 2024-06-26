package dep_tracer

import (
    "fmt"
    "encoding/hex"
    "github.com/holiman/uint256"
)

type OverlayDBSlotKey struct {
    addr Address
    slot uint256.Int
}

type OverlaySlot struct {
    data     []DEPByte
    codeAddr Address
}

type OverlayCode struct {
    data         []DEPByte
    codeAddr     Address
    codeHash     Hash
    initcodeHash Hash
}

type OverlayDB struct {
    simpleDB      *SimpleDB
    slots         map[OverlayDBSlotKey]OverlaySlot
    updatedSlots  map[OverlayDBSlotKey]bool
    codes         map[Address]OverlayCode
    updatedCodes  map[Address]bool
    selfdestruced map[Address]bool
    created       map[Address]bool
    versions      map[Address]uint64
    transient     map[OverlayDBSlotKey][]DEPByte
}

func OverlayDBNew(simpleDB *SimpleDB) *OverlayDB {
    o := new(OverlayDB)
    o.simpleDB = simpleDB
    o.slots = make(map[OverlayDBSlotKey]OverlaySlot)
    o.updatedSlots = make(map[OverlayDBSlotKey]bool)
    o.codes = make(map[Address]OverlayCode)
    o.updatedCodes = make(map[Address]bool)
    o.selfdestruced = make(map[Address]bool)
    o.created = make(map[Address]bool)
    o.versions = make(map[Address]uint64)
    return o
}

func (o *OverlayDB) Copy() *OverlayDB {
    res := new(OverlayDB)
    res.simpleDB = o.simpleDB
    res.slots = make(map[OverlayDBSlotKey]OverlaySlot)
    for k,v := range o.slots {
        res.slots[k] = v
    }
    res.updatedSlots = make(map[OverlayDBSlotKey]bool)
    for k,_ := range o.updatedSlots {
        res.updatedSlots[k] = true
    }
    res.codes = make(map[Address]OverlayCode)
    for k,v := range o.codes {
        res.codes[k] = v
    }
    res.updatedCodes = make(map[Address]bool)
    for k,_ := range o.updatedCodes {
        res.updatedCodes[k] = true
    }
    res.selfdestruced = make(map[Address]bool)
    for k,v := range o.selfdestruced {
        res.selfdestruced[k] = v
    }
    res.created = make(map[Address]bool)
    for k,v := range o.created {
        res.created[k] = v
    }
    res.versions = make(map[Address]uint64)
    for k,v := range o.versions {
        res.versions[k] = v
    }
    res.transient = make(map[OverlayDBSlotKey][]DEPByte)
    for k,v := range o.transient {
        res.transient[k] = v
    }
    return res
}

func (o *OverlayDB) GetAddressVersion(addr Address) uint64 {
    val, ok := o.versions[addr]
    if ok {
        return val
    }
    val = o.simpleDB.GetAddressVersion(addr)
    o.versions[addr] = val
    return val
}

func (o *OverlayDB) GetSlot(addr Address, slot *uint256.Int, value Hash) OverlaySlot {
    key := OverlayDBSlotKey{addr, *slot}
    val, ok := o.slots[key]
    if ok {
        return val
    }
    val = OverlaySlot{o.simpleDB.GetSlot(addr, slot, value), Address{}}
    o.slots[key] = val
    return val
}

func (o *OverlayDB) SetSlot(addr, codeAddress Address, slot *uint256.Int, val []DEPByte) {
    key := OverlayDBSlotKey{addr, *slot}
    o.slots[key] = OverlaySlot{val, codeAddress}
    o.updatedSlots[key] = true
}

func (o *OverlayDB) GetTransient(addr Address, slot *uint256.Int) []DEPByte {
    key := OverlayDBSlotKey{addr, *slot}
    val, ok := o.transient[key]
    if ok {
        return val
    }
    val = InitDEPBytes(32)
    o.transient[key] = val
    return val
}

func (o *OverlayDB) SetTransient(addr Address, slot *uint256.Int, val []DEPByte) {
    key := OverlayDBSlotKey{addr, *slot}
    o.transient[key] = val
}

func (o *OverlayDB) GetCode(addr Address, code []byte) OverlayCode {
    val, ok := o.codes[addr]
    if ok {
        return val
    }
    codeHash, initcodeHash, res := o.simpleDB.GetCode(addr, code)
    val = OverlayCode{res, Address{}, codeHash, initcodeHash}
    o.codes[addr] = val
    return val
}

func (o *OverlayDB) SetCode(addr, codeAddress Address, val []DEPByte, valBytes []byte, initcodeHash Hash) {
    o.codes[addr] = OverlayCode{val, codeAddress, CodeHash(valBytes), initcodeHash}
    o.updatedCodes[addr] = true
    o.created[addr] = true
}

func (o *OverlayDB) Destruct(addr Address) {
    o.selfdestruced[addr] = true
}

func (o *OverlayDB) Created(addr Address) bool {
    _, ok := o.created[addr]
    return ok
}

func (o *OverlayDB) Commit() {
    for k,_ := range o.updatedSlots {
        value := o.slots[k]
        o.simpleDB.CommitDEPBytesWithShorts(value.data)
        o.simpleDB.SetSlot(k.addr, &k.slot, value.data)
        o.simpleDB.logger.LogFinalSlot(k.addr, o.GetAddressVersion(k.addr), value.codeAddr, value.data, &k.slot)
    }
    for addr, _ := range o.updatedCodes {
        code := o.codes[addr]
        o.simpleDB.CommitDEPBytesWithShorts(code.data)
        o.simpleDB.SetCode(addr, code.data, code.codeHash, code.initcodeHash)
        o.simpleDB.logger.LogFinalCode(addr, o.GetAddressVersion(addr), code.codeAddr, code.data)
    }
    for addr, _ := range o.selfdestruced {
        o.simpleDB.IncreaseAddressVersion(addr)
    }
}

func (o *OverlayDB) Print(f Formula) {
    o.simpleDB.Print(f)
}

func (o *OverlayDB) FullPrint(f Formula) {
    o.simpleDB.FullPrint(f)
}

func (o *OverlayDB) PrintData(data []DEPByte) {
    o.simpleDB.PrintData(data)
}

func (o *OverlayDB) FullPrintData(data []DEPByte) {
    o.simpleDB.FullPrintData(data)
}

func (o *OverlayDB) PrintCommit() {
    fmt.Println("-- SLOTS --")
    for k, _ := range o.updatedSlots {
        value := o.slots[k]
        slotBytes := k.slot.Bytes32()
        fmt.Println("[", hex.EncodeToString(k.addr[:]), "->", hex.EncodeToString(slotBytes[:]), "]")
        o.FullPrintData(value.data)
    }
    fmt.Println("-- CODES --")
    for addr, _ := range o.updatedCodes {
        code := o.codes[addr]
        fmt.Println("[", hex.EncodeToString(addr[:]), "]")
        o.FullPrintData(code.data)
    }
    fmt.Println("-- SELFDESTRUCTS --")
    for addr, _ := range o.selfdestruced {
        fmt.Println("[", hex.EncodeToString(addr[:]), "]")
    }
}
