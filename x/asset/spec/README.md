# `x/asset`

## Overview

The Asset module provides functionality for creating and managing digital assets on the Provenance blockchain. It leverages the Cosmos SDK's NFT module to represent assets as non-fungible tokens while adding additional features like asset classes, pools, tokenizations, and securitizations.

## Key Features

- **Asset Classes**: Define schemas and classifications for digital assets with JSON schema validation
- **Assets**: Create individual digital assets within asset classes with data validation
- **Pools**: Bundle multiple NFTs into tradeable marker tokens
- **Tokenizations**: Fractionalize individual NFTs into tradeable tokens
- **Securitizations**: Create structured financial products with multiple tranches
- **Integration**: Seamless integration with Provenance's ledger, registry, and marker modules

## Contents

1. **[Concepts](01_concepts.md)** - Core concepts and terminology
2. **[State](02_state.md)** - State management and storage
3. **[Events](03_events.md)** - Event types and attributes
4. **[Messages](04_messages.md)** - Message types and validation
5. **[Queries](05_queries.md)** - Query endpoints and responses
6. **[CLI Examples](06_cli_examples.md)** - Command-line interface usage examples 