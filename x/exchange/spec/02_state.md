# Exchange State

The Exchange module manages several things in state.

Big-endian ordering is used for all conversions between numbers and byte arrays.

---
<!-- TOC -->
  - [Params](#params)
    - [Default Split](#default-split)
    - [Specific Denom Split](#specific-denom-split)
  - [Markets](#markets)
    - [Market Create-Ask Flat Fee](#market-create-ask-flat-fee)
    - [Market Create-Bid Flat Fee](#market-create-bid-flat-fee)
    - [Market Seller Settlement Flat Fee](#market-seller-settlement-flat-fee)
    - [Market Seller Settlement Ratio Fee](#market-seller-settlement-ratio-fee)
    - [Market Buyer Settlement Flat Fee](#market-buyer-settlement-flat-fee)
    - [Market Buyer Settlement Ratio Fee](#market-buyer-settlement-ratio-fee)
    - [Market Inactive Indicator](#market-inactive-indicator)
    - [Market User-Settle Indicator](#market-user-settle-indicator)
    - [Market Permissions](#market-permissions)
    - [Market Create-Ask Required Attributes](#market-create-ask-required-attributes)
    - [Market Create-Bid Required Attributes](#market-create-bid-required-attributes)
    - [Market Account](#market-account)
    - [Market Details](#market-details)
    - [Known Market ID](#known-market-id)
    - [Last Market ID](#last-market-id)
  - [Orders](#orders)
    - [Ask Orders](#ask-orders)
    - [Bid Orders](#bid-orders)
    - [Last Order ID](#last-order-id)
  - [Indexes](#indexes)
    - [Market to Order](#market-to-order)
    - [Owner Address to Order](#owner-address-to-order)
    - [Asset Denom to Order](#asset-denom-to-order)
    - [Market External ID to Order](#market-external-id-to-order)


## Params

All params entries start with the type byte `0x00` followed by a string identifying the entry type.

Each `<split>` is stored as a `uint16` (2 bytes) in big-endian order.

The byte `0x1E` is used in a few places as a record separator.

See also: [Params](06_params.md#params).


### Default Split

The default split defines the split amount (in basis points) the exchange receives of fees when there is not an applicable specific denom split.

* Key:`0x00 | "split" (5 bytes)`
* Value: `<split (2 bytes)>`


### Specific Denom Split

A specific denom split is a split amount (in basis points) the exchange receives of fees for fees paid in a specific denom.

* Key: `0x00 | "split" (5 bytes) | <denom (string)>`
* Value: `<split (2 bytes)>`

See also: [DenomSplit](06_params.md#denomsplit).


## Markets

Each aspect of a market is stored separately for specific lookup.

Each `<market id>` is a `uint32` (4 bytes) in big-endian order.

Most aspects of a market have keys that start with the type byte `0x01`, followed by the `<market id>` then another type byte.

See also: [Market](03_messages.md#market).


### Market Create-Ask Flat Fee

One entry per configured denom.

* Key: `0x01 | <market id (4 bytes)> | 0x00 | <denom (string)>`
* Value: `<amount (string)>`


### Market Create-Bid Flat Fee

One entry per configured denom.

* Key: `0x01 | <market id (4 bytes)> | 0x01 | <denom (string)>`
* Value: `<amount (string)>`


### Market Seller Settlement Flat Fee

One entry per configured denom.

* Key: `0x01 | <market id (4 bytes)> | 0x02 | <denom (string)>`
* Value: `<amount (string)>`


### Market Seller Settlement Ratio Fee

One entry per configured price:fee denom pair.

* Key: `0x01 | <market id (4 bytes)> | 0x03 | <price denom (string)> | 0x1E | <fee denom (string)>`
* Value: `<price amount (string)> | 0x1E | <fee amount (string)>`

See also: [FeeRatio](03_messages.md#feeratio).


### Market Buyer Settlement Flat Fee

One entry per configured denom.

* Key: `0x01 | <market id (4 bytes)> | 0x04 | <denom (string)>`
* Value: `<amount (string)>`


### Market Buyer Settlement Ratio Fee

One entry per configured price:fee denom pair.

* Key: `0x01 | <market id (4 bytes)> | 0x05 | <price denom (string)> | 0x1E | <fee denom (string)>`
* Value: `<price amount (string)> | 0x1E | <fee amount (string)>`

See also: [FeeRatio](03_messages.md#feeratio).


### Market Inactive Indicator

When a market has `accepting_orders = false`, this state entry will exist.
When it has `accepting_orders = true`, this entry will not exist.

* Key: `0x01 | <market id (4 bytes)> | 0x06`
* Value: `<nil (0 bytes)>`


### Market User-Settle Indicator

When a market has `allow_user_settlement = true`, this state entry will exist.
When it has `allow_user_settlement = false`, this entry will not exist.

* Key: `0x01 | <market id (4 bytes)> | 0x07`
* Value: `<nil (0 bytes)>`


### Market Permissions

When an address has a given permission in a market, the following entry will exist.

* Key: `0x01 | <market id (4 bytes)> | 0x08 | <addr len (1 byte)> | <addr> | <permission type byte (1 byte)>`
* Value: `<nil (0 bytes)>`

The `<permission type byte>` is a single byte as `uint8` with the same values as the enum entries, e.g. `PERMISSION_CANCEL` is `0x03`.

See also: [AccessGrant](03_messages.md#accessgrant) and [Permission](03_messages.md#permission).


### Market Create-Ask Required Attributes

* Key: `0x01 | <market id (4 bytes)> | 0x09 | 0x00`
* Value: `<list of strings separated by 0x1E>`


### Market Create-Bid Required Attributes

* Key: `0x01 | <market id (4 bytes)> | 0x09 | 0x01`
* Value: `<list of strings separated by 0x1E>`


### Market Account

Each market has an associated `MarketAccount` with an address derived from the `market_id`.
Each `MarketAccount` is stored using the `Accounts` module.

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/market.proto#L14-L26


### Market Details

The [MarketDetails](03_messages.md#marketdetails) are stored as part of the `MarketAccount` (in the `x/auth` module).


### Known Market ID

These entries are used to indicate that a given market exists.

* Key: `0x07 | <market id (4 bytes)>`
* Value: `<nil (0 bytes)>`


### Last Market ID

This indicates the last market-id that was auto-selected for use.

When a `MsgGovCreateMarketRequest` is processed that has a `market_id` of `0` (zero), the next available market id is auto selected.
Starting with the number after what's in this state entry, each market id is sequentially checked until an available one is found.
The new market gets that id, then this entry is then updated to indicate what that was.

* Key: `0x06`
* Value: `<market id (4 bytes)>`

When a `MsgGovCreateMarketRequest` is processed that has a non-zero `market_id`, this entry is not considered or altered.


## Orders

Each `<order id>` is a `uint64` (8 bytes) in big-endian order.

Orders are stored using the following format:

* Key: `0x02 | <order id (8 bytes)>`
* Value `<order type byte> | protobuf(order type)`

The `<order type byte>` has these possible values:
* `0x00` => Ask Order
* `0x01` => Bid Order


### Ask Orders

* Key: `0x02 | <order id (8 bytes)>`
* Value: `0x00 | protobuf(AskOrder)`

See also: [AskOrder](03_messages.md#askorder).


### Bid Orders

* Key: `0x02 | <order id (8 bytes)>`
* Value: `0x01 | protobuf(BidOrder)`

See also: [BidOrder](03_messages.md#bidorder).


### Last Order ID

Whenever an order is created, this value is looked up and incremented to get the new order's id.
Then this entry is updated to reflect the new order.

* Key: `0x08`
* Value: `<order id (8 bytes)>`


## Indexes

Several index entries are maintained to help facilitate look-ups.

The `<order type byte>` values are the same as those described in [Orders](#orders).


### Market to Order

This index can be used to find orders in a given market.

* Key: `0x03 | <market id (4 bytes)> | <order id (8 bytes)>`
* Value: `<order type byte (1 byte)>`


### Owner Address to Order

This index can be used to find orders with a given buyer or seller.

* Key: `0x04 | <addr len (1 byte)> | <addr> | <order id (8 bytes)>`
* Value: `<order type byte (1 byte)>`


### Asset Denom to Order

This index can be used to find orders involving a given `assets` denom.

* Key: `0x05 | <asset denom> | <order id (8 bytes)>`
* Value: `<order type byte (1 byte)>`


### Market External ID to Order

This index is used to look up orders by their market and external id.

* Key: `0x09 | <market id (4 bytes)> | <external id (string)>`
* Value: `<order id (8 bytes)>`
