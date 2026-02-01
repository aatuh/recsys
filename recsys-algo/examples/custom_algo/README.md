# Custom Algorithm Plugin (Dev Only)

This example builds a Go plugin that implements the recsys-algo contract.

Build:

```bash
go build -buildmode=plugin -o ./hello_algo.so ./examples/custom_algo
```

Run the API with:

```bash
RECSYS_ALGO_PLUGIN_ENABLED=true \
RECSYS_ALGO_PLUGIN_PATH=./recsys-algo/examples/custom_algo/hello_algo.so \
make dev
```

The plugin exports `RecsysAlgorithmPlugin` and implements a minimal
popularity-only algorithm to demonstrate the contract.
