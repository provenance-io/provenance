# Metadata Messages

In this section we describe the processing of the metadata messages and the corresponding updates to the state.
All created/modified state objects specified by each message are defined within the [state](02_state.md) section.

These endpoints, requests, and responses are defined in [tx.proto](https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto).

<!-- TOC -->
  - [Entries](#entries)
    - [Msg/WriteScope](#msg-writescope)
    - [Msg/DeleteScope](#msg-deletescope)
    - [Msg/WriteSession](#msg-writesession)
    - [Msg/WriteRecord](#msg-writerecord)
    - [Msg/DeleteRecord](#msg-deleterecord)
  - [Specifications](#specifications)
    - [Msg/WriteScopeSpecification](#msg-writescopespecification)
    - [Msg/DeleteScopeSpecification](#msg-deletescopespecification)
    - [Msg/WriteContractSpecification](#msg-writecontractspecification)
    - [Msg/DeleteContractSpecification](#msg-deletecontractspecification)
    - [Msg/WriteRecordSpecification](#msg-writerecordspecification)
    - [Msg/DeleteRecordSpecification](#msg-deleterecordspecification)
  - [Object Store Locators](#object-store-locators)
    - [Msg/BindOSLocator](#msg-bindoslocator)
    - [Msg/DeleteOSLocator](#msg-deleteoslocator)
    - [Msg/ModifyOSLocator](#msg-modifyoslocator)
  - [Deprecated](#deprecated)
    - [Msg/WriteP8eContractSpec](#msg-writep8econtractspec)
    - [Msg/P8eMemorializeContract](#msg-p8ememorializecontract)



## Entries

### Msg/WriteScope

A scope is created or updated using the `WriteScope` service method.

Scopes are identified using their `scope_id`.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L74-L98

The `scope_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate scope id for use in the `scope.scope_id` field.

The `spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate scope specification id for use in the `scope.specification_id` field.

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L100-L104

#### Expected failures

This service message is expected to fail if:
* The `scope_id` is missing or invalid.
* The `specification_id` is missing or invalid.
* The `owners` list is empty.
* Any of the owner `address` values aren't bech32 address strings.
* Any of the `data_access` values aren't bech32 address strings.
* A `value_owner_address` is provided that isn't a bech32 address string.
* One or more `owners` are not `signers`.
* The `value_owner` is changing, and the existing value owner is a marker, but none of the signers have `withdraw` access.
* The `value_owner` is changing, and the existing value owner is not a marker, and is also not in `signers`.
* The `value_owner` is changing, and the proposed value owner is a marker, but none of the signers have `deposit` access.


### Msg/DeleteScope

A scope is deleted using the `DeleteScope` service method.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L106-L120

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L122-L123

#### Expected failures

This service message is expected to fail if:
* No scope exists with the given `scope_id`.
* One or more `owners` are not `signers`.



### Msg/WriteSession

A session is created or updated using the `WriteSession` service method.

Sessions are identified using their `session_id`.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L125-L151

The `session_id_components` field is optional.
If supplied, it will be used to generate the appropriate session id for use in the `session.session_id` field.

The `spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate contract specification id for use in the `session.specification_id` field.

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L166-L170

#### Expected failures

This service message is expected to fail if:
* The `session_id` is missing or invalid.
* The `specification_id` is missing or invalid.
* The `parties` list is empty.
* Any of the `parties` have an `address` that isn't a bech32 address string.
* Any of the `parties` have a `role` of `unspecified`.
* The `audit.message` string is longer than 200 characters.
* The `specification_id` is being changed.
* The session is being updated, but no `name` is provided.
* The session's scope does not exist.
* The session's contract specification does not exist.
* A party type required by the contract specification is not in the `parties` list.
* One or more of the `owners` are not `signers`.
* The `audit` fields are changed.


### Msg/WriteRecord

A record is created or updated using the `WriteRecord` service method.

Records are identified using their `name` and `session_id`.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L172-L200

The `session_id_components` field is optional.
If supplied, it will be used to generate the appropriate session id for use in the `record.session_id` field.

The `contract_spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used with `record.name` to generate the appropriate record specification id for use in the `record.specification_id` field.

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L202-L206

#### Expected failures

This service message is expected to fail if:
TODO: WriteRecord failure points.



### Msg/DeleteRecord

A record is deleted using the `DeleteRecord` service method.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L208-L222

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L224-L225

#### Expected failures

This service message is expected to fail if:
TODO: DeleteRecord failure points.



## Specifications

### Msg/WriteScopeSpecification

A scope specification is created or updated using the `WriteScopeSpecification` service method.

Scope specifications are identified using their `specification_id`.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L227-L246

The `spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate scope specification id for use in the `specification.specification_id` field.

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L248-L252

#### Expected failures

This service message is expected to fail if:
TODO: WriteScopeSpecification failure points.



### Msg/DeleteScopeSpecification

A scope specification is deleted using the `DeleteScopeSpecification` service method.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L254-L268

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L270-L271

#### Expected failures

This service message is expected to fail if:
TODO: DeleteScopeSpecification failure points.



### Msg/WriteContractSpecification

A contract specification is created or updated using the `WriteContractSpecification` service method.

Contract specifications are identified using their `specification_id`.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L273-L292

The `spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate contract specification id for use in the `specification.specification_id` field.

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L294-L299

#### Expected failures

This service message is expected to fail if:
TODO: WriteContractSpecification failure points.



### Msg/DeleteContractSpecification

A contract specification is deleted using the `DeleteContractSpecification` service method.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L301-L315

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L317-L318

#### Expected failures

This service message is expected to fail if:
TODO: DeleteContractSpecification failure points.



### Msg/WriteRecordSpecification

A record specification is created or updated using the `WriteRecordSpecification` service method.

Record specifications are identified using their `specification_id`.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L320-L339

The `contract_spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used with the `specification.name` to generate the appropriate record specification id for use in the `specification.specification_id` field.

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L341-L346

#### Expected failures

This service message is expected to fail if:
TODO: WriteRecordSpecification failure points.



### Msg/DeleteRecordSpecification

A record specification is deleted using the `DeleteRecordSpecification` service method.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L348-L362

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L364-L365 

#### Expected failures

This service message is expected to fail if:
TODO: DeleteRecordSpecification failure points.



## Object Store Locators

### Msg/BindOSLocator

An Object Store Locator entry is created using the `BindOSLocator` service method.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L422-L428

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L430-L433

#### Expected failures

This service message is expected to fail if:
TODO: BindOSLocator failure points.



### Msg/DeleteOSLocator

An Object Store Locator entry is deleted using the `DeleteOSLocator` service method.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L435-L442

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L444-L447

#### Expected failures

This service message is expected to fail if:
TODO: DeleteOSLocator failure points.



### Msg/ModifyOSLocator

An Object Store Locator entry is updated using the `DeleteOSLocator` service method.

Object Store Locators are identified by their `owner`.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L449-L455

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L457-L460

#### Expected failures

This service message is expected to fail if:
TODO: ModifyOSLocator failure points.




## Deprecated

These are messages associated with deprecated endpoints.
These endpoints exist only to facilitate a transition to the new models.
As such, they are sparsely documented and probably shouldn't be trusted.

### Msg/WriteP8eContractSpec

The `WriteP8eContractSpec` service method converts an old contract specification message structure into the new stuff.
Then it either creates or updates the provided contract specification and record specifications.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L367-L377

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L379-L387

#### Expected failures

This service message is expected to fail if:
TODO: WriteP8eContractSpec failure points.



### Msg/P8eMemorializeContract

The `P8eMemorializeContract` service endpoint converts in an old contract message structure into the new stuff.
Then it either creates or updates a scope, session, and records.

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L389-L410

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L412-L420

#### Expected failures

This service message is expected to fail if:
TODO: P8eMemorializeContract failure points.
