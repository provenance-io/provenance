# Events

The `x/hold` module emits the following events:

<!-- TOC -->
  - [EventHoldAdded](#eventholdadded)
  - [EventHoldReleased](#eventholdreleased)

## EventHoldAdded

This event is emitted when a hold is placed on some funds.

`@Type`: `provenance.hold.v1.EventHoldAdded`

| Attribute Key | Attribute Value                         |
|---------------|-----------------------------------------|
| address       | bech32 string of account with the funds |
| amount        | string of coins newly placed on hold    |
| reason        | human readable string                   |

All values are wrapped in double quotes.

Example:

```json
{
  "type": "provenance.hold.v1.EventHoldAdded",
  "attributes": [
    {"key": "address", "value": "\"pb1v9jxgun9wde476twta6xse2lv4mx2mn56s5hm4\""},
    {"key": "amount", "value": "\"1000000000nhash,5000musdf\""}
    {"key": "reason", "value": "\"order 66\""}
  ]
}
```

## EventHoldReleased

This event is emitted when some held funds are released.

`@Type`: `provenance.hold.v1.EventHoldReleased`

| Attribute Key | Attribute Value                         |
|---------------|-----------------------------------------|
| address       | bech32 string of account with the funds |
| amount        | string of the coins just released       |

Both values are wrapped in double quotes.

Example:

```json
{
  "type": "provenance.hold.v1.EventHoldReleased",
  "attributes": [
    {"key": "address", "value": "\"pb1v9jxgun9wde476twta6xse2lv4mx2mn56s5hm4\""},
    {"key": "amount", "value": "\"1000000000nhash,5000musdf\""}
  ]
}
```
