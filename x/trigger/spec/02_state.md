<!--
order: 2
-->

# State

The trigger module manages the state of every trigger.

---
<!-- TOC 2 4 -->
  - [Trigger](#trigger)
    - [TriggerEventI](#triggereventi)
      - [BlockHeightEvent](#blockheightevent)
      - [BlockTimeEvent](#blocktimeevent)
      - [TransactionEvent](#transactionevent)
  - [Queue](#queue)



## Trigger

A `Trigger` is the main data structure used by the module. It keeps track of the owner, event, and actions for a single `Trigger`. Every `Trigger` gets its own unique identifier, and a unique entry within the `Event Listener` and `Gas Limit` tables. The `Event Listener` table allows the event detection system to quickly filter applicable `Triggers` by name and type. A trigger can vary in size making it difficult to calculate gas usage on store, thus we opted to store remaining transaction gas in the `Gas Limit` table. It gives us a predictable way to calculate and store remaining gas.

The excess gas on a MsgCreateTrigger transaction will be used for the `Trigger's` `Gas Limit` table. The maximum `Gas Limit` for a `Trigger` is `2000000`.

* Trigger: `0x01 | Trigger ID (8 bytes) -> ProtocolBuffers(Trigger)`
* Trigger ID: `0x05 -> uint64(TriggerID)`
* Event Listener: `0x02 | Event Type (32 bytes) | Order (8 bytes) -> []byte{}`
* Gas Limit: `0x04 | Trigger ID (8 bytes) -> uint64(GasLimit)`

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/trigger.proto#L13-L25

### TriggerEventI

A `Trigger` must have an event that implements the `TriggerEventI` interface. Currently, the system supports `BlockHeightEvent`, `BlockTimeEvent`, and `TransactionEvent`.

#### BlockHeightEvent

The `BlockHeightEvent` allows the user to configure their `Trigger` to fire when the current block's `Block Height` is greater than or equal to the defined one.

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/trigger.proto#L39-L46

#### BlockTimeEvent

The `BlockTimeEvent` allows the user to configure their `Trigger` to fire when the current block's `Block Time` is greater than or equal to the defined one.

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/trigger.proto#L48-L55

#### TransactionEvent

The `TransactionEvent` allows the user to configure their `Trigger` to fire when a transaction event matching the user defined one has been emitted.

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/trigger.proto#L57-L66

##### Attribute

The `Attribute` is used by the `TransactionEvent` to allow the user to configure which attributes must be present on the transaction event. An `Attribute` with an empty `value` will only require the `name` to match.

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/trigger.proto#L68-L76

---
## Queue
<!-- link message: QueuedTrigger -->

The `Queue` is an internal structure that we use to store and throttle the execution of `Triggers` on the `BeginBlock`. We store each `Trigger` as a `QueuedTrigger`, and then manipulate the `Queue Start Index` and `Queue Length` whenever we add or remove from the `Queue`. When we add to the `Queue`, the new element is added to the `QueueStartIndex` + `Length`. The `QueueLength` is then incremented by one. When we dequeue from the Queue, the `QueueStartIndex` will be incremented by 1 and the `QueueLength` is decremented by 1. We also ensure the key of the dequeued element is removed.

* Queue Item: `0x03 | Queue Index (8 bytes) -> ProtocolBuffers(QueuedTrigger)`
* Queue Start Index: `0x06 -> uint64(QueueStartIndex)`
* Queue Length: `0x07 -> uint64(QueueLength)`

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/trigger.proto#L27-L37
