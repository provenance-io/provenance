# Metadata Messages

In this section we describe the processing of the metadata messages and the corresponding updates to the state.
All created/modified state objects specified by each message are defined within the [state](02_state.md) section.

## Msg/WriteScope
TODO: WriteScope messages

## Msg/DeleteScope
TODO: DeleteScope messages

## Msg/WriteSession
TODO: WriteSession messages

## Msg/WriteRecord
TODO: WriteRecord messages

## Msg/DeleteRecord
TODO: DeleteRecord messages

## Msg/WriteScopeSpecification
TODO: WriteScopeSpecification messages

## Msg/DeleteScopeSpecification
TODO: DeleteScopeSpecification messages

## Msg/WriteContractSpecification
TODO: WriteContractSpecification messages

## Msg/DeleteContractSpecification
TODO: DeleteContractSpecification messages

## Msg/WriteRecordSpecification
TODO: WriteRecordSpecification messages

## Msg/DeleteRecordSpecification
TODO: DeleteRecordSpecification messages

## Msg/BindOSLocator
TODO: BindOSLocator messages

## Msg/DeleteOSLocator
TODO: DeleteOSLocator messages

## Msg/ModifyOSLocator
TODO: ModifyOSLocator messages

## Deprecated Endpoints

These are messages associated with deprecated endpoints.
These endpoints exist only to facilitate a transition to the new models.
As such, they are sparsely documented a probably shouldn't be trusted. 

### Msg/WriteP8eContractSpec
TODO: WriteP8eContractSpec messages

### Msg/P8eMemorializeContract
TODO: P8eMemorializeContract messages






<!-- This was given in slack as example formatting:
# Messages

In this section we describe the processing of the marker messages and the corresponding updates to the state.
All created/modified state objects specified by each message are defined within the 
[state](./02_state_transitions.md) section.

## Msg/AddMarkerRequest

A marker is created using the Add Marker service message.
The created marker can not be directly added in an Active (or Cancelled/Destroyed) status.  Markers
must have a valid supply and denomination value

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L44-L54

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L57
-->
