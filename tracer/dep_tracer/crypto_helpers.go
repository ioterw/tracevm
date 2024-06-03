package dep_tracer

import (
    "hash"
    "encoding/binary"
    "golang.org/x/crypto/sha3"
)

func CreateAddress(caller Address, nonce uint64) Address {
    data := rlpEncodeCreate(caller, nonce)
    return bytesToAddress(Keccak256(data)[12:])
}

func rlpEncodeCreate(caller Address, nonce uint64) []byte {
    res := []byte{}
    res = append(res, rlpEncodeAddress(caller)...)
    res = append(res, rlpEncodeNonce(nonce)...)
    var prefix byte = 0xc0 + byte(len(res))
    return append([]byte{prefix}, res...)
}

func rlpEncodeAddress(addr Address) []byte {
    var prefix byte = 0x94
    return append([]byte{prefix}, addr[:]...)
}

func rlpEncodeNonce(nonce uint64) []byte {
    res := binary.BigEndian.AppendUint64(nil, nonce)
    for len(res) > 0 && res[0] == 0 {
        res = res[1:]
    }
    if len(res) == 1 && 0x00 <= res[0] && res[0] <= 0x7f {
        return res
    }
    var prefix byte = 0x80 + byte(len(res))
    return append([]byte{prefix}, res...)
}

func CreateAddress2(b Address, salt Hash, inithash []byte) Address {
    return bytesToAddress(Keccak256([]byte{0xff}, b[:], salt[:], inithash)[12:])
}

func bytesToAddress(b []byte) Address {
    setBytes := func(a Address, b []byte) {
        if len(b) > len(a) {
            b = b[len(b)-20:]
        }
        copy(a[20-len(b):], b)
    }

    var a Address
    setBytes(a, b)
    return a
}

func Keccak256(data ...[]byte) []byte {
    type KeccakState interface {
        hash.Hash
        Read([]byte) (int, error)
    }

    b := make([]byte, 32)
    d := sha3.NewLegacyKeccak256().(KeccakState)
    for _, b := range data {
        d.Write(b)
    }
    d.Read(b)
    return b
}
