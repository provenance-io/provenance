# `x/ledger`

## Abstract

This document specifies the ledger module of the Provenance blockchain.

The ledger module provides comprehensive financial tracking capabilities for NFTs and metadata scopes, enabling detailed accounting of transactions, disbursements, payments, and fees associated with asset ownership. It maintains chronological records of all financial activities and their impact on principal, interest, and other balances.

## Context

Financial tracking on blockchain requires the ability to maintain detailed accounting records for non-fungible assets with complex ownership structures. Each asset requires rules governing financial transactions, balance tracking, and historical record keeping. Examples include loan servicing, asset-backed securities, and fractional ownership tracking. The rules governing financial activities must be enforced by the blockchain itself to ensure transparency, auditability, and trust in the financial data.

## Overview

The ledger module provides various tools for defining and managing financial tracking for assets. Ledgers can be created and managed by authorized maintainers through normal Msg requests. A ledger can track multiple types of financial activities including disbursements, payments, fees, and adjustments. The module supports configurable entry types, status types, and balance buckets to accommodate various financial tracking requirements.

## Contents

1. **[Concepts](01_concepts.md)**
    - [Ledger Classes](01_concepts.md#ledger-class)
    - [Ledgers](01_concepts.md#ledger)
    - [Ledger Entries](01_concepts.md#ledger-entry)
    - [Balance Buckets](01_concepts.md#balance-buckets)
    - [Entry Types](01_concepts.md#entry-types)
    - [Status Types](01_concepts.md#status-types)
2. **[State](02_state.md)**
    - [Ledger Classes](02_state.md#ledger-class)
    - [Ledgers](02_state.md#ledger)
    - [Ledger Entries](02_state.md#ledger-entries)
    - [Balances](02_state.md#balances)
    - [State Storage](02_state.md#state-storage)
3. **[Events](03_events.md)**
    - [Ledger Events](03_events.md#ledger-events)
    - [Entry Events](03_events.md#entry-events)
    - [Transfer Events](03_events.md#transfer-events)
4. **[Queries](04_queries.md)**
    - [Ledger Queries](04_queries.md#ledger-queries)
    - [Entry Queries](04_queries.md#entry-queries)
    - [Balance Queries](04_queries.md#balance-queries)
5. **[Messages](05_messages.md)**
    - [Ledger Management](05_messages.md#ledger-management)
    - [Entry Management](05_messages.md#entry-management)
    - [Class Management](05_messages.md#class-management)
    - [Transfer Management](05_messages.md#transfer-management)
6. **[Bulk Import](06_bulk_import.md)** - Bulk import functionality and usage

## Integration

The Ledger module can be integrated with:
- NFT marketplaces and exchanges
- Financial tracking and accounting systems
- Loan servicing platforms
- Asset-backed security systems
- Audit and compliance systems
- External monitoring and reporting tools
- Payment processing systems 