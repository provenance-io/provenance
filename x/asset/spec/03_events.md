# Events

The `x/asset` module emits events through the underlying NFT module and custom events for asset-specific operations.

<!-- TOC -->
  - [NFT Events](#nft-events)
  - [Asset-Specific Events](#asset-specific-events)

## NFT Events

Since the asset module uses the Cosmos SDK NFT module as its underlying infrastructure, it inherits and emits all standard NFT events:

### EventClassCreated

Emitted when an asset class is created.

`@Type`: `cosmos.nft.v1beta1.EventClassCreated`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| class_id      | string of the created asset class ID |
| name          | string of the asset class name |
| symbol        | string of the asset class symbol |
| description   | string of the asset class description |
| uri           | string of the asset class URI |
| uri_hash      | string of the asset class URI hash |

### EventNFTMinted

Emitted when an asset is created (minted as an NFT).

`@Type`: `cosmos.nft.v1beta1.EventNFTMinted`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| class_id      | string of the asset class ID |
| id            | string of the asset ID |
| uri           | string of the asset URI |
| uri_hash      | string of the asset URI hash |
| owner         | string of the asset owner address |

### EventNFTTransferred

Emitted when an asset is transferred between addresses.

`@Type`: `cosmos.nft.v1beta1.EventNFTTransferred`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| class_id      | string of the asset class ID |
| id            | string of the asset ID |
| sender        | string of the sender address |
| receiver      | string of the receiver address |

### EventNFTBurned

Emitted when an asset is burned (destroyed).

`@Type`: `cosmos.nft.v1beta1.EventNFTBurned`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| class_id      | string of the asset class ID |
| id            | string of the asset ID |
| owner         | string of the asset owner address |

## Asset-Specific Events

The asset module emits additional events for asset-specific operations:

### EventAssetClassCreated

Emitted when an asset class is created with additional asset-specific metadata.

`@Type`: `asset_class_created`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| asset_class_id | string of the created asset class ID |
| asset_name    | string of the asset class name |
| asset_symbol  | string of the asset class symbol |
| ledger_class  | string of the associated ledger class ID |
| owner         | string of the creator address |

### EventAssetCreated

Emitted when an asset is created with additional asset-specific metadata.

`@Type`: `asset_created`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| asset_class_id | string of the asset class ID |
| asset_id      | string of the created asset ID |
| owner         | string of the creator address |

### EventPoolCreated

Emitted when a pool is created.

`@Type`: `pool_created`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| pool_denom    | string of the pool marker denomination |
| pool_amount   | string of the pool marker amount |
| nft_count     | string of the number of NFTs in the pool |
| owner         | string of the creator address |

### EventTokenizationCreated

Emitted when a tokenization is created.

`@Type`: `tokenization_created`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| tokenization_denom | string of the tokenization denomination |
| pool_amount   | string of the tokenization amount |
| nft_class_id  | string of the underlying NFT class ID |
| nft_id        | string of the underlying NFT ID |
| owner         | string of the creator address |

### EventSecuritizationCreated

Emitted when a securitization is created.

`@Type`: `securitization_created`

| Attribute Key | Attribute Value |
|---------------|-----------------|
| securitization_id | string of the securitization ID |
| tranche_count | string of the number of tranches |
| pool_count    | string of the number of pools in the securitization |
| owner         | string of the creator address |

## Example Event

Example of an asset creation event:

```json
{
  "type": "asset_created",
  "attributes": [
    {"key": "asset_class_id", "value": "\"real-estate-token\""},
    {"key": "asset_id", "value": "\"property-001\""},
    {"key": "owner", "value": "\"pb1v9jxgun9wde476twta6xse2lv4mx2mn56s5hm4\""}
  ]
}
```

The asset module leverages the robust event system of the NFT module while adding domain-specific events for financial applications. 