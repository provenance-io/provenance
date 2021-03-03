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

* Add support for creating, updating, removing, finding, and iterating over `Record`s #54
* Add migration support for v039 account into v040 attributes module #100
* Remove setting default no-op upgrade handlers.
* Add an explicit no-op upgrade handler for release v0.1.5.
* Add support for creating, updating, removing, finding, and iterating over `ContractSpecification`s #57
* Add support for record specification metadata addresses #58
* Enhance build process to release cosmovisor compatible zip and plan #119

## [v0.1.4](https://github.com/provenance-io/provenance/releases/tag/v0.1.4) - 2021-02-24

* Update `ScopeSpecification` proto and create `Description` proto #71
* Update `Scope` proto: change field `owner_address` to `owners` #89
* Add support for migrating Marker Accesslist from v39 to v40 #46.
* Add migration command for previous version of Provenance blockchain #78
* Add support for creating, updating, removing, finding, and iterating over `ScopeSpecification`s #56
* Implemented v39 to v40 migration for name module.
* Add support for github actions to build binary releases on tag #30.

## [v0.1.3](https://github.com/provenance-io/provenance/releases/tag/v0.1.3) - 2021-02-12

* Add support for Scope objects to Metadata module #53
* Denom Metadata config for nhash in testnet #42
* Denom Metadata support for marker module #47
* WASM support for Marker module #28
### Bug Fixes

* Name service allows uuids as segments despite length restrictions #48
* Protogen breaks on marker uint64 equals #38
* Fix for marker module beginblock wiring #34
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
