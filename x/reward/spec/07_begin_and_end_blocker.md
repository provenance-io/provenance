<!--
order: 7
-->

# Begin Blocker
The `BeginBlocker` abci call is invoked on the beginning of each block. The newly created block's `BlockTime` field is used to update the `Reward Programs` and `Reward Claims`.

## State Machine Update
The following conditional logic is evaluated to help a `Reward Program` transition between states:
1. Starts a `Reward Program` if the `BlockTime` >= `program_start_time`.
2. Evaluates if `BlockTime` >= `claim_period_end_time`, and if it evaluates to true the `Reward Program` will *attempt* to progress to the next claim period.
3. The Reward Program will successfully progress to the next claim period, if all of the following criteria are true:
   1. `remaining_pool_balance` >= `minimum_rollover_amount`
   2. `BlockTime` < `program_end_time_max`
4. If either of the previously mentioned criteria is not met, then the `Reward Program` will end.
5. A completed `Reward Program` will then expire after `reward_claim_expiration_offset` seconds from its completion time.
6. All expired `Reward Program` will return any unused funds to the reward creator and expire any unclaimed rewards.

# End Blocker
The `EndBlocker` abci call is ran at the end of each block. The `EventManager` is monitored and `Qualifying Actions` are deduced from newly created events and prior internal state.

## Qualifying Action Detection
The following is logic is used to detect `Qualifying Actions` and grant shares:
1. The `EventManager` is utilized to traverse the events from the newly created block.
2. Using the `QualifyingAction` types defined in each active `RewardProgram`, the module attempts to build actions using the event and any prior events.
3. Each detected action will then be evaluated and only deemed a `QualifyingAction` if it matches the `Evaluation Criteria`.
4. One or more shares will be granted to the participant performing the `QualifyingAction`.
5. Participants can then claim these shares once the claim period that they were earned in completes.
6. Detects qualifying events based on events in EventManager for reward programs currently running.
7. Gives out shares for a running rewards program for that claim period for a given address performing qualifying events.
