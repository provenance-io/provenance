# FlatFees Concepts

The `x/flatfees` module, in conjunction with a custom gas meter, antehandler, message handler, and post handler, make Provenance Blockchain charge fees based on the `Msg` types instead of gas.

Provenance Blockchain mainnet and testnet users should ALWAYS use `1nhash` for their gas prices, and should not use a gas adjustment.

---
<!-- TOC -->
  - [Fee definition](#fee-definition)
    - [Default Fee](#default-fee)
    - [Msg Fee](#msg-fee)
    - [Conversion Factor](#conversion-factor)
    - [Added Fees](#added-fees)
  - [Simulation](#simulation)
    - [Gas](#gas)
    - [The CalculateTxFees Query](#the-calculatetxfees-query)
    - [Standard SDK Simulation](#standard-sdk-simulation)
  - [Fee Collection](#fee-collection)
    - [Up-Front Cost](#up-front-cost)
    - [Post-Processing Cost](#post-processing-cost)
  - [Blockchain Customizations](#blockchain-customizations)
    - [Flat Fee Gas Meter](#flat-fee-gas-meter)
    - [Antehandler](#antehandler)
    - [Msg Handler](#msg-handler)
    - [Post Handler](#post-handler)


## Fee definition

Fees for messages are defined in one denom, then a conversion factor allows us to get the equivalent amount in the fee denom.
For example, we might say that `MsgSend` takes `25cusd`. Then, we use a conversion factor to calculate the amount of `nhash` that is worth `25cusd`.
This conversion factor can be updated in order to keep the cost of Msgs steady even as the price of `nhash` fluctuates.

### Default Fee

The default fee is defined in [Params](06_params.md).
Any message type that does not have specific fee defined will cost this amount.
The conversion factor is applied to this to identify the actual default cost.

This amount is also the most that will be charged for a Msg that passes `ValidateBasic()` but ultimately fails.

### Msg Fee

A Msg can cost more or less (or otherwise differently) than the default by defining a [MsgFee](03_messages.md#msgfee) for it.
The conversion factor is applied to these to identify the actual cost for the Msg.

### Conversion Factor

The conversion factor is also defined in [Params](06_params.md#conversionfactor).
It is used to convert the fees as defined into the actual fee required.

```
required fee = defined cost * converted_amount / definition_amount
```

When the conversion factor is applied, it only converts the coin with the same denom as the `definition_amount`, all other denominations are left unchanged.
For example, if the conversion factor is `1cusd` = `2nhash`, and a `MsgFee` has a cost of `10cusd,15peach`, then the required (converted) fee will be `20nhash,15peach`.

### Added Fees

While processing a Msg, an endpoint can use the `antehandler.ConsumeAdditionalFee(ctx, fee)` helper to assess an additional fee, possibly based on the contents of that Msg.
Such fees are added to the Msg fee (or default).
Any fees added this way are included in the fee returned when simulating a Tx.


## Simulation

Since the fee is no longer dependent on the amount of gas, users will need both amounts (gas and fees) when simulating a Tx.

There are two ways to simulate a Tx on a Provenance Blockchain:
1. Using this module's [CalculateTxFees](05_queries.md#calculatetxfees) query.
2. Standard SDK simulation.

### Gas

Gas is still metered.
Max gas for a Tx is 4,000,000 (defined in code).
Max gas for a block is 60,000,000 (defined in consensus params).

Users must still provide a `gas_wanted` that is greater than (or equal to) what gets used.
Failure to do so will cause the Tx will fail, and the user will be charged the up-front cost.

Gas used during simulation is still expected to be lower than gas used during actual Tx execution.

### The CalculateTxFees Query

This is the recommended way of simulating a Tx.
This query will simulate a Tx and return the required fee and estimated gas.
After simulating a Tx in this way, update the tx with the values returned and submit it as normal for processing.

### Standard SDK Simulation

Standard SDK simulation only allows communication of gas amounts.
However, we have our own custom `app.Simulate` method that is part of the standard SDKs Tx simulation process.
In it, the `gas_used` field gets set to the amount of `nhash` required as a fee (instead of containing the amount of gas used).
The `gas_wanted` field is set to the amount of gas actually used (as opposed to the amount of gas included in the request).

This is the process that gets used when users provide `--gas=auto` with a tx command.

Users should always use `1nhash` as the gas price, and the gas adjustment should be 1 (or not included).
By doing this, the generic Cosmos-SDK clients will multiply the `gas_used` by 1 (the gas adjustment), and then `1nhash` (the gas prices) to end up with the actual fee required for the Tx.
It essentially allows existing third-party clients to continue to work.

When a Tx is simulated and submitted this way, it will have the correct fee amount, but will also end up with that amount as the gas wanted (standard client behavior).
In our antehandler, we check for that case and use a default amount of gas for the Tx (500,000) instead of what was provided (which is probably more than the block maximum anyway).
This default is large enough to handle most Txs, but won't be enough for some.

If the default isn't enough, users have a few options:

1. Manually set their `gas_wanted` based on the `gas_wanted` returned from the simulation (probably still need to apply a multiplier).
2. Use the `CalculateTxFees` query to get both fee and gas information, and use those values instead.

This method of simulation also has a blind spot.
If the required fee has an amount in a denom other than the standard fee denom, info about it is not returned using this method of simulation.
In such a case, the `CalculateTxFees` must be used.


## Fee Collection

The fee for a Tx is collected in two parts.

1. Before its Msgs are processed (the up-front cost).
2. After its Msgs have been successfully processed (skipped if the tx fails).

### Up-Front Cost

A portion (or all) of the tx fee is collected before the Msgs are processed (as long as they all pass `ValidateBasic`).
This fee is collected even if the Tx fails.

The up-front cost of a Msg is the smaller of the cost of the Msg or the default Msg cost, and is only the portion in the fee denom.
This is calculated for each Msg in a tx, then summed.
For example, say the default cost is `5cusd`, and tx has three Msgs with costs `4cusd`, `5cusd`, and `6cusd`.
The up-front cost for the tx would be `14cusd` = `4cusd` (less than default) + `5cusd` (equal to default) + `5cusd` (more than default).
In that example, `1cusd` would still need to be collected at the end.

When calculating the up-front cost, only Msgs contained in the Tx (either directly or as sub-Msgs) are considered.
For example, if a Tx has a smart contract execution in it that will issue Msgs as part of its operation, those Msgs are not accounted for in the up-front cost.
Such messages are accounted for as they are run, and their fees will be collected upon success.

### Post-Processing Cost

After all the Msgs in a tx have been processed, if they were all successful, the rest of the fee is collected.
This collection only happens if all the Msgs were successful.
The amount collected is the amount of fees provided with the tx, less the up-front cost (that was already collected).
In other words, with a successful Tx, all of the fees provided with the Tx are collected, even if that is more than what was required.

If, while processing Msgs, new Msgs are identified (e.g. from a smart contract execution), or custom fees were added (e.g. from `ConsumeAdditionalFee`),
the fee provided with the Tx is re-verified against the updated required fee amount to make sure enough was provided.
That means it's possible to get charged the up-front cost, but still have the Tx fail due to insufficient fees, (but should be very rare).

In many cases, the entirety of the cost is collected up-front, and there is nothing left to collect in the end.


## Blockchain Customizations

This flatfees module is only part of how we charge by msg type (instead of gas).
Several aspects of the blockchain have been customized too.

### Flat Fee Gas Meter

Provenance Blockchain uses a custom gas meter that enhances another gas meter to track Msgs and costs as well as gas.
The antehandler creates it and sets the initial costs, the collects the up-front cost.
The post handler finalizes it before collecting anything remaining.

If `FlatFeeGasMeter.ConsumeAddedFee` or `antehandler.ConsumeAdditionalFee` is called while processing a Msg, 
that amount is added to the amount of fee required for the Tx (which is checked again in the post handler).

### Antehandler

The Provenance Blockchain antehandler has some customized processes.
The base gas meter is created with a limit possibly different from the gas wanted (to account for the generic simulation process).
It is then wrapped in a Flat Fee Gas Meter and set in the context.
This is also where we ensure that gas isn't more than the tx or block limits.

Our antehandler also identifies the up-front costs and ensures that the fee provided with the Tx is at least that amount.
Eventually, it then collects that up-front cost.

It does most of the standard SDK things too.

The antehandler emits [two](04_events.md#generic-fee-event) [events](04_events.md#initial-fee-event).
The amount in the generic fee event is initially the full amount of fee.
If one or more of its Msgs fail, though, the event aggregator is used to change that amount to the `min_fee_charged` from the other event.
We do this so that the generic fee event always has the amount of fee collected.

### Msg Handler

The Provence Blockchain uses a custom Msg handler that is almost identical to the standard SDK one.
Ours calls `ConsumeMsg` on the gas meter, to record that a message is being processed.
The gas meter keeps track of Msgs in this way so that we can charge for Msgs that aren't directly in the Tx (e.g. from a smart contract execution).

### Post Handler

The Provenance Blockchain uses a post handler to collect the remainder of the fee.
If the Tx was not successful, the post handler does nothing.

In the post handler, it looks for costs added since the antehandler.
If there were any, the provided fee amount is rechecked to make sure it is still enough.
If not enough, the Tx fails (and we keep the up-front cost).
If enough (or nothing new was added), the remainder of the provided tx fee is collected (unless it was all collected up-front).

The post handler also emits [an event](04_events.md#success-fee-event) with a breakdown of the fees.
