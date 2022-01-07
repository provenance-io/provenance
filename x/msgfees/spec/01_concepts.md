<!--
order: 1
-->

# Concepts
The msg fees modules manages additional fees that can be applied to tx msgs specified through governance.

<!-- TOC -->
  - [Additional Msg Fees](#additional-msg-fees)
  - [Base Fee](#base-fee)
  - [Total Fees](#total-fees)
  - [Additional Fee Assessed in Base Denom i.e nhash](#additional-fee-assessed-in-base-denom-i-e-nhash)
  - [Authz and Wamsd Messages](#authz-and-wamsd-messages)
  - [Simulation and Calculating the Additional Fee to be Paid](#simulation-and-calculating-the-additional-fee-to-be-paid)



## Additional Msg Fees

Fees is one of the most important tools available to secure a PoS network since it incentives staking and encourages spam prevention etc.

As part of the provenance blockchain economics certain messages *may* require an additional fee to be paid in
addition to the normal gas consumption. 

Additional fees are assessed and finally consumed based on msgType of the msgs contained in the transaction,
and the fee schedule that is persisted on chain.  These additional fees are created/updated/removed by governance through `AddMsgFeeProposal`, `UpdateMsgFeeProposal`, and `RemoveMsgFeeProposal` proposals.

Additional fee can be in any *denom*.

## Base Fee

Base fee is the current fee implementation. Fees are paid in base denom and determined by gas value passed into the Tx.
The value collected remains the same.

## Total Fees

Total fees = Additional Fees (if any) + Base Fee
Total fees continue to be passed as `sdk.Coins` the Tx accepts for fee entry currently.
e.g usd.example is the denom(assuming there is a marker/coin of type usd.example) in which additional fee is being charged in
```bash
--fees 382199010nhash,99usd.example 
```

## Additional Fee Assessed in Base Denom i.e nhash

To preserve backwards compatibility of all invokes, clients continue accepting fees in sdk.Coins `type Coins []Coin`, and because the code needs to distinguish between base fee and additional fee, the msgfees module introduces an additional param, described in [params documentation](06_params.md), called `DefaultFloorGasPrice` to differentiate between base fee and additional fee when additional fee is in same denom as default base denom i.e nhash.

This fee is charged initially by the antehandler, if any excess fee is left, once additional fee are paid, that's collected
at the end of the Tx also(same as current behavior)

For e.g
Additional fee = 10000nhash
Gas = 10000
Fee passed in = 19070000nhash

In this client passes in an extra 10000nhash (1905 * 10000 +10000 = 19060000nhash).
Current behavior is maintained and tx passes and charges 19050000 initially and 1000 nhash plus 1000nhash extra fee passed in the deliverTx stage.
Thus, this will protect against future changes like priority mempool as well as keep current behavior same as current production. 

## Authz and Wamsd Messages

Authz and wasmd messages are dispatched via the submessages route, so they get charged and assessed the same additional
fee if set on a submessage, caveat being they forfeit all their fees if they fail (since we have no way upfront of knowing what 
the submessages maybe)

For Example, let's say a `MsgSend` has a fee of 100usd.local and a smart contract does 3 MsgSend operations as per the logic of the smart contract, the code will expect additional fees of 300 usd.local (3 msgs x 100usd.local) to be present for the Tx to be successful.

## Simulation and Calculating the Additional Fee to be Paid

Current simulation method looks like this:  
```kotlin
val cosmosService = cosmos.tx.v1beta1.ServiceGrpc.newBlockingStub(channel)
cosmosService.simulate(SimulateRequest.newBuilder().setTx(txFinal).build()).gasInfo.gasUsed
            
```

In the future we recommend using the method: 
```kotlin
val msgFeeClient = io.provenance.msgfees.v1.QueryGrpc.newBlockingStub(channel)
msgFeeClient.calculateTxFees(CalculateTxFeesRequest.newBuilder().setTx(txFinal).build())

```

or from the cmd line as:

```bash
provenanced tx simulate <required params>
```
