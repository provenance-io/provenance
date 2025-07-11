# Queries

The Ledger module provides several query endpoints to access ledger state and information.

## Query Types

### Get Ledger Class
Retrieves the ledger class configuration:

```protobuf
message QueryLedgerClassRequest {
    string ledger_class_id = 1;
}

message QueryLedgerClassResponse {
    LedgerClass ledger_class = 1;
}
```

### Get Ledger Configuration
Retrieves the ledger configuration for a specific asset:

```protobuf
message QueryLedgerConfigRequest {
    string nft_id = 1;
    string asset_class_id = 2;
}

message QueryLedgerConfigResponse {
    Ledger ledger = 1;
}
```

### Get Ledger Entries
Retrieves all ledger entries for a specific asset:

```protobuf
message QueryLedgerRequest {
    string nft_id = 1;
    string asset_class_id = 2;
}

message QueryLedgerResponse {
    repeated LedgerEntry entries = 1;
}
```

### Get Ledger Entry
Retrieves a specific ledger entry by asset identifier and correlation ID:

```protobuf
message QueryLedgerEntryRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    string correlation_id = 3;  // Free-form string up to 50 characters
}

message QueryLedgerEntryResponse {
    LedgerEntry entry = 1;
}
```

### Get Balances As Of Date
Retrieves the balances for a specific asset as of a given date:

```protobuf
message QueryBalancesAsOfRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    string as_of_date = 3;  // Date in ISO 8601 format: YYYY-MM-DD
}

message QueryBalancesAsOfResponse {
    Balances balances = 1;
}
```

## Query Implementation

### Ledger Class Queries
1. **Get Ledger Class**
   - Validates ledger class ID
   - Retrieves ledger class from store
   - Returns ledger class configuration

### Ledger Queries
1. **Get Ledger Configuration**
   - Validates asset identifiers
   - Retrieves ledger from store
   - Returns ledger configuration

2. **Get Ledger Entries**
   - Validates asset identifiers
   - Retrieves all entries from store
   - Returns list of entries

3. **Get Ledger Entry**
   - Validates asset identifiers and correlation ID
   - Retrieves specific entry from store
   - Returns entry details

### Balance Queries
1. **Get Balances As Of Date**
   - Validates asset identifiers and date format
   - Calculates balances as of specified date
   - Returns bucket balances

## Query Response Format

All query responses follow a standard format:

```protobuf
message QueryResponse {
    // Response data
    oneof response {
        QueryLedgerClassResponse ledger_class = 1;
        QueryLedgerConfigResponse config = 2;
        QueryLedgerResponse entries = 3;
        QueryLedgerEntryResponse entry = 4;
        QueryBalancesAsOfResponse balances = 5;
    }
}
```

## Error Handling

Queries may return the following errors:

1. **Invalid Asset Identifier**
   - Code: 1
   - Message: "invalid asset identifier"

2. **Invalid Asset Class**
   - Code: 2
   - Message: "invalid asset class"

3. **Ledger Class Not Found**
   - Code: 3
   - Message: "ledger class not found"

4. **Ledger Not Found**
   - Code: 4
   - Message: "ledger not found"

5. **Entry Not Found**
   - Code: 5
   - Message: "entry not found"

6. **Invalid Correlation ID**
   - Code: 6
   - Message: "invalid correlation id"

7. **Invalid Date Format**
   - Code: 7
   - Message: "invalid date format"

## Query Usage Examples

### CLI
```bash
# Get ledger class
provenanced q ledger class [ledger-class-id]

# Get ledger configuration
provenanced q ledger config [nft-id] [asset-class-id]

# Get ledger entries
provenanced q ledger entries [nft-id] [asset-class-id]

# Get ledger entry
provenanced q ledger entry [nft-id] [asset-class-id] [correlation-id]

# Get balances as of date
provenanced q ledger balances [nft-id] [asset-class-id] [as-of-date]
```

### REST
```http
# Get ledger class
GET /provenance/ledger/v1/class/{ledger_class_id}

# Get ledger configuration
GET /provenance/ledger/v1/config?nft_id={nft_id}&asset_class_id={asset_class_id}

# Get ledger entries
GET /provenance/ledger/v1/entries?nft_id={nft_id}&asset_class_id={asset_class_id}

# Get ledger entry
GET /provenance/ledger/v1/ledger/{nft_id}/{asset_class_id}/entry/{correlation_id}

# Get balances as of date
GET /provenance/ledger/v1/ledger/{nft_id}/{asset_class_id}/balances/{as_of_date}
```

## Notes

- All dates should be provided in ISO8601 format (e.g., "2024-01-01")
- The balances query will return the cumulative balances up to and including the specified date
- Entries are sorted by effective date when calculating balances
- The module maintains balances in configurable buckets as defined by the ledger class
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
provenanced query ledger entries [nft_id] [asset_class_id] --limit=10

# Get next page using the next_key from previous response
provenanced query ledger entries [nft_id] [asset_class_id] --limit=10 --page-key=[next_key]
``` 