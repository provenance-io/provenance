# Concepts

## Ledger

A Ledger is the primary data structure that tracks financial activities for a specific NFT. Each ledger is associated with a unique NFT address and maintains a specific denomination for its entries.

### Fields
- `nft_address`: The address of the NFT to which this ledger is linked
- `denom`: The denomination used for all entries in this ledger
- `next_pmt_date`: The next scheduled payment date in ISO 8601 format (YYYY-MM-DD)
- `next_pmt_amt`: The amount of the next scheduled payment
- `status`: The current status of the ledger
- `interest_rate`: The interest rate applied to the ledger
- `maturity_date`: The maturity date of the ledger in ISO 8601 format (YYYY-MM-DD)

## Ledger Entry

A Ledger Entry represents a single financial transaction or activity in the ledger. Each entry has a specific type and tracks various amounts related to principal, interest, and other balances.

### Fields
- `correlation_id`: Unique identifier for tracking with external systems (max 50 characters)
- `sequence`: Sequence number for ordering entries with the same effective date
- `type`: The type of ledger entry (see LedgerEntryType)
- `posted_date`: The date when the entry was recorded (ISO 8601 format: YYYY-MM-DD)
- `effective_date`: The date when the entry takes effect (ISO 8601 format: YYYY-MM-DD)
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
3. `LEDGER_ENTRY_TYPE_SCHEDULED_PAYMENT`: Represents a scheduled payment
   - Example: Regular monthly payment
4. `LEDGER_ENTRY_TYPE_UNSCHEDULED_PAYMENT`: Represents an unscheduled payment
   - Example: Extra payment or early payoff
5. `LEDGER_ENTRY_TYPE_FORECLOSURE_PAYMENT`: Represents a foreclosure-related payment
   - Example: Payment from foreclosure proceeds
6. `LEDGER_ENTRY_TYPE_FEE`: Represents a fee charged
   - Example: Origination fee, late payment fee
7. `LEDGER_ENTRY_TYPE_OTHER`: Represents other types of financial activities
   - Example: Adjustments, corrections, or special transactions

Ledgers are not meant to track every potential transaction type since there are only a few that are proven on-chain.

## Balance Tracking

The module maintains several types of balances:

1. **Principal Balance**
   - Original amount disbursed
   - Reduced by principal payments
   - Increased by disbursements
   - Reduced/Increased based on "other" adjustments

2. **Interest Balance**
   - Accrued interest
   - Reduced by interest payments
   - Reduced/Increased based on "other" adjustments

3. **Other Balance**
   - Fees and charges
   - Special adjustments
   - Miscellaneous amounts
   - Can affect principal/interest applications and balances

## Fund Transfers

The module supports two types of fund transfers:

1. **Basic Fund Transfer**
   - Tracks a single transfer amount
   - Includes status tracking (Pending, Processing, Completed, Failed)
   - Supports settlement timing control
   - Includes optional memo

2. **Fund Transfer with Settlement**
   - Supports multiple settlement instructions
   - Each instruction includes:
     - Amount
     - Recipient address
     - Optional memo
     - Settlement timing control
   - Used for complex payment scenarios

## State Management

The module maintains the following state:
- Ledger configurations for each NFT
- Historical ledger entries for each NFT
- Current balances and status for each NFT's financial position
- Fund transfer information and settlement instructions

## Query System

The module provides query endpoints to:
- Retrieve ledger configuration for a specific NFT
- Access historical ledger entries for a specific NFT
- View current balances and financial status
- Filter and search ledger entries
- Get aggregated financial information
- Track fund transfer status and settlement instructions 