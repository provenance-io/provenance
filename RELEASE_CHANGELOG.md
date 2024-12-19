## [v1.20.3](https://github.com/provenance-io/provenance/releases/tag/v1.20.3) 2024-12-19

Provenance Blockchain version `v1.20.3` patches a vulnerability that might cause a node to crash, or the chain to halt. Everyone should switch to this version at their earliest convenience.

This version fixes a [security vulnerability](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-8wcc-m6j2-qxvm) in the Cosmos SDK.

Building or installing `provenanced` from source now requires you to use [Go 1.23](https://golang.org/dl/).

### Dependencies

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
* `google.golang.org/grpc` bumped to v1.69.0 (from v1.67.1) ([PR 2215](https://github.com/provenance-io/provenance/pull/2215), [PR 2235](https://github.com/provenance-io/provenance/pull/2235)).
* `google.golang.org/protobuf` bumped to v1.36.0 (from v1.35.1) ([PR 2218](https://github.com/provenance-io/provenance/pull/2218), [PR 2237](https://github.com/provenance-io/provenance/pull/2237)).
* `go.opentelemetry.io/otel/metric` bumped to v1.31.0 (from v1.24.0) [PR 2235](https://github.com/provenance-io/provenance/pull/2235).
* `go.opentelemetry.io/otel/trace` bumped to v1.31.0 (from v1.24.0) [PR 2235](https://github.com/provenance-io/provenance/pull/2235).
* `go.opentelemetry.io/otel` bumped to v1.31.0 (from v1.24.0) [PR 2235](https://github.com/provenance-io/provenance/pull/2235).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.20.2...v1.20.3

