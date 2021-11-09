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
  - [Authz Grants](#authz-grants)
  - [Deprecated](#deprecated)
    - [Msg/WriteP8eContractSpec](#msg-writep8econtractspec)
    - [Msg/P8eMemorializeContract](#msg-p8ememorializecontract)



---
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

---
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

---
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
* The session's scope cannot be found.
* The session's contract specification does not exist.
* A party type required by the contract specification is not in the `parties` list.
* One or more of the `owners` are not `signers`.
* The `audit` fields are changed.

---
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
* The `session_id` is missing or invalid.
* The `specification_id` is provided but invalid.
* An entry in `inputs` does not have a `name`.
* An entry in `inputs` does not have a `source`.
* An entry in `inputs` has a `source` type that doesn't match the input's `status`.
* An entry in `inputs` has a `record_id` `source` but the `record_id` is missing or invalid.
* An entry in `inputs` does not have a `type_name`.
* An entry in `outputs` has a `status` of `unspecified`.
* An entry in `outputs` has a `status` of `pass` or `fail`, and doesn't have a `hash`.
* The `name` is missing.
* The `process.method` is missing.
* The `process.name` is missing.
* The `process.process_id` is missing.
* A record is being updated and the `name` values are different.
* A record is being updated and the `session` values are different.
* A record is being updated and the `specification_id` values are different.
* The record's scope cannot be found.
* The record's session cannot be found.
* The record's contract specification cannot be found.
* The record's record specification cannot be found.
* The `parties_involved` is missing an entry required by the contract specification.
* There are duplicate `inputs` by `name`.
* An entry in `inputs` exists that is not part of the record specification.
* The `inputs` list does not contain one or more inputs defined in the record specification.
* An entry in `inputs` has a `type_name` different from its input specification.
* An entry in `inputs` has a `source` type that doesn't match the input specification.
* An entry in `inputs` has a `source` value that doesn't match the intput specification.
* The record specification has a result type of `record` but there isn't exactly one entry in `outputs`.
* The record specification has a result type of `record_list` but the `outputs` list is empty.

---
### Msg/DeleteRecord

A record is deleted using the `DeleteRecord` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L208-L222

#### Response

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L224-L225

#### Expected failures

This service message is expected to fail if:
* No record exists with the given `record_id`.
* The record's scope cannot be found.
* One or more scope `owners` are not `signers`.



---
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
* The `specificatio_id` is missing or invalid.
* The `description` has an empty `name` or the `name` is longer than 200 characters.
* The `description` has a `description` longer than 5000 characters.
* The `description` has a `website_url` or `icon_url` that is empty or longer than 2048 characters.
* The `description` has a `website_url` or `icon_url` that has a protocol other than `http`, `https`, or `data`.
* The `owners` list is empty.
* One of the entries in `owners` is not a valid bech32 address.
* The `parties_involved` list is empty.
* One of the entries in `contract_spec_ids` is invalid.
* One of the entries in `contract_spec_ids` does not exist.
* One or more `owners` of the existing scope specification are not `signers`.

---
### Msg/DeleteScopeSpecification

A scope specification is deleted using the `DeleteScopeSpecification` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L254-L268

#### Response

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L270-L271

#### Expected failures

This service message is expected to fail if:
* No scope specification exists with the given `specification_id`
* One or more `owners` are not `signers`.

---
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
* The `specification_id` is missing or invalid.
* The `description` has an empty `name` or the `name` is longer than 200 characters.
* The `description` has a `description` longer than 5000 characters.
* The `description` has a `website_url` or `icon_url` that is empty or longer than 2048 characters.
* The `description` has a `website_url` or `icon_url` that has a protocol other than `http`, `https`, or `data`.
* The `owners` list is empty.
* One of the entries in `owners` is not a valid bech32 address.
* The `parties_involved` list is empty.
* The `source` is empty.
* The `source` is a resource id, that is invalid.
* The `source` is a hash that is empty.
* The `class_name` is empty or longer than 1000 characters.
* One or more `owners` of the existing contract specification are not `signers`.

---
### Msg/DeleteContractSpecification

A contract specification is deleted using the `DeleteContractSpecification` service method.

This will also delete all record specifications associated with this contract specification.

#### Request

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L301-L315

#### Response

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L317-L318

#### Expected failures

This service message is expected to fail if:
* No contract specification exists with the given `specification_id`
* One or more `owners` are not `signers`.
* One of the record specifications associated with this contract specification cannot be deleted.

---
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
* The `specification_id` is missing or invalid.
* No contract specification exists with the given contract specification id portion of the `specification_id`.
* One or more contract specification `owners` are not `signers`.
* The `name` is longer than 200 characters.
* One of the `input_specifications` is missing a `name` or its `name` is longer than 200 characters.
* One of the `input_specifications` is missing a `type_name` or its `type_name` is longer than 1000 characters.
* One of the `input_specifications` is missing a `source`.
* One of the `input_specifications` has a `source` that is a record id that is missing or invalid.
* One of the `input_specifications` has a `source` that is a hash that is missing.
* The `type_name` is longer than 1000 characters.
* The `responsible_parties` list is empty.
* The `result_type` is unspecified.
* A record specification is being updated and the `name` values are different.
* A record specification is being updated and the `specification_id` values are different.

---
### Msg/DeleteRecordSpecification

A record specification is deleted using the `DeleteRecordSpecification` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L348-L362

#### Response

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L364-L365 

#### Expected failures

This service message is expected to fail if:
* No record specification exists with the given `specification_id`.
* No contract specification exists with the given contract specification id portion of the `specification_id`.
* One or more `owners` of the contracts specification are not `signers`.

---
## Object Store Locators

### Msg/BindOSLocator

An Object Store Locator entry is created using the `BindOSLocator` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L422-L428

#### Response

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L430-L433

#### Expected failures

This service message is expected to fail if:
* The `owner` is missing.
* The `owner` is not a valid bech32 address.
* The `uri` is empty.
* The `uri` is not a valid URI.
* The `owner` does not match an existing account.
* An object store locator already exists for the given `owner`.

---
### Msg/DeleteOSLocator

An Object Store Locator entry is deleted using the `DeleteOSLocator` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L435-L442

#### Response

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L444-L447

#### Expected failures

This service message is expected to fail if:
* The `owner` is missing.
* The `owner` is not a valid bech32 address.
* The `uri` is empty.
* The `uri` is not a valid URI.
* The `owner` does not match an existing account.
* An object store locator does not exist for the given `owner`.

---
### Msg/ModifyOSLocator

An Object Store Locator entry is updated using the `DeleteOSLocator` service method.

Object Store Locators are identified by their `owner`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L449-L455

#### Response

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L457-L460

#### Expected failures

This service message is expected to fail if:
* The `owner` is missing.
* The `owner` is not a valid bech32 address.
* The `uri` is empty.
* The `uri` is not a valid URI.
* The `owner` does not match an existing account.
* An object store locator does not exist for the given `owner`.

---
## Authz Grants

Authz requires the use of fully qualified message type URLs when applying grants to an address. See [04_authz.md](04_authz.md) for more details.

Fully qualified `metadata` message type URLs:
- `/provenance.metadata.v1.MsgWriteScopeRequest`
- `/provenance.metadata.v1.MsgDeleteScopeRequest`
- `/provenance.metadata.v1.MsgAddScopeDataAccessRequest`
- `/provenance.metadata.v1.MsgDeleteScopeDataAccessRequest`
- `/provenance.metadata.v1.MsgAddScopeOwnerRequest`
- `/provenance.metadata.v1.MsgDeleteScopeOwnerRequest`
- `/provenance.metadata.v1.MsgWriteSessionRequest`
- `/provenance.metadata.v1.MsgWriteRecordRequest`
- `/provenance.metadata.v1.MsgDeleteRecordRequest`
- `/provenance.metadata.v1.MsgWriteScopeSpecificationRequest`
- `/provenance.metadata.v1.MsgDeleteScopeSpecificationRequest`
- `/provenance.metadata.v1.MsgWriteContractSpecificationRequest`
- `/provenance.metadata.v1.MsgDeleteContractSpecificationRequest`
- `/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest`
- `/provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecRequest`
- `/provenance.metadata.v1.MsgWriteRecordSpecificationRequest`
- `/provenance.metadata.v1.MsgDeleteRecordSpecificationRequest`
- `/provenance.metadata.v1.MsgBindOSLocatorRequest`
- `/provenance.metadata.v1.MsgDeleteOSLocatorRequest`
- `/provenance.metadata.v1.MsgModifyOSLocatorRequest`
- `/provenance.metadata.v1.MsgWriteP8eContractSpecRequest`
- `/provenance.metadata.v1.MsgP8eMemorializeContractRequest`

---
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
* The converted contract specification meets one of the failure criteria for [contract specifications](#msg-writecontractspecification).
* One of the converted record specifications meets one of the failure criteria for [record specifications](#msg-writerecordspecification).

---
### Msg/P8eMemorializeContract

The `P8eMemorializeContract` service endpoint converts in an old contract message structure into the new stuff.
Then it either creates or updates a scope, session, and records.

#### Request

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L389-L410

#### Response

+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L412-L420

#### Expected failures

This service message is expected to fail if:
* The converted scope meets one of the failure criteria for [scopes](#msg-writescope).
* The converted session meets one of the failure criteria for [sessions](#msg-writesession).
* One of the converted records meets one of the failure criteria for [records](#msg-writerecord).
