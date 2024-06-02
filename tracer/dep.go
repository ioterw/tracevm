package live

import (
    "github.com/ethereum/go-ethereum/eth/tracers"

    "github.com/ethereum/go-ethereum/eth/tracers/live/dep_tracer"
)

func init() {
    tracers.LiveDirectory.Register("dep", dep_tracer.NewDep)
}
