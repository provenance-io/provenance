# Concepts

The name service builds up a heirarchy of names similar to DNS using dot separated strings.  Each level in the heirarchy
can be setup with an account that "owns" the name.  This owner must sign transactions that seek to add new names under 
this level.  Names created under another name can have a new owner thus transfering control from one account to another.

## Delegating Control

Every label in a name is owned by an address.  Starting from the root address each level can be configured to allow any user to add a new child or for the exclusive control of the creator to add child names.  The `Restricted` flag is used to indicate the permission requirements for adding child nodes.

```proto
// NameRecord is a structure used to bind ownership of a name heirarchy to a collection of addresses
message NameRecord {
  option (gogoproto.goproto_stringer) = false;

  // The bound name
  string name = 1;
  // The address the name resolves to.
  string address = 2;
  // Whether owner signature is required to add sub-names.
  bool restricted = 3;
}
```

## Normalization

Name records are normalized before being processed for creation or query.  Each component of the name must conform to a standard set of rules.  The sha256 of the normalized value is used internally for comparision purposes.

1. Names are always stored and compared using a lower case form or a hash derived from this normalized form.
2. Unicode values that are not graphic, lower case, or digits are considered invalid.
3. A single occurance of the hyphen-minus character is allowed unless the value conforms to a valid uuid.
```value: -
HYPHEN-MINUS
Unicode: U+002D, UTF-8: 2D
```
4. Each component of the name is subject to length restrictions for minimum and maxium length.  These limits are configurable in the module [parameters](./05_params.md)
5. A maximum number of components for a name (levels in the heirarchy) is also enforced.
6. Leading and trailing spaces are always trimmed off of names for consistency during processing and evaluation.