## [v1.13.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.13.0-rc1) - 2022-10-05

### Improvements

* Ignore hardcoded tx gas limit when `consensus_params.block.max_gas` is set to -1 for local nodes
* Bump Cosmos-SDK to v0.46.2-pio-1 (from v0.45.5-pio-1). [#995](https://github.com/provenance-io/provenance/issues/995)
  See https://github.com/provenance-io/cosmos-sdk/blob/v0.46.2-pio-1/RELEASE_NOTES.md for more info.
* Refactor the `x/marker` module's `Holding` query to utilize the `x/bank` module's new `DenomOwners` query. [#995](https://github.com/provenance-io/provenance/issues/995)
  The only real difference between those two queries is that the `Holding` query accepts either a denom or marker address.
* Update the third-party protos and swagger files after the cosmos v0.46 bump. [#1017](https://github.com/provenance-io/provenance/issues/1017)
* Stop using the deprecated Wrap and Wrapf functions in the sdk/types/errors package in favor of those functions off specific errors, or else the cosmossdk.io/errors package. [#1013](https://github.com/provenance-io/provenance/issues/995)
* For newly added reward's module, Voting incentive program, validator votes should count for higher shares, since they vote for all their delegations.
  This feature allows the reward creator to introduce the multiplier to achieve the above.
* Refactored the fee handling [#1006](https://github.com/provenance-io/provenance/issues/1006):
  * Created a `MinGasPricesDecorator` to replace the `MempoolFeeDecorator` that was removed from the SDK. It makes sure the fee is greater than the validators min-gas fee.
  * Refactored the `MsgFeesDecorator` to only make sure there's enough fee provided. It no longer deducts/consumes anything and it no longer checks the payer's account.
  * Refactored the `ProvenanceDeductFeeDecorator`. It now makes sure the payer has enough in their account to cover the additional fees. It also now deducts/consumes the `floor gas price * gas`.
  * Added the `fee_payer` attribute to events of type `tx` involving fees (i.e. the ones with attributes `fee`, `min_fee_charged`, `additionalfee` and/or `baseFee`).
  * Moved the additional fees calculation logic into the msgfees keeper.
* Update `fee` event with amount charged even on failure and emit SendCoin events from `DeductFeesDistributions` [#1092](https://github.com/provenance-io/provenance/issues/1092)
* Bump IBC to `5.0.0-pio-1` (from `v2.3.0`) to add a check for SendEnabled [#1100](https://github.com/provenance-io/provenance/issues/1100)
*  [#1067](https://github.com/provenance-io/provenance/issues/1067) This feature makes it so that you can start the chain with custom denoms for a chain, by passing in the required flags, also makes MsgFee not coupled only to the nhash denom.
   For running the chain locally `make run DENOM=vspn MIN_FLOOR_PRICE=0` and `make clean localnet-start DENOM=vspn MIN_FLOOR_PRICE=0` make targets were also updated.
* Use latest ProvWasm contract in wasm tests [#731](https://github.com/provenance-io/provenance/issues/731)
* Update wasmd to 0.28 with 0.46 sdk version from notional-labs [#1015](https://github.com/provenance-io/provenance/issues/1015)

### Bug Fixes

* Remove the workaround for the index-events configuration field (now fixed in the SDK). [#995](https://github.com/provenance-io/provenance/issues/995)

### Client Breaking

* Remove the custom/legacy rest endpoints from the `x/attribute`, `x/marker`, and `x/name` modules. [#995](https://github.com/provenance-io/provenance/issues/995)
* Remove the state-listening/plugin system (and `librdkafka` dependencies). [#995](https://github.com/provenance-io/provenance/issues/995)
