# Upgrade Data Directory

This directory contains data files that are embedded into the binary so that they can be loaded during the upgrade process.

The `embed` library is used to make the file accessible programmatically in the `upgradeDataFS` variable (file system).
The paths will be `upgrade_data/<filename>`.

## File Size Guidelines

- **Small files (< 1MB uncompressed)**: Can be single files. Does not need to be compressed.
- **Medium files (1-10MB uncompressed)**: Consider splitting into chunks. Also consider gzipping it.
- **Large files (> 10MB uncompressed)**: Should be split into chunks. Should also be gzipped.

These files will increase the size of the compiled `provenanced` executable, and should be as small as possible.

## bouvardia_ledger_genesis.json.gz

This file contains a set of ledger data that is to be loaded during the `bouvardia` upgrade.
It is a gzipped JSON file containing a `ledgerTypes.GenesisState` object.

