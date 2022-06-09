# State Sync

## What is State sync?

The Provenance 1.10.0+ release comes with Tendermint Core 0.34 which includes support for state sync. State sync allows a new node to join a network by fetching a snapshot of the application state at a recent height vs. fetching and replaying all historical blocks (which can take days). This can reduce the time needed to sync with the network down to minutes.

<!-- TOC -->
- [What is State sync?](#what-is-state-sync)
    - [Starting a new node from State Sync](#starting-a-new-node-from-state-sync)
        - [Mainnet](#mainnet)
        - [Testnet](#testnet)
    - [Power users](#power-users)


### Starting a new node from State Sync
___

#### Mainnet

##### 1. Install the `provenanced` binary

```
# setup home directory (or directory of your choosing)
export PIO_HOME=~/.provenanced

# clone and checkout
git clone https://github.com/provenance-io/provenance.git
cd provenance
git checkout tags/v1.10.0 -b v1.10.0

# install binary
make clean install

# initialize node
provenanced init choose-a-moniker --chain-id pio-mainnet-1
curl https://raw.githubusercontent.com/provenance-io/mainnet/main/pio-mainnet-1/genesis.json > genesis.json
mv genesis.json $PIO_HOME/config

# update config to use 'cleveldb'
sed -i.orig -E "s|^(db_backend[[:space:]]+=[[:space:]]+).*$|\1\"cleveldb\"| ;" $PIO_HOME/config/config.toml
```

##### 2. Start a new node from State Sync

```  
# setup sync node
# Chose one of three available host: rpc-{0, 1 or 2}.provenance.io
PIO_RPC=`host rpc-0.provenance.io | awk '{print $4}'`:26657

# State Sync Configuration Options
LATEST_HEIGHT=$(curl -s $PIO_RPC/block | jq -r .result.block.header.height); \
TRUST_HEIGHT=$((LATEST_HEIGHT - 1000)); \
TRUST_HASH=$(curl -s "$PIO_RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash); \
echo -e "\nLATEST_HEIGH=$LATEST_HEIGHT\nTRUST_HEIGHT=$TRUST_HEIGHT\nTRUST_HASH=$TRUST_HASH\n"

# Enable state sync
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; \
    s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"$PIO_RPC,$PIO_RPC\"| ; \
    s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$TRUST_HEIGHT| ; \
    s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"| ; \
    s|^(seeds[[:space:]]+=[[:space:]]+).*$|\1\"\"|" $PIO_HOME/config/config.toml

# start node
provenanced start \
--p2p.seeds 4bd2fb0ae5a123f1db325960836004f980ee09b4@seed-0.provenance.io:26656,048b991204d7aac7209229cbe457f622eed96e5d@seed-1.provenance.io:26656 \
--x-crisis-skip-assert-invariants

# start node (capture stdout & stderr to log file)
provenanced start \
--p2p.seeds 4bd2fb0ae5a123f1db325960836004f980ee09b4@seed-0.provenance.io:26656,048b991204d7aac7209229cbe457f622eed96e5d@seed-1.provenance.io:26656 \
--x-crisis-skip-assert-invariants \
--log_level=info &>> pio.log
```

#### Testnet

##### 1. Install the `provenanced` binary

```
# setup home directory (or directory of your choosing)
export PIO_HOME=~/.provenanced

# clone and checkout
git clone https://github.com/provenance-io/provenance.git
cd provenance
git checkout tags/v1.10.0 -b v1.10.0

# install binary
make clean install

# initialize node
provenanced init choose-a-moniker --chain-id pio-testnet-1
curl https://raw.githubusercontent.com/provenance-io/testnet/main/pio-testnet-1/genesis.json > genesis.json
mv genesis.json $PIO_HOME/config

# update config to use 'cleveldb'
sed -i.orig -E "s|^(db_backend[[:space:]]+=[[:space:]]+).*$|\1\"cleveldb\"| ;" $PIO_HOME/config/config.toml
```

##### 2. Start a new node from State Sync

```  
# setup sync node
PIO_RPC=34.66.209.228:26657

# State Sync Configuration Options
LATEST_HEIGHT=$(curl -s $PIO_RPC/block | jq -r .result.block.header.height); \
TRUST_HEIGHT=$((LATEST_HEIGHT - 1000)); \
TRUST_HASH=$(curl -s "$PIO_RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash); \
echo -e "\nLATEST_HEIGH=$LATEST_HEIGHT\nTRUST_HEIGHT=$TRUST_HEIGHT\nTRUST_HASH=$TRUST_HASH\n"

# Enable state sync
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; \
    s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"$PIO_RPC,$PIO_RPC\"| ; \
    s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$TRUST_HEIGHT| ; \
    s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"| ; \
    s|^(seeds[[:space:]]+=[[:space:]]+).*$|\1\"\"|" $PIO_HOME/config/config.toml

# start node
provenanced start \
--testnet \
--p2p.seeds 2de841ce706e9b8cdff9af4f137e52a4de0a85b2@104.196.26.176:26656,add1d50d00c8ff79a6f7b9873cc0d9d20622614e@34.71.242.51:26656 \
--x-crisis-skip-assert-invariants

# start node (capture stdout & stderr to log file)
provenanced start \
--testnet \
--p2p.seeds 2de841ce706e9b8cdff9af4f137e52a4de0a85b2@104.196.26.176:26656,add1d50d00c8ff79a6f7b9873cc0d9d20622614e@34.71.242.51:26656 \
--x-crisis-skip-assert-invariants \
--log_level=info &>> pio.log
```
---

### Power users

#### Using [direnv](https://github.com/direnv/direnv) to run both `mainnet` and `testnet`

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


See [installing direnv](https://github.com/direnv/direnv/blob/master/docs/installation.md) for instructions if it is not installed.