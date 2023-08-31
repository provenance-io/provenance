<!--
order: 3
-->

# Messages

In this section we describe the processing of the trigger messages and the corresponding updates to the state.

<!-- TOC 2 -->
  - [Msg/CreateTriggerRequest](#msgcreatetriggerrequest)
  - [Msg/DestroyTriggerRequest](#msgdestroytriggerrequest)


## Msg/CreateTriggerRequest

Creates a `Trigger` that will fire when its event has been detected. If the message has more than one signer, then the newly created `Trigger` will designate the first signer as the owner.

### Request

+++ https://github.com/provenance-io/provenance/blob/bda28e5f58a4a58e8fef21141400ad362b84518b/proto/provenance/trigger/v1/tx.proto#L20-L31

### Response

+++ https://github.com/provenance-io/provenance/blob/bda28e5f58a4a58e8fef21141400ad362b84518b/proto/provenance/trigger/v1/tx.proto#L33-L37

The message will fail under the following conditions:
* The authority is an invalid bech32 address
* The event does not implement `TriggerEventI`
* The actions list is empty
* At least one action is not a valid `sdk.Msg`
* The signers on one or more actions aren't in the set of the request's signers.

## Msg/DestroyTriggerRequest

Destroys a `Trigger` that has been created and is still registered.

### Request

+++ https://github.com/provenance-io/provenance/blob/bda28e5f58a4a58e8fef21141400ad362b84518b/proto/provenance/trigger/v1/tx.proto#L39-L48

### Response

+++ https://github.com/provenance-io/provenance/blob/bda28e5f58a4a58e8fef21141400ad362b84518b/proto/provenance/trigger/v1/tx.proto#L50-L51

The message will fail under the following conditions:
* The `Trigger` does not exist
* The `Trigger` owner does not match the specified address
