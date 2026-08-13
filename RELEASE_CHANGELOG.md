## [v1.30.0-rc2](https://github.com/provenance-io/provenance/releases/tag/v1.30.0-rc2) 2026-08-13

Provenance Blockchain version `v1.30.0` contains some exciting new features, improvements and bug fixes.

### Improvements

* Detect vault principal accounts, which are marker accounts, and exclude them from force transfers [PR 2804](https://github.com/provenance-io/provenance/pull/2804).
* Remove the marker all-supply permissions bypass [PR 2807](https://github.com/provenance-io/provenance/pull/2807).
* Set the flatfees for some vault messages [PR 2815](https://github.com/provenance-io/provenance/pull/2815).

### Bug Fixes

* Honor `require_deposit_access` when withdrawing to or depositing into an unrestricted marker account [PR 2809](https://github.com/provenance-io/provenance/pull/2809).

### Dependencies

* `docker/login-action` bumped to 4.6.0 (from 4.5.2) [PR 2802](https://github.com/provenance-io/provenance/pull/2802).
* `github.com/provlabs/vault` bumped to v1.2.4 (from v1.2.3) [PR 2811](https://github.com/provenance-io/provenance/pull/2811).
* `github/codeql-action` bumped to 4.37.6 (from 4.37.3) [PR 2806](https://github.com/provenance-io/provenance/pull/2806).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.30.0-rc1...v1.30.0-rc2
* https://github.com/provenance-io/provenance/compare/v1.29.0...v1.30.0-rc2

---

## [v1.30.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.30.0-rc1) 2026-07-31

Provenance Blockchain version `v1.30.0` contains some exciting new features, improvements and bug fixes.

### Features

* Switch to collections in the marker module [#2494](https://github.com/provenance-io/provenance/issues/2494).
* Add an opt-in `require_deposit_access` marker flag that enforces deposit-access control on coins sent into a marker regardless of marker type [#2769](https://github.com/provenance-io/provenance/issues/2769).

### Improvements

* Reduce race test runtime by excluding long-running simulation tests from race builds [#2738](https://github.com/provenance-io/provenance/issues/2738).
* Add the forsythia upgrade and remove older upgrades [PR 2780](https://github.com/provenance-io/provenance/pull/2780).
* Short circuit at the lookup step when deleting expired events [PR 2789](https://github.com/provenance-io/provenance/pull/2789).

### Bug Fixes

* Fix marker `ExportGenesis` dropping the `require_deposit_access` flag from exported markers [#2769](https://github.com/provenance-io/provenance/issues/2769).
* Enforce the unrestricted denom regex on markers created by the asset module [PR 2784](https://github.com/provenance-io/provenance/pull/2784).
* Propagate the AllowList when updating a MarkerTransferAuthorization [PR 2786](https://github.com/provenance-io/provenance/pull/2786).
* Verify smart contract signers in metadata's AddNetAssetValues endpoint [PR 2787](https://github.com/provenance-io/provenance/pull/2787).

### Deprecated

* Deprecate and deactivate the quarantine module  [#2695](https://github.com/provenance-io/provenance/issues/2695).
  There is no replacement for the functionality that the quarantine module provided.
  All quarantined funds will be sent to their associated accounts (i.e. all quarantined fund will be accepted).

### Dependencies

* `4d63.com/gocheckcompilerdirectives` bumped to v1.3.0 (from v1.2.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `4d63.com/gochecknoglobals` bumped to v0.2.2 (from v0.2.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `actions/checkout` bumped to 7 (from 6) [PR 2763](https://github.com/provenance-io/provenance/pull/2763).
* `actions/setup-go` bumped to 7 (from 6) [PR 2777](https://github.com/provenance-io/provenance/pull/2777).
* `cel.dev/expr` bumped to v0.25.2 (from v0.25.1) [PR 2796](https://github.com/provenance-io/provenance/pull/2796).
* `charm.land/lipgloss/v2` added at v2.0.3 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `codeberg.org/chavacava/garif` added at v0.2.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `codeberg.org/polyfloyd/go-errorlint` added at v1.9.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `codecov/codecov-action` bumped to 7 (from 6) [PR 2756](https://github.com/provenance-io/provenance/pull/2756).
* `dev.gaijin.team/go/exhaustruct/v4` added at v4.0.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `dev.gaijin.team/go/golib` added at v0.6.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `docker/login-action` bumped to 4.5.2 (from 4) [PR 2797](https://github.com/provenance-io/provenance/pull/2797).
* `github.com/4meepo/tagalign` bumped to v1.4.3 (from v1.3.4) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/Abirdcfly/dupword` bumped to v0.1.7 (from v0.1.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/AdminBenni/iota-mixing` added at v1.0.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/AlwxSin/noinlineerr` added at v1.0.5 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/Antonboom/errname` bumped to v1.1.1 (from v0.1.13) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/Antonboom/nilnil` bumped to v1.1.1 (from v0.1.9) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/Antonboom/testifylint` bumped to v1.6.4 (from v1.4.3) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/BurntSushi/toml` bumped to v1.6.0 (from v1.4.1-0.20240526193622-a339e1f7089c) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ClickHouse/clickhouse-go-linter` added at v1.2.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/CosmWasm/wasmvm/v3` bumped to v3.0.7 (from v3.0.5) ([PR 2740](https://github.com/provenance-io/provenance/pull/2740), [PR 2765](https://github.com/provenance-io/provenance/pull/2765)).
* `github.com/Crocmagnon/fatcontext` removed at v0.5.2 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/Djarvur/go-err113` bumped to v0.1.1 (from v0.0.0-20210108212216-aea10b59be24) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/GaijinEntertainment/go-exhaustruct/v3` removed at v3.3.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp` bumped to v1.33.0 (from v1.31.0) ([PR 2778](https://github.com/provenance-io/provenance/pull/2778), [PR 2796](https://github.com/provenance-io/provenance/pull/2796)).
* `github.com/Masterminds/semver/v3` bumped to v3.5.0 (from v3.3.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/MirrexOne/unqueryvet` added at v1.5.4 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/OpenPeeDeeP/depguard/v2` bumped to v2.2.1 (from v2.2.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/alecthomas/chroma/v2` added at v2.24.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/alecthomas/go-check-sumtype` bumped to v0.3.1 (from v0.1.4) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/alexkohler/nakedret/v2` bumped to v2.0.6 (from v2.0.4) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/alexkohler/prealloc` bumped to v1.1.0 (from v1.0.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/alfatraining/structtag` added at v1.0.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/alingse/nilnesserr` added at v0.2.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ashanbrown/forbidigo/v2` added at v2.3.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ashanbrown/forbidigo` removed at v1.6.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ashanbrown/makezero/v2` added at v2.2.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ashanbrown/makezero` removed at v1.1.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/bits-and-blooms/bitset` bumped to v1.24.4 (from v1.24.3) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/bkielbasa/cyclop` bumped to v1.2.3 (from v1.2.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/bombsimon/wsl/v4` bumped to v4.7.0 (from v4.4.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/bombsimon/wsl/v5` added at v5.8.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/breml/bidichk` bumped to v0.3.3 (from v0.2.7) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/breml/errchkjson` bumped to v0.4.1 (from v0.3.6) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/butuzov/ireturn` bumped to v0.4.1 (from v0.3.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/butuzov/mirror` bumped to v1.3.0 (from v1.2.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/catenacyber/perfsprint` bumped to v0.10.1 (from v0.7.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ccojocar/zxcvbn-go` bumped to v1.0.4 (from v1.0.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/charithe/durationcheck` bumped to v0.0.11 (from v0.0.10) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/charmbracelet/colorprofile` added at v0.4.3 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/charmbracelet/ultraviolet` added at v0.0.0-20251205161215-1948445e3318 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/charmbracelet/x/ansi` added at v0.11.7 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/charmbracelet/x/termios` added at v0.1.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/charmbracelet/x/term` added at v0.2.2 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/charmbracelet/x/windows` added at v0.2.2 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/chavacava/garif` removed at v0.1.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ckaznocha/intrange` bumped to v0.3.1 (from v0.2.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/clipperhouse/displaywidth` added at v0.11.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/clipperhouse/uax29/v2` added at v2.7.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/cosmos/cosmos-sdk` bumped to v0.53.8-pio-1 of `github.com/provenance-io/cosmos-sdk` (from v0.53.5-pio-2 of `github.com/provenance-io/cosmos-sdk`) [PR 2792](https://github.com/provenance-io/provenance/pull/2792).
* `github.com/curioswitch/go-reassign` bumped to v0.3.0 (from v0.2.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/daixiang0/gci` bumped to v0.13.7 (from v0.13.5) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/dave/dst` added at v0.27.3 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/dlclark/regexp2` added at v1.12.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/fatih/color` bumped to v1.19.0 (from v1.18.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/firefart/nonamedreturns` bumped to v1.0.6 (from v1.0.5) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ghostiam/protogetter` bumped to v0.3.20 (from v0.3.6) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/godoc-lint/godoc-lint` added at v0.11.2 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/gofrs/flock` bumped to v0.13.0 (from v0.12.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/asciicheck` added at v0.5.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/dupl` bumped to v0.0.0-20260401084720-c99c5cf5c202 (from v0.0.0-20180902072040-3e9179ac440a) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/gofmt` bumped to v0.0.0-20250106114630-d62b90e6713d (from v0.0.0-20240816233607-d8596aa466a9) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/golangci-lint/v2` added at v2.12.2 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/golangci-lint` removed at v1.61.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/golines` added at v0.15.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/go-printf-func-name` added at v0.1.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/misspell` bumped to v0.8.0 (from v0.6.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/modinfo` removed at v0.3.4 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/plugin-module-register` bumped to v0.1.2 (from v0.1.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/revgrep` bumped to v0.8.0 (from v0.5.3) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/rowserrcheck` added at v0.0.0-20260419091836-c5f79b8a11ba [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/swaggoswag` added at v0.0.0-20250504205917-77f2aca3143e [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/golangci/unconvert` bumped to v0.0.0-20250410112200-a129a6e6413e (from v0.0.0-20240309020433-c5143eacb3ed) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/gordonklaus/ineffassign` bumped to v0.2.0 (from v0.1.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/gostaticanalysis/comment` bumped to v1.5.0 (from v1.4.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/gostaticanalysis/forcetypeassert` bumped to v0.2.0 (from v0.1.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/gostaticanalysis/nilerr` bumped to v0.1.2 (from v0.1.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/go-critic/go-critic` bumped to v0.14.3 (from v0.11.4) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/go-viper/mapstructure/v2` bumped to v2.5.0 (from v2.4.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/go-xmlfmt/xmlfmt` bumped to v1.1.3 (from v1.1.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/hashicorp/go-immutable-radix/v2` added at v2.1.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/hashicorp/go-metrics` bumped to v0.6.1 (from v0.5.4) ([PR 2764](https://github.com/provenance-io/provenance/pull/2764), [PR 2795](https://github.com/provenance-io/provenance/pull/2795)).
* `github.com/hashicorp/go-version` bumped to v1.9.0 (from v1.8.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/jgautheron/goconst` bumped to v1.10.0 (from v1.7.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/jingyugao/rowserrcheck` removed at v1.1.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/jirfag/go-printf-func-name` removed at v0.0.0-20200119135958-7558a9eaa5af [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/jjti/go-spancheck` bumped to v0.6.5 (from v0.6.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/julz/importas` bumped to v0.2.0 (from v0.1.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/karamaru-alpha/copyloopvar` bumped to v1.2.2 (from v1.1.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/kisielk/errcheck` bumped to v1.10.0 (from v1.7.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/kkHAIKE/contextcheck` bumped to v1.1.6 (from v1.1.5) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/klauspost/compress` bumped to v1.19.1 (from v1.18.5) [PR 2795](https://github.com/provenance-io/provenance/pull/2795).
* `github.com/kulti/thelper` bumped to v0.7.1 (from v0.6.3) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/kunwardeep/paralleltest` bumped to v1.0.15 (from v1.0.10) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/kyoh86/exportloopref` removed at v0.1.11 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/lasiar/canonicalheader` bumped to v1.1.2 (from v1.1.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ldez/exptostd` added at v0.4.5 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ldez/gomoddirectives` bumped to v0.8.0 (from v0.2.4) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ldez/grignotin` added at v0.10.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ldez/structtags` added at v0.6.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ldez/tagliatelle` bumped to v0.7.2 (from v0.5.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ldez/usetesting` added at v0.5.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/lib/pq` bumped to v1.12.3 (from v1.12.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/lucasb-eyer/go-colorful` added at v1.4.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/lufeee/execinquery` removed at v1.2.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/macabu/inamedparam` bumped to v0.2.0 (from v0.1.3) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/manuelarte/embeddedstructfieldcheck` added at v0.4.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/manuelarte/funcorder` added at v0.6.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/maratori/testableexamples` bumped to v1.0.1 (from v1.0.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/maratori/testpackage` bumped to v1.1.2 (from v1.1.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/matoous/godox` bumped to v1.1.0 (from v0.0.0-20230222163458-006bad1f9d26) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/mattn/go-runewidth` bumped to v0.0.23 (from v0.0.13) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/mgechev/revive` bumped to v1.15.0 (from v1.3.9) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/muesli/cancelreader` added at v0.2.2 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/nunnatsa/ginkgolinter` bumped to v0.23.0 (from v0.16.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/olekukonko/tablewriter` removed at v0.0.5 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/pelletier/go-toml/v2` bumped to v2.3.1 (from v2.2.4) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/polyfloyd/go-errorlint` removed at v1.6.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/prometheus/client_golang` bumped to v1.24.1 (from v1.23.2) [PR 2795](https://github.com/provenance-io/provenance/pull/2795).
* `github.com/prometheus/common` bumped to v0.70.1 (from v0.67.5) [PR 2795](https://github.com/provenance-io/provenance/pull/2795).
* `github.com/prometheus/procfs` bumped to v0.21.1 (from v0.19.2) [PR 2795](https://github.com/provenance-io/provenance/pull/2795).
* `github.com/provlabs/vault` bumped to v1.2.3 (from v1.1.0) ([PR 2779](https://github.com/provenance-io/provenance/pull/2779), [PR 2790](https://github.com/provenance-io/provenance/pull/2790)).
* `github.com/quasilyte/go-ruleguard/dsl` bumped to v0.3.23 (from v0.3.22) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/quasilyte/go-ruleguard` bumped to v0.4.5 (from v0.4.3-0.20240823090925-0fe6f58b47b1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/raeperd/recvcheck` added at v0.2.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/rivo/uniseg` bumped to v0.4.7 (from v0.2.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ryancurrah/gomodguard/v2` added at v2.1.3 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ryancurrah/gomodguard` bumped to v1.4.1 (from v1.3.5) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ryanrolds/sqlclosecheck` bumped to v0.6.0 (from v0.5.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/sanposhiho/wastedassign/v2` bumped to v2.1.0 (from v2.0.7) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/santhosh-tekuri/jsonschema/v5` removed at v5.3.1 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/santhosh-tekuri/jsonschema/v6` added at v6.0.2 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/sashamelentyev/usestdlibvars` bumped to v1.29.0 (from v1.27.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/securego/gosec/v2` bumped to v2.26.1 (from v2.21.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/shamaton/msgpack/v2` bumped to v2.4.1 (from v2.2.3) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/shazow/go-diff` removed at v0.0.0-20160112020656-b6b7b6733b8c [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/sirupsen/logrus` bumped to v1.9.4 (from v1.9.3) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/sivchari/tenv` removed at v1.10.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/sonatard/noctx` bumped to v0.5.1 (from v0.0.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/sourcegraph/go-diff` bumped to v0.8.0 (from v0.7.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/spiffe/go-spiffe/v2` bumped to v2.7.0 (from v2.6.0) [PR 2796](https://github.com/provenance-io/provenance/pull/2796).
* `github.com/stbenjam/no-sprintf-host-port` bumped to v0.3.1 (from v0.1.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/tdakkota/asciicheck` removed at v0.2.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/tetafro/godot` bumped to v1.5.6 (from v1.4.17) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/timakin/bodyclose` bumped to v0.0.0-20260129054331-73d1f95b84b4 (from v0.0.0-20230421092635-574207250966) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/timonwong/loggercheck` bumped to v0.11.0 (from v0.9.4) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/tomarrell/wrapcheck/v2` bumped to v2.12.0 (from v2.9.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ultraware/funlen` bumped to v0.2.0 (from v0.1.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/ultraware/whitespace` bumped to v0.2.0 (from v0.1.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/uudashr/gocognit` bumped to v1.2.1 (from v1.1.3) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/uudashr/iface` added at v1.4.2 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/xen0n/gosmopolitan` bumped to v1.3.0 (from v1.2.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github.com/xo/terminfo` added at v0.0.0-20220910002029-abceb7e1c41e [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `github/codeql-action` bumped to 4.37.3 (from 4) [PR 2798](https://github.com/provenance-io/provenance/pull/2798).
* `golang.org/x/crypto` bumped to v0.54.0 (from v0.50.0) ([PR 2759](https://github.com/provenance-io/provenance/pull/2759), [PR 2776](https://github.com/provenance-io/provenance/pull/2776), [PR 2775](https://github.com/provenance-io/provenance/pull/2775), [PR 2795](https://github.com/provenance-io/provenance/pull/2795)).
* `golang.org/x/exp/typeparams` bumped to v0.0.0-20260209203927-2842357ff358 (from v0.0.0-20240314144324-c7f7c6466f7f) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `golang.org/x/exp` bumped to v0.0.0-20250620022241-b7579e27df2b (from v0.0.0-20250305212735-054e65f0b394) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `golang.org/x/mod` bumped to v0.37.0 (from v0.35.0) ([PR 2759](https://github.com/provenance-io/provenance/pull/2759), [PR 2775](https://github.com/provenance-io/provenance/pull/2775)).
* `golang.org/x/net` bumped to v0.57.0 (from v0.53.0) ([PR 2759](https://github.com/provenance-io/provenance/pull/2759), [PR 2773](https://github.com/provenance-io/provenance/pull/2773), [PR 2775](https://github.com/provenance-io/provenance/pull/2775), [PR 2795](https://github.com/provenance-io/provenance/pull/2795)).
* `golang.org/x/sync` bumped to v0.22.0 (from v0.20.0) ([PR 2759](https://github.com/provenance-io/provenance/pull/2759), [PR 2775](https://github.com/provenance-io/provenance/pull/2775)).
* `golang.org/x/sys` bumped to v0.47.0 (from v0.43.0) ([PR 2759](https://github.com/provenance-io/provenance/pull/2759), [PR 2776](https://github.com/provenance-io/provenance/pull/2776), [PR 2775](https://github.com/provenance-io/provenance/pull/2775), [PR 2795](https://github.com/provenance-io/provenance/pull/2795)).
* `golang.org/x/term` bumped to v0.45.0 (from v0.42.0) ([PR 2759](https://github.com/provenance-io/provenance/pull/2759), [PR 2775](https://github.com/provenance-io/provenance/pull/2775), [PR 2795](https://github.com/provenance-io/provenance/pull/2795)).
* `golang.org/x/text` bumped to v0.40.0 (from v0.37.0) ([PR 2759](https://github.com/provenance-io/provenance/pull/2759), [PR 2775](https://github.com/provenance-io/provenance/pull/2775)).
* `golang.org/x/tools/go/expect` removed at v0.1.1-deprecated [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `golang.org/x/tools/go/packages/packagestest` removed at v0.1.1-deprecated [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `golang.org/x/tools` bumped to v0.47.0 (from v0.44.0) ([PR 2759](https://github.com/provenance-io/provenance/pull/2759), [PR 2775](https://github.com/provenance-io/provenance/pull/2775)).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20260526163538-3dc84a4a5aaa (from v0.0.0-20260226221140-a57be14db171) ([PR 2778](https://github.com/provenance-io/provenance/pull/2778), [PR 2796](https://github.com/provenance-io/provenance/pull/2796)).
* `google.golang.org/genproto/googleapis/rpc` bumped to v0.0.0-20260526163538-3dc84a4a5aaa (from v0.0.0-20260226221140-a57be14db171) ([PR 2778](https://github.com/provenance-io/provenance/pull/2778), [PR 2796](https://github.com/provenance-io/provenance/pull/2796)).
* `google.golang.org/grpc` bumped to v1.83.0 (from v1.81.1) ([PR 2778](https://github.com/provenance-io/provenance/pull/2778), [PR 2796](https://github.com/provenance-io/provenance/pull/2796)).
* `go-simpler.org/musttag` bumped to v0.14.0 (from v0.12.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `go-simpler.org/sloglint` bumped to v0.12.0 (from v0.7.2) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `go.augendre.info/arangolint` added at v0.4.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `go.augendre.info/fatcontext` added at v0.9.0 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `go.opentelemetry.io/contrib/detectors/gcp` bumped to v1.44.0 (from v1.42.0) ([PR 2778](https://github.com/provenance-io/provenance/pull/2778), [PR 2796](https://github.com/provenance-io/provenance/pull/2796)).
* `go.opentelemetry.io/otel/metric` bumped to v1.44.0 (from v1.43.0) [PR 2796](https://github.com/provenance-io/provenance/pull/2796).
* `go.opentelemetry.io/otel/sdk/metric` bumped to v1.44.0 (from v1.43.0) [PR 2796](https://github.com/provenance-io/provenance/pull/2796).
* `go.opentelemetry.io/otel/sdk` bumped to v1.44.0 (from v1.43.0) [PR 2796](https://github.com/provenance-io/provenance/pull/2796).
* `go.opentelemetry.io/otel/trace` bumped to v1.44.0 (from v1.43.0) [PR 2796](https://github.com/provenance-io/provenance/pull/2796).
* `go.opentelemetry.io/otel` bumped to v1.44.0 (from v1.43.0) [PR 2796](https://github.com/provenance-io/provenance/pull/2796).
* `go.uber.org/automaxprocs` removed at v1.5.3 [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `go.yaml.in/yaml/v2` bumped to v2.4.4 (from v2.4.3) [PR 2795](https://github.com/provenance-io/provenance/pull/2795).
* `honnef.co/go/tools` bumped to v0.7.0 (from v0.5.1) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `mvdan.cc/gofumpt` bumped to v0.9.2 (from v0.7.0) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).
* `mvdan.cc/unparam` bumped to v0.0.0-20251027182757-5beb8c8f8f15 (from v0.0.0-20240528143540-8a5130ca722f) [PR 2779](https://github.com/provenance-io/provenance/pull/2779).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.29.0...v1.30.0-rc1

