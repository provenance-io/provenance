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

* Update `ScopeSpecification` proto and create `Info` proto #71.

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
