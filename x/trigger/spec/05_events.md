<!--
order: 5
-->

# Events

The trigger module emits the following events:

<!-- TOC -->
  - [Trigger Created](#trigger-created)
  - [Trigger Destroyed](#trigger-destroyed)
  - [Trigger Detected](#trigger-detected)
  - [Trigger Executed](#trigger-executed)

---
## Trigger Created

Fires when a trigger is created with the CreateTriggerMsg.

| Type           | Attribute Key | Attribute Value               |
| -------------- | ------------- | ----------------------------- |
| TriggerCreated | trigger_id    | The ID of the created trigger |

---
## Trigger Destroyed

Fires when a trigger is destroyed with the DestroyTriggerMsg.

| Type             | Attribute Key | Attribute Value                       |
| ---------------- | ------------- | ------------------------------------- |
| TriggerDestroyed | trigger_id    | The ID of the trigger being destroyed |
---
## Trigger Detected

Fires when a trigger's event is detected in the EndBlocker.

| Type            | Attribute Key | Attribute Value                      |
| --------------- | ------------- | ------------------------------------ |
| TriggerDetected | trigger_id    | The ID of the trigger being detected |
---
## Trigger Executed

Fires when a trigger's actions are executed in the BeginBlocker.

| Type            | Attribute Key | Attribute Value                                               |
| --------------- | ------------- | ------------------------------------------------------------- |
| TriggerExecuted | trigger_id    | The ID of the trigger being executed                          |
| TriggerExecuted | owner         | The sdk.Address of the trigger's owner                        |
| TriggerExecuted | success       | A boolean indicating if all the actions successfully executed |
