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
