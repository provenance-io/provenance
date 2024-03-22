## [v1.18.0](https://github.com/provenance-io/provenance/releases/tag/v1.18.0) - 2024-03-22

Release v1.18.0 expands the exchange capabilities, and includes various other features, improvements, and bug fixes.

### Overview

With the release of v1.18.0, the Provenance Blockchain protocol's exchange module adds support for both commitments and payments.
The metadata module now allows recording Net-Asset-Values (NAVs) on scopes.
And in the marker module, various enhancements have been made to help facilitate and safeguard movements of funds.

### Exchange Module Updates

The exchange module now has functionality for commitments and payments.

Commitments allow an account to give a specific market the ability to manage a specified amount of funds.
Committed funds remain in the account (with a hold on them) until the market moves them.

Payments allow two parties to securely exchange funds.
When a payment is created, a hold is placed on the source funds until the payment is accepted, rejected, or cancelled.
When a payment is accepted, the source funds are sent to the target and the target funds are sent to the source.

Additionally, improvements have been made that allow a market actor (e.g. with `settle` permission) to act as a transfer agent (with respects to the markers of the funds being moved).
This allows an account with both `settle` permission in a market, and `transfer` permission on a marker, to use the exchange module to facilitate movements of that marker's funds.

### Marker Module Updates

The marker module received several updates related to the movement of funds.

* There is a new `force_transfer` permission that is required in order to force a transfer of a marker's funds, and the `transfer` permission no longer has the ability to do forced transfers.
* Safeguards were added to prevent funds from being sent to addresses that are specially used by the blockchain (e.g. the fee collector).
* Transfers now work for groups accounts.
* Net-Asset-Values (NAVs) now use mills (instead of cents).

### Features

* Support commitments in the exchange module [#1789](https://github.com/provenance-io/provenance/issues/1789), [PR 1830](https://github.com/provenance-io/provenance/pull/1830).
  Commitments allow a party to give a market access to a specified amount of funds in their account.
* Add [Payments](x/exchange/spec/01_concepts.md#payments) to the exchange module [#1703](https://github.com/provenance-io/provenance/issues/1703).
  Payments allow two parties to trade assets securely and asynchronously.
* Add the ibcratelimit module [#1498](https://github.com/provenance-io/provenance/issues/1498).
* Add NAV support for metadata scopes [#1749](https://github.com/provenance-io/provenance/issues/1749).
* Add fix for NAV units to tourmaline upgrade handler [#1815](https://github.com/provenance-io/provenance/issues/1815).
* In the marker module's `SendRestrictionFn`, allow a transfer agent to be identified through the context [#1834](https://github.com/provenance-io/provenance/issues/1834).
* In the exchange module, provide the admin as the transfer agent when attempting to move funds [#1834](https://github.com/provenance-io/provenance/issues/1834).
* Add upgrade handler to set net asset values to markers in pio-testnet-1 [PR 1881](https://github.com/provenance-io/provenance/pull/1881).
* Add upgrade handler to set net asset values and update block height for pio-mainnet-1 [PR 1888](https://github.com/provenance-io/provenance/pull/1888).

### Improvements

* Add `tourmaline` upgrade handlers for 1.18 [PR 1756](https://github.com/provenance-io/provenance/pull/1756), [#1834](https://github.com/provenance-io/provenance/issues/1834), [#1703](https://github.com/provenance-io/provenance/issues/1703).
* Remove the rust upgrade handlers [PR 1774](https://github.com/provenance-io/provenance/pull/1774).
* Allow bypassing the config warning wait using an environment variable [PR 1810](https://github.com/provenance-io/provenance/pull/1810).
* Filter out empty distribution events from begin blocker [PR 1822](https://github.com/provenance-io/provenance/pull/1822).
* Add new force_transfer access that is required for an account to do a forced transfer ([#1829](https://github.com/provenance-io/provenance/issues/1829)).
* Update the MsgFees Params to set the nhash per usd-mil to 40,000,000 ($0.025/hash) [PR 1833](https://github.com/provenance-io/provenance/pull/1833).
* Bid order prices are no longer restricted to amounts that can be evenly applied to a buyer settlement fee ratio [PR 1834](https://github.com/provenance-io/provenance/pull/1843).
* In the marker and exchange modules, help ensure funds don't get sent to blocked addresses [#1834](https://github.com/provenance-io/provenance/issues/1834).
* Update marker and exchange spec docs to include info about transfer agents [#1834](https://github.com/provenance-io/provenance/issues/1834).
* Prevent restricted markers from being sent to the fee collector account [#1845](https://github.com/provenance-io/provenance/issues/1845).
* Allow force transfers from marker and market accounts [PR 1855](https://github.com/provenance-io/provenance/pull/1855).
* Remove the startup warning issued when disable-iavl-fastnode is true (we recommend keeping it as true if you already have it that way) [PR 1874](https://github.com/provenance-io/provenance/pull/1874).
* Switch to `github.com/cometbft/cometbft-db` `v0.7.0` (from `github.com/tendermint/tm-db` `v0.6.7`) [PR 1874](https://github.com/provenance-io/provenance/pull/1874).
* Allow NAV volume to exceed a marker's supply [PR 1883](https://github.com/provenance-io/provenance/pull/1883).

### Deprecated

* The concept of an "active" market (in the exchange module) has been removed in favor of specifying whether it accepts orders [#1789](https://github.com/provenance-io/provenance/issues/1789).
  * The `MarketUpdateEnabled` endpoint has been deprecated and is no longer usable. It is replaced with the `MarketUpdateAcceptingOrders` endpoint.
  * `MsgMarketUpdateEnabledRequest` is replaced with `MsgMarketUpdateAcceptingOrdersRequest`.
  * `MsgMarketUpdateEnabledResponse` is replaced with `MsgMarketUpdateAcceptingOrdersResponse`.
  * `EventMarketEnabled` is replaced with `EventMarketOrdersEnabled`.
  * `EventMarketDisabled` is replaced with `EventMarketOrdersDisabled`.

### Bug Fixes

* Remove deleted marker send deny entries [#1666](https://github.com/provenance-io/provenance/issues/1666).
* Update protos, naming, and documentation to use mills for NAVs [#1813](https://github.com/provenance-io/provenance/issues/1813).
* Update marker transfer to work with groups [#1818](https://github.com/provenance-io/provenance/issues/1818).
* Prevent funds from going to or from a marker without the transfer agent having deposit or withdraw access (respectively) [#1834](https://github.com/provenance-io/provenance/issues/1834).
* Ensure the store loader isn't nil when the handling an upgrade [PR 1852](https://github.com/provenance-io/provenance/pull/1852).
* Fix `MarkerTransferAuthorization` validation to ensure the coins and addresses are all valid [PR 1856](https://github.com/provenance-io/provenance/pull/1856).
* In `MarketCommitmentSettle`, only consume the settlement fee if the settlement succeeds [#1703](https://github.com/provenance-io/provenance/issues/1703).

### Client Breaking

* The `provenanced tx exchange market-enabled` command has been changed to `provenanced tx exchange market-accepting-orders` [#1789](https://github.com/provenance-io/provenance/issues/1789).

### API Breaking

* The `MarketUpdateEnabled` endpoint has been deprecated and replaced with `MarketUpdateAcceptingOrders` along with its request, response, and events [#1789](https://github.com/provenance-io/provenance/issues/1789).
  The old endpoint is no longer usable. See the Deprecated section for more details.
* Accounts that have transfer access in a marker are no longer allowed to do forced transfers ([#1829](https://github.com/provenance-io/provenance/issues/1829)).
  Accounts must now have the force_transfer access for that.

### Dependencies

- Bump `bufbuild/buf-setup-action` from 1.27.1 to 1.30.0 ([#1724](https://github.com/provenance-io/provenance/pull/1724), [#1744](https://github.com/provenance-io/provenance/pull/1744), [#1750](https://github.com/provenance-io/provenance/pull/1750), [#1825](https://github.com/provenance-io/provenance/pull/1825), [#1871](https://github.com/provenance-io/provenance/pull/1871))
- Bump `github.com/google/uuid` from 1.3.1 to 1.6.0 ([#1723](https://github.com/provenance-io/provenance/pull/1723), [#1781](https://github.com/provenance-io/provenance/pull/1781), [#1819](https://github.com/provenance-io/provenance/pull/1819))
- Bump `github.com/gorilla/mux` from 1.8.0 to 1.8.1 ([#1734](https://github.com/provenance-io/provenance/pull/1734))
- Bump `golang.org/x/text` from 0.13.0 to 0.14.0 ([#1735](https://github.com/provenance-io/provenance/pull/1735))
- Bump `cosmossdk.io/math` from 1.1.2 to 1.3.0 ([#1739](https://github.com/provenance-io/provenance/pull/1739), [#1857](https://github.com/provenance-io/provenance/pull/1857))
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
- Bump `google.golang.org/grpc` from 1.59.0 to 1.62.1 ([#1794](https://github.com/provenance-io/provenance/pull/1794), [#1820](https://github.com/provenance-io/provenance/pull/1820), [#1842](https://github.com/provenance-io/provenance/pull/1842), [#1850](https://github.com/provenance-io/provenance/pull/1850), [#1864](https://github.com/provenance-io/provenance/pull/1864))
- Bump `golang.org/x/crypto` from 0.14.0 to 0.17.0 ([#1788](https://github.com/provenance-io/provenance/pull/1788))
- Bump `cosmossdk.io/errors` from 1.0.0 to 1.0.1 ([#1806](https://github.com/provenance-io/provenance/pull/1806))
- Bump `actions/cache` from 3 to 4 ([#1817](https://github.com/provenance-io/provenance/pull/1817))
- Bump `codecov/codecov-action` from 3 to 4 ([#1828](https://github.com/provenance-io/provenance/pull/1828))
- Bump `peter-evans/create-pull-request` from 5.0.2 to 6.0.2 ([#1827](https://github.com/provenance-io/provenance/pull/1827), [#1858](https://github.com/provenance-io/provenance/pull/1858), [#1872](https://github.com/provenance-io/provenance/pull/1872))
- Bump `github.com/rs/zerolog` from 1.31.0 to 1.32.0 ([#1832](https://github.com/provenance-io/provenance/pull/1832))
- Bump `serde-json-wasm` from 1.0.0 to 1.0.1 in /testutil/contracts/rate-limiter ([#1836](https://github.com/provenance-io/provenance/pull/1836))
- Bump `serde-json-wasm` from 0.5.1 to 0.5.2 in /testutil/contracts/counter ([#1837](https://github.com/provenance-io/provenance/pull/1837))
- Bump `serde-json-wasm` from 0.5.1 to 0.5.2 in /testutil/contracts/echo ([#1838](https://github.com/provenance-io/provenance/pull/1838))
- Bump `golangci/golangci-lint-action` from 3 to 4 ([#1840](https://github.com/provenance-io/provenance/pull/1840))
- Bump `github.com/cosmos/cosmos-sdk` from v0.46.13-pio-2 to v0.46.13-pio-4 ([#1848](https://github.com/provenance-io/provenance/pull/1848), [#1874](https://github.com/provenance-io/provenance/pull/1874))
- Bump `github.com/golang/protobuf` from 1.5.3 to 1.5.4 ([#1863](https://github.com/provenance-io/provenance/pull/1863))
- Bump `github.com/stretchr/testify` from 1.8.4 to 1.9.0 ([#1860](https://github.com/provenance-io/provenance/pull/1860))
- Bump `github.com/cosmos/ibc-go/v6` from v6.2.0-pio-1 to v6.2.0-pio-2 ([#1874](https://github.com/provenance-io/provenance/pull/1874))
- Bump `github.com/CosmWasm/wasmd` from v0.30.0-pio-6 to v0.30.0-pio-7 ([#1874](https://github.com/provenance-io/provenance/pull/1874))
- Bump `github.com/cosmos/iavl` from v0.19.6 to v0.20.1 ([#1874](https://github.com/provenance-io/provenance/pull/1874))

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.17.1...v1.18.0
