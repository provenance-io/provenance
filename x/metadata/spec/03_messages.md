# Metadata Messages

In this section we describe the processing of the metadata messages and the corresponding updates to the state.
All created/modified state objects specified by each message are defined within the [state](02_state.md) section.

These endpoints, requests, and responses are defined in [tx.proto](https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto).

<!-- TOC -->
  - [Entries](#entries)
    - [Msg/WriteScope](#msgwritescope)
    - [Msg/DeleteScope](#msgdeletescope)
    - [Msg/AddScopeDataAccess](#msgaddscopedataaccess)
    - [Msg/DeleteScopeDataAccess](#msgdeletescopedataaccess)
    - [Msg/AddScopeOwner](#msgaddscopeowner)
    - [Msg/DeleteScopeOwner](#msgdeletescopeowner)
    - [Msg/UpdateValueOwners](#msgupdatevalueowners)
    - [Msg/MigrateValueOwner](#msgmigratevalueowner)
    - [Msg/WriteSession](#msgwritesession)
    - [Msg/WriteRecord](#msgwriterecord)
    - [Msg/DeleteRecord](#msgdeleterecord)
  - [Specifications](#specifications)
    - [Msg/WriteScopeSpecification](#msgwritescopespecification)
    - [Msg/DeleteScopeSpecification](#msgdeletescopespecification)
    - [Msg/WriteContractSpecification](#msgwritecontractspecification)
    - [Msg/DeleteContractSpecification](#msgdeletecontractspecification)
    - [Msg/AddContractSpecToScopeSpec](#msgaddcontractspectoscopespec)
    - [Msg/DeleteContractSpecFromScopeSpec](#msgdeletecontractspecfromscopespec)
    - [Msg/WriteRecordSpecification](#msgwriterecordspecification)
    - [Msg/DeleteRecordSpecification](#msgdeleterecordspecification)
  - [Object Store Locators](#object-store-locators)
    - [Msg/BindOSLocator](#msgbindoslocator)
    - [Msg/DeleteOSLocator](#msgdeleteoslocator)
    - [Msg/ModifyOSLocator](#msgmodifyoslocator)
  - [Account Data](#account-data)
    - [Msg/SetAccountData](#msgsetaccountdata)
  - [Authz Grants](#authz-grants)



---
## Entries

### Msg/WriteScope

A scope is created or updated using the `WriteScope` service method.

Scopes are identified using their `scope_id`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L76-L99

The `scope_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate scope id for use in the `scope.scope_id` field.

The `spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate scope specification id for use in the `scope.specification_id` field.

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L101-L105

#### Expected failures

This service message is expected to fail if:
* The `scope_id` is missing or invalid.
* The `specification_id` is missing or invalid.
* The `owners` list is empty.
* Any of the owner `address` values aren't bech32 address strings.
* Any of the `data_access` values aren't bech32 address strings.
* A `value_owner_address` is provided that isn't a bech32 address string.
* The `signers` do not have permission to write the scope.

---
### Msg/DeleteScope

A scope is deleted using the `DeleteScope` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L107-L120

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L122-L123

#### Expected failures

This service message is expected to fail if:
* No scope exists with the given `scope_id`.
* The `signers` do not have permission to delete the scope.

---
### Msg/AddScopeDataAccess

Addresses can be added to a scope's data access list using the `AddScopeDataAccess` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L125-L142

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L144-L145

#### Expected failures

This service message is expected to fail if:
* Any provided address is invalid.
* Any provided address is already in the scope's data access list.
* The `signers` do not have permission to update the scope.

---
### Msg/DeleteScopeDataAccess

Addresses can be deleted from a scope's data access list using the `DeleteScopeDataAccess` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L147-L164

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L166-L167

#### Expected failures

This service message is expected to fail if:
* Any provided address is not already in the scope's data access list.
* The `signers` do not have permission to update the scope.

---
### Msg/AddScopeOwner

Scope owners can be added to a scope using the `AddScopeOwner` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L169-L186

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L188-L189

#### Expected failures

This service message is expected to fail if:
* Any new party is invalid.
* An `optional = true` party is being added to a `require_party_rollup = false` scope.
* The `signers` do not have permission to update the scope.

---
### Msg/DeleteScopeOwner

Scope owners can be deleted from a scope using the `DeleteScopeOwner` service method.
All owner parties with any of the provided addresses will be removed from the scope.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L191-L208

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L210-L211

#### Expected failures

This service message is expected to fail if:
* Any provided `owners` (addresses) are not an address in a party in the scope.
* The resulting scope owners do not meet scope specification requirements.
* The `signers` do not have permission to update the scope.

---
### Msg/UpdateValueOwners

The value owner address of one or more scopes can be updated using the `UpdateValueOwners` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/37cdb0c84db7b2f91aef057a606c5ba6aece06a1/proto/provenance/metadata/v1/tx.proto#L219-L235

#### Response

+++ https://github.com/provenance-io/provenance/blob/37cdb0c84db7b2f91aef057a606c5ba6aece06a1/proto/provenance/metadata/v1/tx.proto#L237-L238

#### Expected failures

This service message is expected to fail if:
* The new value owner address is invalid.
* Any of the provided scope ids are not metadata scope identifiers or do not exist.
* The signers are not allowed to update the value owner address of a provided scope.

---
### Msg/MigrateValueOwner

All scopes with a given existing value owner address can be updated to have a new proposed value owner address using the `MigrateValueOwner` endpoint.

#### Request

+++ https://github.com/provenance-io/provenance/blob/37cdb0c84db7b2f91aef057a606c5ba6aece06a1/proto/provenance/metadata/v1/tx.proto#L240-L252

#### Response

+++ https://github.com/provenance-io/provenance/blob/37cdb0c84db7b2f91aef057a606c5ba6aece06a1/proto/provenance/metadata/v1/tx.proto#L254-L255

#### Expected failures

This service message is expected to fail if:
* Either the existing or proposed values are not valid bech32 addresses.
* The existing address is not a value owner on any scopes.
* The signers are not allowed to update the value owner address of a scope being updated.

---
### Msg/WriteSession

A session is created or updated using the `WriteSession` service method.

Sessions are identified using their `session_id`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L213-L238

The `session_id_components` field is optional.
If supplied, it will be used to generate the appropriate session id for use in the `session.session_id` field.

The `spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate contract specification id for use in the `session.specification_id` field.

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L253-L257

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
* The `signers` do not have permission to write the session.
* The `audit` fields are changed.

---
### Msg/WriteRecord

A record is created or updated using the `WriteRecord` service method.

Records are identified using their `name` and `session_id`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L260-L289

The `session_id_components` field is optional.
If supplied, it will be used to generate the appropriate session id for use in the `record.session_id` field.

The `contract_spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used with `record.name` to generate the appropriate record specification id for use in the `record.specification_id` field.

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L291-L295

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
* There are duplicate `inputs` by `name`.
* An entry in `inputs` exists that is not part of the record specification.
* The `inputs` list does not contain one or more inputs defined in the record specification.
* An entry in `inputs` has a `type_name` different from its input specification.
* An entry in `inputs` has a `source` type that doesn't match the input specification.
* An entry in `inputs` has a `source` value that doesn't match the intput specification.
* The record specification has a result type of `record` but there isn't exactly one entry in `outputs`.
* The record specification has a result type of `record_list` but the `outputs` list is empty.
* The `signers` do not have permission to write the record.

---
### Msg/DeleteRecord

A record is deleted using the `DeleteRecord` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L297-L310

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L312-L313

#### Expected failures

This service message is expected to fail if:
* No record exists with the given `record_id`.
* The `signers` do not have permission to delete the record.



---
## Specifications

### Msg/WriteScopeSpecification

A scope specification is created or updated using the `WriteScopeSpecification` service method.

Scope specifications are identified using their `specification_id`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L315-L333

The `spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate scope specification id for use in the `specification.specification_id` field.

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L335-L339

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

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L341-L354

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L356-L357

#### Expected failures

This service message is expected to fail if:
* No scope specification exists with the given `specification_id`
* One or more `owners` are not `signers`.

---
### Msg/WriteContractSpecification

A contract specification is created or updated using the `WriteContractSpecification` service method.

Contract specifications are identified using their `specification_id`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L359-L377

The `spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate contract specification id for use in the `specification.specification_id` field.

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L379-L384

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

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L437-L450

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L452-L453

#### Expected failures

This service message is expected to fail if:
* No contract specification exists with the given `specification_id`
* One or more `owners` are not `signers`.
* One of the record specifications associated with this contract specification cannot be deleted.

---
### Msg/AddContractSpecToScopeSpec

A contract specification can be added to a scope specification using the `AddContractSpecToScopeSpec` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L386-L406

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L408-L409

#### Expected failures

This service message is expected to fail if:
* The `contract_specification_id` is missing or invalid.
* The `scope_specification_id` is missing or invalid.
* The contract specification does not exist.
* The scope specification does not exist.
* * The contract specification is already allowed in the provided scope specification.
* One or more of the scope specification `owners` are not `signers`.

---
### Msg/DeleteContractSpecFromScopeSpec

A contract specification can be removed from a scope specification using the `AddContractSpecToScopeSpec` service method.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L411-L431

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L433-L435

#### Expected failures

This service message is expected to fail if:
* The `contract_specification_id` is missing or invalid.
* The `scope_specification_id` is missing or invalid.
* The scope specification does not exist.
* The contract specification is not already allowed in the provided scope specification.
* One or more of the scope specification `owners` are not `signers`.

---
### Msg/WriteRecordSpecification

A record specification is created or updated using the `WriteRecordSpecification` service method.

Record specifications are identified using their `specification_id`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L455-L473

The `contract_spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used with the `specification.name` to generate the appropriate record specification id for use in the `specification.specification_id` field.

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L475-L480

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

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L482-L495

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L497-L498 

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

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L500-L506

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L508-L511

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

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L513-L520

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L522-L525

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

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L527-L533

#### Response

+++ https://github.com/provenance-io/provenance/blob/812cb97c77036b8df59e10845fa8a04f4ba84c43/proto/provenance/metadata/v1/tx.proto#L535-L538

#### Expected failures

This service message is expected to fail if:
* The `owner` is missing.
* The `owner` is not a valid bech32 address.
* The `uri` is empty.
* The `uri` is not a valid URI.
* The `owner` does not match an existing account.
* An object store locator does not exist for the given `owner`.

---
## Account Data

### Msg/SetAccountData

Simple data (a string) can be associated with scopes using the `SetAccountData` service method.

+++ https://github.com/provenance-io/provenance/blob/e83f1955cba07e2ba87790c4487d22632ae9e69c/proto/provenance/metadata/v1/tx.proto#L589-L606

+++ https://github.com/provenance-io/provenance/blob/e83f1955cba07e2ba87790c4487d22632ae9e69c/proto/provenance/metadata/v1/tx.proto#L608-L609

This service message is expected to fail if:
* The provided address is not a scope id.
* The provided scope id does not exist.
* The signers do not have authority to update the entry.
* The provided value is too long (as defined by the attribute module params).

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
- `/provenance.metadata.v1.MsgUpdateValueOwnersRequest`
- `/provenance.metadata.v1.MsgMigrateValueOwnerRequest`
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
- `/provenance.metadata.v1.MsgSetAccountDataRequest`
