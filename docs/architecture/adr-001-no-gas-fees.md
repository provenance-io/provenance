# ADR 001: No gas fees

## Changelog

- 2022-10-26: Initial Draft

## Status

Initial research

## Context
Get rid of gas as a construct in provenance blockchain, and only keep the fee construct.
Gas was a means to put a limit on the system, we should still set it as having 
a max threshold value internally.

We think it will really make client interaction simpler if there was only fee to deal with, in proposing tx's to the provenance blockchain and to get rid of the gas construct.
So Proposal is to get rid of gas as a construct in provenance blockchain, and only keep the fee construct.
Gas was a means to put a limit on the system, we should still set it(internally only) as having a max threshold value but clients need not set it and also pay fees based on it. i.e if messages which exceed max threshold(currently 4 million) will still fail so as to protect the system.
Fees will be set using MsgFee module per message, for messages which are not in the list will be charged the min fee(let's say 0.,5 hash)

Tendermint considerations
```go
// addNewTransaction handles the ABCI CheckTx response for the first time a
// transaction is added to the mempool.  A recheck after a block is committed
// goes to handleRecheckResult.
//
// If either the application rejected the transaction or a post-check hook is
// defined and rejects the transaction, it is discarded.
//
// Otherwise, if the mempool is full, check for lower-priority transactions
// that can be evicted to make room for the new one. If no such transactions
// exist, this transaction is logged and dropped; otherwise the selected
// transactions are evicted.
//
// Finally, the new transaction is added and size stats updated.
func (txmp *TxMempool) addNewTransaction(wtx *WrappedTx, checkTxRes *abci.ResponseCheckTx) {
```

```go
	wtx.SetGasWanted(checkTxRes.GasWanted)
	wtx.SetPriority(priority)
	wtx.SetSender(sender)
	txmp.insertTx(wtx)
```

Proposal
1.All clients only pass in fee param with the flat fee required for that MsgType(or the base min fee if fee not defined for that Msg type)
2. Clients can simulate the fee from the provenance simualtion endpoint
3. We can change the cosmos endpoint if required, not sure if it's required.(and will add to the fork)
4. Set every tx to start at gasmeter of 4 million (or whatever the max threshold is defined)
5. If they fail because of running out of gas even at 4m (or whatever the max threshold is defined) then we take the fees provided.
6. All failed tx's get charged the whole fee or a percentage of the fee, whatever is seen to be more appropriate.
7. Fee table is to be set up as params in the msg fee module which already exists.
8. Fee checks are already assessed in msgservicerouter for additional fees, so they probably will work similarly.
9. wasm module also go through the same checks in msgservicerouter so should be maneagable to change.
10. Major downside: Clients will have to understand this new construct but overall it should make the fee system easier hopefully.


Tender