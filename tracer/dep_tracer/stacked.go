package dep_tracer

import(
    "github.com/ethereum/go-ethereum/common"
)

type StackedElement struct {
    isCreate     bool
    addr         common.Address
    addrVersion  uint64
    codeAddr     common.Address
    calldata     []DEPByte
    code         []DEPByte
    codeHash     common.Hash
    initcodeHash common.Hash
    stack        *Stack
    memory       *Memory
}

func StackedElementNew(isCreate bool, addr common.Address, addrVersion uint64, codeAddr common.Address, calldata []DEPByte, code []DEPByte, codeHash, initcodeHash common.Hash) *StackedElement {
    se := new(StackedElement)
    se.isCreate = isCreate
    se.addr = addr
    se.addrVersion = addrVersion
    se.codeAddr = codeAddr
    se.calldata = CopyDEPBytes(calldata)
    se.code = CopyDEPBytes(code)
    se.codeHash = codeHash
    se.initcodeHash = initcodeHash
    se.stack = StackNew()
    se.memory = MemoryNew()
    return se
}

func (se *StackedElement) Copy() *StackedElement {
    res := new(StackedElement)
    res.isCreate = se.isCreate
    res.addr = se.addr
    res.addrVersion = se.addrVersion
    res.codeAddr = se.codeAddr
    res.calldata = CopyDEPBytes(se.calldata)
    res.code = CopyDEPBytes(se.code)
    res.codeHash = se.codeHash
    res.stack = se.stack.Copy()
    res.memory = se.memory.Copy()
    return res
}


type Stacked struct {
    elements []*StackedElement
}

func StackedNew(isCreate bool, addr common.Address, addrVersion uint64, codeAddr common.Address, calldata []DEPByte, code []DEPByte, codeHash, initcodeHash common.Hash) *Stacked {
    s := new(Stacked)
    s.elements = make([]*StackedElement, 0)
    s.Push(isCreate, addr, addrVersion, codeAddr, calldata, code, codeHash, initcodeHash)
    return s
}

func (s *Stacked) Copy() *Stacked {
    res := new(Stacked)
    res.elements = make([]*StackedElement, len(s.elements))
    for i, element := range s.elements {
        res.elements[i] = element.Copy()
    }
    return res
}

func (s *Stacked) Push(isCreate bool, addr common.Address, addrVersion uint64, codeAddr common.Address, calldata []DEPByte, code []DEPByte, codeHash, initcodeHash common.Hash) {
    se := StackedElementNew(isCreate, addr, addrVersion, codeAddr, calldata, code, codeHash, initcodeHash)
    s.elements = append(s.elements, se)
}

func (s *Stacked) Cur() *StackedElement {
    if len(s.elements) < 1 {
        panic("no elements in stacked")
    }
    return s.elements[len(s.elements)-1]
}

func (s *Stacked) Pop() {
    s.elements = s.elements[:len(s.elements)-1]
}
