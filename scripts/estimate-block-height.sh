#!/usr/bin/env bash
usage() {
    echo "Usage: $0 -d <desired-date> -b <block-cut-time(5s)> -c <chain-id(pio-testnet-1)>
Example: ./estimate-block-height.sh -d '2022-06-22 15:30:00' -b 4.56 -c pio-testnet-1"
}
desiredDate=""
cutTime=5
chainId="pio-testnet-1"
while getopts "d:b:c:" arg; do
    case "${arg}" in
        d)
            desiredDate="${OPTARG}"
            ;;
        b)
            cutTime="${OPTARG}"
            ;;
        c)
            chainId="${OPTARG}"
            ;;
        *)
            usage
            exit 0
    esac
done
if [ "$desiredDate" = "" ]; then
    echo "-d is a required flag. You must specify your desired date and time in '%Y-%m-%d %H:%M:%S' format. Current time zone of the machine is used. Don't forget to use a 24 hour time format (1pm = 13)"
    exit 1
fi
case $chainId in
    pio-testnet-1)
        host="https://rpc.test.provenance.io:443"
        ;;
    pio-mainnet-1)
        host="https://rpc.provenance.io"
        ;;
    testnet-beta)
        host="35.243.142.236:26657"
        ;;
esac
status=$(curl -sSL "${host}/status")
currentHeight=$(jq -r .result.sync_info.latest_block_height <<< $status )
desiredSeconds=$(date -j -f "%Y-%m-%d %H:%M:%S" "$desiredDate" "+%s")
currentSeconds=$(date +%s)
secondsDiff=$((desiredSeconds-currentSeconds))
if ! command -v provenanced &> /dev/null; then
    isValid="provenanced is not installed and the upgrade height could not be checked for validity. Please install provenanced to ensure a valid block height is chosen"
else
    echo ""
    votingPeriodNano=$(provenanced q gov params --node $host -o json | jq -r '.voting_params.voting_period')
    votingPeriodSec=$((votingPeriodNano/1000000000))
    if [ "$secondsDiff" -gt "$votingPeriodSec" ]; then
        isValid="This block height is valid and will occur after the voting period ends"
    else
        isValid="WARNING: This block height is INVALID. The chosen block height is likely to occur before the voting period ends. Please choose a later date and time."
    fi
fi
futureBlocks=$(echo "($secondsDiff/$cutTime)/1" | bc)
targetBlock=$((currentHeight+futureBlocks))
echo "Target block height is $targetBlock"
echo $isValid
