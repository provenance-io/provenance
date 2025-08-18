# Ledger State

The Ledger module maintains several types of state to track financial activities and balances for assets (NFTs or Metadata Scopes). The state is organized into collections that store ledger classes, ledgers, entries, and settlements. The module uses the Cosmos SDK Collections framework for efficient state management.

<!-- TOC -->
- [Ledger State](#ledger-state)
  - [Ledger Class](#ledger-class)
  - [Ledger](#ledger)
  - [Ledger Entries](#ledger-entries)
  - [Balances](#balances)
  - [State Storage](#state-storage)
    - [Collections Structure](#collections-structure)
    - [Key Generation](#key-generation)
    - [Balance Calculation](#balance-calculation)

## Ledger Class

The module stores configuration information for each class of assets:

```protobuf
message LedgerClass {
    string ledger_class_id = 1;      // Unique ID for the ledger class
    string asset_class_id = 2;       // Scope Specification ID or NFT Class ID
    string denom = 3;                // Denomination used for all entries in this class
    string maintainer_address = 4;   // Address of the maintainer for the ledger class
}

message LedgerClassEntryType {
    int32 id = 1;                    // Unique ID for the entry type
    string code = 2;                 // Code for the entry type
    string description = 3;          // Description of the entry type
}

message LedgerClassStatusType {
    int32 id = 1;                    // Unique ID for the status type
    string code = 2;                 // Code for the status type
    string description = 3;          // Description of the status type
}

message LedgerClassBucketType {
    int32 id = 1;                    // Unique ID for the bucket type
    string code = 2;                 // Code for the bucket type
    string description = 3;          // Description of the bucket type
}
```

## Ledger

The module stores ledger information for each asset:

```protobuf
message LedgerKey {
    string nft_id = 1;               // NFT or Scope identifier
    string asset_class_id = 2;       // Scope Specification ID or NFT Class ID
}

message Ledger {
    LedgerKey key = 1;               // Unique identifier for the ledger
    string ledger_class_id = 2;      // Reference to the ledger class
    int32 status_type_id = 3;        // Current status of the ledger
    int32 next_pmt_date = 4;         // Next payment date in epoch days
    int64 next_pmt_amt = 5;          // Next payment amount
    int32 interest_rate = 6;         // Interest rate (10000000 = 10.000000%)
    int32 maturity_date = 7;         // Maturity date in epoch days
    DayCountConvention interest_day_count_convention = 8;  // Day count convention
    InterestAccrualMethod interest_accrual_method = 9;     // Interest accrual method
    PaymentFrequency payment_frequency = 10;               // Payment frequency
}
```

## Ledger Entries

Historical ledger entries are stored for each asset:

```protobuf
message LedgerEntry {
    string correlation_id = 1;           // Correlation ID for tracking with external systems (max 50 characters)
    string reverses_correlation_id = 2;  // If this entry reverses another entry, the correlation ID of the reversed entry
    bool is_void = 3;                    // Indicates if this entry is void and should be excluded from balance calculations
    uint32 sequence = 4;                 // Sequence number for ordering entries with same effective date (less than 100)
    int32 entry_type_id = 5;             // The type of ledger entry
    int32 posted_date = 7;               // Posted date in epoch days
    int32 effective_date = 8;            // Effective date in epoch days
    string total_amt = 9;                // Total amount of the entry
    repeated LedgerBucketAmount applied_amounts = 10;  // Amounts applied to different buckets
    repeated BucketBalance balance_amounts = 11;       // Current balances for each bucket after this entry
}

message LedgerBucketAmount {
    int32 bucket_type_id = 1;            // The bucket type ID
    string applied_amt = 2;              // Amount applied to this bucket
}

message BucketBalance {
    int32 bucket_type_id = 1;            // The bucket type ID
    string balance_amt = 2;              // Current balance in this bucket
}
```

## Balances

Balances are calculated on-the-fly from ledger entries rather than stored separately. The `GetBalancesAsOf` function processes all entries up to a specific date to determine the current state of each bucket:

```protobuf
message BucketBalances {
    repeated BucketBalance bucket_balances = 1;  // Current balances for each bucket type
}
```

## State Storage

### Collections Structure
The module uses the Cosmos SDK Collections framework for state storage, which provides type-safe access to the state store:

1. **Ledgers**: Stores ledger configurations
   - Collection: `Ledgers`
   - Key: `ledger_id` (bech32 string)
   - Value: `Ledger`
   - Prefix: `0x01`

2. **LedgerEntries**: Stores historical ledger entries
   - Collection: `LedgerEntries`
   - Key: `(ledger_id, correlation_id)` pair
   - Value: `LedgerEntry`
   - Prefix: `0x02`

3. **FundTransfersWithSettlement**: Stores fund transfer settlement instructions
   - Collection: `FundTransfersWithSettlement`
   - Key: `(ledger_id, settlement_id)` pair
   - Value: `StoredSettlementInstructions`
   - Prefix: `0x08`

4. **LedgerClasses**: Stores ledger class configurations
   - Collection: `LedgerClasses`
   - Key: `ledger_class_id` (string)
   - Value: `LedgerClass`
   - Prefix: `0x03`

5. **LedgerClassEntryTypes**: Stores entry type configurations for each ledger class
   - Collection: `LedgerClassEntryTypes`
   - Key: `(ledger_class_id, entry_type_id)` pair
   - Value: `LedgerClassEntryType`
   - Prefix: `0x04`

6. **LedgerClassStatusTypes**: Stores status type configurations for each ledger class
   - Collection: `LedgerClassStatusTypes`
   - Key: `(ledger_class_id, status_type_id)` pair
   - Value: `LedgerClassStatusType`
   - Prefix: `0x05`

7. **LedgerClassBucketTypes**: Stores bucket type configurations for each ledger class
   - Collection: `LedgerClassBucketTypes`
   - Key: `(ledger_class_id, bucket_type_id)` pair
   - Value: `LedgerClassBucketType`
   - Prefix: `0x06`

### Key Generation

**Ledger ID Generation**: The ledger ID is generated as a bech32 string using the following process:
- Combines `asset_class_id` and `nft_id` with a null byte delimiter (`\x00`)
- Encodes the combined bytes using bech32 with the human-readable part `"ledger"`
- Example: `ledger1w3jhxapdden8gttrd3shxuedd9jqqcm0wdkk7ue3x44hjwtyw5uxzvnhd3ehg73kvec8svmsx3khzur209ex6dtrvack5amv8pehzl09ezy`

**Entry Keys**: Ledger entries are keyed by `(ledger_id, correlation_id)` pairs, where:
- `ledger_id` is the bech32-encoded ledger identifier
- `correlation_id` is the unique identifier for the entry (up to 50 characters)

**Class Configuration Keys**: Class-level configurations use composite keys:
- Entry types: `(ledger_class_id, entry_type_id)`
- Status types: `(ledger_class_id, status_type_id)`
- Bucket types: `(ledger_class_id, bucket_type_id)`

### Balance Calculation

Balances are not stored as separate state but calculated dynamically by:
1. Retrieving all ledger entries for a specific ledger
2. Filtering entries by effective date (up to the requested "as of" date)
3. Processing entries in chronological order
4. Using the `balance_amounts` field from the most recent entry as of the requested date
5. Returning the current bucket balances for that point in time

This approach ensures that balances are always consistent with the ledger entries and provides accurate point-in-time financial state information. 