# State

The Ledger module maintains several types of state to track financial activities and balances for NFTs.

## State Structure

### Ledger Configuration
The module stores configuration information for each NFT's ledger:

```protobuf
message Ledger {
    string nft_address = 1;
    string denom = 2;
    int32 next_pmt_date = 3;  // Next payment date in epoch days
    int64 next_pmt_amt = 4;   // Next payment amount
    string status = 5;         // Status of the ledger
    int32 interest_rate = 6;  // Interest rate
    int32 maturity_date = 7;  // Maturity date in epoch days
}
```

### Ledger Entries
Historical ledger entries are stored for each NFT:

```protobuf
message LedgerEntry {
    string correlation_id = 1;  // Correlation ID for tracking with external systems (max 50 characters)
    uint32 sequence = 2;        // Sequence number for ordering entries with same effective date
    LedgerEntryType type = 3;
    string sub_type = 4;
    int32 posted_date = 5;     // Posted date in epoch days
    int32 effective_date = 6;  // Effective date in epoch days
    int64 total_amt = 7;       // Total amount
    repeated LedgerBucketAmount applied_amounts = 8;
}

message LedgerBucketAmount {
    string bucket = 1;      // The bucket name (e.g., "principal", "interest", "other")
    int64 applied_amt = 2;  // Amount applied to this bucket
    int64 balance_amt = 3;  // Remaining balance in this bucket
}
```

### Fund Transfers
Fund transfer information is stored for processing:

```protobuf
message FundTransfer {
    string nft_address = 1;
    string ledger_entry_correlation_id = 2;  // Correlation ID of the associated ledger entry
    string amount = 3;
    FundingTransferStatus status = 4;
    string memo = 5;
    int64 settlement_block = 6;  // Minimum block height or timestamp for settlement
}

message FundTransferWithSettlement {
    string nft_address = 1;
    string ledger_entry_correlation_id = 2;
    repeated SettlementInstruction settlementInstructions = 3;
}

message SettlementInstruction {
    string amount = 1;
    string recipient_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];  // The recipient's blockchain address
    string memo = 3;               // Optional memo or note for the transaction
    int64 settlement_block = 4;    // Minimum block height or timestamp for settlement
}
```

### Balances
Current balances for principal, interest, and other amounts:

```protobuf
message Balances {
    repeated BucketBalance bucket_balances = 1;
}

message BucketBalance {
    string bucket = 1;  // The bucket name (e.g., "principal", "interest", "other")
    int64 balance = 2;  // Current balance in this bucket
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
   - Key: `nft_address + correlation_id` (as a pair)
   - Value: `LedgerEntry` protobuf

3. `FundTransfers`: Stores fund transfer information
   - Prefix: "ledger_fund_transfers"
   - Key: `nft_address`
   - Value: `FundTransfer` or `FundTransferWithSettlement` protobuf

4. `Balances`: Stores current balances
   - Prefix: "ledger_balances"
   - Key: `nft_address`
   - Value: `Balances` protobuf

## State Transitions

State transitions occur when:

1. **Ledger Creation**
   - New ledger configuration is stored
   - Initial balances are set
   - Creation event is emitted

2. **Entry Addition**
   - New entry is stored with correlation ID
   - Balances are updated
   - Entry event is emitted

3. **Balance Updates**
   - Principal balance changes
   - Interest balance changes
   - Other balance changes
   - Balance update event is emitted

4. **Configuration Changes**
   - Denomination updates
   - Payment schedule updates
   - Interest rate updates
   - Maturity date updates
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