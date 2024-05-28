# TracEVM

This tool is used to track the values and addresses of slots (storage variables in Solidity) inside the Ethereum contract, as well as tracking logs (Solidity events).

- Partial symbolic execution provides the complete data on how a specific variable or slot address was calculated
- Suitable for learning about Solidity internals
- Written in Go and Python

## Prerequesits

Project includes submodules, therefore it is needed to clone the project this way.

```bash
git clone --recurse-submodules https://github.com/ioterw/tracevm.git
```

Also, `make`, `go` and `python3` should be preinstalled.

Additionally, [Geth prerequisites](https://geth.ethereum.org/docs/getting-started/installing-geth#build-from-source) are required to compile Geth.

## Building

It is possible to build TracEVM with such command.

```bash
./build.py
```

## Running

To open webview on address [127.0.0.1:4334](http://127.0.0.1:4334) run the following command.

```bash
./run.py conf.json
```

TracEVM will be running!

![](images/webview.png)

## Connecting with Remix

It is expected to have Remix installed.

Choose External HTTP Provider

![](images/remix1.png)

Connect by default address.

![](images/remix2.png)

## Usage example

Let's imagine we have such contract.

![](images/sample1.png)

Let's deploy the contract and see the output.

![](images/sample2.png)

We see that:

- event type is final_slot, which means that this slot was written at the end of transaction (not reverted).
- further we see slot offsets (constant - initial slot, mapping - solidity keccak mapping magic, offset - offset from last value)
- short slot formula, which shows all cryptographic operations which were performed with slot

Also there is a full formula, which computes all needed data from initial initcode (or calldata).

![](images/sample3.png)

## Editing config

```Javascript
{
    "kv": {
        // Possible values for engine:
        // memory: data is stored in memory (root is ignored)
        // leveldb: data is stored in leveldb (root is path for leveldb folders)
        // riak: data is stored in riak (root is riak address)
        // amnesia: data is not stored (root is ignored, past_unknown is switched to true)
        "engine": "memory",
        "root": ""
    },
    "logger": {
        // _short postfix generally counts only cryptographic formulas (sha256, keccak etc.)
        // as significant, other formulas are folded 
        "final_slots_short": true,
        // outputs final slots which are set at the end of transaction
        "final_slots": true,
        "codes_short": false,
        // outputs code of contracts which is set at the end of transaction
        "codes": false,
        "return_data_short": false,
        // outputs return data at the end of transaction
        "return_data": true,
        "logs_short": false,
        // outputs logs (events)
        "logs": true,
        // outputs solidity view of final slots (final_slots should be enabled)
        "sol_view": true
    },
    // Possible values:
    // path to output file
    // if starts with "http://", starts http server on specified address
    // if empty, outputs to terminal
    "output": "http://127.0.0.1:4334",
    // special mode, if enabled TracEVM thinks that there are some slots or code which
    // existed before, therefore unknown, so it is marked as UNKNOWNSLOT or UNKNOWNCODE
    "past_unknown": false
}
```

## Found a bug?

Please open an [issue](https://github.com/ioterw/tracevm/issues) and supply solidity code that produced unexpected behaviour.
