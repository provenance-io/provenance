<!--
order: 1
-->

# Concepts

The oracle module is very minimal, but users should understand what the `Oracle` is and how it interacts with `ICQ`.

<!-- TOC 2 -->
  - [Oracle](#oracle)
  - [Interchain Queries (ICQ)](#interchain-queries-icq)


---
## Oracle

The `Oracle` is a custom built CosmWasm smart contract that the chain queries for data. Chain users can update the address with a proposal.

## Interchain Queries (ICQ)

`ICQ` is heavily leveraged in order to allow one Provenance Blockcahin to query another Provenance Blockchain's `Oracle`. This module acts as both the `Controller` and receiver of the `Host` in the `ICQ` realm.

When a user intends to query another chain, they initiate the process by submitting a query through a transaction on the `ICQ Controller`. This `Controller` delivers the query from the transaction to the `ICQ Host` module of the destination chain via `IBC`. Subsequently, the received query is routed by the `ICQ Host` to this module. Upon receipt, the module queries the `Oracle` using the provided input, and the resulting information is then transmitted back to the `ICQ Controller` in the form of an `ACK` message.

It should be noted that responses, which arrive in the form of the `ACK`, indicate that queries operate asynchronously. Consequently, these results will not be immediately accessible, requiring the user to wait for an emitted event on the response. For additional details, you can refer to the [Async ICQ Module](https://github.com/cosmos/ibc-apps/tree/main/modules/async-icq) developed by strangelove-ventures.

### Note

For `ICQ` to function correctly, it is essential to establish an `unordered channel` connecting the two chains. This channel should be configured utilizing the `oracle` and `icqhost` ports on the `ICQ Controller` and `ICQ Host` correspondingly. The `version` should be designated as `icq-1`. Moreover, it is crucial to ensure that the `HostEnabled` parameter is enabled with a value of `true`, while the `AllowQueries` parameter should encompass the path `"/provenance.oracle.v1.Query/Oracle"`.
