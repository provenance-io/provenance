#!/bin/bash -e

PROV_DIR="$HOME/1.7.6/bin"
BUILD_DIR="./build"

if [ ! -d "$BUILDDIR/run/provenanced/config" ]; then
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" init --chain-id=testing testing
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" keys add validator --keyring-backend test
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" add-genesis-root-name validator pio --keyring-backend test
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" add-genesis-root-name validator pb --restrict=false --keyring-backend test
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" add-genesis-root-name validator io --restrict --keyring-backend test
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" add-genesis-root-name validator provenance --keyring-backend test
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" add-genesis-account validator 100000000000000000000nhash --keyring-backend test
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" gentx validator 1000000000000000nhash --keyring-backend test --chain-id=testing
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" add-genesis-marker 100000000000000000000nhash --manager validator --access mint,burn,admin,withdraw,deposit --activate --keyring-backend test
		"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" collect-gentxs
fi ;

"$PROV_DIR/provenanced" -t --home "$BUILD_DIR/run/provenanced" start