# Messages

The `x/asset` module provides several message types for creating and managing digital assets, pools, tokenizations, and securitizations.

<!-- TOC -->
  - [MsgCreateAssetClass](#msgcreateassetclass)
  - [MsgCreateAsset](#msgcreateasset)
  - [MsgCreatePool](#msgcreatepool)
  - [MsgCreateTokenization](#msgcreatetokenization)
  - [MsgCreateSecuritization](#msgcreatesecuritization)

## MsgCreateAssetClass

Creates a new asset class that defines the classification and schema for digital assets.

**Signer**: `from_address`

| Field | Type | Description |
|-------|------|-------------|
| asset_class | AssetClass | The asset class definition |
| ledger_class | string | The associated ledger class ID |
| from_address | string | The address creating the asset class |

### AssetClass Fields

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier for the asset class |
| name | string | Human-readable name for the asset class |
| symbol | string | Abbreviated name for the asset class |
| description | string | Brief description of the asset class |
| uri | string | Link to off-chain metadata |
| uri_hash | string | Hash of the metadata document |
| data | string | JSON schema for asset data validation |

### Validation Rules

- `asset_class` cannot be nil
- `asset_class.id` cannot be empty
- `asset_class.name` cannot be empty
- `asset_class.data` must be valid JSON schema if provided
- `ledger_class` cannot be empty and must exist in the ledger module
- The ledger class's asset class ID must match the asset class ID
- `from_address` must be a valid bech32 address

### Response

```protobuf
message MsgCreateAssetClassResponse {}
```

## MsgCreateAsset

Creates a new digital asset within an existing asset class.

**Signer**: `from_address`

| Field | Type | Description |
|-------|------|-------------|
| asset | Asset | The asset definition |
| from_address | string | The address creating the asset |

### Asset Fields

| Field | Type | Description |
|-------|------|-------------|
| class_id | string | The asset class ID this asset belongs to |
| id | string | Unique identifier for the asset |
| uri | string | Link to off-chain metadata |
| uri_hash | string | Hash of the metadata document |
| data | string | JSON data conforming to the asset class schema |

### Validation Rules

- `asset` cannot be nil
- `asset.class_id` cannot be empty
- `asset.id` cannot be empty
- `asset.data` must be valid JSON if provided
- `asset.data` must conform to the asset class schema
- `from_address` must be a valid bech32 address
- The asset class must exist

### Response

```protobuf
message MsgCreateAssetResponse {}
```

## MsgCreatePool

Creates a pool of NFTs represented by a marker token.

**Signer**: `from_address`

| Field | Type | Description |
|-------|------|-------------|
| pool | Coin | The pool marker coin definition |
| nfts | []Nft | List of NFTs to include in the pool |
| from_address | string | The address creating the pool |

### Nft Fields

| Field | Type | Description |
|-------|------|-------------|
| class_id | string | The asset class ID |
| id | string | The asset ID |

### Validation Rules

- `pool` cannot be nil and must be valid
- `nfts` cannot be empty
- All NFTs in `nfts` must have valid `class_id` and `id`
- The `from_address` must own all specified NFTs
- `from_address` must be a valid bech32 address

### Response

```protobuf
message MsgCreatePoolResponse {}
```

## MsgCreateTokenization

Creates a tokenization marker representing fractional ownership of an individual NFT.

**Signer**: `from_address`

| Field | Type | Description |
|-------|------|-------------|
| denom | Coin | The tokenization marker denomination |
| nft | Nft | The NFT to tokenize |
| from_address | string | The address creating the tokenization |

### Validation Rules

- `denom` cannot be nil and must be valid
- `nft` cannot be nil
- `nft.class_id` cannot be empty
- `nft.id` cannot be empty
- The `from_address` must own the specified NFT
- `from_address` must be a valid bech32 address

### Response

```protobuf
message MsgCreateTokenizationResponse {}
```

## MsgCreateSecuritization

Creates a securitization with multiple pools and tranches.

**Signer**: `from_address`

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier for the securitization |
| pools | []string | List of pool IDs included in the securitization |
| tranches | []Coin | Different classes of securities with varying characteristics |
| from_address | string | The address creating the securitization |

### Validation Rules

- `id` cannot be empty
- `pools` cannot be empty
- All pool IDs in `pools` must be non-empty
- `tranches` cannot be empty
- All tranches must be valid coins
- `from_address` must be a valid bech32 address

### Response

```protobuf
message MsgCreateSecuritizationResponse {}
```

## Message Flow

1. **Asset Class Creation**: First, create an asset class to define the schema and classification
2. **Asset Creation**: Create individual assets within the asset class
3. **Pool Creation**: Optionally create pools by bundling multiple assets
4. **Tokenization Creation**: Create tokenization markers for fractional ownership of individual assets
5. **Securitization Creation**: Create structured financial products with multiple tranches

Each message type includes proper validation and integration with other Provenance modules to ensure data consistency and proper asset lifecycle management. 