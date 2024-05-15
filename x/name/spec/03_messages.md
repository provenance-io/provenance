# Messages

In this section we describe the processing of the staking messages and the corresponding updates to the state.

<!-- TOC -->
  - [MsgBindNameRequest](#msgbindnamerequest)
  - [MsgDeleteNameRequest](#msgdeletenamerequest)
  - [MsgModifyNameRequest](#msgmodifynamerequest)
  - [MsgCreateRootNameRequest](#msgcreaterootnamerequest)

## MsgBindNameRequest

A name record is created using the `MsgBindNameRequest` message.

```proto
message MsgBindNameRequest {
  option (gogoproto.equal)           = false;
  option (gogoproto.goproto_getters) = false;

  // The parent record to bind this name under.
  NameRecord parent = 1 [(gogoproto.nullable) = false];
  // The name record to bind under the parent
  NameRecord record = 2 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:
- The parent name record does not exist
- The requestor does not match the owner listed on the parent record _and_ the parent record indicates creation of child records is restricted.
- The record being created is otherwise invalid due to format or contents of the name value itself
    - Insuffient length of name
    - Excessive length of name
    - Not deriving from the parent record (targets another root)

If successful a name record will be created as described and an address index record will be created for the address associated with the name.
## MsgDeleteNameRequest

The delete name request method allows a name record that does not contain any children records to be removed from the system.  All 
associated attributes on account addresses will be deleted.

```proto
// MsgDeleteNameRequest defines an sdk.Msg type that is used to remove an existing address/name binding.  The binding
// may not have any child names currently bound for this request to be successful. All associated attributes on account addresses will be deleted.
message MsgDeleteNameRequest {
  option (gogoproto.equal)           = false;
  option (gogoproto.goproto_getters) = false;

  // The parent record the record to remove is under.
  NameRecord parent = 1 [(gogoproto.nullable) = false];
  // The record being removed
  NameRecord record = 2 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:
- Any components of the request do not pass basic integrity and format checks
- The parent name record does not exist
- The record to remove does not exist
- Any child records exist under the record being removed
- The requestor does not match the owner listed on the record.

## MsgModifyNameRequest

A name record is modified by proposing the `MsgModifyNameRequest` message.

```proto
// MsgModifyNameRequest defines a method that is used to update an existing address/name binding.
message MsgModifyNameRequest {
  option (cosmos.msg.v1.signer) = "authority";

  // The address signing the message
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // The record being updated
  NameRecord record = 2 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:
- Any components of the request do not pass basic integrity and format checks
- The record to update does not exist
- The authority does not match the gov module or the name owner.

If successful a name record will be updated with the new address and restriction.

## MsgCreateRootNameRequest

The `MsgCreateRootNameRequest` is a governance proposal that allows new root level names to be established after the genesis of the blockchain.

```proto
message MsgCreateRootNameRequest {
  option (cosmos.msg.v1.signer)    = "authority";

  // The signing authority for the request
  string authority   = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // NameRecord is a structure used to bind ownership of a name hierarchy to a collection of addresses
  NameRecord record  = 2;
}
```

This message is expected to fail if:
- The name already exists
- Insuffient length of name
- Excessive length of name
- The authority does not match the gov module.

If successful a name record will be created with the provided address and restriction.
