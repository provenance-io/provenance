# Messages

In this section we describe the processing of the marker messages and the corresponding updates to the state.
All created/modified state objects specified by each message are defined within the 
[state](./02_state_transitions.md) section.


## Msg/AddMarkerRequest

A marker is created using the Add Marker service message.
The created marker can not be directly added in an Active (or Cancelled/Destroyed) status.  Markers
must have a valid supply and denomination value.

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L44-L54

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L57

This service message is expected to fail if:
- The Denom string:
  - Is already in use by another marker
  - Does not conform to the "Marker Denom Validation Expression"
  - Does not conform to the base coin denom validation expression parameter
- The supply value:
  - Is less than zero
  - Is greater than the "max supply" parameter
- The Marker Status:
  - Is Active (markers can not be created as active the must transition from Finalized)
  - Is Cancelled
  - Is Destroyed
- The manager address is invalid. (Note: an empty manager address will be set to the Msg from address)

The service message will create a marker account object and request the auth module persist it.  No coin will be minted
or disbursed as a result of adding a marker using this endpoint.

## Msg/AddAccessRequest

Add Access Request is used to add permissions to a marker that allow the specified accounts to perform the specified actions.

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L60-L64

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L67

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is pending:
  - And the request is not signed with an administrator address that matches the manager address or:
  - The given administrator address does not currently have the "admin" access granted on the marker
- The accesslist:
  - Contains more than one entry for a given address
  - Contains a grant with an invalid address
  - Contains a grant with an invalid access enum value (Unspecified/0)

The Add Access request can be called many times on a marker with some or all of the access grant values.  The method may
only be used against markers in the Pending status when called by the current marker manager address or against `Finalized`
and `Active` markers when the caller is currently assigned the `Admin` access type.

## Msg/DeleteAccessRequest

DeleteAccess Request defines the Msg/DeleteAccess request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L70-L74

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L76

This service message is expected to fail if:

## Msg/FinalizeRequest

Finalize Request defines the Msg/Finalize request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L79-L82

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L84

This service message is expected to fail if:

## Msg/ActivateRequest

Activate Request defines the Msg/Activate request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L87-L90

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L92

This service message is expected to fail if:

## Msg/CancelRequest

Cancel Request defines the Msg/Cancel request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L95-L98

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L100

This service message is expected to fail if:

## Msg/DeleteRequest

Delete Request defines the Msg/Delete request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L103-L106

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L108

This service message is expected to fail if:

## Msg/MintRequest

Mint Request defines the Msg/Mint request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L111-L115

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L117

This service message is expected to fail if:

## Msg/BurnRequest

Burn Request defines the Msg/Burn request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L120-L124

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L126
This service message is expected to fail if:

## Msg/WithdrawRequest

Withdraw Request defines the Msg/Withdraw request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L129-L135

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L137

This service message is expected to fail if:

## Msg/TransferRequest

Transfer Request defines the Msg/Transfer request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L140-L146

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L149

This service message is expected to fail if:

## Msg/SetDenomMetadataRequest

SetDenomMetadata Request defines the Msg/SetDenomMetadata request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L152-L156

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L159

This service message is expected to fail if:
