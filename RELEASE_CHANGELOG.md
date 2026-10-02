## [v1.31.0](https://github.com/provenance-io/provenance/releases/tag/v1.31.0) 2026-10-02

Provenance Blockchain version `v1.31.0` contains some exciting new features, improvements and bug fixes.

### RHEL 8 / glibc < 2.30: Replacement `libwasmvm.x86_64.so`

The `libwasmvm.x86_64.so` in `provenance-linux-amd64-v1.31.0.zip` comes from wasmvm v3.0.8, and it requires glibc 2.30 or newer (previous releases only needed 2.28). If your system has an older glibc, `provenanced` will fail to start. This affects RHEL 8.x and other EL8 distros (Rocky Linux 8, AlmaLinux 8, Oracle Linux 8), and any other Linux with glibc older than 2.30.

**You are affected if** `ldd --version` reports a glibc older than 2.30, or `provenanced` fails with an error like:

```
provenanced: /lib64/libc.so.6: version `GLIBC_2.30' not found (required by .../libwasmvm.x86_64.so)
provenanced: /lib64/libm.so.6: version `GLIBC_2.29' not found (required by .../libwasmvm.x86_64.so)
```

If you're affected, use the `libwasmvm.x86_64.so` attached to this release instead of the one in the zip. It's built from the same wasmvm v3.0.8 source with the same Rust version (1.95.0), but on Rocky Linux 8, so it only requires glibc 2.28. It's state-compatible with the standard library. The only difference is the glibc version it's linked against.

To use it:

1. Download `libwasmvm.x86_64.so` and `libwasmvm.x86_64.so.sha256` from this release, and verify the file:
   ```
   sha256sum -c libwasmvm.x86_64.so.sha256
   ```
   Expected sha256: `aa735c44bee2f0a31d93657a76856a72a8f1dcc50edbcf687d8fd6831ad67429`
2. Put the `provenanced` from `provenance-linux-amd64-v1.31.0.zip` in your upgrade directory as usual (e.g. `cosmovisor/upgrades/geranium/bin/`). Then **replace** the `libwasmvm.x86_64.so` next to it with the downloaded one. `provenanced` loads the library from its own directory first, so no `LD_LIBRARY_PATH` changes are needed.
3. If you use cosmovisor, set `DAEMON_ALLOW_DOWNLOAD_BINARIES=false` and put the binaries in place yourself before the upgrade height. Otherwise cosmovisor's automatic download gets the standard zip, which won't run on your system.
4. Check it: `provenanced query wasm libwasmvm-version` should print `3.0.8`.

If your system has glibc 2.30 or newer, you don't need this; use the standard zip as-is. Only use this library with `provenanced` builds that use wasmvm v3.0.8.

The build script is `scripts/build-libwasmvm-el8.sh`. The underlying issue has been reported upstream: https://github.com/CosmWasm/cosmwasm/issues/2710.

### Features

* Migrated name module from kv-store to collections [#2411](https://github.com/provenance-io/provenance/issues/2411).
* Use hashed name for key instead of just the segments [#2683](https://github.com/provenance-io/provenance/issues/2683).
* Add geranium and geranium-rc1 upgrades [#2832](https://github.com/provenance-io/provenance/issues/2832).

### Improvements

* Remove the quarantine module [#2695](https://github.com/provenance-io/provenance/issues/2695).
* Use an arm runner to build the arm docker images [PR 2833](https://github.com/provenance-io/provenance/pull/2833).
* Improve required attribute checking (in the marker and exchange modules) to make it read less from state [PR 2843](https://github.com/provenance-io/provenance/pull/2843).
* Reduce the number of expired attributes that can be deleted per block to 500 (from 100,000) [PR 2850](https://github.com/provenance-io/provenance/pull/2850).
* Store name records with keys that use the name with segments reversed instead of hashing the name [#2851](https://github.com/provenance-io/provenance/issues/2851).
* Do not allow the denoms to change when updating the flatfees params or conversion factor [PR 2854](https://github.com/provenance-io/provenance/pull/2854).
* Update the flatfees conversion factor [PR 2857](https://github.com/provenance-io/provenance/pull/2857).

### Bug Fixes

* Fix the heighliner docker build GitHub action [PR 2833](https://github.com/provenance-io/provenance/pull/2833).
* When setting an attribute, delete any previously recorded expiration entry [PR 2842](https://github.com/provenance-io/provenance/pull/2842).
* Ensure attributes are actually expired before deleting them [PR 2842](https://github.com/provenance-io/provenance/pull/2842).
* Prevent a smart contract that only holds an `x/authz` grant from one `x/metadata` party from vouching for the signatures of the other parties listed after it [PR 2848](https://github.com/provenance-io/provenance/pull/2848).
* Fix marker withdraw to not panic if there's no to address [PR 2849](https://github.com/provenance-io/provenance/pull/2849).

### Api Breaking

* The previously deprecated quarantine module has been fully removed [#2695](https://github.com/provenance-io/provenance/issues/2695).

### Dependencies

* `actions/download-artifact` bumped to 8 (from 4) [PR 2837](https://github.com/provenance-io/provenance/pull/2837).
* `actions/setup-java` bumped to 6 (from 5) [PR 2828](https://github.com/provenance-io/provenance/pull/2828).
* `actions/upload-artifact` bumped to 7 (from 4) [PR 2836](https://github.com/provenance-io/provenance/pull/2836).
* `cloud.google.com/go/auth` bumped to v0.20.0 (from v0.18.2) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `github.com/CosmWasm/wasmd` bumped to v0.61.15-pio-1 of `github.com/provenance-io/wasmd` (from v0.61.10-pio-2 of `github.com/provenance-io/wasmd`) [PR 2856](https://github.com/provenance-io/provenance/pull/2856).
* `github.com/CosmWasm/wasmvm/v3` bumped to v3.0.8 (from v3.0.7) [PR 2856](https://github.com/provenance-io/provenance/pull/2856).
* `github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp` bumped to v1.34.0 (from v1.33.0) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `github.com/cosmos/cosmos-sdk` bumped to v0.53.8-pio-2 of `github.com/provenance-io/cosmos-sdk` (from v0.53.8-pio-1 of `github.com/provenance-io/cosmos-sdk`) [PR 2839](https://github.com/provenance-io/provenance/pull/2839).
* `github.com/felixge/httpsnoop` bumped to v1.1.0 (from v1.0.4) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `github.com/googleapis/enterprise-certificate-proxy` bumped to v0.3.15 (from v0.3.14) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `github.com/googleapis/gax-go/v2` bumped to v2.22.0 (from v2.17.0) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `github.com/go-logr/logr` bumped to v1.4.4 (from v1.4.3) [PR 2844](https://github.com/provenance-io/provenance/pull/2844).
* `github.com/hashicorp/go-metrics` bumped to v0.7.0 (from v0.6.1) [PR 2846](https://github.com/provenance-io/provenance/pull/2846).
* `github.com/pmezard/go-difflib` removed at v1.0.1-0.20181226105442-5d4384ee4fb2 [PR 2825](https://github.com/provenance-io/provenance/pull/2825).
* `github.com/prometheus/client_model` bumped to v0.6.3 (from v0.6.2) [PR 2846](https://github.com/provenance-io/provenance/pull/2846).
* `github.com/prometheus/common` bumped to v0.71.0 (from v0.70.1) [PR 2846](https://github.com/provenance-io/provenance/pull/2846).
* `github.com/spiffe/go-spiffe/v2` bumped to v2.8.1 (from v2.7.0) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `github.com/stretchr/objx` bumped to v0.5.3 (from v0.5.2) [PR 2825](https://github.com/provenance-io/provenance/pull/2825).
* `github.com/stretchr/testify` bumped to v1.12.1 (from v1.11.1) [PR 2825](https://github.com/provenance-io/provenance/pull/2825).
* `github/codeql-action` bumped to 4.38.1 (from 4.37.6) ([PR 2830](https://github.com/provenance-io/provenance/pull/2830), [PR 2841](https://github.com/provenance-io/provenance/pull/2841), [PR 2847](https://github.com/provenance-io/provenance/pull/2847)).
* `golang.org/x/crypto` bumped to v0.55.0 (from v0.54.0) [PR 2829](https://github.com/provenance-io/provenance/pull/2829).
* `golang.org/x/mod` bumped to v0.38.0 (from v0.37.0) [PR 2829](https://github.com/provenance-io/provenance/pull/2829).
* `golang.org/x/net` bumped to v0.58.0 (from v0.57.0) [PR 2829](https://github.com/provenance-io/provenance/pull/2829).
* `golang.org/x/text` bumped to v0.41.0 (from v0.40.0) [PR 2829](https://github.com/provenance-io/provenance/pull/2829).
* `golang.org/x/tools` bumped to v0.48.0 (from v0.47.0) [PR 2829](https://github.com/provenance-io/provenance/pull/2829).
* `google.golang.org/api` bumped to v0.278.0 (from v0.271.0) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20260706201446-f0a921348800 (from v0.0.0-20260526163538-3dc84a4a5aaa) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `google.golang.org/genproto/googleapis/rpc` bumped to v0.0.0-20260706201446-f0a921348800 (from v0.0.0-20260526163538-3dc84a4a5aaa) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `google.golang.org/genproto` bumped to v0.0.0-20260319201613-d00831a3d3e7 (from v0.0.0-20260128011058-8636f8732409) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `google.golang.org/grpc` bumped to v1.84.0 (from v1.83.0) ([PR 2829](https://github.com/provenance-io/provenance/pull/2829), [PR 2845](https://github.com/provenance-io/provenance/pull/2845)).
* `google.golang.org/protobuf` bumped to v1.36.12 (from v1.36.11) [PR 2819](https://github.com/provenance-io/provenance/pull/2819).
* `go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc` bumped to v0.67.0 (from v0.63.0) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` bumped to v0.69.0 (from v0.62.0) [PR 2845](https://github.com/provenance-io/provenance/pull/2845).
* `go.opentelemetry.io/otel/metric` bumped to v1.45.0 (from v1.44.0) [PR 2844](https://github.com/provenance-io/provenance/pull/2844).
* `go.opentelemetry.io/otel/sdk/metric` bumped to v1.45.0 (from v1.44.0) [PR 2844](https://github.com/provenance-io/provenance/pull/2844).
* `go.opentelemetry.io/otel/sdk` bumped to v1.45.0 (from v1.44.0) [PR 2844](https://github.com/provenance-io/provenance/pull/2844).
* `go.opentelemetry.io/otel/trace` bumped to v1.45.0 (from v1.44.0) [PR 2844](https://github.com/provenance-io/provenance/pull/2844).
* `go.opentelemetry.io/otel` bumped to v1.45.0 (from v1.44.0) [PR 2844](https://github.com/provenance-io/provenance/pull/2844).
* `go.yaml.in/yaml/v3` bumped to v3.0.5 (from v3.0.4) [PR 2825](https://github.com/provenance-io/provenance/pull/2825).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.30.0...v1.31.0

