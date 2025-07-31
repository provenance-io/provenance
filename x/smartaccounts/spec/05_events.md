<!--
order: 5
-->

# Events

The smart accounts module emits various events during credential management operations. These events provide a comprehensive audit trail for smart account operations and enable external systems to monitor account activity.

## Event Types

### EventSmartAccountInit

Emitted when a new smart account is initialized with credentials.

```protobuf
message EventSmartAccountInit {
  string address = 1;
  uint64 credential_count = 2;
}
```

**Attributes:**
- `address`: The bech32 address of the newly created smart account
- `credential_count`: Number of credentials registered during initialization

**When Emitted:**
- During the first credential registration for an account
- When converting a regular account to a smart account

**Example:**
```json
{
  "type": "provenance.smartaccounts.v1.EventSmartAccountInit",
  "attributes": [
    {
      "key": "address",
      "value": "tp1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpq"
    },
    {
      "key": "credential_count",
      "value": "1"
    }
  ]
}
```

### EventFido2CredentialAdd

Emitted when a FIDO2/WebAuthn credential is successfully added to a smart account.

```protobuf
message EventFido2CredentialAdd {
  string address = 1;
  uint64 credential_number = 2;
  string credential_id = 3;
}
```

**Attributes:**
- `address`: The bech32 address of the smart account
- `credential_number`: Globally unique identifier assigned to the credential
- `credential_id`: The base64url-encoded credential ID from the authenticator

**When Emitted:**
- After successful FIDO2 credential registration
- During MsgRegisterFido2Credential processing

**Example:**
```json
{
  "type": "provenance.smartaccounts.v1.EventFido2CredentialAdd",
  "attributes": [
    {
      "key": "address",
      "value": "tp1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpq"
    },
    {
      "key": "credential_number",
      "value": "42"
    },
    {
      "key": "credential_id",
      "value": "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA"
    }
  ]
}
```

### EventCosmosCredentialAdd

Emitted when a traditional cryptographic credential is added to a smart account.

```protobuf
message EventCosmosCredentialAdd {
  string address = 1;
  uint64 credential_number = 2;
}
```

**Attributes:**
- `address`: The bech32 address of the smart account
- `credential_number`: Globally unique identifier assigned to the credential

**When Emitted:**
- After successful secp256k1, Ed25519, or P256 credential registration
- During MsgRegisterCosmosCredential processing

**Example:**
```json
{
  "type": "provenance.smartaccounts.v1.EventCosmosCredentialAdd",
  "attributes": [
    {
      "key": "address",
      "value": "tp1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpq"
    },
    {
      "key": "credential_number",
      "value": "43"
    }
  ]
}
```

### EventCredentialDelete

Emitted when a credential is removed from a smart account.

```protobuf
message EventCredentialDelete {
  string address = 1;
  uint64 credential_number = 2;
}
```

**Attributes:**
- `address`: The bech32 address of the smart account
- `credential_number`: The unique identifier of the deleted credential

**When Emitted:**
- After successful credential deletion
- During MsgDeleteCredential processing

**Example:**
```json
{
  "type": "provenance.smartaccounts.v1.EventCredentialDelete",
  "attributes": [
    {
      "key": "address",
      "value": "tp1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpq"
    },
    {
      "key": "credential_number",
      "value": "42"
    }
  ]
}
```

## Event Context

### Block Height and Timestamp

All events include standard Cosmos SDK event metadata:
- Block height when the event occurred
- Block timestamp
- Transaction hash that triggered the event
- Event index within the transaction

### Transaction Context

Events are emitted within the context of successful message execution:
- Events are only emitted for successful operations
- Failed transactions do not emit module-specific events
- Events are included in transaction receipts and can be queried

## Event Attributes

### Address Format

All address attributes use the bech32 format:
- Mainnet: `pb1...` prefix
- Testnet: `tp1...` prefix  
- Local/development: configurable prefix

### Credential Numbers

Credential numbers are globally unique across all smart accounts:
- Monotonically increasing sequence
- Never reused, even after credential deletion
- Useful for tracking credential lifecycle

### Credential IDs

For FIDO2 credentials, the credential_id is:
- Base64url-encoded binary data
- Directly from the WebAuthn authenticator
- Unique per authenticator device
- Used for authentication assertion verification

## Event Filtering and Querying

### By Event Type

Query specific event types:

```bash
# Query FIDO2 credential additions
provenanced query txs --events 'provenance.smartaccounts.v1.EventFido2CredentialAdd.address=tp1...'

# Query credential deletions
provenanced query txs --events 'provenance.smartaccounts.v1.EventCredentialDelete.credential_number=42'
```

### By Account Address

Query all smart account events for a specific address:

```bash
provenanced query txs --events 'provenance.smartaccounts.v1.EventSmartAccountInit.address=tp1...'
```

### By Block Range

Query events within a specific block range:

```bash
provenanced query txs --events 'provenance.smartaccounts.v1.EventFido2CredentialAdd.address=tp1...' --height 1000 --limit 100
```

## Event Monitoring

### Real-time Monitoring

Applications can monitor events in real-time using:

**WebSocket Subscriptions:**
```javascript
const client = await StargateClient.connect(rpcEndpoint);
const stream = client.subscribeTx({
  type: 'provenance.smartaccounts.v1.EventFido2CredentialAdd'
});

stream.addListener((event) => {
  console.log('New FIDO2 credential added:', event);
});
```

**Tendermint Event Subscription:**
```bash
curl -X POST http://localhost:26657/subscribe \
  -H "Content-Type: application/json" \
  -d '{"query": "tm.event='\''Tx'\'' AND provenance.smartaccounts.v1.EventFido2CredentialAdd.address EXISTS"}'
```

### Historical Analysis

For historical event analysis:

```javascript
// Query events by block range
const events = await client.searchTx({
  events: [
    { type: 'provenance.smartaccounts.v1.EventSmartAccountInit' }
  ],
  minHeight: 1000,
  maxHeight: 2000
});
```

## Integration Examples

### Event Processing Service

```typescript
interface SmartAccountEvent {
  type: string;
  address: string;
  credentialNumber?: number;
  credentialId?: string;
  blockHeight: number;
  txHash: string;
}

class SmartAccountEventProcessor {
  async processEvent(event: SmartAccountEvent) {
    switch (event.type) {
      case 'EventSmartAccountInit':
        await this.handleAccountInit(event);
        break;
      case 'EventFido2CredentialAdd':
        await this.handleFido2Add(event);
        break;
      case 'EventCosmosCredentialAdd':
        await this.handleCosmosAdd(event);
        break;
      case 'EventCredentialDelete':
        await this.handleCredentialDelete(event);
        break;
    }
  }

  private async handleAccountInit(event: SmartAccountEvent) {
    // Update user database
    // Send notification
    // Update analytics
  }

  private async handleFido2Add(event: SmartAccountEvent) {
    // Log security event
    // Update credential inventory
    // Trigger backup reminder
  }
}
```

### Audit Trail

```sql
-- Example database schema for event storage
CREATE TABLE smart_account_events (
  id SERIAL PRIMARY KEY,
  event_type VARCHAR(100) NOT NULL,
  account_address VARCHAR(50) NOT NULL,
  credential_number BIGINT,
  credential_id TEXT,
  block_height BIGINT NOT NULL,
  tx_hash VARCHAR(64) NOT NULL,
  timestamp TIMESTAMP NOT NULL,
  
  INDEX idx_address (account_address),
  INDEX idx_credential (credential_number),
  INDEX idx_block (block_height),
  INDEX idx_type (event_type)
);
```

## Security and Privacy Considerations

### Information Disclosure

**Public Information in Events:**
- Account addresses (already public)
- Credential numbers (non-sensitive identifiers)
- Credential IDs (public keys, not private data)
- Block heights and timestamps

**No Sensitive Data:**
- Private keys are never included
- Biometric data is never exposed
- User personal information is not included

### Event Immutability

Events provide an immutable audit trail:
- Events cannot be modified after emission
- Blockchain provides cryptographic proof of event integrity
- Historical events remain accessible indefinitely

### Privacy Best Practices

**For Application Developers:**
- Avoid logging sensitive user data in event handlers
- Use credential numbers instead of credential IDs when possible
- Implement proper access controls for event data storage

**For Users:**
- Understand that account activity is publicly visible
- Credential operations create permanent audit trails
- Consider privacy implications when choosing usernames or identifiers

## Testing Events

### Unit Tests

```go
func TestEventEmission(t *testing.T) {
  app := setupTestApp()
  ctx := app.BaseApp.NewContext(false, tmproto.Header{})
  
  // Register FIDO2 credential
  msg := &types.MsgRegisterFido2Credential{
    Sender: testAddress,
    EncodedAttestation: testAttestation,
    UserIdentifier: "test-user",
  }
  
  _, err := msgServer.RegisterFido2Credential(ctx, msg)
  require.NoError(t, err)
  
  // Verify event emission
  events := ctx.EventManager().Events()
  require.Len(t, events, 1)
  
  event := events[0]
  require.Equal(t, "provenance.smartaccounts.v1.EventFido2CredentialAdd", event.Type)
  
  // Check event attributes
  attrs := event.Attributes
  require.Equal(t, testAddress, string(attrs[0].Value))
}
```

### Integration Tests

```bash
#!/bin/bash
# Test event emission during credential operations

# Register credential and capture events
TX_HASH=$(provenanced tx smartaccounts register-fido2 \
  --encoded-attestation "$ATTESTATION" \
  --user-identifier "test" \
  --from alice \
  --chain-id testing \
  --yes \
  --output json | jq -r '.txhash')

# Query transaction events
provenanced query tx $TX_HASH --output json | \
  jq '.logs[0].events[] | select(.type | contains("smartaccounts"))'
```

## Performance Considerations

### Event Storage

Events contribute to blockchain state growth:
- Each event adds data to transaction logs
- Events are stored permanently on all nodes
- Consider event size when designing operations

### Query Performance

Event queries can be resource-intensive:
- Index events by commonly queried attributes
- Use block range limits for large queries
- Consider off-chain event storage for high-frequency queries

### Subscription Management

Real-time event monitoring:
- Limit concurrent subscriptions
- Implement proper error handling and reconnection logic
- Use appropriate buffer sizes for event streams