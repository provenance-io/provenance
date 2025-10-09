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

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L175-L182

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).

### QueryOrderFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L184-L203


## GetOrder

Use the `GetOrder` query to look up an order by its id.

### QueryGetOrderRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L205-L209

### QueryGetOrderResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L211-L215

### Order

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/orders.proto#L15-L28

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).


## GetOrderByExternalID

Orders with external ids can be looked up using the `GetOrderByExternalID` query.

### QueryGetOrderByExternalIDRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L217-L223

### QueryGetOrderByExternalIDResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L225-L229

See also: [Order](#order).


## GetMarketOrders

To get all of the orders in a given market, use the `GetMarketOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetMarketOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L231-L242

### QueryGetMarketOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L244-L251

See also: [Order](#order).


## GetOwnerOrders

To get all of the orders with a specific owner (e.g. buyer or seller), use the `GetOwnerOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetOwnerOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L253-L264

### QueryGetOwnerOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L266-L273

See also: [Order](#order).


## GetAssetOrders

To get all of the orders with a specific asset denom, use the `GetAssetOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetAssetOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L275-L286

### QueryGetAssetOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L288-L295

See also: [Order](#order).


## GetAllOrders

To get all existing orders, use the `GetAllOrders` query.

This query is paginated.

### QueryGetAllOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L297-L301

### QueryGetAllOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L303-L310

See also: [Order](#order).


## GetCommitment

To find out how much an account has committed to a market, use the `GetCommitment` query.

### QueryGetCommitmentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L312-L318

### QueryGetCommitmentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L320-L329


## GetAccountCommitments

To look up the amounts an account has committed to any market, use the `GetAccountCommitments` query.
You can optionally filter the results for a specific denomination using the `denom` query parameter.

### QueryGetAccountCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L331-L337

### QueryGetAccountCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L339-L343


## GetMarketCommitments

To get the amounts committed to a market by any account, use the `GetMarketCommitments` query.

### QueryGetMarketCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L345-L352

### QueryGetMarketCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L354-L361


## GetAllCommitments

To get all funds committed by any account to any market, use the `GetAllCommitments` query.

### QueryGetAllCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L363-L367

### QueryGetAllCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L369-L376


## GetMarket

All the information and setup for a market can be looked up using the `GetMarket` query.

### QueryGetMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L378-L382

### QueryGetMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L384-L390

See also: [Market](03_messages.md#market).


## GetAllMarkets

Use the `GetAllMarkets` query to get brief information about all markets.

### QueryGetAllMarketsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L392-L396

### QueryGetAllMarketsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L398-L405

### MarketBrief

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/market.proto#L42-L50


## Params

The exchange module params can be looked up using the `Params` query.

### QueryParamsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L407-L408

### QueryParamsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L410-L414

See also: [Params](06_params.md#params).


## CommitmentSettlementFeeCalc

To find out the additional tx fee required for a commitment settlement, use the `CommitmentSettlementFeeCalc` query.

### QueryCommitmentSettlementFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L416-L426

See also: [MsgMarketCommitmentSettleRequest](03_messages.md#msgmarketcommitmentsettlerequest).

### QueryCommitmentSettlementFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L428-L455


## ValidateCreateMarket

It's possible for a [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest) to result in a market setup that is problematic.
To verify that one is not problematic, this `ValidateCreateMarket` can be used.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateCreateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L457-L461

See also: [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest).

### QueryValidateCreateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L463-L473


## ValidateMarket

An existing market's setup can be checked for problems using the `ValidateMarket` query.

Any problems detected will be returned in the `error` field.

### QueryValidateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L475-L479

### QueryValidateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L481-L485


## ValidateManageFees

It's possible for a [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest) to result in a problematic setup for a market.
To verify that one does not result in such a state, use this `ValidateManageFees` query.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateManageFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L487-L491

See also: [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest).

### QueryValidateManageFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L493-L503


## GetPayment

Use the `GetPayment` query to look up a payment by `source` and `external_id`.

### QueryGetPaymentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L505-L511

### QueryGetPaymentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L513-L517

See also: [Payment](03_messages.md#payment).


## GetPaymentsWithSource

To get all payments with a specific `source`, use the `GetPaymentsWithSource` query.

This query is paginated.

### QueryGetPaymentsWithSourceRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L519-L526

### QueryGetPaymentsWithSourceResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L528-L535

See also: [Payment](03_messages.md#payment).


## GetPaymentsWithTarget

To get all payments with a specific `target`, use the `GetPaymentsWithTarget` query.

This query is paginated.

### QueryGetPaymentsWithTargetRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L537-L544

### QueryGetPaymentsWithTargetResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L546-L553

See also: [Payment](03_messages.md#payment).


## GetAllPayments

A listing of all existing payments can be found using the `GetAllPayments` query.

This query is paginated.

### QueryGetAllPaymentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L555-L559

### QueryGetAllPaymentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L561-L568

See also: [Payment](03_messages.md#payment).


## PaymentFeeCalc

The `PaymentFeeCalc` query can be used to calculate the fees for creating or accepting a payment.

### QueryPaymentFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L570-L574

See also: [Payment](03_messages.md#payment).

### QueryPaymentFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/exchange/v1/query.proto#L576-L592
