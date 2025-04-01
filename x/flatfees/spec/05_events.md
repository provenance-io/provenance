<!--
order: 5
-->

# Events

Existing fee event continue to show total fee charged

<!-- TOC -->
  - [Any Tx](#any-tx)
  - [Tx with Additional Fee](#tx-with-additional-fee)
  - [Tx Summary Event](#tx-summary-event)
  - [Add/Update/Remove Proposal](#addupdateremove-proposal)

## Any Tx

If a Tx was successful, or if it failed, but the min fee was charged, these two events are emitted:

| Type     | Attribute Key    | Attribute Value               |
| -------- |------------------|-------------------------------|
| tx       | fee              | total fee (coins)             |
| tx       | min_fee_charged  | floor gas price * gas (coins) |


## Tx with Additional Fee

If there are tx msgs that have additional fees, and those fees were successfully charged, a breakdown event will be emitted.

Type: tx

| Attribute Key | Attribute Value                                                    |
| ------------- | -------------------------------------------------------------------|
| additionalfee | additional fee charged (coins)                                     |
| basefee       | total fee - additional fee, should always cover gas costs (coins)  |

## Tx Summary Event

If there are tx msgs that have additional fees, and those fees were successfully charged, a summary event will be emitted.

Type: provenance.msgfees.v1.EventMsgFees

| Type         | Attribute Key | Attribute Value                                                             |
| ------------ | ------------- | --------------------------------------------------------------------------- |
| EventMsgFees | MsgFees       | A JSON list of EventMsgFee entries summarizing each msg type and recipient. |

Each `EventMsgFee` has the following fields:

| Field Name    | Field Value                                                                                            |
| ------------- | ------------------------------------------------------------------------------------------------------ |
| type_url      | The type url for the tx msg that has a msg fee.                                                        |
| count         | A count of txs with this msg type.                                                                     |
| total         | The total amount of additional fees for this msg type and recipient (type_url count * msg fee = total) |
| recipient     | the bech32 address that the fee was sent to. An empty string indicates the module is the recipient.    |

## Add/Update/Remove Proposal

Governance proposals events(for proposed msg fees) will continue to be emitted by cosmos sdk.
 (https://github.com/cosmos/cosmos-sdk/blob/master/x/gov/spec/04_events.md)
