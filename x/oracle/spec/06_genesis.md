<!--
order: 6
-->

# Oracle Genesis

In this section we describe the processing of the oracle messages and the corresponding updates to the state.

<!-- TOC 2 -->
  - [Msg/GenesisState](#msggenesisstate)


---
## Msg/GenesisState

The GenesisState encompasses the upcoming sequence ID for an ICQ packet, the associated parameters, the designated port ID for the module, and the oracle address. These values are both extracted for export and imported for storage within the store.

+++ https://github.com/provenance-io/provenance/blob/ba0b65c54f61f99c951fe4694271847dbad0fb00/proto/provenance/oracle/v1/genesis.proto#L11-L24
