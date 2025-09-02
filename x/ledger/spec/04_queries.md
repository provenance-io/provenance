# Ledger Queries

The Ledger module provides several query endpoints to access ledger state and information. These queries allow for retrieving ledger configurations, entries, balances, and other financial data.

<!-- TOC -->
  - [Ledger Queries](#ledger-queries)
  - [Entry Queries](#entry-queries)
  - [Balance Queries](#balance-queries)
  - [Class Configuration Queries](#class-configuration-queries)
  - [Settlement Queries](#settlement-queries)
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
message QueryLedgerRequest {
    LedgerKey key = 1;  // Contains nft_id and asset_class_id
}

message QueryLedgerResponse {
    Ledger ledger = 1;
}
```

## Entry Queries

### Get Ledger Entries
Retrieves all ledger entries for a specific asset:

```protobuf
message QueryLedgerEntriesRequest {
    LedgerKey key = 1;  // Contains nft_id and asset_class_id
}

message QueryLedgerEntriesResponse {
    repeated LedgerEntry entries = 1;
}
```

### Get Ledger Entry
Retrieves a specific ledger entry by asset identifier and correlation ID:

```protobuf
message QueryLedgerEntryRequest {
    LedgerKey key = 1;              // Contains nft_id and asset_class_id
    string correlation_id = 2;       // Free-form string up to 50 characters
}

message QueryLedgerEntryResponse {
    LedgerEntry entry = 1;
}
```

## Balance Queries

### Get Balances As Of Date
Retrieves the balances for a specific asset as of a given date:

```protobuf
message QueryLedgerBalancesAsOfRequest {
    LedgerKey key = 1;              // Contains nft_id and asset_class_id
    string as_of_date = 2;          // Date in ISO 8601 format: YYYY-MM-DD
}

message QueryLedgerBalancesAsOfResponse {
    BucketBalances bucket_balances = 1;
}
```

## Class Configuration Queries

### Get Ledger Class Entry Types
Retrieves all entry types configured for a specific ledger class:

```protobuf
message QueryLedgerClassEntryTypesRequest {
    string ledger_class_id = 1;
}

message QueryLedgerClassEntryTypesResponse {
    repeated LedgerClassEntryType entry_types = 1;
}
```

### Get Ledger Class Status Types
Retrieves all status types configured for a specific ledger class:

```protobuf
message QueryLedgerClassStatusTypesRequest {
    string ledger_class_id = 1;
}

message QueryLedgerClassStatusTypesResponse {
    repeated LedgerClassStatusType status_types = 1;
}
```

### Get Ledger Class Bucket Types
Retrieves all bucket types configured for a specific ledger class:

```protobuf
message QueryLedgerClassBucketTypesRequest {
    string ledger_class_id = 1;
}

message QueryLedgerClassBucketTypesResponse {
    repeated LedgerClassBucketType bucket_types = 1;
}
```

## Settlement Queries

### Get Ledger Settlements
Retrieves all settlements for a specific ledger:

```protobuf
message QueryLedgerSettlementsRequest {
    LedgerKey key = 1;  // Contains nft_id and asset_class_id
}

message QueryLedgerSettlementsResponse {
    repeated StoredSettlementInstructions settlements = 1;
}
```

### Get Ledger Settlements By Correlation ID
Retrieves settlements by correlation ID:

```protobuf
message QueryLedgerSettlementsByCorrelationIDRequest {
    string correlation_id = 1;
}

message QueryLedgerSettlementsByCorrelationIDResponse {
    repeated StoredSettlementInstructions settlements = 1;
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
   - Validates ledger key (asset identifiers)
   - Retrieves ledger from store
   - Returns ledger configuration

### Entry Queries
1. **Get Ledger Entries**
   - Validates ledger key (asset identifiers)
   - Retrieves all entries from store
   - Returns list of entries

2. **Get Ledger Entry**
   - Validates ledger key and correlation ID
   - Retrieves specific entry from store
   - Returns entry details

### Balance Queries
1. **Get Balances As Of Date**
   - Validates ledger key and date format
   - Calculates balances as of specified date
   - Returns bucket balances

### Class Configuration Queries
1. **Get Ledger Class Entry Types**
   - Validates ledger class ID
   - Retrieves entry types from store
   - Returns list of entry types

2. **Get Ledger Class Status Types**
   - Validates ledger class ID
   - Retrieves status types from store
   - Returns list of status types

3. **Get Ledger Class Bucket Types**
   - Validates ledger class ID
   - Retrieves bucket types from store
   - Returns list of bucket types

### Settlement Queries
1. **Get Ledger Settlements**
   - Validates ledger key
   - Retrieves settlements from store
   - Returns list of settlements

2. **Get Ledger Settlements By Correlation ID**
   - Validates correlation ID
   - Retrieves settlements by correlation ID
   - Returns list of settlements

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
provenanced q ledger class <ledger-class-id>

# Get ledger configuration
provenanced q ledger get <asset-class-id> <nft-id>

# Get ledger entries
provenanced q ledger entries <asset-class-id> <nft-id>

# Get ledger entry
provenanced q ledger entry <asset-class-id> <nft-id> <correlation-id>

# Get balances as of date
provenanced q ledger balances-as-of <asset-class-id> <nft-id> <as-of-date>

# Get ledger class entry types
provenanced q ledger entry-types <ledger-class-id>

# Get ledger class status types
provenanced q ledger status-types <ledger-class-id>

# Get ledger class bucket types
provenanced q ledger bucket-types <ledger-class-id>

# Get all settlements for a ledger
provenanced q ledger settlements <asset-class-id> <nft-id>

# Get settlements by correlation ID
provenanced q ledger settlements-by-correlation <correlation-id>
```

### REST
```http
# Get ledger class
GET /provenance/ledger/v1/class/{ledger_class_id}

# Get ledger configuration
GET /provenance/ledger/v1/ledger?key.asset_class_id={asset_class_id}&key.nft_id={nft_id}

# Get ledger entries
GET /provenance/ledger/v1/ledger/{asset_class_id}/{nft_id}/entries

# Get ledger entry
GET /provenance/ledger/v1/ledger/{asset_class_id}/{nft_id}/entry/{correlation_id}

# Get balances as of date
GET /provenance/ledger/v1/ledger/{asset_class_id}/{nft_id}/balances/{as_of_date}

# Get ledger class entry types
GET /provenance/ledger/v1/class/{ledger_class_id}/entry-types

# Get ledger class status types
GET /provenance/ledger/v1/class/{ledger_class_id}/status-types

# Get ledger class bucket types
GET /provenance/ledger/v1/class/{ledger_class_id}/bucket-types

# Get ledger settlements
GET /provenance/ledger/v1/ledger/{asset_class_id}/{nft_id}/settlements

# Get settlements by correlation ID
GET /provenance/ledger/v1/settlements/correlation/{correlation_id}
```

## Notes

- All dates should be provided in ISO8601 format (e.g., "2024-01-01")
- The balances query calculates cumulative balances up to and including the specified date
- Entries are sorted by effective date when calculating balances
- The module maintains balances in configurable buckets as defined by the ledger class
- Correlation IDs are free-form strings up to 50 characters, used to track and correlate ledger entries with external systems
- Interest rates are stored with 6 decimal places (10000000 = 10.000000%)
- Ledger keys are bech32-encoded strings that combine asset class ID and NFT ID
- Balances are calculated on-the-fly from ledger entries rather than stored separately 