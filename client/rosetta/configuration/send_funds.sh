#!/bin/sh

set -e
addr=$(simd keys show fd -a --keyring-backend=test)
echo "12345678" | provenanced tx bank send "$addr" "$1" 100stake --chain-id="testing" --node tcp://api.provenance.io:26657 --yes --keyring-backend=test