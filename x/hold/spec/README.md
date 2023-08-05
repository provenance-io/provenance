# `Hold`

## Abstract

The Hold module keeps track of funds that are being held in an account.
Funds with a hold on them cannot be sent, spent, delegated, or otherwise be moved from the account.
Funds are held in-place in their account; they are not moved anywhere.

Holds are created and released through internal-only keeper functions.
The bank module's `SpendableBalances` query will not return funds that have holds on them.
I.e. held funds are not reported as spendable.

TODO[1607]: Finish spec docs.

## Contents

1. **[State](01_state.md)**
2. **[Messages](02_messages.md)**
3. **[Events](03_events.md)**
4. **[Queries](04_queries.md)**
5. **[Genesis](04_genesis.md)**
