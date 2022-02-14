#!/bin/bash

# Shorten the voting period from the normal 2 weeks for rapid local testing
VOTING_PERIOD=21s

# This was taken from https://github.com/channa-figure/pio-scratch/blob/main/pio_clean_setup_run_env.sh
# This should be run after a `make clean build run-config`
# Further instructions on doing a local software upgrade can be found in the docs/testing_software_upgrade.md file

cat ./build/run/provenanced/config/genesis.json | jq ' .app_state.gov.voting_params.voting_period="'${VOTING_PERIOD}'" ' | tee ./build/run/provenanced/config/genesis.json
cat ./build/run/provenanced/config/genesis.json | grep voting