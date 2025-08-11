<!--
order: 1
-->

# Concepts

## Overview

Smart accounts are an enhanced type of account in the Provenance blockchain that extends the standard Cosmos SDK `BaseAccount` with flexible authentication capabilities. Unlike traditional externally owned accounts (EOAs) that rely on a single private key, smart accounts support multiple authentication methods and provide a more user-friendly experience.

## Traditional Account Limitations

### Externally Owned Accounts (EOAs)

Traditional blockchain accounts have several significant limitations:

- **Single Point of Failure**: One private key controls the entire account. If lost, access to funds is permanently gone.
- **Poor User Experience**: Users must manage complex recovery phrases and private keys manually.
- **Limited Device Support**: Transferring accounts between devices requires manual sharing of sensitive key material.
- **No Key Rotation**: Once created, the private key cannot be changed without creating a new account.
- **Security Risks**: Exposure to phishing attacks, malware, and human error in key management.
- **High Barrier to Entry**: New users must quickly learn complex security practices.

### Multi-Device Challenges

EOA-based systems typically lack robust multi-device support. Users accessing accounts on different devices often compromise security by:
- Manually transferring recovery phrases
- Using cloud storage for sensitive key material
- Reusing the same keys across multiple applications

## Smart Account Solution

### Core Concepts

Smart accounts address these limitations by introducing:

1. **Multiple Authentication Methods**: Support for various credential types including FIDO2/WebAuthn and traditional cryptographic keys.
2. **Credential Management**: Ability to add, remove, and rotate authentication credentials.
3. **Flexible Authentication**: Choose between different authentication methods for different use cases.
4. **Enhanced Security**: Support for hardware-backed authentication and biometric verification.

### Supported Authentication Methods

#### FIDO2/WebAuthn Credentials
- **Biometric Authentication**: Face ID, Touch ID, fingerprint readers
- **Hardware Security Keys**: YubiKey, Titan Keys, and other FIDO2-compliant devices
- **Platform Authenticators**: Built-in device authenticators (TPM, Secure Enclave)
- **Passkeys**: Cross-platform, phishing-resistant credentials

#### Traditional Cryptographic Credentials
- **secp256k1 (K256)**: Standard Cosmos SDK key format
- **ED25519**: Edwards-curve digital signature algorithm
- **P256**: NIST P-256 elliptic curve

#### Session-Based Authentication (Future)
- **Temporary Credentials**: Short-lived authentication tokens
- **Block Height Expiry**: Credentials that expire at specific block heights
- **Use Case**: Improved UX for frequent transactions (e.g., trading)

## Account Structure

### ProvenanceAccount

A smart account is represented by the `ProvenanceAccount` type, which embeds a standard `BaseAccount` and adds:

```protobuf
message ProvenanceAccount {
  cosmos.auth.v1beta1.BaseAccount base_account = 1;
  uint64 smart_account_number = 2;
  repeated Credential credentials = 3;
  bool is_smart_account_only_authentication = 4;
}
```

### Credentials

Each credential represents an authentication method:

```protobuf
message Credential {
  BaseCredential base_credential = 1;
  oneof authenticator {
    Fido2Authenticator   fido2_authenticator   = 2;
    K256Authenticator    k256_authenticator    = 3;
    SessionAuthenticator session_authenticator = 4;
  }
}
```

### Credential Types

The module supports different credential variants:

- `CREDENTIAL_TYPE_K256`: Standard secp256k1 keys (Cosmos default)
- `CREDENTIAL_TYPE_ED25519`: Edwards curve keys
- `CREDENTIAL_TYPE_P256`: NIST P-256 keys
- `CREDENTIAL_TYPE_WEBAUTHN`: Basic WebAuthn credentials
- `CREDENTIAL_TYPE_WEBAUTHN_UV`: WebAuthn with user verification

## Authentication Flow

### Transaction Signing Process

1. **Transaction Creation**: User creates a transaction as normal
2. **Signature Verification**: Custom ante handler intercepts the transaction
3. **Account Type Check**: Determines if the signer is a smart account
4. **Credential Selection**: Uses appropriate credential for authentication
5. **Verification**: Validates the signature/assertion against registered credentials

### FIDO2/WebAuthn Flow

For WebAuthn credentials:
1. **Challenge Generation**: Transaction hash becomes the challenge
2. **Assertion Creation**: Browser/device generates authentication assertion
3. **Signature Verification**: Module validates the assertion against stored credential

## Key Benefits

### Enhanced Security
- **Phishing Resistance**: FIDO2 credentials are bound to origins
- **Hardware Backing**: Support for secure enclaves and TPMs
- **No Shared Secrets**: Public key cryptography eliminates password-like vulnerabilities

### Improved User Experience
- **Familiar Authentication**: Users can use Face ID, Touch ID, and other familiar methods
- **Multi-Device Support**: Register multiple devices without sharing keys
- **Recovery Options**: Multiple credentials provide backup authentication methods

### Developer Benefits
- **Standard Compliance**: Built on established FIDO2/WebAuthn standards
- **Cosmos Integration**: Seamless integration with existing Cosmos SDK applications
- **Flexible Implementation**: Support for various authentication strategies

## Use Cases

### Individual Users
- **New Users**: Easier onboarding with familiar authentication methods
- **Multi-Device Users**: Seamless access across phones, laptops, and tablets
- **Security-Conscious Users**: Hardware key and biometric authentication

### Enterprise Applications
- **Employee Access**: Corporate device-based authentication
- **Compliance**: Auditable authentication with non-repudiation
- **Risk Management**: Granular control over authentication methods

### DeFi and Trading
- **Frequent Transactions**: Session-based authentication for improved UX
- **Security**: Hardware-backed signing for high-value transactions
- **Accessibility**: Lower barrier to entry for traditional finance users

## Compatibility

### Backward Compatibility
- Traditional EOAs continue to work unchanged
- Existing applications work without modification
- Users can choose when to upgrade to smart accounts

### Cosmos SDK Integration
- Built on standard Cosmos SDK account interfaces
- Compatible with existing ante handlers and middleware
- Integrates with standard transaction processing pipeline

## Security Considerations

### Threat Model
- **Key Compromise**: Multiple credentials reduce single point of failure
- **Device Loss**: Other registered devices can maintain access
- **Phishing Attacks**: FIDO2 provides origin binding protection
- **Malware**: Hardware-backed credentials resist local attacks

### Best Practices
- **Multi-Factor Setup**: Register multiple credential types
- **Regular Rotation**: Periodically update credentials
- **Secure Storage**: Leverage hardware security modules when available
- **Recovery Planning**: Maintain backup authentication methods