# Registry Messages

The `x/registry` module has several `Msg` endpoints.

---
<!-- TOC 2 2 -->
  - [RegisterNFT](#registernft)
  - [GrantRole](#grantrole)
  - [RevokeRole](#revokerole)
  - [UnregisterNFT](#unregisternft)
  - [RegistryBulkUpdate](#registrybulkupdate)


## RegisterNFT

A registry is initially created and assigned to an NFT using a `MsgRegisterNFT`.

It is expected to fail if:
* The NFT does not exist.
* The NFT already has a registry.
* The `signer` is not the owner of the NFT.
* The msg is invalid.

### MsgRegisterNFT

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L39-L55

See also: [RegistryKey](01_concepts.md#registrykey), [RolesEntry](01_concepts.md#rolesentry).

### MsgRegisterNFTResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L57-L59


## GrantRole

After the registry has been created for an NFT, roles can be added using a `MsgGrantRole`.

It is expected to fail if:
* The NFT does not yet have a registry.
* The `signer` is not the owner of the NFT.
* A provided address already has the provided role.
* The msg is invalid.

### MsgGrantRole

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L61-L81

See also: [RegistryKey](01_concepts.md#registrykey), [RegistryRole](01_concepts.md#registryrole).

### MsgGrantRoleResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L83-L85


## RevokeRole

After the registry has been created for an NFT, roles can be removed using a `MsgRevokeRole`.

It is expected to fail if:
* The NFT does not yet have a registry.
* The `signer` is not the owner of the NFT.
* A provided address does not have the provided role.
* The msg is invalid.

### MsgRevokeRole

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L87-L107

See also: [RegistryKey](01_concepts.md#registrykey), [RegistryRole](01_concepts.md#registryrole).

### MsgRevokeRoleResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L109-L111


## UnregisterNFT

A registry can be deleted using a `MsgUnregisterNFT`.

It is expected to fail if:
* The NFT does not yet have a registry.
* The `signer` is not the owner of the NFT.
* The msg is invalid.

### MsgUnregisterNFT

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L113-L125

See also: [RegistryKey](01_concepts.md#registrykey).

### MsgUnregisterNFTResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L127-L129


## RegistryBulkUpdate

Multiple registries can be created or updated at once using a `MsgRegistryBulkUpdate`.
This action will overwrite any existing registry entry to the one provided.

It is expected to fail if:
* The `signer` is not the owner of one or more NFTs involved.
* The msg is invalid.

### MsgRegistryBulkUpdate

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L131-L143

See also: [RegistryEntry](01_concepts.md#registryentry).

### MsgRegistryBulkUpdateResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/registry/v1/tx.proto#L145-L147
