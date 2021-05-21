# Messages

In this section we describe the processing of the marker messages and the corresponding updates to the state.
All created/modified state objects specified by each message are defined within the
[state](./02_state_transitions.md) section.

<!-- TOC 2 2 -->
  - [Msg/AddMarkerRequest](#msg-addmarkerrequest)
  - [Msg/AddAccessRequest](#msg-addaccessrequest)
  - [Msg/DeleteAccessRequest](#msg-deleteaccessrequest)
  - [Msg/FinalizeRequest](#msg-finalizerequest)
  - [Msg/ActivateRequest](#msg-activaterequest)
  - [Msg/CancelRequest](#msg-cancelrequest)
  - [Msg/DeleteRequest](#msg-deleterequest)
  - [Msg/MintRequest](#msg-mintrequest)
  - [Msg/BurnRequest](#msg-burnrequest)
  - [Msg/WithdrawRequest](#msg-withdrawrequest)
  - [Msg/TransferRequest](#msg-transferrequest)
  - [Msg/SetDenomMetadataRequest](#msg-setdenommetadatarequest)



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
only be used against markers in the `Pending` status when called by the current marker manager address or against `Finalized`
and `Active` markers when the caller is currently assigned the `Admin` access type.

## Msg/DeleteAccessRequest

DeleteAccess Request defines the Msg/DeleteAccess request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L70-L74

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L76

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not pending or:
  - The request is not signed with an administrator address that matches the manager address or:
  - The given administrator address does not currently have the "admin" access granted on the marker

The Delete Access request will remove all access granted to the given address on the specified marker.  The method may
only be used against markers in the `Pending` status when called by the current marker manager address or against `Finalized`
and `Active` markers when the caller is currently assigned the `Admin` access type.

## Msg/FinalizeRequest

Finalize Request defines the Msg/Finalize request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L79-L82

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L84

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `proposed` status or:
  - The request is not signed with an administrator address that matches the manager address or:
  - The given administrator address does not currently have the "admin" access granted on the marker

The `Finalize` marker status performs a set of checks to ensure the marker is ready to be activated.  It is designed to
serve as an intermediate step prior to activation that indicates marker configuration is complete.

## Msg/ActivateRequest

Activate Request defines the Msg/Activate request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L87-L90

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L92

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `Finalized` status or:
  - The request is not signed with an administrator address that matches the manager address or:
  - The given administrator address does not currently have the "admin" access granted on the marker
- The marker has a supply less than the current in circulation supply (for markers created against existing coin)

The Activate marker request will mint any coin required to achieve a circulation target set by the total supply.  In
addition the marker will no longer be managed by an indicated "manager" account but will instead require explicit
rights assigned as access grants for any modification.

If a marker has a fixed supply the begin block/invariant supply checks are also performed.  If the supply is expected to
float then the `total_supply` value will be set to zero upon activation.

## Msg/CancelRequest

Cancel Request defines the Msg/Cancel request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L95-L98

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L100

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `Pending` or `Active` status
- If marker is in a `Pending` status and:
  - The given administrator address does not currently have the "admin" access granted on the marker
  - Or given administrator is not listed as the manager on the marker
- If marker is in a `Active` status and:
  - The given administrator address does not currently have the "admin" access granted on the marker
- The amount in circulation is greater than zero or any remaining amount is not currently held in escrow within the
  marker account.

## Msg/DeleteRequest

Delete Request defines the Msg/Delete request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L103-L106

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L108

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `Cancelled` status
- The given administrator address does not currently have the "admin" access granted on the marker or:
  - If the marker was previously in a `Proposed` status when cancelled the administrator must be the marker manager.
- The amount in circulation is greater than zero or any remaining amount is not currently held in escrow within the
  marker account.
- There are any other coins remaining in escrow after supply has been fully burned.

## Msg/MintRequest

Mint Request defines the Msg/Mint request type

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L111-L115

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L117

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `Active` status or:
  - The request is not signed with an administrator address that matches the manager address or:
- The given administrator address does not currently have the "mint" access granted on the marker
- The requested amount of mint would increase the total supply in circulation above the configured supply limit set in
  the marker module params

## Msg/BurnRequest

Burn Request defines the Msg/Burn request type that is used to remove supply of the marker coin from circulation.  In
order to successfully burn supply the amount to burn must be held by the marker account itself (in escrow).

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L120-L124

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L126

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in an `Active` status or:
  - The request is not signed with an administrator address that matches the manager address or:
- The given administrator address does not currently have the "burn" access granted on the marker
- The amount of coin to burn is not currently held in escrow within the marker account.

## Msg/WithdrawRequest

Withdraw Request defines the Msg/Withdraw request type and is used to withdraw coin from escrow within the marker.

NOTE: any denom coin can be held within a marker "in escrow", these values are not restricted to just the denom of the
marker itself.

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L129-L135

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L137

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- If marker is not in a `Active` status:
  - The request is not signed with an administrator address that matches the manager address
  - For `Pending` status: the denom being withdrawn from the marker matches the marker denom
- If the marker is `Active`, `Cancelled`
 - The given administrator address does not currently have the "withdraw" access granted on the marker
- The amount of coin requested for withdraw is not currently held by the marker account

## Msg/TransferRequest

Transfer Request defines the Msg/Transfer request type.  A transfer request is used to transfer coin between two
accounts for `RESTRICTED_COIN` type markers that have `send_enabled=false` configured with the `bank` module and thus
can not be sent using a normal `send_coin` operation.  A transfer request requires a signature from an account with
the transfer permission as well as approval from the account the funds will be withdrawn from.

NOTE: the withdraw approval has been suspended pending integration with the `auth` module.  See [Issue #262](https://github.com/provenance-io/provenance/issues/262)

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L140-L146

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L149

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `Active` status or:
  - The given administrator address does not currently have the "transfer" access granted on the marker
  - The marker types is not `RESTRICTED_COIN`

## Msg/SetDenomMetadataRequest

SetDenomMetadata Request defines the Msg/SetDenomMetadata request type.  This request is used to set the informational
denom metadata held within the bank module.  Denom metadata can be used to provide a more streamlined user experience
within block explorers or similar applications.

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L152-L156

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/tx.proto#L159

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The request is not signed with an administrator address that matches the manager address or:
- The given administrator address does not currently have the "admin" access granted on the marker
- Any of the provided display denoms is found to be invalid
  - Does not match the proper form with an SI unit prefix matching the associated exponent
  - Is missing the denom unit for the indicated base denom or display denom unit.
  - If there is an existing record the update will fail if:
     - The Base denom is changed.
       If marker status is `Active` or `Finalized`:
        - Any DenomUnit entries are removed.
        - DenomUnit Denom fields are modified.
        - Any aliases are removed from a DenomUnit.
