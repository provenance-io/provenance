<!--
order: 6
-->

# Begin Blocker
The `BeginBlocker` abci call is invoked on the beginning of each block. `Triggers` will be dequeued and ran.

## Trigger Execution
The following steps are performed on each `BeginBlocker`:
1. A `Trigger` is removed from the `Queue`.
2. The `Gas Limit` for the `Trigger` is retrieved from the store.
3. A `GasMeter` is created for the `Trigger`.
4. An `Action` on the `Trigger` is ran updating and verifying gas usage against the `GasMeter`
5. The events for the `Action` are emitted.
6. Step 4 is repeated until no more `Actions` exist for the trigger.
7. Step 1 is repeated until the `Queue` is empty or the `throttling` limit has been reached.

# End Blocker
The `EndBlocker` abci call is ran at the end of each block. The `EventManager`, `BlockHeight`, and `BlockTime` are monitored and used to detect `Trigger` activation.

## Block Event Detection
The following is logic is used to detect the activation of a `Trigger`:
1. The `EventManager` is utilized to traverse the transaction events from the newly created block.
2. The `Event Listener` table filters for `Triggers` containing a `TransactionEvent` matching the transaction event types and containing the defined `Attributes`.
3. The `Event Listener` table filters for `Triggers` containing a `BlockHeightEvent` that is greater than or equal to the current `BlockHeight`.
4. The `Event Listener` table filters for `Triggers` containing a `BlockTimeEvent` that is greater than or equal to the current `BlockTime`.
5. These `Triggers` are then unregistered and added to the `Queue`.
