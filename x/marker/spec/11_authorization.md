# Authorization

The marker module supports granting authorizations for restricted coin transfers.  This is implemented using
the `authz` module's `Authorization` interface.

### MarkerTransferAuthorization :
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

### MultiAuthorization :

A `MultiAuthorization` contains multiple sub-authorizations that must each be satisfied for the authorization to be valid.

For example, consider a `MultiAuthorization` with a `CountAuthorization` with `2` uses, and a `SendAuthorization` with `500nhash`.

* If the grantee uses it once to send `500nhash`, then the `SendAuthorization` will be depleted and the whole `MultiAuthorization` will become unusable (be deleted) even though there's `1` more use left on the `CountAuthorization`.
* If the grantee uses it twice to send a total of `400nhash`, then the `CountAuthorization` will be depleted and the whole `MultiAuthorization` will become unusable (be deleted) even though there's still `100nhash` left in the `SendAuthorization`.

```
// MultiAuthorization lets you combine several authorizations.
// All sub-authorizations must accept the message for it to be allowed.
message MultiAuthorization {
  option (cosmos_proto.implements_interface) = "Authorization";

  // The message type this authorization is for.
  string msg_type_url = 1;

  // A list of sub-authorizations that must all accept the message.
  // Must have at least 2 and at most 10 items.
  repeated google.protobuf.Any sub_authorizations = 2 [(cosmos_proto.accepts_interface) = "Authorization"];
}
```
With `MultiAuthorization`, a `granter` can grant permission to a `grantee` with multiple restrictions.
* `msg_type_url` defines which message type this applies to (e.g. /cosmos.bank.v1beta1.MsgSend).
* `sub_authorizations` is a list of other authorizations (like GenericAuthorization) that all must accept the message.
