#!/bin/sh

set -e
addr=$(provenanced -t keys show testUser -a --keyring-backend=test)
echo "12345678" | provenanced -t tx bank send "$addr" "$1" 100nhash --chain-id="testing" --node tcp://provenance:26657 --yes --keyring-backend=test