# Asset State

The `x/asset` module leverages the Cosmos SDK's NFT module for state storage, with additional integration points to other Provenance modules.

The asset module itself does not maintain any state. All asset and asset class data is stored in the underlying NFT module.

---
<!-- TOC -->
  - [Asset Classes](#asset-classes)
  - [Assets](#assets)
  - [Pools](#pools)
  - [Tokenizations](#tokenizations)
  - [Securitizations](#securitizations)
  - [Integration State](#integration-state)


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

Assets are owned by their creators (or specified owner), not by the asset module account.


## Pools

Pools are represented by marker accounts in the `x/marker` module. The pool creation process:

1. Creates a marker account with the specified denom.
2. Transfers the specified NFTs to the marker account address.
3. The marker account becomes the owner of the underlying assets.

Pool state is stored in the marker module's state structure, with the marker account holding the NFTs that comprise the pool.

The pool marker is activated and finalized, making it immediately available for use.


## Tokenizations

Tokenizations are represented by marker accounts in the `x/marker` module. When a tokenization is created:

1. A marker account is created with the specified denomination and amount.
2. The underlying NFT is transferred to the tokenization marker address.
3. The marker represents fractional ownership of the underlying asset.

Tokenization state follows the marker module's storage format. The tokenization marker is activated and finalized.


## Securitizations

Securitizations are structured financial products that reference pools and define tranches. The securitization creation process:

1. Creates a main securitization marker with denom equal to the securitization ID (amount is 0).
2. Creates individual tranche markers with denoms `{id}.tranche.{tranche_denom}` and the specified amounts.
3. Revokes all existing access permissions from the pool markers.
4. Grants full administrative access to the asset module account for each pool marker.
5. Saves the updated pool markers with new permissions.

Securitization state includes:
- **Main Marker**: The securitization marker account (denom = securitization ID, amount = 0)
- **Tranche Markers**: Individual tranche marker accounts with specified denoms and amounts
- **Pool Control**: The asset module account has exclusive control over the underlying pool markers


## Integration State

The asset module creates additional state entries in other modules:

### Registry Module

When an asset is created, a default registry entry is created:
```
Registry Key: <asset_class_id> | <nft_id> -> <registry_data>
```

The default registry entry provides role-based access control for the asset. It is preserved even when an asset is burned, maintaining historical records.

### Marker Module

Pools, tokenizations, and securitizations create marker accounts:
```
Marker Key: <marker_address> -> <marker_data>
```

Each marker account stores:
- The marker denomination.
- The marker type (COIN).
- The marker status (ACTIVE).
- The marker supply and permissions.
- Access grants for managing the marker.

This multi-module state design allows the asset module to leverage the proven NFT infrastructure while adding specialized functionality for financial applications.
