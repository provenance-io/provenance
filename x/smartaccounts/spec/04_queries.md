<!--
order: 4
-->

# Queries

The smart accounts module provides query services to retrieve information about smart accounts and module parameters. All queries are accessible via gRPC, REST API, and CLI.

## Query Service

```protobuf
service Query {
  rpc SmartAccount(SmartAccountQueryRequest) returns (SmartAccountResponse);
  rpc Params(QueryParamsRequest) returns (QueryParamsResponse);
}
```

## Query Types

### SmartAccount Query

Retrieves complete information about a smart account including all registered credentials.

#### Request

```protobuf
message SmartAccountQueryRequest {
  string address = 1;
}
```

**Fields:**
- `address`: The bech32 address of the account to query

**Validation:**
- `address` must be a valid bech32 address
- Account must exist in the system

#### Response

```protobuf
message SmartAccountResponse {
  ProvenanceAccount provenanceaccount = 1;
}
```

**Fields:**
- `provenanceaccount`: Complete smart account data including base account and all credentials

**Response Details:**
The response includes:
- Base account information (address, account number, sequence)
- Smart account number
- List of all registered credentials with metadata
- Authentication settings

#### Example Response

```json
{
  "provenanceaccount": {
    "base_account": {
      "@type": "/cosmos.auth.v1beta1.BaseAccount",
      "address": "tp1...",
      "account_number": "123",
      "sequence": "5"
    },
    "smart_account_number": "456",
    "credentials": [
      {
        "base_credential": {
          "credential_number": "1",
          "public_key": {
            "@type": "/cosmos.crypto.secp256k1.PubKey",
            "key": "..."
          },
          "variant": "CREDENTIAL_TYPE_K256",
          "create_time": "1640995200"
        },
        "k256_authenticator": {}
      },
      {
        "base_credential": {
          "credential_number": "2",
          "public_key": {
            "@type": "/provenance.smartaccounts.v1.PubKeyWebauthn",
            "key": "..."
          },
          "variant": "CREDENTIAL_TYPE_WEBAUTHN_UV",
          "create_time": "1640995300"
        },
        "fido2_authenticator": {
          "id": "credential-id-base64",
          "username": "user@example.com",
          "aaguid": "...",
          "credential_creation_response": "...",
          "rp_id": "example.com",
          "rp_origin": "https://example.com"
        }
      }
    ],
    "is_smart_account_only_authentication": false
  }
}
```

### Parameters Query

Retrieves the current module parameters.

#### Request

```protobuf
message QueryParamsRequest {}
```

Empty request message.

#### Response

```protobuf
message QueryParamsResponse {
  Params params = 1;
}
```

**Fields:**
- `params`: Current module parameters

#### Example Response

```json
{
  "params": {
    "enabled": true,
    "max_credential_allowed": 10
  }
}
```

## API Endpoints

### REST API

**Smart Account Query:**
```
GET /provenance/smartaccount/v1/account?address={address}
```

**Parameters Query:**
```
GET /provenance/smartaccount/v1/params
```

### gRPC

**Smart Account Query:**
```
provenance.smartaccounts.v1.Query/SmartAccount
```

**Parameters Query:**
```
provenance.smartaccounts.v1.Query/Params
```

## CLI Commands

### Query Smart Account

```bash
provenanced query smartaccounts account [address]
```

**Example:**
```bash
provenanced query smartaccounts account tp1..address..
```

**Flags:**
- `--height`: Query at specific block height
- `--output`: Output format (json, text)
- `--node`: RPC node endpoint

### Query Parameters

```bash
provenanced query smartaccounts params
```

**Example:**
```bash
provenanced query smartaccounts params --output json
```

## Error Handling

### Common Query Errors

**Account Not Found:**
- Status: `NotFound`
- Description: The specified address does not exist or is not a smart account
- Resolution: Verify the address is correct and the account exists

**Invalid Address:**
- Status: `InvalidArgument`
- Description: The provided address is not a valid bech32 address
- Resolution: Provide a properly formatted address

**Module Disabled:**
- Status: `FailedPrecondition`
- Description: Smart accounts module is disabled
- Resolution: Enable module through governance or wait for activation

## Query Filtering and Pagination

### Current Limitations

The current implementation does not support:
- Pagination for credential lists (accounts have a maximum credential limit)
- Filtering credentials by type or status
- Bulk account queries

### Future Enhancements

Planned query improvements include:
- Credential filtering by type (`CREDENTIAL_TYPE_WEBAUTHN`, `CREDENTIAL_TYPE_K256`, etc.)
- Credential status queries (active, expired, revoked)
- Account listing with pagination
- Credential usage statistics

## Performance Considerations

### Query Optimization

**Smart Account Queries:**
- Cached at the RPC level for frequently accessed accounts
- Single key-value lookup in account store
- Credential data is stored inline (no additional lookups required)

**Parameter Queries:**
- Highly cacheable (parameters change infrequently)
- Single key-value lookup
- Minimal serialization overhead

### Rate Limiting

Standard Cosmos SDK query rate limiting applies:
- No special rate limits for smart account queries
- Follow standard RPC node configuration
- Consider caching for high-frequency applications

## Integration Examples

### JavaScript/TypeScript

```typescript
import { QueryClient } from '@cosmjs/stargate';

const client = QueryClient.withExtensions(
  await Tendermint34Client.connect(rpcEndpoint),
  setupSmartAccountsExtension
);

// Query smart account
const account = await client.smartaccounts.smartAccount({
  address: 'tp1...'
});

// Query parameters
const params = await client.smartaccounts.params({});
```

### Go

```go
import (
  "github.com/provenance-io/provenance/x/smartaccounts/types"
)

// Query smart account
req := &types.SmartAccountQueryRequest{
  Address: "tp1...",
}
res, err := queryClient.SmartAccount(ctx, req)

// Query parameters
paramsReq := &types.QueryParamsRequest{}
paramsRes, err := queryClient.Params(ctx, paramsReq)
```

### Python

```python
from provenance_client import Client

client = Client()

# Query smart account
account = client.smartaccounts.smart_account(address="tp1...")

# Query parameters
params = client.smartaccounts.params()
```

## Monitoring and Observability

### Query Metrics

Applications should monitor:
- Query response times
- Query error rates
- Cache hit rates (if applicable)
- Account access patterns

### Logging

Recommended logging for query operations:
- Account address in query requests
- Query duration and response size
- Error conditions and error codes
- Cache utilization (if implemented)

## Security Considerations

### Information Disclosure

**Public Information:**
- Account addresses are public
- Credential metadata (types, creation times) is public
- Public keys are public by design

**Private Information:**
- Private keys are never exposed
- Biometric data is never stored on-chain
- User identifiers may contain sensitive information

### Query Authentication

**No Authentication Required:**
- Queries are read-only operations
- No sensitive information is exposed
- Standard RPC rate limiting provides protection

**Access Control:**
- All smart account information is publicly queryable
- No special permissions required for queries
- Consider privacy implications when storing user identifiers

## Testing Queries

### Unit Testing

```go
func TestSmartAccountQuery(t *testing.T) {
  // Setup test environment
  app := setupTestApp()
  ctx := app.BaseApp.NewContext(false, tmproto.Header{})
  
  // Create test account
  account := createTestSmartAccount(ctx, app)
  
  // Query the account
  req := &types.SmartAccountQueryRequest{
    Address: account.GetAddress().String(),
  }
  
  res, err := app.SmartAccountsKeeper.SmartAccount(ctx, req)
  require.NoError(t, err)
  require.NotNil(t, res.ProvenanceAccount)
}
```

### Integration Testing

```bash
# Test CLI queries
provenanced query smartaccounts account tp1test... --node http://localhost:26657
provenanced query smartaccounts params --node http://localhost:26657

# Test REST API
curl "http://localhost:1317/provenance/smartaccount/v1/account?address=tp1test..."
curl "http://localhost:1317/provenance/smartaccount/v1/params"
```
