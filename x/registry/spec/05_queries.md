# Registry Queries

There are several queries for getting information about things in the registry module.

---
<!-- TOC 2 2 -->


## GetRegistry

All roles and addresses associated with an NFT can be looked up using the `GetRegistry` query.

### QueryGetRegistryRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/query.proto#L35-L41

See also: [RegistryKey](01_concepts.md#registrykey).

### QueryGetRegistryResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/query.proto#L43-L49

See also: [RegistryEntry](01_concepts.md#registryentry).


## GetRegistries

All registries can be looked up using the `GetRegistries`.
This query can be optionally restricted to only those with a given asset class id.

This query is paginated.

### QueryGetRegistriesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/query.proto#L51-L59

### QueryGetRegistriesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/query.proto#L61-L70

See also: [RegistryEntry](01_concepts.md#registryentry).


## HasRole

To check if a specific address has a certain role for an NFT, use the `HasRole` query.

### QueryHasRoleRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/query.proto#L72-L86

See also: [RegistryKey](01_concepts.md#registrykey), [RegistryRole](01_concepts.md#registryrole).

### QueryHasRoleResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/query.proto#L88-L94
