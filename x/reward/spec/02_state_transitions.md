# State Transitions

This document describes the state transition operations pertaining to:

1. [Reward Programs](./02_state_transitions.md#reward-programs)
2. [Reward Claims](./02_state_transitions.md#reward-claims)

<!-- TOC 2 2 -->
 
## Reward Programs
State transition for Reward Programs happen on `BeginBlock` and make use of the `BlockTime` attribute.

A Reward Program can be `Pending`, `Started`, `Finished`, or `Expired`. A Reward Program will move through all these states, and will initially be in the `Pending` state.

### Pending 
Reward program has *not* started.

### Started 
Reward program has started, rewards can be claimed if the minimum claim period has ended and addresses 
have preformed eligible qualifying events and accrued shares for that claim period.

### Finished 
Reward program had ended, however rewards can still be claimed by eligible addresses.

### Expired
Reward program has passed it's expiry date.
All funds will be returned to the reward creator address

<p align="center">
  <img src="./diagrams/reward-program/RewardProgram.png">
</p>

## Reward Claims
State transitions for a Reward Claim happen on `BeginBlock` and on claim transactions.

A Reward Claim can be `Unclaimable`, `Claimable`, `Claimed`, or `Expired`. A Reward Claim will always start as `Unclaimable` and eventually become `Claimable`. If a participant claims their reward then the Reward Claim will become `Claimed`, otherwise it will timeout and enter the `Expired` state where they can no longer claim it.

### Unclaimable
The reward has been granted to a participant, but it cannot be claimed until the current claim period ends.

### Claimable
The reward has been granted to the participant, and it's claimable by the participant via a transaction. If the reward is not claimed it will eventually expire.

### Claimed
The reward has been granted and received by the participant. A reward cannot be claimed more than once.

### Expired
The reward has been cleaned up and the participant can no longer claim it. The funds attached to the reward claim are refunded back to the program creator.

<p align="center">
  <img src="./diagrams/reward-claim/RewardClaim.png">
</p>
