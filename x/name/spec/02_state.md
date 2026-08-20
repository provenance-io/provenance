
# State

The name module holds a very simple state collection.  


## Name Record KV Values
Name records are stored using a key made up of the name record type byte, `0x03`, followed by a SHA-256 hash of the
normalized name.  Hashing the whole name, dots included, keeps each name in its own record: names made up of the same
letters split into different segments, e.g. `bar.foo` and `foo` with the segment `bar`, get different keys.

```
Name: foo
key = 03 || 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae

Name: foo.bar
key = 03 || 2595d08ad22c733f7a1ce713e767563e13a8dfa35baa74919c28e0f586cb424b

Name:  foo.bar.baz
key = 03 || 31f52e7760a01e3ff5b6b395411d1dbace56483562480570bec00e03237b9373

```

## Address Record KV Index
In addition to the records stored by name an address cache is maintained for the addresses associated with each name
record.  This allows simple and fast reverse lookup queries to be performed.  Those keys are made up of the address
index type byte, `0x05`, the length of the address, the address itself, and then the name record key.

```
Address: pb1tg3ktger9ttlscehl3r5j4pqw7qzmvs4qr9vpm
key = 05 || 14 || 5A2365A3232AD7F86337FC4749542077802DB215 || 03 || 2595d08ad22c733f7a1ce713e767563e13a8dfa35baa74919c28e0f586cb424b
value = foo.bar
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