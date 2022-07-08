# State Transitions

This document describes the state transition operations pertaining markers:

<!-- TOC 2 2 -->
 
## State transitions for Reward Program

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
