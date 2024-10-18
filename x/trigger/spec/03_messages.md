<!--
order: 3
-->

# Messages

In this section we describe the processing of the trigger messages and the corresponding updates to the state.

<!-- TOC 2 -->
  - [Msg/CreateTrigger](#msgcreatetrigger)
  - [Msg/DestroyTrigger](#msgdestroytrigger)


## Msg/CreateTrigger

Creates a `Trigger` that will fire when its event has been detected. If the message has more than one signer, then the newly created `Trigger` will designate the first signer as the owner.

### Request

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/tx.proto#L23-L34

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/tx.proto#L36-L40

The message will fail under the following conditions:
* The authority is an invalid bech32 address
* The event does not implement `TriggerEventI`
* The actions list is empty
* At least one action is not a valid `sdk.Msg`
* The signers on one or more actions aren't in the set of the request's signers.

## Msg/DestroyTrigger

Destroys a `Trigger` that has been created and is still registered.

### Request

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/tx.proto#L42-L51

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/tx.proto#L53-L54

The message will fail under the following conditions:
* The `Trigger` does not exist
* The `Trigger` owner does not match the specified address
