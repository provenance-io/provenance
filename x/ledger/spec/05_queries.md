# Ledger Queries

The Ledger module provides several query endpoints to access ledger state and information.

---
<!-- TOC 2 2 -->
  - [LedgerClass](#ledgerclass)
  - [LedgerClasses](#ledgerclasses)
  - [LedgerClassEntryTypes](#ledgerclassentrytypes)
  - [LedgerClassStatusTypes](#ledgerclassstatustypes)
  - [LedgerClassBucketTypes](#ledgerclassbuckettypes)
  - [Ledger](#ledger)
  - [Ledgers](#ledgers)
  - [LedgerEntries](#ledgerentries)
  - [LedgerEntry](#ledgerentry)
  - [LedgerBalancesAsOf](#ledgerbalancesasof)
  - [LedgerSettlements](#ledgersettlements)
  - [LedgerSettlementsByCorrelationID](#ledgersettlementsbycorrelationid)

## LedgerClass

Use the `LedgerClass` query to look up a specific ledger class.

### QueryLedgerClassRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L81-L84

### QueryLedgerClassResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L86-L89

See also: [LedgerClass](03_messages.md#ledgerclass)


## LedgerClasses

To get all ledger classes, use the `LedgerClasses` query.

This query is paginated.

### QueryLedgerClassesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L91-L95

### QueryLedgerClassesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L97-L104

See also: [LedgerClass](03_messages.md#ledgerclass)


## LedgerClassEntryTypes

To get all ledger class entry types for a ledger class, use the `LedgerClassEntryTypes` query.

### QueryLedgerClassEntryTypesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L106-L109

### QueryLedgerClassEntryTypesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L111-L114

See also: [LedgerClassEntryType](03_messages.md#ledgerclassentrytype)


## LedgerClassStatusTypes

To get all ledger class status types for a ledger class, use the `LedgerClassEntryTypes` query.

### QueryLedgerClassStatusTypesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L116-L119

### QueryLedgerClassStatusTypesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L121-L124

See also: [LedgerClassStatusType](03_messages.md#ledgerclassstatustype)


## LedgerClassBucketTypes

The bucket types for a ledger class can be looked up using the `LedgerClassBucketTypes` query.

### QueryLedgerClassBucketTypesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L126-L129

### QueryLedgerClassBucketTypesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L131-L134

See also: [LedgerClassBucketType](03_messages.md#ledgerclassbuckettype)


## Ledger

To look up a specific ledger, use the `Ledger` query.

### QueryLedgerRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L136-L140

See also: [LedgerKey](03_messages.md#ledgerkey)

### QueryLedgerResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L142-L146

See also: [Ledger](03_messages.md#ledger)


## Ledgers

To get all ledgers, use the `Ledgers` query.

This query is paginated.

### QueryLedgersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L148-L152

### QueryLedgersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L154-L161

See also: [Ledger](03_messages.md#ledger)


## LedgerEntries

A ledger's entries are looked up using this `LedgerEntries` query.

### QueryLedgerEntriesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L163-L167

See also: [LedgerKey](03_messages.md#ledgerkey)

### QueryLedgerEntriesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L169-L173

See also: [LedgerEntry](03_messages.md#ledgerentry)


## LedgerEntry

To get a specific ledger entry, use the `LedgerEntry` query.

### QueryLedgerEntryRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L175-L181

See also: [LedgerKey](03_messages.md#ledgerkey)

### QueryLedgerEntryResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L183-L187

See also: [LedgerEntry](03_messages.md#ledgerentry)


## LedgerBalancesAsOf

The `LedgerBalancesAsOf` returns the balance of a ledger as of a given date.

### QueryLedgerBalancesAsOfRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L189-L195

See also: [LedgerKey](03_messages.md#ledgerkey)

### QueryLedgerBalancesAsOfResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L197-L201

See also: [BucketBalance](03_messages.md#bucketbalance)


## LedgerSettlements

To look up the settlemets for a ledger, use the `LedgerSettlements` query.

### QueryLedgerSettlementsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L203-L206

See also: [LedgerKey](03_messages.md#ledgerkey)

### QueryLedgerSettlementsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L208-L211

#### StoredSettlementInstructions

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger_settlement.proto#L51-L55

See also: [SettlementInstructions](03_messages.md#settlementinstructions)


## LedgerSettlementsByCorrelationID

Specific ledger settlements can be looked up with the `LedgerSettlementsByCorrelationID` query.

### QueryLedgerSettlementsByCorrelationIDRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L213-L218

See also: [LedgerKey](03_messages.md#ledgerkey)

### QueryLedgerSettlementsByCorrelationIDResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/query.proto#L220-L223

See also: [StoredSettlementInstructions](#storedsettlementinstructions)
