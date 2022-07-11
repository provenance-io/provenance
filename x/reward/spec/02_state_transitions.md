# State Transitions

This document describes the state transitions that occur within the Active Participation and Engagement (APE) module.

<!-- TOC 2 2 -->
 
## State transitions for a Reward Program

![RewardProgram state diagram](./diagrams/reward-program/RewardProgram.png?raw=true "RewardProgram state diagram")

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

## State transitions for a Reward Claim

![RewardClaim state diagram](./diagrams/reward-claim/RewardClaim.png#center?raw=true "RewardClaim state diagram")

### Unclaimable
The reward has been granted to a participant, but it cannot be claimed until the current claim period ends.

### Claimable
The reward has been granted to the participant, and it's claimable by the participant via a transaction. If the reward is not claimed it will eventually expire.

### Claimed
The reward has been granted and received by the participant. A reward cannot be claimed more than once.

### Expired
The reward has been cleaned up and the participant can no longer claim it. The funds attached to the reward claim are refunded back to the program creator.