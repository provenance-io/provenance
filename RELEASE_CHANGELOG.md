## [v1.20.0-rc4](https://github.com/provenance-io/provenance/releases/tag/v1.20.0-rc4) 2024-10-25

Building or installing `provenanced` from source now requires you to use [Go 1.23](https://golang.org/dl/).
Linting now requires `golangci-lint` v1.60.2. You can update yours using `make golangci-lint-update` or install it using `make golangci-lint`.

Version `v1.20.0-rc4` is state-compatible with `v1.20.0-rc3` and `v1.20.0-rc2` and it is recommended that nodes update at their earliest convenience.

### Improvements

* Hard code the mainnet `consensus.timeout_commit` config value to 3.5s [#2121](https://github.com/provenance-io/provenance/issues/2121).
* Update the prep-release script to combine dependency changelog entries [PR 2181](https://github.com/provenance-io/provenance/pull/2181).
* Update the proto file links in the spec docs to point to `v1.20.0` (instead of `v1.19.0`) [PR 2192](https://github.com/provenance-io/provenance/pull/2192).
* Suppress the events emitted during the metadata migration that changes how scope value owners are recorded [PR 2195](https://github.com/provenance-io/provenance/pull/2195).

### Bug Fixes

* Fix the query metadata recordspec command to use the RecordSpecification query when provided a recspec id [#2148](https://github.com/provenance-io/provenance/issues/2148).
* Register the params types with the codecs so old gov props can be read [PR 2198](https://github.com/provenance-io/provenance/pull/2198).
* Add the query flags to the query wasm build-addr command [PR 2199](https://github.com/provenance-io/provenance/pull/2199).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.20.0-rc3...v1.20.0-rc4
* https://github.com/provenance-io/provenance/compare/v1.19.1...v1.20.0-rc4

---

## [v1.20.0-rc3](https://github.com/provenance-io/provenance/releases/tag/v1.20.0-rc3) 2024-10-16

Building or installing `provenanced` from source now requires you to use [Go 1.23](https://golang.org/dl/).
Linting now requires `golangci-lint` v1.60.2. You can update yours using `make golangci-lint-update` or install it using `make golangci-lint`.

Version `v1.20.0-rc3` should be used in place of `v1.20.0-rc2`. Version `v1.20.0-rc2` doesn't allow restarting a node once it has been stopped (after applying the `viridian-rc1` upgrade). Switching to `v1.20.0-rc3` will fix the error `failed to load latest version: version of store params mismatch root store's version`. It is also safe to use `v1.20.0-rc3` to apply the upgrade (even though the upgrade says to use `v1.20.0-rc2`).

### Bug Fixes

* Remove the params store key and transient store key from the app [PR 2189](https://github.com/provenance-io/provenance/pull/2189).
  This fixes a problem in `v1.20.0-rc2` that prevented nodes from restarting if stopped after the upgrade.

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.20.0-rc2...v1.20.0-rc3
* https://github.com/provenance-io/provenance/compare/v1.19.1...v1.20.0-rc3

---

## [v1.20.0-rc2](https://github.com/provenance-io/provenance/releases/tag/v1.20.0-rc2) 2024-10-16

Building or installing `provenanced` from source now requires you to use [Go 1.23](https://golang.org/dl/).
Linting now requires `golangci-lint` v1.60.2. You can update yours using `make golangci-lint-update` or install it using `make golangci-lint`.

### Bug Fixes

* Rename the RELEASE_NOTES.md file to RELEASE_CHANGELOG.md [PR 2182](https://github.com/provenance-io/provenance/pull/2182).
* Fix the heighliner build [PR 2184](https://github.com/provenance-io/provenance/pull/2184).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.20.0-rc1...v1.20.0-rc2
* https://github.com/provenance-io/provenance/compare/v1.19.1...v1.20.0-rc2

---

## [v1.20.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.20.0-rc1) 2024-10-14

Building or installing `provenanced` from source now requires you to use [Go 1.23](https://golang.org/dl/).
Linting now requires `golangci-lint` v1.60.2. You can update yours using `make golangci-lint-update` or install it using `make golangci-lint`.

### Features

* Create the `viridian` upgrade [#2137](https://github.com/provenance-io/provenance/issues/2137).

### Improvements

* Only set a NAV record when explicitly provided [#2030](https://github.com/provenance-io/provenance/issues/2030).
* Create the `build-debug` make target for building a provenanced binary that allows debugging [#2062](https://github.com/provenance-io/provenance/issues/2062).
* Switch to `unclog` for unreleased changelog entries [PR 2112](https://github.com/provenance-io/provenance/pull/2112).
* Address missing documentation on marker nav command [#2128](https://github.com/provenance-io/provenance/issues/2128).
* Address missing documentation on metadata/scope nav command [#2134](https://github.com/provenance-io/provenance/issues/2134).
* Clean up some unused stuff from our makefiles [PR 2136](https://github.com/provenance-io/provenance/pull/2136).
* Use the bank module to keep track of the value owner of scopes [#2137](https://github.com/provenance-io/provenance/issues/2137).
* Create the `add-change.sh` script to make it easier to add changelog entries [PR 2166](https://github.com/provenance-io/provenance/pull/2166).
* Delete the `umber` upgrade and the stuff it needed that nothing else needed [PR 2176](https://github.com/provenance-io/provenance/pull/2176).

### Bug Fixes

* Fix proto markdown generation and regenerate proto-docs.md [#376](https://github.com/provenance-io/provenance/issues/376).
* Allow marker funds to be used via feegrant again [#2110](https://github.com/provenance-io/provenance/issues/2110).
* Make our proto generation stuff work again [#2135](https://github.com/provenance-io/provenance/issues/2135).
* Remove the telemetry counters from the metadata module since it wasn't actually doing anything [#2144](https://github.com/provenance-io/provenance/issues/2144).
* Fix telemetry to include data from cometbft that got unknowningly removed with v1.19 [PR 2177](https://github.com/provenance-io/provenance/pull/2177).

### Client Breaking

* remove old provwasm bindings [PR 2119](https://github.com/provenance-io/provenance/pull/2119).
* In proofs.proto, the `HashOp` enum value 3 has changed to `KECCAK256` (from `KECCAK`) [PR 2153](https://github.com/provenance-io/provenance/pull/2153).
* Fixes the metadata nav cli command example to use the correct module name [#2058](https://github.com/provenance-io/provenance/issues/2058).
  During this fix it was discovered that the volume parameter was not present but was required for proper price ratios.  The volume
  parameter has been added to the NAV entry and when not present a default value of 1 (which should be the most common case for a scope) is
  used instead.

### Api Breaking

* The `Ownership` query in the `x/metadata` module now only returns scopes that have the provided address in the `owners` list [#2137](https://github.com/provenance-io/provenance/issues/2137).
  Previously, if an address was the value owner of a scope, but not in the `owners` list, the scope would be returned
  by the `Ownership` query when given that address.  That is no longer the case.
  The `ValueOwnership` query can be to identify scopes with a specific value owner (like before).
  If a scope has a value owner that is also in its `owners` list, it will still be returned by both queries.
* The `WriteScope` endpoint now uses the `scope.value_owner_address` differently [#2137](https://github.com/provenance-io/provenance/issues/2137).
  If it is empty, it indicates that there is no change to the value owner of the scope and the releated lookups and validation
  are skipped. If it isn't empty, the current value owner will be looked up and the coin for the scope will be transferred to
  the provided address (assuming signer validation passed).
* An authz grant on `MsgWriteScope` no longer also applies to the `UpdateValueOwners` or `MigrateValueOwner` endpoints [#2137](https://github.com/provenance-io/provenance/issues/2137).
* The params module has been removed [PR 2176](https://github.com/provenance-io/provenance/pull/2176).
  All params module endpoints have been removed. All modules now manage their params on their own.

### Dependencies

* `bufbuild/buf-setup-action` bumped to 1.36.0 (from 1.34.0) [PR 2122](https://github.com/provenance-io/provenance/pull/2122).
* `bufbuild/buf-setup-action` bumped to 1.37.0 (from 1.36.0) [PR 2131](https://github.com/provenance-io/provenance/pull/2131).
* `bufbuild/buf-setup-action` bumped to 1.38.0 (from 1.37.0) [PR 2133](https://github.com/provenance-io/provenance/pull/2133).
* `bufbuild/buf-setup-action` bumped to 1.39.0 (from 1.38.0) [PR 2138](https://github.com/provenance-io/provenance/pull/2138).
* `bufbuild/buf-setup-action` bumped to 1.41.0 (from 1.39.0) [PR 2151](https://github.com/provenance-io/provenance/pull/2151).
* `bufbuild/buf-setup-action` bumped to 1.42.0 (from 1.41.0) [PR 2155](https://github.com/provenance-io/provenance/pull/2155).
* `bufbuild/buf-setup-action` bumped to 1.43.0 (from 1.42.0) [PR 2164](https://github.com/provenance-io/provenance/pull/2164).
* `bufbuild/buf-setup-action` bumped to 1.44.0 (from 1.43.0) [PR 2168](https://github.com/provenance-io/provenance/pull/2168).
* `bufbuild/buf-setup-action` bumped to 1.45.0 (from 1.44.0) [PR 2174](https://github.com/provenance-io/provenance/pull/2174).
* `cloud.google.com/go/compute/metadata` bumped to v0.5.0 (from v0.3.0) [PR 2157](https://github.com/provenance-io/provenance/pull/2157).
* `cosmossdk.io/api` bumped to v0.7.6 (from v0.7.5) [PR 2162](https://github.com/provenance-io/provenance/pull/2162).
* `cosmossdk.io/client/v2` bumped to v2.0.0-beta.4 (from v2.0.0-beta.2) [PR 2100](https://github.com/provenance-io/provenance/pull/2100).
* `cosmossdk.io/client/v2` bumped to v2.0.0-beta.5 (from v2.0.0-beta.4) [PR 2153](https://github.com/provenance-io/provenance/pull/2153).
* `cosmossdk.io/core` bumped to v0.11.1 (from v0.11.0) [PR 2107](https://github.com/provenance-io/provenance/pull/2107).
* `cosmossdk.io/core` bumped to v0.11.2 (from v0.11.1) [PR 2130](https://github.com/provenance-io/provenance/pull/2130).
* `cosmossdk.io/depinject` bumped to v1.0.0 (from v1.0.0-alpha.4) [PR 2107](https://github.com/provenance-io/provenance/pull/2107).
* `cosmossdk.io/log` bumped to v1.4.0 (from v1.3.1) [PR 2116](https://github.com/provenance-io/provenance/pull/2116).
* `cosmossdk.io/log` bumped to v1.4.1 (from v1.4.0) [PR 2129](https://github.com/provenance-io/provenance/pull/2129).
* `cosmossdk.io/store` bumped to v1.1.1 (from v1.1.0) [PR 2175](https://github.com/provenance-io/provenance/pull/2175).
* `cosmossdk.io/x/tx` bumped to v0.13.4 (from v0.13.3) [PR 2113](https://github.com/provenance-io/provenance/pull/2113).
* `cosmossdk.io/x/tx` bumped to v0.13.5 (from v0.13.4) [PR 2154](https://github.com/provenance-io/provenance/pull/2154).
* `github.com/CosmWasm/wasmvm/v2` bumped to v2.1.2 (from v2.1.0) [PR 2126](https://github.com/provenance-io/provenance/pull/2126).
* `github.com/CosmWasm/wasmvm/v2` bumped to v2.1.3 (from v2.1.2) [PR 2161](https://github.com/provenance-io/provenance/pull/2161).
* `github.com/btcsuite/btcd/btcec/v2` bumped to v2.3.4 (from v2.3.2) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/cockroachdb/errors` bumped to v1.11.3 (from v1.11.1) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/cockroachdb/fifo` added at v0.0.0-20240606204812-0bbfbd93a7ce [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/cockroachdb/pebble` bumped to v1.1.1 (from v1.1.0) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/cometbft/cometbft-db` bumped to v0.11.0 (from v0.9.1) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/cometbft/cometbft` bumped to v0.38.11 (from v0.38.10) [PR 2120](https://github.com/provenance-io/provenance/pull/2120).
* `github.com/cometbft/cometbft` bumped to v0.38.12 (from v0.38.11) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/cosmos/cosmos-sdk` bumped to v0.50.10-pio-1 of `github.com/provenance-io/cosmos-sdk` (from v0.50.7-pio-1 of `github.com/provenance-io/cosmos-sdk`) [PR 2175](https://github.com/provenance-io/provenance/pull/2175).
* `github.com/cosmos/gogoproto` bumped to v1.7.0 (from v1.5.0) [PR 2125](https://github.com/provenance-io/provenance/pull/2125).
* `github.com/cosmos/ics23/go` bumped to v0.11.0 (from v0.10.0) [PR 2153](https://github.com/provenance-io/provenance/pull/2153).
* `github.com/golang/glog` bumped to v1.2.2 (from v1.2.1) [PR 2157](https://github.com/provenance-io/provenance/pull/2157).
* `github.com/gorilla/websocket` bumped to v1.5.3 (from v1.5.1) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/klauspost/compress` bumped to v1.17.9 (from v1.17.7) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/munnerz/goautoneg` added at v0.0.0-20191010083416-a7dc8b61c822 [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/prometheus/client_golang` bumped to v1.20.1 (from v1.19.1) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/prometheus/common` bumped to v0.55.0 (from v0.52.2) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/prometheus/procfs` bumped to v0.15.1 (from v0.13.0) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/rs/cors` bumped to v1.11.1 (from v1.11.0) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `github.com/spf13/cast` bumped to v1.7.0 (from v1.6.0) [PR 2114](https://github.com/provenance-io/provenance/pull/2114).
* `golangci-lint` bumped to v1.60.2 (from v1.54.2) [PR 2132](https://github.com/provenance-io/provenance/pull/2132).
* `golang.org/x/crypto` bumped to v0.25.0 (from v0.23.0) [PR 2107](https://github.com/provenance-io/provenance/pull/2107).
* `golang.org/x/crypto` bumped to v0.26.0 (from v0.25.0) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `golang.org/x/net` bumped to v0.27.0 (from v0.25.0) [PR 2107](https://github.com/provenance-io/provenance/pull/2107).
* `golang.org/x/net` bumped to v0.28.0 (from v0.27.0) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `golang.org/x/oauth2` bumped to v0.21.0 (from v0.20.0) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `golang.org/x/oauth2` bumped to v0.22.0 (from v0.21.0) [PR 2157](https://github.com/provenance-io/provenance/pull/2157).
* `golang.org/x/sync` bumped to v0.8.0 (from v0.7.0) [PR 2115](https://github.com/provenance-io/provenance/pull/2115).
* `golang.org/x/sys` bumped to v0.22.0 (from v0.20.0) [PR 2107](https://github.com/provenance-io/provenance/pull/2107).
* `golang.org/x/sys` bumped to v0.23.0 (from v0.22.0) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `golang.org/x/sys` bumped to v0.24.0 (from v0.23.0) [PR 2157](https://github.com/provenance-io/provenance/pull/2157).
* `golang.org/x/term` bumped to v0.22.0 (from v0.20.0) [PR 2107](https://github.com/provenance-io/provenance/pull/2107).
* `golang.org/x/term` bumped to v0.23.0 (from v0.22.0) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `golang.org/x/text` bumped to v0.17.0 (from v0.16.0) [PR 2115](https://github.com/provenance-io/provenance/pull/2115).
* `golang.org/x/text` bumped to v0.18.0 (from v0.17.0) [PR 2143](https://github.com/provenance-io/provenance/pull/2143).
* `golang.org/x/text` bumped to v0.19.0 (from v0.18.0) [PR 2170](https://github.com/provenance-io/provenance/pull/2170).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20240604185151-ef581f913117 (from v0.0.0-20240528184218-531527333157) [PR 2150](https://github.com/provenance-io/provenance/pull/2150).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20240814211410-ddb44dafa142 (from v0.0.0-20240604185151-ef581f913117) [PR 2157](https://github.com/provenance-io/provenance/pull/2157).
* `google.golang.org/genproto/googleapis/rpc` bumped to v0.0.0-20240709173604-40e1e62336c5 (from v0.0.0-20240528184218-531527333157) [PR 2107](https://github.com/provenance-io/provenance/pull/2107).
* `google.golang.org/genproto/googleapis/rpc` bumped to v0.0.0-20240814211410-ddb44dafa142 (from v0.0.0-20240709173604-40e1e62336c5) [PR 2157](https://github.com/provenance-io/provenance/pull/2157).
* `google.golang.org/grpc` bumped to v1.66.2 (from v1.65.0) [PR 2150](https://github.com/provenance-io/provenance/pull/2150).
* `google.golang.org/grpc` bumped to v1.67.0 (from v1.66.2) [PR 2157](https://github.com/provenance-io/provenance/pull/2157).
* `google.golang.org/grpc` bumped to v1.67.1 (from v1.67.0) [PR 2165](https://github.com/provenance-io/provenance/pull/2165).
* `google.golang.org/protobuf` bumped to v1.35.1 (from v1.34.2) [PR 2173](https://github.com/provenance-io/provenance/pull/2173).
* `go.etcd.io/bbolt` bumped to v1.3.10 (from v1.3.8) [PR 2142](https://github.com/provenance-io/provenance/pull/2142).
* `go` bumped to 1.23 (from 1.21) [PR 2132](https://github.com/provenance-io/provenance/pull/2132).
* `peter-evans/create-pull-request` bumped to 7.0.0 (from 6.1.0) [PR 2141](https://github.com/provenance-io/provenance/pull/2141).
* `peter-evans/create-pull-request` bumped to 7.0.2 (from 7.0.0) [PR 2152](https://github.com/provenance-io/provenance/pull/2152).
* `peter-evans/create-pull-request` bumped to 7.0.5 (from 7.0.2) [PR 2156](https://github.com/provenance-io/provenance/pull/2156).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.19.1...v1.20.0-rc1

