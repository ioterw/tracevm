package dep_tracer

var (
    ConstantInitZero = ConstantNew(OPInitZero, []byte{0})
    ConstantZero     = ConstantNew(OPConstant, []byte{0})
)
