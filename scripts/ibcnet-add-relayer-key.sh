#!/bin/bash

# Verify that jq exists
if ! command -v jq &> /dev/null
then
    echo "jq could not be found."
    exit 1
fi

if [ "$#" -ne 1 ]; then
    echo "The relayer key must be the only argument passed in."
    exit 1
fi

RELAYER_KEY=$1
TEMP=tmp

# We probably want to verify that the address doesn't exist
ACCOUNTS=$(jq . build/ibc0-0/config/genesis.json | jq '.app_state.auth.accounts' | jq length)
if [ "$ACCOUNTS" -eq "1" ]; then
    GENESIS=build/ibc0-0/config/genesis.json
    echo "Updating $GENESIS"
    jq . "$GENESIS" | jq --arg KEY "$RELAYER_KEY" '.app_state.auth.accounts += [{"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": "\($KEY)", "pub_key": null, "account_number": "1", "sequence": "0"}]' | jq --arg KEY "$RELAYER_KEY" '.app_state.bank.balances += [{"address": "\($KEY)", "coins": [{"denom": "nhash","amount": "100000000000000000"}]}]' | jq '.app_state.bank.balances[0].coins[0].amount = "99900000000000000000"' > "$TEMP" && mv "$TEMP" "$GENESIS"
else
    echo "Genesis file is already updated for ibc0. Skipping..."
fi

ACCOUNTS=$(jq . build/ibc1-0/config/genesis.json | jq '.app_state.auth.accounts' | jq length)
if [ "$ACCOUNTS" -eq "1" ]; then
    GENESIS=build/ibc1-0/config/genesis.json
    echo "Updating $GENESIS"
    jq . "$GENESIS" | jq --arg KEY "$RELAYER_KEY" '.app_state.auth.accounts += [{"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": "\($KEY)", "pub_key": null, "account_number": "1", "sequence": "0"}]' | jq --arg KEY "$RELAYER_KEY" '.app_state.bank.balances += [{"address": "\($KEY)", "coins": [{"denom": "nhash","amount": "100000000000000000"}]}]' | jq '.app_state.bank.balances[0].coins[0].amount = "99900000000000000000"' > "$TEMP" && mv "$TEMP" "$GENESIS"
else
    echo "Genesis file is already updated for ibc1. Skipping..."
fi