<!--
order: 7
-->

# Trigger Genesis

In this section we describe the processing of the trigger messages and the corresponding updates to the state.


## GenesisState

GenesisState contains a list of triggers, and queued triggers. It also tracks the triggerID and the queue start. The `gas_limits` field has been deprecated and is no longer used; as such, it must always be empty.

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/trigger/v1/genesis.proto#L11-L30
