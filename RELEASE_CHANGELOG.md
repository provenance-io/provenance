## [v1.14.1](https://github.com/provenance-io/provenance/releases/tag/v1.14.1) - 2023-02-28

The Provenance Blockchain `v1.14.1` release fixes a couple bugs.
Notably, state listening now behaves as expected with `stop-node-on-err = true`.

This release is state compatible with the `v1.14.0` release.
Users should upgrade from `v1.14.0` at their convenience.

### Improvements

* Bump Cosmos-SDK to `v0.46.10-pio-2` (from `v0.46.10-pio-1`). [PR 1396](https://github.com/provenance-io/provenance/pull/1396). \
  See the following `RELEASE_NOTES.md` for details: \
  [v0.46.10-pio-2](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.10-pio-2/RELEASE_NOTES.md). \
  Full Commit History: https://github.com/provenance-io/cosmos-sdk/compare/v0.46.10-pio-1...v0.46.10-pio-2

### Bug Fixes

* Fix `start` using default home directory [PR 1393](https://github.com/provenance-io/provenance/pull/1393).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.14.0...v1.14.1

