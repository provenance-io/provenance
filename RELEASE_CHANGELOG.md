## [v1.20.1](https://github.com/provenance-io/provenance/releases/tag/v1.20.1) 2024-11-07

Provenance Blockchain version `v1.20.1` contains several improvements and bug fixes. Validators will switch to it with the `viridian` upgrade, which is tentatively scheduled for 2024-11-14 5:15 PM Eastern time.

This version fixes a [security vulnerability](https://github.com/cometbft/cometbft/security/advisories/GHSA-p7mv-53f2-4cwj) in the cometbft library.

Building or installing `provenanced` from source now requires you to use [Go 1.23](https://golang.org/dl/).
Linting now requires `golangci-lint` v1.60.2. You can update yours using `make golangci-lint-update` or install it using `make golangci-lint`.

### Dependencies

* `bufbuild/buf-setup-action` bumped to 1.46.0 (from 1.45.0) [PR 2206](https://github.com/provenance-io/provenance/pull/2206).
* `github.com/btcsuite/btcd/btcec/v2` removed at v2.3.4 [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/cespare/xxhash` removed at v1.1.0 [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/cometbft/cometbft-db` bumped to v0.14.1 (from v0.11.0) [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/cometbft/cometbft` bumped to v0.38.15 (from v0.38.12) ([PR 2202](https://github.com/provenance-io/provenance/pull/2202), [PR 2209](https://github.com/provenance-io/provenance/pull/2209)).
* `github.com/decred/dcrd/dcrec/secp256k1/v4` bumped to v4.3.0 (from v4.2.0) [PR 2209](https://github.com/provenance-io/provenance/pull/2209).
* `github.com/dgraph-io/badger/v2` removed at v2.2007.4 [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/dgraph-io/badger/v4` added at v4.2.0 [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/dgryski/go-farm` removed at v0.0.0-20200201041132-a6ae2369ad13 [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/google/btree` bumped to v1.1.3 (from v1.1.2) [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/google/flatbuffers` added at v1.12.1 [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/lib/pq` bumped to v1.10.9 (from v1.10.7) [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/minio/highwayhash` bumped to v1.0.3 (from v1.0.2) [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/petermattis/goid` bumped to v0.0.0-20240813172612-4fcff4a6cae7 (from v0.0.0-20231207134359-e60b3f734c67) [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `github.com/prometheus/client_golang` bumped to v1.20.5 (from v1.20.1) ([PR 2202](https://github.com/provenance-io/provenance/pull/2202), [PR 2209](https://github.com/provenance-io/provenance/pull/2209)).
* `github.com/prometheus/common` bumped to v0.60.1 (from v0.55.0) ([PR 2202](https://github.com/provenance-io/provenance/pull/2202), [PR 2209](https://github.com/provenance-io/provenance/pull/2209)).
* `github.com/sasha-s/go-deadlock` bumped to v0.3.5 (from v0.3.1) [PR 2202](https://github.com/provenance-io/provenance/pull/2202).
* `golang.org/x/crypto` bumped to v0.28.0 (from v0.26.0) ([PR 2202](https://github.com/provenance-io/provenance/pull/2202), [PR 2209](https://github.com/provenance-io/provenance/pull/2209)).
* `golang.org/x/net` bumped to v0.30.0 (from v0.28.0) ([PR 2202](https://github.com/provenance-io/provenance/pull/2202), [PR 2209](https://github.com/provenance-io/provenance/pull/2209)).
* `golang.org/x/oauth2` bumped to v0.23.0 (from v0.22.0) [PR 2209](https://github.com/provenance-io/provenance/pull/2209).
* `golang.org/x/sys` bumped to v0.26.0 (from v0.24.0) ([PR 2202](https://github.com/provenance-io/provenance/pull/2202), [PR 2209](https://github.com/provenance-io/provenance/pull/2209)).
* `golang.org/x/term` bumped to v0.25.0 (from v0.23.0) ([PR 2202](https://github.com/provenance-io/provenance/pull/2202), [PR 2209](https://github.com/provenance-io/provenance/pull/2209)).
* `go.etcd.io/bbolt` bumped to v1.4.0-alpha.0.0.20240404170359-43604f3112c5 (from v1.3.10) [PR 2202](https://github.com/provenance-io/provenance/pull/2202).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.20.0...v1.20.1
* https://github.com/provenance-io/provenance/compare/v1.19.1...v1.20.1
