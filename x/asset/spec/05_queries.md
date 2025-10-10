# Asset Queries

There are several queries for getting information about assets and asset classes in the asset module.

---
<!-- TOC 2 2 -->
  - [Asset](#asset)
  - [Assets](#assets)
  - [AssetClass](#assetclass)
  - [AssetClasses](#assetclasses)


## Asset

Use the `Asset` query to look up a specific asset by its class ID and asset ID.

### QueryAssetRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/query.proto#L36-L43

### QueryAssetResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/query.proto#L45-L49

See also: [Asset](03_messages.md#asset).


## Assets

The `Assets` query retrieves all assets for a given owner and optionally filters by class ID.

This query is paginated.

### QueryAssetsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/query.proto#L51-L61

### QueryAssetsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/query.proto#L63-L70

See also: [Asset](03_messages.md#asset).


## AssetClass

Use the `AssetClass` query to look up a specific asset class by its ID.

### QueryAssetClassRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/query.proto#L72-L76

### QueryAssetClassResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/query.proto#L78-L82

See also: [AssetClass](03_messages.md#assetclass).


## AssetClasses

Use the `AssetClasses` query to get all asset classes in the system.

This query is paginated.

### QueryAssetClassesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/query.proto#L84-L88

### QueryAssetClassesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/asset/v1/query.proto#L90-L97

See also: [AssetClass](03_messages.md#assetclass).
