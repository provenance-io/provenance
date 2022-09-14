<!--
order: 2
-->

# State

`MsgFee` is the core of what gets stored on the blockchain it consists of four parts
 1. the msg type url, i.e. /cosmos.bank.v1beta1.MsgSend
 2. minimum additional fees(can be of any denom)
 3. optional recipient of fee based on `recipient_basis_points`
 4. if recipient is declared they will recieve the basis points of the fee (0-10,000)
 
 [MsgFee proto](../../../proto/provenance/msgfees/v1/msgfees.proto#L25-L37) 
```protobuf
message MsgFee {
  string msg_type_url = 1;
  // additional_fee can pay in any Coin( basically a Denom and Amount, Amount can be zero)
  cosmos.base.v1beta1.Coin additional_fee = 2
      [(gogoproto.nullable) = false, (gogoproto.moretags) = "yaml:\"additional_fee\""];
  string recipient = 3; // optional recipient address, the amount is split between recipient and fee module
  uint32 recipient_basis_points =
      4; // optional split of funds between the recipient and fee module defaults to 50:50 split
}
```

This state is created via governance proposals.
