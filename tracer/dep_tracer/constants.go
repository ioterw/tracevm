package dep_tracer

type Hash    [32]byte
type Address [20]byte

var (
    ConstantInitZero = ConstantNew(OPInitZero, []byte{0})
    ConstantZero     = ConstantNew(OPConstant, []byte{0})
)

