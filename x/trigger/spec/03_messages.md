<!--
order: 3
-->

# Messages

In this section we describe the processing of the trigger messages and the corresponding updates to the state.

<!-- TOC 3 -->
- [Messages](#messages)
  - [Msg/CreateTriggerRequest](#msgcreatetriggerrequest)
    - [Request](#request)
    - [Response](#response)
  - [Msg/DestroyTriggerRequest](#msgdestroytriggerrequest)
    - [Request](#request-1)
    - [Response](#response-1)


## Msg/CreateTriggerRequest

Creates a `Trigger` that will fire when its event has been detected. If the message has more than one signer, then the newly created `Trigger` will designate the first signer as the owner.

### Request
+++ https://github.com/provenance-io/provenance/blob/288f8b1b60861da811c61840dbf4220a3f906071/proto/provenance/trigger/v1/tx.proto#L21-L31

### Response
+++ https://github.com/provenance-io/provenance/blob/288f8b1b60861da811c61840dbf4220a3f906071/proto/provenance/trigger/v1/tx.proto#L34-L37

The message will fail under the following conditions:
* The authority is an invalid bech32 address
* The event does not implement `TriggerEventI`
* The actions list is empty
* At least one action is not a valid `sdk.Msg`
* The signers on one or more actions don't match the signers of this message

## Msg/DestroyTriggerRequest

Destroys a `Trigger` that has been created and is still registered.

### Request
+++ https://github.com/provenance-io/provenance/blob/288f8b1b60861da811c61840dbf4220a3f906071/proto/provenance/trigger/v1/tx.proto#L40-L48

### Response
+++ https://github.com/provenance-io/provenance/blob/288f8b1b60861da811c61840dbf4220a3f906071/proto/provenance/trigger/v1/tx.proto#L51

The message will fail under the following conditions:
* The `Trigger` does not exist
* The `Trigger` owner does not match the specified address
