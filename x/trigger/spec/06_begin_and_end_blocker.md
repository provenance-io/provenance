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
2. A `Trigger` is removed from the `Queue`.
3. The `Gas Limit` for the `Trigger` is retrieved from the store.
4. A `GasMeter` is created for the `Trigger`.
5. An `Action` on the `Trigger` is ran updating and verifying gas usage against the `GasMeter`
6. The events for the `Action` are emitted.
7. Step 5 is repeated until no more `Actions` exist for the trigger.
8. Step 1 is repeated until the `Queue` is empty or the `throttling limit` has been reached.

### Note

We have implemented a `throttling limit` within the module's `BeginBlocker`, effectively enforcing a maximum of 5 actions and a gas limit of 2,000,000 per `BeginBlock`.

# End Blocker

The `EndBlocker` abci call is ran at the end of each block. The `EventManager`, `BlockHeight`, and `BlockTime` are monitored and used to detect `Trigger` activation.

## Block Event Detection

The following is logic is used to detect the activation of a `Trigger`:
1. The `EventManager` is utilized to traverse the transaction events from the newly created block.
2. The `Event Listener` table filters for `Triggers` containing a `TransactionEvent` matching the transaction event types and containing the defined `Attributes`.
3. The `Event Listener` table filters for `Triggers` containing a `BlockHeightEvent` that is greater than or equal to the current `BlockHeight`.
4. The `Event Listener` table filters for `Triggers` containing a `BlockTimeEvent` that is greater than or equal to the current `BlockTime`.
5. These `Triggers` are then unregistered and added to the `Queue`.
