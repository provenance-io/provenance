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

## [v1.8.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.8.0-rc1) - 2022-01-28

### Features

* Add check for `authz` grants when there are missing signatures in `metadata` transactions [#516](https://github.com/provenance-io/provenance/issues/516)
* Add support for publishing Java and Kotlin Protobuf compiled sources to Maven Central [#562](https://github.com/provenance-io/provenance/issues/562)
* Adds support for creating root name governance proposals from the cli [#599](https://github.com/provenance-io/provenance/issues/599)
* Adding of the msg based fee module [#354](https://github.com/provenance-io/provenance/issues/354)
* Upgrade provenance to 0.45 cosmos sdk release [#607](https://github.com/provenance-io/provenance/issues/607)
* Upgrade wasmd to v0.22.0 Note: this removes dependency on provenance-io's wasmd fork [#479](https://github.com/provenance-io/provenance/issues/479)
* Add support for Scope mutation via wasm Smart Contracts [#531](https://github.com/provenance-io/provenance/issues/531)

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
* Limit the maximum attribute value length to 1000 (down from 10,000 currently) in the `green` upgrade [#616](https://github.com/provenance-io/provenance/issues/616)
* Add additional fees for specified operations in the `green` upgrade [#616](https://github.com/provenance-io/provenance/issues/616)
  * `provenance.name.v1.MsgBindNameRequest` 10 hash (10,000,000,000 nhash)
  * `provenance.marker.v1.MsgAddMarkerRequest` 100 hash (100,000,000,000 nhash)
  * `provenance.attribute.v1.MsgAddAttributeRequest` 10 hash (10,000,000,000 nhash)
  * `provenance.metadata.v1.MsgWriteScopeRequest`  10 hash (10,000,000,000 nhash)
  * `provenance.metadata.v1.MsgP8eMemorializeContractRequest` 10 hash (10,000,000,000 nhash)

### Bug Fixes

* When deleting a scope, require the same permissions as when updating it [#473](https://github.com/provenance-io/provenance/issues/473)
* Allow manager to adjust grants on finalized markers [#545](https://github.com/provenance-io/provenance/issues/545)
* Add migration to re-index the metadata indexes involving addresses [#541](https://github.com/provenance-io/provenance/issues/541)
* Add migration to delete empty sessions [#480](https://github.com/provenance-io/provenance/issues/480)
