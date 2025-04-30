# Events

The Ledger module emits events to track state changes and provide transparency for ledger operations.

## Event Types

### Ledger Class Creation
Emitted when a new ledger class is created:

```go
Event: "ledger_class_created"
Attributes:
- "ledger_class_id": The unique identifier for the ledger class
- "asset_class_id": The Scope Specification ID or NFT Class ID
- "denom": The denomination used for the ledger class
- "maintainer_address": The address of the maintainer
```

### Ledger Creation
Emitted when a new ledger is created for an asset:

```go
Event: "ledger_created"
Attributes:
- "nft_id": The NFT or Scope identifier
- "asset_class_id": The Scope Specification ID or NFT Class ID
- "ledger_class_id": The ledger class identifier
```

### Ledger Configuration Update
Emitted when a ledger's configuration is updated:

```go
Event: "ledger_config_updated"
Attributes:
- "nft_id": The NFT or Scope identifier
- "asset_class_id": The Scope Specification ID or NFT Class ID
```

### Ledger Entry Addition
Emitted when a new entry is added to a ledger:

```go
Event: "ledger_entry_added"
Attributes:
- "nft_id": The NFT or Scope identifier
- "asset_class_id": The Scope Specification ID or NFT Class ID
- "correlation_id": The correlation ID of the entry (max 50 characters)
- "entry_type_id": The type of ledger entry
- "is_void": Whether the entry is void
```

## Event Attributes

Each event includes standard attributes:

- `module`: Always set to "ledger"
- `action`: The type of event (e.g., "ledger_class_created", "ledger_created", "ledger_config_updated", "ledger_entry_added")
- `nft_id`: The NFT or Scope identifier associated with the event
- `asset_class_id`: The Scope Specification ID or NFT Class ID

Additional attributes specific to each event type:

### Ledger Class Creation
- `ledger_class_id`: The unique identifier for the ledger class
- `denom`: The denomination used for the ledger class
- `maintainer_address`: The address of the maintainer

### Ledger Creation
- `ledger_class_id`: The ledger class identifier

### Ledger Configuration Update
- No additional attributes beyond the standard ones

### Ledger Entry Addition
- `correlation_id`: The correlation ID of the entry
- `entry_type_id`: The type of ledger entry
- `is_void`: Whether the entry is void

## Event Indexing

Events are indexed for efficient querying:

1. By Asset Identifier
   - All events related to a specific NFT or Scope
   - Used for tracking asset-specific activities

2. By Asset Class
   - All events related to a specific asset class
   - Used for monitoring class-specific operations

3. By Event Type
   - All events of a specific type
   - Used for monitoring specific operations

## Event Processing

Events are processed in the following order:

1. **Event Creation**
   - Events are created during state transitions
   - All required attributes are populated
   - Events are validated before emission

2. **Event Emission**
   - Events are emitted to the event system
   - Events are indexed for querying
   - Events are made available to subscribers

3. **Event Consumption**
   - Events can be consumed by external systems
   - Events can trigger additional actions
   - Events can be used for auditing and monitoring

## Event Usage

Events can be used to:

1. **Real-time Monitoring**
   - Track financial activities as they occur
   - Monitor ledger creation and updates
   - Track entry additions

2. **External System Integration**
   - Update accounting systems
   - Sync with external databases
   - Trigger business processes

3. **Audit and Compliance**
   - Maintain audit trails
   - Track changes over time
   - Verify transaction history

## Event Subscription

Events can be subscribed to using:
- Tendermint event subscription
- Cosmos SDK event system
- External event listeners 