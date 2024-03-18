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

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L156-L163

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).

### QueryOrderFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L165-L184


## GetOrder

Use the `GetOrder` query to look up an order by its id.

### QueryGetOrderRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L186-L190

### QueryGetOrderResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L192-L196

### Order

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/orders.proto#L13-L26

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).


## GetOrderByExternalID

Orders with external ids can be looked up using the `GetOrderByExternalID` query.

### QueryGetOrderByExternalIDRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L198-L204

### QueryGetOrderByExternalIDResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L206-L210

See also: [Order](#order).


## GetMarketOrders

To get all of the orders in a given market, use the `GetMarketOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetMarketOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L212-L223

### QueryGetMarketOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L225-L232

See also: [Order](#order).


## GetOwnerOrders

To get all of the orders with a specific owner (e.g. buyer or seller), use the `GetOwnerOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetOwnerOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L234-L245

### QueryGetOwnerOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L247-L254

See also: [Order](#order).


## GetAssetOrders

To get all of the orders with a specific asset denom, use the `GetAssetOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetAssetOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L256-L267

### QueryGetAssetOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L269-L276

See also: [Order](#order).


## GetAllOrders

To get all existing orders, use the `GetAllOrders` query.

This query is paginated.

### QueryGetAllOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L278-L282

### QueryGetAllOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L284-L291

See also: [Order](#order).


## GetCommitment

To find out how much an account has committed to a market, use the `GetCommitment` query.

### QueryGetCommitmentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L293-L299

### QueryGetCommitmentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L301-L306


## GetAccountCommitments

To look up the amounts an account has committed to any market, use the `GetAccountCommitments` query.

### QueryGetAccountCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L308-L312

### QueryGetAccountCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L314-L318


## GetMarketCommitments

To get the amounts committed to a market by any account, use the `GetMarketCommitments` query.

### QueryGetMarketCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L320-L327

### QueryGetMarketCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L329-L336


## GetAllCommitments

To get all funds committed by any account to any market, use the `GetAllCommitments` query.

### QueryGetAllCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L338-L342

### QueryGetAllCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L344-L351


## GetMarket

All the information and setup for a market can be looked up using the `GetMarket` query.

### QueryGetMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L353-L357

### QueryGetMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L359-L365

See also: [Market](03_messages.md#market).


## GetAllMarkets

Use the `GetAllMarkets` query to get brief information about all markets.

### QueryGetAllMarketsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L367-L371

### QueryGetAllMarketsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L373-L380

### MarketBrief

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/market.proto#L42-L50


## Params

The exchange module params can be looked up using the `Params` query.

### QueryParamsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L382-L383

### QueryParamsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L385-L389

See also: [Params](06_params.md#params).


## CommitmentSettlementFeeCalc

To find out the additional tx fee required for a commitment settlement, use the `CommitmentSettlementFeeCalc` query.

### QueryCommitmentSettlementFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L391-L401

See also: [MsgMarketCommitmentSettleRequest](03_messages.md#msgmarketcommitmentsettlerequest).

### QueryCommitmentSettlementFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L403-L418


## ValidateCreateMarket

It's possible for a [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest) to result in a market setup that is problematic.
To verify that one is not problematic, this `ValidateCreateMarket` can be used.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateCreateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L420-L424

See also: [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest).

### QueryValidateCreateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L426-L436


## ValidateMarket

An existing market's setup can be checked for problems using the `ValidateMarket` query.

Any problems detected will be returned in the `error` field.

### QueryValidateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L438-L442

### QueryValidateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L444-L448


## ValidateManageFees

It's possible for a [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest) to result in a problematic setup for a market.
To verify that one does not result in such a state, use this `ValidateManageFees` query.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateManageFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L450-L454

See also: [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest).

### QueryValidateManageFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L456-L466


## GetPayment

Use the `GetPayment` query to look up a payment by `source` and `external_id`.

### QueryGetPaymentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L468-L474

### QueryGetPaymentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L476-L480

See also: [Payment](03_messages.md#payment).


## GetPaymentsWithSource

To get all payments with a specific `source`, use the `GetPaymentsWithSource` query.

This query is paginated.

### QueryGetPaymentsWithSourceRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L482-L489

### QueryGetPaymentsWithSourceResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L491-L498

See also: [Payment](03_messages.md#payment).


## GetPaymentsWithTarget

To get all payments with a specific `target`, use the `GetPaymentsWithTarget` query.

This query is paginated.

### QueryGetPaymentsWithTargetRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L500-L507

### QueryGetPaymentsWithTargetResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L509-L516

See also: [Payment](03_messages.md#payment).


## GetAllPayments

A listing of all existing payments can be found using the `GetAllPayments` query.

This query is paginated.

### QueryGetAllPaymentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L518-L522

### QueryGetAllPaymentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L524-L531

See also: [Payment](03_messages.md#payment).


## PaymentFeeCalc

The `PaymentFeeCalc` query can be used to calculate the fees for creating or accepting a payment.

### QueryPaymentFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L533-L537

See also: [Payment](03_messages.md#payment).

### QueryPaymentFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/query.proto#L539-L547
