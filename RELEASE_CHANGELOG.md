## [v1.7.0](https://github.com/provenance-io/provenance/releases/tag/v1.7.0) - 2021-09-03

### Features

* Marker governance proposal are supported in cli [#367](https://github.com/provenance-io/provenance/issues/367)
* Add ability to query metadata sessions by record [#212](https://github.com/provenance-io/provenance/issues/212)
* Add Name and Symbol Cosmos features to Marker Metadata [#372](https://github.com/provenance-io/provenance/issues/372)
* Add authz support to Marker module transfer `MarkerTransferAuthorization` [#265](https://github.com/provenance-io/provenance/issues/265)
  * Add authz grant/revoke command to `marker` cli
  * Add documentation around how to grant/revoke authz [#449](https://github.com/provenance-io/provenance/issues/449)
* Add authz and feegrant modules [PR 384](https://github.com/provenance-io/provenance/pull/384)
* Add Marker governance proposal for setting denom metadata [#369](https://github.com/provenance-io/provenance/issues/369)
* Add `config` command to cli for client configuration [#394](https://github.com/provenance-io/provenance/issues/394)
* Add updated wasmd for Cosmos 0.43 [#409](https://github.com/provenance-io/provenance/issues/409)
* Add Rosetta support and automated testing [#365](https://github.com/provenance-io/provenance/issues/365)
* Update wasm parameters to only allow smart contracts to be uploaded with gov proposal [#440](https://github.com/provenance-io/provenance/issues/440)
* Update `config` command [#403](https://github.com/provenance-io/provenance/issues/403)
  * Get and set any configuration field.
  * Get or set multiple configuration fields in a single invocation.
  * Easily identify fields with changed (non-default) values.
  * Pack the configs into a single json file with only changed (non-default) values.
  * Unpack the config back into the multiple config files (that also have documentation in them).

### Bug Fixes

* Fix for creating non-coin type markers through governance addmarker proposals [#431](https://github.com/provenance-io/provenance/issues/431)
* Marker Withdraw Escrow Proposal type is properly registered [#367](https://github.com/provenance-io/provenance/issues/367)
  * Target Address field spelling error corrected in Withdraw Escrow and Increase Supply Governance Proposals.
* Fix DeleteScopeOwner endpoint to store the correct scope [PR 377](https://github.com/provenance-io/provenance/pull/377)
* Marker module import/export issues  [PR384](https://github.com/provenance-io/provenance/pull/384)
  * Add missing marker attributes to state export
  * Fix account numbering issues with marker accounts and auth module accounts during import
  * Export marker accounts as a base account entry and a separate marker module record
  * Add Marker module governance proposals, genesis, and marker operations to simulation testing [#94](https://github.com/provenance-io/provenance/issues/94)
* Fix an encoding issue with the `--page-key` CLI arguments used in paged queries [#332](https://github.com/provenance-io/provenance/issues/332)
* Fix handling of optional fields in Metadata Write messages [#412](https://github.com/provenance-io/provenance/issues/412)
* Fix cli marker new example is incorrect [#415](https://github.com/provenance-io/provenance/issues/415)
* Fix home directory setup for app export [#457](https://github.com/provenance-io/provenance/issues/457)
* Correct an error message that was providing an illegal amount of gas as an example [#425](https://github.com/provenance-io/provenance/issues/425)

### API Breaking

* Fix for missing validation for marker permissions according to marker type.  Markers of type COIN can no longer have
  the Transfer permission assigned.  Existing permission entries on Coin type markers of type Transfer are removed
  during migration [#428](https://github.com/provenance-io/provenance/issues/428)

### Improvements

* Updated to Cosmos SDK Release v0.44 to resolve security issues in v0.43 [#463](https://github.com/provenance-io/provenance/issues/463)
  * Updated to Cosmos SDK Release v0.43  [#154](https://github.com/provenance-io/provenance/issues/154)
* Updated to go 1.17 [#454](https://github.com/provenance-io/provenance/issues/454)
* Updated wasmd for Cosmos SDK Release v0.43 [#409](https://github.com/provenance-io/provenance/issues/409)
  * CosmWasm wasmvm v0.16.0 [CHANGELOG](https://github.com/CosmWasm/wasmvm/blob/v0.16.0/CHANGELOG.md)
  * CosmWasm cosmwasm v0.16.0 [CHANGELOG](https://github.com/CosmWasm/cosmwasm/blob/v0.16.0/CHANGELOG.md)
* Updated to IBC-Go Module v1.0.1 [PR 445](https://github.com/provenance-io/provenance/pull/445)
* Updated log message for circulation adjustment [#381](https://github.com/provenance-io/provenance/issues/381)
* Updated third party proto files to pull from cosmos 0.43 [#391](https://github.com/provenance-io/provenance/issues/391)
* Removed legacy api endpoints [#380](https://github.com/provenance-io/provenance/issues/380)
* Removed v039 and v040 migrations [#374](https://github.com/provenance-io/provenance/issues/374)
* Dependency Version Updates
  * Build/CI - cache [PR 420](https://github.com/provenance-io/provenance/pull/420), workflow clean up
  [PR 417](https://github.com/provenance-io/provenance/pull/417), diff action [PR 418](https://github.com/provenance-io/provenance/pull/418)
  code coverage [PR 416](https://github.com/provenance-io/provenance/pull/416) and [PR 439](https://github.com/provenance-io/provenance/pull/439),
  setup go [PR 419](https://github.com/provenance-io/provenance/pull/419), [PR 451](https://github.com/provenance-io/provenance/pull/451)
  * Google UUID 1.3.0 [PR 446](https://github.com/provenance-io/provenance/pull/446)
  * GRPC 1.3.0 [PR 443](https://github.com/provenance-io/provenance/pull/443)
  * cast 1.4.1 [PR 442](https://github.com/provenance-io/provenance/pull/442)
* Updated `provenanced init` for better testnet support and defaults [#403](https://github.com/provenance-io/provenance/issues/403)
* Fixed some example address to use the appropriate prefix [#453](https://github.com/provenance-io/provenance/issues/453)

