# State

The Ledger module maintains several types of state to track financial activities and balances for NFTs.

## State Structure

### Ledger Configuration
The module stores configuration information for each NFT's ledger:

```protobuf
message Ledger {
    string nft_address = 1;
    string denom = 2;
    string next_pmt_date = 3;  // Next payment date in ISO 8601 format: YYYY-MM-DD
    string next_pmt_amt = 4;   // Next payment amount
    string status = 5;         // Status of the ledger
    string interest_rate = 6;  // Interest rate
    string maturity_date = 7;  // Maturity date in ISO 8601 format: YYYY-MM-DD
}
```

### Ledger Entries
Historical ledger entries are stored for each NFT:

```protobuf
message LedgerEntry {
    string correlation_id = 1;  // Correlation ID for tracking with external systems (max 50 characters)
    uint32 sequence = 2;        // Sequence number for ordering entries with same effective date
    LedgerEntryType type = 3;
    string posted_date = 4;     // Posted date in ISO 8601 format: YYYY-MM-DD
    string effective_date = 5;  // Effective date in ISO 8601 format: YYYY-MM-DD
    string amt = 6;             // Total amount
    string prin_applied_amt = 7;  // Principal applied amount
    string prin_bal_amt = 8;      // Principal balance amount
    string int_applied_amt = 9;   // Interest applied amount
    string int_bal_amt = 10;      // Interest balance amount
    string other_applied_amt = 11; // Other applied amount
    string other_bal_amt = 12;     // Other balance amount
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
    string recipient_address = 2;  // The recipient's blockchain address
    string memo = 3;               // Optional memo or note for the transaction
    int64 settlement_block = 4;    // Minimum block height or timestamp for settlement
}
```

### Balances
Current balances for principal, interest, and other amounts:

```protobuf
message Balances {
    string principal = 1;  // Current principal balance
    string interest = 2;   // Current interest balance
    string other = 3;      // Current other balance
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