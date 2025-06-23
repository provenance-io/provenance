# State

The `x/asset` module leverages the Cosmos SDK's NFT module for state storage, with additional integration points to other Provenance modules.

<!-- TOC -->
  - [Asset Classes](#asset-classes)
  - [Assets](#assets)
  - [Pools](#pools)
  - [Tokenizations](#tokenizations)
  - [Securitizations](#securitizations)

## Asset Classes

Asset classes are stored as NFT classes in the underlying NFT module. The storage format follows the NFT module's standard:

```
NFT Class Key: 0x01 | <class_id> -> <nft_class_data>
```

Where:
- `0x01` is the NFT module's class prefix
- `<class_id>` is the unique identifier for the asset class
- `<nft_class_data>` is the protobuf-serialized NFT class containing:
  - `id`: The asset class ID
  - `name`: Human-readable name
  - `symbol`: Abbreviated name
  - `description`: Brief description
  - `uri`: Link to off-chain metadata
  - `uri_hash`: Hash of the metadata document
  - `data`: JSON schema as an Any type

Each asset class must be associated with a ledger class that exists in the `x/ledger` module.

## Assets

Individual assets are stored as NFTs in the underlying NFT module:

```
NFT Key: 0x02 | <class_id> | <asset_id> -> <nft_data>
```

Where:
- `0x02` is the NFT module's NFT prefix
- `<class_id>` is the asset class identifier
- `<asset_id>` is the unique asset identifier within the class
- `<nft_data>` is the protobuf-serialized NFT containing:
  - `class_id`: References the asset class
  - `id`: The asset identifier
  - `uri`: Link to off-chain metadata
  - `uri_hash`: Hash of the metadata document
  - `data`: JSON data as an Any type

Assets are owned by their creators, not by the asset module account.

## Pools

Pools are represented by marker accounts in the `x/marker` module. The pool creation process:

1. Creates a marker account with denom `pool.{denom}`
2. Transfers the specified NFTs to the marker account address
3. The marker account becomes the owner of the underlying assets

Pool state is stored in the marker module's state structure, with the marker account holding the NFTs that comprise the pool.

## Tokenizations

Tokenizations are represented by marker accounts in the `x/marker` module. When a tokenization is created:

1. A marker account is created with the specified denomination
2. The underlying NFT is transferred to the tokenization marker address
3. The marker represents fractional ownership of the underlying asset

Tokenization state follows the marker module's storage format.

## Securitizations

Securitizations are structured financial products that reference pools and define tranches. The securitization creation process:

1. Creates a main securitization marker with denom `sec.{id}`
2. Creates individual tranche markers with denoms `sec.{id}.tranche.{denom}`
3. Transfers the pools to the asset module account to prevent further transfers
4. Updates pool marker permissions to grant the asset module account control

Securitization state includes:
- **Main Marker**: The securitization marker account
- **Tranche Markers**: Individual tranche marker accounts
- **Pool Control**: The asset module account controls the underlying pools

## Integration State

The asset module creates additional state entries in other modules:

### Ledger Module
When an asset is created, a ledger entry is created:
```
Ledger Key: <asset_class_id> | <nft_id> -> <ledger_data>
```

### Registry Module
When an asset is created, a default registry entry is created:
```
Registry Key: <asset_class_id> | <nft_id> -> <registry_data>
```

### Marker Module
Pools, tokenizations, and securitizations create marker accounts:
```
Marker Key: <marker_address> -> <marker_data>
```

This multi-module state design allows the asset module to leverage the proven NFT infrastructure while adding specialized functionality for financial applications. 