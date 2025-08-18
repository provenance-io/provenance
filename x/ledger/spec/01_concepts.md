# Ledger Concepts

The `x/ledger` module manages financial tracking for NFTs and metadata scopes. There are several key concepts that define how financial activities are tracked and managed: ledger classes, ledgers, entries, balance buckets, and settlements. Each ledger has a unique identifier and maintains historical records of all financial activities.

<!-- TOC -->
- [Ledger Concepts](#ledger-concepts)
  - [Ledger Class](#ledger-class)
    - [Fields](#fields)
  - [Ledger](#ledger)
    - [Fields](#fields-1)
  - [Ledger Entry](#ledger-entry)
    - [Fields](#fields-2)
  - [Balance Buckets](#balance-buckets)
  - [Entry Types](#entry-types)
  - [Status Types](#status-types)
  - [Settlements](#settlements)
    - [Settlement Instructions](#settlement-instructions)
  - [Day Count Conventions](#day-count-conventions)
  - [Interest Accrual Methods](#interest-accrual-methods)
  - [Payment Frequencies](#payment-frequencies)
  - [Query System](#query-system)
  - [Key Generation](#key-generation)
  - [Balance Calculation](#balance-calculation)

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
- `interest_rate`: The interest rate applied to the ledger (10000000 = 10.000000% - 6 decimal places)
- `maturity_date`: The maturity date in epoch days
- `interest_day_count_convention`: Day count convention for interest calculations
- `interest_accrual_method`: Method used for interest accrual
- `payment_frequency`: Frequency of scheduled payments

## Ledger Entry

A Ledger Entry represents a single financial transaction or activity in the ledger. Each entry has a specific type, tracks time-based information on how to order the entry, indicates how the `total_amt` should be applied to various buckets, and stores the resulting balances for each of the buckets.

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
- `balance_amounts`: Current balances for each bucket type after this entry

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

## Settlements

Settlements represent fund transfer instructions with associated settlement logic. They allow owners or servicers of assets to record transfers of tokens against a ledger entry. 

### Settlement Instructions
- `settlement_instructions`: Array of settlement instructions for fund transfers
- Each instruction contains the logic for processing a specific type of fund transfer
- Settlements are stored and can be queried by ledger or correlation ID

## Day Count Conventions

The module supports several day count conventions for interest calculations:

1. **ACTUAL_365**: Uses actual days with 365-day denominator (or 365.25 for leap years)
2. **ACTUAL_360**: Uses actual days with 360-day denominator
3. **THIRTY_360**: Assumes 30 days per month and 360 days per year
4. **ACTUAL_ACTUAL**: Uses actual days and actual year length (365 or 366)
5. **DAYS_365**: Always uses 365 days regardless of leap years
6. **DAYS_360**: Always uses 360 days

## Interest Accrual Methods

The module supports various interest accrual methods:

1. **SIMPLE_INTEREST**: Interest calculated only on principal amount
2. **COMPOUND_INTEREST**: Interest calculated on principal and accumulated interest
3. **DAILY_COMPOUNDING**: Interest compounded daily
4. **MONTHLY_COMPOUNDING**: Interest compounded monthly
5. **QUARTERLY_COMPOUNDING**: Interest compounded quarterly
6. **ANNUAL_COMPOUNDING**: Interest compounded annually
7. **CONTINUOUS_COMPOUNDING**: Theoretical limit of continuous compounding

## Payment Frequencies

The module supports various payment frequencies:

1. **DAILY**: Daily payments
2. **WEEKLY**: Weekly or biweekly payments
3. **MONTHLY**: Monthly payments (most common for consumer loans and mortgages)
4. **QUARTERLY**: Quarterly payments
5. **ANNUALLY**: Annual payments

## Query System

The module provides query endpoints to:
- Retrieve ledger configuration for a specific asset
- Access historical ledger entries
- View current balances and financial status
- Get ledger class configurations and types
- Query settlements that have been record

## Key Generation

Ledger identifiers are generated using bech32 encoding:
- Combines `asset_class_id` and `nft_id` with a null byte delimiter
- Uses the human-readable part `"ledger"`
- Provides a readable, unique identifier for each ledger
- Example: `ledger1w3jhxapdden8gttrd3shxuedd9jqqcm0wdkk7ue3x44hjwtyw5uxzvnhd3ehg73kvec8svmsx3khzur209ex6dtrvack5amv8pehzl09ezy`

## Balance Calculation

Balance is determined by querying the balance of each bucket as of the effective date queried.
- Process all ledger entries up to a specific date
- Use the `balance_amounts` field from the most recent entry
- Ensure chronological ordering by effective date and sequence
- Provide point-in-time financial state information 