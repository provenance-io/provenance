# Metadata Concepts

The metadata service manages things that define and reference off-chain data.
There are three categories of things stored: entries, specifications, and object store locators.
Each entry and specification has a unique Metadata Address that is often simply called its "id".

## Entries

The term "entries" refers to scopes, sessions, and records.
They group and identify information.

### Scope

A scope is a high-level grouping information combined with some access control.

* A scope must conform to a pre-determined scope specification.
* A scope is used to group together many sessions and records.

### Session

A session is a grouping of records that are all generally created together.
It is conceptually similar to the running of a program.

* A session must conform to a pre-determined contract specification.
* A session groups together a collection of records.
* A session is part of exactly one scope.

### Record

A record identifies the inputs and outputs of a process.
It is conceptually similar to the values involved in a method call.

* A record must conform to a pre-determined record specification.
* A record is part of exactly one scope.
* A record is part of exactly one session.

## Specifications

The term "specifications" refers to scope specifications, contract specifications, and record specifications.
They define validation parameters for the various entries.
Ideally, specifications will be used for multiple entries.

### Scope Specification

A scope specification defines validation parameters for scopes.
They group together contract specifications and define roles that must be involved in a scope.

### Contract Specification

A contract specification defines validation parameters for sessions.
They contain source information and roles that must be involved in a session.
They also group together record specifications.

A contract specification can be part of multiple scope specifications.

### Record Specification

A record specification defines validation parameters for records.
They contain expected inputs and outputs and parties that must be involved in a record.

A record specification is part of exactly one contract specification.

## Metadata Addresses

Entries and Specifications must each have a unique metadata address.
These addresses are byte arrays that are commonly referered to as "ids".
As strings, they should be represented using the bech32 address format.
The byte arrays for the different messages have specific formats that help facilitate grouping and indexing.

### Scope Metadata Address:

Byte Array Length: `17`

| Byte range | Description
|------------|---
| 0          | `0x00`
| 1-16       | UUID of this scope.

* Field Name: `Scope.scope_id`
* Bech32 HRP: `"scope"`
* Bech32 Example: `"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"`

### Session Metadata Address:

Byte Array Length: `33`

| Byte range | Description
|------------|---
| 0          | `0x01`
| 1-16       | UUID of the scope that this session is part of
| 17-32      | UUID of this session

* Field Name: `Session.session_id`
* Bech32 HRP: `"session"`
* Bech32 Example: `"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"`

### Record Metadata Address:

Byte Array Length: `33`

| Byte range | Description
|------------|---
| 0          | `0x02`
| 1-16       | UUID of the scope that this record is part of
| 17-32      | First 16 bytes of the SHA256 checksum of this record's name

* Field Name: `Record.record_id`
* Bech32 HRP: `"record"`
* Bech32 Example: `"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"`

### Scope Spec Metadata Address:

Byte Array Length: `17`

| Byte range | Description
|------------|---
| 0          | `0x03`
| 1-16       | UUID of this scope specification

* Field Name: `ScopeSpecification.specification_id`
* Bech32 HRP: `"scopespec"`
* Bech32 Example: `"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"`

### Contract Spec Metadata Address:

Byte Array Length: `17`

| Byte range | Description
|------------|---
| 0          | `0x04`
| 1-16       | UUID of this contract specification

* Field Name: `ContractSpecification.specification_id`
* Bech32 HRP: `"contractspec"`
* Bech32 Example: `"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"`

### Record Spec Metadata Address:

Byte Array Length: `33`

| Byte range | Description
|------------|---
| 0          | `0x05`
| 1-16       | UUID of the contract specification that this record specification is part of
| 17-32      | First 16 bytes of the SHA256 checksum of this record specification's name

* Field Name: `RecordSpecification.specification_id`
* Bech32 HRP: `"recspec"`
* Bech32 Example: `"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"`

### Generalities

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

## Object Store Locators

An object store locator defines the off-chain location of objects referenced on-chain.
