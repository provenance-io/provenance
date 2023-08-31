# Concepts

The `x/hold` module is designed for use by other modules that need to lock funds in place in an account.

<!-- TOC -->
  - [Holds](#holds)
  - [Managing Holds](#managing-holds)
  - [Locked Coins](#locked-coins)

## Holds

"Holds" are an amount of funds in an account that must not be moved out of the account.
When a hold is placed on some funds, a record of them is created in the `x/hold` module.
When funds are released from a hold, those records are updated appropriately.

Funds with a hold on them remain in the owners account but cannot be spent, sent, delegated or otherwise removed from the account until they are released from hold.

A hold can only be placed on funds that would otherwise be spendable. E.g. you can place a hold on vested funds, but not unvested funds.

## Managing Holds

The `x/hold` module does not have any `Msg` or `Tx` endpoints for managing holds.
Putting holds on funds and releasing holds are actions that are only available via keeper functions.
It is expected that other modules will use the keeper functions (e.g.`AddHold` and `ReleaseHold`) as needed.

## Locked Coins

The `x/hold` module injects a `GetLockedCoinsFn` into the bank keeper in order to tell it which funds have a hold on them.
This allows the bank module and keeper functions to take holds into account when reporting bank account information.
Specifically, the bank keeper functions, `LockedCoins`, and `SpendableCoins` will reflect holds, as well as the `SpendableBalances` query.
The `AllBalances` query and similar keeper functions will still include the held funds though, since the funds actually **are** still in the account.
