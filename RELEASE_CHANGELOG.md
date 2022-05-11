## [v1.10.0](https://github.com/provenance-io/provenance/releases/tag/v1.10.0) - 2022-05-11

### Summary

Provenance 1.10.0 includes upgrades to the underlying CosmWasm dependencies and adds functionality to
remove orphaned metadata in the bank module left over after markers have been deleted.

### Improvements

* Update wasmvm dependencies and update Dockerfile for localnet [#818](https://github.com/provenance-io/provenance/issues/818)
* Remove "send enabled" on marker removal and in bulk on 1.10.0 upgrade [#821](https://github.com/provenance-io/provenance/issues/821)
