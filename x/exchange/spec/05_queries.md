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
  - [GetCommitment](#getcommitment)
  - [GetAccountCommitments](#getaccountcommitments)
  - [GetMarketCommitments](#getmarketcommitments)
  - [GetAllCommitments](#getallcommitments)
  - [GetMarket](#getmarket)
  - [GetAllMarkets](#getallmarkets)
  - [Params](#params)
  - [CommitmentSettlementFeeCalc](#commitmentsettlementfeecalc)
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

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L126-L133

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).

### QueryOrderFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L135-L154


## GetOrder

Use the `GetOrder` query to look up an order by its id.

### QueryGetOrderRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L156-L160

### QueryGetOrderResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L162-L166

### Order

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/orders.proto#L13-L26

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).


## GetOrderByExternalID

Orders with external ids can be looked up using the `GetOrderByExternalID` query.

### QueryGetOrderByExternalIDRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L168-L174

### QueryGetOrderByExternalIDResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L176-L180

See also: [Order](#order).


## GetMarketOrders

To get all of the orders in a given market, use the `GetMarketOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetMarketOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L182-L193

### QueryGetMarketOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L195-L202

See also: [Order](#order).


## GetOwnerOrders

To get all of the orders with a specific owner (e.g. buyer or seller), use the `GetOwnerOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetOwnerOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L204-L215

### QueryGetOwnerOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L217-L224

See also: [Order](#order).


## GetAssetOrders

To get all of the orders with a specific asset denom, use the `GetAssetOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetAssetOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L226-L237

### QueryGetAssetOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L239-L246

See also: [Order](#order).


## GetAllOrders

To get all existing orders, use the `GetAllOrders` query.

This query is paginated.

### QueryGetAllOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L248-L252

### QueryGetAllOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L254-L261

See also: [Order](#order).


## GetCommitment

To find out how much an account has committed to a market, use the `GetCommitment` query.

### QueryGetCommitmentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L263-L269

### QueryGetCommitmentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L271-L276


## GetAccountCommitments

To look up the amounts an account has committed to any market, use the `GetAccountCommitments` query.

### QueryGetAccountCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L278-L282

### QueryGetAccountCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L284-L288


## GetMarketCommitments

To get the amounts committed to a market by any account, use the `GetMarketCommitments` query.

### QueryGetMarketCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L290-L297

### QueryGetMarketCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L299-L306


## GetAllCommitments

To get all funds committed by any account to any market, use the `GetAllCommitments` query.

### QueryGetAllCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L308-L312

### QueryGetAllCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L314-L321


## GetMarket

All the information and setup for a market can be looked up using the `GetMarket` query.

### QueryGetMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L323-L327

### QueryGetMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L329-L335

See also: [Market](03_messages.md#market).


## GetAllMarkets

Use the `GetAllMarkets` query to get brief information about all markets.

### QueryGetAllMarketsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L337-L341

### QueryGetAllMarketsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L343-L350

### MarketBrief

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/market.proto#L42-L50


## Params

The exchange module params can be looked up using the `Params` query.

### QueryParamsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L352-L353

### QueryParamsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L355-L359

See also: [Params](06_params.md#params).


## CommitmentSettlementFeeCalc

To find out the additional tx fee required for a commitment settlement, use the `CommitmentSettlementFeeCalc` query.

### QueryCommitmentSettlementFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L361-L371

See also: [MsgMarketCommitmentSettleRequest](03_messages.md#msgmarketcommitmentsettlerequest).

### QueryCommitmentSettlementFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L373-L388


## ValidateCreateMarket

It's possible for a [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest) to result in a market setup that is problematic.
To verify that one is not problematic, this `ValidateCreateMarket` can be used.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateCreateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L390-L394

See also: [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest).

### QueryValidateCreateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L396-L406


## ValidateMarket

An existing market's setup can be checked for problems using the `ValidateMarket` query.

Any problems detected will be returned in the `error` field.

### QueryValidateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L408-L412

### QueryValidateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L414-L418


## ValidateManageFees

It's possible for a [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest) to result in a problematic setup for a market.
To verify that one does not result in such a state, use this `ValidateManageFees` query.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateManageFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L420-L424

See also: [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest).

### QueryValidateManageFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L426-L436
