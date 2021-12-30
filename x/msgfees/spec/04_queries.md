<!--
order: 4
-->

# MsgFees Queries


## Msg/GenesisState

GenesisState contains a set of msg fees, exported and later imported from/to the store.
[genesis.proto](../../../proto/provenance/msgfees/v1/genesis.proto?plain=1)


## Query Request/Response Object
get params for the module. [get params](../../../proto/provenance/msgfees/v1/query.proto?plain=1)  

[query all msgfees in the system](../../../proto/provenance/msgfees/v1/query.proto?plain=1)
QueryAllMsgFeesRequest/QueryAllMsgFeesResponse resquest/response for all messages
which have fees associated with them.

[simuate fees(including additional fees to be paid for a Tx)](../../../proto/provenance/msgfees/v1/query.proto?plain=1)
To simulate the fees required on the Tx use CalculateTxFeesRequest

Request: [CalculateTxFeesRequest](../../../proto/provenance/msgfees/v1/query.proto#L59-L68)
```protobuf
// CalculateTxFeesRequest is the request type for the Query RPC method.
message CalculateTxFeesRequest {
  // tx_bytes is the transaction to simulate.
  bytes tx_bytes = 1;
  // default_base_denom is used to set the denom used for gas fees
  // if not set it will default to nhash.
  string default_base_denom = 2;
  // gas_adjustment is the adjustment factor to be multiplied against the estimate returned by the tx simulation
  float gas_adjustment = 3;
}
```
Response: [CalculateTxFeesResponse](../../../proto/provenance/msgfees/v1/query.proto#L70-L81)
```protobuf
// CalculateTxFeesResponse is the response type for the Query RPC method.
message CalculateTxFeesResponse {
  // additional_fees are the amount of coins to be for addition msg fees
  repeated cosmos.base.v1beta1.Coin additional_fees = 1
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
  // total_fees are the total amount of fees needed for the transactions (msg fees + gas fee)
  // note: the gas fee is calculated with the min gas fee param as a constant
  repeated cosmos.base.v1beta1.Coin total_fees = 2
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
  // estimated_gas is the amount of gas needed for the transaction
  uint64 estimated_gas = 3 [(gogoproto.moretags) = "yaml:\"estimated_gas\""];
}
```
total fee is calculated based on `floor_gas_price` param set to 1905nhash for now.
