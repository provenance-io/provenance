# Ledger Messages

The Ledger module provides several message types for creating and managing ledger classes, ledgers, entries, and transfers. These messages allow authorized users to perform various operations on ledger data.

<!-- TOC -->
  - [Ledger Management](#ledger-management)
  - [Entry Management](#entry-management)
  - [Class Management](#class-management)
  - [Transfer Management](#transfer-management)
  - [Message Validation](#message-validation)

## Ledger Management

### MsgCreateLedger

`MsgCreateLedger` creates a new ledger for an asset.

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
provenanced tx ledger create [ledger-fields...] --from [authority]
```

### MsgUpdateLedgerStatus

`MsgUpdateLedgerStatus` updates the status of an existing ledger.

```protobuf
message MsgUpdateLedgerStatusRequest {
    LedgerKey key = 1;
    int32 status_type_id = 2;
    string authority = 3;
}

message MsgUpdateLedgerStatusResponse {}
```

#### Fields
- `key`: The unique identifier for the ledger
  - `nft_id`: The NFT or Scope identifier
  - `asset_class_id`: The Scope Specification ID or NFT Class ID
- `status_type_id`: The new status type ID
- `authority`: The address of the authority who is updating the status

#### CLI Command
```bash
provenanced tx ledger update-status [nft-id] [asset-class-id] [status-type-id] --from [authority]
```

### MsgUpdateLedgerInterestRate

`MsgUpdateLedgerInterestRate` updates the interest rate of an existing ledger.

```protobuf
message MsgUpdateLedgerInterestRateRequest {
    LedgerKey key = 1;
    int32 interest_rate = 2;
    string authority = 3;
}

message MsgUpdateLedgerInterestRateResponse {}
```

#### Fields
- `key`: The unique identifier for the ledger
- `interest_rate`: The new interest rate (10000000 = 10.000000%)
- `authority`: The address of the authority who is updating the interest rate

#### CLI Command
```bash
provenanced tx ledger update-interest-rate [nft-id] [asset-class-id] [interest-rate] --from [authority]
```

### MsgUpdateLedgerPayment

`MsgUpdateLedgerPayment` updates the payment schedule of an existing ledger.

```protobuf
message MsgUpdateLedgerPaymentRequest {
    LedgerKey key = 1;
    int32 next_pmt_date = 2;
    int64 next_pmt_amt = 3;
    string authority = 4;
}

message MsgUpdateLedgerPaymentResponse {}
```

#### Fields
- `key`: The unique identifier for the ledger
- `next_pmt_date`: The new next payment date in epoch days
- `next_pmt_amt`: The new next payment amount
- `authority`: The address of the authority who is updating the payment

#### CLI Command
```bash
provenanced tx ledger update-payment [nft-id] [asset-class-id] [next-pmt-date] [next-pmt-amt] --from [authority]
```

### MsgUpdateLedgerMaturityDate

`MsgUpdateLedgerMaturityDate` updates the maturity date of an existing ledger.

```protobuf
message MsgUpdateLedgerMaturityDateRequest {
    LedgerKey key = 1;
    int32 maturity_date = 2;
    string authority = 3;
}

message MsgUpdateLedgerMaturityDateResponse {}
```

#### Fields
- `key`: The unique identifier for the ledger
- `maturity_date`: The new maturity date in epoch days
- `authority`: The address of the authority who is updating the maturity date

#### CLI Command
```bash
provenanced tx ledger update-maturity-date [nft-id] [asset-class-id] [maturity-date] --from [authority]
```

### MsgDestroyLedger

`MsgDestroyLedger` removes a ledger and all associated data.

```protobuf
message MsgDestroyLedgerRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    string authority = 3;
}

message MsgDestroyLedgerResponse {}
```

#### Fields
- `nft_id`: The NFT or Scope identifier
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `authority`: The address of the authority who is destroying the ledger

#### CLI Command
```bash
provenanced tx ledger destroy [nft-id] [asset-class-id] --from [authority]
```

## Entry Management

### MsgAppendLedgerEntry

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
  - `bucket_balances`: Current balances for each bucket
    - `bucket_type_id`: The bucket type ID
    - `balance`: Current balance in this bucket
- `authority`: The address of the authority who is appending the entries

#### CLI Command
Only allows a single entry, use the RPC endpoint to add multiple entries in a single call.
```bash
provenanced tx ledger append [nft-id] [asset-class-id] [entries-fields...] --from [authority]
```

### MsgUpdateLedgerBalances

`MsgUpdateLedgerBalances` updates the balances of an existing ledger.

```protobuf
message MsgUpdateLedgerBalancesRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    repeated BucketBalance bucket_balances = 3;
    string authority = 4;
}

message MsgUpdateLedgerBalancesResponse {}
```

#### Fields
- `nft_id`: The NFT or Scope identifier
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `bucket_balances`: List of bucket balances to update
  - `bucket_type_id`: The bucket type ID
  - `balance`: The new balance for this bucket
- `authority`: The address of the authority who is updating the balances

#### CLI Command
```bash
provenanced tx ledger update-balances [nft-id] [asset-class-id] [bucket-balances...] --from [authority]
```

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
provenanced tx ledger create-class [ledger-class-fields...] --from [authority]
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
provenanced tx ledger add-entry-type [ledger-class-id] [entry-type-fields...] --from [authority]
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
provenanced tx ledger add-status-type [ledger-class-id] [status-type-fields...] --from [authority]
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
provenanced tx ledger add-bucket-type [ledger-class-id] [bucket-type-fields...] --from [authority]
```

## Transfer Management

### MsgTransferFundsWithSettlement

`MsgTransferFundsWithSettlement` transfers funds with settlement instructions.

```protobuf
message MsgTransferFundsWithSettlementRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    string correlation_id = 3;
    string settlement_instructions = 4;
    string authority = 5;
}

message MsgTransferFundsWithSettlementResponse {}
```

#### Fields
- `nft_id`: The NFT or Scope identifier
- `asset_class_id`: The Scope Specification ID or NFT Class ID
- `correlation_id`: The correlation ID for the transfer
- `settlement_instructions`: The settlement instructions
- `authority`: The address of the authority who is performing the transfer

#### CLI Command
```bash
provenanced tx ledger transfer-funds [nft-id] [asset-class-id] [correlation-id] [settlement-instructions] --from [authority]
```

## Message Validation

All messages are validated before processing:

### Ledger Management
1. **MsgCreateLedger**
   - Ledger configuration must be valid
   - Asset identifiers must be valid
   - Ledger class must exist
   - Authority must have permission
   - Dates must be in correct format
   - Amounts must be valid

2. **MsgUpdateLedgerStatus**
   - Asset identifiers must be valid
   - Ledger must exist
   - Status type must be valid
   - Authority must have permission

3. **MsgUpdateLedgerInterestRate**
   - Asset identifiers must be valid
   - Ledger must exist
   - Interest rate must be valid
   - Authority must have permission

4. **MsgUpdateLedgerPayment**
   - Asset identifiers must be valid
   - Ledger must exist
   - Payment date and amount must be valid
   - Authority must have permission

5. **MsgUpdateLedgerMaturityDate**
   - Asset identifiers must be valid
   - Ledger must exist
   - Maturity date must be valid
   - Authority must have permission

6. **MsgDestroyLedger**
   - Asset identifiers must be valid
   - Ledger must exist
   - Authority must have permission
   - All associated data must be properly cleaned up

### Entry Management
1. **MsgAppendLedgerEntry**
   - Asset identifiers must be valid
   - Ledger must exist
   - Entries must be valid
   - Authority must have permission
   - Correlation IDs must be unique
   - Sequences must be valid
   - Bucket types must be valid
   - Amounts must be valid

2. **MsgUpdateLedgerBalances**
   - Asset identifiers must be valid
   - Ledger must exist
   - Bucket types must be valid
   - Authority must have permission
   - Balances must be valid

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