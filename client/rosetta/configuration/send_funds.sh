#!/bin/sh

set -e
addr=$(provenanced -t keys show fredkneeland -a --keyring-backend=test --home=/Users/fredkneeland/code/provenance/build/run/provenanced)
echo "12345678" | provenanced -t tx bank send "$addr" "$1" 100000000000nhash --chain-id="testing" --node tcp://localhost:26657 --yes --keyring-backend=test  --home=/Users/fredkneeland/code/provenance/build/run/provenanced --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5
echo "12345678" | provenanced -t tx bank send "$addr" "$1" 10bitcoin --chain-id="testing" --node tcp://localhost:26657 --yes --keyring-backend=test  --home=/Users/fredkneeland/code/provenance/build/run/provenanced --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5
