# `Attribute`

## Abstract

The purpose of the Attributes Module is to act as a registry that allows an Address to store <Name, Value> pairs.
Every Name must be registered by the Name Service, and an Address can have duplicate Names/Keys. Values are required
to have a type, and they can be set or retrieved with a Name.

This feature provides the blockchain with the capability to store and retrieve values by name. It plays a major
part in the creation of smart contracts. An Address can create and store a named contract on the blockchain.

## Contents

1. **[State](01_state.md)**
1. **[State_transitions](02_state_transitions.md)**
1. **[Messages](03_messages.md)**
1. **[Begin Block](04_begin_block.md)**
1. **[End Block](05_end_block.md)**
1. **[Hooks](06_hooks.md)**
1. **[Events](07_events.md)**
1. **[Telemetry](08_telemetry.md)**
1. **[Params](09_params.md)**
1. **[Governance](10_governance.md)**