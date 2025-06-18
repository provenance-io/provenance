# CLI Examples

The `x/asset` module provides command-line interface (CLI) commands for creating and querying assets, asset classes, pools, tokenizations, and securitizations.

<!-- TOC -->
  - [Transaction Commands](#transaction-commands)
    - [Create Asset Class](#create-asset-class)
    - [Create Asset](#create-asset)
    - [Create Pool](#create-pool)
    - [Create Tokenization](#create-tokenization)
    - [Create Securitization](#create-securitization)
  - [Query Commands](#query-commands)
    - [List Asset Classes](#list-asset-classes)
    - [Get Asset Class](#get-asset-class)
    - [List Assets](#list-assets)
  - [Common Flags](#common-flags)

## Transaction Commands

### Create Asset Class

Creates a new asset class that defines the schema and classification for digital assets.

```bash
provenanced tx asset create-class [id] [name] [symbol] [description] [uri] [uri-hash] [data] [ledger-class-id]
```

**Arguments:**
- `id`: Unique identifier for the asset class
- `name`: Human-readable name for the asset class
- `symbol`: Abbreviated name for the asset class
- `description`: Brief description of the asset class
- `uri`: Link to off-chain metadata
- `uri-hash`: Hash of the metadata document
- `data`: JSON schema for asset data validation
- `ledger-class-id`: Associated ledger class ID

**Example:**
```bash
provenanced tx asset create-class \
  "real-estate" \
  "Real Estate Assets" \
  "REAL" \
  "Real estate properties" \
  "https://example.com/class-metadata.json" \
  "def456" \
  '{"type":"object","properties":{"location":{"type":"string"},"value":{"type":"number"}}}' \
  "ledger-class-001" \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3
```

### Create Asset

Creates a new digital asset within an existing asset class.

```bash
provenanced tx asset create-asset [class-id] [id] [uri] [uri-hash] [data]
```

**Arguments:**
- `class-id`: The asset class ID this asset belongs to
- `id`: Unique identifier for the asset
- `uri`: Link to off-chain metadata
- `uri-hash`: Hash of the metadata document
- `data`: JSON data conforming to the asset class schema

**Example:**
```bash
provenanced tx asset create-asset \
  "real-estate" \
  "property-001" \
  "https://example.com/metadata/property-001.json" \
  "abc123" \
  '{"location": "New York", "value": 500000}' \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3
```

### Create Pool

Creates a pool of NFTs represented by a marker token.

```bash
provenanced tx asset create-pool [pool] [nfts]
```

**Arguments:**
- `pool`: The pool marker coin definition (e.g., "10pooltoken")
- `nfts`: Semicolon-separated list of asset entries, where each entry is a comma-separated class-id and asset-id

**Example:**
```bash
# Create a pool with multiple assets
provenanced tx asset create-pool \
  10pooltoken \
  "real-estate,property-001;real-estate,property-002" \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3

# Create a pool with a single asset
provenanced tx asset create-pool \
  1000pooltoken \
  "real-estate,property-001" \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3
```

### Create Tokenization

Creates a tokenization marker representing fractional ownership of an individual NFT.

```bash
provenanced tx asset create-tokenization [amount] [nft-class-id] [nft-id]
```

**Arguments:**
- `amount`: The tokenization marker denomination (e.g., "1000pooltoken")
- `nft-class-id`: The asset class ID of the NFT to tokenize
- `nft-id`: The asset ID of the NFT to tokenize

**Example:**
```bash
provenanced tx asset create-tokenization \
  1000pooltoken \
  real-estate \
  property-001 \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3
```

### Create Securitization

Creates a securitization with multiple pools and tranches.

```bash
provenanced tx asset create-securitization [id] [pools] [tranches]
```

**Arguments:**
- `id`: Unique identifier for the securitization
- `pools`: Comma-separated list of pool names
- `tranches`: Comma-separated list of coins representing different tranches

**Example:**
```bash
provenanced tx asset create-securitization \
  "mortgage-sec-001" \
  "mortgage-pool-1,mortgage-pool-2" \
  "1000000senior-tranche,500000mezzanine-tranche" \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3

# Simple example
provenanced tx asset create-securitization \
  sec1 \
  "pool1,pool2" \
  "100tranche1,200tranche2" \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3
```

## Query Commands

### List Asset Classes

Retrieves all asset classes in the system.

```bash
provenanced query asset list-classes
```

**Example:**
```bash
provenanced query asset list-classes \
  --node=http://localhost:26657 \
  --output=json
```

**Sample Output:**
```json
{
  "assetClasses": [
    {
      "id": "real-estate",
      "name": "Real Estate Assets",
      "symbol": "REAL",
      "description": "Real estate properties",
      "uri": "https://example.com/class-metadata.json",
      "uri_hash": "def456",
      "data": "{\"type\":\"object\",\"properties\":{\"location\":{\"type\":\"string\"},\"value\":{\"type\":\"number\"}}}"
    }
  ]
}
```

### Get Asset Class

Retrieves a specific asset class by its ID.

```bash
provenanced query asset get-class [id]
```

**Arguments:**
- `id`: The asset class ID to retrieve

**Example:**
```bash
provenanced query asset get-class real-estate \
  --node=http://localhost:26657 \
  --output=json
```

**Sample Output:**
```json
{
  "assetClass": {
    "id": "real-estate",
    "name": "Real Estate Assets",
    "symbol": "REAL",
    "description": "Real estate properties",
    "uri": "https://example.com/class-metadata.json",
    "uri_hash": "def456",
    "data": "{\"type\":\"object\",\"properties\":{\"location\":{\"type\":\"string\"},\"value\":{\"type\":\"number\"}}}"
  }
}
```

### List Assets

Retrieves all assets owned by a specific address.

```bash
provenanced query asset list-assets [address]
```

**Arguments:**
- `address`: The bech32 address to query assets for

**Example:**
```bash
provenanced query asset list-assets pb1v9jxgun9wde476twta6xse2lv4mx2mn56s5hm4 \
  --node=http://localhost:26657 \
  --output=json
```

**Sample Output:**
```json
{
  "assets": [
    {
      "class_id": "real-estate",
      "id": "property-001",
      "uri": "https://example.com/metadata/property-001.json",
      "uri_hash": "abc123",
      "data": "{\"location\": \"New York\", \"value\": 500000}"
    }
  ]
}
```

## Common Flags

### Transaction Flags

- `--from`: Key name or address of the transaction signer
- `--chain-id`: Chain ID of the network
- `--gas`: Gas limit for the transaction
- `--gas-adjustment`: Gas adjustment factor
- `--gas-prices`: Gas prices in decimal format
- `--fees`: Fees to pay for the transaction
- `--broadcast-mode`: Transaction broadcasting mode (sync, async, block)

### Query Flags

- `--node`: Node to connect to (default: tcp://localhost:26657)
- `--output`: Output format (text, json, indent)
- `--height`: Use a specific height to query state at (this can error if the node is pruning state)

## Complete Workflow Example

Here's a complete example of creating an asset class, assets, and a pool:

```bash
# 1. Create an asset class
provenanced tx asset create-class \
  "real-estate" \
  "Real Estate Assets" \
  "REAL" \
  "Real estate properties" \
  "https://example.com/class-metadata.json" \
  "def456" \
  '{"type":"object","properties":{"location":{"type":"string"},"value":{"type":"number"}}}' \
  "ledger-class-001" \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3

# 2. Create assets
provenanced tx asset create-asset \
  "real-estate" \
  "property-001" \
  "https://example.com/metadata/property-001.json" \
  "abc123" \
  '{"location": "New York", "value": 500000}' \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3

provenanced tx asset create-asset \
  "real-estate" \
  "property-002" \
  "https://example.com/metadata/property-002.json" \
  "def456" \
  '{"location": "Los Angeles", "value": 750000}' \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3

# 3. Create a pool with the assets
provenanced tx asset create-pool \
  10pooltoken \
  "real-estate,property-001;real-estate,property-002" \
  --from=alice \
  --chain-id=testing \
  --gas=auto \
  --gas-adjustment=1.3

# 4. Query the results
provenanced query asset list-classes --node=http://localhost:26657
provenanced query asset list-assets pb1v9jxgun9wde476twta6xse2lv4mx2mn56s5hm4 --node=http://localhost:26657
```

This workflow demonstrates the typical usage pattern for the asset module, from creating asset classes and individual assets to bundling them into pools for trading and investment purposes. 