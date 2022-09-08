<!--
order: 1
-->

# Concepts

<!-- TOC -->
  - [Reward Program](#reward-program)
  - [Qualifying Actions and Eligibility Criteria](#qualifying-actions-and-eligibility-criteria)
  - [Claim Period](#claim-period)
  - [Reward Claim](#reward-claim)
  - [Rollover](#rollover)
  - [Refunding](#refunding)

## Reward Program
Reward Programs are configurable campaigns that encourage users to participate in the Provenance Blockchain. Entities interested in creating a Reward Program will supply their new program with funds, set the duration of their program, and provide the participation requirements.

## Qualifying Actions and Eligibility Criteria
A `Qualifying Action` is one or more transactions that a user performs on the Provenance Blockchain that has been listed within the `Reward Program`. These actions are then evaluated against a set of criteria that are also defined within the `Reward Program` known as `Eligiblity Criteria`. Users become participants in the Reward Program by performing a `Qualifying Action` and meeting all conditions specified by its `Eligiblity Criteria`.

## Claim Period
A `Reward Program` is split into one or more time intervals known as `Claim Periods`. Each of these `Claim Periods` gets an equal portion of the `Reward Program Reward Pool` known as the `Claim Period Reward Pool`. Users can participate within these `Claim Periods` and are rewarded for their actions.

## Reward Claim
When a user participates in a `Reward Program` they are granted one or more shares of the `Claim Period Reward Pool`. Once the`Claim Period` ends, the participant will be able to claim their reward by performing a claim transaction. The participant's reward is proportional to their activity compared to everyone else within a `Claim Period`. Users must claim their rewards before the `Reward Program` expires. Additionally, users will be limited to `max_reward_per_address` of the `Claim Period Reward Pool`.

**Reward For Claim Period**

$$\left( ClaimPeriodRewardPool \right) \times \left( EarnedShares \over ClaimPeriodShares \right) $$

## Rollover
It is possible that not all of the `Claim Period Reward Pool` will be distributed. This can happen when there is not enough activity, and participants are gated by the `max_reward_per_address`. The `Reward Program` will attempt to move these funds into a `Rollover Claim Period`. A `Rollover Claim Period` behaves exactly like any other `Claim Period`, but it is not guaranteed to have an equal portion of the original `Reward Program Reward Pool`. A `Reward Program` may run up to `max_rollover_claim_periods`, but is not guaranteed to run any of them. This is dependent on user activity, `program_end_time_max` field, and the `minimum_rollover_amount` field. Currently, the `minimum_rollover_amount` is set to 10% of the `Claim Period Reward Pool`.

## Refunding
When a `Reward Program` ends it gives all participants `expiration_offset` seconds to claim their rewards. After `expiration_offset` seconds the `Reward Program` expires and prevents participants from claiming. The unclaimed rewards and any funds still remaining within the `Reward Program Reward Pool` will be given back to the creator.
