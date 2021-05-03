# Telemetry

The marker module exposes a limited set of telemetry for monitoring is operations.

## Transferred Amount

For transfers of restricted coins the amount moved and the associated denom are published.

| Labels                  | Value          |
| ----------------------- | -------------- |
| `tx`, `msg`, `transfer` | amount `int64` |
| `denom`                 | marker denom   |