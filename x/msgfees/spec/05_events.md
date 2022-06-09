<!--
order: 5
-->

# Events

Existing fee event continue to show total fee charged

## Any Tx

| Type     | Attribute Key | Attribute Value    |
| -------- | ------------- | ------------------ |
| message  | fee           | total fee (coins)  |

If additional fee is assessed, these events will also be emitted (reason for not always emitting them mainly saving space on block output)

## Tx with Additional Fee

| Type     | Attribute Key | Attribute Value                                                    |
| -------- | ------------- | -------------------------------------------------------------------|
| message  | additionalfee | additional fee charged (coins)                                     |
| message  | basefee       | total fee - additional fee, should always cover gas costs (coins)  |

## Tx Summary Event

If there are tx msgs that have additional fees.  A summary event will be emitted that contains the `type_url`, `count`, and `total_fees`.

| Type         | Attribute Key | Attribute Value                                                                                |
| ------------ | ------------- | ---------------------------------------------------------------------------------------------- |
| EventMsgFees | MsgFees       | a list of EventMsgFee that summarize the msg fees for each msg type                            |
| EventMsgFee  | type_url      | the type url for the tx msg that has a msg fee                                                 |
| EventMsgFee  | count         | count of txs with this msg type                                                                |
| EventMsgFee  | total         | the total amount that of additional fees for this msg type (type_url count * msg fee = total)  |
| EventMsgFee  | recipient     | the bech32 address that the fee was sent to.  This is when an assess custom fee msg dispatched |

Note: EventMsgFee `total` is not the cost per msg type, but the sum of msg fees for the number of calls.  Cost per message: `total / count = fee per msg`
## Add/Update/Remove Proposal
 
Governance proposals events(for proposed msg fees) will continue to be emitted by cosmos sdk.
 (https://github.com/cosmos/cosmos-sdk/blob/master/x/gov/spec/04_events.md)
