# Ledger Queries

The Ledger module provides several query endpoints to access ledger state and information. These queries allow for retrieving ledger configurations, entries, balances, and other financial data.

<!-- TOC -->
  - [Ledger Queries](#ledger-queries)
  - [Entry Queries](#entry-queries)
  - [Balance Queries](#balance-queries)
  - [Query Implementation](#query-implementation)
  - [Error Handling](#error-handling)
  - [Query Usage Examples](#query-usage-examples)

## Ledger Queries

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

### Get Ledger Status
Retrieves the current status of a ledger:

```protobuf
message QueryLedgerStatusRequest {
    string nft_id = 1;
    string asset_class_id = 2;
}

message QueryLedgerStatusResponse {
    int32 status_type_id = 1;
    string status_code = 2;
    string status_description = 3;
}
```

## Entry Queries

### Get Ledger Entries
Retrieves all ledger entries for a specific asset:

```protobuf
message QueryLedgerRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    cosmos.base.query.v1beta1.PageRequest pagination = 3;
}

message QueryLedgerResponse {
    repeated LedgerEntry entries = 1;
    cosmos.base.query.v1beta1.PageResponse pagination = 2;
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

### Get Ledger Entries by Type
Retrieves ledger entries filtered by entry type:

```protobuf
message QueryLedgerEntriesByTypeRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    int32 entry_type_id = 3;
    cosmos.base.query.v1beta1.PageRequest pagination = 4;
}

message QueryLedgerEntriesByTypeResponse {
    repeated LedgerEntry entries = 1;
    cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

## Balance Queries

### Get Current Balances
Retrieves the current balances for a specific asset:

```protobuf
message QueryBalancesRequest {
    string nft_id = 1;
    string asset_class_id = 2;
}

message QueryBalancesResponse {
    Balances balances = 1;
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

### Get Balance by Bucket Type
Retrieves the balance for a specific bucket type:

```protobuf
message QueryBalanceByBucketRequest {
    string nft_id = 1;
    string asset_class_id = 2;
    int32 bucket_type_id = 3;
}

message QueryBalanceByBucketResponse {
    BucketBalance balance = 1;
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

2. **Get Ledger Status**
   - Validates asset identifiers
   - Retrieves ledger from store
   - Returns current status information

### Entry Queries
1. **Get Ledger Entries**
   - Validates asset identifiers
   - Retrieves all entries from store
   - Supports pagination
   - Returns list of entries

2. **Get Ledger Entry**
   - Validates asset identifiers and correlation ID
   - Retrieves specific entry from store
   - Returns entry details

3. **Get Ledger Entries by Type**
   - Validates asset identifiers and entry type ID
   - Filters entries by type
   - Supports pagination
   - Returns filtered list of entries

### Balance Queries
1. **Get Current Balances**
   - Validates asset identifiers
   - Retrieves current balances from store
   - Returns bucket balances

2. **Get Balances As Of Date**
   - Validates asset identifiers and date format
   - Calculates balances as of specified date
   - Returns bucket balances

3. **Get Balance by Bucket Type**
   - Validates asset identifiers and bucket type ID
   - Retrieves balance for specific bucket
   - Returns bucket balance

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

8. **Invalid Bucket Type**
   - Code: 8
   - Message: "invalid bucket type"

## Query Usage Examples

### CLI
```bash
# Get ledger class
provenanced q ledger class [ledger-class-id]

# Get ledger configuration
provenanced q ledger config [nft-id] [asset-class-id]

# Get ledger status
provenanced q ledger status [nft-id] [asset-class-id]

# Get ledger entries
provenanced q ledger entries [nft-id] [asset-class-id]

# Get ledger entry
provenanced q ledger entry [nft-id] [asset-class-id] [correlation-id]

# Get ledger entries by type
provenanced q ledger entries-by-type [nft-id] [asset-class-id] [entry-type-id]

# Get current balances
provenanced q ledger balances [nft-id] [asset-class-id]

# Get balances as of date
provenanced q ledger balances-as-of [nft-id] [asset-class-id] [as-of-date]

# Get balance by bucket type
provenanced q ledger balance-by-bucket [nft-id] [asset-class-id] [bucket-type-id]
```

### REST
```http
# Get ledger class
GET /provenance/ledger/v1/class/{ledger_class_id}

# Get ledger configuration
GET /provenance/ledger/v1/config?nft_id={nft_id}&asset_class_id={asset_class_id}

# Get ledger status
GET /provenance/ledger/v1/status?nft_id={nft_id}&asset_class_id={asset_class_id}

# Get ledger entries
GET /provenance/ledger/v1/entries?nft_id={nft_id}&asset_class_id={asset_class_id}

# Get ledger entry
GET /provenance/ledger/v1/ledger/{nft_id}/{asset_class_id}/entry/{correlation_id}

# Get ledger entries by type
GET /provenance/ledger/v1/entries-by-type?nft_id={nft_id}&asset_class_id={asset_class_id}&entry_type_id={entry_type_id}

# Get current balances
GET /provenance/ledger/v1/ledger/{nft_id}/{asset_class_id}/balances

# Get balances as of date
GET /provenance/ledger/v1/ledger/{nft_id}/{asset_class_id}/balances/{as_of_date}

# Get balance by bucket type
GET /provenance/ledger/v1/ledger/{nft_id}/{asset_class_id}/balance/{bucket_type_id}
```

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

## Notes

- All dates should be provided in ISO8601 format (e.g., "2024-01-01")
- The balances query will return the cumulative balances up to and including the specified date
- Entries are sorted by effective date when calculating balances
- The module maintains balances in configurable buckets as defined by the ledger class
- Correlation IDs are free-form strings up to 50 characters, used to track and correlate ledger entries with external systems
- Interest rates are stored with 6 decimal places (10000000 = 10.000000%) 