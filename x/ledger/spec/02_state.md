# State

The Ledger module maintains several types of state to track financial activities and balances for NFTs.

## State Structure

### Ledger Configuration
The module stores configuration information for each NFT's ledger:

```protobuf
message Ledger {
    string nft_address = 1;
    string denom = 2;
}
```

### Ledger Entries
Historical ledger entries are stored for each NFT:

```protobuf
message LedgerEntry {
    string uuid = 1;
    LedgerEntryType type = 2;
    google.protobuf.Timestamp posted_date = 3;
    google.protobuf.Timestamp effective_date = 4;
    string amt = 5;
    string prin_applied_amt = 6;
    string prin_bal_amt = 7;
    string int_applied_amt = 8;
    string int_bal_amt = 9;
    string other_applied_amt = 10;
    string other_bal_amt = 11;
}
```

### Fund Transfers
Fund transfer information is stored for processing:

```protobuf
message FundTransfer {
    string nft_address = 1;
    string ledger_entry_uuid = 2;
    string amount = 3;
    FundingTransferStatus status = 4;
    string memo = 5;
    int64 settlement_block = 6;
}
```

## State Storage

### KV Store Structure
The module uses the following collections for state storage:

1. `Ledgers`: Stores ledger configurations
   - Prefix: "ledgers"
   - Key: `nft_address`
   - Value: `Ledger` protobuf

2. `LedgerEntries`: Stores ledger entries
   - Prefix: "ledger_entries"
   - Key: `nft_address + entry_uuid` (as a pair)
   - Value: `LedgerEntry` protobuf

3. `FundTransfers`: Stores fund transfer information
   - Prefix: "ledger_fund_transfers"
   - Key: `nft_address`
   - Value: `FundTransfer` protobuf

## State Transitions

State transitions occur when:

1. **Ledger Creation**
   - New ledger configuration is stored
   - Initial balances are set
   - Creation event is emitted

2. **Entry Addition**
   - New entry is stored
   - Balances are updated
   - Entry event is emitted

3. **Balance Updates**
   - Principal balance changes
   - Interest balance changes
   - Other balance changes
   - Balance update event is emitted

4. **Configuration Changes**
   - Denomination updates
   - Configuration update event is emitted

5. **Fund Transfer Processing**
   - Fund transfer status updates
   - Settlement instructions are processed
   - Transfer events are emitted

## State Access

State can be accessed through:

1. **Query Endpoints**
   - Get ledger configuration
   - Get ledger entries
   - Get current balances
   - Filter and search entries
   - Get fund transfer status

2. **Transaction Handlers**
   - Create ledger
   - Add entries
   - Update balances
   - Modify configuration
   - Process fund transfers

3. **Event System**
   - Track state changes
   - Monitor activities
   - Maintain audit trail 