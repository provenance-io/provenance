# Software upgrade in Provenance
---

## Purpose
This document outlines how to do a local software upgrade for testing purposes only.  The purpose is to help developers who want to test an upgrade with a local node running from the `make run` command inside of the Provenance repository.

## Overview
An upgrade in Provenance is done by a node submitting a proposal with a deposit in hash.  This proposal contains the blockchain height as well as the upgrade name.  Once a sufficient deposit has been made the proposal enters a voting period where any user with hash may vote.  After the voting period, defined in the genesis.json, the proposal will either pass or fail.  If the proposal passes, the chain will stop at the height of the proposal.  Then the chain must be restarted with the upgrade handler with the upgrade name.  This upgrade handler will execute and the chain will continue running with the software upgrade.

## Steps
1. [ ] Check out previous version of Provenance
    - For example: `git checkout v1.7.5`
   
2. [ ] Build and configure Provenance
    - `go mod vendor && make clean build run-config`
   
3. [ ] Update the genesis.json file to shorten the voting period to 20s
    - `./scripts/upgrade-config.sh`
   
4. [ ] Run the chain
    - `make run`
   
5. [ ] In a separate terminal window save the validator address:
    - `provenanced -t keys list --home ./build/run/provenanced`
    - `export valAddr="<addr listed above>"`
   
6. [ ] Submit a software upgrade proposal with the desired name and upgrade height.  The upgrade height should be at least 100 blocks above the current height to allow for the 20s voting period to finish before the height is reached or the upgrade will become invalid.
   ```bash
   provenanced -t tx gov submit-proposal software-upgrade "<name>" \
       --title "<title>" \
       --description "<description>" \
       --upgrade-info="<name>" \
       --from "$valAddr" \
       --upgrade-height 200 \
       --deposit 10000000nhash \
       --chain-id=testing \
       --keyring-backend test \
       --gas-prices="1905nhash" \
       --gas=auto \
       --gas-adjustment=1.5 \
       --home=./build/run/provenanced \
       --yes
   ```
   
7. [ ] Query the proposal with the following command:
   ```bash
   provenanced query gov proposals
   ```
   
8. [ ] Before the 20s voting period has elapsed vote with the following cmd:
   ```bash
   provenanced -t tx gov vote "1" yes\
       --gas-prices="1905nhash" \
       --gas=auto \
       --gas-adjustment=1.5 \
       --from "$valAddr" \
       --keyring-backend=test \
       --chain-id=testing \
       --home=./build/run/provenanced \
       --yes
   ```
   If there are more than one proposals you will need to replace "1" with the number of the proposal you wish to vote on.

9. [ ] Query the vote tally before the voting period has ended
   ```bash
   provenanced query gov tally 1
   ```

10. [ ] Wait for the chain to halt with an error when it reaches the upgrade height specified
11. [ ] Check out the branch with the upgrade handler
12. [ ] Run `go mod vendor && make build run` to restart the chain triggering the software upgrade
13. [ ] Verify and enjoy any new functionality from the upgraded version :) 