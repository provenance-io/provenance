## [v1.28.0-rc2](https://github.com/provenance-io/provenance/releases/tag/v1.28.0-rc2) 2026-04-16

Provenance Blockchain version `v1.28.0` contains some exciting new features, improvements, and bug fixes.

Key new features and improvements:

* Compiling Provenance Blockchain version `v1.28.0` requires [Go 1.25](https://golang.org/dl/) (specifically).
* The `CosmWasm` library was bumped, so there is a new `libwasmvm` shared library file (e.g. `libwasmvm.x86_64.so`).
* Smart contracts can now be stored without using a governance proposal.
* Now based on Cosmos-SDK v0.53.5.
* IBC libraries bumped to v10 and IBC v2 is now available.
* Enable unordered transactions.

### Bug Fixes

* When transferring commitments, properly update the commitment to the old market [PR 2676](https://github.com/provenance-io/provenance/pull/2676).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.28.0-rc1...v1.28.0-rc2
* https://github.com/provenance-io/provenance/compare/v1.27.2...v1.28.0-rc2

---

## [v1.28.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.28.0-rc1) 2026-04-13

Provenance Blockchain version `v1.28.0` contains some exciting new features, improvements, and bug fixes.

Key new features and improvements:

* Compiling Provenance Blockchain version `v1.28.0` requires [Go 1.25](https://golang.org/dl/) (specifically).
* The `CosmWasm` library was bumped, so there is a new `libwasmvm` shared library file (e.g. `libwasmvm.x86_64.so`).
* Smart contracts can now be stored without using a governance proposal.
* Now based on Cosmos-SDK v0.53.5.
* IBC libraries bumped to v10 and IBC v2 is now available.

### Features

* Allow trusted oracles to adjust flatfees conversion factor without governance [#2550](https://github.com/provenance-io/provenance/issues/2550).
* Configure circuit breaker admin permissions during upgrades [#2585](https://github.com/provenance-io/provenance/issues/2585).
* Allow smart contract storage changes without requiring a governance proposal [#2589](https://github.com/provenance-io/provenance/issues/2589).
* Create the daisy upgrades [PR 2591](https://github.com/provenance-io/provenance/pull/2591).
* Update costs of msgs `MsgWriteRecordRequest`, `MsgWriteSessionRequest` & `MsgUpdateClient` [#2604](https://github.com/provenance-io/provenance/issues/2604).
* Publish events for the flatfees module [#2651](https://github.com/provenance-io/provenance/issues/2651).
* Add ability to send funds to an account then have them committed to a market [#2659](https://github.com/provenance-io/provenance/issues/2659).

### Improvements

* Add CI check for Protobuf formatting and generation in proto.yml file [#1403](https://github.com/provenance-io/provenance/issues/1403).
* Allow denom metadata to be defined for restricted denoms using standard SDK validation [#2556](https://github.com/provenance-io/provenance/issues/2556).
* Update 3rd party swagger file [#2586](https://github.com/provenance-io/provenance/issues/2586).
* Update Sims to match v0.53.x style [#2593](https://github.com/provenance-io/provenance/issues/2593).
* Increase the max size of wasm code to 800kb (from 600kb) [PR 2660](https://github.com/provenance-io/provenance/pull/2660).
* Enable IBC v2 [PR 2663](https://github.com/provenance-io/provenance/pull/2663).
* Enhance the x/ibchooks module to be compatible with IBC v2 [PR 2663](https://github.com/provenance-io/provenance/pull/2663).
* Enhance the x/ibcratelimit module to be compatible with IBC v2 [PR 2663](https://github.com/provenance-io/provenance/pull/2663).
* Remove specialized ibc memo processing for custom marker access [PR 2663](https://github.com/provenance-io/provenance/pull/2663).

### Bug Fixes

* Use non-auto-cli version of tx hold CLI commands [PR 2581](https://github.com/provenance-io/provenance/pull/2581).
* Fixes attribute sim test to not create nil values [PR 2619](https://github.com/provenance-io/provenance/pull/2619).
* Fix admin check in CreateSecuritization [#2648](https://github.com/provenance-io/provenance/issues/2648).

### Deprecated

* `MarkerSupply` query has been deprecated. Please use the `SupplyOf` query from the `bank` module instead, which provides equivalent functionality [#1676](https://github.com/provenance-io/provenance/issues/1676).
* Remove carnation upgrade [#2587](https://github.com/provenance-io/provenance/issues/2587).

### Api Breaking

* Removed the `x/async-icq` module [#2630](https://github.com/provenance-io/provenance/issues/2630).
* Removed the `x/oracle` module [#2630](https://github.com/provenance-io/provenance/issues/2630).

### Dependencies

* `Docker images` bumped to 1.25-bookworm (from 1.23-bullseye) [#2750](https://github.com/provenance-io/provenance/issues/2750).
* `Golang` bumped to 1.25 (from 1.23) [#2750](https://github.com/provenance-io/provenance/issues/2750).
* `actions/checkout` bumped to 6 (from 5) ([PR 2545](https://github.com/provenance-io/provenance/pull/2545), [PR 2579](https://github.com/provenance-io/provenance/pull/2579)).
* `actions/download-artifact` bumped to 8 (from 6) ([PR 2572](https://github.com/provenance-io/provenance/pull/2572), [PR 2634](https://github.com/provenance-io/provenance/pull/2634)).
* `actions/upload-artifact` bumped to 7 (from 5) ([PR 2573](https://github.com/provenance-io/provenance/pull/2573), [PR 2633](https://github.com/provenance-io/provenance/pull/2633)).
* `cel.dev/expr` added at v0.24.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cel.dev/expr` bumped to v0.25.1 (from v0.24.0) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `cloud.google.com/go/auth/oauth2adapt` bumped to v0.2.8 (from v0.2.4) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cloud.google.com/go/auth` bumped to v0.18.2 (from v0.9.3) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `cloud.google.com/go/compute/metadata` bumped to v0.9.0 (from v0.7.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627)).
* `cloud.google.com/go/iam` bumped to v1.5.3 (from v1.2.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `cloud.google.com/go/monitoring` added at v1.24.2 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cloud.google.com/go/monitoring` bumped to v1.24.3 (from v1.24.2) [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `cloud.google.com/go/storage` bumped to v1.61.3 (from v1.43.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `cloud.google.com/go` bumped to v0.123.0 (from v0.115.1) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `codecov/codecov-action` bumped to 6 (from 5) [PR 2654](https://github.com/provenance-io/provenance/pull/2654).
* `cosmossdk.io/api` bumped to v0.9.2 (from v0.7.6) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cosmossdk.io/collections` bumped to v1.4.0 (from v0.4.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627)).
* `cosmossdk.io/core` bumped to v0.11.3 (from v0.11.2) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cosmossdk.io/depinject` bumped to v1.2.1 (from v1.1.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cosmossdk.io/errors` bumped to v1.1.0 (from v1.0.1) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627)).
* `cosmossdk.io/math` bumped to v1.5.3 (from v1.4.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cosmossdk.io/schema` added at v1.1.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cosmossdk.io/store` bumped to v1.1.2 (from v1.1.1) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cosmossdk.io/x/evidence` bumped to v0.2.0 (from v0.1.1) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `cosmossdk.io/x/feegrant` bumped to v0.2.0 (from v0.1.1) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `cosmossdk.io/x/nft` bumped to v0.2.0 (from v0.1.1) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `cosmossdk.io/x/tx` bumped to v0.14.0 (from v0.13.8) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `cosmossdk.io/x/upgrade` bumped to v0.2.0 (from v0.1.4) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `crazy-max/ghaction-import-gpg` bumped to 7 (from 6) [PR 2636](https://github.com/provenance-io/provenance/pull/2636).
* `docker/build-push-action` bumped to 7 (from 6) [PR 2642](https://github.com/provenance-io/provenance/pull/2642).
* `docker/login-action` bumped to 4 (from 3) [PR 2639](https://github.com/provenance-io/provenance/pull/2639).
* `docker/metadata-action` bumped to 6 (from 5) [PR 2643](https://github.com/provenance-io/provenance/pull/2643).
* `docker/setup-buildx-action` bumped to 4 (from 3) [PR 2640](https://github.com/provenance-io/provenance/pull/2640).
* `docker/setup-qemu-action` bumped to 4 (from 3) [PR 2638](https://github.com/provenance-io/provenance/pull/2638).
* `filippo.io/edwards25519` bumped to v1.1.1 (from v1.1.0) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/CosmWasm/wasmd` bumped to v0.61.10-pio-1 of `github.com/provenance-io/wasmd` (from v0.52.0-pio-1 of `github.com/provenance-io/wasmd`) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/CosmWasm/wasmvm/v2` bumped to v2.3.2 (from v2.2.4) [PR 2621](https://github.com/provenance-io/provenance/pull/2621).
* `github.com/CosmWasm/wasmvm/v3` added at v3.0.3 [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/DataDog/datadog-go` bumped to v4.8.3+incompatible (from v3.2.0+incompatible) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/DataDog/zstd` bumped to v1.5.7 (from v1.5.5) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp` added at v1.29.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp` bumped to v1.31.0 (from v1.29.0) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2658](https://github.com/provenance-io/provenance/pull/2658)).
* `github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric` added at v0.50.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric` bumped to v0.55.0 (from v0.50.0) [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping` added at v0.50.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping` bumped to v0.55.0 (from v0.50.0) [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/Microsoft/go-winio` added at v0.6.2 [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/ProjectZKM/Ziren/crates/go-runtime/zkvm_runtime` added at v0.0.0-20251001021608-1fe7b43fc4d6 [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream` added at v1.7.7 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream` bumped to v1.7.8 (from v1.7.7) [PR 2669](https://github.com/provenance-io/provenance/pull/2669).
* `github.com/aws/aws-sdk-go-v2/config` added at v1.32.12 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/credentials` added at v1.19.12 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/feature/ec2/imds` added at v1.18.20 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/internal/configsources` added at v1.4.20 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/internal/configsources` bumped to v1.4.21 (from v1.4.20) [PR 2669](https://github.com/provenance-io/provenance/pull/2669).
* `github.com/aws/aws-sdk-go-v2/internal/endpoints/v2` added at v2.7.20 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/internal/endpoints/v2` bumped to v2.7.21 (from v2.7.20) [PR 2669](https://github.com/provenance-io/provenance/pull/2669).
* `github.com/aws/aws-sdk-go-v2/internal/ini` added at v1.8.6 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/internal/v4a` added at v1.4.21 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/internal/v4a` bumped to v1.4.22 (from v1.4.21) [PR 2669](https://github.com/provenance-io/provenance/pull/2669).
* `github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding` added at v1.13.7 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/service/internal/checksum` added at v1.9.12 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/service/internal/checksum` bumped to v1.9.13 (from v1.9.12) [PR 2669](https://github.com/provenance-io/provenance/pull/2669).
* `github.com/aws/aws-sdk-go-v2/service/internal/presigned-url` added at v1.13.20 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/service/internal/presigned-url` bumped to v1.13.21 (from v1.13.20) [PR 2669](https://github.com/provenance-io/provenance/pull/2669).
* `github.com/aws/aws-sdk-go-v2/service/internal/s3shared` added at v1.19.20 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/service/internal/s3shared` bumped to v1.19.21 (from v1.19.20) [PR 2669](https://github.com/provenance-io/provenance/pull/2669).
* `github.com/aws/aws-sdk-go-v2/service/s3` added at v1.97.1 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/service/s3` bumped to v1.97.3 (from v1.97.1) [PR 2669](https://github.com/provenance-io/provenance/pull/2669).
* `github.com/aws/aws-sdk-go-v2/service/signin` added at v1.0.8 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/service/ssooidc` added at v1.35.17 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/service/sso` added at v1.30.13 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2/service/sts` added at v1.41.9 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2` added at v1.41.4 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/aws-sdk-go-v2` bumped to v1.41.5 (from v1.41.4) [PR 2669](https://github.com/provenance-io/provenance/pull/2669).
* `github.com/aws/aws-sdk-go` bumped to v1.49.0 (from v1.44.224) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/aws/aws-sdk-go` removed at v1.49.0 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/aws/smithy-go` added at v1.24.2 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/bgentry/speakeasy` bumped to v0.2.0 (from v0.1.1-0.20220910012023-760eaf8b6816) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/bits-and-blooms/bitset` bumped to v1.24.3 (from v1.13.0) [#2592](https://github.com/provenance-io/provenance/issues/2592).
* `github.com/bytedance/gopkg` added at v0.1.3 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/bytedance/sonic/loader` bumped to v0.4.0 (from v0.3.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/bytedance/sonic` bumped to v1.14.2 (from v1.14.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/cenkalti/backoff/v4` bumped to v4.3.0 (from v4.2.1) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/cloudwego/base64x` bumped to v0.1.6 (from v0.1.5) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/cncf/xds/go` added at v0.0.0-20250501225837-2ac532fd4443 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/cncf/xds/go` bumped to v0.0.0-20251210132809-ee656c7534f5 (from v0.0.0-20250501225837-2ac532fd4443) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/cockroachdb/errors` bumped to v1.12.0 (from v1.11.3) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/cockroachdb/fifo` bumped to v0.0.0-20240616162244-4768e80dfb9a (from v0.0.0-20240606204812-0bbfbd93a7ce) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/cockroachdb/logtags` bumped to v0.0.0-20241215232642-bb51bb14a506 (from v0.0.0-20230118201751-21c54148d20b) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/cockroachdb/redact` bumped to v1.1.6 (from v1.1.5) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/cometbft/cometbft` bumped to v0.38.21 (from v0.38.19) ([PR 2576](https://github.com/provenance-io/provenance/pull/2576), [PR 2602](https://github.com/provenance-io/provenance/pull/2602)).
* `github.com/cosmos/cosmos-sdk` bumped to v0.53.6 (from v0.50.14) but is still replaced by v0.53.5-pio-2 of `github.com/provenance-io/cosmos-sdk` ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [#2592](https://github.com/provenance-io/provenance/issues/2592), [PR 2627](https://github.com/provenance-io/provenance/pull/2627)).
* `github.com/cosmos/gogoproto` bumped to v1.7.2 (from v1.7.0) [PR 2514](https://github.com/provenance-io/provenance/pull/2514).
* `github.com/cosmos/iavl` bumped to v1.2.6 (from v1.2.2) but is still replaced by v1.2.6 of `github.com/cosmos/iavl` [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/cosmos/ibc-apps/modules/async-icq/v8` removed at v8.0.0-prov-1 of `github.com/provenance-io/ibc-apps/modules/async-icq/v8` [#2630](https://github.com/provenance-io/provenance/issues/2630).
* `github.com/cosmos/ibc-go/v8` removed at v8.6.1-pio-1 of `github.com/provenance-io/ibc-go/v8` [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/cosmos/ibc-go/v10` added at v10.5.0 [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/cosmos/ibc-go/v10` bumped to v10.5.1 (from v10.5.0) [PR 2667](https://github.com/provenance-io/provenance/pull/2667).
* `github.com/cosmos/ledger-cosmos-go` bumped to v1.0.0 (from v0.14.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627)).
* `github.com/danieljoos/wincred` bumped to v1.2.1 (from v1.2.0) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/desertbit/timer` bumped to v1.0.1 (from v0.0.0-20180107155436-c41aec40b27f) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/dgraph-io/ristretto` bumped to v0.2.0 (from v0.1.2-0.20240116140435-c67e07994f91) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/dvsekhvalnov/jose2go` bumped to v1.7.0 (from v1.6.0) [PR 2538](https://github.com/provenance-io/provenance/pull/2538).
* `github.com/envoyproxy/go-control-plane/envoy` added at v1.32.4 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/envoyproxy/go-control-plane/envoy` bumped to v1.36.0 (from v1.32.4) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/envoyproxy/protoc-gen-validate` added at v1.2.1 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/envoyproxy/protoc-gen-validate` bumped to v1.3.0 (from v1.2.1) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/ethereum/go-ethereum` added at v1.17.0 [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/fatih/color` bumped to v1.18.0 (from v1.17.0) [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/getsentry/sentry-go` bumped to v0.42.0 (from v0.27.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627)).
* `github.com/golang/groupcache` bumped to v0.0.0-20241129210726-2c02b8208cf8 (from v0.0.0-20210331224755-41bb18bfe9da) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/golang/mock` removed at v1.6.0 [#2592](https://github.com/provenance-io/provenance/issues/2592).
* `github.com/golang/snappy` bumped to v1.0.0 (from v0.0.5-0.20220116011046-fa5810519dcb) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/googleapis/enterprise-certificate-proxy` bumped to v0.3.14 (from v0.3.3) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `github.com/googleapis/gax-go/v2` bumped to v2.17.0 (from v2.13.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `github.com/google/flatbuffers` bumped to v24.3.25+incompatible (from v2.0.8+incompatible) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/google/s2a-go` bumped to v0.1.9 (from v0.1.8) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/go-jose/go-jose/v4` added at v4.1.1 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/go-jose/go-jose/v4` bumped to v4.1.4 (from v4.1.1) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2661](https://github.com/provenance-io/provenance/pull/2661)).
* `github.com/hashicorp/aws-sdk-go-base/v2` added at v2.0.0-beta.72 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/hashicorp/go-getter` bumped to v1.8.6 (from v1.7.9) [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/hashicorp/go-hclog` bumped to v1.6.3 (from v1.5.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/hashicorp/go-plugin` bumped to v1.6.3 (from v1.6.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/hashicorp/go-safetemp` removed at v1.0.0 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/hashicorp/go-version` bumped to v1.8.0 (from v1.7.0) [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/hashicorp/yamux` bumped to v0.1.2 (from v0.1.1) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/hdevalence/ed25519consensus` bumped to v0.2.0 (from v0.1.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/holiman/uint256` added at v1.3.2 [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/huandu/skiplist` bumped to v1.2.1 (from v1.2.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/jmespath/go-jmespath` removed at v0.4.0 [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/klauspost/compress` bumped to v1.18.5 (from v1.17.11) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `github.com/mdp/qrterminal/v3` added at v3.2.1 [#2592](https://github.com/provenance-io/provenance/issues/2592).
* `github.com/mitchellh/go-testing-interface` removed at v1.14.1 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/planetscale/vtprotobuf` added at v0.6.1-0.20240319094008-0393e58bdf10 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/prometheus/client_golang` bumped to v1.23.2 (from v1.21.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627)).
* `github.com/prometheus/client_model` bumped to v0.6.2 (from v0.6.1) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/prometheus/common` bumped to v0.67.5 (from v0.62.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627)).
* `github.com/prometheus/procfs` bumped to v0.19.2 (from v0.15.1) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627)).
* `github.com/provlabs/vault` bumped to v1.0.15 (from v1.0.13) ([PR 2590](https://github.com/provenance-io/provenance/pull/2590), [PR 2620](https://github.com/provenance-io/provenance/pull/2620)).
* `github.com/rogpeppe/go-internal` bumped to v1.14.1 (from v1.13.1) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/rs/zerolog` bumped to v1.35.0 (from v1.34.0) [PR 2657](https://github.com/provenance-io/provenance/pull/2657).
* `github.com/shamaton/msgpack/v2` bumped to v2.2.3 (from v2.2.0) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/spf13/cobra` bumped to v1.10.2 (from v1.10.1) [PR 2555](https://github.com/provenance-io/provenance/pull/2555).
* `github.com/spiffe/go-spiffe/v2` added at v2.5.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/spiffe/go-spiffe/v2` bumped to v2.6.0 (from v2.5.0) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/tidwall/btree` bumped to v1.8.1 (from v1.7.0) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/ulikunitz/xz` bumped to v0.5.15 (from v0.5.14) [PR 2668](https://github.com/provenance-io/provenance/pull/2668).
* `github.com/zeebo/errs` added at v1.4.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/zeebo/errs` removed at v1.4.0 [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `github.com/zondax/golem` added at v0.27.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `github.com/zondax/ledger-go` bumped to v1.0.1 (from v0.14.3) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `golangci-lint` bumped to v2.1.6 (from v2.1.6) ([#2750](https://github.com/provenance-io/provenance/issues/2750), [#2346](https://github.com/provenance-io/provenance/issues/2346)).
* `golangci/golangci-lint-action` bumped to 8 (from 6) [#2346](https://github.com/provenance-io/provenance/issues/2346).
* `golang.org/x/crypto` bumped to v0.49.0 (from v0.40.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `golang.org/x/exp` bumped to v0.0.0-20250305212735-054e65f0b394 (from v0.0.0-20240904232852-e7e105dedf7e) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `golang.org/x/mod` bumped to v0.34.0 (from v0.26.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668), [PR 2666](https://github.com/provenance-io/provenance/pull/2666)).
* `golang.org/x/net` bumped to v0.52.0 (from v0.42.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `golang.org/x/oauth2` bumped to v0.36.0 (from v0.30.0) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `golang.org/x/sync` bumped to v0.20.0 (from v0.16.0) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `golang.org/x/sys` bumped to v0.42.0 (from v0.34.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `golang.org/x/term` bumped to v0.41.0 (from v0.33.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `golang.org/x/text` bumped to v0.36.0 (from v0.28.0) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668), [PR 2666](https://github.com/provenance-io/provenance/pull/2666)).
* `golang.org/x/time` bumped to v0.15.0 (from v0.6.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `golang.org/x/tools` bumped to v0.43.0 (from v0.35.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668), [PR 2666](https://github.com/provenance-io/provenance/pull/2666)).
* `google.golang.org/api` bumped to v0.271.0 (from v0.196.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20260203192932-546029d2fa20 (from v0.0.0-20250707201910-8d1bb00bc6a7) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2658](https://github.com/provenance-io/provenance/pull/2658), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `google.golang.org/genproto/googleapis/rpc` bumped to v0.0.0-20260226221140-a57be14db171 (from v0.0.0-20250707201910-8d1bb00bc6a7) ([PR 2514](https://github.com/provenance-io/provenance/pull/2514), [PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2658](https://github.com/provenance-io/provenance/pull/2658), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `google.golang.org/genproto` bumped to v0.0.0-20260128011058-8636f8732409 (from v0.0.0-20240903143218-8af14fe29dc1) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `google.golang.org/grpc` bumped to v1.80.0 (from v1.75.1) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2658](https://github.com/provenance-io/provenance/pull/2658)).
* `google.golang.org/protobuf` bumped to v1.36.11 (from v1.36.10) [PR 2574](https://github.com/provenance-io/provenance/pull/2574).
* `gotest.tools/v3` bumped to v3.5.2 (from v3.5.1) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `go.etcd.io/bbolt` bumped to v1.4.0-alpha.1 (from v1.4.0-alpha.0.0.20240404170359-43604f3112c5) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `go.opentelemetry.io/auto/sdk` bumped to v1.2.1 (from v1.1.0) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `go.opentelemetry.io/contrib/detectors/gcp` added at v1.36.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `go.opentelemetry.io/contrib/detectors/gcp` bumped to v1.39.0 (from v1.36.0) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc` bumped to v0.63.0 (from v0.54.0) ([PR 2620](https://github.com/provenance-io/provenance/pull/2620), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` bumped to v0.62.0 (from v0.54.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `go.opentelemetry.io/otel/metric` bumped to v1.42.0 (from v1.37.0) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `go.opentelemetry.io/otel/sdk/metric` added at v1.37.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `go.opentelemetry.io/otel/sdk/metric` bumped to v1.42.0 (from v1.37.0) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `go.opentelemetry.io/otel/sdk` added at v1.37.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `go.opentelemetry.io/otel/sdk` bumped to v1.42.0 (from v1.37.0) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `go.opentelemetry.io/otel/trace` bumped to v1.42.0 (from v1.37.0) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `go.opentelemetry.io/otel` bumped to v1.42.0 (from v1.37.0) ([PR 2627](https://github.com/provenance-io/provenance/pull/2627), [PR 2668](https://github.com/provenance-io/provenance/pull/2668)).
* `go.uber.org/atomic` removed at v1.10.0 [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `go.uber.org/mock` added at v0.6.0 [#2592](https://github.com/provenance-io/provenance/issues/2592).
* `go.uber.org/zap` bumped to v1.27.0 (from v1.24.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `go.yaml.in/yaml/v2` bumped to v2.4.3 (from v2.4.2) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `heighliner` bumped to v1.7.5 (from v1.7.0) [PR 2632](https://github.com/provenance-io/provenance/pull/2632).
* `nhooyr.io/websocket` bumped to v1.8.17 (from v1.8.10) [PR 2627](https://github.com/provenance-io/provenance/pull/2627).
* `peter-evans/create-pull-request` bumped to 8.1.1 (from 7.0.8) ([PR 2565](https://github.com/provenance-io/provenance/pull/2565), [PR 2597](https://github.com/provenance-io/provenance/pull/2597), [PR 2671](https://github.com/provenance-io/provenance/pull/2671)).
* `pgregory.net/rapid` bumped to v1.2.0 (from v1.1.0) [PR 2620](https://github.com/provenance-io/provenance/pull/2620).
* `rsc.io/qr` added at v0.2.0 [#2592](https://github.com/provenance-io/provenance/issues/2592).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.27.2...v1.28.0-rc1

