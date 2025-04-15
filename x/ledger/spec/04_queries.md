# Queries

The Ledger module provides several query endpoints to access ledger state and information.

## Query Types

### Get Ledger Configuration
Retrieves the ledger configuration for a specific NFT:

```protobuf
message QueryLedgerConfigRequest {
    string nft_address = 1;
}

message QueryLedgerConfigResponse {
    Ledger ledger = 1;
}
```

### Get Ledger Entries
Retrieves all ledger entries for a specific NFT:

```protobuf
message QueryLedgerRequest {
    string nft_address = 1;
}

message QueryLedgerResponse {
    repeated LedgerEntry entries = 1;
}
```

### Get Ledger Entry
Retrieves a specific ledger entry by NFT address and correlation ID:

```protobuf
message QueryLedgerEntryRequest {
    string nft_address = 1;
    string correlation_id = 2;  // Free-form string up to 50 characters
}

message QueryLedgerEntryResponse {
    LedgerEntry entry = 1;
}
```

### Get Balances As Of Date
Retrieves the balances for a specific NFT as of a given date:

```protobuf
message QueryBalancesAsOfRequest {
    string nft_address = 1;
    string as_of_date = 2;  // Date in ISO 8601 format: YYYY-MM-DD
}

message QueryBalancesAsOfResponse {
    Balances balances = 1;
}
```

## Query Implementation

### Ledger Queries
1. **Get Ledger Configuration**
   - Validates NFT address
   - Retrieves ledger from store
   - Returns ledger configuration

2. **Get Ledger Entries**
   - Validates NFT address
   - Retrieves all entries from store
   - Returns list of entries

3. **Get Ledger Entry**
   - Validates NFT address and correlation ID
   - Retrieves specific entry from store
   - Returns entry details

### Balance Queries
1. **Get Balances As Of Date**
   - Validates NFT address and date format
   - Calculates balances as of specified date
   - Returns principal, interest, and other balances

## Query Response Format

All query responses follow a standard format:

```protobuf
message QueryResponse {
    // Response data
    oneof response {
        QueryLedgerConfigResponse config = 1;
        QueryLedgerResponse entries = 2;
        QueryLedgerEntryResponse entry = 3;
        QueryBalancesAsOfResponse balances = 4;
    }
}
```

## Error Handling

Queries may return the following errors:

1. **Invalid NFT Address**
   - Code: 1
   - Message: "invalid nft address"

2. **Ledger Not Found**
   - Code: 2
   - Message: "ledger not found"

3. **Entry Not Found**
   - Code: 3
   - Message: "entry not found"

4. **Invalid Correlation ID**
   - Code: 4
   - Message: "invalid correlation id"

5. **Invalid Date Format**
   - Code: 5
   - Message: "invalid date format"

## Query Usage Examples

### CLI
```bash
# Get ledger configuration
provenanced q ledger config [nft-address]

# Get ledger entries
provenanced q ledger entries [nft-address]

# Get ledger entry
provenanced q ledger entry [nft-address] [correlation-id]

# Get balances as of date
provenanced q ledger balances [nft-address] [as-of-date]
```

### REST
```http
# Get ledger configuration
GET /provenance/ledger/v1/config?nft_address={nft_address}

# Get ledger entries
GET /provenance/ledger/v1/entries?nft_address={nft_address}

# Get ledger entry
GET /provenance/ledger/v1/ledger/{nft_address}/entry/{correlation_id}

# Get balances as of date
GET /provenance/ledger/v1/ledger/{nft_address}/balances/{as_of_date}
```

## Notes

- All dates should be provided in ISO8601 format (e.g., "2024-01-01")
- The balances query will return the cumulative balances up to and including the specified date
- Entries are sorted by effective date when calculating balances
- The module maintains separate balances for principal, interest, and other amounts
- Correlation IDs are free-form strings up to 50 characters, used to track and correlate ledger entries with external systems

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