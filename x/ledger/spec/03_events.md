# Ledger Events

The Ledger module emits events to track state changes and provide transparency for ledger operations. Events are defined as protobuf types in `proto/provenance/ledger/v1/events.proto` and are emitted for all major state transitions including ledger creation, updates, entry additions, and destruction.

<!-- TOC -->
  - [Event Types](#event-types)
  - [Event Attributes](#event-attributes)
  - [Event Indexing](#event-indexing)
  - [Event Processing](#event-processing)

## Event Types

All events are defined as protobuf messages in `proto/provenance/ledger/v1/events.proto`:

### EventLedgerCreated
Emitted when a new ledger is created for an asset:

```protobuf
message EventLedgerCreated {
  // asset class of the ledger
  string asset_class_id = 1;

  // nft id of the ledger (scope id or nft id)
  string nft_id = 2;
}
```

### EventLedgerUpdated
Emitted when a ledger's configuration is updated:

```protobuf
message EventLedgerUpdated {
  // asset class of the ledger
  string asset_class_id = 1;

  // nft id of the ledger (scope id or nft id)
  string nft_id = 2;

  // What type of data update caused this event to be emitted
  UpdateType update_type = 3;
}

// UpdateType is the type of data update that caused the EventLedgerUpdated event to be emitted
enum UpdateType {
  // The update type is unspecified
  UPDATE_TYPE_UNSPECIFIED = 0;

  // The status of the ledger was updated
  UPDATE_TYPE_STATUS = 1;

  // The interest rate of the ledger was updated
  UPDATE_TYPE_INTEREST_RATE = 2;

  // The payment of the ledger was updated
  UPDATE_TYPE_PAYMENT = 3;

  // The maturity date of the ledger was updated
  UPDATE_TYPE_MATURITY_DATE = 4;
}
```

### EventLedgerEntryAdded
Emitted when a new entry is added to a ledger:

```protobuf
message EventLedgerEntryAdded {
  // asset class of the ledger
  string asset_class_id = 1;

  // nft id of the ledger (scope id or nft id)
  string nft_id = 2;

  // correlation id of the ledger entry
  string correlation_id = 3;
}
```

### EventFundTransferWithSettlement
Emitted when funds are transferred with settlement instructions:

```protobuf
message EventFundTransferWithSettlement {
  // asset class of the ledger
  string asset_class_id = 1;

  // nft id of the ledger (scope id or nft id)
  string nft_id = 2;

  // correlation id of the ledger entry
  string correlation_id = 3;
}
```

### EventLedgerDestroyed
Emitted when a ledger is destroyed:

```protobuf
message EventLedgerDestroyed {
  // asset class of the ledger
  string asset_class_id = 1;

  // nft id of the ledger (scope id or nft id)
  string nft_id = 2;
}
```

## Event Attributes

Each event includes standard attributes:

- `module`: Always set to "ledger"
- `action`: The type of event (e.g., "ledger_created", "ledger_updated", "ledger_entry_added", "fund_transfer_with_settlement", "ledger_destroyed")
- `nft_id`: The NFT or Scope identifier associated with the event
- `asset_class_id`: The Scope Specification ID or NFT Class ID

Additional attributes specific to each event type:

### EventLedgerCreated
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `nft_id`: The NFT or Scope identifier

### EventLedgerUpdated
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `nft_id`: The NFT or Scope identifier
- `update_type`: The specific type of update that occurred (status, interest rate, payment, or maturity date)

### EventLedgerEntryAdded
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `nft_id`: The NFT or Scope identifier
- `correlation_id`: The correlation ID of the entry

### EventFundTransferWithSettlement
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `nft_id`: The NFT or Scope identifier
- `correlation_id`: The correlation ID of the transfer

### EventLedgerDestroyed
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `nft_id`: The NFT or Scope identifier

## Event Indexing

Events are indexed for efficient querying:

1. **By Asset Identifier**
   - All events related to a specific NFT or Scope
   - Used for tracking asset-specific activities

2. **By Asset Class**
   - All events related to a specific asset class
   - Used for monitoring class-specific operations

3. **By Event Type**
   - All events of a specific type
   - Used for monitoring specific operations

4. **By Correlation ID**
   - All events related to a specific correlation ID
   - Used for tracking specific transactions

5. **By Update Type**
   - All ledger update events of a specific type
   - Used for monitoring specific types of ledger changes

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
   - Track entry additions and fund transfers

2. **External System Integration**
   - Update accounting systems
   - Sync with external databases
   - Trigger business processes

3. **Audit and Compliance**
   - Maintain audit trails
   - Track changes over time
   - Verify transaction history

4. **Settlement Processing**
   - Process fund transfers
   - Execute settlement instructions
   - Track transfer status

5. **Change Tracking**
   - Monitor specific types of ledger updates
   - Track configuration changes
   - Maintain change history

## Event Subscription

Events can be subscribed to using:
- Tendermint event subscription
- Cosmos SDK event system
- External event listeners

## Event Implementation

Events are implemented as protobuf messages in the `provenance.ledger.v1` package. The events are emitted by the keeper methods and can be consumed by external systems through the standard Cosmos SDK event system.

### Event Emission Examples

```go
// Example of emitting an EventLedgerCreated
ctx.EventManager().EmitTypedEvent(
    types.NewEventLedgerCreated(ledgerKey),
)

// Example of emitting an EventLedgerUpdated with update type
ctx.EventManager().EmitTypedEvent(
    types.NewEventLedgerUpdated(ledgerKey, types.UpdateType_UPDATE_TYPE_STATUS),
)

// Example of emitting an EventLedgerEntryAdded
ctx.EventManager().EmitTypedEvent(
    types.NewEventLedgerEntryAdded(ledgerKey, correlationID),
)

// Example of emitting an EventFundTransferWithSettlement
ctx.EventManager().EmitTypedEvent(
    types.NewEventFundTransferWithSettlement(ledgerKey, correlationID),
)

// Example of emitting an EventLedgerDestroyed
ctx.EventManager().EmitTypedEvent(
    types.NewEventLedgerDestroyed(ledgerKey),
)
```

### Event Consumption Example

```go
// Example of consuming events
func (k Keeper) AfterTx(ctx sdk.Context, tx sdk.Tx, events []abci.Event) {
    for _, event := range events {
        if event.Type == "provenance.ledger.v1.EventLedgerCreated" {
            // Process ledger created event
            for _, attr := range event.Attributes {
                switch string(attr.Key) {
                case "asset_class_id":
                    assetClassID := string(attr.Value)
                case "nft_id":
                    nftID := string(attr.Value)
                }
            }
        } else if event.Type == "provenance.ledger.v1.EventLedgerUpdated" {
            // Process ledger updated event
            for _, attr := range event.Attributes {
                switch string(attr.Key) {
                case "asset_class_id":
                    assetClassID := string(attr.Value)
                case "nft_id":
                    nftID := string(attr.Value)
                case "update_type":
                    updateType := string(attr.Value)
                }
            }
        }
    }
}
```

## Notes

- All events include the asset class ID and NFT ID for proper identification
- The `EventLedgerUpdated` event includes an `update_type` field to specify what was changed
- Events are emitted using the Cosmos SDK typed event system for better type safety
- No events are emitted for ledger class management operations (create, add types)
- No events are emitted for bulk operations (they emit individual events for each ledger/entry)
- Events are emitted synchronously during transaction processing
- All events are indexed and can be queried through the standard Cosmos SDK event system 