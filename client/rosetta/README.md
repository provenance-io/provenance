# rosetta

This directory contains the files required to run the rosetta CI. It builds `provenanced` based on the current codebase.

## docker-compose.yaml

Builds:

- provenanced, setup using the testnode configuration just like localnet
- faucet is required so we can test the construction API by transfering funds through send_funds.sh
- rosetta is the rosetta node used by rosetta-cli to interact with the cosmos-sdk app
- test_rosetta in the Makefile runs the rosetta-cli test against construction API and data API

## configuration

Contains the required files to set up rosetta cli and make it work against its workflows.

## rosetta-cli

Contains the files for a deterministic network, with fixed keys and some actions on there, to test parsing of msgs and historical balances.

## Notes

- Keyring password is 12345678
- This implementation is a modification from cosmos-sdk: https://github.com/cosmos/cosmos-sdk/tree/master/contrib/rosetta
