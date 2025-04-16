# Events

The Ledger module emits events to track state changes and provide transparency for ledger operations.

## Event Types

### Ledger Creation
Emitted when a new ledger is created for an NFT:

```go
Event: "ledger_created"
Attributes:
- "nft_address": The address of the NFT
- "denom": The denomination used for the ledger
```

### Ledger Configuration Update
Emitted when a ledger's configuration is updated:

```go
Event: "ledger_config_updated"
Attributes:
- "nft_address": The address of the NFT
```

### Ledger Entry Addition
Emitted when a new entry is added to a ledger:

```go
Event: "ledger_entry_added"
Attributes:
- "nft_address": The address of the NFT
- "correlation_id": The correlation ID of the entry (max 50 characters)
```

## Event Attributes

Each event includes standard attributes:

- `module`: Always set to "ledger"
- `action`: The type of event (e.g., "ledger_created", "ledger_config_updated", "ledger_entry_added")
- `nft_address`: The NFT address associated with the event

Additional attributes specific to each event type:

### Ledger Creation
- `denom`: The denomination used for the ledger

### Ledger Configuration Update
- No additional attributes beyond the standard ones

### Ledger Entry Addition
- `correlation_id`: The correlation ID of the entry

## Event Indexing

Events are indexed for efficient querying:

1. By NFT Address
   - All events related to a specific NFT
   - Used for tracking NFT-specific activities

2. By Event Type
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