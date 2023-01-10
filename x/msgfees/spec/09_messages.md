# Messages

In this section we describe the messages that are used in the msgfees module.

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