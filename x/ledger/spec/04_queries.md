# Queries

The Ledger module provides several query endpoints to access ledger information and financial data.

## Query Endpoints

### Get Ledger Configuration
Retrieves the configuration for a specific NFT's ledger.

```protobuf
message QueryLedgerConfigRequest {
    string nft_address = 1;
}

message QueryLedgerConfigResponse {
    Ledger ledger = 1;
}
```

### Get Ledger Entries
Retrieves ledger entries for a specific NFT with optional filtering.

```protobuf
message QueryLedgerRequest {
    string nft_address = 1;
    LedgerEntryType entry_type = 2;
    google.protobuf.Timestamp start_date = 3;
    google.protobuf.Timestamp end_date = 4;
    cosmos.base.query.v1beta1.PageRequest pagination = 5;
}

message QueryLedgerResponse {
    repeated LedgerEntry entries = 1;
    cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

### Get Current Balances
Retrieves the current balances for a specific NFT.

```protobuf
message QueryBalancesRequest {
    string nft_address = 1;
}

message QueryBalancesResponse {
    string principal_balance = 1;
    string interest_balance = 2;
    string other_balance = 3;
    string total_balance = 4;
    google.protobuf.Timestamp as_of = 5;
}
```

### Get Entry by UUID
Retrieves a specific ledger entry by its UUID.

```protobuf
message QueryEntryRequest {
    string nft_address = 1;
    string entry_uuid = 2;
}

message QueryEntryResponse {
    LedgerEntry entry = 1;
}
```

## Query Usage Examples

### Get Ledger Configuration
```bash
provenanced query ledger config [nft-address]
```

### Get Ledger Entries
```bash
# Get all entries
provenanced query ledger entries [nft-address]

# Get entries with filters
provenanced query ledger entries [nft-address] --entry-type=PAYMENT --start-date=2024-01-01 --end-date=2024-03-31
```

### Get Current Balances
```bash
provenanced query ledger balances [nft-address]
```

### Get Entry by UUID
```bash
provenanced query ledger entry [nft-address] [entry-uuid]
```

## Response Data

### Ledger Configuration
- NFT address
- Denomination
- Creation timestamp
- Update timestamp
- Status
- Metadata

### Ledger Entries
- Entry UUID
- Entry type
- Posted date
- Effective date
- Amounts (total, principal, interest, other)
- Balance information
- Metadata
- Created by

### Current Balances
- Principal balance
- Interest balance
- Other balance
- Total balance
- As of timestamp

## Error Handling

The module returns appropriate error messages for:
- Invalid NFT addresses
- Non-existent ledgers
- Invalid entry UUIDs
- Permission issues
- Invalid query parameters
- Pagination errors

## Pagination

All list queries support pagination using the Cosmos SDK pagination system:

```protobuf
message PageRequest {
    string key = 1;
    uint64 offset = 2;
    uint64 limit = 3;
    bool count_total = 4;
    bool reverse = 5;
}

message PageResponse {
    string next_key = 1;
    uint64 total = 2;
}
```

Example pagination usage:
```bash
# Get first page
provenanced query ledger entries [nft-address] --limit=10

# Get next page using the next_key from previous response
provenanced query ledger entries [nft-address] --limit=10 --page-key=[next_key]
``` 