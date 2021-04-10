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

#### Request
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L74-L98

The `scope_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate scope id for use in the `scope.scope_id` field.

The `spec_uuid` field is optional.
It should be a uuid formated as a string using the standard UUID format.
If supplied, it will be used to generate the appropriate scope id for use in the `scope.specification_id` field.

#### Response
+++ https://github.com/provenance-io/provenance/blob/b295b03b5584741041d8a4e19ef0a03f2300bd2f/proto/provenance/metadata/v1/tx.proto#L100-L104

#### Expected failures

This service message is expected to fail if:
TODO: WriteScope failure points.



### Msg/DeleteScope

TODO: DeleteScope summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: DeleteScope failure points.



### Msg/WriteSession

TODO: WriteSession summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: WriteSession failure points.



### Msg/WriteRecord

TODO: WriteRecord summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: WriteRecord failure points.



### Msg/DeleteRecord

TODO: DeleteRecord summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: DeleteRecord failure points.



## Specifications

### Msg/WriteScopeSpecification

TODO: WriteScopeSpecification summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: WriteScopeSpecification failure points.



### Msg/DeleteScopeSpecification

TODO: DeleteScopeSpecification summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: DeleteScopeSpecification failure points.



### Msg/WriteContractSpecification

TODO: WriteContractSpecification summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: WriteContractSpecification failure points.



### Msg/DeleteContractSpecification

TODO: DeleteContractSpecification summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: DeleteContractSpecification failure points.



### Msg/WriteRecordSpecification

TODO: WriteRecordSpecification summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: WriteRecordSpecification failure points.



### Msg/DeleteRecordSpecification

TODO: DeleteRecordSpecification summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: DeleteRecordSpecification failure points.



## Object Store Locators

### Msg/BindOSLocator

TODO: BindOSLocator summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: BindOSLocator failure points.



### Msg/DeleteOSLocator

TODO: DeleteOSLocator summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: DeleteOSLocator failure points.



### Msg/ModifyOSLocator

TODO: ModifyOSLocator summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: ModifyOSLocator failure points.

## Deprecated

These are messages associated with deprecated endpoints.
These endpoints exist only to facilitate a transition to the new models.
As such, they are sparsely documented and probably shouldn't be trusted.



### Msg/WriteP8eContractSpec

TODO: WriteP8eContractSpec summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: WriteP8eContractSpec failure points.



### Msg/P8eMemorializeContract

TODO: P8eMemorializeContract summary

#### Request
+++ 

#### Response
+++ 

#### Expected failures

This service message is expected to fail if:
TODO: P8eMemorializeContract failure points.
