# Messages

In this section we describe the processing of the staking messages and the corresponding updates to the state.

<!-- TOC -->
  - [MsgAddAttributeRequest](#msgaddattributerequest)
  - [MsgUpdateAttributeRequest](#msgupdateattributerequest)
  - [MsgUpdateAttributeExpirationRequest](#msgupdateattributeexpirationrequest)
  - [MsgDeleteAttributeRequest](#msgdeleteattributerequest)
  - [MsgDeleteDistinctAttributeRequest](#msgdeletedistinctattributerequest)
  - [MsgSetAccountDataRequest](#msgsetaccountdatarequest)



## MsgAddAttributeRequest

An attribute record is created using the `MsgAddAttributeRequest` message.

```proto
// MsgAddAttributeRequest defines an sdk.Msg type that is used to add a new attribute to an account.
// Attributes may only be set in an account by the account that the attribute name resolves to.
message MsgAddAttributeRequest {
  option (gogoproto.equal)            = false;
  option (gogoproto.goproto_stringer) = false;
  option (gogoproto.stringer)         = false;
  option (gogoproto.goproto_getters)  = false;

  // The attribute name.
  string name = 1;
  // The attribute value.
  bytes value = 2;
  // The attribute value type.
  AttributeType attribute_type = 3;
  // The account to add the attribute to.
  string account = 4;
  // The address that the name must resolve to.
  string owner = 5;
  // Time that an attribute will expire.
  google.protobuf.Timestamp expiration_date = 6 [(gogoproto.stdtime) = true, (gogoproto.nullable) = true];
}
```

This message is expected to fail if:
- Any components of the request do not pass basic integrity and format checks
- Attribute value exceeds the maximum length
- Unable to normalize the name
- The account does not exist
- The name does not resolve to the owner address

If successful, an attribute record will be created for the account.

## MsgUpdateAttributeRequest

The update attribute request method allows an existing attribute record to replace its value with a new one.

```proto
// MsgUpdateAttributeRequest defines an sdk.Msg type that is used to update an existing attribute to an account.
// Attributes may only be set in an account by the account that the attribute name resolves to.
message MsgUpdateAttributeRequest {
  option (gogoproto.equal)            = false;
  option (gogoproto.goproto_stringer) = false;
  option (gogoproto.stringer)         = false;
  option (gogoproto.goproto_getters)  = false;

  // The attribute name.
  string name = 1;
  // The original attribute value.
  bytes original_value = 2;
  // The update attribute value.
  bytes update_value = 3;
  // The original attribute value type.
  AttributeType original_attribute_type = 4;
  // The update attribute value type.
  AttributeType update_attribute_type = 5;
  // The account to add the attribute to.
  string account = 6;
  // The address that the name must resolve to.
  string owner = 7;
}
```

This message is expected to fail if:
- Any components of the request do not pass basic integrity and format checks
- Updated attribute value exceeds the maximum length
- Unable to normalize the original or updated attribute name
- Updated name and the original name don't match
- The owner account does not exist
- The updated name does not resolve to the owner address
- The original attribute does not exist

If successful, the value of an attribute will be updated.

## MsgUpdateAttributeExpirationRequest

The update attribute expiration request method updates the attribute's expiration date.

```proto
// MsgUpdateAttributeExpirationRequest defines an sdk.Msg type that is used to update an existing attribute's expiration
// date
message MsgUpdateAttributeExpirationRequest {
  option (gogoproto.equal)            = true;
  option (gogoproto.stringer)         = true;
  option (gogoproto.goproto_stringer) = false;

  // The attribute name.
  string name = 1;
  // The original attribute value.
  bytes value = 2;
  // Time that an attribute will expire.
  google.protobuf.Timestamp expiration_date = 3 [(gogoproto.stdtime) = true, (gogoproto.nullable) = true];
  // The account to add the attribute to.
  string account = 4;
  // The address that the name must resolve to.
  string owner = 5;
}
```

This message is expected to fail if:
- Any components of the request do not pass basic integrity and format checks
- The owner account does not exist
- The name does not resolve to the owner address
- The attribute does not exist
- The expiration date is before current block height

## MsgDeleteAttributeRequest

The delete distinct attribute request method removes an existing account attribute.

```proto
// MsgDeleteAttributeRequest defines a message to delete an attribute from an account
// Attributes may only be removed from an account by the account that the attribute name resolves to.
message MsgDeleteAttributeRequest {
  option (gogoproto.equal)            = false;
  option (gogoproto.goproto_stringer) = false;
  option (gogoproto.stringer)         = false;
  option (gogoproto.goproto_getters)  = false;

  // The attribute name.
  string name = 1;
  // The account to add the attribute to.
  string account = 2;
  // The address that the name must resolve to.
  string owner = 3;
}
```

This message is expected to fail if:
- Any components of the request do not pass basic integrity and format checks
- The owner account does not exist
- The name does not resolve to the owner address
- The attribute does not exist

## MsgDeleteDistinctAttributeRequest

The delete distinct attribute request method removes an existing account attribute with a specific value.

```proto
// MsgDeleteDistinctAttributeRequest defines a message to delete an attribute with matching name, value, and type from
// an account. Attributes may only be removed from an account by the account that the attribute name resolves to.
message MsgDeleteDistinctAttributeRequest {
  option (gogoproto.equal)            = false;
  option (gogoproto.goproto_stringer) = false;
  option (gogoproto.stringer)         = false;
  option (gogoproto.goproto_getters)  = false;

  // The attribute name.
  string name = 1;
  // The attribute value.
  bytes value = 2;
  // The account to add the attribute to.
  string account = 3;
  // The address that the name must resolve to.
  string owner = 4;
}
```

This message is expected to fail if:
- Any components of the request do not pass basic integrity and format checks
- The owner account does not exist
- The name does not resolve to the owner address
- The attribute does not exist

## MsgSetAccountDataRequest

The set account data request method associates some data (a string) with an account.

```protobuf
// MsgSetAccountDataRequest defines a message to set an account's accountdata attribute.
message MsgSetAccountDataRequest {
  option (cosmos.msg.v1.signer) = "account";

  string value   = 1;
  string account = 2;
}
```

This message is expected to fail if:
- The value is too long (as defined in attribute module params).
- The message is not signed by the provided account.
