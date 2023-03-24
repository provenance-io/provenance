# Authorization

The marker module supports granting authorizations for restricted coin transfers.  This is implemented using
the `authz` module's `Authorization` interface.

```
// MarkerTransferAuthorization gives the grantee permissions to execute
// a restricted coin transfer on behalf of the granter's account.
message MarkerTransferAuthorization {
  option (cosmos_proto.implements_interface) = "Authorization";

  // transfer_limit is the total amount the grantee can transfer
  repeated cosmos.base.v1beta1.Coin transfer_limit = 1
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];

  // allow_list specifies an optional list of addresses to whom the grantee can send tokens on behalf of the
  // granter. If omitted, any recipient is allowed.
  repeated string allow_list = 2;
}
```

With the `MarkerTransferAuthorization` a `granter` can allow a `grantee` to do transfers on their behalf.
A `transfer_limit` is required to be set for the `grantee`.
The `allow_list` is optional.
An empty list means any destination address is allowed, otherwise, the destination must be in the `allow_list`.