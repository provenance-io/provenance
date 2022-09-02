#!/bin/bash

RESET_CONFIG="${RESET_CONFIG:-0}"
VERSION="${VERSION:-v2.0.0}"
RELAYER_PATH=$HOME/.relayer
RELAYER_CONFIG=$RELAYER_PATH/config/config.yaml
PROVENANCE_CONFIG=scripts/relayer-config.yaml

# Install relayer if it doesn't exist
if ! [ -x "$(command -v rly)" ]; then
    echo "relayer not installed. Installing..."
    # Install relayer
    DIR=$(pwd)
    git clone https://github.com/cosmos/relayer.git /tmp/relayer
    cd /tmp/relayer
    git checkout $VERSION
    make install

    # Cleanup installer
    cd $DIR
    rm -rf /tmp/relayer
fi

# Confirm the user wants to delete their previous config
if test -f "$RELAYER_CONFIG"; then
    if [[ $RESET_CONFIG != 1 ]]; then
        exit
    fi
    echo "Removing old config.yaml"
    rm $RELAYER_CONFIG
fi

# Setup configuration
echo "Creating new provenance relayer config"
rly config init
cp $PROVENANCE_CONFIG $RELAYER_CONFIG

echo "Setup complete"
