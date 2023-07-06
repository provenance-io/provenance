#!/usr/bin/env bash

##
## Input parameters
##

export RELAYER_HOME="/relayer"
export RELAY_PATH="local_local2"

# Setup the connection, client, and channel on first run
PATH_COUNT=$(rly --home "$RELAYER_HOME" q connections local | wc -l)
if [[ "$PATH_COUNT" -eq 0 ]];then
    echo "Initializing the relayer"
    rly tx link "$RELAY_PATH" --home "$RELAYER_HOME"
fi

# Start the relayer
echo "Starting the relayer"
rly start "$RELAY_PATH" -p events --home "$RELAYER_HOME"