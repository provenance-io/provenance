<!--
order: 9
-->

# Integration Guide

This guide provides comprehensive instructions for integrating smart accounts into applications, wallets, and services built on the Provenance blockchain.

## Overview

Smart accounts provide enhanced authentication capabilities that can significantly improve user experience while maintaining security. This guide covers integration patterns, best practices, and common use cases.

## Prerequisites

### System Requirements

**Blockchain Integration:**
- Provenance blockchain node access (mainnet, testnet, or local)
- Cosmos SDK v0.47+ compatible client libraries
- gRPC and REST API connectivity

**Client-Side Requirements:**
- WebAuthn/FIDO2 capable browser or environment
- HTTPS connection (required for WebAuthn)
- Modern JavaScript runtime for web applications

**Development Environment:**
- Go 1.19+ for blockchain development
- Node.js 16+ for web client development
- TypeScript support recommended

### Module Availability

Check if smart accounts are enabled:

```bash
# Query module parameters
provenanced query smartaccounts params

# Expected response when enabled:
# params:
#   enabled: true
#   max_credential_allowed: 10
```

## Quick Start

### 1. Basic Smart Account Creation

**Using CLI:**
```bash
# Register first credential (creates smart account)
provenanced tx smartaccounts register-fido2 \
  --encoded-attestation "$WEBAUTHN_ATTESTATION" \
  --user-identifier "user@example.com" \
  --from alice \
  --chain-id provenance-1
```

**Using Go SDK:**
```go
import (
    "github.com/provenance-io/provenance/x/smartaccounts/types"
)

// Create registration message
msg := &types.MsgRegisterFido2Credential{
    Sender:             senderAddr.String(),
    EncodedAttestation: base64EncodedAttestation,
    UserIdentifier:     "user@example.com",
}

// Submit transaction
txResponse, err := client.BroadcastTx(ctx, msg)
```

### 2. Query Smart Account

```bash
# Query account details
provenanced query smartaccounts account tp1...address...
```

### 3. Authenticate Transaction

**WebAuthn Authentication:**
```javascript
// Generate assertion for transaction
const assertion = await navigator.credentials.get({
    publicKey: {
        challenge: transactionHash,
        allowCredentials: [{ 
            id: credentialId, 
            type: 'public-key' 
        }]
    }
});
```

## Integration Patterns

### Wallet Integration

#### Web Wallet Implementation

**1. WebAuthn Registration Flow:**

```typescript
class SmartAccountWallet {
    async registerWebAuthnCredential(username: string): Promise<CredentialRegistration> {
        // Step 1: Generate registration challenge
        const challenge = crypto.getRandomValues(new Uint8Array(32));
        
        // Step 2: Create WebAuthn credential
        const credential = await navigator.credentials.create({
            publicKey: {
                challenge: challenge,
                rp: { 
                    id: "wallet.provenance.io",
                    name: "Provenance Wallet" 
                },
                user: {
                    id: new TextEncoder().encode(username),
                    name: username,
                    displayName: username
                },
                pubKeyCredParams: [
                    { alg: -7, type: "public-key" },  // ES256
                    { alg: -257, type: "public-key" } // RS256
                ],
                authenticatorSelection: {
                    userVerification: "required"
                }
            }
        }) as PublicKeyCredential;
        
        // Step 3: Format attestation for blockchain
        const attestationResponse = credential.response as AuthenticatorAttestationResponse;
        const attestation = {
            id: credential.id,
            rawId: arrayBufferToBase64(credential.rawId),
            response: {
                attestationObject: arrayBufferToBase64(attestationResponse.attestationObject),
                clientDataJSON: arrayBufferToBase64(attestationResponse.clientDataJSON)
            },
            type: credential.type
        };
        
        return {
            encodedAttestation: btoa(JSON.stringify(attestation)),
            userIdentifier: username
        };
    }
    
    async signTransaction(tx: StdTx, credentialId: string): Promise<WebAuthnAssertion> {
        // Generate challenge from transaction
        const txBytes = serializeTransaction(tx);
        const challenge = await crypto.subtle.digest('SHA-256', txBytes);
        
        // Create WebAuthn assertion
        const assertion = await navigator.credentials.get({
            publicKey: {
                challenge: challenge,
                allowCredentials: [{
                    id: base64ToArrayBuffer(credentialId),
                    type: 'public-key'
                }],
                userVerification: 'required'
            }
        }) as PublicKeyCredential;
        
        const assertionResponse = assertion.response as AuthenticatorAssertionResponse;
        
        return {
            credentialId: credentialId,
            clientDataJSON: arrayBufferToBase64(assertionResponse.clientDataJSON),
            authenticatorData: arrayBufferToBase64(assertionResponse.authenticatorData),
            signature: arrayBufferToBase64(assertionResponse.signature),
            userHandle: assertionResponse.userHandle ? 
                arrayBufferToBase64(assertionResponse.userHandle) : null
        };
    }
}
```

**2. Account Management:**

```typescript
class SmartAccountManager {
    async getAccountCredentials(address: string): Promise<Credential[]> {
        const response = await this.queryClient.smartaccounts.smartAccount({
            address: address
        });
        return response.provenanceaccount?.credentials || [];
    }
    
    async addNewCredential(address: string, credentialType: 'fido2' | 'cosmos'): Promise<void> {
        if (credentialType === 'fido2') {
            const registration = await this.wallet.registerWebAuthnCredential(`user-${Date.now()}`);
            
            const msg = {
                typeUrl: '/provenance.smartaccounts.v1.MsgRegisterFido2Credential',
                value: {
                    sender: address,
                    encodedAttestation: registration.encodedAttestation,
                    userIdentifier: registration.userIdentifier
                }
            };
            
            await this.client.signAndBroadcast(address, [msg], 'auto');
        }
    }
    
    async removeCredential(address: string, credentialNumber: number): Promise<void> {
        const msg = {
            typeUrl: '/provenance.smartaccounts.v1.MsgDeleteCredential',
            value: {
                sender: address,
                credentialNumber: credentialNumber
            }
        };
        
        await this.client.signAndBroadcast(address, [msg], 'auto');
    }
}
```

#### Mobile Wallet Integration

**React Native Example:**

```typescript
import { WebAuthn } from 'react-native-webauthn';

class MobileSmartAccountWallet {
    async registerBiometricCredential(): Promise<CredentialRegistration> {
        const challenge = await this.generateChallenge();
        
        const credential = await WebAuthn.create({
            challenge: challenge,
            userVerification: 'required',
            authenticatorType: 'platform' // Use platform authenticator
        });
        
        return this.formatCredentialForBlockchain(credential);
    }
    
    async authenticateWithBiometrics(transactionData: any): Promise<WebAuthnAssertion> {
        const challenge = await this.generateTransactionChallenge(transactionData);
        
        const assertion = await WebAuthn.get({
            challenge: challenge,
            userVerification: 'required'
        });
        
        return this.formatAssertionForBlockchain(assertion);
    }
}
```

### DApp Integration

#### Frontend Integration

**1. Smart Account Detection:**

```typescript
class DAppSmartAccountIntegration {
    async detectSmartAccount(address: string): Promise<boolean> {
        try {
            const account = await this.queryClient.smartaccounts.smartAccount({
                address: address
            });
            return account.provenanceaccount !== null;
        } catch (error) {
            return false; // Not a smart account or doesn't exist
        }
    }
    
    async getAvailableAuthMethods(address: string): Promise<AuthMethod[]> {
        const account = await this.queryClient.smartaccounts.smartAccount({
            address: address
        });
        
        const methods: AuthMethod[] = [];
        
        for (const credential of account.provenanceaccount?.credentials || []) {
            switch (credential.baseCredential?.variant) {
                case 'CREDENTIAL_TYPE_WEBAUTHN':
                case 'CREDENTIAL_TYPE_WEBAUTHN_UV':
                    methods.push({
                        type: 'webauthn',
                        credentialNumber: credential.baseCredential.credentialNumber,
                        displayName: credential.fido2Authenticator?.username || 'WebAuthn Device'
                    });
                    break;
                case 'CREDENTIAL_TYPE_K256':
                    methods.push({
                        type: 'traditional',
                        credentialNumber: credential.baseCredential.credentialNumber,
                        displayName: 'Traditional Key'
                    });
                    break;
            }
        }
        
        return methods;
    }
}
```

**2. Transaction Signing Interface:**

```typescript
interface TransactionSigningOptions {
    authMethod?: 'webauthn' | 'traditional' | 'auto';
    credentialNumber?: number;
    userVerification?: 'required' | 'preferred' | 'discouraged';
}

class SmartAccountTransactionSigner {
    async signTransaction(
        address: string, 
        transaction: any, 
        options: TransactionSigningOptions = {}
    ): Promise<SignedTransaction> {
        
        const isSmartAccount = await this.detectSmartAccount(address);
        
        if (!isSmartAccount) {
            // Fallback to traditional signing
            return this.signTraditionally(address, transaction);
        }
        
        const authMethods = await this.getAvailableAuthMethods(address);
        const selectedMethod = this.selectAuthMethod(authMethods, options);
        
        switch (selectedMethod.type) {
            case 'webauthn':
                return this.signWithWebAuthn(address, transaction, selectedMethod);
            case 'traditional':
                return this.signTraditionally(address, transaction);
            default:
                throw new Error('No suitable authentication method available');
        }
    }
    
    private async signWithWebAuthn(
        address: string,
        transaction: any,
        authMethod: AuthMethod
    ): Promise<SignedTransaction> {
        
        const challenge = await this.generateTransactionChallenge(transaction);
        
        try {
            const assertion = await navigator.credentials.get({
                publicKey: {
                    challenge: challenge,
                    allowCredentials: [{
                        id: base64ToArrayBuffer(authMethod.credentialId),
                        type: 'public-key'
                    }],
                    userVerification: 'required'
                }
            });
            
            return this.formatSignedTransaction(transaction, assertion);
            
        } catch (error) {
            if (error.name === 'NotAllowedError') {
                throw new Error('User cancelled authentication');
            } else if (error.name === 'InvalidStateError') {
                throw new Error('Authenticator not available');
            }
            throw error;
        }
    }
}
```

### Backend Service Integration

#### Account Service

```go
package smartaccounts

import (
    "context"
    "fmt"
    
    "github.com/provenance-io/provenance/x/smartaccounts/types"
)

type SmartAccountService struct {
    queryClient types.QueryClient
    msgClient   types.MsgClient
}

func NewSmartAccountService(conn grpc.ClientConnInterface) *SmartAccountService {
    return &SmartAccountService{
        queryClient: types.NewQueryClient(conn),
        msgClient:   types.NewMsgClient(conn),
    }
}

func (s *SmartAccountService) IsSmartAccount(ctx context.Context, address string) (bool, error) {
    req := &types.SmartAccountQueryRequest{Address: address}
    resp, err := s.queryClient.SmartAccount(ctx, req)
    if err != nil {
        return false, nil // Not a smart account
    }
    return resp.Provenanceaccount != nil, nil
}

func (s *SmartAccountService) GetAccountCredentials(ctx context.Context, address string) ([]*types.Credential, error) {
    req := &types.SmartAccountQueryRequest{Address: address}
    resp, err := s.queryClient.SmartAccount(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to query smart account: %w", err)
    }
    
    if resp.Provenanceaccount == nil {
        return nil, fmt.Errorf("not a smart account")
    }
    
    return resp.Provenanceaccount.Credentials, nil
}

func (s *SmartAccountService) RegisterFido2Credential(
    ctx context.Context,
    address string,
    attestation string,
    userID string,
) (*types.MsgRegisterFido2CredentialResponse, error) {
    
    msg := &types.MsgRegisterFido2Credential{
        Sender:             address,
        EncodedAttestation: attestation,
        UserIdentifier:     userID,
    }
    
    return s.msgClient.RegisterFido2Credential(ctx, msg)
}
```

#### Authentication Middleware

```go
package middleware

import (
    "context"
    "net/http"
    "strings"
)

type SmartAccountAuthMiddleware struct {
    smartAccountService *SmartAccountService
}

func (m *SmartAccountAuthMiddleware) AuthenticateRequest(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract authentication information from request
        authHeader := r.Header.Get("Authorization")
        if !strings.HasPrefix(authHeader, "SmartAccount ") {
            // Fall back to traditional authentication
            next.ServeHTTP(w, r)
            return
        }
        
        // Parse smart account authentication
        token := strings.TrimPrefix(authHeader, "SmartAccount ")
        claims, err := m.verifySmartAccountToken(r.Context(), token)
        if err != nil {
            http.Error(w, "Invalid smart account authentication", http.StatusUnauthorized)
            return
        }
        
        // Add claims to request context
        ctx := context.WithValue(r.Context(), "smartAccountClaims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (m *SmartAccountAuthMiddleware) verifySmartAccountToken(ctx context.Context, token string) (*SmartAccountClaims, error) {
    // Implementation depends on your token format
    // Could be JWT with smart account information
    // Or custom format with WebAuthn assertion
    return nil, nil
}
```

## Best Practices

### Security Best Practices

**1. Credential Management:**
```typescript
class SecureCredentialManager {
    // Always verify credentials before use
    async verifyCredentialValidity(credentialId: string): Promise<boolean> {
        try {
            // Test credential with a dummy assertion
            await navigator.credentials.get({
                publicKey: {
                    challenge: new Uint8Array(32),
                    allowCredentials: [{
                        id: base64ToArrayBuffer(credentialId),
                        type: 'public-key'
                    }],
                    userVerification: 'required'
                }
            });
            return true;
        } catch {
            return false;
        }
    }
    
    // Implement credential backup strategy
    async ensureBackupCredentials(address: string): Promise<void> {
        const credentials = await this.getAccountCredentials(address);
        
        if (credentials.length < 2) {
            console.warn('Account has only one credential - consider adding backup');
            // Prompt user to add backup credential
        }
    }
}
```

**2. Error Handling:**
```typescript
class RobustSmartAccountClient {
    async executeWithFallback<T>(
        smartAccountOperation: () => Promise<T>,
        traditionalFallback: () => Promise<T>
    ): Promise<T> {
        try {
            return await smartAccountOperation();
        } catch (error) {
            console.warn('Smart account operation failed, falling back:', error);
            return await traditionalFallback();
        }
    }
    
    async signTransactionWithRetry(
        address: string,
        transaction: any,
        maxRetries: number = 3
    ): Promise<SignedTransaction> {
        let lastError: Error;
        
        for (let attempt = 1; attempt <= maxRetries; attempt++) {
            try {
                return await this.signTransaction(address, transaction);
            } catch (error) {
                lastError = error as Error;
                
                if (error.name === 'NotAllowedError') {
                    // User cancelled - don't retry
                    throw error;
                }
                
                if (attempt < maxRetries) {
                    console.warn(`Signing attempt ${attempt} failed, retrying:`, error);
                    await this.delay(1000 * attempt); // Exponential backoff
                }
            }
        }
        
        throw lastError!;
    }
    
    private delay(ms: number): Promise<void> {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}
```

### Performance Optimization

**1. Caching Strategy:**
```typescript
class OptimizedSmartAccountClient {
    private accountCache = new Map<string, CachedAccount>();
    private cacheTimeout = 5 * 60 * 1000; // 5 minutes
    
    async getAccountWithCache(address: string): Promise<ProvenanceAccount> {
        const cached = this.accountCache.get(address);
        const now = Date.now();
        
        if (cached && (now - cached.timestamp) < this.cacheTimeout) {
            return cached.account;
        }
        
        const account = await this.queryClient.smartaccounts.smartAccount({
            address: address
        });
        
        this.accountCache.set(address, {
            account: account.provenanceaccount!,
            timestamp: now
        });
        
        return account.provenanceaccount!;
    }
    
    // Pre-load frequently accessed accounts
    async preloadAccounts(addresses: string[]): Promise<void> {
        const promises = addresses.map(addr => this.getAccountWithCache(addr));
        await Promise.all(promises);
    }
}
```

**2. Batch Operations:**
```typescript
class BatchSmartAccountOperations {
    async batchQueryAccounts(addresses: string[]): Promise<Map<string, ProvenanceAccount | null>> {
        const results = new Map<string, ProvenanceAccount | null>();
        
        // Execute queries in parallel
        const promises = addresses.map(async (address) => {
            try {
                const response = await this.queryClient.smartaccounts.smartAccount({
                    address: address
                });
                return [address, response.provenanceaccount] as const;
            } catch (error) {
                return [address, null] as const;
            }
        });
        
        const responses = await Promise.all(promises);
        
        for (const [address, account] of responses) {
            results.set(address, account);
        }
        
        return results;
    }
}
```

### User Experience Guidelines

**1. Progressive Enhancement:**
```typescript
class ProgressiveSmartAccountUX {
    async setupAccount(userPreferences: UserPreferences): Promise<AccountSetup> {
        // Check WebAuthn availability
        const hasWebAuthn = await this.checkWebAuthnSupport();
        
        if (hasWebAuthn && userPreferences.preferBiometrics) {
            return this.setupWithWebAuthn();
        } else {
            return this.setupWithTraditionalKeys();
        }
    }
    
    private async checkWebAuthnSupport(): Promise<boolean> {
        return !!(
            window.PublicKeyCredential && 
            await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable()
        );
    }
    
    async promptForAuthMethod(availableMethods: AuthMethod[]): Promise<AuthMethod> {
        // Show user-friendly selection interface
        return this.showAuthMethodSelector(availableMethods);
    }
}
```

**2. Clear User Communication:**
```typescript
class UserFriendlyMessages {
    getErrorMessage(error: Error): string {
        switch (error.name) {
            case 'NotAllowedError':
                return 'Authentication was cancelled. Please try again.';
            case 'InvalidStateError':
                return 'Your authenticator is not available. Please check your device.';
            case 'NotSupportedError':
                return 'This authentication method is not supported on your device.';
            case 'AbortError':
                return 'Authentication timed out. Please try again.';
            default:
                return 'Authentication failed. Please try again or use a different method.';
        }
    }
    
    getCredentialTypeDescription(credentialType: string): string {
        switch (credentialType) {
            case 'CREDENTIAL_TYPE_WEBAUTHN_UV':
                return 'Biometric authentication (Face ID, Touch ID, etc.)';
            case 'CREDENTIAL_TYPE_WEBAUTHN':
                return 'Security key or device authenticator';
            case 'CREDENTIAL_TYPE_K256':
                return 'Traditional cryptographic key';
            default:
                return 'Unknown authentication method';
        }
    }
}
```

## Testing Integration

### Unit Testing

```typescript
describe('SmartAccountIntegration', () => {
    let mockQueryClient: jest.Mocked<QueryClient>;
    let smartAccountService: SmartAccountService;
    
    beforeEach(() => {
        mockQueryClient = createMockQueryClient();
        smartAccountService = new SmartAccountService(mockQueryClient);
    });
    
    test('should detect smart account correctly', async () => {
        mockQueryClient.smartaccounts.smartAccount.mockResolvedValue({
            provenanceaccount: mockSmartAccount
        });
        
        const isSmartAccount = await smartAccountService.isSmartAccount('tp1test...');
        expect(isSmartAccount).toBe(true);
    });
    
    test('should handle authentication errors gracefully', async () => {
        const mockError = new Error('NotAllowedError');
        mockError.name = 'NotAllowedError';
        
        jest.spyOn(navigator.credentials, 'get').mockRejectedValue(mockError);
        
        await expect(
            smartAccountService.authenticateTransaction('tx', 'credId')
        ).rejects.toThrow('User cancelled authentication');
    });
});
```

### Integration Testing

```bash
#!/bin/bash
# Integration test script

echo "Testing smart account registration..."
CRED_RESPONSE=$(provenanced tx smartaccounts register-fido2 \
  --encoded-attestation "$TEST_ATTESTATION" \
  --user-identifier "test-user" \
  --from alice \
  --chain-id testing \
  --yes \
  --output json)

CRED_NUMBER=$(echo $CRED_RESPONSE | jq -r '.logs[0].events[] | select(.type=="provenance.smartaccounts.v1.EventFido2CredentialAdd") | .attributes[] | select(.key=="credential_number") | .value')

echo "Credential registered with number: $CRED_NUMBER"

echo "Testing account query..."
provenanced query smartaccounts account $(echo $CRED_RESPONSE | jq -r '.logs[0].events[0].attributes[0].value')

echo "Testing transaction authentication..."
provenanced tx bank send alice bob 1000stake \
  --webauthn-assertion "$TEST_ASSERTION" \
  --chain-id testing \
  --yes

echo "Integration tests completed successfully!"
```

## Migration Guide

### From Traditional Accounts

**1. Account Upgrade Process:**
```typescript
class AccountMigrationService {
    async upgradeToSmartAccount(
        traditionalAddress: string,
        privateKey: string
    ): Promise<string> {
        // Step 1: Create WebAuthn credential
        const webauthnCredential = await this.createWebAuthnCredential();
        
        // Step 2: Register first smart credential (creates smart account)
        await this.registerFirstCredential(
            traditionalAddress,
            webauthnCredential,
            privateKey
        );
        
        // Step 3: Add traditional key as backup credential
        const publicKey = extractPublicKey(privateKey);
        await this.addTraditionalCredential(traditionalAddress, publicKey);
        
        // Step 4: Test both authentication methods
        await this.validateAllCredentials(traditionalAddress);
        
        return traditionalAddress; // Same address, now smart account
    }
}
```

**2. Backward Compatibility:**
```typescript
class BackwardCompatibleWallet {
    async signTransaction(address: string, transaction: any): Promise<SignedTransaction> {
        const isSmartAccount = await this.isSmartAccount(address);
        
        if (isSmartAccount) {
            // Use smart account signing
            return this.signWithSmartAccount(address, transaction);
        } else {
            // Use traditional signing
            return this.signWithTraditionalKey(address, transaction);
        }
    }
    
    // Gradual feature rollout
    async enableSmartAccountFeatures(address: string): Promise<void> {
        if (await this.isSmartAccount(address)) {
            this.showSmartAccountUI(address);
        } else {
            this.showUpgradePrompt(address);
        }
    }
}
```

## Troubleshooting

### Common Issues

**1. WebAuthn Not Working:**
```typescript
class WebAuthnTroubleshooting {
    async diagnoseWebAuthnIssues(): Promise<DiagnosticReport> {
        const report: DiagnosticReport = {
            webauthnSupport: !!window.PublicKeyCredential,
            platformAuthenticator: false,
            secureContext: window.isSecureContext,
            userAgent: navigator.userAgent
        };
        
        if (window.PublicKeyCredential) {
            report.platformAuthenticator = await PublicKeyCredential
                .isUserVerifyingPlatformAuthenticatorAvailable();
        }
        
        return report;
    }
    
    getRecommendations(report: DiagnosticReport): string[] {
        const recommendations: string[] = [];
        
        if (!report.webauthnSupport) {
            recommendations.push('Browser does not support WebAuthn');
        }
        
        if (!report.secureContext) {
            recommendations.push('WebAuthn requires HTTPS connection');
        }
        
        if (!report.platformAuthenticator) {
            recommendations.push('No platform authenticator available');
        }
        
        return recommendations;
    }
}
```

**2. Common Error Resolution:**

| Error | Cause | Solution |
|-------|-------|----------|
| `NotAllowedError` | User cancelled or timeout | Prompt user to try again |
| `InvalidStateError` | Authenticator unavailable | Check device/browser support |
| `NotSupportedError` | WebAuthn not supported | Fall back to traditional auth |
| `ConstraintError` | Invalid parameters | Verify credential requirements |
| `NetworkError` | Blockchain connectivity | Check node connection |

### Support Resources

**Documentation:**
- [WebAuthn Specification](https://www.w3.org/TR/webauthn-2/)
- [FIDO2 Developer Guide](https://fidoalliance.org/fido2/)
- [Cosmos SDK Documentation](https://docs.cosmos.network/)

**Developer Tools:**
- Browser DevTools for WebAuthn debugging
- Provenance CLI for testing
- Smart account simulation tools

**Community:**
- Provenance Discord for support
- GitHub Issues for bug reports
- Developer forums for best practices