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
