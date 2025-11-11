## [v1.26.0-rc2](https://github.com/provenance-io/provenance/releases/tag/v1.26.0-rc2) 2025-11-11

Provenance Blockchain version `v1.26.0` contains some exciting new features.

**Important**: All users should now use `1nhash` for `gas-prices` and a multiplier of `1.0` (the default).

Fees on Provenance Blockchain are now based on msg type instead of gas.
The standard Tx simulation process now returns the fee amount as gas wanted (i.e. it no longer reflects an actual gas amount).
By using `1nhash` for gas prices when simulating, existing client(s) will properly set the fee for the Tx.

The new `x/flatfees` module manages the costs of each Msg type.
Costs are defined in milli-US-dollars (musd).
A conversion factor (defined in module params) is used to determine the amount of nhash equivalent to the musd cost of a Msg.
This conversion factor will be constant to start, but can be updated manually (via governance proposal) or in the future, might update automatically.
By keeping the conversion factor in-line with the cost of hash, fees will remain constant in terms of how much they cost in US dollars (even though the required amounts of hash might change).

### Improvements

* Recognize 9525nhash as old gas prices too [PR 2517](https://github.com/provenance-io/provenance/pull/2517).
* Fix how the ledger bulk endpoint charges fees [PR 2518](https://github.com/provenance-io/provenance/pull/2518).
* Emit events when a ledger class or class type is created, and when a ledger entry is updated [PR 2522](https://github.com/provenance-io/provenance/pull/2522).
* Add _NOT_DEFINED values to the ledger enums [PR 2528](https://github.com/provenance-io/provenance/pull/2528).

### Bug Fixes

* Bring back (but deprecate) some msgfees proto stuff [PR 2523](https://github.com/provenance-io/provenance/pull/2523).
* In the bouvardia-rc2 upgrade, fix the ledger class and the registry entries added with -rc1 [PR 2524](https://github.com/provenance-io/provenance/pull/2524).

### Dependencies

* `github.com/provlabs/vault` bumped to v1.0.12 (from v1.0.9) [PR 2525](https://github.com/provenance-io/provenance/pull/2525).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.26.0-rc1...v1.26.0-rc2
* https://github.com/provenance-io/provenance/compare/v1.25.1...v1.26.0-rc2

---

## [v1.26.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.26.0-rc1) 2025-10-29

Provenance Blockchain version `v1.26.0` contains some exciting new features.

**Important**: All users should now use `1nhash` for `gas-prices` and a multiplier of `1.0` (the default).

Fees on Provenance Blockchain are now based on msg type instead of gas.
The standard Tx simulation process now returns the fee amount as gas wanted (i.e. it no longer reflects an actual gas amount).
By using `1nhash` for gas prices when simulating, existing client(s) will properly set the fee for the Tx.

The new `x/flatfees` module manages the costs of each Msg type.
Costs are defined in milli-US-dollars (musd).
A conversion factor (defined in module params) is used to determine the amount of nhash equivalent to the musd cost of a Msg.
This conversion factor will be constant to start, but can be updated manually (via governance proposal) or in the future, might update automatically.
By keeping the conversion factor in-line with the cost of hash, fees will remain constant in terms of how much they cost in US dollars (even though the required amounts of hash might change).

### Features

* Publish Docker images to GitHub Container Registry (GHCR) [#904](https://github.com/provenance-io/provenance/issues/904).
* Enabled `module_query_safe` option on deterministic queries [#2005](https://github.com/provenance-io/provenance/issues/2005).
* Support for flexible party types added [#2149](https://github.com/provenance-io/provenance/issues/2149).
* Charge fees based only on `Msg` type (and not gas) [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
  Gas is still monitored to prevent DDOS and is still limited to 4,000,000 per transaction and 60,000,000 per block.
* Create the `x/flatfees` module for managing the costs of msgs [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
* We now use a custom `app.Simulate` method when simulating a Tx [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
  It reports the `gas_wanted` as the amount of nhash required in the fee (instead of the actual gas wanted). When simulating (e.g. using `--gas auto`), users should always use `1nhash` for the gas prices, and the default gas multiplier of `1.0`. This will cause existing/standard clients to set the fee to `gas-wanted` * `gas-prides` * `gas-multiplier` which will equal the required fee. The `CalculateTxFees` query in the flatfees module can also be used to simulate a tx and it explicitly returns the required fee (as well as actual gas used). The `gas_wanted` provided with a Tx is now largely ignored since the existing/standard clients will set it to the amount of fee, which will usually be more than the max tx gas.
* Create the `bouvardia` upgrade for getting us to `v1.26.0` [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
* Add the `x/nft` module [PR 2422](https://github.com/provenance-io/provenance/pull/2422).
* Create the `x/ledger` module for tracking ledger data [PR 2422](https://github.com/provenance-io/provenance/pull/2422).
* Create the `x/registry` module for assigning roles to NFTs [PR 2422](https://github.com/provenance-io/provenance/pull/2422).
* Create the `x/asset` module for creating various types of assets using the NFT, ledger, and registry modules [PR 2422](https://github.com/provenance-io/provenance/pull/2422).
* Added `x/vault` module providing ERC-4626–inspired vault functionality [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* Add vault queries to startgate whitelist [PR 2473](https://github.com/provenance-io/provenance/pull/2473).
* Add vault v1.0.7 `VaultPendingSwapOuts` query to stargate whitelist [PR 2502](https://github.com/provenance-io/provenance/pull/2502).

### Improvements

* Refactor test setups to use testutil.MutateGenesisState [#2013](https://github.com/provenance-io/provenance/issues/2013).
* Added documentation to NetAssetValues fields in proto clarifying that amounts are in `usd` units, where 1usd =$1.00 [#2291](https://github.com/provenance-io/provenance/issues/2291).
* Triggers no longer track or use the extra gas provided when creating the trigger [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
  Users pay for the trigger msg execution when creating the trigger (based on msg type).
* Renamed `pioconfig` stuff to use `Prov` instead of `Provenance` [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
* Make `pioconfig.GetProvConfig` return the defaults if `SetProvConfig` hasn't been called yet [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
* Create the `ConsumeAdditionalFlatFee` helper for the FlatFeeGasMeter [PR 2413](https://github.com/provenance-io/provenance/pull/2413).
* Fix the casing of 'as' in our dockerfiles [#2452](https://github.com/provenance-io/provenance/issues/2452).
* Add gitignore entry for goenv and gvm config file [PR 2467](https://github.com/provenance-io/provenance/pull/2467).
* Increase the max memo length to 1024 bytes (from 256) [PR 2482](https://github.com/provenance-io/provenance/pull/2482).

### Bug Fixes

* Moved `MsgExecuteContract.proto` from proto to `legacy_protos` directory [#2399](https://github.com/provenance-io/provenance/issues/2399).
* Fix the staking restriction error message that contained the wrong amount [PR 2433](https://github.com/provenance-io/provenance/pull/2433).
* Use 6 digits (instead of 2) from the percent when calculating max staking amount [PR 2433](https://github.com/provenance-io/provenance/pull/2433).
* Fix docker builds by generating the go code to x/wasm instead of the uncopied legacy_protos/ [PR 2443](https://github.com/provenance-io/provenance/pull/2443).
* Add missing `AddTxFlagsToCmd` to `gov-root-name` [PR 2456](https://github.com/provenance-io/provenance/pull/2456).
* Fix vault stargate whitelist type url typo [PR 2479](https://github.com/provenance-io/provenance/pull/2479).

### Deprecated

* The `CalculateTxFees` query in the `x/msgfees` module is deprecated [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
  Users should switch to query in the `x/flatfees` module with the same name.
* The `minimum-gas-prices` config field (in app.toml) is now ignored and is always treated as `1nhash` [PR 2318](https://github.com/provenance-io/provenance/pull/2318).

### Client Breaking

* Users must now use `1nhash` for their `gas-prices` and should use `1.0` for the `gas-multiplier` [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
  Using anything else will result in paying significantly more than the required cost for a tx.

### Api Breaking

* The `x/msgfees` module has been removed except for the (deprecated) `CalculateTxFees` query [PR 2318](https://github.com/provenance-io/provenance/pull/2318).

### Dependencies

* `4d63.com/gocheckcompilerdirectives` added at v1.2.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `4d63.com/gochecknoglobals` added at v0.2.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `actions/checkout` bumped to 5 (from 4) ([PR 2421](https://github.com/provenance-io/provenance/pull/2421), [PR 2453](https://github.com/provenance-io/provenance/pull/2453)).
* `actions/download-artifact` bumped to 6 (from 4) ([PR 2414](https://github.com/provenance-io/provenance/pull/2414), [PR 2510](https://github.com/provenance-io/provenance/pull/2510)).
* `actions/setup-go` bumped to 6 (from 5) [PR 2441](https://github.com/provenance-io/provenance/pull/2441).
* `actions/setup-java` bumped to 5 (from 4) [PR 2428](https://github.com/provenance-io/provenance/pull/2428).
* `actions/upload-artifact` bumped to 5 (from 4) [PR 2509](https://github.com/provenance-io/provenance/pull/2509).
* `cloud.google.com/go/auth/oauth2adapt` bumped to v0.2.4 (from v0.2.2) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `cloud.google.com/go/auth` bumped to v0.9.3 (from v0.6.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `cloud.google.com/go/compute/metadata` bumped to v0.7.0 (from v0.6.0) [PR 2405](https://github.com/provenance-io/provenance/pull/2405).
* `cloud.google.com/go/iam` bumped to v1.2.0 (from v1.1.9) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `cloud.google.com/go/storage` bumped to v1.43.0 (from v1.41.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `cloud.google.com/go` bumped to v0.115.1 (from v0.115.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `cosmossdk.io/log` bumped to v1.6.1 (from v1.6.0) [PR 2424](https://github.com/provenance-io/provenance/pull/2424).
* `cosmossdk.io/x/nft` added at v0.1.1 [PR 2422](https://github.com/provenance-io/provenance/pull/2422).
* `github.com/4meepo/tagalign` added at v1.3.4 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/Abirdcfly/dupword` added at v0.1.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/Antonboom/errname` added at v0.1.13 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/Antonboom/nilnil` added at v0.1.9 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/Antonboom/testifylint` added at v1.4.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/BurntSushi/toml` added at v1.4.1-0.20240526193622-a339e1f7089c [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/Crocmagnon/fatcontext` added at v0.5.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/Djarvur/go-err113` added at v0.0.0-20210108212216-aea10b59be24 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/GaijinEntertainment/go-exhaustruct/v3` added at v3.3.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/Masterminds/semver/v3` added at v3.3.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/OpenPeeDeeP/depguard/v2` added at v2.2.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/alecthomas/go-check-sumtype` added at v0.1.4 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/alexkohler/nakedret/v2` added at v2.0.4 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/alexkohler/prealloc` added at v1.0.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/alingse/asasalint` added at v0.0.11 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ashanbrown/forbidigo` added at v1.6.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ashanbrown/makezero` added at v1.1.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/bits-and-blooms/bitset` bumped to v1.13.0 (from v1.8.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/bkielbasa/cyclop` added at v1.2.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/blizzy78/varnamelen` added at v0.8.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/bombsimon/wsl/v4` added at v4.4.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/breml/bidichk` added at v0.2.7 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/breml/errchkjson` added at v0.3.6 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/butuzov/ireturn` added at v0.3.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/butuzov/mirror` added at v1.2.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/bytedance/sonic/loader` bumped to v0.3.0 (from v0.2.4) [PR 2424](https://github.com/provenance-io/provenance/pull/2424).
* `github.com/bytedance/sonic` bumped to v1.14.0 (from v1.13.1) [PR 2424](https://github.com/provenance-io/provenance/pull/2424).
* `github.com/catenacyber/perfsprint` added at v0.7.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ccojocar/zxcvbn-go` added at v1.0.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/charithe/durationcheck` added at v0.0.10 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/chavacava/garif` added at v0.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ckaznocha/intrange` added at v0.2.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/cosmos/cosmos-sdk` bumped to v0.50.14-pio-2 of `github.com/provenance-io/cosmos-sdk` (from v0.50.14-pio-1 of `github.com/provenance-io/cosmos-sdk`) [PR 2318](https://github.com/provenance-io/provenance/pull/2318).
* `github.com/cosmos/iavl` bumped to v1.2.6 of `github.com/cosmos/iavl` (from v1.2.0 of `github.com/cosmos/iavl`) [PR 2437](https://github.com/provenance-io/provenance/pull/2437).
* `github.com/curioswitch/go-reassign` added at v0.2.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/daixiang0/gci` added at v0.13.5 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/denis-tingaikin/go-header` added at v0.5.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ettle/strcase` added at v0.2.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/fatih/color` bumped to v1.17.0 (from v1.16.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/fatih/structtag` added at v1.2.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/firefart/nonamedreturns` added at v1.0.5 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/fsnotify/fsnotify` bumped to v1.9.0 (from v1.7.0) [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/fzipp/gocyclo` added at v0.6.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ghostiam/protogetter` added at v0.3.6 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/gobwas/glob` added at v0.2.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/gofrs/flock` added at v0.12.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/golangci/dupl` added at v0.0.0-20180902072040-3e9179ac440a [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/golangci/gofmt` added at v0.0.0-20240816233607-d8596aa466a9 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/golangci/golangci-lint` added at v1.61.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/golangci/misspell` added at v0.6.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/golangci/modinfo` added at v0.3.4 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/golangci/plugin-module-register` added at v0.1.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/golangci/revgrep` added at v0.5.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/golangci/unconvert` added at v0.0.0-20240309020433-c5143eacb3ed [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/golang/snappy` bumped to v0.0.5-0.20220116011046-fa5810519dcb (from v0.0.4) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/googleapis/enterprise-certificate-proxy` bumped to v0.3.3 (from v0.3.2) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/googleapis/gax-go/v2` bumped to v2.13.0 (from v2.12.5) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/google/flatbuffers` bumped to v2.0.8+incompatible (from v1.12.1) [PR 2426](https://github.com/provenance-io/provenance/pull/2426).
* `github.com/google/s2a-go` bumped to v0.1.8 (from v0.1.7) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/gordonklaus/ineffassign` added at v0.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/gostaticanalysis/analysisutil` added at v0.7.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/gostaticanalysis/comment` added at v1.4.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/gostaticanalysis/forcetypeassert` added at v0.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/gostaticanalysis/nilerr` added at v0.1.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/go-critic/go-critic` added at v0.11.4 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/go-logr/logr` bumped to v1.4.3 (from v1.4.2) [PR 2405](https://github.com/provenance-io/provenance/pull/2405).
* `github.com/go-toolsmith/astcast` added at v1.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/go-toolsmith/astcopy` added at v1.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/go-toolsmith/astequal` added at v1.2.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/go-toolsmith/astfmt` added at v1.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/go-toolsmith/astp` added at v1.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/go-toolsmith/strparse` added at v1.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/go-toolsmith/typep` added at v1.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/go-viper/mapstructure/v2` added at v2.4.0 [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/go-xmlfmt/xmlfmt` added at v1.1.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/hashicorp/go-getter` bumped to v1.7.9 (from v1.7.5) [PR 2426](https://github.com/provenance-io/provenance/pull/2426).
* `github.com/hashicorp/go-version` bumped to v1.7.0 (from v1.6.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/hashicorp/hcl` removed at v1.0.0 [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/hexops/gotextdiff` added at v1.0.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/incu6us/goimports-reviser/v3` added at v3.8.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/jgautheron/goconst` added at v1.7.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/jingyugao/rowserrcheck` added at v1.1.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/jirfag/go-printf-func-name` added at v0.0.0-20200119135958-7558a9eaa5af [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/jjti/go-spancheck` added at v0.6.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/julz/importas` added at v0.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/karamaru-alpha/copyloopvar` added at v1.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/kisielk/errcheck` added at v1.7.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/kkHAIKE/contextcheck` added at v1.1.5 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/kulti/thelper` added at v0.6.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/kunwardeep/paralleltest` added at v1.0.10 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/kyoh86/exportloopref` added at v0.1.11 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/lasiar/canonicalheader` added at v1.1.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ldez/gomoddirectives` added at v0.2.4 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ldez/tagliatelle` added at v0.5.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/leonklingele/grouper` added at v1.1.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/lufeee/execinquery` added at v1.2.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/macabu/inamedparam` added at v0.1.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/magiconair/properties` removed at v1.8.7 [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/maratori/testableexamples` added at v1.0.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/maratori/testpackage` added at v1.1.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/matoous/godox` added at v0.0.0-20230222163458-006bad1f9d26 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/mattn/go-runewidth` added at v0.0.13 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/mgechev/revive` added at v1.3.9 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/mitchellh/mapstructure` removed at v1.5.0 [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/moricho/tparallel` added at v0.3.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/nakabonne/nestif` added at v0.3.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/nishanths/exhaustive` added at v0.12.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/nishanths/predeclared` added at v0.2.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/nunnatsa/ginkgolinter` added at v0.16.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/olekukonko/tablewriter` added at v0.0.5 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/pelletier/go-toml/v2` bumped to v2.2.4 (from v2.2.2) [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/polyfloyd/go-errorlint` added at v1.6.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/provlabs/vault` added at v1.0.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/provlabs/vault` bumped to v1.0.9 (from v1.0.3) ([PR 2465](https://github.com/provenance-io/provenance/pull/2465), [PR 2477](https://github.com/provenance-io/provenance/pull/2477), [PR 2501](https://github.com/provenance-io/provenance/pull/2501), [PR 2506](https://github.com/provenance-io/provenance/pull/2506)).
* `github.com/quasilyte/gogrep` added at v0.5.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/quasilyte/go-ruleguard/dsl` added at v0.3.22 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/quasilyte/go-ruleguard` added at v0.4.3-0.20240823090925-0fe6f58b47b1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/quasilyte/regex/syntax` added at v0.0.0-20210819130434-b3f0c404a727 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/quasilyte/stdinfo` added at v0.0.0-20220114132959-f7386bf02567 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/rivo/uniseg` added at v0.2.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ryancurrah/gomodguard` added at v1.3.5 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ryanrolds/sqlclosecheck` added at v0.5.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/sagikazarmark/locafero` bumped to v0.11.0 (from v0.4.0) [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/sagikazarmark/slog-shim` removed at v0.1.0 [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/sanposhiho/wastedassign/v2` added at v2.0.7 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/santhosh-tekuri/jsonschema/v5` added at v5.3.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/sashamelentyev/interfacebloat` added at v1.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/sashamelentyev/usestdlibvars` added at v1.27.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/securego/gosec/v2` added at v2.21.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/shazow/go-diff` added at v0.0.0-20160112020656-b6b7b6733b8c [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/sirupsen/logrus` added at v1.9.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/sivchari/containedctx` added at v1.0.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/sivchari/tenv` added at v1.10.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/sonatard/noctx` added at v0.0.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/sourcegraph/conc` bumped to v0.3.1-0.20240121214520-5f936abd7ae8 (from v0.3.0) [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/sourcegraph/go-diff` added at v0.7.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/spf13/afero` bumped to v1.15.0 (from v1.11.0) [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/spf13/cast` bumped to v1.10.0 (from v1.9.2) [PR 2444](https://github.com/provenance-io/provenance/pull/2444).
* `github.com/spf13/cobra` bumped to v1.10.1 (from v1.9.1) [PR 2446](https://github.com/provenance-io/provenance/pull/2446).
* `github.com/spf13/pflag` bumped to v1.0.10 (from v1.0.6) ([PR 2401](https://github.com/provenance-io/provenance/pull/2401), [PR 2440](https://github.com/provenance-io/provenance/pull/2440)).
* `github.com/spf13/viper` bumped to v1.21.0 (from v1.19.0) [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `github.com/ssgreg/nlreturn/v2` added at v2.2.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/stbenjam/no-sprintf-host-port` added at v0.1.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/stretchr/objx` added at v0.5.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/stretchr/testify` bumped to v1.11.1 (from v1.10.0) [PR 2434](https://github.com/provenance-io/provenance/pull/2434).
* `github.com/tdakkota/asciicheck` added at v0.2.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/tetafro/godot` added at v1.4.17 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/timakin/bodyclose` added at v0.0.0-20230421092635-574207250966 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/timonwong/loggercheck` added at v0.9.4 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/tomarrell/wrapcheck/v2` added at v2.9.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/tommy-muehle/go-mnd/v2` added at v2.5.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ulikunitz/xz` bumped to v0.5.14 (from v0.5.11) [PR 2436](https://github.com/provenance-io/provenance/pull/2436).
* `github.com/ultraware/funlen` added at v0.1.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ultraware/whitespace` added at v0.1.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/uudashr/gocognit` added at v1.1.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/xen0n/gosmopolitan` added at v1.2.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/yagipy/maintidx` added at v1.0.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/yeya24/promlinter` added at v0.3.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github.com/ykadowak/zerologlint` added at v0.1.5 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `github/codeql-action` bumped to 4 (from 3) [PR 2472](https://github.com/provenance-io/provenance/pull/2472).
* `gitlab.com/bosi/decorder` added at v0.4.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `golang.org/x/arch` bumped to v0.17.0 (from v0.15.0) [PR 2424](https://github.com/provenance-io/provenance/pull/2424).
* `golang.org/x/crypto` bumped to v0.40.0 (from v0.36.0) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2425](https://github.com/provenance-io/provenance/pull/2425), [PR 2458](https://github.com/provenance-io/provenance/pull/2458)).
* `golang.org/x/exp/typeparams` added at v0.0.0-20240314144324-c7f7c6466f7f [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `golang.org/x/exp` bumped to v0.0.0-20240904232852-e7e105dedf7e (from v0.0.0-20240719175910-8a7402abbf56) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `golang.org/x/mod` added at v0.26.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `golang.org/x/net` bumped to v0.42.0 (from v0.38.0) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2425](https://github.com/provenance-io/provenance/pull/2425), [PR 2458](https://github.com/provenance-io/provenance/pull/2458)).
* `golang.org/x/oauth2` bumped to v0.30.0 (from v0.28.0) [PR 2405](https://github.com/provenance-io/provenance/pull/2405).
* `golang.org/x/sys` bumped to v0.34.0 (from v0.31.0) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2458](https://github.com/provenance-io/provenance/pull/2458)).
* `golang.org/x/term` bumped to v0.33.0 (from v0.30.0) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2458](https://github.com/provenance-io/provenance/pull/2458)).
* `golang.org/x/text` bumped to v0.28.0 (from v0.27.0) [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `golang.org/x/time` bumped to v0.6.0 (from v0.5.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `golang.org/x/tools/go/expect` added at v0.1.1-deprecated [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `golang.org/x/tools/go/packages/packagestest` added at v0.1.1-deprecated [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `golang.org/x/tools` added at v0.35.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `google.golang.org/api` bumped to v0.196.0 (from v0.186.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `google.golang.org/genproto/googleapis/api` bumped to v0.0.0-20250707201910-8d1bb00bc6a7 (from v0.0.0-20250324211829-b45e905df463) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2425](https://github.com/provenance-io/provenance/pull/2425)).
* `google.golang.org/genproto/googleapis/rpc` bumped to v0.0.0-20250707201910-8d1bb00bc6a7 (from v0.0.0-20250324211829-b45e905df463) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2425](https://github.com/provenance-io/provenance/pull/2425)).
* `google.golang.org/genproto` bumped to v0.0.0-20240903143218-8af14fe29dc1 (from v0.0.0-20240701130421-f6361c86f094) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `google.golang.org/grpc` bumped to v1.75.1 (from v1.73.0) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2425](https://github.com/provenance-io/provenance/pull/2425), [PR 2448](https://github.com/provenance-io/provenance/pull/2448)).
* `google.golang.org/protobuf` bumped to v1.36.10 (from v1.36.6) ([PR 2427](https://github.com/provenance-io/provenance/pull/2427), [PR 2447](https://github.com/provenance-io/provenance/pull/2447), [PR 2460](https://github.com/provenance-io/provenance/pull/2460)).
* `gopkg.in/ini.v1` removed at v1.67.0 [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `go-simpler.org/musttag` added at v0.12.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `go-simpler.org/sloglint` added at v0.7.2 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc` bumped to v0.54.0 (from v0.49.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` bumped to v0.54.0 (from v0.49.0) [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `go.opentelemetry.io/otel/metric` bumped to v1.37.0 (from v1.35.0) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2425](https://github.com/provenance-io/provenance/pull/2425)).
* `go.opentelemetry.io/otel/trace` bumped to v1.37.0 (from v1.35.0) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2425](https://github.com/provenance-io/provenance/pull/2425)).
* `go.opentelemetry.io/otel` bumped to v1.37.0 (from v1.35.0) ([PR 2405](https://github.com/provenance-io/provenance/pull/2405), [PR 2425](https://github.com/provenance-io/provenance/pull/2425)).
* `go.uber.org/atomic` added at v1.10.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `go.uber.org/automaxprocs` added at v1.5.3 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `go.uber.org/multierr` added at v1.11.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `go.uber.org/multierr` removed at v1.11.0 [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `go.uber.org/zap` added at v1.24.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `go.yaml.in/yaml/v3` added at v3.0.4 [PR 2445](https://github.com/provenance-io/provenance/pull/2445).
* `honnef.co/go/tools` added at v0.5.1 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `mvdan.cc/gofumpt` added at v0.7.0 [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `mvdan.cc/unparam` added at v0.0.0-20240528143540-8a5130ca722f [PR 2458](https://github.com/provenance-io/provenance/pull/2458).
* `sigs.k8s.io/yaml` bumped to v1.6.0 (from v1.5.0) [PR 2407](https://github.com/provenance-io/provenance/pull/2407).
* `stefanzweifel/git-auto-commit-action` bumped to 7 (from 5) ([PR 2377](https://github.com/provenance-io/provenance/pull/2377), [PR 2481](https://github.com/provenance-io/provenance/pull/2481)).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.25.1...v1.26.0-rc1

