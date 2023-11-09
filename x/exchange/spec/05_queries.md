# Exchange Queries

There are several queries for getting information about things in the exchange module.

---
<!-- TOC 2 2 -->
  - [OrderFeeCalc](#orderfeecalc)
  - [GetOrder](#getorder)
  - [GetOrderByExternalID](#getorderbyexternalid)
  - [GetMarketOrders](#getmarketorders)
  - [GetOwnerOrders](#getownerorders)
  - [GetAssetOrders](#getassetorders)
  - [GetAllOrders](#getallorders)
  - [GetMarket](#getmarket)
  - [GetAllMarkets](#getallmarkets)
  - [Params](#params)
  - [ValidateCreateMarket](#validatecreatemarket)
  - [ValidateMarket](#validatemarket)
  - [ValidateManageFees](#validatemanagefees)


## OrderFeeCalc

The `OrderFeeCalc` query is used to find out the various required fee options for a given order.
The idea is that you can provide your [AskOrder](03_messages.md#askorder) or [BidOrder](03_messages.md#bidorder) in this query in order to identify what fees you'll need to pay.

Either an `ask_order` or a `bid_order` must be provided, but not both.

Each response field is a list of options available for the requested order.
If a response field is empty, then no fee of that type is required.

When creating the `AskOrder`, choose one entry from `creation_fee_options` to provide as the `order_creation_fee`.
Then, choose one entry from `settlement_flat_fee_options` and provide that as the `seller_settlement_flat_fee`.
For ask orders, the `settlement_ratio_fee_options` is purely informational and is the minimum that seller's settlement ratio fee that will be for the order.

When creating the `BidOrder`, choose one entry from `creation_fee_options` to provide as the `order_creation_fee`.
Then choose one entry from each of `settlement_flat_fee_options` and `settlement_ratio_fee_options`, add them together, and provide that as the `buyer_settlement_fees`.

### QueryOrderFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L96-L103

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).

### QueryOrderFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L105-L124


## GetOrder

Use the `GetOrder` query to look up an order by its id.

### QueryGetOrderRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L126-L130

### QueryGetOrderResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L132-L136

### Order

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/orders.proto#L13-L26

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).


## GetOrderByExternalID

Orders with external ids can be looked up using the `GetOrderByExternalID` query.

### QueryGetOrderByExternalIDRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L138-L144

### QueryGetOrderByExternalIDResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L146-L150

See also: [Order](#order).


## GetMarketOrders

To get all of the orders in a given market, use the `GetMarketOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetMarketOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L152-L163

### QueryGetMarketOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L165-L172

See also: [Order](#order).


## GetOwnerOrders

To get all of the orders with a specific owner (e.g. buyer or seller), use the `GetOwnerOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetOwnerOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L174-L185

### QueryGetOwnerOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L187-L194

See also: [Order](#order).


## GetAssetOrders

To get all of the orders with a specific asset denom, use the `GetAssetOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetAssetOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L196-L207

### QueryGetAssetOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L209-L216

See also: [Order](#order).


## GetAllOrders

To get all existing orders, use the `GetAllOrders` query.

This query is paginated.

### QueryGetAllOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L218-L222

### QueryGetAllOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L224-L231

See also: [Order](#order).


## GetMarket

All the information and setup for a market can be looked up using the `GetMarket` query.

### QueryGetMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L233-L237

### QueryGetMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L239-L245

See also: [Market](03_messages.md#market).


## GetAllMarkets

Use the `GetAllMarkets` query to get brief information about all markets.

### QueryGetAllMarketsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L247-L251

### QueryGetAllMarketsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L253-L260

### MarketBrief

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/market.proto#L42-L50


## Params

The exchange module params can be looked up using the `Params` query.

### QueryParamsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L262-L263

### QueryParamsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L265-L269

See also: [Params](06_params.md#params).


## ValidateCreateMarket

It's possible for a [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest) to result in a market setup that is problematic.
To verify that one is not problematic, this `ValidateCreateMarket` can be used.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateCreateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L271-L275

See also: [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest).

### QueryValidateCreateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L277-L287


## ValidateMarket

An existing market's setup can be checked for problems using the `ValidateMarket` query.

Any problems detected will be returned in the `error` field.

### QueryValidateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L289-L293

### QueryValidateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L295-L299


## ValidateManageFees

It's possible for a [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest) to result in a problematic setup for a market.
To verify that one does not result in such a state, use this `ValidateManageFees` query.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateManageFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L301-L305

See also: [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest).

### QueryValidateManageFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.17.0/proto/provenance/exchange/v1/query.proto#L307-L317
