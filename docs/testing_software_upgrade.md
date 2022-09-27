# Software upgrade in Provenance
---

## Purpose
This document outlines how to do a local software upgrade for testing purposes only.  The purpose is to help developers who want to test an upgrade with a local node running from the `make run` command inside of the Provenance repository.

## Overview
An upgrade in Provenance is done by a node submitting a proposal with a deposit in hash.  This proposal contains the blockchain height as well as the upgrade name.  Once a sufficient deposit has been made the proposal enters a voting period where any user with hash may vote.  After the voting period, defined in the genesis.json, the proposal will either pass or fail.  If the proposal passes, the chain will stop at the height of the proposal.  Then the chain must be restarted with the upgrade handler with the upgrade name.  This upgrade handler will execute and the chain will continue running with the software upgrade.

## Steps
1. [ ] Check out previous version of Provenance.
    - For example: `git checkout v1.7.5`

2. [ ] Build and configure Provenance:
    - `go mod vendor && make clean build run-config`

3. [ ] Update the genesis.json file to shorten the voting period to 20s:
    - `./scripts/upgrade-config.sh`

4. [ ] Run the chain:
    - `make run`

5. [ ] In a separate terminal window, submit a new proposal, vote on it, and get the tally:
    - `./scripts/upgrade-test.sh <upgrade color>`

6. [ ] Wait for the chain to halt with an error when it reaches the upgrade height specified.
7. [ ] Check out the branch with the upgrade handler.
8. [ ] Run `go mod vendor && make build run` to restart the chain triggering the software upgrade.
9. [ ] Verify and enjoy any new functionality from the upgraded version. :)
