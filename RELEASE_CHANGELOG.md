## [v1.29.0](https://github.com/provenance-io/provenance/releases/tag/v1.29.0) 2026-06-08

Provenance Blockchain version `v1.29.0` contains some exciting new features, improvements, and bug fixes.

### Features

* Wire the vault module params and initialize them (`tech_fee_address`, `default_aum_fee_bips`) during the `edelweiss-rc3` and `edelweiss` upgrades [PR 2749](https://github.com/provenance-io/provenance/pull/2749).

### Improvements

* Limit a trigger's total gas to the same 4,000,000 gas that a tx gets [PR 2572](https://github.com/provenance-io/provenance/pull/2572).
* Add duplication checks to CreateSecuritization [#2682](https://github.com/provenance-io/provenance/issues/2682).
* Use private struct types for bypass context keys to avoid cross-package collisions [#2694](https://github.com/provenance-io/provenance/issues/2694).
* Refactor MustExtractDenomFromPacketOnRecv to return an error instead of panicking [#2699](https://github.com/provenance-io/provenance/issues/2699).
* Refactor SetMarker to return an error instead of panicking [#2700](https://github.com/provenance-io/provenance/issues/2700).
* Fix metadata account-data REST query [#2708](https://github.com/provenance-io/provenance/issues/2708).
* Improved type checking when processing triggers [PR 2730](https://github.com/provenance-io/provenance/pull/2730).
* Properly restrict block-time trigger timestamps to prevent overflow-based trigger suppression [PR 2732](https://github.com/provenance-io/provenance/pull/2732).
* Improved supply check during some marker operations [PR 2734](https://github.com/provenance-io/provenance/pull/2734).
* Add the `edelweiss` upgrades [PR 2735](https://github.com/provenance-io/provenance/pull/2735).

### Bug Fixes

* Fix faulty test assertions in metadata address tests [#2392](https://github.com/provenance-io/provenance/issues/2392).
* PurgeAttribute no longer leaves orphaned expiration lookup entries [#2684](https://github.com/provenance-io/provenance/issues/2684).
* UpdateAttribute now correctly propagates the expiration date in the record [#2685](https://github.com/provenance-io/provenance/issues/2685).
* Add missing `ledger_class_id` to CLI ledger response [#2692](https://github.com/provenance-io/provenance/issues/2692).
* Ensure DestroyLedger removes settlement instructions to prevent orphaned records in FundTransfersWithSettlement [#2698](https://github.com/provenance-io/provenance/issues/2698).
* Make GetWhitelistedQuery return a copy of the query response [#2701](https://github.com/provenance-io/provenance/issues/2701).
* Fix the wasm snapshotter so that statesync works again [PR 2725](https://github.com/provenance-io/provenance/pull/2725).
* Properly handle an IBC v2 error acknowledgment [PR 2726](https://github.com/provenance-io/provenance/pull/2726).
* Store a backup flatfees gas meter in the context [PR 2727](https://github.com/provenance-io/provenance/pull/2727).
* ~~Make the Wasm GRPC Querier return the correct response types [PR 2728](https://github.com/provenance-io/provenance/pull/2728).~~
* Revert PR 2728 and make the wasm grpc querier return a ResponseQuery again [PR 2741](https://github.com/provenance-io/provenance/pull/2741).
* Attributes: Truncate the expiration date to the second [PR 2729](https://github.com/provenance-io/provenance/pull/2729).
* Attributes: Include the current block second when finding expired attributes to delete [PR 2729](https://github.com/provenance-io/provenance/pull/2729).
* Fix event trigger processing by correcting event scanning and ensuring proper queuing of matching events for trigger execution [PR 2733](https://github.com/provenance-io/provenance/pull/2733).
* Prevent markers from using the some denoms [PR 2747](https://github.com/provenance-io/provenance/pull/2747).

### Dependencies

* `github.com/CosmWasm/wasmd` bumped to v0.61.10-pio-2 of `github.com/provenance-io/wasmd` (from v0.61.10-pio-1 of `github.com/provenance-io/wasmd`) [PR 2731](https://github.com/provenance-io/provenance/pull/2731).
* `github.com/CosmWasm/wasmvm/v2` removed at v2.3.2 [PR 2728](https://github.com/provenance-io/provenance/pull/2728).
* `github.com/CosmWasm/wasmvm/v3` bumped to v3.0.5 (from v3.0.3) [PR 2723](https://github.com/provenance-io/provenance/pull/2723).
* `github.com/cncf/xds/go` bumped to v0.0.0-20260202195803-dba9d589def2 (from v0.0.0-20251210132809-ee656c7534f5) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `github.com/cometbft/cometbft` bumped to v0.38.23 (from v0.38.21) [PR 2718](https://github.com/provenance-io/provenance/pull/2718).
* `github.com/cosmos/ibc-go/v10` bumped to v10.7.0 (from v10.5.1) ([PR 2697](https://github.com/provenance-io/provenance/pull/2697), [PR 2736](https://github.com/provenance-io/provenance/pull/2736)).
* `github.com/envoyproxy/go-control-plane/envoy` bumped to v1.37.0 (from v1.36.0) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `github.com/envoyproxy/protoc-gen-validate` bumped to v1.3.3 (from v1.3.0) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `github.com/lib/pq` bumped to v1.12.0 (from v1.10.9) [PR 2718](https://github.com/provenance-io/provenance/pull/2718).
* `github.com/minio/highwayhash` bumped to v1.0.4 (from v1.0.3) [PR 2718](https://github.com/provenance-io/provenance/pull/2718).
* `github.com/petermattis/goid` bumped to v0.0.0-20250813065127-a731cc31b4fe (from v0.0.0-20240813172612-4fcff4a6cae7) [PR 2718](https://github.com/provenance-io/provenance/pull/2718).
* `github.com/provlabs/vault` bumped to v1.1.0 (from v1.0.15) [PR 2749](https://github.com/provenance-io/provenance/pull/2749).
* `github.com/rs/zerolog` bumped to v1.35.1 (from v1.35.0) [PR 2681](https://github.com/provenance-io/provenance/pull/2681).
* `github.com/sasha-s/go-deadlock` bumped to v0.3.9 (from v0.3.5) [PR 2718](https://github.com/provenance-io/provenance/pull/2718).
* `golang.org/x/crypto` bumped to v0.50.0 (from v0.49.0) [PR 2719](https://github.com/provenance-io/provenance/pull/2719).
* `golang.org/x/mod` bumped to v0.35.0 (from v0.34.0) [PR 2719](https://github.com/provenance-io/provenance/pull/2719).
* `golang.org/x/net` bumped to v0.53.0 (from v0.52.0) [PR 2719](https://github.com/provenance-io/provenance/pull/2719).
* `golang.org/x/sys` bumped to v0.43.0 (from v0.42.0) [PR 2719](https://github.com/provenance-io/provenance/pull/2719).
* `golang.org/x/term` bumped to v0.42.0 (from v0.41.0) [PR 2719](https://github.com/provenance-io/provenance/pull/2719).
* `golang.org/x/text` bumped to v0.37.0 (from v0.36.0) [PR 2719](https://github.com/provenance-io/provenance/pull/2719).
* `golang.org/x/tools` bumped to v0.44.0 (from v0.43.0) [PR 2719](https://github.com/provenance-io/provenance/pull/2719).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20260226221140-a57be14db171 (from v0.0.0-20260203192932-546029d2fa20) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `google.golang.org/grpc` bumped to v1.81.1 (from v1.80.0) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `go.opentelemetry.io/contrib/detectors/gcp` bumped to v1.42.0 (from v1.39.0) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `go.opentelemetry.io/otel/metric` bumped to v1.43.0 (from v1.42.0) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `go.opentelemetry.io/otel/sdk/metric` bumped to v1.43.0 (from v1.42.0) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `go.opentelemetry.io/otel/sdk` bumped to v1.43.0 (from v1.42.0) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `go.opentelemetry.io/otel/trace` bumped to v1.43.0 (from v1.42.0) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).
* `go.opentelemetry.io/otel` bumped to v1.43.0 (from v1.42.0) [PR 2724](https://github.com/provenance-io/provenance/pull/2724).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.28.0...v1.29.0

