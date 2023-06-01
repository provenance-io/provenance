## [v1.15.1](https://github.com/provenance-io/provenance/releases/tag/v1.15.1) - 2023-06-01

Provenance blockchain version `v1.15.1` is a state-compatible upgrade to `v1.15.0`. Users are encouraged to upgrade to `v1.15.1` at their earliest convenience.

This version updates the IBC module to address an unlikely (but possible) security issue. It also returns some protobuf messages so that the governance proposals query can work again.

### Improvements

* Bumped ibc-go to 6.1.1 [PR 1563](https://github.com/provenance-io/provenance/pull/1563).

### Bug Fixes

* Bring back some proto messages that were deleted but still needed for historical queries [#1554](https://github.com/provenance-io/provenance/issues/1554).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.15.0...v1.15.1

---

## [v1.15.0](https://github.com/provenance-io/provenance/releases/tag/v1.15.0) - 2023-05-05

### Features

* Add support for tokens restricted marker sends with required attributes [#1256](https://github.com/provenance-io/provenance/issues/1256)
* Allow markers to be configured to allow forced transfers [#1368](https://github.com/provenance-io/provenance/issues/1368).
* Publish Provenance Protobuf API as a NPM module [#1449](https://github.com/provenance-io/provenance/issues/1449).
