<!--
order: 2
-->

# State

The oracle module manages the address of the Oracle and the ICQ state.

## Oracle

The `Oracle` is a CosmWasm smart contract that the module forwards its queries to and relays responses from. Users can manipulate this state by submitting a update oracle proposal.

* Oracle `0x01 -> []byte{}`

---
## IBC

`IBC` communication exists between the `oracle` and `icqhost` modules. The `oracle` module tracks its channel's `port` in state. It also ensures that each `ICQ` packet has a unique `sequence number` by trackinig this in state.

* Port `0x03 -> []byte{}`
* Last Query Sequence Number `0x02 -> uint64`

---
## ICQ Responses

The asynchronous nature of the `ICQ` implementation necessitates the module to retain both the request and the corresponding response. This enables users to monitor the progress of their requests and subsequently review the received responses.

* QueryRequest `0x04 | Sequence ID (8 Bytes) -> ProtocolBuffers(QueryOracleRequest)`
* QueryResponse `0x05 | Sequence ID (8 Bytes) -> ProtocolBuffers(QueryOracleResponse)`

+++ https://github.com/provenance-io/provenance/blob/a10304ad4fe1ddb20bc4f54a413942b8898ce887/proto/provenance/oracle/v1/query.proto#L51C1-L61