<!--
order: 2
-->

# State

The smart accounts module manages several key data structures to support multi-credential authentication. This section describes the on-chain state and how data is stored and accessed.

## Account Storage

### ProvenanceAccount

Smart accounts are stored using the standard Cosmos SDK account keeper, but with an extended account type:

```protobuf
message ProvenanceAccount {
  cosmos.auth.v1beta1.BaseAccount base_account = 1;
  uint64 smart_account_number = 2;
  repeated Credential credentials = 3;
  bool is_smart_account_only_authentication = 4;
}
```

**Fields:**
- `base_account`: Standard Cosmos SDK account containing address, account number, sequence, and public key
- `smart_account_number`: Globally unique identifier for the smart account (separate from account_number)
- `credentials`: List of registered authentication credentials
- `is_smart_account_only_authentication`: When true, only smart account authentication is allowed (traditional signature verification is disabled)

**Storage Key:** Accounts are stored in the auth module's account store using the account address as the key.

## Credential Management

### Credential Structure

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

### BaseCredential

Common credential metadata:

```protobuf
message BaseCredential {
  uint64 credential_number = 1;
  google.protobuf.Any public_key = 2;
  CredentialType variant = 3;
  int64 create_time = 4;
}
```

**Fields:**
- `credential_number`: Globally unique identifier assigned in order of creation
- `public_key`: The public key portion of the credential (format depends on credential type)
- `variant`: Type of credential (see CredentialType enum)
- `create_time`: Unix timestamp when the credential was created

### Credential Types

```protobuf
enum CredentialType {
  CREDENTIAL_TYPE_UNSPECIFIED = 0;
  CREDENTIAL_TYPE_ED25519 = 1;
  CREDENTIAL_TYPE_K256 = 2;
  CREDENTIAL_TYPE_P256 = 3;
  CREDENTIAL_TYPE_WEBAUTHN = 4;
  CREDENTIAL_TYPE_WEBAUTHN_UV = 5;
}
```

## Authenticator Types

### Fido2Authenticator

Stores FIDO2/WebAuthn credential information:

```protobuf
message Fido2Authenticator {
  string id = 1;
  string username = 2;
  bytes aaguid = 3;
  string credential_creation_response = 4;
  string rp_id = 5;
  string rp_origin = 6;
}
```

**Fields:**
- `id`: Base64url-encoded credential ID from the authenticator
- `username`: Human-readable username associated with the credential
- `aaguid`: Authenticator Attestation GUID (identifies device type/model)
- `credential_creation_response`: Base64-encoded attestation response from registration
- `rp_id`: Relying Party identifier (domain)
- `rp_origin`: Origin URL where the credential was created

### K256Authenticator

For traditional secp256k1 credentials:

```protobuf
message K256Authenticator {}
```

This authenticator type has no additional fields beyond the BaseCredential. The public key in BaseCredential contains the secp256k1 public key.

### SessionAuthenticator

For temporary session-based credentials (future implementation):

```protobuf
message SessionAuthenticator {
  int64 end_session_height = 1;
  bool timed_out = 2;
}
```

**Fields:**
- `end_session_height`: Block height at which the session expires
- `timed_out`: Flag indicating if the session has timed out

## Module Parameters

### Params

Global module configuration:

```protobuf
message Params {
  bool enabled = 1;
  uint32 max_credential_allowed = 2;
}
```

**Fields:**
- `enabled`: Whether the smart accounts module is active (controlled by governance)
- `max_credential_allowed`: Maximum number of credentials per smart account

**Storage Key:** `ParamsKey = []byte{0x01}`

**Default Values:**
```go
func DefaultParams() Params {
    return Params{
        Enabled:              false,  // Disabled by default
        MaxCredentialAllowed: 10,     // Allow up to 10 credentials per account
    }
}
```

## State Keys

The module uses the following key prefixes for state storage:

```go
const (
    // ParamsKey stores module parameters
    ParamsKey = iota + 1
)
```

**Note:** Smart accounts themselves are stored in the standard auth module account store, not in the smart accounts module store. The smart accounts module only stores its parameters.

## Genesis State

The genesis state contains the initial module parameters:

```protobuf
message GenesisState {
  Params params = 1;
  repeated ProvenanceAccount accounts = 2;
}
```

**Fields:**
- `params`: Initial module parameters
- `accounts`: List of smart accounts to create at genesis (typically empty)

## State Transitions

### Credential Registration

When a new credential is registered:

1. **Validation**: Check that the account exists and credential limit is not exceeded
2. **Credential Number Assignment**: Generate globally unique credential number
3. **Public Key Extraction**: Extract and validate the public key from the credential
4. **Account Update**: Add the credential to the account's credential list
5. **Storage**: Update the account in the auth module store

### Credential Deletion

When a credential is deleted:

1. **Validation**: Verify the credential exists and the sender owns the account
2. **Safety Check**: Ensure at least one credential remains (prevent account lockout)
3. **Removal**: Remove the credential from the account's credential list
4. **Storage**: Update the account in the auth module store

### Account Creation

Smart accounts are created through the standard auth module account creation process, but with additional smart account metadata.

## Query State

The module provides queries to access stored state:

### Account Query

Retrieve a smart account by address:
- **Request**: Account address
- **Response**: Full ProvenanceAccount structure
- **Path**: `/provenance/smartaccount/v1/account`

### Parameters Query

Retrieve current module parameters:
- **Request**: Empty
- **Response**: Current Params
- **Path**: `/provenance/smartaccount/v1/params`

## State Consistency

### Invariants

The module maintains the following invariants:

1. **Credential Uniqueness**: Each credential number is globally unique
2. **Credential Limit**: No account exceeds the maximum credential limit
3. **Account Integrity**: Smart accounts always have at least one valid credential
4. **Type Consistency**: Credential authenticator type matches the credential variant

### Validation

State validation occurs at:
- Transaction processing (msg validation)
- Block processing (ante handler validation)
- Genesis import/export
- Invariant checks (if implemented)

## Migration Considerations

### Upgrading Regular Accounts

Regular Cosmos SDK accounts can be upgraded to smart accounts by:
1. Creating a new ProvenanceAccount with the same BaseAccount
2. Adding initial credentials during the upgrade process
3. Replacing the account in the auth module store

### Backward Compatibility

Smart accounts maintain full backward compatibility:
- BaseAccount fields remain accessible
- Standard Cosmos SDK account interfaces are implemented
- Existing tooling continues to work without modification