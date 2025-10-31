# Ledger State

The `x/ledger` module uses key/value pairs to store ledger data in state.

---
<!-- TOC -->
  - [Ledgers](#ledgers)
  - [Ledger Entries](#ledger-entries)
  - [Fund Transfers With Settlement](#fund-transfers-with-settlement)
  - [Ledger Classes](#ledger-classes)
  - [Ledger Class Entry Types](#ledger-class-entry-types)
  - [Ledger Class Status Types](#ledger-class-status-types)
  - [Ledger Class Bucket Types](#ledger-class-bucket-types)


## Ledgers

Each ledger is recorded by its [ledger id](01_concepts.md#ledger-identifiers).

```
0x01 | <ledger id> -> protobuf(Ledger)
```

Where:
* `0x01` is the type byte, and has a value of `1` for these records.
* `<ledger id>` is a string with the [ledger identifier](01_concepts.md#ledger-identifiers).
* `protobuf(Ledger)` is a protobuf-encoded [Ledger](03_messages.md#ledger) object.

The `Ledger.Key` field is set to `nil` before storing the ledger in state, and is re-added upon retrieval.


## Ledger Entries

Each ledger entry is stored by its [ledger id](01_concepts.md#ledger-identifiers) and correlation id.

```
0x02 | <ledger id> | 0x00 | <correlation id> -> protobuf(LedgerEntry)
```

Where:
* `0x02` is the type byte, and has a value of `2` for these records.
* `<ledger id>` is a string with the [ledger identifier](01_concepts.md#ledger-identifiers).
* `0x00` is a null byte separator.
* `<correlation id>` is a string with the correlation identifier for the ledger entry that these settlement instructions belong to.
* `protobuf(LedgerEntry)` is a protobuf-encoded [LedgerEntry](03_messages.md#ledgerentry) object.


## Fund Transfers With Settlement

Information about each fund transfer with settlement is stored by its [ledger id](01_concepts.md#ledger-identifiers) and ledger entry correlation id.

```
0x08 | <ledger id> | 0x00 | <correlation id> -> protobuf(StoredSettlementInstructions)
```

Where:
* `0x08` is the type byte, and has a value of `8` for these records.
* `<ledger id>` is a string with the [ledger identifier](01_concepts.md#ledger-identifiers).
* `0x00` is a null byte separator.
* `<correlation id>` is a string with the correlation identifier for the ledger entry that these settlement instructions belong to.
* `protobuf(StoredSettlementInstructions)` is a protobuf-encoded `StoredSettlementInstructions` object (see below).

#### StoredSettlementInstructions

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger_settlement.proto#L54-L58

See also: [SettlementInstruction](03_messages.md#settlementinstruction).


## Ledger Classes

Each ledger class is stored by its class id.

```
0x03 | <class id> -> protobuf(LedgerClass)
```

Where:
* `0x03` is the type byte, and has a value of `3` for these records.
* `<class id>` is a string containing the ledger class identifier.
* `protobuf(LedgerClass)` is a protobuf-encoded [LedgerClass](03_messages.md#ledgerclass) object.


## Ledger Class Entry Types

Each class's entry types are stored by its class id and entry type id.

```
0x04 | <class id> | 0x00 | <entry type id> -> protobuf(LedgerClassEntryType)
```

Where:
* `0x04` is the type byte, and has a value of `4` for these records.
* `<class id>` is a string containing the ledger class identifier.
* `0x00` is a null byte separator.
* `<entry type id>` is a string with the entry type's identifier.
* `protobuf(LedgerClassEntryType)` is a protobuf-encoded [LedgerClassEntryType](03_messages.md#ledgerclassentrytype) object.


## Ledger Class Status Types

Each class status type is stored by its class id and status type id.

```
0x05 | <class id> | 0x00 | <status type id> -> protobuf(LedgerClassStatusType)
```

Where:
* `0x05` is the type byte, and has a value of `5` for these records.
* `<class id>` is a string containing the ledger class identifier.
* `0x00` is a null byte separator.
* `<status type id>` is a string containing the status type's identifier.
* `protobuf(LedgerClassStatusType)` is a protobuf-encoded [LedgerClassStatusType](03_messages.md#ledgerclassstatustype) object.


## Ledger Class Bucket Types

Each class bucket type is stored by its class id and bucket type id.

```
0x06 | <class id> | 0x00 | <bucket type id> -> protobuf(LedgerClassBucketType)
```

Where:
* `0x06` is the type byte, and has a value of `6` for these records.
* `<class id>` is a string containing the ledger class identifier.
* `0x00` is a null byte separator.
* `<bucket type id>` is a string containing the bucket type's identifier.
* `protobuf(LedgerClassBucketType)` is a protobuf-encoded [LedgerClassBucketType](03_messages.md#ledgerclassbuckettype) object.
