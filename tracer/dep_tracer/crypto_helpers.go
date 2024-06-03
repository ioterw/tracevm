package dep_tracer

import (
    "hash"
    "golang.org/x/crypto/sha3"
    
    "github.com/ethereum/go-ethereum/rlp"
)

func CreateAddress(b Address, nonce uint64) Address {
    data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})
    return bytesToAddress(Keccak256(data)[12:])
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
