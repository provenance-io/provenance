# State Sync

The Provenance 1.10.0+ release comes with Tendermint Core 0.34 which includes support for state sync. State sync allows a new node to join a network by fetching a snapshot of the application state at a recent height vs. fetching and replaying all historical blocks (which can take days). This can reduce the time needed to sync with the network down to minutes.

<!-- TOC -->
  - [Prerequisites](#prerequisites)
    - [provenanced binary](#provenanced-binary)
    - [data directory](#data-directory)
  - [Starting a new node from State Sync](#starting-a-new-node-from-state-sync)
    - [Mainnet](#mainnet)
    - [Testnet](#testnet)
    - [Power users](#power-users)


## Prerequisites

### provenanced binary

The instructions in this document require the `provenanced` binary to be installed on your system. 
See [docs/Building.md](https://github.com/provenance-io/provenance/blob/main/docs/Building.md) for installation instructions.

### data directory

If `$PIO_HOME/data/` contains prior data, you must first clean it up in order for `state sync` to work. Delete everything BUT `priv_validator_state.json`.

## Starting a new node from State Sync

### Mainnet

```  
# Initialization is required ONLY on first time setups. Otherwise, skip this step.
export PIO_HOME=~/.provenanced
# Change "choose-a-moniker" below to your own moniker, e.g. pio-node-1
provenanced init choose-a-moniker --chain-id pio-mainnet-1
curl https://raw.githubusercontent.com/provenance-io/mainnet/main/pio-mainnet-1/genesis.json > "$PIO_HOME/config/genesis.json"

# backup config
cp $PIO_HOME/config/config.toml $PIO_HOME/config/config.toml.orig

# update config to use 'cleveldb'
provenanced config set db_backend "cleveldb"

# setup sync node
PIO_RPC="$( host rpc-$(( $RANDOM % 3 )).provenance.io | awk '{print $4}' ):26657"

# State Sync Configuration Options
LATEST_HEIGHT="$(curl -s "$PIO_RPC/block" | jq -r .result.block.header.height)"
TRUST_HEIGHT="$((LATEST_HEIGHT - 1000))"
TRUST_HASH="$(curl -s "$PIO_RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)"
echo -e "PIO_RPC=$PIO_RPC\nLATEST_HEIGH=$LATEST_HEIGHT\nTRUST_HEIGHT=$TRUST_HEIGHT\nTRUST_HASH=$TRUST_HASH\n"

# Enable state sync
provenanced config set \
    statesync.enable true \
    statesync.rpc_servers "[\"$PIO_RPC\",\"$PIO_RPC\"]" \
    statesync.trust_height "$TRUST_HEIGHT" \
    statesync.trust_hash "$TRUST_HASH" \
    p2p.seeds '4bd2fb0ae5a123f1db325960836004f980ee09b4@seed-0.provenance.io:26656, 048b991204d7aac7209229cbe457f622eed96e5d@seed-1.provenance.io:26656'

# start node
provenanced start --x-crisis-skip-assert-invariants

# start node (capture stdout & stderr to log file)
provenanced start --x-crisis-skip-assert-invariants --log_level=info &>> pio.log
```

### Testnet

```
# Initialization is required ONLY on first time setups. Otherwise, skip this step.
export PIO_HOME=~/.provenanced
# Change "choose-a-moniker" below to your own moniker, e.g. pio-node-1
provenanced init --testnet choose-a-moniker --chain-id pio-testnet-1
curl https://raw.githubusercontent.com/provenance-io/testnet/main/pio-testnet-1/genesis.json > "$PIO_HOME/config/genesis.json"

# backup config
cp $PIO_HOME/config/config.toml $PIO_HOME/config/config.toml.orig

# update config to use 'cleveldb'
provenanced config set db_backend "cleveldb"

# setup sync node
# PIO_RPC="$( host rpc.test.provenance.io | awk '{print $4}' ):26657"
# (Temporary workaround due to how the tesntet hosts are currently configured)
PIO_RPC=34.66.209.228:26657

# State Sync Configuration Options
LATEST_HEIGHT="$(curl -s "$PIO_RPC/block" | jq -r .result.block.header.height)"
TRUST_HEIGHT="$((LATEST_HEIGHT - 1000))"
TRUST_HASH="$(curl -s "$PIO_RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)"
echo -e "PIO_RPC=$PIO_RPC\nLATEST_HEIGH=$LATEST_HEIGHT\nTRUST_HEIGHT=$TRUST_HEIGHT\nTRUST_HASH=$TRUST_HASH\n"

# Enable state sync
provenanced config set \
    statesync.enable true \
    statesync.rpc_servers "[\"$PIO_RPC\",\"$PIO_RPC\"]" \
    statesync.trust_height "$TRUST_HEIGHT" \
    statesync.trust_hash "$TRUST_HASH" \
    p2p.seeds '2de841ce706e9b8cdff9af4f137e52a4de0a85b2@104.196.26.176:26656,add1d50d00c8ff79a6f7b9873cc0d9d20622614e@34.71.242.51:26656'

# start node
provenanced start --testnet --x-crisis-skip-assert-invariants

# start node (capture stdout & stderr to log file)
provenanced start --testnet --x-crisis-skip-assert-invariants --log_level=info &>> pio.log
```
---

### Power users

#### Using [direnv](https://github.com/direnv/direnv) to run both `mainnet` and `testnet`

See [installing direnv](https://github.com/direnv/direnv/blob/master/docs/installation.md) if it is not installed on your system.
```
# create directories (or directory of your choosing)
mkdir ~/.provenanced/{mainnet,testnet}

# generate mainnet .envrc
echo -e "export PIO_HOME=\$(pwd)" > ~/.provenanced/mainnet/.envrc

# generate testnet .envrc
echo -e "export PIO_HOME=\$(pwd)\nexport PIO_TESTNET=true" > ~/.provenanced/testnet/.envrc
```

```
# open a new terminal window
cd ~/.provenanced/mainnet

# you should see
direnv: loading ~/.provenanced/mainnet/.envrc
direnv: export ~PIO_HOME
```

```
# open a new terminal window
cd ~/.provenanced/testnet

# you should see
direnv: loading ~/.provenanced/testnet/.envrc
direnv: export +PIO_TESTNET ~PIO_HOME
```
