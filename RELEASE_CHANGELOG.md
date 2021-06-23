## [v1.5.0](https://github.com/provenance-io/provenance/releases/tag/v1.5.0) - 2021-06-23

### Features

* Update Cosmos SDK to 0.42.6 with Tendermint 0.34.11 [#355](https://github.com/provenance-io/provenance/issues/355)
  * Refund gas support added to gas meter trace
  * `ibc-transfer` now contains an `escrow-address` command for querying current escrow balances
* Add `update` and `delete-distinct` attributes to `attribute` module [#314](https://github.com/provenance-io/provenance/issues/314)
* Add support to `metadata` module for adding and removing contract specifications to scope specification [#302](https://github.com/provenance-io/provenance/issues/302)
  * Added `MsgAddContractSpecToScopeSpecRequest`and `MsgDeleteContractSpecFromScopeSpecRequest` messages for adding/removing
  * Added cli commands for adding/removing
* Add smart contract query support to the `metadata` module [#65](https://github.com/provenance-io/provenance/issues/65)

### API Breaking

* Redundant account parameter was removed from Attribute module SetAttribute API. [PR 348](https://github.com/provenance-io/provenance/pull/348)

### Bug Fixes

* Value owner changes are independent of scope owner signature requirements after transfer [#347](https://github.com/provenance-io/provenance/issues/347)
* Attribute module allows removal of orphan attributes, attributes against root names [PR 348](https://github.com/provenance-io/provenance/pull/348)
* `marker` cli query for marker does not cast marker argument to lower case [#329](https://github.com/provenance-io/provenance/issues/329)

### Improvements

* Bump `wasmd` to v0.17.0 [#345](https://github.com/provenance-io/provenance/issues/345)
* Attribute module simulation support [#25](https://github.com/provenance-io/provenance/issues/25)
* Add transfer cli command to `marker` module [#264](https://github.com/provenance-io/provenance/issues/264)
* Refactor `name` module to emit typed events from keeper [#267](https://github.com/provenance-io/provenance/issues/267)

