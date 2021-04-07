# Metadata State

The Metadata module manages the state of seven different types:

1. [Scopes](#scopes)
1. [Sessions](#sessions)
1. [Records](#records)
1. [Scope Specifications](#scope-specifications)
1. [Contract Specifications](#contractsspecifications)
1. [Record Specifications](#record-specifications)
1. [Object Store Locators](#object-store-locators)

## Entries

The term "entries" refers to scopes, sessions, and records.
They group and identify information.

### Scopes

A scope is a high-level grouping information combined with some access control.

* A scope must conform to a pre-determined scope specification.
* A scope is used to group together many sessions and records.

#### Scope Metadata Addresses

Byte Array Length: `17`

| Byte range | Description
|------------|---
| 0          | `0x00`
| 1-16       | UUID of this scope.

* Field Name: `Scope.scope_id`
* Bech32 HRP: `"scope"`
* Bech32 Example: `"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"`

#### Scope Definition

TODO: Scope definition

#### Scope Indexes

TODO: Indexes for Scopes

### Sessions

A session is a grouping of records that are all generally created together.
It is conceptually similar to the running of a program.

* A session must conform to a pre-determined contract specification.
* A session groups together a collection of records.
* A session is part of exactly one scope.

#### Session Metadata Addresses

Byte Array Length: `33`

| Byte range | Description
|------------|---
| 0          | `0x01`
| 1-16       | UUID of the scope that this session is part of
| 17-32      | UUID of this session

* Field Name: `Session.session_id`
* Bech32 HRP: `"session"`
* Bech32 Example: `"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"`

#### Session Definition

TODO: Session definition

#### Session Indexes

TODO: Indexes for Sessions

### Records

A record identifies the inputs and outputs of a process.
It is conceptually similar to the values involved in a method call.

* A record must conform to a pre-determined record specification.
* A record is part of exactly one scope.
* A record is part of exactly one session.

#### Record Metadata Addresses

Byte Array Length: `33`

| Byte range | Description
|------------|---
| 0          | `0x02`
| 1-16       | UUID of the scope that this record is part of
| 17-32      | First 16 bytes of the SHA256 checksum of this record's name

* Field Name: `Record.record_id`
* Bech32 HRP: `"record"`
* Bech32 Example: `"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"`

#### Record Definition

TODO: Record definition

#### Record Indexes

TODO: Indexes for Records

## Specifications

The term "specifications" refers to scope specifications, contract specifications, and record specifications.
They define validation parameters for the various entries.
Ideally, specifications will be used for multiple entries.

### Scope Specifications

A scope specification defines validation parameters for scopes.
They group together contract specifications and define roles that must be involved in a scope.

#### Scope Specification Metadata Addresses

Byte Array Length: `17`

| Byte range | Description
|------------|---
| 0          | `0x03`
| 1-16       | UUID of this scope specification

* Field Name: `ScopeSpecification.specification_id`
* Bech32 HRP: `"scopespec"`
* Bech32 Example: `"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"`

#### Scope Specification Definition

TODO: Scope Specification definition

#### Scope Specification Indexes

TODO: Indexes for Scope Specifications

### Contract Specifications

A contract specification defines validation parameters for sessions.
They contain source information and roles that must be involved in a session.
They also group together record specifications.

A contract specification can be part of multiple scope specifications.

#### Contract Specification Metadata Addresses

Byte Array Length: `17`

| Byte range | Description
|------------|---
| 0          | `0x04`
| 1-16       | UUID of this contract specification

* Field Name: `ContractSpecification.specification_id`
* Bech32 HRP: `"contractspec"`
* Bech32 Example: `"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"`

#### Contract Specification Definition

TODO: Contract Specification definition

#### Contract Specification Indexes

TODO: Indexes for Contract Specifications

### Record Specifications

A record specification defines validation parameters for records.
They contain expected inputs and outputs and parties that must be involved in a record.

A record specification is part of exactly one contract specification.

#### Record Specification Metadata Addresses

Byte Array Length: `33`

| Byte range | Description
|------------|---
| 0          | `0x05`
| 1-16       | UUID of the contract specification that this record specification is part of
| 17-32      | First 16 bytes of the SHA256 checksum of this record specification's name

* Field Name: `RecordSpecification.specification_id`
* Bech32 HRP: `"recspec"`
* Bech32 Example: `"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"`

#### Record Specification Definition

TODO: Record Specification definition

#### Record Specification Indexes

TODO: Indexes for Record Specifications

## Object Store Locators

An object store locator indicates the location of off-chain data.

### Object Store Locator Addresses

TODO: Object Store Locator Addresses

### Object Store Locator Definition

TODO: Record Specification definition

### Object Store Locator Indexes

TODO: Indexes Object Store Locators

