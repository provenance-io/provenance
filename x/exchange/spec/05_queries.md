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
  - [GetPayment](#getpayment)
  - [GetPaymentsWithSource](#getpaymentswithsource)
  - [GetPaymentsWithTarget](#getpaymentswithtarget)
  - [GetAllPayments](#getallpayments)
  - [PaymentFeeCalc](#paymentfeecalc)


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

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L157-L164

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).

### QueryOrderFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L166-L185


## GetOrder

Use the `GetOrder` query to look up an order by its id.

### QueryGetOrderRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L187-L191

### QueryGetOrderResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L193-L197

### Order

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/orders.proto#L15-L28

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).


## GetOrderByExternalID

Orders with external ids can be looked up using the `GetOrderByExternalID` query.

### QueryGetOrderByExternalIDRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L199-L205

### QueryGetOrderByExternalIDResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L207-L211

See also: [Order](#order).


## GetMarketOrders

To get all of the orders in a given market, use the `GetMarketOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetMarketOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L213-L224

### QueryGetMarketOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L226-L233

See also: [Order](#order).


## GetOwnerOrders

To get all of the orders with a specific owner (e.g. buyer or seller), use the `GetOwnerOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetOwnerOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L235-L246

### QueryGetOwnerOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L248-L255

See also: [Order](#order).


## GetAssetOrders

To get all of the orders with a specific asset denom, use the `GetAssetOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetAssetOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L257-L268

### QueryGetAssetOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L270-L277

See also: [Order](#order).


## GetAllOrders

To get all existing orders, use the `GetAllOrders` query.

This query is paginated.

### QueryGetAllOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L279-L283

### QueryGetAllOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L285-L292

See also: [Order](#order).


## GetCommitment

To find out how much an account has committed to a market, use the `GetCommitment` query.

### QueryGetCommitmentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L294-L300

### QueryGetCommitmentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L302-L311


## GetAccountCommitments

To look up the amounts an account has committed to any market, use the `GetAccountCommitments` query.

### QueryGetAccountCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L313-L317

### QueryGetAccountCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L319-L323


## GetMarketCommitments

To get the amounts committed to a market by any account, use the `GetMarketCommitments` query.

### QueryGetMarketCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L325-L332

### QueryGetMarketCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L334-L341


## GetAllCommitments

To get all funds committed by any account to any market, use the `GetAllCommitments` query.

### QueryGetAllCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L343-L347

### QueryGetAllCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L349-L356


## GetMarket

All the information and setup for a market can be looked up using the `GetMarket` query.

### QueryGetMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L358-L362

### QueryGetMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L364-L370

See also: [Market](03_messages.md#market).


## GetAllMarkets

Use the `GetAllMarkets` query to get brief information about all markets.

### QueryGetAllMarketsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L372-L376

### QueryGetAllMarketsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L378-L385

### MarketBrief

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/market.proto#L42-L50


## Params

The exchange module params can be looked up using the `Params` query.

### QueryParamsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L387-L388

### QueryParamsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L390-L394

See also: [Params](06_params.md#params).


## CommitmentSettlementFeeCalc

To find out the additional tx fee required for a commitment settlement, use the `CommitmentSettlementFeeCalc` query.

### QueryCommitmentSettlementFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L396-L406

See also: [MsgMarketCommitmentSettleRequest](03_messages.md#msgmarketcommitmentsettlerequest).

### QueryCommitmentSettlementFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L408-L435


## ValidateCreateMarket

It's possible for a [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest) to result in a market setup that is problematic.
To verify that one is not problematic, this `ValidateCreateMarket` can be used.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateCreateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L437-L441

See also: [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest).

### QueryValidateCreateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L443-L453


## ValidateMarket

An existing market's setup can be checked for problems using the `ValidateMarket` query.

Any problems detected will be returned in the `error` field.

### QueryValidateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L455-L459

### QueryValidateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L461-L465


## ValidateManageFees

It's possible for a [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest) to result in a problematic setup for a market.
To verify that one does not result in such a state, use this `ValidateManageFees` query.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateManageFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L467-L471

See also: [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest).

### QueryValidateManageFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L473-L483


## GetPayment

Use the `GetPayment` query to look up a payment by `source` and `external_id`.

### QueryGetPaymentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L485-L491

### QueryGetPaymentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L493-L497

See also: [Payment](03_messages.md#payment).


## GetPaymentsWithSource

To get all payments with a specific `source`, use the `GetPaymentsWithSource` query.

This query is paginated.

### QueryGetPaymentsWithSourceRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L499-L506

### QueryGetPaymentsWithSourceResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L508-L515

See also: [Payment](03_messages.md#payment).


## GetPaymentsWithTarget

To get all payments with a specific `target`, use the `GetPaymentsWithTarget` query.

This query is paginated.

### QueryGetPaymentsWithTargetRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L517-L524

### QueryGetPaymentsWithTargetResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L526-L533

See also: [Payment](03_messages.md#payment).


## GetAllPayments

A listing of all existing payments can be found using the `GetAllPayments` query.

This query is paginated.

### QueryGetAllPaymentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L535-L539

### QueryGetAllPaymentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L541-L548

See also: [Payment](03_messages.md#payment).


## PaymentFeeCalc

The `PaymentFeeCalc` query can be used to calculate the fees for creating or accepting a payment.

### QueryPaymentFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L550-L554

See also: [Payment](03_messages.md#payment).

### QueryPaymentFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.20.0/proto/provenance/exchange/v1/query.proto#L556-L572
