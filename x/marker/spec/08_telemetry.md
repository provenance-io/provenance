# Telemetry

The marker module exposes a limited set of telemetry for monitoring is operations.  

> NOTE: The majority of the telemetry that applies to the marker module is exposed by the `bank` module and the `auth` 
> module which the marker module uses to perform most of its functions.

## Transferred Amount

For transfers of restricted coins the amount moved and the associated denom are published.

| Labels                  | Value          |
| ----------------------- | -------------- |
| `tx`, `msg`, `transfer` | amount `int64` |
| `denom`                 | marker denom   |