# MsgFees Queries

Since the msgfees module is deprecated, most queries have been removed except `CalculateTxFees`.

## CalculateTxFees

This query has been deprecated in favor of a query in the `x/flatfees` module with the same name.
This query will be removed in a future version (possibly the next), but is still available to help users transition to the new query.
This query is implemented by calling the `x/flatfees` version and returning its response anyway.
Now, only the `total_fees` and `estimated_gas` fields will be populated.

Also, the `gas_adjustment` now only applies to the returned `estimated_gas` field, and does not affect the `total_fees` (since those are flat now).

### CalculateTxFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/msgfees/v1/query.proto#L26-L38

### CalculateTxFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/msgfees/v1/query.proto#L40-L62

### CalculateTxFees Client Info

The standard SDK's simulation method looks like this:

```kotlin
val cosmosService = cosmos.tx.v1beta1.ServiceGrpc.newBlockingStub(channel)
cosmosService.simulate(SimulateRequest.newBuilder().setTx(txFinal).build()).gasInfo.gasUsed
```

Using the msgsfees query looks like this (deprecated):

```kotlin
val msgFeeClient = io.provenance.msgfees.v1.QueryGrpc.newBlockingStub(channel)
msgFeeClient.calculateTxFees(CalculateTxFeesRequest.newBuilder().setTx(txFinal).build())
```

Using the flatfees query looks like this:

```kotlin
val flatFeeClient = io.provenance.flatfees.v1.QueryGrpc.newBlockingStub(channel)
flatFeeClient.calculateTxFees(QueryCalculateTxFeesRequest.newBuilder().setTx(txFinal).build())
```

Or from the cmd line as (this uses the flatfees query now):

```bash
provenanced tx simulate <required params>
```
