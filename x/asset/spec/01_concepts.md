# Concepts

The `x/asset` module is designed to provide a comprehensive digital asset management system on the Provenance blockchain, leveraging the Cosmos SDK's NFT module as the underlying infrastructure.

<!-- TOC -->
  - [Asset Classes](#asset-classes)
  - [Assets](#assets)
  - [Pools](#pools)
  - [Tokenizations](#tokenizations)
  - [Securitizations](#securitizations)
  - [Integration with Other Modules](#integration-with-other-modules)

## Asset Classes

Asset Classes define the classification and schema for digital assets. Each asset class has:

- **ID**: A unique identifier (similar to ERC721 contract address)
- **Name**: Human-readable name for the asset class
- **Symbol**: Abbreviated name for the asset class
- **Description**: Brief description of the asset class
- **URI**: Link to off-chain metadata that can define schema for the class and asset data attributes
- **URI Hash**: Hash of the document pointed by URI
- **Data**: JSON schema that defines the structure for asset data within this class

Asset classes are stored as NFT classes in the underlying NFT module, providing a standardized way to categorize and validate assets. Each asset class must be associated with a ledger class that exists in the `x/ledger` module.

## Assets

Assets are individual digital assets that belong to an asset class. Each asset has:

- **Class ID**: References the asset class it belongs to
- **ID**: Unique identifier within the asset class
- **URI**: Link to off-chain metadata for the specific asset
- **URI Hash**: Hash of the document pointed by URI
- **Data**: JSON data that conforms to the asset class schema

Assets are stored as NFTs in the underlying NFT module, with the creator as the owner. When an asset is created, it also creates:
- A ledger entry in the `x/ledger` module for tracking asset lifecycle
- A default registry entry in the `x/registry` module for asset metadata

## Pools

Pools are collections of NFTs that are grouped together and represented by a marker token. When a pool is created:

1. A marker account is created with a denom in the format `pool.{denom}`
2. The specified NFTs are transferred to the pool marker address
3. The pool marker can be used to represent ownership of the underlying assets

Pools enable the bundling of multiple assets into a single tradeable unit, useful for creating diversified investment vehicles or asset-backed tokens.

## Tokenizations

Tokenizations represent the fractionalization of individual NFTs into tradeable tokens. When a tokenization is created:

1. A marker account is created with the specified denomination
2. The underlying NFT is transferred to the tokenization marker address
3. The marker tokens represent fractional ownership of the underlying asset

Tokenizations allow for the fractionalization of individual assets, enabling smaller investors to participate in larger asset ownership.

## Securitizations

Securitizations are structured financial products that pool multiple assets and issue different tranches of securities. Each securitization has:

- **ID**: Unique identifier for the securitization
- **Pools**: List of pool IDs that are included in the securitization
- **Tranches**: Different classes of securities with varying risk/return profiles

When a securitization is created:
1. A main securitization marker is created with denom `sec.{id}`
2. Individual tranche markers are created with denoms `sec.{id}.tranche.{denom}`
3. The pools are transferred to the asset module account to prevent further transfers
4. The tranche markers represent different risk/return profiles of the underlying assets

Securitizations enable the creation of complex financial instruments with different risk profiles and return characteristics.

## Integration with Other Modules

The Asset module integrates with several other Provenance modules:

- **NFT Module**: Uses the Cosmos SDK NFT module as the underlying storage and transfer mechanism
- **Ledger Module**: Creates ledger entries for tracking asset lifecycle and status
- **Registry Module**: Creates registry entries for asset metadata and provenance tracking
- **Marker Module**: Uses markers for creating pool, tokenization, and securitization tokens
- **Exchange Module**: Assets can be traded on the exchange module

This integration provides a comprehensive ecosystem for digital asset management, trading, and financial product creation. 