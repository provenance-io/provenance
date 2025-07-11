# Concepts

## Ledger Class

A Ledger Class defines the configuration for a specific class of assets. It determines how ledgers will be managed for a particular type of asset, whether it's defined by a scope specification or NFT class.

### Fields
- `ledger_class_id`: Unique identifier for the ledger class
- `asset_class_id`: Scope Specification ID or NFT Class ID
- `denom`: The denomination used for all entries in ledgers of this class
- `maintainer_address`: Address of the maintainer for the ledger class

## Ledger

A Ledger is the primary data structure that tracks financial activities for a specific NFT or scope. Each ledger is associated with a unique asset identifier and maintains balances according to its ledger class configuration.

### Fields
- `key`: Contains the NFT/scope identifier and asset class ID (Metadata Scope Specification or NFT Class)
- `ledger_class_id`: Reference to the ledger class configuration
- `status_type_id`: Current status of the ledger
- `next_pmt_date`: Next scheduled payment date in epoch days
- `next_pmt_amt`: Amount of the next scheduled payment
- `interest_rate`: The interest rate applied to the ledger
- `maturity_date`: The maturity date in epoch days

## Ledger Entry

A Ledger Entry represents a single financial transaction or activity in the ledger. Each entry has a specific type, tracks time based information on how to order the entry, indicates how the `total_amt` should be applied to various buckets, and allows the servicing entity to store balances for each of the buckets.

### Fields
- `correlation_id`: Unique identifier for tracking with external systems (max 50 characters)
- `reverses_correlation_id`: If this entry reverses another entry, the correlation ID of the reversed entry
- `is_void`: Indicates if this entry is void and should be excluded from balance calculations
- `sequence`: Sequence number for ordering entries with the same effective date (less than 100)
- `entry_type_id`: The type of ledger entry
- `posted_date`: The date when the entry was recorded in epoch days
- `effective_date`: The date when the entry takes effect in epoch days
- `total_amt`: The total amount of the entry
- `applied_amounts`: Amounts applied to different balance buckets
- `bucket_balances`: Current balances for each bucket type

## Balance Buckets

The module maintains balances in different buckets. Each of these buckets is configurable by the maintainer of the ledger class.

Examples of potential buckets that may be configured by the ledger class maintainer:

1. **Principal Bucket**
   - Tracks the principal amount
   - Affected by disbursements and principal payments

2. **Interest Bucket**
   - Tracks interest amounts
   - Affected by interest accruals and payments

3. **Other Bucket**
   - Tracks miscellaneous amounts
   - Used for fees, adjustments, and special transactions

## Entry Types

The module supports any type of entry type configured by the maintainer of the ledger class.

Examples of potential entry types that may be configured by the ledger class maintainer:

1. `DISBURSEMENT`: Represents funds being disbursed
   - Example: Initial loan amount disbursed to borrower

2. `SCHEDULED_PAYMENT`: Represents a scheduled payment
   - Example: Regular monthly payment

3. `UNSCHEDULED_PAYMENT`: Represents an unscheduled payment
   - Example: Extra payment or early payoff

4. `FORECLOSURE_PAYMENT`: Represents a foreclosure-related payment
   - Example: Payment from foreclosure proceeds

5. `FEE`: Represents a fee charged
   - Example: Origination fee, late payment fee

6. `OTHER`: Represents other types of financial activities
   - Example: Adjustments, corrections, or special transactions

## Status Types

The module supports any status for a ledger as configured by the maintainer of the ledger class.

Examples of potential ledger statuses:
1. `IN_REPAYMENT`: Normal repayment status
2. `IN_FORECLOSURE`: Foreclosure process initiated
3. `FORBEARANCE`: Temporary payment relief granted
4. `DEFERMENT`: Payment deferral period
5. `BANKRUPTCY`: Bankruptcy status
6. `CLOSED`: Ledger is closed
7. `CANCELLED`: Ledger is cancelled
8. `SUSPENDED`: Ledger is suspended
9. `PIF`: Ledger is Paid-in-full
10. `OTHER`: Other status types

## Query System

The module provides query endpoints to:
- Retrieve ledger configuration for a specific asset
- Access historical ledger entries
- View current balances and financial status
- Filter and search ledger entries
- Get aggregated financial information
- Track entry status and balances 