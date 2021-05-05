# Events

The marker module emits the following events:

<!-- TOC 2 2 -->
  - [Marker Added](#marker-added)
  - [Marker Updated](#marker-updated)
  - [Grant Access](#grant-access)
  - [Revoke Access](#revoke-access)
  - [Finalize](#finalize)
  - [Activate](#activate)
  - [Cancel](#cancel)
  - [Destroy](#destroy)
  - [Mint](#mint)
  - [Burn](#burn)
  - [Withdraw](#withdraw)
  - [Transfer](#transfer)
  - [Withdraw Asset](#withdraw-asset)



---
## Marker Added

Fires when a marker is added using the Add Marker Msg.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_added           | denom                 | {denom string}            |
| marker_added           | amount                | {supply amount}           |
| marker_added           | administrator         | {admin account address}   |
| marker_added           | status                | {current marker status}   |
| marker_added           | type                  | {type of marker}          |


---
## Marker Updated

Fires when a marker is updated (ie SetMarkerDenomMetadata)

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_updated         | denom                 | {denom string}            |
| marker_updated         | administrator         | {admin account address}   |


---
## Grant Access

Fires when administrative access is granted for a marker

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_AccessGranted   | denom                 | {denom string}            |
| marker_AccessGranted   | administrator         | {admin account address}   |
| marker_AccessGranted   | marker_AccessGrant    | {access grant format}     |

### Access Grant Format

| Attribute Key         | Attribute Value          |
| --------------------- | ------------------------ |
| address               | {bech32 address string}  |
| permissions           | {csv list of role names} |


---
## Revoke Access

Fires when all access grants are removed for a given address.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_access_revoked  | denom                 | {denom string}            |
| marker_access_revoked  | administrator         | {admin account address}   |
| marker_access_revoked  | marker_access_revoked | {address removed}         |



---
## Finalize

Fires when a marker is finalized.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_finalized       | denom                 | {denom string}            |
| marker_finalized       | administrator         | {admin account address}   |



---
## Activate

Fires when a marker is activated.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_activated       | denom                 | {denom string}            |
| marker_activated       | administrator         | {admin account address}   |


---
## Cancel

Fired when a marker is cancelled successfully.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_cancelled       | denom                 | {denom string}            |
| marker_cancelled       | administrator         | {admin account address}   |


---
## Destroy

Fires when a marker is marked as destroyed and ready for removal.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_destroyed       | denom                 | {denom string}            |
| marker_destroyed       | administrator         | {admin account address}   |

---
## Mint

Fires when coins are minted for a marker.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_minted_coins    | denom                 | {denom string}            |
| marker_minted_coins    | amount                | {supply amount}           |
| marker_minted_coins    | administrator         | {admin account address}   |



---
## Burn

Fires when coins are burned from a marker account.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_burned_coins    | denom                 | {denom string}            |
| marker_burned_coins    | amount                | {supply amount}           |
| marker_burned_coins    | administrator         | {admin account address}   |


---

Fires when coin is removed from a marker account and transferred to another.
## Withdraw

| Type                   | Attribute Key         | Attribute Value             |
| ---------------------- | --------------------- | --------------------------- |
| marker_withdraw_coins  | denom                 | {denom string}              |
| marker_withdraw_coins  | amount                | {supply amount}             |
| marker_withdraw_coins  | administrator         | {admin account address}     |
| marker_withdraw_coins  | toAddress             | {recipient account address} |


---
## Transfer

Fires when a facilitated transfer is performed of the marker's coin between accounts by an administrator

| Type                   | Attribute Key         | Attribute Value             |
| ---------------------- | --------------------- | --------------------------- |
| marker_transfer_coin   | denom                 | {denom string}              |
| marker_transfer_coin   | amount                | {supply amount}             |
| marker_transfer_coin   | administrator         | {admin account address}     |
| marker_transfer_coin   | fromAddress           | {source account address}    |
| marker_transfer_coin   | toAddress             | {recipient account address} |
---
