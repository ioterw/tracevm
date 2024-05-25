package dep_tracer

type Memory struct {
    Data []DEPByte
}

func MemoryNew() *Memory {
    m := new(Memory)
    m.Data = make([]DEPByte, 0)
    return m
}

func (m *Memory) Copy() *Memory {
    res := new(Memory)
    res.Data = CopyDEPBytes(m.Data)
    return res
}

func (m *Memory) Set(offset uint64, value DEPByte) {
    m.Extend(int(offset) + 1)
    m.Data[offset] = value
}

func (m *Memory) Set32(offset uint64, value [32]DEPByte) {
    m.Extend(int(offset) + len(value))
    copy(m.Data[offset:offset+32], value[:])
}

func (m *Memory) SetN(offset uint64, value []DEPByte) {
    if len(value) < 1 {
        return
    }
    m.Extend(int(offset) + len(value))
    copy(m.Data[offset:offset+uint64(len(value))], value)
}

func (m *Memory) Load(offset, size uint64) []DEPByte {
    if size < 1 {
        return []DEPByte{}
    }
    m.Extend(int(offset) + int(size))
    return CopyDEPBytes(m.Data[offset:offset+size])
}

func (m *Memory) Extend(size int) {
    if size % 32 != 0 {
        size = (size / 32 + 1) * 32
    }
    if size <= len(m.Data) {
        return
    }
    l := size - len(m.Data)
    b := DEPByte{0, ConstantInitZero.hash}
    for i := 0; i < l; i++ {
        m.Data = append(m.Data, b)
    }
}
