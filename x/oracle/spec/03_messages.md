<!--
order: 3
-->

# Messages

In this section we describe the processing of the oracle messages and their corresponding updates to the state.

<!-- TOC 2 -->
  - [Msg/UpdateOracle](#msgupdateoracle)
  - [Msg/SendQueryOracle](#msgsendqueryoracle)


---
## Msg/UpdateOracle

The oracle's address is modified by proposing the `MsgUpdateOracleRequest` message.

### Request

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/oracle/v1/tx.proto#L40-L49

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/oracle/v1/tx.proto#L51-L52

The message will fail under the following conditions:
* The authority does not match the gov module.
* The new address does not pass basic integrity and format checks.

## Msg/SendQueryOracle

Sends a query to another chain's `Oracle` using `ICQ`.

### Request

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/oracle/v1/tx.proto#L22-L32

### Response

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/oracle/v1/tx.proto#L34-L38

The message will fail under the following conditions:
* The authority does not pass basic integrity and format checks.
* The query does not have the correct format.
* The channel is invalid or does not pass basic integrity and format checks.
