## [v1.17.1](https://github.com/provenance-io/provenance/releases/tag/v1.17.1) - 2024-01-11

TODO: Writeup/Instructions

### Features

* Add CLI commands for the exchange module endpoints and queries [#1701](https://github.com/provenance-io/provenance/issues/1701).
* Create CLI commands for adding a market to a genesis file [#1757](https://github.com/provenance-io/provenance/issues/1757).
* Add CLI command to generate autocomplete shell scripts [#1762](https://github.com/provenance-io/provenance/pull/1762).

### Improvements

* Add StoreLoader wrapper to check configuration settings [#1792](https://github.com/provenance-io/provenance/pull/1792).
* Create a default market in `make run`, `localnet`, `devnet` and the `provenanced testnet` command [#1757](https://github.com/provenance-io/provenance/issues/1757).
* Updated documentation for each module to work with docusaurus [PR 1763](https://github.com/provenance-io/provenance/pull/1763)

### Bug Fixes

* Deprecate marker proposal transaction [#1797](https://github.com/provenance-io/provenance/issues/1797).

### Dependencies

- Bump `github.com/spf13/cobra` from 1.7.0 to 1.8.0 ([#1733](https://github.com/provenance-io/provenance/pull/1733))
- Bump `github.com/CosmWasm/wasmvm` from 1.2.4 to 1.2.6 ([#1799](https://github.com/provenance-io/provenance/issues/1799))
- Bump `github.com/CosmWasm/wasmd` from v0.30.0-pio-5 to v0.30.0-pio-6 ([#1799](https://github.com/provenance-io/provenance/issues/1799))

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.17.0...v1.17.1

