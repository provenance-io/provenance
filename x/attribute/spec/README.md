# `Attribute`

## Abstract

The purpose of the Attributes Module is to act as a registry that allows an Address to store <Name, Value> pairs.
Every Name must be registered by the Name Service, and a Name have multiple Values associated with it. Values are required to have a type, and they can be set or retrieved by Name.

This feature provides the blockchain with the capability to store and retrieve values by Name. It plays a major
part in some of our components such as smart contract creation process. It allows an Address to create and store 
a named smart contract on the blockchain.

## Contents

1. **[State](01_state.md)**
1. **[Messages](02_messages.md)**
1. **[Events](03_events.md)**
1. **[Params](04_params.md)**