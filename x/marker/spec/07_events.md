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
  - [Set Net Asset Value](#set-net-asset-value)
  - [Marker Params Updated](#marker-params-updated)



---
## Marker Added

Fires when a marker is added using the Add Marker Msg.

Type: `provenance.marker.v1.EventMarkerAdd`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Denom         | \{denom string\}          |
| Address       | \{marker  address\}       |
| Amount        | \{supply amount\}         |
| Manager       | \{admin account address\} |
| Status        | \{current marker status\} |
| MarkerType    | \{type of marker\}        |

---
## Grant Access

Fires when administrative access is granted for a marker

Type: `provenance.marker.v1.EventMarkerAddAccess`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Denom         | \{denom string\}          |
| Administrator | \{admin account address\} |
| Access        | \{access grant format\}   |

### Access Grant Format

Type: `provenance.marker.v1.EventMarkerAccess`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Address       | \{bech32 address string\} |
| Permissions   | \{array of role names\}   |

---
## Revoke Access

Fires when all access grants are removed for a given address.

Type: `provenance.marker.v1.EventMarkerDeleteAccess`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Denom         | \{denom string\}          |
| Administrator | \{admin account address\} |
| RemoveAddress | \{address removed\}       |

---
## Finalize

Fires when a marker is finalized.

Type: `provenance.marker.v1.EventMarkerFinalize`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Denom         | \{denom string\}          |
| Administrator | \{admin account address\} |

---
## Activate

Fires when a marker is activated.

Type: `provenance.marker.v1.EventMarkerActivate`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Denom         | \{denom string\}          |
| Administrator | \{admin account address\} |

---
## Cancel

Fired when a marker is cancelled successfully.

Type: `provenance.marker.v1.EventMarkerCancel`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Denom         | \{denom string\}          |
| Administrator | \{admin account address\} |

---
## Destroy

Fires when a marker is marked as destroyed and ready for removal.

Type: `provenance.marker.v1.EventMarkerDelete`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Denom         | \{denom string\}          |
| Administrator | \{admin account address\} |

---
## Mint

Fires when coins are minted for a marker.

Type: `provenance.marker.v1.EventMarkerMint`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Denom         | \{denom string\}          |
| Amount        | \{supply amount\}         |
| Administrator | \{admin account address\} |

---
## Burn

Fires when coins are burned from a marker account.

Type: `provenance.marker.v1.EventMarkerBurn`

| Attribute Key | Attribute Value           |
|---------------|---------------------------|
| Denom         | \{denom string\}          |
| Amount        | \{supply amount\}         |
| Administrator | \{admin account address\} |

---
## Withdraw

Fires during a marker `Withdraw`.

Type: `provenance.marker.v1.EventMarkerWithdraw`

| Attribute Key | Attribute Value               |
|---------------|-------------------------------|
| Denom         | \{denom string\}              |
| Amount        | \{supply amount\}             |
| Administrator | \{admin account address\}     |
| ToAddress     | \{recipient account address\} |

---
## Transfer

Fires during a marker `Transfer`.

Type: `provenance.marker.v1.EventMarkerTransfer`

| Attribute Key | Attribute Value               |
|---------------|-------------------------------|
| Denom         | \{denom string\}              |
| Amount        | \{supply amount\}             |
| Administrator | \{admin account address\}     |
| FromAddress   | \{source account address\}    |
| ToAddress     | \{recipient account address\} |

---
## Set Denom Metadata

Fires when the denom metadata is set for a marker

Type: `provenance.marker.v1.EventMarkerSetDenomMetadata`

| Attribute Key       | Attribute Value           |
|---------------------|---------------------------|
| MetadataBase        | \{marker's denom string\} |
| MetadataDescription | \{description string\}    |
| MetadataDisplay     | \{denom string\}          |
| MetadataName        | \{name string\}           |
| MetadataSymbol      | \{symbol string\}         |
| MetadataDenomUnits  | \{array of  denom units\} |
| Administrator       | \{admin account address\} |

### Denom Unit Format

Denom units have a specified exponent (1-18), a specified denom, and a list of optional aliases.  Example
aliases for `uhash` might be `microhash` or `µhash`

Type: `provenance.marker.v1.EventDenomUnit`

| Attribute Key | Attribute Value            |
|---------------|----------------------------|
| Denom         | \{denom string\}           |
| Exponent      | \{uint\}                   |
| Aliases       | \{array of denom strings\} |

---
## Set Net Asset Value

Fires when a `NetAssetValue` is added or updated for a marker.

Type: `provenance.marker.v1.EventSetNetAssetValue`

| Attribute Key | Attribute Value                                     |
|---------------|-----------------------------------------------------|
| Denom         | \{marker's denom string\}                           |
| Price         | \{token amount the marker is valued at for volume\} |
| Volume        | \{total volume/shares associated with price\}       |
| Source        | \{source address of caller\}                        |

---
## Marker Params Updated

Fires when an `EventMarkerParamsUpdated` event occurs, indicating that the marker module's parameters have been updated via a governance proposal.

Type: `provenance.marker.v1.EventMarkerParamsUpdated`

| Attribute Key           | Attribute Value                                     |
|-------------------------|-----------------------------------------------------|
| EnableGovernance        | \{value for if governance control is enabled\}      |
| UnrestrictedDenomRegex  | \{regex for unrestricted denom validation\}         | 
| MaxSupply               | \{value for the max allowed supply\}                |
