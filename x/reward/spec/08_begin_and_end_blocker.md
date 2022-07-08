<!--
order: 8
-->

# Begin Blocker
Begin blocker updates unexpired reward programs with the following details based on block time
1. Starts a rewards program if the block time > program_start_time
2. Ends a claim period if block time > claim_period_end_time, and if the reward program needs to be ended it is done after the current claim period is emded.
3. If Rewards are expiring mark program as expired and return any unused funds to reward creator.No addresses can claim for expired rewards program going forward.


# End Blocker
End blocker does
1. Detects qualifying events based on events in EventManager for reward programs currently running.
2. Gives out shares for a running rewards program for that claim period for a given address performing qualifying events.
