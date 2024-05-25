package dep_tracer

type Stack struct {
    Data [][32]DEPByte
}

func StackNew() *Stack {
    st := new(Stack)
    st.Data = make([][32]DEPByte, 0)
    return st
}

func (st *Stack) Copy() *Stack {
    res := new(Stack)
    res.Data = make([][32]DEPByte, len(st.Data))
    for i, b := range st.Data {
        res.Data[i] = b
    }
    return res
}

func (st *Stack) Push(val [32]DEPByte) {
    st.Data = append(st.Data, val)
}

func (st *Stack) PushN(val []DEPByte) {
    if len(val) > 32 {
        panic("can't push more than 32 bytes")
    }
    extraBytes := 32 - len(val)
    res := [32]DEPByte{}
    b := DEPByte{0, ConstantInitZero.hash}
    for i := 0; i < extraBytes; i++ {
        res[i] = b
    }
    copy(res[extraBytes:32], val)
    st.Data = append(st.Data, res)
}

func (st *Stack) Pop() [32]DEPByte {
    res := st.Data[len(st.Data)-1]
    st.Data = st.Data[:len(st.Data)-1]
    return res
}

func (st *Stack) Swap(n int) {
    st.Data[len(st.Data)-n], st.Data[len(st.Data)-1] = st.Data[len(st.Data)-1], st.Data[len(st.Data)-n]
}

func (st *Stack) Dup(n int) {
    st.Push(st.Data[len(st.Data)-n])
}

func (st *Stack) Size() int {
    return len(st.Data)
}
