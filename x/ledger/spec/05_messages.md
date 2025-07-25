# Messages

The Ledger module provides several message types for creating and managing ledger classes, ledgers, and entries.

## MsgCreateLedgerClass

`MsgCreateLedgerClass` creates a new ledger class configuration.

```protobuf
message MsgCreateLedgerClassRequest {
    LedgerClass ledger_class = 1;
    string authority = 2;
}

message MsgCreateLedgerClassResponse {}
```

### Fields
- `ledger_class`: The ledger class configuration to create
  - `ledger_class_id`: The unique identifier for the ledger class
  - `asset_class_id`: The Scope Specification ID or NFT Class ID
  - `denom`: The denomination to use for the ledger entries
  - `maintainer_address`: The address of the maintainer
- `authority`: The address of the authority who is creating the ledger class

### CLI Command
```bash
provenanced tx ledger create-class [ledger-class-fields...] --from [authority]
```

## MsgCreateLedger

`MsgCreateLedger` creates a new ledger for an asset.

```protobuf
message MsgCreateLedgerRequest {
    Ledger ledger = 1;
    string authority = 2;
}

message MsgCreateLedgerResponse {}
```

### Fields
- `ledger`: The ledger configuration to create
  - `key`: The unique identifier for the ledger
    - `nft_id`: The NFT or Scope identifier
    - `asset_class_id`: The Scope Specification ID or NFT Class ID
  - `ledger_class_id`: The ledger class identifier
  - `status_type_id`: The current status of the ledger
  - `next_pmt_date`: The next scheduled payment date in epoch days
  - `next_pmt_amt`: The amount of the next scheduled payment
  - `interest_rate`: The interest rate for the ledger
  - `maturity_date`: The maturity date in epoch days
- `authority`: The address of the authority who is creating the ledger

### CLI Command
```bash
provenanced tx ledger create [ledger-fields...] --from [authority]
```

## MsgAppendLedgerEntry

`MsgAppendLedgerEntry` adds one or more new entries to an existing ledger.

```protobuf
message MsgAppendLedgerEntryRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    repeated LedgerEntry entries = 3;
    string authority = 4;
}

message MsgAppendLedgerEntryResponse {}
```

### Fields
- `nft_id`: The NFT or Scope identifier
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `entries`: One or more ledger entries to append
  - `correlation_id`: Unique identifier for tracking with external systems (max 50 characters)
  - `reverses_correlation_id`: If this entry reverses another entry, the correlation ID of the reversed entry
  - `is_void`: Indicates if this entry is void and should be excluded from balance calculations
  - `sequence`: Sequence number for ordering entries with same effective date (less than 100)
  - `entry_type_id`: The type of ledger entry
  - `posted_date`: The date when the entry was recorded (epoch days)
  - `effective_date`: The date when the entry takes effect (epoch days)
  - `total_amt`: The total amount of the entry
  - `applied_amounts`: List of amounts applied to different buckets
    - `bucket_type_id`: The bucket type ID
    - `applied_amt`: Amount applied to this bucket
  - `bucket_balances`: Current balances for each bucket
    - `bucket_type_id`: The bucket type ID
    - `balance`: Current balance in this bucket
- `authority`: The address of the authority who is appending the entries

### CLI Command
Only allows a single entry, use the RPC endpoint to add multiple entries in a single call.
```bash
provenanced tx ledger append [nft-id] [asset-class-id] [entries-fields...] --from [authority]
```

## MsgDestroyLedger

`MsgDestroyLedger` removes a ledger and all associated data.

```protobuf
message MsgDestroyLedgerRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    string authority = 3;
}

message MsgDestroyLedgerResponse {}
```

### Fields
- `nft_id`: The NFT or Scope identifier
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `authority`: The address of the authority who is destroying the ledger

### CLI Command
```bash
provenanced tx ledger destroy [nft-id] [asset-class-id] --from [authority]
```

## Message Validation

All messages are validated before processing:

1. **MsgCreateLedgerClass**
   - Ledger class configuration must be valid
   - Asset class ID must be valid
   - Denomination must be valid
   - Maintainer address must be valid
   - Authority must have permission

2. **MsgCreateLedger**
   - Ledger configuration must be valid
   - Asset identifiers must be valid
   - Ledger class must exist
   - Authority must have permission
   - Dates must be in correct format
   - Amounts must be valid

3. **MsgAppendLedgerEntry**
   - Asset identifiers must be valid
   - Ledger must exist
   - Entries must be valid
   - Authority must have permission
   - Correlation IDs must be unique
   - Sequences must be valid
   - Bucket types must be valid
   - Amounts must be valid

4. **MsgDestroyLedger**
   - Asset identifiers must be valid
   - Ledger must exist
   - Authority must have permission
   - All associated data must be properly cleaned up 