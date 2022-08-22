## [v1.12.0](https://github.com/provenance-io/provenance/releases/tag/v1.12.0) - 2022-08-22

### Improvements

* Update the swagger files (including third-party changes). [#728](https://github.com/provenance-io/provenance/issues/728)
* Bump IBC to 2.3.0 and update third-party protos [PR 868](https://github.com/provenance-io/provenance/pull/868)
* Update docker images from `buster` to b`bullseye` [#963](https://github.com/provenance-io/provenance/issues/963)
* Add documentation for `gRPCurl` to `docs/grpcurl.md` [#953](https://github.com/provenance-io/provenance/issues/953)
* Updated to go 1.18 [#996](https://github.com/provenance-io/provenance/issues/996)
* Add docker files for local psql indexing [#997](https://github.com/provenance-io/provenance/issues/997)

### Features

* Bump Cosmos-SDK to `v0.45.4-pio-4` (from `v0.45.4-pio-2`) to utilize the new `CountAuthorization` authz grant type. [#807](https://github.com/provenance-io/provenance/issues/807)
* Update metadata module authz handling to properly call `Accept` and delete/update authorizations as they're used [#905](https://github.com/provenance-io/provenance/issues/905)
* Read the `custom.toml` config file if it exists. This is read before the other config files, and isn't managed by the `config` commands [#989](https://github.com/provenance-io/provenance/issues/989)

### Bug Fixes

* Support standard flags on msgfees params query cli command [#936](https://github.com/provenance-io/provenance/issues/936)
* Fix the `MarkerTransferAuthorization` Accept function and `TransferCoin` authz handling to prevent problems when other authorization types are used [#903](https://github.com/provenance-io/provenance/issues/903)
* Bump Cosmos-SDK to `v0.45.5-pio-1` (from `v0.45.4-pio-4`) to remove buggy ADR 038 plugin system. [#983](https://github.com/provenance-io/provenance/issues/983)
* Remove ADR 038 plugin system implementation due to `AppHash` error [#983](https://github.com/provenance-io/provenance/issues/983)
* Fix fee charging to sweep remaining fees on successful transaction [#1019](https://github.com/provenance-io/provenance/issues/1019)

### State Machine Breaking

* Fix the `MarkerTransferAuthorization` Accept function and `TransferCoin` authz handling to prevent problems when other authorization types are used [#903](https://github.com/provenance-io/provenance/issues/903)
* Fix fee charging to sweep remaining fees on successful transaction [#1019](https://github.com/provenance-io/provenance/issues/1019)
