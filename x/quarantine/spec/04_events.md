# Events

The `x/quarantine` module emits the following events:

## EventOptIn

This event is emitted when an account opts into quarantine.

`@Type`: `/provenance.quarantine.v1.EventOptIn`

| Attribute Key | Attribute Value                        |
|---------------|----------------------------------------|
| to_address    | \{bech32 string of account opting in\} |

## EventOptOut

This event is emitted when an account opts out of quarantine.

`@Type`: `/provenance.quarantine.v1.EventOptOut`

| Attribute Key | Attribute Value                         |
|---------------|-----------------------------------------|
| to_address    | \{bech32 string of account opting out\} |

## EventFundsQuarantined

When funds are quarantined, the `recipient` in events emitted by the `x/bank` module will be the quarantined funds holder account instead of the intended recipient.
The following event is also emitted.

`@Type`: `/provenance.quarantine.v1.EventFundsQuarantined`

| Attribute Key | Attribute Value                         |
|---------------|-----------------------------------------|
| to_address    | \{bech32 string of intended recipient\} |
| coins         | \{sdk.Coins of funds quarantined\}      |

## EventFundsReleased

This event is emitted when funds are fully accepted and sent from the quarantine funds holder to the originally intended recipient.

`@Type`: `/provenance.quarantine.v1.EventFundsReleased`

| Attribute Key | Attribute Value                 |
|---------------|---------------------------------|
| to_address    | \{bech32 string of recipient\}  |
| coins         | \{sdk.Coins of funds released\} |
