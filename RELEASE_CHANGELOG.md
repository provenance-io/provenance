## [v1.17.0-rc2](https://github.com/provenance-io/provenance/releases/tag/v1.17.0-rc2) - 2023-11-02

### Features

* Add the (empty) `saffron-rc2` upgrade [#1699](https://github.com/provenance-io/provenance/issues/1699).

### Improvements

* Wrote unit tests on the keeper methods [#1699](https://github.com/provenance-io/provenance/issues/1699).
* During `FillBids`, the seller settlement fee is now calculated on the total price instead of each order individually [#1699](https://github.com/provenance-io/provenance/issues/1699).
* In the `OrderFeeCalc` query, ensure the market exists [#1699](https://github.com/provenance-io/provenance/issues/1699).

### Bug Fixes

* During `InitGenesis`, ensure LastOrderId is at least the largest order id [#1699](https://github.com/provenance-io/provenance/issues/1699).
* Properly populate the permissions lists when reading access grants from state [#1699](https://github.com/provenance-io/provenance/issues/1699).
* Fixed the paginated order queries to properly look up orders [#1699](https://github.com/provenance-io/provenance/issues/1699).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.17.0-rc1...v1.17.0-rc2
* https://github.com/provenance-io/provenance/compare/v1.16.0...v1.17.0-rc2

---

## [v1.17.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.17.0-rc1) - 2023-10-18

### Features

* Create the `x/exchange` module which facilitates the buying and selling of assets [#1658](https://github.com/provenance-io/provenance/issues/1658).
  Assets and funds remain in their owner's account (with a hold on them) until the order is settled (or cancelled).
  Market's are created to manage order matching and define fees.
  The chain will receive a portion of the fees a market collects.
* Allow marker's transfer authority to prevent transfer of restricted coin with deny list on send [#1518](https://github.com/provenance-io/provenance/issues/1518).
* Add net asset value to markers [#1328](https://github.com/provenance-io/provenance/issues/1328).
* Add ICQHost and Oracle module to allow cross chain oracle queries [#1497](https://github.com/provenance-io/provenance/issues/1497).
* New `GetByAddr` metadata query [#1443](https://github.com/provenance-io/provenance/issues/1443).
* Add Trigger module queries to stargate whitelist for smart contracts [#1636](https://github.com/provenance-io/provenance/issues/1636)
* Added the saffron upgrade handlers [PR 1648](https://github.com/provenance-io/provenance/pull/1648).
* Create the `x/hold` module which facilitates locking funds in an owners account [#1607](https://github.com/provenance-io/provenance/issues/1607).
  Funds with a hold on them cannot be transferred until the hold is removed.
  Management of holds is internal, but there are queries for looking up holds on accounts.
  Holds are also reflected in the `x/bank` module's `SpendableBalances` query.
* Add new MaxSupply param to marker module and deprecate MaxTotalSupply. [#1292](https://github.com/provenance-io/provenance/issues/1292).
* Add hidden docgen command to output documentation in different formats. [#1468](https://github.com/provenance-io/provenance/issues/1468).
* Add ics20 marker creation for receiving marker via ibc sends [#1127](https://github.com/provenance-io/provenance/issues/1127).

### Improvements

* Add IBC-Hooks module for Axelar GMP support [PR 1659](https://github.com/provenance-io/provenance/pull/1659)
* Update ibcnet ports so they don't conflict with host machine. [#1622](https://github.com/provenance-io/provenance/issues/1622)
* Replace custom ibc-go v6.1.1 fork with official module.  [#1616](https://github.com/provenance-io/provenance/issues/1616)
* Migrate `msgfees` gov proposals to v1. [#1328](https://github.com/provenance-io/provenance/issues/1328)
* Updated metadata queries to optionally include the request and id info [#1443](https://github.com/provenance-io/provenance/issues/1443).
  The request is now omitted by default, but will be included if `include_request` is `true`.
  The id info is still included by default, but will be excluded if `exclude_id_info` is `true`.
* Removed the quicksilver upgrade handlers [PR 1648](https://github.com/provenance-io/provenance/pull/1648).
* Bump cometbft to v0.34.29 (from v0.34.28) [PR 1649](https://github.com/provenance-io/provenance/pull/1649).
* Add genesis/init for Marker module send deny list addresses. [#1660](https://github.com/provenance-io/provenance/issues/1660)
* Add automatic changelog entries for dependabot. [#1674](https://github.com/provenance-io/provenance/issues/1674)
* Ensure IBC marker has matching supply [#1706](https://github.com/provenance-io/provenance/issues/1706).

### Bug Fixes

* Fix ibcnet relayer creating multiple connections on restart [#1620](https://github.com/provenance-io/provenance/issues/1620).
* Fix for incorrect resource-id type casting on contract specification [#1647](https://github.com/provenance-io/provenance/issues/1647).
* Allow restricted coins to be quarantined [#1626](https://github.com/provenance-io/provenance/issues/1626).
* Prevent marker forced transfers from module accounts [#1626](https://github.com/provenance-io/provenance/issues/1626).
* Change config load order so custom.toml can override other config. [#1262](https://github.com/provenance-io/provenance/issues/1262)
* Fix the saffron and saffron-rc1 upgrade handlers to add correct ibchooks store key [PR 1715](https://github.com/provenance-io/provenance/pull/1715).

### Client Breaking

* Metadata query responses no longer include the request by default [#1443](https://github.com/provenance-io/provenance/issues/1443).
  They are still available by setting the `include_request` flag in the requests.
* The `provenanced query metadata get` command has been changed to use the new `GetByAddr` query [#1443](https://github.com/provenance-io/provenance/issues/1443).
  The command can now take in multiple ids.
  The output of this command reflects the `GetByAddrResponse` instead of specific type queries.
  The command no longer has any `--include-<thing>` flags since they don't pertain to the `GetByAddr` query.
  The specific queries (e.g. `provenanced query metadata scope`) are still available with all appropriate flags.

### Dependencies

- Bump `google.golang.org/grpc` from 1.56.1 to 1.59.0 ([#1624](https://github.com/provenance-io/provenance/pull/1624), [#1635](https://github.com/provenance-io/provenance/pull/1635), [#1672](https://github.com/provenance-io/provenance/pull/1672), [#1685](https://github.com/provenance-io/provenance/pull/1685), [#1689](https://github.com/provenance-io/provenance/pull/1689), [#1710](https://github.com/provenance-io/provenance/pull/1710))
- Bump `crazy-max/ghaction-import-gpg` from 5 to 6 ([#1677](https://github.com/provenance-io/provenance/pull/1677))
- Bump `golang.org/x/text` from 0.12.0 to 0.13.0 ([#1667](https://github.com/provenance-io/provenance/pull/1667))
- Bump `actions/checkout` from 3 to 4 ([#1668](https://github.com/provenance-io/provenance/pull/1668))
- Bump `bufbuild/buf-breaking-action` from 1.1.2 to 1.1.3 ([#1663](https://github.com/provenance-io/provenance/pull/1663))
- Bump `cosmossdk.io/math` from 1.0.1 to 1.1.2 ([#1656](https://github.com/provenance-io/provenance/pull/1656))
- Bump `github.com/google/uuid` from 1.3.0 to 1.3.1 ([#1657](https://github.com/provenance-io/provenance/pull/1657))
- Bump `golangci/golangci-lint-action` from 3.6.0 to 3.7.0 ([#1651](https://github.com/provenance-io/provenance/pull/1651))
- Bump `bufbuild/buf-setup-action` from 1.21.0 to 1.27.1 ([#1610](https://github.com/provenance-io/provenance/pull/1610), [#1613](https://github.com/provenance-io/provenance/pull/1613), [#1631](https://github.com/provenance-io/provenance/pull/1631), [#1632](https://github.com/provenance-io/provenance/pull/1632), [#1642](https://github.com/provenance-io/provenance/pull/1642), [#1645](https://github.com/provenance-io/provenance/pull/1645), [#1650](https://github.com/provenance-io/provenance/pull/1650), [#1694](https://github.com/provenance-io/provenance/pull/1694), [#1711](https://github.com/provenance-io/provenance/pull/1711))
- Bump `cometbft to v0.34.29 `(from v0.34.28) ([#1649](https://github.com/provenance-io/provenance/pull/1649))
- Bump `golang.org/x/text` from 0.11.0 to 0.12.0 ([#1644](https://github.com/provenance-io/provenance/pull/1644))
- Bump `github.com/rs/zerolog` from 1.29.1 to 1.31.0 ([#1639](https://github.com/provenance-io/provenance/pull/1639), [#1691](https://github.com/provenance-io/provenance/pull/1691))
- Bump `github.com/cosmos/ibc-go/v6` from 6.1.1 to 6.2.0 ([#1629](https://github.com/provenance-io/provenance/pull/1629))
- Bump `cosmossdk.io/errors` from 1.0.0-beta.7 to 1.0.0 ([#1628](https://github.com/provenance-io/provenance/pull/1628))
- Bump `google.golang.org/protobuf` from 1.30.0 to 1.31.0 ([#1611](https://github.com/provenance-io/provenance/pull/1611))
- Bump `docker/setup-buildx-action` from 2 to 3 ([#1681](https://github.com/provenance-io/provenance/pull/1681))
- Bump `docker/login-action` from 2 to 3 ([#1680](https://github.com/provenance-io/provenance/pull/1680))
- Bump `docker/build-push-action` from 4 to 5 ([#1679](https://github.com/provenance-io/provenance/pull/1679))
- Bump `github.com/spf13/viper` from 1.16.0 to 1.17.0 ([#1695](https://github.com/provenance-io/provenance/pull/1695))
- Bump `github.com/otiai10/copy` from 1.12.0 to 1.14.0 ([#1693](https://github.com/provenance-io/provenance/pull/1693))
- Bump `stefanzweifel/git-auto-commit-action` from 4 to 5 ([#1696](https://github.com/provenance-io/provenance/pull/1696))
- Bump `golang.org/x/net` from 0.15.0 to 0.17.0 ([#1704](https://github.com/provenance-io/provenance/pull/1704))
- Bump `bufbuild/buf-lint-action` from 1.0.3 to 1.1.0 ([#1705](https://github.com/provenance-io/provenance/pull/1705))

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.16.0...v1.17.0-rc1
