<!--
order: 2
-->

# State

The rewards module manages the state of every reward program and each of its participants.

<!-- TOC -->
  - [Reward Program](#reward-program)
  - [Claim Period Reward Distribution](#claim-period-reward-distribution)
  - [Reward Account State](#reward-account-state)
    - [Action Counter](#action-counter)
  - [Qualifying Actions](#qualifying-actions)
    - [Action Delegate](#action-delegate)
    - [Action Transfer](#action-transfer)
    - [Action Vote](#action-vote)

---
## Reward Program

A `RewardProgram` is the main data structure used by the Active Participation and Engagement (APE) module. It keeps track of the state, balances, qualifying actions, timers, and counters for a single Reward Program. Every Reward Program gets its own unique identifier that we track within the store.

* Reward Program: `0x01 | RewardProgram ID (8 bytes) -> ProtocolBuffers(RewardProgram)`
* Reward Program ID: `0x02 -> uint64(RewardProgramID)`

+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/reward.proto#L12-L73

---
## Claim Period Reward Distribution

A `ClaimPeriodRewardDistribution` is created for each claim period of every `RewardProgram`. Its purpose is to track live claim period specific information. Examples of this include the total number of granted shares in the claim period, sum of of all its rewards given out as claims, and the amount of reward allocated to it from the `RewardProgram`.

* Claim Period Reward Distribution: `0x03 | Reward Program ID (8 bytes) | Claim Period ID (8 bytes) -> ProtocolBuffers(ClaimPeriodRewardDistribution)`

+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/reward.proto#L75-L92

---
## Reward Account State

The purpose of `RewardAccountState` is to track state at the address level of a claim period. It counts the number of claim period shares the user obtained, the status of their `RewardClaim`, and other stateful information that assists the system in properly granting rewards.

* AccountStateAddressLookupKeyPrefix: `0x04 | Account Address (n bytes, with the address length being stored in the first byte {int64(address[1:2][0])}) | Reward Program ID (8 bytes) | Claim Period ID (8 bytes) -> ProtocolBuffers(RewardAccountState)`
* AccountStateKeyPrefix: `0x05 | Reward Program ID (8 bytes) | Claim Period ID (8 bytes) | Account Address (n bytes, with the address length being stored in the first byte {int64(address[1:2][0])}) -> ProtocolBuffers(RewardAccountState)`

+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/reward.proto#L94-L123

### Action Counter

`ActionCounter` tracks the number of times an action has been performed.

+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/reward.proto#L190-L199

---
## Qualifying Actions

A list of one or more actions that a user can perform to attempt to participate in a `RewardProgram`. In order to be considered a participant and granted a share then all the `EligiblityCriteria` on the action must be met. Each action has its own `EligiblityCriteria`, which is independently evaluated against system state and `RewardAccountState` for that user. Each `Qualifying Action` is evaluated independently, thus it is possible for a user to earn more than one reward for a single action.

+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/reward.proto#L125-L141

### Action Delegate

`ActionDelegate` is when a user performs a delegate.

+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/reward.proto#L143-L162

The triggering account must have a total delegation amount within the bands of [`minimum_delegation_amount`,`maximum_delegation_amount`]. Additionally, the validator they are staking to must be within the [`minimum_active_stake_percentile`,`maximum_active_stake_percentile`] power percentile. If both of these criteria are met then the delegate is considered successful. The `minimum_actions` and `maximum_actions` fields are the number of successful delegate that must be performed. Once all these conditions are met then the user will receive a share.

### Action Transfer

`ActionTransfer` is when a user transfers coins.

+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/reward.proto#L164-L175

If the triggering account has delegated at least the `minimum_delegation_amount`, then the transfer action will be considered successful. The `minimum_actions` and `maximum_actions` fields are the number of successful transfer that must be performed. When all these conditions are met, then the user will receive a share.

### Action Vote

`ActionVote` is when a user votes on a proposal.

+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/reward.proto#L177-L188

If the triggering account has delegated at least the `minimum_delegation_amount`, then the vote action will be considered successful. The `minimum_actions` and `maximum_actions` fields are the number of successful votes that must be performed. When all these conditions are met, then the user will receive a share.
