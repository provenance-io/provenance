# Metadata Events

The metadata module emits the following events and telemetry information.

<!-- TOC 2 5 -->
  - [Counters](#counters)
    - [Stored Objects](#stored-objects)
      - [Stored Object: Keys](#stored-object-keys)
      - [Stored Object: Labels](#stored-object-labels)
        - [Stored Object: Label: Category](#stored-object-label-category)
        - [Stored Object: Label: Object Type](#stored-object-label-object-type)
    - [Object Actions](#object-actions)
      - [Object Action: Keys](#object-action-keys)
      - [Object Action: Labels](#object-action-labels)
        - [Object Action: Label: Category](#object-action-label-category)
        - [Object Action: Label: Object Type](#object-action-label-object-type)
        - [Object Action: Label: Action](#object-action-label-action)
  - [Timers](#timers)
    - [TX Keys](#tx-keys)
    - [Query Keys](#query-keys)



---
## Counters

### Stored Objects

This counter is used to get counts of things stored on the chain.

When this module writes a new object to the chain, this counter is incremented by 1.
When this module deletes an object from the chain, this counter is decremented by 1.
When this module updates an object on the chain, this counter is not updated.

#### Stored Object: Keys

`"metadata"`, `"stored-object"`

#### Stored Object: Labels

`"category"`, `"object-type"`

##### Stored Object: Label: Category

This label groups the objects into a general type.

The string for this label is `"category"`.

Possible values:
- `"entry"`
- `"specification"`
- `"object-store-locator"`

##### Stored Object: Label: Object Type

This label specifically identifies objects.
Each value belongs to exactly one "category" label.

The string for this label is `"object-type"`.

Possible values:
- `"scope"` (is an `"entry"`)
- `"session"` (is an `"entry"`)
- `"record"` (is an `"entry"`)
- `"scope-specification"` (is a `"specification"`)
- `"contract-specification"` (is a `"specification"`)
- `"record-specification"` (is a `"specification"`)
- `"object-store-locator"` (is an `"object-store-locator"`)



### Object Actions

This counter is used to get counts of actions taken on the chain.

Every time this module writes to or deletes from the chain, this counter is incremented.

#### Object Action: Keys

`"metadata"`, `"object-action"`

#### Object Action: Labels

`"category"`, `"object-type"`, `"action"`

##### Object Action: Label: Category

This is the same label used by the stored object counter: [Stored Object: Label: Category](#stored-object-label-category)

##### Object Action: Label: Object Type

This is the same label used by the stored object counter: [Stored Object: Label: Object Type](#stored-object-label-object-type)

##### Object Action: Label: Action

This label defines the actions taken with respects to the various objects.

The string for this label is `"action"`.

Possible values:
- `"created"`
- `"updated"`
- `"deleted"`



---
## Timers

All TX and Query endpoints have related timing metrics.

### TX Keys

`"metadata"`, `"tx"`, `{endpoint}`

Example `{endpoint}` values: `"WriteScope"`, `"DeleteContractSpecification"`, `"ModifyOSLocator"`.

### Query Keys

`"metadata"`, `"query"`, `{endpoint}`

Example `{endpoint}` values: `"Scope"`, `"ContractSpecificationsAll"`, `"OSLocatorsByScope"`.
