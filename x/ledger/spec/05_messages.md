# Messages

The Ledger module provides several message types for creating and managing NFT ledgers, entries, and fund transfers.

## MsgCreate

`MsgCreate` creates a new ledger for an NFT.

```protobuf
message MsgCreateRequest {
    Ledger ledger = 1;
    string authority = 2;
}

message MsgCreateResponse {}
```

### Fields
- `ledger`: The ledger configuration to create
  - `nft_address`: The address of the NFT to create a ledger for
  - `denom`: The denomination to use for the ledger entries
  - `next_pmt_date`: The next scheduled payment date in epoch days (int32)
  - `next_pmt_amt`: The amount of the next scheduled payment (int64)
  - `status`: The current status of the ledger
  - `interest_rate`: The interest rate for the ledger (int32)
  - `maturity_date`: The maturity date in epoch days (int32)
- `authority`: The address of the authority who is creating the ledger

### CLI Command
```bash
provenanced tx ledger create [ledger-fields...] --from [authority]
```

## MsgAppend

`MsgAppend` adds one or more new entries to an existing ledger.

```protobuf
message MsgAppendRequest {
    string nft_address = 1;
    repeated LedgerEntry entries = 2;
    string authority = 3;
}

message MsgAppendResponse {}
```

### Fields
- `nft_address`: The address of the NFT whose ledger to append to
- `entries`: One or more ledger entries to append
  - `correlation_id`: Unique identifier for tracking with external systems (max 50 characters)
  - `sequence`: Sequence number for ordering entries with same effective date (less than 100)
  - `type`: The type of ledger entry (see LedgerEntryType)
  - `sub_type`: Additional classification for the entry
  - `posted_date`: The date when the entry was recorded (epoch days, int32)
  - `effective_date`: The date when the entry takes effect (epoch days, int32)
  - `total_amt`: The total amount of the entry (int64)
  - `applied_amounts`: List of amounts applied to different buckets
    - `bucket`: The bucket name (e.g., "principal", "interest", "other")
    - `applied_amt`: Amount applied to this bucket (int64)
    - `balance_amt`: Remaining balance in this bucket (int64)
- `authority`: The address of the authority who is appending the entries

### CLI Command
Only allows a single entry, use the RPC endpoint to add multiple entries in a single call.
```bash
provenanced tx ledger append [nft-address] [entries fields...] --from [authority]
```

## MsgProcessFundTransfers

`MsgProcessFundTransfers` processes multiple fund transfers (payments and disbursements).

```protobuf
message MsgProcessFundTransfersRequest {
    string authority = 1;
    repeated FundTransfer transfers = 2;
}

message MsgProcessFundTransfersResponse {}
```

### Fields
- `authority`: The address of the authority processing the transfers
- `transfers`: List of fund transfers to process
  - `nft_address`: The address of the NFT
  - `ledger_entry_correlation_id`: Correlation ID of the associated ledger entry
  - `amount`: The amount to transfer
  - `status`: The status of the transfer (see FundingTransferStatus)
  - `memo`: Optional memo for the transfer
  - `settlement_block`: Minimum block height or timestamp for settlement (int64)

### CLI Command
```bash
provenanced tx ledger process-transfers [transfers-json] --from [authority]
```

## MsgProcessFundTransfersWithSettlement

`MsgProcessFundTransfersWithSettlement` processes fund transfers with manual settlement instructions.

```protobuf
message MsgProcessFundTransfersWithSettlementRequest {
    string authority = 1;
    repeated FundTransferWithSettlement transfers = 2;
}

message MsgProcessFundTransfersResponse {}
```

### Fields
- `authority`: The address of the authority processing the transfers
- `transfers`: List of fund transfers with settlement instructions to process
  - `nft_address`: The address of the NFT
  - `ledger_entry_correlation_id`: Correlation ID of the associated ledger entry
  - `settlementInstructions`: List of settlement instructions
    - `amount`: The amount to transfer
    - `recipient_address`: The recipient's blockchain address
    - `memo`: Optional memo for the transaction
    - `settlement_block`: Minimum block height or timestamp for settlement (int64)

### CLI Command
```bash
provenanced tx ledger process-transfers-with-settlement [transfers-json] --from [authority]
```

## MsgDestroy

`MsgDestroy` removes a ledger and all associated data.

```protobuf
message MsgDestroyRequest {
    string nft_address = 1;
    string authority = 2;
}

message MsgDestroyResponse {}
```

### Fields
- `nft_address`: The address of the NFT whose ledger to destroy
- `authority`: The address of the authority who is destroying the ledger

### CLI Command
```bash
provenanced tx ledger destroy [nft-address] --from [authority]
```

## Ledger Entry Types

The module supports several types of ledger entries:

1. `LEDGER_ENTRY_TYPE_UNSPECIFIED`: Default type, not used in normal operations
2. `LEDGER_ENTRY_TYPE_DISBURSEMENT`: Represents funds being disbursed
   - Example: Initial loan amount disbursed to borrower
3. `LEDGER_ENTRY_TYPE_SCHEDULED_PAYMENT`: Represents a scheduled payment
   - Example: Regular monthly payment
4. `LEDGER_ENTRY_TYPE_UNSCHEDULED_PAYMENT`: Represents an unscheduled payment
   - Example: Extra payment or early payoff
5. `LEDGER_ENTRY_TYPE_FORECLOSURE_PAYMENT`: Represents a foreclosure-related payment
   - Example: Payment from foreclosure proceeds
6. `LEDGER_ENTRY_TYPE_FEE`: Represents a fee charged
   - Example: Origination fee, late payment fee
7. `LEDGER_ENTRY_TYPE_OTHER`: Represents other types of financial activities
   - Example: Adjustments, corrections, or special transactions

## Funding Transfer Status

The module supports several statuses for fund transfers:

1. `FUNDING_TRANSFER_STATUS_UNSPECIFIED`: Default status, not used in normal operations
2. `FUNDING_TRANSFER_STATUS_PENDING`: Transfer is pending processing
3. `FUNDING_TRANSFER_STATUS_PROCESSING`: Transfer is being processed
4. `FUNDING_TRANSFER_STATUS_COMPLETED`: Transfer has been completed
5. `FUNDING_TRANSFER_STATUS_FAILED`: Transfer has failed

## Message Validation

All messages are validated before processing:

1. **MsgCreate**
   - Ledger configuration must be valid
   - NFT address must be valid
   - Denomination must be valid
   - Authority must have permission
   - Dates must be in correct format
   - Amounts must be valid

2. **MsgAppend**
   - NFT address must be valid
   - Ledger must exist
   - Entries must be valid
   - Authority must have permission
   - Correlation IDs must be unique
   - Sequences must be valid

3. **MsgProcessFundTransfers**
   - Authority must have permission
   - Transfers must be valid
   - Status transitions must be valid
   - Settlement timing must be valid

4. **MsgProcessFundTransfersWithSettlement**
   - Authority must have permission
   - Transfers must be valid
   - Settlement instructions must be valid
   - Recipient addresses must be valid

5. **MsgDestroy**
   - NFT address must be valid
   - Ledger must exist
   - Authority must have permission
   - All associated data must be properly cleaned up 