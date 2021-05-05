# Metadata Events

The metadata module emits the following events and telemetry information.

<!-- TOC 2 5 -->
  - [Counters](#counters)
    - [Stored Objects](#stored-objects)
      - [Keys](#keys)
      - [Labels](#labels)
        - [category](#category)
        - [object-type](#object-type)
        - [action](#action)
  - [Timers](#timers)
    - [TX Keys](#tx-keys)
    - [Query Keys](#query-keys)

## Counters

### Stored Objects

This counter is used to get counts of things stored in the chain.

The counter value is updated based on the `"action"` label in the following ways:
- When the action is `"created"`, 1 is added to the counter.
- When the action is `"updated"`, the counter is not changed.
- When the action is `"deleted"`, 1 is subtracted from the counter.

#### Keys

`"metadata"`, `"stored-object"`

#### Labels

`"category"`, `"object-type"`, `"action"`

##### category

This label groups the objects into a general type.

Possible values:
- `"entry"`
- `"specification"`
- `"object-store-locator"`

##### object-type

This label specifically identifies objects.
Each value belongs to exactly one "category" label.

Possible values:
- `"scope"` (is an `"entry"`)
- `"session"` (is an `"entry"`)
- `"record"` (is an `"entry"`)
- `"scope-specification"` (is a `"specification"`)
- `"contract-specification"` (is a `"specification"`)
- `"record-specification"` (is a `"specification"`)
- `"object-store-locator"` (is an `"object-store-locator"`)

##### action

This label defines the actions taken with respects to the various objects.

Possible values:
- `"created"`
- `"updated"`
- `"deleted"`

## Timers

All TX and Query endpoints have related timing metrics.

### TX Keys

`"metadata"`, `"tx"`, `{endpoint}`

Example `{endpoint}` values: `"WriteScope"`, `"DeleteContractSpecification"`, `"ModifyOSLocator"`.

### Query Keys

`"metadata"`, `"query"`, `{endpoint}`

Example `{endpoint}` values: `"Scope"`, `"ContractSpecificationsAll"`, `"OSLocatorsByScope"`.
