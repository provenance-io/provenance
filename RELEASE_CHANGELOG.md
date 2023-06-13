## [v1.16.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.16.0-rc1) - 2023-06-13

Important: When building this version of `provenanced`, you **MUST** use go v1.20.

Version `1.16.0` of the Provenance Blockchain brings several features, improvements, and bug fixes.

### Features

* Add support to add/remove required attributes for a restricted marker. [#1512](https://github.com/provenance-io/provenance/issues/1512)
* Add trigger module for delayed execution. [#1462](https://github.com/provenance-io/provenance/issues/1462)
* Add support to update the `allow_forced_transfer` field of a restricted marker [#1545](https://github.com/provenance-io/provenance/issues/1545).
* Add expiration date value to `attribute` [#1435](https://github.com/provenance-io/provenance/issues/1435).
* Add endpoints to update the value owner address of scopes [#1329](https://github.com/provenance-io/provenance/issues/1329).

### Improvements

* Bump go to `1.20` (from `1.18`) [#1539](https://github.com/provenance-io/provenance/issues/1539).
* Bump golangci-lint to `v1.52.2` (from `v1.48`) [#1539](https://github.com/provenance-io/provenance/issues/1539).
  * New `make golangci-lint` target created for installing golangci-lint.
  * New `make golangci-lint-update` target created for installing the current version even if you already have a version installed.
* Add marker deposit access check for sends to marker escrow account [#1525](https://github.com/provenance-io/provenance/issues/1525).
* Add support for `name` owner to execute `MsgModifyName` transaction [#1536](https://github.com/provenance-io/provenance/issues/1536).
* Add usage of `AddGovPropFlagsToCmd` and `ReadGovPropFlags` cli for `GetModifyNameCmd` [#1542](https://github.com/provenance-io/provenance/issues/1542).
* Bump Cosmos-SDK to `v0.46.10-pio-4` (from `v0.46.10-pio-3`) for the `SendRestrictionFn` changes [PR 1506](https://github.com/provenance-io/provenance/pull/1506).
* Switch to using a `SendRestrictionFn` for restricting sends of marker funds [PR 1506](https://github.com/provenance-io/provenance/pull/1506).
* Create `rust` upgrade handlers [PR 1549](https://github.com/provenance-io/provenance/pull/1549).
* Remove mutation of store from `attribute` keeper iterators [#1557](https://github.com/provenance-io/provenance/issues/1557).
* Bumped ibc-go to 6.1.1 [#1563](https://github.com/provenance-io/provenance/pull/1563).
* Update `marker` module spec documentation with new proto references [#1580](https://github.com/provenance-io/provenance/pull/1580).
* Bumped `wasmd` to v0.30.0-pio-5 and `wasmvm` to v1.2.4 [#1582](https://github.com/provenance-io/provenance/pull/1582).
* Inactive validator delegation cleanup process [#1556](https://github.com/provenance-io/provenance/issues/1556).
* Bump Cosmos-SDK to [v0.46.13-pio-1](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.13-pio-1/RELEASE_NOTES.md) (from `v0.46.10-pio-4`) [PR 1585](https://github.com/provenance-io/provenance/pull/1585).

### Bug Fixes

* Bring back some proto messages that were deleted but still needed for historical queries [#1554](https://github.com/provenance-io/provenance/issues/1554).
* Fix the `MsgModifyNameRequest` endpoint to properly clean up old index data [PR 1565](https://github.com/provenance-io/provenance/pull/1565).

### API Breaking

* Add marker deposit access check for sends to marker escrow account.  Will break any current address that is sending to the
marker escrow account if it does not have deposit access.  In order for it to work, deposit access needs to be added.  This can be done using the `MsgAddAccessRequest` tx  [#1525](https://github.com/provenance-io/provenance/issues/1525).
* `MsgMultiSend` is now limited to a single `Input` [PR 1506](https://github.com/provenance-io/provenance/pull/1506).
* SDK errors returned from Metadata module endpoints [#978](https://github.com/provenance-io/provenance/issues/978).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.15.2...v1.16.0-rc1

