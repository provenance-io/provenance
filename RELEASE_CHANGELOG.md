## [v1.15.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.15.0-rc1) - 2023-04-17

### Features

* Add support for tokens restricted marker sends with required attributes [#1256](https://github.com/provenance-io/provenance/issues/1256))
* Allow markers to be configured to allow forced transfers [#1368](https://github.com/provenance-io/provenance/issues/1368).
* Publish Provenance Protobuf API as a NPM module [#1449](https://github.com/provenance-io/provenance/issues/1449).
* Add support for account addresses by attribute name lookup [#1447](https://github.com/provenance-io/provenance/issues/1447).
* Add allow forced transfers support to creating markers from smart contracts [#1458](https://github.com/provenance-io/provenance/issues/1458).
* Metadata party rollup and optional parties [#1438](https://github.com/provenance-io/provenance/issues/1438).
* Repeated roles in a spec require multiple different parties [#1437](https://github.com/provenance-io/provenance/issues/1437).
* The `PROVENANCE` role can only be used by smart contract addresses, and vice versa [#1381](https://github.com/provenance-io/provenance/issues/1381).

### Improvements

* Add the `gci` linter that enforces import group ordering. Create a 'lint-fix' make target [PR 1366](https://github.com/provenance-io/provenance/pull/1366).
* Add gRPC query to get all contract specs and record specs for a scope spec [#677](https://github.com/provenance-io/provenance/issues/677).
* Disable `cleveldb` and `badgerdb` by default [#1411](https://github.com/provenance-io/provenance/issues/1411).
  Official builds still have `cleveldb` support though.
* Expand the `additional_bindings` gRPC tag to use object form to allow for Typescript transpiling [#1405](https://github.com/provenance-io/provenance/issues/1405).
* Add attribute cli command to query account addresses by attribute name [#1451](https://github.com/provenance-io/provenance/issues/1451).
* Add removal of attributes from accounts on name deletion [#1410](https://github.com/provenance-io/provenance/issues/1410).
* Enhance ability of smart contracts to use the metadata module [#1280](https://github.com/provenance-io/provenance/issues/1280).
* Enhance the `AddMarker` endpoint to bypass some validation if issued via governance proposal [#1358](https://github.com/provenance-io/provenance/pull/1358).
  This replaces the old `AddMarkerProposal` governance proposal.

### Deprecated

* The `MsgWriteRecordRequest.parties` field has been deprecated and is ignored. The parties in question are identified by the session [PR 1453](https://github.com/provenance-io/provenance/pull/1453).

### Bug Fixes

* Fix third party Protobuf workflow checks on Provenance release steps [#1339](https://github.com/provenance-io/provenance/issues/1339)
* Fix committer email format in third party Protobuf workflow (for [#1339](https://github.com/provenance-io/provenance/issues/1339)) [PR 1385](https://github.com/provenance-io/provenance/pull/1385)
* Fix `make proto-gen` [PR 1404](https://github.com/provenance-io/provenance/pull/1404).
* Fix wasmd transactions that are run by gov module [#1414](https://github.com/provenance-io/provenance/issues/1414)

### Client Breaking

* Removed the `WriteP8eContractSpec` and `P8eMemorializeContract` endpoints [#1402](https://github.com/provenance-io/provenance/issues/1402).
* Removed the `github.com/provenance-io/provenance/x/metadata/types/p8e` proto package [#1402](https://github.com/provenance-io/provenance/issues/1402).
  Users that generate code from the Provenance protos might need to delete their `p8e/` directory.
* The `write-scope` CLI command now takes in `[owners]` as semicolon-delimited parties (instead of comma-delimited `[owner-addresses]`) [PR 1453](https://github.com/provenance-io/provenance/pull/1453).
* Removed the `AddMarkerProposal` [#1358](https://github.com/provenance-io/provenance/pull/1358).
  It is replaced by putting a `MsgAddMarker` (with the `from_address` of the gov module account), in a `MsgSubmitProposal`.

### API Breaking

* Removed the `WriteP8eContractSpec` and `P8eMemorializeContract` endpoints [#1402](https://github.com/provenance-io/provenance/issues/1402).
* Removed the `AddMarkerProposal` [#1358](https://github.com/provenance-io/provenance/pull/1358).
  It is replaced by putting a `MsgAddMarker` (with the `from_address` of the gov module account), in a `MsgSubmitProposal`.

### State Machine Breaking

* The `AddScopeOwner` endpoint now adds a new owner party even if an owner already exists in the scope with that address [PR 1453](https://github.com/provenance-io/provenance/pull/1453).
  I.e. it no longer updates the role of an existing owner with the same address.

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.14.1...v1.15.0-rc1

