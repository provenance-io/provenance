# Asset Messages

The asset module has `Msg` endpoints for creating and managing assets, asset classes, pools, tokenizations, and securitizations.

---
<!-- TOC 2 2 -->
  - [BurnAsset](#burnasset)
  - [CreateAsset](#createasset)
  - [CreateAssetClass](#createassetclass)
  - [CreatePool](#createpool)
  - [CreateTokenization](#createtokenization)
  - [CreateSecuritization](#createsecuritization)


## BurnAsset

The `BurnAsset` endpoint removes an NFT from circulation by burning it. The asset registry entry is preserved for historical tracking purposes.

It is expected to fail if:
* The asset does not exist.
* The `signer` is not the owner of the asset.
* The `signer` is not a valid bech32 address.

### MsgBurnAsset

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L39-L48

#### AssetKey

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/asset.proto#L51-L57

### MsgBurnAssetResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L50-L51


## CreateAsset

The `CreateAsset` endpoint creates a new digital asset within an existing asset class. When an asset is created:
* An NFT is minted with the specified owner (or the signer if no owner is provided).
* A default registry entry is created in the `x/registry` module for role-based access control.

It is expected to fail if:
* The `asset` is nil.
* The `asset.class_id` is empty.
* The `asset.id` is empty.
* The asset class does not exist.
* The `asset.data` is not valid JSON.
* The `asset.data` does not conform to the asset class schema.
* The `signer` is not a valid bech32 address.
* The asset already exists.

### MsgCreateAsset

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L53-L64

#### Asset

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/asset.proto#L33-L49

### MsgCreateAssetResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L66-L67


## CreateAssetClass

The `CreateAssetClass` endpoint creates a new asset class that defines the classification and schema for digital assets. Asset classes are stored as NFT classes in the underlying NFT module.

It is expected to fail if:
* The `asset_class` is nil.
* The `asset_class.id` is empty.
* The `asset_class.name` is empty.
* The `asset_class.data` is not valid JSON schema.
* The `signer` is not a valid bech32 address.
* The asset class already exists.

### MsgCreateAssetClass

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L69-L77

#### AssetClass

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/asset.proto#L9-L31

### MsgCreateAssetClassResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L79-L80


## CreatePool

The `CreatePool` endpoint creates a pool of NFTs represented by a marker token. When a pool is created:
* A marker account is created with the specified denom.
* The specified NFTs are transferred to the pool marker address.
* The pool marker represents ownership of the underlying assets.

It is expected to fail if:
* The `pool` is nil or invalid.
* The `assets` list is empty.
* Any of the assets in the `assets` list have an empty `class_id` or `id`.
* The `signer` does not own all specified assets.
* The `signer` is not a valid bech32 address.
* Any of the specified assets do not exist.

### MsgCreatePool

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L82-L92

See also: [AssetKey](#assetkey).

### MsgCreatePoolResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L94-L95


## CreateTokenization

The `CreateTokenization` endpoint creates a tokenization marker representing fractional ownership of an individual NFT. When a tokenization is created:
* A marker account is created with the specified denomination.
* The underlying NFT is transferred to the tokenization marker address.
* The marker tokens represent fractional ownership of the underlying asset.

It is expected to fail if:
* The `token` is nil or invalid.
* The `asset` is nil.
* The `asset.class_id` is empty.
* The `asset.id` is empty.
* The `signer` does not own the specified asset.
* The `signer` is not a valid bech32 address.
* The specified asset does not exist.

### MsgCreateTokenization

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L97-L107

See also: [AssetKey](#assetkey).

### MsgCreateTokenizationResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L109-L110


## CreateSecuritization

The `CreateSecuritization` endpoint creates a securitization with multiple pools and tranches. When a securitization is created:
* A main securitization marker is created with denom equal to the securitization `id`.
* Individual tranche markers are created with denoms `{id}.tranche.{tranche_denom}`.
* All access permissions are revoked from the pool markers and the asset module account is granted full administrative access.
* The tranche markers represent different risk/return profiles of the underlying assets.

It is expected to fail if:
* The `id` is empty.
* The `pools` list is empty.
* Any pool ID in the `pools` list is empty.
* The `tranches` list is empty.
* Any tranche in the `tranches` list is invalid.
* The `signer` is not a valid bech32 address.
* Any of the specified pools do not exist.

### MsgCreateSecuritization

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L112-L124

### MsgCreateSecuritizationResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/tx.proto#L126-L127
