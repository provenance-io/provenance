# State

The Ledger module maintains several types of state to track financial activities and balances for NFTs.

## Ledger Configuration

The module stores configuration information for each NFT's ledger, including:
- The NFT address
- The denomination used for entries
- Any specific settings or parameters for the ledger

## Ledger Entries

Historical ledger entries are stored for each NFT, maintaining a complete record of all financial activities. Each entry includes:
- Transaction details (type, amounts, dates)
- Impact on various balances (principal, interest, other)
- Unique identifiers for tracking and reference

## Balance Tracking

The module maintains current balance information for each NFT, including:
- Principal balance
- Interest balance
- Other balance categories
- Total outstanding amounts

## State Transitions

State transitions occur when:
1. A new ledger is created for an NFT
2. New entries are added to an existing ledger
3. Balances are updated due to transactions
4. Configuration changes are made to a ledger

## State Access

State can be accessed through:
- Query endpoints for reading ledger information
- Transaction handlers for modifying ledger state
- Event emission for tracking state changes 