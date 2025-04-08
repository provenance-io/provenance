# Concepts

## Ledger

A Ledger is the primary data structure that tracks financial activities for a specific NFT. Each ledger is associated with a unique NFT address and maintains a specific denomination for its entries.

### Fields
- `nft_address`: The address of the NFT to which this ledger is linked
- `denom`: The denomination used for all entries in this ledger

## Ledger Entry

A Ledger Entry represents a single financial transaction or activity in the ledger. Each entry has a specific type and tracks various amounts related to principal, interest, and other balances.

### Fields
- `uuid`: Unique identifier for the ledger entry
- `type`: The type of ledger entry (see LedgerEntryType)
- `posted_date`: The date when the entry was recorded
- `effective_date`: The date when the entry takes effect
- `amt`: The total amount of the entry
- `prin_applied_amt`: Amount applied to principal
- `prin_bal_amt`: Remaining principal balance
- `int_applied_amt`: Amount applied to interest
- `int_bal_amt`: Remaining interest balance
- `other_applied_amt`: Amount applied to other categories
- `other_bal_amt`: Remaining other balance

## Ledger Entry Types

The module supports several types of ledger entries:

1. `LEDGER_ENTRY_TYPE_UNSPECIFIED`: Default type, not used in normal operations
2. `LEDGER_ENTRY_TYPE_DISBURSEMENT`: Represents funds being disbursed
   - Example: Initial loan amount disbursed to borrower
3. `LEDGER_ENTRY_TYPE_PAYMENT`: Represents a payment made
   - Example: Monthly payment from borrower
4. `LEDGER_ENTRY_TYPE_FEE`: Represents a fee charged
   - Example: Origination fee, late payment fee
5. `LEDGER_ENTRY_TYPE_OTHER`: Represents other types of financial activities
   - Example: Adjustments, corrections, or special transactions

## Balance Tracking

The module maintains several types of balances:

1. **Principal Balance**
   - Original amount disbursed
   - Reduced by principal payments
   - Increased by disbursements

2. **Interest Balance**
   - Accrued interest
   - Reduced by interest payments
   - Updated based on interest calculations

3. **Other Balance**
   - Fees and charges
   - Special adjustments
   - Miscellaneous amounts

## State Management

The module maintains the following state:
- Ledger configurations for each NFT
- Historical ledger entries for each NFT
- Current balances and status for each NFT's financial position

## Query System

The module provides query endpoints to:
- Retrieve ledger configuration for a specific NFT
- Access historical ledger entries for a specific NFT
- View current balances and financial status
- Filter and search ledger entries
- Get aggregated financial information 