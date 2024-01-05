<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a message and either
an issue number or pull request number using one of these formats:

* message #<issue-number>

If there is no issue number, you can add a reference to a Pull Request like this:
* message PR<pull-request-number>

The issue numbers and pull request numbers will later be link-ified during the release process
so you do not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->

## [Unreleased]

### Features

* Add the ibcratelimit module [#1498](https://github.com/provenance-io/provenance/issues/1498).
* Add CLI commands for the exchange module endpoints and queries [#1701](https://github.com/provenance-io/provenance/issues/1701).
* Add CLI command to generate autocomplete shell scripts [#1762](https://github.com/provenance-io/provenance/pull/1762).
* Create CLI commands for adding a market to a genesis file [#1757](https://github.com/provenance-io/provenance/issues/1757).

### Improvements

* Add upgrade handler for 1.18 [#1756](https://github.com/provenance-io/provenance/pull/1756).
* Updated documentation for each module to work with docusaurus [PR 1763](https://github.com/provenance-io/provenance/pull/1763).
* Create a default market in `make run`, `localnet`, `devnet` and the `provenanced testnet` command [#1757](https://github.com/provenance-io/provenance/issues/1757).
* Remove the rust upgrade handlers [PR 1774](https://github.com/provenance-io/provenance/pull/1774).

### Dependencies

- Bump `bufbuild/buf-setup-action` from 1.27.1 to 1.28.1 ([#1724](https://github.com/provenance-io/provenance/pull/1724), [#1744](https://github.com/provenance-io/provenance/pull/1744), [#1750](https://github.com/provenance-io/provenance/pull/1750))
- Bump `github.com/google/uuid` from 1.3.1 to 1.5.0 ([#1723](https://github.com/provenance-io/provenance/pull/1723), [#1781](https://github.com/provenance-io/provenance/pull/1781))
- Bump `github.com/gorilla/mux` from 1.8.0 to 1.8.1 ([#1734](https://github.com/provenance-io/provenance/pull/1734))
- Bump `golang.org/x/text` from 0.13.0 to 0.14.0 ([#1735](https://github.com/provenance-io/provenance/pull/1735))
- Bump `github.com/spf13/cobra` from 1.7.0 to 1.8.0 ([#1733](https://github.com/provenance-io/provenance/pull/1733))
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
- Bump `google.golang.org/grpc` from 1.59.0 to 1.60.1 ([#1794](https://github.com/provenance-io/provenance/pull/1794))
- Bump `golang.org/x/crypto` from 0.14.0 to 0.17.0 ([#1788](https://github.com/provenance-io/provenance/pull/1788))

---

## [v1.17.0](https://github.com/provenance-io/provenance/releases/tag/v1.17.0) - 2023-11-13

### Features

* Create the `x/exchange` module which facilitates the buying and selling of assets [#1658](https://github.com/provenance-io/provenance/issues/1658), [#1699](https://github.com/provenance-io/provenance/issues/1699), [#1700](https://github.com/provenance-io/provenance/issues/1700).
  Assets and funds remain in their owner's account (with a hold on them) until the order is settled (or cancelled).
  Market's are created to manage order matching and define fees.
  The chain will receive a portion of the fees a market collects.
* Allow marker's transfer authority to prevent transfer of restricted coin with deny list on send [#1518](https://github.com/provenance-io/provenance/issues/1518).
* Add net asset value to markers [#1328](https://github.com/provenance-io/provenance/issues/1328).
* Add ICQHost and Oracle module to allow cross chain oracle queries [#1497](https://github.com/provenance-io/provenance/issues/1497).
* New `GetByAddr` metadata query [#1443](https://github.com/provenance-io/provenance/issues/1443).
* Add Trigger module queries to stargate whitelist for smart contracts [#1636](https://github.com/provenance-io/provenance/issues/1636).
* Added the saffron upgrade handlers [PR 1648](https://github.com/provenance-io/provenance/pull/1648).
  * Add the ICQ, Oracle, IBC Hooks, Hold, and Exchange modules.
  * Run module migrations.
  * Set IBC Hooks params [PR 1659](https://github.com/provenance-io/provenance/pull/1659).
  * Remove inactive validators.
  * Migrate marker max supplies to BigInt [#1292](https://github.com/provenance-io/provenance/issues/1292).
  * Add initial marker NAVs [PR 1712](https://github.com/provenance-io/provenance/pull/1712).
  * Create denom metadata for IBC markers [#1728](https://github.com/provenance-io/provenance/issues/1728).
* Create the `x/hold` module which facilitates locking funds in an owners account [#1607](https://github.com/provenance-io/provenance/issues/1607).
  Funds with a hold on them cannot be transferred until the hold is removed.
  Management of holds is internal, but there are queries for looking up holds on accounts.
  Holds are also reflected in the `x/bank` module's `SpendableBalances` query.
* Add new MaxSupply param to marker module and deprecate MaxTotalSupply [#1292](https://github.com/provenance-io/provenance/issues/1292).
* Add hidden docgen command to output documentation in different formats [#1468](https://github.com/provenance-io/provenance/issues/1468).
* Add ics20 marker creation for receiving marker via ibc sends [#1127](https://github.com/provenance-io/provenance/issues/1127).

### Improvements

* Add IBC-Hooks module for Axelar GMP support [PR 1659](https://github.com/provenance-io/provenance/pull/1659).
* Update ibcnet ports so they don't conflict with host machine [#1622](https://github.com/provenance-io/provenance/issues/1622).
* Replace custom ibc-go v6.1.1 fork with official module [#1616](https://github.com/provenance-io/provenance/issues/1616).
* Migrate `msgfees` gov proposals to v1 [#1328](https://github.com/provenance-io/provenance/issues/1328).
* Updated metadata queries to optionally include the request and id info [#1443](https://github.com/provenance-io/provenance/issues/1443).
  The request is now omitted by default, but will be included if `include_request` is `true`.
  The id info is still included by default, but will be excluded if `exclude_id_info` is `true`.
* Removed the quicksilver upgrade handlers [PR 1648](https://github.com/provenance-io/provenance/pull/1648).
* Bump cometbft to v0.34.29 (from v0.34.28) [PR 1649](https://github.com/provenance-io/provenance/pull/1649).
* Add genesis/init for Marker module send deny list addresses [#1660](https://github.com/provenance-io/provenance/issues/1660).
* Add automatic changelog entries for dependabot [#1674](https://github.com/provenance-io/provenance/issues/1674).
* Ensure IBC marker has matching supply [#1706](https://github.com/provenance-io/provenance/issues/1706).
* Add publishing of docker arm64 container builds [#1634](https://github.com/provenance-io/provenance/issues/1634).
* Add additional logging to trigger module [#1718](https://github.com/provenance-io/provenance/issues/1718).
* When the exchange module settles orders, update the marker net-asset-values [#1736](https://github.com/provenance-io/provenance/pull/1736).
* Add the EventTriggerDetected and EventTriggerExecuted events [#1717](https://github.com/provenance-io/provenance/issues/1717).

### Bug Fixes

* Fix ibcnet relayer creating multiple connections on restart [#1620](https://github.com/provenance-io/provenance/issues/1620).
* Fix for incorrect resource-id type casting on contract specification [#1647](https://github.com/provenance-io/provenance/issues/1647).
* Allow restricted coins to be quarantined [#1626](https://github.com/provenance-io/provenance/issues/1626).
* Prevent marker forced transfers from module accounts [#1626](https://github.com/provenance-io/provenance/issues/1626).
* Change config load order so custom.toml can override other config [#1262](https://github.com/provenance-io/provenance/issues/1262).
* Fixed denom metadata source chain-id retrieval for new ibc markers [#1726](https://github.com/provenance-io/provenance/issues/1726).

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

* https://github.com/provenance-io/provenance/compare/v1.16.0...v1.17.0

---

## [v1.16.0](https://github.com/provenance-io/provenance/releases/tag/v1.16.0) - 2023-06-23

### Features

* Add support to add/remove required attributes for a restricted marker. [#1512](https://github.com/provenance-io/provenance/issues/1512)
* Add trigger module for delayed execution. [#1462](https://github.com/provenance-io/provenance/issues/1462)
* Add support to update the `allow_forced_transfer` field of a restricted marker [#1546](https://github.com/provenance-io/provenance/issues/1546).
* Add expiration date value to `attribute` [#1435](https://github.com/provenance-io/provenance/issues/1435).
* Add endpoints to update the value owner address of scopes [#1329](https://github.com/provenance-io/provenance/issues/1329).
* Add pre-upgrade command that updates config files to newest format and sets `consensus.timeout_commit` to `1500ms` [PR 1594](https://github.com/provenance-io/provenance/pull/1594), [PR 1600](https://github.com/provenance-io/provenance/pull/1600).

### Improvements

* Bump go to `1.20` (from `1.18`) [#1539](https://github.com/provenance-io/provenance/issues/1539).
* Bump golangci-lint to `v1.52.2` (from `v1.48`) [#1539](https://github.com/provenance-io/provenance/issues/1539).
  * New `make golangci-lint` target created for installing golangci-lint.
  * New `make golangci-lint-update` target created for installing the current version even if you already have a version installed.
* Add marker deposit access check for sends to marker escrow account [#1525](https://github.com/provenance-io/provenance/issues/1525).
* Add support for `name` owner to execute `MsgModifyName` transaction [#1536](https://github.com/provenance-io/provenance/issues/1536).
* Add usage of `AddGovPropFlagsToCmd` and `ReadGovPropFlags` cli for `GetModifyNameCmd` [#1542](https://github.com/provenance-io/provenance/issues/1542).
* Bump Cosmos-SDK to `v0.46.10-pio-4` (from `v0.46.10-pio-3`) for the `SendRestrictionFn` changes [PR 1506](https://github.com/provenance-io/provenance/pull/1506).
* Switch to using a `SendRestrictionFn` for restricting sends of marker funds [PR 1506](https://github.com/provenance-io/provenance/pull/1506).
* Create `rust` upgrade handlers [PR 1549](https://github.com/provenance-io/provenance/pull/1549).
* Remove mutation of store from `attribute` keeper iterators [#1557](https://github.com/provenance-io/provenance/issues/1557).
* Bumped ibc-go to 6.1.1 [#1563](https://github.com/provenance-io/provenance/pull/1563).
* Update `marker` module spec documentation with new proto references [#1580](https://github.com/provenance-io/provenance/pull/1580).
* Bumped `wasmd` to v0.30.0-pio-5 and `wasmvm` to v1.2.4 [#1582](https://github.com/provenance-io/provenance/pull/1582).
* Inactive validator delegation cleanup process [#1556](https://github.com/provenance-io/provenance/issues/1556).
* Bump Cosmos-SDK to [v0.46.13-pio-1](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.13-pio-1/RELEASE_NOTES.md) (from `v0.46.10-pio-4`) [PR 1585](https://github.com/provenance-io/provenance/pull/1585).

### Bug Fixes

* Bring back some proto messages that were deleted but still needed for historical queries [#1554](https://github.com/provenance-io/provenance/issues/1554).
* Fix the `MsgModifyNameRequest` endpoint to properly clean up old index data [PR 1565](https://github.com/provenance-io/provenance/pull/1565).
* Add `NewUpdateAccountAttributeExpirationCmd` to the CLI [#1592](https://github.com/provenance-io/provenance/issues/1592).
* Fix `minimum-gas-prices` from sometimes getting unset in the configs [PR 1594](https://github.com/provenance-io/provenance/pull/1594).

### API Breaking

* Add marker deposit access check for sends to marker escrow account.  Will break any current address that is sending to the
marker escrow account if it does not have deposit access.  In order for it to work, deposit access needs to be added.  This can be done using the `MsgAddAccessRequest` tx  [#1525](https://github.com/provenance-io/provenance/issues/1525).
* `MsgMultiSend` is now limited to a single `Input` [PR 1506](https://github.com/provenance-io/provenance/pull/1506).
* SDK errors returned from Metadata module endpoints [#978](https://github.com/provenance-io/provenance/issues/978).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.15.2...v1.16.0

---

## [v1.15.2](https://github.com/provenance-io/provenance/releases/tag/v1.15.2) - 2023-06-08

### Bug Fixes

* Address the [Barberry security advisory](https://forum.cosmos.network/t/cosmos-sdk-security-advisory-barberry/10825) [PR 1576](https://github.com/provenance-io/provenance/pull/1576)

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.15.1...v1.15.2

---

## [v1.15.1](https://github.com/provenance-io/provenance/releases/tag/v1.15.1) - 2023-06-01

### Improvements

* Bumped ibc-go to 6.1.1 [PR 1563](https://github.com/provenance-io/provenance/pull/1563).

### Bug Fixes

* Bring back some proto messages that were deleted but still needed for historical queries [#1554](https://github.com/provenance-io/provenance/issues/1554).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.15.0...v1.15.1

---

## [v1.15.0](https://github.com/provenance-io/provenance/releases/tag/v1.15.0) - 2023-05-05

### Features

* Add support for tokens restricted marker sends with required attributes [#1256](https://github.com/provenance-io/provenance/issues/1256)
* Allow markers to be configured to allow forced transfers [#1368](https://github.com/provenance-io/provenance/issues/1368).
* Publish Provenance Protobuf API as a NPM module [#1449](https://github.com/provenance-io/provenance/issues/1449).
* Add support for account addresses by attribute name lookup [#1447](https://github.com/provenance-io/provenance/issues/1447).
* Add allow forced transfers support to creating markers from smart contracts [#1458](https://github.com/provenance-io/provenance/issues/1458).
* Metadata party rollup and optional parties [#1438](https://github.com/provenance-io/provenance/issues/1438).
* Repeated roles in a spec require multiple different parties [#1437](https://github.com/provenance-io/provenance/issues/1437).
* The `PROVENANCE` role can only be used by smart contract addresses, and vice versa [#1381](https://github.com/provenance-io/provenance/issues/1381).
* Add stargate query from wasm support [#1481](https://github.com/provenance-io/provenance/issues/1481).
* Create methods for storing and retrieving account data for accounts, markers, and scopes [#1552](https://github.com/provenance-io/provenance/issues/1552).

### Improvements

* Add the `gci` linter that enforces import group ordering. Create a 'lint-fix' make target [PR 1366](https://github.com/provenance-io/provenance/pull/1366).
* Add gRPC query to get all contract specs and record specs for a scope spec [#677](https://github.com/provenance-io/provenance/issues/677).
* Disable `cleveldb` and `badgerdb` by default [#1411](https://github.com/provenance-io/provenance/issues/1411).
  Official builds still have `cleveldb` support though.
* Expand the `additional_bindings` gRPC tag to use object form to allow for Typescript transpiling [#1405](https://github.com/provenance-io/provenance/issues/1405).
* Add attribute cli command to query account addresses by attribute name [#1451](https://github.com/provenance-io/provenance/issues/1451).
* Add removal of attributes from accounts on name deletion [#1410](https://github.com/provenance-io/provenance/issues/1410).
* Enhance ability of smart contracts to use the metadata module [#1280](https://github.com/provenance-io/provenance/issues/1280).
* Enhance the `AddMarker` endpoint to bypass some validation if issued via governance proposal [#1358](https://github.com/provenance-io/provenance/pull/1358).
  This replaces the old `AddMarkerProposal` governance proposal.
* Bump wasmvm to 1.1.2 [#1484](https://github.com/provenance-io/provenance/pull/1358).
* Documented proposing a transaction [#1489](https://github.com/provenance-io/provenance/pull/1489).
* Add marker address to add marker event [#1499](https://github.com/provenance-io/provenance/issues/1499).

### Deprecated

* The `MsgWriteRecordRequest.parties` field has been deprecated and is ignored. The parties in question are identified by the session [PR 1453](https://github.com/provenance-io/provenance/pull/1453).

### Bug Fixes

* Fix third party Protobuf workflow checks on Provenance release steps [#1339](https://github.com/provenance-io/provenance/issues/1339)
* Fix committer email format in third party Protobuf workflow (for [#1339](https://github.com/provenance-io/provenance/issues/1339)) [PR 1385](https://github.com/provenance-io/provenance/pull/1385)
* Fix `make proto-gen` [PR 1404](https://github.com/provenance-io/provenance/pull/1404).
* Fix wasmd transactions that are run by gov module [#1414](https://github.com/provenance-io/provenance/issues/1414)
* Add support for ibc transfers of restricted tokens [#1502](https://github.com/provenance-io/provenance/issues/1502).
* Fix authz + smart contract + value owner updates being too permissive [PR 1519](https://github.com/provenance-io/provenance/pull/1519).
* Fix metadata params query path in stargate whitelist [#1514](https://github.com/provenance-io/provenance/issues/1514)

### Client Breaking

* Removed the `WriteP8eContractSpec` and `P8eMemorializeContract` endpoints [#1402](https://github.com/provenance-io/provenance/issues/1402).
* Removed the `github.com/provenance-io/provenance/x/metadata/types/p8e` proto package [#1402](https://github.com/provenance-io/provenance/issues/1402).
  Users that generate code from the Provenance protos might need to delete their `p8e/` directory.
* The `write-scope` CLI command now takes in `[owners]` as semicolon-delimited parties (instead of comma-delimited `[owner-addresses]`) [PR 1453](https://github.com/provenance-io/provenance/pull/1453).
* Removed the `AddMarkerProposal` [#1358](https://github.com/provenance-io/provenance/pull/1358).
  It is replaced by putting a `MsgAddMarker` (with the `from_address` of the gov module account), in a `MsgSubmitProposal`.

### API Breaking

* Removed the `WriteP8eContractSpec` and `P8eMemorializeContract` endpoints [#1402](https://github.com/provenance-io/provenance/issues/1402).
* Removed the `AddMarkerProposal` [#1358](https://github.com/provenance-io/provenance/pull/1358).
  It is replaced by putting a `MsgAddMarker` (with the `from_address` of the gov module account), in a `MsgSubmitProposal`.

### State Machine Breaking

* The `AddScopeOwner` endpoint now adds a new owner party even if an owner already exists in the scope with that address [PR 1453](https://github.com/provenance-io/provenance/pull/1453).
  I.e. it no longer updates the role of an existing owner with the same address.

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.14.1...v1.15.0

---

## [v1.14.1](https://github.com/provenance-io/provenance/releases/tag/v1.14.1) - 2023-02-28

### Improvements

* Bump Cosmos-SDK to `v0.46.10-pio-2` (from `v0.46.10-pio-1`). [PR 1396](https://github.com/provenance-io/provenance/pull/1396). \
  See the following `RELEASE_NOTES.md` for details: \
  [v0.46.10-pio-2](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.10-pio-2/RELEASE_NOTES.md). \
  Full Commit History: https://github.com/provenance-io/cosmos-sdk/compare/v0.46.10-pio-1...v0.46.10-pio-2

### Bug Fixes

* Fix `start` using default home directory [PR 1393](https://github.com/provenance-io/provenance/pull/1393).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.14.0...v1.14.1

---

## [v1.14.0](https://github.com/provenance-io/provenance/releases/tag/v1.14.0) - 2023-02-23

The Provenance Blockchain `v1.14.0` release includes several new features, improvements and bug fixes.
Noteably, support is added for state listing using plugins.
Also, new limitations are put in place preventing the concentration of voting power.
The `x/quarantine` and `x/sanction` modules have been added too.

The `paua` upgrade will increase all validators' max commission to 100% and max change in commission to 5% (if currently less than that).

### Features

* Enable ADR-038 State Listening in Provenance [PR 1334](https://github.com/provenance-io/provenance/pull/1334).
* Added the `x/quarantine` module [PR 1317](https://github.com/provenance-io/provenance/pull/1317).
* Added the `x/sanction` module [PR 1317](https://github.com/provenance-io/provenance/pull/1317).
* Added support to set a list of specific recipients allowed for send authorizations in the marker module [#1237](https://github.com/provenance-io/provenance/issues/1237).
* Added a new name governance proposal that allows the fields of a name record to be updated. [PR 1266](https://github.com/provenance-io/provenance/pull/1266).
* Added msg to add, finalize, and activate a marker in a single request [#770](https://github.com/provenance-io/provenance/issues/770).
* Staking concentration limit protection (prevents delegations to nodes with high voting power) [#1331](https://github.com/provenance-io/provenance/issues/1331).

### Improvements

* Bump Cosmos-SDK to `v0.46.10-pio-1` (from `v0.46.6-pio-1`).
  [PR 1371](https://github.com/provenance-io/provenance/pull/1371),
  [PR 1348](https://github.com/provenance-io/provenance/pull/1348),
  [PR 1334](https://github.com/provenance-io/provenance/pull/1334),
  [PR 1348](https://github.com/provenance-io/provenance/pull/1348),
  [PR 1317](https://github.com/provenance-io/provenance/pull/1317),
  [PR 1278](https://github.com/provenance-io/provenance/pull/1278). \
  See the following `RELEASE_NOTES.md` for details:
  [v0.46.10-pio-1](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.10-pio-1/RELEASE_NOTES.md),
  [v0.46.8-pio-3](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.8-pio-3/RELEASE_NOTES.md),
  [v0.46.7-pio-1](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.7-pio-1/RELEASE_NOTES.md). \
  Full Commit History: https://github.com/provenance-io/cosmos-sdk/compare/v0.46.6-pio-1...v0.46.10-pio-1
* Added assess msg fees spec documentation [#1172](https://github.com/provenance-io/provenance/issues/1172).
* Build dbmigrate and include it as an artifact with releases [#1264](https://github.com/provenance-io/provenance/issues/1264).
* Removed, MsgFees Module 50/50 Fee Split on MsgAssessCustomMsgFeeRequest [#1263](https://github.com/provenance-io/provenance/issues/1263).
* Add basis points field to MsgAssessCustomMsgFeeRequest for split of fee between Fee Module and Recipient [#1268](https://github.com/provenance-io/provenance/issues/1268).
* Updated ibc-go to v6.1 [#1273](https://github.com/provenance-io/provenance/issues/1273).
* Update adding of marker to do additional checks for ibc denoms [#1289](https://github.com/provenance-io/provenance/issues/1289).
* Add validate basic check to msg router service [#1308](https://github.com/provenance-io/provenance/issues/1308).
* Removed legacy-amino [#1275](https://github.com/provenance-io/provenance/issues/1275).
* Opened port 9091 on ibcnet container ibc1 to allow for reaching GRPC [PR 1314](https://github.com/provenance-io/provenance/pull/1314).
* Increase all validators' max commission to 100% [PR 1333](https://github.com/provenance-io/provenance/pull/1333).
* Increase all validators' max commission change rate to 5% [PR 1360](https://github.com/provenance-io/provenance/pull/1360).
* Set the immediate sanction/unsanction min deposits to 1,000,000 hash in the `paua` upgrade [PR 1345](https://github.com/provenance-io/provenance/pull/1345).

### Bug Fixes

* Update Maven publishing email to provenance [#1270](https://github.com/provenance-io/provenance/issues/1270).

### Client Breaking

* No longer sign the mac binary, and stop including it in the release [PR 1367](https://github.com/provenance-io/provenance/pull/1367).
* The `--restrict` flag has been replaced with an `--unrestrict` flag in the `tx name` commands `bind` and `root-name-proposal` [PR 1266](https://github.com/provenance-io/provenance/pull/1266).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.13.1...v1.14.0

---

## [v1.13.1](https://github.com/provenance-io/provenance/releases/tag/v1.13.1)

### Improvements

* Updated Cosmos-SDK to `v0.46.6-pio-3` (from `v0.46.6-pio-1`) [PR 1274](https://github.com/provenance-io/provenance/pull/1274).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.13.0...v1.13.1

---

## [v1.13.0](https://github.com/provenance-io/provenance/releases/tag/v1.13.0) - 2022-11-28

The `v1.13.0 - ochre` release is focused on group based management of on chain accounts and important improvements for public/private zone communication within the Provenance Blockchain network.
The `v1.13.0` release includes minor bug fixes and enhancements along with a resolution for the [dragonberry]((https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702)) security update.

### Features

* Add restricted marker transfer over ibc support [#1136](https://github.com/provenance-io/provenance/issues/1136).
* Enable the node query service [PR 1173](https://github.com/provenance-io/provenance/pull/1173).
* Add the `x/groups` module [#1007](https://github.com/provenance-io/provenance/issues/1007).
* Allow starting a `provenanced` chain using a custom denom [#1067](https://github.com/provenance-io/provenance/issues/1067).
  For running the chain locally, `make run DENOM=vspn MIN_FLOOR_PRICE=0` or `make clean localnet-start DENOM=vspn MIN_FLOOR_PRICE=0`.
* [#627](https://github.com/provenance-io/provenance/issues/627) Added Active Participation and Engagement module, see [specification](https://github.com/provenance-io/provenance/blob/main/x/reward/spec/01_concepts.md) for details.

### Improvements

* Updated Cosmos-SDK to `v0.46.6-pio-1` (from `v0.45.10-pio-4`) [PR 1235](https://github.com/provenance-io/provenance/pull/1235).
  This brings several new features and improvements. For details, see the [release notes](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.6-pio-1/RELEASE_NOTES.md) and [changelog](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.6-pio-1/CHANGELOG.md).
* Bump IBC to `v5.0.0-pio-2` (from `v2.3.0`) to add a check for SendEnabled [#1100](https://github.com/provenance-io/provenance/issues/1100), [#1158](https://github.com/provenance-io/provenance/issues/1158).
* Update wasmd to `v0.29.0-pio-1` (from `v0.26.0`) with SDK v0.46 support from notional-labs [#1015](https://github.com/provenance-io/provenance/issues/1015), [PR 1148](https://github.com/provenance-io/provenance/pull/1148).
* Allow MsgFee fees to be denoms other than `nhash` [#1067](https://github.com/provenance-io/provenance/issues/1067).
* Ignore hardcoded tx gas limit when `consensus_params.block.max_gas` is set to -1 for local nodes [#1000](https://github.com/provenance-io/provenance/issues/1000).
* Refactor the `x/marker` module's `Holding` query to utilize the `x/bank` module's new `DenomOwners` query [#995](https://github.com/provenance-io/provenance/issues/995).
  The only real difference between those two queries is that the `Holding` query accepts either a denom or marker address.
* Stop using the deprecated `Wrap` and `Wrapf` functions in the `sdk/types/errors` package in favor of those functions off specific errors, or else the `cosmossdk.io/errors` package [#1013](https://github.com/provenance-io/provenance/issues/995).
* For newly added reward's module, Voting incentive program, validator votes should count for higher shares, since they vote for all their delegations.
  This improvement allows the reward creator to introduce the multiplier to achieve the above.
* Refactored the fee handling [#1006](https://github.com/provenance-io/provenance/issues/1006):
  * Created a `MinGasPricesDecorator` to replace the `MempoolFeeDecorator` that was removed from the SDK. It makes sure the fee is greater than the validators min-gas fee.
  * Refactored the `MsgFeesDecorator` to only make sure there's enough fee provided. It no longer deducts/consumes anything and it no longer checks the payer's account.
  * Refactored the `ProvenanceDeductFeeDecorator`. It now makes sure the payer has enough in their account to cover the additional fees. It also now deducts/consumes the `floor gas price * gas`.
  * Added the `fee_payer` attribute to events of type `tx` involving fees (i.e. the ones with attributes `fee`, `min_fee_charged`, `additionalfee` and/or `baseFee`).
  * Moved the additional fees calculation logic into the msgfees keeper.
* Update `fee` event with amount charged even on failure and emit SendCoin events from `DeductFeesDistributions` [#1092](https://github.com/provenance-io/provenance/issues/1092).
* Alias the `config unpack` command to `config update`. It can be used to update config files to include new fields [PR 1233](https://github.com/provenance-io/provenance/pull/1233).
* When loading the unpacked configs, always load the defaults before reading the files (instead of only loading the defaults if the file doesn't exist) [PR 1233](https://github.com/provenance-io/provenance/pull/1233).
* Add prune command available though cosmos sdk to provenanced. [#1208](https://github.com/provenance-io/provenance/issues/1208).
* Updated name restrictions documentation [#808](https://github.com/provenance-io/provenance/issues/808).
* Update swagger files [PR 1229](https://github.com/provenance-io/provenance/pull/1229).
* Improve CodeQL workflow to run on Go file changes only [#1225](https://github.com/provenance-io/provenance/issues/1225).
* Use latest ProvWasm contract in wasm tests [#731](https://github.com/provenance-io/provenance/issues/731).
* Publish Java/Kotlin JARs to Maven for release candidates [#1223](https://github.com/provenance-io/provenance/issues/1223).
* Added two new Makefile targets to install and start the relayer [#1051] (https://github.com/provenance-io/provenance/pull/1051)
* Updated relayer scripts to make them headless for external services [#1068] (https://github.com/provenance-io/provenance/pull/1068)
* Added docker environment for testing IBC and added Makefile targets to bring this environment up/down [#1248] (https://github.com/provenance-io/provenance/pull/1248).
* Updated the dbmigrate tool to allow the user to force the source database type with the source-db-backend option [#1258] (https://github.com/provenance-io/provenance/pull/1258)
* Updated provenance-io/blockchain image to work with arm64 [#1261]. (https://github.com/provenance-io/provenance/pull/1261)

### Bug Fixes

* Fixed outdated devnet docker configurations [#1062](https://github.com/provenance-io/provenance/issues/1062).
* Fix the [Dragonberry security advisory](https://forum.cosmos.network/t/ibc-security-advisory-dragonberry/7702) [PR 1173](https://github.com/provenance-io/provenance/pull/1173).
* Fix GetParams in `msgfees` modules to return ConversionFeeDenom [#1214](https://github.com/provenance-io/provenance/issues/1214).
* Pay attention to the `iavl-disable-fastnode` config field/flag [PR 1193](https://github.com/provenance-io/provenance/pull/1193).
* Remove the workaround for the index-events configuration field (now fixed in the SDK) [#995](https://github.com/provenance-io/provenance/issues/995).

### Client Breaking

* Remove the state-listening/plugin system (and `librdkafka` dependencies) [#995](https://github.com/provenance-io/provenance/issues/995).
* Remove the custom/legacy rest endpoints from the `x/attribute`, `x/marker`, and `x/name` modules [#995](https://github.com/provenance-io/provenance/issues/995).
  * The following REST endpoints have been removed in favor of `/provenance/...` counterparts:
    * `GET` `attribute/{address}/attributes` -> `/provenance/attribute/v1/attributes/{address}`
    * `GET` `attribute/{address}/attributes/{name}` -> `/provenance/attribute/v1/attribute/{address}/{name}`
    * `GET` `attribute/{address}/scan/{suffix}` -> `/provenance/attribute/v1/attribute/{address}/scan/{suffix}`
    * `GET` `marker/all` -> `/provenance/marker/v1/all`
    * `GET` `marker/holders/{id}` -> `/provenance/marker/v1/holding/{id}`
    * `GET` `marker/detail/{id}` -> `/provenance/marker/v1/detail/{id}`
    * `GET` `marker/accesscontrol/{id}` -> `/provenance/marker/v1/accesscontrol/{id}`
    * `GET` `marker/escrow/{id}` -> `/provenance/marker/v1/escrow/{id}`
    * `GET` `marker/supply/{id}` -> `/provenance/marker/v1/supply/{id}`
    * `GET` `marker/assets/{id}` -> `/provenance/metadata/v1/ownership/{address}` (you can get the `{address}` from `/provenance/marker/v1/detail/{id}`).
    * `GET` `name/{name}` -> `/provenance/name/v1/resolve/{name}`
    * `GET` `name/{address}/names` -> `/provenance/name/v1/lookup/{address}`
  * The following REST endpoints have been removed. They do not have any REST replacement counterparts. Use GRPC instead.
    * `DELETE` `attribute/attributes` -> `DeleteAttribute(MsgDeleteAttributeRequest)`
    * `POST` `/marker/{denom}/mint` -> `Mint(MsgMintRequest)`
    * `POST` `/marker/{denom}/burn` -> `Burn(MsgBurnRequest)`
    * `POST` `/marker/{denom}/status` -> One of:
      * `Activate(MsgActivateRequest)`
      * `Finalize(MsgFinalizeRequest)`
      * `Cancel(MsgCancelRequest)`
      * `Delete(MsgDeleteRequest)`
  * The following short-form `GET` endpoints were removed in favor of longer ones:
    * `/node_info` -> `/cosmos/base/tendermint/v1beta1/node_info`
    * `/syncing` -> `/cosmos/base/tendermint/v1beta1/syncing`
    * `/blocks/latest` -> `/cosmos/base/tendermint/v1beta1/blocks/latest`
    * `/blocks/{height}` -> `/cosmos/base/tendermint/v1beta1/blocks/{height}`
    * `/validatorsets/latest` -> `/cosmos/base/tendermint/v1beta1/validatorsets/latest`
    * `/validatorsets/{height}` -> `/cosmos/base/tendermint/v1beta1/validatorsets/{height}`
  * The denom owners `GET` endpoint changed from `/cosmos/bank/v1beta1/denom_owners/{denom}` to `/cosmos/bank/v1beta1/supply/by_denom?denom={denom}`.

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.12.2...v1.13.0

---

## [v1.12.2](https://github.com/provenance-io/provenance/releases/tag/v1.12.2) - 2022-11-01

Provenance v1.12.2 enables the ability to upgrade your IAVL state store to be faster and handle errors better. This upgrade is recommended and should be done at your convenience prior to the v1.13 chain upgrade.

The IAVL store upgrade is expected to take 30 to 90 minutes. During that time, your node will be down. There will be some log output (info level), but it is sparce and there may be long periods (25+ minutes) without any new log output. Once it has started, it's best to not interrupt the process.

It is highly recommended that you do one of these two prior to the v1.13 chain upgrade:

Either

- Upgrade your node's IAVL store:
  1. Stop your node.
  2. Upgrade `provenanced` to v1.12.2.
  3. Run the command: `provenanced config set iavl-disable-fastnode false`.
  4. Restart your node. Once the upgrade has finished, your node will automatically run as normal.

Or

- Explicitly define that you don't want to upgrade your node's IAVL store:
   1. Ensure that you have `provenanced` v1.12.1 (or higher), e.g. Run the command: `provenanced version`. If you are on 1.12.0, upgrade to at least v1.12.1.
   2. Run the command: `provenanced config set iavl-disable-fastnode true`.

---

You can manually update your `app.toml` file, but using the `config set` command is the recommended method. The `iavl-disable-fastnode` field was added in v1.12.1 and most likely does not yet exist in your `app.toml` file. There are other new sections and fields too. Using the command will add them all (using defaults) as well as their descriptions. If you want to update your `app.toml` manually, the `iavl-disable-fastnode` entry should go below the `index-events` entry and before the `[telemetry]` section.

If you do nothing before the v1.13 chain upgrade, your node will most likely upgrade the IAVL store when v1.13 first runs. The v1.13 chain upgrade and migrations are expected to only take a minute. If your node is also upgrading the IAVL store at that time, it will take 30-90+ minutes.

Note: The command `provenanced config get iavl-disable-fastnode` will report a value regardless of whether the field exists in `app.toml`. As such, that command is insufficient for determining whether the value exists in the `app.toml` file.

### Improvements

* Bump Cosmos-SDK to v0.45.10-pio-4 (from v0.45.9-pio-1) [PR 1202](https://github.com/provenance-io/provenance/pull/1202)
* Allow the IAVL store to be upgraded [PR 1193](https://github.com/provenance-io/provenance/pull/1193).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.12.1...v1.12.2

---

## [v1.12.1](https://github.com/provenance-io/provenance/releases/tag/v1.12.1) - 2022-10-14

### Improvements

* Bump Cosmos-SDK to v0.45.9-pio-1 (from v0.45.5-pio-1) [PR 1159](https://github.com/provenance-io/provenance/pull/1159).

### Bug Fixes

* Bump ics23/go to Cosmos-SDK's v0.8.0 (from confio's v0.7.0) [PR 1159](https://github.com/provenance-io/provenance/pull/1159).

---

## [v1.12.2](https://github.com/provenance-io/provenance/releases/tag/v1.12.2) - 2022-11-01

Provenance v1.12.2 enables the ability to upgrade your IAVL state store to be faster and handle errors better. This upgrade is recommended and should be done at your convenience prior to the v1.13 chain upgrade.

The IAVL store upgrade is expected to take 30 to 90 minutes. During that time, your node will be down. There will be some log output (info level), but it is sparce and there may be long periods (25+ minutes) without any new log output. Once it has started, it's best to not interrupt the process.

It is highly recommended that you do one of these two prior to the v1.13 chain upgrade:

Either

- Upgrade your node's IAVL store:
  1. Stop your node.
  2. Upgrade `provenanced` to v1.12.2.
  3. Run the command: `provenanced config set iavl-disable-fastnode false`.
  4. Restart your node. Once the upgrade has finished, your node will automatically run as normal.

Or

- Explicitly define that you don't want to upgrade your node's IAVL store:
   1. Ensure that you have `provenanced` v1.12.1 (or higher), e.g. Run the command: `provenanced version`. If you are on 1.12.0, upgrade to at least v1.12.1.
   2. Run the command: `provenanced config set iavl-disable-fastnode true`.

---

You can manually update your `app.toml` file, but using the `config set` command is the recommended method. The `iavl-disable-fastnode` field was added in v1.12.1 and most likely does not yet exist in your `app.toml` file. There are other new sections and fields too. Using the command will add them all (using defaults) as well as their descriptions. If you want to update your `app.toml` manually, the `iavl-disable-fastnode` entry should go below the `index-events` entry and before the `[telemetry]` section.

If you do nothing before the v1.13 chain upgrade, your node will most likely upgrade the IAVL store when v1.13 first runs. The v1.13 chain upgrade and migrations are expected to only take a minute. If your node is also upgrading the IAVL store at that time, it will take 30-90+ minutes.

Note: The command `provenanced config get iavl-disable-fastnode` will report a value regardless of whether the field exists in `app.toml`. As such, that command is insufficient for determining whether the value exists in the `app.toml` file.

### Improvements

* Bump Cosmos-SDK to v0.45.10-pio-4 (from v0.45.9-pio-1) [PR 1202](https://github.com/provenance-io/provenance/pull/1202)
* Allow the IAVL store to be upgraded [PR 1193](https://github.com/provenance-io/provenance/pull/1193).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.12.1...v1.12.2

---

## [v1.12.1](https://github.com/provenance-io/provenance/releases/tag/v1.12.1) - 2022-10-14

### Improvements

* Bump Cosmos-SDK to v0.45.9-pio-1 (from v0.45.5-pio-1) [PR 1159](https://github.com/provenance-io/provenance/pull/1159).

### Bug Fixes

* Bump ics23/go to Cosmos-SDK's v0.8.0 (from confio's v0.7.0) [PR 1159](https://github.com/provenance-io/provenance/pull/1159).

---

## [v1.12.0](https://github.com/provenance-io/provenance/releases/tag/v1.12.0) - 2022-08-22

### Improvements

* Update the swagger files (including third-party changes). [#728](https://github.com/provenance-io/provenance/issues/728)
* Bump IBC to 2.3.0 and update third-party protos [PR 868](https://github.com/provenance-io/provenance/pull/868)
* Update docker images from `buster` to b`bullseye` [#963](https://github.com/provenance-io/provenance/issues/963)
* Add documentation for `gRPCurl` to `docs/grpcurl.md` [#953](https://github.com/provenance-io/provenance/issues/953)
* Updated to go 1.18 [#996](https://github.com/provenance-io/provenance/issues/996)
* Add docker files for local psql indexing [#997](https://github.com/provenance-io/provenance/issues/997)


### Features

* Bump Cosmos-SDK to `v0.45.4-pio-4` (from `v0.45.4-pio-2`) to utilize the new `CountAuthorization` authz grant type. [#807](https://github.com/provenance-io/provenance/issues/807)
* Update metadata module authz handling to properly call `Accept` and delete/update authorizations as they're used [#905](https://github.com/provenance-io/provenance/issues/905)
* Read the `custom.toml` config file if it exists. This is read before the other config files, and isn't managed by the `config` commands [#989](https://github.com/provenance-io/provenance/issues/989)
* Allow a msg fee to be paid to a specific address with basis points [#690](https://github.com/provenance-io/provenance/issues/690)
* Bump IBC to 5.0.0 and add support for ICA Host module [PR 1076](https://github.com/provenance-io/provenance/pull/1076)

### Bug Fixes

* Support standard flags on msgfees params query cli command [#936](https://github.com/provenance-io/provenance/issues/936)
* Fix the `MarkerTransferAuthorization` Accept function and `TransferCoin` authz handling to prevent problems when other authorization types are used [#903](https://github.com/provenance-io/provenance/issues/903)
* Bump Cosmos-SDK to `v0.45.5-pio-1` (from `v0.45.4-pio-4`) to remove buggy ADR 038 plugin system. [#983](https://github.com/provenance-io/provenance/issues/983)
* Remove ADR 038 plugin system implementation due to `AppHash` error [#983](https://github.com/provenance-io/provenance/issues/983)
* Fix fee charging to sweep remaining fees on successful transaction [#1019](https://github.com/provenance-io/provenance/issues/1019)

### State Machine Breaking

* Fix the `MarkerTransferAuthorization` Accept function and `TransferCoin` authz handling to prevent problems when other authorization types are used [#903](https://github.com/provenance-io/provenance/issues/903)
* Fix fee charging to sweep remaining fees on successful transaction [#1019](https://github.com/provenance-io/provenance/issues/1019)
* Allow a msg fee to be paid to a specific address with basis points [#690](https://github.com/provenance-io/provenance/issues/690)

---

## [v1.11.1](https://github.com/provenance-io/provenance/releases/tag/v1.11.1) - 2022-07-13

### Bug Fixes

* Add `mango` upgrade handler.
* Add new `msgfees` `NhashPerUsdMil`  default param to param space store on upgrade (PR [#875](https://github.com/provenance-io/provenance/issues/875))
* Run the module migrations as part of the mango upgrade [PR 896](https://github.com/provenance-io/provenance/pull/896)
* Update Cosmos-SDK to v0.45.4-pio-2 to fix a non-deterministic map iteration [PR 928](https://github.com/provenance-io/provenance/pull/928)

---

## [v1.11.0](https://github.com/provenance-io/provenance/releases/tag/v1.11.0) - 2022-06-13

### Features

* Add CONTROLLER, and VALIDATOR PartyTypes for contract execution. [\#824](https://github.com/provenance-io/provenance/pull/824])
* Add FeeGrant allowance support for marker escrow accounts [#406](https://github.com/provenance-io/provenance/issues/406)
* Bump Cosmos-SDK to v0.45.4-pio-1, which contains Cosmos-SDK v0.45.4 and the update to storage of the bank module's SendEnabled information. [PR 850](https://github.com/provenance-io/provenance/pull/850)
* Add `MsgAssessCustomMsgFeeRequest` to add the ability for a smart contract author to charge a custom fee [#831](https://github.com/provenance-io/provenance/issues/831)

### Bug Fixes

* Move buf.build push action to occur after PRs are merged to main branch [#838](https://github.com/provenance-io/provenance/issues/838)
* Update third party proto dependencies [#842](https://github.com/provenance-io/provenance/issues/842)

### Improvements

* Add restricted status info to name module cli queries [#806](https://github.com/provenance-io/provenance/issues/806)
* Store the bank module's SendEnabled flags directly in state instead of as part of Params. This will drastically reduce the costs of sending coins and managing markers. [PR 850](https://github.com/provenance-io/provenance/pull/850)
* Add State Sync readme [#859](https://github.com/provenance-io/provenance/issues/859)

### State Machine Breaking

* Move storage of denomination SendEnabled flags into bank module state (from Params), and update the marker module to correctly manipulate the flags in their new location. [PR 850](https://github.com/provenance-io/provenance/pull/850)

---

## [v1.10.0](https://github.com/provenance-io/provenance/releases/tag/v1.10.0) - 2022-05-11

### Summary

Provenance 1.10.0 includes upgrades to the underlying CosmWasm dependencies and adds functionality to
remove orphaned metadata in the bank module left over after markers have been deleted.

### Improvements

* Update wasmvm dependencies and update Dockerfile for localnet [#818](https://github.com/provenance-io/provenance/issues/818)
* Remove "send enabled" on marker removal and in bulk on 1.10.0 upgrade [#821](https://github.com/provenance-io/provenance/issues/821)

---

## [v1.9.0](https://github.com/provenance-io/provenance/releases/tag/v1.9.0) - 2022-04-25

### Summary

Provenance 1.9.0 brings some minor features and security improvements.

### Features

* Add `add-genesis-msg-fee` command to add msg fees to genesis.json and update Makefile to have pre-defined msg fees [#667](https://github.com/provenance-io/provenance/issues/667)
* Add msgfees summary event to be emitted when there are txs that have fees [#678](https://github.com/provenance-io/provenance/issues/678)
* Adds home subcommand to the cli's config command [#620] (https://github.com/provenance-io/provenance/issues/620)
* Add support for rocksdb and badgerdb [#702](https://github.com/provenance-io/provenance/issues/702)
* Create `dbmigrate` utility for migrating a data folder to use a different db backend [#696](https://github.com/provenance-io/provenance/issues/696)

### Improvements

* When the `start` command encounters an error, it no longer outputs command usage [#670](https://github.com/provenance-io/provenance/issues/670)
* Change max length on marker unresticted denom from 64 to 83 [#719](https://github.com/provenance-io/provenance/issues/719)
* Set prerelease to `true` for release candidates. [#666](https://github.com/provenance-io/provenance/issues/666)
* Allow authz grants to work on scope value owners [#755](https://github.com/provenance-io/provenance/issues/755)
* Bump wasmd to v0.26 (from v0.24). [#799](https://github.com/provenance-io/provenance/pull/799)

---

## [v1.8.2](https://github.com/provenance-io/provenance/releases/tag/v1.8.2) - 2022-04-22

### Summary

Provenance 1.8.2 is a point release to fix an issue with "downgrade detection" in Cosmos SDK. A panic condition
occurs in cases where no update handler is found for the last known upgrade, but the process for determining
the last known upgrade is flawed in Cosmos SDK 0.45.3. This released uses an updated Cosmos fork to patch the
issue until an official patch is released. Version 1.8.2 also adds some remaining pieces for  ADR-038 that were
missing in the 1.8.1 release.

### Bug Fixes

* Order upgrades by block height rather than name to prevent panic [\#106](https://github.com/provenance-io/cosmos-sdk/pull/106)

### Improvements

* Add remaining updates for ADR-038 support [\#786](https://github.com/provenance-io/provenance/pull/786)

---

## [v1.8.1](https://github.com/provenance-io/provenance/releases/tag/v1.8.1) - 2022-04-13

### Summary

Provenance 1.8.1 includes upgrades to the underlying Cosmos SDK and adds initial support for ADR-038.

This release addresses issues related to IAVL concurrency and Tendermint performance that resulted in occasional panics when under high-load conditions such as replay from quicksync. In particular, nodes which experienced issues with "Value missing for hash" and similar panic conditions should work properly with this release. The underlying Cosmos SDK `0.45.3` release that has been incorporated includes a number of improvements around IAVL locking and performance characteristics.

** NOTE: Although Provenance supports multiple database backends, some issues have been reported when using the `goleveldb` backend. If experiencing issues, using the `cleveldb` backend is preferred **

### Improvements

* Update Provenance to use Cosmos SDK 0.45.3 Release [\#781](https://github.com/provenance-io/provenance/issues/781)
* Plugin architecture for ADR-038 + FileStreamingService plugin [\#10639](https://github.com/cosmos/cosmos-sdk/pull/10639)
* Fix for sporadic error "panic: Value missing for hash" [\#611](https://github.com/provenance-io/provenance/issues/611)

---

## [v1.8.0](https://github.com/provenance-io/provenance/releases/tag/v1.8.0) - 2022-03-17

### Summary

Provenance 1.8.0 is focused on improving the fee structures for transactions on the blockchain. While the Cosmos SDK has traditionally offered a generic fee structure focused on gas/resource utilization, the Provenance blockchain has found that certain transactions have additional long term costs and value beyond simple resources charges. This is the reason we are adding the new MsgFee module which allows governance based control of additional fee charges on certain message types.

NOTE: The second major change in the 1.8.0 release is part of the migration process which removes many orphaned state objects that were left in 1.7.x chains. This cleanup process will require a significant amount of time to perform during the green upgrade handler execution. The upgrade will print status messages showing the progress of this process.

### Features

* Add check for `authz` grants when there are missing signatures in `metadata` transactions [#516](https://github.com/provenance-io/provenance/issues/516)
* Add support for publishing Java and Kotlin Protobuf compiled sources to Maven Central [#562](https://github.com/provenance-io/provenance/issues/562)
* Adds support for creating root name governance proposals from the cli [#599](https://github.com/provenance-io/provenance/issues/599)
* Adding of the msg based fee module [#354](https://github.com/provenance-io/provenance/issues/354)
* Upgrade provenance to 0.45 cosmos sdk release [#607](https://github.com/provenance-io/provenance/issues/607)
* Upgrade wasmd to v0.22.0 Note: this removes dependency on provenance-io's wasmd fork [#479](https://github.com/provenance-io/provenance/issues/479)
* Add support for Scope mutation via wasm Smart Contracts [#531](https://github.com/provenance-io/provenance/issues/531)
* Increase governance deposit amount and add create proposal msg fee [#632](https://github.com/provenance-io/provenance/issues/632)
* Allow attributes to be associated with scopes [#631](https://github.com/provenance-io/provenance/issues/631)

### Improvements

* Add `bank` and `authz` module query `proto` files required by `grpcurl` [#482](https://github.com/provenance-io/provenance/issues/482)
* Fix typeos in marker log statements [#502](https://github.com/provenance-io/provenance/issues/502)
* Set default coin type to network default [#534](https://github.com/provenance-io/provenance/issues/534)
* Add logger to upgrade handler [#507](https://github.com/provenance-io/provenance/issues/507)
* Allow markers to be created over existing accounts if they are not a marker and have a zero sequence [#520](https://github.com/provenance-io/provenance/issues/520)
* Removed extraneous Metadata index deletes/rewrites [#543](https://github.com/provenance-io/provenance/issues/543)
* Delete empty sessions when the last record is updated to a new session [#480](https://github.com/provenance-io/provenance/issues/480)
* Refactor the migration to be faster and have more log output [PR 586](https://github.com/provenance-io/provenance/pull/586)
* Capture all included protobufs into release zip file [#556](https://github.com/provenance-io/provenance/issues/556)
* Add Protobuf support with buf.build [#614](https://github.com/provenance-io/provenance/issues/614)
* Limit the maximum attribute value length to 1000 (down from 10,000 currently) in the `green` upgrade [#616](https://github.com/provenance-io/provenance/issues/616)
* Add additional fees for specified operations in the `green` upgrade [#616](https://github.com/provenance-io/provenance/issues/616)
  * `provenance.name.v1.MsgBindNameRequest` 10 hash (10,000,000,000 nhash)
  * `provenance.marker.v1.MsgAddMarkerRequest` 100 hash (100,000,000,000 nhash)
  * `provenance.attribute.v1.MsgAddAttributeRequest` 10 hash (10,000,000,000 nhash)
  * `provenance.metadata.v1.MsgWriteScopeRequest`  10 hash (10,000,000,000 nhash)
  * `provenance.metadata.v1.MsgP8eMemorializeContractRequest` 10 hash (10,000,000,000 nhash)
* Add integration tests for smart contracts [#392](https://github.com/provenance-io/provenance/issues/392)
* Use provwasm release artifact for smart contract tests [#731](https://github.com/provenance-io/provenance/issues/731)

### Client Breaking

* Enforce a maximum gas limit on individual transactions so that at least 20 can fit in any given block. [#681](https://github.com/provenance-io/provenance/issues/681)
  Previously transactions were only limited by their size in bytes as well as the overall gas limit on a given block.

  _With this update transactions must be no more than 5% of the maximum amount of gas allowed per block when a gas limit
  per block is set (this restriction has no effect when a gas limit has not been set).  The current limits on Provenance
  mainnet are 60,000,000 gas per block which will yield a maximum transaction size of 3,000,000 gas using this new AnteHandler
  restriction._

### Bug Fixes

* When deleting a scope, require the same permissions as when updating it [#473](https://github.com/provenance-io/provenance/issues/473)
* Allow manager to adjust grants on finalized markers [#545](https://github.com/provenance-io/provenance/issues/545)
* Add migration to re-index the metadata indexes involving addresses [#541](https://github.com/provenance-io/provenance/issues/541)
* Add migration to delete empty sessions [#480](https://github.com/provenance-io/provenance/issues/480)
* Add Java distribution tag to workflow [#624](https://github.com/provenance-io/provenance/issues/624)
* Add `msgfees` module to added store upgrades [#640](https://github.com/provenance-io/provenance/issues/640)
* Use `nhash` for base denom in gov proposal upgrade [#648](https://github.com/provenance-io/provenance/issues/648)
* Bump `cosmowasm` from `v1.0.0-beta5` to `v1.0.0-beta6` [#655](https://github.com/provenance-io/provenance/issues/655)
* Fix maven publish release version number reference [#650](https://github.com/provenance-io/provenance/issues/650)
* Add `iterator` as feature for wasm [#658](https://github.com/provenance-io/provenance/issues/658)
* String "v" from Jar artifact version number [#653](https://github.com/provenance-io/provenance/issues/653)
* Fix `wasm` contract migration failure to find contract history [#662](https://github.com/provenance-io/provenance/issues/662)

## [v1.7.6](https://github.com/provenance-io/provenance/releases/tag/v1.7.6) - 2021-12-15

* Upgrade Rosetta to v0.7.2 [#560](https://github.com/provenance-io/provenance/issues/560)

## [v1.7.5](https://github.com/provenance-io/provenance/releases/tag/v1.7.5) - 2021-10-22

### Improvements

* Update Cosmos SDK to 0.44.3 [PR 536](https://github.com/provenance-io/provenance/pull/536)

## [v1.7.4](https://github.com/provenance-io/provenance/releases/tag/v1.7.4) - 2021-10-12

### Improvements

* Update github actions to always run required tests [#508](https://github.com/provenance-io/provenance/issues/508)
* Update Cosmos SDK to 0.44.2 [PR 527](https://github.com/provenance-io/provenance/pull/527)

## [v1.7.3](https://github.com/provenance-io/provenance/releases/tag/v1.7.3) - 2021-09-30

### Bug Fixes

* Update Cosmos SDK to 0.44.1 with IAVL 0.17 to resolve locking issues in queries.
* Fix logger config being ignored [PR 510](https://github.com/provenance-io/provenance/pull/510)

## [v1.7.2](https://github.com/provenance-io/provenance/releases/tag/v1.7.2) - 2021-09-27

### Bug Fixes

* Fix for non-deterministic upgrades in cosmos sdk [#505](https://github.com/provenance-io/provenance/issues/505)

## [v1.7.1](https://github.com/provenance-io/provenance/releases/tag/v1.7.1) - 2021-09-20

### Improvements

* Ensure marker state transition validation does not panic [#492](https://github.com/provenance-io/provenance/issues/492)
* Refactor Examples for cobra cli commands to have examples [#399](https://github.com/provenance-io/provenance/issues/399)
* Verify go version on `make build` [#483](https://github.com/provenance-io/provenance/issues/483)

### Bug Fixes

* Fix marker permissions migration and add panic on `eigengrau` upgrade [#484](https://github.com/provenance-io/provenance/issues/484)
* Fixed marker with more than uint64 causes panic [#489](https://github.com/provenance-io/provenance/issues/489)
* Fixed issue with rosetta tests timing out occasionally, because the timeout was too short [#500](https://github.com/provenance-io/provenance/issues/500)

## [v1.7.0](https://github.com/provenance-io/provenance/releases/tag/v1.7.0) - 2021-09-03

### Features

* Add a single node docker based development environment [#311](https://github.com/provenance-io/provenance/issues/311)
  * Add make targets `devnet-start` and `devnet-stop`
  * Add `networks/dev/mnemonics` for adding accounts to development environment

### Improvements

* Updated some of the documentation of Metadata type bytes (prefixes) [#474](https://github.com/provenance-io/provenance/issues/474)
* Update the Marker Holding query to fully utilize pagination fields [#400](https://github.com/provenance-io/provenance/issues/400)
* Update the Metadata OSLocatorsByURI query to fully utilize pagination fields [#401](https://github.com/provenance-io/provenance/issues/401)
* Update the Metadata OSAllLocators query to fully utilize pagination fields [#402](https://github.com/provenance-io/provenance/issues/402)
* Validate `marker` before setting it to prevent panics [#491](https://github.com/provenance-io/provenance/issues/491)

### Bug Fixes

* Removed some unneeded code from the persistent record update validation [#471](https://github.com/provenance-io/provenance/issues/471)
* Fixed packed config loading bug [#487](https://github.com/provenance-io/provenance/issues/487)
* Fixed marker with more than uint64 causes panic [#489](https://github.com/provenance-io/provenance/issues/489)

## [v1.7.0](https://github.com/provenance-io/provenance/releases/tag/v1.7.0) - 2021-09-03

### Features

* Marker governance proposal are supported in cli [#367](https://github.com/provenance-io/provenance/issues/367)
* Add ability to query metadata sessions by record [#212](https://github.com/provenance-io/provenance/issues/212)
* Add Name and Symbol Cosmos features to Marker Metadata [#372](https://github.com/provenance-io/provenance/issues/372)
* Add authz support to Marker module transfer `MarkerTransferAuthorization` [#265](https://github.com/provenance-io/provenance/issues/265)
  * Add authz grant/revoke command to `marker` cli
  * Add documentation around how to grant/revoke authz [#449](https://github.com/provenance-io/provenance/issues/449)
* Add authz and feegrant modules [PR 384](https://github.com/provenance-io/provenance/pull/384)
* Add Marker governance proposal for setting denom metadata [#369](https://github.com/provenance-io/provenance/issues/369)
* Add `config` command to cli for client configuration [#394](https://github.com/provenance-io/provenance/issues/394)
* Add updated wasmd for Cosmos 0.43 [#409](https://github.com/provenance-io/provenance/issues/409)
* Add Rosetta support and automated testing [#365](https://github.com/provenance-io/provenance/issues/365)
* Update wasm parameters to only allow smart contracts to be uploaded with gov proposal [#440](https://github.com/provenance-io/provenance/issues/440)
* Update `config` command [#403](https://github.com/provenance-io/provenance/issues/403)
  * Get and set any configuration field.
  * Get or set multiple configuration fields in a single invocation.
  * Easily identify fields with changed (non-default) values.
  * Pack the configs into a single json file with only changed (non-default) values.
  * Unpack the config back into the multiple config files (that also have documentation in them).

### Bug Fixes

* Fix for creating non-coin type markers through governance addmarker proposals [#431](https://github.com/provenance-io/provenance/issues/431)
* Marker Withdraw Escrow Proposal type is properly registered [#367](https://github.com/provenance-io/provenance/issues/367)
  * Target Address field spelling error corrected in Withdraw Escrow and Increase Supply Governance Proposals.
* Fix DeleteScopeOwner endpoint to store the correct scope [PR 377](https://github.com/provenance-io/provenance/pull/377)
* Marker module import/export issues  [PR384](https://github.com/provenance-io/provenance/pull/384)
  * Add missing marker attributes to state export
  * Fix account numbering issues with marker accounts and auth module accounts during import
  * Export marker accounts as a base account entry and a separate marker module record
  * Add Marker module governance proposals, genesis, and marker operations to simulation testing [#94](https://github.com/provenance-io/provenance/issues/94)
* Fix an encoding issue with the `--page-key` CLI arguments used in paged queries [#332](https://github.com/provenance-io/provenance/issues/332)
* Fix handling of optional fields in Metadata Write messages [#412](https://github.com/provenance-io/provenance/issues/412)
* Fix cli marker new example is incorrect [#415](https://github.com/provenance-io/provenance/issues/415)
* Fix home directory setup for app export [#457](https://github.com/provenance-io/provenance/issues/457)
* Correct an error message that was providing an illegal amount of gas as an example [#425](https://github.com/provenance-io/provenance/issues/425)

### API Breaking

* Fix for missing validation for marker permissions according to marker type.  Markers of type COIN can no longer have
  the Transfer permission assigned.  Existing permission entries on Coin type markers of type Transfer are removed
  during migration [#428](https://github.com/provenance-io/provenance/issues/428)

### Improvements

* Updated to Cosmos SDK Release v0.44 to resolve security issues in v0.43 [#463](https://github.com/provenance-io/provenance/issues/463)
  * Updated to Cosmos SDK Release v0.43  [#154](https://github.com/provenance-io/provenance/issues/154)
* Updated to go 1.17 [#454](https://github.com/provenance-io/provenance/issues/454)
* Updated wasmd for Cosmos SDK Release v0.43 [#409](https://github.com/provenance-io/provenance/issues/409)
  * CosmWasm wasmvm v0.16.0 [CHANGELOG](https://github.com/CosmWasm/wasmvm/blob/v0.16.0/CHANGELOG.md)
  * CosmWasm cosmwasm v0.16.0 [CHANGELOG](https://github.com/CosmWasm/cosmwasm/blob/v0.16.0/CHANGELOG.md)
* Updated to IBC-Go Module v1.0.1 [PR 445](https://github.com/provenance-io/provenance/pull/445)
* Updated log message for circulation adjustment [#381](https://github.com/provenance-io/provenance/issues/381)
* Updated third party proto files to pull from cosmos 0.43 [#391](https://github.com/provenance-io/provenance/issues/391)
* Removed legacy api endpoints [#380](https://github.com/provenance-io/provenance/issues/380)
* Removed v039 and v040 migrations [#374](https://github.com/provenance-io/provenance/issues/374)
* Dependency Version Updates
  * Build/CI - cache [PR 420](https://github.com/provenance-io/provenance/pull/420), workflow clean up
  [PR 417](https://github.com/provenance-io/provenance/pull/417), diff action [PR 418](https://github.com/provenance-io/provenance/pull/418)
  code coverage [PR 416](https://github.com/provenance-io/provenance/pull/416) and [PR 439](https://github.com/provenance-io/provenance/pull/439),
  setup go [PR 419](https://github.com/provenance-io/provenance/pull/419), [PR 451](https://github.com/provenance-io/provenance/pull/451)
  * Google UUID 1.3.0 [PR 446](https://github.com/provenance-io/provenance/pull/446)
  * GRPC 1.3.0 [PR 443](https://github.com/provenance-io/provenance/pull/443)
  * cast 1.4.1 [PR 442](https://github.com/provenance-io/provenance/pull/442)
* Updated `provenanced init` for better testnet support and defaults [#403](https://github.com/provenance-io/provenance/issues/403)
* Fixed some example address to use the appropriate prefix [#453](https://github.com/provenance-io/provenance/issues/453)

## [v1.6.0](https://github.com/provenance-io/provenance/releases/tag/v1.6.0) - 2021-08-23

### Bug Fixes

* Fix for creating non-coin type markers through governance addmarker proposals [#431](https://github.com/provenance-io/provenance/issues/431)
* Upgrade handler migrates usdf.c to the right marker_type.

## [v1.5.0](https://github.com/provenance-io/provenance/releases/tag/v1.5.0) - 2021-06-23

### Features

* Update Cosmos SDK to 0.42.6 with Tendermint 0.34.11 [#355](https://github.com/provenance-io/provenance/issues/355)
  * Refund gas support added to gas meter trace
  * `ibc-transfer` now contains an `escrow-address` command for querying current escrow balances
* Add `update` and `delete-distinct` attributes to `attribute` module [#314](https://github.com/provenance-io/provenance/issues/314)
* Add support to `metadata` module for adding and removing contract specifications to scope specification [#302](https://github.com/provenance-io/provenance/issues/302)
  * Added `MsgAddContractSpecToScopeSpecRequest`and `MsgDeleteContractSpecFromScopeSpecRequest` messages for adding/removing
  * Added cli commands for adding/removing
* Add smart contract query support to the `metadata` module [#65](https://github.com/provenance-io/provenance/issues/65)

### API Breaking

* Redundant account parameter was removed from Attribute module SetAttribute API. [PR 348](https://github.com/provenance-io/provenance/pull/348)

### Bug Fixes

* Value owner changes are independent of scope owner signature requirements after transfer [#347](https://github.com/provenance-io/provenance/issues/347)
* Attribute module allows removal of orphan attributes, attributes against root names [PR 348](https://github.com/provenance-io/provenance/pull/348)
* `marker` cli query for marker does not cast marker argument to lower case [#329](https://github.com/provenance-io/provenance/issues/329)

### Improvements

* Bump `wasmd` to v0.17.0 [#345](https://github.com/provenance-io/provenance/issues/345)
* Attribute module simulation support [#25](https://github.com/provenance-io/provenance/issues/25)
* Add transfer cli command to `marker` module [#264](https://github.com/provenance-io/provenance/issues/264)
* Refactor `name` module to emit typed events from keeper [#267](https://github.com/provenance-io/provenance/issues/267)

## [v1.4.1](https://github.com/provenance-io/provenance/releases/tag/v1.4.1) - 2021-06-02

* Updated github binary release workflow.  No code changes from 1.4.0.

## [v1.4.0](https://github.com/provenance-io/provenance/releases/tag/v1.4.0) - 2021-06-02

### Features

* ENV config support, SDK v0.42.5 update [#320](https://github.com/provenance-io/provenance/issues/320)
* Upgrade handler set version name to `citrine` [#339](https://github.com/provenance-io/provenance/issues/339)

### Bug Fixes

* P8EMemorializeContract: preserve some Scope fields if the scope already exists [PR 336](https://github.com/provenance-io/provenance/pull/336)
* Set default standard err/out for `provenanced` commands [PR 337](https://github.com/provenance-io/provenance/pull/337)
* Fix for invalid help text permissions list on marker access grant command [PR 337](https://github.com/provenance-io/provenance/pull/337)
* When writing a session, make sure the scope spec of the containing scope, contains the session's contract spec. [#322](https://github.com/provenance-io/provenance/issues/322)

###  Improvements

* Informative error message for `min-gas-prices` invalid config panic on startup [#333](https://github.com/provenance-io/provenance/issues/333)
* Update marker event documentation to match typed event namespaces [#304](https://github.com/provenance-io/provenance/issues/304)


## [v1.3.1](https://github.com/provenance-io/provenance/releases/tag/v1.3.1) - 2021-05-21

### Bug Fixes

* Remove broken gauge on attribute module. Fixes prometheus metrics [#315](https://github.com/provenance-io/provenance/issues/315)
* Correct logging levels for marker mint/burn requests [#318](https://github.com/provenance-io/provenance/issues/318)
* Fix the CLI metaaddress commands [#321](https://github.com/provenance-io/provenance/issues/321)

### Improvements

* Add Kotlin and Javascript examples for Metadata Addresses [#301](https://github.com/provenance-io/provenance/issues/301)
* Updated swagger docs [PR 313](https://github.com/provenance-io/provenance/pull/313)
* Fix swagger docs [PR 317](https://github.com/provenance-io/provenance/pull/317)
* Updated default min-gas-prices to reflect provenance network nhash economics [#310](https://github.com/provenance-io/provenance/pull/310)
* Improved marker error message when marker is not found [#325](https://github.com/provenance-io/provenance/issues/325)


## [v1.3.0](https://github.com/provenance-io/provenance/releases/tag/v1.3.0) - 2021-05-06

### Features

* Add grpc messages and cli command to add/remove addresses from metadata scope data access [#220](https://github.com/provenance-io/provenance/issues/220)
* Add a `context` field to the `Session` [#276](https://github.com/provenance-io/provenance/issues/276)
* Add typed events and telemetry metrics to attribute module [#86](https://github.com/provenance-io/provenance/issues/86)
* Add rpc and cli support for adding/updating/removing owners on a `Scope` [#283](https://github.com/provenance-io/provenance/issues/283)
* Add transaction and query time measurements to marker module [#284](https://github.com/provenance-io/provenance/issues/284)
* Upgrade handler included that sets denom metadata for `hash` bond denom [#294](https://github.com/provenance-io/provenance/issues/294)
* Upgrade wasmd to v0.16.0 [#291](https://github.com/provenance-io/provenance/issues/291)
* Add params query endpoint to the marker module cli [#271](https://github.com/provenance-io/provenance/issues/271)

### Improvements

* Added linkify script for changelog issue links [#107](https://github.com/provenance-io/provenance/issues/107)
* Changed Metadata events to be typed events [#88](https://github.com/provenance-io/provenance/issues/88)
* Updated marker module spec documentation [#93](https://github.com/provenance-io/provenance/issues/93)
* Gas consumption telemetry and tracing [#299](https://github.com/provenance-io/provenance/issues/299)

### Bug Fixes

* More mapping fixes related to `WriteP8EContractSpec` and `P8EMemorializeContract` [#275](https://github.com/provenance-io/provenance/issues/275)
* Fix event manager scope in attribute, name, marker, and metadata modules to prevent event duplication [#289](https://github.com/provenance-io/provenance/issues/289)
* Proposed markers that are cancelled can be deleted without ADMIN role being assigned [#280](https://github.com/provenance-io/provenance/issues/280)
* Fix to ensure markers have no balances in Escrow prior to being deleted. [#303](https://github.com/provenance-io/provenance/issues/303)

### State Machine Breaking

* Add support for purging destroyed markers [#282](https://github.com/provenance-io/provenance/issues/282)

## [v1.2.0](https://github.com/provenance-io/provenance/releases/tag/v1.2.0) - 2021-04-26

### Improvements

* Add spec documentation for the metadata module [#224](https://github.com/provenance-io/provenance/issues/224)

### Features

* Add typed events and telemetry metrics to marker module [#247](https://github.com/provenance-io/provenance/issues/247)

### Bug Fixes

* Wired recovery flag into `init` command [#254](https://github.com/provenance-io/provenance/issues/254)
* Always anchor unrestricted denom validation expressions, Do not allow slashes in marker denom expressions [#258](https://github.com/provenance-io/provenance/issues/258)
* Mapping and validation fixes found while trying to use `P8EMemorializeContract` [#256](https://github.com/provenance-io/provenance/issues/256)

### Client Breaking

* Update marker transfer request signing behavior [#246](https://github.com/provenance-io/provenance/issues/246)


## [v1.1.1](https://github.com/provenance-io/provenance/releases/tag/v1.1.1) - 2021-04-15

### Bug Fixes

* Add upgrade plan v1.1.1

## [v1.1.0](https://github.com/provenance-io/provenance/releases/tag/v1.1.0) - 2021-04-15

### Features

* Add marker cli has two new flags to set SupplyFixed and AllowGovernanceControl [#241](https://github.com/provenance-io/provenance/issues/241)
* Modify 'enable governance' behavior on marker module [#227](https://github.com/provenance-io/provenance/issues/227)
* Typed Events and Metric counters in Name Module [#85](https://github.com/provenance-io/provenance/issues/85)

### Improvements

* Add some extra aliases for the CLI query metadata commands.
* Make p8e contract spec id easier to communicate.

### Bug Fixes

* Add pagination flags to the CLI query metadata commands.
* Fix handling of Metadata Write message id helper fields.
* Fix cli metadata address encoding/decoding command tree [#231](https://github.com/provenance-io/provenance/issues/231)
* Metadata Module parsing of base64 public key fixed [#225](https://github.com/provenance-io/provenance/issues/225)
* Fix some conversion pieces in `P8EMemorializeContract`.
* Remove extra Object Store Locator storage.
* Fix input status mapping.
* Add MsgSetDenomMetadataRequest to the marker handler.

## [v1.0.0](https://github.com/provenance-io/provenance/releases/tag/v1.0.0) - 2021-03-31

### Bug Fixes

* Resolves an issue where Gov Proposals to Create a new name would fail for new root domains [#192](https://github.com/provenance-io/provenance/issues/192)
* Remove deprecated ModuleCdc amino encoding from Metadata Locator records [#187](https://github.com/provenance-io/provenance/issues/187)
* Update Cosmos SDK to 0.42.3
* Remove deprecated ModuleCdc amino encoding from name module [#189](https://github.com/provenance-io/provenance/issues/189)
* Remove deprecated ModuleCdc amino encoding from attribute module [#188](https://github.com/provenance-io/provenance/issues/188)

### Features

* Allow withdrawals of any coin type from a marker account in WASM smart contracts. [#151](https://github.com/provenance-io/provenance/issues/151)
* Added cli tx commands `write-contract-specification` `remove-contract-specification` for updating/adding/removing metadata `ContractSpecification`s. [#195](https://github.com/provenance-io/provenance/issues/195)
* Added cli tx commands `write-record-specification` `remove-record-specification` for updating/adding/removing metadata `RecordSpecification`s. [#176](https://github.com/provenance-io/provenance/issues/176)
* Added cli tx commands `write-scope-specification` `remove-scope-specification` for updating/adding/removing metadata `ScopeSpecification`s. [#202](https://github.com/provenance-io/provenance/issues/202)
* Added cli tx commands `write-scope` `remove-scope` for updating/adding/removing metadata `Scope`s. [#199](https://github.com/provenance-io/provenance/issues/199)
* Added cli tx commands `write-record` `remove-record` for updating/adding/removing metadata `Record`s. [#205](https://github.com/provenance-io/provenance/issues/205)
* Simulation testing support [#95](https://github.com/provenance-io/provenance/issues/95)
* Name module simulation testing [#24](https://github.com/provenance-io/provenance/issues/24)
* Added default IBC parameters for v039 chain genesis migration script [#102](https://github.com/provenance-io/provenance/issues/102)
* Expand and simplify querying. [#169](https://github.com/provenance-io/provenance/issues/169)
  * Added endpoints for getting all entries of a type, e.g. `RecordsAll`.
  * Combined some endpoints (see notesin "API Breaking" section).
  * Allow searching for related entries. E.g. you can provide a record id to the scope search.
  * Add ability to return related entries. E.g. the `Sessions` endpoint has a `include_records` flag that will cause the response to contain the records that are part of the sessions.
* Add optional identification fields in tx `Write...` messages. [#169](https://github.com/provenance-io/provenance/issues/169)
* The `Write` endpoints now return information about the written entries. [#169](https://github.com/provenance-io/provenance/issues/169)
* Added a CLI command for getting all entries of a type, `query metadata all <type>`, or `query metadata <type> all`. [#169](https://github.com/provenance-io/provenance/issues/169)
* Restrict denom metadata. [#208](https://github.com/provenance-io/provenance/issues/208)

### API Breaking

* Change `Add...` metadata tx endpoints to `Write...` (e.g. `AddScope` is now `WriteScope`). [#169](https://github.com/provenance-io/provenance/issues/169)
* Expand and simplify metadata querying. [#169](https://github.com/provenance-io/provenance/issues/169)
  * Removed the `SessionContextByID` and `SessionContextByUUID` endponts. Replaced with the `Sessions` endpoint.
  * Removed the `RecordsByScopeID` and `RecordsByScopeUUID` endpoints. Replaced with the `Records` endpoint.
  * Removed the `ContractSpecificationExtended` endpoint. Use `ContractSpecification` now with the `include_record_specs` flag.
  * Removed the `RecordSpecificationByID` endpoint. Use the `RecordSpecification` endpoint.
  * Change the `_uuid` fields in the queries to `_id` to allow for either address or uuid input.
  * The `Scope` query no longer returns `Sessions` and `Records` by default. Use the `include_sessions` and `include_records` if you want them.
  * Query result entries are now wrapped to include extra id information alongside an entry.
    E.g. Where a `Scope` used to be returned, now a `ScopeWrapper` is returned containing a `Scope` and its `ScopeIdInfo`.
    So where you previously had `resp.Scope` you will now want `resp.Scope.Scope`.
  * Pluralized both the message name and field name of locator queries that return multiple entries.
    * `OSLocatorByScopeUUIDRequest` and `OSLocatorByScopeUUIDResponse` changed to `OSLocatorsByScopeUUIDRequest` and `OSLocatorsByScopeUUIDResponse`.
    * `OSLocatorByURIRequest` and `OSLocatorByURIResponse` changed to `OSLocatorsByURIRequest` and `OSLocatorsByURIResponse`.
    * Field name `locator` changed to `locators` in `OSLocatorsByURIResponse`, `OSLocatorsByScopeUUIDResponse`, `OSAllLocatorsResponse`.

### Client Breaking

* The paths for querying metadata have changed. See API Breaking section for an overview, and the proto file for details. [#169](https://github.com/provenance-io/provenance/issues/169)
* The CLI has been updated for metadata querying. [#169](https://github.com/provenance-io/provenance/issues/169)
  * Removed the `fullscope` command. Use `query metadata scope --include-sessions --include-records` now.
  * Combined the `locator-by-addr`, `locator-by-uri`, `locator-by-scope`, and `locator-all` into a single `locator` command.
* Changed the CLI metadata tx `add-...` commands to `write-...`. [#166](https://github.com/provenance-io/provenance/issues/166)

## [v0.3.0](https://github.com/provenance-io/provenance/releases/tag/v0.3.0) - 2021-03-19

### Features

* Governance proposal support for marker module
* Decentralized discovery for object store instances [#105](https://github.com/provenance-io/provenance/issues/105)
* Add `AddP8eContractSpec` endpoint to convert v39 contract spec into v40 contract specification  [#167](https://github.com/provenance-io/provenance/issues/167)
* Refactor `Attribute` validate to sdk standard validate basic and validate size of attribute value [#175](https://github.com/provenance-io/provenance/issues/175)
* Add the temporary `P8eMemorializeContract` endpoint to help facilitate the transition. [#164](https://github.com/provenance-io/provenance/issues/164)
* Add handler for 0.3.0 testnet upgrade.

### Bug Fixes

* Gov module route added for name module root name proposal
* Update Cosmos SDK to 0.42.2 for bug fixes and improvements


## [v0.2.1](https://github.com/provenance-io/provenance/releases/tag/v0.2.1) - 2021-03-11

* Update to Cosmos SDK 0.42.1
* Add github action for docker publishing [#156](https://github.com/provenance-io/provenance/issues/156)
* Add `MetaAddress` encoder and parser commands [#147](https://github.com/provenance-io/provenance/issues/147)
* Add build support for publishing protos used in this release [#69](https://github.com/provenance-io/provenance/issues/69)
* Support for setting a marker denom validation expression [#84](https://github.com/provenance-io/provenance/issues/84)
* Expand cli metadata query functionality [#142](https://github.com/provenance-io/provenance/issues/142)

## [v0.2.0](https://github.com/provenance-io/provenance/releases/tag/v0.2.0) - 2021-03-05

* Truncate hashes used in metadata addresses for Record, Record Specification [#132](https://github.com/provenance-io/provenance/issues/132)
* Add support for creating, updating, removing, finding, and iterating over `Session`s [#55](https://github.com/provenance-io/provenance/issues/55)
* Add support for creating, updating, removing, finding, and iterating over `RecordSpecification`s [#59](https://github.com/provenance-io/provenance/issues/59)

## [v0.1.10](https://github.com/provenance-io/provenance/releases/tag/v0.1.10) - 2021-03-04

### Bug fixes

* Ensure all upgrade handlers apply always before storeLoader is created.
* Add upgrade handler for v0.1.10

## [v0.1.9](https://github.com/provenance-io/provenance/releases/tag/v0.1.9) - 2021-03-03

### Bug fixes

* Add module for metadata for v0.1.9

## [v0.1.8](https://github.com/provenance-io/provenance/releases/tag/v0.1.8) - 2021-03-03

### Bug fixes

* Add handlers for v0.1.7, v0.1.8

## [v0.1.7](https://github.com/provenance-io/provenance/releases/tag/v0.1.7) - 2021-03-03

### Bug Fixes

* Fix npe caused by always loading custom storeLoader.

## [v0.1.6](https://github.com/provenance-io/provenance/releases/tag/v0.1.6) - 2021-03-02

### Bug Fixes

* Add metadata module to the IAVL store during upgrade

## [v0.1.5](https://github.com/provenance-io/provenance/releases/tag/v0.1.5) - 2021-03-02

* Add support for creating, updating, removing, finding, and iterating over `Record`s [#54](https://github.com/provenance-io/provenance/issues/54)
* Add migration support for v039 account into v040 attributes module [#100](https://github.com/provenance-io/provenance/issues/100)
* Remove setting default no-op upgrade handlers.
* Add an explicit no-op upgrade handler for release v0.1.5.
* Add support for creating, updating, removing, finding, and iterating over `ContractSpecification`s [#57](https://github.com/provenance-io/provenance/issues/57)
* Add support for record specification metadata addresses [#58](https://github.com/provenance-io/provenance/issues/58)
* Enhance build process to release cosmovisor compatible zip and plan [#119](https://github.com/provenance-io/provenance/issues/119)

## [v0.1.4](https://github.com/provenance-io/provenance/releases/tag/v0.1.4) - 2021-02-24

* Update `ScopeSpecification` proto and create `Description` proto [#71](https://github.com/provenance-io/provenance/issues/71)
* Update `Scope` proto: change field `owner_address` to `owners` [#89](https://github.com/provenance-io/provenance/issues/89)
* Add support for migrating Marker Accesslist from v39 to v40 [#46](https://github.com/provenance-io/provenance/issues/46).
* Add migration command for previous version of Provenance blockchain [#78](https://github.com/provenance-io/provenance/issues/78)
* Add support for creating, updating, removing, finding, and iterating over `ScopeSpecification`s [#56](https://github.com/provenance-io/provenance/issues/56)
* Implemented v39 to v40 migration for name module.
* Add support for github actions to build binary releases on tag [#30](https://github.com/provenance-io/provenance/issues/30).

## [v0.1.3](https://github.com/provenance-io/provenance/releases/tag/v0.1.3) - 2021-02-12

* Add support for Scope objects to Metadata module [#53](https://github.com/provenance-io/provenance/issues/53)
* Denom Metadata config for nhash in testnet [#42](https://github.com/provenance-io/provenance/issues/42)
* Denom Metadata support for marker module [#47](https://github.com/provenance-io/provenance/issues/47)
* WASM support for Marker module [#28](https://github.com/provenance-io/provenance/issues/28)

### Bug Fixes

* Name service allows uuids as segments despite length restrictions [#48](https://github.com/provenance-io/provenance/issues/48)
* Protogen breaks on marker uint64 equals [#38](https://github.com/provenance-io/provenance/issues/38)
* Fix for marker module beginblock wiring [#34](https://github.com/provenance-io/provenance/issues/34)
* Fix for marker get cli command
* Updated the links in PULL_REQUEST_TEMPLATE.md to use correct 'main' branch

## [v0.1.2](https://github.com/provenance-io/provenance/releases/tag/v0.1.2) - 2021-01-27

### Bug Fixes

* Update goreleaser configuration to match `provenance` repository name

## [v0.1.1](https://github.com/provenance-io/provenance/releases/tag/v0.1.1) - 2021-01-27

This is the intial beta release for the first Provenance public TESTNET.  This release is not intended for any type of
production or reliable development as extensive work is still in progress to migrate the private network functionality
into the public network.

### Features

* Initial port of private Provenance blockchain modules `name`, `attribute`, and `marker` from v0.39.x Cosmos SDK chain
into new 0.40.x base.  Minimal unit test coverage and features in place to begin setup of testnet process.

## PRE-HISTORY

## [v0.1.0](https://github.com/provenance-io/provenance/releases/tag/v0.1.0) - 2021-01-26

* Test tag prior to initial testnet release.

The Provenance Blockchain was started by Figure Technologies in 2018 using a Hyperledger Fabric derived private network.
A subsequent migration was made to a new internal private network based on the 0.38-0.39 series of Cosmos SDK and
Tendermint.  The Provence-IO/Provenance Cosmos SDK derived public network is the
