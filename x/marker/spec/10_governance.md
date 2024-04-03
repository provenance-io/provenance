# Governance Proposal Control

The marker module supports an extensive amount of control over markers via governance proposal.  This allows a
marker to be defined where no single account is allowed to make modifications and yet it is still possible to
issue change requests through passing a governance proposal.

<!-- TOC 2 2 -->
  - [Add Marker Proposal](#add-marker-proposal)
  - [Supply Increase Proposal](#supply-increase-proposal)
  - [Supply Decrease Proposal](#supply-decrease-proposal)
  - [Set Administrator Proposal](#set-administrator-proposal)
  - [Remove Administrator Proposal](#remove-administrator-proposal)
  - [Change Status Proposal](#change-status-proposal)
  - [Withdraw Escrow Proposal](#withdraw-escrow-proposal)
  - [Set Denom Metadata Proposal](#set-denom-metadata-proposal)



## Add Marker Proposal

AddMarkerProposal defines a governance proposal to create a new marker.

In a typical add marker situation the `UnrestrictedDenomRegex` parameter would be used to enforce longer denom
values (preventing users from creating coins with well known symbols such as BTC, ETH, etc.).  Markers added
via governance proposal are only limited by the more generic Coin Validation Denom expression enforced by the
bank module.

A further difference from the standard add marker flow is that governance proposals to add a marker can directly
set a marker to the `Active` status with the appropriate minting operations performed immediately.

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/proposals.proto#L15-L34

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- The marker request contains an invalid denom value
- The marker already exists
- The amount of coin in circulation could not be set.
  - There is already coin in circulation [perhaps from genesis] and the configured supply is less than this amount and
    it is not possible to burn sufficient coin to make the requested supply match actual supply
- The mint operation fails for any reason (see bank module)

## Supply Increase Proposal

SupplyIncreaseProposal defines a governance proposal to administer a marker and increase total supply of the marker
through minting coin and placing it within the marker or assigning it directly to an account.

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/proposals.proto#L36-L47

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- The requested supply exceeds the configuration parameter for `MaxSupply`

## Supply Decrease Proposal

SupplyDecreaseProposal defines a governance proposal to administer a marker and decrease the total supply through
burning coin held within the marker

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/proposals.proto#L49-L59

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- Marker does not allow governance control (`AllowGovernanceControl`)
- The marker account itself is not holding sufficient supply to cover the amount of coin requested to burn
- The amount of resulting supply would be less than zero

The chain will panic and halt if:
- The bank burn operation fails for any reason (see bank module)

## Set Administrator Proposal

SetAdministratorProposal defines a governance proposal to administer a marker and set administrators with specific
access on the marker

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/proposals.proto#L61-L71

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- The marker does not exist
- Marker does not allow governance control (`AllowGovernanceControl`)
- Any of the access grants are invalid

## Remove Administrator Proposal

RemoveAdministratorProposal defines a governance proposal to administer a marker and remove all permissions for a
given address

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/proposals.proto#L73-L83

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- The marker does not exist
- Marker does not allow governance control (`AllowGovernanceControl`)
- The address to be removed is not present

## Change Status Proposal

ChangeStatusProposal defines a governance proposal to administer a marker to change its status

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/proposals.proto#L85-L94

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- Marker does not allow governance control (`AllowGovernanceControl`)
- The requested status is invalid
- The new status is not a valid transition from the current status
- For destroyed markers
  - The supply of the marker is greater than zero and the amount held by the marker account does not equal this value
    resulting in the failure to burn all remaining supply.

## Withdraw Escrow Proposal

WithdrawEscrowProposal defines a governance proposal to withdraw escrow coins from a marker

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/proposals.proto#L96-L107

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- Marker does not allow governance control (`AllowGovernanceControl`)
- The marker account is not holding sufficient assets to cover the requested withdrawal amounts.

## Set Denom Metadata Proposal

SetDenomMetadataProposal defines a governance proposal to set the metadata for a denom.

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/proposals.proto#L109-L117

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- Marker does not allow governance control (`AllowGovernanceControl`)
