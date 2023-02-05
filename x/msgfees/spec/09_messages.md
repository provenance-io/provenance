# Messages

In this section we describe the messages that are used in the msgfees module.
  - [MsgAssessCustomMsgFeeRequest](#MsgAssessCustomMsgFeeRequest)
  - [MsgAddMsgFeeProposalRequest](#MsgAddMsgFeeProposalRequest)
  - [MsgUpdateMsgFeeProposalRequest](#MsgUpdateMsgFeeProposalRequest)
  - [MsgRemoveMsgFeeProposalRequest](#MsgRemoveMsgFeeProposalRequest)
  - [MsgUpdateNhashPerUsdMilProposalRequest](#MsgUpdateNhashPerUsdMilProposalRequest)
  - [MsgUpdateConversionFeeDenomProposalRequest](#MsgUpdateConversionFeeDenomProposalRequest)

## MsgAssessCustomMsgFeeRequest

A custom fee is applied when this message is broadcast. This would be used in a smart contract to charge a custom fee for the usage.  

```proto
// MsgAssessCustomMsgFeeRequest defines an sdk.Msg type
message MsgAssessCustomMsgFeeRequest {
  option (gogoproto.equal)            = false;
  option (gogoproto.goproto_getters)  = false;
  option (gogoproto.goproto_stringer) = true;

  string name = 1; // optional short name for custom msg fee, this will be emitted as a property of the event
  cosmos.base.v1beta1.Coin amount = 2 [(gogoproto.nullable) = false]; // amount of additional fee that must be paid
  string recipient = 3; // optional recipient address, the basis points amount is sent to the recipient
  string from      = 4; // the signer of the msg
  string recipient_basis_points = 5; // optional basis points 0 - 10,000 for recipient defaults to 10,000
}
```

The `amount` must be in `usd` or `nhash` else the msg will not pass validation.  If the amount is specified as `usd` this will be converted
to `nhash` using the `UsdConversionRate` param.  Note: `usd` and `UsdConversionRate` are specified in mils.  Example: 1234 = $1.234

The `recipient` is a bech32 address of an account that will receive the amount calculated from the `recipient_basis_points`.  If the `recipient_basis_points` is left empty the whole `amount` will be sent to the recipient.  The remainder is sent the the Fee Module.

## MsgAddMsgFeeProposalRequest
```proto
// AddMsgFeeProposal defines a governance proposal to add additional msg based fee
message MsgAddMsgFeeProposalRequest {
  option (gogoproto.goproto_stringer) = true;
  option (cosmos.msg.v1.signer) = "authority";

  // type url of msg to add fee
  string msg_type_url = 1;

  // additional fee for msg type
  cosmos.base.v1beta1.Coin additional_fee = 2 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (gogoproto.moretags)     = "yaml:\"additional_fee\""
  ];

  // optional recipient to receive basis points
  string recipient = 3;
  // basis points to use when recipient is present (1 - 10,000)
  string recipient_basis_points = 4;
  // the signing authority for the proposal
  string authority = 5 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

## MsgUpdateMsgFeeProposalRequest

```proto 
// UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee
message MsgUpdateMsgFeeProposalRequest {
  option (gogoproto.goproto_stringer) = true;
  option (cosmos.msg.v1.signer) = "authority";

  // type url of msg to update fee
  string msg_type_url = 1;

  // additional fee for msg type
  cosmos.base.v1beta1.Coin additional_fee = 2
  [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
  // optional recipient to recieve basis points
  string recipient = 3;
  // basis points to use when recipient is present (1 - 10,000)
  string recipient_basis_points = 4;
  // the signing authority for the proposal
  string authority = 5 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

## MsgRemoveMsgFeeProposalRequest

```proto 
// RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee
message MsgRemoveMsgFeeProposalRequest {
  option (gogoproto.goproto_stringer) = true;
  option (cosmos.msg.v1.signer) = "authority";

  // type url of msg fee to remove
  string msg_type_url = 1;
  // the signing authority for the proposal
  string authority = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"]; //
}

```

## MsgUpdateNhashPerUsdMilProposalRequest

```proto
// UpdateNhashPerUsdMilProposal defines a governance proposal to update the nhash per usd mil param
message MsgUpdateNhashPerUsdMilProposalRequest {
  option (gogoproto.goproto_stringer) = true;
  option (cosmos.msg.v1.signer) = "authority";

  // nhash_per_usd_mil is number of nhash per usd mil
  uint64 nhash_per_usd_mil = 1;
  // the signing authority for the proposal
  string authority = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"]; //
}
```

## MsgUpdateConversionFeeDenomProposalRequest

```proto 
// UpdateConversionFeeDenomProposal defines a governance proposal to update the msg fee conversion denom
message MsgUpdateConversionFeeDenomProposalRequest {
  option (gogoproto.goproto_stringer) = true;
  option (cosmos.msg.v1.signer) = "authority";

  // conversion_fee_denom is the denom that usd will be converted to
  string conversion_fee_denom = 1;
  // the signing authority for the proposal
  string authority = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"]; //
}
```
