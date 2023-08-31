<!--
order: 5
-->

# Events

The trigger module emits the following events:

<!-- TOC -->
  - [Trigger Created](#trigger-created)
  - [Trigger Destroyed](#trigger-destroyed)


---
## Trigger Created

Fires when a trigger is created with the CreateTriggerMsg.

| Type           | Attribute Key | Attribute Value |
| -------------- | ------------- | --------------- |
| TriggerCreated | trigger_id    | {ID string}     |

---
## Trigger Destroyed

Fires when a trigger is destroyed with the DestroyTriggerMsg.

| Type             | Attribute Key | Attribute Value |
| ---------------- | ------------- | --------------- |
| TriggerDestroyed | trigger_id    | {ID string}     |
