# Metadata Queries

In this section we describe the queries available for looking up metadata information.
All state objects specified by each message are defined within the [state](02_state.md) section.

Each entry or specification state object is wrapped with an `*_id_info` message containing information about that state object's address/id.
By default, the `*_id_info` fields are populated with information about the metadata address(es) involved, but each applicable request has an `exclude_id_info` flag to cause those field to not be populated in the result.
If a requested entry or specification isn't found, an empty wrapper containing only id info is returned.

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
  - [GetByAddr](#getbyaddr)
  - [OSLocatorParams](#oslocatorparams)
  - [OSLocator](#oslocator)
  - [OSLocatorsByURI](#oslocatorsbyuri)
  - [OSLocatorsByScope](#oslocatorsbyscope)
  - [OSAllLocators](#osalllocators)
  - [AccountData](#accountdata)


---
## Params

The `Params` query gets the parameters of the metadata module.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L277-L281

There are no inputs for this query.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L283-L290


---
## Scope

The `Scope` query gets a scope.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L292-L312

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
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L314-L325


---
## ScopesAll

The `ScopesAll` query gets all scopes.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L337-L346

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L348-L357


---
## Sessions

The `Sessions` query gets sessions.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L359-L382

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
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L384-L395


---
## SessionsAll

The `SessionsAll` query gets all sessions.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L407-L416

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L418-L427


---
## Records

The `Records` query gets records.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L429-L452

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
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L454-L465


---
## RecordsAll

The `RecordsAll` query gets all records.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L477-L486

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L488-L497


---
## Ownership

The `Ownership` query gets the ids of scopes owned by an address.

A scope is owned by an address if the address is listed as either an owner, or the value owner.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L499-L507

The `address` should be a bech32 address string.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L509-L518


---
## ValueOwnership

The `ValueOwnership` query gets the ids of scopes that list an address as the value owner.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L520-L528

The `address` should be a bech32 address string.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L530-L539


---
## ScopeSpecification

The `ScopeSpecification` query gets a scope specification.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L541-L558

The `specification_id` can either be a uuid, e.g. `dc83ea70-eacd-40fe-9adf-1cf6148bf8a2` or a bech32 scope
specification address, e.g. `scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m`.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L560-L571


---
## ScopeSpecificationsAll

The `ScopeSpecificationsAll` query gets all scope specifications.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L581-L590

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L592-L601


---
## ContractSpecification

The `ContractSpecification` query gets a contract specification.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L603-L619

The `specification_id` can either be a uuid, e.g. `def6bc0a-c9dd-4874-948f-5206e6060a84`, a bech32 contract
specification address, e.g. `contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn`, or a bech32 record specification
address, e.g. `recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44`. If it is a record specification
address, then the contract specification that contains that record specification is looked up.

By default, the record specifications for this contract specification are not included.
Set `include_record_specs` to true to include them in the result.


### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L621-L631


---
## ContractSpecificationsAll

The `ContractSpecificationsAll` query gets all contract specifications.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L641-L650

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L652-L661


---
## RecordSpecificationsForContractSpecification

The `RecordSpecificationsForContractSpecification` query gets all record specifications for a contract specification.

The only difference between this query and `ContractSpecification` with `include_record_specs = true` is that
this query does not return the contract specification.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L663-L677

The `specification_id` can either be a uuid, e.g. `def6bc0a-c9dd-4874-948f-5206e6060a84`, a bech32 contract
specification address, e.g. `contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn`, or a bech32 record specification
address, e.g. `recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44`. If it is a record specification
address, then the contract specification that contains that record specification is used.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L679-L691


---
## RecordSpecification

The `RecordSpecification` query gets a record specification.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L693-L710

The `specification_id` can either be a uuid, e.g. `def6bc0a-c9dd-4874-948f-5206e6060a84` or a bech32 contract specification
address, e.g. `contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn`.
It can also be a record specification address, e.g.
`recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44`.

The `name` is the name of the record to look up.
It is required if the `specification_id` is a uuid or contract specification address.
It is ignored if the `specification_id` is a record specification address.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L712-L719


---
## RecordSpecificationsAll

The `RecordSpecificationsAll` query gets all record specifications.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L729-L738

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L740-L749


---
## GetByAddr

The `GetByAddr` query looks up metadata entries and/or specifications for a given list of addresses.
The results of this query are not wrapped with id information like the other queries, and only returns the exact entries requested.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L751-L755

The `addrs` can contain any valid metadata address bech32 strings.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L757-L773

Any invalid or nonexistent `addrs` will be in the `not_found` list.

---
## OSLocatorParams

The `OSLocatorParams` query gets the parameters of the Object Store Locator sub-module.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L775-L779

There are no inputs for this query.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L781-L788


---
## OSLocator

The `OSLocator` query gets an Object Store Locator for an address.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L790-L796

The `owner` should be a bech32 address string.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L798-L804


---
## OSLocatorsByURI

The `OSLocatorsByURI` query gets the object store locators by URI.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L806-L814

The `uri` is string the URI to find object store locators for.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L816-L824


---
## OSLocatorsByScope

The `OSLocatorsByScope` query gets the object store locators for the owners and value owner of a scope.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L826-L832

The `scope_id`, must either be scope uuid, e.g. `91978ba2-5f35-459a-86a7-feca1b0512e0` or a scope address,
e.g. `scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel`

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L834-L840


---
## OSAllLocators

The `OSAllLocators` query gets all object store locators.

This query is paginated.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L842-L848

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L850-L858

---
## AccountData

The `AccountData` query gets the account data associated with a scope.

### Request
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L860-L865

The `metadata_addr` must be a scope id, e.g. `scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel`.

### Response
+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/metadata/v1/query.proto#L867-L871
