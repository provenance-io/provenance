<!--
order: 8
-->

# Reward Genesis

In this section we describe the processing of the reward messages and the corresponding updates to the state.


## Msg/GenesisState
GenesisState contains a list of reward programs, claim period reward distributions, and reward account states. These are exported and later imported from/to the store.

+++ https://github.com/provenance-io/provenance/blob/ccaef3a7024f0ccd73d175465e91577373127858/proto/provenance/reward/v1/genesis.proto#L13-L22
