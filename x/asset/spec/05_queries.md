# Queries

The `x/asset` module provides queries for retrieving asset and asset class information.

<!-- TOC -->
  - [ListAssets](#listassets)
  - [ListAssetClasses](#listassetclasses)
  - [GetClass](#getclass)

## ListAssets

Retrieves all assets owned by a specific address.

**Endpoint**: `GET /provenance/asset/v1/asset`

### Request

| Field | Type | Description |
|-------|------|-------------|
| address | string | The bech32 address to query assets for |

### Response

```protobuf
message QueryListAssetsResponse {
  repeated Asset assets = 1;
}
```

### Asset Fields

| Field | Type | Description |
|-------|------|-------------|
| class_id | string | The asset class ID |
| id | string | The asset ID |
| uri | string | Link to off-chain metadata |
| uri_hash | string | Hash of the metadata document |
| data | string | JSON data for the asset |

### Validation

- The `address` must be a valid bech32 address
- If the address is invalid or missing, the query will fail
- If the address doesn't own any assets, an empty list will be returned

### Example

Request:
```bash
curl "http://localhost:1317/provenance/asset/v1/asset?address=pb1v9jxgun9wde476twta6xse2lv4mx2mn56s5hm4"
```

Response:
```json
{
  "assets": [
    {
      "class_id": "real-estate-token",
      "id": "property-001",
      "uri": "https://example.com/metadata/property-001.json",
      "uri_hash": "a1b2c3d4e5f6",
      "data": "{\"address\":\"123 Main St\",\"value\":500000}"
    }
  ]
}
```

## ListAssetClasses

Retrieves all asset classes in the system.

**Endpoint**: `GET /provenance/asset/v1/class`

### Request

No parameters required.

### Response

```protobuf
message QueryListAssetClassesResponse {
  repeated AssetClass assetClasses = 1;
}
```

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

### Example

Request:
```bash
curl "http://localhost:1317/provenance/asset/v1/class"
```

Response:
```json
{
  "assetClasses": [
    {
      "id": "real-estate-token",
      "name": "Real Estate Token",
      "symbol": "RET",
      "description": "Tokenized real estate properties",
      "uri": "https://example.com/schema/real-estate.json",
      "uri_hash": "schema-hash-123",
      "data": "{\"type\":\"object\",\"properties\":{\"address\":{\"type\":\"string\"},\"value\":{\"type\":\"number\"}}}"
    }
  ]
}
```

## GetClass

Retrieves a specific asset class by its ID.

**Endpoint**: `GET /provenance/asset/v1/class/{id}`

### Request

| Field | Type | Description |
|-------|------|-------------|
| id | string | The asset class ID to retrieve |

### Response

```protobuf
message QueryGetClassResponse {
  AssetClass assetClass = 1;
}
```

### Validation

- The `id` must be a non-empty string
- If the asset class doesn't exist, the query will fail with "class not found"
- If the `id` is invalid, the query will fail

### Example

Request:
```bash
curl "http://localhost:1317/provenance/asset/v1/class/real-estate-token"
```

Response:
```json
{
  "assetClass": {
    "id": "real-estate-token",
    "name": "Real Estate Token",
    "symbol": "RET",
    "description": "Tokenized real estate properties",
    "uri": "https://example.com/schema/real-estate.json",
    "uri_hash": "schema-hash-123",
    "data": "{\"type\":\"object\",\"properties\":{\"address\":{\"type\":\"string\"},\"value\":{\"type\":\"number\"}}}"
  }
}
```

## Query Implementation Details

The asset module queries leverage the underlying NFT module's storage:

- **ListAssets**: Queries all NFT classes, then filters NFTs by owner address
- **ListAssetClasses**: Queries all NFT classes and converts them to asset classes
- **GetClass**: Queries a specific NFT class by ID

The queries handle the conversion between NFT module data structures and asset module data structures, including:

- Converting `Any` types to JSON strings for the `data` field
- Mapping NFT class fields to asset class fields
- Filtering and transforming data as needed

This design allows the asset module to benefit from the NFT module's proven query infrastructure while providing domain-specific asset functionality. 