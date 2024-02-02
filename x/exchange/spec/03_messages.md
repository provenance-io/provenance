# Exchange Messages

The exchange module has `Msg` endpoints for users, markets, and governance proposals.

---
<!-- TOC -->
  - [User Endpoints](#user-endpoints)
    - [CreateAsk](#createask)
    - [CreateBid](#createbid)
    - [CommitFunds](#commitfunds)
    - [CancelOrder](#cancelorder)
    - [FillBids](#fillbids)
    - [FillAsks](#fillasks)
  - [Market Endpoints](#market-endpoints)
    - [MarketSettle](#marketsettle)
    - [MarketCommitmentSettle](#marketcommitmentsettle)
    - [MarketReleaseCommitments](#marketreleasecommitments)
    - [MarketSetOrderExternalID](#marketsetorderexternalid)
    - [MarketWithdraw](#marketwithdraw)
    - [MarketUpdateDetails](#marketupdatedetails)
    - [MarketUpdateAcceptingOrders](#marketupdateacceptingorders)
    - [MarketUpdateUserSettle](#marketupdateusersettle)
    - [MarketUpdateAcceptingCommitments](#marketupdateacceptingcommitments)
    - [MarketUpdateIntermediaryDenom](#marketupdateintermediarydenom)
    - [MarketManagePermissions](#marketmanagepermissions)
    - [MarketManageReqAttrs](#marketmanagereqattrs)
  - [Governance Proposals](#governance-proposals)
    - [GovCreateMarket](#govcreatemarket)
    - [GovManageFees](#govmanagefees)
    - [GovCloseMarket](#govclosemarket)
    - [GovUpdateParams](#govupdateparams)


## User Endpoints

There are several endpoints available for all users, but some markets might have restrictions on their use.


### CreateAsk

An ask order indicates the desire to sell some `assets` at a minimum `price`.
They are created using the `CreateAsk` endpoint.

Markets can define a set of attributes that an account must have in order to create ask orders in them.
So, this endpoint might not be available, depending on the `seller` and the `market_id`.
Markets can also disable order creation altogether, making this endpoint unavailable for that `market_id`.

It is expected to fail if:
* The `market_id` does not exist.
* The market is not allowing orders to be created.
* The market requires attributes in order to create ask orders and the `seller` is missing one or more.
* The `assets` are not in the `seller`'s account.
* The `price` is in a denom not supported by the market.
* The `seller_settlement_flat_fee` is in a denom different from the `price`, and is not in the `seller`'s account.
* The `seller_settlement_flat_fee` is insufficient (as dictated by the market).
* The `external_id` value is not empty and is already in use in the market.
* The `order_creation_fee` is not in the `seller`'s account.

#### MsgCreateAskRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L97-L105

#### AskOrder

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/orders.proto#L28-L53

#### MsgCreateAskResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L107-L111


### CreateBid

A bid order indicates the desire to buy some `assets` at a specific `price`.
They are created using the `CreateBid` endpoint.

Markets can define a set of attributes that an account must have in order to create bid orders in them.
So, this endpoint might not be available, depending on the `buyer` and the `market_id`.
Markets can also disable order creation altogether, making this endpoint unavailable for that `market_id`.

It is expected to fail if:
* The `market_id` does not exist.
* The market is not allowing orders to be created.
* The market requires attributes in order to create bid orders and the `buyer` is missing one or more.
* The `price` funds are not in the `buyer`'s account.
* The `price` is in a denom not supported by the market.
* The `buyer_settlement_fees` are not in the `buyer`'s account.
* The `buyer_settlement_fees` are insufficient (as dictated by the market).
* The `external_id` value is not empty and is already in use in the market.
* The `order_creation_fee` is not in the `buyer`'s account.

#### MsgCreateBidRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L113-L121

#### BidOrder

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/orders.proto#L55-L78

#### MsgCreateBidResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L123-L127


### CommitFunds

Funds can be committed to a market using the `CommitFunds` endpoint.
If the account already has funds committed to the market, the provided funds are added to that commitment amount.

It is expected to fail if:
* The market does not exist.
* The market is not accepting commitments.
* The market requires attributes in order to create commitments and the `account` is missing one or more.
* The `creation_fee` is insufficient (as dictated by the market).
* The `amount` is not spendable in the account (after paying the creation fee).

#### MsgCommitFundsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L129-L144

#### MsgCommitFundsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L146-L147


### CancelOrder

Orders can be cancelled using the `CancelOrder` endpoint.
When an order is cancelled, the hold on its funds is released and the order is deleted.

Users can cancel their own orders at any time.
Market actors with the `PERMISSION_CANCEL` permission can also cancel orders in that market at any time.

Order creation fees are **not** refunded when an order is cancelled.

It is expected to fail if:
* The order does not exist.
* The `signer` is not one of:
  * The order's owner (e.g. `buyer` or `seller`).
  * An account with `PERMISSION_CANCEL` in the order's market.
  * The governance module account (`authority`).

#### MsgCancelOrderRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L149-L159

#### MsgCancelOrderResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L161-L162


### FillBids

If a market allows user-settlement, users can use the `FillBids` endpoint to settle one or more bids with their own `assets`.
This is similar to an "Immediate or cancel" `AskOrder` with the sum of the provided bids' assets and prices.
Fees are paid the same as if an `AskOrder` were actually created and settled normally with the provided bids.
The `seller` must be allowed to create an `AskOrder` in the given market.

It is expected to fail if:
* The market does not exist.
* The market is not allowing orders to be created.
* The market does not allow user-settlement.
* The market requires attributes in order to create ask orders and the `seller` is missing one or more.
* One or more `bid_order_ids` are not bid orders (or do not exist).
* One or more `bid_order_ids` are in a market other than the provided `market_id`.
* The `total_assets` are not in the `seller`'s account.
* The sum of bid order `assets` does not equal the provided `total_assets`.
* The `seller` or one of the `buyer`s are sanctioned, or are not allowed to possess the funds they are to receive.
* The `seller_settlement_flat_fee` is insufficient.
* The `seller_settlement_flat_fee` is not in the `seller`'s account (after `assets` and `price` funds have been transferred).
* The `ask_order_creation_fee` is insufficient.
* The `ask_order_creation_fee` is not in the `seller`'s account (after all other transfers have been made).

#### MsgFillBidsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L164-L184

#### MsgFillBidsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L186-L187


### FillAsks

If a market allows user-settlement, users can use the `FillAsks` endpoint to settle one or more asks with their own price funds.
This is similar to an "Immediate or cancel" `BidOrder` with the sum of the provided asks' assets and prices.
Fees are paid the same as if a `BidOrder` were actually created and settled normally with the provided asks.
The `buyer` must be allowed to create a `BidOrder` in the given market.

It is expected to fail if:
* The market does not exist.
* The market is not allowing orders to be created.
* The market does not allow user-settlement.
* The market requires attributes in order to create bid orders and the `buyer` is missing one or more.
* One or more `ask_order_ids` are not ask orders (or do not exist).
* One or more `ask_order_ids` are in a market other than the provided `market_id`.
* The `total_price` funds are not in the `buyer`'s account.
* The sum of ask order `price`s does not equal the provided `total_price`.
* The `buyer` or one of the `seller`s are sanctioned, or are not allowed to possess the funds they are to receive.
* The `buyer_settlement_fees` are insufficient.
* The `buyer_settlement_fees` are not in the `buyer`'s account (after `assets` and `price` funds have been transferred).
* The `bid_order_creation_fee` is insufficient.
* The `bid_order_creation_fee` is not in the `buyer`'s account (after all other transfers have been made).

#### MsgFillAsksRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L189-L210

#### MsgFillAsksResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L212-L213


## Market Endpoints

Several endpoints are only available to accounts designated by the market.
These are all also available for use in governance proposals using the governance module account (aka `authority`) as the `admin`.


### MarketSettle

Orders are settled using the `MarketSettle` endpoint.
The `admin` must have the `PERMISSION_SETTLE` permission in the market (or be the `authority`).

The market is responsible for identifying order matches.
Once identified, this endpoint is used to settle and clear the matched orders.

All orders in a settlement must have the same asset denom and the same price denom.

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_SETTLE` in the market, and is not the `authority`.
* One or more `ask_order_ids` are not ask orders, or do not exist, or are in a market other than the provided `market_id`.
* One or more `bid_order_ids` are not bid orders, or do not exist, or are in a market other than the provided `market_id`.
* There is more than one denom in the `assets` of all the provided orders.
* There is more than one denom in the `price` of all the provided orders.
* The market requires a seller settlement ratio fee, but there is no ratio defined for the `price` denom.
* Two or more orders are being partially filled.
* One or more orders cannot be filled at all with the `assets` or `price` funds available in the settlement.
* An order is being partially filled, but `expect_partial` is `false`.
* All orders are being filled in full, but `expect_partial` is `true`.
* One or more of the `buyer`s and `seller`s are sanctioned, or are not allowed to possess the funds they are to receive.

#### MsgMarketSettleRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L215-L232

#### MsgMarketSettleResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L234-L235


### MarketCommitmentSettle

A market can move committed funds using the `MarketCommitmentSettle` endpoint.
The `admin` must have the `PERMISSION_SETTLE` permission in the market (or be the `authority`).

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_SETTLE` in the market, and is not the `authority`.
* The sum of the `inputs` does not equal the sum of the `outputs`.
* Not enough funds have been committed by one or more accounts to the market.
* A NAV is needed (for fee calculation) that does not exist and was not provided.

#### MsgMarketCommitmentSettleRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L237-L257

#### MsgMarketCommitmentSettleResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L259-L260


### MarketReleaseCommitments

A market can release committed funds using the `MarketReleaseCommitments` endpoint.
The `admin` must have the `PERMISSION_CANCEL` permission in the market (or be the `authority`).

Providing an empty amount indicates that all funds currently committed in that account (to the market) should be released.

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_CANCEL` in the market, and is not the `authority`.
* One or more of the amounts is more than what is currently committed by the associated `account`.

#### MsgMarketReleaseCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L262-L275

#### MsgMarketReleaseCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L277-L278


### MarketSetOrderExternalID

Some markets might want to attach their own identifiers to orders.
This is done using the `MarketSetOrderExternalID` endpoint.
The `admin` must have the `PERMISSION_SET_IDS` permission in the market (or be the `authority`).

Orders with external ids can be looked up using the [GetOrderByExternalID](05_queries.md#getorderbyexternalid) query.

External ids must be unique in a market, but multiple markets can use the same external id.

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_SET_IDS` in the market, and is not the `authority`.
* The order does not exist, or is in a different market than the provided `market_id`.
* The provided `external_id` equals the order's current `external_id`.
* The provided `external_id` is already associated with another order in the same market.

#### MsgMarketSetOrderExternalIDRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L280-L294

#### MsgMarketSetOrderExternalIDResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L296-L297


### MarketWithdraw

When fees are collected by a market, they are given to the market's account.
Those funds can then be withdrawn/transferred using the `MarketWithdraw` endpoint.
The `admin` must have the `PERMISSION_WITHDRAW` permission in the market (or be the `authority`).

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_WITHDRAW` in the market, and is not the `authority`.
* The `amount` funds are not in the market's account.
* The `to_address` is not allowed to possess the requested funds.

#### MsgMarketWithdrawRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L299-L313

#### MsgMarketWithdrawResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L315-L316


### MarketUpdateDetails

A market's details can be updated using the `MarketUpdateDetails` endpoint.
The `admin` must have the `PERMISSION_UPDATE` permission in the market (or be the `authority`).

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_UPDATE` in the market, and is not the `authority`.
* One or more of the [MarketDetails](#marketdetails) fields is too large.

#### MsgMarketUpdateDetailsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L318-L329

See also: [MarketDetails](#marketdetails).

#### MsgMarketUpdateDetailsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L331-L332


### MarketUpdateAcceptingOrders

A market can enable or disable order creation using the `MarketUpdateAcceptingOrders` endpoint.
The `admin` must have the `PERMISSION_UPDATE` permission in the market (or be the `authority`).

With `accepting_orders` = `false`, no one can create any new orders in the market, but existing orders can still be settled or cancelled.

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_UPDATE` in the market, and is not the `authority`.
* The provided `accepting_orders` value equals the market's current setting.

#### MsgMarketUpdateAcceptingOrdersRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L358-L369

#### MsgMarketUpdateAcceptingOrdersResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L371-L372


### MarketUpdateUserSettle

Using the `MarketUpdateUserSettle` endpoint, markets can control whether user-settlement is allowed.
The `admin` must have the `PERMISSION_UPDATE` permission in the market (or be the `authority`).

The [FillBids](#fillbids) and [FillAsks](#fillasks) endpoints are only available for markets where `allow_user_settlement` = `true`.
The [MarketSettle](#marketsettle) endpoint is usable regardless of this setting.

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_UPDATE` in the market, and is not the `authority`.
* The provided `allow_user_settlement` value equals the market's current setting.

#### MsgMarketUpdateUserSettleRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L374-L387

#### MsgMarketUpdateUserSettleResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L389-L390


### MarketUpdateAcceptingCommitments

Using the `MarketUpdateAcceptingCommitments` endpoint, a market can control whether it is accepting commitments.
The `admin` must have the `PERMISSION_UPDATE` permission in the market (or be the `authority`).

The [CommitFunds](#CommitFunds) endpoint is only available for markets where `accepting_orders` = `true`.

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_UPDATE` in the market, and is not the `authority`.
* The provided `accepting_orders` value equals the market's current setting.
* The provided `accepting_orders` is `true` but no commitment-related fees are defined.
* The provided `accepting_orders` is `true` and bips are set, but either no intermediary denom is defined or there is no NAV associating the intermediary denom with the chain's fee denom.

#### MsgMarketUpdateAcceptingCommitmentsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L392-L405

#### MsgMarketUpdateAcceptingCommitmentsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L407-L408


### MarketUpdateIntermediaryDenom

The `MarketUpdateIntermediaryDenom` endpoint allows a market to change its intermediary denom (used for commitment settlement fee calculation).
The `admin` must have the `PERMISSION_UPDATE` permission in the market (or be the `authority`).

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_UPDATE` in the market, and is not the `authority`.
* The provided `intermediary_denom` is not a valid denom string.

#### MsgMarketUpdateIntermediaryDenomRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L410-L421

#### MsgMarketUpdateIntermediaryDenomResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L423-L424


### MarketManagePermissions

Permissions in a market are managed using the `MarketManagePermissions` endpoint.
The `admin` must have the `PERMISSION_PERMISSIONS` permission in the market (or be the `authority`).

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_PERMISSIONS` in the market, and is not the `authority`.
* One or more `revoke_all` addresses do not currently have any permissions in the market.
* One or more `to_revoke` entries do not currently exist in the market.
* One or more `to_grant` entries already exist in the market (after `revoke_all` and `to_revoke` are processed).

#### MsgMarketManagePermissionsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L426-L441

See also: [AccessGrant](#accessgrant) and [Permission](#permission).

#### MsgMarketManagePermissionsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L443-L444


### MarketManageReqAttrs

The attributes required to create orders in a market can be managed using the `MarketManageReqAttrs` endpoint.
The `admin` must have the `PERMISSION_ATTRIBUTES` permission in the market (or be the `authority`).

See also: [Required Attributes](01_concepts.md#required-attributes).

It is expected to fail if:
* The market does not exist.
* The `admin` does not have `PERMISSION_ATTRIBUTES` in the market, and is not the `authority`.
* One or more attributes to add are already required by the market (for the given order type).
* One or more attributes to remove are not currently required by the market (for the given order type).

#### MsgMarketManageReqAttrsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#446-L467

#### MsgMarketManageReqAttrsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L469-L470


## Governance Proposals

There are several governance-proposal-only endpoints.


### GovCreateMarket

Market creation must be done via governance proposal with a `MsgGovCreateMarketRequest`.

If the provided `market_id` is `0` (zero), the next available market id will be assigned to the new market.
If it is not zero, the provided `market_id` will be used (unless it's already in use by another market).
If it's already in use, the proposal will fail.

It is recommended that the message be checked using the [ValidateCreateMarket](05_queries.md#validatecreatemarket) query first, to reduce the risk of failure or problems.

It is expected to fail if:
* The provided `authority` is not the governance module's account.
* The provided `market_id` is not zero, and is already in use by another market.
* One or more of the [MarketDetails](#marketdetails) fields is too large.
* One or more required attributes are invalid.

#### MsgGovCreateMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L472-L483

#### Market

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/market.proto#L52-L148

#### MarketDetails

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/market.proto#L28-L40

* The `name` is limited to 250 characters max.
* The `description` is limited to 2000 characters max.
* The `website_url` is limited to 200 characters max.
* The `icon_uri` is limited to 2000 characters max.

#### FeeRatio

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/market.proto#L150-L158

#### AccessGrant

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/market.proto#L160-L166

#### Permission

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/market.proto#L168-L186

#### MsgGovCreateMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L485-L486


### GovManageFees

A market's fees can only be altered via governance proposal with a `MsgGovManageFeesRequest`.

It is recommended that the message be checked using the [ValidateManageFees](05_queries.md#validatemanagefees) query first, to ensure the updated fees do not present any problems.

It is expected to fail if:
* The provided `authority` is not the governance module's account.

#### MsgGovManageFeesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L488-L538

See also: [FeeRatio](#feeratio).

#### MsgGovManageFeesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L540-L541


### GovCloseMarket

A market can be closed via governance proposal with a `MsgGovCloseMarketRequest`.

When a market is closed, it stops accepting orders and commitments, all orders are cancelled, and all commitments are released.

It is expected to fail if:
* The provided `authority` is not the governance module's account.

#### MsgGovCloseMarketRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L543-L551

#### MsgGovCloseMarketResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L553-L554


### GovUpdateParams

The exchange module params are updated via governance proposal with a `MsgGovUpdateParamsRequest`.

It is expected to fail if:
* The provided `authority` is not the governance module's account.

#### MsgGovUpdateParamsRequest

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L556-L565

See also: [Params](06_params.md#params).

#### MsgGovUpdateParamsResponse

+++ https://github.com/provenance-io/provenance/blob/v1.18.0/proto/provenance/exchange/v1/tx.proto#L567-L568
