<!--
order: 6
-->

# Events

The rewards module emits the following events:

<!-- TOC -->
  - [Reward Program Created](#reward-program-created)
  - [Reward Program Started](#reward-program-started)
  - [Reward Program Finished](#reward-program-finished)
  - [Reward Program Expired](#reward-program-expired)
  - [Reward Program Ended](#reward-program-ended)
  - [Claim Rewards](#claim-rewards)
  - [Claim All Rewards](#claim-all-rewards)


---
## Reward Program Created

Fires when a reward program is created with the Create Reward Program Msg.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| RewardProgramCreated   | reward_program_created| \{ID string\}               |

---
## Reward Program Started

Fires when a reward program transitions to the STARTED state.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| RewardProgramStarted   | reward_program_id     | \{ID string\}               |

---
## Reward Program Finished

Fires when a reward program transitions to the FINISHED state.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| RewardProgramFinished  | reward_program_id     | \{ID string\}               |

---
## Reward Program Expired

Fires when a reward program transitions to the EXPIRED state.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| RewardProgramExpired   | reward_program_id     | \{ID string\}               |

---
## Reward Program Ended

Fires when a reward program is ended with the End Reward Program Msg.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| RewardProgramEnded     | reward_program_id     | \{ID string\}               |

---
## Claim Rewards

Fires when a participant claims a reward claim with Claim Reward Msg.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| ClaimRewards           | reward_program_id     | \{ID string\}               |
| ClaimRewards           | rewards_claim_address | \{bech32address string\}    |

This event will not fire if the user has no claims or if they have already claimed all their rewards.

---
## Claim All Rewards

Fires when a participant claims all their reward claims with Claim Reward Msg.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| ClaimAllRewards        | rewards_claim_address | \{bech32address string\}    |

This event will not fire if the user has no claims or if they have already claimed all their rewards.

---
