# FlatFees Messages

The `x/flatfees` module only has `Msg` endpoints for governance proposals.

---
<!-- TOC -->
  - [Governance Proposals](#governance-proposals)
    - [UpdateParams](#updateparams)
    - [UpdateMsgFees](#updatemsgfees)


## Governance Proposals

There are a couple endpoints available as governance proposals that manage the FlatFees module behavior.


### UpdateParams

The flatfees module params are updated via governance proposal with a `MsgUpdateParamsRequest`.

It is expected to fail if:
* The provided `authority` is not the governance module's account.
* The provided params are invalid.

#### MsgUpdateParamsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.24.0/proto/provenance/flatfees/v1/tx.proto#L24-L33

See also: [Params](06_params.md#params).

#### MsgUpdateParamsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.24.0/proto/provenance/flatfees/v1/tx.proto#L35-L36


### UpdateMsgFees

The costs for specific msg types are managed via governance proposal with a `MsgUpdateMsgFeesRequest`.

To add or update a msg fee, include it in the `to_set` field.
To unset a msg fee, include the Msg type URL in the `to_unset` field.
Unsetting a msg fee will make that msg type cost the default (defined in params).

It is expected to fail if:
* The provided `authority` is not the governance module's account.
* Any `MsgFee` entries in `to_set` are invalid.
* Both `to_set` and `to_unset` are empty.
* A Msg type URL is listed more than once among both the `to_set` and `to_unset` fields.

#### MsgUpdateMsgFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.24.0/proto/provenance/flatfees/v1/tx.proto#L38-L50

#### MsgFee

+++ https://github.com/provenance-io/provenance/blob/v1.24.0/proto/provenance/flatfees/v1/flatfees.proto#L22-L35

#### MsgUpdateMsgFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.24.0/proto/provenance/flatfees/v1/tx.proto#L52-L53
