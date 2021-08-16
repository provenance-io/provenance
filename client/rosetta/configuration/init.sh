#!/bin/sh

set -e

echo "initializing provenanced"

provenanced init provenanced --chain-id testing --home /testrosetta/node0
#
## generate some data for the config
rovenanced -t keys add testUser --keyring-backend=test --home /testrosetta/node0
provenanced -t add-genesis-account testUser 1000000000000nhash --keyring-backend=test --home /testrosetta/node0
provenanced gentx testUser 10000000nhash --chain-id=chain-local --keyring-backend=test -t --home=/testrosetta/node0
# --generate-only
provenanced -t collect-gentxs --home /testrosetta/node0
sed -i 's/127.0.0.1/0.0.0.0/g' /testrosetta/node0/config/config.toml