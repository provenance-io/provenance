<!-- This file is auto-generated. Please do not modify it yourself. -->
# Provenance API Documentation
<a name="top"></a>

## Table of Contents

- [cosmos/quarantine/v1beta1/tx.proto](#cosmos_quarantine_v1beta1_tx-proto)
    - [MsgAccept](#cosmos-quarantine-v1beta1-MsgAccept)
    - [MsgAcceptResponse](#cosmos-quarantine-v1beta1-MsgAcceptResponse)
    - [MsgDecline](#cosmos-quarantine-v1beta1-MsgDecline)
    - [MsgDeclineResponse](#cosmos-quarantine-v1beta1-MsgDeclineResponse)
    - [MsgOptIn](#cosmos-quarantine-v1beta1-MsgOptIn)
    - [MsgOptInResponse](#cosmos-quarantine-v1beta1-MsgOptInResponse)
    - [MsgOptOut](#cosmos-quarantine-v1beta1-MsgOptOut)
    - [MsgOptOutResponse](#cosmos-quarantine-v1beta1-MsgOptOutResponse)
    - [MsgUpdateAutoResponses](#cosmos-quarantine-v1beta1-MsgUpdateAutoResponses)
    - [MsgUpdateAutoResponsesResponse](#cosmos-quarantine-v1beta1-MsgUpdateAutoResponsesResponse)
  
    - [Msg](#cosmos-quarantine-v1beta1-Msg)
  
- [cosmos/quarantine/v1beta1/events.proto](#cosmos_quarantine_v1beta1_events-proto)
    - [EventFundsQuarantined](#cosmos-quarantine-v1beta1-EventFundsQuarantined)
    - [EventFundsReleased](#cosmos-quarantine-v1beta1-EventFundsReleased)
    - [EventOptIn](#cosmos-quarantine-v1beta1-EventOptIn)
    - [EventOptOut](#cosmos-quarantine-v1beta1-EventOptOut)
  
- [cosmos/quarantine/v1beta1/query.proto](#cosmos_quarantine_v1beta1_query-proto)
    - [QueryAutoResponsesRequest](#cosmos-quarantine-v1beta1-QueryAutoResponsesRequest)
    - [QueryAutoResponsesResponse](#cosmos-quarantine-v1beta1-QueryAutoResponsesResponse)
    - [QueryIsQuarantinedRequest](#cosmos-quarantine-v1beta1-QueryIsQuarantinedRequest)
    - [QueryIsQuarantinedResponse](#cosmos-quarantine-v1beta1-QueryIsQuarantinedResponse)
    - [QueryQuarantinedFundsRequest](#cosmos-quarantine-v1beta1-QueryQuarantinedFundsRequest)
    - [QueryQuarantinedFundsResponse](#cosmos-quarantine-v1beta1-QueryQuarantinedFundsResponse)
  
    - [Query](#cosmos-quarantine-v1beta1-Query)
  
- [cosmos/quarantine/v1beta1/quarantine.proto](#cosmos_quarantine_v1beta1_quarantine-proto)
    - [AutoResponseEntry](#cosmos-quarantine-v1beta1-AutoResponseEntry)
    - [AutoResponseUpdate](#cosmos-quarantine-v1beta1-AutoResponseUpdate)
    - [QuarantineRecord](#cosmos-quarantine-v1beta1-QuarantineRecord)
    - [QuarantineRecordSuffixIndex](#cosmos-quarantine-v1beta1-QuarantineRecordSuffixIndex)
    - [QuarantinedFunds](#cosmos-quarantine-v1beta1-QuarantinedFunds)
  
    - [AutoResponse](#cosmos-quarantine-v1beta1-AutoResponse)
  
- [cosmos/quarantine/v1beta1/genesis.proto](#cosmos_quarantine_v1beta1_genesis-proto)
    - [GenesisState](#cosmos-quarantine-v1beta1-GenesisState)
  
- [cosmos/sanction/v1beta1/tx.proto](#cosmos_sanction_v1beta1_tx-proto)
    - [MsgSanction](#cosmos-sanction-v1beta1-MsgSanction)
    - [MsgSanctionResponse](#cosmos-sanction-v1beta1-MsgSanctionResponse)
    - [MsgUnsanction](#cosmos-sanction-v1beta1-MsgUnsanction)
    - [MsgUnsanctionResponse](#cosmos-sanction-v1beta1-MsgUnsanctionResponse)
    - [MsgUpdateParams](#cosmos-sanction-v1beta1-MsgUpdateParams)
    - [MsgUpdateParamsResponse](#cosmos-sanction-v1beta1-MsgUpdateParamsResponse)
  
    - [Msg](#cosmos-sanction-v1beta1-Msg)
  
- [cosmos/sanction/v1beta1/events.proto](#cosmos_sanction_v1beta1_events-proto)
    - [EventAddressSanctioned](#cosmos-sanction-v1beta1-EventAddressSanctioned)
    - [EventAddressUnsanctioned](#cosmos-sanction-v1beta1-EventAddressUnsanctioned)
    - [EventParamsUpdated](#cosmos-sanction-v1beta1-EventParamsUpdated)
    - [EventTempAddressSanctioned](#cosmos-sanction-v1beta1-EventTempAddressSanctioned)
    - [EventTempAddressUnsanctioned](#cosmos-sanction-v1beta1-EventTempAddressUnsanctioned)
  
- [cosmos/sanction/v1beta1/query.proto](#cosmos_sanction_v1beta1_query-proto)
    - [QueryIsSanctionedRequest](#cosmos-sanction-v1beta1-QueryIsSanctionedRequest)
    - [QueryIsSanctionedResponse](#cosmos-sanction-v1beta1-QueryIsSanctionedResponse)
    - [QueryParamsRequest](#cosmos-sanction-v1beta1-QueryParamsRequest)
    - [QueryParamsResponse](#cosmos-sanction-v1beta1-QueryParamsResponse)
    - [QuerySanctionedAddressesRequest](#cosmos-sanction-v1beta1-QuerySanctionedAddressesRequest)
    - [QuerySanctionedAddressesResponse](#cosmos-sanction-v1beta1-QuerySanctionedAddressesResponse)
    - [QueryTemporaryEntriesRequest](#cosmos-sanction-v1beta1-QueryTemporaryEntriesRequest)
    - [QueryTemporaryEntriesResponse](#cosmos-sanction-v1beta1-QueryTemporaryEntriesResponse)
  
    - [Query](#cosmos-sanction-v1beta1-Query)
  
- [cosmos/sanction/v1beta1/genesis.proto](#cosmos_sanction_v1beta1_genesis-proto)
    - [GenesisState](#cosmos-sanction-v1beta1-GenesisState)
  
- [cosmos/sanction/v1beta1/sanction.proto](#cosmos_sanction_v1beta1_sanction-proto)
    - [Params](#cosmos-sanction-v1beta1-Params)
    - [TemporaryEntry](#cosmos-sanction-v1beta1-TemporaryEntry)
  
    - [TempStatus](#cosmos-sanction-v1beta1-TempStatus)
  
- [provenance/exchange/v1/tx.proto](#provenance_exchange_v1_tx-proto)
    - [MsgAcceptPaymentRequest](#provenance-exchange-v1-MsgAcceptPaymentRequest)
    - [MsgAcceptPaymentResponse](#provenance-exchange-v1-MsgAcceptPaymentResponse)
    - [MsgCancelOrderRequest](#provenance-exchange-v1-MsgCancelOrderRequest)
    - [MsgCancelOrderResponse](#provenance-exchange-v1-MsgCancelOrderResponse)
    - [MsgCancelPaymentsRequest](#provenance-exchange-v1-MsgCancelPaymentsRequest)
    - [MsgCancelPaymentsResponse](#provenance-exchange-v1-MsgCancelPaymentsResponse)
    - [MsgChangePaymentTargetRequest](#provenance-exchange-v1-MsgChangePaymentTargetRequest)
    - [MsgChangePaymentTargetResponse](#provenance-exchange-v1-MsgChangePaymentTargetResponse)
    - [MsgCommitFundsRequest](#provenance-exchange-v1-MsgCommitFundsRequest)
    - [MsgCommitFundsResponse](#provenance-exchange-v1-MsgCommitFundsResponse)
    - [MsgCreateAskRequest](#provenance-exchange-v1-MsgCreateAskRequest)
    - [MsgCreateAskResponse](#provenance-exchange-v1-MsgCreateAskResponse)
    - [MsgCreateBidRequest](#provenance-exchange-v1-MsgCreateBidRequest)
    - [MsgCreateBidResponse](#provenance-exchange-v1-MsgCreateBidResponse)
    - [MsgCreatePaymentRequest](#provenance-exchange-v1-MsgCreatePaymentRequest)
    - [MsgCreatePaymentResponse](#provenance-exchange-v1-MsgCreatePaymentResponse)
    - [MsgFillAsksRequest](#provenance-exchange-v1-MsgFillAsksRequest)
    - [MsgFillAsksResponse](#provenance-exchange-v1-MsgFillAsksResponse)
    - [MsgFillBidsRequest](#provenance-exchange-v1-MsgFillBidsRequest)
    - [MsgFillBidsResponse](#provenance-exchange-v1-MsgFillBidsResponse)
    - [MsgGovCloseMarketRequest](#provenance-exchange-v1-MsgGovCloseMarketRequest)
    - [MsgGovCloseMarketResponse](#provenance-exchange-v1-MsgGovCloseMarketResponse)
    - [MsgGovCreateMarketRequest](#provenance-exchange-v1-MsgGovCreateMarketRequest)
    - [MsgGovCreateMarketResponse](#provenance-exchange-v1-MsgGovCreateMarketResponse)
    - [MsgGovManageFeesRequest](#provenance-exchange-v1-MsgGovManageFeesRequest)
    - [MsgGovManageFeesResponse](#provenance-exchange-v1-MsgGovManageFeesResponse)
    - [MsgGovUpdateParamsRequest](#provenance-exchange-v1-MsgGovUpdateParamsRequest)
    - [MsgGovUpdateParamsResponse](#provenance-exchange-v1-MsgGovUpdateParamsResponse)
    - [MsgMarketCommitmentSettleRequest](#provenance-exchange-v1-MsgMarketCommitmentSettleRequest)
    - [MsgMarketCommitmentSettleResponse](#provenance-exchange-v1-MsgMarketCommitmentSettleResponse)
    - [MsgMarketManagePermissionsRequest](#provenance-exchange-v1-MsgMarketManagePermissionsRequest)
    - [MsgMarketManagePermissionsResponse](#provenance-exchange-v1-MsgMarketManagePermissionsResponse)
    - [MsgMarketManageReqAttrsRequest](#provenance-exchange-v1-MsgMarketManageReqAttrsRequest)
    - [MsgMarketManageReqAttrsResponse](#provenance-exchange-v1-MsgMarketManageReqAttrsResponse)
    - [MsgMarketReleaseCommitmentsRequest](#provenance-exchange-v1-MsgMarketReleaseCommitmentsRequest)
    - [MsgMarketReleaseCommitmentsResponse](#provenance-exchange-v1-MsgMarketReleaseCommitmentsResponse)
    - [MsgMarketSetOrderExternalIDRequest](#provenance-exchange-v1-MsgMarketSetOrderExternalIDRequest)
    - [MsgMarketSetOrderExternalIDResponse](#provenance-exchange-v1-MsgMarketSetOrderExternalIDResponse)
    - [MsgMarketSettleRequest](#provenance-exchange-v1-MsgMarketSettleRequest)
    - [MsgMarketSettleResponse](#provenance-exchange-v1-MsgMarketSettleResponse)
    - [MsgMarketUpdateAcceptingCommitmentsRequest](#provenance-exchange-v1-MsgMarketUpdateAcceptingCommitmentsRequest)
    - [MsgMarketUpdateAcceptingCommitmentsResponse](#provenance-exchange-v1-MsgMarketUpdateAcceptingCommitmentsResponse)
    - [MsgMarketUpdateAcceptingOrdersRequest](#provenance-exchange-v1-MsgMarketUpdateAcceptingOrdersRequest)
    - [MsgMarketUpdateAcceptingOrdersResponse](#provenance-exchange-v1-MsgMarketUpdateAcceptingOrdersResponse)
    - [MsgMarketUpdateDetailsRequest](#provenance-exchange-v1-MsgMarketUpdateDetailsRequest)
    - [MsgMarketUpdateDetailsResponse](#provenance-exchange-v1-MsgMarketUpdateDetailsResponse)
    - [MsgMarketUpdateEnabledRequest](#provenance-exchange-v1-MsgMarketUpdateEnabledRequest)
    - [MsgMarketUpdateEnabledResponse](#provenance-exchange-v1-MsgMarketUpdateEnabledResponse)
    - [MsgMarketUpdateIntermediaryDenomRequest](#provenance-exchange-v1-MsgMarketUpdateIntermediaryDenomRequest)
    - [MsgMarketUpdateIntermediaryDenomResponse](#provenance-exchange-v1-MsgMarketUpdateIntermediaryDenomResponse)
    - [MsgMarketUpdateUserSettleRequest](#provenance-exchange-v1-MsgMarketUpdateUserSettleRequest)
    - [MsgMarketUpdateUserSettleResponse](#provenance-exchange-v1-MsgMarketUpdateUserSettleResponse)
    - [MsgMarketWithdrawRequest](#provenance-exchange-v1-MsgMarketWithdrawRequest)
    - [MsgMarketWithdrawResponse](#provenance-exchange-v1-MsgMarketWithdrawResponse)
    - [MsgRejectPaymentRequest](#provenance-exchange-v1-MsgRejectPaymentRequest)
    - [MsgRejectPaymentResponse](#provenance-exchange-v1-MsgRejectPaymentResponse)
    - [MsgRejectPaymentsRequest](#provenance-exchange-v1-MsgRejectPaymentsRequest)
    - [MsgRejectPaymentsResponse](#provenance-exchange-v1-MsgRejectPaymentsResponse)
    - [MsgUpdateParamsRequest](#provenance-exchange-v1-MsgUpdateParamsRequest)
    - [MsgUpdateParamsResponse](#provenance-exchange-v1-MsgUpdateParamsResponse)
  
    - [Msg](#provenance-exchange-v1-Msg)
  
- [provenance/exchange/v1/events.proto](#provenance_exchange_v1_events-proto)
    - [EventCommitmentReleased](#provenance-exchange-v1-EventCommitmentReleased)
    - [EventFundsCommitted](#provenance-exchange-v1-EventFundsCommitted)
    - [EventMarketCommitmentsDisabled](#provenance-exchange-v1-EventMarketCommitmentsDisabled)
    - [EventMarketCommitmentsEnabled](#provenance-exchange-v1-EventMarketCommitmentsEnabled)
    - [EventMarketCreated](#provenance-exchange-v1-EventMarketCreated)
    - [EventMarketDetailsUpdated](#provenance-exchange-v1-EventMarketDetailsUpdated)
    - [EventMarketDisabled](#provenance-exchange-v1-EventMarketDisabled)
    - [EventMarketEnabled](#provenance-exchange-v1-EventMarketEnabled)
    - [EventMarketFeesUpdated](#provenance-exchange-v1-EventMarketFeesUpdated)
    - [EventMarketIntermediaryDenomUpdated](#provenance-exchange-v1-EventMarketIntermediaryDenomUpdated)
    - [EventMarketOrdersDisabled](#provenance-exchange-v1-EventMarketOrdersDisabled)
    - [EventMarketOrdersEnabled](#provenance-exchange-v1-EventMarketOrdersEnabled)
    - [EventMarketPermissionsUpdated](#provenance-exchange-v1-EventMarketPermissionsUpdated)
    - [EventMarketReqAttrUpdated](#provenance-exchange-v1-EventMarketReqAttrUpdated)
    - [EventMarketUserSettleDisabled](#provenance-exchange-v1-EventMarketUserSettleDisabled)
    - [EventMarketUserSettleEnabled](#provenance-exchange-v1-EventMarketUserSettleEnabled)
    - [EventMarketWithdraw](#provenance-exchange-v1-EventMarketWithdraw)
    - [EventOrderCancelled](#provenance-exchange-v1-EventOrderCancelled)
    - [EventOrderCreated](#provenance-exchange-v1-EventOrderCreated)
    - [EventOrderExternalIDUpdated](#provenance-exchange-v1-EventOrderExternalIDUpdated)
    - [EventOrderFilled](#provenance-exchange-v1-EventOrderFilled)
    - [EventOrderPartiallyFilled](#provenance-exchange-v1-EventOrderPartiallyFilled)
    - [EventParamsUpdated](#provenance-exchange-v1-EventParamsUpdated)
    - [EventPaymentAccepted](#provenance-exchange-v1-EventPaymentAccepted)
    - [EventPaymentCancelled](#provenance-exchange-v1-EventPaymentCancelled)
    - [EventPaymentCreated](#provenance-exchange-v1-EventPaymentCreated)
    - [EventPaymentRejected](#provenance-exchange-v1-EventPaymentRejected)
    - [EventPaymentUpdated](#provenance-exchange-v1-EventPaymentUpdated)
  
- [provenance/exchange/v1/market.proto](#provenance_exchange_v1_market-proto)
    - [AccessGrant](#provenance-exchange-v1-AccessGrant)
    - [FeeRatio](#provenance-exchange-v1-FeeRatio)
    - [Market](#provenance-exchange-v1-Market)
    - [MarketAccount](#provenance-exchange-v1-MarketAccount)
    - [MarketBrief](#provenance-exchange-v1-MarketBrief)
    - [MarketDetails](#provenance-exchange-v1-MarketDetails)
  
    - [Permission](#provenance-exchange-v1-Permission)
  
- [provenance/exchange/v1/payments.proto](#provenance_exchange_v1_payments-proto)
    - [Payment](#provenance-exchange-v1-Payment)
  
- [provenance/exchange/v1/commitments.proto](#provenance_exchange_v1_commitments-proto)
    - [AccountAmount](#provenance-exchange-v1-AccountAmount)
    - [Commitment](#provenance-exchange-v1-Commitment)
    - [MarketAmount](#provenance-exchange-v1-MarketAmount)
    - [NetAssetPrice](#provenance-exchange-v1-NetAssetPrice)
  
- [provenance/exchange/v1/query.proto](#provenance_exchange_v1_query-proto)
    - [QueryCommitmentSettlementFeeCalcRequest](#provenance-exchange-v1-QueryCommitmentSettlementFeeCalcRequest)
    - [QueryCommitmentSettlementFeeCalcResponse](#provenance-exchange-v1-QueryCommitmentSettlementFeeCalcResponse)
    - [QueryGetAccountCommitmentsRequest](#provenance-exchange-v1-QueryGetAccountCommitmentsRequest)
    - [QueryGetAccountCommitmentsResponse](#provenance-exchange-v1-QueryGetAccountCommitmentsResponse)
    - [QueryGetAllCommitmentsRequest](#provenance-exchange-v1-QueryGetAllCommitmentsRequest)
    - [QueryGetAllCommitmentsResponse](#provenance-exchange-v1-QueryGetAllCommitmentsResponse)
    - [QueryGetAllMarketsRequest](#provenance-exchange-v1-QueryGetAllMarketsRequest)
    - [QueryGetAllMarketsResponse](#provenance-exchange-v1-QueryGetAllMarketsResponse)
    - [QueryGetAllOrdersRequest](#provenance-exchange-v1-QueryGetAllOrdersRequest)
    - [QueryGetAllOrdersResponse](#provenance-exchange-v1-QueryGetAllOrdersResponse)
    - [QueryGetAllPaymentsRequest](#provenance-exchange-v1-QueryGetAllPaymentsRequest)
    - [QueryGetAllPaymentsResponse](#provenance-exchange-v1-QueryGetAllPaymentsResponse)
    - [QueryGetAssetOrdersRequest](#provenance-exchange-v1-QueryGetAssetOrdersRequest)
    - [QueryGetAssetOrdersResponse](#provenance-exchange-v1-QueryGetAssetOrdersResponse)
    - [QueryGetCommitmentRequest](#provenance-exchange-v1-QueryGetCommitmentRequest)
    - [QueryGetCommitmentResponse](#provenance-exchange-v1-QueryGetCommitmentResponse)
    - [QueryGetMarketCommitmentsRequest](#provenance-exchange-v1-QueryGetMarketCommitmentsRequest)
    - [QueryGetMarketCommitmentsResponse](#provenance-exchange-v1-QueryGetMarketCommitmentsResponse)
    - [QueryGetMarketOrdersRequest](#provenance-exchange-v1-QueryGetMarketOrdersRequest)
    - [QueryGetMarketOrdersResponse](#provenance-exchange-v1-QueryGetMarketOrdersResponse)
    - [QueryGetMarketRequest](#provenance-exchange-v1-QueryGetMarketRequest)
    - [QueryGetMarketResponse](#provenance-exchange-v1-QueryGetMarketResponse)
    - [QueryGetOrderByExternalIDRequest](#provenance-exchange-v1-QueryGetOrderByExternalIDRequest)
    - [QueryGetOrderByExternalIDResponse](#provenance-exchange-v1-QueryGetOrderByExternalIDResponse)
    - [QueryGetOrderRequest](#provenance-exchange-v1-QueryGetOrderRequest)
    - [QueryGetOrderResponse](#provenance-exchange-v1-QueryGetOrderResponse)
    - [QueryGetOwnerOrdersRequest](#provenance-exchange-v1-QueryGetOwnerOrdersRequest)
    - [QueryGetOwnerOrdersResponse](#provenance-exchange-v1-QueryGetOwnerOrdersResponse)
    - [QueryGetPaymentRequest](#provenance-exchange-v1-QueryGetPaymentRequest)
    - [QueryGetPaymentResponse](#provenance-exchange-v1-QueryGetPaymentResponse)
    - [QueryGetPaymentsWithSourceRequest](#provenance-exchange-v1-QueryGetPaymentsWithSourceRequest)
    - [QueryGetPaymentsWithSourceResponse](#provenance-exchange-v1-QueryGetPaymentsWithSourceResponse)
    - [QueryGetPaymentsWithTargetRequest](#provenance-exchange-v1-QueryGetPaymentsWithTargetRequest)
    - [QueryGetPaymentsWithTargetResponse](#provenance-exchange-v1-QueryGetPaymentsWithTargetResponse)
    - [QueryOrderFeeCalcRequest](#provenance-exchange-v1-QueryOrderFeeCalcRequest)
    - [QueryOrderFeeCalcResponse](#provenance-exchange-v1-QueryOrderFeeCalcResponse)
    - [QueryParamsRequest](#provenance-exchange-v1-QueryParamsRequest)
    - [QueryParamsResponse](#provenance-exchange-v1-QueryParamsResponse)
    - [QueryPaymentFeeCalcRequest](#provenance-exchange-v1-QueryPaymentFeeCalcRequest)
    - [QueryPaymentFeeCalcResponse](#provenance-exchange-v1-QueryPaymentFeeCalcResponse)
    - [QueryValidateCreateMarketRequest](#provenance-exchange-v1-QueryValidateCreateMarketRequest)
    - [QueryValidateCreateMarketResponse](#provenance-exchange-v1-QueryValidateCreateMarketResponse)
    - [QueryValidateManageFeesRequest](#provenance-exchange-v1-QueryValidateManageFeesRequest)
    - [QueryValidateManageFeesResponse](#provenance-exchange-v1-QueryValidateManageFeesResponse)
    - [QueryValidateMarketRequest](#provenance-exchange-v1-QueryValidateMarketRequest)
    - [QueryValidateMarketResponse](#provenance-exchange-v1-QueryValidateMarketResponse)
  
    - [Query](#provenance-exchange-v1-Query)
  
- [provenance/exchange/v1/genesis.proto](#provenance_exchange_v1_genesis-proto)
    - [GenesisState](#provenance-exchange-v1-GenesisState)
  
- [provenance/exchange/v1/orders.proto](#provenance_exchange_v1_orders-proto)
    - [AskOrder](#provenance-exchange-v1-AskOrder)
    - [BidOrder](#provenance-exchange-v1-BidOrder)
    - [Order](#provenance-exchange-v1-Order)
  
- [provenance/exchange/v1/params.proto](#provenance_exchange_v1_params-proto)
    - [DenomSplit](#provenance-exchange-v1-DenomSplit)
    - [Params](#provenance-exchange-v1-Params)
  
- [provenance/ledger/v1/tx.proto](#provenance_ledger_v1_tx-proto)
    - [MsgAppendRequest](#provenance-ledger-v1-MsgAppendRequest)
    - [MsgAppendResponse](#provenance-ledger-v1-MsgAppendResponse)
    - [MsgCreateRequest](#provenance-ledger-v1-MsgCreateRequest)
    - [MsgCreateResponse](#provenance-ledger-v1-MsgCreateResponse)
    - [MsgDestroyRequest](#provenance-ledger-v1-MsgDestroyRequest)
    - [MsgDestroyResponse](#provenance-ledger-v1-MsgDestroyResponse)
    - [MsgProcessFundTransfersRequest](#provenance-ledger-v1-MsgProcessFundTransfersRequest)
    - [MsgProcessFundTransfersResponse](#provenance-ledger-v1-MsgProcessFundTransfersResponse)
    - [MsgProcessFundTransfersWithSettlementRequest](#provenance-ledger-v1-MsgProcessFundTransfersWithSettlementRequest)
    - [MsgUpdateBalancesRequest](#provenance-ledger-v1-MsgUpdateBalancesRequest)
    - [MsgUpdateBalancesResponse](#provenance-ledger-v1-MsgUpdateBalancesResponse)
  
    - [Msg](#provenance-ledger-v1-Msg)
  
- [provenance/ledger/v1/ledger.proto](#provenance_ledger_v1_ledger-proto)
    - [Balances](#provenance-ledger-v1-Balances)
    - [BucketBalance](#provenance-ledger-v1-BucketBalance)
    - [Ledger](#provenance-ledger-v1-Ledger)
    - [LedgerBucketAmount](#provenance-ledger-v1-LedgerBucketAmount)
    - [LedgerClass](#provenance-ledger-v1-LedgerClass)
    - [LedgerClassBucketType](#provenance-ledger-v1-LedgerClassBucketType)
    - [LedgerClassEntryType](#provenance-ledger-v1-LedgerClassEntryType)
    - [LedgerClassStatusType](#provenance-ledger-v1-LedgerClassStatusType)
    - [LedgerEntry](#provenance-ledger-v1-LedgerEntry)
    - [LedgerEntry.BucketBalancesEntry](#provenance-ledger-v1-LedgerEntry-BucketBalancesEntry)
    - [LedgerKey](#provenance-ledger-v1-LedgerKey)
  
- [provenance/ledger/v1/query.proto](#provenance_ledger_v1_query-proto)
    - [QueryBalancesAsOfRequest](#provenance-ledger-v1-QueryBalancesAsOfRequest)
    - [QueryBalancesAsOfResponse](#provenance-ledger-v1-QueryBalancesAsOfResponse)
    - [QueryLedgerClassBucketTypesRequest](#provenance-ledger-v1-QueryLedgerClassBucketTypesRequest)
    - [QueryLedgerClassBucketTypesResponse](#provenance-ledger-v1-QueryLedgerClassBucketTypesResponse)
    - [QueryLedgerClassEntryTypesRequest](#provenance-ledger-v1-QueryLedgerClassEntryTypesRequest)
    - [QueryLedgerClassEntryTypesResponse](#provenance-ledger-v1-QueryLedgerClassEntryTypesResponse)
    - [QueryLedgerClassStatusTypesRequest](#provenance-ledger-v1-QueryLedgerClassStatusTypesRequest)
    - [QueryLedgerClassStatusTypesResponse](#provenance-ledger-v1-QueryLedgerClassStatusTypesResponse)
    - [QueryLedgerConfigRequest](#provenance-ledger-v1-QueryLedgerConfigRequest)
    - [QueryLedgerConfigResponse](#provenance-ledger-v1-QueryLedgerConfigResponse)
    - [QueryLedgerEntryRequest](#provenance-ledger-v1-QueryLedgerEntryRequest)
    - [QueryLedgerEntryResponse](#provenance-ledger-v1-QueryLedgerEntryResponse)
    - [QueryLedgerRequest](#provenance-ledger-v1-QueryLedgerRequest)
    - [QueryLedgerResponse](#provenance-ledger-v1-QueryLedgerResponse)
  
    - [Query](#provenance-ledger-v1-Query)
  
- [provenance/ledger/v1/ledger_settlement.proto](#provenance_ledger_v1_ledger_settlement-proto)
    - [FundTransfer](#provenance-ledger-v1-FundTransfer)
    - [FundTransferWithSettlement](#provenance-ledger-v1-FundTransferWithSettlement)
    - [SettlementInstruction](#provenance-ledger-v1-SettlementInstruction)
  
    - [FundingTransferStatus](#provenance-ledger-v1-FundingTransferStatus)
  
- [provenance/ledger/v1/ledger_query.proto](#provenance_ledger_v1_ledger_query-proto)
    - [LedgerBucketAmountPlainText](#provenance-ledger-v1-LedgerBucketAmountPlainText)
    - [LedgerEntryPlainText](#provenance-ledger-v1-LedgerEntryPlainText)
    - [LedgerPlainText](#provenance-ledger-v1-LedgerPlainText)
    - [QueryLedgerEntryResponsePlainText](#provenance-ledger-v1-QueryLedgerEntryResponsePlainText)
  
- [provenance/ledger/v1/genesis.proto](#provenance_ledger_v1_genesis-proto)
    - [GenesisState](#provenance-ledger-v1-GenesisState)
  
- [provenance/trigger/v1/tx.proto](#provenance_trigger_v1_tx-proto)
    - [MsgCreateTriggerRequest](#provenance-trigger-v1-MsgCreateTriggerRequest)
    - [MsgCreateTriggerResponse](#provenance-trigger-v1-MsgCreateTriggerResponse)
    - [MsgDestroyTriggerRequest](#provenance-trigger-v1-MsgDestroyTriggerRequest)
    - [MsgDestroyTriggerResponse](#provenance-trigger-v1-MsgDestroyTriggerResponse)
  
    - [Msg](#provenance-trigger-v1-Msg)
  
- [provenance/trigger/v1/query.proto](#provenance_trigger_v1_query-proto)
    - [QueryTriggerByIDRequest](#provenance-trigger-v1-QueryTriggerByIDRequest)
    - [QueryTriggerByIDResponse](#provenance-trigger-v1-QueryTriggerByIDResponse)
    - [QueryTriggersRequest](#provenance-trigger-v1-QueryTriggersRequest)
    - [QueryTriggersResponse](#provenance-trigger-v1-QueryTriggersResponse)
  
    - [Query](#provenance-trigger-v1-Query)
  
- [provenance/trigger/v1/event.proto](#provenance_trigger_v1_event-proto)
    - [EventTriggerCreated](#provenance-trigger-v1-EventTriggerCreated)
    - [EventTriggerDestroyed](#provenance-trigger-v1-EventTriggerDestroyed)
    - [EventTriggerDetected](#provenance-trigger-v1-EventTriggerDetected)
    - [EventTriggerExecuted](#provenance-trigger-v1-EventTriggerExecuted)
  
- [provenance/trigger/v1/genesis.proto](#provenance_trigger_v1_genesis-proto)
    - [GasLimit](#provenance-trigger-v1-GasLimit)
    - [GenesisState](#provenance-trigger-v1-GenesisState)
  
- [provenance/trigger/v1/trigger.proto](#provenance_trigger_v1_trigger-proto)
    - [Attribute](#provenance-trigger-v1-Attribute)
    - [BlockHeightEvent](#provenance-trigger-v1-BlockHeightEvent)
    - [BlockTimeEvent](#provenance-trigger-v1-BlockTimeEvent)
    - [QueuedTrigger](#provenance-trigger-v1-QueuedTrigger)
    - [TransactionEvent](#provenance-trigger-v1-TransactionEvent)
    - [Trigger](#provenance-trigger-v1-Trigger)
  
- [provenance/attribute/v1/tx.proto](#provenance_attribute_v1_tx-proto)
    - [MsgAddAttributeRequest](#provenance-attribute-v1-MsgAddAttributeRequest)
    - [MsgAddAttributeResponse](#provenance-attribute-v1-MsgAddAttributeResponse)
    - [MsgDeleteAttributeRequest](#provenance-attribute-v1-MsgDeleteAttributeRequest)
    - [MsgDeleteAttributeResponse](#provenance-attribute-v1-MsgDeleteAttributeResponse)
    - [MsgDeleteDistinctAttributeRequest](#provenance-attribute-v1-MsgDeleteDistinctAttributeRequest)
    - [MsgDeleteDistinctAttributeResponse](#provenance-attribute-v1-MsgDeleteDistinctAttributeResponse)
    - [MsgSetAccountDataRequest](#provenance-attribute-v1-MsgSetAccountDataRequest)
    - [MsgSetAccountDataResponse](#provenance-attribute-v1-MsgSetAccountDataResponse)
    - [MsgUpdateAttributeExpirationRequest](#provenance-attribute-v1-MsgUpdateAttributeExpirationRequest)
    - [MsgUpdateAttributeExpirationResponse](#provenance-attribute-v1-MsgUpdateAttributeExpirationResponse)
    - [MsgUpdateAttributeRequest](#provenance-attribute-v1-MsgUpdateAttributeRequest)
    - [MsgUpdateAttributeResponse](#provenance-attribute-v1-MsgUpdateAttributeResponse)
    - [MsgUpdateParamsRequest](#provenance-attribute-v1-MsgUpdateParamsRequest)
    - [MsgUpdateParamsResponse](#provenance-attribute-v1-MsgUpdateParamsResponse)
  
    - [Msg](#provenance-attribute-v1-Msg)
  
- [provenance/attribute/v1/attribute.proto](#provenance_attribute_v1_attribute-proto)
    - [Attribute](#provenance-attribute-v1-Attribute)
    - [EventAccountDataUpdated](#provenance-attribute-v1-EventAccountDataUpdated)
    - [EventAttributeAdd](#provenance-attribute-v1-EventAttributeAdd)
    - [EventAttributeDelete](#provenance-attribute-v1-EventAttributeDelete)
    - [EventAttributeDistinctDelete](#provenance-attribute-v1-EventAttributeDistinctDelete)
    - [EventAttributeExpirationUpdate](#provenance-attribute-v1-EventAttributeExpirationUpdate)
    - [EventAttributeExpired](#provenance-attribute-v1-EventAttributeExpired)
    - [EventAttributeParamsUpdated](#provenance-attribute-v1-EventAttributeParamsUpdated)
    - [EventAttributeUpdate](#provenance-attribute-v1-EventAttributeUpdate)
    - [Params](#provenance-attribute-v1-Params)
  
    - [AttributeType](#provenance-attribute-v1-AttributeType)
  
- [provenance/attribute/v1/query.proto](#provenance_attribute_v1_query-proto)
    - [QueryAccountDataRequest](#provenance-attribute-v1-QueryAccountDataRequest)
    - [QueryAccountDataResponse](#provenance-attribute-v1-QueryAccountDataResponse)
    - [QueryAttributeAccountsRequest](#provenance-attribute-v1-QueryAttributeAccountsRequest)
    - [QueryAttributeAccountsResponse](#provenance-attribute-v1-QueryAttributeAccountsResponse)
    - [QueryAttributeRequest](#provenance-attribute-v1-QueryAttributeRequest)
    - [QueryAttributeResponse](#provenance-attribute-v1-QueryAttributeResponse)
    - [QueryAttributesRequest](#provenance-attribute-v1-QueryAttributesRequest)
    - [QueryAttributesResponse](#provenance-attribute-v1-QueryAttributesResponse)
    - [QueryParamsRequest](#provenance-attribute-v1-QueryParamsRequest)
    - [QueryParamsResponse](#provenance-attribute-v1-QueryParamsResponse)
    - [QueryScanRequest](#provenance-attribute-v1-QueryScanRequest)
    - [QueryScanResponse](#provenance-attribute-v1-QueryScanResponse)
  
    - [Query](#provenance-attribute-v1-Query)
  
- [provenance/attribute/v1/genesis.proto](#provenance_attribute_v1_genesis-proto)
    - [GenesisState](#provenance-attribute-v1-GenesisState)
  
- [provenance/asset/v1/tx.proto](#provenance_asset_v1_tx-proto)
    - [MsgAddAsset](#provenance-asset-v1-MsgAddAsset)
    - [MsgAddAssetClass](#provenance-asset-v1-MsgAddAssetClass)
    - [MsgAddAssetClassResponse](#provenance-asset-v1-MsgAddAssetClassResponse)
    - [MsgAddAssetResponse](#provenance-asset-v1-MsgAddAssetResponse)
  
    - [Msg](#provenance-asset-v1-Msg)
  
- [provenance/asset/v1/asset.proto](#provenance_asset_v1_asset-proto)
    - [Asset](#provenance-asset-v1-Asset)
    - [AssetClass](#provenance-asset-v1-AssetClass)
  
- [provenance/asset/v1/query.proto](#provenance_asset_v1_query-proto)
    - [QueryGetClass](#provenance-asset-v1-QueryGetClass)
    - [QueryGetClassResponse](#provenance-asset-v1-QueryGetClassResponse)
    - [QueryListAssetClasses](#provenance-asset-v1-QueryListAssetClasses)
    - [QueryListAssetClassesResponse](#provenance-asset-v1-QueryListAssetClassesResponse)
    - [QueryListAssets](#provenance-asset-v1-QueryListAssets)
    - [QueryListAssetsResponse](#provenance-asset-v1-QueryListAssetsResponse)
  
    - [Query](#provenance-asset-v1-Query)
  
- [provenance/asset/v1/genesis.proto](#provenance_asset_v1_genesis-proto)
    - [GenesisState](#provenance-asset-v1-GenesisState)
  
- [provenance/msgfees/v1/tx.proto](#provenance_msgfees_v1_tx-proto)
    - [MsgAddMsgFeeProposalRequest](#provenance-msgfees-v1-MsgAddMsgFeeProposalRequest)
    - [MsgAddMsgFeeProposalResponse](#provenance-msgfees-v1-MsgAddMsgFeeProposalResponse)
    - [MsgAssessCustomMsgFeeRequest](#provenance-msgfees-v1-MsgAssessCustomMsgFeeRequest)
    - [MsgAssessCustomMsgFeeResponse](#provenance-msgfees-v1-MsgAssessCustomMsgFeeResponse)
    - [MsgRemoveMsgFeeProposalRequest](#provenance-msgfees-v1-MsgRemoveMsgFeeProposalRequest)
    - [MsgRemoveMsgFeeProposalResponse](#provenance-msgfees-v1-MsgRemoveMsgFeeProposalResponse)
    - [MsgUpdateConversionFeeDenomProposalRequest](#provenance-msgfees-v1-MsgUpdateConversionFeeDenomProposalRequest)
    - [MsgUpdateConversionFeeDenomProposalResponse](#provenance-msgfees-v1-MsgUpdateConversionFeeDenomProposalResponse)
    - [MsgUpdateMsgFeeProposalRequest](#provenance-msgfees-v1-MsgUpdateMsgFeeProposalRequest)
    - [MsgUpdateMsgFeeProposalResponse](#provenance-msgfees-v1-MsgUpdateMsgFeeProposalResponse)
    - [MsgUpdateNhashPerUsdMilProposalRequest](#provenance-msgfees-v1-MsgUpdateNhashPerUsdMilProposalRequest)
    - [MsgUpdateNhashPerUsdMilProposalResponse](#provenance-msgfees-v1-MsgUpdateNhashPerUsdMilProposalResponse)
  
    - [Msg](#provenance-msgfees-v1-Msg)
  
- [provenance/msgfees/v1/query.proto](#provenance_msgfees_v1_query-proto)
    - [CalculateTxFeesRequest](#provenance-msgfees-v1-CalculateTxFeesRequest)
    - [CalculateTxFeesResponse](#provenance-msgfees-v1-CalculateTxFeesResponse)
    - [QueryAllMsgFeesRequest](#provenance-msgfees-v1-QueryAllMsgFeesRequest)
    - [QueryAllMsgFeesResponse](#provenance-msgfees-v1-QueryAllMsgFeesResponse)
    - [QueryParamsRequest](#provenance-msgfees-v1-QueryParamsRequest)
    - [QueryParamsResponse](#provenance-msgfees-v1-QueryParamsResponse)
  
    - [Query](#provenance-msgfees-v1-Query)
  
- [provenance/msgfees/v1/genesis.proto](#provenance_msgfees_v1_genesis-proto)
    - [GenesisState](#provenance-msgfees-v1-GenesisState)
  
- [provenance/msgfees/v1/msgfees.proto](#provenance_msgfees_v1_msgfees-proto)
    - [EventMsgFee](#provenance-msgfees-v1-EventMsgFee)
    - [EventMsgFees](#provenance-msgfees-v1-EventMsgFees)
    - [MsgFee](#provenance-msgfees-v1-MsgFee)
    - [Params](#provenance-msgfees-v1-Params)
  
- [provenance/msgfees/v1/proposals.proto](#provenance_msgfees_v1_proposals-proto)
    - [AddMsgFeeProposal](#provenance-msgfees-v1-AddMsgFeeProposal)
    - [RemoveMsgFeeProposal](#provenance-msgfees-v1-RemoveMsgFeeProposal)
    - [UpdateConversionFeeDenomProposal](#provenance-msgfees-v1-UpdateConversionFeeDenomProposal)
    - [UpdateMsgFeeProposal](#provenance-msgfees-v1-UpdateMsgFeeProposal)
    - [UpdateNhashPerUsdMilProposal](#provenance-msgfees-v1-UpdateNhashPerUsdMilProposal)
  
- [provenance/oracle/v1/tx.proto](#provenance_oracle_v1_tx-proto)
    - [MsgSendQueryOracleRequest](#provenance-oracle-v1-MsgSendQueryOracleRequest)
    - [MsgSendQueryOracleResponse](#provenance-oracle-v1-MsgSendQueryOracleResponse)
    - [MsgUpdateOracleRequest](#provenance-oracle-v1-MsgUpdateOracleRequest)
    - [MsgUpdateOracleResponse](#provenance-oracle-v1-MsgUpdateOracleResponse)
  
    - [Msg](#provenance-oracle-v1-Msg)
  
- [provenance/oracle/v1/query.proto](#provenance_oracle_v1_query-proto)
    - [QueryOracleAddressRequest](#provenance-oracle-v1-QueryOracleAddressRequest)
    - [QueryOracleAddressResponse](#provenance-oracle-v1-QueryOracleAddressResponse)
    - [QueryOracleRequest](#provenance-oracle-v1-QueryOracleRequest)
    - [QueryOracleResponse](#provenance-oracle-v1-QueryOracleResponse)
  
    - [Query](#provenance-oracle-v1-Query)
  
- [provenance/oracle/v1/event.proto](#provenance_oracle_v1_event-proto)
    - [EventOracleQueryError](#provenance-oracle-v1-EventOracleQueryError)
    - [EventOracleQuerySuccess](#provenance-oracle-v1-EventOracleQuerySuccess)
    - [EventOracleQueryTimeout](#provenance-oracle-v1-EventOracleQueryTimeout)
  
- [provenance/oracle/v1/genesis.proto](#provenance_oracle_v1_genesis-proto)
    - [GenesisState](#provenance-oracle-v1-GenesisState)
  
- [provenance/registry/v1/tx.proto](#provenance_registry_v1_tx-proto)
    - [MsgGrantRole](#provenance-registry-v1-MsgGrantRole)
    - [MsgGrantRoleResponse](#provenance-registry-v1-MsgGrantRoleResponse)
    - [MsgRegisterNFT](#provenance-registry-v1-MsgRegisterNFT)
    - [MsgRegisterNFT.RolesEntry](#provenance-registry-v1-MsgRegisterNFT-RolesEntry)
    - [MsgRegisterNFTResponse](#provenance-registry-v1-MsgRegisterNFTResponse)
    - [MsgRevokeRole](#provenance-registry-v1-MsgRevokeRole)
    - [MsgRevokeRoleResponse](#provenance-registry-v1-MsgRevokeRoleResponse)
    - [MsgUnregisterNFT](#provenance-registry-v1-MsgUnregisterNFT)
    - [MsgUnregisterNFTResponse](#provenance-registry-v1-MsgUnregisterNFTResponse)
  
    - [Msg](#provenance-registry-v1-Msg)
  
- [provenance/registry/v1/query.proto](#provenance_registry_v1_query-proto)
    - [QueryGetRegistryRequest](#provenance-registry-v1-QueryGetRegistryRequest)
    - [QueryGetRegistryResponse](#provenance-registry-v1-QueryGetRegistryResponse)
    - [QueryHasRoleRequest](#provenance-registry-v1-QueryHasRoleRequest)
    - [QueryHasRoleResponse](#provenance-registry-v1-QueryHasRoleResponse)
  
    - [Query](#provenance-registry-v1-Query)
  
- [provenance/registry/v1/registry.proto](#provenance_registry_v1_registry-proto)
    - [GenesisState](#provenance-registry-v1-GenesisState)
    - [RegistryEntry](#provenance-registry-v1-RegistryEntry)
    - [RegistryEntry.RolesEntry](#provenance-registry-v1-RegistryEntry-RolesEntry)
    - [RegistryKey](#provenance-registry-v1-RegistryKey)
    - [RoleAddresses](#provenance-registry-v1-RoleAddresses)
  
    - [RegistryRole](#provenance-registry-v1-RegistryRole)
  
- [provenance/ibchooks/v1/tx.proto](#provenance_ibchooks_v1_tx-proto)
    - [MsgEmitIBCAck](#provenance-ibchooks-v1-MsgEmitIBCAck)
    - [MsgEmitIBCAckResponse](#provenance-ibchooks-v1-MsgEmitIBCAckResponse)
    - [MsgUpdateParamsRequest](#provenance-ibchooks-v1-MsgUpdateParamsRequest)
    - [MsgUpdateParamsResponse](#provenance-ibchooks-v1-MsgUpdateParamsResponse)
  
    - [Msg](#provenance-ibchooks-v1-Msg)
  
- [provenance/ibchooks/v1/query.proto](#provenance_ibchooks_v1_query-proto)
    - [QueryParamsRequest](#provenance-ibchooks-v1-QueryParamsRequest)
    - [QueryParamsResponse](#provenance-ibchooks-v1-QueryParamsResponse)
  
    - [Query](#provenance-ibchooks-v1-Query)
  
- [provenance/ibchooks/v1/event.proto](#provenance_ibchooks_v1_event-proto)
    - [EventIBCHooksParamsUpdated](#provenance-ibchooks-v1-EventIBCHooksParamsUpdated)
  
- [provenance/ibchooks/v1/genesis.proto](#provenance_ibchooks_v1_genesis-proto)
    - [GenesisState](#provenance-ibchooks-v1-GenesisState)
  
- [provenance/ibchooks/v1/params.proto](#provenance_ibchooks_v1_params-proto)
    - [Params](#provenance-ibchooks-v1-Params)
  
- [provenance/ibcratelimit/v1/tx.proto](#provenance_ibcratelimit_v1_tx-proto)
    - [MsgGovUpdateParamsRequest](#provenance-ibcratelimit-v1-MsgGovUpdateParamsRequest)
    - [MsgGovUpdateParamsResponse](#provenance-ibcratelimit-v1-MsgGovUpdateParamsResponse)
    - [MsgUpdateParamsRequest](#provenance-ibcratelimit-v1-MsgUpdateParamsRequest)
    - [MsgUpdateParamsResponse](#provenance-ibcratelimit-v1-MsgUpdateParamsResponse)
  
    - [Msg](#provenance-ibcratelimit-v1-Msg)
  
- [provenance/ibcratelimit/v1/query.proto](#provenance_ibcratelimit_v1_query-proto)
    - [ParamsRequest](#provenance-ibcratelimit-v1-ParamsRequest)
    - [ParamsResponse](#provenance-ibcratelimit-v1-ParamsResponse)
  
    - [Query](#provenance-ibcratelimit-v1-Query)
  
- [provenance/ibcratelimit/v1/event.proto](#provenance_ibcratelimit_v1_event-proto)
    - [EventAckRevertFailure](#provenance-ibcratelimit-v1-EventAckRevertFailure)
    - [EventParamsUpdated](#provenance-ibcratelimit-v1-EventParamsUpdated)
    - [EventTimeoutRevertFailure](#provenance-ibcratelimit-v1-EventTimeoutRevertFailure)
  
- [provenance/ibcratelimit/v1/genesis.proto](#provenance_ibcratelimit_v1_genesis-proto)
    - [GenesisState](#provenance-ibcratelimit-v1-GenesisState)
  
- [provenance/ibcratelimit/v1/params.proto](#provenance_ibcratelimit_v1_params-proto)
    - [Params](#provenance-ibcratelimit-v1-Params)
  
- [provenance/marker/v1/tx.proto](#provenance_marker_v1_tx-proto)
    - [MsgActivateRequest](#provenance-marker-v1-MsgActivateRequest)
    - [MsgActivateResponse](#provenance-marker-v1-MsgActivateResponse)
    - [MsgAddAccessRequest](#provenance-marker-v1-MsgAddAccessRequest)
    - [MsgAddAccessResponse](#provenance-marker-v1-MsgAddAccessResponse)
    - [MsgAddFinalizeActivateMarkerRequest](#provenance-marker-v1-MsgAddFinalizeActivateMarkerRequest)
    - [MsgAddFinalizeActivateMarkerResponse](#provenance-marker-v1-MsgAddFinalizeActivateMarkerResponse)
    - [MsgAddMarkerRequest](#provenance-marker-v1-MsgAddMarkerRequest)
    - [MsgAddMarkerResponse](#provenance-marker-v1-MsgAddMarkerResponse)
    - [MsgAddNetAssetValuesRequest](#provenance-marker-v1-MsgAddNetAssetValuesRequest)
    - [MsgAddNetAssetValuesResponse](#provenance-marker-v1-MsgAddNetAssetValuesResponse)
    - [MsgBurnRequest](#provenance-marker-v1-MsgBurnRequest)
    - [MsgBurnResponse](#provenance-marker-v1-MsgBurnResponse)
    - [MsgCancelRequest](#provenance-marker-v1-MsgCancelRequest)
    - [MsgCancelResponse](#provenance-marker-v1-MsgCancelResponse)
    - [MsgChangeStatusProposalRequest](#provenance-marker-v1-MsgChangeStatusProposalRequest)
    - [MsgChangeStatusProposalResponse](#provenance-marker-v1-MsgChangeStatusProposalResponse)
    - [MsgDeleteAccessRequest](#provenance-marker-v1-MsgDeleteAccessRequest)
    - [MsgDeleteAccessResponse](#provenance-marker-v1-MsgDeleteAccessResponse)
    - [MsgDeleteRequest](#provenance-marker-v1-MsgDeleteRequest)
    - [MsgDeleteResponse](#provenance-marker-v1-MsgDeleteResponse)
    - [MsgFinalizeRequest](#provenance-marker-v1-MsgFinalizeRequest)
    - [MsgFinalizeResponse](#provenance-marker-v1-MsgFinalizeResponse)
    - [MsgGrantAllowanceRequest](#provenance-marker-v1-MsgGrantAllowanceRequest)
    - [MsgGrantAllowanceResponse](#provenance-marker-v1-MsgGrantAllowanceResponse)
    - [MsgIbcTransferRequest](#provenance-marker-v1-MsgIbcTransferRequest)
    - [MsgIbcTransferResponse](#provenance-marker-v1-MsgIbcTransferResponse)
    - [MsgMintRequest](#provenance-marker-v1-MsgMintRequest)
    - [MsgMintResponse](#provenance-marker-v1-MsgMintResponse)
    - [MsgRemoveAdministratorProposalRequest](#provenance-marker-v1-MsgRemoveAdministratorProposalRequest)
    - [MsgRemoveAdministratorProposalResponse](#provenance-marker-v1-MsgRemoveAdministratorProposalResponse)
    - [MsgSetAccountDataRequest](#provenance-marker-v1-MsgSetAccountDataRequest)
    - [MsgSetAccountDataResponse](#provenance-marker-v1-MsgSetAccountDataResponse)
    - [MsgSetAdministratorProposalRequest](#provenance-marker-v1-MsgSetAdministratorProposalRequest)
    - [MsgSetAdministratorProposalResponse](#provenance-marker-v1-MsgSetAdministratorProposalResponse)
    - [MsgSetDenomMetadataProposalRequest](#provenance-marker-v1-MsgSetDenomMetadataProposalRequest)
    - [MsgSetDenomMetadataProposalResponse](#provenance-marker-v1-MsgSetDenomMetadataProposalResponse)
    - [MsgSetDenomMetadataRequest](#provenance-marker-v1-MsgSetDenomMetadataRequest)
    - [MsgSetDenomMetadataResponse](#provenance-marker-v1-MsgSetDenomMetadataResponse)
    - [MsgSupplyDecreaseProposalRequest](#provenance-marker-v1-MsgSupplyDecreaseProposalRequest)
    - [MsgSupplyDecreaseProposalResponse](#provenance-marker-v1-MsgSupplyDecreaseProposalResponse)
    - [MsgSupplyIncreaseProposalRequest](#provenance-marker-v1-MsgSupplyIncreaseProposalRequest)
    - [MsgSupplyIncreaseProposalResponse](#provenance-marker-v1-MsgSupplyIncreaseProposalResponse)
    - [MsgTransferRequest](#provenance-marker-v1-MsgTransferRequest)
    - [MsgTransferResponse](#provenance-marker-v1-MsgTransferResponse)
    - [MsgUpdateForcedTransferRequest](#provenance-marker-v1-MsgUpdateForcedTransferRequest)
    - [MsgUpdateForcedTransferResponse](#provenance-marker-v1-MsgUpdateForcedTransferResponse)
    - [MsgUpdateParamsRequest](#provenance-marker-v1-MsgUpdateParamsRequest)
    - [MsgUpdateParamsResponse](#provenance-marker-v1-MsgUpdateParamsResponse)
    - [MsgUpdateRequiredAttributesRequest](#provenance-marker-v1-MsgUpdateRequiredAttributesRequest)
    - [MsgUpdateRequiredAttributesResponse](#provenance-marker-v1-MsgUpdateRequiredAttributesResponse)
    - [MsgUpdateSendDenyListRequest](#provenance-marker-v1-MsgUpdateSendDenyListRequest)
    - [MsgUpdateSendDenyListResponse](#provenance-marker-v1-MsgUpdateSendDenyListResponse)
    - [MsgWithdrawEscrowProposalRequest](#provenance-marker-v1-MsgWithdrawEscrowProposalRequest)
    - [MsgWithdrawEscrowProposalResponse](#provenance-marker-v1-MsgWithdrawEscrowProposalResponse)
    - [MsgWithdrawRequest](#provenance-marker-v1-MsgWithdrawRequest)
    - [MsgWithdrawResponse](#provenance-marker-v1-MsgWithdrawResponse)
  
    - [Msg](#provenance-marker-v1-Msg)
  
- [provenance/marker/v1/si.proto](#provenance_marker_v1_si-proto)
    - [SIPrefix](#provenance-marker-v1-SIPrefix)
  
- [provenance/marker/v1/marker.proto](#provenance_marker_v1_marker-proto)
    - [EventDenomUnit](#provenance-marker-v1-EventDenomUnit)
    - [EventMarkerAccess](#provenance-marker-v1-EventMarkerAccess)
    - [EventMarkerActivate](#provenance-marker-v1-EventMarkerActivate)
    - [EventMarkerAdd](#provenance-marker-v1-EventMarkerAdd)
    - [EventMarkerAddAccess](#provenance-marker-v1-EventMarkerAddAccess)
    - [EventMarkerBurn](#provenance-marker-v1-EventMarkerBurn)
    - [EventMarkerCancel](#provenance-marker-v1-EventMarkerCancel)
    - [EventMarkerDelete](#provenance-marker-v1-EventMarkerDelete)
    - [EventMarkerDeleteAccess](#provenance-marker-v1-EventMarkerDeleteAccess)
    - [EventMarkerFinalize](#provenance-marker-v1-EventMarkerFinalize)
    - [EventMarkerMint](#provenance-marker-v1-EventMarkerMint)
    - [EventMarkerParamsUpdated](#provenance-marker-v1-EventMarkerParamsUpdated)
    - [EventMarkerSetDenomMetadata](#provenance-marker-v1-EventMarkerSetDenomMetadata)
    - [EventMarkerTransfer](#provenance-marker-v1-EventMarkerTransfer)
    - [EventMarkerWithdraw](#provenance-marker-v1-EventMarkerWithdraw)
    - [EventSetNetAssetValue](#provenance-marker-v1-EventSetNetAssetValue)
    - [MarkerAccount](#provenance-marker-v1-MarkerAccount)
    - [NetAssetValue](#provenance-marker-v1-NetAssetValue)
    - [Params](#provenance-marker-v1-Params)
  
    - [MarkerStatus](#provenance-marker-v1-MarkerStatus)
    - [MarkerType](#provenance-marker-v1-MarkerType)
  
- [provenance/marker/v1/query.proto](#provenance_marker_v1_query-proto)
    - [Balance](#provenance-marker-v1-Balance)
    - [QueryAccessRequest](#provenance-marker-v1-QueryAccessRequest)
    - [QueryAccessResponse](#provenance-marker-v1-QueryAccessResponse)
    - [QueryAccountDataRequest](#provenance-marker-v1-QueryAccountDataRequest)
    - [QueryAccountDataResponse](#provenance-marker-v1-QueryAccountDataResponse)
    - [QueryAllMarkersRequest](#provenance-marker-v1-QueryAllMarkersRequest)
    - [QueryAllMarkersResponse](#provenance-marker-v1-QueryAllMarkersResponse)
    - [QueryDenomMetadataRequest](#provenance-marker-v1-QueryDenomMetadataRequest)
    - [QueryDenomMetadataResponse](#provenance-marker-v1-QueryDenomMetadataResponse)
    - [QueryEscrowRequest](#provenance-marker-v1-QueryEscrowRequest)
    - [QueryEscrowResponse](#provenance-marker-v1-QueryEscrowResponse)
    - [QueryHoldingRequest](#provenance-marker-v1-QueryHoldingRequest)
    - [QueryHoldingResponse](#provenance-marker-v1-QueryHoldingResponse)
    - [QueryMarkerRequest](#provenance-marker-v1-QueryMarkerRequest)
    - [QueryMarkerResponse](#provenance-marker-v1-QueryMarkerResponse)
    - [QueryNetAssetValuesRequest](#provenance-marker-v1-QueryNetAssetValuesRequest)
    - [QueryNetAssetValuesResponse](#provenance-marker-v1-QueryNetAssetValuesResponse)
    - [QueryParamsRequest](#provenance-marker-v1-QueryParamsRequest)
    - [QueryParamsResponse](#provenance-marker-v1-QueryParamsResponse)
    - [QuerySupplyRequest](#provenance-marker-v1-QuerySupplyRequest)
    - [QuerySupplyResponse](#provenance-marker-v1-QuerySupplyResponse)
  
    - [Query](#provenance-marker-v1-Query)
  
- [provenance/marker/v1/accessgrant.proto](#provenance_marker_v1_accessgrant-proto)
    - [AccessGrant](#provenance-marker-v1-AccessGrant)
  
    - [Access](#provenance-marker-v1-Access)
  
- [provenance/marker/v1/authz.proto](#provenance_marker_v1_authz-proto)
    - [MarkerTransferAuthorization](#provenance-marker-v1-MarkerTransferAuthorization)
  
- [provenance/marker/v1/genesis.proto](#provenance_marker_v1_genesis-proto)
    - [DenySendAddress](#provenance-marker-v1-DenySendAddress)
    - [GenesisState](#provenance-marker-v1-GenesisState)
    - [MarkerNetAssetValues](#provenance-marker-v1-MarkerNetAssetValues)
  
- [provenance/marker/v1/proposals.proto](#provenance_marker_v1_proposals-proto)
    - [AddMarkerProposal](#provenance-marker-v1-AddMarkerProposal)
    - [ChangeStatusProposal](#provenance-marker-v1-ChangeStatusProposal)
    - [RemoveAdministratorProposal](#provenance-marker-v1-RemoveAdministratorProposal)
    - [SetAdministratorProposal](#provenance-marker-v1-SetAdministratorProposal)
    - [SetDenomMetadataProposal](#provenance-marker-v1-SetDenomMetadataProposal)
    - [SupplyDecreaseProposal](#provenance-marker-v1-SupplyDecreaseProposal)
    - [SupplyIncreaseProposal](#provenance-marker-v1-SupplyIncreaseProposal)
    - [WithdrawEscrowProposal](#provenance-marker-v1-WithdrawEscrowProposal)
  
- [provenance/name/v1/tx.proto](#provenance_name_v1_tx-proto)
    - [MsgBindNameRequest](#provenance-name-v1-MsgBindNameRequest)
    - [MsgBindNameResponse](#provenance-name-v1-MsgBindNameResponse)
    - [MsgCreateRootNameRequest](#provenance-name-v1-MsgCreateRootNameRequest)
    - [MsgCreateRootNameResponse](#provenance-name-v1-MsgCreateRootNameResponse)
    - [MsgDeleteNameRequest](#provenance-name-v1-MsgDeleteNameRequest)
    - [MsgDeleteNameResponse](#provenance-name-v1-MsgDeleteNameResponse)
    - [MsgModifyNameRequest](#provenance-name-v1-MsgModifyNameRequest)
    - [MsgModifyNameResponse](#provenance-name-v1-MsgModifyNameResponse)
    - [MsgUpdateParamsRequest](#provenance-name-v1-MsgUpdateParamsRequest)
    - [MsgUpdateParamsResponse](#provenance-name-v1-MsgUpdateParamsResponse)
  
    - [Msg](#provenance-name-v1-Msg)
  
- [provenance/name/v1/name.proto](#provenance_name_v1_name-proto)
    - [CreateRootNameProposal](#provenance-name-v1-CreateRootNameProposal)
    - [EventNameBound](#provenance-name-v1-EventNameBound)
    - [EventNameParamsUpdated](#provenance-name-v1-EventNameParamsUpdated)
    - [EventNameUnbound](#provenance-name-v1-EventNameUnbound)
    - [EventNameUpdate](#provenance-name-v1-EventNameUpdate)
    - [NameRecord](#provenance-name-v1-NameRecord)
    - [Params](#provenance-name-v1-Params)
  
- [provenance/name/v1/query.proto](#provenance_name_v1_query-proto)
    - [QueryParamsRequest](#provenance-name-v1-QueryParamsRequest)
    - [QueryParamsResponse](#provenance-name-v1-QueryParamsResponse)
    - [QueryResolveRequest](#provenance-name-v1-QueryResolveRequest)
    - [QueryResolveResponse](#provenance-name-v1-QueryResolveResponse)
    - [QueryReverseLookupRequest](#provenance-name-v1-QueryReverseLookupRequest)
    - [QueryReverseLookupResponse](#provenance-name-v1-QueryReverseLookupResponse)
  
    - [Query](#provenance-name-v1-Query)
  
- [provenance/name/v1/genesis.proto](#provenance_name_v1_genesis-proto)
    - [GenesisState](#provenance-name-v1-GenesisState)
  
- [provenance/metadata/v1/tx.proto](#provenance_metadata_v1_tx-proto)
    - [MsgAddContractSpecToScopeSpecRequest](#provenance-metadata-v1-MsgAddContractSpecToScopeSpecRequest)
    - [MsgAddContractSpecToScopeSpecResponse](#provenance-metadata-v1-MsgAddContractSpecToScopeSpecResponse)
    - [MsgAddNetAssetValuesRequest](#provenance-metadata-v1-MsgAddNetAssetValuesRequest)
    - [MsgAddNetAssetValuesResponse](#provenance-metadata-v1-MsgAddNetAssetValuesResponse)
    - [MsgAddScopeDataAccessRequest](#provenance-metadata-v1-MsgAddScopeDataAccessRequest)
    - [MsgAddScopeDataAccessResponse](#provenance-metadata-v1-MsgAddScopeDataAccessResponse)
    - [MsgAddScopeOwnerRequest](#provenance-metadata-v1-MsgAddScopeOwnerRequest)
    - [MsgAddScopeOwnerResponse](#provenance-metadata-v1-MsgAddScopeOwnerResponse)
    - [MsgBindOSLocatorRequest](#provenance-metadata-v1-MsgBindOSLocatorRequest)
    - [MsgBindOSLocatorResponse](#provenance-metadata-v1-MsgBindOSLocatorResponse)
    - [MsgDeleteContractSpecFromScopeSpecRequest](#provenance-metadata-v1-MsgDeleteContractSpecFromScopeSpecRequest)
    - [MsgDeleteContractSpecFromScopeSpecResponse](#provenance-metadata-v1-MsgDeleteContractSpecFromScopeSpecResponse)
    - [MsgDeleteContractSpecificationRequest](#provenance-metadata-v1-MsgDeleteContractSpecificationRequest)
    - [MsgDeleteContractSpecificationResponse](#provenance-metadata-v1-MsgDeleteContractSpecificationResponse)
    - [MsgDeleteOSLocatorRequest](#provenance-metadata-v1-MsgDeleteOSLocatorRequest)
    - [MsgDeleteOSLocatorResponse](#provenance-metadata-v1-MsgDeleteOSLocatorResponse)
    - [MsgDeleteRecordRequest](#provenance-metadata-v1-MsgDeleteRecordRequest)
    - [MsgDeleteRecordResponse](#provenance-metadata-v1-MsgDeleteRecordResponse)
    - [MsgDeleteRecordSpecificationRequest](#provenance-metadata-v1-MsgDeleteRecordSpecificationRequest)
    - [MsgDeleteRecordSpecificationResponse](#provenance-metadata-v1-MsgDeleteRecordSpecificationResponse)
    - [MsgDeleteScopeDataAccessRequest](#provenance-metadata-v1-MsgDeleteScopeDataAccessRequest)
    - [MsgDeleteScopeDataAccessResponse](#provenance-metadata-v1-MsgDeleteScopeDataAccessResponse)
    - [MsgDeleteScopeOwnerRequest](#provenance-metadata-v1-MsgDeleteScopeOwnerRequest)
    - [MsgDeleteScopeOwnerResponse](#provenance-metadata-v1-MsgDeleteScopeOwnerResponse)
    - [MsgDeleteScopeRequest](#provenance-metadata-v1-MsgDeleteScopeRequest)
    - [MsgDeleteScopeResponse](#provenance-metadata-v1-MsgDeleteScopeResponse)
    - [MsgDeleteScopeSpecificationRequest](#provenance-metadata-v1-MsgDeleteScopeSpecificationRequest)
    - [MsgDeleteScopeSpecificationResponse](#provenance-metadata-v1-MsgDeleteScopeSpecificationResponse)
    - [MsgMigrateValueOwnerRequest](#provenance-metadata-v1-MsgMigrateValueOwnerRequest)
    - [MsgMigrateValueOwnerResponse](#provenance-metadata-v1-MsgMigrateValueOwnerResponse)
    - [MsgModifyOSLocatorRequest](#provenance-metadata-v1-MsgModifyOSLocatorRequest)
    - [MsgModifyOSLocatorResponse](#provenance-metadata-v1-MsgModifyOSLocatorResponse)
    - [MsgP8eMemorializeContractRequest](#provenance-metadata-v1-MsgP8eMemorializeContractRequest)
    - [MsgP8eMemorializeContractResponse](#provenance-metadata-v1-MsgP8eMemorializeContractResponse)
    - [MsgSetAccountDataRequest](#provenance-metadata-v1-MsgSetAccountDataRequest)
    - [MsgSetAccountDataResponse](#provenance-metadata-v1-MsgSetAccountDataResponse)
    - [MsgUpdateValueOwnersRequest](#provenance-metadata-v1-MsgUpdateValueOwnersRequest)
    - [MsgUpdateValueOwnersResponse](#provenance-metadata-v1-MsgUpdateValueOwnersResponse)
    - [MsgWriteContractSpecificationRequest](#provenance-metadata-v1-MsgWriteContractSpecificationRequest)
    - [MsgWriteContractSpecificationResponse](#provenance-metadata-v1-MsgWriteContractSpecificationResponse)
    - [MsgWriteP8eContractSpecRequest](#provenance-metadata-v1-MsgWriteP8eContractSpecRequest)
    - [MsgWriteP8eContractSpecResponse](#provenance-metadata-v1-MsgWriteP8eContractSpecResponse)
    - [MsgWriteRecordRequest](#provenance-metadata-v1-MsgWriteRecordRequest)
    - [MsgWriteRecordResponse](#provenance-metadata-v1-MsgWriteRecordResponse)
    - [MsgWriteRecordSpecificationRequest](#provenance-metadata-v1-MsgWriteRecordSpecificationRequest)
    - [MsgWriteRecordSpecificationResponse](#provenance-metadata-v1-MsgWriteRecordSpecificationResponse)
    - [MsgWriteScopeRequest](#provenance-metadata-v1-MsgWriteScopeRequest)
    - [MsgWriteScopeResponse](#provenance-metadata-v1-MsgWriteScopeResponse)
    - [MsgWriteScopeSpecificationRequest](#provenance-metadata-v1-MsgWriteScopeSpecificationRequest)
    - [MsgWriteScopeSpecificationResponse](#provenance-metadata-v1-MsgWriteScopeSpecificationResponse)
    - [MsgWriteSessionRequest](#provenance-metadata-v1-MsgWriteSessionRequest)
    - [MsgWriteSessionResponse](#provenance-metadata-v1-MsgWriteSessionResponse)
    - [SessionIdComponents](#provenance-metadata-v1-SessionIdComponents)
  
    - [Msg](#provenance-metadata-v1-Msg)
  
- [provenance/metadata/v1/events.proto](#provenance_metadata_v1_events-proto)
    - [EventContractSpecificationCreated](#provenance-metadata-v1-EventContractSpecificationCreated)
    - [EventContractSpecificationDeleted](#provenance-metadata-v1-EventContractSpecificationDeleted)
    - [EventContractSpecificationUpdated](#provenance-metadata-v1-EventContractSpecificationUpdated)
    - [EventOSLocatorCreated](#provenance-metadata-v1-EventOSLocatorCreated)
    - [EventOSLocatorDeleted](#provenance-metadata-v1-EventOSLocatorDeleted)
    - [EventOSLocatorUpdated](#provenance-metadata-v1-EventOSLocatorUpdated)
    - [EventRecordCreated](#provenance-metadata-v1-EventRecordCreated)
    - [EventRecordDeleted](#provenance-metadata-v1-EventRecordDeleted)
    - [EventRecordSpecificationCreated](#provenance-metadata-v1-EventRecordSpecificationCreated)
    - [EventRecordSpecificationDeleted](#provenance-metadata-v1-EventRecordSpecificationDeleted)
    - [EventRecordSpecificationUpdated](#provenance-metadata-v1-EventRecordSpecificationUpdated)
    - [EventRecordUpdated](#provenance-metadata-v1-EventRecordUpdated)
    - [EventScopeCreated](#provenance-metadata-v1-EventScopeCreated)
    - [EventScopeDeleted](#provenance-metadata-v1-EventScopeDeleted)
    - [EventScopeSpecificationCreated](#provenance-metadata-v1-EventScopeSpecificationCreated)
    - [EventScopeSpecificationDeleted](#provenance-metadata-v1-EventScopeSpecificationDeleted)
    - [EventScopeSpecificationUpdated](#provenance-metadata-v1-EventScopeSpecificationUpdated)
    - [EventScopeUpdated](#provenance-metadata-v1-EventScopeUpdated)
    - [EventSessionCreated](#provenance-metadata-v1-EventSessionCreated)
    - [EventSessionDeleted](#provenance-metadata-v1-EventSessionDeleted)
    - [EventSessionUpdated](#provenance-metadata-v1-EventSessionUpdated)
    - [EventSetNetAssetValue](#provenance-metadata-v1-EventSetNetAssetValue)
    - [EventTxCompleted](#provenance-metadata-v1-EventTxCompleted)
  
- [provenance/metadata/v1/specification.proto](#provenance_metadata_v1_specification-proto)
    - [ContractSpecification](#provenance-metadata-v1-ContractSpecification)
    - [Description](#provenance-metadata-v1-Description)
    - [InputSpecification](#provenance-metadata-v1-InputSpecification)
    - [RecordSpecification](#provenance-metadata-v1-RecordSpecification)
    - [ScopeSpecification](#provenance-metadata-v1-ScopeSpecification)
  
    - [DefinitionType](#provenance-metadata-v1-DefinitionType)
    - [PartyType](#provenance-metadata-v1-PartyType)
  
- [provenance/metadata/v1/scope.proto](#provenance_metadata_v1_scope-proto)
    - [AuditFields](#provenance-metadata-v1-AuditFields)
    - [NetAssetValue](#provenance-metadata-v1-NetAssetValue)
    - [Party](#provenance-metadata-v1-Party)
    - [Process](#provenance-metadata-v1-Process)
    - [Record](#provenance-metadata-v1-Record)
    - [RecordInput](#provenance-metadata-v1-RecordInput)
    - [RecordOutput](#provenance-metadata-v1-RecordOutput)
    - [Scope](#provenance-metadata-v1-Scope)
    - [Session](#provenance-metadata-v1-Session)
  
    - [RecordInputStatus](#provenance-metadata-v1-RecordInputStatus)
    - [ResultStatus](#provenance-metadata-v1-ResultStatus)
  
- [provenance/metadata/v1/query.proto](#provenance_metadata_v1_query-proto)
    - [AccountDataRequest](#provenance-metadata-v1-AccountDataRequest)
    - [AccountDataResponse](#provenance-metadata-v1-AccountDataResponse)
    - [ContractSpecificationRequest](#provenance-metadata-v1-ContractSpecificationRequest)
    - [ContractSpecificationResponse](#provenance-metadata-v1-ContractSpecificationResponse)
    - [ContractSpecificationWrapper](#provenance-metadata-v1-ContractSpecificationWrapper)
    - [ContractSpecificationsAllRequest](#provenance-metadata-v1-ContractSpecificationsAllRequest)
    - [ContractSpecificationsAllResponse](#provenance-metadata-v1-ContractSpecificationsAllResponse)
    - [GetByAddrRequest](#provenance-metadata-v1-GetByAddrRequest)
    - [GetByAddrResponse](#provenance-metadata-v1-GetByAddrResponse)
    - [OSAllLocatorsRequest](#provenance-metadata-v1-OSAllLocatorsRequest)
    - [OSAllLocatorsResponse](#provenance-metadata-v1-OSAllLocatorsResponse)
    - [OSLocatorParamsRequest](#provenance-metadata-v1-OSLocatorParamsRequest)
    - [OSLocatorParamsResponse](#provenance-metadata-v1-OSLocatorParamsResponse)
    - [OSLocatorRequest](#provenance-metadata-v1-OSLocatorRequest)
    - [OSLocatorResponse](#provenance-metadata-v1-OSLocatorResponse)
    - [OSLocatorsByScopeRequest](#provenance-metadata-v1-OSLocatorsByScopeRequest)
    - [OSLocatorsByScopeResponse](#provenance-metadata-v1-OSLocatorsByScopeResponse)
    - [OSLocatorsByURIRequest](#provenance-metadata-v1-OSLocatorsByURIRequest)
    - [OSLocatorsByURIResponse](#provenance-metadata-v1-OSLocatorsByURIResponse)
    - [OwnershipRequest](#provenance-metadata-v1-OwnershipRequest)
    - [OwnershipResponse](#provenance-metadata-v1-OwnershipResponse)
    - [QueryParamsRequest](#provenance-metadata-v1-QueryParamsRequest)
    - [QueryParamsResponse](#provenance-metadata-v1-QueryParamsResponse)
    - [QueryScopeNetAssetValuesRequest](#provenance-metadata-v1-QueryScopeNetAssetValuesRequest)
    - [QueryScopeNetAssetValuesResponse](#provenance-metadata-v1-QueryScopeNetAssetValuesResponse)
    - [RecordSpecificationRequest](#provenance-metadata-v1-RecordSpecificationRequest)
    - [RecordSpecificationResponse](#provenance-metadata-v1-RecordSpecificationResponse)
    - [RecordSpecificationWrapper](#provenance-metadata-v1-RecordSpecificationWrapper)
    - [RecordSpecificationsAllRequest](#provenance-metadata-v1-RecordSpecificationsAllRequest)
    - [RecordSpecificationsAllResponse](#provenance-metadata-v1-RecordSpecificationsAllResponse)
    - [RecordSpecificationsForContractSpecificationRequest](#provenance-metadata-v1-RecordSpecificationsForContractSpecificationRequest)
    - [RecordSpecificationsForContractSpecificationResponse](#provenance-metadata-v1-RecordSpecificationsForContractSpecificationResponse)
    - [RecordWrapper](#provenance-metadata-v1-RecordWrapper)
    - [RecordsAllRequest](#provenance-metadata-v1-RecordsAllRequest)
    - [RecordsAllResponse](#provenance-metadata-v1-RecordsAllResponse)
    - [RecordsRequest](#provenance-metadata-v1-RecordsRequest)
    - [RecordsResponse](#provenance-metadata-v1-RecordsResponse)
    - [ScopeRequest](#provenance-metadata-v1-ScopeRequest)
    - [ScopeResponse](#provenance-metadata-v1-ScopeResponse)
    - [ScopeSpecificationRequest](#provenance-metadata-v1-ScopeSpecificationRequest)
    - [ScopeSpecificationResponse](#provenance-metadata-v1-ScopeSpecificationResponse)
    - [ScopeSpecificationWrapper](#provenance-metadata-v1-ScopeSpecificationWrapper)
    - [ScopeSpecificationsAllRequest](#provenance-metadata-v1-ScopeSpecificationsAllRequest)
    - [ScopeSpecificationsAllResponse](#provenance-metadata-v1-ScopeSpecificationsAllResponse)
    - [ScopeWrapper](#provenance-metadata-v1-ScopeWrapper)
    - [ScopesAllRequest](#provenance-metadata-v1-ScopesAllRequest)
    - [ScopesAllResponse](#provenance-metadata-v1-ScopesAllResponse)
    - [SessionWrapper](#provenance-metadata-v1-SessionWrapper)
    - [SessionsAllRequest](#provenance-metadata-v1-SessionsAllRequest)
    - [SessionsAllResponse](#provenance-metadata-v1-SessionsAllResponse)
    - [SessionsRequest](#provenance-metadata-v1-SessionsRequest)
    - [SessionsResponse](#provenance-metadata-v1-SessionsResponse)
    - [ValueOwnershipRequest](#provenance-metadata-v1-ValueOwnershipRequest)
    - [ValueOwnershipResponse](#provenance-metadata-v1-ValueOwnershipResponse)
  
    - [Query](#provenance-metadata-v1-Query)
  
- [provenance/metadata/v1/objectstore.proto](#provenance_metadata_v1_objectstore-proto)
    - [OSLocatorParams](#provenance-metadata-v1-OSLocatorParams)
    - [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator)
  
- [provenance/metadata/v1/metadata.proto](#provenance_metadata_v1_metadata-proto)
    - [ContractSpecIdInfo](#provenance-metadata-v1-ContractSpecIdInfo)
    - [Params](#provenance-metadata-v1-Params)
    - [RecordIdInfo](#provenance-metadata-v1-RecordIdInfo)
    - [RecordSpecIdInfo](#provenance-metadata-v1-RecordSpecIdInfo)
    - [ScopeIdInfo](#provenance-metadata-v1-ScopeIdInfo)
    - [ScopeSpecIdInfo](#provenance-metadata-v1-ScopeSpecIdInfo)
    - [SessionIdInfo](#provenance-metadata-v1-SessionIdInfo)
  
- [provenance/metadata/v1/p8e/p8e.proto](#provenance_metadata_v1_p8e_p8e-proto)
    - [Condition](#provenance-metadata-v1-p8e-Condition)
    - [ConditionSpec](#provenance-metadata-v1-p8e-ConditionSpec)
    - [Consideration](#provenance-metadata-v1-p8e-Consideration)
    - [ConsiderationSpec](#provenance-metadata-v1-p8e-ConsiderationSpec)
    - [Contract](#provenance-metadata-v1-p8e-Contract)
    - [ContractSpec](#provenance-metadata-v1-p8e-ContractSpec)
    - [DefinitionSpec](#provenance-metadata-v1-p8e-DefinitionSpec)
    - [ExecutionResult](#provenance-metadata-v1-p8e-ExecutionResult)
    - [Fact](#provenance-metadata-v1-p8e-Fact)
    - [Location](#provenance-metadata-v1-p8e-Location)
    - [OutputSpec](#provenance-metadata-v1-p8e-OutputSpec)
    - [ProposedFact](#provenance-metadata-v1-p8e-ProposedFact)
    - [ProvenanceReference](#provenance-metadata-v1-p8e-ProvenanceReference)
    - [PublicKey](#provenance-metadata-v1-p8e-PublicKey)
    - [Recital](#provenance-metadata-v1-p8e-Recital)
    - [Recitals](#provenance-metadata-v1-p8e-Recitals)
    - [Signature](#provenance-metadata-v1-p8e-Signature)
    - [SignatureSet](#provenance-metadata-v1-p8e-SignatureSet)
    - [SigningAndEncryptionPublicKeys](#provenance-metadata-v1-p8e-SigningAndEncryptionPublicKeys)
    - [Timestamp](#provenance-metadata-v1-p8e-Timestamp)
    - [UUID](#provenance-metadata-v1-p8e-UUID)
  
    - [DefinitionSpecType](#provenance-metadata-v1-p8e-DefinitionSpecType)
    - [ExecutionResultType](#provenance-metadata-v1-p8e-ExecutionResultType)
    - [PartyType](#provenance-metadata-v1-p8e-PartyType)
    - [PublicKeyCurve](#provenance-metadata-v1-p8e-PublicKeyCurve)
    - [PublicKeyType](#provenance-metadata-v1-p8e-PublicKeyType)
  
- [provenance/metadata/v1/genesis.proto](#provenance_metadata_v1_genesis-proto)
    - [GenesisState](#provenance-metadata-v1-GenesisState)
    - [MarkerNetAssetValues](#provenance-metadata-v1-MarkerNetAssetValues)
  
- [provenance/hold/v1/events.proto](#provenance_hold_v1_events-proto)
    - [EventHoldAdded](#provenance-hold-v1-EventHoldAdded)
    - [EventHoldReleased](#provenance-hold-v1-EventHoldReleased)
  
- [provenance/hold/v1/hold.proto](#provenance_hold_v1_hold-proto)
    - [AccountHold](#provenance-hold-v1-AccountHold)
  
- [provenance/hold/v1/query.proto](#provenance_hold_v1_query-proto)
    - [GetAllHoldsRequest](#provenance-hold-v1-GetAllHoldsRequest)
    - [GetAllHoldsResponse](#provenance-hold-v1-GetAllHoldsResponse)
    - [GetHoldsRequest](#provenance-hold-v1-GetHoldsRequest)
    - [GetHoldsResponse](#provenance-hold-v1-GetHoldsResponse)
  
    - [Query](#provenance-hold-v1-Query)
  
- [provenance/hold/v1/genesis.proto](#provenance_hold_v1_genesis-proto)
    - [GenesisState](#provenance-hold-v1-GenesisState)
  
- [Scalar Value Types](#scalar-value-types)



<a name="cosmos_quarantine_v1beta1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/quarantine/v1beta1/tx.proto



<a name="cosmos-quarantine-v1beta1-MsgAccept"></a>

### MsgAccept
MsgAccept represents a message for accepting quarantined funds.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  | to_address is the address of the quarantined account that is accepting funds. |
| `from_addresses` | [string](#string) | repeated | from_addresses is one or more addresses that have sent funds to the quarantined account. All funds quarantined for to_address from any from_addresses are marked as accepted and released if appropriate. At least one is required. |
| `permanent` | [bool](#bool) |  | permanent, if true, sets up auto-accept for the to_address from each from_address. If false (default), only the currently quarantined funds will be accepted. |






<a name="cosmos-quarantine-v1beta1-MsgAcceptResponse"></a>

### MsgAcceptResponse
MsgAcceptResponse defines the Msg/Accept response type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `funds_released` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | funds_released is the amount that was quarantined but has now been released and sent to the requester. |






<a name="cosmos-quarantine-v1beta1-MsgDecline"></a>

### MsgDecline
MsgDecline represents a message for declining quarantined funds.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  | to_address is the address of the quarantined account that is accepting funds. |
| `from_addresses` | [string](#string) | repeated | from_addresses is one or more addresses that have sent funds to the quarantined account. All funds quarantined for to_address from any from_addresses are marked as declined. At least one is required. |
| `permanent` | [bool](#bool) |  | permanent, if true, sets up auto-decline for the to_address from each from_address. If false (default), only the currently quarantined funds will be declined. |






<a name="cosmos-quarantine-v1beta1-MsgDeclineResponse"></a>

### MsgDeclineResponse
MsgDeclineResponse defines the Msg/Decline response type.






<a name="cosmos-quarantine-v1beta1-MsgOptIn"></a>

### MsgOptIn
MsgOptIn represents a message for opting in to account quarantine.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  |  |






<a name="cosmos-quarantine-v1beta1-MsgOptInResponse"></a>

### MsgOptInResponse
MsgOptInResponse defines the Msg/OptIn response type.






<a name="cosmos-quarantine-v1beta1-MsgOptOut"></a>

### MsgOptOut
MsgOptOut represents a message for opting in to account quarantine.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  |  |






<a name="cosmos-quarantine-v1beta1-MsgOptOutResponse"></a>

### MsgOptOutResponse
MsgOptOutResponse defines the Msg/OptOut response type.






<a name="cosmos-quarantine-v1beta1-MsgUpdateAutoResponses"></a>

### MsgUpdateAutoResponses
MsgUpdateAutoResponses represents a message for updating quarantine auto-responses for a receiving address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  | to_address is the quarantined address that would be accepting or declining funds. |
| `updates` | [AutoResponseUpdate](#cosmos-quarantine-v1beta1-AutoResponseUpdate) | repeated | updates is the list of addresses and auto-responses that should be updated for the to_address. |






<a name="cosmos-quarantine-v1beta1-MsgUpdateAutoResponsesResponse"></a>

### MsgUpdateAutoResponsesResponse
MsgUpdateAutoResponsesResponse defines the Msg/UpdateAutoResponse response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="cosmos-quarantine-v1beta1-Msg"></a>

### Msg
Query defines the quarantine gRPC msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `OptIn` | [MsgOptIn](#cosmos-quarantine-v1beta1-MsgOptIn) | [MsgOptInResponse](#cosmos-quarantine-v1beta1-MsgOptInResponse) | OptIn defines a method for opting in to account quarantine. Funds sent to a quarantined account must be approved before they can be received. |
| `OptOut` | [MsgOptOut](#cosmos-quarantine-v1beta1-MsgOptOut) | [MsgOptOutResponse](#cosmos-quarantine-v1beta1-MsgOptOutResponse) | OptOut defines a method for opting out of account quarantine. Any pending funds for the account must still be accepted, but new sends will no longer be quarantined. |
| `Accept` | [MsgAccept](#cosmos-quarantine-v1beta1-MsgAccept) | [MsgAcceptResponse](#cosmos-quarantine-v1beta1-MsgAcceptResponse) | Accept defines a method for accepting quarantined funds. |
| `Decline` | [MsgDecline](#cosmos-quarantine-v1beta1-MsgDecline) | [MsgDeclineResponse](#cosmos-quarantine-v1beta1-MsgDeclineResponse) | Decline defines a method for declining quarantined funds. |
| `UpdateAutoResponses` | [MsgUpdateAutoResponses](#cosmos-quarantine-v1beta1-MsgUpdateAutoResponses) | [MsgUpdateAutoResponsesResponse](#cosmos-quarantine-v1beta1-MsgUpdateAutoResponsesResponse) | UpdateAutoResponses defines a method for updating the auto-response settings for a quarantined address. |

 <!-- end services -->



<a name="cosmos_quarantine_v1beta1_events-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/quarantine/v1beta1/events.proto



<a name="cosmos-quarantine-v1beta1-EventFundsQuarantined"></a>

### EventFundsQuarantined
EventFundsQuarantined is an event emitted when funds are quarantined.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated |  |






<a name="cosmos-quarantine-v1beta1-EventFundsReleased"></a>

### EventFundsReleased
EventFundsReleased is an event emitted when quarantined funds are accepted and released.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  |  |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated |  |






<a name="cosmos-quarantine-v1beta1-EventOptIn"></a>

### EventOptIn
EventOptIn is an event emitted when an address opts into quarantine.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  |  |






<a name="cosmos-quarantine-v1beta1-EventOptOut"></a>

### EventOptOut
EventOptOut is an event emitted when an address opts out of quarantine.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="cosmos_quarantine_v1beta1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/quarantine/v1beta1/query.proto



<a name="cosmos-quarantine-v1beta1-QueryAutoResponsesRequest"></a>

### QueryAutoResponsesRequest
QueryAutoResponsesRequest defines the RPC request for getting auto-response settings for an address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  | to_address is the quarantined account to get info on. |
| `from_address` | [string](#string) |  | from_address is an optional sender address to limit results. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="cosmos-quarantine-v1beta1-QueryAutoResponsesResponse"></a>

### QueryAutoResponsesResponse
QueryAutoResponsesResponse defines the RPC response of a AutoResponses query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `auto_responses` | [AutoResponseEntry](#cosmos-quarantine-v1beta1-AutoResponseEntry) | repeated | auto_responses are the auto-response entries from the provided query. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines the pagination parameters of the response. |






<a name="cosmos-quarantine-v1beta1-QueryIsQuarantinedRequest"></a>

### QueryIsQuarantinedRequest
QueryIsQuarantinedRequest defines the RPC request for checking if an account has opted into quarantine.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  | to_address is the address to check. |






<a name="cosmos-quarantine-v1beta1-QueryIsQuarantinedResponse"></a>

### QueryIsQuarantinedResponse
QueryIsQuarantinedResponse defines the RPC response of an IsQuarantined query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `is_quarantined` | [bool](#bool) |  | is_quarantined is true if the to_address has opted into quarantine. |






<a name="cosmos-quarantine-v1beta1-QueryQuarantinedFundsRequest"></a>

### QueryQuarantinedFundsRequest
QueryQuarantinedFundsRequest defines the RPC request for looking up quarantined funds.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  | to_address is the intended recipient of the coins that have been quarantined. |
| `from_address` | [string](#string) |  | from_address is the sender of the coins. If provided, a to_address must also be provided. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="cosmos-quarantine-v1beta1-QueryQuarantinedFundsResponse"></a>

### QueryQuarantinedFundsResponse
QueryQuarantinedFundsResponse defines the RPC response of a QuarantinedFunds query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `quarantinedFunds` | [QuarantinedFunds](#cosmos-quarantine-v1beta1-QuarantinedFunds) | repeated | quarantinedFunds is info about coins sitting in quarantine. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines the pagination parameters of the response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="cosmos-quarantine-v1beta1-Query"></a>

### Query
Query defines the quarantine gRPC query service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `IsQuarantined` | [QueryIsQuarantinedRequest](#cosmos-quarantine-v1beta1-QueryIsQuarantinedRequest) | [QueryIsQuarantinedResponse](#cosmos-quarantine-v1beta1-QueryIsQuarantinedResponse) | IsQuarantined checks if an account has opted into quarantine. |
| `QuarantinedFunds` | [QueryQuarantinedFundsRequest](#cosmos-quarantine-v1beta1-QueryQuarantinedFundsRequest) | [QueryQuarantinedFundsResponse](#cosmos-quarantine-v1beta1-QueryQuarantinedFundsResponse) | QuarantinedFunds gets information about funds that have been quarantined.<br>If both a to_address and from_address are provided, any such quarantined funds will be returned regardless of whether they've been declined. If only a to_address is provided, the unaccepted and undeclined funds waiting on a response from to_address will be returned. If neither a to_address nor from_address is provided, all non-declined quarantined funds for any address will be returned. The request is invalid if only a from_address is provided. |
| `AutoResponses` | [QueryAutoResponsesRequest](#cosmos-quarantine-v1beta1-QueryAutoResponsesRequest) | [QueryAutoResponsesResponse](#cosmos-quarantine-v1beta1-QueryAutoResponsesResponse) | AutoResponses gets the auto-response settings for a quarantined account.<br>The to_address is required. If a from_address is provided only the auto response for that from_address will be returned. If no from_address is provided, all auto-response settings for the given to_address will be returned. |

 <!-- end services -->



<a name="cosmos_quarantine_v1beta1_quarantine-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/quarantine/v1beta1/quarantine.proto



<a name="cosmos-quarantine-v1beta1-AutoResponseEntry"></a>

### AutoResponseEntry
AutoResponseEntry defines the auto response to one address from another.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  | to_address is the receiving address. |
| `from_address` | [string](#string) |  | from_address is the sending address. |
| `response` | [AutoResponse](#cosmos-quarantine-v1beta1-AutoResponse) |  | response is the auto-response setting for these two addresses. |






<a name="cosmos-quarantine-v1beta1-AutoResponseUpdate"></a>

### AutoResponseUpdate
AutoResponseUpdate defines a quarantine auto response update that should be applied.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from_address` | [string](#string) |  | from_address is the address that funds would be coming from. |
| `response` | [AutoResponse](#cosmos-quarantine-v1beta1-AutoResponse) |  | response is the automatic action to take on funds sent from from_address. Provide AUTO_RESPONSE_UNSPECIFIED to turn off an auto-response. |






<a name="cosmos-quarantine-v1beta1-QuarantineRecord"></a>

### QuarantineRecord
QuarantineRecord defines information regarding quarantined funds that is stored in state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `unaccepted_from_addresses` | [bytes](#bytes) | repeated | unaccepted_from_addresses are the senders that have not been part of an accept yet for these coins. |
| `accepted_from_addresses` | [bytes](#bytes) | repeated | accepted_from_addresses are the senders that have already been part of an accept for these coins. |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | coins is the amount that has been quarantined. |
| `declined` | [bool](#bool) |  | declined is whether these funds have been declined. |






<a name="cosmos-quarantine-v1beta1-QuarantineRecordSuffixIndex"></a>

### QuarantineRecordSuffixIndex
QuarantineRecordSuffixIndex defines a list of record suffixes that can be stored in state and used as an index.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_suffixes` | [bytes](#bytes) | repeated |  |






<a name="cosmos-quarantine-v1beta1-QuarantinedFunds"></a>

### QuarantinedFunds
QuarantinedFunds defines structure that represents coins that have been quarantined.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `to_address` | [string](#string) |  | to_address is the intended recipient of the coins that have been quarantined. |
| `unaccepted_from_addresses` | [string](#string) | repeated | unaccepted_from_addresses are the senders that have not been part of an accept yet for these coins. |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | coins is the amount currently in quarantined for the two addresses. |
| `declined` | [bool](#bool) |  | declined is true if these funds were previously declined. |





 <!-- end messages -->


<a name="cosmos-quarantine-v1beta1-AutoResponse"></a>

### AutoResponse
AutoResponse enumerates the quarantine auto-response options.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `AUTO_RESPONSE_UNSPECIFIED` | `0` | AUTO_RESPONSE_UNSPECIFIED defines that an automatic response has not been specified. This means that no automatic action should be taken, i.e. this auto-response is off, and default quarantine behavior is used. |
| `AUTO_RESPONSE_ACCEPT` | `1` | AUTO_RESPONSE_ACCEPT defines that sends should be automatically accepted, bypassing quarantine. |
| `AUTO_RESPONSE_DECLINE` | `2` | AUTO_RESPONSE_DECLINE defines that sends should be automatically declined. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="cosmos_quarantine_v1beta1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/quarantine/v1beta1/genesis.proto



<a name="cosmos-quarantine-v1beta1-GenesisState"></a>

### GenesisState
GenesisState defines the quarantine module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `quarantined_addresses` | [string](#string) | repeated | quarantined_addresses defines account addresses that are opted into quarantine. |
| `auto_responses` | [AutoResponseEntry](#cosmos-quarantine-v1beta1-AutoResponseEntry) | repeated | auto_responses defines the quarantine auto-responses for addresses. |
| `quarantined_funds` | [QuarantinedFunds](#cosmos-quarantine-v1beta1-QuarantinedFunds) | repeated | quarantined_funds defines funds that are quarantined. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="cosmos_sanction_v1beta1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/sanction/v1beta1/tx.proto



<a name="cosmos-sanction-v1beta1-MsgSanction"></a>

### MsgSanction
MsgSanction represents a message for the governance operation of sanctioning addresses.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) | repeated | addresses are the addresses to sanction. |
| `authority` | [string](#string) |  | authority is the address of the account with the authority to enact sanctions (most likely the governance module account). |






<a name="cosmos-sanction-v1beta1-MsgSanctionResponse"></a>

### MsgSanctionResponse
MsgOptInResponse defines the Msg/Sanction response type.






<a name="cosmos-sanction-v1beta1-MsgUnsanction"></a>

### MsgUnsanction
MsgSanction represents a message for the governance operation of unsanctioning addresses.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) | repeated | addresses are the addresses to unsanction. |
| `authority` | [string](#string) |  | authority is the address of the account with the authority to retract sanctions (most likely the governance module account). |






<a name="cosmos-sanction-v1beta1-MsgUnsanctionResponse"></a>

### MsgUnsanctionResponse
MsgOptInResponse defines the Msg/Unsanction response type.






<a name="cosmos-sanction-v1beta1-MsgUpdateParams"></a>

### MsgUpdateParams
MsgUpdateParams represents a message for the governance operation of updating the sanction module params.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#cosmos-sanction-v1beta1-Params) |  | params are the sanction module parameters. |
| `authority` | [string](#string) |  | authority is the address of the account with the authority to update params (most likely the governance module account). |






<a name="cosmos-sanction-v1beta1-MsgUpdateParamsResponse"></a>

### MsgUpdateParamsResponse
MsgUpdateParamsResponse defined the Msg/UpdateParams response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="cosmos-sanction-v1beta1-Msg"></a>

### Msg
Msg defines the sanction Msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Sanction` | [MsgSanction](#cosmos-sanction-v1beta1-MsgSanction) | [MsgSanctionResponse](#cosmos-sanction-v1beta1-MsgSanctionResponse) | Sanction is a governance operation for sanctioning addresses. |
| `Unsanction` | [MsgUnsanction](#cosmos-sanction-v1beta1-MsgUnsanction) | [MsgUnsanctionResponse](#cosmos-sanction-v1beta1-MsgUnsanctionResponse) | Unsanction is a governance operation for unsanctioning addresses. |
| `UpdateParams` | [MsgUpdateParams](#cosmos-sanction-v1beta1-MsgUpdateParams) | [MsgUpdateParamsResponse](#cosmos-sanction-v1beta1-MsgUpdateParamsResponse) | UpdateParams is a governance operation for updating the sanction module params. |

 <!-- end services -->



<a name="cosmos_sanction_v1beta1_events-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/sanction/v1beta1/events.proto



<a name="cosmos-sanction-v1beta1-EventAddressSanctioned"></a>

### EventAddressSanctioned
EventAddressSanctioned is an event emitted when an address is sanctioned.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="cosmos-sanction-v1beta1-EventAddressUnsanctioned"></a>

### EventAddressUnsanctioned
EventAddressUnsanctioned is an event emitted when an address is unsanctioned.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="cosmos-sanction-v1beta1-EventParamsUpdated"></a>

### EventParamsUpdated
EventParamsUpdated is an event emitted when the sanction module params are updated.






<a name="cosmos-sanction-v1beta1-EventTempAddressSanctioned"></a>

### EventTempAddressSanctioned
EventTempAddressSanctioned is an event emitted when an address is temporarily sanctioned.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="cosmos-sanction-v1beta1-EventTempAddressUnsanctioned"></a>

### EventTempAddressUnsanctioned
EventTempAddressUnsanctioned is an event emitted when an address is temporarily unsanctioned.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="cosmos_sanction_v1beta1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/sanction/v1beta1/query.proto



<a name="cosmos-sanction-v1beta1-QueryIsSanctionedRequest"></a>

### QueryIsSanctionedRequest
QueryIsSanctionedRequest defines the RPC request for checking if an account is sanctioned.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="cosmos-sanction-v1beta1-QueryIsSanctionedResponse"></a>

### QueryIsSanctionedResponse
QueryIsSanctionedResponse defines the RPC response of an IsSanctioned query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `is_sanctioned` | [bool](#bool) |  | is_sanctioned is true if the address is sanctioned. |






<a name="cosmos-sanction-v1beta1-QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest defines the RPC request for getting the sanction module params.






<a name="cosmos-sanction-v1beta1-QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse defines the RPC response of a Params query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#cosmos-sanction-v1beta1-Params) |  | params are the sanction module parameters. |






<a name="cosmos-sanction-v1beta1-QuerySanctionedAddressesRequest"></a>

### QuerySanctionedAddressesRequest
QuerySanctionedAddressesRequest defines the RPC request for listing sanctioned accounts.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="cosmos-sanction-v1beta1-QuerySanctionedAddressesResponse"></a>

### QuerySanctionedAddressesResponse
QuerySanctionedAddressesResponse defines the RPC response of a SanctionedAddresses query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) | repeated | addresses is the list of sanctioned account addresses. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines the pagination in the response. |






<a name="cosmos-sanction-v1beta1-QueryTemporaryEntriesRequest"></a>

### QueryTemporaryEntriesRequest
QueryTemporaryEntriesRequest defines the RPC request for listing temporary sanction/unsanction entries.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is an optional address to restrict results to. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="cosmos-sanction-v1beta1-QueryTemporaryEntriesResponse"></a>

### QueryTemporaryEntriesResponse
QueryTemporaryEntriesResponse defines the RPC response of a TemporaryEntries query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entries` | [TemporaryEntry](#cosmos-sanction-v1beta1-TemporaryEntry) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines the pagination in the response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="cosmos-sanction-v1beta1-Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `IsSanctioned` | [QueryIsSanctionedRequest](#cosmos-sanction-v1beta1-QueryIsSanctionedRequest) | [QueryIsSanctionedResponse](#cosmos-sanction-v1beta1-QueryIsSanctionedResponse) | IsSanctioned checks if an account has been sanctioned. |
| `SanctionedAddresses` | [QuerySanctionedAddressesRequest](#cosmos-sanction-v1beta1-QuerySanctionedAddressesRequest) | [QuerySanctionedAddressesResponse](#cosmos-sanction-v1beta1-QuerySanctionedAddressesResponse) | SanctionedAddresses returns a list of sanctioned addresses. |
| `TemporaryEntries` | [QueryTemporaryEntriesRequest](#cosmos-sanction-v1beta1-QueryTemporaryEntriesRequest) | [QueryTemporaryEntriesResponse](#cosmos-sanction-v1beta1-QueryTemporaryEntriesResponse) | TemporaryEntries returns temporary sanction/unsanction info. |
| `Params` | [QueryParamsRequest](#cosmos-sanction-v1beta1-QueryParamsRequest) | [QueryParamsResponse](#cosmos-sanction-v1beta1-QueryParamsResponse) | Params returns the sanction module's params. |

 <!-- end services -->



<a name="cosmos_sanction_v1beta1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/sanction/v1beta1/genesis.proto



<a name="cosmos-sanction-v1beta1-GenesisState"></a>

### GenesisState
GenesisState defines the sanction module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#cosmos-sanction-v1beta1-Params) |  | params are the sanction module parameters. |
| `sanctioned_addresses` | [string](#string) | repeated | sanctioned_addresses defines account addresses that are sanctioned. |
| `temporary_entries` | [TemporaryEntry](#cosmos-sanction-v1beta1-TemporaryEntry) | repeated | temporary_entries defines the temporary entries associated with on-going governance proposals. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="cosmos_sanction_v1beta1_sanction-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## cosmos/sanction/v1beta1/sanction.proto



<a name="cosmos-sanction-v1beta1-Params"></a>

### Params
Params defines the configurable parameters of the sanction module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `immediate_sanction_min_deposit` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | immediate_sanction_min_deposit is the minimum deposit for a sanction to happen immediately. If this is zero, immediate sanctioning is not available. Otherwise, if a sanction governance proposal is issued with a deposit at least this large, a temporary sanction will be immediately issued that will expire when voting ends on the governance proposal. |
| `immediate_unsanction_min_deposit` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | immediate_unsanction_min_deposit is the minimum deposit for an unsanction to happen immediately. If this is zero, immediate unsanctioning is not available. Otherwise, if an unsanction governance proposal is issued with a deposit at least this large, a temporary unsanction will be immediately issued that will expire when voting ends on the governance proposal. |






<a name="cosmos-sanction-v1beta1-TemporaryEntry"></a>

### TemporaryEntry
TemporaryEntry defines the information involved in a temporary sanction or unsanction.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of this temporary entry. |
| `proposal_id` | [uint64](#uint64) |  | proposal_id is the governance proposal id associated with this temporary entry. |
| `status` | [TempStatus](#cosmos-sanction-v1beta1-TempStatus) |  | status is whether the entry is a sanction or unsanction. |





 <!-- end messages -->


<a name="cosmos-sanction-v1beta1-TempStatus"></a>

### TempStatus
TempStatus is whether a temporary entry is a sanction or unsanction.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `TEMP_STATUS_UNSPECIFIED` | `0` | TEMP_STATUS_UNSPECIFIED represents and unspecified status value. |
| `TEMP_STATUS_SANCTIONED` | `1` | TEMP_STATUS_SANCTIONED indicates a sanction is in place. |
| `TEMP_STATUS_UNSANCTIONED` | `2` | TEMP_STATUS_UNSANCTIONED indicates an unsanctioned is in place. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_exchange_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/tx.proto



<a name="provenance-exchange-v1-MsgAcceptPaymentRequest"></a>

### MsgAcceptPaymentRequest
MsgAcceptPaymentRequest is a request message for the AcceptPayment endpoint.

The signer is the payment.target, but we can't define that using the cosmos.msg.v1.signer option.
So signers for this msg are defined in code using a custom get-signers function.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payment` | [Payment](#provenance-exchange-v1-Payment) |  | payment is the details of the payment to accept. |






<a name="provenance-exchange-v1-MsgAcceptPaymentResponse"></a>

### MsgAcceptPaymentResponse
MsgAcceptPaymentResponse is a response message for the AcceptPayment endpoint.






<a name="provenance-exchange-v1-MsgCancelOrderRequest"></a>

### MsgCancelOrderRequest
MsgCancelOrderRequest is a request message for the CancelOrder endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signer` | [string](#string) |  | signer is the account requesting the order cancellation. It must be either the order owner (e.g. the buyer or seller), the governance module account address, or an account with cancel permission with the market that the order is in. |
| `order_id` | [uint64](#uint64) |  | order_id is the id of the order to cancel. |






<a name="provenance-exchange-v1-MsgCancelOrderResponse"></a>

### MsgCancelOrderResponse
MsgCancelOrderResponse is a response message for the CancelOrder endpoint.






<a name="provenance-exchange-v1-MsgCancelPaymentsRequest"></a>

### MsgCancelPaymentsRequest
MsgCancelPaymentsRequest is a request message for the CancelPayments endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the account that wishes to cancel some of their payments. |
| `external_ids` | [string](#string) | repeated | external_ids is all of the external ids of the payments to cancel. |






<a name="provenance-exchange-v1-MsgCancelPaymentsResponse"></a>

### MsgCancelPaymentsResponse
MsgCancelPaymentsResponse is a response message for the CancelPayments endpoint.






<a name="provenance-exchange-v1-MsgChangePaymentTargetRequest"></a>

### MsgChangePaymentTargetRequest
MsgChangePaymentTargetRequest is a request message for the ChangePaymentTarget endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the account that wishes to update the target of one of their payments. |
| `external_id` | [string](#string) |  | external_id is the external id of the payment to update. |
| `new_target` | [string](#string) |  | new_target is the new target account of the payment. |






<a name="provenance-exchange-v1-MsgChangePaymentTargetResponse"></a>

### MsgChangePaymentTargetResponse
MsgChangePaymentTargetResponse is a response message for the ChangePaymentTarget endpoint.






<a name="provenance-exchange-v1-MsgCommitFundsRequest"></a>

### MsgCommitFundsRequest
MsgCommitFundsRequest is a request message for the CommitFunds endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account is the address of the account with the funds being committed. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market the funds will be committed to. |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | amount is the funds being committed to the market. |
| `creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | creation_fee is the fee that is being paid to create this commitment. |
| `event_tag` | [string](#string) |  | event_tag is a string that is included in the funds-committed event. Max length is 100 characters. |






<a name="provenance-exchange-v1-MsgCommitFundsResponse"></a>

### MsgCommitFundsResponse
MsgCommitFundsResponse is a response message for the CommitFunds endpoint.






<a name="provenance-exchange-v1-MsgCreateAskRequest"></a>

### MsgCreateAskRequest
MsgCreateAskRequest is a request message for the CreateAsk endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ask_order` | [AskOrder](#provenance-exchange-v1-AskOrder) |  | ask_order is the details of the order being created. |
| `order_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | order_creation_fee is the fee that is being paid to create this order. |






<a name="provenance-exchange-v1-MsgCreateAskResponse"></a>

### MsgCreateAskResponse
MsgCreateAskResponse is a response message for the CreateAsk endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the id of the order created. |






<a name="provenance-exchange-v1-MsgCreateBidRequest"></a>

### MsgCreateBidRequest
MsgCreateBidRequest is a request message for the CreateBid endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bid_order` | [BidOrder](#provenance-exchange-v1-BidOrder) |  | bid_order is the details of the order being created. |
| `order_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | order_creation_fee is the fee that is being paid to create this order. |






<a name="provenance-exchange-v1-MsgCreateBidResponse"></a>

### MsgCreateBidResponse
MsgCreateBidResponse is a response message for the CreateBid endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the id of the order created. |






<a name="provenance-exchange-v1-MsgCreatePaymentRequest"></a>

### MsgCreatePaymentRequest
MsgCreatePaymentRequest is a request message for the CreatePayment endpoint.

The signer is the payment.source, but we can't define that using the cosmos.msg.v1.signer option.
So signers for this msg are defined in code using a custom get-signers function.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payment` | [Payment](#provenance-exchange-v1-Payment) |  | payment is the details of the payment to create. |






<a name="provenance-exchange-v1-MsgCreatePaymentResponse"></a>

### MsgCreatePaymentResponse
MsgCreatePaymentResponse is a response message for the CreatePayment endpoint.






<a name="provenance-exchange-v1-MsgFillAsksRequest"></a>

### MsgFillAsksRequest
MsgFillAsksRequest is a request message for the FillAsks endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `buyer` | [string](#string) |  | buyer is the address of the account attempting to buy some assets. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market with the asks to fill. All ask orders being filled must be in this market. |
| `total_price` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | total_price is the total amount being spent on some assets. It must be the sum of all ask order prices. |
| `ask_order_ids` | [uint64](#uint64) | repeated | ask_order_ids are the ids of the ask orders that you are trying to fill. All ids must be for ask orders, and must be in the same market as the market_id. |
| `buyer_settlement_fees` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | buyer_settlement_fees are the fees (both flat and proportional) that the buyer will pay (in addition to the price) for this settlement. |
| `bid_order_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | bid_order_creation_fee is the fee that is being paid to create this order (which is immediately then settled). |






<a name="provenance-exchange-v1-MsgFillAsksResponse"></a>

### MsgFillAsksResponse
MsgFillAsksResponse is a response message for the FillAsks endpoint.






<a name="provenance-exchange-v1-MsgFillBidsRequest"></a>

### MsgFillBidsRequest
MsgFillBidsRequest is a request message for the FillBids endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `seller` | [string](#string) |  | seller is the address of the account with the assets to sell. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market with the bids to fill. All bid orders being filled must be in this market. |
| `total_assets` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | total_assets are the things that the seller wishes to sell. It must be the sum of all bid order assets. |
| `bid_order_ids` | [uint64](#uint64) | repeated | bid_order_ids are the ids of the bid orders that you are trying to fill. All ids must be for bid orders, and must be in the same market as the market_id. |
| `seller_settlement_flat_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | seller_settlement_flat_fee is the flat fee for sellers that will be charged for this settlement. |
| `ask_order_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | ask_order_creation_fee is the fee that is being paid to create this order (which is immediately then settled). |






<a name="provenance-exchange-v1-MsgFillBidsResponse"></a>

### MsgFillBidsResponse
MsgFillBidsResponse is a response message for the FillBids endpoint.






<a name="provenance-exchange-v1-MsgGovCloseMarketRequest"></a>

### MsgGovCloseMarketRequest
MsgGovCloseMarketRequest is a request message for the GovCloseMarket endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority must be the governance module account. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to close. |






<a name="provenance-exchange-v1-MsgGovCloseMarketResponse"></a>

### MsgGovCloseMarketResponse
MsgGovCloseMarketResponse is a response message for the GovCloseMarket endpoint.






<a name="provenance-exchange-v1-MsgGovCreateMarketRequest"></a>

### MsgGovCreateMarketRequest
MsgGovCreateMarketRequest is a request message for the GovCreateMarket endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `market` | [Market](#provenance-exchange-v1-Market) |  | market is the initial market configuration. If the market_id is 0, the next available market_id will be used (once voting ends). If it is not zero, it must not yet be in use when the voting period ends. |






<a name="provenance-exchange-v1-MsgGovCreateMarketResponse"></a>

### MsgGovCreateMarketResponse
MsgGovCreateMarketResponse is a response message for the GovCreateMarket endpoint.






<a name="provenance-exchange-v1-MsgGovManageFeesRequest"></a>

### MsgGovManageFeesRequest
MsgGovManageFeesRequest is a request message for the GovManageFees endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `market_id` | [uint32](#uint32) |  | market_id is the market id that will get these fee updates. |
| `add_fee_create_ask_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | add_fee_create_ask_flat are the create-ask flat fee options to add. |
| `remove_fee_create_ask_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | remove_fee_create_ask_flat are the create-ask flat fee options to remove. |
| `add_fee_create_bid_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | add_fee_create_bid_flat are the create-bid flat fee options to add. |
| `remove_fee_create_bid_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | remove_fee_create_bid_flat are the create-bid flat fee options to remove. |
| `add_fee_seller_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | add_fee_seller_settlement_flat are the seller settlement flat fee options to add. |
| `remove_fee_seller_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | remove_fee_seller_settlement_flat are the seller settlement flat fee options to remove. |
| `add_fee_seller_settlement_ratios` | [FeeRatio](#provenance-exchange-v1-FeeRatio) | repeated | add_fee_seller_settlement_ratios are the seller settlement fee ratios to add. |
| `remove_fee_seller_settlement_ratios` | [FeeRatio](#provenance-exchange-v1-FeeRatio) | repeated | remove_fee_seller_settlement_ratios are the seller settlement fee ratios to remove. |
| `add_fee_buyer_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | add_fee_buyer_settlement_flat are the buyer settlement flat fee options to add. |
| `remove_fee_buyer_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | remove_fee_buyer_settlement_flat are the buyer settlement flat fee options to remove. |
| `add_fee_buyer_settlement_ratios` | [FeeRatio](#provenance-exchange-v1-FeeRatio) | repeated | add_fee_buyer_settlement_ratios are the buyer settlement fee ratios to add. |
| `remove_fee_buyer_settlement_ratios` | [FeeRatio](#provenance-exchange-v1-FeeRatio) | repeated | remove_fee_buyer_settlement_ratios are the buyer settlement fee ratios to remove. |
| `add_fee_create_commitment_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | add_fee_create_commitment_flat are the create-commitment flat fee options to add. |
| `remove_fee_create_commitment_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | remove_fee_create_commitment_flat are the create-commitment flat fee options to remove. |
| `set_fee_commitment_settlement_bips` | [uint32](#uint32) |  | set_fee_commitment_settlement_bips is the new fee_commitment_settlement_bips for the market. It is ignored if it is zero. To set it to zero set unset_fee_commitment_settlement_bips to true. |
| `unset_fee_commitment_settlement_bips` | [bool](#bool) |  | unset_fee_commitment_settlement_bips, if true, sets the fee_commitment_settlement_bips to zero. If false, it is ignored. |






<a name="provenance-exchange-v1-MsgGovManageFeesResponse"></a>

### MsgGovManageFeesResponse
MsgGovManageFeesResponse is a response message for the GovManageFees endpoint.






<a name="provenance-exchange-v1-MsgGovUpdateParamsRequest"></a>

### MsgGovUpdateParamsRequest
MsgGovUpdateParamsRequest is a request message for the GovUpdateParams endpoint.
Deprecated: Use MsgUpdateParamsRequest instead.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `params` | [Params](#provenance-exchange-v1-Params) |  | params are the new param values to set |






<a name="provenance-exchange-v1-MsgGovUpdateParamsResponse"></a>

### MsgGovUpdateParamsResponse
MsgGovUpdateParamsResponse is a response message for the GovUpdateParams endpoint.
Deprecated: Use MsgUpdateParamsResponse instead.






<a name="provenance-exchange-v1-MsgMarketCommitmentSettleRequest"></a>

### MsgMarketCommitmentSettleRequest
MsgMarketCommitmentSettleRequest is a request message for the MarketCommitmentSettle endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "settle" permission requesting this settlement. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market requesting this settlement. |
| `inputs` | [AccountAmount](#provenance-exchange-v1-AccountAmount) | repeated | inputs defines where the funds are coming from. All of these funds must be already committed to the market. |
| `outputs` | [AccountAmount](#provenance-exchange-v1-AccountAmount) | repeated | outputs defines how the funds are to be distributed. These funds will be re-committed in the destination accounts. |
| `fees` | [AccountAmount](#provenance-exchange-v1-AccountAmount) | repeated | fees is the funds that the market is collecting as part of this settlement. All of these funds must be already committed to the market. |
| `navs` | [NetAssetPrice](#provenance-exchange-v1-NetAssetPrice) | repeated | navs are any NAV info that should be updated at the beginning of this settlement. |
| `event_tag` | [string](#string) |  | event_tag is a string that is included in the funds-committed/released events. Max length is 100 characters. |






<a name="provenance-exchange-v1-MsgMarketCommitmentSettleResponse"></a>

### MsgMarketCommitmentSettleResponse
MsgMarketCommitmentSettleResponse is a response message for the MarketCommitmentSettle endpoint.






<a name="provenance-exchange-v1-MsgMarketManagePermissionsRequest"></a>

### MsgMarketManagePermissionsRequest
MsgMarketManagePermissionsRequest is a request message for the MarketManagePermissions endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "permissions" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to manage permissions for. |
| `revoke_all` | [string](#string) | repeated | revoke_all are addresses that should have all their permissions revoked. |
| `to_revoke` | [AccessGrant](#provenance-exchange-v1-AccessGrant) | repeated | to_revoke are the specific permissions to remove for addresses. |
| `to_grant` | [AccessGrant](#provenance-exchange-v1-AccessGrant) | repeated | to_grant are the permissions to grant to addresses. |






<a name="provenance-exchange-v1-MsgMarketManagePermissionsResponse"></a>

### MsgMarketManagePermissionsResponse
MsgMarketManagePermissionsResponse is a response message for the MarketManagePermissions endpoint.






<a name="provenance-exchange-v1-MsgMarketManageReqAttrsRequest"></a>

### MsgMarketManageReqAttrsRequest
MsgMarketManageReqAttrsRequest is a request message for the MarketManageReqAttrs endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "attributes" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to update required attributes for. |
| `create_ask_to_add` | [string](#string) | repeated | create_ask_to_add are the attributes that should now also be required to create an ask order. |
| `create_ask_to_remove` | [string](#string) | repeated | create_ask_to_remove are the attributes that should no longer be required to create an ask order. |
| `create_bid_to_add` | [string](#string) | repeated | create_bid_to_add are the attributes that should now also be required to create a bid order. |
| `create_bid_to_remove` | [string](#string) | repeated | create_bid_to_remove are the attributes that should no longer be required to create a bid order. |
| `create_commitment_to_add` | [string](#string) | repeated | create_commitment_to_add are the attributes that should now also be required to create a commitment. |
| `create_commitment_to_remove` | [string](#string) | repeated | create_commitment_to_remove are the attributes that should no longer be required to create a commitment. |






<a name="provenance-exchange-v1-MsgMarketManageReqAttrsResponse"></a>

### MsgMarketManageReqAttrsResponse
MsgMarketManageReqAttrsResponse is a response message for the MarketManageReqAttrs endpoint.






<a name="provenance-exchange-v1-MsgMarketReleaseCommitmentsRequest"></a>

### MsgMarketReleaseCommitmentsRequest
MsgMarketReleaseCommitmentsRequest is a request message for the MarketReleaseCommitments endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "cancel" permission requesting this release. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market releasing these funds. |
| `to_release` | [AccountAmount](#provenance-exchange-v1-AccountAmount) | repeated | to_release is the funds that are to be released. An entry with a zero amount indicates that all committed funds for that account should be released. |
| `event_tag` | [string](#string) |  | event_tag is a string that is included in the funds-released events. Max length is 100 characters. |






<a name="provenance-exchange-v1-MsgMarketReleaseCommitmentsResponse"></a>

### MsgMarketReleaseCommitmentsResponse
MsgMarketReleaseCommitmentsResponse is a response message for the MarketReleaseCommitments endpoint.






<a name="provenance-exchange-v1-MsgMarketSetOrderExternalIDRequest"></a>

### MsgMarketSetOrderExternalIDRequest
MsgMarketSetOrderExternalIDRequest is a request message for the MarketSetOrderExternalID endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "set_ids" permission requesting this settlement. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market with the orders to update. |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order to update. |
| `external_id` | [string](#string) |  | external_id is the new external id to associate with the order. Max length is 100 characters. If the external id is already associated with another order in this market, this update will fail. |






<a name="provenance-exchange-v1-MsgMarketSetOrderExternalIDResponse"></a>

### MsgMarketSetOrderExternalIDResponse
MsgMarketSetOrderExternalIDResponse is a response message for the MarketSetOrderExternalID endpoint.






<a name="provenance-exchange-v1-MsgMarketSettleRequest"></a>

### MsgMarketSettleRequest
MsgMarketSettleRequest is a request message for the MarketSettle endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "settle" permission requesting this settlement. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market requesting this settlement. |
| `ask_order_ids` | [uint64](#uint64) | repeated | ask_order_ids are the ask orders being filled. |
| `bid_order_ids` | [uint64](#uint64) | repeated | bid_order_ids are the bid orders being filled. |
| `expect_partial` | [bool](#bool) |  | expect_partial is whether to expect an order to only be partially filled. Set to true to indicate that either the last ask order, or last bid order will be partially filled by this settlement. Set to false to indicate that all provided orders will be filled in full during this settlement. |






<a name="provenance-exchange-v1-MsgMarketSettleResponse"></a>

### MsgMarketSettleResponse
MsgMarketSettleResponse is a response message for the MarketSettle endpoint.






<a name="provenance-exchange-v1-MsgMarketUpdateAcceptingCommitmentsRequest"></a>

### MsgMarketUpdateAcceptingCommitmentsRequest
MsgMarketUpdateAcceptingCommitmentsRequest is a request message for the MarketUpdateAcceptingCommitments endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "update" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to enable or disable commitments for. |
| `accepting_commitments` | [bool](#bool) |  | accepting_commitments is whether this market allows users to commit funds to it. For example, the CommitFunds endpoint is available if and only if this is true. The MarketCommitmentSettle endpoint is available (only to market actors) regardless of the value of this field. |






<a name="provenance-exchange-v1-MsgMarketUpdateAcceptingCommitmentsResponse"></a>

### MsgMarketUpdateAcceptingCommitmentsResponse
MsgMarketUpdateAcceptingCommitmentsResponse is a response message for the MarketUpdateAcceptingCommitments endpoint.






<a name="provenance-exchange-v1-MsgMarketUpdateAcceptingOrdersRequest"></a>

### MsgMarketUpdateAcceptingOrdersRequest
MsgMarketUpdateAcceptingOrdersRequest is a request message for the MarketUpdateAcceptingOrders endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "update" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to enable or disable. |
| `accepting_orders` | [bool](#bool) |  | accepting_orders is whether this market is allowing orders to be created for it. |






<a name="provenance-exchange-v1-MsgMarketUpdateAcceptingOrdersResponse"></a>

### MsgMarketUpdateAcceptingOrdersResponse
MsgMarketUpdateAcceptingOrdersResponse is a response message for the MarketUpdateAcceptingOrders endpoint.






<a name="provenance-exchange-v1-MsgMarketUpdateDetailsRequest"></a>

### MsgMarketUpdateDetailsRequest
MsgMarketUpdateDetailsRequest is a request message for the MarketUpdateDetails endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "update" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to update details for. |
| `market_details` | [MarketDetails](#provenance-exchange-v1-MarketDetails) |  | market_details is some information about this market. |






<a name="provenance-exchange-v1-MsgMarketUpdateDetailsResponse"></a>

### MsgMarketUpdateDetailsResponse
MsgMarketUpdateDetailsResponse is a response message for the MarketUpdateDetails endpoint.






<a name="provenance-exchange-v1-MsgMarketUpdateEnabledRequest"></a>

### MsgMarketUpdateEnabledRequest
MsgMarketUpdateEnabledRequest is a request message for the MarketUpdateEnabled endpoint.
Deprecated: This endpoint is no longer usable. It is replaced by MarketUpdateAcceptingOrders.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | **Deprecated.** admin is the account with "update" permission requesting this change. Deprecated: This endpoint is no longer usable. It is replaced by MarketUpdateAcceptingOrders. |
| `market_id` | [uint32](#uint32) |  | **Deprecated.** market_id is the numerical identifier of the market to enable or disable. Deprecated: This endpoint is no longer usable. It is replaced by MarketUpdateAcceptingOrders. |
| `accepting_orders` | [bool](#bool) |  | **Deprecated.** accepting_orders is whether this market is allowing orders to be created for it. Deprecated: This endpoint is no longer usable. It is replaced by MarketUpdateAcceptingOrders. |






<a name="provenance-exchange-v1-MsgMarketUpdateEnabledResponse"></a>

### MsgMarketUpdateEnabledResponse
MsgMarketUpdateEnabledResponse is a response message for the MarketUpdateEnabled endpoint.
Deprecated: This endpoint is no longer usable. It is replaced by MarketUpdateAcceptingOrders.






<a name="provenance-exchange-v1-MsgMarketUpdateIntermediaryDenomRequest"></a>

### MsgMarketUpdateIntermediaryDenomRequest
MsgMarketUpdateIntermediaryDenomRequest is a request message for the MarketUpdateIntermediaryDenom endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "update" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market changing the intermediary denom. |
| `intermediary_denom` | [string](#string) |  | intermediary_denom is the new intermediary denom for this market to use. |






<a name="provenance-exchange-v1-MsgMarketUpdateIntermediaryDenomResponse"></a>

### MsgMarketUpdateIntermediaryDenomResponse
MsgMarketUpdateIntermediaryDenomResponse is a response message for the MarketUpdateIntermediaryDenom endpoint.






<a name="provenance-exchange-v1-MsgMarketUpdateUserSettleRequest"></a>

### MsgMarketUpdateUserSettleRequest
MsgMarketUpdateUserSettleRequest is a request message for the MarketUpdateUserSettle endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "update" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to enable or disable user-settlement for. |
| `allow_user_settlement` | [bool](#bool) |  | allow_user_settlement is whether this market allows users to initiate their own settlements. For example, the FillBids and FillAsks endpoints are available if and only if this is true. The MarketSettle endpoint is available (only to market actors) regardless of the value of this field. |






<a name="provenance-exchange-v1-MsgMarketUpdateUserSettleResponse"></a>

### MsgMarketUpdateUserSettleResponse
MsgMarketUpdateUserSettleResponse is a response message for the MarketUpdateUserSettle endpoint.






<a name="provenance-exchange-v1-MsgMarketWithdrawRequest"></a>

### MsgMarketWithdrawRequest
MsgMarketWithdrawRequest is a request message for the MarketWithdraw endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with withdraw permission requesting the withdrawal. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to withdraw from. |
| `to_address` | [string](#string) |  | to_address is the address that will receive the funds. |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | amount is the funds to withdraw. |






<a name="provenance-exchange-v1-MsgMarketWithdrawResponse"></a>

### MsgMarketWithdrawResponse
MsgMarketWithdrawResponse is a response message for the MarketWithdraw endpoint.






<a name="provenance-exchange-v1-MsgRejectPaymentRequest"></a>

### MsgRejectPaymentRequest
MsgRejectPaymentRequest is a request message for the RejectPayment endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `target` | [string](#string) |  | target is the target account of the payment to reject. |
| `source` | [string](#string) |  | source is the source account of the payment to reject. |
| `external_id` | [string](#string) |  | external_id is the external id of the payment to reject. |






<a name="provenance-exchange-v1-MsgRejectPaymentResponse"></a>

### MsgRejectPaymentResponse
MsgRejectPaymentResponse is a response message for the RejectPayment endpoint.






<a name="provenance-exchange-v1-MsgRejectPaymentsRequest"></a>

### MsgRejectPaymentsRequest
MsgRejectPaymentsRequest is a request message for the RejectPayments endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `target` | [string](#string) |  | target is the account that wishes to reject some payments. |
| `sources` | [string](#string) | repeated | sources is the source accounts of the payments to reject. |






<a name="provenance-exchange-v1-MsgRejectPaymentsResponse"></a>

### MsgRejectPaymentsResponse
MsgRejectPaymentsResponse is a response message for the RejectPayments endpoint.






<a name="provenance-exchange-v1-MsgUpdateParamsRequest"></a>

### MsgUpdateParamsRequest
MsgGovUpdateParamsRequest is a request message for the GovUpdateParams endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `params` | [Params](#provenance-exchange-v1-Params) |  | params are the new param values to set |






<a name="provenance-exchange-v1-MsgUpdateParamsResponse"></a>

### MsgUpdateParamsResponse
MsgUpdateParamsResponse is a response message for the GovUpdateParams endpoint.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-exchange-v1-Msg"></a>

### Msg
Msg is the service for exchange module's tx endpoints.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `CreateAsk` | [MsgCreateAskRequest](#provenance-exchange-v1-MsgCreateAskRequest) | [MsgCreateAskResponse](#provenance-exchange-v1-MsgCreateAskResponse) | CreateAsk creates an ask order (to sell something you own). |
| `CreateBid` | [MsgCreateBidRequest](#provenance-exchange-v1-MsgCreateBidRequest) | [MsgCreateBidResponse](#provenance-exchange-v1-MsgCreateBidResponse) | CreateBid creates a bid order (to buy something you want). |
| `CommitFunds` | [MsgCommitFundsRequest](#provenance-exchange-v1-MsgCommitFundsRequest) | [MsgCommitFundsResponse](#provenance-exchange-v1-MsgCommitFundsResponse) | CommitFunds marks funds in an account as manageable by a market. |
| `CancelOrder` | [MsgCancelOrderRequest](#provenance-exchange-v1-MsgCancelOrderRequest) | [MsgCancelOrderResponse](#provenance-exchange-v1-MsgCancelOrderResponse) | CancelOrder cancels an order. |
| `FillBids` | [MsgFillBidsRequest](#provenance-exchange-v1-MsgFillBidsRequest) | [MsgFillBidsResponse](#provenance-exchange-v1-MsgFillBidsResponse) | FillBids uses the assets in your account to fulfill one or more bids (similar to a fill-or-cancel ask). |
| `FillAsks` | [MsgFillAsksRequest](#provenance-exchange-v1-MsgFillAsksRequest) | [MsgFillAsksResponse](#provenance-exchange-v1-MsgFillAsksResponse) | FillAsks uses the funds in your account to fulfill one or more asks (similar to a fill-or-cancel bid). |
| `MarketSettle` | [MsgMarketSettleRequest](#provenance-exchange-v1-MsgMarketSettleRequest) | [MsgMarketSettleResponse](#provenance-exchange-v1-MsgMarketSettleResponse) | MarketSettle is a market endpoint to trigger the settlement of orders. |
| `MarketCommitmentSettle` | [MsgMarketCommitmentSettleRequest](#provenance-exchange-v1-MsgMarketCommitmentSettleRequest) | [MsgMarketCommitmentSettleResponse](#provenance-exchange-v1-MsgMarketCommitmentSettleResponse) | MarketCommitmentSettle is a market endpoint to transfer committed funds. |
| `MarketReleaseCommitments` | [MsgMarketReleaseCommitmentsRequest](#provenance-exchange-v1-MsgMarketReleaseCommitmentsRequest) | [MsgMarketReleaseCommitmentsResponse](#provenance-exchange-v1-MsgMarketReleaseCommitmentsResponse) | MarketReleaseCommitments is a market endpoint return control of funds back to the account owner(s). |
| `MarketSetOrderExternalID` | [MsgMarketSetOrderExternalIDRequest](#provenance-exchange-v1-MsgMarketSetOrderExternalIDRequest) | [MsgMarketSetOrderExternalIDResponse](#provenance-exchange-v1-MsgMarketSetOrderExternalIDResponse) | MarketSetOrderExternalID updates an order's external id field. |
| `MarketWithdraw` | [MsgMarketWithdrawRequest](#provenance-exchange-v1-MsgMarketWithdrawRequest) | [MsgMarketWithdrawResponse](#provenance-exchange-v1-MsgMarketWithdrawResponse) | MarketWithdraw is a market endpoint to withdraw fees that have been collected. |
| `MarketUpdateDetails` | [MsgMarketUpdateDetailsRequest](#provenance-exchange-v1-MsgMarketUpdateDetailsRequest) | [MsgMarketUpdateDetailsResponse](#provenance-exchange-v1-MsgMarketUpdateDetailsResponse) | MarketUpdateDetails is a market endpoint to update its details. |
| `MarketUpdateEnabled` | [MsgMarketUpdateEnabledRequest](#provenance-exchange-v1-MsgMarketUpdateEnabledRequest) | [MsgMarketUpdateEnabledResponse](#provenance-exchange-v1-MsgMarketUpdateEnabledResponse) | MarketUpdateEnabled is a market endpoint to update whether its accepting orders. Deprecated: This endpoint is no longer usable. It is replaced by MarketUpdateAcceptingOrders. |
| `MarketUpdateAcceptingOrders` | [MsgMarketUpdateAcceptingOrdersRequest](#provenance-exchange-v1-MsgMarketUpdateAcceptingOrdersRequest) | [MsgMarketUpdateAcceptingOrdersResponse](#provenance-exchange-v1-MsgMarketUpdateAcceptingOrdersResponse) | MarketUpdateAcceptingOrders is a market endpoint to update whether its accepting orders. |
| `MarketUpdateUserSettle` | [MsgMarketUpdateUserSettleRequest](#provenance-exchange-v1-MsgMarketUpdateUserSettleRequest) | [MsgMarketUpdateUserSettleResponse](#provenance-exchange-v1-MsgMarketUpdateUserSettleResponse) | MarketUpdateUserSettle is a market endpoint to update whether it allows user-initiated settlement. |
| `MarketUpdateAcceptingCommitments` | [MsgMarketUpdateAcceptingCommitmentsRequest](#provenance-exchange-v1-MsgMarketUpdateAcceptingCommitmentsRequest) | [MsgMarketUpdateAcceptingCommitmentsResponse](#provenance-exchange-v1-MsgMarketUpdateAcceptingCommitmentsResponse) | MarketUpdateAcceptingCommitments is a market endpoint to update whether it accepts commitments. |
| `MarketUpdateIntermediaryDenom` | [MsgMarketUpdateIntermediaryDenomRequest](#provenance-exchange-v1-MsgMarketUpdateIntermediaryDenomRequest) | [MsgMarketUpdateIntermediaryDenomResponse](#provenance-exchange-v1-MsgMarketUpdateIntermediaryDenomResponse) | MarketUpdateIntermediaryDenom sets a market's intermediary denom. |
| `MarketManagePermissions` | [MsgMarketManagePermissionsRequest](#provenance-exchange-v1-MsgMarketManagePermissionsRequest) | [MsgMarketManagePermissionsResponse](#provenance-exchange-v1-MsgMarketManagePermissionsResponse) | MarketManagePermissions is a market endpoint to manage a market's user permissions. |
| `MarketManageReqAttrs` | [MsgMarketManageReqAttrsRequest](#provenance-exchange-v1-MsgMarketManageReqAttrsRequest) | [MsgMarketManageReqAttrsResponse](#provenance-exchange-v1-MsgMarketManageReqAttrsResponse) | MarketManageReqAttrs is a market endpoint to manage the attributes required to interact with it. |
| `CreatePayment` | [MsgCreatePaymentRequest](#provenance-exchange-v1-MsgCreatePaymentRequest) | [MsgCreatePaymentResponse](#provenance-exchange-v1-MsgCreatePaymentResponse) | CreatePayment creates a payment to facilitate a trade between two accounts. |
| `AcceptPayment` | [MsgAcceptPaymentRequest](#provenance-exchange-v1-MsgAcceptPaymentRequest) | [MsgAcceptPaymentResponse](#provenance-exchange-v1-MsgAcceptPaymentResponse) | AcceptPayment is used by a target to accept a payment. |
| `RejectPayment` | [MsgRejectPaymentRequest](#provenance-exchange-v1-MsgRejectPaymentRequest) | [MsgRejectPaymentResponse](#provenance-exchange-v1-MsgRejectPaymentResponse) | RejectPayment can be used by a target to reject a payment. |
| `RejectPayments` | [MsgRejectPaymentsRequest](#provenance-exchange-v1-MsgRejectPaymentsRequest) | [MsgRejectPaymentsResponse](#provenance-exchange-v1-MsgRejectPaymentsResponse) | RejectPayments can be used by a target to reject all payments from one or more sources. |
| `CancelPayments` | [MsgCancelPaymentsRequest](#provenance-exchange-v1-MsgCancelPaymentsRequest) | [MsgCancelPaymentsResponse](#provenance-exchange-v1-MsgCancelPaymentsResponse) | CancelPayments can be used by a source to cancel one or more payments. |
| `ChangePaymentTarget` | [MsgChangePaymentTargetRequest](#provenance-exchange-v1-MsgChangePaymentTargetRequest) | [MsgChangePaymentTargetResponse](#provenance-exchange-v1-MsgChangePaymentTargetResponse) | ChangePaymentTarget can be used by a source to change the target in one of their payments. |
| `GovCreateMarket` | [MsgGovCreateMarketRequest](#provenance-exchange-v1-MsgGovCreateMarketRequest) | [MsgGovCreateMarketResponse](#provenance-exchange-v1-MsgGovCreateMarketResponse) | GovCreateMarket is a governance proposal endpoint for creating a market. |
| `GovManageFees` | [MsgGovManageFeesRequest](#provenance-exchange-v1-MsgGovManageFeesRequest) | [MsgGovManageFeesResponse](#provenance-exchange-v1-MsgGovManageFeesResponse) | GovManageFees is a governance proposal endpoint for updating a market's fees. |
| `GovCloseMarket` | [MsgGovCloseMarketRequest](#provenance-exchange-v1-MsgGovCloseMarketRequest) | [MsgGovCloseMarketResponse](#provenance-exchange-v1-MsgGovCloseMarketResponse) | GovCloseMarket is a governance proposal endpoint that will disable order and commitment creation, cancel all orders, and release all commitments. |
| `GovUpdateParams` | [MsgGovUpdateParamsRequest](#provenance-exchange-v1-MsgGovUpdateParamsRequest) | [MsgGovUpdateParamsResponse](#provenance-exchange-v1-MsgGovUpdateParamsResponse) | GovUpdateParams is a governance proposal endpoint for updating the exchange module's params. Deprecated: Use UpdateParams instead. |
| `UpdateParams` | [MsgUpdateParamsRequest](#provenance-exchange-v1-MsgUpdateParamsRequest) | [MsgUpdateParamsResponse](#provenance-exchange-v1-MsgUpdateParamsResponse) | UpdateParams is a governance proposal endpoint for updating the exchange module's params. |

 <!-- end services -->



<a name="provenance_exchange_v1_events-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/events.proto



<a name="provenance-exchange-v1-EventCommitmentReleased"></a>

### EventCommitmentReleased
EventCommitmentReleased is an event emitted when funds are released from their commitment.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account is the bech32 address string of the account. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `amount` | [string](#string) |  | amount is the coins string of the funds that were released from commitment. |
| `tag` | [string](#string) |  | tag is the string provided in the message causing this event. |






<a name="provenance-exchange-v1-EventFundsCommitted"></a>

### EventFundsCommitted
EventFundsCommitted is an event emitted when funds are committed to a market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account is the bech32 address string of the account. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `amount` | [string](#string) |  | amount is the coins string of the newly committed funds. |
| `tag` | [string](#string) |  | tag is the string provided in the message causing this event. |






<a name="provenance-exchange-v1-EventMarketCommitmentsDisabled"></a>

### EventMarketCommitmentsDisabled
EventMarketCommitmentsDisabled is an event emitted when a market's accepting_commitments option is disabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the accepting_commitments option. |






<a name="provenance-exchange-v1-EventMarketCommitmentsEnabled"></a>

### EventMarketCommitmentsEnabled
EventMarketCommitmentsEnabled is an event emitted when a market's accepting_commitments option is enabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the accepting_commitments option. |






<a name="provenance-exchange-v1-EventMarketCreated"></a>

### EventMarketCreated
EventMarketCreated is an event emitted when a market has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |






<a name="provenance-exchange-v1-EventMarketDetailsUpdated"></a>

### EventMarketDetailsUpdated
EventMarketDetailsUpdated is an event emitted when a market's details are updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the details. |






<a name="provenance-exchange-v1-EventMarketDisabled"></a>

### EventMarketDisabled
EventMarketDisabled is an event emitted when a market is disabled.
Deprecated: This event is no longer used. It is replaced with EventMarketOrdersDisabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that disabled the market. |






<a name="provenance-exchange-v1-EventMarketEnabled"></a>

### EventMarketEnabled
EventMarketEnabled is an event emitted when a market is enabled.
Deprecated: This event is no longer used. It is replaced with EventMarketOrdersEnabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that enabled the market. |






<a name="provenance-exchange-v1-EventMarketFeesUpdated"></a>

### EventMarketFeesUpdated
EventMarketFeesUpdated is an event emitted when a market's fees have been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |






<a name="provenance-exchange-v1-EventMarketIntermediaryDenomUpdated"></a>

### EventMarketIntermediaryDenomUpdated
EventMarketIntermediaryDenomUpdated is an event emitted when a market updates its
commitment_settlement_intermediary_denom field.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the intermediary denom. |






<a name="provenance-exchange-v1-EventMarketOrdersDisabled"></a>

### EventMarketOrdersDisabled
EventMarketOrdersEnabled is an event emitted when a market disables order creation.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the accepting_orders option. |






<a name="provenance-exchange-v1-EventMarketOrdersEnabled"></a>

### EventMarketOrdersEnabled
EventMarketOrdersEnabled is an event emitted when a market enables order creation.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the accepting_orders option. |






<a name="provenance-exchange-v1-EventMarketPermissionsUpdated"></a>

### EventMarketPermissionsUpdated
EventMarketPermissionsUpdated is an event emitted when a market's permissions are updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the permissions. |






<a name="provenance-exchange-v1-EventMarketReqAttrUpdated"></a>

### EventMarketReqAttrUpdated
EventMarketReqAttrUpdated is an event emitted when a market's required attributes are updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the required attributes. |






<a name="provenance-exchange-v1-EventMarketUserSettleDisabled"></a>

### EventMarketUserSettleDisabled
EventMarketUserSettleDisabled is an event emitted when a market's user_settle option is disabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the user_settle option. |






<a name="provenance-exchange-v1-EventMarketUserSettleEnabled"></a>

### EventMarketUserSettleEnabled
EventMarketUserSettleEnabled is an event emitted when a market's user_settle option is enabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the user_settle option. |






<a name="provenance-exchange-v1-EventMarketWithdraw"></a>

### EventMarketWithdraw
EventMarketWithdraw is an event emitted when a withdrawal of a market's collected fees is made.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `amount` | [string](#string) |  | amount is the coins amount string of funds withdrawn from the market account. |
| `destination` | [string](#string) |  | destination is the account that received the funds. |
| `withdrawn_by` | [string](#string) |  | withdrawn_by is the account that requested the withdrawal. |






<a name="provenance-exchange-v1-EventOrderCancelled"></a>

### EventOrderCancelled
EventOrderCancelled is an event emitted when an order is cancelled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order cancelled. |
| `cancelled_by` | [string](#string) |  | cancelled_by is the account that triggered the cancellation of the order. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `external_id` | [string](#string) |  | external_id is the order's external id. |






<a name="provenance-exchange-v1-EventOrderCreated"></a>

### EventOrderCreated
EventOrderCreated is an event emitted when an order is created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order created. |
| `order_type` | [string](#string) |  | order_type is the type of order, e.g. "ask" or "bid". |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `external_id` | [string](#string) |  | external_id is the order's external id. |






<a name="provenance-exchange-v1-EventOrderExternalIDUpdated"></a>

### EventOrderExternalIDUpdated
EventOrderExternalIDUpdated is an event emitted when an order's external id is updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order partially filled. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `external_id` | [string](#string) |  | external_id is the order's new external id. |






<a name="provenance-exchange-v1-EventOrderFilled"></a>

### EventOrderFilled
EventOrderFilled is an event emitted when an order has been filled in full.
This event is also used for orders that were previously partially filled, but have now been filled in full.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order filled. |
| `assets` | [string](#string) |  | assets is the coins amount string of assets bought/sold for this order. |
| `price` | [string](#string) |  | price is the coins amount string of the price payed/received for this order. |
| `fees` | [string](#string) |  | fees is the coins amount string of settlement fees paid with this order. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `external_id` | [string](#string) |  | external_id is the order's external id. |






<a name="provenance-exchange-v1-EventOrderPartiallyFilled"></a>

### EventOrderPartiallyFilled
EventOrderPartiallyFilled is an event emitted when an order filled in part and still has more left to fill.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order partially filled. |
| `assets` | [string](#string) |  | assets is the coins amount string of assets that were filled and removed from the order. |
| `price` | [string](#string) |  | price is the coins amount string of the price payed/received for this order. For ask orders, this might be more than the amount that was removed from the order's price. |
| `fees` | [string](#string) |  | fees is the coins amount string of settlement fees paid with this partial order. For ask orders, this might be more than the amount that was removed from the order's settlement fees. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `external_id` | [string](#string) |  | external_id is the order's external id. |






<a name="provenance-exchange-v1-EventParamsUpdated"></a>

### EventParamsUpdated
EventParamsUpdated is an event emitted when the exchange module's params have been updated.






<a name="provenance-exchange-v1-EventPaymentAccepted"></a>

### EventPaymentAccepted
EventPaymentAccepted is an event emitted when a payment is accepted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the account that created the Payment. |
| `source_amount` | [string](#string) |  | source_amount is the coins amount string of the funds that the source will pay (to the target). |
| `target` | [string](#string) |  | target is the account that accepted the Payment. |
| `target_amount` | [string](#string) |  | target_amount is the coins amount string of the funds that the target will pay (to the source). |
| `external_id` | [string](#string) |  | external_id is used along with the source to uniquely identify this Payment. |






<a name="provenance-exchange-v1-EventPaymentCancelled"></a>

### EventPaymentCancelled
EventPaymentCancelled is an event emitted when a payment is cancelled (by the source).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the account that cancelled (and created) the Payment. |
| `target` | [string](#string) |  | target is the account that could have accepted the Payment. |
| `external_id` | [string](#string) |  | external_id is used along with the source to uniquely identify this Payment. |






<a name="provenance-exchange-v1-EventPaymentCreated"></a>

### EventPaymentCreated
EventPaymentCreated is an event emitted when a payment is created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the account that created the Payment. |
| `source_amount` | [string](#string) |  | source_amount is the coins amount string of the funds that the source will pay (to the target). |
| `target` | [string](#string) |  | target is the account that can accept the Payment. |
| `target_amount` | [string](#string) |  | target_amount is the coins amount string of the funds that the target will pay (to the source). |
| `external_id` | [string](#string) |  | external_id is used along with the source to uniquely identify this Payment. |






<a name="provenance-exchange-v1-EventPaymentRejected"></a>

### EventPaymentRejected
EventPaymentRejected is an event emitted when a payment is rejected (by the target).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the account that created the Payment. |
| `target` | [string](#string) |  | target is the account that rejected the Payment. |
| `external_id` | [string](#string) |  | external_id is used along with the source to uniquely identify this Payment. |






<a name="provenance-exchange-v1-EventPaymentUpdated"></a>

### EventPaymentUpdated
EventPaymentUpdated is an event emitted when a payment is updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the account that updated (and previously created) the Payment. |
| `source_amount` | [string](#string) |  | source_amount is the coins amount string of the funds that the source will pay (to the target). |
| `old_target` | [string](#string) |  | old_target is the account that used to be able to accept the Payment (but not any more). |
| `new_target` | [string](#string) |  | new_target is the account that is now able to accept the Payment. |
| `target_amount` | [string](#string) |  | target_amount is the coins amount string of the funds that the target will pay (to the source). |
| `external_id` | [string](#string) |  | external_id is used along with the source to uniquely identify this Payment. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_exchange_v1_market-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/market.proto



<a name="provenance-exchange-v1-AccessGrant"></a>

### AccessGrant
AddrPermissions associates an address with a list of permissions available for that address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address that these permissions apply to. |
| `permissions` | [Permission](#provenance-exchange-v1-Permission) | repeated | allowed is the list of permissions available for the address. |






<a name="provenance-exchange-v1-FeeRatio"></a>

### FeeRatio
FeeRatio defines a ratio of price amount to fee amount.
For an order to be valid, its price must be evenly divisible by a FeeRatio's price.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `price` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | price is the unit the order price is divided by to get how much of the fee should apply. |
| `fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | fee is the amount to charge per price unit. |






<a name="provenance-exchange-v1-Market"></a>

### Market
Market contains all information about a market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier for this market. |
| `market_details` | [MarketDetails](#provenance-exchange-v1-MarketDetails) |  | market_details is some information about this market. |
| `fee_create_ask_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | fee_create_ask_flat is the flat fee charged for creating an ask order. Each coin entry is a separate option. When an ask is created, one of these must be paid. If empty, no fee is required to create an ask order. |
| `fee_create_bid_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | fee_create_bid_flat is the flat fee charged for creating a bid order. Each coin entry is a separate option. When a bid is created, one of these must be paid. If empty, no fee is required to create a bid order. |
| `fee_seller_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | fee_seller_settlement_flat is the flat fee charged to the seller during settlement. Each coin entry is a separate option. When an ask is settled, the seller will pay the amount in the denom that matches the price they received. |
| `fee_seller_settlement_ratios` | [FeeRatio](#provenance-exchange-v1-FeeRatio) | repeated | fee_seller_settlement_ratios is the fee to charge a seller during settlement based on the price they are receiving. The price and fee denoms must be equal for each entry, and only one entry for any given denom is allowed. |
| `fee_buyer_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | fee_buyer_settlement_flat is the flat fee charged to the buyer during settlement. Each coin entry is a separate option. When a bid is created, the settlement fees provided must contain one of these. |
| `fee_buyer_settlement_ratios` | [FeeRatio](#provenance-exchange-v1-FeeRatio) | repeated | fee_buyer_settlement_ratios is the fee to charge a buyer during settlement based on the price they are spending. The price and fee denoms do not have to equal. Multiple entries for any given price or fee denom are allowed, but each price denom to fee denom pair can only have one entry. |
| `accepting_orders` | [bool](#bool) |  | accepting_orders is whether this market is allowing orders to be created for it. |
| `allow_user_settlement` | [bool](#bool) |  | allow_user_settlement is whether this market allows users to initiate their own settlements. For example, the FillBids and FillAsks endpoints are available if and only if this is true. The MarketSettle endpoint is only available to market actors regardless of the value of this field. |
| `access_grants` | [AccessGrant](#provenance-exchange-v1-AccessGrant) | repeated | access_grants is the list of addresses and permissions granted for this market. |
| `req_attr_create_ask` | [string](#string) | repeated | req_attr_create_ask is a list of attributes required on an account for it to be allowed to create an ask order. An account must have all of these attributes in order to create an ask order in this market. If the list is empty, any account can create ask orders in this market.<br>An entry that starts with "*." will match any attributes that end with the rest of it. E.g. "*.b.a" will match all of "c.b.a", "x.b.a", and "e.d.c.b.a"; but not "b.a", "xb.a", "b.x.a", or "c.b.a.x". |
| `req_attr_create_bid` | [string](#string) | repeated | req_attr_create_ask is a list of attributes required on an account for it to be allowed to create a bid order. An account must have all of these attributes in order to create a bid order in this market. If the list is empty, any account can create bid orders in this market.<br>An entry that starts with "*." will match any attributes that end with the rest of it. E.g. "*.b.a" will match all of "c.b.a", "x.b.a", and "e.d.c.b.a"; but not "b.a", "xb.a", "c.b.x.a", or "c.b.a.x". |
| `accepting_commitments` | [bool](#bool) |  | accepting_commitments is whether the market is allowing users to commit funds to it. |
| `fee_create_commitment_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | fee_create_commitment_flat is the flat fee charged for creating a commitment. Each coin entry is a separate option. When a commitment is created, one of these must be paid. If empty, no fee is required to create a commitment. |
| `commitment_settlement_bips` | [uint32](#uint32) |  | commitment_settlement_bips is the fraction of a commitment settlement that will be paid to the exchange. It is represented in basis points (1/100th of 1%, e.g. 0.0001) and is limited to 0 to 10,000 inclusive. During a commitment settlement, the inputs are summed and NAVs are used to convert that total to the intermediary denom, then to the fee denom. That is then multiplied by this value to get the fee amount that will be transferred out of the market's account into the exchange for that settlement.<br>Summing the inputs effectively doubles the value of the settlement from what what is usually thought of as the value of a trade. That should be taken into account when setting this value. E.g. if two accounts are trading 10apples for 100grapes, the inputs total will be 10apples,100grapes (which might then be converted to USD then nhash before applying this ratio); Usually, though, the value of that trade would be viewed as either just 10apples or just 100grapes. |
| `intermediary_denom` | [string](#string) |  | intermediary_denom is the denom that funds get converted to (before being converted to the chain's fee denom) when calculating the fees that are paid to the exchange. NAVs are used for this conversion and actions will fail if a NAV is needed but not available. |
| `req_attr_create_commitment` | [string](#string) | repeated | req_attr_create_commitment is a list of attributes required on an account for it to be allowed to create a commitment. An account must have all of these attributes in order to create a commitment in this market. If the list is empty, any account can create commitments in this market.<br>An entry that starts with "*." will match any attributes that end with the rest of it. E.g. "*.b.a" will match all of "c.b.a", "x.b.a", and "e.d.c.b.a"; but not "b.a", "xb.a", "c.b.x.a", or "c.b.a.x". |






<a name="provenance-exchange-v1-MarketAccount"></a>

### MarketAccount
MarketAccount is an account type for use with the accounts module to hold some basic information about a market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_account` | [cosmos.auth.v1beta1.BaseAccount](#cosmos-auth-v1beta1-BaseAccount) |  | base_account is the base cosmos account information. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier for this market. |
| `market_details` | [MarketDetails](#provenance-exchange-v1-MarketDetails) |  | market_details is some human-consumable information about this market. |






<a name="provenance-exchange-v1-MarketBrief"></a>

### MarketBrief
MarketBrief is a message containing brief, superficial information about a market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier for this market. |
| `market_address` | [string](#string) |  | market_address is the bech32 address string of this market's account. |
| `market_details` | [MarketDetails](#provenance-exchange-v1-MarketDetails) |  | market_details is some information about this market. |






<a name="provenance-exchange-v1-MarketDetails"></a>

### MarketDetails
MarketDetails contains information about a market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | name is a moniker that people can use to refer to this market. |
| `description` | [string](#string) |  | description extra information about this market. The field is meant to be human-readable. |
| `website_url` | [string](#string) |  | website_url is a url people can use to get to this market, or at least get more information about this market. |
| `icon_uri` | [string](#string) |  | icon_uri is a uri for an icon to associate with this market. |





 <!-- end messages -->


<a name="provenance-exchange-v1-Permission"></a>

### Permission
Permission defines the different types of permission that can be given to an account for a market.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `PERMISSION_UNSPECIFIED` | `0` | PERMISSION_UNSPECIFIED is the zero-value Permission; it is an error to use it. |
| `PERMISSION_SETTLE` | `1` | PERMISSION_SETTLE is the ability to use the Settle Tx endpoint on behalf of a market. |
| `PERMISSION_SET_IDS` | `2` | PERMISSION_SET_IDS is the ability to use the SetOrderExternalID Tx endpoint on behalf of a market. |
| `PERMISSION_CANCEL` | `3` | PERMISSION_CANCEL is the ability to use the Cancel Tx endpoint on behalf of a market. |
| `PERMISSION_WITHDRAW` | `4` | PERMISSION_WITHDRAW is the ability to use the MarketWithdraw Tx endpoint. |
| `PERMISSION_UPDATE` | `5` | PERMISSION_UPDATE is the ability to use the MarketUpdate* Tx endpoints. |
| `PERMISSION_PERMISSIONS` | `6` | PERMISSION_PERMISSIONS is the ability to use the MarketManagePermissions Tx endpoint. |
| `PERMISSION_ATTRIBUTES` | `7` | PERMISSION_ATTRIBUTES is the ability to use the MarketManageReqAttrs Tx endpoint. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_exchange_v1_payments-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/payments.proto



<a name="provenance-exchange-v1-Payment"></a>

### Payment
Payment represents one account's desire to trade funds with another account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the account that created this Payment. It is considered the owner of the payment. |
| `source_amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | source_amount is the funds that the source is will pay the target in exchange for the target_amount. A hold will be placed on this amount in the source account until this Payment is accepted, rejected or cancelled. If the source_amount is zero, this Payment can be considered a "payment request." |
| `target` | [string](#string) |  | target is the account that can accept this Payment. The target is the only thing allowed to change in a payment. I.e. it can be empty initially and updated later as needed. |
| `target_amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | target_amount is the funds that the target will pay the source in exchange for the source_amount. If the target_amount is zero, this Payment can be considered a "peer-to-peer (P2P) payment." |
| `external_id` | [string](#string) |  | external_id is used along with the source to uniquely identify this Payment.<br>A source can only have one Payment with any given external id. A source can have two payments with two different external ids. Two different sources can each have a payment with the same external id. But a source cannot have two different payments each with the same external id.<br>An external id can be reused by a source once the payment is accepted, rejected, or cancelled.<br>The external id is limited to 100 bytes. An empty string is a valid external id. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_exchange_v1_commitments-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/commitments.proto



<a name="provenance-exchange-v1-AccountAmount"></a>

### AccountAmount
AccountAmount associates an account with a coins amount.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account is the bech32 address string of the account associated with the amount. |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | amount is the funds associated with the address. |






<a name="provenance-exchange-v1-Commitment"></a>

### Commitment
Commitment contains information on committed funds.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account is the bech32 address string with the committed funds. |
| `market_id` | [uint32](#uint32) |  | market_id is the numeric identifier of the market the funds are committed to. |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | amount is the funds that have been committed by the account to the market. |






<a name="provenance-exchange-v1-MarketAmount"></a>

### MarketAmount
MarketAmount associates a market with a coins amount.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numeric identifier the amount has been committed to. |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | amount is the funds associated with the address. |






<a name="provenance-exchange-v1-NetAssetPrice"></a>

### NetAssetPrice
NetAssetPrice is an association of assets and price used to record the value of things.
It is related to the NetAssetValue message from the x/marker module, and is therefore often referred to as "a NAV".


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assets` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | assets is the volume and denom that has been bought or sold. |
| `price` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | price is what was paid for the assets. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_exchange_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/query.proto



<a name="provenance-exchange-v1-QueryCommitmentSettlementFeeCalcRequest"></a>

### QueryCommitmentSettlementFeeCalcRequest
QueryCommitmentSettlementFeeCalcRequest is a request message for the CommitmentSettlementFeeCalc query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `settlement` | [MsgMarketCommitmentSettleRequest](#provenance-exchange-v1-MsgMarketCommitmentSettleRequest) |  | settlement is a market's commitment settlement request message. If no inputs are provided, only the to_fee_nav field will be populated in the response. |
| `include_breakdown_fields` | [bool](#bool) |  | include_breakdown_fields controls the fields that are populated in the response. If false, only the exchange_fees field is populated. If true, all of the fields are populated as possible. If the settlement does not have any inputs, this field defaults to true. |






<a name="provenance-exchange-v1-QueryCommitmentSettlementFeeCalcResponse"></a>

### QueryCommitmentSettlementFeeCalcResponse
QueryCommitmentSettlementFeeCalcResponse is a response message for the CommitmentSettlementFeeCalc query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exchange_fees` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | exchange_fees is the total that the exchange would currently pay for the provided settlement. |
| `input_total` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | input_total is the sum of all the inputs in the provided settlement. |
| `converted_total` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | converted_total is the input_total converted to a single intermediary denom or left as the fee denom. |
| `conversion_navs` | [NetAssetPrice](#provenance-exchange-v1-NetAssetPrice) | repeated | conversion_navs are the NAVs used to convert the input_total to the converted_total. |
| `to_fee_nav` | [NetAssetPrice](#provenance-exchange-v1-NetAssetPrice) |  | to_fee_nav is the NAV used to convert the converted_total into the fee denom. |






<a name="provenance-exchange-v1-QueryGetAccountCommitmentsRequest"></a>

### QueryGetAccountCommitmentsRequest
QueryGetAccountCommitmentsRequest is a request message for the GetAccountCommitments query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account is the bech32 address string of the account with the commitments. |






<a name="provenance-exchange-v1-QueryGetAccountCommitmentsResponse"></a>

### QueryGetAccountCommitmentsResponse
QueryGetAccountCommitmentsResponse is a response message for the GetAccountCommitments query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `commitments` | [MarketAmount](#provenance-exchange-v1-MarketAmount) | repeated | commitments is the amounts committed from the account to the any market. |






<a name="provenance-exchange-v1-QueryGetAllCommitmentsRequest"></a>

### QueryGetAllCommitmentsRequest
QueryGetAllCommitmentsRequest is a request message for the GetAllCommitments query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetAllCommitmentsResponse"></a>

### QueryGetAllCommitmentsResponse
QueryGetAllCommitmentsResponse is a response message for the GetAllCommitments query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `commitments` | [Commitment](#provenance-exchange-v1-Commitment) | repeated | commitments is the requested commitment information. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryGetAllMarketsRequest"></a>

### QueryGetAllMarketsRequest
QueryGetAllMarketsRequest is a request message for the GetAllMarkets query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetAllMarketsResponse"></a>

### QueryGetAllMarketsResponse
QueryGetAllMarketsResponse is a response message for the GetAllMarkets query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `markets` | [MarketBrief](#provenance-exchange-v1-MarketBrief) | repeated | markets are a page of the briefs for all markets. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryGetAllOrdersRequest"></a>

### QueryGetAllOrdersRequest
QueryGetAllOrdersRequest is a request message for the GetAllOrders query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetAllOrdersResponse"></a>

### QueryGetAllOrdersResponse
QueryGetAllOrdersResponse is a response message for the GetAllOrders query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `orders` | [Order](#provenance-exchange-v1-Order) | repeated | orders are a page of the all orders. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryGetAllPaymentsRequest"></a>

### QueryGetAllPaymentsRequest
QueryGetAllPaymentsRequest is a request message for the GetAllPayments query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetAllPaymentsResponse"></a>

### QueryGetAllPaymentsResponse
QueryGetAllPaymentsResponse is a response message for the GetAllPayments query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payments` | [Payment](#provenance-exchange-v1-Payment) | repeated | payments is all the payments on this page of results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryGetAssetOrdersRequest"></a>

### QueryGetAssetOrdersRequest
QueryGetAssetOrdersRequest is a request message for the GetAssetOrders query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [string](#string) |  | asset is the denom of assets to get orders for. |
| `order_type` | [string](#string) |  | order_type is optional and can limit orders to only "ask" or "bid" orders. |
| `after_order_id` | [uint64](#uint64) |  | after_order_id is a minimum (exclusive) order id. All results will be strictly greater than this. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetAssetOrdersResponse"></a>

### QueryGetAssetOrdersResponse
QueryGetAssetOrdersResponse is a response message for the GetAssetOrders query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `orders` | [Order](#provenance-exchange-v1-Order) | repeated | orders are a page of the orders for the provided asset. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryGetCommitmentRequest"></a>

### QueryGetCommitmentRequest
QueryGetCommitmentRequest is a request message for the GetCommitment query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account is the bech32 address string of the account in the commitment. |
| `market_id` | [uint32](#uint32) |  | market_id is the numeric identifier of the market in the commitment. |






<a name="provenance-exchange-v1-QueryGetCommitmentResponse"></a>

### QueryGetCommitmentResponse
QueryGetCommitmentResponse is a response message for the GetCommitment query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | amount is the total funds committed to the market by the account. |






<a name="provenance-exchange-v1-QueryGetMarketCommitmentsRequest"></a>

### QueryGetMarketCommitmentsRequest
QueryGetMarketCommitmentsRequest is a request message for the GetMarketCommitments query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numeric identifier of the market with the commitment. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetMarketCommitmentsResponse"></a>

### QueryGetMarketCommitmentsResponse
QueryGetMarketCommitmentsResponse is a response message for the GetMarketCommitments query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `commitments` | [AccountAmount](#provenance-exchange-v1-AccountAmount) | repeated | commitments is the amounts committed to the market from any account. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryGetMarketOrdersRequest"></a>

### QueryGetMarketOrdersRequest
QueryGetMarketOrdersRequest is a request message for the GetMarketOrders query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the id of the market to get all the orders for. |
| `order_type` | [string](#string) |  | order_type is optional and can limit orders to only "ask" or "bid" orders. |
| `after_order_id` | [uint64](#uint64) |  | after_order_id is a minimum (exclusive) order id. All results will be strictly greater than this. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetMarketOrdersResponse"></a>

### QueryGetMarketOrdersResponse
QueryGetMarketOrdersResponse is a response message for the GetMarketOrders query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `orders` | [Order](#provenance-exchange-v1-Order) | repeated | orders are a page of the orders in the provided market. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryGetMarketRequest"></a>

### QueryGetMarketRequest
QueryGetMarketRequest is a request message for the GetMarket query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the id of the market to look up. |






<a name="provenance-exchange-v1-QueryGetMarketResponse"></a>

### QueryGetMarketResponse
QueryGetMarketResponse is a response message for the GetMarket query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the bech32 address string of this market's account. |
| `market` | [Market](#provenance-exchange-v1-Market) |  | market is all information and details of the market. |






<a name="provenance-exchange-v1-QueryGetOrderByExternalIDRequest"></a>

### QueryGetOrderByExternalIDRequest
QueryGetOrderByExternalIDRequest is a request message for the GetOrderByExternalID query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the id of the market that's expected to have the order. |
| `external_id` | [string](#string) |  | external_id the external id to look up. |






<a name="provenance-exchange-v1-QueryGetOrderByExternalIDResponse"></a>

### QueryGetOrderByExternalIDResponse
QueryGetOrderByExternalIDResponse is a response message for the GetOrderByExternalID query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order` | [Order](#provenance-exchange-v1-Order) |  | order is the requested order. |






<a name="provenance-exchange-v1-QueryGetOrderRequest"></a>

### QueryGetOrderRequest
QueryGetOrderRequest is a request message for the GetOrder query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the id of the order to look up. |






<a name="provenance-exchange-v1-QueryGetOrderResponse"></a>

### QueryGetOrderResponse
QueryGetOrderResponse is a response message for the GetOrder query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order` | [Order](#provenance-exchange-v1-Order) |  | order is the requested order. |






<a name="provenance-exchange-v1-QueryGetOwnerOrdersRequest"></a>

### QueryGetOwnerOrdersRequest
QueryGetOwnerOrdersRequest is a request message for the GetOwnerOrders query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | owner is the bech32 address string of the owner to get the orders for. |
| `order_type` | [string](#string) |  | order_type is optional and can limit orders to only "ask" or "bid" orders. |
| `after_order_id` | [uint64](#uint64) |  | after_order_id is a minimum (exclusive) order id. All results will be strictly greater than this. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetOwnerOrdersResponse"></a>

### QueryGetOwnerOrdersResponse
QueryGetOwnerOrdersResponse is a response message for the GetOwnerOrders query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `orders` | [Order](#provenance-exchange-v1-Order) | repeated | orders are a page of the orders for the provided address. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryGetPaymentRequest"></a>

### QueryGetPaymentRequest
QueryGetPaymentRequest is a request message for the GetPayment query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the source account of the payment to get. |
| `external_id` | [string](#string) |  | external_id is the external id of the payment to get. |






<a name="provenance-exchange-v1-QueryGetPaymentResponse"></a>

### QueryGetPaymentResponse
QueryGetPaymentResponse is a response message for the GetPayment query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payment` | [Payment](#provenance-exchange-v1-Payment) |  | payment is the info on the requested payment. |






<a name="provenance-exchange-v1-QueryGetPaymentsWithSourceRequest"></a>

### QueryGetPaymentsWithSourceRequest
QueryGetPaymentsWithSourceRequest is a request message for the GetPaymentsWithSource query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `source` | [string](#string) |  | source is the source account of the payments to get. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetPaymentsWithSourceResponse"></a>

### QueryGetPaymentsWithSourceResponse
QueryGetPaymentsWithSourceResponse is a response message for the GetPaymentsWithSource query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payments` | [Payment](#provenance-exchange-v1-Payment) | repeated | payments is all the payments with the requested source. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryGetPaymentsWithTargetRequest"></a>

### QueryGetPaymentsWithTargetRequest
QueryGetPaymentsWithTargetRequest is a request message for the GetPaymentsWithTarget query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `target` | [string](#string) |  | target is the target account of the payments to get. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-exchange-v1-QueryGetPaymentsWithTargetResponse"></a>

### QueryGetPaymentsWithTargetResponse
QueryGetPaymentsWithTargetResponse is a response message for the GetPaymentsWithTarget query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payments` | [Payment](#provenance-exchange-v1-Payment) | repeated | payments is all the payments with the requested target. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination is the resulting pagination parameters. |






<a name="provenance-exchange-v1-QueryOrderFeeCalcRequest"></a>

### QueryOrderFeeCalcRequest
QueryOrderFeeCalcRequest is a request message for the OrderFeeCalc query.
Exactly one of ask_order or bid_order must be provided.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ask_order` | [AskOrder](#provenance-exchange-v1-AskOrder) |  | ask_order is the ask order to calculate the fees for. |
| `bid_order` | [BidOrder](#provenance-exchange-v1-BidOrder) |  | bid_order is the bid order to calculate the fees for. |






<a name="provenance-exchange-v1-QueryOrderFeeCalcResponse"></a>

### QueryOrderFeeCalcResponse
QueryOrderFeeCalcResponse is a response message for the OrderFeeCalc query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `creation_fee_options` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | creation_fee_options are the order creation flat fee options available for creating the provided order. If it's empty, no order creation fee is required. When creating the order, you should include exactly one of these. |
| `settlement_flat_fee_options` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | settlement_flat_fee_options are the settlement flat fee options available for the provided order. If it's empty, no settlement flat fee is required. When creating an order, you should include exactly one of these in the settlement fees field. |
| `settlement_ratio_fee_options` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | settlement_ratio_fee_options are the settlement ratio fee options available for the provided order. If it's empty, no settlement ratio fee is required.<br>If the provided order was a bid order, you should include exactly one of these in the settlement fees field. If the flat and ratio options you've chose have the same denom, a single entry should be included with their sum.<br>If the provided order was an ask order, these are purely informational and represent how much will be removed from your price if it settles at that price. If it settles for more, the actual amount will probably be larger. |






<a name="provenance-exchange-v1-QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is a request message for the Params query.






<a name="provenance-exchange-v1-QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is a response message for the Params query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-exchange-v1-Params) |  | params are the exchange module parameter values. |






<a name="provenance-exchange-v1-QueryPaymentFeeCalcRequest"></a>

### QueryPaymentFeeCalcRequest
QueryPaymentFeeCalcRequest is a request message for the PaymentFeeCalc query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `payment` | [Payment](#provenance-exchange-v1-Payment) |  | payment is the details of the payment to create or accept. |






<a name="provenance-exchange-v1-QueryPaymentFeeCalcResponse"></a>

### QueryPaymentFeeCalcResponse
QueryPaymentFeeCalcResponse is a response message for the PaymentFeeCalc query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `fee_create` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | fee_create is the fee required to create the provided payment. |
| `fee_accept` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | fee_accept is the fee required to accept the provided payment. |






<a name="provenance-exchange-v1-QueryValidateCreateMarketRequest"></a>

### QueryValidateCreateMarketRequest
QueryValidateCreateMarketRequest is a request message for the ValidateCreateMarket query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `create_market_request` | [MsgGovCreateMarketRequest](#provenance-exchange-v1-MsgGovCreateMarketRequest) |  | create_market_request is the request to run validation on. |






<a name="provenance-exchange-v1-QueryValidateCreateMarketResponse"></a>

### QueryValidateCreateMarketResponse
QueryValidateCreateMarketResponse is a response message for the ValidateCreateMarket query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `error` | [string](#string) |  | error is any problems or inconsistencies in the provided gov prop msg. This goes above and beyond the validation done when actually processing the governance proposal. If an error is returned, and gov_prop_will_pass is true, it means the error is more of an inconsistency that might cause certain aspects of the market to behave unexpectedly. |
| `gov_prop_will_pass` | [bool](#bool) |  | gov_prop_will_pass will be true if the the provided msg will be successfully processed at the end of it's voting period (assuming it passes). |






<a name="provenance-exchange-v1-QueryValidateManageFeesRequest"></a>

### QueryValidateManageFeesRequest
QueryValidateManageFeesRequest is a request message for the ValidateManageFees query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `manage_fees_request` | [MsgGovManageFeesRequest](#provenance-exchange-v1-MsgGovManageFeesRequest) |  | manage_fees_request is the request to run validation on. |






<a name="provenance-exchange-v1-QueryValidateManageFeesResponse"></a>

### QueryValidateManageFeesResponse
QueryValidateManageFeesResponse is a response message for the ValidateManageFees query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `error` | [string](#string) |  | error is any problems or inconsistencies in the provided gov prop msg. This goes above and beyond the validation done when actually processing the governance proposal. If an error is returned, and gov_prop_will_pass is true, it means the error is more of an inconsistency that might cause certain aspects of the market to behave unexpectedly. |
| `gov_prop_will_pass` | [bool](#bool) |  | gov_prop_will_pass will be true if the the provided msg will be successfully processed at the end of it's voting period (assuming it passes). |






<a name="provenance-exchange-v1-QueryValidateMarketRequest"></a>

### QueryValidateMarketRequest
QueryValidateMarketRequest is a request message for the ValidateMarket query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the id of the market to check. |






<a name="provenance-exchange-v1-QueryValidateMarketResponse"></a>

### QueryValidateMarketResponse
QueryValidateMarketResponse is a response message for the ValidateMarket query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `error` | [string](#string) |  | error is any problems or inconsistencies in the provided market. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-exchange-v1-Query"></a>

### Query
Query is the service for exchange module's query endpoints.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `OrderFeeCalc` | [QueryOrderFeeCalcRequest](#provenance-exchange-v1-QueryOrderFeeCalcRequest) | [QueryOrderFeeCalcResponse](#provenance-exchange-v1-QueryOrderFeeCalcResponse) | OrderFeeCalc calculates the fees that will be associated with the provided order. |
| `GetOrder` | [QueryGetOrderRequest](#provenance-exchange-v1-QueryGetOrderRequest) | [QueryGetOrderResponse](#provenance-exchange-v1-QueryGetOrderResponse) | GetOrder looks up an order by id. |
| `GetOrderByExternalID` | [QueryGetOrderByExternalIDRequest](#provenance-exchange-v1-QueryGetOrderByExternalIDRequest) | [QueryGetOrderByExternalIDResponse](#provenance-exchange-v1-QueryGetOrderByExternalIDResponse) | GetOrderByExternalID looks up an order by market id and external id. |
| `GetMarketOrders` | [QueryGetMarketOrdersRequest](#provenance-exchange-v1-QueryGetMarketOrdersRequest) | [QueryGetMarketOrdersResponse](#provenance-exchange-v1-QueryGetMarketOrdersResponse) | GetMarketOrders looks up the orders in a market. |
| `GetOwnerOrders` | [QueryGetOwnerOrdersRequest](#provenance-exchange-v1-QueryGetOwnerOrdersRequest) | [QueryGetOwnerOrdersResponse](#provenance-exchange-v1-QueryGetOwnerOrdersResponse) | GetOwnerOrders looks up the orders from the provided owner address. |
| `GetAssetOrders` | [QueryGetAssetOrdersRequest](#provenance-exchange-v1-QueryGetAssetOrdersRequest) | [QueryGetAssetOrdersResponse](#provenance-exchange-v1-QueryGetAssetOrdersResponse) | GetAssetOrders looks up the orders for a specific asset denom. |
| `GetAllOrders` | [QueryGetAllOrdersRequest](#provenance-exchange-v1-QueryGetAllOrdersRequest) | [QueryGetAllOrdersResponse](#provenance-exchange-v1-QueryGetAllOrdersResponse) | GetAllOrders gets all orders in the exchange module. |
| `GetCommitment` | [QueryGetCommitmentRequest](#provenance-exchange-v1-QueryGetCommitmentRequest) | [QueryGetCommitmentResponse](#provenance-exchange-v1-QueryGetCommitmentResponse) | GetCommitment gets the funds in an account that are committed to the market. |
| `GetAccountCommitments` | [QueryGetAccountCommitmentsRequest](#provenance-exchange-v1-QueryGetAccountCommitmentsRequest) | [QueryGetAccountCommitmentsResponse](#provenance-exchange-v1-QueryGetAccountCommitmentsResponse) | GetAccountCommitments gets all the funds in an account that are committed to any market. |
| `GetMarketCommitments` | [QueryGetMarketCommitmentsRequest](#provenance-exchange-v1-QueryGetMarketCommitmentsRequest) | [QueryGetMarketCommitmentsResponse](#provenance-exchange-v1-QueryGetMarketCommitmentsResponse) | GetMarketCommitments gets all the funds committed to a market from any account. |
| `GetAllCommitments` | [QueryGetAllCommitmentsRequest](#provenance-exchange-v1-QueryGetAllCommitmentsRequest) | [QueryGetAllCommitmentsResponse](#provenance-exchange-v1-QueryGetAllCommitmentsResponse) | GetAllCommitments gets all fund committed to any market from any account. |
| `GetMarket` | [QueryGetMarketRequest](#provenance-exchange-v1-QueryGetMarketRequest) | [QueryGetMarketResponse](#provenance-exchange-v1-QueryGetMarketResponse) | GetMarket returns all the information and details about a market. |
| `GetAllMarkets` | [QueryGetAllMarketsRequest](#provenance-exchange-v1-QueryGetAllMarketsRequest) | [QueryGetAllMarketsResponse](#provenance-exchange-v1-QueryGetAllMarketsResponse) | GetAllMarkets returns brief information about each market. |
| `Params` | [QueryParamsRequest](#provenance-exchange-v1-QueryParamsRequest) | [QueryParamsResponse](#provenance-exchange-v1-QueryParamsResponse) | Params returns the exchange module parameters. |
| `CommitmentSettlementFeeCalc` | [QueryCommitmentSettlementFeeCalcRequest](#provenance-exchange-v1-QueryCommitmentSettlementFeeCalcRequest) | [QueryCommitmentSettlementFeeCalcResponse](#provenance-exchange-v1-QueryCommitmentSettlementFeeCalcResponse) | CommitmentSettlementFeeCalc calculates the fees a market will pay for a commitment settlement using current NAVs. |
| `ValidateCreateMarket` | [QueryValidateCreateMarketRequest](#provenance-exchange-v1-QueryValidateCreateMarketRequest) | [QueryValidateCreateMarketResponse](#provenance-exchange-v1-QueryValidateCreateMarketResponse) | ValidateCreateMarket checks the provided MsgGovCreateMarketResponse and returns any errors it might have. |
| `ValidateMarket` | [QueryValidateMarketRequest](#provenance-exchange-v1-QueryValidateMarketRequest) | [QueryValidateMarketResponse](#provenance-exchange-v1-QueryValidateMarketResponse) | ValidateMarket checks for any problems with a market's setup. |
| `ValidateManageFees` | [QueryValidateManageFeesRequest](#provenance-exchange-v1-QueryValidateManageFeesRequest) | [QueryValidateManageFeesResponse](#provenance-exchange-v1-QueryValidateManageFeesResponse) | ValidateManageFees checks the provided MsgGovManageFeesRequest and returns any errors that it might have. |
| `GetPayment` | [QueryGetPaymentRequest](#provenance-exchange-v1-QueryGetPaymentRequest) | [QueryGetPaymentResponse](#provenance-exchange-v1-QueryGetPaymentResponse) | GetPayment gets a single specific payment. |
| `GetPaymentsWithSource` | [QueryGetPaymentsWithSourceRequest](#provenance-exchange-v1-QueryGetPaymentsWithSourceRequest) | [QueryGetPaymentsWithSourceResponse](#provenance-exchange-v1-QueryGetPaymentsWithSourceResponse) | GetPaymentsWithSource gets all payments with a specific source account. |
| `GetPaymentsWithTarget` | [QueryGetPaymentsWithTargetRequest](#provenance-exchange-v1-QueryGetPaymentsWithTargetRequest) | [QueryGetPaymentsWithTargetResponse](#provenance-exchange-v1-QueryGetPaymentsWithTargetResponse) | GetPaymentsWithTarget gets all payments with a specific target account. |
| `GetAllPayments` | [QueryGetAllPaymentsRequest](#provenance-exchange-v1-QueryGetAllPaymentsRequest) | [QueryGetAllPaymentsResponse](#provenance-exchange-v1-QueryGetAllPaymentsResponse) | GetAllPayments gets all payments. |
| `PaymentFeeCalc` | [QueryPaymentFeeCalcRequest](#provenance-exchange-v1-QueryPaymentFeeCalcRequest) | [QueryPaymentFeeCalcResponse](#provenance-exchange-v1-QueryPaymentFeeCalcResponse) | PaymentFeeCalc calculates the fees that must be paid for creating or accepting a specific payment. |

 <!-- end services -->



<a name="provenance_exchange_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/genesis.proto



<a name="provenance-exchange-v1-GenesisState"></a>

### GenesisState
GenesisState is the data that should be loaded into the exchange module during genesis.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-exchange-v1-Params) |  | params defines all the parameters of the exchange module. |
| `markets` | [Market](#provenance-exchange-v1-Market) | repeated | markets are all of the markets to create at genesis. |
| `orders` | [Order](#provenance-exchange-v1-Order) | repeated | orders are all the orders to create at genesis. |
| `last_market_id` | [uint32](#uint32) |  | last_market_id is the value of the last auto-selected market id. |
| `last_order_id` | [uint64](#uint64) |  | last_order_id is the value of the last order id created. |
| `commitments` | [Commitment](#provenance-exchange-v1-Commitment) | repeated | commitments are all of the commitments to create at genesis. |
| `payments` | [Payment](#provenance-exchange-v1-Payment) | repeated | payments are all the payments to create at genesis. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_exchange_v1_orders-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/orders.proto



<a name="provenance-exchange-v1-AskOrder"></a>

### AskOrder
AskOrder represents someone's desire to sell something at a minimum price.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id identifies the market that this order belongs to. |
| `seller` | [string](#string) |  | seller is the address of the account that owns this order and has the assets to sell. |
| `assets` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | assets are the things that the seller wishes to sell. A hold is placed on this until the order is filled or cancelled. |
| `price` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | price is the minimum amount that the seller is willing to accept for the assets. The seller's settlement proportional fee (and possibly the settlement flat fee) is taken out of the amount the seller receives, so it's possible that the seller will still receive less than this price. |
| `seller_settlement_flat_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | seller_settlement_flat_fee is the flat fee for sellers that will be charged during settlement. If this denom is the same denom as the price, it will come out of the actual price received. If this denom is different, the amount must be in the seller's account and a hold is placed on it until the order is filled or cancelled. |
| `allow_partial` | [bool](#bool) |  | allow_partial should be true if partial fulfillment of this order should be allowed, and should be false if the order must be either filled in full or not filled at all. |
| `external_id` | [string](#string) |  | external_id is an optional string used to externally identify this order. Max length is 100 characters. If an order in this market with this external id already exists, this order will be rejected. |






<a name="provenance-exchange-v1-BidOrder"></a>

### BidOrder
BidOrder represents someone's desire to buy something at a specific price.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id identifies the market that this order belongs to. |
| `buyer` | [string](#string) |  | buyer is the address of the account that owns this order and has the price to spend. |
| `assets` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | assets are the things that the buyer wishes to buy. |
| `price` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | price is the amount that the buyer will pay for the assets. A hold is placed on this until the order is filled or cancelled. |
| `buyer_settlement_fees` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | buyer_settlement_fees are the fees (both flat and proportional) that the buyer will pay (in addition to the price) when the order is settled. A hold is placed on this until the order is filled or cancelled. |
| `allow_partial` | [bool](#bool) |  | allow_partial should be true if partial fulfillment of this order should be allowed, and should be false if the order must be either filled in full or not filled at all. |
| `external_id` | [string](#string) |  | external_id is an optional string used to externally identify this order. Max length is 100 characters. If an order in this market with this external id already exists, this order will be rejected. |






<a name="provenance-exchange-v1-Order"></a>

### Order
Order associates an order id with one of the order types.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier for this order. |
| `ask_order` | [AskOrder](#provenance-exchange-v1-AskOrder) |  | ask_order is the information about this order if it represents an ask order. |
| `bid_order` | [BidOrder](#provenance-exchange-v1-BidOrder) |  | bid_order is the information about this order if it represents a bid order. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_exchange_v1_params-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/params.proto



<a name="provenance-exchange-v1-DenomSplit"></a>

### DenomSplit
DenomSplit associates a coin denomination with an amount the exchange receives for that denom.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | denom is the coin denomination this split applies to. |
| `split` | [uint32](#uint32) |  | split is the proportion of fees the exchange receives for this denom in basis points. E.g. 100 = 1%. Min = 0, Max = 10000. |






<a name="provenance-exchange-v1-Params"></a>

### Params
Params is a representation of the exchange module parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `default_split` | [uint32](#uint32) |  | default_split is the default proportion of fees the exchange receives in basis points. It is used if there isn't an applicable denom-specific split defined. E.g. 100 = 1%. Min = 0, Max = 10000. |
| `denom_splits` | [DenomSplit](#provenance-exchange-v1-DenomSplit) | repeated | denom_splits are the denom-specific amounts the exchange receives. |
| `fee_create_payment_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | fee_create_payment_flat is the flat fee options for creating a payment. If the source amount is not zero then one of these fee entries is required to create the payment. This field is currently limited to zero or one entries. |
| `fee_accept_payment_flat` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | fee_accept_payment_flat is the flat fee options for accepting a payment. If the target amount is not zero then one of these fee entries is required to accept the payment. This field is currently limited to zero or one entries. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ledger_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ledger/v1/tx.proto



<a name="provenance-ledger-v1-MsgAppendRequest"></a>

### MsgAppendRequest
MsgAppendRequest


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [LedgerKey](#provenance-ledger-v1-LedgerKey) |  |  |
| `entries` | [LedgerEntry](#provenance-ledger-v1-LedgerEntry) | repeated |  |
| `authority` | [string](#string) |  |  |






<a name="provenance-ledger-v1-MsgAppendResponse"></a>

### MsgAppendResponse
MsgAppendResponse






<a name="provenance-ledger-v1-MsgCreateRequest"></a>

### MsgCreateRequest
MsgCreateRequest


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ledger` | [Ledger](#provenance-ledger-v1-Ledger) |  |  |
| `authority` | [string](#string) |  |  |






<a name="provenance-ledger-v1-MsgCreateResponse"></a>

### MsgCreateResponse
MsgCreateResponse






<a name="provenance-ledger-v1-MsgDestroyRequest"></a>

### MsgDestroyRequest
MsgDestroyRequest represents a request to destroy a ledger


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [LedgerKey](#provenance-ledger-v1-LedgerKey) |  |  |
| `authority` | [string](#string) |  |  |






<a name="provenance-ledger-v1-MsgDestroyResponse"></a>

### MsgDestroyResponse
MsgDestroyResponse represents the response from destroying a ledger






<a name="provenance-ledger-v1-MsgProcessFundTransfersRequest"></a>

### MsgProcessFundTransfersRequest
MsgProcessFundTransfersRequest represents a request to process multiple fund transfers


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `transfers` | [FundTransfer](#provenance-ledger-v1-FundTransfer) | repeated |  |






<a name="provenance-ledger-v1-MsgProcessFundTransfersResponse"></a>

### MsgProcessFundTransfersResponse
MsgProcessFundTransfersResponse represents the response from processing fund transfers






<a name="provenance-ledger-v1-MsgProcessFundTransfersWithSettlementRequest"></a>

### MsgProcessFundTransfersWithSettlementRequest
MsgProcessFundTransfersWithSettlementRequest represents a request to process fund transfers with settlement
instructions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  |  |
| `transfers` | [FundTransferWithSettlement](#provenance-ledger-v1-FundTransferWithSettlement) | repeated |  |






<a name="provenance-ledger-v1-MsgUpdateBalancesRequest"></a>

### MsgUpdateBalancesRequest
MsgUpdateBalancesRequest


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [LedgerKey](#provenance-ledger-v1-LedgerKey) |  |  |
| `authority` | [string](#string) |  |  |
| `correlation_id` | [string](#string) |  |  |
| `bucket_balances` | [BucketBalance](#provenance-ledger-v1-BucketBalance) | repeated |  |






<a name="provenance-ledger-v1-MsgUpdateBalancesResponse"></a>

### MsgUpdateBalancesResponse
MsgUpdateBalancesResponse





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-ledger-v1-Msg"></a>

### Msg
Msg defines the attribute module Msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Create` | [MsgCreateRequest](#provenance-ledger-v1-MsgCreateRequest) | [MsgCreateResponse](#provenance-ledger-v1-MsgCreateResponse) | Create a new NFT ledger |
| `Append` | [MsgAppendRequest](#provenance-ledger-v1-MsgAppendRequest) | [MsgAppendResponse](#provenance-ledger-v1-MsgAppendResponse) | Append a ledger entry |
| `UpdateBalances` | [MsgUpdateBalancesRequest](#provenance-ledger-v1-MsgUpdateBalancesRequest) | [MsgUpdateBalancesResponse](#provenance-ledger-v1-MsgUpdateBalancesResponse) | Balances can be updated for a ledger entry allowing for retroactive adjustments to be applied |
| `ProcessFundTransfers` | [MsgProcessFundTransfersRequest](#provenance-ledger-v1-MsgProcessFundTransfersRequest) | [MsgProcessFundTransfersResponse](#provenance-ledger-v1-MsgProcessFundTransfersResponse) | Process multiple fund transfers (payments and disbursements) |
| `ProcessFundTransfersWithSettlement` | [MsgProcessFundTransfersWithSettlementRequest](#provenance-ledger-v1-MsgProcessFundTransfersWithSettlementRequest) | [MsgProcessFundTransfersResponse](#provenance-ledger-v1-MsgProcessFundTransfersResponse) | Process multiple fund transfers with manual settlement instructions |
| `Destroy` | [MsgDestroyRequest](#provenance-ledger-v1-MsgDestroyRequest) | [MsgDestroyResponse](#provenance-ledger-v1-MsgDestroyResponse) | Destroy a ledger by NFT address |

 <!-- end services -->



<a name="provenance_ledger_v1_ledger-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ledger/v1/ledger.proto



<a name="provenance-ledger-v1-Balances"></a>

### Balances
Balances represents the current balances for principal, interest, and other amounts


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bucket_balances` | [BucketBalance](#provenance-ledger-v1-BucketBalance) | repeated |  |






<a name="provenance-ledger-v1-BucketBalance"></a>

### BucketBalance



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bucket_type_id` | [int32](#int32) |  | The bucket type specified by the LedgerClassBucketType.id |
| `balance` | [string](#string) |  | The balance of the bucket |






<a name="provenance-ledger-v1-Ledger"></a>

### Ledger
Ledger


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [LedgerKey](#provenance-ledger-v1-LedgerKey) |  |  |
| `ledger_class_id` | [string](#string) |  | Ledger class id for the ledger |
| `status_type_id` | [int32](#int32) |  | Status of the ledger |
| `next_pmt_date` | [int32](#int32) |  | Next payment date days since epoch |
| `next_pmt_amt` | [int64](#int64) |  | Next payment amount |
| `interest_rate` | [int32](#int32) |  | Interest rate |
| `maturity_date` | [int32](#int32) |  | Maturity date days since epoch |






<a name="provenance-ledger-v1-LedgerBucketAmount"></a>

### LedgerBucketAmount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bucket_type_id` | [int32](#int32) |  | The bucket type specified by the LedgerClassBucketType.id |
| `applied_amt` | [string](#string) |  | The amount applied to the bucket |






<a name="provenance-ledger-v1-LedgerClass"></a>

### LedgerClass
LedgerClass contains the configuration for a ledger related to a particular class of asset. The asset class
is defined by the either a scope spec `x/metadata`, or nft class `x/nft`. Ultimately, the configuration will
assist in verifying the types that are associated with particular ledger entries.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ledger_class_id` | [string](#string) |  | Unique ID for the ledger class (eg. 1, 2, 3, etc.) This is necessary since the nft class does not have an owner. |
| `asset_class_id` | [string](#string) |  | Scope Specification ID or NFT Class ID |
| `denom` | [string](#string) |  | Denom that this class of asset will be ledgered in |






<a name="provenance-ledger-v1-LedgerClassBucketType"></a>

### LedgerClassBucketType



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [int32](#int32) |  | Unique ID for the bucket type (eg. 1, 2, 3, etc.) |
| `code` | [string](#string) |  | Code for the bucket type (eg. "PRINCIPAL", "INTEREST", "OTHER") |
| `description` | [string](#string) |  | Description of the bucket type (eg. "Principal", "Interest", "Other") |






<a name="provenance-ledger-v1-LedgerClassEntryType"></a>

### LedgerClassEntryType
LedgerClassEntryType defines the types of possible ledger entries for a given asset class. These type codes allow
for minimal data storage while providing a human readable description of the entry type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [int32](#int32) |  | Unique ID for the entry type (eg. 1, 2, 3, etc.) |
| `code` | [string](#string) |  | Code for the entry type (eg. "DISBURSEMENT", "SCHEDULED_PAYMENT", "UNSCHEDULED_PAYMENT", "FORECLOSURE_PAYMENT", "FEE", "OTHER") |
| `description` | [string](#string) |  | Description of the entry type (eg. "Disbursement", "Scheduled Payment", "Unscheduled Payment", "Foreclosure Payment", "Fee", "Other") |






<a name="provenance-ledger-v1-LedgerClassStatusType"></a>

### LedgerClassStatusType



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [int32](#int32) |  | Unique ID for the status type (eg. 1, 2, 3, etc.) |
| `code` | [string](#string) |  | Code for the status type (eg. "IN_REPAYMENT", "IN_FORECLOSURE", "FORBEARANCE", "DEFERMENT", "BANKRUPTCY""CLOSED", "CANCELLED", "SUSPENDED", "OTHER") |
| `description` | [string](#string) |  | Description of the status type (eg. "In Repayment", "In Foreclosure", "Forbearance", "Deferment", "Bankruptcy", "Closed", "Cancelled", "Suspended", "Other") |






<a name="provenance-ledger-v1-LedgerEntry"></a>

### LedgerEntry
LedgerEntry


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `correlation_id` | [string](#string) |  | Correlation ID for tracking ledger entries with external systems (max 50 characters) |
| `reverses_correlation_id` | [string](#string) |  | If this entry reverses another entry, the correlation id of the entry it reverses |
| `is_void` | [bool](#bool) |  | If true, this entry is a void and should not be included in the ledger balance calculations |
| `sequence` | [uint32](#uint32) |  | The NFT address that this ledger entry pertains to Sequence number of the ledger entry (less than 100) This field is used to maintain the correct order of entries when multiple entries share the same effective date. Entries are sorted first by effective date, then by sequence. |
| `entry_type_id` | [int32](#int32) |  | The type of ledger entry specified by the LedgerClassEntryType.id |
| `posted_date` | [int32](#int32) |  | Posted date days since epoch |
| `effective_date` | [int32](#int32) |  | Effective date days since epoch |
| `total_amt` | [string](#string) |  |  |
| `applied_amounts` | [LedgerBucketAmount](#provenance-ledger-v1-LedgerBucketAmount) | repeated | Applied amounts for each bucket |
| `bucket_balances` | [LedgerEntry.BucketBalancesEntry](#provenance-ledger-v1-LedgerEntry-BucketBalancesEntry) | repeated | Balances for each bucket The key is the bucket type id |






<a name="provenance-ledger-v1-LedgerEntry-BucketBalancesEntry"></a>

### LedgerEntry.BucketBalancesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [int32](#int32) |  |  |
| `value` | [BucketBalance](#provenance-ledger-v1-BucketBalance) |  |  |






<a name="provenance-ledger-v1-LedgerKey"></a>

### LedgerKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `nft_id` | [string](#string) |  | Identifier for the nft that this ledger is linked to. This could be a `x/metadata` scope id or an `x/nft` nft id. In order to create a ledger for an nft, the nft class must be registered in the ledger module as a LedgerClass. |
| `asset_class_id` | [string](#string) |  | Scope Specification ID or NFT Class ID |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ledger_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ledger/v1/query.proto



<a name="provenance-ledger-v1-QueryBalancesAsOfRequest"></a>

### QueryBalancesAsOfRequest
QueryBalancesAsOfRequest is the request type for the Query/GetBalancesAsOf RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [LedgerKey](#provenance-ledger-v1-LedgerKey) |  |  |
| `as_of_date` | [string](#string) |  |  |






<a name="provenance-ledger-v1-QueryBalancesAsOfResponse"></a>

### QueryBalancesAsOfResponse
QueryBalancesAsOfResponse is the response type for the Query/GetBalancesAsOf RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balances` | [Balances](#provenance-ledger-v1-Balances) |  |  |






<a name="provenance-ledger-v1-QueryLedgerClassBucketTypesRequest"></a>

### QueryLedgerClassBucketTypesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_class_id` | [string](#string) |  |  |






<a name="provenance-ledger-v1-QueryLedgerClassBucketTypesResponse"></a>

### QueryLedgerClassBucketTypesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bucket_types` | [LedgerClassBucketType](#provenance-ledger-v1-LedgerClassBucketType) | repeated |  |






<a name="provenance-ledger-v1-QueryLedgerClassEntryTypesRequest"></a>

### QueryLedgerClassEntryTypesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_class_id` | [string](#string) |  |  |






<a name="provenance-ledger-v1-QueryLedgerClassEntryTypesResponse"></a>

### QueryLedgerClassEntryTypesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entry_types` | [LedgerClassEntryType](#provenance-ledger-v1-LedgerClassEntryType) | repeated |  |






<a name="provenance-ledger-v1-QueryLedgerClassStatusTypesRequest"></a>

### QueryLedgerClassStatusTypesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_class_id` | [string](#string) |  |  |






<a name="provenance-ledger-v1-QueryLedgerClassStatusTypesResponse"></a>

### QueryLedgerClassStatusTypesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status_types` | [LedgerClassStatusType](#provenance-ledger-v1-LedgerClassStatusType) | repeated |  |






<a name="provenance-ledger-v1-QueryLedgerConfigRequest"></a>

### QueryLedgerConfigRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [LedgerKey](#provenance-ledger-v1-LedgerKey) |  |  |






<a name="provenance-ledger-v1-QueryLedgerConfigResponse"></a>

### QueryLedgerConfigResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ledger` | [Ledger](#provenance-ledger-v1-Ledger) |  |  |






<a name="provenance-ledger-v1-QueryLedgerEntryRequest"></a>

### QueryLedgerEntryRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [LedgerKey](#provenance-ledger-v1-LedgerKey) |  |  |
| `correlation_id` | [string](#string) |  | Free-form string up to 50 characters |






<a name="provenance-ledger-v1-QueryLedgerEntryResponse"></a>

### QueryLedgerEntryResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entry` | [LedgerEntry](#provenance-ledger-v1-LedgerEntry) |  |  |






<a name="provenance-ledger-v1-QueryLedgerRequest"></a>

### QueryLedgerRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [LedgerKey](#provenance-ledger-v1-LedgerKey) |  |  |






<a name="provenance-ledger-v1-QueryLedgerResponse"></a>

### QueryLedgerResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entries` | [LedgerEntry](#provenance-ledger-v1-LedgerEntry) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-ledger-v1-Query"></a>

### Query
Query defines the gRPC querier service for ledger module.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Config` | [QueryLedgerConfigRequest](#provenance-ledger-v1-QueryLedgerConfigRequest) | [QueryLedgerConfigResponse](#provenance-ledger-v1-QueryLedgerConfigResponse) | Params queries params of the ledger module. |
| `Entries` | [QueryLedgerRequest](#provenance-ledger-v1-QueryLedgerRequest) | [QueryLedgerResponse](#provenance-ledger-v1-QueryLedgerResponse) |  |
| `ClassEntryTypes` | [QueryLedgerClassEntryTypesRequest](#provenance-ledger-v1-QueryLedgerClassEntryTypesRequest) | [QueryLedgerClassEntryTypesResponse](#provenance-ledger-v1-QueryLedgerClassEntryTypesResponse) |  |
| `ClassStatusTypes` | [QueryLedgerClassStatusTypesRequest](#provenance-ledger-v1-QueryLedgerClassStatusTypesRequest) | [QueryLedgerClassStatusTypesResponse](#provenance-ledger-v1-QueryLedgerClassStatusTypesResponse) |  |
| `ClassBucketTypes` | [QueryLedgerClassBucketTypesRequest](#provenance-ledger-v1-QueryLedgerClassBucketTypesRequest) | [QueryLedgerClassBucketTypesResponse](#provenance-ledger-v1-QueryLedgerClassBucketTypesResponse) |  |
| `GetLedgerEntry` | [QueryLedgerEntryRequest](#provenance-ledger-v1-QueryLedgerEntryRequest) | [QueryLedgerEntryResponse](#provenance-ledger-v1-QueryLedgerEntryResponse) | GetLedgerEntry returns a specific ledger entry for an NFT |
| `GetBalancesAsOf` | [QueryBalancesAsOfRequest](#provenance-ledger-v1-QueryBalancesAsOfRequest) | [QueryBalancesAsOfResponse](#provenance-ledger-v1-QueryBalancesAsOfResponse) | GetBalancesAsOf returns the balances for a specific NFT as of a given date |

 <!-- end services -->



<a name="provenance_ledger_v1_ledger_settlement-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ledger/v1/ledger_settlement.proto



<a name="provenance-ledger-v1-FundTransfer"></a>

### FundTransfer
FundTransfer represents a single fund transfer to process


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `nft_id` | [string](#string) |  |  |
| `ledger_entry_correlation_id` | [string](#string) |  |  |
| `amount` | [string](#string) |  |  |
| `status` | [FundingTransferStatus](#provenance-ledger-v1-FundingTransferStatus) |  |  |
| `memo` | [string](#string) |  |  |
| `settlement_block` | [int64](#int64) |  | The minimum block height or timestamp for settlement |






<a name="provenance-ledger-v1-FundTransferWithSettlement"></a>

### FundTransferWithSettlement
FundTransferEntryWithSettlement represents a fund transfer with settlement instructions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `nft_id` | [string](#string) |  |  |
| `ledger_entry_correlation_id` | [string](#string) |  |  |
| `settlementInstructions` | [SettlementInstruction](#provenance-ledger-v1-SettlementInstruction) | repeated |  |






<a name="provenance-ledger-v1-SettlementInstruction"></a>

### SettlementInstruction
SettlementInstruction represents blockchain-specific settlement instructions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [string](#string) |  |  |
| `recipient_address` | [string](#string) |  | The recipient's blockchain address |
| `memo` | [string](#string) |  | Optional memo or note for the transaction |
| `settlement_block` | [int64](#int64) |  | The minimum block height or timestamp for settlement |





 <!-- end messages -->


<a name="provenance-ledger-v1-FundingTransferStatus"></a>

### FundingTransferStatus
FlowStatus represents the current status of a flow

| Name | Number | Description |
| ---- | ------ | ----------- |
| `FUNDING_TRANSFER_STATUS_UNSPECIFIED` | `0` |  |
| `FUNDING_TRANSFER_STATUS_PENDING` | `1` |  |
| `FUNDING_TRANSFER_STATUS_PROCESSING` | `2` |  |
| `FUNDING_TRANSFER_STATUS_COMPLETED` | `3` |  |
| `FUNDING_TRANSFER_STATUS_FAILED` | `4` |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ledger_v1_ledger_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ledger/v1/ledger_query.proto



<a name="provenance-ledger-v1-LedgerBucketAmountPlainText"></a>

### LedgerBucketAmountPlainText



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bucket` | [LedgerClassBucketType](#provenance-ledger-v1-LedgerClassBucketType) |  |  |
| `applied_amt` | [string](#string) |  |  |
| `balance_amt` | [string](#string) |  |  |






<a name="provenance-ledger-v1-LedgerEntryPlainText"></a>

### LedgerEntryPlainText
LedgerEntry


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `correlation_id` | [string](#string) |  | Correlation ID for tracking ledger entries with external systems (max 50 characters) |
| `sequence` | [uint32](#uint32) |  | Sequence number of the ledger entry (less than 100) This field is used to maintain the correct order of entries when multiple entries share the same effective date. Entries are sorted first by effective date, then by sequence. |
| `type` | [LedgerClassEntryType](#provenance-ledger-v1-LedgerClassEntryType) |  | The type of ledger entry specified by the LedgerClassEntryType.id |
| `posted_date` | [string](#string) |  | Posted date |
| `effective_date` | [string](#string) |  | Effective date |
| `total_amt` | [string](#string) |  | The total amount of the ledger entry |
| `applied_amounts` | [LedgerBucketAmountPlainText](#provenance-ledger-v1-LedgerBucketAmountPlainText) | repeated | The amounts applied to each bucket |






<a name="provenance-ledger-v1-LedgerPlainText"></a>

### LedgerPlainText
Ledger


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [LedgerKey](#provenance-ledger-v1-LedgerKey) |  | Ledger key |
| `status` | [string](#string) |  | Status of the ledger |
| `next_pmt_date` | [string](#string) |  | Next payment date |
| `next_pmt_amt` | [string](#string) |  | Next payment amount |
| `interest_rate` | [string](#string) |  | Interest rate |
| `maturity_date` | [string](#string) |  | Maturity date |






<a name="provenance-ledger-v1-QueryLedgerEntryResponsePlainText"></a>

### QueryLedgerEntryResponsePlainText



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entries` | [LedgerEntryPlainText](#provenance-ledger-v1-LedgerEntryPlainText) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ledger_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ledger/v1/genesis.proto



<a name="provenance-ledger-v1-GenesisState"></a>

### GenesisState
Initial state of the ledger store.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_trigger_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/tx.proto



<a name="provenance-trigger-v1-MsgCreateTriggerRequest"></a>

### MsgCreateTriggerRequest
MsgCreateTriggerRequest is the request type for creating a trigger RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authorities` | [string](#string) | repeated | The signing authorities for the request |
| `event` | [google.protobuf.Any](#google-protobuf-Any) |  | The event that must be detected for the trigger to fire. |
| `actions` | [google.protobuf.Any](#google-protobuf-Any) | repeated | The messages to run when the trigger fires. |






<a name="provenance-trigger-v1-MsgCreateTriggerResponse"></a>

### MsgCreateTriggerResponse
MsgCreateTriggerResponse is the response type for creating a trigger RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | trigger id that is generated on creation. |






<a name="provenance-trigger-v1-MsgDestroyTriggerRequest"></a>

### MsgDestroyTriggerRequest
MsgDestroyTriggerRequest is the request type for creating a trigger RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | the id of the trigger to destroy. |
| `authority` | [string](#string) |  | The signing authority for the request |






<a name="provenance-trigger-v1-MsgDestroyTriggerResponse"></a>

### MsgDestroyTriggerResponse
MsgDestroyTriggerResponse is the response type for creating a trigger RPC





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-trigger-v1-Msg"></a>

### Msg
Msg

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `CreateTrigger` | [MsgCreateTriggerRequest](#provenance-trigger-v1-MsgCreateTriggerRequest) | [MsgCreateTriggerResponse](#provenance-trigger-v1-MsgCreateTriggerResponse) | CreateTrigger is the RPC endpoint for creating a trigger |
| `DestroyTrigger` | [MsgDestroyTriggerRequest](#provenance-trigger-v1-MsgDestroyTriggerRequest) | [MsgDestroyTriggerResponse](#provenance-trigger-v1-MsgDestroyTriggerResponse) | DestroyTrigger is the RPC endpoint for creating a trigger |

 <!-- end services -->



<a name="provenance_trigger_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/query.proto



<a name="provenance-trigger-v1-QueryTriggerByIDRequest"></a>

### QueryTriggerByIDRequest
QueryTriggerByIDRequest queries for the Trigger with an identifier of id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | The id of the trigger to query. |






<a name="provenance-trigger-v1-QueryTriggerByIDResponse"></a>

### QueryTriggerByIDResponse
QueryTriggerByIDResponse contains the requested Trigger.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger` | [Trigger](#provenance-trigger-v1-Trigger) |  | The trigger object that was queried for. |






<a name="provenance-trigger-v1-QueryTriggersRequest"></a>

### QueryTriggersRequest
QueryTriggersRequest queries for all triggers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-trigger-v1-QueryTriggersResponse"></a>

### QueryTriggersResponse
QueryTriggersResponse contains the list of Triggers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `triggers` | [Trigger](#provenance-trigger-v1-Trigger) | repeated | List of Trigger objects. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-trigger-v1-Query"></a>

### Query
Query defines the gRPC querier service for trigger module.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `TriggerByID` | [QueryTriggerByIDRequest](#provenance-trigger-v1-QueryTriggerByIDRequest) | [QueryTriggerByIDResponse](#provenance-trigger-v1-QueryTriggerByIDResponse) | TriggerByID returns a trigger matching the ID. |
| `Triggers` | [QueryTriggersRequest](#provenance-trigger-v1-QueryTriggersRequest) | [QueryTriggersResponse](#provenance-trigger-v1-QueryTriggersResponse) | Triggers returns the list of triggers. |

 <!-- end services -->



<a name="provenance_trigger_v1_event-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/event.proto



<a name="provenance-trigger-v1-EventTriggerCreated"></a>

### EventTriggerCreated
EventTriggerCreated is an event for when a trigger is created


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [string](#string) |  | trigger_id is a unique identifier of the trigger. |






<a name="provenance-trigger-v1-EventTriggerDestroyed"></a>

### EventTriggerDestroyed
EventTriggerDestroyed is an event for when a trigger is destroyed


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [string](#string) |  | trigger_id is a unique identifier of the trigger. |






<a name="provenance-trigger-v1-EventTriggerDetected"></a>

### EventTriggerDetected
EventTriggerDetected is an event for when a trigger's event is detected


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [string](#string) |  | trigger_id is a unique identifier of the trigger. |






<a name="provenance-trigger-v1-EventTriggerExecuted"></a>

### EventTriggerExecuted
EventTriggerExecuted is an event for when a trigger is executed.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [string](#string) |  | trigger_id is a unique identifier of the trigger. |
| `owner` | [string](#string) |  | owner is the creator of the trigger. |
| `success` | [bool](#bool) |  | success indicates if all executed actions were successful. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_trigger_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/genesis.proto



<a name="provenance-trigger-v1-GasLimit"></a>

### GasLimit
GasLimit defines the trigger module's grouping of a trigger and a gas limit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [uint64](#uint64) |  | The identifier of the trigger this GasLimit belongs to. |
| `amount` | [uint64](#uint64) |  | The maximum amount of gas that the trigger can use. |






<a name="provenance-trigger-v1-GenesisState"></a>

### GenesisState
GenesisState defines the trigger module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [uint64](#uint64) |  | Trigger id is the next auto incremented id to be assigned to the next created trigger |
| `queue_start` | [uint64](#uint64) |  | Queue start is the starting index of the queue. |
| `triggers` | [Trigger](#provenance-trigger-v1-Trigger) | repeated | Triggers to initially start with. |
| `gas_limits` | [GasLimit](#provenance-trigger-v1-GasLimit) | repeated | Maximum amount of gas that the triggers can use. |
| `queued_triggers` | [QueuedTrigger](#provenance-trigger-v1-QueuedTrigger) | repeated | Triggers to initially start with in the queue. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_trigger_v1_trigger-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/trigger.proto



<a name="provenance-trigger-v1-Attribute"></a>

### Attribute
Attribute


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The name of the attribute that the event must have to be considered a match. |
| `value` | [string](#string) |  | The value of the attribute that the event must have to be considered a match. |






<a name="provenance-trigger-v1-BlockHeightEvent"></a>

### BlockHeightEvent
BlockHeightEvent


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `block_height` | [uint64](#uint64) |  | The height that the trigger should fire at. |






<a name="provenance-trigger-v1-BlockTimeEvent"></a>

### BlockTimeEvent
BlockTimeEvent


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `time` | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The time the trigger should fire at. |






<a name="provenance-trigger-v1-QueuedTrigger"></a>

### QueuedTrigger
QueuedTrigger


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `block_height` | [uint64](#uint64) |  | The block height the trigger was detected and queued. |
| `time` | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The time the trigger was detected and queued. |
| `trigger` | [Trigger](#provenance-trigger-v1-Trigger) |  | The trigger that was detected. |






<a name="provenance-trigger-v1-TransactionEvent"></a>

### TransactionEvent
TransactionEvent


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The name of the event for a match. |
| `attributes` | [Attribute](#provenance-trigger-v1-Attribute) | repeated | The attributes that must be present for a match. |






<a name="provenance-trigger-v1-Trigger"></a>

### Trigger
Trigger


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | An integer to uniquely identify the trigger. |
| `owner` | [string](#string) |  | The owner of the trigger. |
| `event` | [google.protobuf.Any](#google-protobuf-Any) |  | The event that must be detected for the trigger to fire. |
| `actions` | [google.protobuf.Any](#google-protobuf-Any) | repeated | The messages to run when the trigger fires. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_attribute_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/attribute/v1/tx.proto



<a name="provenance-attribute-v1-MsgAddAttributeRequest"></a>

### MsgAddAttributeRequest
MsgAddAttributeRequest defines an sdk.Msg type that is used to add a new attribute to an account.
Attributes may only be set in an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The attribute value. |
| `attribute_type` | [AttributeType](#provenance-attribute-v1-AttributeType) |  | The attribute value type. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |
| `expiration_date` | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time that an attribute will expire. |






<a name="provenance-attribute-v1-MsgAddAttributeResponse"></a>

### MsgAddAttributeResponse
MsgAddAttributeResponse defines the Msg/AddAttribute response type.






<a name="provenance-attribute-v1-MsgDeleteAttributeRequest"></a>

### MsgDeleteAttributeRequest
MsgDeleteAttributeRequest defines a message to delete an attribute from an account
Attributes may only be removed from an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance-attribute-v1-MsgDeleteAttributeResponse"></a>

### MsgDeleteAttributeResponse
MsgDeleteAttributeResponse defines the Msg/DeleteAttribute response type.






<a name="provenance-attribute-v1-MsgDeleteDistinctAttributeRequest"></a>

### MsgDeleteDistinctAttributeRequest
MsgDeleteDistinctAttributeRequest defines a message to delete an attribute with matching name, value, and type from
an account. Attributes may only be removed from an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The attribute value. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance-attribute-v1-MsgDeleteDistinctAttributeResponse"></a>

### MsgDeleteDistinctAttributeResponse
MsgDeleteDistinctAttributeResponse defines the Msg/DeleteDistinctAttribute response type.






<a name="provenance-attribute-v1-MsgSetAccountDataRequest"></a>

### MsgSetAccountDataRequest
MsgSetAccountDataRequest defines a message to set an account's accountdata attribute.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |






<a name="provenance-attribute-v1-MsgSetAccountDataResponse"></a>

### MsgSetAccountDataResponse
MsgSetAccountDataResponse defines the Msg/SetAccountData response type.






<a name="provenance-attribute-v1-MsgUpdateAttributeExpirationRequest"></a>

### MsgUpdateAttributeExpirationRequest
MsgUpdateAttributeExpirationRequest defines an sdk.Msg type that is used to update an existing attribute's expiration
date


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The original attribute value. |
| `expiration_date` | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time that an attribute will expire. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance-attribute-v1-MsgUpdateAttributeExpirationResponse"></a>

### MsgUpdateAttributeExpirationResponse
MsgUpdateAttributeExpirationResponse defines the Msg/Vote response type.






<a name="provenance-attribute-v1-MsgUpdateAttributeRequest"></a>

### MsgUpdateAttributeRequest
MsgUpdateAttributeRequest defines an sdk.Msg type that is used to update an existing attribute to an account.
Attributes may only be set in an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `original_value` | [bytes](#bytes) |  | The original attribute value. |
| `update_value` | [bytes](#bytes) |  | The update attribute value. |
| `original_attribute_type` | [AttributeType](#provenance-attribute-v1-AttributeType) |  | The original attribute value type. |
| `update_attribute_type` | [AttributeType](#provenance-attribute-v1-AttributeType) |  | The update attribute value type. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance-attribute-v1-MsgUpdateAttributeResponse"></a>

### MsgUpdateAttributeResponse
MsgUpdateAttributeResponse defines the Msg/UpdateAttribute response type.






<a name="provenance-attribute-v1-MsgUpdateParamsRequest"></a>

### MsgUpdateParamsRequest
MsgUpdateParamsRequest is a request message for the UpdateParams endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `params` | [Params](#provenance-attribute-v1-Params) |  | params are the new param values to set. |






<a name="provenance-attribute-v1-MsgUpdateParamsResponse"></a>

### MsgUpdateParamsResponse
MsgUpdateParamsResponse is a response message for the UpdateParams endpoint.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-attribute-v1-Msg"></a>

### Msg
Msg defines the attribute module Msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `AddAttribute` | [MsgAddAttributeRequest](#provenance-attribute-v1-MsgAddAttributeRequest) | [MsgAddAttributeResponse](#provenance-attribute-v1-MsgAddAttributeResponse) | AddAttribute defines a method to verify a particular invariance. |
| `UpdateAttribute` | [MsgUpdateAttributeRequest](#provenance-attribute-v1-MsgUpdateAttributeRequest) | [MsgUpdateAttributeResponse](#provenance-attribute-v1-MsgUpdateAttributeResponse) | UpdateAttribute defines a method to verify a particular invariance. |
| `UpdateAttributeExpiration` | [MsgUpdateAttributeExpirationRequest](#provenance-attribute-v1-MsgUpdateAttributeExpirationRequest) | [MsgUpdateAttributeExpirationResponse](#provenance-attribute-v1-MsgUpdateAttributeExpirationResponse) | UpdateAttributeExpiration defines a method to verify a particular invariance. |
| `DeleteAttribute` | [MsgDeleteAttributeRequest](#provenance-attribute-v1-MsgDeleteAttributeRequest) | [MsgDeleteAttributeResponse](#provenance-attribute-v1-MsgDeleteAttributeResponse) | DeleteAttribute defines a method to verify a particular invariance. |
| `DeleteDistinctAttribute` | [MsgDeleteDistinctAttributeRequest](#provenance-attribute-v1-MsgDeleteDistinctAttributeRequest) | [MsgDeleteDistinctAttributeResponse](#provenance-attribute-v1-MsgDeleteDistinctAttributeResponse) | DeleteDistinctAttribute defines a method to verify a particular invariance. |
| `SetAccountData` | [MsgSetAccountDataRequest](#provenance-attribute-v1-MsgSetAccountDataRequest) | [MsgSetAccountDataResponse](#provenance-attribute-v1-MsgSetAccountDataResponse) | SetAccountData defines a method for setting/updating an account's accountdata attribute. |
| `UpdateParams` | [MsgUpdateParamsRequest](#provenance-attribute-v1-MsgUpdateParamsRequest) | [MsgUpdateParamsResponse](#provenance-attribute-v1-MsgUpdateParamsResponse) | UpdateParams is a governance proposal endpoint for updating the attribute module's params. |

 <!-- end services -->



<a name="provenance_attribute_v1_attribute-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/attribute/v1/attribute.proto



<a name="provenance-attribute-v1-Attribute"></a>

### Attribute
Attribute holds a typed key/value structure for data associated with an account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The attribute value. |
| `attribute_type` | [AttributeType](#provenance-attribute-v1-AttributeType) |  | The attribute value type. |
| `address` | [string](#string) |  | The address the attribute is bound to |
| `expiration_date` | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Time that an attribute will expire. |






<a name="provenance-attribute-v1-EventAccountDataUpdated"></a>

### EventAccountDataUpdated
EventAccountDataUpdated event emitted when accountdata is set, updated, or deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  |  |






<a name="provenance-attribute-v1-EventAttributeAdd"></a>

### EventAttributeAdd
EventAttributeAdd event emitted when attribute is added


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |
| `type` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |
| `expiration` | [string](#string) |  |  |






<a name="provenance-attribute-v1-EventAttributeDelete"></a>

### EventAttributeDelete
EventAttributeDelete event emitted when attribute is deleted


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="provenance-attribute-v1-EventAttributeDistinctDelete"></a>

### EventAttributeDistinctDelete
EventAttributeDistinctDelete event emitted when attribute is deleted with matching value


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |
| `attribute_type` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="provenance-attribute-v1-EventAttributeExpirationUpdate"></a>

### EventAttributeExpirationUpdate
EventAttributeExpirationUpdate event emitted when attribute expiration is updated


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |
| `original_expiration` | [string](#string) |  |  |
| `updated_expiration` | [string](#string) |  |  |






<a name="provenance-attribute-v1-EventAttributeExpired"></a>

### EventAttributeExpired
EventAttributeExpired event emitted when attribute has expired and been deleted in BeginBlocker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `value_hash` | [string](#string) |  |  |
| `attribute_type` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `expiration` | [string](#string) |  |  |






<a name="provenance-attribute-v1-EventAttributeParamsUpdated"></a>

### EventAttributeParamsUpdated
EventAttributeParamsUpdated event emitted when attribute params are updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_value_length` | [string](#string) |  |  |






<a name="provenance-attribute-v1-EventAttributeUpdate"></a>

### EventAttributeUpdate
EventAttributeUpdate event emitted when attribute is updated


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `original_value` | [string](#string) |  |  |
| `original_type` | [string](#string) |  |  |
| `update_value` | [string](#string) |  |  |
| `update_type` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="provenance-attribute-v1-Params"></a>

### Params
Params defines the set of params for the attribute module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_value_length` | [uint32](#uint32) |  | maximum length of data to allow in an attribute value |





 <!-- end messages -->


<a name="provenance-attribute-v1-AttributeType"></a>

### AttributeType
AttributeType defines the type of the data stored in the attribute value

| Name | Number | Description |
| ---- | ------ | ----------- |
| `ATTRIBUTE_TYPE_UNSPECIFIED` | `0` | ATTRIBUTE_TYPE_UNSPECIFIED defines an unknown/invalid type |
| `ATTRIBUTE_TYPE_UUID` | `1` | ATTRIBUTE_TYPE_UUID defines an attribute value that contains a string value representation of a V4 uuid |
| `ATTRIBUTE_TYPE_JSON` | `2` | ATTRIBUTE_TYPE_JSON defines an attribute value that contains a byte string containing json data |
| `ATTRIBUTE_TYPE_STRING` | `3` | ATTRIBUTE_TYPE_STRING defines an attribute value that contains a generic string value |
| `ATTRIBUTE_TYPE_URI` | `4` | ATTRIBUTE_TYPE_URI defines an attribute value that contains a URI |
| `ATTRIBUTE_TYPE_INT` | `5` | ATTRIBUTE_TYPE_INT defines an attribute value that contains an integer (cast as int64) |
| `ATTRIBUTE_TYPE_FLOAT` | `6` | ATTRIBUTE_TYPE_FLOAT defines an attribute value that contains a float |
| `ATTRIBUTE_TYPE_PROTO` | `7` | ATTRIBUTE_TYPE_PROTO defines an attribute value that contains a serialized proto value in bytes |
| `ATTRIBUTE_TYPE_BYTES` | `8` | ATTRIBUTE_TYPE_BYTES defines an attribute value that contains an untyped array of bytes |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_attribute_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/attribute/v1/query.proto



<a name="provenance-attribute-v1-QueryAccountDataRequest"></a>

### QueryAccountDataRequest
QueryAccountDataRequest is the request type for the Query/AccountData method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account is the bech32 address of the account to get the data for |






<a name="provenance-attribute-v1-QueryAccountDataResponse"></a>

### QueryAccountDataResponse
QueryAccountDataResponse is the response type for the Query/AccountData method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  | value is the accountdata attribute value for the requested account. |






<a name="provenance-attribute-v1-QueryAttributeAccountsRequest"></a>

### QueryAttributeAccountsRequest
QueryAttributeAccountsRequest is the request type for the Query/AttributeAccounts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `attribute_name` | [string](#string) |  | name is the attribute name to query for |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-attribute-v1-QueryAttributeAccountsResponse"></a>

### QueryAttributeAccountsResponse
QueryAttributeAccountsResponse is the response type for the Query/AttributeAccounts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `accounts` | [string](#string) | repeated | list of account addresses that have attributes of request name |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance-attribute-v1-QueryAttributeRequest"></a>

### QueryAttributeRequest
QueryAttributeRequest is the request type for the Query/Attribute method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account defines the address to query for. |
| `name` | [string](#string) |  | name is the attribute name to query for |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-attribute-v1-QueryAttributeResponse"></a>

### QueryAttributeResponse
QueryAttributeResponse is the response type for the Query/Attribute method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | a string containing the address of the account the attributes are assigned to. |
| `attributes` | [Attribute](#provenance-attribute-v1-Attribute) | repeated | a list of attribute values |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance-attribute-v1-QueryAttributesRequest"></a>

### QueryAttributesRequest
QueryAttributesRequest is the request type for the Query/Attributes method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account defines the address to query for. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-attribute-v1-QueryAttributesResponse"></a>

### QueryAttributesResponse
QueryAttributesResponse is the response type for the Query/Attributes method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | a string containing the address of the account the attributes are assigned to= |
| `attributes` | [Attribute](#provenance-attribute-v1-Attribute) | repeated | a list of attribute values |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance-attribute-v1-QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance-attribute-v1-QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-attribute-v1-Params) |  | params defines the parameters of the module. |






<a name="provenance-attribute-v1-QueryScanRequest"></a>

### QueryScanRequest
QueryScanRequest is the request type for the Query/Scan method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account defines the address to query for. |
| `suffix` | [string](#string) |  | name defines the partial attribute name to search for base on names being in RDNS format. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-attribute-v1-QueryScanResponse"></a>

### QueryScanResponse
QueryScanResponse is the response type for the Query/Scan method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | a string containing the address of the account the attributes are assigned to= |
| `attributes` | [Attribute](#provenance-attribute-v1-Attribute) | repeated | a list of attribute values |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the request. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-attribute-v1-Query"></a>

### Query
Query defines the gRPC querier service for attribute module.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Params` | [QueryParamsRequest](#provenance-attribute-v1-QueryParamsRequest) | [QueryParamsResponse](#provenance-attribute-v1-QueryParamsResponse) | Params queries params of the attribute module. |
| `Attribute` | [QueryAttributeRequest](#provenance-attribute-v1-QueryAttributeRequest) | [QueryAttributeResponse](#provenance-attribute-v1-QueryAttributeResponse) | Attribute queries attributes on a given account (address) for one (or more) with the given name |
| `Attributes` | [QueryAttributesRequest](#provenance-attribute-v1-QueryAttributesRequest) | [QueryAttributesResponse](#provenance-attribute-v1-QueryAttributesResponse) | Attributes queries attributes on a given account (address) for any defined attributes |
| `Scan` | [QueryScanRequest](#provenance-attribute-v1-QueryScanRequest) | [QueryScanResponse](#provenance-attribute-v1-QueryScanResponse) | Scan queries attributes on a given account (address) for any that match the provided suffix |
| `AttributeAccounts` | [QueryAttributeAccountsRequest](#provenance-attribute-v1-QueryAttributeAccountsRequest) | [QueryAttributeAccountsResponse](#provenance-attribute-v1-QueryAttributeAccountsResponse) | AttributeAccounts queries accounts on a given attribute name |
| `AccountData` | [QueryAccountDataRequest](#provenance-attribute-v1-QueryAccountDataRequest) | [QueryAccountDataResponse](#provenance-attribute-v1-QueryAccountDataResponse) | AccountData returns the accountdata for a specified account. |

 <!-- end services -->



<a name="provenance_attribute_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/attribute/v1/genesis.proto



<a name="provenance-attribute-v1-GenesisState"></a>

### GenesisState
GenesisState defines the attribute module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-attribute-v1-Params) |  | params defines all the parameters of the module. |
| `attributes` | [Attribute](#provenance-attribute-v1-Attribute) | repeated | deposits defines all the deposits present at genesis. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_asset_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/asset/v1/tx.proto



<a name="provenance-asset-v1-MsgAddAsset"></a>

### MsgAddAsset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [Asset](#provenance-asset-v1-Asset) |  |  |
| `from_address` | [string](#string) |  |  |






<a name="provenance-asset-v1-MsgAddAssetClass"></a>

### MsgAddAssetClass



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset_class` | [AssetClass](#provenance-asset-v1-AssetClass) |  |  |
| `from_address` | [string](#string) |  |  |






<a name="provenance-asset-v1-MsgAddAssetClassResponse"></a>

### MsgAddAssetClassResponse







<a name="provenance-asset-v1-MsgAddAssetResponse"></a>

### MsgAddAssetResponse






 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-asset-v1-Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `AddAsset` | [MsgAddAsset](#provenance-asset-v1-MsgAddAsset) | [MsgAddAssetResponse](#provenance-asset-v1-MsgAddAssetResponse) |  |
| `AddAssetClass` | [MsgAddAssetClass](#provenance-asset-v1-MsgAddAssetClass) | [MsgAddAssetClassResponse](#provenance-asset-v1-MsgAddAssetClassResponse) |  |

 <!-- end services -->



<a name="provenance_asset_v1_asset-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/asset/v1/asset.proto



<a name="provenance-asset-v1-Asset"></a>

### Asset
Asset defines the asset.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `class_id` | [string](#string) |  | class_id associated with the asset, similar to the contract address of ERC721 |
| `id` | [string](#string) |  | id is a unique identifier of the asseet |
| `uri` | [string](#string) |  | uri for the asset metadata stored off chain |
| `uri_hash` | [string](#string) |  | uri_hash is a hash of the document pointed by uri |
| `data` | [string](#string) |  | data is an app specific json data of the asset |






<a name="provenance-asset-v1-AssetClass"></a>

### AssetClass



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | id defines the unique identifier of the asset classification, similar to the contract address of ERC721 |
| `name` | [string](#string) |  | name defines the human-readable name of the asset classification |
| `symbol` | [string](#string) |  | symbol is an abbreviated name for asset classification |
| `description` | [string](#string) |  | description is a brief description of asset classification |
| `uri` | [string](#string) |  | uri for the class metadata stored off chain. It can define schema for Class and asset `Data` attributes |
| `uri_hash` | [string](#string) |  | uri_hash is a hash of the document pointed by uri |
| `data` | [string](#string) |  | data is the app specific json schema of the asset class |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_asset_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/asset/v1/query.proto



<a name="provenance-asset-v1-QueryGetClass"></a>

### QueryGetClass



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  |  |






<a name="provenance-asset-v1-QueryGetClassResponse"></a>

### QueryGetClassResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assetClass` | [AssetClass](#provenance-asset-v1-AssetClass) |  |  |






<a name="provenance-asset-v1-QueryListAssetClasses"></a>

### QueryListAssetClasses







<a name="provenance-asset-v1-QueryListAssetClassesResponse"></a>

### QueryListAssetClassesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assetClasses` | [AssetClass](#provenance-asset-v1-AssetClass) | repeated |  |






<a name="provenance-asset-v1-QueryListAssets"></a>

### QueryListAssets







<a name="provenance-asset-v1-QueryListAssetsResponse"></a>

### QueryListAssetsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `assets` | [Asset](#provenance-asset-v1-Asset) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-asset-v1-Query"></a>

### Query


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `ListAssets` | [QueryListAssets](#provenance-asset-v1-QueryListAssets) | [QueryListAssetsResponse](#provenance-asset-v1-QueryListAssetsResponse) |  |
| `ListAssetClasses` | [QueryListAssetClasses](#provenance-asset-v1-QueryListAssetClasses) | [QueryListAssetClassesResponse](#provenance-asset-v1-QueryListAssetClassesResponse) |  |
| `GetClass` | [QueryGetClass](#provenance-asset-v1-QueryGetClass) | [QueryGetClassResponse](#provenance-asset-v1-QueryGetClassResponse) |  |

 <!-- end services -->



<a name="provenance_asset_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/asset/v1/genesis.proto



<a name="provenance-asset-v1-GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `asset` | [Asset](#provenance-asset-v1-Asset) | repeated |  |
| `asset_classes` | [AssetClass](#provenance-asset-v1-AssetClass) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_msgfees_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/tx.proto



<a name="provenance-msgfees-v1-MsgAddMsgFeeProposalRequest"></a>

### MsgAddMsgFeeProposalRequest
AddMsgFeeProposal defines a governance proposal to add additional msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type_url` | [string](#string) |  | type url of msg to add fee |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | additional fee for msg type |
| `recipient` | [string](#string) |  | optional recipient to receive basis points |
| `recipient_basis_points` | [string](#string) |  | basis points to use when recipient is present (1 - 10,000) |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance-msgfees-v1-MsgAddMsgFeeProposalResponse"></a>

### MsgAddMsgFeeProposalResponse
MsgAddMsgFeeProposalResponse defines the Msg/AddMsgFeeProposal response type






<a name="provenance-msgfees-v1-MsgAssessCustomMsgFeeRequest"></a>

### MsgAssessCustomMsgFeeRequest
MsgAssessCustomMsgFeeRequest defines an sdk.Msg type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | optional short name for custom msg fee, this will be emitted as a property of the event |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | amount of additional fee that must be paid |
| `recipient` | [string](#string) |  | optional recipient address, the basis points amount is sent to the recipient |
| `from` | [string](#string) |  | the signer of the msg |
| `recipient_basis_points` | [string](#string) |  | optional basis points 0 - 10,000 for recipient defaults to 10,000 |






<a name="provenance-msgfees-v1-MsgAssessCustomMsgFeeResponse"></a>

### MsgAssessCustomMsgFeeResponse
MsgAssessCustomMsgFeeResponse defines the Msg/AssessCustomMsgFeee response type.






<a name="provenance-msgfees-v1-MsgRemoveMsgFeeProposalRequest"></a>

### MsgRemoveMsgFeeProposalRequest
RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type_url` | [string](#string) |  | type url of msg fee to remove |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance-msgfees-v1-MsgRemoveMsgFeeProposalResponse"></a>

### MsgRemoveMsgFeeProposalResponse
MsgRemoveMsgFeeProposalResponse defines the Msg/RemoveMsgFeeProposal response type






<a name="provenance-msgfees-v1-MsgUpdateConversionFeeDenomProposalRequest"></a>

### MsgUpdateConversionFeeDenomProposalRequest
UpdateConversionFeeDenomProposal defines a governance proposal to update the msg fee conversion denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `conversion_fee_denom` | [string](#string) |  | conversion_fee_denom is the denom that usd will be converted to |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance-msgfees-v1-MsgUpdateConversionFeeDenomProposalResponse"></a>

### MsgUpdateConversionFeeDenomProposalResponse
MsgUpdateConversionFeeDenomProposalResponse defines the Msg/UpdateConversionFeeDenomProposal response type






<a name="provenance-msgfees-v1-MsgUpdateMsgFeeProposalRequest"></a>

### MsgUpdateMsgFeeProposalRequest
UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type_url` | [string](#string) |  | type url of msg to update fee |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | additional fee for msg type |
| `recipient` | [string](#string) |  | optional recipient to receive basis points |
| `recipient_basis_points` | [string](#string) |  | basis points to use when recipient is present (1 - 10,000) |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance-msgfees-v1-MsgUpdateMsgFeeProposalResponse"></a>

### MsgUpdateMsgFeeProposalResponse
MsgUpdateMsgFeeProposalResponse defines the Msg/RemoveMsgFeeProposal response type






<a name="provenance-msgfees-v1-MsgUpdateNhashPerUsdMilProposalRequest"></a>

### MsgUpdateNhashPerUsdMilProposalRequest
UpdateNhashPerUsdMilProposal defines a governance proposal to update the nhash per usd mil param


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `nhash_per_usd_mil` | [uint64](#uint64) |  | nhash_per_usd_mil is number of nhash per usd mil |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance-msgfees-v1-MsgUpdateNhashPerUsdMilProposalResponse"></a>

### MsgUpdateNhashPerUsdMilProposalResponse
MsgUpdateNhashPerUsdMilProposalResponse defines the Msg/UpdateNhashPerUsdMilProposal response type





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-msgfees-v1-Msg"></a>

### Msg
Msg defines the msgfees Msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `AssessCustomMsgFee` | [MsgAssessCustomMsgFeeRequest](#provenance-msgfees-v1-MsgAssessCustomMsgFeeRequest) | [MsgAssessCustomMsgFeeResponse](#provenance-msgfees-v1-MsgAssessCustomMsgFeeResponse) | AssessCustomMsgFee endpoint executes the additional fee charges. This will only emit the event and not persist it to the keeper. Fees are handled with the custom msg fee handlers Use Case: smart contracts will be able to charge additional fees and direct partial funds to specified recipient for executing contracts |
| `AddMsgFeeProposal` | [MsgAddMsgFeeProposalRequest](#provenance-msgfees-v1-MsgAddMsgFeeProposalRequest) | [MsgAddMsgFeeProposalResponse](#provenance-msgfees-v1-MsgAddMsgFeeProposalResponse) | AddMsgFeeProposal defines a governance proposal to add additional msg based fee |
| `UpdateMsgFeeProposal` | [MsgUpdateMsgFeeProposalRequest](#provenance-msgfees-v1-MsgUpdateMsgFeeProposalRequest) | [MsgUpdateMsgFeeProposalResponse](#provenance-msgfees-v1-MsgUpdateMsgFeeProposalResponse) | UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee |
| `RemoveMsgFeeProposal` | [MsgRemoveMsgFeeProposalRequest](#provenance-msgfees-v1-MsgRemoveMsgFeeProposalRequest) | [MsgRemoveMsgFeeProposalResponse](#provenance-msgfees-v1-MsgRemoveMsgFeeProposalResponse) | RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee |
| `UpdateNhashPerUsdMilProposal` | [MsgUpdateNhashPerUsdMilProposalRequest](#provenance-msgfees-v1-MsgUpdateNhashPerUsdMilProposalRequest) | [MsgUpdateNhashPerUsdMilProposalResponse](#provenance-msgfees-v1-MsgUpdateNhashPerUsdMilProposalResponse) | UpdateNhashPerUsdMilProposal defines a governance proposal to update the nhash per usd mil param |
| `UpdateConversionFeeDenomProposal` | [MsgUpdateConversionFeeDenomProposalRequest](#provenance-msgfees-v1-MsgUpdateConversionFeeDenomProposalRequest) | [MsgUpdateConversionFeeDenomProposalResponse](#provenance-msgfees-v1-MsgUpdateConversionFeeDenomProposalResponse) | UpdateConversionFeeDenomProposal defines a governance proposal to update the msg fee conversion denom |

 <!-- end services -->



<a name="provenance_msgfees_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/query.proto



<a name="provenance-msgfees-v1-CalculateTxFeesRequest"></a>

### CalculateTxFeesRequest
CalculateTxFeesRequest is the request type for the Query RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_bytes` | [bytes](#bytes) |  | tx_bytes is the transaction to simulate. |
| `default_base_denom` | [string](#string) |  | default_base_denom is used to set the denom used for gas fees if not set it will default to nhash. |
| `gas_adjustment` | [float](#float) |  | gas_adjustment is the adjustment factor to be multiplied against the estimate returned by the tx simulation |






<a name="provenance-msgfees-v1-CalculateTxFeesResponse"></a>

### CalculateTxFeesResponse
CalculateTxFeesResponse is the response type for the Query RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `additional_fees` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | additional_fees are the amount of coins to be for addition msg fees |
| `total_fees` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | total_fees are the total amount of fees needed for the transactions (msg fees + gas fee) note: the gas fee is calculated with the floor gas price module param. |
| `estimated_gas` | [uint64](#uint64) |  | estimated_gas is the amount of gas needed for the transaction |






<a name="provenance-msgfees-v1-QueryAllMsgFeesRequest"></a>

### QueryAllMsgFeesRequest
QueryAllMsgFeesRequest queries all Msg which have fees associated with them.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-msgfees-v1-QueryAllMsgFeesResponse"></a>

### QueryAllMsgFeesResponse
response for querying all msg's with fees associated with them


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_fees` | [MsgFee](#provenance-msgfees-v1-MsgFee) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance-msgfees-v1-QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance-msgfees-v1-QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-msgfees-v1-Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-msgfees-v1-Query"></a>

### Query
Query defines the gRPC querier service for marker module.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Params` | [QueryParamsRequest](#provenance-msgfees-v1-QueryParamsRequest) | [QueryParamsResponse](#provenance-msgfees-v1-QueryParamsResponse) | Params queries the parameters for x/msgfees |
| `QueryAllMsgFees` | [QueryAllMsgFeesRequest](#provenance-msgfees-v1-QueryAllMsgFeesRequest) | [QueryAllMsgFeesResponse](#provenance-msgfees-v1-QueryAllMsgFeesResponse) | Query all Msgs which have fees associated with them. |
| `CalculateTxFees` | [CalculateTxFeesRequest](#provenance-msgfees-v1-CalculateTxFeesRequest) | [CalculateTxFeesResponse](#provenance-msgfees-v1-CalculateTxFeesResponse) | CalculateTxFees simulates executing a transaction for estimating gas usage and additional fees. |

 <!-- end services -->



<a name="provenance_msgfees_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/genesis.proto



<a name="provenance-msgfees-v1-GenesisState"></a>

### GenesisState
GenesisState contains a set of msg fees, persisted from the store


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-msgfees-v1-Params) |  | params defines all the parameters of the module. |
| `msg_fees` | [MsgFee](#provenance-msgfees-v1-MsgFee) | repeated | msg_based_fees are the additional fees on specific tx msgs |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_msgfees_v1_msgfees-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/msgfees.proto



<a name="provenance-msgfees-v1-EventMsgFee"></a>

### EventMsgFee
EventMsgFee final event property for msg fee on type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type` | [string](#string) |  |  |
| `count` | [string](#string) |  |  |
| `total` | [string](#string) |  |  |
| `recipient` | [string](#string) |  |  |






<a name="provenance-msgfees-v1-EventMsgFees"></a>

### EventMsgFees
EventMsgFees event emitted with summary of msg fees


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_fees` | [EventMsgFee](#provenance-msgfees-v1-EventMsgFee) | repeated |  |






<a name="provenance-msgfees-v1-MsgFee"></a>

### MsgFee
MsgFee is the core of what gets stored on the blockchain to define a msg-based fee.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type_url` | [string](#string) |  | msg_type_url is the type-url of the message with the added fee, e.g. "/cosmos.bank.v1beta1.MsgSend". |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | additional_fee is the extra fee that is required for the given message type (can be in any denom). |
| `recipient` | [string](#string) |  | recipient is an option address that will receive a portion of the additional fee. There can only be a recipient if the recipient_basis_points is not zero. |
| `recipient_basis_points` | [uint32](#uint32) |  | recipient_basis_points is an optional portion of the additional fee to be sent to the recipient. Must be between 0 and 10,000 (inclusive).<br>If there is a recipient, this must not be zero. If there is not a recipient, this must be zero.<br>The recipient will receive additional_fee * recipient_basis_points / 10,000. The fee collector will receive the rest, i.e. additional_fee * (10,000 - recipient_basis_points) / 10,000. |






<a name="provenance-msgfees-v1-Params"></a>

### Params
Params defines the set of params for the msgfees module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `floor_gas_price` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | floor_gas_price is the constant used to calculate fees when gas fees shares denom with msg fee.<br>Conversions: - x nhash/usd-mil = 1,000,000/x usd/hash - y usd/hash = 1,000,000/y nhash/usd-mil<br>Examples: - 40,000,000 nhash/usd-mil = 1,000,000/40,000,000 usd/hash = $0.025/hash, - $0.040/hash = 1,000,000/0.040 nhash/usd-mil = 25,000,000 nhash/usd-mil |
| `nhash_per_usd_mil` | [uint64](#uint64) |  | nhash_per_usd_mil is the total nhash per usd mil for converting usd to nhash. |
| `conversion_fee_denom` | [string](#string) |  | conversion_fee_denom is the denom usd is converted to. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_msgfees_v1_proposals-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/proposals.proto



<a name="provenance-msgfees-v1-AddMsgFeeProposal"></a>

### AddMsgFeeProposal
AddMsgFeeProposal defines a governance proposal to add additional msg based fee
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgAddMsgFeeProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | propsal title |
| `description` | [string](#string) |  | propsal description |
| `msg_type_url` | [string](#string) |  | type url of msg to add fee |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | additional fee for msg type |
| `recipient` | [string](#string) |  | optional recipient to recieve basis points |
| `recipient_basis_points` | [string](#string) |  | basis points to use when recipient is present (1 - 10,000) |






<a name="provenance-msgfees-v1-RemoveMsgFeeProposal"></a>

### RemoveMsgFeeProposal
RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgRemoveMsgFeeProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | propsal title |
| `description` | [string](#string) |  | propsal description |
| `msg_type_url` | [string](#string) |  | type url of msg fee to remove |






<a name="provenance-msgfees-v1-UpdateConversionFeeDenomProposal"></a>

### UpdateConversionFeeDenomProposal
UpdateConversionFeeDenomProposal defines a governance proposal to update the msg fee conversion denom
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgUpdateConversionFeeDenomProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | proposal title |
| `description` | [string](#string) |  | proposal description |
| `conversion_fee_denom` | [string](#string) |  | conversion_fee_denom is the denom that usd will be converted to |






<a name="provenance-msgfees-v1-UpdateMsgFeeProposal"></a>

### UpdateMsgFeeProposal
UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgUpdateMsgFeeProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | propsal title |
| `description` | [string](#string) |  | propsal description |
| `msg_type_url` | [string](#string) |  | type url of msg to update fee |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | additional fee for msg type |
| `recipient` | [string](#string) |  | optional recipient to recieve basis points |
| `recipient_basis_points` | [string](#string) |  | basis points to use when recipient is present (1 - 10,000) |






<a name="provenance-msgfees-v1-UpdateNhashPerUsdMilProposal"></a>

### UpdateNhashPerUsdMilProposal
UpdateNhashPerUsdMilProposal defines a governance proposal to update the nhash per usd mil param
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgUpdateNhashPerUsdMilProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | proposal title |
| `description` | [string](#string) |  | proposal description |
| `nhash_per_usd_mil` | [uint64](#uint64) |  | nhash_per_usd_mil is number of nhash per usd mil |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_oracle_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/oracle/v1/tx.proto



<a name="provenance-oracle-v1-MsgSendQueryOracleRequest"></a>

### MsgSendQueryOracleRequest
MsgSendQueryOracleRequest queries an oracle on another chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `query` | [bytes](#bytes) |  | Query contains the query data passed to the oracle. |
| `channel` | [string](#string) |  | Channel is the channel to the oracle. |
| `authority` | [string](#string) |  | The signing authority for the request |






<a name="provenance-oracle-v1-MsgSendQueryOracleResponse"></a>

### MsgSendQueryOracleResponse
MsgSendQueryOracleResponse contains the id of the oracle query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sequence` | [uint64](#uint64) |  | The sequence number that uniquely identifies the query. |






<a name="provenance-oracle-v1-MsgUpdateOracleRequest"></a>

### MsgUpdateOracleRequest
MsgUpdateOracleRequest is the request type for updating an oracle's contract address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | The address of the oracle's contract |
| `authority` | [string](#string) |  | The signing authorities for the request |






<a name="provenance-oracle-v1-MsgUpdateOracleResponse"></a>

### MsgUpdateOracleResponse
MsgUpdateOracleResponse is the response type for updating the oracle.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-oracle-v1-Msg"></a>

### Msg
Msg

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `UpdateOracle` | [MsgUpdateOracleRequest](#provenance-oracle-v1-MsgUpdateOracleRequest) | [MsgUpdateOracleResponse](#provenance-oracle-v1-MsgUpdateOracleResponse) | UpdateOracle is the RPC endpoint for updating the oracle |
| `SendQueryOracle` | [MsgSendQueryOracleRequest](#provenance-oracle-v1-MsgSendQueryOracleRequest) | [MsgSendQueryOracleResponse](#provenance-oracle-v1-MsgSendQueryOracleResponse) | SendQueryOracle sends a query to an oracle on another chain |

 <!-- end services -->



<a name="provenance_oracle_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/oracle/v1/query.proto



<a name="provenance-oracle-v1-QueryOracleAddressRequest"></a>

### QueryOracleAddressRequest
QueryOracleAddressRequest queries for the address of the oracle.






<a name="provenance-oracle-v1-QueryOracleAddressResponse"></a>

### QueryOracleAddressResponse
QueryOracleAddressResponse contains the address of the oracle.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | The address of the oracle |






<a name="provenance-oracle-v1-QueryOracleRequest"></a>

### QueryOracleRequest
QueryOracleRequest queries the module's oracle.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `query` | [bytes](#bytes) |  | Query contains the query data passed to the oracle. |






<a name="provenance-oracle-v1-QueryOracleResponse"></a>

### QueryOracleResponse
QueryOracleResponse contains the result of the query sent to the oracle.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `data` | [bytes](#bytes) |  | Data contains the json data returned from the oracle. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-oracle-v1-Query"></a>

### Query
Query defines the gRPC querier service for oracle module.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `OracleAddress` | [QueryOracleAddressRequest](#provenance-oracle-v1-QueryOracleAddressRequest) | [QueryOracleAddressResponse](#provenance-oracle-v1-QueryOracleAddressResponse) | OracleAddress returns the address of the oracle |
| `Oracle` | [QueryOracleRequest](#provenance-oracle-v1-QueryOracleRequest) | [QueryOracleResponse](#provenance-oracle-v1-QueryOracleResponse) | Oracle forwards a query to the module's oracle |

 <!-- end services -->



<a name="provenance_oracle_v1_event-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/oracle/v1/event.proto



<a name="provenance-oracle-v1-EventOracleQueryError"></a>

### EventOracleQueryError
EventOracleQueryError is an event for when the chain receives an error response from an oracle query


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channel` | [string](#string) |  | channel is the local channel that the oracle query response was received from |
| `sequence_id` | [string](#string) |  | sequence_id is a unique identifier of the query |
| `error` | [string](#string) |  | error is the error message received from the query |






<a name="provenance-oracle-v1-EventOracleQuerySuccess"></a>

### EventOracleQuerySuccess
EventOracleQuerySuccess is an event for when the chain receives a successful response from an oracle query


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channel` | [string](#string) |  | channel is the local channel that the oracle query response was received from |
| `sequence_id` | [string](#string) |  | sequence_id is a unique identifier of the query |
| `result` | [string](#string) |  | result is the data received from the query |






<a name="provenance-oracle-v1-EventOracleQueryTimeout"></a>

### EventOracleQueryTimeout
EventOracleQueryTimeout is an event for when the chain receives a timeout from an oracle query


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `channel` | [string](#string) |  | channel is the local channel that the oracle timeout was received from |
| `sequence_id` | [string](#string) |  | sequence_id is a unique identifier of the query |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_oracle_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/oracle/v1/genesis.proto



<a name="provenance-oracle-v1-GenesisState"></a>

### GenesisState
GenesisState defines the oracle module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `port_id` | [string](#string) |  | The port to assign to the module |
| `oracle` | [string](#string) |  | The address of the oracle |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_registry_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/registry/v1/tx.proto



<a name="provenance-registry-v1-MsgGrantRole"></a>

### MsgGrantRole
MsgGrantRole represents a message to grant a role to an address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority is the address that is authorized to grant the role |
| `key` | [RegistryKey](#provenance-registry-v1-RegistryKey) |  | key is the key to grant the role to |
| `role` | [RegistryRole](#provenance-registry-v1-RegistryRole) |  | role is the role to grant |
| `addresses` | [string](#string) | repeated | addresses is the list of addresses to grant the role to |






<a name="provenance-registry-v1-MsgGrantRoleResponse"></a>

### MsgGrantRoleResponse
MsgGrantRoleResponse defines the response for GrantRole






<a name="provenance-registry-v1-MsgRegisterNFT"></a>

### MsgRegisterNFT
MsgRegisterNFT represents a message to register a new NFT


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority is the address that is authorized to register addresses |
| `key` | [RegistryKey](#provenance-registry-v1-RegistryKey) |  | key is the key to register |
| `roles` | [MsgRegisterNFT.RolesEntry](#provenance-registry-v1-MsgRegisterNFT-RolesEntry) | repeated | roles is a map of role names to lists of addresses that can perform that role |






<a name="provenance-registry-v1-MsgRegisterNFT-RolesEntry"></a>

### MsgRegisterNFT.RolesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [RoleAddresses](#provenance-registry-v1-RoleAddresses) |  |  |






<a name="provenance-registry-v1-MsgRegisterNFTResponse"></a>

### MsgRegisterNFTResponse
MsgRegisterNFTResponse defines the response for RegisterNFT






<a name="provenance-registry-v1-MsgRevokeRole"></a>

### MsgRevokeRole
MsgRevokeRole represents a message to revoke a role from an address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority is the address that is authorized to revoke the role |
| `key` | [RegistryKey](#provenance-registry-v1-RegistryKey) |  | key is the key to revoke the role from |
| `role` | [RegistryRole](#provenance-registry-v1-RegistryRole) |  | role is the role to revoke |
| `addresses` | [string](#string) | repeated | addresses is the list of addresses to revoke the role from |






<a name="provenance-registry-v1-MsgRevokeRoleResponse"></a>

### MsgRevokeRoleResponse
MsgRevokeRoleResponse defines the response for RevokeRole






<a name="provenance-registry-v1-MsgUnregisterNFT"></a>

### MsgUnregisterNFT
MsgUnregisterNFT represents a message to unregister an NFT from the registry


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority is the address that is authorized to remove addresses |
| `key` | [RegistryKey](#provenance-registry-v1-RegistryKey) |  | key is the key to remove |






<a name="provenance-registry-v1-MsgUnregisterNFTResponse"></a>

### MsgUnregisterNFTResponse
MsgUnregisterNFTResponse defines the response for UnregisterNFT





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-registry-v1-Msg"></a>

### Msg
Msg defines the registry Msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `RegisterNFT` | [MsgRegisterNFT](#provenance-registry-v1-MsgRegisterNFT) | [MsgRegisterNFTResponse](#provenance-registry-v1-MsgRegisterNFTResponse) | RegisterNFT registers a new NFT |
| `GrantRole` | [MsgGrantRole](#provenance-registry-v1-MsgGrantRole) | [MsgGrantRoleResponse](#provenance-registry-v1-MsgGrantRoleResponse) | GrantRole grants a role to an address |
| `RevokeRole` | [MsgRevokeRole](#provenance-registry-v1-MsgRevokeRole) | [MsgRevokeRoleResponse](#provenance-registry-v1-MsgRevokeRoleResponse) | RevokeRole revokes a role from an address |
| `UnregisterNFT` | [MsgUnregisterNFT](#provenance-registry-v1-MsgUnregisterNFT) | [MsgUnregisterNFTResponse](#provenance-registry-v1-MsgUnregisterNFTResponse) | UnregisterNFT unregisters an NFT from the registry |

 <!-- end services -->



<a name="provenance_registry_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/registry/v1/query.proto



<a name="provenance-registry-v1-QueryGetRegistryRequest"></a>

### QueryGetRegistryRequest
QueryGetRegistryRequest is the request type for the Query/GetRegistry RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [RegistryKey](#provenance-registry-v1-RegistryKey) |  | key is the key to query |






<a name="provenance-registry-v1-QueryGetRegistryResponse"></a>

### QueryGetRegistryResponse
QueryGetRegistryResponse is the response type for the Query/GetRegistry RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `registry` | [RegistryEntry](#provenance-registry-v1-RegistryEntry) |  | entry is the registry entry for the requested key |






<a name="provenance-registry-v1-QueryHasRoleRequest"></a>

### QueryHasRoleRequest
QueryHasRoleRequest is the request type for the Query/HasRole RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [RegistryKey](#provenance-registry-v1-RegistryKey) |  | key is the key to query |
| `address` | [string](#string) |  | address is the address to query |
| `role` | [RegistryRole](#provenance-registry-v1-RegistryRole) |  | role is the role to query |






<a name="provenance-registry-v1-QueryHasRoleResponse"></a>

### QueryHasRoleResponse
QueryHasRoleResponse is the response type for the Query/HasRole RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `has_role` | [bool](#bool) |  | has_role is true if the address has the role for the given key |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-registry-v1-Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetRegistry` | [QueryGetRegistryRequest](#provenance-registry-v1-QueryGetRegistryRequest) | [QueryGetRegistryResponse](#provenance-registry-v1-QueryGetRegistryResponse) | GetRegistry returns the registry for a given key |
| `HasRole` | [QueryHasRoleRequest](#provenance-registry-v1-QueryHasRoleRequest) | [QueryHasRoleResponse](#provenance-registry-v1-QueryHasRoleResponse) | HasRole returns true if the address has the role for the given key |

 <!-- end services -->



<a name="provenance_registry_v1_registry-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/registry/v1/registry.proto



<a name="provenance-registry-v1-GenesisState"></a>

### GenesisState
GenesisState defines the registry module's genesis state


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entries` | [RegistryEntry](#provenance-registry-v1-RegistryEntry) | repeated | entries is the list of registry entries |






<a name="provenance-registry-v1-RegistryEntry"></a>

### RegistryEntry
RegistryEntry represents a single entry in the registry, mapping a blockchain address to its roles


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [RegistryKey](#provenance-registry-v1-RegistryKey) |  | Key ties the registry entry to an asset class and nft id |
| `roles` | [RegistryEntry.RolesEntry](#provenance-registry-v1-RegistryEntry-RolesEntry) | repeated | roles is a map of role names to lists of addresses that can perform that role |






<a name="provenance-registry-v1-RegistryEntry-RolesEntry"></a>

### RegistryEntry.RolesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `key` | [string](#string) |  |  |
| `value` | [RoleAddresses](#provenance-registry-v1-RoleAddresses) |  |  |






<a name="provenance-registry-v1-RegistryKey"></a>

### RegistryKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `nft_id` | [string](#string) |  | Identifier for the nft that this ledger is linked to. This could be a `x/metadata` scope id or an `x/nft` nft id. In order to create a ledger for an nft, the nft class must be registered in the ledger module as a LedgerClass. |
| `asset_class_id` | [string](#string) |  | Scope Specification ID or NFT Class ID |






<a name="provenance-registry-v1-RoleAddresses"></a>

### RoleAddresses
RoleAddresses contains a list of addresses that can perform a specific role


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) | repeated | addresses is the list of blockchain addresses that can perform this role |





 <!-- end messages -->


<a name="provenance-registry-v1-RegistryRole"></a>

### RegistryRole
Role defines the different types of roles that can be assigned to addresses

| Name | Number | Description |
| ---- | ------ | ----------- |
| `REGISTRY_ROLE_UNSPECIFIED` | `0` | REGISTRY_ROLE_UNSPECIFIED indicates no role is assigned |
| `REGISTRY_ROLE_SERVICER` | `1` | REGISTRY_ROLE_SERVICER indicates the address has servicer privileges |
| `REGISTRY_ROLE_SUBSERVICER` | `2` | REGISTRY_ROLE_SUBSERVICER indicates the address has subservicer privileges |
| `REGISTRY_ROLE_CONTROLLER` | `3` | REGISTRY_ROLE_CONTROLLER indicates the address has controller privileges |
| `REGISTRY_ROLE_CUSTODIAN` | `4` | REGISTRY_ROLE_CUSTODIAN indicates the address has custodian privileges |
| `REGISTRY_ROLE_BORROWER` | `5` | REGISTRY_ROLE_BORROWER indicates the address has borrower privileges |
| `REGISTRY_ROLE_ORIGINATOR` | `6` | REGISTRY_ROLE_ORIGINATOR indicates the address has originator privileges |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ibchooks_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibchooks/v1/tx.proto



<a name="provenance-ibchooks-v1-MsgEmitIBCAck"></a>

### MsgEmitIBCAck
MsgEmitIBCAck is the IBC Acknowledgement


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `packet_sequence` | [uint64](#uint64) |  |  |
| `channel` | [string](#string) |  |  |






<a name="provenance-ibchooks-v1-MsgEmitIBCAckResponse"></a>

### MsgEmitIBCAckResponse
MsgEmitIBCAckResponse is the IBC Acknowledgement response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_result` | [string](#string) |  |  |
| `ibc_ack` | [string](#string) |  |  |






<a name="provenance-ibchooks-v1-MsgUpdateParamsRequest"></a>

### MsgUpdateParamsRequest
MsgUpdateParamsRequest is a request message for the UpdateParams endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `params` | [Params](#provenance-ibchooks-v1-Params) |  | params are the new param values to set. |






<a name="provenance-ibchooks-v1-MsgUpdateParamsResponse"></a>

### MsgUpdateParamsResponse
MsgUpdateParamsResponse is a response message for the UpdateParams endpoint.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-ibchooks-v1-Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `EmitIBCAck` | [MsgEmitIBCAck](#provenance-ibchooks-v1-MsgEmitIBCAck) | [MsgEmitIBCAckResponse](#provenance-ibchooks-v1-MsgEmitIBCAckResponse) | EmitIBCAck checks the sender can emit the ack and writes the IBC acknowledgement |
| `UpdateParams` | [MsgUpdateParamsRequest](#provenance-ibchooks-v1-MsgUpdateParamsRequest) | [MsgUpdateParamsResponse](#provenance-ibchooks-v1-MsgUpdateParamsResponse) | UpdateParams is a governance proposal endpoint for updating the ibchooks module's params. |

 <!-- end services -->



<a name="provenance_ibchooks_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibchooks/v1/query.proto



<a name="provenance-ibchooks-v1-QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance-ibchooks-v1-QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-ibchooks-v1-Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-ibchooks-v1-Query"></a>

### Query
Query defines the gRPC querier service for attribute module.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Params` | [QueryParamsRequest](#provenance-ibchooks-v1-QueryParamsRequest) | [QueryParamsResponse](#provenance-ibchooks-v1-QueryParamsResponse) | Params queries params of the ihchooks module. |

 <!-- end services -->



<a name="provenance_ibchooks_v1_event-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibchooks/v1/event.proto



<a name="provenance-ibchooks-v1-EventIBCHooksParamsUpdated"></a>

### EventIBCHooksParamsUpdated
EventIBCHooksParamsUpdated defines the event emitted after updating ibchooks parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `allowed_async_ack_contracts` | [string](#string) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ibchooks_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibchooks/v1/genesis.proto



<a name="provenance-ibchooks-v1-GenesisState"></a>

### GenesisState
GenesisState is the IBC Hooks genesis state (params)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-ibchooks-v1-Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ibchooks_v1_params-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibchooks/v1/params.proto



<a name="provenance-ibchooks-v1-Params"></a>

### Params
Params defines the allowed async ack contracts


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `allowed_async_ack_contracts` | [string](#string) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ibcratelimit_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibcratelimit/v1/tx.proto



<a name="provenance-ibcratelimit-v1-MsgGovUpdateParamsRequest"></a>

### MsgGovUpdateParamsRequest
MsgGovUpdateParamsRequest is a request message for the GovUpdateParams endpoint.
Deprecated: Use MsgUpdateParamsRequest instead.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `params` | [Params](#provenance-ibcratelimit-v1-Params) |  | params are the new param values to set. |






<a name="provenance-ibcratelimit-v1-MsgGovUpdateParamsResponse"></a>

### MsgGovUpdateParamsResponse
MsgGovUpdateParamsResponse is a response message for the GovUpdateParams endpoint.
Deprecated: Use MsgUpdateParamsResponse instead.






<a name="provenance-ibcratelimit-v1-MsgUpdateParamsRequest"></a>

### MsgUpdateParamsRequest
MsgUpdateParamsRequest is a request message for the UpdateParams endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `params` | [Params](#provenance-ibcratelimit-v1-Params) |  | params are the new param values to set. |






<a name="provenance-ibcratelimit-v1-MsgUpdateParamsResponse"></a>

### MsgUpdateParamsResponse
MsgUpdateParamsResponse is a response message for the UpdateParams endpoint.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-ibcratelimit-v1-Msg"></a>

### Msg
Msg is the service for ibcratelimit module's tx endpoints.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GovUpdateParams` | [MsgGovUpdateParamsRequest](#provenance-ibcratelimit-v1-MsgGovUpdateParamsRequest) | [MsgGovUpdateParamsResponse](#provenance-ibcratelimit-v1-MsgGovUpdateParamsResponse) | GovUpdateParams is a governance proposal endpoint for updating the exchange module's params. Deprecated: Use UpdateParams instead. |
| `UpdateParams` | [MsgUpdateParamsRequest](#provenance-ibcratelimit-v1-MsgUpdateParamsRequest) | [MsgUpdateParamsResponse](#provenance-ibcratelimit-v1-MsgUpdateParamsResponse) | UpdateParams is a governance proposal endpoint for updating the ibcratelimit module's params. |

 <!-- end services -->



<a name="provenance_ibcratelimit_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibcratelimit/v1/query.proto



<a name="provenance-ibcratelimit-v1-ParamsRequest"></a>

### ParamsRequest
ParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance-ibcratelimit-v1-ParamsResponse"></a>

### ParamsResponse
ParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-ibcratelimit-v1-Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-ibcratelimit-v1-Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Params` | [ParamsRequest](#provenance-ibcratelimit-v1-ParamsRequest) | [ParamsResponse](#provenance-ibcratelimit-v1-ParamsResponse) | Params defines a gRPC query method that returns the ibcratelimit module's parameters. |

 <!-- end services -->



<a name="provenance_ibcratelimit_v1_event-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibcratelimit/v1/event.proto



<a name="provenance-ibcratelimit-v1-EventAckRevertFailure"></a>

### EventAckRevertFailure
EventAckRevertFailure is emitted when an Ack revert fails


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  | module is the name of the module that emitted it. |
| `packet` | [string](#string) |  | packet is the packet received on acknowledgement. |
| `ack` | [string](#string) |  | ack is the packet's inner acknowledgement message. |






<a name="provenance-ibcratelimit-v1-EventParamsUpdated"></a>

### EventParamsUpdated
EventParamsUpdated is an event emitted when the ibcratelimit module's params have been updated.






<a name="provenance-ibcratelimit-v1-EventTimeoutRevertFailure"></a>

### EventTimeoutRevertFailure
EventTimeoutRevertFailure is emitted when a Timeout revert fails


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  | module is the name of the module that emitted it. |
| `packet` | [string](#string) |  | packet is the packet received on timeout. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ibcratelimit_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibcratelimit/v1/genesis.proto



<a name="provenance-ibcratelimit-v1-GenesisState"></a>

### GenesisState
GenesisState defines the ibcratelimit module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-ibcratelimit-v1-Params) |  | params are all the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_ibcratelimit_v1_params-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibcratelimit/v1/params.proto



<a name="provenance-ibcratelimit-v1-Params"></a>

### Params
Params defines the parameters for the ibcratelimit module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract_address is the address of the rate limiter contract. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_marker_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/tx.proto



<a name="provenance-marker-v1-MsgActivateRequest"></a>

### MsgActivateRequest
MsgActivateRequest defines the Msg/Activate request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgActivateResponse"></a>

### MsgActivateResponse
MsgActivateResponse defines the Msg/Activate response type






<a name="provenance-marker-v1-MsgAddAccessRequest"></a>

### MsgAddAccessRequest
MsgAddAccessRequest defines the Msg/AddAccess request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `access` | [AccessGrant](#provenance-marker-v1-AccessGrant) | repeated |  |






<a name="provenance-marker-v1-MsgAddAccessResponse"></a>

### MsgAddAccessResponse
MsgAddAccessResponse defines the Msg/AddAccess response type






<a name="provenance-marker-v1-MsgAddFinalizeActivateMarkerRequest"></a>

### MsgAddFinalizeActivateMarkerRequest
MsgAddFinalizeActivateMarkerRequest defines the Msg/AddFinalizeActivateMarker request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |
| `manager` | [string](#string) |  |  |
| `from_address` | [string](#string) |  |  |
| `marker_type` | [MarkerType](#provenance-marker-v1-MarkerType) |  |  |
| `access_list` | [AccessGrant](#provenance-marker-v1-AccessGrant) | repeated |  |
| `supply_fixed` | [bool](#bool) |  |  |
| `allow_governance_control` | [bool](#bool) |  |  |
| `allow_forced_transfer` | [bool](#bool) |  |  |
| `required_attributes` | [string](#string) | repeated |  |
| `usd_cents` | [uint64](#uint64) |  | **Deprecated.**  |
| `volume` | [uint64](#uint64) |  |  |
| `usd_mills` | [uint64](#uint64) |  |  |






<a name="provenance-marker-v1-MsgAddFinalizeActivateMarkerResponse"></a>

### MsgAddFinalizeActivateMarkerResponse
MsgAddFinalizeActivateMarkerResponse defines the Msg/AddFinalizeActivateMarker response type






<a name="provenance-marker-v1-MsgAddMarkerRequest"></a>

### MsgAddMarkerRequest
MsgAddMarkerRequest defines the Msg/AddMarker request type.
If being provided as a governance proposal, set the from_address to the gov module's account address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |
| `manager` | [string](#string) |  |  |
| `from_address` | [string](#string) |  |  |
| `status` | [MarkerStatus](#provenance-marker-v1-MarkerStatus) |  |  |
| `marker_type` | [MarkerType](#provenance-marker-v1-MarkerType) |  |  |
| `access_list` | [AccessGrant](#provenance-marker-v1-AccessGrant) | repeated |  |
| `supply_fixed` | [bool](#bool) |  |  |
| `allow_governance_control` | [bool](#bool) |  |  |
| `allow_forced_transfer` | [bool](#bool) |  |  |
| `required_attributes` | [string](#string) | repeated |  |
| `usd_cents` | [uint64](#uint64) |  | **Deprecated.**  |
| `volume` | [uint64](#uint64) |  |  |
| `usd_mills` | [uint64](#uint64) |  |  |






<a name="provenance-marker-v1-MsgAddMarkerResponse"></a>

### MsgAddMarkerResponse
MsgAddMarkerResponse defines the Msg/AddMarker response type






<a name="provenance-marker-v1-MsgAddNetAssetValuesRequest"></a>

### MsgAddNetAssetValuesRequest
MsgAddNetAssetValuesRequest defines the Msg/AddNetAssetValues request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `net_asset_values` | [NetAssetValue](#provenance-marker-v1-NetAssetValue) | repeated |  |






<a name="provenance-marker-v1-MsgAddNetAssetValuesResponse"></a>

### MsgAddNetAssetValuesResponse
MsgAddNetAssetValuesResponse defines the Msg/AddNetAssetValue response type






<a name="provenance-marker-v1-MsgBurnRequest"></a>

### MsgBurnRequest
MsgBurnRequest defines the Msg/Burn request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgBurnResponse"></a>

### MsgBurnResponse
MsgBurnResponse defines the Msg/Burn response type






<a name="provenance-marker-v1-MsgCancelRequest"></a>

### MsgCancelRequest
MsgCancelRequest defines the Msg/Cancel request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgCancelResponse"></a>

### MsgCancelResponse
MsgCancelResponse defines the Msg/Cancel response type






<a name="provenance-marker-v1-MsgChangeStatusProposalRequest"></a>

### MsgChangeStatusProposalRequest
MsgChangeStatusProposalRequest defines the Msg/ChangeStatusProposal request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `new_status` | [MarkerStatus](#provenance-marker-v1-MarkerStatus) |  |  |
| `authority` | [string](#string) |  | The signer of the message. Must have admin authority to marker or be governance module account address. |






<a name="provenance-marker-v1-MsgChangeStatusProposalResponse"></a>

### MsgChangeStatusProposalResponse
MsgChangeStatusProposalResponse defines the Msg/ChangeStatusProposal response type






<a name="provenance-marker-v1-MsgDeleteAccessRequest"></a>

### MsgDeleteAccessRequest
MsgDeleteAccessRequest defines the Msg/DeleteAccess request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `removed_address` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgDeleteAccessResponse"></a>

### MsgDeleteAccessResponse
MsgDeleteAccessResponse defines the Msg/DeleteAccess response type






<a name="provenance-marker-v1-MsgDeleteRequest"></a>

### MsgDeleteRequest
MsgDeleteRequest defines the Msg/Delete request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgDeleteResponse"></a>

### MsgDeleteResponse
MsgDeleteResponse defines the Msg/Delete response type






<a name="provenance-marker-v1-MsgFinalizeRequest"></a>

### MsgFinalizeRequest
MsgFinalizeRequest defines the Msg/Finalize request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgFinalizeResponse"></a>

### MsgFinalizeResponse
MsgFinalizeResponse defines the Msg/Finalize response type






<a name="provenance-marker-v1-MsgGrantAllowanceRequest"></a>

### MsgGrantAllowanceRequest
MsgGrantAllowanceRequest validates permission to create a fee grant based on marker admin access. If
successful a feegrant is recorded where the marker account itself is the grantor


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `grantee` | [string](#string) |  | grantee is the address of the user being granted an allowance of another user's funds. |
| `allowance` | [google.protobuf.Any](#google-protobuf-Any) |  | allowance can be any of basic and filtered fee allowance (fee FeeGrant module). |






<a name="provenance-marker-v1-MsgGrantAllowanceResponse"></a>

### MsgGrantAllowanceResponse
MsgGrantAllowanceResponse defines the Msg/GrantAllowanceResponse response type.






<a name="provenance-marker-v1-MsgIbcTransferRequest"></a>

### MsgIbcTransferRequest
MsgIbcTransferRequest defines the Msg/IbcTransfer request type for markers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfer` | [ibc.applications.transfer.v1.MsgTransfer](#ibc-applications-transfer-v1-MsgTransfer) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgIbcTransferResponse"></a>

### MsgIbcTransferResponse
MsgIbcTransferResponse defines the Msg/IbcTransfer response type






<a name="provenance-marker-v1-MsgMintRequest"></a>

### MsgMintRequest
MsgMintRequest defines the Msg/Mint request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgMintResponse"></a>

### MsgMintResponse
MsgMintResponse defines the Msg/Mint response type






<a name="provenance-marker-v1-MsgRemoveAdministratorProposalRequest"></a>

### MsgRemoveAdministratorProposalRequest
MsgRemoveAdministratorProposalRequest defines the Msg/RemoveAdministratorProposal request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `removed_address` | [string](#string) | repeated |  |
| `authority` | [string](#string) |  | The signer of the message. Must have admin authority to marker or be governance module account address. |






<a name="provenance-marker-v1-MsgRemoveAdministratorProposalResponse"></a>

### MsgRemoveAdministratorProposalResponse
MsgRemoveAdministratorProposalResponse defines the Msg/RemoveAdministratorProposal response type






<a name="provenance-marker-v1-MsgSetAccountDataRequest"></a>

### MsgSetAccountDataRequest
MsgSetAccountDataRequest defines a msg to set/update/delete the account data for a marker.
Signer must have deposit authority or be a gov proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | The denomination of the marker to update. |
| `value` | [string](#string) |  | The desired accountdata value. |
| `signer` | [string](#string) |  | The signer of this message. Must have deposit authority or be the governance module account address. |






<a name="provenance-marker-v1-MsgSetAccountDataResponse"></a>

### MsgSetAccountDataResponse
MsgSetAccountDataResponse defines the Msg/SetAccountData response type






<a name="provenance-marker-v1-MsgSetAdministratorProposalRequest"></a>

### MsgSetAdministratorProposalRequest
MsgSetAdministratorProposalRequest defines the Msg/SetAdministratorProposal request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `access` | [AccessGrant](#provenance-marker-v1-AccessGrant) | repeated |  |
| `authority` | [string](#string) |  | The signer of the message. Must have admin authority to marker or be governance module account address. |






<a name="provenance-marker-v1-MsgSetAdministratorProposalResponse"></a>

### MsgSetAdministratorProposalResponse
MsgSetAdministratorProposalResponse defines the Msg/SetAdministratorProposal response type






<a name="provenance-marker-v1-MsgSetDenomMetadataProposalRequest"></a>

### MsgSetDenomMetadataProposalRequest
MsgSetDenomMetadataProposalRequest defines the Msg/SetDenomMetadataProposal request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata` | [cosmos.bank.v1beta1.Metadata](#cosmos-bank-v1beta1-Metadata) |  |  |
| `authority` | [string](#string) |  | The signer of the message. Must have admin authority to marker or be governance module account address. |






<a name="provenance-marker-v1-MsgSetDenomMetadataProposalResponse"></a>

### MsgSetDenomMetadataProposalResponse
MsgSetDenomMetadataProposalResponse defines the Msg/SetDenomMetadataProposal response type






<a name="provenance-marker-v1-MsgSetDenomMetadataRequest"></a>

### MsgSetDenomMetadataRequest
MsgSetDenomMetadataRequest defines the Msg/SetDenomMetadata request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata` | [cosmos.bank.v1beta1.Metadata](#cosmos-bank-v1beta1-Metadata) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgSetDenomMetadataResponse"></a>

### MsgSetDenomMetadataResponse
MsgSetDenomMetadataResponse defines the Msg/SetDenomMetadata response type






<a name="provenance-marker-v1-MsgSupplyDecreaseProposalRequest"></a>

### MsgSupplyDecreaseProposalRequest
MsgSupplyDecreaseProposalRequest defines a governance proposal to decrease total supply of the marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |
| `authority` | [string](#string) |  | signer of the proposal |






<a name="provenance-marker-v1-MsgSupplyDecreaseProposalResponse"></a>

### MsgSupplyDecreaseProposalResponse
MsgSupplyIncreaseProposalResponse defines the Msg/SupplyDecreaseProposal response type






<a name="provenance-marker-v1-MsgSupplyIncreaseProposalRequest"></a>

### MsgSupplyIncreaseProposalRequest
MsgSupplyIncreaseProposalRequest defines a governance proposal to administer a marker and increase total supply of
the marker through minting coin and placing it within the marker or assigning it directly to an account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |
| `target_address` | [string](#string) |  | an optional target address for the minted coin from this request |
| `authority` | [string](#string) |  | signer of the proposal |






<a name="provenance-marker-v1-MsgSupplyIncreaseProposalResponse"></a>

### MsgSupplyIncreaseProposalResponse
MsgSupplyIncreaseProposalResponse defines the Msg/SupplyIncreaseProposal response type






<a name="provenance-marker-v1-MsgTransferRequest"></a>

### MsgTransferRequest
MsgTransferRequest defines the Msg/Transfer request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |
| `administrator` | [string](#string) |  |  |
| `from_address` | [string](#string) |  |  |
| `to_address` | [string](#string) |  |  |






<a name="provenance-marker-v1-MsgTransferResponse"></a>

### MsgTransferResponse
MsgTransferResponse defines the Msg/Transfer response type






<a name="provenance-marker-v1-MsgUpdateForcedTransferRequest"></a>

### MsgUpdateForcedTransferRequest
MsgUpdateForcedTransferRequest defines a msg to update the allow_forced_transfer field of a marker.
It is only usable via governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | The denomination of the marker to update. |
| `allow_forced_transfer` | [bool](#bool) |  | Whether an admin can transfer restricted coins from a 3rd-party account without their signature. |
| `authority` | [string](#string) |  | The signer of this message. Must be the governance module account address. |






<a name="provenance-marker-v1-MsgUpdateForcedTransferResponse"></a>

### MsgUpdateForcedTransferResponse
MsgUpdateForcedTransferResponse defines the Msg/UpdateForcedTransfer response type






<a name="provenance-marker-v1-MsgUpdateParamsRequest"></a>

### MsgUpdateParamsRequest
MsgUpdateParamsRequest is a request message for the UpdateParams endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `params` | [Params](#provenance-marker-v1-Params) |  | params are the new param values to set. |






<a name="provenance-marker-v1-MsgUpdateParamsResponse"></a>

### MsgUpdateParamsResponse
MsgUpdateParamsResponse is a response message for the UpdateParams endpoint.






<a name="provenance-marker-v1-MsgUpdateRequiredAttributesRequest"></a>

### MsgUpdateRequiredAttributesRequest
MsgUpdateRequiredAttributesRequest defines a msg to update/add/remove required attributes from a resticted marker
signer must have transfer authority to change attributes, to update attribute add current to remove list and new to
add list


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | The denomination of the marker to update. |
| `remove_required_attributes` | [string](#string) | repeated | List of required attributes to remove from marker. |
| `add_required_attributes` | [string](#string) | repeated | List of required attributes to add to marker. |
| `transfer_authority` | [string](#string) |  | The signer of the message. Must have transfer authority to marker or be governance module account address. |






<a name="provenance-marker-v1-MsgUpdateRequiredAttributesResponse"></a>

### MsgUpdateRequiredAttributesResponse
MsgUpdateRequiredAttributesResponse defines the Msg/UpdateRequiredAttributes response type






<a name="provenance-marker-v1-MsgUpdateSendDenyListRequest"></a>

### MsgUpdateSendDenyListRequest
MsgUpdateSendDenyListRequest defines a msg to add/remove addresses to send deny list for a resticted marker
signer must have transfer authority


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | The denomination of the marker to update. |
| `remove_denied_addresses` | [string](#string) | repeated | List of bech32 addresses to remove from the deny send list. |
| `add_denied_addresses` | [string](#string) | repeated | List of bech32 addresses to add to the deny send list. |
| `authority` | [string](#string) |  | The signer of the message. Must have admin authority to marker or be governance module account address. |






<a name="provenance-marker-v1-MsgUpdateSendDenyListResponse"></a>

### MsgUpdateSendDenyListResponse
MsgUpdateSendDenyListResponse defines the Msg/UpdateSendDenyList response type






<a name="provenance-marker-v1-MsgWithdrawEscrowProposalRequest"></a>

### MsgWithdrawEscrowProposalRequest
MsgWithdrawEscrowProposalRequest defines the Msg/WithdrawEscrowProposal request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated |  |
| `target_address` | [string](#string) |  |  |
| `authority` | [string](#string) |  | The signer of the message. Must have admin authority to marker or be governance module account address. |






<a name="provenance-marker-v1-MsgWithdrawEscrowProposalResponse"></a>

### MsgWithdrawEscrowProposalResponse
MsgWithdrawEscrowProposalResponse defines the Msg/WithdrawEscrowProposal response type






<a name="provenance-marker-v1-MsgWithdrawRequest"></a>

### MsgWithdrawRequest
MsgWithdrawRequest defines the Msg/Withdraw request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `to_address` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated |  |






<a name="provenance-marker-v1-MsgWithdrawResponse"></a>

### MsgWithdrawResponse
MsgWithdrawResponse defines the Msg/Withdraw response type





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-marker-v1-Msg"></a>

### Msg
Msg defines the Marker Msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Finalize` | [MsgFinalizeRequest](#provenance-marker-v1-MsgFinalizeRequest) | [MsgFinalizeResponse](#provenance-marker-v1-MsgFinalizeResponse) | Finalize |
| `Activate` | [MsgActivateRequest](#provenance-marker-v1-MsgActivateRequest) | [MsgActivateResponse](#provenance-marker-v1-MsgActivateResponse) | Activate |
| `Cancel` | [MsgCancelRequest](#provenance-marker-v1-MsgCancelRequest) | [MsgCancelResponse](#provenance-marker-v1-MsgCancelResponse) | Cancel |
| `Delete` | [MsgDeleteRequest](#provenance-marker-v1-MsgDeleteRequest) | [MsgDeleteResponse](#provenance-marker-v1-MsgDeleteResponse) | Delete |
| `Mint` | [MsgMintRequest](#provenance-marker-v1-MsgMintRequest) | [MsgMintResponse](#provenance-marker-v1-MsgMintResponse) | Mint |
| `Burn` | [MsgBurnRequest](#provenance-marker-v1-MsgBurnRequest) | [MsgBurnResponse](#provenance-marker-v1-MsgBurnResponse) | Burn |
| `AddAccess` | [MsgAddAccessRequest](#provenance-marker-v1-MsgAddAccessRequest) | [MsgAddAccessResponse](#provenance-marker-v1-MsgAddAccessResponse) | AddAccess |
| `DeleteAccess` | [MsgDeleteAccessRequest](#provenance-marker-v1-MsgDeleteAccessRequest) | [MsgDeleteAccessResponse](#provenance-marker-v1-MsgDeleteAccessResponse) | DeleteAccess |
| `Withdraw` | [MsgWithdrawRequest](#provenance-marker-v1-MsgWithdrawRequest) | [MsgWithdrawResponse](#provenance-marker-v1-MsgWithdrawResponse) | Withdraw |
| `AddMarker` | [MsgAddMarkerRequest](#provenance-marker-v1-MsgAddMarkerRequest) | [MsgAddMarkerResponse](#provenance-marker-v1-MsgAddMarkerResponse) | AddMarker |
| `Transfer` | [MsgTransferRequest](#provenance-marker-v1-MsgTransferRequest) | [MsgTransferResponse](#provenance-marker-v1-MsgTransferResponse) | Transfer marker denominated coin between accounts |
| `IbcTransfer` | [MsgIbcTransferRequest](#provenance-marker-v1-MsgIbcTransferRequest) | [MsgIbcTransferResponse](#provenance-marker-v1-MsgIbcTransferResponse) | Transfer over ibc any marker(including restricted markers) between ibc accounts. The relayer is still needed to accomplish ibc middleware relays. |
| `SetDenomMetadata` | [MsgSetDenomMetadataRequest](#provenance-marker-v1-MsgSetDenomMetadataRequest) | [MsgSetDenomMetadataResponse](#provenance-marker-v1-MsgSetDenomMetadataResponse) | Allows Denom Metadata (see bank module) to be set for the Marker's Denom |
| `GrantAllowance` | [MsgGrantAllowanceRequest](#provenance-marker-v1-MsgGrantAllowanceRequest) | [MsgGrantAllowanceResponse](#provenance-marker-v1-MsgGrantAllowanceResponse) | GrantAllowance grants fee allowance to the grantee on the granter's account with the provided expiration time. |
| `AddFinalizeActivateMarker` | [MsgAddFinalizeActivateMarkerRequest](#provenance-marker-v1-MsgAddFinalizeActivateMarkerRequest) | [MsgAddFinalizeActivateMarkerResponse](#provenance-marker-v1-MsgAddFinalizeActivateMarkerResponse) | AddFinalizeActivateMarker |
| `SupplyIncreaseProposal` | [MsgSupplyIncreaseProposalRequest](#provenance-marker-v1-MsgSupplyIncreaseProposalRequest) | [MsgSupplyIncreaseProposalResponse](#provenance-marker-v1-MsgSupplyIncreaseProposalResponse) | SupplyIncreaseProposal can only be called via gov proposal |
| `SupplyDecreaseProposal` | [MsgSupplyDecreaseProposalRequest](#provenance-marker-v1-MsgSupplyDecreaseProposalRequest) | [MsgSupplyDecreaseProposalResponse](#provenance-marker-v1-MsgSupplyDecreaseProposalResponse) | SupplyDecreaseProposal can only be called via gov proposal |
| `UpdateRequiredAttributes` | [MsgUpdateRequiredAttributesRequest](#provenance-marker-v1-MsgUpdateRequiredAttributesRequest) | [MsgUpdateRequiredAttributesResponse](#provenance-marker-v1-MsgUpdateRequiredAttributesResponse) | UpdateRequiredAttributes will only succeed if signer has transfer authority |
| `UpdateForcedTransfer` | [MsgUpdateForcedTransferRequest](#provenance-marker-v1-MsgUpdateForcedTransferRequest) | [MsgUpdateForcedTransferResponse](#provenance-marker-v1-MsgUpdateForcedTransferResponse) | UpdateForcedTransfer updates the allow_forced_transfer field of a marker via governance proposal. |
| `SetAccountData` | [MsgSetAccountDataRequest](#provenance-marker-v1-MsgSetAccountDataRequest) | [MsgSetAccountDataResponse](#provenance-marker-v1-MsgSetAccountDataResponse) | SetAccountData sets the accountdata for a denom. Signer must have deposit authority. |
| `UpdateSendDenyList` | [MsgUpdateSendDenyListRequest](#provenance-marker-v1-MsgUpdateSendDenyListRequest) | [MsgUpdateSendDenyListResponse](#provenance-marker-v1-MsgUpdateSendDenyListResponse) | UpdateSendDenyList will only succeed if signer has admin authority |
| `AddNetAssetValues` | [MsgAddNetAssetValuesRequest](#provenance-marker-v1-MsgAddNetAssetValuesRequest) | [MsgAddNetAssetValuesResponse](#provenance-marker-v1-MsgAddNetAssetValuesResponse) | AddNetAssetValues set the net asset value for a marker |
| `SetAdministratorProposal` | [MsgSetAdministratorProposalRequest](#provenance-marker-v1-MsgSetAdministratorProposalRequest) | [MsgSetAdministratorProposalResponse](#provenance-marker-v1-MsgSetAdministratorProposalResponse) | SetAdministratorProposal sets administrators with specific access on the marker |
| `RemoveAdministratorProposal` | [MsgRemoveAdministratorProposalRequest](#provenance-marker-v1-MsgRemoveAdministratorProposalRequest) | [MsgRemoveAdministratorProposalResponse](#provenance-marker-v1-MsgRemoveAdministratorProposalResponse) | RemoveAdministratorProposal removes administrators with specific access on the marker |
| `ChangeStatusProposal` | [MsgChangeStatusProposalRequest](#provenance-marker-v1-MsgChangeStatusProposalRequest) | [MsgChangeStatusProposalResponse](#provenance-marker-v1-MsgChangeStatusProposalResponse) | ChangeStatusProposal is a governance proposal change marker status |
| `WithdrawEscrowProposal` | [MsgWithdrawEscrowProposalRequest](#provenance-marker-v1-MsgWithdrawEscrowProposalRequest) | [MsgWithdrawEscrowProposalResponse](#provenance-marker-v1-MsgWithdrawEscrowProposalResponse) | WithdrawEscrowProposal is a governance proposal to withdraw escrow coins from a marker |
| `SetDenomMetadataProposal` | [MsgSetDenomMetadataProposalRequest](#provenance-marker-v1-MsgSetDenomMetadataProposalRequest) | [MsgSetDenomMetadataProposalResponse](#provenance-marker-v1-MsgSetDenomMetadataProposalResponse) | SetDenomMetadataProposal is a governance proposal to set marker metadata |
| `UpdateParams` | [MsgUpdateParamsRequest](#provenance-marker-v1-MsgUpdateParamsRequest) | [MsgUpdateParamsResponse](#provenance-marker-v1-MsgUpdateParamsResponse) | UpdateParams is a governance proposal endpoint for updating the marker module's params. |

 <!-- end services -->



<a name="provenance_marker_v1_si-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/si.proto


 <!-- end messages -->


<a name="provenance-marker-v1-SIPrefix"></a>

### SIPrefix
SIPrefix represents an International System of Units (SI) Prefix.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `SI_PREFIX_NONE` | `0` | 10^0 (none) |
| `SI_PREFIX_DEKA` | `1` | 10^1 deka da |
| `SI_PREFIX_HECTO` | `2` | 10^2 hecto h |
| `SI_PREFIX_KILO` | `3` | 10^3 kilo k |
| `SI_PREFIX_MEGA` | `6` | 10^6 mega M |
| `SI_PREFIX_GIGA` | `9` | 10^9 giga G |
| `SI_PREFIX_TERA` | `12` | 10^12 tera T |
| `SI_PREFIX_PETA` | `15` | 10^15 peta P |
| `SI_PREFIX_EXA` | `18` | 10^18 exa E |
| `SI_PREFIX_ZETTA` | `21` | 10^21 zetta Z |
| `SI_PREFIX_YOTTA` | `24` | 10^24 yotta Y |
| `SI_PREFIX_DECI` | `-1` | 10^-1 deci d |
| `SI_PREFIX_CENTI` | `-2` | 10^-2 centi c |
| `SI_PREFIX_MILLI` | `-3` | 10^-3 milli m |
| `SI_PREFIX_MICRO` | `-6` | 10^-6 micro µ |
| `SI_PREFIX_NANO` | `-9` | 10^-9 nano n |
| `SI_PREFIX_PICO` | `-12` | 10^-12 pico p |
| `SI_PREFIX_FEMTO` | `-15` | 10^-15 femto f |
| `SI_PREFIX_ATTO` | `-18` | 10^-18 atto a |
| `SI_PREFIX_ZEPTO` | `-21` | 10^-21 zepto z |
| `SI_PREFIX_YOCTO` | `-24` | 10^-24 yocto y |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_marker_v1_marker-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/marker.proto



<a name="provenance-marker-v1-EventDenomUnit"></a>

### EventDenomUnit
EventDenomUnit denom units for set denom metadata event


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `exponent` | [string](#string) |  |  |
| `aliases` | [string](#string) | repeated |  |






<a name="provenance-marker-v1-EventMarkerAccess"></a>

### EventMarkerAccess
EventMarkerAccess event access permissions for address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `permissions` | [string](#string) | repeated |  |






<a name="provenance-marker-v1-EventMarkerActivate"></a>

### EventMarkerActivate
EventMarkerActivate event emitted when marker is activated


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerAdd"></a>

### EventMarkerAdd
EventMarkerAdd event emitted when marker is added


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `amount` | [string](#string) |  |  |
| `status` | [string](#string) |  |  |
| `manager` | [string](#string) |  |  |
| `marker_type` | [string](#string) |  |  |
| `address` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerAddAccess"></a>

### EventMarkerAddAccess
EventMarkerAddAccess event emitted when marker access is added


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `access` | [EventMarkerAccess](#provenance-marker-v1-EventMarkerAccess) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerBurn"></a>

### EventMarkerBurn
EventMarkerBurn event emitted when coin is burned from marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerCancel"></a>

### EventMarkerCancel
EventMarkerCancel event emitted when marker is cancelled


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerDelete"></a>

### EventMarkerDelete
EventMarkerDelete event emitted when marker is deleted


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerDeleteAccess"></a>

### EventMarkerDeleteAccess
EventMarkerDeleteAccess event emitted when marker access is revoked


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `remove_address` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerFinalize"></a>

### EventMarkerFinalize
EventMarkerFinalize event emitted when marker is finalized


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerMint"></a>

### EventMarkerMint
EventMarkerMint event emitted when additional marker supply is minted


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerParamsUpdated"></a>

### EventMarkerParamsUpdated
EventMarkerParamsUpdated event emitted when marker params are updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `enable_governance` | [string](#string) |  |  |
| `unrestricted_denom_regex` | [string](#string) |  |  |
| `max_supply` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerSetDenomMetadata"></a>

### EventMarkerSetDenomMetadata
EventMarkerSetDenomMetadata event emitted when metadata is set on marker with denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata_base` | [string](#string) |  |  |
| `metadata_description` | [string](#string) |  |  |
| `metadata_display` | [string](#string) |  |  |
| `metadata_denom_units` | [EventDenomUnit](#provenance-marker-v1-EventDenomUnit) | repeated |  |
| `administrator` | [string](#string) |  |  |
| `metadata_name` | [string](#string) |  |  |
| `metadata_symbol` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerTransfer"></a>

### EventMarkerTransfer
EventMarkerTransfer event emitted when coins are transfered to from account to another


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `to_address` | [string](#string) |  |  |
| `from_address` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventMarkerWithdraw"></a>

### EventMarkerWithdraw
EventMarkerWithdraw event emitted when coins are withdrew from marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `to_address` | [string](#string) |  |  |






<a name="provenance-marker-v1-EventSetNetAssetValue"></a>

### EventSetNetAssetValue
EventSetNetAssetValue event emitted when Net Asset Value for marker is update or added


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `price` | [string](#string) |  |  |
| `volume` | [string](#string) |  |  |
| `source` | [string](#string) |  |  |






<a name="provenance-marker-v1-MarkerAccount"></a>

### MarkerAccount
MarkerAccount holds the marker configuration information in addition to a base account structure.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_account` | [cosmos.auth.v1beta1.BaseAccount](#cosmos-auth-v1beta1-BaseAccount) |  | base cosmos account information including address and coin holdings. |
| `manager` | [string](#string) |  | Address that owns the marker configuration. This account must sign any requests to change marker config (only valid for statuses prior to finalization) |
| `access_control` | [AccessGrant](#provenance-marker-v1-AccessGrant) | repeated | Access control lists |
| `status` | [MarkerStatus](#provenance-marker-v1-MarkerStatus) |  | Indicates the current status of this marker record. |
| `denom` | [string](#string) |  | value denomination and total supply for the token. |
| `supply` | [string](#string) |  | the total supply expected for a marker. This is the amount that is minted when a marker is created. |
| `marker_type` | [MarkerType](#provenance-marker-v1-MarkerType) |  | Marker type information |
| `supply_fixed` | [bool](#bool) |  | A fixed supply will mint additional coin automatically if the total supply decreases below a set value. This may occur if the coin is burned or an account holding the coin is slashed. (default: true) |
| `allow_governance_control` | [bool](#bool) |  | indicates that governance based control is allowed for this marker |
| `allow_forced_transfer` | [bool](#bool) |  | Whether an admin can transfer restricted coins from a 3rd-party account without their signature. |
| `required_attributes` | [string](#string) | repeated | list of required attributes on restricted marker in order to send and receive transfers if sender does not have transfer authority |






<a name="provenance-marker-v1-NetAssetValue"></a>

### NetAssetValue
NetAssetValue defines a marker's net asset value


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `price` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | price is the complete value of the asset's volume |
| `volume` | [uint64](#uint64) |  | volume is the number of tokens of the marker that were purchased for the price |
| `updated_block_height` | [uint64](#uint64) |  | updated_block_height is the block height of last update |






<a name="provenance-marker-v1-Params"></a>

### Params
Params defines the set of params for the account module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_total_supply` | [uint64](#uint64) |  | **Deprecated.** Deprecated: Prefer to use `max_supply` instead. Maximum amount of supply to allow a marker to be created with |
| `enable_governance` | [bool](#bool) |  | indicates if governance based controls of markers is allowed. |
| `unrestricted_denom_regex` | [string](#string) |  | a regular expression used to validate marker denom values from normal create requests (governance requests are only subject to platform coin validation denom expression) |
| `max_supply` | [string](#string) |  | maximum amount of supply to allow a marker to be created with |





 <!-- end messages -->


<a name="provenance-marker-v1-MarkerStatus"></a>

### MarkerStatus
MarkerStatus defines the various states a marker account can be in.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `MARKER_STATUS_UNSPECIFIED` | `0` | MARKER_STATUS_UNSPECIFIED - Unknown/Invalid Marker Status |
| `MARKER_STATUS_PROPOSED` | `1` | MARKER_STATUS_PROPOSED - Initial configuration period, updates allowed, token supply not created. |
| `MARKER_STATUS_FINALIZED` | `2` | MARKER_STATUS_FINALIZED - Configuration finalized, ready for supply creation |
| `MARKER_STATUS_ACTIVE` | `3` | MARKER_STATUS_ACTIVE - Supply is created, rules are in force. |
| `MARKER_STATUS_CANCELLED` | `4` | MARKER_STATUS_CANCELLED - Marker has been cancelled, pending destroy |
| `MARKER_STATUS_DESTROYED` | `5` | MARKER_STATUS_DESTROYED - Marker supply has all been recalled, marker is considered destroyed and no further actions allowed. |



<a name="provenance-marker-v1-MarkerType"></a>

### MarkerType
MarkerType defines the types of marker

| Name | Number | Description |
| ---- | ------ | ----------- |
| `MARKER_TYPE_UNSPECIFIED` | `0` | MARKER_TYPE_UNSPECIFIED is an invalid/unknown marker type. |
| `MARKER_TYPE_COIN` | `1` | MARKER_TYPE_COIN is a marker that represents a standard fungible coin (default). |
| `MARKER_TYPE_RESTRICTED` | `2` | MARKER_TYPE_RESTRICTED is a marker that represents a denom with send_enabled = false. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_marker_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/query.proto



<a name="provenance-marker-v1-Balance"></a>

### Balance
Balance defines an account address and balance pair used in queries for accounts holding a marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of the balance holder. |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | coins defines the different coins this balance holds. |






<a name="provenance-marker-v1-QueryAccessRequest"></a>

### QueryAccessRequest
QueryAccessRequest is the request type for the Query/MarkerAccess method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | address or denom for the marker |






<a name="provenance-marker-v1-QueryAccessResponse"></a>

### QueryAccessResponse
QueryAccessResponse is the response type for the Query/MarkerAccess method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `accounts` | [AccessGrant](#provenance-marker-v1-AccessGrant) | repeated |  |






<a name="provenance-marker-v1-QueryAccountDataRequest"></a>

### QueryAccountDataRequest
QueryAccountDataRequest is the request type for the Query/AccountData


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | The denomination to look up. |






<a name="provenance-marker-v1-QueryAccountDataResponse"></a>

### QueryAccountDataResponse
QueryAccountDataResponse is the response type for the Query/AccountData


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  | The accountdata for the requested denom. |






<a name="provenance-marker-v1-QueryAllMarkersRequest"></a>

### QueryAllMarkersRequest
QueryAllMarkersRequest is the request type for the Query/AllMarkers method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status` | [MarkerStatus](#provenance-marker-v1-MarkerStatus) |  | Optional status to filter request |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-marker-v1-QueryAllMarkersResponse"></a>

### QueryAllMarkersResponse
QueryAllMarkersResponse is the response type for the Query/AllMarkers method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `markers` | [google.protobuf.Any](#google-protobuf-Any) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance-marker-v1-QueryDenomMetadataRequest"></a>

### QueryDenomMetadataRequest
QueryDenomMetadataRequest is the request type for Query/DenomMetadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="provenance-marker-v1-QueryDenomMetadataResponse"></a>

### QueryDenomMetadataResponse
QueryDenomMetadataResponse is the response type for the Query/DenomMetadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata` | [cosmos.bank.v1beta1.Metadata](#cosmos-bank-v1beta1-Metadata) |  |  |






<a name="provenance-marker-v1-QueryEscrowRequest"></a>

### QueryEscrowRequest
QueryEscrowRequest is the request type for the Query/MarkerEscrow method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | address or denom for the marker |






<a name="provenance-marker-v1-QueryEscrowResponse"></a>

### QueryEscrowResponse
QueryEscrowResponse is the response type for the Query/MarkerEscrow method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `escrow` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated |  |






<a name="provenance-marker-v1-QueryHoldingRequest"></a>

### QueryHoldingRequest
QueryHoldingRequest is the request type for the Query/MarkerHolders method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | the address or denom of the marker |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-marker-v1-QueryHoldingResponse"></a>

### QueryHoldingResponse
QueryHoldingResponse is the response type for the Query/MarkerHolders method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balances` | [Balance](#provenance-marker-v1-Balance) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance-marker-v1-QueryMarkerRequest"></a>

### QueryMarkerRequest
QueryMarkerRequest is the request type for the Query/Marker method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | the address or denom of the marker |






<a name="provenance-marker-v1-QueryMarkerResponse"></a>

### QueryMarkerResponse
QueryMarkerResponse is the response type for the Query/Marker method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `marker` | [google.protobuf.Any](#google-protobuf-Any) |  |  |






<a name="provenance-marker-v1-QueryNetAssetValuesRequest"></a>

### QueryNetAssetValuesRequest
QueryNetAssetValuesRequest is the request type for the Query/NetAssetValues method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | address or denom for the marker |






<a name="provenance-marker-v1-QueryNetAssetValuesResponse"></a>

### QueryNetAssetValuesResponse
QueryNetAssetValuesRequest is the response type for the Query/NetAssetValues method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `net_asset_values` | [NetAssetValue](#provenance-marker-v1-NetAssetValue) | repeated | net asset values for marker denom |






<a name="provenance-marker-v1-QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance-marker-v1-QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-marker-v1-Params) |  | params defines the parameters of the module. |






<a name="provenance-marker-v1-QuerySupplyRequest"></a>

### QuerySupplyRequest
QuerySupplyRequest is the request type for the Query/MarkerSupply method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | address or denom for the marker |






<a name="provenance-marker-v1-QuerySupplyResponse"></a>

### QuerySupplyResponse
QuerySupplyResponse is the response type for the Query/MarkerSupply method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | amount is the supply of the marker. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-marker-v1-Query"></a>

### Query
Query defines the gRPC querier service for marker module.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Params` | [QueryParamsRequest](#provenance-marker-v1-QueryParamsRequest) | [QueryParamsResponse](#provenance-marker-v1-QueryParamsResponse) | Params queries the parameters of x/bank module. |
| `AllMarkers` | [QueryAllMarkersRequest](#provenance-marker-v1-QueryAllMarkersRequest) | [QueryAllMarkersResponse](#provenance-marker-v1-QueryAllMarkersResponse) | Returns a list of all markers on the blockchain |
| `Marker` | [QueryMarkerRequest](#provenance-marker-v1-QueryMarkerRequest) | [QueryMarkerResponse](#provenance-marker-v1-QueryMarkerResponse) | query for a single marker by denom or address |
| `Holding` | [QueryHoldingRequest](#provenance-marker-v1-QueryHoldingRequest) | [QueryHoldingResponse](#provenance-marker-v1-QueryHoldingResponse) | query for all accounts holding the given marker coins |
| `Supply` | [QuerySupplyRequest](#provenance-marker-v1-QuerySupplyRequest) | [QuerySupplyResponse](#provenance-marker-v1-QuerySupplyResponse) | query for supply of coin on a marker account |
| `Escrow` | [QueryEscrowRequest](#provenance-marker-v1-QueryEscrowRequest) | [QueryEscrowResponse](#provenance-marker-v1-QueryEscrowResponse) | query for coins on a marker account |
| `Access` | [QueryAccessRequest](#provenance-marker-v1-QueryAccessRequest) | [QueryAccessResponse](#provenance-marker-v1-QueryAccessResponse) | query for access records on an account |
| `DenomMetadata` | [QueryDenomMetadataRequest](#provenance-marker-v1-QueryDenomMetadataRequest) | [QueryDenomMetadataResponse](#provenance-marker-v1-QueryDenomMetadataResponse) | query for access records on an account |
| `AccountData` | [QueryAccountDataRequest](#provenance-marker-v1-QueryAccountDataRequest) | [QueryAccountDataResponse](#provenance-marker-v1-QueryAccountDataResponse) | query for account data associated with a denom |
| `NetAssetValues` | [QueryNetAssetValuesRequest](#provenance-marker-v1-QueryNetAssetValuesRequest) | [QueryNetAssetValuesResponse](#provenance-marker-v1-QueryNetAssetValuesResponse) | NetAssetValues returns net asset values for marker |

 <!-- end services -->



<a name="provenance_marker_v1_accessgrant-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/accessgrant.proto



<a name="provenance-marker-v1-AccessGrant"></a>

### AccessGrant
AccessGrant associates a collection of permissions with an address for delegated marker account control.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `permissions` | [Access](#provenance-marker-v1-Access) | repeated |  |





 <!-- end messages -->


<a name="provenance-marker-v1-Access"></a>

### Access
Access defines the different types of permissions that a marker supports granting to an address.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `ACCESS_UNSPECIFIED` | `0` | ACCESS_UNSPECIFIED defines a no-op vote option. |
| `ACCESS_MINT` | `1` | ACCESS_MINT is the ability to increase the supply of a marker. |
| `ACCESS_BURN` | `2` | ACCESS_BURN is the ability to decrease the supply of the marker using coin held by the marker. |
| `ACCESS_DEPOSIT` | `3` | ACCESS_DEPOSIT is the ability to transfer funds from another account to this marker account or to set a reference to this marker in the metadata/scopes module. |
| `ACCESS_WITHDRAW` | `4` | ACCESS_WITHDRAW is the ability to transfer funds from this marker account to another account or to remove a reference to this marker in the metadata/scopes module. |
| `ACCESS_DELETE` | `5` | ACCESS_DELETE is the ability to move a proposed, finalized or active marker into the cancelled state. This access also allows cancelled markers to be marked for deletion. |
| `ACCESS_ADMIN` | `6` | ACCESS_ADMIN is the ability to add access grants for accounts to the list of marker permissions. This access also gives the ability to update the marker's denom metadata. |
| `ACCESS_TRANSFER` | `7` | ACCESS_TRANSFER is the ability to manage transfer settings and broker transfers of the marker. Accounts with this access can: - Update the marker's required attributes. - Update the send-deny list. - Use the transfer or bank send endpoints to move marker funds out of their own account. This access right is only supported on RESTRICTED markers. |
| `ACCESS_FORCE_TRANSFER` | `8` | ACCESS_FORCE_TRANSFER is the ability to transfer restricted coins from a 3rd-party account without their signature. This access right is only supported on RESTRICTED markers and only has meaning when allow_forced_transfer is true. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_marker_v1_authz-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/authz.proto



<a name="provenance-marker-v1-MarkerTransferAuthorization"></a>

### MarkerTransferAuthorization
MarkerTransferAuthorization gives the grantee permissions to execute
a marker transfer on behalf of the granter's account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfer_limit` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | transfer_limit is the total amount the grantee can transfer |
| `allow_list` | [string](#string) | repeated | allow_list specifies an optional list of addresses to whom the grantee can send restricted coins on behalf of the granter. If omitted, any recipient is allowed. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_marker_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/genesis.proto



<a name="provenance-marker-v1-DenySendAddress"></a>

### DenySendAddress
DenySendAddress defines addresses that are denied sends for marker denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `marker_address` | [string](#string) |  | marker_address is the marker's address for denied address |
| `deny_address` | [string](#string) |  | deny_address defines all wallet addresses that are denied sends for the marker |






<a name="provenance-marker-v1-GenesisState"></a>

### GenesisState
GenesisState defines the account module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-marker-v1-Params) |  | params defines all the parameters of the module. |
| `markers` | [MarkerAccount](#provenance-marker-v1-MarkerAccount) | repeated | A collection of marker accounts to create on start |
| `net_asset_values` | [MarkerNetAssetValues](#provenance-marker-v1-MarkerNetAssetValues) | repeated | list of marker net asset values |
| `deny_send_addresses` | [DenySendAddress](#provenance-marker-v1-DenySendAddress) | repeated | list of denom based denied send addresses |






<a name="provenance-marker-v1-MarkerNetAssetValues"></a>

### MarkerNetAssetValues
MarkerNetAssetValues defines the net asset values for a marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address defines the marker address |
| `net_asset_values` | [NetAssetValue](#provenance-marker-v1-NetAssetValue) | repeated | net_asset_values that are assigned to marker |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_marker_v1_proposals-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/proposals.proto



<a name="provenance-marker-v1-AddMarkerProposal"></a>

### AddMarkerProposal
AddMarkerProposal is deprecated and can no longer be used.
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgAddMarkerRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |
| `manager` | [string](#string) |  |  |
| `status` | [MarkerStatus](#provenance-marker-v1-MarkerStatus) |  |  |
| `marker_type` | [MarkerType](#provenance-marker-v1-MarkerType) |  |  |
| `access_list` | [AccessGrant](#provenance-marker-v1-AccessGrant) | repeated |  |
| `supply_fixed` | [bool](#bool) |  |  |
| `allow_governance_control` | [bool](#bool) |  |  |






<a name="provenance-marker-v1-ChangeStatusProposal"></a>

### ChangeStatusProposal
ChangeStatusProposal defines a governance proposal to administer a marker to change its status
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgChangeStatusProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `new_status` | [MarkerStatus](#provenance-marker-v1-MarkerStatus) |  |  |






<a name="provenance-marker-v1-RemoveAdministratorProposal"></a>

### RemoveAdministratorProposal
RemoveAdministratorProposal defines a governance proposal to administer a marker and remove all permissions for a
given address
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgRemoveAdministratorProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `removed_address` | [string](#string) | repeated |  |






<a name="provenance-marker-v1-SetAdministratorProposal"></a>

### SetAdministratorProposal
SetAdministratorProposal defines a governance proposal to administer a marker and set administrators with specific
access on the marker
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgSetAdministratorProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `access` | [AccessGrant](#provenance-marker-v1-AccessGrant) | repeated |  |






<a name="provenance-marker-v1-SetDenomMetadataProposal"></a>

### SetDenomMetadataProposal
SetDenomMetadataProposal defines a governance proposal to set the metadata for a denom
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgSetDenomMetadataProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `metadata` | [cosmos.bank.v1beta1.Metadata](#cosmos-bank-v1beta1-Metadata) |  |  |






<a name="provenance-marker-v1-SupplyDecreaseProposal"></a>

### SupplyDecreaseProposal
SupplyDecreaseProposal defines a governance proposal to administer a marker and decrease the total supply through
burning coin held within the marker
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgSupplyDecreaseProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |






<a name="provenance-marker-v1-SupplyIncreaseProposal"></a>

### SupplyIncreaseProposal
SupplyIncreaseProposal defines a governance proposal to administer a marker and increase total supply of the marker
through minting coin and placing it within the marker or assigning it directly to an account
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgSupplyIncreaseProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  |  |
| `target_address` | [string](#string) |  | an optional target address for the minted coin from this request |






<a name="provenance-marker-v1-WithdrawEscrowProposal"></a>

### WithdrawEscrowProposal
WithdrawEscrowProposal defines a governance proposal to withdraw escrow coins from a marker
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgWithdrawEscrowProposalRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated |  |
| `target_address` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_name_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/name/v1/tx.proto



<a name="provenance-name-v1-MsgBindNameRequest"></a>

### MsgBindNameRequest
MsgBindNameRequest defines an sdk.Msg type that is used to add an address/name binding under an optional parent name.
The record may optionally be restricted to prevent additional names from being added under this one without the
owner signing the request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `parent` | [NameRecord](#provenance-name-v1-NameRecord) |  | The parent record to bind this name under. |
| `record` | [NameRecord](#provenance-name-v1-NameRecord) |  | The name record to bind under the parent |






<a name="provenance-name-v1-MsgBindNameResponse"></a>

### MsgBindNameResponse
MsgBindNameResponse defines the Msg/BindName response type.






<a name="provenance-name-v1-MsgCreateRootNameRequest"></a>

### MsgCreateRootNameRequest
MsgCreateRootNameRequest defines an sdk.Msg type to create a new root name
that is controlled by a given owner and optionally restricted to the owner
for the sole creation of sub names.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | The signing authority for the request |
| `record` | [NameRecord](#provenance-name-v1-NameRecord) |  | NameRecord is a structure used to bind ownership of a name hierarchy to a collection of addresses |






<a name="provenance-name-v1-MsgCreateRootNameResponse"></a>

### MsgCreateRootNameResponse
MsgCreateRootNameResponse defines Msg/CreateRootName response type.






<a name="provenance-name-v1-MsgDeleteNameRequest"></a>

### MsgDeleteNameRequest
MsgDeleteNameRequest defines an sdk.Msg type that is used to remove an existing address/name binding.  The binding
may not have any child names currently bound for this request to be successful. All associated attributes on account
addresses will be deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record` | [NameRecord](#provenance-name-v1-NameRecord) |  | The record being removed |






<a name="provenance-name-v1-MsgDeleteNameResponse"></a>

### MsgDeleteNameResponse
MsgDeleteNameResponse defines the Msg/DeleteName response type.






<a name="provenance-name-v1-MsgModifyNameRequest"></a>

### MsgModifyNameRequest
MsgModifyNameRequest defines a governance method that is used to update an existing address/name binding.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | The address signing the message |
| `record` | [NameRecord](#provenance-name-v1-NameRecord) |  | The record being updated |






<a name="provenance-name-v1-MsgModifyNameResponse"></a>

### MsgModifyNameResponse
MsgModifyNameResponse defines the Msg/ModifyName response type.






<a name="provenance-name-v1-MsgUpdateParamsRequest"></a>

### MsgUpdateParamsRequest
MsgUpdateParamsRequest is a request message for the UpdateParams endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `params` | [Params](#provenance-name-v1-Params) |  | params are the new param values to set. |






<a name="provenance-name-v1-MsgUpdateParamsResponse"></a>

### MsgUpdateParamsResponse
MsgUpdateParamsResponse is a response message for the UpdateParams endpoint.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-name-v1-Msg"></a>

### Msg
Msg defines the bank Msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `BindName` | [MsgBindNameRequest](#provenance-name-v1-MsgBindNameRequest) | [MsgBindNameResponse](#provenance-name-v1-MsgBindNameResponse) | BindName binds a name to an address under a root name. |
| `DeleteName` | [MsgDeleteNameRequest](#provenance-name-v1-MsgDeleteNameRequest) | [MsgDeleteNameResponse](#provenance-name-v1-MsgDeleteNameResponse) | DeleteName defines a method to verify a particular invariance. |
| `ModifyName` | [MsgModifyNameRequest](#provenance-name-v1-MsgModifyNameRequest) | [MsgModifyNameResponse](#provenance-name-v1-MsgModifyNameResponse) | ModifyName defines a method to modify the attributes of an existing name. |
| `CreateRootName` | [MsgCreateRootNameRequest](#provenance-name-v1-MsgCreateRootNameRequest) | [MsgCreateRootNameResponse](#provenance-name-v1-MsgCreateRootNameResponse) | CreateRootName defines a governance method for creating a root name. |
| `UpdateParams` | [MsgUpdateParamsRequest](#provenance-name-v1-MsgUpdateParamsRequest) | [MsgUpdateParamsResponse](#provenance-name-v1-MsgUpdateParamsResponse) | UpdateParams is a governance proposal endpoint for updating the name module's params. |

 <!-- end services -->



<a name="provenance_name_v1_name-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/name/v1/name.proto



<a name="provenance-name-v1-CreateRootNameProposal"></a>

### CreateRootNameProposal
CreateRootNameProposal details a proposal to create a new root name
that is controlled by a given owner and optionally restricted to the owner
for the sole creation of sub names.
Deprecated: This legacy proposal is deprecated in favor of Msg-based gov
proposals, see MsgCreateRootNameRequest.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | proposal title |
| `description` | [string](#string) |  | proposal description |
| `name` | [string](#string) |  | the bound name |
| `owner` | [string](#string) |  | the address the name will resolve to |
| `restricted` | [bool](#bool) |  | a flag that indicates if an owner signature is required to add sub-names |






<a name="provenance-name-v1-EventNameBound"></a>

### EventNameBound
Event emitted when name is bound.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `restricted` | [bool](#bool) |  |  |






<a name="provenance-name-v1-EventNameParamsUpdated"></a>

### EventNameParamsUpdated
EventNameParamsUpdated event emitted when name params are updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `allow_unrestricted_names` | [string](#string) |  |  |
| `max_name_levels` | [string](#string) |  |  |
| `min_segment_length` | [string](#string) |  |  |
| `max_segment_length` | [string](#string) |  |  |






<a name="provenance-name-v1-EventNameUnbound"></a>

### EventNameUnbound
Event emitted when name is unbound.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `restricted` | [bool](#bool) |  |  |






<a name="provenance-name-v1-EventNameUpdate"></a>

### EventNameUpdate
Event emitted when name is updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `restricted` | [bool](#bool) |  |  |






<a name="provenance-name-v1-NameRecord"></a>

### NameRecord
NameRecord is a structure used to bind ownership of a name hierarchy to a collection of addresses


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | the bound name |
| `address` | [string](#string) |  | the address the name resolved to |
| `restricted` | [bool](#bool) |  | whether owner signature is required to add sub-names |






<a name="provenance-name-v1-Params"></a>

### Params
Params defines the set of params for the name module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_segment_length` | [uint32](#uint32) |  | maximum length of name segment to allow |
| `min_segment_length` | [uint32](#uint32) |  | minimum length of name segment to allow |
| `max_name_levels` | [uint32](#uint32) |  | maximum number of name segments to allow. Example: `foo.bar.baz` would be 3 |
| `allow_unrestricted_names` | [bool](#bool) |  | determines if unrestricted name keys are allowed or not |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_name_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/name/v1/query.proto



<a name="provenance-name-v1-QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance-name-v1-QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-name-v1-Params) |  | params defines the parameters of the module. |






<a name="provenance-name-v1-QueryResolveRequest"></a>

### QueryResolveRequest
QueryResolveRequest is the request type for the Query/Resolve method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | name to resolve the address for |






<a name="provenance-name-v1-QueryResolveResponse"></a>

### QueryResolveResponse
QueryResolveResponse is the response type for the Query/Resolve method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | a string containing the address the name resolves to |
| `restricted` | [bool](#bool) |  | Whether owner signature is required to add sub-names. |






<a name="provenance-name-v1-QueryReverseLookupRequest"></a>

### QueryReverseLookupRequest
QueryReverseLookupRequest is the request type for the Query/ReverseLookup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address to find name records for |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-name-v1-QueryReverseLookupResponse"></a>

### QueryReverseLookupResponse
QueryReverseLookupResponse is the response type for the Query/Resolve method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) | repeated | an array of names bound against a given address |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the request. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-name-v1-Query"></a>

### Query
Query defines the gRPC querier service for distribution module.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Params` | [QueryParamsRequest](#provenance-name-v1-QueryParamsRequest) | [QueryParamsResponse](#provenance-name-v1-QueryParamsResponse) | Params queries params of the name module. |
| `Resolve` | [QueryResolveRequest](#provenance-name-v1-QueryResolveRequest) | [QueryResolveResponse](#provenance-name-v1-QueryResolveResponse) | Resolve queries for the address associated with a given name |
| `ReverseLookup` | [QueryReverseLookupRequest](#provenance-name-v1-QueryReverseLookupRequest) | [QueryReverseLookupResponse](#provenance-name-v1-QueryReverseLookupResponse) | ReverseLookup queries for all names bound against a given address |

 <!-- end services -->



<a name="provenance_name_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/name/v1/genesis.proto



<a name="provenance-name-v1-GenesisState"></a>

### GenesisState
GenesisState defines the name module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-name-v1-Params) |  | params defines all the parameters of the module. |
| `bindings` | [NameRecord](#provenance-name-v1-NameRecord) | repeated | bindings defines all the name records present at genesis |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_metadata_v1_tx-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/tx.proto



<a name="provenance-metadata-v1-MsgAddContractSpecToScopeSpecRequest"></a>

### MsgAddContractSpecToScopeSpecRequest
MsgAddContractSpecToScopeSpecRequest is the request type for the Msg/AddContractSpecToScopeSpec RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_id` | [bytes](#bytes) |  | MetadataAddress for the contract specification to add. |
| `scope_specification_id` | [bytes](#bytes) |  | MetadataAddress for the scope specification to add contract specification to. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance-metadata-v1-MsgAddContractSpecToScopeSpecResponse"></a>

### MsgAddContractSpecToScopeSpecResponse
MsgAddContractSpecToScopeSpecResponse is the response type for the Msg/AddContractSpecToScopeSpec RPC method.






<a name="provenance-metadata-v1-MsgAddNetAssetValuesRequest"></a>

### MsgAddNetAssetValuesRequest
MsgAddNetAssetValuesRequest defines the Msg/AddNetAssetValues request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [string](#string) |  |  |
| `signers` | [string](#string) | repeated |  |
| `net_asset_values` | [NetAssetValue](#provenance-metadata-v1-NetAssetValue) | repeated |  |






<a name="provenance-metadata-v1-MsgAddNetAssetValuesResponse"></a>

### MsgAddNetAssetValuesResponse
MsgAddNetAssetValuesResponse defines the Msg/AddNetAssetValue response type






<a name="provenance-metadata-v1-MsgAddScopeDataAccessRequest"></a>

### MsgAddScopeDataAccessRequest
MsgAddScopeDataAccessRequest is the request to add data access AccAddress to scope


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope MetadataAddress for updating data access |
| `data_access` | [string](#string) | repeated | AccAddress addresses to be added to scope |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |






<a name="provenance-metadata-v1-MsgAddScopeDataAccessResponse"></a>

### MsgAddScopeDataAccessResponse
MsgAddScopeDataAccessResponse is the response for adding data access AccAddress to scope






<a name="provenance-metadata-v1-MsgAddScopeOwnerRequest"></a>

### MsgAddScopeOwnerRequest
MsgAddScopeOwnerRequest is the request to add owner AccAddress to scope


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope MetadataAddress for updating data access |
| `owners` | [Party](#provenance-metadata-v1-Party) | repeated | owner parties to add to the scope |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |






<a name="provenance-metadata-v1-MsgAddScopeOwnerResponse"></a>

### MsgAddScopeOwnerResponse
MsgAddScopeOwnerResponse is the response for adding owner AccAddresses to scope






<a name="provenance-metadata-v1-MsgBindOSLocatorRequest"></a>

### MsgBindOSLocatorRequest
MsgBindOSLocatorRequest is the request type for the Msg/BindOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) |  | The object locator to bind the address to bind to the URI. |






<a name="provenance-metadata-v1-MsgBindOSLocatorResponse"></a>

### MsgBindOSLocatorResponse
MsgBindOSLocatorResponse is the response type for the Msg/BindOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) |  |  |






<a name="provenance-metadata-v1-MsgDeleteContractSpecFromScopeSpecRequest"></a>

### MsgDeleteContractSpecFromScopeSpecRequest
MsgDeleteContractSpecFromScopeSpecRequest is the request type for the Msg/DeleteContractSpecFromScopeSpec RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_id` | [bytes](#bytes) |  | MetadataAddress for the contract specification to add. |
| `scope_specification_id` | [bytes](#bytes) |  | MetadataAddress for the scope specification to add contract specification to. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance-metadata-v1-MsgDeleteContractSpecFromScopeSpecResponse"></a>

### MsgDeleteContractSpecFromScopeSpecResponse
MsgDeleteContractSpecFromScopeSpecResponse is the response type for the Msg/DeleteContractSpecFromScopeSpec RPC
method.






<a name="provenance-metadata-v1-MsgDeleteContractSpecificationRequest"></a>

### MsgDeleteContractSpecificationRequest
MsgDeleteContractSpecificationRequest is the request type for the Msg/DeleteContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | MetadataAddress for the contract specification to delete. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance-metadata-v1-MsgDeleteContractSpecificationResponse"></a>

### MsgDeleteContractSpecificationResponse
MsgDeleteContractSpecificationResponse is the response type for the Msg/DeleteContractSpecification RPC method.






<a name="provenance-metadata-v1-MsgDeleteOSLocatorRequest"></a>

### MsgDeleteOSLocatorRequest
MsgDeleteOSLocatorRequest is the request type for the Msg/DeleteOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) |  | The record being removed |






<a name="provenance-metadata-v1-MsgDeleteOSLocatorResponse"></a>

### MsgDeleteOSLocatorResponse
MsgDeleteOSLocatorResponse is the response type for the Msg/DeleteOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) |  |  |






<a name="provenance-metadata-v1-MsgDeleteRecordRequest"></a>

### MsgDeleteRecordRequest
MsgDeleteRecordRequest is the request type for the Msg/DeleteRecord RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_id` | [bytes](#bytes) |  |  |
| `signers` | [string](#string) | repeated |  |






<a name="provenance-metadata-v1-MsgDeleteRecordResponse"></a>

### MsgDeleteRecordResponse
MsgDeleteRecordResponse is the response type for the Msg/DeleteRecord RPC method.






<a name="provenance-metadata-v1-MsgDeleteRecordSpecificationRequest"></a>

### MsgDeleteRecordSpecificationRequest
MsgDeleteRecordSpecificationRequest is the request type for the Msg/DeleteRecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | MetadataAddress for the record specification to delete. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance-metadata-v1-MsgDeleteRecordSpecificationResponse"></a>

### MsgDeleteRecordSpecificationResponse
MsgDeleteRecordSpecificationResponse is the response type for the Msg/DeleteRecordSpecification RPC method.






<a name="provenance-metadata-v1-MsgDeleteScopeDataAccessRequest"></a>

### MsgDeleteScopeDataAccessRequest
MsgDeleteScopeDataAccessRequest is the request to remove data access AccAddress to scope


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope MetadataAddress for removing data access |
| `data_access` | [string](#string) | repeated | AccAddress address to be removed from scope |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |






<a name="provenance-metadata-v1-MsgDeleteScopeDataAccessResponse"></a>

### MsgDeleteScopeDataAccessResponse
MsgDeleteScopeDataAccessResponse is the response from removing data access AccAddress to scope






<a name="provenance-metadata-v1-MsgDeleteScopeOwnerRequest"></a>

### MsgDeleteScopeOwnerRequest
MsgDeleteScopeOwnerRequest is the request to remove owner AccAddresses to scope


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope MetadataAddress for removing data access |
| `owners` | [string](#string) | repeated | AccAddress owner addresses to be removed from scope |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |






<a name="provenance-metadata-v1-MsgDeleteScopeOwnerResponse"></a>

### MsgDeleteScopeOwnerResponse
MsgDeleteScopeOwnerResponse is the response from removing owner AccAddress to scope






<a name="provenance-metadata-v1-MsgDeleteScopeRequest"></a>

### MsgDeleteScopeRequest
MsgDeleteScopeRequest is the request type for the Msg/DeleteScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | Unique ID for the scope to delete |
| `signers` | [string](#string) | repeated |  |






<a name="provenance-metadata-v1-MsgDeleteScopeResponse"></a>

### MsgDeleteScopeResponse
MsgDeleteScopeResponse is the response type for the Msg/DeleteScope RPC method.






<a name="provenance-metadata-v1-MsgDeleteScopeSpecificationRequest"></a>

### MsgDeleteScopeSpecificationRequest
MsgDeleteScopeSpecificationRequest is the request type for the Msg/DeleteScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | MetadataAddress for the scope specification to delete. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance-metadata-v1-MsgDeleteScopeSpecificationResponse"></a>

### MsgDeleteScopeSpecificationResponse
MsgDeleteScopeSpecificationResponse is the response type for the Msg/DeleteScopeSpecification RPC method.






<a name="provenance-metadata-v1-MsgMigrateValueOwnerRequest"></a>

### MsgMigrateValueOwnerRequest
MsgMigrateValueOwnerRequest is the request to migrate all scopes with one value owner to another value owner.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `existing` | [string](#string) |  | existing is the value owner address that is being migrated. |
| `proposed` | [string](#string) |  | proposed is the new value owner address for all of existing's scopes. |
| `signers` | [string](#string) | repeated | signers is the list of addresses of those signing this request. |






<a name="provenance-metadata-v1-MsgMigrateValueOwnerResponse"></a>

### MsgMigrateValueOwnerResponse
MsgMigrateValueOwnerResponse is the response from migrating a value owner address.






<a name="provenance-metadata-v1-MsgModifyOSLocatorRequest"></a>

### MsgModifyOSLocatorRequest
MsgModifyOSLocatorRequest is the request type for the Msg/ModifyOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) |  | The object locator to bind the address to bind to the URI. |






<a name="provenance-metadata-v1-MsgModifyOSLocatorResponse"></a>

### MsgModifyOSLocatorResponse
MsgModifyOSLocatorResponse is the response type for the Msg/ModifyOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) |  |  |






<a name="provenance-metadata-v1-MsgP8eMemorializeContractRequest"></a>

### MsgP8eMemorializeContractRequest
MsgP8eMemorializeContractRequest  has been deprecated and is no longer usable.
Deprecated: This message is no longer part of any endpoint and cannot be used for anything.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [string](#string) |  |  |
| `group_id` | [string](#string) |  |  |
| `scope_specification_id` | [string](#string) |  |  |
| `recitals` | [p8e.Recitals](#provenance-metadata-v1-p8e-Recitals) |  |  |
| `contract` | [p8e.Contract](#provenance-metadata-v1-p8e-Contract) |  |  |
| `signatures` | [p8e.SignatureSet](#provenance-metadata-v1-p8e-SignatureSet) |  |  |
| `invoker` | [string](#string) |  |  |






<a name="provenance-metadata-v1-MsgP8eMemorializeContractResponse"></a>

### MsgP8eMemorializeContractResponse
MsgP8eMemorializeContractResponse  has been deprecated and is no longer usable.
Deprecated: This message is no longer part of any endpoint and cannot be used for anything.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id_info` | [ScopeIdInfo](#provenance-metadata-v1-ScopeIdInfo) |  |  |
| `session_id_info` | [SessionIdInfo](#provenance-metadata-v1-SessionIdInfo) |  |  |
| `record_id_infos` | [RecordIdInfo](#provenance-metadata-v1-RecordIdInfo) | repeated |  |






<a name="provenance-metadata-v1-MsgSetAccountDataRequest"></a>

### MsgSetAccountDataRequest
MsgSetAccountDataRequest is the request to set/update/delete a scope's account data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata_addr` | [bytes](#bytes) |  | The identifier to associate the data with. Currently, only scope ids are supported. |
| `value` | [string](#string) |  | The desired accountdata value. |
| `signers` | [string](#string) | repeated | The signers of this message. Must fulfill owner requirements of the scope. |






<a name="provenance-metadata-v1-MsgSetAccountDataResponse"></a>

### MsgSetAccountDataResponse
MsgSetAccountDataResponse is the response from setting/updating/deleting a scope's account data.






<a name="provenance-metadata-v1-MsgUpdateValueOwnersRequest"></a>

### MsgUpdateValueOwnersRequest
MsgUpdateValueOwnersRequest is the request to update the value owner addresses in one or more scopes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_ids` | [bytes](#bytes) | repeated | scope_ids are the scope metadata addresses of all scopes to be updated. |
| `value_owner_address` | [string](#string) |  | value_owner_address is the address of the new value owner for the provided scopes. |
| `signers` | [string](#string) | repeated | signers is the list of addresses of those signing this request. |






<a name="provenance-metadata-v1-MsgUpdateValueOwnersResponse"></a>

### MsgUpdateValueOwnersResponse
MsgUpdateValueOwnersResponse is the response from updating value owner addresses in one or more scopes.






<a name="provenance-metadata-v1-MsgWriteContractSpecificationRequest"></a>

### MsgWriteContractSpecificationRequest
MsgWriteContractSpecificationRequest is the request type for the Msg/WriteContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [ContractSpecification](#provenance-metadata-v1-ContractSpecification) |  | specification is the ContractSpecification you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `spec_uuid` | [string](#string) |  | spec_uuid is an optional contract specification uuid string, e.g. "def6bc0a-c9dd-4874-948f-5206e6060a84" If provided, it will be used to generate the MetadataAddress for the contract specification which will override the specification_id in the provided specification. If not provided (or it is an empty string), nothing special happens. If there is a value in specification.specification_id that is different from the one created from this uuid, an error is returned. |






<a name="provenance-metadata-v1-MsgWriteContractSpecificationResponse"></a>

### MsgWriteContractSpecificationResponse
MsgWriteContractSpecificationResponse is the response type for the Msg/WriteContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance-metadata-v1-ContractSpecIdInfo) |  | contract_spec_id_info contains information about the id/address of the contract specification that was added or updated. |






<a name="provenance-metadata-v1-MsgWriteP8eContractSpecRequest"></a>

### MsgWriteP8eContractSpecRequest
MsgWriteP8eContractSpecRequest has been deprecated and is no longer usable.
Deprecated: This message is no longer part of any endpoint and cannot be used for anything.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contractspec` | [p8e.ContractSpec](#provenance-metadata-v1-p8e-ContractSpec) |  |  |
| `signers` | [string](#string) | repeated |  |






<a name="provenance-metadata-v1-MsgWriteP8eContractSpecResponse"></a>

### MsgWriteP8eContractSpecResponse
MsgWriteP8eContractSpecResponse  has been deprecated and is no longer usable.
Deprecated: This message is no longer part of any endpoint and cannot be used for anything.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance-metadata-v1-ContractSpecIdInfo) |  |  |
| `record_spec_id_infos` | [RecordSpecIdInfo](#provenance-metadata-v1-RecordSpecIdInfo) | repeated |  |






<a name="provenance-metadata-v1-MsgWriteRecordRequest"></a>

### MsgWriteRecordRequest
MsgWriteRecordRequest is the request type for the Msg/WriteRecord RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record` | [Record](#provenance-metadata-v1-Record) |  | record is the Record you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `session_id_components` | [SessionIdComponents](#provenance-metadata-v1-SessionIdComponents) |  | SessionIDComponents is an optional (alternate) way of defining what the session_id should be in the provided record. If provided, it must have both a scope and session_uuid. Those components will be used to create the MetadataAddress for the session which will override the session_id in the provided record. If not provided (or all empty), nothing special happens. If there is a value in record.session_id that is different from the one created from these components, an error is returned. |
| `contract_spec_uuid` | [string](#string) |  | contract_spec_uuid is an optional contract specification uuid string, e.g. "def6bc0a-c9dd-4874-948f-5206e6060a84" If provided, it will be combined with the record name to generate the MetadataAddress for the record specification which will override the specification_id in the provided record. If not provided (or it is an empty string), nothing special happens. If there is a value in record.specification_id that is different from the one created from this uuid and record.name, an error is returned. |
| `parties` | [Party](#provenance-metadata-v1-Party) | repeated | parties is the list of parties involved with this record. Deprecated: This field is ignored. The parties are identified in the session and as signers. |






<a name="provenance-metadata-v1-MsgWriteRecordResponse"></a>

### MsgWriteRecordResponse
MsgWriteRecordResponse is the response type for the Msg/WriteRecord RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_id_info` | [RecordIdInfo](#provenance-metadata-v1-RecordIdInfo) |  | record_id_info contains information about the id/address of the record that was added or updated. |






<a name="provenance-metadata-v1-MsgWriteRecordSpecificationRequest"></a>

### MsgWriteRecordSpecificationRequest
MsgWriteRecordSpecificationRequest is the request type for the Msg/WriteRecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [RecordSpecification](#provenance-metadata-v1-RecordSpecification) |  | specification is the RecordSpecification you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `contract_spec_uuid` | [string](#string) |  | contract_spec_uuid is an optional contract specification uuid string, e.g. "def6bc0a-c9dd-4874-948f-5206e6060a84" If provided, it will be combined with the record specification name to generate the MetadataAddress for the record specification which will override the specification_id in the provided specification. If not provided (or it is an empty string), nothing special happens. If there is a value in specification.specification_id that is different from the one created from this uuid and specification.name, an error is returned. |






<a name="provenance-metadata-v1-MsgWriteRecordSpecificationResponse"></a>

### MsgWriteRecordSpecificationResponse
MsgWriteRecordSpecificationResponse is the response type for the Msg/WriteRecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_spec_id_info` | [RecordSpecIdInfo](#provenance-metadata-v1-RecordSpecIdInfo) |  | record_spec_id_info contains information about the id/address of the record specification that was added or updated. |






<a name="provenance-metadata-v1-MsgWriteScopeRequest"></a>

### MsgWriteScopeRequest
MsgWriteScopeRequest is the request type for the Msg/WriteScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [Scope](#provenance-metadata-v1-Scope) |  | scope is the Scope you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `scope_uuid` | [string](#string) |  | scope_uuid is an optional uuid string, e.g. "91978ba2-5f35-459a-86a7-feca1b0512e0" If provided, it will be used to generate the MetadataAddress for the scope which will override the scope_id in the provided scope. If not provided (or it is an empty string), nothing special happens. If there is a value in scope.scope_id that is different from the one created from this uuid, an error is returned. |
| `spec_uuid` | [string](#string) |  | spec_uuid is an optional scope specification uuid string, e.g. "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2" If provided, it will be used to generate the MetadataAddress for the scope specification which will override the specification_id in the provided scope. If not provided (or it is an empty string), nothing special happens. If there is a value in scope.specification_id that is different from the one created from this uuid, an error is returned. |
| `usd_mills` | [uint64](#uint64) |  | usd_mills value of scope in usd mills (1234 = $1.234) used for net asset value |






<a name="provenance-metadata-v1-MsgWriteScopeResponse"></a>

### MsgWriteScopeResponse
MsgWriteScopeResponse is the response type for the Msg/WriteScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id_info` | [ScopeIdInfo](#provenance-metadata-v1-ScopeIdInfo) |  | scope_id_info contains information about the id/address of the scope that was added or updated. |






<a name="provenance-metadata-v1-MsgWriteScopeSpecificationRequest"></a>

### MsgWriteScopeSpecificationRequest
MsgWriteScopeSpecificationRequest is the request type for the Msg/WriteScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [ScopeSpecification](#provenance-metadata-v1-ScopeSpecification) |  | specification is the ScopeSpecification you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `spec_uuid` | [string](#string) |  | spec_uuid is an optional scope specification uuid string, e.g. "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2" If provided, it will be used to generate the MetadataAddress for the scope specification which will override the specification_id in the provided specification. If not provided (or it is an empty string), nothing special happens. If there is a value in specification.specification_id that is different from the one created from this uuid, an error is returned. |






<a name="provenance-metadata-v1-MsgWriteScopeSpecificationResponse"></a>

### MsgWriteScopeSpecificationResponse
MsgWriteScopeSpecificationResponse is the response type for the Msg/WriteScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_spec_id_info` | [ScopeSpecIdInfo](#provenance-metadata-v1-ScopeSpecIdInfo) |  | scope_spec_id_info contains information about the id/address of the scope specification that was added or updated. |






<a name="provenance-metadata-v1-MsgWriteSessionRequest"></a>

### MsgWriteSessionRequest
MsgWriteSessionRequest is the request type for the Msg/WriteSession RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session` | [Session](#provenance-metadata-v1-Session) |  | session is the Session you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `session_id_components` | [SessionIdComponents](#provenance-metadata-v1-SessionIdComponents) |  | SessionIDComponents is an optional (alternate) way of defining what the session_id should be in the provided session. If provided, it must have both a scope and session_uuid. Those components will be used to create the MetadataAddress for the session which will override the session_id in the provided session. If not provided (or all empty), nothing special happens. If there is a value in session.session_id that is different from the one created from these components, an error is returned. |
| `spec_uuid` | [string](#string) |  | spec_uuid is an optional contract specification uuid string, e.g. "def6bc0a-c9dd-4874-948f-5206e6060a84" If provided, it will be used to generate the MetadataAddress for the contract specification which will override the specification_id in the provided session. If not provided (or it is an empty string), nothing special happens. If there is a value in session.specification_id that is different from the one created from this uuid, an error is returned. |






<a name="provenance-metadata-v1-MsgWriteSessionResponse"></a>

### MsgWriteSessionResponse
MsgWriteSessionResponse is the response type for the Msg/WriteSession RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_id_info` | [SessionIdInfo](#provenance-metadata-v1-SessionIdInfo) |  | session_id_info contains information about the id/address of the session that was added or updated. |






<a name="provenance-metadata-v1-SessionIdComponents"></a>

### SessionIdComponents
SessionIDComponents contains fields for the components that make up a session id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_uuid` | [string](#string) |  | scope_uuid is the uuid string for the scope, e.g. "91978ba2-5f35-459a-86a7-feca1b0512e0" |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string for the scope, g.g. "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel" |
| `session_uuid` | [string](#string) |  | session_uuid is a uuid string for identifying this session, e.g. "5803f8bc-6067-4eb5-951f-2121671c2ec0" |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-metadata-v1-Msg"></a>

### Msg
Msg defines the Metadata Msg service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `WriteScope` | [MsgWriteScopeRequest](#provenance-metadata-v1-MsgWriteScopeRequest) | [MsgWriteScopeResponse](#provenance-metadata-v1-MsgWriteScopeResponse) | WriteScope adds or updates a scope. |
| `DeleteScope` | [MsgDeleteScopeRequest](#provenance-metadata-v1-MsgDeleteScopeRequest) | [MsgDeleteScopeResponse](#provenance-metadata-v1-MsgDeleteScopeResponse) | DeleteScope deletes a scope and all associated Records, Sessions. |
| `AddScopeDataAccess` | [MsgAddScopeDataAccessRequest](#provenance-metadata-v1-MsgAddScopeDataAccessRequest) | [MsgAddScopeDataAccessResponse](#provenance-metadata-v1-MsgAddScopeDataAccessResponse) | AddScopeDataAccess adds data access AccAddress to scope |
| `DeleteScopeDataAccess` | [MsgDeleteScopeDataAccessRequest](#provenance-metadata-v1-MsgDeleteScopeDataAccessRequest) | [MsgDeleteScopeDataAccessResponse](#provenance-metadata-v1-MsgDeleteScopeDataAccessResponse) | DeleteScopeDataAccess removes data access AccAddress from scope |
| `AddScopeOwner` | [MsgAddScopeOwnerRequest](#provenance-metadata-v1-MsgAddScopeOwnerRequest) | [MsgAddScopeOwnerResponse](#provenance-metadata-v1-MsgAddScopeOwnerResponse) | AddScopeOwner adds new owner parties to a scope |
| `DeleteScopeOwner` | [MsgDeleteScopeOwnerRequest](#provenance-metadata-v1-MsgDeleteScopeOwnerRequest) | [MsgDeleteScopeOwnerResponse](#provenance-metadata-v1-MsgDeleteScopeOwnerResponse) | DeleteScopeOwner removes owner parties (by addresses) from a scope |
| `UpdateValueOwners` | [MsgUpdateValueOwnersRequest](#provenance-metadata-v1-MsgUpdateValueOwnersRequest) | [MsgUpdateValueOwnersResponse](#provenance-metadata-v1-MsgUpdateValueOwnersResponse) | UpdateValueOwners sets the value owner of one or more scopes. |
| `MigrateValueOwner` | [MsgMigrateValueOwnerRequest](#provenance-metadata-v1-MsgMigrateValueOwnerRequest) | [MsgMigrateValueOwnerResponse](#provenance-metadata-v1-MsgMigrateValueOwnerResponse) | MigrateValueOwner updates all scopes that have one value owner to have a another value owner. |
| `WriteSession` | [MsgWriteSessionRequest](#provenance-metadata-v1-MsgWriteSessionRequest) | [MsgWriteSessionResponse](#provenance-metadata-v1-MsgWriteSessionResponse) | WriteSession adds or updates a session context. |
| `WriteRecord` | [MsgWriteRecordRequest](#provenance-metadata-v1-MsgWriteRecordRequest) | [MsgWriteRecordResponse](#provenance-metadata-v1-MsgWriteRecordResponse) | WriteRecord adds or updates a record. |
| `DeleteRecord` | [MsgDeleteRecordRequest](#provenance-metadata-v1-MsgDeleteRecordRequest) | [MsgDeleteRecordResponse](#provenance-metadata-v1-MsgDeleteRecordResponse) | DeleteRecord deletes a record. |
| `WriteScopeSpecification` | [MsgWriteScopeSpecificationRequest](#provenance-metadata-v1-MsgWriteScopeSpecificationRequest) | [MsgWriteScopeSpecificationResponse](#provenance-metadata-v1-MsgWriteScopeSpecificationResponse) | WriteScopeSpecification adds or updates a scope specification. |
| `DeleteScopeSpecification` | [MsgDeleteScopeSpecificationRequest](#provenance-metadata-v1-MsgDeleteScopeSpecificationRequest) | [MsgDeleteScopeSpecificationResponse](#provenance-metadata-v1-MsgDeleteScopeSpecificationResponse) | DeleteScopeSpecification deletes a scope specification. |
| `WriteContractSpecification` | [MsgWriteContractSpecificationRequest](#provenance-metadata-v1-MsgWriteContractSpecificationRequest) | [MsgWriteContractSpecificationResponse](#provenance-metadata-v1-MsgWriteContractSpecificationResponse) | WriteContractSpecification adds or updates a contract specification. |
| `DeleteContractSpecification` | [MsgDeleteContractSpecificationRequest](#provenance-metadata-v1-MsgDeleteContractSpecificationRequest) | [MsgDeleteContractSpecificationResponse](#provenance-metadata-v1-MsgDeleteContractSpecificationResponse) | DeleteContractSpecification deletes a contract specification. |
| `AddContractSpecToScopeSpec` | [MsgAddContractSpecToScopeSpecRequest](#provenance-metadata-v1-MsgAddContractSpecToScopeSpecRequest) | [MsgAddContractSpecToScopeSpecResponse](#provenance-metadata-v1-MsgAddContractSpecToScopeSpecResponse) | AddContractSpecToScopeSpec adds contract specification to a scope specification. |
| `DeleteContractSpecFromScopeSpec` | [MsgDeleteContractSpecFromScopeSpecRequest](#provenance-metadata-v1-MsgDeleteContractSpecFromScopeSpecRequest) | [MsgDeleteContractSpecFromScopeSpecResponse](#provenance-metadata-v1-MsgDeleteContractSpecFromScopeSpecResponse) | DeleteContractSpecFromScopeSpec deletes a contract specification from a scope specification. |
| `WriteRecordSpecification` | [MsgWriteRecordSpecificationRequest](#provenance-metadata-v1-MsgWriteRecordSpecificationRequest) | [MsgWriteRecordSpecificationResponse](#provenance-metadata-v1-MsgWriteRecordSpecificationResponse) | WriteRecordSpecification adds or updates a record specification. |
| `DeleteRecordSpecification` | [MsgDeleteRecordSpecificationRequest](#provenance-metadata-v1-MsgDeleteRecordSpecificationRequest) | [MsgDeleteRecordSpecificationResponse](#provenance-metadata-v1-MsgDeleteRecordSpecificationResponse) | DeleteRecordSpecification deletes a record specification. |
| `BindOSLocator` | [MsgBindOSLocatorRequest](#provenance-metadata-v1-MsgBindOSLocatorRequest) | [MsgBindOSLocatorResponse](#provenance-metadata-v1-MsgBindOSLocatorResponse) | BindOSLocator binds an owner address to a uri. |
| `DeleteOSLocator` | [MsgDeleteOSLocatorRequest](#provenance-metadata-v1-MsgDeleteOSLocatorRequest) | [MsgDeleteOSLocatorResponse](#provenance-metadata-v1-MsgDeleteOSLocatorResponse) | DeleteOSLocator deletes an existing ObjectStoreLocator record. |
| `ModifyOSLocator` | [MsgModifyOSLocatorRequest](#provenance-metadata-v1-MsgModifyOSLocatorRequest) | [MsgModifyOSLocatorResponse](#provenance-metadata-v1-MsgModifyOSLocatorResponse) | ModifyOSLocator updates an ObjectStoreLocator record by the current owner. |
| `SetAccountData` | [MsgSetAccountDataRequest](#provenance-metadata-v1-MsgSetAccountDataRequest) | [MsgSetAccountDataResponse](#provenance-metadata-v1-MsgSetAccountDataResponse) | SetAccountData associates some basic data with a metadata address. Currently, only scope ids are supported. |
| `AddNetAssetValues` | [MsgAddNetAssetValuesRequest](#provenance-metadata-v1-MsgAddNetAssetValuesRequest) | [MsgAddNetAssetValuesResponse](#provenance-metadata-v1-MsgAddNetAssetValuesResponse) | AddNetAssetValues set the net asset value for a scope |

 <!-- end services -->



<a name="provenance_metadata_v1_events-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/events.proto



<a name="provenance-metadata-v1-EventContractSpecificationCreated"></a>

### EventContractSpecificationCreated
EventContractSpecificationCreated is an event message indicating a contract specification has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the specification id of the contract specification that was created. |






<a name="provenance-metadata-v1-EventContractSpecificationDeleted"></a>

### EventContractSpecificationDeleted
EventContractSpecificationDeleted is an event message indicating a contract specification has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the specification id of the contract specification that was deleted. |






<a name="provenance-metadata-v1-EventContractSpecificationUpdated"></a>

### EventContractSpecificationUpdated
EventContractSpecificationUpdated is an event message indicating a contract specification has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the specification id of the contract specification that was updated. |






<a name="provenance-metadata-v1-EventOSLocatorCreated"></a>

### EventOSLocatorCreated
EventOSLocatorCreated is an event message indicating an object store locator has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | owner is the owner in the object store locator that was created. |






<a name="provenance-metadata-v1-EventOSLocatorDeleted"></a>

### EventOSLocatorDeleted
EventOSLocatorDeleted is an event message indicating an object store locator has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | owner is the owner in the object store locator that was deleted. |






<a name="provenance-metadata-v1-EventOSLocatorUpdated"></a>

### EventOSLocatorUpdated
EventOSLocatorUpdated is an event message indicating an object store locator has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | owner is the owner in the object store locator that was updated. |






<a name="provenance-metadata-v1-EventRecordCreated"></a>

### EventRecordCreated
EventRecordCreated is an event message indicating a record has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_addr` | [string](#string) |  | record_addr is the bech32 address string of the record id that was created. |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id this record belongs to. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this record belongs to. |






<a name="provenance-metadata-v1-EventRecordDeleted"></a>

### EventRecordDeleted
EventRecordDeleted is an event message indicating a record has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_addr` | [string](#string) |  | record is the bech32 address string of the record id that was deleted. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this record belonged to. |






<a name="provenance-metadata-v1-EventRecordSpecificationCreated"></a>

### EventRecordSpecificationCreated
EventRecordSpecificationCreated is an event message indicating a record specification has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specification_addr` | [string](#string) |  | record_specification_addr is the bech32 address string of the specification id of the record specification that was created. |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the contract specification id this record specification belongs to. |






<a name="provenance-metadata-v1-EventRecordSpecificationDeleted"></a>

### EventRecordSpecificationDeleted
EventRecordSpecificationDeleted is an event message indicating a record specification has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specification_addr` | [string](#string) |  | record_specification_addr is the bech32 address string of the specification id of the record specification that was deleted. |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the contract specification id this record specification belongs to. |






<a name="provenance-metadata-v1-EventRecordSpecificationUpdated"></a>

### EventRecordSpecificationUpdated
EventRecordSpecificationUpdated is an event message indicating a record specification has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specification_addr` | [string](#string) |  | record_specification_addr is the bech32 address string of the specification id of the record specification that was updated. |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the contract specification id this record specification belongs to. |






<a name="provenance-metadata-v1-EventRecordUpdated"></a>

### EventRecordUpdated
EventRecordUpdated is an event message indicating a record has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_addr` | [string](#string) |  | record_addr is the bech32 address string of the record id that was updated. |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id this record belongs to. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this record belongs to. |






<a name="provenance-metadata-v1-EventScopeCreated"></a>

### EventScopeCreated
EventScopeCreated is an event message indicating a scope has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id that was created. |






<a name="provenance-metadata-v1-EventScopeDeleted"></a>

### EventScopeDeleted
EventScopeDeleted is an event message indicating a scope has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id that was deleted. |






<a name="provenance-metadata-v1-EventScopeSpecificationCreated"></a>

### EventScopeSpecificationCreated
EventScopeSpecificationCreated is an event message indicating a scope specification has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specification_addr` | [string](#string) |  | scope_specification_addr is the bech32 address string of the specification id of the scope specification that was created. |






<a name="provenance-metadata-v1-EventScopeSpecificationDeleted"></a>

### EventScopeSpecificationDeleted
EventScopeSpecificationDeleted is an event message indicating a scope specification has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specification_addr` | [string](#string) |  | scope_specification_addr is the bech32 address string of the specification id of the scope specification that was deleted. |






<a name="provenance-metadata-v1-EventScopeSpecificationUpdated"></a>

### EventScopeSpecificationUpdated
EventScopeSpecificationUpdated is an event message indicating a scope specification has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specification_addr` | [string](#string) |  | scope_specification_addr is the bech32 address string of the specification id of the scope specification that was updated. |






<a name="provenance-metadata-v1-EventScopeUpdated"></a>

### EventScopeUpdated
EventScopeUpdated is an event message indicating a scope has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id that was updated. |






<a name="provenance-metadata-v1-EventSessionCreated"></a>

### EventSessionCreated
EventSessionCreated is an event message indicating a session has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id that was created. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this session belongs to. |






<a name="provenance-metadata-v1-EventSessionDeleted"></a>

### EventSessionDeleted
EventSessionDeleted is an event message indicating a session has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id that was deleted. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this session belongs to. |






<a name="provenance-metadata-v1-EventSessionUpdated"></a>

### EventSessionUpdated
EventSessionUpdated is an event message indicating a session has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id that was updated. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this session belongs to. |






<a name="provenance-metadata-v1-EventSetNetAssetValue"></a>

### EventSetNetAssetValue
EventSetNetAssetValue event emitted when Net Asset Value for a scope is update or added


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [string](#string) |  |  |
| `price` | [string](#string) |  |  |
| `source` | [string](#string) |  |  |
| `volume` | [string](#string) |  |  |






<a name="provenance-metadata-v1-EventTxCompleted"></a>

### EventTxCompleted
EventTxCompleted is an event message indicating that a TX has completed.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `module` | [string](#string) |  | module is the module the TX belongs to. |
| `endpoint` | [string](#string) |  | endpoint is the rpc endpoint that was just completed. |
| `signers` | [string](#string) | repeated | signers are the bech32 address strings of the signers of this TX. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_metadata_v1_specification-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/specification.proto



<a name="provenance-metadata-v1-ContractSpecification"></a>

### ContractSpecification
ContractSpecification defines the required parties, resources, conditions, and consideration outputs for a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | unique identifier for this specification on chain |
| `description` | [Description](#provenance-metadata-v1-Description) |  | Description information for this contract specification |
| `owner_addresses` | [string](#string) | repeated | Address of the account that owns this specificaiton |
| `parties_involved` | [PartyType](#provenance-metadata-v1-PartyType) | repeated | a list of party roles that must be fullfilled when signing a transaction for this contract specification |
| `resource_id` | [bytes](#bytes) |  | the address of a record on chain that represents this contract |
| `hash` | [string](#string) |  | the hash of contract binary (off-chain instance) |
| `class_name` | [string](#string) |  | name of the class/type of this contract executable |






<a name="provenance-metadata-v1-Description"></a>

### Description
Description holds general information that is handy to associate with a structure.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | A Name for this thing. |
| `description` | [string](#string) |  | A description of this thing. |
| `website_url` | [string](#string) |  | URL to find even more info. |
| `icon_url` | [string](#string) |  | URL of an icon. |






<a name="provenance-metadata-v1-InputSpecification"></a>

### InputSpecification
InputSpecification defines a name, type_name, and source reference (either on or off chain) to define an input
parameter


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | name for this input |
| `type_name` | [string](#string) |  | a type_name (typically a proto name or class_name) |
| `record_id` | [bytes](#bytes) |  | the address of a record on chain (For Established Records) |
| `hash` | [string](#string) |  | the hash of an off-chain piece of information (For Proposed Records) |






<a name="provenance-metadata-v1-RecordSpecification"></a>

### RecordSpecification
RecordSpecification defines the specification for a Record including allowed/required inputs/outputs


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | unique identifier for this specification on chain |
| `name` | [string](#string) |  | Name of Record that will be created when this specification is used |
| `inputs` | [InputSpecification](#provenance-metadata-v1-InputSpecification) | repeated | A set of inputs that must be satisified to apply this RecordSpecification and create a Record |
| `type_name` | [string](#string) |  | A type name for data associated with this record (typically a class or proto name) |
| `result_type` | [DefinitionType](#provenance-metadata-v1-DefinitionType) |  | Type of result for this record specification (must be RECORD or RECORD_LIST) |
| `responsible_parties` | [PartyType](#provenance-metadata-v1-PartyType) | repeated | Type of party responsible for this record |






<a name="provenance-metadata-v1-ScopeSpecification"></a>

### ScopeSpecification
ScopeSpecification defines the required parties, resources, conditions, and consideration outputs for a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | unique identifier for this specification on chain |
| `description` | [Description](#provenance-metadata-v1-Description) |  | General information about this scope specification. |
| `owner_addresses` | [string](#string) | repeated | Addresses of the owners of this scope specification. |
| `parties_involved` | [PartyType](#provenance-metadata-v1-PartyType) | repeated | A list of parties that must be present on a scope (and their associated roles) |
| `contract_spec_ids` | [bytes](#bytes) | repeated | A list of contract specification ids allowed for a scope based on this specification. |





 <!-- end messages -->


<a name="provenance-metadata-v1-DefinitionType"></a>

### DefinitionType
DefinitionType indicates the required definition type for this value

| Name | Number | Description |
| ---- | ------ | ----------- |
| `DEFINITION_TYPE_UNSPECIFIED` | `0` | DEFINITION_TYPE_UNSPECIFIED indicates an unknown/invalid value |
| `DEFINITION_TYPE_PROPOSED` | `1` | DEFINITION_TYPE_PROPOSED indicates a proposed value is used here (a record that is not on-chain) |
| `DEFINITION_TYPE_RECORD` | `2` | DEFINITION_TYPE_RECORD indicates the value must be a reference to a record on chain |
| `DEFINITION_TYPE_RECORD_LIST` | `3` | DEFINITION_TYPE_RECORD_LIST indicates the value maybe a reference to a collection of values on chain having the same name |



<a name="provenance-metadata-v1-PartyType"></a>

### PartyType
PartyType are the different roles parties on a contract may use

| Name | Number | Description |
| ---- | ------ | ----------- |
| `PARTY_TYPE_UNSPECIFIED` | `0` | PARTY_TYPE_UNSPECIFIED is an error condition |
| `PARTY_TYPE_ORIGINATOR` | `1` | PARTY_TYPE_ORIGINATOR is an asset originator |
| `PARTY_TYPE_SERVICER` | `2` | PARTY_TYPE_SERVICER provides debt servicing functions |
| `PARTY_TYPE_INVESTOR` | `3` | PARTY_TYPE_INVESTOR is a generic investor |
| `PARTY_TYPE_CUSTODIAN` | `4` | PARTY_TYPE_CUSTODIAN is an entity that provides custodian services for assets |
| `PARTY_TYPE_OWNER` | `5` | PARTY_TYPE_OWNER indicates this party is an owner of the item |
| `PARTY_TYPE_AFFILIATE` | `6` | PARTY_TYPE_AFFILIATE is a party with an affiliate agreement |
| `PARTY_TYPE_OMNIBUS` | `7` | PARTY_TYPE_OMNIBUS is a special type of party that controls an omnibus bank account |
| `PARTY_TYPE_PROVENANCE` | `8` | PARTY_TYPE_PROVENANCE is used to indicate this party represents the blockchain or a smart contract action |
| `PARTY_TYPE_CONTROLLER` | `10` | PARTY_TYPE_CONTROLLER is an entity which controls a specific asset on chain (ie enote) |
| `PARTY_TYPE_VALIDATOR` | `11` | PARTY_TYPE_VALIDATOR is an entity which validates given assets on chain |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_metadata_v1_scope-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/scope.proto



<a name="provenance-metadata-v1-AuditFields"></a>

### AuditFields
AuditFields capture information about the last account to make modifications and when they were made


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `created_date` | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | the date/time when this entry was created |
| `created_by` | [string](#string) |  | the address of the account that created this record |
| `updated_date` | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | the date/time when this entry was last updated |
| `updated_by` | [string](#string) |  | the address of the account that modified this record |
| `version` | [uint32](#uint32) |  | an optional version number that is incremented with each update |
| `message` | [string](#string) |  | an optional message associated with the creation/update event |






<a name="provenance-metadata-v1-NetAssetValue"></a>

### NetAssetValue
NetAssetValue defines a scope's net asset value


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `price` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) |  | price is the complete value of the asset's volume |
| `updated_block_height` | [uint64](#uint64) |  | updated_block_height is the block height of last update |
| `volume` | [uint64](#uint64) |  | volume is the number of scope instances that were purchased for the price Typically this will be null (equivalent to one) or one. The only reason this would be more than one is for cases where the precision of the price denom is insufficient to represent the actual price |






<a name="provenance-metadata-v1-Party"></a>

### Party
A Party is an address with/in a given role associated with a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address of the account (on chain) |
| `role` | [PartyType](#provenance-metadata-v1-PartyType) |  | a role for this account within the context of the processes used |
| `optional` | [bool](#bool) |  | whether this party's signature is optional |






<a name="provenance-metadata-v1-Process"></a>

### Process
Process contains information used to uniquely identify what was used to generate this record


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | the address of a smart contract used for this process |
| `hash` | [string](#string) |  | the hash of an off-chain process used |
| `name` | [string](#string) |  | a name associated with the process (type_name, classname or smart contract common name) |
| `method` | [string](#string) |  | method is a name or reference to a specific operation (method) within a class/contract that was invoked |






<a name="provenance-metadata-v1-Record"></a>

### Record
A record (of fact) is attached to a session or each consideration output from a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | name/identifier for this record. Value must be unique within the scope. Also known as a Fact name |
| `session_id` | [bytes](#bytes) |  | id of the session context that was used to create this record (use with filtered kvprefix iterator) |
| `process` | [Process](#provenance-metadata-v1-Process) |  | process contain information used to uniquely identify an execution on or off chain that generated this record |
| `inputs` | [RecordInput](#provenance-metadata-v1-RecordInput) | repeated | inputs used with the process to achieve the output on this record |
| `outputs` | [RecordOutput](#provenance-metadata-v1-RecordOutput) | repeated | output(s) is the results of executing the process on the given process indicated in this record |
| `specification_id` | [bytes](#bytes) |  | specification_id is the id of the record specification that was used to create this record. |






<a name="provenance-metadata-v1-RecordInput"></a>

### RecordInput
Tracks the inputs used to establish this record


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | Name value included to link back to the definition spec. |
| `record_id` | [bytes](#bytes) |  | the address of a record on chain (For Established Records) |
| `hash` | [string](#string) |  | the hash of an off-chain piece of information (For Proposed Records) |
| `type_name` | [string](#string) |  | from proposed fact structure to unmarshal |
| `status` | [RecordInputStatus](#provenance-metadata-v1-RecordInputStatus) |  | Indicates if this input was a recorded fact on chain or just a given hashed input |






<a name="provenance-metadata-v1-RecordOutput"></a>

### RecordOutput
RecordOutput encapsulates the output of a process recorded on chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `hash` | [string](#string) |  | Hash of the data output that was output/generated for this record |
| `status` | [ResultStatus](#provenance-metadata-v1-ResultStatus) |  | Status of the process execution associated with this output indicating success,failure, or pending |






<a name="provenance-metadata-v1-Scope"></a>

### Scope
Scope defines a root reference for a collection of records owned by one or more parties.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | Unique ID for this scope. Implements sdk.Address interface for use where addresses are required in Cosmos |
| `specification_id` | [bytes](#bytes) |  | the scope specification that contains the specifications for data elements allowed within this scope |
| `owners` | [Party](#provenance-metadata-v1-Party) | repeated | These parties represent top level owners of the records within. These parties must sign any requests that modify the data within the scope. These addresses are in union with parties listed on the sessions. |
| `data_access` | [string](#string) | repeated | Addresses in this list are authorized to receive off-chain data associated with this scope. |
| `value_owner_address` | [string](#string) |  | The address that controls the value associated with this scope.<br>The value owner is actually tracked by the bank module using a coin with the denom "nft/<scope_id>". The value owner can be changed using WriteScope or anything that transfers funds, e.g. MsgSend.<br>During WriteScope: - If this field is empty, it indicates that there should not be a change to the value owner. I.e. Once a scope has a value owner, it will always have one (until it's deleted). - If this field has a value, the existing value owner will be looked up, and - If there's already an existing value owner, they must be a signer, and the coin will be transferred to the new value owner. - If there isn't yet a value owner, the coin will be minted and sent to the new value owner. If the scope already exists, the owners must be signers (just like changing other fields). If it's a new scope, there's no special signer limitations related to the value owner. |
| `require_party_rollup` | [bool](#bool) |  | Whether all parties in this scope and its sessions must be present in this scope's owners field. This also enables use of optional=true scope owners and session parties. |






<a name="provenance-metadata-v1-Session"></a>

### Session
Session defines an execution context against a specific specification instance.
The context will have a specification and set of parties involved.

NOTE: When there are no more Records within a Scope that reference a Session, the Session is removed.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_id` | [bytes](#bytes) |  |  |
| `specification_id` | [bytes](#bytes) |  | unique id of the contract specification that was used to create this session. |
| `parties` | [Party](#provenance-metadata-v1-Party) | repeated | parties is the set of identities that signed this contract |
| `name` | [string](#string) |  | name to associate with this session execution context, typically classname |
| `context` | [bytes](#bytes) |  | context is a field for storing client specific data associated with a session. |
| `audit` | [AuditFields](#provenance-metadata-v1-AuditFields) |  | Created by, updated by, timestamps, version number, and related info. |





 <!-- end messages -->


<a name="provenance-metadata-v1-RecordInputStatus"></a>

### RecordInputStatus
A set of types for inputs on a record (of fact)

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RECORD_INPUT_STATUS_UNSPECIFIED` | `0` | RECORD_INPUT_STATUS_UNSPECIFIED indicates an invalid/unknown input type |
| `RECORD_INPUT_STATUS_PROPOSED` | `1` | RECORD_INPUT_STATUS_PROPOSED indicates this input was an arbitrary piece of data that was hashed |
| `RECORD_INPUT_STATUS_RECORD` | `2` | RECORD_INPUT_STATUS_RECORD indicates this input is a reference to a previously recorded fact on blockchain |



<a name="provenance-metadata-v1-ResultStatus"></a>

### ResultStatus
ResultStatus indicates the various states of execution of a record

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RESULT_STATUS_UNSPECIFIED` | `0` | RESULT_STATUS_UNSPECIFIED indicates an unset condition |
| `RESULT_STATUS_PASS` | `1` | RESULT_STATUS_PASS indicates the execution was successful |
| `RESULT_STATUS_SKIP` | `2` | RESULT_STATUS_SKIP indicates condition/consideration was skipped due to missing inputs or delayed execution |
| `RESULT_STATUS_FAIL` | `3` | RESULT_STATUS_FAIL indicates the execution of the condition/consideration failed. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_metadata_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/query.proto



<a name="provenance-metadata-v1-AccountDataRequest"></a>

### AccountDataRequest
AccountDataRequest is the request type for the Query/AccountData RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata_addr` | [bytes](#bytes) |  | The metadata address to look up. Currently, only scope ids are supported. |






<a name="provenance-metadata-v1-AccountDataResponse"></a>

### AccountDataResponse
AccountDataResponse is the response type for the Query/AccountData RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  | The accountdata for the requested metadata address. |






<a name="provenance-metadata-v1-ContractSpecificationRequest"></a>

### ContractSpecificationRequest
ContractSpecificationRequest is the request type for the Query/ContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [string](#string) |  | specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84 or a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn. It can also be a record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. |
| `include_record_specs` | [bool](#bool) |  | include_record_specs is a flag for whether to include the the record specifications of this contract specification in the response. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-ContractSpecificationResponse"></a>

### ContractSpecificationResponse
ContractSpecificationResponse is the response type for the Query/ContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification` | [ContractSpecificationWrapper](#provenance-metadata-v1-ContractSpecificationWrapper) |  | contract_specification is the wrapped contract specification. |
| `record_specifications` | [RecordSpecificationWrapper](#provenance-metadata-v1-RecordSpecificationWrapper) | repeated | record_specifications is any number or wrapped record specifications associated with this contract_specification (if requested). |
| `request` | [ContractSpecificationRequest](#provenance-metadata-v1-ContractSpecificationRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-ContractSpecificationWrapper"></a>

### ContractSpecificationWrapper
ContractSpecificationWrapper contains a single contract specification and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [ContractSpecification](#provenance-metadata-v1-ContractSpecification) |  | specification is the on-chain contract specification message. |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance-metadata-v1-ContractSpecIdInfo) |  | contract_spec_id_info contains information about the id/address of the contract specification. |






<a name="provenance-metadata-v1-ContractSpecificationsAllRequest"></a>

### ContractSpecificationsAllRequest
ContractSpecificationsAllRequest is the request type for the Query/ContractSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-ContractSpecificationsAllResponse"></a>

### ContractSpecificationsAllResponse
ContractSpecificationsAllResponse is the response type for the Query/ContractSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specifications` | [ContractSpecificationWrapper](#provenance-metadata-v1-ContractSpecificationWrapper) | repeated | contract_specifications are the wrapped contract specifications. |
| `request` | [ContractSpecificationsAllRequest](#provenance-metadata-v1-ContractSpecificationsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance-metadata-v1-GetByAddrRequest"></a>

### GetByAddrRequest
GetByAddrRequest is the request type for the Query/GetByAddr RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addrs` | [string](#string) | repeated | ids are the metadata addresses of the things to look up. |






<a name="provenance-metadata-v1-GetByAddrResponse"></a>

### GetByAddrResponse
GetByAddrResponse is the response type for the Query/GetByAddr RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scopes` | [Scope](#provenance-metadata-v1-Scope) | repeated | scopes contains any scopes that were requested and found. |
| `sessions` | [Session](#provenance-metadata-v1-Session) | repeated | sessions contains any sessions that were requested and found. |
| `records` | [Record](#provenance-metadata-v1-Record) | repeated | records contains any records that were requested and found. |
| `scope_specs` | [ScopeSpecification](#provenance-metadata-v1-ScopeSpecification) | repeated | scope_specs contains any scope specifications that were requested and found. |
| `contract_specs` | [ContractSpecification](#provenance-metadata-v1-ContractSpecification) | repeated | contract_specs contains any contract specifications that were requested and found. |
| `record_specs` | [RecordSpecification](#provenance-metadata-v1-RecordSpecification) | repeated | record_specs contains any record specifications that were requested and found. |
| `not_found` | [string](#string) | repeated | not_found contains any addrs requested but not found. |






<a name="provenance-metadata-v1-OSAllLocatorsRequest"></a>

### OSAllLocatorsRequest
OSAllLocatorsRequest is the request type for the Query/OSAllLocators RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-OSAllLocatorsResponse"></a>

### OSAllLocatorsResponse
OSAllLocatorsResponse is the response type for the Query/OSAllLocators RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locators` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) | repeated |  |
| `request` | [OSAllLocatorsRequest](#provenance-metadata-v1-OSAllLocatorsRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance-metadata-v1-OSLocatorParamsRequest"></a>

### OSLocatorParamsRequest
OSLocatorParamsRequest is the request type for the Query/OSLocatorParams RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-OSLocatorParamsResponse"></a>

### OSLocatorParamsResponse
OSLocatorParamsResponse is the response type for the Query/OSLocatorParams RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [OSLocatorParams](#provenance-metadata-v1-OSLocatorParams) |  | params defines the parameters of the module. |
| `request` | [OSLocatorParamsRequest](#provenance-metadata-v1-OSLocatorParamsRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-OSLocatorRequest"></a>

### OSLocatorRequest
OSLocatorRequest is the request type for the Query/OSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-OSLocatorResponse"></a>

### OSLocatorResponse
OSLocatorResponse is the response type for the Query/OSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) |  |  |
| `request` | [OSLocatorRequest](#provenance-metadata-v1-OSLocatorRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-OSLocatorsByScopeRequest"></a>

### OSLocatorsByScopeRequest
OSLocatorsByScopeRequest is the request type for the Query/OSLocatorsByScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-OSLocatorsByScopeResponse"></a>

### OSLocatorsByScopeResponse
OSLocatorsByScopeResponse is the response type for the Query/OSLocatorsByScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locators` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) | repeated |  |
| `request` | [OSLocatorsByScopeRequest](#provenance-metadata-v1-OSLocatorsByScopeRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-OSLocatorsByURIRequest"></a>

### OSLocatorsByURIRequest
OSLocatorsByURIRequest is the request type for the Query/OSLocatorsByURI RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `uri` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-OSLocatorsByURIResponse"></a>

### OSLocatorsByURIResponse
OSLocatorsByURIResponse is the response type for the Query/OSLocatorsByURI RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locators` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) | repeated |  |
| `request` | [OSLocatorsByURIRequest](#provenance-metadata-v1-OSLocatorsByURIRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance-metadata-v1-OwnershipRequest"></a>

### OwnershipRequest
OwnershipRequest is the request type for the Query/Ownership RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-OwnershipResponse"></a>

### OwnershipResponse
OwnershipResponse is the response type for the Query/Ownership RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_uuids` | [string](#string) | repeated | A list of scope ids (uuid) associated with the given address. |
| `request` | [OwnershipRequest](#provenance-metadata-v1-OwnershipRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance-metadata-v1-QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-metadata-v1-Params) |  | params defines the parameters of the module. |
| `request` | [QueryParamsRequest](#provenance-metadata-v1-QueryParamsRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-QueryScopeNetAssetValuesRequest"></a>

### QueryScopeNetAssetValuesRequest
QueryNetAssetValuesRequest is the request type for the Query/NetAssetValues method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | scopeid metadata address |






<a name="provenance-metadata-v1-QueryScopeNetAssetValuesResponse"></a>

### QueryScopeNetAssetValuesResponse
QueryNetAssetValuesRequest is the response type for the Query/NetAssetValues method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `net_asset_values` | [NetAssetValue](#provenance-metadata-v1-NetAssetValue) | repeated | net asset values for scope |






<a name="provenance-metadata-v1-RecordSpecificationRequest"></a>

### RecordSpecificationRequest
RecordSpecificationRequest is the request type for the Query/RecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [string](#string) |  | specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84 or a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn. It can also be a record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. |
| `name` | [string](#string) |  | name is the name of the record to look up. It is required if the specification_id is a uuid or contract specification address. It is ignored if the specification_id is a record specification address. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-RecordSpecificationResponse"></a>

### RecordSpecificationResponse
RecordSpecificationResponse is the response type for the Query/RecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specification` | [RecordSpecificationWrapper](#provenance-metadata-v1-RecordSpecificationWrapper) |  | record_specification is the wrapped record specification. |
| `request` | [RecordSpecificationRequest](#provenance-metadata-v1-RecordSpecificationRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-RecordSpecificationWrapper"></a>

### RecordSpecificationWrapper
RecordSpecificationWrapper contains a single record specification and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [RecordSpecification](#provenance-metadata-v1-RecordSpecification) |  | specification is the on-chain record specification message. |
| `record_spec_id_info` | [RecordSpecIdInfo](#provenance-metadata-v1-RecordSpecIdInfo) |  | record_spec_id_info contains information about the id/address of the record specification. |






<a name="provenance-metadata-v1-RecordSpecificationsAllRequest"></a>

### RecordSpecificationsAllRequest
RecordSpecificationsAllRequest is the request type for the Query/RecordSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-RecordSpecificationsAllResponse"></a>

### RecordSpecificationsAllResponse
RecordSpecificationsAllResponse is the response type for the Query/RecordSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specifications` | [RecordSpecificationWrapper](#provenance-metadata-v1-RecordSpecificationWrapper) | repeated | record_specifications are the wrapped record specifications. |
| `request` | [RecordSpecificationsAllRequest](#provenance-metadata-v1-RecordSpecificationsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance-metadata-v1-RecordSpecificationsForContractSpecificationRequest"></a>

### RecordSpecificationsForContractSpecificationRequest
RecordSpecificationsForContractSpecificationRequest is the request type for the
Query/RecordSpecificationsForContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [string](#string) |  | specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84 or a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn. It can also be a record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-RecordSpecificationsForContractSpecificationResponse"></a>

### RecordSpecificationsForContractSpecificationResponse
RecordSpecificationsForContractSpecificationResponse is the response type for the
Query/RecordSpecificationsForContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specifications` | [RecordSpecificationWrapper](#provenance-metadata-v1-RecordSpecificationWrapper) | repeated | record_specifications is any number of wrapped record specifications associated with this contract_specification. |
| `contract_specification_uuid` | [string](#string) |  | contract_specification_uuid is the uuid of this contract specification. |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the contract specification address as a bech32 encoded string. |
| `request` | [RecordSpecificationsForContractSpecificationRequest](#provenance-metadata-v1-RecordSpecificationsForContractSpecificationRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-RecordWrapper"></a>

### RecordWrapper
RecordWrapper contains a single record and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record` | [Record](#provenance-metadata-v1-Record) |  | record is the on-chain record message. |
| `record_id_info` | [RecordIdInfo](#provenance-metadata-v1-RecordIdInfo) |  | record_id_info contains information about the id/address of the record. |
| `record_spec_id_info` | [RecordSpecIdInfo](#provenance-metadata-v1-RecordSpecIdInfo) |  | record_spec_id_info contains information about the id/address of the record specification. |






<a name="provenance-metadata-v1-RecordsAllRequest"></a>

### RecordsAllRequest
RecordsAllRequest is the request type for the Query/RecordsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-RecordsAllResponse"></a>

### RecordsAllResponse
RecordsAllResponse is the response type for the Query/RecordsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `records` | [RecordWrapper](#provenance-metadata-v1-RecordWrapper) | repeated | records are the wrapped records. |
| `request` | [RecordsAllRequest](#provenance-metadata-v1-RecordsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance-metadata-v1-RecordsRequest"></a>

### RecordsRequest
RecordsRequest is the request type for the Query/Records RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_addr` | [string](#string) |  | record_addr is a bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3. |
| `scope_id` | [string](#string) |  | scope_id can either be a uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a bech32 scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. |
| `session_id` | [string](#string) |  | session_id can either be a uuid, e.g. 5803f8bc-6067-4eb5-951f-2121671c2ec0 or a bech32 session address, e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. This can only be a uuid if a scope_id is also provided. |
| `name` | [string](#string) |  | name is the name of the record to look for |
| `include_scope` | [bool](#bool) |  | include_scope is a flag for whether to include the the scope containing these records in the response. |
| `include_sessions` | [bool](#bool) |  | include_sessions is a flag for whether to include the sessions containing these records in the response. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-RecordsResponse"></a>

### RecordsResponse
RecordsResponse is the response type for the Query/Records RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [ScopeWrapper](#provenance-metadata-v1-ScopeWrapper) |  | scope is the wrapped scope that holds these records (if requested). |
| `sessions` | [SessionWrapper](#provenance-metadata-v1-SessionWrapper) | repeated | sessions is any number of wrapped sessions that hold these records (if requested). |
| `records` | [RecordWrapper](#provenance-metadata-v1-RecordWrapper) | repeated | records is any number of wrapped record results. |
| `request` | [RecordsRequest](#provenance-metadata-v1-RecordsRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-ScopeRequest"></a>

### ScopeRequest
ScopeRequest is the request type for the Query/Scope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [string](#string) |  | scope_id can either be a uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a bech32 scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. |
| `session_addr` | [string](#string) |  | session_addr is a bech32 session address, e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. |
| `record_addr` | [string](#string) |  | record_addr is a bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3. |
| `include_sessions` | [bool](#bool) |  | include_sessions is a flag for whether to include the sessions of the scope in the response. |
| `include_records` | [bool](#bool) |  | include_records is a flag for whether to include the records of the scope in the response. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-ScopeResponse"></a>

### ScopeResponse
ScopeResponse is the response type for the Query/Scope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [ScopeWrapper](#provenance-metadata-v1-ScopeWrapper) |  | scope is the wrapped scope result. |
| `sessions` | [SessionWrapper](#provenance-metadata-v1-SessionWrapper) | repeated | sessions is any number of wrapped sessions in this scope (if requested). |
| `records` | [RecordWrapper](#provenance-metadata-v1-RecordWrapper) | repeated | records is any number of wrapped records in this scope (if requested). |
| `request` | [ScopeRequest](#provenance-metadata-v1-ScopeRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-ScopeSpecificationRequest"></a>

### ScopeSpecificationRequest
ScopeSpecificationRequest is the request type for the Query/ScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [string](#string) |  | specification_id can either be a uuid, e.g. dc83ea70-eacd-40fe-9adf-1cf6148bf8a2 or a bech32 scope specification address, e.g. scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m. |
| `include_contract_specs` | [bool](#bool) |  | include_contract_specs is a flag for whether to include the contract specifications of the scope specification in the response. |
| `include_record_specs` | [bool](#bool) |  | include_record_specs is a flag for whether to include the record specifications of the scope specification in the response. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-ScopeSpecificationResponse"></a>

### ScopeSpecificationResponse
ScopeSpecificationResponse is the response type for the Query/ScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specification` | [ScopeSpecificationWrapper](#provenance-metadata-v1-ScopeSpecificationWrapper) |  | scope_specification is the wrapped scope specification. |
| `contract_specs` | [ContractSpecificationWrapper](#provenance-metadata-v1-ContractSpecificationWrapper) | repeated | contract_specs is any number of wrapped contract specifications in this scope specification (if requested). |
| `record_specs` | [RecordSpecificationWrapper](#provenance-metadata-v1-RecordSpecificationWrapper) | repeated | record_specs is any number of wrapped record specifications in this scope specification (if requested). |
| `request` | [ScopeSpecificationRequest](#provenance-metadata-v1-ScopeSpecificationRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-ScopeSpecificationWrapper"></a>

### ScopeSpecificationWrapper
ScopeSpecificationWrapper contains a single scope specification and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [ScopeSpecification](#provenance-metadata-v1-ScopeSpecification) |  | specification is the on-chain scope specification message. |
| `scope_spec_id_info` | [ScopeSpecIdInfo](#provenance-metadata-v1-ScopeSpecIdInfo) |  | scope_spec_id_info contains information about the id/address of the scope specification. |






<a name="provenance-metadata-v1-ScopeSpecificationsAllRequest"></a>

### ScopeSpecificationsAllRequest
ScopeSpecificationsAllRequest is the request type for the Query/ScopeSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-ScopeSpecificationsAllResponse"></a>

### ScopeSpecificationsAllResponse
ScopeSpecificationsAllResponse is the response type for the Query/ScopeSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specifications` | [ScopeSpecificationWrapper](#provenance-metadata-v1-ScopeSpecificationWrapper) | repeated | scope_specifications are the wrapped scope specifications. |
| `request` | [ScopeSpecificationsAllRequest](#provenance-metadata-v1-ScopeSpecificationsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance-metadata-v1-ScopeWrapper"></a>

### ScopeWrapper
SessionWrapper contains a single scope and its uuid.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [Scope](#provenance-metadata-v1-Scope) |  | scope is the on-chain scope message. |
| `scope_id_info` | [ScopeIdInfo](#provenance-metadata-v1-ScopeIdInfo) |  | scope_id_info contains information about the id/address of the scope. |
| `scope_spec_id_info` | [ScopeSpecIdInfo](#provenance-metadata-v1-ScopeSpecIdInfo) |  | scope_spec_id_info contains information about the id/address of the scope specification. |






<a name="provenance-metadata-v1-ScopesAllRequest"></a>

### ScopesAllRequest
ScopesAllRequest is the request type for the Query/ScopesAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-ScopesAllResponse"></a>

### ScopesAllResponse
ScopesAllResponse is the response type for the Query/ScopesAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scopes` | [ScopeWrapper](#provenance-metadata-v1-ScopeWrapper) | repeated | scopes are the wrapped scopes. |
| `request` | [ScopesAllRequest](#provenance-metadata-v1-ScopesAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance-metadata-v1-SessionWrapper"></a>

### SessionWrapper
SessionWrapper contains a single session and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session` | [Session](#provenance-metadata-v1-Session) |  | session is the on-chain session message. |
| `session_id_info` | [SessionIdInfo](#provenance-metadata-v1-SessionIdInfo) |  | session_id_info contains information about the id/address of the session. |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance-metadata-v1-ContractSpecIdInfo) |  | contract_spec_id_info contains information about the id/address of the contract specification. |






<a name="provenance-metadata-v1-SessionsAllRequest"></a>

### SessionsAllRequest
SessionsAllRequest is the request type for the Query/SessionsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-SessionsAllResponse"></a>

### SessionsAllResponse
SessionsAllResponse is the response type for the Query/SessionsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sessions` | [SessionWrapper](#provenance-metadata-v1-SessionWrapper) | repeated | sessions are the wrapped sessions. |
| `request` | [SessionsAllRequest](#provenance-metadata-v1-SessionsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance-metadata-v1-SessionsRequest"></a>

### SessionsRequest
SessionsRequest is the request type for the Query/Sessions RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [string](#string) |  | scope_id can either be a uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a bech32 scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. |
| `session_id` | [string](#string) |  | session_id can either be a uuid, e.g. 5803f8bc-6067-4eb5-951f-2121671c2ec0 or a bech32 session address, e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. This can only be a uuid if a scope_id is also provided. |
| `record_addr` | [string](#string) |  | record_addr is a bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3. |
| `record_name` | [string](#string) |  | record_name is the name of the record to find the session for in the provided scope. |
| `include_scope` | [bool](#bool) |  | include_scope is a flag for whether to include the scope containing these sessions in the response. |
| `include_records` | [bool](#bool) |  | include_records is a flag for whether to include the records of these sessions in the response. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance-metadata-v1-SessionsResponse"></a>

### SessionsResponse
SessionsResponse is the response type for the Query/Sessions RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [ScopeWrapper](#provenance-metadata-v1-ScopeWrapper) |  | scope is the wrapped scope that holds these sessions (if requested). |
| `sessions` | [SessionWrapper](#provenance-metadata-v1-SessionWrapper) | repeated | sessions is any number of wrapped session results. |
| `records` | [RecordWrapper](#provenance-metadata-v1-RecordWrapper) | repeated | records is any number of wrapped records contained in these sessions (if requested). |
| `request` | [SessionsRequest](#provenance-metadata-v1-SessionsRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance-metadata-v1-ValueOwnershipRequest"></a>

### ValueOwnershipRequest
ValueOwnershipRequest is the request type for the Query/ValueOwnership RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance-metadata-v1-ValueOwnershipResponse"></a>

### ValueOwnershipResponse
ValueOwnershipResponse is the response type for the Query/ValueOwnership RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_uuids` | [string](#string) | repeated | A list of scope ids (uuid) associated with the given address. |
| `request` | [ValueOwnershipRequest](#provenance-metadata-v1-ValueOwnershipRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination provides the pagination information of this response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-metadata-v1-Query"></a>

### Query
Query defines the Metadata Query service.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `Params` | [QueryParamsRequest](#provenance-metadata-v1-QueryParamsRequest) | [QueryParamsResponse](#provenance-metadata-v1-QueryParamsResponse) | Params queries the parameters of x/metadata module. |
| `Scope` | [ScopeRequest](#provenance-metadata-v1-ScopeRequest) | [ScopeResponse](#provenance-metadata-v1-ScopeResponse) | Scope searches for a scope.<br>The scope id, if provided, must either be scope uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. The session addr, if provided, must be a bech32 session address, e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. The record_addr, if provided, must be a bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3.<br>* If only a scope_id is provided, that scope is returned. * If only a session_addr is provided, the scope containing that session is returned. * If only a record_addr is provided, the scope containing that record is returned. * If more than one of scope_id, session_addr, and record_addr are provided, and they don't refer to the same scope, a bad request is returned.<br>Providing a session addr or record addr does not limit the sessions and records returned (if requested). Those parameters are only used to find the scope.<br>By default, sessions and records are not included. Set include_sessions and/or include_records to true to include sessions and/or records. |
| `ScopesAll` | [ScopesAllRequest](#provenance-metadata-v1-ScopesAllRequest) | [ScopesAllResponse](#provenance-metadata-v1-ScopesAllResponse) | ScopesAll retrieves all scopes. |
| `Sessions` | [SessionsRequest](#provenance-metadata-v1-SessionsRequest) | [SessionsResponse](#provenance-metadata-v1-SessionsResponse) | Sessions searches for sessions.<br>The scope_id can either be scope uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. Similarly, the session_id can either be a uuid or session address, e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. The record_addr, if provided, must be a bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3.<br>* If only a scope_id is provided, all sessions in that scope are returned. * If only a session_id is provided, it must be an address, and that single session is returned. * If the session_id is a uuid, then either a scope_id or record_addr must also be provided, and that single session is returned. * If only a record_addr is provided, the session containing that record will be returned. * If a record_name is provided then either a scope_id, session_id as an address, or record_addr must also be provided, and the session containing that record will be returned.<br>A bad request is returned if: * The session_id is a uuid and is provided without a scope_id or record_addr. * A record_name is provided without any way to identify the scope (e.g. a scope_id, a session_id as an address, or a record_addr). * Two or more of scope_id, session_id as an address, and record_addr are provided and don't all refer to the same scope. * A record_addr (or scope_id and record_name) is provided with a session_id and that session does not contain such a record. * A record_addr and record_name are both provided, but reference different records.<br>By default, the scope and records are not included. Set include_scope and/or include_records to true to include the scope and/or records. |
| `SessionsAll` | [SessionsAllRequest](#provenance-metadata-v1-SessionsAllRequest) | [SessionsAllResponse](#provenance-metadata-v1-SessionsAllResponse) | SessionsAll retrieves all sessions. |
| `Records` | [RecordsRequest](#provenance-metadata-v1-RecordsRequest) | [RecordsResponse](#provenance-metadata-v1-RecordsResponse) | Records searches for records.<br>The record_addr, if provided, must be a bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3. The scope-id can either be scope uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. Similarly, the session_id can either be a uuid or session address, e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. The name is the name of the record you're interested in.<br>* If only a record_addr is provided, that single record will be returned. * If only a scope_id is provided, all records in that scope will be returned. * If only a session_id (or scope_id/session_id), all records in that session will be returned. * If a name is provided with a scope_id and/or session_id, that single record will be returned.<br>A bad request is returned if: * The session_id is a uuid and no scope_id is provided. * There are two or more of record_addr, session_id, and scope_id, and they don't all refer to the same scope. * A name is provided, but not a scope_id and/or a session_id. * A name and record_addr are provided and the name doesn't match the record_addr.<br>By default, the scope and sessions are not included. Set include_scope and/or include_sessions to true to include the scope and/or sessions. |
| `RecordsAll` | [RecordsAllRequest](#provenance-metadata-v1-RecordsAllRequest) | [RecordsAllResponse](#provenance-metadata-v1-RecordsAllResponse) | RecordsAll retrieves all records. |
| `Ownership` | [OwnershipRequest](#provenance-metadata-v1-OwnershipRequest) | [OwnershipResponse](#provenance-metadata-v1-OwnershipResponse) | Ownership returns the scope identifiers that have the given address in the owners list. |
| `ValueOwnership` | [ValueOwnershipRequest](#provenance-metadata-v1-ValueOwnershipRequest) | [ValueOwnershipResponse](#provenance-metadata-v1-ValueOwnershipResponse) | ValueOwnership returns the scope identifiers that list the given address as the value owner. |
| `ScopeSpecification` | [ScopeSpecificationRequest](#provenance-metadata-v1-ScopeSpecificationRequest) | [ScopeSpecificationResponse](#provenance-metadata-v1-ScopeSpecificationResponse) | ScopeSpecification returns a scope specification for the given specification id.<br>The specification_id can either be a uuid, e.g. dc83ea70-eacd-40fe-9adf-1cf6148bf8a2 or a bech32 scope specification address, e.g. scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m.<br>By default, the contract and record specifications are not included. Set include_contract_specs and/or include_record_specs to true to include contract and/or record specifications. |
| `ScopeSpecificationsAll` | [ScopeSpecificationsAllRequest](#provenance-metadata-v1-ScopeSpecificationsAllRequest) | [ScopeSpecificationsAllResponse](#provenance-metadata-v1-ScopeSpecificationsAllResponse) | ScopeSpecificationsAll retrieves all scope specifications. |
| `ContractSpecification` | [ContractSpecificationRequest](#provenance-metadata-v1-ContractSpecificationRequest) | [ContractSpecificationResponse](#provenance-metadata-v1-ContractSpecificationResponse) | ContractSpecification returns a contract specification for the given specification id.<br>The specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84, a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn, or a bech32 record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. If it is a record specification address, then the contract specification that contains that record specification is looked up.<br>By default, the record specifications for this contract specification are not included. Set include_record_specs to true to include them in the result. |
| `ContractSpecificationsAll` | [ContractSpecificationsAllRequest](#provenance-metadata-v1-ContractSpecificationsAllRequest) | [ContractSpecificationsAllResponse](#provenance-metadata-v1-ContractSpecificationsAllResponse) | ContractSpecificationsAll retrieves all contract specifications. |
| `RecordSpecificationsForContractSpecification` | [RecordSpecificationsForContractSpecificationRequest](#provenance-metadata-v1-RecordSpecificationsForContractSpecificationRequest) | [RecordSpecificationsForContractSpecificationResponse](#provenance-metadata-v1-RecordSpecificationsForContractSpecificationResponse) | RecordSpecificationsForContractSpecification returns the record specifications for the given input.<br>The specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84, a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn, or a bech32 record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. If it is a record specification address, then the contract specification that contains that record specification is used. |
| `RecordSpecification` | [RecordSpecificationRequest](#provenance-metadata-v1-RecordSpecificationRequest) | [RecordSpecificationResponse](#provenance-metadata-v1-RecordSpecificationResponse) | RecordSpecification returns a record specification for the given input. |
| `RecordSpecificationsAll` | [RecordSpecificationsAllRequest](#provenance-metadata-v1-RecordSpecificationsAllRequest) | [RecordSpecificationsAllResponse](#provenance-metadata-v1-RecordSpecificationsAllResponse) | RecordSpecificationsAll retrieves all record specifications. |
| `GetByAddr` | [GetByAddrRequest](#provenance-metadata-v1-GetByAddrRequest) | [GetByAddrResponse](#provenance-metadata-v1-GetByAddrResponse) | GetByAddr retrieves metadata given any address(es). |
| `OSLocatorParams` | [OSLocatorParamsRequest](#provenance-metadata-v1-OSLocatorParamsRequest) | [OSLocatorParamsResponse](#provenance-metadata-v1-OSLocatorParamsResponse) | OSLocatorParams returns all parameters for the object store locator sub module. |
| `OSLocator` | [OSLocatorRequest](#provenance-metadata-v1-OSLocatorRequest) | [OSLocatorResponse](#provenance-metadata-v1-OSLocatorResponse) | OSLocator returns an ObjectStoreLocator by its owner's address. |
| `OSLocatorsByURI` | [OSLocatorsByURIRequest](#provenance-metadata-v1-OSLocatorsByURIRequest) | [OSLocatorsByURIResponse](#provenance-metadata-v1-OSLocatorsByURIResponse) | OSLocatorsByURI returns all ObjectStoreLocator entries for a locator uri. |
| `OSLocatorsByScope` | [OSLocatorsByScopeRequest](#provenance-metadata-v1-OSLocatorsByScopeRequest) | [OSLocatorsByScopeResponse](#provenance-metadata-v1-OSLocatorsByScopeResponse) | OSLocatorsByScope returns all ObjectStoreLocator entries for a for all signer's present in the specified scope. |
| `OSAllLocators` | [OSAllLocatorsRequest](#provenance-metadata-v1-OSAllLocatorsRequest) | [OSAllLocatorsResponse](#provenance-metadata-v1-OSAllLocatorsResponse) | OSAllLocators returns all ObjectStoreLocator entries. |
| `AccountData` | [AccountDataRequest](#provenance-metadata-v1-AccountDataRequest) | [AccountDataResponse](#provenance-metadata-v1-AccountDataResponse) | AccountData gets the account data associated with a metadata address. Currently, only scope ids are supported. |
| `ScopeNetAssetValues` | [QueryScopeNetAssetValuesRequest](#provenance-metadata-v1-QueryScopeNetAssetValuesRequest) | [QueryScopeNetAssetValuesResponse](#provenance-metadata-v1-QueryScopeNetAssetValuesResponse) | ScopeNetAssetValues returns net asset values for scope |

 <!-- end services -->



<a name="provenance_metadata_v1_objectstore-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/objectstore.proto



<a name="provenance-metadata-v1-OSLocatorParams"></a>

### OSLocatorParams
Params defines the parameters for the metadata-locator module methods.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_uri_length` | [uint32](#uint32) |  |  |






<a name="provenance-metadata-v1-ObjectStoreLocator"></a>

### ObjectStoreLocator
Defines an Locator object stored on chain, which represents a owner( blockchain address) associated with a endpoint
uri for it's associated object store.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | account address the endpoint is owned by |
| `locator_uri` | [string](#string) |  | locator endpoint uri |
| `encryption_key` | [string](#string) |  | owners encryption key address |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_metadata_v1_metadata-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/metadata.proto



<a name="provenance-metadata-v1-ContractSpecIdInfo"></a>

### ContractSpecIdInfo
ContractSpecIdInfo contains various info regarding a contract specification id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_spec_id` | [bytes](#bytes) |  | contract_spec_id is the raw bytes of the contract specification address. |
| `contract_spec_id_prefix` | [bytes](#bytes) |  | contract_spec_id_prefix is the prefix portion of the contract_spec_id. |
| `contract_spec_id_contract_spec_uuid` | [bytes](#bytes) |  | contract_spec_id_contract_spec_uuid is the contract_spec_uuid portion of the contract_spec_id. |
| `contract_spec_addr` | [string](#string) |  | contract_spec_addr is the bech32 string version of the contract_spec_id. |
| `contract_spec_uuid` | [string](#string) |  | contract_spec_uuid is the uuid hex string of the contract_spec_id_contract_spec_uuid. |






<a name="provenance-metadata-v1-Params"></a>

### Params
Params defines the set of params for the metadata module.






<a name="provenance-metadata-v1-RecordIdInfo"></a>

### RecordIdInfo
RecordIdInfo contains various info regarding a record id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_id` | [bytes](#bytes) |  | record_id is the raw bytes of the record address. |
| `record_id_prefix` | [bytes](#bytes) |  | record_id_prefix is the prefix portion of the record_id. |
| `record_id_scope_uuid` | [bytes](#bytes) |  | record_id_scope_uuid is the scope_uuid portion of the record_id. |
| `record_id_hashed_name` | [bytes](#bytes) |  | record_id_hashed_name is the hashed name portion of the record_id. |
| `record_addr` | [string](#string) |  | record_addr is the bech32 string version of the record_id. |
| `scope_id_info` | [ScopeIdInfo](#provenance-metadata-v1-ScopeIdInfo) |  | scope_id_info is information about the scope id referenced in the record_id. |






<a name="provenance-metadata-v1-RecordSpecIdInfo"></a>

### RecordSpecIdInfo
RecordSpecIdInfo contains various info regarding a record specification id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_spec_id` | [bytes](#bytes) |  | record_spec_id is the raw bytes of the record specification address. |
| `record_spec_id_prefix` | [bytes](#bytes) |  | record_spec_id_prefix is the prefix portion of the record_spec_id. |
| `record_spec_id_contract_spec_uuid` | [bytes](#bytes) |  | record_spec_id_contract_spec_uuid is the contract_spec_uuid portion of the record_spec_id. |
| `record_spec_id_hashed_name` | [bytes](#bytes) |  | record_spec_id_hashed_name is the hashed name portion of the record_spec_id. |
| `record_spec_addr` | [string](#string) |  | record_spec_addr is the bech32 string version of the record_spec_id. |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance-metadata-v1-ContractSpecIdInfo) |  | contract_spec_id_info is information about the contract spec id referenced in the record_spec_id. |






<a name="provenance-metadata-v1-ScopeIdInfo"></a>

### ScopeIdInfo
ScopeIdInfo contains various info regarding a scope id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope_id is the raw bytes of the scope address. |
| `scope_id_prefix` | [bytes](#bytes) |  | scope_id_prefix is the prefix portion of the scope_id. |
| `scope_id_scope_uuid` | [bytes](#bytes) |  | scope_id_scope_uuid is the scope_uuid portion of the scope_id. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 string version of the scope_id. |
| `scope_uuid` | [string](#string) |  | scope_uuid is the uuid hex string of the scope_id_scope_uuid. |






<a name="provenance-metadata-v1-ScopeSpecIdInfo"></a>

### ScopeSpecIdInfo
ScopeSpecIdInfo contains various info regarding a scope specification id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_spec_id` | [bytes](#bytes) |  | scope_spec_id is the raw bytes of the scope specification address. |
| `scope_spec_id_prefix` | [bytes](#bytes) |  | scope_spec_id_prefix is the prefix portion of the scope_spec_id. |
| `scope_spec_id_scope_spec_uuid` | [bytes](#bytes) |  | scope_spec_id_scope_spec_uuid is the scope_spec_uuid portion of the scope_spec_id. |
| `scope_spec_addr` | [string](#string) |  | scope_spec_addr is the bech32 string version of the scope_spec_id. |
| `scope_spec_uuid` | [string](#string) |  | scope_spec_uuid is the uuid hex string of the scope_spec_id_scope_spec_uuid. |






<a name="provenance-metadata-v1-SessionIdInfo"></a>

### SessionIdInfo
SessionIdInfo contains various info regarding a session id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_id` | [bytes](#bytes) |  | session_id is the raw bytes of the session address. |
| `session_id_prefix` | [bytes](#bytes) |  | session_id_prefix is the prefix portion of the session_id. |
| `session_id_scope_uuid` | [bytes](#bytes) |  | session_id_scope_uuid is the scope_uuid portion of the session_id. |
| `session_id_session_uuid` | [bytes](#bytes) |  | session_id_session_uuid is the session_uuid portion of the session_id. |
| `session_addr` | [string](#string) |  | session_addr is the bech32 string version of the session_id. |
| `session_uuid` | [string](#string) |  | session_uuid is the uuid hex string of the session_id_session_uuid. |
| `scope_id_info` | [ScopeIdInfo](#provenance-metadata-v1-ScopeIdInfo) |  | scope_id_info is information about the scope id referenced in the session_id. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_metadata_v1_p8e_p8e-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/p8e/p8e.proto



<a name="provenance-metadata-v1-p8e-Condition"></a>

### Condition
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `condition_name` | [string](#string) |  |  |
| `result` | [ExecutionResult](#provenance-metadata-v1-p8e-ExecutionResult) |  |  |






<a name="provenance-metadata-v1-p8e-ConditionSpec"></a>

### ConditionSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `func_name` | [string](#string) |  |  |
| `input_specs` | [DefinitionSpec](#provenance-metadata-v1-p8e-DefinitionSpec) | repeated |  |
| `output_spec` | [OutputSpec](#provenance-metadata-v1-p8e-OutputSpec) |  |  |






<a name="provenance-metadata-v1-p8e-Consideration"></a>

### Consideration
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `consideration_name` | [string](#string) |  |  |
| `inputs` | [ProposedFact](#provenance-metadata-v1-p8e-ProposedFact) | repeated |  |
| `result` | [ExecutionResult](#provenance-metadata-v1-p8e-ExecutionResult) |  |  |






<a name="provenance-metadata-v1-p8e-ConsiderationSpec"></a>

### ConsiderationSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `func_name` | [string](#string) |  |  |
| `responsible_party` | [PartyType](#provenance-metadata-v1-p8e-PartyType) |  |  |
| `input_specs` | [DefinitionSpec](#provenance-metadata-v1-p8e-DefinitionSpec) | repeated |  |
| `output_spec` | [OutputSpec](#provenance-metadata-v1-p8e-OutputSpec) |  |  |






<a name="provenance-metadata-v1-p8e-Contract"></a>

### Contract
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `definition` | [DefinitionSpec](#provenance-metadata-v1-p8e-DefinitionSpec) |  |  |
| `spec` | [Fact](#provenance-metadata-v1-p8e-Fact) |  |  |
| `invoker` | [SigningAndEncryptionPublicKeys](#provenance-metadata-v1-p8e-SigningAndEncryptionPublicKeys) |  |  |
| `inputs` | [Fact](#provenance-metadata-v1-p8e-Fact) | repeated |  |
| `conditions` | [Condition](#provenance-metadata-v1-p8e-Condition) | repeated | **Deprecated.**  |
| `considerations` | [Consideration](#provenance-metadata-v1-p8e-Consideration) | repeated |  |
| `recitals` | [Recital](#provenance-metadata-v1-p8e-Recital) | repeated |  |
| `times_executed` | [int32](#int32) |  |  |
| `start_time` | [Timestamp](#provenance-metadata-v1-p8e-Timestamp) |  |  |
| `context` | [bytes](#bytes) |  |  |






<a name="provenance-metadata-v1-p8e-ContractSpec"></a>

### ContractSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `definition` | [DefinitionSpec](#provenance-metadata-v1-p8e-DefinitionSpec) |  |  |
| `input_specs` | [DefinitionSpec](#provenance-metadata-v1-p8e-DefinitionSpec) | repeated |  |
| `parties_involved` | [PartyType](#provenance-metadata-v1-p8e-PartyType) | repeated |  |
| `condition_specs` | [ConditionSpec](#provenance-metadata-v1-p8e-ConditionSpec) | repeated |  |
| `consideration_specs` | [ConsiderationSpec](#provenance-metadata-v1-p8e-ConsiderationSpec) | repeated |  |






<a name="provenance-metadata-v1-p8e-DefinitionSpec"></a>

### DefinitionSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `resource_location` | [Location](#provenance-metadata-v1-p8e-Location) |  |  |
| `signature` | [Signature](#provenance-metadata-v1-p8e-Signature) |  |  |
| `type` | [DefinitionSpecType](#provenance-metadata-v1-p8e-DefinitionSpecType) |  |  |






<a name="provenance-metadata-v1-p8e-ExecutionResult"></a>

### ExecutionResult
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `output` | [ProposedFact](#provenance-metadata-v1-p8e-ProposedFact) |  |  |
| `result` | [ExecutionResultType](#provenance-metadata-v1-p8e-ExecutionResultType) |  |  |
| `recorded_at` | [Timestamp](#provenance-metadata-v1-p8e-Timestamp) |  |  |
| `error_message` | [string](#string) |  |  |






<a name="provenance-metadata-v1-p8e-Fact"></a>

### Fact
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `data_location` | [Location](#provenance-metadata-v1-p8e-Location) |  |  |






<a name="provenance-metadata-v1-p8e-Location"></a>

### Location
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ref` | [ProvenanceReference](#provenance-metadata-v1-p8e-ProvenanceReference) |  |  |
| `classname` | [string](#string) |  |  |






<a name="provenance-metadata-v1-p8e-OutputSpec"></a>

### OutputSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `spec` | [DefinitionSpec](#provenance-metadata-v1-p8e-DefinitionSpec) |  |  |






<a name="provenance-metadata-v1-p8e-ProposedFact"></a>

### ProposedFact
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `hash` | [string](#string) |  |  |
| `classname` | [string](#string) |  |  |
| `ancestor` | [ProvenanceReference](#provenance-metadata-v1-p8e-ProvenanceReference) |  |  |






<a name="provenance-metadata-v1-p8e-ProvenanceReference"></a>

### ProvenanceReference
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_uuid` | [UUID](#provenance-metadata-v1-p8e-UUID) |  |  |
| `group_uuid` | [UUID](#provenance-metadata-v1-p8e-UUID) |  |  |
| `hash` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |






<a name="provenance-metadata-v1-p8e-PublicKey"></a>

### PublicKey
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `public_key_bytes` | [bytes](#bytes) |  |  |
| `type` | [PublicKeyType](#provenance-metadata-v1-p8e-PublicKeyType) |  |  |
| `curve` | [PublicKeyCurve](#provenance-metadata-v1-p8e-PublicKeyCurve) |  |  |






<a name="provenance-metadata-v1-p8e-Recital"></a>

### Recital
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signer_role` | [PartyType](#provenance-metadata-v1-p8e-PartyType) |  |  |
| `signer` | [SigningAndEncryptionPublicKeys](#provenance-metadata-v1-p8e-SigningAndEncryptionPublicKeys) |  |  |
| `address` | [bytes](#bytes) |  |  |






<a name="provenance-metadata-v1-p8e-Recitals"></a>

### Recitals
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `parties` | [Recital](#provenance-metadata-v1-p8e-Recital) | repeated |  |






<a name="provenance-metadata-v1-p8e-Signature"></a>

### Signature
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `algo` | [string](#string) |  |  |
| `provider` | [string](#string) |  |  |
| `signature` | [string](#string) |  |  |
| `signer` | [SigningAndEncryptionPublicKeys](#provenance-metadata-v1-p8e-SigningAndEncryptionPublicKeys) |  |  |






<a name="provenance-metadata-v1-p8e-SignatureSet"></a>

### SignatureSet
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signatures` | [Signature](#provenance-metadata-v1-p8e-Signature) | repeated |  |






<a name="provenance-metadata-v1-p8e-SigningAndEncryptionPublicKeys"></a>

### SigningAndEncryptionPublicKeys
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signing_public_key` | [PublicKey](#provenance-metadata-v1-p8e-PublicKey) |  |  |
| `encryption_public_key` | [PublicKey](#provenance-metadata-v1-p8e-PublicKey) |  |  |






<a name="provenance-metadata-v1-p8e-Timestamp"></a>

### Timestamp
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `seconds` | [int64](#int64) |  |  |
| `nanos` | [int32](#int32) |  |  |






<a name="provenance-metadata-v1-p8e-UUID"></a>

### UUID
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  |  |





 <!-- end messages -->


<a name="provenance-metadata-v1-p8e-DefinitionSpecType"></a>

### DefinitionSpecType
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `DEFINITION_SPEC_TYPE_UNKNOWN` | `0` | Deprecated: Do not use. |
| `DEFINITION_SPEC_TYPE_PROPOSED` | `1` | Deprecated: Do not use. |
| `DEFINITION_SPEC_TYPE_FACT` | `2` | Deprecated: Do not use. |
| `DEFINITION_SPEC_TYPE_FACT_LIST` | `3` | Deprecated: Do not use. |



<a name="provenance-metadata-v1-p8e-ExecutionResultType"></a>

### ExecutionResultType
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `RESULT_TYPE_UNKNOWN` | `0` | Deprecated: Do not use. |
| `RESULT_TYPE_PASS` | `1` | Deprecated: Do not use. |
| `RESULT_TYPE_SKIP` | `2` | Deprecated: Do not use. |
| `RESULT_TYPE_FAIL` | `3` | Deprecated: Do not use. |



<a name="provenance-metadata-v1-p8e-PartyType"></a>

### PartyType
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `PARTY_TYPE_UNKNOWN` | `0` | Deprecated: Do not use. |
| `PARTY_TYPE_ORIGINATOR` | `1` | Deprecated: Do not use. |
| `PARTY_TYPE_SERVICER` | `2` | Deprecated: Do not use. |
| `PARTY_TYPE_INVESTOR` | `3` | Deprecated: Do not use. |
| `PARTY_TYPE_CUSTODIAN` | `4` | Deprecated: Do not use. |
| `PARTY_TYPE_OWNER` | `5` | Deprecated: Do not use. |
| `PARTY_TYPE_AFFILIATE` | `6` | Deprecated: Do not use. |
| `PARTY_TYPE_OMNIBUS` | `7` | Deprecated: Do not use. |
| `PARTY_TYPE_PROVENANCE` | `8` | Deprecated: Do not use. |
| `PARTY_TYPE_MARKER` | `9` | Deprecated: Do not use. |
| `PARTY_TYPE_CONTROLLER` | `10` | Deprecated: Do not use. |
| `PARTY_TYPE_VALIDATOR` | `11` | Deprecated: Do not use. |



<a name="provenance-metadata-v1-p8e-PublicKeyCurve"></a>

### PublicKeyCurve
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `SECP256K1` | `0` | Deprecated: Do not use. |
| `P256` | `1` | Deprecated: Do not use. |



<a name="provenance-metadata-v1-p8e-PublicKeyType"></a>

### PublicKeyType
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `ELLIPTIC` | `0` | Deprecated: Do not use. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_metadata_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/genesis.proto



<a name="provenance-metadata-v1-GenesisState"></a>

### GenesisState
GenesisState defines the account module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance-metadata-v1-Params) |  | params defines all the parameters of the module. |
| `scopes` | [Scope](#provenance-metadata-v1-Scope) | repeated | A collection of metadata scopes and specs to create on start |
| `sessions` | [Session](#provenance-metadata-v1-Session) | repeated |  |
| `records` | [Record](#provenance-metadata-v1-Record) | repeated |  |
| `scope_specifications` | [ScopeSpecification](#provenance-metadata-v1-ScopeSpecification) | repeated |  |
| `contract_specifications` | [ContractSpecification](#provenance-metadata-v1-ContractSpecification) | repeated |  |
| `record_specifications` | [RecordSpecification](#provenance-metadata-v1-RecordSpecification) | repeated |  |
| `o_s_locator_params` | [OSLocatorParams](#provenance-metadata-v1-OSLocatorParams) |  |  |
| `object_store_locators` | [ObjectStoreLocator](#provenance-metadata-v1-ObjectStoreLocator) | repeated |  |
| `net_asset_values` | [MarkerNetAssetValues](#provenance-metadata-v1-MarkerNetAssetValues) | repeated | Net asset values assigned to scopes |






<a name="provenance-metadata-v1-MarkerNetAssetValues"></a>

### MarkerNetAssetValues
MarkerNetAssetValues defines the net asset values for a scope


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address defines the scope address |
| `net_asset_values` | [NetAssetValue](#provenance-metadata-v1-NetAssetValue) | repeated | net_asset_values that are assigned to scope |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_hold_v1_events-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/hold/v1/events.proto



<a name="provenance-hold-v1-EventHoldAdded"></a>

### EventHoldAdded
EventHoldAdded is an event indicating that some funds were placed on hold in an account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the bech32 address string of the account with the funds. |
| `amount` | [string](#string) |  | amount is a Coins string of the funds placed on hold. |
| `reason` | [string](#string) |  | reason is a human-readable indicator of why this hold was added. |






<a name="provenance-hold-v1-EventHoldReleased"></a>

### EventHoldReleased
EventHoldReleased is an event indicating that some funds were released from hold for an account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the bech32 address string of the account with the funds. |
| `amount` | [string](#string) |  | amount is a Coins string of the funds released from hold. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_hold_v1_hold-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/hold/v1/hold.proto



<a name="provenance-hold-v1-AccountHold"></a>

### AccountHold
AccountHold associates an address with an amount on hold for that address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the account address that holds the funds on hold. |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | amount is the balances that are on hold for the address. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance_hold_v1_query-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/hold/v1/query.proto



<a name="provenance-hold-v1-GetAllHoldsRequest"></a>

### GetAllHoldsRequest
GetAllHoldsRequest is the request type for the Query/GetAllHolds query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos-base-query-v1beta1-PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance-hold-v1-GetAllHoldsResponse"></a>

### GetAllHoldsResponse
GetAllHoldsResponse is the response type for the Query/GetAllHolds query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `holds` | [AccountHold](#provenance-hold-v1-AccountHold) | repeated | holds is a list of addresses with funds on hold and the amounts being held. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos-base-query-v1beta1-PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance-hold-v1-GetHoldsRequest"></a>

### GetHoldsRequest
GetHoldsRequest is the request type for the Query/GetHolds query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the account address to get on-hold balances for. |






<a name="provenance-hold-v1-GetHoldsResponse"></a>

### GetHoldsResponse
GetHoldsResponse is the response type for the Query/GetHolds query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos-base-v1beta1-Coin) | repeated | amount is the total on hold for the requested address. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance-hold-v1-Query"></a>

### Query
Query defines the gRPC querier service for attribute module.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| `GetHolds` | [GetHoldsRequest](#provenance-hold-v1-GetHoldsRequest) | [GetHoldsResponse](#provenance-hold-v1-GetHoldsResponse) | GetHolds looks up the funds that are on hold for an address. |
| `GetAllHolds` | [GetAllHoldsRequest](#provenance-hold-v1-GetAllHoldsRequest) | [GetAllHoldsResponse](#provenance-hold-v1-GetAllHoldsResponse) | GetAllHolds returns all addresses with funds on hold, and the amount held. |

 <!-- end services -->



<a name="provenance_hold_v1_genesis-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/hold/v1/genesis.proto



<a name="provenance-hold-v1-GenesisState"></a>

### GenesisState
GenesisState defines the attribute module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `holds` | [AccountHold](#provenance-hold-v1-AccountHold) | repeated | holds defines the funds on hold at genesis. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
