# State

The Ledger module maintains several types of state to track financial activities and balances for assets (NFTs or Metadata Scopes).

## State Structure

### Ledger Class
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

### Ledger
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
    int32 interest_rate = 6;         // Interest rate
    int32 maturity_date = 7;         // Maturity date in epoch days
}
```

### Ledger Entries
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
    map<int32, BucketBalance> bucket_balances = 11;    // Current balances for each bucket
}

message LedgerBucketAmount {
    int32 bucket_type_id = 1;            // The bucket type ID
    string applied_amt = 2;               // Amount applied to this bucket
}

message BucketBalance {
    int32 bucket_type_id = 1;            // The bucket type ID
    string balance = 2;                   // Current balance in this bucket
}
```

### Balances
Current balances for configured buckets:

```protobuf
message Balances {
    repeated BucketBalance bucket_balances = 1;  // Current balances for each bucket type
}
```

## State Storage

### KV Store Structure
The module uses the following collections for state storage:

1. `LedgerClasses`: Stores ledger class configurations
   - Prefix: "ledger_classes"
   - Key: `ledger_class_id`
   - Value: `LedgerClass` protobuf

2. `LedgerClassEntryTypes`: Stores entry type definitions
   - Prefix: "ledger_class_entry_types"
   - Key: `ledger_class_id + entry_type_id`
   - Value: `LedgerClassEntryType` protobuf

3. `LedgerClassStatusTypes`: Stores status type definitions
   - Prefix: "ledger_class_status_types"
   - Key: `ledger_class_id + status_type_id`
   - Value: `LedgerClassStatusType` protobuf

4. `LedgerClassBucketTypes`: Stores bucket type definitions
   - Prefix: "ledger_class_bucket_types"
   - Key: `ledger_class_id + bucket_type_id`
   - Value: `LedgerClassBucketType` protobuf

5. `Ledgers`: Stores ledger configurations
   - Prefix: "ledgers"
   - Key: `nft_id + asset_class_id`
   - Value: `Ledger` protobuf

6. `LedgerEntries`: Stores ledger entries
   - Prefix: "ledger_entries"
   - Key: `nft_id + asset_class_id + correlation_id`
   - Value: `LedgerEntry` protobuf

7. `Balances`: Stores current balances
   - Prefix: "ledger_balances"
   - Key: `nft_id + asset_class_id`
   - Value: `Balances` protobuf

## State Transitions

State transitions occur when:

1. **Ledger Class Creation**
   - New ledger class configuration is stored
   - Entry types are defined
   - Status types are defined
   - Bucket types are defined
   - Creation event is emitted

2. **Ledger Creation**
   - New ledger configuration is stored
   - Initial balances are set
   - Creation event is emitted

3. **Entry Addition**
   - New entry is stored with correlation ID
   - Balances are updated
   - Entry event is emitted

4. **Balance Updates**
   - Bucket balances are updated
   - Balance update event is emitted

5. **Configuration Changes**
   - Payment schedule updates
   - Interest rate updates
   - Maturity date updates
   - Status updates
   - Configuration update event is emitted

## State Access

State can be accessed through:

1. **Query Endpoints**
   - Get ledger class configuration
   - Get ledger configuration
   - Get ledger entries
   - Get current balances
   - Filter and search entries

2. **Transaction Handlers**
   - Create ledger class
   - Create ledger
   - Add entries
   - Update balances
   - Modify configuration

3. **Event System**
   - Track state changes
   - Monitor activities
   - Maintain audit trail 