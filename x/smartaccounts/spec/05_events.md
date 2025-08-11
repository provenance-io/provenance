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
```go
package main

import (
	"context"
	"cosmossdk.io/math"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/arnabmitra/simple-provenance-client/sa_testing_tools/temp_util"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	txservice "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/bank"
	smartaccountmodule "github.com/provenance-io/provenance/x/smartaccounts/module"
	"github.com/provenance-io/provenance/x/smartaccounts/types"
	"google.golang.org/grpc"
	"log"
	"strings"
)

func init() {
	// Set the Bech32 prefix to "tp"
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("tp", "tp"+sdk.PrefixPublic)
	config.Seal()
}

func broadcastTx() error {
	// Choose your codec: Amino or Protobuf. Here, we use Protobuf, given by the following function.
	encCfg := testutilmod.MakeTestEncodingConfig(bank.AppModuleBasic{}, auth.AppModuleBasic{}, smartaccountmodule.AppModuleBasic{})

	// Create a new TxBuilder.
	txBuilder := encCfg.TxConfig.NewTxBuilder()
	// this is just a test account, for local testing
	privKey, _ := temp_util.PrivKeyFromHex("f109a351d02607503221102905585f29c01dce1e9fb8a3afcb352f357021d2d7")
	pub := privKey.PubKey()
	addr := sdk.AccAddress(pub.Address())
	fmt.Printf("the from address is %s\n", addr)

	// Create the MsgRegisterFido2Credential message
	msg := &types.MsgRegisterFido2Credential{
		Sender: addr.String(),

		EncodedAttestation: "eyJpZCI6InAtOTNIaVpmRVpQX0ZYNURNY3dvaGciLCJyYXdJZCI6InAtOTNIaVpmRVpQX0ZYNURNY3dvaGciLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWaVVTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJOZEFBQUFBT3FialdaTkFSMGhQT1MydEl5MWRkUUFFS2Z2ZHg0bVh4R1RfeFYtUXpITUtJYWxBUUlESmlBQklWZ2dWM2JBbjVaejJ1Z0JuRm9QVXIyR0RIaXZTaE50MjYxWmROaUpuaDVYV00waVdDQmlFblc0MGtzYThreFp6RmkxcV9RN2x0MmU5ZnhnOThXZDN0S0hDZ19tX1EiLCJjbGllbnREYXRhSlNPTiI6ImV5SjBlWEJsSWpvaWQyVmlZWFYwYUc0dVkzSmxZWFJsSWl3aVkyaGhiR3hsYm1kbElqb2lSbTFWVnpaRlVXUnRTMEpOWlMxNlJXRnRaR1ZhYzFOU1lub3RSekZxWWpOellYbDBXVWgxYzJzMlFTSXNJbTl5YVdkcGJpSTZJbWgwZEhBNkx5OXNiMk5oYkdodmMzUTZNVGd3T0RBaWZRIn0sInR5cGUiOiJwdWJsaWMta2V5IiwiYXV0aGVudGljYXRvckF0dGFjaG1lbnQiOiJwbGF0Zm9ybSJ9", // foo6
		
		UserIdentifier: "foo6",
	}

	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return err
	}

	// Create a connection to the gRPC server.
	grpcConn, _ := grpc.Dial(
		"127.0.0.1:9090",    // Or your gRPC server address.
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
	)
	defer grpcConn.Close()

	// Broadcast the tx via gRPC. We create a new client for the Protobuf Tx service.
	clientCtx := context.Background()
	txSvcClient := txservice.NewServiceClient(grpcConn)

	// Get account number and sequence dynamically. This is needed for both simulation and signing.
	accNum, accSeq, err := temp_util.GetAccountInfo(grpcConn, addr, encCfg.Codec)
	if err != nil {
		return err
	}

	// To simulate a transaction, we need to build a temporary transaction with a dummy signature.
	// The signature doesn't need to be valid; it just needs to be present with the correct public key
	// and sequence number for the simulation to accurately estimate gas costs.
	simSigV2 := signing.SignatureV2{
		PubKey: pub,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil, // A nil signature is a valid dummy signature.
		},
		Sequence: accSeq,
	}
	if err := txBuilder.SetSignatures(simSigV2); err != nil {
		return fmt.Errorf("failed to set dummy signature for simulation: %w", err)
	}

	// Encode the transaction with the dummy signature for the simulation request.
	simTxBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return fmt.Errorf("failed to encode tx for simulation: %w", err)
	}

	// Run the simulation.
	simRes, err := txSvcClient.Simulate(
		context.Background(),
		&txservice.SimulateRequest{
			TxBytes: simTxBytes,
		},
	)
	if err != nil {
		return fmt.Errorf("transaction simulation failed: %w", err)
	}

	// We'll use the simulated gas used and add a buffer (e.g., 50%) to get our gas limit.
	// This helps prevent "out of gas" errors if the real execution uses slightly more gas.
	const gasAdjustment = 1.5
	gasLimit := uint64(float64(simRes.GasInfo.GasUsed) * gasAdjustment)

	// Define the gas price. This could be a constant or fetched from chain params.
	// For Provenance, a common gas price is 1905nhash.
	gasPrice, err := sdk.ParseCoinNormalized("1nhash")
	if err != nil {
		return fmt.Errorf("failed to parse gas price: %w", err)
	}

	// Calculate the fee by multiplying the gas limit by the gas price.
	feeAmount := gasPrice.Amount.Mul(math.NewInt(int64(gasLimit)))
	fees := sdk.NewCoins(sdk.NewCoin(gasPrice.Denom, feeAmount))

	fmt.Printf("Dynamic Estimation Complete:\n")
	fmt.Printf("  - Gas Used (Simulated): %d\n", simRes.GasInfo.GasUsed)
	fmt.Printf("  - Gas Limit (%.2fx buffer): %d\n", gasAdjustment, gasLimit)
	fmt.Printf("  - Fee Calculated: %s\n", fees.String())

	// Now, set the dynamically estimated gas limit and fee on the transaction builder.
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(fees)

	privs := []cryptotypes.PrivKey{privKey}
	accNums := []uint64{accNum}
	accSeqs := []uint64{accSeq}

	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return err
	}

	txBytes1, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}

	// Compute the SHA-256 hash of the raw bytes.
	hash := sha256.Sum256(txBytes1)
	txHash := strings.ToUpper(hex.EncodeToString(hash[:]))

	fmt.Printf("Transaction Hash: %s\n", txHash)

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       "testing",
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(context.TODO(),
			signing.SignMode_SIGN_MODE_DIRECT, signerData,
			txBuilder, priv, encCfg.TxConfig, accSeqs[i])
		if err != nil {
			return err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)

	// Generated Protobuf-encoded bytes.
	txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}
	// Compute the SHA-256 hash of the raw bytes.
	hashFinal := sha256.Sum256(txBytes)
	txHashFinal := strings.ToUpper(hex.EncodeToString(hashFinal[:]))
	fmt.Printf("Transaction Hash: %s\n", txHashFinal)

	// Generate a JSON string.
	txJSONBytes, err := encCfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}
	txJSON := string(txJSONBytes)
	fmt.Printf("the txJSON is %s\n", txJSON)

	grpcRes, err := txSvcClient.BroadcastTx(
		clientCtx,
		&txservice.BroadcastTxRequest{
			Mode:    txservice.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes, // Proto-binary of the signed transaction, see previous step.
		},
	)
	if err != nil {
		return err
	}

	fmt.Println(grpcRes.TxResponse.Code) // Should be `0` if the tx is successful
	fmt.Printf("the tx hash is %s\n", grpcRes.TxResponse.TxHash)
	return nil
}

func main() {
	err := broadcastTx()
	if err != nil {
		log.Fatalf("failed to broadcast transaction: %v", err)
	}
}


```
```bash
#!/bin/bash
# Test event emission during credential operations

# Register credential and capture events
TX_HASH=$(provenanced tx smartaccounts add-webauthn-credentials \
  --encodedAttestation "$ATTESTATION" \
  --user-identifier "test" \
  --from alice \
  --chain-id testing \
  --yes \
  --output json | jq -r '.txhash')

# Query transaction events (handles both old and new tx response structures)
provenanced query tx $TX_HASH --output json | \
  jq '
    if .logs then 
      .logs[] | .events[]? | select(.type | contains("smartaccounts"))
    else 
      .events[]? | select(.type | contains("smartaccounts"))
    end'
```

event emitted during credential registration
```json
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "hash": "38205A21BFBABE4A7B2FF516D68DA4412D0C1EAD908F1C929C78C81FFAF1CCD1",
    "height": "167708",
    "index": 0,
    "tx_result": {
      "code": 0,
      "data": "Eq4bCj8vcHJvdmVuYW5jZS5zbWFydGFjY291bnRzLnYxLk1zZ1JlZ2lzdGVyRmlkbzJDcmVkZW50aWFsUmVzcG9uc2US6hoIAhLlGgp3Cil0cDF3NDBxM3E3djI2cGV0dzZnNXN0ejVkdDl4c2V6Z256YWx4Z3c4eBJGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQIYayDKR2UDAKzrg/qSftcodkl6RzUmh3uMxsfwL3vXshgKIAEa0wgK4QES1gEKLS9wcm92ZW5hbmNlLnNtYXJ0YWNjb3VudHMudjEuRUMyUHVibGljS2V5RGF0YRKkAQpcCk2lAQIDJiABIVggW2VYdwEJYJny/mALUsijd6+6peaXdq+n13SmAnpvIj8iWCCi6VIpkSDR1KUlcq6x8ohEBxsljGbR76o532aYh4BP5hACGPn//////////wEQARogW2VYdwEJYJny/mALUsijd6+6peaXdq+n13SmAnpvIj8iIKLpUimRINHUpSVyrrHyiEQHGyWMZtHvqjnfZpiHgE/mGAUgzoKQxAYS7AYKGy1oV2VveWlJVDRQYy1ETm53YTdFb3VQWmVVcxIFZm9vMTMaEPv8MAcVTk7MjAtuAgVX170ikAZleUpwWkNJNklpMW9WMlZ2ZVdsSlZEUlFZeTFFVG01M1lUZEZiM1ZRV21WVmN5SXNJbkpoZDBsa0lqb2lMV2hYWlc5NWFVbFVORkJqTFVST2JuZGhOMFZ2ZFZCYVpWVnpJaXdpY21WemNHOXVjMlVpT25zaVlYUjBaWE4wWVhScGIyNVBZbXBsWTNRaU9pSnZNazV0WWxoU2EySnRPWFZhVjJSb1pFaFNWR1JITVRCdlIyaG9aRmhTYjFKSFJqQlpWbWxaVTFwWlRqVlpaMDlxUjJnd1RrSmpVRnBJV21kWE5GOXJjbkp0YVdocVRFaHRWbnA2ZFc5TlpHd3lUbVJCUVVGQlFWQjJPRTFCWTFaVWF6ZE5ha0YwZFVGblZsZ3hOekJCUmxCdlZtNXhUVzlwUlMxRU0xQm5lbG80UjNWNFMweHFNbGhzVEhCUlJVTkJlVmxuUVZOR1dVbEdkR3hYU0dOQ1ExZERXamgyTldkRE1VeEpiek5sZG5WeFdHMXNNMkYyY0Rsa01IQm5TalppZVVsZlNXeG5aMjkxYkZOTFdrVm5NR1JUYkVwWVMzVnpaa3RKVWtGallrcFplRzB3WlMxeFQyUTViVzFKWlVGVUxWa2lMQ0pqYkdsbGJuUkVZWFJoU2xOUFRpSTZJbVY1U2pCbFdFSnNTV3B2YVdReVZtbFpXRll3WVVjMGRWa3pTbXhaV0ZKc1NXbDNhVmt5YUdoaVIzaHNZbTFrYkVscWIybFBWbFkxVGpOYVdsUXpZM2xYVlRsQ1RsaHNhRlJVUW1sa1dGcDVVbFZWTTJKRlRraFBSM041VjFWVmVHSXdPREZUU0ZwQ1ZURnNkazVEU1hOSmJUbDVZVmRrY0dKcFNUWkpiV2d3WkVoQk5reDVPWE5pTWs1b1lrZG9kbU16VVRaTlZHZDNUMFJCYVdaUkluMHNJblI1Y0dVaU9pSndkV0pzYVdNdGEyVjVJaXdpWVhWMGFHVnVkR2xqWVhSdmNrRjBkR0ZqYUcxbGJuUWlPaUp3YkdGMFptOXliU0o5Kglsb2NhbGhvc3QyFmh0dHA6Ly9sb2NhbGhvc3Q6MTgwODAa1QgK4wEIARLWAQotL3Byb3ZlbmFuY2Uuc21hcnRhY2NvdW50cy52MS5FQzJQdWJsaWNLZXlEYXRhEqQBClwKTaUBAgMmIAEhWCDJljx9dXAv3KgMWbnUaNAx3DFC7z8eOlK+KtJqBcbiICJYIBEkwbiqwMTIJ4GvgBvcOCXN60rmn7pDvPkFv8FATarhEAIY+f//////////ARABGiDJljx9dXAv3KgMWbnUaNAx3DFC7z8eOlK+KtJqBcbiICIgESTBuKrAxMgnga+AG9w4Jc3rSuafukO8+QW/wUBNquEYBSCaiqXEBhLsBgobcWdNbU1sZDdIOG16QWhLN2tKNEdXVDRQeW1JEgVmb28xNBoQ+/wwBxVOTsyMC24CBVfXvSKQBmV5SnBaQ0k2SW5GblRXMU5iR1EzU0RodGVrRm9TemRyU2pSSFYxUTBVSGx0U1NJc0luSmhkMGxrSWpvaWNXZE5iVTFzWkRkSU9HMTZRV2hMTjJ0S05FZFhWRFJRZVcxSklpd2ljbVZ6Y0c5dWMyVWlPbnNpWVhSMFpYTjBZWFJwYjI1UFltcGxZM1FpT2lKdk1rNXRZbGhTYTJKdE9YVmFWMlJvWkVoU1ZHUkhNVEJ2UjJob1pGaFNiMUpIUmpCWlZtbFpVMXBaVGpWWlowOXFSMmd3VGtKalVGcElXbWRYTkY5cmNuSnRhV2hxVEVodFZucDZkVzlOWkd3eVRtUkJRVUZCUVZCMk9FMUJZMVpVYXpkTmFrRjBkVUZuVmxneE56QkJSa3R2UkVwcVNsaGxlRjlLYzNkSlUzVTFRMlZDYkdzdFJEaHdhWEJSUlVOQmVWbG5RVk5HV1VsTmJWZFFTREV4WTBOZlkzRkJlRnAxWkZKdk1FUklZMDFWVEhaUWVEUTJWWEkwY1RCdGIwWjRkVWxuU1d4blowVlRWRUoxUzNKQmVFMW5ibWRoTFVGSE9YYzBTbU16Y2xOMVlXWjFhMDg0TFZGWFgzZFZRazV4ZFVVaUxDSmpiR2xsYm5SRVlYUmhTbE5QVGlJNkltVjVTakJsV0VKc1NXcHZhV1F5Vm1sWldGWXdZVWMwZFZrelNteFpXRkpzU1dsM2FWa3lhR2hpUjNoc1ltMWtiRWxxYjJsV1JWSjBVekpLVDFwRlJuTmpWbWh6VDFSS1IxTldVbmhPUjNBMlQxZHdURTFWYUd0Uk1qbExVMWR6ZUdNeU5IcFdWWGd5VmtSa05sRlRTWE5KYlRsNVlWZGtjR0pwU1RaSmJXZ3daRWhCTmt4NU9YTmlNazVvWWtkb2RtTXpVVFpOVkdkM1QwUkJhV1pSSW4wc0luUjVjR1VpT2lKd2RXSnNhV010YTJWNUlpd2lZWFYwYUdWdWRHbGpZWFJ2Y2tGMGRHRmphRzFsYm5RaU9pSndiR0YwWm05eWJTSjkqCWxvY2FsaG9zdDIWaHR0cDovL2xvY2FsaG9zdDoxODA4MBq7CArjAQgCEtYBCi0vcHJvdmVuYW5jZS5zbWFydGFjY291bnRzLnYxLkVDMlB1YmxpY0tleURhdGESpAEKXApNpQECAyYgASFYIFd2wJ+Wc9roAZxaD1K9hgx4r0oTbdutWXTYiZ4eV1jNIlggYhJ1uNJLGvJMWcxYtav0O5bdnvX8YPfFnd7ShwoP5v0QAhj5//////////8BEAEaIFd2wJ+Wc9roAZxaD1K9hgx4r0oTbdutWXTYiZ4eV1jNIiBiEnW40ksa8kxZzFi1q/Q7lt2e9fxg98Wd3tKHCg/m/RgFIIuAsMQGEtIGChZwLTkzSGlaZkVaUF9GWDVETWN3b2hnEgRmb282GhDqm41mTQEdITzktrSMtXXUIvwFZXlKcFpDSTZJbkF0T1ROSWFWcG1SVnBRWDBaWU5VUk5ZM2R2YUdjaUxDSnlZWGRKWkNJNkluQXRPVE5JYVZwbVJWcFFYMFpZTlVSTlkzZHZhR2NpTENKeVpYTndiMjV6WlNJNmV5SmhkSFJsYzNSaGRHbHZiazlpYW1WamRDSTZJbTh5VG0xaVdGSnJZbTA1ZFZwWFpHaGtTRkpVWkVjeE1HOUhhR2hrV0ZKdlVrZEdNRmxXYVZWVFdsbE9OVmxuVDJwSGFEQk9RbU5RV2toYVoxYzBYMnR5Y20xcGFHcE1TRzFXZW5wMWIwMWtiREpPWkVGQlFVRkJUM0ZpYWxkYVRrRlNNR2hRVDFNeWRFbDVNV1JrVVVGRlMyWjJaSGcwYlZoNFIxUmZlRll0VVhwSVRVdEpZV3hCVVVsRVNtbEJRa2xXWjJkV00ySkJialZhZWpKMVowSnVSbTlRVlhJeVIwUklhWFpUYUU1ME1qWXhXbVJPYVVwdWFEVllWMDB3YVZkRFFtbEZibGMwTUd0ellUaHJlRnA2Um1reGNWOVJOMngwTW1VNVpuaG5PVGhYWkROMFMwaERaMTl0WDFFaUxDSmpiR2xsYm5SRVlYUmhTbE5QVGlJNkltVjVTakJsV0VKc1NXcHZhV1F5Vm1sWldGWXdZVWMwZFZrelNteFpXRkpzU1dsM2FWa3lhR2hpUjNoc1ltMWtiRWxxYjJsU2JURldWbnBhUmxWWFVuUlRNRXBPV2xNeE5sSlhSblJhUjFaaFl6Rk9VMWx1YjNSU2VrWnhXV3BPZWxsWWJEQlhWV2d4WXpKek1sRlRTWE5KYlRsNVlWZGtjR0pwU1RaSmJXZ3daRWhCTmt4NU9YTmlNazVvWWtkb2RtTXpVVFpOVkdkM1QwUkJhV1pSSW4wc0luUjVjR1VpT2lKd2RXSnNhV010YTJWNUlpd2lZWFYwYUdWdWRHbGpZWFJ2Y2tGMGRHRmphRzFsYm5RaU9pSndiR0YwWm05eWJTSjkqCWxvY2FsaG9zdDIWaHR0cDovL2xvY2FsaG9zdDoxODA4MA==",
      "log": "",
      "info": "",
      "gas_wanted": "4000000",
      "gas_used": "226526",
      "events": [
        {
          "type": "coin_spent",
          "attributes": [
            {
              "key": "spender",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            },
            {
              "key": "amount",
              "value": "6000000000nhash",
              "index": true
            }
          ]
        },
        {
          "type": "coin_received",
          "attributes": [
            {
              "key": "receiver",
              "value": "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt",
              "index": true
            },
            {
              "key": "amount",
              "value": "6000000000nhash",
              "index": true
            }
          ]
        },
        {
          "type": "transfer",
          "attributes": [
            {
              "key": "recipient",
              "value": "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt",
              "index": true
            },
            {
              "key": "sender",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            },
            {
              "key": "amount",
              "value": "6000000000nhash",
              "index": true
            }
          ]
        },
        {
          "type": "message",
          "attributes": [
            {
              "key": "sender",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            }
          ]
        },
        {
          "type": "tx",
          "attributes": [
            {
              "key": "fee",
              "value": "9000000000nhash",
              "index": true
            },
            {
              "key": "fee_payer",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            }
          ]
        },
        {
          "type": "tx",
          "attributes": [
            {
              "key": "min_fee_charged",
              "value": "6000000000nhash",
              "index": true
            },
            {
              "key": "fee_payer",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            }
          ]
        },
        {
          "type": "tx",
          "attributes": [
            {
              "key": "acc_seq",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x/2",
              "index": true
            }
          ]
        },
        {
          "type": "tx",
          "attributes": [
            {
              "key": "signature",
              "value": "MqObQ0R34o3kofEbz2k7d1KLBxfmIPmniZERRjznCfM8b/Etz3eJoL+j8rVHXCkB57aR2B6+/AYvR/Lbwttl+A==",
              "index": true
            }
          ]
        },
        {
          "type": "message",
          "attributes": [
            {
              "key": "action",
              "value": "/provenance.smartaccounts.v1.MsgRegisterFido2Credential",
              "index": true
            },
            {
              "key": "sender",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            },
            {
              "key": "module",
              "value": "smartaccounts",
              "index": true
            },
            {
              "key": "msg_index",
              "value": "0",
              "index": true
            }
          ]
        },
        {
          "type": "provenance.smartaccounts.v1.EventFido2CredentialAdd",
          "attributes": [
            {
              "key": "address",
              "value": "\"tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x\"",
              "index": true
            },
            {
              "key": "credential_id",
              "value": "\"p-93HiZfEZP_FX5DMcwohg\"",
              "index": true
            },
            {
              "key": "credential_number",
              "value": "\"2\"",
              "index": true
            },
            {
              "key": "msg_index",
              "value": "0",
              "index": true
            }
          ]
        },
        {
          "type": "coin_spent",
          "attributes": [
            {
              "key": "spender",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            },
            {
              "key": "amount",
              "value": "3000000000nhash",
              "index": true
            }
          ]
        },
        {
          "type": "coin_received",
          "attributes": [
            {
              "key": "receiver",
              "value": "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt",
              "index": true
            },
            {
              "key": "amount",
              "value": "3000000000nhash",
              "index": true
            }
          ]
        },
        {
          "type": "transfer",
          "attributes": [
            {
              "key": "recipient",
              "value": "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt",
              "index": true
            },
            {
              "key": "sender",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            },
            {
              "key": "amount",
              "value": "3000000000nhash",
              "index": true
            }
          ]
        },
        {
          "type": "message",
          "attributes": [
            {
              "key": "sender",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            }
          ]
        },
        {
          "type": "tx",
          "attributes": [
            {
              "key": "fee_payer",
              "value": "tp1w40q3q7v26petw6g5stz5dt9xsezgnzalxgw8x",
              "index": true
            },
            {
              "key": "basefee",
              "value": "6000000000nhash",
              "index": true
            },
            {
              "key": "fee_overage",
              "value": "3000000000nhash",
              "index": true
            },
            {
              "key": "total",
              "value": "9000000000nhash",
              "index": true
            }
          ]
        }
      ],
      "codespace": ""
    },
    "tx": "Cu8GCuwGCjcvcHJvdmVuYW5jZS5zbWFydGFjY291bnRzLnYxLk1zZ1JlZ2lzdGVyRmlkbzJDcmVkZW50aWFsErAGCil0cDF3NDBxM3E3djI2cGV0dzZnNXN0ejVkdDl4c2V6Z256YWx4Z3c4eBL8BWV5SnBaQ0k2SW5BdE9UTklhVnBtUlZwUVgwWllOVVJOWTNkdmFHY2lMQ0p5WVhkSlpDSTZJbkF0T1ROSWFWcG1SVnBRWDBaWU5VUk5ZM2R2YUdjaUxDSnlaWE53YjI1elpTSTZleUpoZEhSbGMzUmhkR2x2Yms5aWFtVmpkQ0k2SW04eVRtMWlXRkpyWW0wNWRWcFhaR2hrU0ZKVVpFY3hNRzlIYUdoa1dGSnZVa2RHTUZsV2FWVlRXbGxPTlZsblQycEhhREJPUW1OUVdraGFaMWMwWDJ0eWNtMXBhR3BNU0cxV2VucDFiMDFrYkRKT1pFRkJRVUZCVDNGaWFsZGFUa0ZTTUdoUVQxTXlkRWw1TVdSa1VVRkZTMloyWkhnMGJWaDRSMVJmZUZZdFVYcElUVXRKWVd4QlVVbEVTbWxCUWtsV1oyZFdNMkpCYmpWYWVqSjFaMEp1Um05UVZYSXlSMFJJYVhaVGFFNTBNall4V21ST2FVcHVhRFZZVjAwd2FWZERRbWxGYmxjME1HdHpZVGhyZUZwNlJta3hjVjlSTjJ4ME1tVTVabmhuT1RoWFpETjBTMGhEWjE5dFgxRWlMQ0pqYkdsbGJuUkVZWFJoU2xOUFRpSTZJbVY1U2pCbFdFSnNTV3B2YVdReVZtbFpXRll3WVVjMGRWa3pTbXhaV0ZKc1NXbDNhVmt5YUdoaVIzaHNZbTFrYkVscWIybFNiVEZXVm5wYVJsVlhVblJUTUVwT1dsTXhObEpYUm5SYVIxWmhZekZPVTFsdWIzUlNla1p4V1dwT2VsbFliREJYVldneFl6SnpNbEZUU1hOSmJUbDVZVmRrY0dKcFNUWkpiV2d3WkVoQk5reDVPWE5pTWs1b1lrZG9kbU16VVRaTlZHZDNUMFJCYVdaUkluMHNJblI1Y0dVaU9pSndkV0pzYVdNdGEyVjVJaXdpWVhWMGFHVnVkR2xqWVhSdmNrRjBkR0ZqYUcxbGJuUWlPaUp3YkdGMFptOXliU0o5GgRmb282Em8KUApGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQIYayDKR2UDAKzrg/qSftcodkl6RzUmh3uMxsfwL3vXshIECgIIARgCEhsKEwoFbmhhc2gSCjkwMDAwMDAwMDAQgLTEwyEaQDKjm0NEd+KN5KHxG89pO3dSiwcX5iD5p4mREUY85wnzPG/xLc93iaC/o/K1R1wpAee2kdgevvwGL0fy28LbZfg="
  }
}
```
