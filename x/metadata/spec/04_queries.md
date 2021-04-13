# Metadata Queries

In this section we describe the queries available for looking up metadata information.
All state objects specified by each message are defined within the [state](02_state.md) section.

Each entry or specification state object is wrapped with an `*IdInfo` message containing information about that state object's address/id.

<!-- TOC 2 -->
  - [Query/Params](#query-params)
  - [Query/Scope](#query-scope)
  - [Query/ScopesAll](#query-scopesall)
  - [Query/Sessions](#query-sessions)
  - [Query/SessionsAll](#query-sessionsall)
  - [Query/Records](#query-records)
  - [Query/RecordsAll](#query-recordsall)
  - [Query/Ownership](#query-ownership)
  - [Query/ValueOwnership](#query-valueownership)
  - [Query/ScopeSpecification](#query-scopespecification)
  - [Query/ScopeSpecificationsAll](#query-scopespecificationsall)
  - [Query/ContractSpecification](#query-contractspecification)
  - [Query/ContractSpecificationsAll](#query-contractspecificationsall)
  - [Query/RecordSpecificationsForContractSpecification](#query-recordspecificationsforcontractspecification)
  - [Query/RecordSpecification](#query-recordspecification)
  - [Query/RecordSpecificationsAll](#query-recordspecificationsall)
  - [Query/OSLocatorParams](#query-oslocatorparams)
  - [Query/OSLocator](#query-oslocator)
  - [Query/OSLocatorsByURI](#query-oslocatorsbyuri)
  - [Query/OSLocatorsByScope](#query-oslocatorsbyscope)
  - [Query/OSAllLocators](#query-osalllocators)



## Query/Params

The `Params` query gets the parameters of the metadata module.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L223-L224

There are no inputs for this query.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L226-L233



## Query/Scope

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



## Query/ScopesAll

The `ScopesAll` query gets all scopes in a paginated manner.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L275-L279

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L281-L290



## Query/Sessions

The `Sessions` query gets sessions.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L292-L306

The `scope_id` can either be scope uuid, e.g. `91978ba2-5f35-459a-86a7-feca1b0512e0` or a scope address, e.g.
`scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel`. Similarly, the `session_id` can either be a uuid or session address, e.g.
`session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr`.

* If only a `scope_id` is provided, all sessions in that scope are returned.
* If only a `session_id` is provided, it must be an address, and that single session is returned.
* If both are provided, that single session is returned.
* If both `scope_id` and `session_id` are addresses, and they don't refer to the same scope, a bad request is returned.

By default, the scope and records are not included.
Set `include_scope` and/or `include_records` to true to include the scope and/or records.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L308-L319



## Query/SessionsAll

The `SessionsAll` query gets all sessions in a paginated manner.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L331-L335

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L337-L346



## Query/Records

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



## Query/RecordsAll

The `RecordsAll` query gets all records in a paginated manner.

### Request
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L391-L395

The only input to this query is pagination information.

### Response
+++ https://github.com/provenance-io/provenance/blob/995c8f6e73eca5f63ebc85b27df6a1c6bdd43e10/proto/provenance/metadata/v1/query.proto#L397-L406



## Query/Ownership
TODO: Ownership messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/ValueOwnership
TODO: ValueOwnership messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/ScopeSpecification
TODO: ScopeSpecification messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/ScopeSpecificationsAll
TODO: ScopeSpecificationsAll messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/ContractSpecification
TODO: ContractSpecification messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/ContractSpecificationsAll
TODO: ContractSpecificationsAll messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/RecordSpecificationsForContractSpecification
TODO: RecordSpecificationsForContractSpecification messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/RecordSpecification
TODO: RecordSpecification messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/RecordSpecificationsAll
TODO: RecordSpecificationsAll messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/OSLocatorParams
TODO: OSLocatorParams messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/OSLocator
TODO: OSLocator messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/OSLocatorsByURI
TODO: OSLocatorsByURI messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/OSLocatorsByScope
TODO: OSLocatorsByScope messages
The `xxx` query gets xxxxxxxxxxx

### Request
+++ 

Info about the query request fields.

### Response
+++ 



## Query/OSAllLocators
TODO: OSAllLocators messages
