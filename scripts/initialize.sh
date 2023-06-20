#!/bin/bash
# This script will initialize a Provenance blockchain.

VERBOSE="${VERBOSE:-}"
for arg in "$@"; do
    case "$arg" in
        -h|--help)
            cat << EOF
Usage: $0 [-h|--help] [-v|--verbose]

The following environment variables control the behavior of this script:
  PIO_HOME ------------ The location of the home directory.
                        Default: $HOME/.provenance
  PROV_CMD ------------ The command to use to execute the provenanced binary.
                        Default: provenanced
  DENOM --------------- The denomination to use as the utility token.
                        Default: nhash
  MIN_FLOOR_PRICE ----- The minimum floor gas price that validators are allowed to set.
                        Default: 1905
  PIO_TESTNET --------- Whether this is a testnet setup.
                        Default: true
  PIO_KEYRING_BACKEND - The keyring backend type to use for the new validator's keys.
                        Default: test
  PIO_CHAIN_ID -------- The chain id to use.
                        Default: testing
  TIMEOUT_COMMIT ------ The consensus.timeout_commit value to set.
                        Default: defined by init command
  SHOW_START ---------- Whether to output how to start the chain (at the end).
                        Default: true

EOF
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=YES
            ;;
        *)
            printf 'Unknown argument: "%s"\n' "$arg"
            exit 1
            ;;
    esac
done

PIO_HOME="${PIO_HOME:-$HOME/.provenance}"
PROV_CMD=${PROV_CMD:-provenanced}
DENOM="${DENOM:-${PIO_CUSTOM_DENOM:-nhash}}"
MIN_FLOOR_PRICE="${MIN_FLOOR_PRICE:-1905}"
PIO_TESTNET="${PIO_TESTNET:-true}"
PIO_KEYRING_BACKEND="${PIO_KEYRING_BACKEND:-test}"
PIO_CHAIN_ID="${PIO_CHAIN_ID:-testing}"
SHOW_START="${SHOW_START:-true}"

# When the PROV_CMD is a docker thing, env vars don't get passed.
# So just always provide the needed ones as args.
if [ "$PIO_TESTNET" == 'true' ]; then
    PROV_CMD="$PROV_CMD -t"
fi
PROV_CMD="$PROV_CMD --home $PIO_HOME"
arg_chain_id="--chain-id=$PIO_CHAIN_ID"
arg_keyring="--keyring-backend $PIO_KEYRING_BACKEND"

if [ -n "$TIMEOUT_COMMIT" ]; then
  arg_timeout_commit="--timeout-commit $TIMEOUT_COMMIT"
fi

if [ -n "$VERBOSE" ]; then
    printf 'Initializing blockchain:\n'
    printf '%s=%s\n' \
        PIO_HOME "$PIO_HOME" \
        PROV_CMD "$PROV_CMD" \
        DENOM "$DENOM" \
        MIN_FLOOR_PRICE "$MIN_FLOOR_PRICE" \
        PIO_TESTNET "$PIO_TESTNET" \
        PIO_KEYRING_BACKEND "$PIO_KEYRING_BACKEND" \
        PIO_CHAIN_ID "$PIO_CHAIN_ID" \
        TIMEOUT_COMMIT "$TIMEOUT_COMMIT"
fi

set -ex
$PROV_CMD init testing --custom-denom "$DENOM" $arg_timeout_commit $arg_chain_id
$PROV_CMD keys add validator $arg_keyring
$PROV_CMD add-genesis-root-name validator pio $arg_keyring
$PROV_CMD add-genesis-root-name validator pb --restrict=false $arg_keyring
$PROV_CMD add-genesis-root-name validator io --restrict $arg_keyring
$PROV_CMD add-genesis-root-name validator provenance $arg_keyring
$PROV_CMD add-genesis-account validator "100000000000000000000$DENOM" $arg_keyring
$PROV_CMD gentx validator "1000000000000000$DENOM" $arg_chain_id $arg_keyring
$PROV_CMD add-genesis-marker "100000000000000000000$DENOM" \
    --manager validator \
    --access mint,burn,admin,withdraw,deposit \
    $arg_keyring \
    --activate
$PROV_CMD add-genesis-msg-fee /provenance.name.v1.MsgBindNameRequest "10000000000$DENOM"
$PROV_CMD add-genesis-msg-fee /provenance.marker.v1.MsgAddMarkerRequest "100000000000$DENOM"
$PROV_CMD add-genesis-msg-fee /provenance.attribute.v1.MsgAddAttributeRequest "10000000000$DENOM"
$PROV_CMD add-genesis-msg-fee /provenance.metadata.v1.MsgWriteScopeRequest "10000000000$DENOM"
$PROV_CMD add-genesis-custom-floor "${MIN_FLOOR_PRICE}${DENOM}"
$PROV_CMD collect-gentxs
$PROV_CMD config set minimum-gas-prices "${MIN_FLOOR_PRICE}${DENOM}"
set +ex

[ -n "$VERBOSE" ] && printf '\nProvenance Blockchain initialized at: %s\n' "$PIO_HOME"

if [ "$SHOW_START" == 'true' ]; then
    start_cmd="$PROV_CMD start"
    if [ "$DENOM" != 'nhash' ]; then
        start_cmd="$start_cmd --custom-denom $DENOM"
    fi
    printf '\nYou can start your node with the following command:\n  %s\n\n' "$start_cmd"
fi

exit 0
