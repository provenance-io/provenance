# Events

The Ledger module emits events to track state changes and provide transparency for ledger operations.

## Event Types

### Ledger Creation
Emitted when a new ledger is created for an NFT:

```protobuf
message EventLedgerCreated {
    // The address of the NFT
    string nft_address = 1;
    // The denomination used for the ledger
    string denom = 2;
}
```

### Ledger Entry Addition
Emitted when a new entry is added to a ledger:

```protobuf
message EventLedgerEntryAdded {
    // The address of the NFT
    string nft_address = 1;
    // The correlation ID of the entry (max 50 characters)
    string correlation_id = 2;
    // The type of the entry
    LedgerEntryType entry_type = 3;
    // The date the entry was posted
    google.protobuf.Timestamp posted_date = 4 [(gogoproto.stdtime) = true];
    // The date the entry takes effect
    google.protobuf.Timestamp effective_date = 5 [(gogoproto.stdtime) = true];
    // The total amount of the entry
    string amount = 6;
}
```

### Balance Updates
Emitted when balances are updated due to a ledger entry:

```protobuf
message EventBalanceUpdated {
    // The address of the NFT
    string nft_address = 1;
    // The new principal balance
    string principal_balance = 2;
    // The new interest balance
    string interest_balance = 3;
    // The new other balance
    string other_balance = 4;
}
```

### Ledger Configuration Updates
Emitted when ledger configuration is modified:

```protobuf
message EventLedgerConfigUpdated {
    // The address of the NFT
    string nft_address = 1;
    // The new denomination
    string denom = 2;
    // The previous denomination (if changed)
    string previous_denom = 3;
}
```

## Event Attributes

Each event includes standard attributes:

- `module`: Always set to "ledger"
- `action`: The type of event (e.g., "ledger_created", "entry_added", "balance_updated", "config_updated")
- `nft_address`: The NFT address associated with the event

Additional attributes specific to each event type:

### Ledger Creation
- `denom`: The denomination used for the ledger

### Ledger Entry Addition
- `correlation_id`: The correlation ID of the entry
- `entry_type`: The type of the entry
- `posted_date`: The date the entry was posted
- `effective_date`: The date the entry takes effect
- `amount`: The total amount of the entry

### Balance Updates
- `principal_balance`: The new principal balance
- `interest_balance`: The new interest balance
- `other_balance`: The new other balance

### Ledger Configuration Updates
- `denom`: The new denomination
- `previous_denom`: The previous denomination (if changed)

## Event Indexing

Events are indexed for efficient querying:

1. By NFT Address
   - All events related to a specific NFT
   - Used for tracking NFT-specific activities

2. By Event Type
   - All events of a specific type
   - Used for monitoring specific operations

3. By Date
   - Events within a specific time range
   - Used for historical analysis

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
   - Monitor balance changes
   - Detect configuration updates

2. **External System Integration**
   - Update accounting systems
   - Sync with external databases
   - Trigger business processes

3. **Audit and Compliance**
   - Maintain audit trails
   - Track changes over time
   - Verify transaction history

4. **Analytics and Reporting**
   - Generate financial reports
   - Analyze transaction patterns
   - Track performance metrics

## Event Subscription

Events can be subscribed to using:
- Tendermint event subscription
- Cosmos SDK event system
- External event listeners 