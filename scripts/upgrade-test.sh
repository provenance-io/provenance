#!/usr/bin/env bash
# This script will submit a test upgrade proposal, and vote for it.
# It's assumed that you're running it from the root of this repo.
# Uses build/provenanced as the executable unless $PROV is set (and exported).
# Uses build/run/provenanced as PIO_HOME (--home) unless $PIO_HOME is set (and exported).

if [[ -z "$1" ]]; then
    printf 'Usage: %s <color>\n' "$0"
    exit 1
fi

color="$1"
shift

prov="${PROV:-build/provenanced}"

set -ex
export PIO_TESTNET=true
export PIO_HOME="${PIO_HOME:-build/run/provenanced}"
export PIO_OUTPUT=json

valAddr="$( "$prov" keys list --output json | jq -r '.[] | select( .name == "validator" ) | .address' )"
curHeight="$( "$prov" status | jq -r '.SyncInfo.latest_block_height' )"
targetHeight="$(( curHeight + 40 ))"

propRes="$( "$prov" tx gov submit-proposal software-upgrade "$color" \
    --title "Upgrade for $color" \
    --description "Upgrading provenance to $color" \
    --upgrade-info="$color" \
    --from "$valAddr" \
    --upgrade-height "$targetHeight" \
    --deposit 10000000nhash \
    --chain-id=testing \
    --keyring-backend test \
    --gas-prices='1905nhash' \
    --gas=auto \
    --gas-adjustment=1.5 \
    --yes )"

propId="$( tail -n 1 <<< "$propRes" | jq -r '.logs[0].events[] | select( .type == "submit_proposal" ) | .attributes[] | select( .key == "proposal_id" ) | .value' )"

"$prov" tx gov vote "$propId" yes \
    --gas-prices='1905nhash' \
    --gas=auto \
    --gas-adjustment=1.5 \
    --from "$valAddr" \
    --keyring-backend=test \
    --chain-id=testing \
    --yes

"$prov" query gov tally "$propId"

printf 'Success: Upgrade will happen at height=%d\n' "$targetHeight"
