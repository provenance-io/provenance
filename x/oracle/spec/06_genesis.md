<!--
order: 6
-->

# Oracle Genesis

In this section we describe the processing of the oracle messages and the corresponding updates to the state.

<!-- TOC 2 -->
- [Oracle Genesis](#oracle-genesis)
  - [Msg/GenesisState](#msggenesisstate)


---
## Msg/GenesisState

The GenesisState encompasses the upcoming sequence ID for an ICQ packet, the associated parameters, the designated port ID for the module, and the oracle address. These values are both extracted for export and imported for storage within the store.

+++ https://github.com/provenance-io/provenance/blob/65865991f93e2c1a7647e29be11f6527f49616e6/proto/provenance/oracle/v1/genesis.proto#L11-L22
