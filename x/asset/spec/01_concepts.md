# Asset Concepts

The `x/asset` module provides a comprehensive digital asset management system on the Provenance blockchain, leveraging the Cosmos SDK's NFT module as the underlying infrastructure.

---
<!-- TOC 2 2 -->
  - [Asset Classes](#asset-classes)
  - [Assets](#assets)
  - [Pools](#pools)
  - [Tokenizations](#tokenizations)
  - [Securitizations](#securitizations)
  - [Integration with Other Modules](#integration-with-other-modules)


## Asset Classes

Asset Classes define the classification and schema for digital assets. Each asset class has:

- **ID**: A unique identifier (similar to ERC721 contract address).
- **Name**: Human-readable name for the asset class.
- **Symbol**: Abbreviated name for the asset class.
- **Description**: Brief description of the asset class.
- **URI**: Link to off-chain metadata that can define schema for the class and asset data attributes.
- **URI Hash**: Hash of the document pointed by URI.
- **Data**: JSON schema that defines the structure for asset data within this class.

Asset classes are stored as NFT classes in the underlying NFT module, providing a standardized way to categorize and validate assets.

See also: [MsgCreateAssetClass](03_messages.md#createassetclass).


## Assets

Assets are individual digital assets that belong to an asset class. Each asset has:

- **Class ID**: References the asset class it belongs to.
- **ID**: Unique identifier within the asset class.
- **URI**: Link to off-chain metadata for the specific asset.
- **URI Hash**: Hash of the document pointed by URI.
- **Data**: JSON data that conforms to the asset class schema.

Assets are stored as NFTs in the underlying NFT module, with the creator as the owner. When an asset is created:
* An NFT is minted with the specified owner.
* A default registry entry is created in the `x/registry` module for role-based access control.

See also: [MsgCreateAsset](03_messages.md#createasset).


## Pools

Pools are collections of NFTs that are grouped together and represented by a marker token. When a pool is created:

1. A marker account is created with the specified denom.
2. The specified NFTs are transferred to the pool marker address.
3. The pool marker represents ownership of the underlying assets.

Pools enable the bundling of multiple assets into a single tradeable unit, useful for creating diversified investment vehicles or asset-backed tokens.

See also: [MsgCreatePool](03_messages.md#createpool).


## Tokenizations

Tokenizations represent the fractionalization of individual NFTs into tradeable tokens. When a tokenization is created:

1. A marker account is created with the specified denomination and amount.
2. The underlying NFT is transferred to the tokenization marker address.
3. The marker tokens represent fractional ownership of the underlying asset.

Tokenizations allow for the fractionalization of individual assets, enabling smaller investors to participate in larger asset ownership.

See also: [MsgCreateTokenization](03_messages.md#createtokenization).


## Securitizations

Securitizations are structured financial products that pool multiple assets and issue different tranches of securities. Each securitization has:

- **ID**: Unique identifier for the securitization.
- **Pools**: List of pool denoms that are included in the securitization.
- **Tranches**: Different classes of securities with varying risk/return profiles.

When a securitization is created:
1. A main securitization marker is created with denom equal to the securitization `id`.
2. Individual tranche markers are created with denoms `{id}.tranche.{tranche_denom}`.
3. All access permissions are revoked from the pool markers.
4. The asset module account is granted full administrative access to the pool markers.
5. The tranche markers represent different risk/return profiles of the underlying assets.

Securitizations enable the creation of complex financial instruments with different risk profiles and return characteristics.

See also: [MsgCreateSecuritization](03_messages.md#createsecuritization).


## Integration with Other Modules

The Asset module integrates with several other Provenance modules:

- **NFT Module**: Uses the Cosmos SDK NFT module as the underlying storage and transfer mechanism for asset classes and assets.
- **Registry Module**: Creates default registry entries for assets to enable role-based access control and permissions management.
- **Marker Module**: Creates marker tokens for pools, tokenizations, and securitizations to represent bundled or fractionalized assets.
- **Exchange Module**: Assets can be traded on the exchange module using their NFT representation.

This integration provides a comprehensive ecosystem for digital asset management, trading, and financial product creation.
