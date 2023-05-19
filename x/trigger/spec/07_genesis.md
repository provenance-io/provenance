<!--
order: 7
-->

# Trigger Genesis

In this section we describe the processing of the trigger messages and the corresponding updates to the state.


## Msg/GenesisState
GenesisState contains a list of triggers, queued triggers, and gas limits. It also tracks the triggerID and the queue start. These are exported and later imported from/to the store.

+++ https://github.com/provenance-io/provenance/blob/ac5d53a1b4f14b7fc9d0d13630ea5262d61b93c0/proto/provenance/trigger/v1/genesis.proto#L12-L30
