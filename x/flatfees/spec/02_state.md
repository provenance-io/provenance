# FlatFees State

The `x/flatfees` module uses key/value pairs to store flat-fee related data in state.

## MsgFees

The fee for each Msg type is recorded by type url.

```
0x00 | <msg type url> -> protobuf(MsgFee)
```

Where:

* `0x00` is the type byte, and has a value of `0` for these records.
* `<msg type url>` is a string containing the `MsgTypeURL` of a `Msg`.
* `protobuf(MsgFee)` is a protobuf encoded `MsgFee` entry.

Records are created, updated, and deleted as needed.

TODO: See also MsgFee.

## Params

The params for the flatfees module is stored in a single state entry.

```
0x01 -> protobuf(Params)
```

Where:

* `0x01` is the type byte, and has a value of `1` for the params record.
* `protobuf(Params)` is a protobuf encoded `Params` entry.

TODO: See also params.