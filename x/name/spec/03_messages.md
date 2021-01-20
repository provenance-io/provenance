# Messages

In this section we describe the processing of the staking messages and the corresponding updates to the state.

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

The delete name request method allows a name record that does not contain any children records to be removed from the system.

```proto
// MsgDeleteNameRequest defines an sdk.Msg type that is used to remove an existing address/name binding.  The binding
// may not have any child names currently bound for this request to be successful.
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

## CreateRootNameProposal

The create root name proposal is a governance proposal that allows new root level names to be established after the genesis of the blockchain.

```proto
message CreateRootNameProposal {
  option (gogoproto.equal)            = false;
  option (gogoproto.goproto_getters)  = false;
  option (gogoproto.goproto_stringer) = false;

  string title       = 1;
  string description = 2;
  string name        = 3;
  string owner       = 4;
  bool   restricted  = 5;
}
```

This message is expected to fail if:
- The name already exists
- Insuffient length of name
- Excessive length of name