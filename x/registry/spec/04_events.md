# Registry Events

The registry module emits several events for various actions.

---
<!-- TOC -->
  - [EventNFTRegistered](#eventnftregistered)
  - [EventRoleGranted](#eventrolegranted)
  - [EventRoleRevoked](#eventrolerevoked)
  - [EventNFTUnregistered](#eventnftunregistered)
  - [EventRegistryBulkUpdated](#eventregistrybulkupdated)


## EventNFTRegistered

Any time a new registry is created, an `EventNFTRegistered` is emitted.

Event type: `provenance.registry.v1.EventNFTRegistered`

| Attribute Key  | Attribute Value                                    |
|----------------|----------------------------------------------------|
| nft_id         | The ID of the NFT the registry was just made for.  |
| asset_class_id | The ID of the asset class the NFT belongs to.      |


## EventRoleGranted

Any time a role is granted to an address for an NFT, an `EventRoleGranted` is emitted.
This includes registry creation and bulk operations.

Event type: `provenance.registry.v1.EventRoleGranted`

| Attribute Key  | Attribute Value                                            |
|----------------|------------------------------------------------------------|
| nft_id         | The ID of the NFT the registry was just made for.          |
| asset_class_id | The ID of the asset class the NFT belongs to.              |
| role           | The role that the addresses were granted.                  |
| addresses      | The addresses granted the role as a JSON array of strings. |


## EventRoleRevoked

Any time a role is revoked for an address on an NFT, an `EventRoleRevoked` is emitted.
This includes registry deletion and bulk operations.

Event type: `provenance.registry.v1.EventRoleRevoked`

| Attribute Key  | Attribute Value                                            |
|----------------|------------------------------------------------------------|
| nft_id         | The ID of the NFT the registry was just made for.          |
| asset_class_id | The ID of the asset class the NFT belongs to.              |
| role           | The role that the addresses were granted.                  |
| addresses      | The addresses granted the role as a JSON array of strings. |


## EventNFTUnregistered

Any time an existing registry is deleted, an `EventNFTUnregistered` is emitted.

Event type: `provenance.registry.v1.EventNFTUnregistered`

| Attribute Key  | Attribute Value                                    |
|----------------|----------------------------------------------------|
| nft_id         | The ID of the NFT the registry was just made for.  |
| asset_class_id | The ID of the asset class the NFT belongs to.      |


## EventRegistryBulkUpdated

During a bulk operation, if a registry is being updated, a `EventRegistryBulkUpdated` is emitted.
This is only emitted when the registry is being UPDATED, though.
When a registry is created during a bulk operation, an `EventNFTRegistered` is emitted instead.

Event type: `provenance.registry.v1.EventRegistryBulkUpdated`

| Attribute Key  | Attribute Value                                    |
|----------------|----------------------------------------------------|
| nft_id         | The ID of the NFT the registry was just made for.  |
| asset_class_id | The ID of the asset class the NFT belongs to.      |
