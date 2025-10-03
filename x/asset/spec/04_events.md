# Asset Events

The asset module emits several events for various asset management operations.

---
<!-- TOC -->
  - [EventAssetBurned](#eventassetburned)
  - [EventAssetClassCreated](#eventassetclasscreated)
  - [EventAssetCreated](#eventassetcreated)
  - [EventPoolCreated](#eventpoolcreated)
  - [EventTokenizationCreated](#eventtokenizationcreated)
  - [EventSecuritizationCreated](#eventsecuritizationcreated)


## EventAssetBurned

When an asset is burned, an `EventAssetBurned` is emitted.

This event is triggered by the [MsgBurnAsset](03_messages.md#burnasset) message handler when an asset is successfully burned and removed from circulation.

Event Type: `provenance.asset.v1.EventAssetBurned`

| Attribute Key | Attribute Value                                              |
|---------------|--------------------------------------------------------------|
| class_id      | The asset class identifier of the burned asset.              |
| id            | The identifier of the burned asset.                          |
| owner         | The bech32 address of the account that owned the asset.      |


## EventAssetClassCreated

When a new asset class is created, an `EventAssetClassCreated` is emitted.

This event is triggered by the [MsgCreateAssetClass](03_messages.md#createassetclass) message handler when an asset class is successfully created.

Event Type: `provenance.asset.v1.EventAssetClassCreated`

| Attribute Key | Attribute Value                                              |
|---------------|--------------------------------------------------------------|
| class_id      | The unique identifier of the created asset class.            |
| class_name    | The human-readable name of the asset class.                  |
| class_symbol  | The symbol or ticker for the asset class.                    |


## EventAssetCreated

When a new asset is created, an `EventAssetCreated` is emitted.

This event is triggered by the [MsgCreateAsset](03_messages.md#createasset) message handler when an asset is successfully created and minted.

Event Type: `provenance.asset.v1.EventAssetCreated`

| Attribute Key | Attribute Value                                              |
|---------------|--------------------------------------------------------------|
| class_id      | The asset class identifier of the created asset.             |
| id            | The identifier of the created asset.                         |
| owner         | The bech32 address of the account that owns the asset.       |


## EventPoolCreated

When a new pool is created, an `EventPoolCreated` is emitted.

This event is triggered by the [MsgCreatePool](03_messages.md#createpool) message handler when a pool is successfully created with assets.

Event Type: `provenance.asset.v1.EventPoolCreated`

| Attribute Key | Attribute Value                                              |
|---------------|--------------------------------------------------------------|
| pool          | The coin representation of the created pool.                 |
| asset_count   | The number of assets added to the pool.                      |
| owner         | The bech32 address of the account that created the pool.     |


## EventTokenizationCreated

When a tokenization marker is created, an `EventTokenizationCreated` is emitted.

This event is triggered by the [MsgCreateTokenization](03_messages.md#createtokenization) message handler when a tokenization is successfully created for an asset.

Event Type: `provenance.asset.v1.EventTokenizationCreated`

| Attribute Key | Attribute Value                                              |
|---------------|--------------------------------------------------------------|
| tokenization  | The coin representation of the tokenization marker.          |
| class_id      | The asset class identifier of the tokenized asset.           |
| id            | The identifier of the tokenized asset.                       |
| owner         | The bech32 address of the account that created the tokenization. |


## EventSecuritizationCreated

When a securitization is created, an `EventSecuritizationCreated` is emitted.

This event is triggered by the [MsgCreateSecuritization](03_messages.md#createsecuritization) message handler when a securitization is successfully created with tranches and pools.

Event Type: `provenance.asset.v1.EventSecuritizationCreated`

| Attribute Key     | Attribute Value                                              |
|-------------------|--------------------------------------------------------------|
| securitization_id | The unique identifier of the created securitization.         |
| tranche_count     | The number of tranches in the securitization.                |
| pool_count        | The number of pools in the securitization.                   |
| owner             | The bech32 address of the account that created the securitization. |

