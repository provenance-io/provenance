# Ledger Concepts

The `x/ledger` module manages financial tracking for NFTs and metadata scopes.
There are several key concepts that define how financial activities are tracked and managed: ledger classes, ledgers, entries, balance buckets, and settlements.
Each ledger has a unique identifier and maintains historical records of all financial activities.

Ledgers can be created and managed by authorized maintainers through normal Msg requests.
A ledger can track multiple types of financial activities including disbursements, payments, fees, and adjustments.
The module supports configurable entry types, status types, and balance buckets to accommodate various financial tracking requirements.

---
<!-- TOC 2 2 -->
  - [Ledger Class](#ledger-class)
  - [Ledger](#ledger)
  - [Ledger Entry](#ledger-entry)
  - [Balance Buckets](#balance-buckets)
  - [Entry Types](#entry-types)
  - [Status Types](#status-types)
  - [Settlements](#settlements)
  - [Ledger Identifiers](#ledger-identifiers)

## Ledger Class

A Ledger Class defines the configuration for a specific class of assets. 
It determines how ledgers will be managed for a particular type of asset, whether it's defined by a scope specification or NFT class.

See also: [LedgerClass](03_messages.md#ledgerclass).

## Ledger

A Ledger is the primary data structure that tracks financial activities for a specific NFT or scope. 
Each ledger is associated with a unique asset identifier and maintains balances according to its ledger class configuration.

See also: [Ledger](03_messages.md#ledger).

## Ledger Entry

A Ledger Entry represents a single financial transaction or activity in the ledger. 
Each entry has a specific type, tracks time-based information on how to order the entry, indicates how the `total_amt` 
should be applied to various buckets, and stores the resulting balances for each of the buckets.

See also: [LedgerEntry](03_messages.md#ledgerentry).

## Balance Buckets

The module maintains balances in different buckets. Each of these buckets is configurable by the maintainer of the ledger class.

Examples of potential buckets that may be configured by the ledger class maintainer:

1. **Principal Bucket**
   - Tracks the principal amount.
   - Affected by disbursements and principal payments.

2. **Interest Bucket**
   - Tracks interest amounts.
   - Affected by interest accruals and payments.

3. **Other Bucket**
   - Tracks miscellaneous amounts.
   - Used for fees, adjustments, and special transactions.

See also: [BucketBalance](03_messages.md#bucketbalance).

## Entry Types

The module supports any type of entry type configured by the maintainer of the ledger class.

Examples of potential entry types that may be configured by the ledger class maintainer:

1. `DISBURSEMENT`: Represents funds being disbursed.
   - Example: Initial loan amount disbursed to borrower.

2. `SCHEDULED_PAYMENT`: Represents a scheduled payment.
   - Example: Regular monthly payment.

3. `UNSCHEDULED_PAYMENT`: Represents an unscheduled payment.
   - Example: Extra payment or early payoff.

4. `FORECLOSURE_PAYMENT`: Represents a foreclosure-related payment.
   - Example: Payment from foreclosure proceeds.

5. `FEE`: Represents a fee charged.
   - Example: Origination fee, late payment fee.

6. `OTHER`: Represents other types of financial activities.
   - Example: Adjustments, corrections, or special transactions.

See also: [LedgerClassEntryType](03_messages.md#ledgerclassentrytype).

## Status Types

The module supports any status for a ledger as configured by the maintainer of the ledger class.

Examples of potential ledger statuses:
1. `IN_REPAYMENT`: Normal repayment status.
2. `IN_FORECLOSURE`: Foreclosure process initiated.
3. `FORBEARANCE`: Temporary payment relief granted.
4. `DEFERMENT`: Payment deferral period.
5. `BANKRUPTCY`: Bankruptcy status.
6. `CLOSED`: Ledger is closed.
7. `CANCELLED`: Ledger is cancelled.
8. `SUSPENDED`: Ledger is suspended.
9. `PIF`: Ledger is Paid-in-full.
10. `OTHER`: Other status types.

See also: [LedgerClassStatusType](03_messages.md#ledgerclassstatustype).

## Settlements

Settlements represent fund transfer instructions with associated settlement logic.
They allow owners or servicers of assets to record transfers of tokens against a ledger entry. 

See also: [SettlementInstruction](03_messages.md#settlementinstruction).

## Ledger Identifiers

Ledger identifiers are generated using bech32 encoding:
- Combines `asset_class_id` and `nft_id` with a null byte delimiter
- Uses the human-readable part `"ledger"`
- Provides a readable, unique identifier for each ledger
- Example: `ledger1w3jhxapdden8gttrd3shxuedd9jqqcm0wdkk7ue3x44hjwtyw5uxzvnhd3ehg73kvec8svmsx3khzur209ex6dtrvack5amv8pehzl09ezy`

See also: [LedgerKey](03_messages.md#ledgerkey).
