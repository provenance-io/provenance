# Metadata Queries

In this section we describe the queries available for looking up metadata information.
All state objects specified by each message are defined within the [state](02_state.md) section.

Each entry or specification state object is wrapped with an `*IdInfo` message containing information about that state object's address/id.

<!-- TOC 2 -->
  - [Params](#params)
  - [Scope](#scope)
  - [ScopesAll](#scopesall)
  - [Sessions](#sessions)
  - [SessionsAll](#sessionsall)
  - [Records](#records)
  - [RecordsAll](#recordsall)
  - [Ownership](#ownership)
  - [ValueOwnership](#valueownership)
  - [ScopeSpecification](#scopespecification)
  - [ScopeSpecificationsAll](#scopespecificationsall)
  - [ContractSpecification](#contractspecification)
  - [ContractSpecificationsAll](#contractspecificationsall)
  - [RecordSpecificationsForContractSpecification](#recordspecificationsforcontractspecification)
  - [RecordSpecification](#recordspecification)
  - [RecordSpecificationsAll](#recordspecificationsall)
  - [OSLocatorParams](#oslocatorparams)
  - [OSLocator](#oslocator)
  - [OSLocatorsByURI](#oslocatorsbyuri)
  - [OSLocatorsByScope](#oslocatorsbyscope)
  - [OSAllLocators](#osalllocators)


---
## Params

The `Params` query gets the parameters of the metadata module.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L223-L224

There are no inputs for this query.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L226-L233


---
## Scope

The `Scope` query gets a scope.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L235-L250

The `scope_id`, if provided, must either be scope uuid, e.g. `91978ba2-5f35-459a-86a7-feca1b0512e0` or a scope address,
e.g. `scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel`. The session addr, if provided, must be a bech32 session address,
e.g. `session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr`. The record_addr, if provided, must be a
bech32 record address, e.g. `record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3`.

* If only a `scope_id` is provided, that scope is returned.
* If only a `session_addr` is provided, the scope containing that session is returned.
* If only a `record_addr` is provided, the scope containing that record is returned.
* If more than one of `scope_id`, `session_addr`, and `record_addr` are provided, and they don't refer to the same scope,
a bad request is returned.

Providing a `session_addr` or `record_addr` does not limit the sessions and records returned (if requested).
Those parameters are only used to find the scope.

By default, sessions and records are not included.
Set `include_sessions` and/or `include_records` to true to include sessions and/or records.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L252-L263


---
## ScopesAll

The `ScopesAll` query gets all scopes.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L275-L279

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L281-L290


---
## Sessions

The `Sessions` query gets sessions.

### Request
+++ https://github.com/provenance-io/provenance/blob/12e927800df502d0625de77b7fb2051632eecd22/proto/provenance/metadata/v1/query.proto#L308-L326

The `scope_id` can either be scope uuid, e.g. `91978ba2-5f35-459a-86a7-feca1b0512e0` or a scope address, e.g.
`scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel`. Similarly, the `session_id` can either be a uuid or session address, e.g.
`session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr`. The `record_addr`, if provided, must be a
bech32 record address, e.g. `record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3`.

* If only a `scope_id` is provided, all sessions in that scope are returned.
* If only a `session_id` is provided, it must be an address, and that single session is returned.
* If the `session_id` is a uuid, then either a `scope_id` or `record_addr` must also be provided, and that single session
is returned.
* If only a `record_addr` is provided, the session containing that record will be returned.
* If a `record_name` is provided then either a `scope_id`, `session_id` as an address, or `record_addr` must also be
provided, and the session containing that record will be returned.

A bad request is returned if:
* The `session_id` is a uuid and is provided without a `scope_id` or `record_addr`.
* A `record_name` is provided without any way to identify the scope (e.g. a `scope_id`, a `session_id` as an address, or
a `record_addr`).
* Two or more of `scope_id`, `session_id` as an address, and `record_addr` are provided and don't all refer to the same
scope.
* A `record_addr` (or `scope_id` and `record_name`) is provided with a `session_id` and that session does not contain such
a record.
* A `record_addr` and `record_name` are both provided, but reference different records.

By default, the scope and records are not included.
Set `include_scope` and/or `include_records` to true to include the scope and/or records.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L308-L319


---
## SessionsAll

The `SessionsAll` query gets all sessions.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L331-L335

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L337-L346


---
## Records

The `Records` query gets records.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L348-L366

The `record_addr`, if provided, must be a bech32 record address, e.g.
`record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3`. The `scope_id` can either be scope uuid, e.g.
`91978ba2-5f35-459a-86a7-feca1b0512e0 `or a scope address, e.g. `scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel`. Similarly,
the `session_id` can either be a uuid or session address, e.g.
`session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr`. The name is the name of the record you're
interested in.

* If only a `record_addr` is provided, that single record will be returned.
* If only a `scope_id` is provided, all records in that scope will be returned.
* If only a `session_id` (or scope_id/session_id), all records in that session will be returned.
* If a `name` is provided with a `scope_id` and/or `session_id`, that single record will be returned.

A bad request is returned if:
* The `session_id` is a uuid and no `scope_id` is provided.
* There are two or more of `record_addr`, `session_id`, and `scope_id`, and they don't all refer to the same scope.
* A `name` is provided, but not a `scope_id` and/or a `session_id`.
* A `name` and `record_addr` are provided and the name doesn't match the record_addr.

By default, the scope and sessions are not included.
Set `include_scope` and/or `include_sessions` to true to include the scope and/or sessions.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L368-L379


---
## RecordsAll

The `RecordsAll` query gets all records.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L391-L395

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L397-L406


---
## Ownership

The `Ownership` query gets the ids of scopes owned by an address.

A scope is owned by an address if the address is listed as either an owner, or the value owner.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L408-L414

The `address` should be a bech32 address string.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L416-L425


---
## ValueOwnership

The `ValueOwnership` query gets gets the ids of scopes that list an address as the value owner.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L427-L433

The `address` should be a bech32 address string.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L435-L444


---
## ScopeSpecification

The `ScopeSpecification` query gets a scope specification.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L446-L451

The `specification_id` can either be a uuid, e.g. `dc83ea70-eacd-40fe-9adf-1cf6148bf8a2` or a bech32 scope
specification address, e.g. `scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m`.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L453-L460


---
## ScopeSpecificationsAll

The `ScopeSpecificationsAll` query gets all scope specifications.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L470-L474

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L476-L485


---
## ContractSpecification

The `ContractSpecification` query gets a contract specification.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L487-L498

The `specification_id` can either be a uuid, e.g. `def6bc0a-c9dd-4874-948f-5206e6060a84`, a bech32 contract
specification address, e.g. `contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn`, or a bech32 record specification
address, e.g. `recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44`. If it is a record specification
address, then the contract specification that contains that record specification is looked up.

By default, the record specifications for this contract specification are not included.
Set `include_record_specs` to true to include them in the result.


### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L500-L511


---
## ContractSpecificationsAll

The `ContractSpecificationsAll` query gets all contract specifications.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L521-L525

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L527-L537


---
## RecordSpecificationsForContractSpecification

The `RecordSpecificationsForContractSpecification` query gets all record specifications for a contract specification.

The only difference between this query and `ContractSpecification` with `include_record_specs = true` is that
this query does not return the contract specification.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L539-L547

The `specification_id` can either be a uuid, e.g. `def6bc0a-c9dd-4874-948f-5206e6060a84`, a bech32 contract
specification address, e.g. `contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn`, or a bech32 record specification
address, e.g. `recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44`. If it is a record specification
address, then the contract specification that contains that record specification is used.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L549-L562


---
## RecordSpecification

The `RecordSpecification` query gets a record specification.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L564-L575

The `specification_id` can either be a uuid, e.g. `def6bc0a-c9dd-4874-948f-5206e6060a84` or a bech32 contract specification
address, e.g. `contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn`.
It can also be a record specification address, e.g.
`recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44`.

The `name` is the name of the record to look up.
It is required if the `specification_id` is a uuid or contract specification address.
It is ignored if the `specification_id` is a record specification address.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L577-L584


---
## RecordSpecificationsAll

The `RecordSpecificationsAll` query gets all record specifications.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L594-L598

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L600-L610


---
## OSLocatorParams

The `OSLocatorParams` query gets the parameters of the Object Store Locator sub-module.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L612-L613

There are no inputs for this query.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L615-L622


---
## OSLocator

The `OSLocator` query gets an Object Store Locator for an address.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L624-L627

The `owner` should be a bech32 address string.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L629-L635


---
## OSLocatorsByURI

The `OSLocatorsByURI` query gets the object store locators by URI.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L637-L643

The `uri` is string the URI to find object store locators for.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L645-L653


---
## OSLocatorsByScope

The `OSLocatorsByScope` query gets the object store locators for the owners and value owner of a scope.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L655-L658

The `scope_id`, must either be scope uuid, e.g. `91978ba2-5f35-459a-86a7-feca1b0512e0` or a scope address,
e.g. `scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel`

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L660-L666


---
## OSAllLocators

The `OSAllLocators` query gets all object store locators.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L668-L672

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L674-L682
