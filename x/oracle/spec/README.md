# `x/oracle`

## Overview

The oracle module provides the Provenance Blockchain with the capability to dynamically expose query endpoints through Interchain Queries (ICQ)

One challenge that the Provenance Blockchain faces is supporting each Provenance Blockchain Zone with a unique set of queries. It is not feasible to create an evolving set of queries for each chain. Furthermore, it is not desirable for other parties to request Provenance to build these endpoints for them and then upgrade. This module resolves these issues by enabling Provenance Blockchain zones to manage their own oracle.

## Acknowledgements

We appreciate the substantial contributions made by Strangelove Ventures and Quasar Finance through their work on the [Async ICQ Module](https://github.com/cosmos/ibc-apps/tree/main/modules/async-icq) and [Interchain Query Demo](https://github.com/quasar-finance/interchain-query-demo). These resources were of paramount importance in informing the development of our oracle module.

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Queries](04_queries.md)**
5. **[Events](05_events.md)**
6. **[Genesis](06_genesis.md)**