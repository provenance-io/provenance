## [v1.14.0](https://github.com/provenance-io/provenance/releases/tag/v1.14.0) - 2023-02-22

### Features

* Added support to set a list of specific recipients allowed for send authorizations in the marker module [#1237](https://github.com/provenance-io/provenance/issues/1237).
* Added a new name governance proposal that allows the fields of a name record to be updated. [PR 1266](https://github.com/provenance-io/provenance/pull/1266).
* Added msg to add, finalize, and activate a marker in a single request [#770](https://github.com/provenance-io/provenance/issues/770).
* Added the `x/quarantine` module [PR 1317](https://github.com/provenance-io/provenance/pull/1317).
* Added the `x/sanction` module [PR 1317](https://github.com/provenance-io/provenance/pull/1317).
* Staking concentration limit protection (prevents delegations to nodes with high voting power) [#1331](https://github.com/provenance-io/provenance/issues/1331).
* Enable ADR-038 State Listening in Provenance [PR 1334](https://github.com/provenance-io/provenance/pull/1334).

### Improvements

* Bump Cosmos-SDK to `v0.46.10-pio-1` (from `v0.46.6-pio-1`).
  [PR 1371](https://github.com/provenance-io/provenance/pull/1371),
  [PR 1348](https://github.com/provenance-io/provenance/pull/1348),
  [PR 1334](https://github.com/provenance-io/provenance/pull/1334),
  [PR 1348](https://github.com/provenance-io/provenance/pull/1348),
  [PR 1317](https://github.com/provenance-io/provenance/pull/1317),
  [PR 1278](https://github.com/provenance-io/provenance/pull/1278).  
  See the following `RELEASE_NOTES.md` for details:
  [v0.46.10-pio-1](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.10-pio-1/RELEASE_NOTES.md),
  [v0.46.8-pio-3](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.8-pio-3/RELEASE_NOTES.md),
  [v0.46.7-pio-1](https://github.com/provenance-io/cosmos-sdk/blob/v0.46.7-pio-1/RELEASE_NOTES.md).  
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

