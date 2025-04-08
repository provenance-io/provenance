# Events

The Ledger module emits events to track important state changes and activities. These events can be used by external systems to monitor ledger activities and maintain synchronization.

## Event Structure

All events include the following common attributes:
- `module`: Always set to "ledger"
- `action`: The type of action that triggered the event
- `nft_address`: The address of the NFT associated with the event

## Event Types

### Ledger Created
Emitted when a new ledger is created for an NFT.

```protobuf
message EventLedgerCreated {
    string nft_address = 1;
    string denom = 2;
}
```

Attributes:
- `nft_address`: The address of the NFT
- `denom`: The denomination used for the ledger

### Ledger Entry Added
Emitted when a new entry is added to a ledger.

```protobuf
message EventLedgerEntryAdded {
    string nft_address = 1;
    string entry_uuid = 2;
    LedgerEntryType entry_type = 3;
    google.protobuf.Timestamp posted_date = 4;
    google.protobuf.Timestamp effective_date = 5;
    string amount = 6;
}
```

Attributes:
- `nft_address`: The address of the NFT
- `entry_uuid`: The unique identifier of the entry
- `entry_type`: The type of the entry
- `posted_date`: The date the entry was posted
- `effective_date`: The date the entry takes effect
- `amount`: The total amount of the entry

### Balance Updated
Emitted when balances are updated due to a ledger entry.

```protobuf
message EventBalanceUpdated {
    string nft_address = 1;
    string principal_balance = 2;
    string interest_balance = 3;
    string other_balance = 4;
}
```

Attributes:
- `nft_address`: The address of the NFT
- `principal_balance`: The new principal balance
- `interest_balance`: The new interest balance
- `other_balance`: The new other balance

### Ledger Configuration Updated
Emitted when ledger configuration is modified.

```protobuf
message EventLedgerConfigUpdated {
    string nft_address = 1;
    string denom = 2;
    string previous_denom = 3;
}
```

Attributes:
- `nft_address`: The address of the NFT
- `denom`: The new denomination
- `previous_denom`: The previous denomination (if changed)

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