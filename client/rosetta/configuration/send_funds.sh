#!/bin/sh

set -e
addr=$(provenanced -t keys show node0 -a --keyring-backend=test --home=/testrosetta/node0)
echo "12345678" | provenanced -t tx bank send "$addr" "$1" 100000000000nhash --chain-id="chain-local" --node tcp://provenance:26657 --yes --keyring-backend=test  --home=/testrosetta/node0 --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5
