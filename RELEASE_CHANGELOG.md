## [v1.13.0](https://github.com/provenance-io/provenance/releases/tag/v1.13.0) - 2022-11-28

### Features

* Add restricted marker transfer over ibc support [#1136](https://github.com/provenance-io/provenance/issues/1136).
* Enable the node query service [PR 1173](https://github.com/provenance-io/provenance/pull/1173).
* Add the `x/groups` module [#1007](https://github.com/provenance-io/provenance/issues/1007).
* Allow starting a `provenanced` chain using a custom denom [#1067](https://github.com/provenance-io/provenance/issues/1067).
  For running the chain locally, `make run DENOM=vspn MIN_FLOOR_PRICE=0` or `make clean localnet-start DENOM=vspn MIN_FLOOR_PRICE=0`.

### Improvements

* Updated Cosmos-SDK to `v0.46.6-pio-1` (from `v0.45.10-pio-4`) [PR 1235](https://github.com/provenance-io/provenance/pull/1235).
  This brings several new features and improvements. For details, see the [release notes](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.6-pio-1/RELEASE_NOTES.md) and [changelog](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.6-pio-1/CHANGELOG.md).
* Bump IBC to `v5.0.0-pio-2` (from `v2.3.0`) to add a check for SendEnabled [#1100](https://github.com/provenance-io/provenance/issues/1100), [#1158](https://github.com/provenance-io/provenance/issues/1158).
* Update wasmd to `v0.29.0-pio-1` (from `v0.26.0`) with SDK v0.46 support from notional-labs [#1015](https://github.com/provenance-io/provenance/issues/1015), [PR 1148](https://github.com/provenance-io/provenance/pull/1148).
* Allow MsgFee fees to be denoms other than `nhash` [#1067](https://github.com/provenance-io/provenance/issues/1067).
* Ignore hardcoded tx gas limit when `consensus_params.block.max_gas` is set to -1 for local nodes [#1000](https://github.com/provenance-io/provenance/issues/1000).
* Refactor the `x/marker` module's `Holding` query to utilize the `x/bank` module's new `DenomOwners` query [#995](https://github.com/provenance-io/provenance/issues/995).
  The only real difference between those two queries is that the `Holding` query accepts either a denom or marker address.
* Stop using the deprecated `Wrap` and `Wrapf` functions in the `sdk/types/errors` package in favor of those functions off specific errors, or else the `cosmossdk.io/errors` package [#1013](https://github.com/provenance-io/provenance/issues/995).
* For newly added reward's module, Voting incentive program, validator votes should count for higher shares, since they vote for all their delegations.
  This improvement allows the reward creator to introduce the multiplier to achieve the above.
* Refactored the fee handling [#1006](https://github.com/provenance-io/provenance/issues/1006):
  * Created a `MinGasPricesDecorator` to replace the `MempoolFeeDecorator` that was removed from the SDK. It makes sure the fee is greater than the validators min-gas fee.
  * Refactored the `MsgFeesDecorator` to only make sure there's enough fee provided. It no longer deducts/consumes anything and it no longer checks the payer's account.
  * Refactored the `ProvenanceDeductFeeDecorator`. It now makes sure the payer has enough in their account to cover the additional fees. It also now deducts/consumes the `floor gas price * gas`.
  * Added the `fee_payer` attribute to events of type `tx` involving fees (i.e. the ones with attributes `fee`, `min_fee_charged`, `additionalfee` and/or `baseFee`).
  * Moved the additional fees calculation logic into the msgfees keeper.
* Update `fee` event with amount charged even on failure and emit SendCoin events from `DeductFeesDistributions` [#1092](https://github.com/provenance-io/provenance/issues/1092).
* Alias the `config unpack` command to `config update`. It can be used to update config files to include new fields [PR 1233](https://github.com/provenance-io/provenance/pull/1233).
* When loading the unpacked configs, always load the defaults before reading the files (instead of only loading the defaults if the file doesn't exist) [PR 1233](https://github.com/provenance-io/provenance/pull/1233).
* Add prune command available though cosmos sdk to provenanced. [#1208](https://github.com/provenance-io/provenance/issues/1208).
* Updated name restrictions documentation [#808](https://github.com/provenance-io/provenance/issues/808).
* Update swagger files [PR 1229](https://github.com/provenance-io/provenance/pull/1229).
* Improve CodeQL workflow to run on Go file changes only [#1225](https://github.com/provenance-io/provenance/issues/1225).
* Use latest ProvWasm contract in wasm tests [#731](https://github.com/provenance-io/provenance/issues/731).
* Publish Java/Kotlin JARs to Maven for release candidates [#1223](https://github.com/provenance-io/provenance/issues/1223).

### Bug Fixes

* Fixed outdated devnet docker configurations [#1062](https://github.com/provenance-io/provenance/issues/1062).
* Fix the [Dragonberry security advisory](https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702) [PR 1173](https://github.com/provenance-io/provenance/pull/1173).
* Fix GetParams in `msgfees` modules to return ConversionFeeDenom [#1214](https://github.com/provenance-io/provenance/issues/1214).
* Pay attention to the `iavl-disable-fastnode` config field/flag [PR 1193](https://github.com/provenance-io/provenance/pull/1193).
* Remove the workaround for the index-events configuration field (now fixed in the SDK) [#995](https://github.com/provenance-io/provenance/issues/995).

### Client Breaking

* Remove the state-listening/plugin system (and `librdkafka` dependencies) [#995](https://github.com/provenance-io/provenance/issues/995).
* Remove the custom/legacy rest endpoints from the `x/attribute`, `x/marker`, and `x/name` modules [#995](https://github.com/provenance-io/provenance/issues/995).
  * The following REST endpoints have been removed in favor of `/provenance/...` counterparts:
    * `GET` `attribute/{address}/attributes` -> `/provenance/attribute/v1/attributes/{address}`
    * `GET` `attribute/{address}/attributes/{name}` -> `/provenance/attribute/v1/attribute/{address}/{name}`
    * `GET` `attribute/{address}/scan/{suffix}` -> `/provenance/attribute/v1/attribute/{address}/scan/{suffix}`
    * `GET` `marker/all` -> `/provenance/marker/v1/all`
    * `GET` `marker/holders/{id}` -> `/provenance/marker/v1/holding/{id}`
    * `GET` `marker/detail/{id}` -> `/provenance/marker/v1/detail/{id}`
    * `GET` `marker/accesscontrol/{id}` -> `/provenance/marker/v1/accesscontrol/{id}`
    * `GET` `marker/escrow/{id}` -> `/provenance/marker/v1/escrow/{id}`
    * `GET` `marker/supply/{id}` -> `/provenance/marker/v1/supply/{id}`
    * `GET` `marker/assets/{id}` -> `/provenance/metadata/v1/ownership/{address}` (you can get the `{address}` from `/provenance/marker/v1/detail/{id}`).
    * `GET` `name/{name}` -> `/provenance/name/v1/resolve/{name}`
    * `GET` `name/{address}/names` -> `/provenance/name/v1/lookup/{address}`
  * The following REST endpoints have been removed. They do not have any REST replacement counterparts. Use GRPC instead.
    * `DELETE` `attribute/attributes` -> `DeleteAttribute(MsgDeleteAttributeRequest)`
    * `POST` `/marker/{denom}/mint` -> `Mint(MsgMintRequest)`
    * `POST` `/marker/{denom}/burn` -> `Burn(MsgBurnRequest)`
    * `POST` `/marker/{denom}/status` -> One of:
      * `Activate(MsgActivateRequest)`
      * `Finalize(MsgFinalizeRequest)`
      * `Cancel(MsgCancelRequest)`
      * `Delete(MsgDeleteRequest)`
  * The following short-form `GET` endpoints were removed in favor of longer ones:
    * `/node_info` -> `/cosmos/base/tendermint/v1beta1/node_info`
    * `/syncing` -> `/cosmos/base/tendermint/v1beta1/syncing`
    * `/blocks/latest` -> `/cosmos/base/tendermint/v1beta1/blocks/latest`
    * `/blocks/{height}` -> `/cosmos/base/tendermint/v1beta1/blocks/{height}`
    * `/validatorsets/latest` -> `/cosmos/base/tendermint/v1beta1/validatorsets/latest`
    * `/validatorsets/{height}` -> `/cosmos/base/tendermint/v1beta1/validatorsets/{height}`
  * The denom owners `GET` endpoint changed from `/cosmos/bank/v1beta1/denom_owners/{denom}` to `/cosmos/bank/v1beta1/supply/by_denom?denom={denom}`.

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.12.2...v1.13.0
