# State

The expiration module maintains a simple state collection. The `Expiration` is the data structure stored on the blockchain.
It consists of the following five parts:

- the module asset ID
- the owner of the expiration record
- the time the expiration will expire
- the deposit held for storing assets on chain
- the `sdk.Msg` type that will later delete the expiring asset upon invocation of the expiration message.

The expiration record is created when expiring module assets are created.

+++ https://github.com/provenance-io/provenance/blob/eb569b71b4d9137272432df5968cd62bf1eca2fb/proto/provenance/expiration/v1/expiration.proto#L23-L35

```protobuf
// Expiration holds a typed key/value structure for data associated with an expiring module asset
message Expiration {
  // the module asset identifier
  string module_asset_id = 1;
  // The bech32 address the expiration is bound to
  string owner = 2;
  // The time the module asset expires
  google.protobuf.Timestamp time = 3 [(gogoproto.stdtime) = true, (gogoproto.nullable) = false];
  // The deposit amount held while module asset is in use
  cosmos.base.v1beta1.Coin deposit = 4 [(gogoproto.nullable) = false];
  // Message relating to the expiring module asset
  google.protobuf.Any message = 5 [(gogoproto.nullable) = false];
}

```
