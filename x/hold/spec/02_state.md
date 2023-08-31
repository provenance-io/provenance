# State

The `x/hold` module uses key/value pairs to store hold-related data in state.

## Holds

Holds on funds are recorded by address and denom using the following record format:

```
0x00 | len(<address>) | <address> | <denom> -> <amount>
```

Where:

* `0x00` is the type byte, and has a value of `0` for these records.
* `len(<address>)` is a single byte containing the length of the `<address>` as an 8-bit byte in big-endian order.
* `<address>` is the raw bytes of the address of the account that the funds are in.
* `<denom>` is the denomination string of the coin being held.
* `<amount>` is a string representation of the numerical amount being held.

Records are created, increased and decreased as needed.
If the `<amount>` is reduced to zero, the record is deleted.
