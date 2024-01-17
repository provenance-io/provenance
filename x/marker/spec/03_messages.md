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
  - [Msg/UpdateRequiredAttributesRequest](#msgupdaterequiredattributesrequest)
  - [Msg/UpdateSendDenyListRequest](#msgupdatesenddenylistrequest)
  - [Msg/UpdateForcedTransferRequest](#msgupdateforcedtransferrequest)
  - [Msg/SetAccountDataRequest](#msgsetaccountdatarequest)
  - [Msg/AddNetAssetValuesRequest](#msgaddnetassetvaluesrequest)


## Msg/AddMarkerRequest

A marker is created using the Add Marker service message.
The created marker can not be directly added in an Active (or Cancelled/Destroyed) status.  Markers
must have a valid supply and denomination value.

+++ https://github.com/provenance-io/provenance/blob/d999e7acefea1d77e2d5ce86ea89524dbf2692f8/proto/provenance/marker/v1/tx.proto#L83-L102

+++ https://github.com/provenance-io/provenance/blob/d999e7acefea1d77e2d5ce86ea89524dbf2692f8/proto/provenance/marker/v1/tx.proto#L104-L105


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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L98-L103

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L105-106

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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L108-L113

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L114-L115

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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L117-L121

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L122-L123

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `proposed` status or:
  - The request is not signed with an administrator address that matches the manager address or:
  - The given administrator address does not currently have the "admin" access granted on the marker

The `Finalize` marker status performs a set of checks to ensure the marker is ready to be activated.  It is designed to
serve as an intermediate step prior to activation that indicates marker configuration is complete.

## Msg/ActivateRequest

Activate Request defines the Msg/Activate request type

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L125-L129

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L130-L131

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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L133-L137

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L138-L139

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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L141-L145

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L146-L147

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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L149-L154

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L155-L156

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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L158-L163

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L164-L165

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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L167-L174

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L175-L176

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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L178-L185

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L187-L188

This service message is expected to fail if:

- The given denom value is invalid or does not match an existing marker on the system
- The marker is not in a `Active` status or:
  - The given administrator address does not currently have the "transfer" access granted on the marker
  - The marker types is not `RESTRICTED_COIN`

## Msg/IbcTransferRequest

Ibc transfer Request defines the Msg/IbcTransfer request type.  The `IbcTransferRequest` is used to transfer `RESTRICTED_COIN` type markers to another chain via ibc. These coins have their `send_enabled` flag disabled by the bank module and thus cannot be sent using a normal `send_coin` operation.

NOTE: A transfer request also requires a signature from an account with the transfer permission as well as approval from the account the funds will be withdrawn from.

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L190-L197

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L199-L200

## Msg/SetDenomMetadataRequest

SetDenomMetadata Request defines the Msg/SetDenomMetadata request type.  This request is used to set the informational
denom metadata held within the bank module.  Denom metadata can be used to provide a more streamlined user experience
within block explorers or similar applications.

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L202-L207

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L209-L210

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

+++ https://github.com/provenance-io/provenance/blob/d999e7acefea1d77e2d5ce86ea89524dbf2692f8/proto/provenance/marker/v1/tx.proto#L221-L236

+++ https://github.com/provenance-io/provenance/blob/d999e7acefea1d77e2d5ce86ea89524dbf2692f8/proto/provenance/marker/v1/tx.proto#L238-L239

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

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L61-L72

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L74-L75

This service message is expected to fail if:

- Any field is empty.
- The allowance is invalid
- The given denom value is invalid or does not match an existing marker on the system
- The administrator or grantee are invalid addresses
- The administrator does not have `ADMIN` access on the marker.

## Msg/SupplyIncreaseProposalRequest

SupplyIncreaseProposal is a governance-only message for increasing the supply of a marker.

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L229-L239

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1//tx.proto#L241-L242

This service message is expected to fail if:

- The authority is not the address of the governance module's account.
- The governance proposal format (title, description, etc) is invalid
- The requested supply exceeds the configuration parameter for `MaxSupply`

See also: [Governance: Supply Increase Proposal](./10_governance.md#supply-increase-proposal)

## Msg/UpdateRequiredAttributesRequest

UpdateRequiredAttributes allows signers that have transfer authority or via gov proposal to add and remove required attributes from a restricted marker.

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L244-L255

+++ https://github.com/provenance-io/provenance/blob/22740319ba4b3ba268b3720d4bee36d6c6b06b40/proto/provenance/marker/v1/tx.proto#L257-L258

This service message is expected to fail if:

- Remove list has an attribute that does not exist in current Required Attributes
- Add list has an attribute that already exist in current Required Attributes
- Attributes cannot be normalized
- Marker denom cannot be found or is not a restricted marker

## Msg/UpdateSendDenyListRequest

UpdateSendDenyList allows signers that have transfer authority or via gov proposal to add and remove addresses to the deny send list for a restricted marker.

+++ https://github.com/provenance-io/provenance/blob/373d0ee8faeaa9e3b70d658e6069ab1781e6ce5e/proto/provenance/marker/v1/tx.proto#L295-L308

+++ https://github.com/provenance-io/provenance/blob/373d0ee8faeaa9e3b70d658e6069ab1781e6ce5e/proto/provenance/marker/v1/tx.proto#L310-L311

This service message is expected to fail if:

- Remove list has an address that does not exist in current deny list
- Add list has an attribute that already exist in current deny list
- Both add and remove lists are empty
- Invalid address format in add/remove lists
- Marker denom cannot be found or is not a restricted marker
- Signer does not have transfer authority or is not from gov proposal

## Msg/UpdateForcedTransferRequest

UpdateForcedTransfer allows for the activation or deactivation of forced transfers for a marker.
This message must be submitted via governance proposal.

+++ https://github.com/provenance-io/provenance/blob/a830a8ecf24199469de52b92ee20475d6912f2eb/proto/provenance/marker/v1/tx.proto#L260-L272

+++ https://github.com/provenance-io/provenance/blob/a830a8ecf24199469de52b92ee20475d6912f2eb/proto/provenance/marker/v1/tx.proto#L274-L275

This service message is expected to fail if:

- The authority is not the governance module account address.
- No marker with the provided denom exists.
- The marker is not a restricted coin.
- The marker does not allow governance control.

## Msg/SetAccountDataRequest

SetAccountData allows the association of some data (a string) with a marker.

+++ https://github.com/provenance-io/provenance/blob/e83f1955cba07e2ba87790c4487d22632ae9e69c/proto/provenance/marker/v1/tx.proto#L279-L291

+++ https://github.com/provenance-io/provenance/blob/e83f1955cba07e2ba87790c4487d22632ae9e69c/proto/provenance/marker/v1/tx.proto#L293-L294

This endpoint can either be used directly or via governance proposal.

This service message is expected to fail if:

- No marker with the provided denom exists.
- The signer is the governance module account address but the marker does not allow governance control.
- The signer is not the governance module account and does not have deposit access on the marker.
- The provided value is too long (as defined by the attribute module params).

## Msg/AddNetAssetValuesRequest

AddNetAssetValuesRequest allows for the adding/updating of net asset values for a marker.

+++ https://github.com/provenance-io/provenance/blob/25070572cc898c476f5bb1a816c6c1c4d07e3d38/proto/provenance/marker/v1/tx.proto#L327-L332

+++ https://github.com/provenance-io/provenance/blob/25070572cc898c476f5bb1a816c6c1c4d07e3d38/proto/provenance/marker/v1/tx.proto#L334-L335

This endpoint can either be used directly or via governance proposal.

This service message is expected to fail if:

- No marker with the provided denom exists.
- The signer is the governance module account address but the marker does not allow governance control.
- The signer is not the governance module account and does not have any access on the marker.
- The provided net value asset properties are invalid.
