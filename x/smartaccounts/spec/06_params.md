<!--
order: 6
-->

# Parameters

The smart accounts module has configurable parameters that control its behavior and limits. These parameters can only be modified through governance proposals.

## Parameter Types

### Params

The module uses a single `Params` structure to store all configuration:

```protobuf
message Params {
  bool enabled = 1;
  uint32 max_credential_allowed = 2;
}
```

## Parameter Details

### enabled

**Type:** `bool`  
**Default:** `false`  
**Description:** Controls whether the smart accounts module is active and accepting new operations.

**Behavior:**
- When `false`: All smart account operations are disabled
  - New credential registrations are rejected
  - Existing smart accounts cannot add/remove credentials
  - Authentication still works for existing smart accounts
  - Parameter updates are still allowed (governance only)
- When `true`: All smart account functionality is available

**Use Cases:**
- **Gradual Rollout**: Enable the module after thorough testing
- **Emergency Disable**: Temporarily disable new registrations if issues are discovered
- **Maintenance Mode**: Disable operations during upgrades or maintenance

**Governance Example:**
```json
{
  "title": "Enable Smart Accounts Module",
  "description": "Enable the smart accounts module to allow FIDO2 and multi-credential authentication",
  "changes": [
    {
      "subspace": "smartaccounts",
      "key": "enabled",
      "value": "true"
    }
  ]
}
```

### max_credential_allowed

**Type:** `uint32`  
**Default:** `10`  
**Description:** Maximum number of credentials that can be registered per smart account.

**Constraints:**
- Must be greater than 0
- Reasonable upper limit to prevent state bloat
- Should account for typical user device scenarios

**Considerations:**
- **Storage Impact**: Each credential adds to account storage size
- **User Experience**: Users may have multiple devices (phone, laptop, hardware keys)
- **Security**: Multiple credentials provide backup authentication methods
- **Gas Costs**: More credentials increase transaction processing costs

**Typical Values:**
- **Conservative**: 5 credentials (basic multi-device support)
- **Standard**: 10 credentials (recommended default)
- **Liberal**: 20+ credentials (enterprise/institutional use)

**Governance Example:**
```json
{
  "title": "Increase Smart Account Credential Limit",
  "description": "Increase the maximum credentials per account from 10 to 15 to support enterprise users",
  "changes": [
    {
      "subspace": "smartaccounts",
      "key": "max_credential_allowed",
      "value": "15"
    }
  ]
}
```

## Default Parameters

```go
func DefaultParams() Params {
    return Params{
        Enabled:              false, // Disabled by default for safety
        MaxCredentialAllowed: 10,    // Reasonable limit for most users
    }
}
```

## Parameter Validation

### Runtime Validation

Parameters are validated when updated:

```go
func (p Params) Validate() error {
    if p.MaxCredentialAllowed == 0 {
        return fmt.Errorf("max_credential_allowed must be greater than 0")
    }
    return nil
}
```

### State Validation

The module enforces parameter constraints during operation:
- Credential registration checks against `max_credential_allowed`
- All operations check `enabled` status
- Invalid parameters prevent module initialization

## Parameter Updates

### Governance Process

Parameters can only be changed through governance:

1. **Proposal Submission**: Community member submits parameter change proposal
2. **Voting Period**: Token holders vote on the proposal
3. **Execution**: If passed, parameters are automatically updated
4. **Effect**: Changes take effect immediately upon proposal execution

### Update Message

Parameter updates use the standard governance parameter change process:

```protobuf
message MsgUpdateParams {
  string authority = 1;  // Governance module address
  Params params = 2;     // New parameter values
}
```

### CLI Commands

**Submit Parameter Change Proposal:**
```bash
provenanced tx gov submit-proposal param-change proposal.json \
  --from validator \
  --chain-id provenance-1 \
  --gas auto \
  --gas-adjustment 1.5
```

**Query Current Parameters:**
```bash
provenanced query smartaccounts params
```

**Query Specific Parameter:**
```bash
provenanced query params subspace smartaccounts enabled
provenanced query params subspace smartaccounts max_credential_allowed
```

## Parameter Migration

### Upgrade Considerations

When upgrading the module:
- Parameters persist across upgrades
- New parameters get default values
- Deprecated parameters are ignored
- Validation ensures consistency

### Backward Compatibility

Parameter changes should maintain backward compatibility:
- Increasing limits is generally safe
- Decreasing limits may affect existing accounts
- Disabling the module preserves existing functionality

## Security Implications

### enabled Parameter

**Security Considerations:**
- Disabling prevents new attack vectors from being introduced
- Does not affect existing smart account security
- Can be used as emergency brake if vulnerabilities are discovered

**Risk Mitigation:**
- Monitor for unusual activity when enabling module
- Have governance process ready for emergency disable
- Test thoroughly before enabling on mainnet

### max_credential_allowed Parameter

**Security Considerations:**
- Higher limits increase potential attack surface per account
- More credentials mean more potential compromise vectors
- Storage and gas costs scale with credential count

**Risk Mitigation:**
- Set reasonable limits based on expected use cases
- Monitor storage growth and gas usage patterns
- Consider gradual increases rather than large jumps

## Monitoring Parameters

### Key Metrics

**Module Enablement:**
- Track when module is enabled/disabled
- Monitor credential registration rates after enabling
- Alert on unexpected parameter changes

**Credential Limits:**
- Monitor accounts approaching credential limits
- Track average credentials per account
- Identify accounts with maximum credentials

### Alerting

**Governance Alerts:**
- Parameter change proposals submitted
- Unusual parameter values proposed
- Emergency parameter changes

**Usage Alerts:**
- High credential registration rate after enabling
- Accounts hitting credential limits
- Unusual authentication patterns

## Best Practices

### For Governance

1. **Gradual Rollout**: Enable module on testnets first
2. **Conservative Limits**: Start with lower credential limits
3. **Monitor Impact**: Watch metrics after parameter changes
4. **Emergency Planning**: Have disable procedures ready

### For Developers

1. **Parameter Awareness**: Check current parameters before operations
2. **Graceful Handling**: Handle disabled module state appropriately
3. **User Communication**: Inform users of parameter limits
4. **Future Proofing**: Design for potential parameter changes

### For Users

1. **Limit Planning**: Understand credential limits before registering many devices
2. **Backup Strategy**: Don't rely on maximum credentials for backup
3. **Governance Participation**: Vote on parameter proposals that affect usage
4. **Stay Informed**: Monitor governance proposals for parameter changes

## Testing Parameters

### Unit Tests

```go
func TestParameterValidation(t *testing.T) {
    // Test valid parameters
    validParams := types.Params{
        Enabled:              true,
        MaxCredentialAllowed: 10,
    }
    require.NoError(t, validParams.Validate())
    
    // Test invalid parameters
    invalidParams := types.Params{
        Enabled:              true,
        MaxCredentialAllowed: 0, // Invalid
    }
    require.Error(t, invalidParams.Validate())
}

func TestParameterEnforcement(t *testing.T) {
    app := setupTestApp()
    ctx := app.BaseApp.NewContext(false, tmproto.Header{})
    
    // Set low credential limit
    params := types.Params{
        Enabled:              true,
        MaxCredentialAllowed: 1,
    }
    app.SmartAccountsKeeper.SetParams(ctx, params)
    
    // Test limit enforcement
    account := createTestAccount(ctx, app)
    
    // First credential should succeed
    err := registerCredential(ctx, app, account, "cred1")
    require.NoError(t, err)
    
    // Second credential should fail
    err = registerCredential(ctx, app, account, "cred2")
    require.Error(t, err)
    require.Contains(t, err.Error(), "max credentials reached")
}
```

### Integration Tests

```bash
#!/bin/bash

# Test parameter queries
provenanced query smartaccounts params --output json

# Test parameter enforcement
# Try to register credential when module is disabled
provenanced tx smartaccounts register-cosmos \
  --pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"..."}' \
  --from alice \
  --chain-id testing \
  --yes

# Should fail with module disabled error
```

## Future Parameters

### Planned Additions

**Session Parameters:**
- `session_timeout_blocks`: Default session duration in blocks
- `max_session_duration`: Maximum allowed session duration

**Security Parameters:**
- `require_user_verification`: Force UV flag for WebAuthn credentials
- `allowed_attestation_formats`: Restrict WebAuthn attestation formats

**Rate Limiting:**
- `credential_registration_rate_limit`: Limit registrations per block
- `credential_deletion_cooldown`: Minimum time between deletions

### Deprecation Policy

When parameters become obsolete:
1. Mark as deprecated in documentation
2. Continue accepting but ignore the parameter
3. Remove from new versions after sufficient notice
4. Provide migration guide for dependent applications