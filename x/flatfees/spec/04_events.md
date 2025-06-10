# FlatFees Events

The `x/flatfees` module does not have any typed events and does not emit any events itself.
However, there are a few key events emitted about fees that are worth noting.

---
<!-- TOC 2 2 -->
  - [Generic Fee Event](#generic-fee-event)
  - [Initial Fee Event](#initial-fee-event)
  - [Success Fee Event](#success-fee-event)


## Generic Fee Event

At the start of Msg processing, when the up-front cost is paid, an event with info on the fee is emitted.
This event matches a standard one emitted by the SDK.

Event type: `tx`

| Attribute Key | Attribute Value                                 |
|---------------|-------------------------------------------------|
| fee           | The amount of fee that was paid (coins string). |
| fee_payer     | The account that paid the fee (bech32).         |

This `fee` attribute always has the total amount of fee that was paid, even if the Msg fails.
I.e. if the Msg fails, the `fee` is the same as the `min_fee_charged` is the same as what would be the `basefee` (if that event were emitted).
But if the Msg succeeds, the `fee` is the same as the `total` from the success fee event.
Either way, the `fee` is the amount collected for the Msg.


## Initial Fee Event

At the same time that the generic fee event is emitted, an event with info on the up-front cost (`min_fee_charged`) is emitted.
This event was originally added when we created the (no-longer-used) `x/msgfees` module and was kept for backwards compatibility.

Event type: `tx`

| Attribute Key   | Attribute Value                                                 |
|-----------------|-----------------------------------------------------------------|
| min_fee_charged | The up-front cost paid regardless of Tx success (coins string). |
| fee_payer       | The account that paid the fee (bech32).                         |


## Success Fee Event

After a Msg is processed, iff it is successful, an event with a fee recap is emitted.

Event type: `tx`

| Attribute Key | Attribute Value                                                       |
|---------------|-----------------------------------------------------------------------|
| fee_payer     | The account that paid the fee (bech32).                               |
| basefee       | The up-front cost (coins string).                                     |
| additionalfee | The additional fee paid because the Tx was successful (coins string). |
| fee_overage   | The amount of fee provided above what was required (coins string).    |
| total         | The total amount of fee paid (coins string).                          |

The `additionalfee` and `fee_overage` attributes are each omitted if they are zero.
The other attributes will always be present.

The `total` equals the `fee` attribute of the generic fee event, and is the sum of the other three coins in this event.
The `basefee` equals the `min_fee_charged` attribute in the minimum fee event.
