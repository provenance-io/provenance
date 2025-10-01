# Registry Concepts

The Registry module assigns roles to addresses associated NFTs.

---
<!-- TOC -->


## Overview

A registry is a set of addresses with specific roles; an address can have multiple roles, and a role can have multiple addresses.

A registry can be added to an NFT using the `RegisterNFT` endpoint.
After that, it can be managed using the `GrantRole` and `RevokeRole` endpoints.
To check if an address has a role for an NFT, use the `HasRole` query.

## Types

There are a few message types used throughout this module.

### RegistryRole

The different roles available are defined in the `RegistryRole` enum.

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/registry.proto#L1-L1

### RegistryKey

A `RegistryKey` contains the information that uniquely identifies an NFT.

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/registry.proto#L1-L1

### RegistryEntry

A registry entry contains all of the role and address associations for a given NFT.

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/registry.proto#L1-L1

See also: [RegistryKey](#registrykey), [RolesEntry](#rolesentry).

### RolesEntry

A `RolesEntry` associates any number of addresses with a specific role.

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/registry/v1/registry.proto#L1-L1

See also: [RegistryRole](#registryrole).
