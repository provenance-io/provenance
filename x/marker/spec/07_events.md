# Events

The marker module emits the following events:

<!-- TOC 2 2 -->
  - [Marker Added](#marker-added)
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
  - [Set Denom Metadata](#set-denom-metadata)



---
## Marker Added

Fires when a marker is added using the Add Marker Msg.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| EventMarkerAdd         | Denom                 | {denom string}            |
| EventMarkerAdd         | Amount                | {supply amount}           |
| EventMarkerAdd         | Manager               | {admin account address}   |
| EventMarkerAdd         | Status                | {current marker status}   |
| EventMarkerAdd         | MarkerType            | {type of marker}          |

`provenance.marker.v1.EventMarkerAdd`

---
## Grant Access

Fires when administrative access is granted for a marker

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| EventMarkerAddAccess   | Denom                 | {denom string}            |
| EventMarkerAddAccess   | Administrator         | {admin account address}   |
| EventMarkerAddAccess   | Access                | {access grant format}     |

`provenance.marker.v1.EventMarkerAddAccess`
### Access Grant Format

| Attribute Key         | Attribute Value          |
| --------------------- | ------------------------ |
| Address               | {bech32 address string}  |
| Permissions           | {array of role names}    |


---
## Revoke Access

Fires when all access grants are removed for a given address.

| Type                     | Attribute Key         | Attribute Value           |
| ------------------------ | --------------------- | ------------------------- |
| EventMarkerDeleteAccess  | Denom                 | {denom string}            |
| EventMarkerDeleteAccess  | Administrator         | {admin account address}   |
| EventMarkerDeleteAccess  | RemoveAddress         | {address removed}         |

`provenance.marker.v1.EventMarkerDeleteAccess`

---
## Finalize

Fires when a marker is finalized.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| EventMarkerFinalize    | Denom                 | {denom string}            |
| EventMarkerFinalize    | Administrator         | {admin account address}   |

`provenance.marker.v1.EventMarkerFinalize`

---
## Activate

Fires when a marker is activated.

| Type                      | Attribute Key         | Attribute Value           |
| ------------------------- | --------------------- | ------------------------- |
| EventMarkerActivate       | Denom                 | {denom string}            |
| EventMarkerActivate       | Administrator         | {admin account address}   |

`provenance.marker.v1.EventMarkerActivate`

---
## Cancel

Fired when a marker is cancelled successfully.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| EventMarkerCancel      | Denom                 | {denom string}            |
| EventMarkerCancel      | Administrator         | {admin account address}   |

`provenance.marker.v1.EventMarkerCancel`

---
## Destroy

Fires when a marker is marked as destroyed and ready for removal.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| EventMarkerDelete      | Denom                 | {denom string}            |
| EventMarkerDelete      | Administrator         | {admin account address}   |

`provenance.marker.v1.EventMarkerDelete`

---
## Mint

Fires when coins are minted for a marker.

| Type               | Attribute Key         | Attribute Value           |
| ------------------ | --------------------- | ------------------------- |
| EventMarkerMint    | Denom                 | {denom string}            |
| EventMarkerMint    | Amount                | {supply amount}           |
| EventMarkerMint    | Administrator         | {admin account address}   |

`provenance.marker.v1.EventMarkerMint`

---
## Burn

Fires when coins are burned from a marker account.

| Type               | Attribute Key         | Attribute Value           |
| ------------------ | --------------------- | ------------------------- |
| EventMarkerBurn    | Denom                 | {denom string}            |
| EventMarkerBurn    | Amount                | {supply amount}           |
| EventMarkerBurn    | Administrator         | {admin account address}   |

`provenance.marker.v1.EventMarkerBurn`

---

Fires when coin is removed from a marker account and transferred to another.
## Withdraw

| Type                 | Attribute Key         | Attribute Value             |
| -------------------- | --------------------- | --------------------------- |
| EventMarkerWithdraw  | Denom                 | {denom string}              |
| EventMarkerWithdraw  | Amount                | {supply amount}             |
| EventMarkerWithdraw  | Administrator         | {admin account address}     |
| EventMarkerWithdraw  | ToAddress             | {recipient account address} |

`provenance.marker.v1.EventMarkerWithdraw`

---
## Transfer

Fires when a facilitated transfer is performed of the marker's coin between accounts by an administrator

| Type                   | Attribute Key         | Attribute Value             |
| ---------------------- | --------------------- | --------------------------- |
| EventMarkerTransfer   | Denom                 | {denom string}               |
| EventMarkerTransfer   | Amount                | {supply amount}              |
| EventMarkerTransfer   | Administrator         | {admin account address}      |
| EventMarkerTransfer   | FromAddress           | {source account address}     |
| EventMarkerTransfer   | ToAddress             | {recipient account address}  |

`provenance.marker.v1.EventMarkerTransfer`

## Set Denom Metadata

Fires when the denom metadata is set for a marker

| Type                          | Attribute Key         | Attribute Value             |
| ----------------------------- | --------------------- | --------------------------- |
| EventMarkerSetDenomMetadata   | MetadataBase          | {marker's denom string}     |
| EventMarkerSetDenomMetadata   | MetadataDescription   | {description string}        |
| EventMarkerSetDenomMetadata   | MetadataDisplay       | {denom string}              |
| EventMarkerSetDenomMetadata   | MetadataName          | {name string}               |
| EventMarkerSetDenomMetadata   | MetadataSymbol        | {symbol string}             |
| EventMarkerSetDenomMetadata   | MetadataDenomUnits    | {array of  denom units}     |
| EventMarkerSetDenomMetadata   | Administrator         | {admin account address}     |

### Denom Unit Format

Denom units have a specified exponent (1-18), a specified denom, and a list of optional aliases.  Example
aliases for `uhash` might be `microhash` or `µhash`

| Attribute Key         | Attribute Value          |
| --------------------- | ------------------------ |
| Denom                 | {denom string}           |
| Exponent              | {uint}                   |
| Aliases               | {array of denom strings} |


`provenance.marker.v1.EventMarkerSetDenomMetadata`

---
