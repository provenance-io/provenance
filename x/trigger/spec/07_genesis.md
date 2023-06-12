<!--
order: 7
-->

# Trigger Genesis

In this section we describe the processing of the trigger messages and the corresponding updates to the state.


## Msg/GenesisState

GenesisState contains a list of triggers, queued triggers, and gas limits. It also tracks the triggerID and the queue start. These are exported and later imported from/to the store.

+++ https://github.com/provenance-io/provenance/blob/bda28e5f58a4a58e8fef21141400ad362b84518b/proto/provenance/trigger/v1/genesis.proto#L11-L30
