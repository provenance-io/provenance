#!/bin/bash
set -e

# Load shell variables
. ./variables.sh

### Configure the clients and connection
echo "Initiating connection handshake..."
$HERMES_BINARY -c $CONFIG_DIR create connection testing testing2

sleep 2
