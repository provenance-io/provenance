# Expiration Messages

In this section we describe the processing of the expiration messages and the corresponding updates to the state.
All created/modified state objects specified by each message are defined within the [state](02_state.md) section.

<!-- TOC -->
  - [Entries](#entries)
    - [Msg/AddExpirationRequest](#msg-addexpirationrequest)
    - [Msg/ExtendExpirationRequest](#msg-extendexpirationrequest)
    - [Msg/InvokeExpirationRequest](#msg-invokeexpirationrequest)
  - [Authz Grants](#authz-grants)

---
## Entries

### Msg/AddExpirationRequest

An expiration is created using the `AddExpiration` service method.

Expirations are identified using their `module_asset_id`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/eb569b71b4d9137272432df5968cd62bf1eca2fb/proto/provenance/expiration/v1/tx.proto#L22-L32

#### Response

+++ https://github.com/provenance-io/provenance/blob/eb569b71b4d9137272432df5968cd62bf1eca2fb/proto/provenance/expiration/v1/tx.proto#L34-L35

#### Expected failures

This service message is expected to fail if:
* The `module_asset_id` is missing or invalid.
* The `owner` is missing or is not a bech32 address string.
* The expiration `time` is missing or invalid.
* The `deposit` is missing or is an invalid coin type.
* The `message` is missing, or does not implement `sdk.Msg` interface.

### Msg/ExtendExpirationRequest

An expiration is extended using the `ExtendExpiration` service method.

Expirations are identified using their `module_asset_id`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/eb569b71b4d9137272432df5968cd62bf1eca2fb/proto/provenance/expiration/v1/tx.proto#L37-L49

Accepted values for a `duration` are n{h,d,w,y} where 1h = 60m, 1d = 24h, 1w = 7d (or 168h), 1y = 365d (or 8760h)

#### Response 

+++ https://github.com/provenance-io/provenance/blob/eb569b71b4d9137272432df5968cd62bf1eca2fb/proto/provenance/expiration/v1/tx.proto#L51-L52

#### Expected failures

This service is expected to fail if:
* The `module_asset_id` is missing or invalid.
* The `duration` is missing or is invalid.

### Msg/InvokeExpirationRequest

An expiration is invoked using the `InvokeExpiration` service method.

Expirations are identified using their `module_asset_id`.

#### Request

+++ https://github.com/provenance-io/provenance/blob/eb569b71b4d9137272432df5968cd62bf1eca2fb/proto/provenance/expiration/v1/tx.proto#L54-L64

#### Response

+++ https://github.com/provenance-io/provenance/blob/eb569b71b4d9137272432df5968cd62bf1eca2fb/proto/provenance/expiration/v1/tx.proto#L66-L67

#### Expected failures

This service is expected to fail if:
* The `module_asset_id` is missing or invalid.

---
## Authz Grants

Authz requires the use of fully qualified message type URLs when applying grants to an expiration. See [04_authz.md](04_authz.md) for more details.

Fully qualified `expiration` message type URLs:
- `/provenance.expiration.v1.MsgAddExpirationRequest`
- `/provenance.expiration.v1.MsgExtendExpirationRequest`
- `/provenance.expiration.v1.MsgInvokeExpirationRequest`
