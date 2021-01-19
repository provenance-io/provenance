
# State

The name module holds a very simple state collection.  


## Name Record KV Values
Name records are stored using a key based upon a concatenated
list of hashes based on each label within the name.  This approach allows all of the names in the tree under a given
name to be quickly queried and iterated over.

```
Name: foo
key = 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae

Name: foo.bar
key = 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae.fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9

Name:  foo.bar.baz
key = 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae.fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9.baa5a0964d3320fbc0c6a922140453c8513ea24ab8fd0577034804a967248096

```

## Address Record KV Index
In addition to the records stored by name an address cache is maintained for the addresses associated with each name
record.  This allows simple and fast reverse lookup queries to be performed.

```
Address: pb1tg3ktger9ttlscehl3r5j4pqw7qzmvs4qr9vpm
key = 5A2365A3232AD7F86337FC4749542077802DB215.2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae.fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9
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