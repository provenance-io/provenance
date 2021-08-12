#!/bin/sh

set -e

wait_provenanced() {
  timeout 30 sh -c 'until nc -z $0 $1; do sleep 1; done' localhost 9090
}
# this script is used to recreate the data dir
echo clearing /root/.provenanced
rm -rf /root/.provenanced
echo initting new chain
# init config files
provenanced init provenanced --chain-id testing

# create accounts
provenanced keys add fd --keyring-backend=test

addr=$(provenanced -t keys show fd -a --keyring-backend=test)
val_addr=$(provenanced -t keys show fd  --keyring-backend=test --bech val -a)

# give the accounts some money
provenanced add-genesis-account "$addr" 1000000000000stake --keyring-backend=test

# save configs for the daemon
provenanced gentx fd 10000000stake --chain-id testing --keyring-backend=test

# input genTx to the genesis file
provenanced collect-gentxs
# verify genesis file is fine
provenanced validate-genesis
echo changing network settings
sed -i 's/127.0.0.1/0.0.0.0/g' /root/.provenance/config/config.toml

# start provenanced
echo starting provenanced...
provenanced start --pruning=nothing &
pid=$!
echo provenanced started with PID $pid

echo awaiting for provenanced to be ready
wait_provenanced
echo provenanced is ready
sleep 10


# send transaction to deterministic address
echo sending transaction with addr $addr
provenanced tx bank send "$addr" cosmos19g9cm8ymzchq2qkcdv3zgqtwayj9asv3hjv5u5 100nhash --yes --keyring-backend=test --broadcast-mode=block --chain-id=testing

sleep 10

echo stopping provenanced...
kill -9 $pid

echo zipping data dir and saving to /tmp/data.tar.gz

tar -czvf /tmp/data.tar.gz /root/.simapp

echo new address for bootstrap.json "$addr" "$val_addr"
