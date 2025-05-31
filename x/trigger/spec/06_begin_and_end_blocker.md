<!--
order: 6
-->

<!-- TOC 1 2 -->
  - [Begin Blocker](#begin-blocker)
    - [Trigger Execution](#trigger-execution)
  - [End Blocker](#end-blocker)
    - [Block Event Detection](#block-event-detection)

# Begin Blocker

The `BeginBlocker` abci call is invoked on the beginning of each block. `Triggers` will be dequeued and ran.

## Trigger Execution

The following steps are performed on each `BeginBlocker`:
1. A `Trigger` is removed from the `Queue`.
2. An `Action` is run.
3. The events for the `Action` are emitted.
4. Step 2-3 are repeated until no more `Actions` exist for the trigger.
5. Step 1-4 is repeated until the `Queue` is empty or the `throttling limit` has been reached.

### Note

We have implemented a `throttling limit` within the module's `BeginBlocker`, effectively enforcing a maximum of 5 actions per `BeginBlock`.

# End Blocker

The `EndBlocker` abci call is ran at the end of each block. The `EventManager`, `BlockHeight`, and `BlockTime` are monitored and used to detect `Trigger` activation.

## Block Event Detection

The following is logic is used to detect the activation of a `Trigger`:
1. The `EventManager` is utilized to traverse the transaction events from the newly created block.
2. The `Event Listener` table filters for `Triggers` containing a `TransactionEvent` matching the transaction event types and containing the defined `Attributes`.
3. The `Event Listener` table filters for `Triggers` containing a `BlockHeightEvent` that is greater than or equal to the current `BlockHeight`.
4. The `Event Listener` table filters for `Triggers` containing a `BlockTimeEvent` that is greater than or equal to the current `BlockTime`.
5. These `Triggers` are then unregistered and added to the `Queue`.
