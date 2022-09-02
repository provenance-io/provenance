#!/bin/bash

RELAY_PATH="${RELAY_PATH:-local_testnet}"

# We have to get CHAIN_1 and CHAIN_2 by splitting the path
CHAIN_1=$(echo "${RELAY_PATH}" | cut -d "_" -f 1)
CHAIN_2=$(echo "${RELAY_PATH}" | cut -d "_" -f 2)

check_keys() {
    CHAIN=$1
    rly keys show $CHAIN &> /dev/null
    status=$?
    if [ $status != 0 ]; then
        echo "No keys exist for $CHAIN" >&2
        echo "Consider using an existing key with 'rly keys restore $CHAIN default \"mneomnic\"'" >&2
        echo "Alternatively, if an account does not exist you can create one with 'rly keys add $CHAIN default' and then funding it" >&2
    fi
    return $status
}

check_keys ${CHAIN_1}
if [ $? != 0 ]; then
    echo "Cannot start relayer until ${CHAIN_1} has a key"
    exit 1
fi

check_keys ${CHAIN_2}
if [ $? != 0 ]; then
    echo "Cannot start relayer until ${CHAIN_2} has a key"
    exit 1
fi

rly start ${RELAY_PATH} -p events