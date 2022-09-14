#!/bin/bash

RELAY_PATH="${RELAY_PATH:=local_local2}"
RELAY_CMD=rly

# We have to get CHAIN_1 and CHAIN_2 by splitting the path
CHAIN_1=$(echo "${RELAY_PATH}" | cut -d "_" -f 1)
CHAIN_2=$(echo "${RELAY_PATH}" | cut -d "_" -f 2)

check_keys() {
    CHAIN=$1
    ${RELAY_CMD} keys show $CHAIN &> /dev/null
    status=$?
    if [ $status != 0 ]; then
        echo "No keys exist for $CHAIN" >&2
        echo "Consider using an existing key with 'rly keys restore $CHAIN default \"mneomnic\"'" >&2
        echo "Alternatively, if an account does not exist you can create one with 'rly keys add $CHAIN default' and then funding it" >&2
    fi
    return $status
}

check_links() {
    PATH_STATUS=$(${RELAY_CMD} paths list | grep ${RELAY_PATH})
    PASSING=$(grep -o "✔" <<< $PATH_STATUS | wc -l | tr -d ' ')
    if [ $PASSING == 3 ]; then
        return 0
    fi
    return 1
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

check_links ${RELAY_PATH}
if [ $? != 0 ]; then
    echo "${RELAY_PATH} is not fully functional. Consider checking the status of the path and trying again." >&2
    echo "Alternatively, if the path has not been setup yet, then you can create one with 'rly tx link ${RELAY_PATH}'" >&2
    exit 1
fi

${RELAY_CMD} start ${RELAY_PATH} -p events