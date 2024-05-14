<!--
order: 2
-->

# State

[MsgFee proto](../../../proto/provenance/msgfees/v1/msgfees.proto#L31-L48)
```protobuf
// MsgFee is the core of what gets stored on the blockchain to define a msg-based fee.
message MsgFee {
  // msg_type_url is the type-url of the message with the added fee, e.g. "/cosmos.bank.v1beta1.MsgSend".
  string msg_type_url = 1;
  // additional_fee is the extra fee that is required for the given message type (can be in any denom).
  cosmos.base.v1beta1.Coin additional_fee = 2 [(gogoproto.nullable) = false];
  // recipient is an option address that will receive a portion of the additional fee.
  // There can only be a recipient if the recipient_basis_points is not zero.
  string recipient = 3;
  // recipient_basis_points is an optional portion of the additional fee to be sent to the recipient.
  // Must be between 0 and 10,000 (inclusive).
  //
  // If there is a recipient, this must not be zero. If there is not a recipient, this must be zero.
  //
  // The recipient will receive additional_fee * recipient_basis_points / 10,000.
  // The fee collector will receive the rest, i.e. additional_fee * (10,000 - recipient_basis_points) / 10,000.
  uint32 recipient_basis_points = 4;
}
```

This state is created via governance proposals.
