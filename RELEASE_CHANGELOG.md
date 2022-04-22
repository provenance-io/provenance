## [v1.8.2](https://github.com/provenance-io/provenance/releases/tag/v1.8.2) - 2022-04-22

### Summary

Provenance 1.8.2 is a point release to fix an issue with "downgrade detection" in Cosmos SDK. A panic condition 
occurs in cases where no update handler is found for the last known upgrade, but the process for determining
the last known upgrade is flawed in Cosmos SDK 0.45.3. This released uses an updated Cosmos fork to patch the
issue until an official patch is released. Version 1.8.2 also adds some remaining pieces for  ADR-038 that were 
missing in the 1.8.1 release.

### Bug Fixes

* Order upgrades by block height rather than name to prevent panic [\#106](https://github.com/provenance-io/cosmos-sdk/pull/106)

### Improvements

* Add remaining updates for ADR-038 support [\#786](https://github.com/provenance-io/provenance/pull/786)
