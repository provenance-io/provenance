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
Retrieves all ledger entries for a specific NFT.

```protobuf
message QueryLedgerRequest {
    string nft_address = 1;
}

message QueryLedgerResponse {
    repeated LedgerEntry entries = 1;
}
```

### Get Balances As Of Date
Retrieves the balances for a specific NFT as of a given date.

```protobuf
message QueryBalancesAsOfRequest {
    string nft_address = 1;
    string as_of_date = 2;  // RFC3339 format (e.g., "2024-01-01T00:00:00Z")
}

message QueryBalancesAsOfResponse {
    Balances balances = 1;
}
```

### Get Entry by UUID
Retrieves a specific ledger entry by its UUID.

```protobuf
message QueryLedgerEntryRequest {
    string nft_address = 1;
    string uuid = 2;
}

message QueryLedgerEntryResponse {
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
```

### Get Balances As Of Date
```bash
# Get balances as of a specific date
provenanced query ledger balances [nft-address] [as-of-date]
# Example:
provenanced query ledger balances [nft-address] "2024-01-01T00:00:00Z"
```

### Get Entry by UUID
```bash
provenanced query ledger entry [nft-address] [uuid]
```

## Response Data

### Ledger Configuration
- NFT address
- Denomination

### Ledger Entries
- Entry UUID
- Entry type
- Posted date
- Effective date
- Amounts (total, principal, interest, other)
- Balance information

### Balances As Of Date
- Principal balance
- Interest balance
- Other balance

### Entry by UUID
- Complete ledger entry details including all amounts and dates

## Error Handling

The module returns appropriate error messages for:
- Invalid NFT addresses
- Non-existent ledgers
- Invalid entry UUIDs
- Invalid date formats
- Permission issues
- Invalid query parameters

## Notes

- All dates should be provided in RFC3339 format (e.g., "2024-01-01T00:00:00Z")
- The balances query will return the cumulative balances up to and including the specified date
- Entries are sorted by effective date when calculating balances
- The module maintains separate balances for principal, interest, and other amounts

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