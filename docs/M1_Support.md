# Running Provenance on an M1 laptop
---

### Summary

While Provenance cannot currently be compiled on an M1 Apple laptop, a chain can be run locally.  Setting up a Rosetta terminal is **not** needed to run a local chain if a release binary is downloaded.  However, at present only a limited subset of functionality on the Provenance blockchain has been tested on an M1 laptop.  What has and has not been tested is documented below along with instructions on downloading and running a local chain.

### Setting up
Download the `darwin-amd64` zipped binary from the Provenance release page: [link](https://github.com/provenance-io/provenance/releases/tag/v1.7.6)

Unzip the binary and move `provenanced` and `libwasmvm.dylib` into a directory in your path or update the `PROV_CMD` in the script below to point to the folder they are located in.


### Running

The following script is modified from the run command in Provenance to configure and start a local chain using the downloaded binary:

```bash
#!/bin/bash -ex

# If the provenanced executable is not in your $PATH, update PROV_CMD to include the full path to it.
PROV_CMD="provenanced"
PIO_HOME="${PIO_HOME:-$HOME/Library/Application Support/Provenance}"
export PIO_HOME

if [ ! -d "$PIO_HOME/config" ]; then
    "$PROV_CMD" -t init --chain-id=testing testing
    "$PROV_CMD" -t keys add validator --keyring-backend test
    "$PROV_CMD" -t add-genesis-root-name validator pio --keyring-backend test
    "$PROV_CMD" -t add-genesis-root-name validator pb --restrict=false \
		--keyring-backend test
    "$PROV_CMD" -t add-genesis-root-name validator io --restrict \
		--keyring-backend test
    "$PROV_CMD" -t add-genesis-root-name validator provenance \ 
		--keyring-backend test
    "$PROV_CMD" -t add-genesis-account validator 100000000000000000000nhash \
		--keyring-backend test
    "$PROV_CMD" -t gentx validator 1000000000000000nhash \ 
		--keyring-backend test --chain-id=testing
    "$PROV_CMD" -t add-genesis-marker 100000000000000000000nhash --manager \
		validator --access mint,burn,admin,withdraw,deposit \ 
		--activate --keyring-backend test
    "$PROV_CMD" -t collect-gentxs
fi
"$PROV_CMD" -t start
```

Running this script will start up a local chain with a default home directory. 

### Testing

The following functionality has been verified to work:

1. Starting a local chain and running it to 50000+ blocks
2. Querying the chain
3. Adding new keys to the chain
4. Sending nhash between various keys
5. Creating a name attribute
6. Creating a new coin
7. Granting permission, activating, finalizing and withdrawing the new coin
8. Storing, instantiating, querying and executing a simple smart contract

However, in order to generally test that the chain is working properly a mainnet node should be started up and run for several days without failure.