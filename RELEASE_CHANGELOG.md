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

