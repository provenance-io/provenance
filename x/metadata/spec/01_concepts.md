# Metadata Concepts

The metadata service manages things that define and reference off-chain data.
There are three categories of things stored: entries, specifications, and object store locators.
Each entry and specification has a unique Metadata Address that is often simply called its "id".
Additionally, several indexes are created to help with linking and iterating over related messages.

<!-- TOC -->
  - [Entries](#entries)
  - [Specifications](#specifications)
  - [Metadata Addresses](#metadata-addresses)
    - [MetadataAddress Example Implementations](#metadataaddress-example-implementations)
    - [MetadataAddress General Guidelines](#metadataaddress-general-guidelines)
  - [Indexes](#indexes)
  - [Signing Requirements](#signing-requirements)
    - [Scope Value Owner Address Requirements](#scope-value-owner-address-requirements)
    - [Smart Contract Requirements](#smart-contract-requirements)
    - [With Party Rollup Required](#with-party-rollup-required)
    - [Without Party Rollup Required](#without-party-rollup-required)



## Entries

The term "entries" refers to scopes, sessions, and records.
See [Entries](02_state.md#entries) for details.

## Specifications

The term "specifications" refers to scope specifications, contract specifications, and record specifications.
See [Specifications](02_state.md#specifications) for details.

## Metadata Addresses

Entries and Specifications must each have a unique metadata address.
These addresses are byte arrays that are commonly referered to as "ids".
As strings, they should be represented using the bech32 address format.
The addresses for the different messages have specific formats that help facilitate grouping and indexing.
All addresses start with a single byte that identifies the type, and are followed by 16 bytes commonly called a UUID.
Some address types contain other elements too.

### MetadataAddress Example Implementations

* Go: [address.go](https://github.com/provenance-io/provenance/blob/main/x/metadata/spec/examples/go/metadata_address.go)
* Kotlin: [MetadataAddress.kt](https://github.com/provenance-io/provenance/blob/main/x/metadata/spec/examples/kotlin/src/main/kotlin/MetadataAddress.kt)
* Javascript: [metadata-address.js](https://github.com/provenance-io/provenance/blob/main/x/metadata/spec/examples/js/lib/metadata-address.js)

### MetadataAddress General Guidelines

* As strings, the metadata addresses are represented using the bech32 address format.
* The `*IdInfo` messages defined in `metadata.proto` (e.g. `RecordIdInfo`) are used in response messages and contain a breakdown of a metadata address.
* Variables that hold the addresses as byte arrays should end in `_id`.
* Variables that hold the addresses as bech32 strings should end in `_addr`.
* Variables that hold UUIDs as strings should use the standard UUID format and end in `_uuid`.
* If a variable is a byte array that ends in `_id`, then it should be the full Metadata Address byte array.
* String variables that end in `_id` should only be used in input messages.
  They should be flexible fields that can accept either the bech32 string version of the Metadata Address byte array, or a UUID in the standard UUID string format.
* If a variable ends in `_addr`, then it should be the bech32 string version of the Metadata Address byte array.
* If a variable ends in `_uuid`, then it should be a UUID in the standard UUID string format.
  The exception to this is the byte array fields in the `*IdInfo` messages that represent the id broken into its various parts.
  For example, `ScopeIdInfo.scope_id_scope_uuid` represents the UUID portion of the `scope_id`, and is left as a byte array,
  but `ScopeIdInfo.scope_uuid` is the standard UUID string representation of those bytes.

## Indexes

Indexes are specially formatted entries in the kvstore used to find associated things.

The keys contain all of the relevant information.
They are byte arrays with three parts:
1. Type byte: A single byte representing the type of index.
1. Part 1: Address of the starting thing in the association.
1. Part 2: Address of the entry to find.

The values are always a single byte: `0x01`.

The general use of them is to create a prefix using the type byte and part 1.
Then use that prefix to iterate over all keys with that same prefix.
During iteration, remove the prefix from the current entry's key in order to get the key of the thing to find.

## Signing Requirements

Scopes have a `require_party_rollup` boolean field that dictates most signer requirements for a scope and all it's sessions and records.
There are also special signer considerations related to a scope's `value_owner_address` field.

### Scope Value Owner Address Requirements

These requirements are applied regardless of a scope's `require_party_rollup` value.
They are applied when writing new scopes, updating existing scopes, and deleting scopes.

If a scope with a value owner address is being updated, and the ONLY change is to that value owner address, then ONLY these signer requirements are applied and all other signer requirements are ignored.
If the value owner address is not changing, these requirements do not apply.
If the value owner address is changing as well as one or more other fields, these requirements apply as well as the other signer requirements.

* When a value owner address is being set to a marker, at least one of the signers must have deposit permission on that marker.
* When a value owner address is a marker and is being changed, at least one of the signers must have withdraw permission on that marker.
* When a value owner address is a non-marker address, and is being changed, that existing address must be one of the signers.
* When a value owner address is empty, and is being changed, standard scope signer requirements are also applied even if that's the only change to the scope.

### Smart Contract Requirements

The following are requirements related to smart contract usage of the `x/metadata` module:

* A party with a smart contract address MUST have the `PROVENANCE` role.
* A party with the `PROVENANCE` role MUST have the address of a smart contract.
* When a smart contract signs a message, it MUST be first or have only smart-contract signers before it, and SHOULD include the invoker address(es) after.
* When a smart contract is a signer, it must either be a party/owner, or have authorizations (via `x/authz`) from all signers after it.
* If a smart contract is a signer, but not a party, it cannot be the only signer, and cannot be the last signer.

### With Party Rollup Required

When a scope has `require_party_rollup = true`, all session parties must also be listed in the scope owners.
The use of `optional = true` parties is also allowed.
The party types (aka roles) defined in specifications, in conjunction with they entry's parties dictate the signers that are required (in addition to any `optional = false` parties).

For example, if a scope has an `optional = false` `CONTROLLER` (address `A`), and two `optional = true` `SERVICER`s (addresses `B`, and `C`), 
and a session is being written using a contract spec that requires just a `SERVICER` signature, then to write that session,
either address `B` or `C` must be a signer (due to the contract spec), and `A` must also sign (because they're `optional = false` in the scope).

#### Writing or Deleting a Scope With Party Rollup

* All roles required by the scope spec must have a party in the owners.
* If not new:
  * All `optional = false` existing owners must be signers.
  * All roles required by the scope spec must have a signer and associated party from the existing scope.
* Scope value owner address requirements are applied.

#### Writing a Session With Party Rollup

* All proposed session parties must be present in this scope's owners.
* All `optional = false` scope owners must be signers.
* If new:
  * All roles required by the contract spec must have a signer and associated party in the proposed session.
* If not new:
  * All roles required by the contract spec must have a signer and associated party in the existing session.
  * All roles required by the contract spec must have parties in the proposed session.
  * All `optional = false` existing parties must also be signers.

#### Writing a Record With Party Rollup

* All roles required by the record spec must have a signer and associated party in the session.
* All `optional = false` scope owners and session parties must be signers.
* If the record is changing sessions, all `optional = false` previous session parties must be signers.

#### Deleting a Record With Party Rollup

* All roles required by the record spec must have a signer and associated party in the scope.
* All `optional = false` scope owners must be signers.

### Without Party Rollup Required

When a scope has `require_party_rollup = false`, then `optional = true` parties are not allowed in the scope or any of its sessions.

#### Writing or Deleting a Scope Without Party Rollup

* All roles required by the scope spec must have a party in the owners.
* If not new, all existing owners must sign.
* Scope value owner address requirements are applied.

#### Writing a Session Without Party Rollup

* All roles required by the contract spec must have a party in the session parties.
* All scope owners must sign.

#### Writing a Record Without Party Rollup

* All roles required by the record spec must have a party in the session parties.
* All session parties must sign.
* If the record is changing to a new session, all previous session parties must sign.

#### Deleting a Record Without Party Rollup

* All scope owners must sign.
