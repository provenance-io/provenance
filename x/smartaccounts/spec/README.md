# `x/smartaccounts`

## Abstract

This document specifies the smart accounts module of the Provenance blockchain.

The smart accounts module extends the authentication capabilities of standard Cosmos SDK accounts by supporting multiple authentication methods beyond traditional private key signatures. Smart accounts enable users to authenticate transactions using modern web standards like FIDO2/WebAuthn (Face ID, Touch ID, Passkeys) while maintaining compatibility with traditional secp256k1 key-based authentication.

**Current Status:** By default, smart accounts are disabled and must be enabled through governance before use.

## Context

Traditional externally owned accounts (EOAs) in blockchain systems have several limitations:

- **Single point of failure**: One private key controls the entire account
- **Poor user experience**: Complex key management and recovery processes
- **Limited device support**: Manual transfer of recovery phrases between devices
- **Security risks**: No key rotation capabilities and exposure to phishing attacks
- **High barrier to entry**: Users must learn complex security practices

The smart accounts module addresses these limitations by providing a more flexible and user-friendly authentication system that bridges the gap between Web2 user experience and Web3 security requirements.

## Overview

Smart accounts are enhanced Cosmos SDK accounts that support multiple authentication credentials. Each smart account extends a base account with additional authentication methods, allowing users to:

- Authenticate using FIDO2/WebAuthn standards (biometrics, hardware keys, passkeys)
- Maintain traditional secp256k1 key-based authentication
- Rotate and manage multiple credentials
- Control authentication methods per account

By default, smart accounts are disabled and must be enabled through governance. Once enabled, users can register multiple credentials and authenticate transactions using any of their registered methods.

## Module Status

The smart accounts module is currently **disabled by default** for security and stability reasons. To enable the module:

1. **Governance Proposal Required**: The module must be activated through a governance proposal
2. **Parameter Configuration**: Key parameters include:
   - `enabled`: Controls module activation (default: `false`)
   - `max_credential_allowed`: Maximum credentials per account (default: `10`)
3. **Gradual Rollout**: Recommended to enable on testnets before mainnet activation

```bash
# Check current module status
provenanced query smartaccounts params

# Expected output when disabled:
# params:
#   enabled: false
#   max_credential_allowed: 10
```

## Contents

1. **[Concepts](01_concepts.md)** - Core concepts, architecture, and benefits
2. **[State](02_state.md)** - On-chain data structures and storage
3. **[Messages](03_messages.md)** - Transaction types and message handling
4. **[Queries](04_queries.md)** - Query endpoints and data retrieval
5. **[Events](05_events.md)** - Blockchain events and monitoring
6. **[Parameters](06_params.md)** - Module configuration and governance
7. **[Authentication Flow](07_authentication.md)** - Transaction signing and verification
8. **[Client Usage](08_client_usage.md)** - WebAuthn client implementation examples
9. **[Integration Guide](09_integration.md)** - Complete integration guide for developers

## Quick Links

- **[Enable Smart Accounts](https://github.com/arnabmitra/go-provenance-client/blob/ce59dea12b41aa2f898e76ee8d6c2b95fc2f399e/sa_testing_tools/proposal_sa.json#L10)** - Example governance proposal
- **[WebAuthn Testing UI](https://github.com/arnabmitra/webauthn_proxy)** - Testing interface for WebAuthn integration
- **[Proto Definitions](../../../proto/provenance/smartaccounts/v1/)** - Protocol buffer definitions