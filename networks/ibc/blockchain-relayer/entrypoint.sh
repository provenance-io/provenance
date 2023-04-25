#!/usr/bin/env bash

##
## Input parameters
##

export RELAYER_HOME="/relayer"
export RELAY_PATH="local_local2"

# Setup the connection, client, and channel
#rly tx link "$RELAY_PATH" --home "$RELAYER_HOME"

# Start the relayer
#rly start "$RELAY_PATH" -p events --home "$RELAYER_HOME"