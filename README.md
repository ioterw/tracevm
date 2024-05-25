# TracEVM

This tool is used to track the values and addresses of slots (storage variables in Solidity) inside the Ethereum contract, as well as tracking logs (Solidity events).

- Partial symbolic execution provides the complete data on how a specific variable or slot address was calculated
- Suitable for learning about Solidity internals
- Written in Go and Python

## Building

Python and Flask are required to run the web view.

```bash
pip3 install Flask
```

Additionally, [Geth prerequisites](https://geth.ethereum.org/docs/getting-started/installing-geth#build-from-source) are required to compile Geth.

Finally, it is possible to build TracEVM.

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
