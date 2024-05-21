#!/bin/bash
# This script shortens the voting period to 21s or whatever the VOTING_PERIOD environment variable is.
# This should be run after a `make clean build run-config`
# Further instructions on doing a local software upgrade can be found in the docs/testing_software_upgrade.md file

fn='./build/run/provenanced/config/genesis.json'
json="$( cat "$fn" )"
jq --arg vp "${VOTING_PERIOD:-21s}" ' .app_state.gov.params.voting_period=$vp | .app_state.gov.params.expedited_voting_period=$vp' <<< "$json" > "$fn"
grep voting_period "$fn"
