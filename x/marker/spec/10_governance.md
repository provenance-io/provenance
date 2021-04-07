
# Governance Proposal Control

The marker module supports an extensive amount of control over markers via governance proposal.  This allows a 
marker to be defined where no single account is allowed to make modifications and yet it is still possible to
issue change requests through passing a governance proposal.

## Add Marker Proposal
AddMarkerProposal defines defines a governance proposal to create a new marker.

In a typical add marker situation the `UnrestrictedDenomRegex` parameter would be used to enforce longer denom
values (preventing users from creating coins with well known symbols such as BTC, ETH, etc).  Markers added
via governance proposal are only limited by the more generic Coin Validation Denom expression enforced by the
bank module.

A further difference from the standard add marker flow is that governance proposals to add a marker can directly
set a marker to the `Active` status with the appropriate minting operations performed immediately.

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/proposals.proto#L15-L30

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- The marker request contains an invalid denom value
- The marker already exists
- The amount of coin in circulation could not be set.
  - There is already coin in circulation [perhaps from genesis] and the configured supply is less than this amount and
    it is not possible to burn sufficient coin to make the requested supply match actual supply
- The mint operation fails for any reason (see bank module)

## SupplyIncrease Proposal
SupplyIncreaseProposal defines a governance proposal to administer a marker and increase total supply of the marker
through minting coin and placing it within the marker or assigning it directly to an account.

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/proposals.proto#L34-L43

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- The requested supply exceeds the configuration parameter for `MaxTotalSupply`

## SupplyDecrease Proposal
SupplyDecreaseProposal defines a governance proposal to administer a marker and decrease the total supply through
burning coin held within the marker

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/proposals.proto#L47-L55

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
- The marker account itself is not holding sufficient supply to cover the amount of coin requested to burn
- The amount of resulting supply would be less than zero

The chain will panic and halt if:
- The bank burn operation fails for any reason (see bank module)

## SetAdministrator Proposal
SetAdministratorProposal defines a governance proposal to administer a marker and set administrators with specific
access on the marker

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/proposals.proto#L59-L67

## RemoveAdministrator Proposal
RemoveAdministratorProposal defines a governance proposal to administer a marker and remove all permissions for a
given address

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/proposals.proto#L71-L79

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid

## ChangeStatus Proposal
ChangeStatusProposal defines a governance proposal to administer a marker to change its status

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/proposals.proto#L82-L90

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid

## WithdrawEscrow Proposal
WithdrawEscrowProposal defines a governance proposal to withdraw escrow coins from a marker

+++ https://github.com/provenance-io/provenance/blob/2e713a82ac71747e99975a98e902efe01286f591/proto/provenance/marker/v1/proposals.proto#L93-L103

This request is expected to fail if:
- The governance proposal format (title, description, etc) is invalid
