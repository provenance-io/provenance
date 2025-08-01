<!--
order: 1
-->

# Concepts

The trigger module allows users to delay the execution of a message until an event is detected. Users should have a strong understanding of what a `Trigger`, `Event`, `Queued Trigger` are, and how `Payment` works before using this module.

<!-- TOC -->
  - [Trigger](#trigger)
  - [Actions](#actions)
  - [Payment](#payment)
  - [Block Event](#block-event)
    - [Transaction Event](#transaction-event)
    - [Block Height Events](#block-height-events)
    - [Block Time Event](#block-time-event)
  - [Queued Trigger](#queued-trigger)



## Trigger

A `Trigger` is an address owned object that registers to a `Block Event`, and then proceeds to fire off its `Actions` when that `Block Event` has been detected by the system. A `Trigger` is single-shot, and it will automatically be destroyed after its `Block Event` has been detected.

## Actions

`Actions` are one or more `Msg`s that should be invoked. Every `Action` follows the same rules as a `Msg`. See the [Payment](#payment) section for more information.

## Payment

The fee required to create a trigger is the sum of the `Msg` fee for the `MsgCreateTriggerRequest` plus the `Msg` fee for each of the `Actions`. This total is the amount of fees to include in the `Tx`.

## Block Event

A `Block Event` is a blanket term that refers to events that occur during the creation of a block. The `Trigger` module currently supports `Transaction Events`, `Block Height Events`, and `Block Time Events`. 

### Transaction Event

These type of events refer to the `ABCI Events` that are emitted by the `DeliverTx` transactions. An `ABCI Event` must have the same `Type` and `Attributes` as the user defined `Transaction Event` for the event criteria to be met. A user defined `Attribute` with an empty `Value` will always match as long as the `Attribute Name` field matches.

### Block Height Events

These type of events refer to the `Block Height` on a newly created block. The `Block Height` must be greater than or equal to the defined value for the event criteria to be met.

### Block Time Event

These type of events refer to the `Block Time` on a newly created block. The `Block Time` must be greater than or equal to the defined value for the event criteria to be met.

## Queued Trigger

The `Queued Trigger` is a `Trigger` that is ready to have its actions be executed at a future block.
