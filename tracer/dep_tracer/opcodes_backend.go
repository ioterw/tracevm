package dep_tracer

const (
    // Constants
    OPInitZero    uint8 = 0x00
    OPInitCode    uint8 = 0x01
    OPCallData    uint8 = 0x02
    OPConstant    uint8 = 0x03
    OPCoinbase    uint8 = 0x04
    OPAddress     uint8 = 0x05
    OPOrigin      uint8 = 0x06
    OPCaller      uint8 = 0x07
    OPCallValue   uint8 = 0x08
    OPGasPrice    uint8 = 0x09
    OPTimestamp   uint8 = 0x0A
    OPNumber      uint8 = 0x0B
    OPDifficulty  uint8 = 0x0C
    OPRandom      uint8 = 0x0D
    OPGasLimit    uint8 = 0x0E
    OPPc          uint8 = 0x0F
    OPMsize       uint8 = 0x10
    OPGas         uint8 = 0x11
    OPChainID     uint8 = 0x12
    OPBaseFee     uint8 = 0x13
    OPCreateAddr  uint8 = 0x14
    OPCreate2Addr uint8 = 0x15
    OPCallResult  uint8 = 0x16
    OPBlobBaseFee uint8 = 0x17

    // Dynamic
    OPSlice           uint8 = 0xA0
    // take some uint8s from bigger value (value + offset + size)
    // do not want to split every 32 byte result to separate bytes without any reason
    OPConcat          uint8 = 0xA1 // opposite of slice, for example can be used when passing multiple
    OPSize            uint8 = 0xA2 // size of data
    OPCodeSize        uint8 = 0xA3
    OPAdd             uint8 = 0xA4
    OPMul             uint8 = 0xA5
    OPSub             uint8 = 0xA6
    OPDiv             uint8 = 0xA7
    OPSDiv            uint8 = 0xA8
    OPMod             uint8 = 0xA9
    OPSMod            uint8 = 0xAA
    OPExp             uint8 = 0xAB
    OPSignExtend      uint8 = 0xAC
    OPNot             uint8 = 0xAD
    OPLt              uint8 = 0xAE
    OPGt              uint8 = 0xAF
    OPSlt             uint8 = 0xB0
    OPSgt             uint8 = 0xB1
    OPEq              uint8 = 0xB2
    OPOr              uint8 = 0xB3
    OPXor             uint8 = 0xB4
    OPAddMod          uint8 = 0xB5
    OPMulMod          uint8 = 0xB6
    OPShl             uint8 = 0xB7
    OPShr             uint8 = 0xB8
    OPSar             uint8 = 0xB9
    OPAnd             uint8 = 0xBA
    OPIsZero          uint8 = 0xBB
    OPKeccak          uint8 = 0xBC
    OPCodeKeccak      uint8 = 0xBD
    OPBalance         uint8 = 0xBE
    OPBlockHash       uint8 = 0xBF
    OPEcRecover       uint8 = 0xC0
    OPSha256          uint8 = 0xC1
    OPRipemd160       uint8 = 0xC2
    OPModExp          uint8 = 0xC3
    OPEcAddX          uint8 = 0xC4
    OPEcAddY          uint8 = 0xC5
    OPEcMulX          uint8 = 0xC6
    OPEcMulY          uint8 = 0xC7
    OPEcPairing       uint8 = 0xC8
    OPBlake2F         uint8 = 0xC9
    OPBlobHash        uint8 = 0xD0
    OPPointEvaluation uint8 = 0xD1

    // Addressable (1st - value, 2nd - address) / when used as operand - shortened to value
    OPSLoad  uint8 = 0xE0
    OPSStore uint8 = 0xE1
)
