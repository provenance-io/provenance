## [v1.14.0-rc2](https://github.com/provenance-io/provenance/releases/tag/v1.14.0-rc2) - 2023-02-06

### Improvements

* Bump tendermint to Notional's v0.34.25 (from base repo's v0.34.24) [PR 1348](https://github.com/provenance-io/provenance/pull/1348).
* Bump Cosmos-SDK to v0.46.8-pio-3 (from [v0.46.7-pio-2](https://github.com/provenance-io/cosmos-sdk/compare/v0.46.7-pio-2...v0.46.8-pio-3)) [PR 1348](https://github.com/provenance-io/provenance/pull/1348).
* Set the immediate sanction/unsanction min deposits to 1,000,000 hash in the `paua` upgrade [PR 1345](https://github.com/provenance-io/provenance/pull/1345).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.14.0-rc1...v1.14.0-rc2
* https://github.com/provenance-io/provenance/compare/v1.13.1...v1.14.0-rc2

---

## [v1.14.0-rc1](https://github.com/provenance-io/provenance/releases/tag/v1.14.0-rc1) - 2023-02-02

### Features

* Added support to set a list of specific recipients allowed for send authorizations in the marker module [#1237](https://github.com/provenance-io/provenance/issues/1237).
* Added a new name governance proposal that allows the fields of a name record to be updated. [PR 1266](https://github.com/provenance-io/provenance/pull/1266).
  The `restrict` flag has been changed to `unrestrict` in the `BindName` request and `CreateRootName` proposal.
* Added msg to add, finalize, and activate a marker in a single request [#770](https://github.com/provenance-io/provenance/issues/770).
* Added the `x/quarantine` module [PR 1317](https://github.com/provenance-io/provenance/pull/1317).
* Added the `x/sanction` module [PR 1317](https://github.com/provenance-io/provenance/pull/1317).
* Staking concentration limit protection (prevents delegations to nodes with high voting power) [#1331](https://github.com/provenance-io/provenance/issues/1331).
* Enable ADR-038 State Listening in Provenance

### Improvements

* Bump Cosmos-SDK to v0.46.8-pio-2 (from [v0.46.7-pio-1](https://github.com/provenance-io/cosmos-sdk/compare/v0.46.7-pio-1...v0.46.8-pio-2)) [#1332](https://github.com/provenance-io/provenance/issues/1332).
  See its [RELEASE_NOTES.md](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.8-pio-2/RELEASE_NOTES.md) for details.
* Added assess msg fees spec documentation [#1172](https://github.com/provenance-io/provenance/issues/1172).
* Build dbmigrate and include it as an artifact with releases [#1264](https://github.com/provenance-io/provenance/issues/1264).
* Removed, MsgFees Module 50/50 Fee Split on MsgAssessCustomMsgFeeRequest [#1263](https://github.com/provenance-io/provenance/issues/1263).
* Add basis points field to MsgAssessCustomMsgFeeRequest for split of fee between Fee Module and Recipient [#1268](https://github.com/provenance-io/provenance/issues/1268).
* Updated ibc-go to v6.1 [#1273](https://github.com/provenance-io/provenance/issues/1273).
* Update adding of marker to do additional checks for ibc denoms [#1289](https://github.com/provenance-io/provenance/issues/1289).
* Add validate basic check to msg router service [#1308](https://github.com/provenance-io/provenance/issues/1308).
* Removed legacy-amino [#1275](https://github.com/provenance-io/provenance/issues/1275).
* Opened port 9091 on ibcnet container ibc1 to allow for reaching GRPC [PR 1314](https://github.com/provenance-io/provenance/pull/1314).
* Increase all validator's max commission to 100% [PR 1333](https://github.com/provenance-io/provenance/pull/1333).
* Increase max gas per block to 120,000,000 (from 60,000,000) [PR 1335](https://github.com/provenance-io/provenance/pull/1335).

### Bug Fixes

* Update Maven publishing email to provenance [#1270](https://github.com/provenance-io/provenance/issues/1270).

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.13.1...v1.14.0-rc1

