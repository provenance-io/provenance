# Messages

The Ledger module provides several message types for creating and managing NFT ledgers, entries, and fund transfers.

## MsgCreate

`MsgCreate` creates a new ledger for an NFT.

```protobuf
message MsgCreateRequest {
    string nft_address = 1;
    string denom = 2;
    string owner = 3;
}

message MsgCreateResponse {}
```

### Fields
- `nft_address`: The address of the NFT to create a ledger for
- `denom`: The denomination to use for the ledger entries
- `owner`: The address of the owner who is creating the ledger

### CLI Command
```bash
provenanced tx ledger create [nft-address] [denom] --from [owner]
```

## MsgAppend

`MsgAppend` adds a new entry to an existing ledger.

```protobuf
message MsgAppendRequest {
    string nft_address = 1;
    LedgerEntry entry = 2;
    string owner = 3;
}

message MsgAppendResponse {}
```

### Fields
- `nft_address`: The address of the NFT whose ledger to append to
- `entry`: The ledger entry to append
- `owner`: The address of the owner who is appending the entry

### CLI Command
```bash
provenanced tx ledger append [nft-address] [entry-json] --from [owner]
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

## Fund Transfer Types

### FundTransfer
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

### FundTransferWithSettlement
```protobuf
message FundTransferWithSettlement {
    string nft_address = 1;
    string ledger_entry_uuid = 2;
    repeated SettlementInstruction settlementInstructions = 3;
}
```

### SettlementInstruction
```protobuf
message SettlementInstruction {
    string amount = 1;
    string recipient_address = 2;
    string memo = 3;
    int64 settlement_block = 4;
}
```

## Funding Transfer Status

The module supports several statuses for fund transfers:

1. `FUNDING_TRANSFER_STATUS_UNSPECIFIED`: Default status, not used in normal operations
2. `FUNDING_TRANSFER_STATUS_PENDING`: Transfer is pending processing
3. `FUNDING_TRANSFER_STATUS_PROCESSING`: Transfer is being processed
4. `FUNDING_TRANSFER_STATUS_COMPLETED`: Transfer has been completed
5. `FUNDING_TRANSFER_STATUS_FAILED`: Transfer has failed 