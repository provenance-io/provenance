# Ledger Messages

The Ledger module provides several message types for creating and managing ledger classes, ledgers, entries, and transfers. These messages allow authorized users to perform various operations on ledger data.

<!-- TOC -->
  - [Ledger Management](#ledger-management)
  - [Entry Management](#entry-management)
  - [Class Management](#class-management)
  - [Transfer Management](#transfer-management)
  - [Bulk Operations](#bulk-operations)
  - [Message Validation](#message-validation)

## Ledger Management

### MsgCreate

`MsgCreate` creates a new ledger for an asset.

```protobuf
message MsgCreateLedgerRequest {
    Ledger ledger = 1;
    string authority = 2;
}

message MsgCreateLedgerResponse {}
```

#### Fields
- `ledger`: The ledger configuration to create
  - `key`: The unique identifier for the ledger
    - `nft_id`: The NFT or Scope identifier
    - `asset_class_id`: The Scope Specification ID or NFT Class ID
  - `ledger_class_id`: The ledger class identifier
  - `status_type_id`: The current status of the ledger
  - `next_pmt_date`: The next scheduled payment date in epoch days
  - `next_pmt_amt`: The amount of the next scheduled payment
  - `interest_rate`: The interest rate for the ledger (10000000 = 10.000000%)
  - `maturity_date`: The maturity date in epoch days
  - `interest_day_count_convention`: Day count convention for interest calculations
  - `interest_accrual_method`: Method used for interest accrual
  - `payment_frequency`: Frequency of scheduled payments
- `authority`: The address of the authority who is creating the ledger

#### CLI Command
```bash
provenanced tx ledger create <asset_class_id> <nft_id> <ledger_class_id> <status_type_id> [flags] --from <authority>
```

**Flags:**
- `--next-pmt-date`: Next payment date (YYYY-MM-DD)
- `--next-pmt-amt`: Next payment amount
- `--interest-rate`: Interest rate (10000000 = 10.000000%)
- `--maturity-date`: Maturity date (YYYY-MM-DD)
- `--day-count-convention`: Day count convention (actual-365, actual-360, thirty-360, actual-actual, days-365, days-360)
- `--interest-accrual-method`: Interest accrual method (simple, compound, daily, monthly, quarterly, annual, continuous)
- `--payment-frequency`: Payment frequency (daily, weekly, monthly, quarterly, annually)

### MsgDestroy

`MsgDestroy` removes a ledger and all associated data.

```protobuf
message MsgDestroyRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    string authority = 3;
}

message MsgDestroyResponse {}
```

#### Fields
- `nft_id`: The NFT or Scope identifier
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `authority`: The address of the authority who is destroying the ledger

#### CLI Command
```bash
provenanced tx ledger destroy <asset_class_id> <nft_id> --from <authority>
```

## Entry Management

### MsgAppend

`MsgAppend` adds one or more new entries to an existing ledger.

```protobuf
message MsgAppendRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    repeated LedgerEntry entries = 3;
    string authority = 4;
}

message MsgAppendResponse {}
```

#### Fields
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
  - `balance_amounts`: Current balances for each bucket after this entry
    - `bucket_type_id`: The bucket type ID
    - `balance_amt`: Current balance in this bucket
- `authority`: The address of the authority who is appending the entries

#### CLI Command
```bash
provenanced tx ledger append <asset_class_id> <nft_id> <json_file_path> --from <authority>
```

**Note:** The JSON file should contain an array of ledger entries with the required fields.

## Class Management

### MsgCreateLedgerClass

`MsgCreateLedgerClass` creates a new ledger class configuration.

```protobuf
message MsgCreateLedgerClassRequest {
    LedgerClass ledger_class = 1;
    string authority = 2;
}

message MsgCreateLedgerClassResponse {}
```

#### Fields
- `ledger_class`: The ledger class configuration to create
  - `ledger_class_id`: The unique identifier for the ledger class
  - `asset_class_id`: The Scope Specification ID or NFT Class ID
  - `denom`: The denomination to use for the ledger entries
  - `maintainer_address`: The address of the maintainer
- `authority`: The address of the authority who is creating the ledger class

#### CLI Command
```bash
provenanced tx ledger create-class <ledger_class_id> <asset_class_id> <denom> --from <authority>
```

### MsgAddLedgerClassEntryType

`MsgAddLedgerClassEntryType` adds an entry type to a ledger class.

```protobuf
message MsgAddLedgerClassEntryTypeRequest {
    string ledger_class_id = 1;
    LedgerClassEntryType entry_type = 2;
    string authority = 3;
}

message MsgAddLedgerClassEntryTypeResponse {}
```

#### Fields
- `ledger_class_id`: The ledger class identifier
- `entry_type`: The entry type to add
  - `id`: The unique ID for the entry type
  - `code`: The code for the entry type
  - `description`: The description of the entry type
- `authority`: The address of the authority who is adding the entry type

#### CLI Command
```bash
provenanced tx ledger add-entry-type <ledger_class_id> <id> <code> <description> --from <authority>
```

### MsgAddLedgerClassStatusType

`MsgAddLedgerClassStatusType` adds a status type to a ledger class.

```protobuf
message MsgAddLedgerClassStatusTypeRequest {
    string ledger_class_id = 1;
    LedgerClassStatusType status_type = 2;
    string authority = 3;
}

message MsgAddLedgerClassStatusTypeResponse {}
```

#### Fields
- `ledger_class_id`: The ledger class identifier
- `status_type`: The status type to add
  - `id`: The unique ID for the status type
  - `code`: The code for the status type
  - `description`: The description of the status type
- `authority`: The address of the authority who is adding the status type

#### CLI Command
```bash
provenanced tx ledger add-status-type <ledger_class_id> <id> <code> <description> --from <authority>
```

### MsgAddLedgerClassBucketType

`MsgAddLedgerClassBucketType` adds a bucket type to a ledger class.

```protobuf
message MsgAddLedgerClassBucketTypeRequest {
    string ledger_class_id = 1;
    LedgerClassBucketType bucket_type = 2;
    string authority = 3;
}

message MsgAddLedgerClassBucketTypeResponse {}
```

#### Fields
- `ledger_class_id`: The ledger class identifier
- `bucket_type`: The bucket type to add
  - `id`: The unique ID for the bucket type
  - `code`: The code for the bucket type
  - `description`: The description of the bucket type
- `authority`: The address of the authority who is adding the bucket type

#### CLI Command
```bash
provenanced tx ledger add-bucket-type <ledger_class_id> <id> <code> <description> --from <authority>
```

## Transfer Management

### MsgTransferFundsWithSettlement

`MsgTransferFundsWithSettlement` transfers funds with settlement instructions.

```protobuf
message MsgTransferFundsWithSettlementRequest {
    repeated FundTransferWithSettlement transfers = 1;
    string authority = 2;
}

message MsgTransferFundsWithSettlementResponse {}
```

#### Fields
- `transfers`: List of fund transfers with settlement instructions
  - `nft_id`: The NFT or Scope identifier
  - `asset_class_id`: The Scope Specification ID or NFT Class ID
  - `correlation_id`: The correlation ID for the transfer
  - `settlement_instructions`: The settlement instructions
- `authority`: The address of the authority who is performing the transfer

#### CLI Command
```bash
provenanced tx ledger xfer <fund_transfers_json_file> --from <authority>
```

**Note:** The JSON file should contain an array of fund transfer objects with the required fields.

## Bulk Operations

### MsgBulkCreate

`MsgBulkCreate` creates multiple ledgers and entries in a single transaction.

```protobuf
message MsgBulkCreateRequest {
    repeated LedgerToEntries ledger_to_entries = 1;
    string authority = 2;
}

message MsgBulkCreateResponse {}
```

#### Fields
- `ledger_to_entries`: List of ledgers with their associated entries
  - `ledger_key`: The unique identifier for the ledger
  - `ledger`: The ledger configuration
  - `entries`: List of ledger entries to create
- `authority`: The address of the authority who is performing the bulk creation

#### CLI Command
```bash
provenanced tx ledger bulk-create <ledger_entries_json_file> --from <authority>
```

**Note:** The JSON file should contain an array of ledger-to-entries objects with the required fields.

## Message Validation

All messages are validated before processing:

### Ledger Management
1. **MsgCreate**
   - Ledger configuration must be valid
   - Asset identifiers must be valid
   - Ledger class must exist
   - Authority must have permission
   - Dates must be in correct format
   - Amounts must be valid

2. **MsgDestroy**
   - Asset identifiers must be valid
   - Ledger must exist
   - Authority must have permission
   - All associated data must be properly cleaned up

### Entry Management
1. **MsgAppend**
   - Asset identifiers must be valid
   - Ledger must exist
   - Entries must be valid
   - Authority must have permission
   - Correlation IDs must be unique
   - Sequences must be valid
   - Bucket types must be valid
   - Amounts must be valid

### Class Management
1. **MsgCreateLedgerClass**
   - Ledger class configuration must be valid
   - Asset class ID must be valid
   - Denomination must be valid
   - Maintainer address must be valid
   - Authority must have permission

2. **MsgAddLedgerClassEntryType**
   - Ledger class must exist
   - Entry type must be valid
   - Authority must have permission

3. **MsgAddLedgerClassStatusType**
   - Ledger class must exist
   - Status type must be valid
   - Authority must have permission

4. **MsgAddLedgerClassBucketType**
   - Ledger class must exist
   - Bucket type must be valid
   - Authority must have permission

### Transfer Management
1. **MsgTransferFundsWithSettlement**
   - Asset identifiers must be valid
   - Ledger must exist
   - Correlation ID must be valid
   - Settlement instructions must be valid
   - Authority must have permission

### Bulk Operations
1. **MsgBulkCreate**
   - All ledger configurations must be valid
   - All entries must be valid
   - Authority must have permission
   - Transaction size must be within limits 