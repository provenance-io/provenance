# Queries

The Ledger module provides several query endpoints to access ledger information and financial data.

## Query Endpoints

### Get Ledger Configuration
Retrieves the configuration for a specific NFT's ledger.

Request:
```protobuf
message QueryLedgerConfigRequest {
    string nft_address = 1;
}
```

Response:
```protobuf
message QueryLedgerConfigResponse {
    Ledger ledger = 1;
}
```

### Get Ledger Entries
Retrieves all ledger entries for a specific NFT.

Request:
```protobuf
message QueryLedgerRequest {
    string nft_address = 1;
}
```

Response:
```protobuf
message QueryLedgerResponse {
    repeated LedgerEntry entries = 1;
}
```

## Query Usage Examples

### Get Ledger Configuration
```bash
provenanced query ledger config [nft-address]
```

### Get Ledger Entries
```bash
provenanced query ledger entries [nft-address]
```

## Response Data

### Ledger Configuration
- NFT address
- Denomination
- Any additional configuration parameters

### Ledger Entries
- Entry UUID
- Entry type
- Posted date
- Effective date
- Amounts (total, principal, interest, other)
- Balance information

## Error Handling

The module returns appropriate error messages for:
- Invalid NFT addresses
- Non-existent ledgers
- Permission issues
- Invalid query parameters 