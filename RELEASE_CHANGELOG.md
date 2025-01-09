## [v1.21.0](https://github.com/provenance-io/provenance/releases/tag/v1.21.0) 2025-01-09

Provenance Blockchain version `v1.21.0` updates all validators to have a 60% commission with 60% max. It also updates the staking parameters to set a minimum commission of 60%.

### Features

* Create the wisteria upgrade that will set validator commission rates [PR 2260](https://github.com/provenance-io/provenance/pull/2260).

### Dependencies

* `bufbuild/buf-setup-action` bumped to 1.48.0 (from 1.46.0) [PR 2247](https://github.com/provenance-io/provenance/pull/2247).
* `cloud.google.com/go/compute/metadata` bumped to v0.5.2 (from v0.5.0) [PR 2235](https://github.com/provenance-io/provenance/pull/2235).
* `codecov/codecov-action` bumped to 5 (from 4) [PR 2217](https://github.com/provenance-io/provenance/pull/2217).
* `cosmossdk.io/client/v2` bumped to v2.0.0-beta.7 (from v2.0.0-beta.5) ([PR 2224](https://github.com/provenance-io/provenance/pull/2224), [PR 2231](https://github.com/provenance-io/provenance/pull/2231)).
* `cosmossdk.io/depinject` bumped to v1.1.0 (from v1.0.0) [PR 2239](https://github.com/provenance-io/provenance/pull/2239).
* `cosmossdk.io/log` bumped to v1.5.0 (from v1.4.1) [PR 2214](https://github.com/provenance-io/provenance/pull/2214).
* `cosmossdk.io/x/tx` bumped to v0.13.7 (from v0.13.5) ([PR 2234](https://github.com/provenance-io/provenance/pull/2234), [PR 2239](https://github.com/provenance-io/provenance/pull/2239)).
* `github.com/CosmWasm/wasmvm/v2` bumped to v2.1.4 (from v2.1.3) [PR 2230](https://github.com/provenance-io/provenance/pull/2230).
* `github.com/bytedance/sonic/loader` added at v0.2.0 [PR 2214](https://github.com/provenance-io/provenance/pull/2214).
* `github.com/bytedance/sonic` added at v1.12.3 [PR 2214](https://github.com/provenance-io/provenance/pull/2214).
* `github.com/cloudwego/base64x` added at v0.1.4 [PR 2214](https://github.com/provenance-io/provenance/pull/2214).
* `github.com/cloudwego/iasm` added at v0.2.0 [PR 2214](https://github.com/provenance-io/provenance/pull/2214).
* `github.com/cockroachdb/pebble` bumped to v1.1.2 (from v1.1.1) [PR 2225](https://github.com/provenance-io/provenance/pull/2225).
* `github.com/cosmos/cosmos-db` bumped to v1.1.0 (from v1.0.2) [PR 2225](https://github.com/provenance-io/provenance/pull/2225).
* `github.com/cosmos/cosmos-sdk` bumped to v0.50.11-pio-1 of `github.com/provenance-io/cosmos-sdk` (from v0.50.10-pio-1 of `github.com/provenance-io/cosmos-sdk`) [PR 2239](https://github.com/provenance-io/provenance/pull/2239).
* `github.com/cosmos/iavl` bumped to v1.2.2 (from v1.2.0) but is still replaced by v1.2.0 of `github.com/cosmos/iavl` [PR 2239](https://github.com/provenance-io/provenance/pull/2239).
* `github.com/emicklei/dot` bumped to v1.6.2 (from v1.6.1) [PR 2239](https://github.com/provenance-io/provenance/pull/2239).
* `github.com/go-logr/logr` bumped to v1.4.2 (from v1.4.1) [PR 2235](https://github.com/provenance-io/provenance/pull/2235).
* `github.com/klauspost/cpuid/v2` added at v2.2.4 [PR 2214](https://github.com/provenance-io/provenance/pull/2214).
* `github.com/spf13/cast` bumped to v1.7.1 (from v1.7.0) [PR 2246](https://github.com/provenance-io/provenance/pull/2246).
* `github.com/stretchr/testify` bumped to v1.10.0 (from v1.9.0) [PR 2226](https://github.com/provenance-io/provenance/pull/2226).
* `github.com/twitchyliquid64/golang-asm` added at v0.15.1 [PR 2214](https://github.com/provenance-io/provenance/pull/2214).
* `golang.org/x/arch` added at v0.3.0 [PR 2214](https://github.com/provenance-io/provenance/pull/2214).
* `golang.org/x/crypto` bumped to v0.31.0 (from v0.28.0) [PR 2233](https://github.com/provenance-io/provenance/pull/2233).
* `golang.org/x/sync` bumped to v0.10.0 (from v0.8.0) ([PR 2213](https://github.com/provenance-io/provenance/pull/2213), [PR 2233](https://github.com/provenance-io/provenance/pull/2233)).
* `golang.org/x/sys` bumped to v0.28.0 (from v0.26.0) [PR 2233](https://github.com/provenance-io/provenance/pull/2233).
* `golang.org/x/term` bumped to v0.27.0 (from v0.25.0) [PR 2233](https://github.com/provenance-io/provenance/pull/2233).
* `golang.org/x/text` bumped to v0.21.0 (from v0.19.0) ([PR 2213](https://github.com/provenance-io/provenance/pull/2213), [PR 2233](https://github.com/provenance-io/provenance/pull/2233)).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20241015192408-796eee8c2d53 (from v0.0.0-20240814211410-ddb44dafa142) ([PR 2215](https://github.com/provenance-io/provenance/pull/2215), [PR 2235](https://github.com/provenance-io/provenance/pull/2235)).
* `google.golang.org/genproto/googleapis/rpc` bumped to v0.0.0-20241015192408-796eee8c2d53 (from v0.0.0-20240814211410-ddb44dafa142) ([PR 2215](https://github.com/provenance-io/provenance/pull/2215), [PR 2235](https://github.com/provenance-io/provenance/pull/2235)).
* `google.golang.org/grpc` bumped to v1.69.2 (from v1.67.1) ([PR 2215](https://github.com/provenance-io/provenance/pull/2215), [PR 2235](https://github.com/provenance-io/provenance/pull/2235), [PR 2244](https://github.com/provenance-io/provenance/pull/2244)).
* `google.golang.org/protobuf` bumped to v1.36.1 (from v1.35.1) ([PR 2218](https://github.com/provenance-io/provenance/pull/2218), [PR 2237](https://github.com/provenance-io/provenance/pull/2237), [PR 2250](https://github.com/provenance-io/provenance/pull/2250)).
* `go.opentelemetry.io/otel/metric` bumped to v1.31.0 (from v1.24.0) [PR 2235](https://github.com/provenance-io/provenance/pull/2235).
* `go.opentelemetry.io/otel/trace` bumped to v1.31.0 (from v1.24.0) [PR 2235](https://github.com/provenance-io/provenance/pull/2235).
* `go.opentelemetry.io/otel` bumped to v1.31.0 (from v1.24.0) [PR 2235](https://github.com/provenance-io/provenance/pull/2235).
* `peter-evans/create-pull-request` bumped to 7.0.6 (from 7.0.5) [PR 2251](https://github.com/provenance-io/provenance/pull/2251).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.20.2...v1.21.0

