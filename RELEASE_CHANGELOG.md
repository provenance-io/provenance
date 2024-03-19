## [v1.18.0-rc3](https://github.com/provenance-io/provenance/releases/tag/v1.18.0-rc3) - 2024-03-19

### Features

* Add [Payments](x/exchange/spec/01_concepts.md#payments) to the exchange module [#1703](https://github.com/provenance-io/provenance/issues/1703).
  Payments allow two parties to trade assets securely and asynchronously.
* Add upgrade handler to set net asset values to markers in pio-testnet-1 [PR 1881](https://github.com/provenance-io/provenance/pull/1881).

### Improvements

* Allow force transfers from marker and market accounts [#1855](https://github.com/provenance-io/provenance/pull/1855).
* Add a `tourmaline-rc3` upgrade handler to set some new exchange module params related to payments [#1703](https://github.com/provenance-io/provenance/issues/1703).
* Remove the startup warning issued when disable-iavl-fastnode is true (we recommend keeping it as true if you already have it that way) [#1874](https://github.com/provenance-io/provenance/pull/1874).
* Switch to `github.com/cometbft/cometbft-db` `v0.7.0` (from `github.com/tendermint/tm-db` `v0.6.7`) [#1874](https://github.com/provenance-io/provenance/pull/1874).
* Allow NAV volume to exceed a marker's supply [#1883](https://github.com/provenance-io/provenance/pull/1883).

### Bug Fixes

* Fix `MarkerTransferAuthorization` validation to ensure the coins and addresses are all valid [#1856](https://github.com/provenance-io/provenance/pull/1856).
* In `MarketCommitmentSettle`, only consume the settlement fee if the settlement succeeds [#1703](https://github.com/provenance-io/provenance/issues/1703).

### Dependencies

- Bump `google.golang.org/grpc` from 1.61.1 to 1.62.1 ([#1850](https://github.com/provenance-io/provenance/pull/1850), [#1864](https://github.com/provenance-io/provenance/pull/1864))
- Bump `cosmossdk.io/math` from 1.2.0 to 1.3.0 ([#1857](https://github.com/provenance-io/provenance/pull/1857))
- Bump `peter-evans/create-pull-request` from 6.0.0 to 6.0.2 ([#1858](https://github.com/provenance-io/provenance/pull/1858), [#1872](https://github.com/provenance-io/provenance/pull/1872))
- Bump `github.com/golang/protobuf` from 1.5.3 to 1.5.4 ([#1863](https://github.com/provenance-io/provenance/pull/1863))
- Bump `github.com/stretchr/testify` from 1.8.4 to 1.9.0 ([#1860](https://github.com/provenance-io/provenance/pull/1860))
- Bump `bufbuild/buf-setup-action` from 1.29.0 to 1.30.0 ([#1871](https://github.com/provenance-io/provenance/pull/1871))
- Bump `github.com/cosmos/cosmos-sdk` from v0.46.13-pio-3 to v0.46.13-pio-4 ([#1874](https://github.com/provenance-io/provenance/pull/1874)).
- Bump `github.com/cosmos/ibc-go/v6` from v6.2.0-pio-1 to v6.2.0-pio-2 ([#1874](https://github.com/provenance-io/provenance/pull/1874)).
- Bump `github.com/CosmWasm/wasmd` from v0.30.0-pio-6 to v0.30.0-pio-7 ([#1874](https://github.com/provenance-io/provenance/pull/1874)).
- Bump `github.com/cosmos/iavl` from v0.19.6 to v0.20.1 ([#1874](https://github.com/provenance-io/provenance/pull/1874)).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.18.0-rc2...v1.18.0-rc3
* https://github.com/provenance-io/provenance/compare/v1.17.1...v1.18.0-rc3

---

## [v1.18.0-rc2](https://github.com/provenance-io/provenance/releases/tag/v1.18.0-rc2) - 2024-02-22

### Features

* In the marker module's `SendRestrictionFn`, allow a transfer agent to be identified through the context [#1834](https://github.com/provenance-io/provenance/issues/1834).
* In the exchange module, provide the admin as the transfer agent when attepting to move funds [#1834](https://github.com/provenance-io/provenance/issues/1834).

### Improvements

* Add an empty `tourmaline-rc2` upgrade handler [#1834](https://github.com/provenance-io/provenance/issues/1834).
* Add new force_transfer access that is required for an account to do a forced transfer ([#1829](https://github.com/provenance-io/provenance/issues/1829)).
* Add exchange commitment stuff to CLI [PR 1830](https://github.com/provenance-io/provenance/pull/1830).
* Update the MsgFees Params to set the nhash per usd-mil to 40,000,000 ($0.025/hash) [#1833](https://github.com/provenance-io/provenance/pull/1833).
* Bid order prices are no longer restricted to amounts that can be evenly applied to a buyer settlement fee ratio [1834](https://github.com/provenance-io/provenance/pull/1843).
* In the marker and exchange modules, help ensure funds don't get sent to blocked addresses [#1834](https://github.com/provenance-io/provenance/issues/1834).
* Update marker and exchange spec docs to include info about transfer agents [#1834](https://github.com/provenance-io/provenance/issues/1834).
* Prevent restricted markers from being sent to the fee collector account [#1845](https://github.com/provenance-io/provenance/issues/1845).

### Bug Fixes

* Prevent funds from going to or from a marker without the transfer agent having deposit or withdraw access (respectively) [#1834](https://github.com/provenance-io/provenance/issues/1834).
* Ensure the store loader isn't nil when the handling an upgrade [1852](https://github.com/provenance-io/provenance/pull/1852).

### API Breaking

* Accounts that have transfer access in a marker are no longer allowed to do forced transfers ([#1829](https://github.com/provenance-io/provenance/issues/1829)).
  Accounts must now have the force_transfer access for that.

### Dependencies

- Bump `codecov/codecov-action` from 3 to 4 ([#1828](https://github.com/provenance-io/provenance/pull/1828))
- Bump `peter-evans/create-pull-request` from 5.0.2 to 6.0.0 ([#1827](https://github.com/provenance-io/provenance/pull/1827))
- Bump `bufbuild/buf-setup-action` from 1.28.1 to 1.29.0 ([#1825](https://github.com/provenance-io/provenance/pull/1825))
- Bump `github.com/rs/zerolog` from 1.31.0 to 1.32.0 ([#1832](https://github.com/provenance-io/provenance/pull/1832))
- Bump `serde-json-wasm` from 1.0.0 to 1.0.1 in /testutil/contracts/rate-limiter ([#1836](https://github.com/provenance-io/provenance/pull/1836))
- Bump `serde-json-wasm` from 0.5.1 to 0.5.2 in /testutil/contracts/counter ([#1837](https://github.com/provenance-io/provenance/pull/1837))
- Bump `serde-json-wasm` from 0.5.1 to 0.5.2 in /testutil/contracts/echo ([#1838](https://github.com/provenance-io/provenance/pull/1838))
- Bump `golangci/golangci-lint-action` from 3 to 4 ([#1840](https://github.com/provenance-io/provenance/pull/1840))
- Bump `google.golang.org/grpc` from 1.61.0 to 1.61.1 ([#1842](https://github.com/provenance-io/provenance/pull/1842))
- Bump `cosmos-sdk` from v0.46.13-pio-2 to v0.46.13-pio-3 ([#1848](https://github.com/provenance-io/provenance/pull/1848))

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.18.0-rc1...v1.18.0-rc2
* https://github.com/provenance-io/provenance/compare/v1.17.1...v1.18.0-rc2

---

## [v1.18.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.18.0-rc1) - 2024-01-30

### Features

* Add the ibcratelimit module [#1498](https://github.com/provenance-io/provenance/issues/1498).
* Add NAV support for metadata scopes [#1749](https://github.com/provenance-io/provenance/issues/1749).
* Add fix for NAV units to tourmaline upgrade handler [#1815](https://github.com/provenance-io/provenance/issues/1815).
* Support commitments in the exchange module [#1789](https://github.com/provenance-io/provenance/issues/1789).

### Improvements

* Add upgrade handler for 1.18 [#1756](https://github.com/provenance-io/provenance/pull/1756).
* Remove the rust upgrade handlers [PR 1774](https://github.com/provenance-io/provenance/pull/1774).
* Allow bypassing the config warning wait using an environment variable [PR 1810](https://github.com/provenance-io/provenance/pull/1810).
* Filter out empty distribution events from begin blocker [#1822](https://github.com/provenance-io/provenance/pull/1822).

### Deprecated

* The concept of an "active" market (in the exchange module) has been removed in favor of specifying whether it accepts orders [#1789](https://github.com/provenance-io/provenance/issues/1789).
  * The `MarketUpdateEnabled` endpoint has been deprecated and is no longer usable. It is replaced with the `MarketUpdateAcceptingOrders` endpoint.
  * `MsgMarketUpdateEnabledRequest` is replaced with `MsgMarketUpdateAcceptingOrdersRequest`.
  * `MsgMarketUpdateEnabledResponse` is replaced with `MsgMarketUpdateAcceptingOrdersResponse`.
  * `EventMarketEnabled` is replaced with `EventMarketOrdersEnabled`.
  * `EventMarketDisabled` is replaced with `EventMarketOrdersDisabled`.

### Bug Fixes

* Remove deleted marker send deny entries [#1666](https://github.com/provenance-io/provenance/issues/1666).
* Update protos, naming, and documentation to use mills [#1813](https://github.com/provenance-io/provenance/issues/1813).
* Update marker transfer to work with groups [#1818](https://github.com/provenance-io/provenance/issues/1818).

### Client Breaking

* The `provenanced tx exchange market-enabled` command has been changed to `provenanced tx exchange market-accepting-orders` [#1789](https://github.com/provenance-io/provenance/issues/1789).

### API Breaking

* The `MarketUpdateEnabled` has been deprecated and replaced with `MarketUpdateAcceptingOrders` along with its request, response, and events [#1789](https://github.com/provenance-io/provenance/issues/1789).
  The old endpoint is no longer usable. See the Deprecated section for more details.

### Dependencies

- Bump `bufbuild/buf-setup-action` from 1.27.1 to 1.28.1 ([#1724](https://github.com/provenance-io/provenance/pull/1724), [#1744](https://github.com/provenance-io/provenance/pull/1744), [#1750](https://github.com/provenance-io/provenance/pull/1750))
- Bump `github.com/google/uuid` from 1.3.1 to 1.6.0 ([#1723](https://github.com/provenance-io/provenance/pull/1723), [#1781](https://github.com/provenance-io/provenance/pull/1781), [#1819](https://github.com/provenance-io/provenance/pull/1819))
- Bump `github.com/gorilla/mux` from 1.8.0 to 1.8.1 ([#1734](https://github.com/provenance-io/provenance/pull/1734))
- Bump `golang.org/x/text` from 0.13.0 to 0.14.0 ([#1735](https://github.com/provenance-io/provenance/pull/1735))
- Bump `cosmossdk.io/math` from 1.1.2 to 1.2.0 ([#1739](https://github.com/provenance-io/provenance/pull/1739))
- Update `async-icq` from `github.com/strangelove-ventures/async-icq/v6` to `github.com/cosmos/ibc-apps/modules/async-icq/v6.1.0` ([#1748](https://github.com/provenance-io/provenance/pull/1748))
- Bump `github.com/spf13/viper` from 1.17.0 to 1.18.2 ([#1777](https://github.com/provenance-io/provenance/pull/1777), [#1795](https://github.com/provenance-io/provenance/pull/1795))
- Bump `actions/setup-go` from 4 to 5 ([#1776](https://github.com/provenance-io/provenance/pull/1776))
- Bump `github.com/spf13/cast` from 1.5.1 to 1.6.0 ([#1769](https://github.com/provenance-io/provenance/pull/1769))
- Bump `actions/setup-java` from 3 to 4 ([#1770](https://github.com/provenance-io/provenance/pull/1770))
- Bump `github.com/dvsekhvalnov/jose2go` from 1.5.0 to 1.6.0 ([#1793](https://github.com/provenance-io/provenance/pull/1793))
- Bump `google.golang.org/protobuf` from 1.31.0 to 1.32.0 ([#1790](https://github.com/provenance-io/provenance/pull/1790))
- Bump `github/codeql-action` from 2 to 3 ([#1784](https://github.com/provenance-io/provenance/pull/1784))
- Bump `actions/download-artifact` from 3 to 4 ([#1785](https://github.com/provenance-io/provenance/pull/1785))
- Bump `actions/upload-artifact` from 3 to 4 ([#1785](https://github.com/provenance-io/provenance/pull/1785))
- Bump `google.golang.org/grpc` from 1.59.0 to 1.61.0 ([#1794](https://github.com/provenance-io/provenance/pull/1794), [#1820](https://github.com/provenance-io/provenance/pull/1820))
- Bump `golang.org/x/crypto` from 0.14.0 to 0.17.0 ([#1788](https://github.com/provenance-io/provenance/pull/1788))
- Bump `cosmossdk.io/errors` from 1.0.0 to 1.0.1 ([#1806](https://github.com/provenance-io/provenance/pull/1806))
- Bump `actions/cache` from 3 to 4 ([#1817](https://github.com/provenance-io/provenance/pull/1817))

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.17.1...v1.18.0-rc1

