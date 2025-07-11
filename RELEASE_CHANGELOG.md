## [v1.25.0](https://github.com/provenance-io/provenance/releases/tag/v1.25.0) 2025-07-11

Provenance Blockchain version `v1.25.0` contains some exciting new features and improves security.

### Features

* Added support for an optional `concrete_type` in attributes to describe value types [#1588](https://github.com/provenance-io/provenance/issues/1588).
* Added optional recipient address to support withdrawal when minting [#1841](https://github.com/provenance-io/provenance/issues/1841).
* Marker: Added msg to support for revoking fee grants issued by admin [#2098](https://github.com/provenance-io/provenance/issues/2098).
* Added MultiAuthorization: allows creating multiple sub-authorizations that must all approve a message before it can execute [#2207](https://github.com/provenance-io/provenance/issues/2207).
* The `GetAccountCommitments` query now lets you limit results to a specific denom [#2252](https://github.com/provenance-io/provenance/issues/2252).
* Exchange: Added Market Transfer Commitment message for direct fund transfers between markets [#2322](https://github.com/provenance-io/provenance/issues/2322).
* Added governance-only tx endpoint to the hold module for unlocking vesting accounts back to base accounts [#2347](https://github.com/provenance-io/provenance/issues/2347).
* Emit upgrade event at start of upgrade with plan name [#2356](https://github.com/provenance-io/provenance/issues/2356).
* Create the alyssum upgrades and remove the zomp upgrades [PR 2389](https://github.com/provenance-io/provenance/pull/2389).

### Improvements

* Replace use of `SpendableBalances` in the hold module with a combination of `LockedCoins` and `GetBalance` [PR 2393](https://github.com/provenance-io/provenance/pull/2393).

### Bug Fixes

* Registered legacy `MsgExecuteContract` (v1beta1) for backward compatibility in transaction decoding [#2311](https://github.com/provenance-io/provenance/issues/2311).
* Updated Linux build to use Ubuntu 20.04 for glibc compatibility [#2376](https://github.com/provenance-io/provenance/issues/2376).

### Dependencies

* `cosmossdk.io/log` bumped to v1.6.0 (from v1.5.0) [PR 2341](https://github.com/provenance-io/provenance/pull/2341).
* `github.com/CosmWasm/wasmvm/v2` bumped to v2.2.4 (from v2.2.3) [PR 2352](https://github.com/provenance-io/provenance/pull/2352).
* `github.com/bytedance/sonic/loader` bumped to v0.2.4 (from v0.2.0) [PR 2341](https://github.com/provenance-io/provenance/pull/2341).
* `github.com/bytedance/sonic` bumped to v1.13.1 (from v1.12.3) [PR 2341](https://github.com/provenance-io/provenance/pull/2341).
* `github.com/cloudwego/base64x` bumped to v0.1.5 (from v0.1.4) [PR 2341](https://github.com/provenance-io/provenance/pull/2341).
* `github.com/cloudwego/iasm` removed at v0.2.0 [PR 2341](https://github.com/provenance-io/provenance/pull/2341).
* `github.com/cockroachdb/pebble` bumped to v1.1.5 (from v1.1.2) [PR 2368](https://github.com/provenance-io/provenance/pull/2368).
* `github.com/cometbft/cometbft` bumped to v0.38.18 (from v0.38.17) [PR 2386](https://github.com/provenance-io/provenance/pull/2386).
* `github.com/cosmos/cosmos-db` bumped to v1.1.3 (from v1.1.1) [PR 2368](https://github.com/provenance-io/provenance/pull/2368).
* `github.com/cosmos/cosmos-sdk` bumped to v0.50.14-pio-1 of `github.com/provenance-io/cosmos-sdk` (from v0.50.13-pio-1 of `github.com/provenance-io/cosmos-sdk`) [PR 2389](https://github.com/provenance-io/provenance/pull/2389).
* `github.com/decred/dcrd/dcrec/secp256k1/v4` bumped to v4.4.0 (from v4.3.0) [PR 2386](https://github.com/provenance-io/provenance/pull/2386).
* `github.com/google/go-cmp` bumped to v0.7.0 (from v0.6.0) [PR 2371](https://github.com/provenance-io/provenance/pull/2371).
* `github.com/klauspost/compress` bumped to v1.17.11 (from v1.17.9) [PR 2386](https://github.com/provenance-io/provenance/pull/2386).
* `github.com/klauspost/cpuid/v2` bumped to v2.2.10 (from v2.2.4) [PR 2341](https://github.com/provenance-io/provenance/pull/2341).
* `github.com/mattn/go-colorable` bumped to v0.1.14 (from v0.1.13) [PR 2341](https://github.com/provenance-io/provenance/pull/2341).
* `github.com/prometheus/client_golang` bumped to v1.21.0 (from v1.20.5) [PR 2386](https://github.com/provenance-io/provenance/pull/2386).
* `github.com/spf13/cast` bumped to v1.9.2 (from v1.7.1) ([PR 2353](https://github.com/provenance-io/provenance/pull/2353), [PR 2367](https://github.com/provenance-io/provenance/pull/2367)).
* `golang.org/x/arch` bumped to v0.15.0 (from v0.3.0) [PR 2341](https://github.com/provenance-io/provenance/pull/2341).
* `golang.org/x/crypto` bumped to v0.36.0 (from v0.35.0) [PR 2327](https://github.com/provenance-io/provenance/pull/2327).
* `golang.org/x/net` bumped to v0.38.0 (from v0.36.0) [PR 2327](https://github.com/provenance-io/provenance/pull/2327).
* `golang.org/x/oauth2` bumped to v0.28.0 (from v0.25.0) ([PR 2342](https://github.com/provenance-io/provenance/pull/2342), [PR 2371](https://github.com/provenance-io/provenance/pull/2371)).
* `golang.org/x/sync` bumped to v0.16.0 (from v0.12.0) ([PR 2337](https://github.com/provenance-io/provenance/pull/2337), [PR 2372](https://github.com/provenance-io/provenance/pull/2372), [PR 2391](https://github.com/provenance-io/provenance/pull/2391)).
* `golang.org/x/sys` bumped to v0.31.0 (from v0.30.0) [PR 2341](https://github.com/provenance-io/provenance/pull/2341).
* `golang.org/x/term` bumped to v0.30.0 (from v0.29.0) [PR 2327](https://github.com/provenance-io/provenance/pull/2327).
* `golang.org/x/text` bumped to v0.27.0 (from v0.23.0) ([PR 2337](https://github.com/provenance-io/provenance/pull/2337), [PR 2372](https://github.com/provenance-io/provenance/pull/2372), [PR 2391](https://github.com/provenance-io/provenance/pull/2391)).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20250324211829-b45e905df463 (from v0.0.0-20250106144421-5f5ef82da422) ([PR 2342](https://github.com/provenance-io/provenance/pull/2342), [PR 2371](https://github.com/provenance-io/provenance/pull/2371)).
* `google.golang.org/genproto/googleapis/rpc` bumped to v0.0.0-20250324211829-b45e905df463 (from v0.0.0-20250115164207-1a7da9e5054f) ([PR 2342](https://github.com/provenance-io/provenance/pull/2342), [PR 2371](https://github.com/provenance-io/provenance/pull/2371)).
* `google.golang.org/grpc` bumped to v1.73.0 (from v1.71.0) ([PR 2342](https://github.com/provenance-io/provenance/pull/2342), [PR 2360](https://github.com/provenance-io/provenance/pull/2360), [PR 2371](https://github.com/provenance-io/provenance/pull/2371)).
* `go.opentelemetry.io/otel/metric` bumped to v1.35.0 (from v1.34.0) [PR 2371](https://github.com/provenance-io/provenance/pull/2371).
* `go.opentelemetry.io/otel/trace` bumped to v1.35.0 (from v1.34.0) [PR 2371](https://github.com/provenance-io/provenance/pull/2371).
* `go.opentelemetry.io/otel` bumped to v1.35.0 (from v1.34.0) [PR 2371](https://github.com/provenance-io/provenance/pull/2371).
* `go.yaml.in/yaml/v2` added at v2.4.2 [PR 2382](https://github.com/provenance-io/provenance/pull/2382).
* `sigs.k8s.io/yaml` bumped to v1.5.0 (from v1.4.0) [PR 2382](https://github.com/provenance-io/provenance/pull/2382).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.24.0...v1.25.0

