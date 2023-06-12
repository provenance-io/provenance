<!--
order: 7
-->

# Trigger Genesis

In this section we describe the processing of the trigger messages and the corresponding updates to the state.


## Msg/GenesisState

GenesisState contains a list of triggers, queued triggers, and gas limits. It also tracks the triggerID and the queue start. These are exported and later imported from/to the store.

+++ https://github.com/provenance-io/provenance/blob/f560c43f9e0e8079e3b62b4e8fc8411baee5590c/proto/provenance/trigger/v1/genesis.proto#L12-L30
