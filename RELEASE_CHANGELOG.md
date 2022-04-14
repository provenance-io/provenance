## [v1.8.1](https://github.com/provenance-io/provenance/releases/tag/v1.8.1) - 2022-04-13

### Summary

Provenance 1.8.1 includes upgrades to the underlying Cosmos SDK and adds initial support for ADR-038.

This release addresses issues related to IAVL concurrency and Tendermint performance that resulted in occasional panics when under high-load conditions such as replay from quicksync. In particular, nodes which experienced issues with "Value missing for hash" and similar panic conditions should work properly with this release. The underlying Cosmos SDK `0.45.3` release that has been incorporated includes a number of improvements around IAVL locking and performance characteristics.

** NOTE: Although Provenance supports multiple database backends, some issues have been reported when using the `goleveldb` backend. If experiencing issues, using the `cleveldb` backend is preferred **

### Improvements

* Update Provenance to use Cosmos SDK 0.45.3 Release [\#781](https://github.com/provenance-io/provenance/issues/781)
* Plugin architecture for ADR-038 + FileStreamingService plugin [\#10639](https://github.com/cosmos/cosmos-sdk/pull/10639)
* Fix for sporadic error "panic: Value missing for hash" [\#611](https://github.com/provenance-io/provenance/issues/611) 
