# Messages

In this section we describe the processing of the marker messages and the corresponding updates to the state.
All created/modified state objects specified by each message are defined within the
[state](./02_state_transitions.md) section.

<!-- TOC 2 2 -->
  - [Msg/AddMarkerRequest](#msgaddmarkerrequest)
  - [Msg/AddAccessRequest](#msgaddaccessrequest)
  - [Msg/DeleteAccessRequest](#msgdeleteaccessrequest)
  - [Msg/FinalizeRequest](#msgfinalizerequest)
  - [Msg/ActivateRequest](#msgactivaterequest)
  - [Msg/CancelRequest](#msgcancelrequest)
  - [Msg/DeleteRequest](#msgdeleterequest)
  - [Msg/MintRequest](#msgmintrequest)
  - [Msg/BurnRequest](#msgburnrequest)
  - [Msg/WithdrawRequest](#msgwithdrawrequest)
  - [Msg/TransferRequest](#msgtransferrequest)
  - [Msg/IbcTransferRequest](#msgibctransferrequest)
  - [Msg/SetDenomMetadataRequest](#msgsetdenommetadatarequest)
  - [Msg/AddFinalizeActivateMarkerRequest](#msgaddfinalizeactivatemarkerrequest)
  - [Msg/GrantAllowanceRequest](#msggrantallowancerequest)
  - [Msg/SupplyIncreaseProposalRequest](#msgsupplyincreaseproposalrequest)



## Msg/AddMarkerRequest

A marker is created using the Add Marker service message.
The created marker can not be directly added in an Active (or Cancelled/Destroyed) status.  Markers
must have a valid supply and denomination value.

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L74-L85

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L88


This service message is expected to fail if:
- The Denom string:
  - Is already in use by another marker
  - Does not conform to the "Marker Denom Validation Expression" (`unrestricted_denom_regex` param)
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

If issued via governance proposal, and has a `from_address` of the governance module account:
- The marker status can be Active.
- The `unrestricted_denom_regex` check is not applied. Denoms still need to conform to the base coin denom format though.
- The marker's `allow_governance_control` flag ignores the `enable_governance` param value, and is set to the provided value.
- If the marker status is Active, and no `manager` is provided, it is left blank (instead of being populated with the `from_address`).

## Msg/AddAccessRequest

Add Access Request is used to add permissions to a marker that allow the specified accounts to perform the specified actions.

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L91-L95

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L98

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

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L101-L105

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L107

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

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L110-L113

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L115

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `proposed` status or:
  - The request is not signed with an administrator address that matches the manager address or:
  - The given administrator address does not currently have the "admin" access granted on the marker

The `Finalize` marker status performs a set of checks to ensure the marker is ready to be activated.  It is designed to
serve as an intermediate step prior to activation that indicates marker configuration is complete.

## Msg/ActivateRequest

Activate Request defines the Msg/Activate request type

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L118-L121

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L123

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

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L126-L129

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L131

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

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L134-L137

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L139

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

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L142-L146

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L148

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

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L151-L155

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L157

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

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L160-L166

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L168

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
accounts for `RESTRICTED_COIN` type markers. Such markers have `send_enabled=false` configured with the `x/bank` module,
and thus cannot be sent using a normal `MsgSend` operation.  A transfer request requires a signature from an account
with `TRANSFER` access. If force transfer is not enabled for the marker, the source account must have granted the admin
permission (via `authz`) to do the transfer. If force transfer is allowed for the marker, the source account does not
need to approve of the transfer.

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L171-L177

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L180

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `Active` status or:
  - The given administrator address does not currently have the "transfer" access granted on the marker
  - The marker types is not `RESTRICTED_COIN`

## Msg/IbcTransferRequest

Ibc transfer Request defines the Msg/IbcTransfer request type.  The `IbcTransferRequest` is used to transfer `RESTRICTED_COIN` type markers to another chain via ibc. These coins have their `send_enabled` flag disabled by the bank module and thus cannot be sent using a normal `send_coin` operation.

NOTE: A transfer request also requires a signature from an account with the transfer permission as well as approval from the account the funds will be withdrawn from.

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L183-L189

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L192

## Msg/SetDenomMetadataRequest

SetDenomMetadata Request defines the Msg/SetDenomMetadata request type.  This request is used to set the informational
denom metadata held within the bank module.  Denom metadata can be used to provide a more streamlined user experience
within block explorers or similar applications.

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L195-L199

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L202

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

## Msg/AddFinalizeActivateMarkerRequest

AddFinalizeActivate requested is used for adding, finalizing, and activating a marker in a single request.

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L205-L215

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L218

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is pending:
  - And the request is not signed with an administrator address that matches the manager address or:
  - The given administrator address does not currently have the "admin" access granted on the marker
- The accesslist:
  - Contains more than one entry for a given address
  - Contains a grant with an invalid address
  - Contains a grant with an invalid access enum value (Unspecified/0)

## Msg/GrantAllowanceRequest

GrantAllowance grants a fee allowance to the grantee on the granter's account.

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L59-L68

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L71

This service message is expected to fail if:

- Any field is empty.
- The allowance is invalid
- The given denom value is invalid or does not match an existing marker on the system
- The administrator or grantee are invalid addresses
- The administrator does not have `ADMIN` access on the marker.

## Msg/SupplyIncreaseProposalRequest

SupplyIncreaseProposal is a governance-only message for increasing the supply of a marker.

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L222-L230

+++ https://github.com/provenance-io/provenance/blob/0b005ca855eb0dcda86a87d585e8a021c87d985d/proto/provenance/marker/v1/tx.proto#L233

This service message is expected to fail if:

- The authority is not the address of the governance module's account.
- The governance proposal format (title, description, etc) is invalid
- The requested supply exceeds the configuration parameter for `MaxTotalSupply`

See also: [Governance: Supply Increase Proposal](./10_governance.md#supply-increase-proposal)
