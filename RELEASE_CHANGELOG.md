## [v1.8.0](https://github.com/provenance-io/provenance/releases/tag/v1.8.0) - 2022-03-17

## Summary
Provenance 1.8.0 is focused on improving the fee structures for transactions on the blockchain. While the Cosmos SDK has traditionally offered a generic fee structure focused on gas/resource utilization, the Provenance blockchain has found that certain transactions have additional long term costs and value beyond simple resources charges. This is the reason we are adding the new MsgFee module which allows governance based control of additional fee charges on certain message types.

NOTE: The second major change in the 1.8.0 release is part of the migration process which removes many orphaned state objects that were left in 1.7.x chains. This cleanup process will require a significant amount of time to perform during the green upgrade handler execution. The upgrade will print status messages showing the progress of this process.

### Features

* Add check for `authz` grants when there are missing signatures in `metadata` transactions [#516](https://github.com/provenance-io/provenance/issues/516)
* Add support for publishing Java and Kotlin Protobuf compiled sources to Maven Central [#562](https://github.com/provenance-io/provenance/issues/562)
* Adds support for creating root name governance proposals from the cli [#599](https://github.com/provenance-io/provenance/issues/599)
* Adding of the msg based fee module [#354](https://github.com/provenance-io/provenance/issues/354)
* Upgrade provenance to 0.45 cosmos sdk release [#607](https://github.com/provenance-io/provenance/issues/607)
* Upgrade wasmd to v0.22.0 Note: this removes dependency on provenance-io's wasmd fork [#479](https://github.com/provenance-io/provenance/issues/479)
* Add support for Scope mutation via wasm Smart Contracts [#531](https://github.com/provenance-io/provenance/issues/531)
* Increase governance deposit amount and add create proposal msg fee [#632](https://github.com/provenance-io/provenance/issues/632)
* Allow attributes to be associated with scopes [#631](https://github.com/provenance-io/provenance/issues/631)

### Improvements

* Add `bank` and `authz` module query `proto` files required by `grpcurl` [#482](https://github.com/provenance-io/provenance/issues/482)
* Fix typeos in marker log statements [#502](https://github.com/provenance-io/provenance/issues/502)
* Set default coin type to network default [#534](https://github.com/provenance-io/provenance/issues/534)
* Add logger to upgrade handler [#507](https://github.com/provenance-io/provenance/issues/507)
* Allow markers to be created over existing accounts if they are not a marker and have a zero sequence [#520](https://github.com/provenance-io/provenance/issues/520)
* Removed extraneous Metadata index deletes/rewrites [#543](https://github.com/provenance-io/provenance/issues/543)
* Delete empty sessions when the last record is updated to a new session [#480](https://github.com/provenance-io/provenance/issues/480)
* Refactor the migration to be faster and have more log output [PR 586](https://github.com/provenance-io/provenance/pull/586)
* Capture all included protobufs into release zip file [#556](https://github.com/provenance-io/provenance/issues/556)
* Add Protobuf support with buf.build [#614](https://github.com/provenance-io/provenance/issues/614)
* Limit the maximum attribute value length to 1000 (down from 10,000 currently) in the `green` upgrade [#616](https://github.com/provenance-io/provenance/issues/616)
* Add additional fees for specified operations in the `green` upgrade [#616](https://github.com/provenance-io/provenance/issues/616)
  * `provenance.name.v1.MsgBindNameRequest` 10 hash (10,000,000,000 nhash)
  * `provenance.marker.v1.MsgAddMarkerRequest` 100 hash (100,000,000,000 nhash)
  * `provenance.attribute.v1.MsgAddAttributeRequest` 10 hash (10,000,000,000 nhash)
  * `provenance.metadata.v1.MsgWriteScopeRequest`  10 hash (10,000,000,000 nhash)
  * `provenance.metadata.v1.MsgP8eMemorializeContractRequest` 10 hash (10,000,000,000 nhash)
* Add integration tests for smart contracts [#392](https://github.com/provenance-io/provenance/issues/392)  
* Use provwasm release artifact for smart contract tests [#731](https://github.com/provenance-io/provenance/issues/731)

### Client Breaking

* Enforce a maximum gas limit on individual transactions so that at least 20 can fit in any given block. [#681](https://github.com/provenance-io/provenance/issues/681)
  Previously transactions were only limited by their size in bytes as well as the overall gas limit on a given block.  

  _With this update transactions must be no more than 5% of the maximum amount of gas allowed per block when a gas limit
  per block is set (this restriction has no effect when a gas limit has not been set).  The current limits on Provenance
  mainnet are 60,000,000 gas per block which will yield a maximum transaction size of 3,000,000 gas using this new AnteHandler
  restriction._

### Bug Fixes

* When deleting a scope, require the same permissions as when updating it [#473](https://github.com/provenance-io/provenance/issues/473)
* Allow manager to adjust grants on finalized markers [#545](https://github.com/provenance-io/provenance/issues/545)
* Add migration to re-index the metadata indexes involving addresses [#541](https://github.com/provenance-io/provenance/issues/541)
* Add migration to delete empty sessions [#480](https://github.com/provenance-io/provenance/issues/480)
* Add Java distribution tag to workflow [#624](https://github.com/provenance-io/provenance/issues/624)
* Add `msgfees` module to added store upgrades [#640](https://github.com/provenance-io/provenance/issues/640)
* Use `nhash` for base denom in gov proposal upgrade [#648](https://github.com/provenance-io/provenance/issues/648)
* Bump `cosmowasm` from `v1.0.0-beta5` to `v1.0.0-beta6` [#655](https://github.com/provenance-io/provenance/issues/655)
* Fix maven publish release version number reference [#650](https://github.com/provenance-io/provenance/issues/650)
* Add `iterator` as feature for wasm [#658](https://github.com/provenance-io/provenance/issues/658)
* String "v" from Jar artifact version number [#653](https://github.com/provenance-io/provenance/issues/653)
* Fix `wasm` contract migration failure to find contract history [#662](https://github.com/provenance-io/provenance/issues/662)
