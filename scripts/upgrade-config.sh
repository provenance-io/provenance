#!/bin/bash

# This was taken from https://github.com/channa-figure/pio-scratch/blob/main/pio_clean_setup_run_env.sh
# This should be run after a `make clean build run-config`
# Further instructions on doing a local software upgrade can be found in the docs/software_upgrade.md file

cat ./build/run/provenanced/config/genesis.json | jq ' .app_state.gov.voting_params.voting_period="'${VOTING_PERIOD}'" ' | tee ./build/run/provenanced/config/genesis.json
cat ./build/run/provenanced/config/genesis.json | grep voting