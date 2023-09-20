<!--
order: 2
-->

# State

The oracle module manages the address of the Oracle and the ICQ state.

<!-- TOC 2 -->
  - [Oracle](#oracle)
  - [IBC](#ibc)


---
## Oracle

The `Oracle` is a CosmWasm smart contract that the module forwards its queries to and relays responses from. Users can manipulate this state by submitting a update oracle proposal.

* Oracle `0x01 -> []byte{}`

---
## IBC

`IBC` communication exists between the `oracle` and `icqhost` modules. The `oracle` module tracks its channel's `port` in state.

* Port `0x02 -> []byte{}`
