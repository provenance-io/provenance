# Messages

The Ledger module provides several message types for creating and managing NFT ledgers, entries, and fund transfers.

## MsgCreate

`MsgCreate` creates a new ledger for an NFT.

```protobuf
message MsgCreateRequest {
    Ledger ledger = 1;
    string owner = 2;
}

message MsgCreateResponse {}
```

### Fields
- `ledger`: The ledger configuration to create
  - `nft_address`: The address of the NFT to create a ledger for
  - `denom`: The denomination to use for the ledger entries
  - `next_pmt_date`: The next scheduled payment date in ISO 8601 format: YYYY-MM-DD
  - `next_pmt_amt`: The amount of the next scheduled payment
  - `status`: The current status of the ledger
  - `interest_rate`: The interest rate for the ledger
  - `maturity_date`: The maturity date in ISO 8601 format: YYYY-MM-DD
- `owner`: The address of the owner who is creating the ledger

### CLI Command
```bash
provenanced tx ledger create [ledger-fields...] --from [owner]
```

## MsgAppend

`MsgAppend` adds one or more new entries to an existing ledger.

```protobuf
message MsgAppendRequest {
    string nft_address = 1;
    repeated LedgerEntry entries = 2;
    string owner = 3;
}

message MsgAppendResponse {}
```

### Fields
- `nft_address`: The address of the NFT whose ledger to append to
- `entries`: One or more ledger entries to append (each with correlation_id and sequence)
- `owner`: The address of the owner who is appending the entries

### CLI Command
Only allows a single entry, use the RPC endpoint to add multiple entries in a single call.
```bash
provenanced tx ledger append [nft-address] [entries fields...] --from [owner]
```

## MsgProcessFundTransfers

`MsgProcessFundTransfers` processes multiple fund transfers (payments and disbursements).

```protobuf
message MsgProcessFundTransfersRequest {
    string owner = 1;
    repeated FundTransfer transfers = 2;
}

message MsgProcessFundTransfersResponse {}
```

### Fields
- `owner`: The address of the owner processing the transfers
- `transfers`: List of fund transfers to process

### CLI Command
```bash
provenanced tx ledger process-transfers [transfers-json] --from [owner]
```

## MsgProcessFundTransfersWithSettlement

`MsgProcessFundTransfersWithSettlement` processes fund transfers with manual settlement instructions.

```protobuf
message MsgProcessFundTransfersWithSettlementRequest {
    string owner = 1;
    repeated FundTransferWithSettlement transfers = 2;
}

message MsgProcessFundTransfersResponse {}
```

### Fields
- `owner`: The address of the owner processing the transfers
- `transfers`: List of fund transfers with settlement instructions to process

### CLI Command
```bash
provenanced tx ledger process-transfers-with-settlement [transfers-json] --from [owner]
```

## MsgDestroy

`MsgDestroy` removes a ledger and all associated data.

```protobuf
message MsgDestroyRequest {
    string nft_address = 1;
    string owner = 2;
}

message MsgDestroyResponse {}
```

### Fields
- `nft_address`: The address of the NFT whose ledger to destroy
- `owner`: The address of the owner who is destroying the ledger

### CLI Command
```bash
provenanced tx ledger destroy [nft-address] --from [owner]
```

## Fund Transfer Types

### FundTransfer
```protobuf
message FundTransfer {
    string nft_address = 1;
    string ledger_entry_correlation_id = 2;  // Correlation ID of the associated ledger entry
    string amount = 3;
    FundingTransferStatus status = 4;
    string memo = 5;
    int64 settlement_block = 6;  // Minimum block height or timestamp for settlement
}
```

### FundTransferWithSettlement
```protobuf
message FundTransferWithSettlement {
    string nft_address = 1;
    string ledger_entry_correlation_id = 2;  // Correlation ID of the associated ledger entry
    repeated SettlementInstruction settlementInstructions = 3;
}
```

### SettlementInstruction
```protobuf
message SettlementInstruction {
    string amount = 1;
    string recipient_address = 2;  // The recipient's blockchain address
    string memo = 3;               // Optional memo or note for the transaction
    int64 settlement_block = 4;    // Minimum block height or timestamp for settlement
}
```

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
   - Owner must have permission
   - Dates must be in correct format
   - Amounts must be valid

2. **MsgAppend**
   - NFT address must be valid
   - Ledger must exist
   - Entries must be valid
   - Owner must have permission
   - Correlation IDs must be unique
   - Sequences must be valid

3. **MsgProcessFundTransfers**
   - Owner must have permission
   - Transfers must be valid
   - Status transitions must be valid
   - Settlement timing must be valid

4. **MsgProcessFundTransfersWithSettlement**
   - Owner must have permission
   - Transfers must be valid
   - Settlement instructions must be valid
   - Recipient addresses must be valid

5. **MsgDestroy**
   - NFT address must be valid
   - Ledger must exist
   - Owner must have permission
   - All associated data must be properly cleaned up 