#!/bin/bash

RESET_CONFIG="${RESET_CONFIG:-0}"
VERSION="${VERSION:-v2.3.1}"
RELAYER_HOME="${RELAYER_HOME:-$HOME/.relayer}"
RELAY_CMD="${RELAY_CMD:-rly}"
RELAYER_CONFIG="$RELAYER_HOME/config/config.yaml"
PROVENANCE_CONFIG=scripts/relayer-config.yaml

# Install relayer if it doesn't exist
if ! command -v "$RELAY_CMD" > /dev/null 2>&1; then
    echo "relayer not installed. Installing..."
    # Install relayer
    DIR="$( pwd )"
    git clone https://github.com/cosmos/relayer.git /tmp/relayer
    cd /tmp/relayer || exit 1 
    git checkout "$VERSION"
    make install

    # Cleanup installer
    cd "$DIR" || exit 1 
    rm -rf /tmp/relayer
fi

# Confirm the user wants to delete their previous config
if test -f "$RELAYER_CONFIG"; then
    if [[ "$RESET_CONFIG" != 1 ]]; then
        exit
    fi
    echo "Removing old config.yaml"
    rm "$RELAYER_CONFIG"
fi

# Setup configuration
echo "Creating new provenance relayer config"
"$RELAY_CMD" config init --home "$RELAYER_HOME"
cp "$PROVENANCE_CONFIG" "$RELAYER_CONFIG"

echo "Setup complete"
