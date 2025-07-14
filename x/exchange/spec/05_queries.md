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

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L158-L165

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).

### QueryOrderFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L167-L186


## GetOrder

Use the `GetOrder` query to look up an order by its id.

### QueryGetOrderRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L188-L192

### QueryGetOrderResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L194-L198

### Order

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/orders.proto#L15-L28

See also: [AskOrder](03_messages.md#askorder), and [BidOrder](03_messages.md#bidorder).


## GetOrderByExternalID

Orders with external ids can be looked up using the `GetOrderByExternalID` query.

### QueryGetOrderByExternalIDRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L200-L206

### QueryGetOrderByExternalIDResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L208-L212

See also: [Order](#order).


## GetMarketOrders

To get all of the orders in a given market, use the `GetMarketOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetMarketOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L214-L225

### QueryGetMarketOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L227-L234

See also: [Order](#order).


## GetOwnerOrders

To get all of the orders with a specific owner (e.g. buyer or seller), use the `GetOwnerOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetOwnerOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L236-L247

### QueryGetOwnerOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L249-L256

See also: [Order](#order).


## GetAssetOrders

To get all of the orders with a specific asset denom, use the `GetAssetOrders` query.
Results can be optionally limited by order type (e.g. "ask" or "bid") and/or a minimum (exclusive) order id.

This query is paginated.

### QueryGetAssetOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L258-L269

### QueryGetAssetOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L271-L278

See also: [Order](#order).


## GetAllOrders

To get all existing orders, use the `GetAllOrders` query.

This query is paginated.

### QueryGetAllOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L280-L284

### QueryGetAllOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L286-L293

See also: [Order](#order).


## GetCommitment

To find out how much an account has committed to a market, use the `GetCommitment` query.

### QueryGetCommitmentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L295-L301

### QueryGetCommitmentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L303-L312


## GetAccountCommitments

To look up the amounts an account has committed to any market, use the `GetAccountCommitments` query.
You can optionally filter the results for a specific denomination using the `denom` query parameter.

### QueryGetAccountCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L314-L320

### QueryGetAccountCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L322-L326


## GetMarketCommitments

To get the amounts committed to a market by any account, use the `GetMarketCommitments` query.

### QueryGetMarketCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L328-L335

### QueryGetMarketCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L337-L344


## GetAllCommitments

To get all funds committed by any account to any market, use the `GetAllCommitments` query.

### QueryGetAllCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L346-L350

### QueryGetAllCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L352-L359


## GetMarket

All the information and setup for a market can be looked up using the `GetMarket` query.

### QueryGetMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L361-L365

### QueryGetMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L367-L373

See also: [Market](03_messages.md#market).


## GetAllMarkets

Use the `GetAllMarkets` query to get brief information about all markets.

### QueryGetAllMarketsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L375-L379

### QueryGetAllMarketsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L381-L388

### MarketBrief

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/market.proto#L42-L50


## Params

The exchange module params can be looked up using the `Params` query.

### QueryParamsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L390-L391

### QueryParamsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L393-L397

See also: [Params](06_params.md#params).


## CommitmentSettlementFeeCalc

To find out the additional tx fee required for a commitment settlement, use the `CommitmentSettlementFeeCalc` query.

### QueryCommitmentSettlementFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L399-L409

See also: [MsgMarketCommitmentSettleRequest](03_messages.md#msgmarketcommitmentsettlerequest).

### QueryCommitmentSettlementFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L411-L438


## ValidateCreateMarket

It's possible for a [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest) to result in a market setup that is problematic.
To verify that one is not problematic, this `ValidateCreateMarket` can be used.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateCreateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L440-L444

See also: [MsgGovCreateMarketRequest](03_messages.md#msggovcreatemarketrequest).

### QueryValidateCreateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L446-L456


## ValidateMarket

An existing market's setup can be checked for problems using the `ValidateMarket` query.

Any problems detected will be returned in the `error` field.

### QueryValidateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L458-L462

### QueryValidateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L464-L468


## ValidateManageFees

It's possible for a [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest) to result in a problematic setup for a market.
To verify that one does not result in such a state, use this `ValidateManageFees` query.

If the result has:
* `gov_prop_will_pass` = `false`, then either submitting the proposal will fail, or the `Msg` will result in an error ("failed") after the proposal is passed. The `error` field will have details.
* `gov_prop_will_pass` = `true` and a non-empty `error` field, then the `Msg` would successfully run, but would result in the problems identified in the `error` field.
* `gov_prop_will_pass` = `true` and an empty `error` field, then there are no problems with the provided `Msg`.

### QueryValidateManageFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L470-L474

See also: [MsgGovManageFeesRequest](03_messages.md#msggovmanagefeesrequest).

### QueryValidateManageFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L476-L486


## GetPayment

Use the `GetPayment` query to look up a payment by `source` and `external_id`.

### QueryGetPaymentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L488-L494

### QueryGetPaymentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L496-L500

See also: [Payment](03_messages.md#payment).


## GetPaymentsWithSource

To get all payments with a specific `source`, use the `GetPaymentsWithSource` query.

This query is paginated.

### QueryGetPaymentsWithSourceRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L502-L509

### QueryGetPaymentsWithSourceResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L511-L518

See also: [Payment](03_messages.md#payment).


## GetPaymentsWithTarget

To get all payments with a specific `target`, use the `GetPaymentsWithTarget` query.

This query is paginated.

### QueryGetPaymentsWithTargetRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L520-L527

### QueryGetPaymentsWithTargetResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L529-L536

See also: [Payment](03_messages.md#payment).


## GetAllPayments

A listing of all existing payments can be found using the `GetAllPayments` query.

This query is paginated.

### QueryGetAllPaymentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L538-L542

### QueryGetAllPaymentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L544-L551

See also: [Payment](03_messages.md#payment).


## PaymentFeeCalc

The `PaymentFeeCalc` query can be used to calculate the fees for creating or accepting a payment.

### QueryPaymentFeeCalcRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L553-L557

See also: [Payment](03_messages.md#payment).

### QueryPaymentFeeCalcResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/exchange/v1/query.proto#L559-L575
