package dep_tracer

type StackedElement struct {
    isCreate     bool
    addr         Address
    addrVersion  uint64
    codeAddr     Address
    calldata     []DEPByte
    code         []DEPByte
    codeHash     Hash
    initcodeHash Hash
    stack        *Stack
    memory       *Memory
}

func StackedElementNew(isCreate bool, addr Address, addrVersion uint64, codeAddr Address, calldata []DEPByte, code []DEPByte, codeHash, initcodeHash Hash) *StackedElement {
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

func StackedNew(isCreate bool, addr Address, addrVersion uint64, codeAddr Address, calldata []DEPByte, code []DEPByte, codeHash, initcodeHash Hash) *Stacked {
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

func (s *Stacked) Push(isCreate bool, addr Address, addrVersion uint64, codeAddr Address, calldata []DEPByte, code []DEPByte, codeHash, initcodeHash Hash) {
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
