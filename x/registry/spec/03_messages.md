# Registry Messages

The `x/registry` module has several `Msg` endpoints.

---
<!-- TOC 2 2 -->


## RegisterNFT

A registry is initially created and assigned to an NFT using a `MsgRegisterNFT`.

It is expected to fail if:
* The NFT does not exist.
* The NFT already has a registry.
* The `signer` is not the owner of the NFT.
* The msg is invalid.

### MsgRegisterNFT

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1

### RegistryKey

A `RegistryKey` contains the information that uniquely identifies an NFT.

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/registry.proto#L1-L1

### RolesEntry

A `RolesEntry` associates any number of addresses with a specific role.

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/registry.proto#L1-L1

### MsgRegisterNFTResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1


## GrantRole

After the registry has been created for an NFT, roles can be added using a `MsgGrantRole`.

It is expected to fail if:
* The NFT does not yet have a registry.
* The `signer` is not the owner of the NFT.
* A provided address already has the provided role.
* The msg is invalid.

### MsgGrantRole

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1

See also: [RegistryKey](#registrykey)

### MsgGrantRoleResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1


## RevokeRole

After the registry has been created for an NFT, roles can be removed using a `MsgRevokeRole`.

It is expected to fail if:
* The NFT does not yet have a registry.
* The `signer` is not the owner of the NFT.
* A provided address does not have the provided role.
* The msg is invalid.

### MsgRevokeRole

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1

See also: [RegistryKey](#registrykey)

### MsgRevokeRoleResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1


## UnregisterNFT

A registry can be deleted using a `MsgUnregisterNFT`.

It is expected to fail if:
* The NFT does not yet have a registry.
* The `signer` is not the owner of the NFT.
* The msg is invalid.

### MsgUnregisterNFT

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1

See also: [RegistryKey](#registrykey)

### MsgUnregisterNFTResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1


## RegistryBulkUpdate

Multiple registries can be created or updated at once using a `MsgRegistryBulkUpdate`.
This action will overwrite any existing registry entry to the one provided.

It is expected to fail if:
* The `signer` is not the owner of one or more NFTs involved.
* The msg is invalid.

### MsgRegistryBulkUpdate

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1

### RegistryEntry

A registry entry contains all of the role and address associations for a given NFT.

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/registry.proto#L1-L1

See also: [RegistryKey](#registrykey)

### MsgRegistryBulkUpdateResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/tx.proto#L1-L1
