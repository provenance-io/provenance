# Metadata Events

The metadata module emits the following events and telemetry information.

<!-- TOC 2 5 -->
  - [Events](#events)
    - [Generic](#generic)
      - [EventTxCompleted](#eventtxcompleted)
    - [Scope](#scope)
      - [EventScopeCreated](#eventscopecreated)
      - [EventScopeUpdated](#eventscopeupdated)
      - [EventScopeDeleted](#eventscopedeleted)
    - [Session](#session)
      - [EventSessionCreated](#eventsessioncreated)
      - [EventSessionUpdated](#eventsessionupdated)
      - [EventSessionDeleted](#eventsessiondeleted)
    - [Record](#record)
      - [NewEventRecordCreated](#neweventrecordcreated)
      - [EventRecordUpdated](#eventrecordupdated)
      - [EventRecordDeleted](#eventrecorddeleted)
    - [Scope Specification](#scope-specification)
      - [EventScopeSpecificationCreated](#eventscopespecificationcreated)
      - [EventScopeSpecificationUpdated](#eventscopespecificationupdated)
      - [EventScopeSpecificationDeleted](#eventscopespecificationdeleted)
    - [Contract Specification](#contract-specification)
      - [EventContractSpecificationCreated](#eventcontractspecificationcreated)
      - [EventContractSpecificationUpdated](#eventcontractspecificationupdated)
      - [EventContractSpecificationDeleted](#eventcontractspecificationdeleted)
    - [Record Specification](#record-specification)
      - [EventRecordSpecificationCreated](#eventrecordspecificationcreated)
      - [EventRecordSpecificationUpdated](#eventrecordspecificationupdated)
      - [EventRecordSpecificationDeleted](#eventrecordspecificationdeleted)
    - [Object Store Locator](#object-store-locator)
      - [EventOSLocatorCreated](#eventoslocatorcreated)
      - [EventOSLocatorUpdated](#eventoslocatorupdated)
      - [EventOSLocatorDeleted](#eventoslocatordeleted)
  - [Telemetry](#telemetry)
    - [Counters](#counters)
      - [Stored Objects](#stored-objects)
        - [Keys](#keys)
        - [Labels](#labels)
        - [Label: category](#label--category)
        - [Label: object-type](#label--object-type)
        - [Label: action](#label--action)

## Events

### Generic

#### EventTxCompleted

This event is emitted whenever a TX has completed without issues.
It will usually be accompanied by one or more of the other events.

| Attribute Key         | Attribute Value                                   |
| --------------------- | ------------------------------------------------- |
| Module                | "metadata"                                        |
| Endpoint              | The name of the rpc called, e.g. "WriteScope"     |
| Signers               | List of bech32 address strings of the msg signers |

### Scope

#### EventScopeCreated

This event is emitted whenever a new scope is written.

| Attribute Key         | Attribute Value                                   |
| --------------------- | ------------------------------------------------- |
| ScopeAddr             | The bech32 address string of the ScopeId          |

#### EventScopeUpdated

This event is emitted whenever an existing scope is updated.

| Attribute Key         | Attribute Value                                   |
| --------------------- | ------------------------------------------------- |
| ScopeAddr             | The bech32 address string of the ScopeId          |

#### EventScopeDeleted

This event is emitted whenever an existing scope is deleted.

| Attribute Key         | Attribute Value                                   |
| --------------------- | ------------------------------------------------- |
| ScopeAddr             | The bech32 address string of the ScopeId          |

### Session

#### EventSessionCreated

This event is emitted whenever a new session is written.

| Attribute Key         | Attribute Value                                    |
| --------------------- | -------------------------------------------------- |
| SessionAddr           | The bech32 address string of the SessionId         |
| ScopeAddr             | The bech32 address string of the session's ScopeId |

#### EventSessionUpdated

This event is emitted whenever an existing session is updated.

| Attribute Key         | Attribute Value                                    |
| --------------------- | -------------------------------------------------- |
| SessionAddr           | The bech32 address string of the SessionId         |
| ScopeAddr             | The bech32 address string of the session's ScopeId |

#### EventSessionDeleted

This event is emitted whenever an existing session is deleted.

| Attribute Key         | Attribute Value                                    |
| --------------------- | -------------------------------------------------- |
| SessionAddr           | The bech32 address string of the SessionId         |
| ScopeAddr             | The bech32 address string of the session's ScopeId |

### Record

#### NewEventRecordCreated

This event is emitted whenever a new record is written.

| Attribute Key         | Attribute Value                                     |
| --------------------- | --------------------------------------------------- |
| RecordAddr            | The bech32 address string of the RecordId           |
| SessionAddr           | The bech32 address string of the record's SessionId |
| ScopeAddr             | The bech32 address string of the record's ScopeId   |

#### EventRecordUpdated

This event is emitted whenever an existing record is updated.

| Attribute Key         | Attribute Value                                     |
| --------------------- | --------------------------------------------------- |
| RecordAddr            | The bech32 address string of the RecordId           |
| SessionAddr           | The bech32 address string of the record's SessionId |
| ScopeAddr             | The bech32 address string of the record's ScopeId   |

#### EventRecordDeleted

This event is emitted whenever an existing record is deleted.

| Attribute Key         | Attribute Value                                   |
| --------------------- | ------------------------------------------------- |
| RecordAddr            | The bech32 address string of the RecordId         |
| ScopeAddr             | The bech32 address string of the record's ScopeId |

### Scope Specification

#### EventScopeSpecificationCreated

This event is emitted whenever a new scope specification is written.

| Attribute Key          | Attribute Value                                   |
| ---------------------- | ------------------------------------------------- |
| ScopeSpecificationAddr | The bech32 address string of the SpecificationId  |

#### EventScopeSpecificationUpdated

This event is emitted whenever an existing scope specification is updated.

| Attribute Key          | Attribute Value                                   |
| ---------------------- | ------------------------------------------------- |
| ScopeSpecificationAddr | The bech32 address string of the SpecificationId  |

#### EventScopeSpecificationDeleted

This event is emitted whenever an existing scope specification is deleted.

| Attribute Key          | Attribute Value                                   |
| ---------------------- | ------------------------------------------------- |
| ScopeSpecificationAddr | The bech32 address string of the SpecificationId  |

### Contract Specification

#### EventContractSpecificationCreated

This event is emitted whenever a new contract specification is written.

| Attribute Key             | Attribute Value                                   |
| ------------------------- | ------------------------------------------------- |
| ContractSpecificationAddr | The bech32 address string of the SpecificationId  |

#### EventContractSpecificationUpdated

This event is emitted whenever an existing contract specification is updated.

| Attribute Key             | Attribute Value                                   |
| ------------------------- | ------------------------------------------------- |
| ContractSpecificationAddr | The bech32 address string of the SpecificationId  |

#### EventContractSpecificationDeleted

This event is emitted whenever an existing contract specification is deleted.

| Attribute Key             | Attribute Value                                   |
| ------------------------- | ------------------------------------------------- |
| ContractSpecificationAddr | The bech32 address string of the SpecificationId  |

### Record Specification

#### EventRecordSpecificationCreated

This event is emitted whenever a new record specification is written.

| Attribute Key             | Attribute Value                                            |
| ------------------------- | ---------------------------------------------------------- |
| RecordSpecificationAddr   | The bech32 address string of the SpecificationId           |
| ContractSpecificationAddr | The bech32 address string of the Contract SpecificationId  |

#### EventRecordSpecificationUpdated

This event is emitted whenever an existing record specification is updated.

| Attribute Key             | Attribute Value                                            |
| ------------------------- | ---------------------------------------------------------- |
| RecordSpecificationAddr   | The bech32 address string of the SpecificationId           |
| ContractSpecificationAddr | The bech32 address string of the Contract SpecificationId  |

#### EventRecordSpecificationDeleted

This event is emitted whenever an existing record specification is deleted.

| Attribute Key             | Attribute Value                                            |
| ------------------------- | ---------------------------------------------------------- |
| RecordSpecificationAddr   | The bech32 address string of the SpecificationId           |
| ContractSpecificationAddr | The bech32 address string of the Contract SpecificationId  |

### Object Store Locator

#### EventOSLocatorCreated

This event is emitted whenever a new object store locator is written.

| Attribute Key    | Attribute Value                        |
| ---------------- | -------------------------------------- |
| Owner            | The bech32 address string of the Owner |

#### EventOSLocatorUpdated

This event is emitted whenever an existing object store locator is updated.

| Attribute Key    | Attribute Value                        |
| ---------------- | -------------------------------------- |
| Owner            | The bech32 address string of the Owner |

#### EventOSLocatorDeleted

This event is emitted whenever an existing object store locator is deleted.

| Attribute Key    | Attribute Value                        |
| ---------------- | -------------------------------------- |
| Owner            | The bech32 address string of the Owner |

## Telemetry

### Counters

#### Stored Objects

This counter is used to get counts of things stored in the chain.

The counter value is updated based on the "action" label in the following ways:
- When the action is "created", 1 is added to the counter.
- When the action is "updated", the counter is not changed.
- When the action is "deleted", 1 is subtracted from the counter.

##### Keys

"metadata", "stored-object"

##### Labels

"category", "object-type", "action"

##### Label: category

This label groups the objects into a general type.

Possible values:
- "entry"
- "specification"
- "object-store-locator"

##### Label: object-type

This label specifically identifies objects.
Each value belongs to exactly one "category" label.

Possible values:
- "scope" (is an "entry")
- "session" (is an "entry")
- "record" (is an "entry")
- "scope-specification" (is a "specification")
- "contract-specification" (is a "specification")
- "record-specification" (is a "specification")
- "object-store-locator" (is an "object-store-locator")

##### Label: action

This label defines the actions taken with respects to the various objects.

Possible values:
- "created"
- "updated"
- "deleted"
