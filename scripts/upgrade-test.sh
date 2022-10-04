#!/usr/bin/env bash
# This script will submit a test upgrade proposal, and vote for it.
# It's assumed that you're running it from the root of this repo.
# This script looks to the following env vars for configuration:
#   PIO_HOME:  Default: build/run/provenanced
#   PIO_CHAIN_ID:  Default: testing
#   PIO_KEYRING_BACKEND:  Default: test
#   PROV:  The provenanced executable. Default: build/provenanced
#   DEPOSIT:  The --deposit amount for the upgrade gov proposal. Default: 10000000nhash
#   GAS_PRICES:  The --gas-prices amount for the tx commands. Default: 1905nhash
#   GAS:  The --gas amount for the tx commands. Default: auto
#   GAS_ADJUSTMENT:  The --gas-adjustment amount for tx commands. Only used for --gas=auto. Default: 1.5
#   HEIGHT_PLUS:  The number of blocks in the future to do the upgrade (must happen after voting ends). Default: 40

if [[ -z "$1" ]]; then
    printf 'Usage: %s <color>\n' "$0"
    exit 1
fi

color="$1"
shift


set -ex
export PIO_TESTNET=true
export PIO_OUTPUT=json
export PIO_HOME="${PIO_HOME:-build/run/provenanced}"
export PIO_CHAIN_ID="${PIO_CHAIN_ID:-testing}"
export PIO_KEYRING_BACKEND="${PIO_KEYRING_BACKEND:-test}"
prov="${PROV:-build/provenanced}"
deposit="${DEPOSIT:-10000000nhash}"
gasPrices="${GAS_PRICES:-1905nhash}"
gas="${GAS:-auto}"
gasAdj="${GAS_ADJUSTMENT:-1.5}"
gasArgs="--gas-prices=$gasPrices --gas=$gas"
if [ "$gas" == 'auto' ]; then
    gasArgs="$gasArgs --gas-adjustment=$gasAdj"
fi
heightPlus="${HEIGHT_PLUS:-40}"

prop_cmd='submit-proposal'
no_val=''
if "$prov" tx gov "$prop_cmd" software-upgrade --help 2>&1 | grep -qF 'proposal.json' > /dev/null 2>&1; then
    prop_cmd='submit-legacy-proposal'
    no_val='--no-validate'
fi


valAddr="$( "$prov" keys list --output json | jq -r '.[] | select( .name == "validator" ) | .address' )"
curHeight="$( "$prov" status | jq -r '.SyncInfo.latest_block_height' )"
targetHeight="$(( curHeight + heightPlus ))"

propRes="$( "$prov" tx gov "$prop_cmd" software-upgrade "$color" \
    --title "Upgrade for $color" \
    --description "Upgrading provenance to $color" \
    --upgrade-info="$color" \
    --upgrade-height "$targetHeight" \
    --deposit "$deposit" \
    --from "$valAddr" \
    $gasArgs $no_val \
    --yes )"

propId="$( tail -n 1 <<< "$propRes" | jq -r '.logs[0].events[] | select( .type == "submit_proposal" ) | .attributes[] | select( .key == "proposal_id" ) | .value' )"

"$prov" tx gov vote "$propId" yes \
    --from "$valAddr" \
    $gasArgs \
    --yes

"$prov" query gov tally "$propId"

printf 'Success: Upgrade will happen at height=%d\n' "$targetHeight"
