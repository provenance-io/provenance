
# State

The name module holds a very simple state collection.  

## Name Record KV Values

Name records are stored using a key made from the normalized name with its segments in reverse order.
This keeps a name and all the names under it next to each other in state (they share a key prefix),
so all the names in the tree under a given name can be quickly queried and iterated over.

```
Name: foo
key = 0x07 | "foo"

Name: bar.foo
key = 0x07 | "foo.bar"

Name: baz.bar.foo
key = 0x07 | "foo.bar.baz"
```

## Address Record KV Index

In addition to the records stored by name, an index is maintained on the addresses associated with each name
record. This allows simple and fast reverse lookup queries to be performed.

```

Address: pb1tg3ktger9ttlscehl3r5j4pqw7qzmvs4qr9vpm
key = 0x08 | len(address) | address | "baz.bar.foo"
value = (empty)

```

## Name Record

Name records are encoded using the following protobuf type

```

// NameRecord is a structure used to bind ownership of a name heirarchy to a collection of addresses
message NameRecord {
  option (gogoproto.goproto_stringer) = false;

  // The bound name
  string name = 1;
  // The address the name resolved to.
  string address = 2;
  // Whether owner signature is required to add sub-names.
  bool restricted = 3;
}

```
