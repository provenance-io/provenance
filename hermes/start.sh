#!/bin/bash

# Load shell variables
. ./variables.sh

# Start the hermes relayer in multi-paths mode
echo "Starting hermes relayer..."
$HERMES_BINARY -c $CONFIG_DIR start
