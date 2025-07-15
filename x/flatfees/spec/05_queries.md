# FlatFees Queries

There are a few queries available in the flatfees module.

---
<!-- TOC 2 2 -->
  - [Params](#params)
  - [AllMsgFees](#allmsgfees)
  - [MsgFee](#msgfee)
  - [CalculateTxFees](#calculatetxfees)


## Params

The flatfees module params can be looked up using the `Params` query.

### QueryParamsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/query.proto#L42-L43

### QueryParamsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/query.proto#L45-L49

See also: [Params](06_params.md#params).


## AllMsgFees

To get info on all msg types that have a customized flat fee, use the `AllMsgFees` query.
If a Msg type is not returned by this query, it uses the default cost.

By default, the amounts are converted to the fee denom (using the params conversion factor).
To skip this conversion (and get the costs as they are defined) set `do_not_convert` to `true`.

### QueryAllMsgFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/query.proto#L51-L57

### QueryAllMsgFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/query.proto#L59-L65

See also: [MsgFee](03_messages.md#msgfee).


## MsgFee

To get the cost of a specific msg type, use the `MsgFee` query.
If there isn't a specific entry for the provided Msg type URL, the default cost is returned.

By default, the amount is converted to the fee denom (using the params conversion factor).
To skip this conversion (and get the cost as it is defined) set `do_not_convert` to `true`.

### QueryMsgFeeRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/query.proto#L67-L73

### QueryMsgFeeResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/query.proto#L75-L79

See also: [MsgFee](03_messages.md#msgfee).


## CalculateTxFees

The `CalculateTxFees` is a replacement for tx simulation that returns both gas and fees needed for a Tx.

The `gas_adjustment` only applies to the `estimated_gas` (it does not affect the `total_fees`).

### QueryCalculateTxFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/query.proto#L81-L89

### QueryCalculateTxFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/flatfees/v1/query.proto#L91-L102
