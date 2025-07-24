# `x/smartaccounts`

By Default smart accounts are disabled for now, you can turn them on by governance
```go
// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		// enabled by default
		Enabled:              false,
		MaxCredentialAllowed: 10, // Set default max credentials per account
	}
}

```

[example proposal to turn it on](https://github.com/arnabmitra/go-provenance-client/blob/ce59dea12b41aa2f898e76ee8d6c2b95fc2f399e/sa_testing_tools/proposal_sa.json#L10)



[This can be used for testing via an UI](https://github.com/arnabmitra/webauthn_proxy)


## Overview

The `x/smartaccounts` module introduces a new type of account that extends the authentication capabilities of a base account, allowing for more flexible and user-friendly authentication methods beyond a single private key.

## Specifications

1. **[Concepts](01_concepts.md)**
