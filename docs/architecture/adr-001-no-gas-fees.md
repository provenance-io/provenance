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
1. All clients only pass in fee param with the flat fee required for that MsgType(or the base min fee if fee not defined for that Msg type)
2. Clients can simulate the fee from the provenance simulation endpoint.(ignore the gas returned, we will deprecate it and eventually remove it)
3. Add params to msgfees module to account for a base fee charged to prevent spam and failed tx's detailed below.

Two new params added to msgfee module params, `floor_fee` and `default_msg_fee`

```protobuf
// Params defines the set of params for the msgfees module.
message Params {
  option (gogoproto.equal)            = false;
  option (gogoproto.goproto_stringer) = true;
  // constant used to calculate fees when gas fees shares denom with msg fee
  cosmos.base.v1beta1.Coin floor_gas_price = 2 [(gogoproto.nullable) = false,deprecated=true];
  // total nhash per usd mil for converting usd to nhash
  uint64 nhash_per_usd_mil = 3;
  // conversion fee denom is the denom usd is converted to
  string conversion_fee_denom = 4;
  // constant to be used to charge a minimum fee for each transaction. for e.g 0.381 hash (1905nhash * 200000) 200k gas,
  // is the default gas set by the cli. This will be the fee to be charged for spam prevention and failed tx's.
  cosmos.base.v1beta1.Coin floor_fee = 5 [(gogoproto.nullable) = false];
  // since gas is not used for fees anymore, the design should accomodate a fee that should be charged if not present 
  // in the fee table, should be >= floor_fee
  cosmos.base.v1beta1.Coin default_msg_fee = 6 [(gogoproto.nullable) = false];
}
```

3. Change the `NewSetUpContextDecorator` to take in the limit specific for provenance
and apply it tot the `tx` , right now the limit is 4m and will continue to be so.
```go
cosmosante.NewSetUpContextDecorator(cosmosante.GasLimit{
Limit:         gasTxLimit,
OverrideGasTx: true,
})
```

additional flag is just not to have the cosmos-dk tests fail and will continue to apply the tx gas limit for that project.

4. The limit for any tx is 4million gas, no chnages need to happen the gas meters, except that when we run the messages in `baseapp.go` we set the limit to gas consumed, this keeps the tendermint check for maxBytes of a block 
and what can be included in a block in line to current scenario where the user estimates gas.
What we need in the baseapp is do do this
```go
	// GasMeter expected to be set in AnteHandler
	// however provenance will depend on message fees and should match up to just gas consumed
	gasWanted = ctx.GasMeter().GasConsumed()

	return gInfo, result, anteEvents, priority, ctx, err
```

4. Set every tx to start at gasmeter of 4 million (or whatever the max threshold is defined)
5. If they fail because of running out of gas even at 4m (or whatever the max threshold is defined) then we take the fees provided.
6. All failed tx's get charged the whole fee or a percentage of the fee, whatever is seen to be more appropriate.
7. Fee table is to be set up as params in the msg fee module which already exists.
8. Fee checks are already assessed in msgservicerouter for additional fees, so they probably will work similarly.
9. wasm module also go through the same checks in msgservicerouter so should be maneagable to change.
10. Major downside: Clients will have to understand this new construct but overall it should make the fee system easier hopefully.




