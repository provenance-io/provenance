<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a message and either
an issue number or pull request number using one of these formats:

* message #<issue-number>

If there is no issue number, you can add a reference to a Pull Request like this:
* message PR<pull-request-number>

The issue numbers and pull request numbers will later be link-ified during the release process
so you do not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->

## Unreleased

### Bug Fixes

* Add new `msgfees` `NhashPerUsdMil`  default param to param space store on upgrade (PR [#875](https://github.com/provenance-io/provenance/issues/875))

---

## [v1.11.1-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.11.1-rc1) - 2022-06-14

### Bug Fixes

* Add `mango` upgrade handler.

---

## [v1.11.0](https://github.com/provenance-io/provenance/releases/tag/v1.11.0) - 2022-06-13

### Features

* Add CONTROLLER, and VALIDATOR PartyTypes for contract execution. [\#824](https://github.com/provenance-io/provenance/pull/824])
* Add FeeGrant allowance support for marker escrow accounts [#406](https://github.com/provenance-io/provenance/issues/406)
* Bump Cosmos-SDK to v0.45.4-pio-1, which contains Cosmos-SDK v0.45.4 and the update to storage of the bank module's SendEnabled information. [PR 850](https://github.com/provenance-io/provenance/pull/850)
* Add `MsgAssessCustomMsgFeeRequest` to add the ability for a smart contract author to charge a custom fee [#831](https://github.com/provenance-io/provenance/issues/831)

### Bug Fixes

* Move buf.build push action to occur after PRs are merged to main branch [#838](https://github.com/provenance-io/provenance/issues/838)
* Update third party proto dependencies [#842](https://github.com/provenance-io/provenance/issues/842)

### Improvements

* Add restricted status info to name module cli queries [#806](https://github.com/provenance-io/provenance/issues/806)
* Store the bank module's SendEnabled flags directly in state instead of as part of Params. This will drastically reduce the costs of sending coins and managing markers. [PR 850](https://github.com/provenance-io/provenance/pull/850)
* Add State Sync readme [#859](https://github.com/provenance-io/provenance/issues/859)

### State Machine Breaking

* Move storage of denomination SendEnabled flags into bank module state (from Params), and update the marker module to correctly manipulate the flags in their new location. [PR 850](https://github.com/provenance-io/provenance/pull/850)

---

## [v1.10.0](https://github.com/provenance-io/provenance/releases/tag/v1.10.0) - 2022-05-11

### Summary

Provenance 1.10.0 includes upgrades to the underlying CosmWasm dependencies and adds functionality to
remove orphaned metadata in the bank module left over after markers have been deleted.

### Improvements

* Update wasmvm dependencies and update Dockerfile for localnet [#818](https://github.com/provenance-io/provenance/issues/818)
* Remove "send enabled" on marker removal and in bulk on 1.10.0 upgrade [#821](https://github.com/provenance-io/provenance/issues/821)

---

## [v1.9.0](https://github.com/provenance-io/provenance/releases/tag/v1.9.0) - 2022-04-25

### Summary

Provenance 1.9.0 brings some minor features and security improvements.

### Features

* Add `add-genesis-msg-fee` command to add msg fees to genesis.json and update Makefile to have pre-defined msg fees [#667](https://github.com/provenance-io/provenance/issues/667)
* Add msgfees summary event to be emitted when there are txs that have fees [#678](https://github.com/provenance-io/provenance/issues/678)
* Adds home subcommand to the cli's config command [#620] (https://github.com/provenance-io/provenance/issues/620)
* Add support for rocksdb and badgerdb [#702](https://github.com/provenance-io/provenance/issues/702)
* Create `dbmigrate` utility for migrating a data folder to use a different db backend [#696](https://github.com/provenance-io/provenance/issues/696)

### Improvements

* When the `start` command encounters an error, it no longer outputs command usage [#670](https://github.com/provenance-io/provenance/issues/670)
* Change max length on marker unresticted denom from 64 to 83 [#719](https://github.com/provenance-io/provenance/issues/719)
* Set prerelease to `true` for release candidates. [#666](https://github.com/provenance-io/provenance/issues/666)
* Allow authz grants to work on scope value owners [#755](https://github.com/provenance-io/provenance/issues/755)
* Bump wasmd to v0.26 (from v0.24). [#799](https://github.com/provenance-io/provenance/pull/799)

---

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

---

## [v1.8.1](https://github.com/provenance-io/provenance/releases/tag/v1.8.1) - 2022-04-13

### Summary

Provenance 1.8.1 includes upgrades to the underlying Cosmos SDK and adds initial support for ADR-038.

This release addresses issues related to IAVL concurrency and Tendermint performance that resulted in occasional panics when under high-load conditions such as replay from quicksync. In particular, nodes which experienced issues with "Value missing for hash" and similar panic conditions should work properly with this release. The underlying Cosmos SDK `0.45.3` release that has been incorporated includes a number of improvements around IAVL locking and performance characteristics.

** NOTE: Although Provenance supports multiple database backends, some issues have been reported when using the `goleveldb` backend. If experiencing issues, using the `cleveldb` backend is preferred **

### Improvements

* Update Provenance to use Cosmos SDK 0.45.3 Release [\#781](https://github.com/provenance-io/provenance/issues/781)
* Plugin architecture for ADR-038 + FileStreamingService plugin [\#10639](https://github.com/cosmos/cosmos-sdk/pull/10639)
* Fix for sporadic error "panic: Value missing for hash" [\#611](https://github.com/provenance-io/provenance/issues/611)

---

## [v1.8.0](https://github.com/provenance-io/provenance/releases/tag/v1.8.0) - 2022-03-17

### Summary

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

## [v1.7.6](https://github.com/provenance-io/provenance/releases/tag/v1.7.6) - 2021-12-15

* Upgrade Rosetta to v0.7.2 [#560](https://github.com/provenance-io/provenance/issues/560)

## [v1.7.5](https://github.com/provenance-io/provenance/releases/tag/v1.7.5) - 2021-10-22

### Improvements

* Update Cosmos SDK to 0.44.3 [PR 536](https://github.com/provenance-io/provenance/pull/536)

## [v1.7.4](https://github.com/provenance-io/provenance/releases/tag/v1.7.4) - 2021-10-12

### Improvements

* Update github actions to always run required tests [#508](https://github.com/provenance-io/provenance/issues/508)
* Update Cosmos SDK to 0.44.2 [PR 527](https://github.com/provenance-io/provenance/pull/527)

## [v1.7.3](https://github.com/provenance-io/provenance/releases/tag/v1.7.3) - 2021-09-30

### Bug Fixes

* Update Cosmos SDK to 0.44.1 with IAVL 0.17 to resolve locking issues in queries.
* Fix logger config being ignored [PR 510](https://github.com/provenance-io/provenance/pull/510)

## [v1.7.2](https://github.com/provenance-io/provenance/releases/tag/v1.7.2) - 2021-09-27

### Bug Fixes

* Fix for non-deterministic upgrades in cosmos sdk [#505](https://github.com/provenance-io/provenance/issues/505)

## [v1.7.1](https://github.com/provenance-io/provenance/releases/tag/v1.7.1) - 2021-09-20

### Improvements

* Ensure marker state transition validation does not panic [#492](https://github.com/provenance-io/provenance/issues/492)
* Refactor Examples for cobra cli commands to have examples [#399](https://github.com/provenance-io/provenance/issues/399)
* Verify go version on `make build` [#483](https://github.com/provenance-io/provenance/issues/483)

### Bug Fixes

* Fix marker permissions migration and add panic on `eigengrau` upgrade [#484](https://github.com/provenance-io/provenance/issues/484)
* Fixed marker with more than uint64 causes panic [#489](https://github.com/provenance-io/provenance/issues/489)
* Fixed issue with rosetta tests timing out occasionally, because the timeout was too short [#500](https://github.com/provenance-io/provenance/issues/500)

## [v1.7.0](https://github.com/provenance-io/provenance/releases/tag/v1.7.0) - 2021-09-03

### Features

* Add a single node docker based development environment [#311](https://github.com/provenance-io/provenance/issues/311)
  * Add make targets `devnet-start` and `devnet-stop`
  * Add `networks/dev/mnemonics` for adding accounts to development environment

### Improvements

* Updated some of the documentation of Metadata type bytes (prefixes) [#474](https://github.com/provenance-io/provenance/issues/474)
* Update the Marker Holding query to fully utilize pagination fields [#400](https://github.com/provenance-io/provenance/issues/400)
* Update the Metadata OSLocatorsByURI query to fully utilize pagination fields [#401](https://github.com/provenance-io/provenance/issues/401)
* Update the Metadata OSAllLocators query to fully utilize pagination fields [#402](https://github.com/provenance-io/provenance/issues/402)
* Validate `marker` before setting it to prevent panics [#491](https://github.com/provenance-io/provenance/issues/491)

### Bug Fixes

* Removed some unneeded code from the persistent record update validation [#471](https://github.com/provenance-io/provenance/issues/471)
* Fixed packed config loading bug [#487](https://github.com/provenance-io/provenance/issues/487)
* Fixed marker with more than uint64 causes panic [#489](https://github.com/provenance-io/provenance/issues/489)

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

## [v1.6.0](https://github.com/provenance-io/provenance/releases/tag/v1.6.0) - 2021-08-23

### Bug Fixes

* Fix for creating non-coin type markers through governance addmarker proposals [#431](https://github.com/provenance-io/provenance/issues/431)
* Upgrade handler migrates usdf.c to the right marker_type.

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

## [v1.4.1](https://github.com/provenance-io/provenance/releases/tag/v1.4.1) - 2021-06-02

* Updated github binary release workflow.  No code changes from 1.4.0.

## [v1.4.0](https://github.com/provenance-io/provenance/releases/tag/v1.4.0) - 2021-06-02

### Features

* ENV config support, SDK v0.42.5 update [#320](https://github.com/provenance-io/provenance/issues/320)
* Upgrade handler set version name to `citrine` [#339](https://github.com/provenance-io/provenance/issues/339)

### Bug Fixes

* P8EMemorializeContract: preserve some Scope fields if the scope already exists [PR 336](https://github.com/provenance-io/provenance/pull/336)
* Set default standard err/out for `provenanced` commands [PR 337](https://github.com/provenance-io/provenance/pull/337)
* Fix for invalid help text permissions list on marker access grant command [PR 337](https://github.com/provenance-io/provenance/pull/337)
* When writing a session, make sure the scope spec of the containing scope, contains the session's contract spec. [#322](https://github.com/provenance-io/provenance/issues/322)

###  Improvements

* Informative error message for `min-gas-prices` invalid config panic on startup [#333](https://github.com/provenance-io/provenance/issues/333)
* Update marker event documentation to match typed event namespaces [#304](https://github.com/provenance-io/provenance/issues/304)


## [v1.3.1](https://github.com/provenance-io/provenance/releases/tag/v1.3.1) - 2021-05-21

### Bug Fixes

* Remove broken gauge on attribute module. Fixes prometheus metrics [#315](https://github.com/provenance-io/provenance/issues/315)
* Correct logging levels for marker mint/burn requests [#318](https://github.com/provenance-io/provenance/issues/318)
* Fix the CLI metaaddress commands [#321](https://github.com/provenance-io/provenance/issues/321)

### Improvements

* Add Kotlin and Javascript examples for Metadata Addresses [#301](https://github.com/provenance-io/provenance/issues/301)
* Updated swagger docs [PR 313](https://github.com/provenance-io/provenance/pull/313)
* Fix swagger docs [PR 317](https://github.com/provenance-io/provenance/pull/317)
* Updated default min-gas-prices to reflect provenance network nhash economics [#310](https://github.com/provenance-io/provenance/pull/310)
* Improved marker error message when marker is not found [#325](https://github.com/provenance-io/provenance/issues/325)


## [v1.3.0](https://github.com/provenance-io/provenance/releases/tag/v1.3.0) - 2021-05-06

### Features

* Add grpc messages and cli command to add/remove addresses from metadata scope data access [#220](https://github.com/provenance-io/provenance/issues/220)
* Add a `context` field to the `Session` [#276](https://github.com/provenance-io/provenance/issues/276)
* Add typed events and telemetry metrics to attribute module [#86](https://github.com/provenance-io/provenance/issues/86)
* Add rpc and cli support for adding/updating/removing owners on a `Scope` [#283](https://github.com/provenance-io/provenance/issues/283)
* Add transaction and query time measurements to marker module [#284](https://github.com/provenance-io/provenance/issues/284)
* Upgrade handler included that sets denom metadata for `hash` bond denom [#294](https://github.com/provenance-io/provenance/issues/294)
* Upgrade wasmd to v0.16.0 [#291](https://github.com/provenance-io/provenance/issues/291)
* Add params query endpoint to the marker module cli [#271](https://github.com/provenance-io/provenance/issues/271)

### Improvements

* Added linkify script for changelog issue links [#107](https://github.com/provenance-io/provenance/issues/107)
* Changed Metadata events to be typed events [#88](https://github.com/provenance-io/provenance/issues/88)
* Updated marker module spec documentation [#93](https://github.com/provenance-io/provenance/issues/93)
* Gas consumption telemetry and tracing [#299](https://github.com/provenance-io/provenance/issues/299)

### Bug Fixes

* More mapping fixes related to `WriteP8EContractSpec` and `P8EMemorializeContract` [#275](https://github.com/provenance-io/provenance/issues/275)
* Fix event manager scope in attribute, name, marker, and metadata modules to prevent event duplication [#289](https://github.com/provenance-io/provenance/issues/289)
* Proposed markers that are cancelled can be deleted without ADMIN role being assigned [#280](https://github.com/provenance-io/provenance/issues/280)
* Fix to ensure markers have no balances in Escrow prior to being deleted. [#303](https://github.com/provenance-io/provenance/issues/303)

### State Machine Breaking

* Add support for purging destroyed markers [#282](https://github.com/provenance-io/provenance/issues/282)

## [v1.2.0](https://github.com/provenance-io/provenance/releases/tag/v1.2.0) - 2021-04-26

### Improvements

* Add spec documentation for the metadata module [#224](https://github.com/provenance-io/provenance/issues/224)

### Features

* Add typed events and telemetry metrics to marker module [#247](https://github.com/provenance-io/provenance/issues/247)

### Bug Fixes

* Wired recovery flag into `init` command [#254](https://github.com/provenance-io/provenance/issues/254)
* Always anchor unrestricted denom validation expressions, Do not allow slashes in marker denom expressions [#258](https://github.com/provenance-io/provenance/issues/258)
* Mapping and validation fixes found while trying to use `P8EMemorializeContract` [#256](https://github.com/provenance-io/provenance/issues/256)

### Client Breaking

* Update marker transfer request signing behavior [#246](https://github.com/provenance-io/provenance/issues/246)


## [v1.1.1](https://github.com/provenance-io/provenance/releases/tag/v1.1.1) - 2021-04-15

### Bug Fixes

* Add upgrade plan v1.1.1

## [v1.1.0](https://github.com/provenance-io/provenance/releases/tag/v1.1.0) - 2021-04-15

### Features

* Add marker cli has two new flags to set SupplyFixed and AllowGovernanceControl [#241](https://github.com/provenance-io/provenance/issues/241)
* Modify 'enable governance' behavior on marker module [#227](https://github.com/provenance-io/provenance/issues/227)
* Typed Events and Metric counters in Name Module [#85](https://github.com/provenance-io/provenance/issues/85)

### Improvements

* Add some extra aliases for the CLI query metadata commands.
* Make p8e contract spec id easier to communicate.

### Bug Fixes

* Add pagination flags to the CLI query metadata commands.
* Fix handling of Metadata Write message id helper fields.
* Fix cli metadata address encoding/decoding command tree [#231](https://github.com/provenance-io/provenance/issues/231)
* Metadata Module parsing of base64 public key fixed [#225](https://github.com/provenance-io/provenance/issues/225)
* Fix some conversion pieces in `P8EMemorializeContract`.
* Remove extra Object Store Locator storage.
* Fix input status mapping.
* Add MsgSetDenomMetadataRequest to the marker handler.

## [v1.0.0](https://github.com/provenance-io/provenance/releases/tag/v1.0.0) - 2021-03-31

### Bug Fixes

* Resolves an issue where Gov Proposals to Create a new name would fail for new root domains [#192](https://github.com/provenance-io/provenance/issues/192)
* Remove deprecated ModuleCdc amino encoding from Metadata Locator records [#187](https://github.com/provenance-io/provenance/issues/187)
* Update Cosmos SDK to 0.42.3
* Remove deprecated ModuleCdc amino encoding from name module [#189](https://github.com/provenance-io/provenance/issues/189)
* Remove deprecated ModuleCdc amino encoding from attribute module [#188](https://github.com/provenance-io/provenance/issues/188)

### Features

* Allow withdrawals of any coin type from a marker account in WASM smart contracts. [#151](https://github.com/provenance-io/provenance/issues/151)
* Added cli tx commands `write-contract-specification` `remove-contract-specification` for updating/adding/removing metadata `ContractSpecification`s. [#195](https://github.com/provenance-io/provenance/issues/195)
* Added cli tx commands `write-record-specification` `remove-record-specification` for updating/adding/removing metadata `RecordSpecification`s. [#176](https://github.com/provenance-io/provenance/issues/176)
* Added cli tx commands `write-scope-specification` `remove-scope-specification` for updating/adding/removing metadata `ScopeSpecification`s. [#202](https://github.com/provenance-io/provenance/issues/202)
* Added cli tx commands `write-scope` `remove-scope` for updating/adding/removing metadata `Scope`s. [#199](https://github.com/provenance-io/provenance/issues/199)
* Added cli tx commands `write-record` `remove-record` for updating/adding/removing metadata `Record`s. [#205](https://github.com/provenance-io/provenance/issues/205)
* Simulation testing support [#95](https://github.com/provenance-io/provenance/issues/95)
* Name module simulation testing [#24](https://github.com/provenance-io/provenance/issues/24)
* Added default IBC parameters for v039 chain genesis migration script [#102](https://github.com/provenance-io/provenance/issues/102)
* Expand and simplify querying. [#169](https://github.com/provenance-io/provenance/issues/169)
  * Added endpoints for getting all entries of a type, e.g. `RecordsAll`.
  * Combined some endpoints (see notesin "API Breaking" section).
  * Allow searching for related entries. E.g. you can provide a record id to the scope search.
  * Add ability to return related entries. E.g. the `Sessions` endpoint has a `include_records` flag that will cause the response to contain the records that are part of the sessions.
* Add optional identification fields in tx `Write...` messages. [#169](https://github.com/provenance-io/provenance/issues/169)
* The `Write` endpoints now return information about the written entries. [#169](https://github.com/provenance-io/provenance/issues/169)
* Added a CLI command for getting all entries of a type, `query metadata all <type>`, or `query metadata <type> all`. [#169](https://github.com/provenance-io/provenance/issues/169)
* Restrict denom metadata. [#208](https://github.com/provenance-io/provenance/issues/208)

### API Breaking

* Change `Add...` metadata tx endpoints to `Write...` (e.g. `AddScope` is now `WriteScope`). [#169](https://github.com/provenance-io/provenance/issues/169)
* Expand and simplify metadata querying. [#169](https://github.com/provenance-io/provenance/issues/169)
  * Removed the `SessionContextByID` and `SessionContextByUUID` endponts. Replaced with the `Sessions` endpoint.
  * Removed the `RecordsByScopeID` and `RecordsByScopeUUID` endpoints. Replaced with the `Records` endpoint.
  * Removed the `ContractSpecificationExtended` endpoint. Use `ContractSpecification` now with the `include_record_specs` flag.
  * Removed the `RecordSpecificationByID` endpoint. Use the `RecordSpecification` endpoint.
  * Change the `_uuid` fields in the queries to `_id` to allow for either address or uuid input.
  * The `Scope` query no longer returns `Sessions` and `Records` by default. Use the `include_sessions` and `include_records` if you want them.
  * Query result entries are now wrapped to include extra id information alongside an entry.
    E.g. Where a `Scope` used to be returned, now a `ScopeWrapper` is returned containing a `Scope` and its `ScopeIdInfo`.
    So where you previously had `resp.Scope` you will now want `resp.Scope.Scope`.
  * Pluralized both the message name and field name of locator queries that return multiple entries.
    * `OSLocatorByScopeUUIDRequest` and `OSLocatorByScopeUUIDResponse` changed to `OSLocatorsByScopeUUIDRequest` and `OSLocatorsByScopeUUIDResponse`.
    * `OSLocatorByURIRequest` and `OSLocatorByURIResponse` changed to `OSLocatorsByURIRequest` and `OSLocatorsByURIResponse`.
    * Field name `locator` changed to `locators` in `OSLocatorsByURIResponse`, `OSLocatorsByScopeUUIDResponse`, `OSAllLocatorsResponse`.

### Client Breaking

* The paths for querying metadata have changed. See API Breaking section for an overview, and the proto file for details. [#169](https://github.com/provenance-io/provenance/issues/169)
* The CLI has been updated for metadata querying. [#169](https://github.com/provenance-io/provenance/issues/169)
  * Removed the `fullscope` command. Use `query metadata scope --include-sessions --include-records` now.
  * Combined the `locator-by-addr`, `locator-by-uri`, `locator-by-scope`, and `locator-all` into a single `locator` command.
* Changed the CLI metadata tx `add-...` commands to `write-...`. [#166](https://github.com/provenance-io/provenance/issues/166)

## [v0.3.0](https://github.com/provenance-io/provenance/releases/tag/v0.3.0) - 2021-03-19

### Features

* Governance proposal support for marker module
* Decentralized discovery for object store instances [#105](https://github.com/provenance-io/provenance/issues/105)
* Add `AddP8eContractSpec` endpoint to convert v39 contract spec into v40 contract specification  [#167](https://github.com/provenance-io/provenance/issues/167)
* Refactor `Attribute` validate to sdk standard validate basic and validate size of attribute value [#175](https://github.com/provenance-io/provenance/issues/175)
* Add the temporary `P8eMemorializeContract` endpoint to help facilitate the transition. [#164](https://github.com/provenance-io/provenance/issues/164)
* Add handler for 0.3.0 testnet upgrade.

### Bug Fixes

* Gov module route added for name module root name proposal
* Update Cosmos SDK to 0.42.2 for bug fixes and improvements


## [v0.2.1](https://github.com/provenance-io/provenance/releases/tag/v0.2.1) - 2021-03-11

* Update to Cosmos SDK 0.42.1
* Add github action for docker publishing [#156](https://github.com/provenance-io/provenance/issues/156)
* Add `MetaAddress` encoder and parser commands [#147](https://github.com/provenance-io/provenance/issues/147)
* Add build support for publishing protos used in this release [#69](https://github.com/provenance-io/provenance/issues/69)
* Support for setting a marker denom validation expression [#84](https://github.com/provenance-io/provenance/issues/84)
* Expand cli metadata query functionality [#142](https://github.com/provenance-io/provenance/issues/142)

## [v0.2.0](https://github.com/provenance-io/provenance/releases/tag/v0.2.0) - 2021-03-05

* Truncate hashes used in metadata addresses for Record, Record Specification [#132](https://github.com/provenance-io/provenance/issues/132)
* Add support for creating, updating, removing, finding, and iterating over `Session`s [#55](https://github.com/provenance-io/provenance/issues/55)
* Add support for creating, updating, removing, finding, and iterating over `RecordSpecification`s [#59](https://github.com/provenance-io/provenance/issues/59)

## [v0.1.10](https://github.com/provenance-io/provenance/releases/tag/v0.1.10) - 2021-03-04

### Bug fixes

* Ensure all upgrade handlers apply always before storeLoader is created.
* Add upgrade handler for v0.1.10

## [v0.1.9](https://github.com/provenance-io/provenance/releases/tag/v0.1.9) - 2021-03-03

### Bug fixes

* Add module for metadata for v0.1.9

## [v0.1.8](https://github.com/provenance-io/provenance/releases/tag/v0.1.8) - 2021-03-03

### Bug fixes

* Add handlers for v0.1.7, v0.1.8

## [v0.1.7](https://github.com/provenance-io/provenance/releases/tag/v0.1.7) - 2021-03-03

### Bug Fixes

* Fix npe caused by always loading custom storeLoader.

## [v0.1.6](https://github.com/provenance-io/provenance/releases/tag/v0.1.6) - 2021-03-02

### Bug Fixes

* Add metadata module to the IAVL store during upgrade

## [v0.1.5](https://github.com/provenance-io/provenance/releases/tag/v0.1.5) - 2021-03-02

* Add support for creating, updating, removing, finding, and iterating over `Record`s [#54](https://github.com/provenance-io/provenance/issues/54)
* Add migration support for v039 account into v040 attributes module [#100](https://github.com/provenance-io/provenance/issues/100)
* Remove setting default no-op upgrade handlers.
* Add an explicit no-op upgrade handler for release v0.1.5.
* Add support for creating, updating, removing, finding, and iterating over `ContractSpecification`s [#57](https://github.com/provenance-io/provenance/issues/57)
* Add support for record specification metadata addresses [#58](https://github.com/provenance-io/provenance/issues/58)
* Enhance build process to release cosmovisor compatible zip and plan [#119](https://github.com/provenance-io/provenance/issues/119)

## [v0.1.4](https://github.com/provenance-io/provenance/releases/tag/v0.1.4) - 2021-02-24

* Update `ScopeSpecification` proto and create `Description` proto [#71](https://github.com/provenance-io/provenance/issues/71)
* Update `Scope` proto: change field `owner_address` to `owners` [#89](https://github.com/provenance-io/provenance/issues/89)
* Add support for migrating Marker Accesslist from v39 to v40 [#46](https://github.com/provenance-io/provenance/issues/46).
* Add migration command for previous version of Provenance blockchain [#78](https://github.com/provenance-io/provenance/issues/78)
* Add support for creating, updating, removing, finding, and iterating over `ScopeSpecification`s [#56](https://github.com/provenance-io/provenance/issues/56)
* Implemented v39 to v40 migration for name module.
* Add support for github actions to build binary releases on tag [#30](https://github.com/provenance-io/provenance/issues/30).

## [v0.1.3](https://github.com/provenance-io/provenance/releases/tag/v0.1.3) - 2021-02-12

* Add support for Scope objects to Metadata module [#53](https://github.com/provenance-io/provenance/issues/53)
* Denom Metadata config for nhash in testnet [#42](https://github.com/provenance-io/provenance/issues/42)
* Denom Metadata support for marker module [#47](https://github.com/provenance-io/provenance/issues/47)
* WASM support for Marker module [#28](https://github.com/provenance-io/provenance/issues/28)

### Bug Fixes

* Name service allows uuids as segments despite length restrictions [#48](https://github.com/provenance-io/provenance/issues/48)
* Protogen breaks on marker uint64 equals [#38](https://github.com/provenance-io/provenance/issues/38)
* Fix for marker module beginblock wiring [#34](https://github.com/provenance-io/provenance/issues/34)
* Fix for marker get cli command
* Updated the links in PULL_REQUEST_TEMPLATE.md to use correct 'main' branch

## [v0.1.2](https://github.com/provenance-io/provenance/releases/tag/v0.1.2) - 2021-01-27

### Bug Fixes

* Update goreleaser configuration to match `provenance` repository name

## [v0.1.1](https://github.com/provenance-io/provenance/releases/tag/v0.1.1) - 2021-01-27

This is the intial beta release for the first Provenance public TESTNET.  This release is not intended for any type of
production or reliable development as extensive work is still in progress to migrate the private network functionality
into the public network.

### Features

* Initial port of private Provenance blockchain modules `name`, `attribute`, and `marker` from v0.39.x Cosmos SDK chain
into new 0.40.x base.  Minimal unit test coverage and features in place to begin setup of testnet process.

## PRE-HISTORY

## [v0.1.0](https://github.com/provenance-io/provenance/releases/tag/v0.1.0) - 2021-01-26

* Test tag prior to initial testnet release.

The Provenance Blockchain was started by Figure Technologies in 2018 using a Hyperledger Fabric derived private network.
A subsequent migration was made to a new internal private network based on the 0.38-0.39 series of Cosmos SDK and
Tendermint.  The Provence-IO/Provenance Cosmos SDK derived public network is the
