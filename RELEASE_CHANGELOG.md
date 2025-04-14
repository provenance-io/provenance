## [v1.22.0](https://github.com/provenance-io/provenance/releases/tag/v1.22.0) 2025-04-14

Provenance Blockchain version `v1.22.0` updates several libraries to address some vulnerabilities.  Upgrading to this version will be done with a coordinated chain upgrade.

### Improvements

* Create the xenon upgrades and remove the wisteria ones [PR 2309](https://github.com/provenance-io/provenance/pull/2309).

### Dependencies

* `bufbuild/buf-setup-action` bumped to 1.50.0 (from 1.48.0) [PR 2273](https://github.com/provenance-io/provenance/pull/2273).
* `cloud.google.com/go/auth/oauth2adapt` added at v0.2.2 [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `cloud.google.com/go/auth` added at v0.6.0 [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `cloud.google.com/go/compute/metadata` bumped to v0.6.0 (from v0.5.2) [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `cloud.google.com/go/iam` bumped to v1.1.9 (from v1.1.6) [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `cloud.google.com/go/storage` bumped to v1.41.0 (from v1.38.0) [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `cloud.google.com/go` bumped to v0.115.0 (from v0.112.1) [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `cosmossdk.io/client/v2` bumped to v2.0.0-beta.8 (from v2.0.0-beta.7) [PR 2276](https://github.com/provenance-io/provenance/pull/2276).
* `cosmossdk.io/x/tx` bumped to v0.13.8 (from v0.13.7) [PR 2275](https://github.com/provenance-io/provenance/pull/2275).
* `cosmossdk.io/x/upgrade` bumped to v0.1.4 (from v0.1.3) [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `filippo.io/edwards25519` bumped to v1.1.0 (from v1.0.0) [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `github.com/CosmWasm/wasmvm/v2` bumped to v2.2.3 (from v2.1.4) ([PR 2278](https://github.com/provenance-io/provenance/pull/2278), [PR 2297](https://github.com/provenance-io/provenance/pull/2297)).
* `github.com/cometbft/cometbft-db` bumped to v0.15.0 (from v0.14.1) [PR 2274](https://github.com/provenance-io/provenance/pull/2274).
* `github.com/cosmos/cosmos-sdk` bumped to v0.50.13-pio-1 of `github.com/provenance-io/cosmos-sdk` (from v0.50.12-pio-1 of `github.com/provenance-io/cosmos-sdk`) [PR 2307](https://github.com/provenance-io/provenance/pull/2307).
* `github.com/cosmos/ibc-go/v8` bumped to v8.6.1-pio-1 of `github.com/provenance-io/ibc-go/v8` (from v8.3.2-pio-1 of `github.com/provenance-io/ibc-go/v8`) [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `github.com/cpuguy83/go-md2man/v2` bumped to v2.0.6 (from v2.0.4) [PR 2281](https://github.com/provenance-io/provenance/pull/2281).
* `github.com/dgraph-io/badger/v4` bumped to v4.3.0 (from v4.2.0) [PR 2274](https://github.com/provenance-io/provenance/pull/2274).
* `github.com/dgraph-io/ristretto` bumped to v0.1.2-0.20240116140435-c67e07994f91 (from v0.1.1) [PR 2274](https://github.com/provenance-io/provenance/pull/2274).
* `github.com/golang/glog` bumped to v1.2.4 (from v1.2.3) [PR 2268](https://github.com/provenance-io/provenance/pull/2268).
* `github.com/golang/glog` removed at v1.2.4 [PR 2274](https://github.com/provenance-io/provenance/pull/2274).
* `github.com/googleapis/gax-go/v2` bumped to v2.12.5 (from v2.12.3) [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `github.com/hashicorp/go-metrics` bumped to v0.5.4 (from v0.5.3) [PR 2262](https://github.com/provenance-io/provenance/pull/2262).
* `github.com/linxGnu/grocksdb` bumped to v1.9.3 (from v1.8.14) [PR 2274](https://github.com/provenance-io/provenance/pull/2274).
* `github.com/rogpeppe/go-internal` bumped to v1.13.1 (from v1.12.0) [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `github.com/rs/zerolog` bumped to v1.34.0 (from v1.33.0) [PR 2304](https://github.com/provenance-io/provenance/pull/2304).
* `github.com/spf13/cobra` bumped to v1.9.1 (from v1.8.1) [PR 2281](https://github.com/provenance-io/provenance/pull/2281).
* `github.com/spf13/pflag` bumped to v1.0.6 (from v1.0.5) [PR 2270](https://github.com/provenance-io/provenance/pull/2270).
* `golang.org/x/crypto` bumped to v0.35.0 (from v0.32.0) [PR 2298](https://github.com/provenance-io/provenance/pull/2298).
* `golang.org/x/net` bumped to v0.36.0 (from v0.34.0) [PR 2298](https://github.com/provenance-io/provenance/pull/2298).
* `golang.org/x/oauth2` bumped to v0.25.0 (from v0.24.0) [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `golang.org/x/sync` bumped to v0.12.0 (from v0.10.0) [PR 2293](https://github.com/provenance-io/provenance/pull/2293).
* `golang.org/x/sys` bumped to v0.30.0 (from v0.29.0) [PR 2298](https://github.com/provenance-io/provenance/pull/2298).
* `golang.org/x/term` bumped to v0.29.0 (from v0.28.0) [PR 2298](https://github.com/provenance-io/provenance/pull/2298).
* `golang.org/x/text` bumped to v0.23.0 (from v0.21.0) [PR 2293](https://github.com/provenance-io/provenance/pull/2293).
* `google.golang.org/api` bumped to v0.186.0 (from v0.171.0) [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20250106144421-5f5ef82da422 (from v0.0.0-20241202173237-19429a94021a) [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `google.golang.org/genproto/googleapis/rpc` bumped to v0.0.0-20250115164207-1a7da9e5054f (from v0.0.0-20241202173237-19429a94021a) [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `google.golang.org/genproto` bumped to v0.0.0-20240701130421-f6361c86f094 (from v0.0.0-20240227224415-6ceb2ff114de) [PR 2290](https://github.com/provenance-io/provenance/pull/2290).
* `google.golang.org/grpc` bumped to v1.71.0 (from v1.70.0) [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `google.golang.org/protobuf` bumped to v1.36.6 (from v1.36.4) ([PR 2280](https://github.com/provenance-io/provenance/pull/2280), [PR 2305](https://github.com/provenance-io/provenance/pull/2305)).
* `go.opentelemetry.io/auto/sdk` added at v1.1.0 [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `go.opentelemetry.io/otel/metric` bumped to v1.34.0 (from v1.32.0) [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `go.opentelemetry.io/otel/trace` bumped to v1.34.0 (from v1.32.0) [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `go.opentelemetry.io/otel` bumped to v1.34.0 (from v1.32.0) [PR 2292](https://github.com/provenance-io/provenance/pull/2292).
* `peter-evans/create-pull-request` bumped to 7.0.8 (from 7.0.6) [PR 2294](https://github.com/provenance-io/provenance/pull/2294).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.21.1...v1.22.0

