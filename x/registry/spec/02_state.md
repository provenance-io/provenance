# Registry State

The `x/registry` module uses key/value paris to store registry data in state.

---

## Registry Entry

Each registry entry is recorded by asset class id and NFT id.

```
0x01 | <asset class id> | 0x00 | <nft id> -> protobuf(RegistryEntry)
```

Where:
* `0x01` is the type byte, and has a value of `1` for these records.
* `<asset class id>` is a string containing the asset class identifier.
* `0x00` is a null byte separator.
* `<nft id>` is a string containing the NFT identifier.

Records are created, updated, and deleted as needed.

See also: [RegistryEntry](03_messages.md#registryentry)