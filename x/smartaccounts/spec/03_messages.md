<!--
order: 3
-->

# Messages

The smart accounts module provides several message types for managing smart accounts and their credentials. All messages are processed by the module's message server and emit corresponding events.

## Message Types

### MsgRegisterFido2Credential

Registers a new FIDO2/WebAuthn credential for a smart account.

```protobuf
message MsgRegisterFido2Credential {
  string sender = 1;
  string encoded_attestation = 2;
  string user_identifier = 3;
}
```

**Fields:**
- `sender`: The address of the account registering the credential
- `encoded_attestation`: Base64-encoded WebAuthn attestation response from the authenticator
- `user_identifier`: FIDO2 user identifier for the authenticator (stored separately during registration)

**Validation:**
- `sender` must be a valid bech32 address
- `encoded_attestation` must be valid base64 and contain a valid WebAuthn attestation
- `user_identifier` must not be empty
- Smart accounts module must be enabled
- Account must not exceed maximum credential limit

**State Changes:**
- Creates or updates a ProvenanceAccount with the new FIDO2 credential
- Assigns a unique credential number
- Stores the credential metadata including AAGUID, RP ID, and origin

**Events Emitted:**
- `EventFido2CredentialAdd` with account address, credential number, and credential ID

#### Response

```protobuf
message MsgRegisterFido2CredentialResponse {
  uint64 credential_number = 1;
  ProvenanceAccount provenance_account = 2;
}
```

**Fields:**
- `credential_number`: Globally unique identifier assigned to the new credential
- `provenance_account`: Complete smart account data after registration

### MsgRegisterCosmosCredential

Registers a traditional cryptographic credential (secp256k1, Ed25519, etc.) for a smart account.

```protobuf
message MsgRegisterCosmosCredential {
  string sender = 1;
  google.protobuf.Any pubkey = 2;
}
```

**Fields:**
- `sender`: The address of the account registering the credential
- `pubkey`: The public key to register (wrapped in protobuf Any for type flexibility)

**Validation:**
- `sender` must be a valid bech32 address
- `pubkey` must contain a valid public key of a supported type
- Smart accounts module must be enabled
- Account must not exceed maximum credential limit
- Public key must not already be registered for this account

**State Changes:**
- Creates or updates a ProvenanceAccount with the new cryptographic credential
- Assigns a unique credential number
- Stores the public key and credential type

**Events Emitted:**
- `EventCosmosCredentialAdd` with account address and credential number

#### Response

```protobuf
message MsgRegisterCosmosCredentialResponse {
  uint64 credential_number = 1;
}
```

**Fields:**
- `credential_number`: Globally unique identifier assigned to the new credential

### MsgDeleteCredential

Removes a credential from a smart account.

```protobuf
message MsgDeleteCredential {
  string sender = 1;
  uint64 credential_number = 2;
}
```

**Fields:**
- `sender`: The address of the account owner
- `credential_number`: The unique identifier of the credential to delete

**Validation:**
- `sender` must be a valid bech32 address
- `credential_number` must exist for the sender's account
- Account must have at least 2 credentials (cannot delete the last credential)
- Sender must own the account containing the credential

**State Changes:**
- Removes the specified credential from the account's credential list
- Updates the ProvenanceAccount in storage

**Events Emitted:**
- `EventCredentialDelete` with account address and credential number

#### Response

```protobuf
message MsgDeleteCredentialResponse {
  uint64 credential_number = 1;
}
```

**Fields:**
- `credential_number`: The credential number that was deleted

### MsgUpdateParams

Updates the module parameters (governance only).

```protobuf
message MsgUpdateParams {
  string authority = 1;
  Params params = 2;
}
```

**Fields:**
- `authority`: The address of the governance account (must match the module's authority)
- `params`: New parameter values

**Validation:**
- `authority` must match the module's configured authority address
- `params` must be valid (max_credential_allowed > 0)

**State Changes:**
- Updates the module parameters in state

**Events Emitted:**
- Standard governance events (not module-specific)

#### Response

```protobuf
message MsgUpdateParamsResponse {}
```

Empty response message.

## Message Processing Flow

### Registration Flow

1. **Message Validation**: Basic field validation and authorization checks
2. **Account Resolution**: Determine if account exists or needs to be created
3. **Credential Processing**: Parse and validate the credential data
4. **Uniqueness Check**: Ensure credential is not already registered
5. **Limit Check**: Verify maximum credential limit is not exceeded
6. **State Update**: Add credential to account and update storage
7. **Event Emission**: Emit appropriate events for indexing

### Deletion Flow

1. **Message Validation**: Basic field validation and authorization checks
2. **Account Lookup**: Retrieve the smart account from storage
3. **Credential Existence**: Verify the credential exists for this account
4. **Safety Check**: Ensure account will have remaining credentials
5. **State Update**: Remove credential from account and update storage
6. **Event Emission**: Emit deletion event

## Error Handling

### Common Error Cases

**ErrSmartAccountsDisabled**
- Code: 2
- Description: Smart accounts module is disabled by governance
- Resolution: Enable module through governance proposal

**ErrMaxCredentialsReached**
- Code: 3
- Description: Account has reached maximum allowed credentials
- Resolution: Delete unused credentials or request parameter update

**ErrCredentialNotFound**
- Code: 4
- Description: Specified credential does not exist
- Resolution: Verify credential number and account ownership

**ErrInvalidCredential**
- Code: 5
- Description: Credential data is malformed or invalid
- Resolution: Check credential format and regenerate if necessary

**ErrCannotDeleteLastCredential**
- Code: 6
- Description: Attempt to delete the only remaining credential
- Resolution: Add another credential before deleting the last one

**ErrDuplicateCredential**
- Code: 7
- Description: Credential already exists for this account
- Resolution: Use existing credential or delete before re-registering

## Authentication During Message Processing

### Signature Requirements

All messages require proper authentication:

**Traditional Accounts:**
- Standard secp256k1 signature verification
- Must sign the transaction with the account's private key

**Smart Accounts:**
- Can use any registered credential for authentication
- FIDO2 credentials use WebAuthn assertion format
- Traditional credentials use standard ECDSA signatures

### Message Authorization

Each message type has specific authorization requirements:

**Credential Registration:**
- Account owner must sign the transaction
- If account doesn't exist, creates new smart account
- Must have permission to modify the account

**Credential Deletion:**
- Account owner must sign the transaction
- Must own the credential being deleted
- Cannot delete if it would leave account with no credentials

**Parameter Updates:**
- Only governance authority can update parameters
- Requires governance proposal and voting process

## Gas Costs

Message gas costs are determined by:

**Base Message Cost:**
- Standard Cosmos SDK message processing overhead
- Signature verification costs

**Credential-Specific Costs:**
- FIDO2 registration: Higher cost due to attestation verification
- Traditional key registration: Standard public key validation cost
- Credential deletion: Minimal cost for storage update

**Storage Costs:**
- Account storage updates
- Credential metadata storage
- Event emission costs

## Best Practices

### For Client Applications

1. **Error Handling**: Implement comprehensive error handling for all error types
2. **Retry Logic**: Add appropriate retry mechanisms for transient failures
3. **User Feedback**: Provide clear feedback about credential registration status
4. **Validation**: Perform client-side validation before submitting messages

### For Credential Management

1. **Backup Credentials**: Always maintain multiple credentials per account
2. **Credential Rotation**: Regularly update credentials for security
3. **Testing**: Test credential functionality after registration
4. **Documentation**: Keep track of registered credentials and their purposes

### For Integration

1. **Module Checks**: Verify module is enabled before attempting operations
2. **Parameter Awareness**: Check current parameter limits before registration
3. **Event Monitoring**: Monitor events for successful credential operations
4. **Graceful Degradation**: Handle disabled module state appropriately