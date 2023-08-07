<!--
order: 3
-->

# Messages

In this section we describe the processing of the oracle messages and their corresponding updates to the state.

<!-- TOC 2 -->
  - [Msg/UpdateOracleRequest](#msgupdateoraclerequest)
  - [Msg/SendQueryOracleRequest](#msgsendqueryoraclerequest)


---
## Msg/UpdateOracleRequest

The oracle's address is modified by proposing the `MsgUpdateOracleRequest` message.

### Request

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/tx.proto#L37-L46

### Response

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/tx.proto#L48-L49

The message will fail under the following conditions:
* The authority does not match the gov module.
* The new address does not pass basic integrity and format checks.

## Msg/SendQueryOracleRequest

Sends a query to another chain's `Oracle` using `ICQ`.

### Request

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/tx.proto#L21-L29

### Response

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/tx.proto#L31-L35

The message will fail under the following conditions:
* The authority does not pass basic integrity and format checks.
* The query does not have the correct format.
* The channel is invalid or does not pass basic integrity and format checks.
