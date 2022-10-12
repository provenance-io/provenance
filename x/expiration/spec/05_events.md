# Expiration Events

The expiration module emits the following events.

<!-- TOC -->
  - [Generic](#generic)
    - [EventTxCompleted](#eventtxcompleted)
  - [Expiration](#expiration)
    - [EventExpirationAdd](#eventexpirationadd)
    - [EventExpirationDeposit](#eventexpirationdeposit)
    - [EventExpirationExtend](#eventexpirationextend)
    - [EventExpirationInvoke](#eventexpirationinvoke)

---
## Generic

### EventTxCompleted

This event is emitted when a TX has completed successfully.
It will usually be accompanied by one or more of the other events.

| Attribute Key         | Attribute Value                                   |
| --------------------- |---------------------------------------------------|
| Module                | "expiration"                                      |
| Endpoint              | The name of the rpc called, e.g. "AddExpiration"  |
| Signers               | List of bech32 address strings of the msg signers |


## Expiration

### EventExpirationAdd

This event is emitted when a new expiration is added.

| Attribute Key | Attribute Value                                      |
|---------------|------------------------------------------------------|
| ModuleAssetId | The ID of the module asset registered for expiration |

### EventExpirationDeposit

This event is emitted when a new expiration is added.
It signals that a deposit was collected.

| Attribute Key | Attribute Value                                         |
|---------------|---------------------------------------------------------|
| ModuleAssetId | The ID of the module asset registered for expiration    |
| Depositor     | The bech32 address from which the deposit was collected |
| Deposit       | The amount collected                                    |

### EventExpirationExtend

This event is emitted when an expiration is extended.

| Attribute Key | Attribute Value                                      |
|---------------|------------------------------------------------------|
| ModuleAssetId | The ID of the module asset registered for expiration |

### EventExpirationInvoke

This event is emitted when an expiration is invoked.

| Attribute Key | Attribute Value                                      |
|---------------|------------------------------------------------------|
| ModuleAssetId | The ID of the module asset registered for expiration |
