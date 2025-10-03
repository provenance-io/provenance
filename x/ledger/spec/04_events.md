# Ledger Events

The Ledger module emits events to track state changes and provide transparency for ledger operations. Events are defined as protobuf types in `proto/provenance/ledger/v1/events.proto` and are emitted for all major state transitions including ledger creation, updates, entry additions, and destruction.

---
<!-- TOC 2 2 -->
  - [EventLedgerCreated](#eventledgercreated)
  - [EventLedgerUpdated](#eventledgerupdated)
  - [EventLedgerEntryAdded](#eventledgerentryadded)
  - [EventFundTransferWithSettlement](#eventfundtransferwithsettlement)
  - [EventLedgerDestroyed](#eventledgerdestroyed)


## EventLedgerCreated

When a new ledger is created, an `EventLedgerCreated` is emitted.

Event type: `provenance.ledger.v1.EventLedgerCreated`

| Attribute Key  | Attribute Value                                    |
|----------------|----------------------------------------------------|
| asset_class_id | The ID of the asset class the NFT belongs to.      |
| nft_id         | The ID of the NFT the registry was just made for.  |


## EventLedgerUpdated

When a ledger's configuration is updated, an `EventLedgerUpdated` is emitted.

Event type: `provenance.ledger.v1.EventLedgerUpdated`

| Attribute Key  | Attribute Value                                   |
|----------------|---------------------------------------------------|
| asset_class_id | The ID of the asset class the NFT belongs to.     |
| nft_id         | The ID of the NFT the registry was just made for. |
| update_type    | The type of thing updated (see below).            |

### UpdateType

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/events.proto#L35-L52


## EventLedgerEntryAdded

When a ledger entry is added, an `EventLedgerEntryAdded` is emitted.

Event type: `provenance.ledger.v1.EventLedgerEntryAdded`

| Attribute Key   | Attribute Value                                   |
|-----------------|---------------------------------------------------|
| asset_class_id  | The ID of the asset class the NFT belongs to.     |
| nft_id          | The ID of the NFT the registry was just made for. |
| correlation_id  | The ID used to correlate ledger entries.          |


## EventFundTransferWithSettlement

When funds are transferred with settlement instructions, an `EventFundTransferWithSettlement` is emitted.

Event type: `provenance.ledger.v1.EventFundTransferWithSettlement`

| Attribute Key   | Attribute Value                                   |
|-----------------|---------------------------------------------------|
| asset_class_id  | The ID of the asset class the NFT belongs to.     |
| nft_id          | The ID of the NFT the registry was just made for. |
| correlation_id  | The ID used to correlate ledger entries.          |


## EventLedgerDestroyed

When a ledger is destroyed, an `EventLedgerDestroyed` is emitted.

Event type: `provenance.ledger.v1.EventLedgerDestroyed`

| Attribute Key  | Attribute Value                                    |
|----------------|----------------------------------------------------|
| asset_class_id | The ID of the asset class the NFT belongs to.      |
| nft_id         | The ID of the NFT the registry was just made for.  |
