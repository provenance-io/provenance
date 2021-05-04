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
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

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

### Features

* Add grpc messages and cli command to add/remove addresses from metadata scope data access [#220](https://github.com/provenance-io/provenance/issues/220)
* Add a `context` field to the `Session` [#276](https://github.com/provenance-io/provenance/issues/276)
* Add typed events and telemetry metrics to attribute module [#86](https://github.com/provenance-io/provenance/issues/86)
* Add transaction and query time measurements to marker module [#284](https://github.com/provenance-io/provenance/issues/284)

### Improvements

* Added linkify script for changelog issue links [#107](https://github.com/provenance-io/provenance/issues/107)

### Bug Fixes

* More mapping fixes related to `WriteP8EContractSpec` and `P8EMemorializeContract` [#275](https://github.com/provenance-io/provenance/issues/275)
* Fix event manager scope in attribute, name, marker, and metadata modules to prevent event duplication [#289](https://github.com/provenance-io/provenance/issues/289)

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
