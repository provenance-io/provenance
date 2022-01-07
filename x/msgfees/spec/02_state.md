<!--
order: 2
-->

# State

`MsgFee` is the core of what gets stored on the blockchain it consists of two parts
 1. the msg type url, i.e. /cosmos.bank.v1beta1.MsgSend
 2. minimum additional fees(can be of any denom)
 
 [MsgFee proto](../../../proto/provenance/msgfees/v1/msgfees.proto#L25-L37) 
```protobuf
message MsgFee {
  string msg_type_url = 1;
  // additional_fee can pay in any Coin( basically a Denom and Amount, Amount can be zero)
  cosmos.base.v1beta1.Coin additional_fee = 2 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (gogoproto.moretags)     = "yaml:\"additional_fee\""
  ];
}
```

This state is created via governance proposals.
