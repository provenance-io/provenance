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


## Add/Update/Remove Proposal
 
Governance proposals events(for proposed msg fees) will continue to be emitted by cosmos sdk.
 (https://github.com/cosmos/cosmos-sdk/blob/master/x/gov/spec/04_events.md)
