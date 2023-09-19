<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [provenance/attribute/v1/attribute.proto](#provenance/attribute/v1/attribute.proto)
    - [Attribute](#provenance.attribute.v1.Attribute)
    - [EventAccountDataUpdated](#provenance.attribute.v1.EventAccountDataUpdated)
    - [EventAttributeAdd](#provenance.attribute.v1.EventAttributeAdd)
    - [EventAttributeDelete](#provenance.attribute.v1.EventAttributeDelete)
    - [EventAttributeDistinctDelete](#provenance.attribute.v1.EventAttributeDistinctDelete)
    - [EventAttributeExpirationUpdate](#provenance.attribute.v1.EventAttributeExpirationUpdate)
    - [EventAttributeExpired](#provenance.attribute.v1.EventAttributeExpired)
    - [EventAttributeUpdate](#provenance.attribute.v1.EventAttributeUpdate)
    - [Params](#provenance.attribute.v1.Params)
  
    - [AttributeType](#provenance.attribute.v1.AttributeType)
  
- [provenance/attribute/v1/genesis.proto](#provenance/attribute/v1/genesis.proto)
    - [GenesisState](#provenance.attribute.v1.GenesisState)
  
- [provenance/attribute/v1/query.proto](#provenance/attribute/v1/query.proto)
    - [QueryAccountDataRequest](#provenance.attribute.v1.QueryAccountDataRequest)
    - [QueryAccountDataResponse](#provenance.attribute.v1.QueryAccountDataResponse)
    - [QueryAttributeAccountsRequest](#provenance.attribute.v1.QueryAttributeAccountsRequest)
    - [QueryAttributeAccountsResponse](#provenance.attribute.v1.QueryAttributeAccountsResponse)
    - [QueryAttributeRequest](#provenance.attribute.v1.QueryAttributeRequest)
    - [QueryAttributeResponse](#provenance.attribute.v1.QueryAttributeResponse)
    - [QueryAttributesRequest](#provenance.attribute.v1.QueryAttributesRequest)
    - [QueryAttributesResponse](#provenance.attribute.v1.QueryAttributesResponse)
    - [QueryParamsRequest](#provenance.attribute.v1.QueryParamsRequest)
    - [QueryParamsResponse](#provenance.attribute.v1.QueryParamsResponse)
    - [QueryScanRequest](#provenance.attribute.v1.QueryScanRequest)
    - [QueryScanResponse](#provenance.attribute.v1.QueryScanResponse)
  
    - [Query](#provenance.attribute.v1.Query)
  
- [provenance/attribute/v1/tx.proto](#provenance/attribute/v1/tx.proto)
    - [MsgAddAttributeRequest](#provenance.attribute.v1.MsgAddAttributeRequest)
    - [MsgAddAttributeResponse](#provenance.attribute.v1.MsgAddAttributeResponse)
    - [MsgDeleteAttributeRequest](#provenance.attribute.v1.MsgDeleteAttributeRequest)
    - [MsgDeleteAttributeResponse](#provenance.attribute.v1.MsgDeleteAttributeResponse)
    - [MsgDeleteDistinctAttributeRequest](#provenance.attribute.v1.MsgDeleteDistinctAttributeRequest)
    - [MsgDeleteDistinctAttributeResponse](#provenance.attribute.v1.MsgDeleteDistinctAttributeResponse)
    - [MsgSetAccountDataRequest](#provenance.attribute.v1.MsgSetAccountDataRequest)
    - [MsgSetAccountDataResponse](#provenance.attribute.v1.MsgSetAccountDataResponse)
    - [MsgUpdateAttributeExpirationRequest](#provenance.attribute.v1.MsgUpdateAttributeExpirationRequest)
    - [MsgUpdateAttributeExpirationResponse](#provenance.attribute.v1.MsgUpdateAttributeExpirationResponse)
    - [MsgUpdateAttributeRequest](#provenance.attribute.v1.MsgUpdateAttributeRequest)
    - [MsgUpdateAttributeResponse](#provenance.attribute.v1.MsgUpdateAttributeResponse)
  
    - [Msg](#provenance.attribute.v1.Msg)
  
- [provenance/exchange/v1/events.proto](#provenance/exchange/v1/events.proto)
    - [EventCreateMarketSubmitted](#provenance.exchange.v1.EventCreateMarketSubmitted)
    - [EventMarketCreated](#provenance.exchange.v1.EventMarketCreated)
    - [EventMarketDetailsUpdated](#provenance.exchange.v1.EventMarketDetailsUpdated)
    - [EventMarketDisabled](#provenance.exchange.v1.EventMarketDisabled)
    - [EventMarketEnabled](#provenance.exchange.v1.EventMarketEnabled)
    - [EventMarketFeesUpdated](#provenance.exchange.v1.EventMarketFeesUpdated)
    - [EventMarketPermissionsUpdated](#provenance.exchange.v1.EventMarketPermissionsUpdated)
    - [EventMarketReqAttrUpdated](#provenance.exchange.v1.EventMarketReqAttrUpdated)
    - [EventMarketUserSettleUpdated](#provenance.exchange.v1.EventMarketUserSettleUpdated)
    - [EventMarketWithdraw](#provenance.exchange.v1.EventMarketWithdraw)
    - [EventOrderCancelled](#provenance.exchange.v1.EventOrderCancelled)
    - [EventOrderCreated](#provenance.exchange.v1.EventOrderCreated)
    - [EventOrderFilled](#provenance.exchange.v1.EventOrderFilled)
    - [EventOrderPartiallyFilled](#provenance.exchange.v1.EventOrderPartiallyFilled)
    - [EventParamsUpdated](#provenance.exchange.v1.EventParamsUpdated)
  
- [provenance/exchange/v1/market.proto](#provenance/exchange/v1/market.proto)
    - [AccessGrant](#provenance.exchange.v1.AccessGrant)
    - [FeeRatio](#provenance.exchange.v1.FeeRatio)
    - [Market](#provenance.exchange.v1.Market)
    - [MarketAccount](#provenance.exchange.v1.MarketAccount)
    - [MarketDetails](#provenance.exchange.v1.MarketDetails)
  
    - [Permission](#provenance.exchange.v1.Permission)
  
- [provenance/exchange/v1/orders.proto](#provenance/exchange/v1/orders.proto)
    - [AskOrder](#provenance.exchange.v1.AskOrder)
    - [BidOrder](#provenance.exchange.v1.BidOrder)
    - [Order](#provenance.exchange.v1.Order)
  
- [provenance/exchange/v1/params.proto](#provenance/exchange/v1/params.proto)
    - [DenomSplit](#provenance.exchange.v1.DenomSplit)
    - [Params](#provenance.exchange.v1.Params)
  
- [provenance/exchange/v1/genesis.proto](#provenance/exchange/v1/genesis.proto)
    - [GenesisState](#provenance.exchange.v1.GenesisState)
  
- [provenance/exchange/v1/query.proto](#provenance/exchange/v1/query.proto)
    - [QueryGetAddressOrdersRequest](#provenance.exchange.v1.QueryGetAddressOrdersRequest)
    - [QueryGetAddressOrdersResponse](#provenance.exchange.v1.QueryGetAddressOrdersResponse)
    - [QueryGetAllOrdersRequest](#provenance.exchange.v1.QueryGetAllOrdersRequest)
    - [QueryGetAllOrdersResponse](#provenance.exchange.v1.QueryGetAllOrdersResponse)
    - [QueryGetMarketOrdersRequest](#provenance.exchange.v1.QueryGetMarketOrdersRequest)
    - [QueryGetMarketOrdersResponse](#provenance.exchange.v1.QueryGetMarketOrdersResponse)
    - [QueryGetOrderRequest](#provenance.exchange.v1.QueryGetOrderRequest)
    - [QueryGetOrderResponse](#provenance.exchange.v1.QueryGetOrderResponse)
    - [QueryMarketInfoRequest](#provenance.exchange.v1.QueryMarketInfoRequest)
    - [QueryMarketInfoResponse](#provenance.exchange.v1.QueryMarketInfoResponse)
    - [QueryOrderFeeCalcRequest](#provenance.exchange.v1.QueryOrderFeeCalcRequest)
    - [QueryOrderFeeCalcResponse](#provenance.exchange.v1.QueryOrderFeeCalcResponse)
    - [QueryParamsRequest](#provenance.exchange.v1.QueryParamsRequest)
    - [QueryParamsResponse](#provenance.exchange.v1.QueryParamsResponse)
    - [QuerySettlementFeeCalcRequest](#provenance.exchange.v1.QuerySettlementFeeCalcRequest)
    - [QuerySettlementFeeCalcResponse](#provenance.exchange.v1.QuerySettlementFeeCalcResponse)
    - [QueryValidateCreateMarketRequest](#provenance.exchange.v1.QueryValidateCreateMarketRequest)
    - [QueryValidateCreateMarketResponse](#provenance.exchange.v1.QueryValidateCreateMarketResponse)
    - [QueryValidateManageFeesRequest](#provenance.exchange.v1.QueryValidateManageFeesRequest)
    - [QueryValidateManageFeesResponse](#provenance.exchange.v1.QueryValidateManageFeesResponse)
  
    - [Query](#provenance.exchange.v1.Query)
  
- [provenance/exchange/v1/tx.proto](#provenance/exchange/v1/tx.proto)
    - [MsgCancelOrderRequest](#provenance.exchange.v1.MsgCancelOrderRequest)
    - [MsgCancelOrderResponse](#provenance.exchange.v1.MsgCancelOrderResponse)
    - [MsgCreateAskRequest](#provenance.exchange.v1.MsgCreateAskRequest)
    - [MsgCreateAskResponse](#provenance.exchange.v1.MsgCreateAskResponse)
    - [MsgCreateBidRequest](#provenance.exchange.v1.MsgCreateBidRequest)
    - [MsgCreateBidResponse](#provenance.exchange.v1.MsgCreateBidResponse)
    - [MsgFillAsksRequest](#provenance.exchange.v1.MsgFillAsksRequest)
    - [MsgFillAsksResponse](#provenance.exchange.v1.MsgFillAsksResponse)
    - [MsgFillBidsRequest](#provenance.exchange.v1.MsgFillBidsRequest)
    - [MsgFillBidsResponse](#provenance.exchange.v1.MsgFillBidsResponse)
    - [MsgGovCreateMarketRequest](#provenance.exchange.v1.MsgGovCreateMarketRequest)
    - [MsgGovCreateMarketResponse](#provenance.exchange.v1.MsgGovCreateMarketResponse)
    - [MsgGovManageFeesRequest](#provenance.exchange.v1.MsgGovManageFeesRequest)
    - [MsgGovManageFeesResponse](#provenance.exchange.v1.MsgGovManageFeesResponse)
    - [MsgGovUpdateParamsRequest](#provenance.exchange.v1.MsgGovUpdateParamsRequest)
    - [MsgGovUpdateParamsResponse](#provenance.exchange.v1.MsgGovUpdateParamsResponse)
    - [MsgMarketManagePermissionsRequest](#provenance.exchange.v1.MsgMarketManagePermissionsRequest)
    - [MsgMarketManagePermissionsResponse](#provenance.exchange.v1.MsgMarketManagePermissionsResponse)
    - [MsgMarketManageReqAttrsRequest](#provenance.exchange.v1.MsgMarketManageReqAttrsRequest)
    - [MsgMarketManageReqAttrsResponse](#provenance.exchange.v1.MsgMarketManageReqAttrsResponse)
    - [MsgMarketSettleRequest](#provenance.exchange.v1.MsgMarketSettleRequest)
    - [MsgMarketSettleResponse](#provenance.exchange.v1.MsgMarketSettleResponse)
    - [MsgMarketUpdateDetailsRequest](#provenance.exchange.v1.MsgMarketUpdateDetailsRequest)
    - [MsgMarketUpdateDetailsResponse](#provenance.exchange.v1.MsgMarketUpdateDetailsResponse)
    - [MsgMarketUpdateEnabledRequest](#provenance.exchange.v1.MsgMarketUpdateEnabledRequest)
    - [MsgMarketUpdateEnabledResponse](#provenance.exchange.v1.MsgMarketUpdateEnabledResponse)
    - [MsgMarketUpdateUserSettleRequest](#provenance.exchange.v1.MsgMarketUpdateUserSettleRequest)
    - [MsgMarketUpdateUserSettleResponse](#provenance.exchange.v1.MsgMarketUpdateUserSettleResponse)
    - [MsgMarketWithdrawRequest](#provenance.exchange.v1.MsgMarketWithdrawRequest)
    - [MsgMarketWithdrawResponse](#provenance.exchange.v1.MsgMarketWithdrawResponse)
  
    - [Msg](#provenance.exchange.v1.Msg)
  
- [provenance/hold/v1/events.proto](#provenance/hold/v1/events.proto)
    - [EventHoldAdded](#provenance.hold.v1.EventHoldAdded)
    - [EventHoldReleased](#provenance.hold.v1.EventHoldReleased)
  
- [provenance/hold/v1/hold.proto](#provenance/hold/v1/hold.proto)
    - [AccountHold](#provenance.hold.v1.AccountHold)
  
- [provenance/hold/v1/genesis.proto](#provenance/hold/v1/genesis.proto)
    - [GenesisState](#provenance.hold.v1.GenesisState)
  
- [provenance/hold/v1/query.proto](#provenance/hold/v1/query.proto)
    - [GetAllHoldsRequest](#provenance.hold.v1.GetAllHoldsRequest)
    - [GetAllHoldsResponse](#provenance.hold.v1.GetAllHoldsResponse)
    - [GetHoldsRequest](#provenance.hold.v1.GetHoldsRequest)
    - [GetHoldsResponse](#provenance.hold.v1.GetHoldsResponse)
  
    - [Query](#provenance.hold.v1.Query)
  
- [provenance/ibchooks/v1/params.proto](#provenance/ibchooks/v1/params.proto)
    - [Params](#provenance.ibchooks.v1.Params)
  
- [provenance/ibchooks/v1/genesis.proto](#provenance/ibchooks/v1/genesis.proto)
    - [GenesisState](#provenance.ibchooks.v1.GenesisState)
  
- [provenance/ibchooks/v1/tx.proto](#provenance/ibchooks/v1/tx.proto)
    - [MsgEmitIBCAck](#provenance.ibchooks.v1.MsgEmitIBCAck)
    - [MsgEmitIBCAckResponse](#provenance.ibchooks.v1.MsgEmitIBCAckResponse)
  
    - [Msg](#provenance.ibchooks.v1.Msg)
  
- [provenance/marker/v1/accessgrant.proto](#provenance/marker/v1/accessgrant.proto)
    - [AccessGrant](#provenance.marker.v1.AccessGrant)
  
    - [Access](#provenance.marker.v1.Access)
  
- [provenance/marker/v1/authz.proto](#provenance/marker/v1/authz.proto)
    - [MarkerTransferAuthorization](#provenance.marker.v1.MarkerTransferAuthorization)
  
- [provenance/marker/v1/marker.proto](#provenance/marker/v1/marker.proto)
    - [EventDenomUnit](#provenance.marker.v1.EventDenomUnit)
    - [EventMarkerAccess](#provenance.marker.v1.EventMarkerAccess)
    - [EventMarkerActivate](#provenance.marker.v1.EventMarkerActivate)
    - [EventMarkerAdd](#provenance.marker.v1.EventMarkerAdd)
    - [EventMarkerAddAccess](#provenance.marker.v1.EventMarkerAddAccess)
    - [EventMarkerBurn](#provenance.marker.v1.EventMarkerBurn)
    - [EventMarkerCancel](#provenance.marker.v1.EventMarkerCancel)
    - [EventMarkerDelete](#provenance.marker.v1.EventMarkerDelete)
    - [EventMarkerDeleteAccess](#provenance.marker.v1.EventMarkerDeleteAccess)
    - [EventMarkerFinalize](#provenance.marker.v1.EventMarkerFinalize)
    - [EventMarkerMint](#provenance.marker.v1.EventMarkerMint)
    - [EventMarkerSetDenomMetadata](#provenance.marker.v1.EventMarkerSetDenomMetadata)
    - [EventMarkerTransfer](#provenance.marker.v1.EventMarkerTransfer)
    - [EventMarkerWithdraw](#provenance.marker.v1.EventMarkerWithdraw)
    - [MarkerAccount](#provenance.marker.v1.MarkerAccount)
    - [Params](#provenance.marker.v1.Params)
  
    - [MarkerStatus](#provenance.marker.v1.MarkerStatus)
    - [MarkerType](#provenance.marker.v1.MarkerType)
  
- [provenance/marker/v1/genesis.proto](#provenance/marker/v1/genesis.proto)
    - [GenesisState](#provenance.marker.v1.GenesisState)
  
- [provenance/marker/v1/proposals.proto](#provenance/marker/v1/proposals.proto)
    - [AddMarkerProposal](#provenance.marker.v1.AddMarkerProposal)
    - [ChangeStatusProposal](#provenance.marker.v1.ChangeStatusProposal)
    - [RemoveAdministratorProposal](#provenance.marker.v1.RemoveAdministratorProposal)
    - [SetAdministratorProposal](#provenance.marker.v1.SetAdministratorProposal)
    - [SetDenomMetadataProposal](#provenance.marker.v1.SetDenomMetadataProposal)
    - [SupplyDecreaseProposal](#provenance.marker.v1.SupplyDecreaseProposal)
    - [SupplyIncreaseProposal](#provenance.marker.v1.SupplyIncreaseProposal)
    - [WithdrawEscrowProposal](#provenance.marker.v1.WithdrawEscrowProposal)
  
- [provenance/marker/v1/query.proto](#provenance/marker/v1/query.proto)
    - [Balance](#provenance.marker.v1.Balance)
    - [QueryAccessRequest](#provenance.marker.v1.QueryAccessRequest)
    - [QueryAccessResponse](#provenance.marker.v1.QueryAccessResponse)
    - [QueryAccountDataRequest](#provenance.marker.v1.QueryAccountDataRequest)
    - [QueryAccountDataResponse](#provenance.marker.v1.QueryAccountDataResponse)
    - [QueryAllMarkersRequest](#provenance.marker.v1.QueryAllMarkersRequest)
    - [QueryAllMarkersResponse](#provenance.marker.v1.QueryAllMarkersResponse)
    - [QueryDenomMetadataRequest](#provenance.marker.v1.QueryDenomMetadataRequest)
    - [QueryDenomMetadataResponse](#provenance.marker.v1.QueryDenomMetadataResponse)
    - [QueryEscrowRequest](#provenance.marker.v1.QueryEscrowRequest)
    - [QueryEscrowResponse](#provenance.marker.v1.QueryEscrowResponse)
    - [QueryHoldingRequest](#provenance.marker.v1.QueryHoldingRequest)
    - [QueryHoldingResponse](#provenance.marker.v1.QueryHoldingResponse)
    - [QueryMarkerRequest](#provenance.marker.v1.QueryMarkerRequest)
    - [QueryMarkerResponse](#provenance.marker.v1.QueryMarkerResponse)
    - [QueryParamsRequest](#provenance.marker.v1.QueryParamsRequest)
    - [QueryParamsResponse](#provenance.marker.v1.QueryParamsResponse)
    - [QuerySupplyRequest](#provenance.marker.v1.QuerySupplyRequest)
    - [QuerySupplyResponse](#provenance.marker.v1.QuerySupplyResponse)
  
    - [Query](#provenance.marker.v1.Query)
  
- [provenance/marker/v1/si.proto](#provenance/marker/v1/si.proto)
    - [SIPrefix](#provenance.marker.v1.SIPrefix)
  
- [provenance/marker/v1/tx.proto](#provenance/marker/v1/tx.proto)
    - [MsgActivateRequest](#provenance.marker.v1.MsgActivateRequest)
    - [MsgActivateResponse](#provenance.marker.v1.MsgActivateResponse)
    - [MsgAddAccessRequest](#provenance.marker.v1.MsgAddAccessRequest)
    - [MsgAddAccessResponse](#provenance.marker.v1.MsgAddAccessResponse)
    - [MsgAddFinalizeActivateMarkerRequest](#provenance.marker.v1.MsgAddFinalizeActivateMarkerRequest)
    - [MsgAddFinalizeActivateMarkerResponse](#provenance.marker.v1.MsgAddFinalizeActivateMarkerResponse)
    - [MsgAddMarkerRequest](#provenance.marker.v1.MsgAddMarkerRequest)
    - [MsgAddMarkerResponse](#provenance.marker.v1.MsgAddMarkerResponse)
    - [MsgBurnRequest](#provenance.marker.v1.MsgBurnRequest)
    - [MsgBurnResponse](#provenance.marker.v1.MsgBurnResponse)
    - [MsgCancelRequest](#provenance.marker.v1.MsgCancelRequest)
    - [MsgCancelResponse](#provenance.marker.v1.MsgCancelResponse)
    - [MsgDeleteAccessRequest](#provenance.marker.v1.MsgDeleteAccessRequest)
    - [MsgDeleteAccessResponse](#provenance.marker.v1.MsgDeleteAccessResponse)
    - [MsgDeleteRequest](#provenance.marker.v1.MsgDeleteRequest)
    - [MsgDeleteResponse](#provenance.marker.v1.MsgDeleteResponse)
    - [MsgFinalizeRequest](#provenance.marker.v1.MsgFinalizeRequest)
    - [MsgFinalizeResponse](#provenance.marker.v1.MsgFinalizeResponse)
    - [MsgGrantAllowanceRequest](#provenance.marker.v1.MsgGrantAllowanceRequest)
    - [MsgGrantAllowanceResponse](#provenance.marker.v1.MsgGrantAllowanceResponse)
    - [MsgIbcTransferRequest](#provenance.marker.v1.MsgIbcTransferRequest)
    - [MsgIbcTransferResponse](#provenance.marker.v1.MsgIbcTransferResponse)
    - [MsgMintRequest](#provenance.marker.v1.MsgMintRequest)
    - [MsgMintResponse](#provenance.marker.v1.MsgMintResponse)
    - [MsgSetAccountDataRequest](#provenance.marker.v1.MsgSetAccountDataRequest)
    - [MsgSetAccountDataResponse](#provenance.marker.v1.MsgSetAccountDataResponse)
    - [MsgSetDenomMetadataRequest](#provenance.marker.v1.MsgSetDenomMetadataRequest)
    - [MsgSetDenomMetadataResponse](#provenance.marker.v1.MsgSetDenomMetadataResponse)
    - [MsgSupplyIncreaseProposalRequest](#provenance.marker.v1.MsgSupplyIncreaseProposalRequest)
    - [MsgSupplyIncreaseProposalResponse](#provenance.marker.v1.MsgSupplyIncreaseProposalResponse)
    - [MsgTransferRequest](#provenance.marker.v1.MsgTransferRequest)
    - [MsgTransferResponse](#provenance.marker.v1.MsgTransferResponse)
    - [MsgUpdateForcedTransferRequest](#provenance.marker.v1.MsgUpdateForcedTransferRequest)
    - [MsgUpdateForcedTransferResponse](#provenance.marker.v1.MsgUpdateForcedTransferResponse)
    - [MsgUpdateRequiredAttributesRequest](#provenance.marker.v1.MsgUpdateRequiredAttributesRequest)
    - [MsgUpdateRequiredAttributesResponse](#provenance.marker.v1.MsgUpdateRequiredAttributesResponse)
    - [MsgUpdateSendDenyListRequest](#provenance.marker.v1.MsgUpdateSendDenyListRequest)
    - [MsgUpdateSendDenyListResponse](#provenance.marker.v1.MsgUpdateSendDenyListResponse)
    - [MsgWithdrawRequest](#provenance.marker.v1.MsgWithdrawRequest)
    - [MsgWithdrawResponse](#provenance.marker.v1.MsgWithdrawResponse)
  
    - [Msg](#provenance.marker.v1.Msg)
  
- [provenance/metadata/v1/events.proto](#provenance/metadata/v1/events.proto)
    - [EventContractSpecificationCreated](#provenance.metadata.v1.EventContractSpecificationCreated)
    - [EventContractSpecificationDeleted](#provenance.metadata.v1.EventContractSpecificationDeleted)
    - [EventContractSpecificationUpdated](#provenance.metadata.v1.EventContractSpecificationUpdated)
    - [EventOSLocatorCreated](#provenance.metadata.v1.EventOSLocatorCreated)
    - [EventOSLocatorDeleted](#provenance.metadata.v1.EventOSLocatorDeleted)
    - [EventOSLocatorUpdated](#provenance.metadata.v1.EventOSLocatorUpdated)
    - [EventRecordCreated](#provenance.metadata.v1.EventRecordCreated)
    - [EventRecordDeleted](#provenance.metadata.v1.EventRecordDeleted)
    - [EventRecordSpecificationCreated](#provenance.metadata.v1.EventRecordSpecificationCreated)
    - [EventRecordSpecificationDeleted](#provenance.metadata.v1.EventRecordSpecificationDeleted)
    - [EventRecordSpecificationUpdated](#provenance.metadata.v1.EventRecordSpecificationUpdated)
    - [EventRecordUpdated](#provenance.metadata.v1.EventRecordUpdated)
    - [EventScopeCreated](#provenance.metadata.v1.EventScopeCreated)
    - [EventScopeDeleted](#provenance.metadata.v1.EventScopeDeleted)
    - [EventScopeSpecificationCreated](#provenance.metadata.v1.EventScopeSpecificationCreated)
    - [EventScopeSpecificationDeleted](#provenance.metadata.v1.EventScopeSpecificationDeleted)
    - [EventScopeSpecificationUpdated](#provenance.metadata.v1.EventScopeSpecificationUpdated)
    - [EventScopeUpdated](#provenance.metadata.v1.EventScopeUpdated)
    - [EventSessionCreated](#provenance.metadata.v1.EventSessionCreated)
    - [EventSessionDeleted](#provenance.metadata.v1.EventSessionDeleted)
    - [EventSessionUpdated](#provenance.metadata.v1.EventSessionUpdated)
    - [EventTxCompleted](#provenance.metadata.v1.EventTxCompleted)
  
- [provenance/metadata/v1/metadata.proto](#provenance/metadata/v1/metadata.proto)
    - [ContractSpecIdInfo](#provenance.metadata.v1.ContractSpecIdInfo)
    - [Params](#provenance.metadata.v1.Params)
    - [RecordIdInfo](#provenance.metadata.v1.RecordIdInfo)
    - [RecordSpecIdInfo](#provenance.metadata.v1.RecordSpecIdInfo)
    - [ScopeIdInfo](#provenance.metadata.v1.ScopeIdInfo)
    - [ScopeSpecIdInfo](#provenance.metadata.v1.ScopeSpecIdInfo)
    - [SessionIdInfo](#provenance.metadata.v1.SessionIdInfo)
  
- [provenance/metadata/v1/specification.proto](#provenance/metadata/v1/specification.proto)
    - [ContractSpecification](#provenance.metadata.v1.ContractSpecification)
    - [Description](#provenance.metadata.v1.Description)
    - [InputSpecification](#provenance.metadata.v1.InputSpecification)
    - [RecordSpecification](#provenance.metadata.v1.RecordSpecification)
    - [ScopeSpecification](#provenance.metadata.v1.ScopeSpecification)
  
    - [DefinitionType](#provenance.metadata.v1.DefinitionType)
    - [PartyType](#provenance.metadata.v1.PartyType)
  
- [provenance/metadata/v1/scope.proto](#provenance/metadata/v1/scope.proto)
    - [AuditFields](#provenance.metadata.v1.AuditFields)
    - [Party](#provenance.metadata.v1.Party)
    - [Process](#provenance.metadata.v1.Process)
    - [Record](#provenance.metadata.v1.Record)
    - [RecordInput](#provenance.metadata.v1.RecordInput)
    - [RecordOutput](#provenance.metadata.v1.RecordOutput)
    - [Scope](#provenance.metadata.v1.Scope)
    - [Session](#provenance.metadata.v1.Session)
  
    - [RecordInputStatus](#provenance.metadata.v1.RecordInputStatus)
    - [ResultStatus](#provenance.metadata.v1.ResultStatus)
  
- [provenance/metadata/v1/objectstore.proto](#provenance/metadata/v1/objectstore.proto)
    - [OSLocatorParams](#provenance.metadata.v1.OSLocatorParams)
    - [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator)
  
- [provenance/metadata/v1/genesis.proto](#provenance/metadata/v1/genesis.proto)
    - [GenesisState](#provenance.metadata.v1.GenesisState)
  
- [provenance/metadata/v1/p8e/p8e.proto](#provenance/metadata/v1/p8e/p8e.proto)
    - [Condition](#provenance.metadata.v1.p8e.Condition)
    - [ConditionSpec](#provenance.metadata.v1.p8e.ConditionSpec)
    - [Consideration](#provenance.metadata.v1.p8e.Consideration)
    - [ConsiderationSpec](#provenance.metadata.v1.p8e.ConsiderationSpec)
    - [Contract](#provenance.metadata.v1.p8e.Contract)
    - [ContractSpec](#provenance.metadata.v1.p8e.ContractSpec)
    - [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec)
    - [ExecutionResult](#provenance.metadata.v1.p8e.ExecutionResult)
    - [Fact](#provenance.metadata.v1.p8e.Fact)
    - [Location](#provenance.metadata.v1.p8e.Location)
    - [OutputSpec](#provenance.metadata.v1.p8e.OutputSpec)
    - [ProposedFact](#provenance.metadata.v1.p8e.ProposedFact)
    - [ProvenanceReference](#provenance.metadata.v1.p8e.ProvenanceReference)
    - [PublicKey](#provenance.metadata.v1.p8e.PublicKey)
    - [Recital](#provenance.metadata.v1.p8e.Recital)
    - [Recitals](#provenance.metadata.v1.p8e.Recitals)
    - [Signature](#provenance.metadata.v1.p8e.Signature)
    - [SignatureSet](#provenance.metadata.v1.p8e.SignatureSet)
    - [SigningAndEncryptionPublicKeys](#provenance.metadata.v1.p8e.SigningAndEncryptionPublicKeys)
    - [Timestamp](#provenance.metadata.v1.p8e.Timestamp)
    - [UUID](#provenance.metadata.v1.p8e.UUID)
  
    - [DefinitionSpecType](#provenance.metadata.v1.p8e.DefinitionSpecType)
    - [ExecutionResultType](#provenance.metadata.v1.p8e.ExecutionResultType)
    - [PartyType](#provenance.metadata.v1.p8e.PartyType)
    - [PublicKeyCurve](#provenance.metadata.v1.p8e.PublicKeyCurve)
    - [PublicKeyType](#provenance.metadata.v1.p8e.PublicKeyType)
  
- [provenance/metadata/v1/query.proto](#provenance/metadata/v1/query.proto)
    - [AccountDataRequest](#provenance.metadata.v1.AccountDataRequest)
    - [AccountDataResponse](#provenance.metadata.v1.AccountDataResponse)
    - [ContractSpecificationRequest](#provenance.metadata.v1.ContractSpecificationRequest)
    - [ContractSpecificationResponse](#provenance.metadata.v1.ContractSpecificationResponse)
    - [ContractSpecificationWrapper](#provenance.metadata.v1.ContractSpecificationWrapper)
    - [ContractSpecificationsAllRequest](#provenance.metadata.v1.ContractSpecificationsAllRequest)
    - [ContractSpecificationsAllResponse](#provenance.metadata.v1.ContractSpecificationsAllResponse)
    - [GetByAddrRequest](#provenance.metadata.v1.GetByAddrRequest)
    - [GetByAddrResponse](#provenance.metadata.v1.GetByAddrResponse)
    - [OSAllLocatorsRequest](#provenance.metadata.v1.OSAllLocatorsRequest)
    - [OSAllLocatorsResponse](#provenance.metadata.v1.OSAllLocatorsResponse)
    - [OSLocatorParamsRequest](#provenance.metadata.v1.OSLocatorParamsRequest)
    - [OSLocatorParamsResponse](#provenance.metadata.v1.OSLocatorParamsResponse)
    - [OSLocatorRequest](#provenance.metadata.v1.OSLocatorRequest)
    - [OSLocatorResponse](#provenance.metadata.v1.OSLocatorResponse)
    - [OSLocatorsByScopeRequest](#provenance.metadata.v1.OSLocatorsByScopeRequest)
    - [OSLocatorsByScopeResponse](#provenance.metadata.v1.OSLocatorsByScopeResponse)
    - [OSLocatorsByURIRequest](#provenance.metadata.v1.OSLocatorsByURIRequest)
    - [OSLocatorsByURIResponse](#provenance.metadata.v1.OSLocatorsByURIResponse)
    - [OwnershipRequest](#provenance.metadata.v1.OwnershipRequest)
    - [OwnershipResponse](#provenance.metadata.v1.OwnershipResponse)
    - [QueryParamsRequest](#provenance.metadata.v1.QueryParamsRequest)
    - [QueryParamsResponse](#provenance.metadata.v1.QueryParamsResponse)
    - [RecordSpecificationRequest](#provenance.metadata.v1.RecordSpecificationRequest)
    - [RecordSpecificationResponse](#provenance.metadata.v1.RecordSpecificationResponse)
    - [RecordSpecificationWrapper](#provenance.metadata.v1.RecordSpecificationWrapper)
    - [RecordSpecificationsAllRequest](#provenance.metadata.v1.RecordSpecificationsAllRequest)
    - [RecordSpecificationsAllResponse](#provenance.metadata.v1.RecordSpecificationsAllResponse)
    - [RecordSpecificationsForContractSpecificationRequest](#provenance.metadata.v1.RecordSpecificationsForContractSpecificationRequest)
    - [RecordSpecificationsForContractSpecificationResponse](#provenance.metadata.v1.RecordSpecificationsForContractSpecificationResponse)
    - [RecordWrapper](#provenance.metadata.v1.RecordWrapper)
    - [RecordsAllRequest](#provenance.metadata.v1.RecordsAllRequest)
    - [RecordsAllResponse](#provenance.metadata.v1.RecordsAllResponse)
    - [RecordsRequest](#provenance.metadata.v1.RecordsRequest)
    - [RecordsResponse](#provenance.metadata.v1.RecordsResponse)
    - [ScopeRequest](#provenance.metadata.v1.ScopeRequest)
    - [ScopeResponse](#provenance.metadata.v1.ScopeResponse)
    - [ScopeSpecificationRequest](#provenance.metadata.v1.ScopeSpecificationRequest)
    - [ScopeSpecificationResponse](#provenance.metadata.v1.ScopeSpecificationResponse)
    - [ScopeSpecificationWrapper](#provenance.metadata.v1.ScopeSpecificationWrapper)
    - [ScopeSpecificationsAllRequest](#provenance.metadata.v1.ScopeSpecificationsAllRequest)
    - [ScopeSpecificationsAllResponse](#provenance.metadata.v1.ScopeSpecificationsAllResponse)
    - [ScopeWrapper](#provenance.metadata.v1.ScopeWrapper)
    - [ScopesAllRequest](#provenance.metadata.v1.ScopesAllRequest)
    - [ScopesAllResponse](#provenance.metadata.v1.ScopesAllResponse)
    - [SessionWrapper](#provenance.metadata.v1.SessionWrapper)
    - [SessionsAllRequest](#provenance.metadata.v1.SessionsAllRequest)
    - [SessionsAllResponse](#provenance.metadata.v1.SessionsAllResponse)
    - [SessionsRequest](#provenance.metadata.v1.SessionsRequest)
    - [SessionsResponse](#provenance.metadata.v1.SessionsResponse)
    - [ValueOwnershipRequest](#provenance.metadata.v1.ValueOwnershipRequest)
    - [ValueOwnershipResponse](#provenance.metadata.v1.ValueOwnershipResponse)
  
    - [Query](#provenance.metadata.v1.Query)
  
- [provenance/metadata/v1/tx.proto](#provenance/metadata/v1/tx.proto)
    - [MsgAddContractSpecToScopeSpecRequest](#provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest)
    - [MsgAddContractSpecToScopeSpecResponse](#provenance.metadata.v1.MsgAddContractSpecToScopeSpecResponse)
    - [MsgAddScopeDataAccessRequest](#provenance.metadata.v1.MsgAddScopeDataAccessRequest)
    - [MsgAddScopeDataAccessResponse](#provenance.metadata.v1.MsgAddScopeDataAccessResponse)
    - [MsgAddScopeOwnerRequest](#provenance.metadata.v1.MsgAddScopeOwnerRequest)
    - [MsgAddScopeOwnerResponse](#provenance.metadata.v1.MsgAddScopeOwnerResponse)
    - [MsgBindOSLocatorRequest](#provenance.metadata.v1.MsgBindOSLocatorRequest)
    - [MsgBindOSLocatorResponse](#provenance.metadata.v1.MsgBindOSLocatorResponse)
    - [MsgDeleteContractSpecFromScopeSpecRequest](#provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecRequest)
    - [MsgDeleteContractSpecFromScopeSpecResponse](#provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecResponse)
    - [MsgDeleteContractSpecificationRequest](#provenance.metadata.v1.MsgDeleteContractSpecificationRequest)
    - [MsgDeleteContractSpecificationResponse](#provenance.metadata.v1.MsgDeleteContractSpecificationResponse)
    - [MsgDeleteOSLocatorRequest](#provenance.metadata.v1.MsgDeleteOSLocatorRequest)
    - [MsgDeleteOSLocatorResponse](#provenance.metadata.v1.MsgDeleteOSLocatorResponse)
    - [MsgDeleteRecordRequest](#provenance.metadata.v1.MsgDeleteRecordRequest)
    - [MsgDeleteRecordResponse](#provenance.metadata.v1.MsgDeleteRecordResponse)
    - [MsgDeleteRecordSpecificationRequest](#provenance.metadata.v1.MsgDeleteRecordSpecificationRequest)
    - [MsgDeleteRecordSpecificationResponse](#provenance.metadata.v1.MsgDeleteRecordSpecificationResponse)
    - [MsgDeleteScopeDataAccessRequest](#provenance.metadata.v1.MsgDeleteScopeDataAccessRequest)
    - [MsgDeleteScopeDataAccessResponse](#provenance.metadata.v1.MsgDeleteScopeDataAccessResponse)
    - [MsgDeleteScopeOwnerRequest](#provenance.metadata.v1.MsgDeleteScopeOwnerRequest)
    - [MsgDeleteScopeOwnerResponse](#provenance.metadata.v1.MsgDeleteScopeOwnerResponse)
    - [MsgDeleteScopeRequest](#provenance.metadata.v1.MsgDeleteScopeRequest)
    - [MsgDeleteScopeResponse](#provenance.metadata.v1.MsgDeleteScopeResponse)
    - [MsgDeleteScopeSpecificationRequest](#provenance.metadata.v1.MsgDeleteScopeSpecificationRequest)
    - [MsgDeleteScopeSpecificationResponse](#provenance.metadata.v1.MsgDeleteScopeSpecificationResponse)
    - [MsgMigrateValueOwnerRequest](#provenance.metadata.v1.MsgMigrateValueOwnerRequest)
    - [MsgMigrateValueOwnerResponse](#provenance.metadata.v1.MsgMigrateValueOwnerResponse)
    - [MsgModifyOSLocatorRequest](#provenance.metadata.v1.MsgModifyOSLocatorRequest)
    - [MsgModifyOSLocatorResponse](#provenance.metadata.v1.MsgModifyOSLocatorResponse)
    - [MsgP8eMemorializeContractRequest](#provenance.metadata.v1.MsgP8eMemorializeContractRequest)
    - [MsgP8eMemorializeContractResponse](#provenance.metadata.v1.MsgP8eMemorializeContractResponse)
    - [MsgSetAccountDataRequest](#provenance.metadata.v1.MsgSetAccountDataRequest)
    - [MsgSetAccountDataResponse](#provenance.metadata.v1.MsgSetAccountDataResponse)
    - [MsgUpdateValueOwnersRequest](#provenance.metadata.v1.MsgUpdateValueOwnersRequest)
    - [MsgUpdateValueOwnersResponse](#provenance.metadata.v1.MsgUpdateValueOwnersResponse)
    - [MsgWriteContractSpecificationRequest](#provenance.metadata.v1.MsgWriteContractSpecificationRequest)
    - [MsgWriteContractSpecificationResponse](#provenance.metadata.v1.MsgWriteContractSpecificationResponse)
    - [MsgWriteP8eContractSpecRequest](#provenance.metadata.v1.MsgWriteP8eContractSpecRequest)
    - [MsgWriteP8eContractSpecResponse](#provenance.metadata.v1.MsgWriteP8eContractSpecResponse)
    - [MsgWriteRecordRequest](#provenance.metadata.v1.MsgWriteRecordRequest)
    - [MsgWriteRecordResponse](#provenance.metadata.v1.MsgWriteRecordResponse)
    - [MsgWriteRecordSpecificationRequest](#provenance.metadata.v1.MsgWriteRecordSpecificationRequest)
    - [MsgWriteRecordSpecificationResponse](#provenance.metadata.v1.MsgWriteRecordSpecificationResponse)
    - [MsgWriteScopeRequest](#provenance.metadata.v1.MsgWriteScopeRequest)
    - [MsgWriteScopeResponse](#provenance.metadata.v1.MsgWriteScopeResponse)
    - [MsgWriteScopeSpecificationRequest](#provenance.metadata.v1.MsgWriteScopeSpecificationRequest)
    - [MsgWriteScopeSpecificationResponse](#provenance.metadata.v1.MsgWriteScopeSpecificationResponse)
    - [MsgWriteSessionRequest](#provenance.metadata.v1.MsgWriteSessionRequest)
    - [MsgWriteSessionResponse](#provenance.metadata.v1.MsgWriteSessionResponse)
    - [SessionIdComponents](#provenance.metadata.v1.SessionIdComponents)
  
    - [Msg](#provenance.metadata.v1.Msg)
  
- [provenance/msgfees/v1/msgfees.proto](#provenance/msgfees/v1/msgfees.proto)
    - [EventMsgFee](#provenance.msgfees.v1.EventMsgFee)
    - [EventMsgFees](#provenance.msgfees.v1.EventMsgFees)
    - [MsgFee](#provenance.msgfees.v1.MsgFee)
    - [Params](#provenance.msgfees.v1.Params)
  
- [provenance/msgfees/v1/genesis.proto](#provenance/msgfees/v1/genesis.proto)
    - [GenesisState](#provenance.msgfees.v1.GenesisState)
  
- [provenance/msgfees/v1/proposals.proto](#provenance/msgfees/v1/proposals.proto)
    - [AddMsgFeeProposal](#provenance.msgfees.v1.AddMsgFeeProposal)
    - [RemoveMsgFeeProposal](#provenance.msgfees.v1.RemoveMsgFeeProposal)
    - [UpdateConversionFeeDenomProposal](#provenance.msgfees.v1.UpdateConversionFeeDenomProposal)
    - [UpdateMsgFeeProposal](#provenance.msgfees.v1.UpdateMsgFeeProposal)
    - [UpdateNhashPerUsdMilProposal](#provenance.msgfees.v1.UpdateNhashPerUsdMilProposal)
  
- [provenance/msgfees/v1/query.proto](#provenance/msgfees/v1/query.proto)
    - [CalculateTxFeesRequest](#provenance.msgfees.v1.CalculateTxFeesRequest)
    - [CalculateTxFeesResponse](#provenance.msgfees.v1.CalculateTxFeesResponse)
    - [QueryAllMsgFeesRequest](#provenance.msgfees.v1.QueryAllMsgFeesRequest)
    - [QueryAllMsgFeesResponse](#provenance.msgfees.v1.QueryAllMsgFeesResponse)
    - [QueryParamsRequest](#provenance.msgfees.v1.QueryParamsRequest)
    - [QueryParamsResponse](#provenance.msgfees.v1.QueryParamsResponse)
  
    - [Query](#provenance.msgfees.v1.Query)
  
- [provenance/msgfees/v1/tx.proto](#provenance/msgfees/v1/tx.proto)
    - [MsgAddMsgFeeProposalRequest](#provenance.msgfees.v1.MsgAddMsgFeeProposalRequest)
    - [MsgAddMsgFeeProposalResponse](#provenance.msgfees.v1.MsgAddMsgFeeProposalResponse)
    - [MsgAssessCustomMsgFeeRequest](#provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest)
    - [MsgAssessCustomMsgFeeResponse](#provenance.msgfees.v1.MsgAssessCustomMsgFeeResponse)
    - [MsgRemoveMsgFeeProposalRequest](#provenance.msgfees.v1.MsgRemoveMsgFeeProposalRequest)
    - [MsgRemoveMsgFeeProposalResponse](#provenance.msgfees.v1.MsgRemoveMsgFeeProposalResponse)
    - [MsgUpdateConversionFeeDenomProposalRequest](#provenance.msgfees.v1.MsgUpdateConversionFeeDenomProposalRequest)
    - [MsgUpdateConversionFeeDenomProposalResponse](#provenance.msgfees.v1.MsgUpdateConversionFeeDenomProposalResponse)
    - [MsgUpdateMsgFeeProposalRequest](#provenance.msgfees.v1.MsgUpdateMsgFeeProposalRequest)
    - [MsgUpdateMsgFeeProposalResponse](#provenance.msgfees.v1.MsgUpdateMsgFeeProposalResponse)
    - [MsgUpdateNhashPerUsdMilProposalRequest](#provenance.msgfees.v1.MsgUpdateNhashPerUsdMilProposalRequest)
    - [MsgUpdateNhashPerUsdMilProposalResponse](#provenance.msgfees.v1.MsgUpdateNhashPerUsdMilProposalResponse)
  
    - [Msg](#provenance.msgfees.v1.Msg)
  
- [provenance/name/v1/name.proto](#provenance/name/v1/name.proto)
    - [CreateRootNameProposal](#provenance.name.v1.CreateRootNameProposal)
    - [EventNameBound](#provenance.name.v1.EventNameBound)
    - [EventNameUnbound](#provenance.name.v1.EventNameUnbound)
    - [EventNameUpdate](#provenance.name.v1.EventNameUpdate)
    - [NameRecord](#provenance.name.v1.NameRecord)
    - [Params](#provenance.name.v1.Params)
  
- [provenance/name/v1/genesis.proto](#provenance/name/v1/genesis.proto)
    - [GenesisState](#provenance.name.v1.GenesisState)
  
- [provenance/name/v1/query.proto](#provenance/name/v1/query.proto)
    - [QueryParamsRequest](#provenance.name.v1.QueryParamsRequest)
    - [QueryParamsResponse](#provenance.name.v1.QueryParamsResponse)
    - [QueryResolveRequest](#provenance.name.v1.QueryResolveRequest)
    - [QueryResolveResponse](#provenance.name.v1.QueryResolveResponse)
    - [QueryReverseLookupRequest](#provenance.name.v1.QueryReverseLookupRequest)
    - [QueryReverseLookupResponse](#provenance.name.v1.QueryReverseLookupResponse)
  
    - [Query](#provenance.name.v1.Query)
  
- [provenance/name/v1/tx.proto](#provenance/name/v1/tx.proto)
    - [MsgBindNameRequest](#provenance.name.v1.MsgBindNameRequest)
    - [MsgBindNameResponse](#provenance.name.v1.MsgBindNameResponse)
    - [MsgCreateRootNameRequest](#provenance.name.v1.MsgCreateRootNameRequest)
    - [MsgCreateRootNameResponse](#provenance.name.v1.MsgCreateRootNameResponse)
    - [MsgDeleteNameRequest](#provenance.name.v1.MsgDeleteNameRequest)
    - [MsgDeleteNameResponse](#provenance.name.v1.MsgDeleteNameResponse)
    - [MsgModifyNameRequest](#provenance.name.v1.MsgModifyNameRequest)
    - [MsgModifyNameResponse](#provenance.name.v1.MsgModifyNameResponse)
  
    - [Msg](#provenance.name.v1.Msg)
  
- [provenance/reward/v1/reward.proto](#provenance/reward/v1/reward.proto)
    - [ActionCounter](#provenance.reward.v1.ActionCounter)
    - [ActionDelegate](#provenance.reward.v1.ActionDelegate)
    - [ActionTransfer](#provenance.reward.v1.ActionTransfer)
    - [ActionVote](#provenance.reward.v1.ActionVote)
    - [ClaimPeriodRewardDistribution](#provenance.reward.v1.ClaimPeriodRewardDistribution)
    - [QualifyingAction](#provenance.reward.v1.QualifyingAction)
    - [QualifyingActions](#provenance.reward.v1.QualifyingActions)
    - [RewardAccountState](#provenance.reward.v1.RewardAccountState)
    - [RewardProgram](#provenance.reward.v1.RewardProgram)
  
    - [RewardAccountState.ClaimStatus](#provenance.reward.v1.RewardAccountState.ClaimStatus)
    - [RewardProgram.State](#provenance.reward.v1.RewardProgram.State)
  
- [provenance/reward/v1/genesis.proto](#provenance/reward/v1/genesis.proto)
    - [GenesisState](#provenance.reward.v1.GenesisState)
  
- [provenance/reward/v1/query.proto](#provenance/reward/v1/query.proto)
    - [QueryClaimPeriodRewardDistributionsByIDRequest](#provenance.reward.v1.QueryClaimPeriodRewardDistributionsByIDRequest)
    - [QueryClaimPeriodRewardDistributionsByIDResponse](#provenance.reward.v1.QueryClaimPeriodRewardDistributionsByIDResponse)
    - [QueryClaimPeriodRewardDistributionsRequest](#provenance.reward.v1.QueryClaimPeriodRewardDistributionsRequest)
    - [QueryClaimPeriodRewardDistributionsResponse](#provenance.reward.v1.QueryClaimPeriodRewardDistributionsResponse)
    - [QueryRewardDistributionsByAddressRequest](#provenance.reward.v1.QueryRewardDistributionsByAddressRequest)
    - [QueryRewardDistributionsByAddressResponse](#provenance.reward.v1.QueryRewardDistributionsByAddressResponse)
    - [QueryRewardProgramByIDRequest](#provenance.reward.v1.QueryRewardProgramByIDRequest)
    - [QueryRewardProgramByIDResponse](#provenance.reward.v1.QueryRewardProgramByIDResponse)
    - [QueryRewardProgramsRequest](#provenance.reward.v1.QueryRewardProgramsRequest)
    - [QueryRewardProgramsResponse](#provenance.reward.v1.QueryRewardProgramsResponse)
    - [RewardAccountResponse](#provenance.reward.v1.RewardAccountResponse)
  
    - [QueryRewardProgramsRequest.QueryType](#provenance.reward.v1.QueryRewardProgramsRequest.QueryType)
  
    - [Query](#provenance.reward.v1.Query)
  
- [provenance/reward/v1/tx.proto](#provenance/reward/v1/tx.proto)
    - [ClaimedRewardPeriodDetail](#provenance.reward.v1.ClaimedRewardPeriodDetail)
    - [MsgClaimAllRewardsRequest](#provenance.reward.v1.MsgClaimAllRewardsRequest)
    - [MsgClaimAllRewardsResponse](#provenance.reward.v1.MsgClaimAllRewardsResponse)
    - [MsgClaimRewardsRequest](#provenance.reward.v1.MsgClaimRewardsRequest)
    - [MsgClaimRewardsResponse](#provenance.reward.v1.MsgClaimRewardsResponse)
    - [MsgCreateRewardProgramRequest](#provenance.reward.v1.MsgCreateRewardProgramRequest)
    - [MsgCreateRewardProgramResponse](#provenance.reward.v1.MsgCreateRewardProgramResponse)
    - [MsgEndRewardProgramRequest](#provenance.reward.v1.MsgEndRewardProgramRequest)
    - [MsgEndRewardProgramResponse](#provenance.reward.v1.MsgEndRewardProgramResponse)
    - [RewardProgramClaimDetail](#provenance.reward.v1.RewardProgramClaimDetail)
  
    - [Msg](#provenance.reward.v1.Msg)
  
- [provenance/trigger/v1/event.proto](#provenance/trigger/v1/event.proto)
    - [EventTriggerCreated](#provenance.trigger.v1.EventTriggerCreated)
    - [EventTriggerDestroyed](#provenance.trigger.v1.EventTriggerDestroyed)
  
- [provenance/trigger/v1/trigger.proto](#provenance/trigger/v1/trigger.proto)
    - [Attribute](#provenance.trigger.v1.Attribute)
    - [BlockHeightEvent](#provenance.trigger.v1.BlockHeightEvent)
    - [BlockTimeEvent](#provenance.trigger.v1.BlockTimeEvent)
    - [QueuedTrigger](#provenance.trigger.v1.QueuedTrigger)
    - [TransactionEvent](#provenance.trigger.v1.TransactionEvent)
    - [Trigger](#provenance.trigger.v1.Trigger)
  
- [provenance/trigger/v1/genesis.proto](#provenance/trigger/v1/genesis.proto)
    - [GasLimit](#provenance.trigger.v1.GasLimit)
    - [GenesisState](#provenance.trigger.v1.GenesisState)
  
- [provenance/trigger/v1/query.proto](#provenance/trigger/v1/query.proto)
    - [QueryTriggerByIDRequest](#provenance.trigger.v1.QueryTriggerByIDRequest)
    - [QueryTriggerByIDResponse](#provenance.trigger.v1.QueryTriggerByIDResponse)
    - [QueryTriggersRequest](#provenance.trigger.v1.QueryTriggersRequest)
    - [QueryTriggersResponse](#provenance.trigger.v1.QueryTriggersResponse)
  
    - [Query](#provenance.trigger.v1.Query)
  
- [provenance/trigger/v1/tx.proto](#provenance/trigger/v1/tx.proto)
    - [MsgCreateTriggerRequest](#provenance.trigger.v1.MsgCreateTriggerRequest)
    - [MsgCreateTriggerResponse](#provenance.trigger.v1.MsgCreateTriggerResponse)
    - [MsgDestroyTriggerRequest](#provenance.trigger.v1.MsgDestroyTriggerRequest)
    - [MsgDestroyTriggerResponse](#provenance.trigger.v1.MsgDestroyTriggerResponse)
  
    - [Msg](#provenance.trigger.v1.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="provenance/attribute/v1/attribute.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/attribute/v1/attribute.proto



<a name="provenance.attribute.v1.Attribute"></a>

### Attribute
Attribute holds a typed key/value structure for data associated with an account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The attribute value. |
| `attribute_type` | [AttributeType](#provenance.attribute.v1.AttributeType) |  | The attribute value type. |
| `address` | [string](#string) |  | The address the attribute is bound to |
| `expiration_date` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time that an attribute will expire. |






<a name="provenance.attribute.v1.EventAccountDataUpdated"></a>

### EventAccountDataUpdated
EventAccountDataUpdated event emitted when accountdata is set, updated, or deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  |  |






<a name="provenance.attribute.v1.EventAttributeAdd"></a>

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






<a name="provenance.attribute.v1.EventAttributeDelete"></a>

### EventAttributeDelete
EventAttributeDelete event emitted when attribute is deleted


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="provenance.attribute.v1.EventAttributeDistinctDelete"></a>

### EventAttributeDistinctDelete
EventAttributeDistinctDelete event emitted when attribute is deleted with matching value


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `value` | [string](#string) |  |  |
| `attribute_type` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |






<a name="provenance.attribute.v1.EventAttributeExpirationUpdate"></a>

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






<a name="provenance.attribute.v1.EventAttributeExpired"></a>

### EventAttributeExpired
EventAttributeExpired event emitted when attribute has expired and been deleted in BeginBlocker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `value_hash` | [string](#string) |  |  |
| `attribute_type` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |
| `expiration` | [string](#string) |  |  |






<a name="provenance.attribute.v1.EventAttributeUpdate"></a>

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






<a name="provenance.attribute.v1.Params"></a>

### Params
Params defines the set of params for the attribute module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_value_length` | [uint32](#uint32) |  | maximum length of data to allow in an attribute value |





 <!-- end messages -->


<a name="provenance.attribute.v1.AttributeType"></a>

### AttributeType
AttributeType defines the type of the data stored in the attribute value

| Name | Number | Description |
| ---- | ------ | ----------- |
| ATTRIBUTE_TYPE_UNSPECIFIED | 0 | ATTRIBUTE_TYPE_UNSPECIFIED defines an unknown/invalid type |
| ATTRIBUTE_TYPE_UUID | 1 | ATTRIBUTE_TYPE_UUID defines an attribute value that contains a string value representation of a V4 uuid |
| ATTRIBUTE_TYPE_JSON | 2 | ATTRIBUTE_TYPE_JSON defines an attribute value that contains a byte string containing json data |
| ATTRIBUTE_TYPE_STRING | 3 | ATTRIBUTE_TYPE_STRING defines an attribute value that contains a generic string value |
| ATTRIBUTE_TYPE_URI | 4 | ATTRIBUTE_TYPE_URI defines an attribute value that contains a URI |
| ATTRIBUTE_TYPE_INT | 5 | ATTRIBUTE_TYPE_INT defines an attribute value that contains an integer (cast as int64) |
| ATTRIBUTE_TYPE_FLOAT | 6 | ATTRIBUTE_TYPE_FLOAT defines an attribute value that contains a float |
| ATTRIBUTE_TYPE_PROTO | 7 | ATTRIBUTE_TYPE_PROTO defines an attribute value that contains a serialized proto value in bytes |
| ATTRIBUTE_TYPE_BYTES | 8 | ATTRIBUTE_TYPE_BYTES defines an attribute value that contains an untyped array of bytes |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/attribute/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/attribute/v1/genesis.proto



<a name="provenance.attribute.v1.GenesisState"></a>

### GenesisState
GenesisState defines the attribute module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.attribute.v1.Params) |  | params defines all the parameters of the module. |
| `attributes` | [Attribute](#provenance.attribute.v1.Attribute) | repeated | deposits defines all the deposits present at genesis. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/attribute/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/attribute/v1/query.proto



<a name="provenance.attribute.v1.QueryAccountDataRequest"></a>

### QueryAccountDataRequest
QueryAccountDataRequest is the request type for the Query/AccountData method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account is the bech32 address of the account to get the data for |






<a name="provenance.attribute.v1.QueryAccountDataResponse"></a>

### QueryAccountDataResponse
QueryAccountDataResponse is the response type for the Query/AccountData method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  | value is the accountdata attribute value for the requested account. |






<a name="provenance.attribute.v1.QueryAttributeAccountsRequest"></a>

### QueryAttributeAccountsRequest
QueryAttributeAccountsRequest is the request type for the Query/AttributeAccounts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `attribute_name` | [string](#string) |  | name is the attribute name to query for |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.attribute.v1.QueryAttributeAccountsResponse"></a>

### QueryAttributeAccountsResponse
QueryAttributeAccountsResponse is the response type for the Query/AttributeAccounts method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `accounts` | [string](#string) | repeated | list of account addresses that have attributes of request name |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance.attribute.v1.QueryAttributeRequest"></a>

### QueryAttributeRequest
QueryAttributeRequest is the request type for the Query/Attribute method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account defines the address to query for. |
| `name` | [string](#string) |  | name is the attribute name to query for |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.attribute.v1.QueryAttributeResponse"></a>

### QueryAttributeResponse
QueryAttributeResponse is the response type for the Query/Attribute method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | a string containing the address of the account the attributes are assigned to. |
| `attributes` | [Attribute](#provenance.attribute.v1.Attribute) | repeated | a list of attribute values |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance.attribute.v1.QueryAttributesRequest"></a>

### QueryAttributesRequest
QueryAttributesRequest is the request type for the Query/Attributes method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account defines the address to query for. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.attribute.v1.QueryAttributesResponse"></a>

### QueryAttributesResponse
QueryAttributesResponse is the response type for the Query/Attributes method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | a string containing the address of the account the attributes are assigned to= |
| `attributes` | [Attribute](#provenance.attribute.v1.Attribute) | repeated | a list of attribute values |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance.attribute.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance.attribute.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.attribute.v1.Params) |  | params defines the parameters of the module. |






<a name="provenance.attribute.v1.QueryScanRequest"></a>

### QueryScanRequest
QueryScanRequest is the request type for the Query/Scan method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account defines the address to query for. |
| `suffix` | [string](#string) |  | name defines the partial attribute name to search for base on names being in RDNS format. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.attribute.v1.QueryScanResponse"></a>

### QueryScanResponse
QueryScanResponse is the response type for the Query/Scan method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | a string containing the address of the account the attributes are assigned to= |
| `attributes` | [Attribute](#provenance.attribute.v1.Attribute) | repeated | a list of attribute values |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the request. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.attribute.v1.Query"></a>

### Query
Query defines the gRPC querier service for attribute module.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#provenance.attribute.v1.QueryParamsRequest) | [QueryParamsResponse](#provenance.attribute.v1.QueryParamsResponse) | Params queries params of the attribute module. | GET|/provenance/attribute/v1/params|
| `Attribute` | [QueryAttributeRequest](#provenance.attribute.v1.QueryAttributeRequest) | [QueryAttributeResponse](#provenance.attribute.v1.QueryAttributeResponse) | Attribute queries attributes on a given account (address) for one (or more) with the given name | GET|/provenance/attribute/v1/attribute/{account}/{name}|
| `Attributes` | [QueryAttributesRequest](#provenance.attribute.v1.QueryAttributesRequest) | [QueryAttributesResponse](#provenance.attribute.v1.QueryAttributesResponse) | Attributes queries attributes on a given account (address) for any defined attributes | GET|/provenance/attribute/v1/attributes/{account}|
| `Scan` | [QueryScanRequest](#provenance.attribute.v1.QueryScanRequest) | [QueryScanResponse](#provenance.attribute.v1.QueryScanResponse) | Scan queries attributes on a given account (address) for any that match the provided suffix | GET|/provenance/attribute/v1/attribute/{account}/scan/{suffix}|
| `AttributeAccounts` | [QueryAttributeAccountsRequest](#provenance.attribute.v1.QueryAttributeAccountsRequest) | [QueryAttributeAccountsResponse](#provenance.attribute.v1.QueryAttributeAccountsResponse) | AttributeAccounts queries accounts on a given attribute name | GET|/provenance/attribute/v1/accounts/{attribute_name}|
| `AccountData` | [QueryAccountDataRequest](#provenance.attribute.v1.QueryAccountDataRequest) | [QueryAccountDataResponse](#provenance.attribute.v1.QueryAccountDataResponse) | AccountData returns the accountdata for a specified account. | GET|/provenance/attribute/v1/accountdata/{account}|

 <!-- end services -->



<a name="provenance/attribute/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/attribute/v1/tx.proto



<a name="provenance.attribute.v1.MsgAddAttributeRequest"></a>

### MsgAddAttributeRequest
MsgAddAttributeRequest defines an sdk.Msg type that is used to add a new attribute to an account.
Attributes may only be set in an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The attribute value. |
| `attribute_type` | [AttributeType](#provenance.attribute.v1.AttributeType) |  | The attribute value type. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |
| `expiration_date` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time that an attribute will expire. |






<a name="provenance.attribute.v1.MsgAddAttributeResponse"></a>

### MsgAddAttributeResponse
MsgAddAttributeResponse defines the Msg/AddAttribute response type.






<a name="provenance.attribute.v1.MsgDeleteAttributeRequest"></a>

### MsgDeleteAttributeRequest
MsgDeleteAttributeRequest defines a message to delete an attribute from an account
Attributes may only be removed from an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance.attribute.v1.MsgDeleteAttributeResponse"></a>

### MsgDeleteAttributeResponse
MsgDeleteAttributeResponse defines the Msg/DeleteAttribute response type.






<a name="provenance.attribute.v1.MsgDeleteDistinctAttributeRequest"></a>

### MsgDeleteDistinctAttributeRequest
MsgDeleteDistinctAttributeRequest defines a message to delete an attribute with matching name, value, and type from
an account. Attributes may only be removed from an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The attribute value. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance.attribute.v1.MsgDeleteDistinctAttributeResponse"></a>

### MsgDeleteDistinctAttributeResponse
MsgDeleteDistinctAttributeResponse defines the Msg/DeleteDistinctAttribute response type.






<a name="provenance.attribute.v1.MsgSetAccountDataRequest"></a>

### MsgSetAccountDataRequest
MsgSetAccountDataRequest defines a message to set an account's accountdata attribute.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  |  |
| `account` | [string](#string) |  |  |






<a name="provenance.attribute.v1.MsgSetAccountDataResponse"></a>

### MsgSetAccountDataResponse
MsgSetAccountDataResponse defines the Msg/SetAccountData response type.






<a name="provenance.attribute.v1.MsgUpdateAttributeExpirationRequest"></a>

### MsgUpdateAttributeExpirationRequest
MsgUpdateAttributeExpirationRequest defines an sdk.Msg type that is used to update an existing attribute's expiration
date


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The original attribute value. |
| `expiration_date` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time that an attribute will expire. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance.attribute.v1.MsgUpdateAttributeExpirationResponse"></a>

### MsgUpdateAttributeExpirationResponse
MsgUpdateAttributeExpirationResponse defines the Msg/Vote response type.






<a name="provenance.attribute.v1.MsgUpdateAttributeRequest"></a>

### MsgUpdateAttributeRequest
MsgUpdateAttributeRequest defines an sdk.Msg type that is used to update an existing attribute to an account.
Attributes may only be set in an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `original_value` | [bytes](#bytes) |  | The original attribute value. |
| `update_value` | [bytes](#bytes) |  | The update attribute value. |
| `original_attribute_type` | [AttributeType](#provenance.attribute.v1.AttributeType) |  | The original attribute value type. |
| `update_attribute_type` | [AttributeType](#provenance.attribute.v1.AttributeType) |  | The update attribute value type. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance.attribute.v1.MsgUpdateAttributeResponse"></a>

### MsgUpdateAttributeResponse
MsgUpdateAttributeResponse defines the Msg/UpdateAttribute response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.attribute.v1.Msg"></a>

### Msg
Msg defines the attribute module Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `AddAttribute` | [MsgAddAttributeRequest](#provenance.attribute.v1.MsgAddAttributeRequest) | [MsgAddAttributeResponse](#provenance.attribute.v1.MsgAddAttributeResponse) | AddAttribute defines a method to verify a particular invariance. | |
| `UpdateAttribute` | [MsgUpdateAttributeRequest](#provenance.attribute.v1.MsgUpdateAttributeRequest) | [MsgUpdateAttributeResponse](#provenance.attribute.v1.MsgUpdateAttributeResponse) | UpdateAttribute defines a method to verify a particular invariance. | |
| `UpdateAttributeExpiration` | [MsgUpdateAttributeExpirationRequest](#provenance.attribute.v1.MsgUpdateAttributeExpirationRequest) | [MsgUpdateAttributeExpirationResponse](#provenance.attribute.v1.MsgUpdateAttributeExpirationResponse) | UpdateAttributeExpiration defines a method to verify a particular invariance. | |
| `DeleteAttribute` | [MsgDeleteAttributeRequest](#provenance.attribute.v1.MsgDeleteAttributeRequest) | [MsgDeleteAttributeResponse](#provenance.attribute.v1.MsgDeleteAttributeResponse) | DeleteAttribute defines a method to verify a particular invariance. | |
| `DeleteDistinctAttribute` | [MsgDeleteDistinctAttributeRequest](#provenance.attribute.v1.MsgDeleteDistinctAttributeRequest) | [MsgDeleteDistinctAttributeResponse](#provenance.attribute.v1.MsgDeleteDistinctAttributeResponse) | DeleteDistinctAttribute defines a method to verify a particular invariance. | |
| `SetAccountData` | [MsgSetAccountDataRequest](#provenance.attribute.v1.MsgSetAccountDataRequest) | [MsgSetAccountDataResponse](#provenance.attribute.v1.MsgSetAccountDataResponse) | SetAccountData defines a method for setting/updating an account's accountdata attribute. | |

 <!-- end services -->



<a name="provenance/exchange/v1/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/events.proto



<a name="provenance.exchange.v1.EventCreateMarketSubmitted"></a>

### EventCreateMarketSubmitted
EventCreateMarketSubmitted is an event emitted during CreateMarket indicating that a governance
proposal was submitted to create a market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `proposal_id` | [uint64](#uint64) |  | proposal_id is the identifier of the governance proposal that was submitted to create the market. |
| `submitted_by` | [string](#string) |  | submitted_by is the account that requested the creation of the market. |






<a name="provenance.exchange.v1.EventMarketCreated"></a>

### EventMarketCreated
EventMarketCreated is an event emitted when a market has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |






<a name="provenance.exchange.v1.EventMarketDetailsUpdated"></a>

### EventMarketDetailsUpdated
EventMarketDetailsUpdated is an event emitted when a market's details are updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the details. |






<a name="provenance.exchange.v1.EventMarketDisabled"></a>

### EventMarketDisabled
EventMarketDisabled is an event emitted when a market is disabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that disabled the market. |






<a name="provenance.exchange.v1.EventMarketEnabled"></a>

### EventMarketEnabled
EventMarketEnabled is an event emitted when a market is enabled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that enabled the market. |






<a name="provenance.exchange.v1.EventMarketFeesUpdated"></a>

### EventMarketFeesUpdated
EventMarketFeesUpdated is an event emitted when a market's fees have been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |






<a name="provenance.exchange.v1.EventMarketPermissionsUpdated"></a>

### EventMarketPermissionsUpdated
EventMarketPermissionsUpdated is an event emitted when a market's permissions are updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the permissions. |






<a name="provenance.exchange.v1.EventMarketReqAttrUpdated"></a>

### EventMarketReqAttrUpdated
EventMarketReqAttrUpdated is an event emitted when a market's required attributes are updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the required attributes. |






<a name="provenance.exchange.v1.EventMarketUserSettleUpdated"></a>

### EventMarketUserSettleUpdated
EventMarketUserSettleUpdated is an event emitted when a market's user_settle option is updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `updated_by` | [string](#string) |  | updated_by is the account that updated the user_settle option. |






<a name="provenance.exchange.v1.EventMarketWithdraw"></a>

### EventMarketWithdraw
EventMarketWithdraw is an event emitted when a withdrawal of a market's collected fees is made.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market. |
| `amount_withdrawn` | [string](#string) |  | amount_withdrawn is the coins amount string of funds withdrawn from the market account. |
| `destination` | [string](#string) |  | destination is the account that received the funds. |
| `withdrawn_by` | [string](#string) |  | withdrawn_by is the account that requested the withdrawal. |






<a name="provenance.exchange.v1.EventOrderCancelled"></a>

### EventOrderCancelled
EventOrderCancelled is an event emitted when an order is cancelled.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order created. |
| `cancelled_by` | [string](#string) |  | cancelled_by is the account that triggered the cancellation of the order. |






<a name="provenance.exchange.v1.EventOrderCreated"></a>

### EventOrderCreated
EventOrderCreated is an event emitted when an order is created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order created. |
| `order_type` | [string](#string) |  | order_type is the type of order, e.g. "ask" or "bid". |






<a name="provenance.exchange.v1.EventOrderFilled"></a>

### EventOrderFilled
EventOrderFilled is an event emitted when an order has been filled in full.
This event is also used for orders that were previously partially filled, but have now been filled in full.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order created. |






<a name="provenance.exchange.v1.EventOrderPartiallyFilled"></a>

### EventOrderPartiallyFilled
EventOrderPartiallyFilled is an event emitted when an order filled in part and still has more left to fill.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier of the order created. |
| `assets_filled` | [string](#string) |  | amount_filled is the coins amount string of assets that were filled (and removed from the order). |
| `fees_filled` | [string](#string) |  | fees_filled is the coins amount string of fees removed from the order. |






<a name="provenance.exchange.v1.EventParamsUpdated"></a>

### EventParamsUpdated
EventParamsUpdated is an event emitted when the exchange module's params have been updated.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/exchange/v1/market.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/market.proto



<a name="provenance.exchange.v1.AccessGrant"></a>

### AccessGrant
AddrPermissions associates an address with a list of permissions available for that address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address that these permissions apply to. |
| `permissions` | [Permission](#provenance.exchange.v1.Permission) | repeated | allowed is the list of permissions available for the address. |






<a name="provenance.exchange.v1.FeeRatio"></a>

### FeeRatio
FeeRatio defines a ratio of price amount to fee amount.
For an order to be valid, its price must be evenly divisible by a FeeRatio's price.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `price` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | price is the unit the order price is divided by to get how much of the fee should apply. |
| `fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | fee is the amount to charge per price unit. |






<a name="provenance.exchange.v1.Market"></a>

### Market
Market contains all information about a market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier for this market. |
| `market_details` | [MarketDetails](#provenance.exchange.v1.MarketDetails) |  | market_details is some information about this market. |
| `fee_create_ask_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | fee_create_ask_flat is the flat fee charged for creating an ask order. Each coin entry is a separate option. When an ask is created, one of these must be paid. If empty, no fee is required to create an ask order. |
| `fee_create_bid_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | fee_create_bid_flat is the flat fee charged for creating a bid order. Each coin entry is a separate option. When a bid is created, one of these must be paid. If empty, no fee is required to create a bid order. |
| `fee_seller_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | fee_seller_settlement_flat is the flat fee charged to the seller during settlement. Each coin entry is a separate option. When an ask is settled, the seller will pay the amount in the denom that matches the price they received. |
| `fee_seller_settlement_ratios` | [FeeRatio](#provenance.exchange.v1.FeeRatio) | repeated | fee_seller_settlement_ratios is the fee to charge a seller during settlement based on the price they are receiving. The price and fee denoms must be equal for each entry, and only one entry for any given denom is allowed. |
| `fee_buyer_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | fee_buyer_settlement_flat is the flat fee charged to the buyer during settlement. Each coin entry is a separate option. When a bid is created, the settlement fees provided must contain one of these. |
| `fee_buyer_settlement_ratios` | [FeeRatio](#provenance.exchange.v1.FeeRatio) | repeated | fee_buyer_settlement_ratios is the fee to charge a buyer during settlement based on the price they are spending. The price and fee denoms do not have to equal. Multiple entries for any given price or fee denom are allowed, but each price denom to fee denom pair can only have one entry. |
| `accepting_orders` | [bool](#bool) |  | accepting_orders is whether this market is allowing orders to be created for it. |
| `allow_user_settlement` | [bool](#bool) |  | allow_user_settlement is whether this market allows users to initiate their own settlements. For example, the FillBids and FillAsks endpoints are available if and only if this is true. The MarketSettle endpoint is only available to market actors regardless of the value of this field. |
| `access_grants` | [AccessGrant](#provenance.exchange.v1.AccessGrant) | repeated | access_grants is the list of addresses and permissions granted for this market. |
| `req_attr_create_ask` | [string](#string) | repeated | req_attr_create_ask is a list of attributes required on an account for it to be allowed to create an ask order. An account must have all of these attributes in order to create an ask order in this market. If the list is empty, any account can create ask orders in this market.

An entry that starts with "*." will match any attributes that end with the rest of it. E.g. "*.b.a" will match all of "c.b.a", "x.b.a", and "e.d.c.b.a"; but not "b.a", "xb.a", "b.x.a", or "c.b.a.x".

An entry of exactly "*" will match any attribute, which is equivalent to leaving this list empty. |
| `req_attr_create_bid` | [string](#string) | repeated | req_attr_create_ask is a list of attributes required on an account for it to be allowed to create a bid order. An account must have all of these attributes in order to create a bid order in this market. If the list is empty, any account can create bid orders in this market.

An entry that starts with "*." will match any attributes that end with the rest of it. E.g. "*.b.a" will match all of "c.b.a", "x.b.a", and "e.d.c.b.a"; but not "b.a", "xb.a", "c.b.x.a", or "c.b.a.x".

An entry of exactly "*" will match any attribute, which is equivalent to leaving this list empty. |






<a name="provenance.exchange.v1.MarketAccount"></a>

### MarketAccount
MarketAccount is an account type for use with the accounts module to hold some basic information about a market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_account` | [cosmos.auth.v1beta1.BaseAccount](#cosmos.auth.v1beta1.BaseAccount) |  | base_account is the base cosmos account information. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier for this market. |
| `market_details` | [MarketDetails](#provenance.exchange.v1.MarketDetails) |  | market_details is some human-consumable information about this market. |






<a name="provenance.exchange.v1.MarketDetails"></a>

### MarketDetails
MarketDetails contains information about a market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | name is a moniker that people can use to refer to this market. |
| `description` | [string](#string) |  | description extra information about this market. The field is meant to be human-readable. |
| `website_url` | [string](#string) |  | website_url is a url people can use to get to this market, or at least get more information about this market. |
| `icon_uri` | [string](#string) |  | icon_uri is a uri for an icon to associate with this market. |





 <!-- end messages -->


<a name="provenance.exchange.v1.Permission"></a>

### Permission
Permission defines the different types of permission that can be given to an account for a market.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PERMISSION_UNSPECIFIED | 0 | PERMISSION_UNSPECIFIED is the zero-value Permission; it is an error to use it. |
| PERMISSION_SETTLE | 1 | PERMISSION_SETTLE is the ability to use the Settle Tx endpoint on behalf of a market. |
| PERMISSION_CANCEL | 2 | PERMISSION_CANCEL is the ability to use the Cancel Tx endpoint on behalf of a market. |
| PERMISSION_WITHDRAW | 3 | PERMISSION_WITHDRAW is the ability to use the MarketWithdraw Tx endpoint. |
| PERMISSION_UPDATE | 4 | PERMISSION_UPDATE is the ability to use the MarketUpdate* Tx endpoints. |
| PERMISSION_PERMISSIONS | 5 | PERMISSION_PERMISSIONS is the ability to use the MarketManagePermissions Tx endpoint. |
| PERMISSION_ATTRIBUTES | 6 | PERMISSION_ATTRIBUTES is the ability to use the MarketManageReqAttrs Tx endpoint. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/exchange/v1/orders.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/orders.proto



<a name="provenance.exchange.v1.AskOrder"></a>

### AskOrder
AskOrder represents someone's desire to sell something at a minimum price.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id identifies the market that this order belongs to. |
| `seller` | [string](#string) |  | seller is the address of the account that owns this order and has the assets to sell. |
| `assets` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | assets are the things that the seller wishes to sell. A hold is placed on this until the order is filled or cancelled. |
| `price` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | price is the minimum amount that the seller is willing to accept for the assets. The seller's settlement proportional fee (and possibly the settlement flat fee) is taken out of the amount the seller receives, so it's possible that the seller will still receive less than this price. |
| `seller_settlement_flat_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | seller_settlement_flat_fee is the flat fee for sellers that will be charged during settlement. If this denom is the same denom as the price, it will come out of the actual price received. If this denom is different, the amount must be in the seller's account and a hold is placed on it until the order is filled or cancelled. |
| `allow_partial` | [bool](#bool) |  | allow_partial should be true if partial fulfillment of this order should be allowed, and should be false if the order must be either filled in full or not filled at all. |






<a name="provenance.exchange.v1.BidOrder"></a>

### BidOrder
BidOrder represents someone's desire to buy something at a specific price.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | market_id identifies the market that this order belongs to. |
| `buyer` | [string](#string) |  | buyer is the address of the account that owns this order and has the price to spend. |
| `assets` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | assets are the things that the buyer wishes to buy. |
| `price` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | price is the amount that the buyer will pay for the assets. A hold is placed on this until the order is filled or cancelled. |
| `buyer_settlement_fees` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | buyer_settlement_fees are the fees (both flat and proportional) that the buyer will pay (in addition to the price) when the order is settled. A hold is placed on this until the order is filled or cancelled. |
| `allow_partial` | [bool](#bool) |  | allow_partial should be true if partial fulfillment of this order should be allowed, and should be false if the order must be either filled in full or not filled at all. |






<a name="provenance.exchange.v1.Order"></a>

### Order
Order associates an order id with one of the order types.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the numerical identifier for this order. |
| `ask_order` | [AskOrder](#provenance.exchange.v1.AskOrder) |  | ask_order is the information about this order if it represents an ask order. |
| `bid_order` | [BidOrder](#provenance.exchange.v1.BidOrder) |  | bid_order is the information about this order if it represents a bid order. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/exchange/v1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/params.proto



<a name="provenance.exchange.v1.DenomSplit"></a>

### DenomSplit
DenomSplit associates a coin denomination with an amount the exchange receives for that denom.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | denom is the coin denomination this split applies to. |
| `split` | [uint32](#uint32) |  | split is the proportion of fees the exchange receives for this denom in basis points. E.g. 100 = 1%. Min = 0, Max = 10000. |






<a name="provenance.exchange.v1.Params"></a>

### Params
Params is a representation of the exchange module parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `default_split` | [uint32](#uint32) |  | default_split is the default proportion of fees the exchange receives in basis points. It is used if there isn't an applicable denom-specific split defined. E.g. 100 = 1%. Min = 0, Max = 10000. |
| `denom_splits` | [DenomSplit](#provenance.exchange.v1.DenomSplit) | repeated | denom_splits are the denom-specific amounts the exchange receives. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/exchange/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/genesis.proto



<a name="provenance.exchange.v1.GenesisState"></a>

### GenesisState
GenesisState is the data that should be loaded into the exchange module during genesis.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.exchange.v1.Params) |  | params defines all the parameters of the exchange module. |
| `markets` | [Market](#provenance.exchange.v1.Market) | repeated | markets are all of the markets to create at genesis. |
| `orders` | [Order](#provenance.exchange.v1.Order) | repeated | orders are all the orders to create at genesis. |
| `last_market_id` | [uint32](#uint32) |  | last_market_id is the value of the last auto-selected market id. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/exchange/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/query.proto



<a name="provenance.exchange.v1.QueryGetAddressOrdersRequest"></a>

### QueryGetAddressOrdersRequest
QueryGetAddressOrdersRequest is a request message for the QueryGetAddressOrders endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | TODO[1658]: QueryGetAddressOrdersRequest |






<a name="provenance.exchange.v1.QueryGetAddressOrdersResponse"></a>

### QueryGetAddressOrdersResponse
QueryGetAddressOrdersResponse is a response message for the QueryGetAddressOrders endpoint.






<a name="provenance.exchange.v1.QueryGetAllOrdersRequest"></a>

### QueryGetAllOrdersRequest
QueryGetAllOrdersRequest is a request message for the QueryGetAllOrders endpoint.






<a name="provenance.exchange.v1.QueryGetAllOrdersResponse"></a>

### QueryGetAllOrdersResponse
QueryGetAllOrdersResponse is a response message for the QueryGetAllOrders endpoint.






<a name="provenance.exchange.v1.QueryGetMarketOrdersRequest"></a>

### QueryGetMarketOrdersRequest
QueryGetMarketOrdersRequest is a request message for the QueryGetMarketOrders endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | TODO[1658]: QueryGetMarketOrdersRequest |






<a name="provenance.exchange.v1.QueryGetMarketOrdersResponse"></a>

### QueryGetMarketOrdersResponse
QueryGetMarketOrdersResponse is a response message for the QueryGetMarketOrders endpoint.






<a name="provenance.exchange.v1.QueryGetOrderRequest"></a>

### QueryGetOrderRequest
QueryGetOrderRequest is a request message for the QueryGetOrder endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | TODO[1658]: QueryGetOrderRequest |






<a name="provenance.exchange.v1.QueryGetOrderResponse"></a>

### QueryGetOrderResponse
QueryGetOrderResponse is a response message for the QueryGetOrder endpoint.






<a name="provenance.exchange.v1.QueryMarketInfoRequest"></a>

### QueryMarketInfoRequest
QueryMarketInfoRequest is a request message for the QueryMarketInfo endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | TODO[1658]: QueryMarketInfoRequest |






<a name="provenance.exchange.v1.QueryMarketInfoResponse"></a>

### QueryMarketInfoResponse
QueryMarketInfoResponse is a response message for the QueryMarketInfo endpoint.






<a name="provenance.exchange.v1.QueryOrderFeeCalcRequest"></a>

### QueryOrderFeeCalcRequest
QueryOrderFeeCalcRequest is a request message for the QueryOrderFeeCalc endpoint.






<a name="provenance.exchange.v1.QueryOrderFeeCalcResponse"></a>

### QueryOrderFeeCalcResponse
QueryOrderFeeCalcResponse is a response message for the QueryOrderFeeCalc endpoint.






<a name="provenance.exchange.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is a request message for the QueryParams endpoint.






<a name="provenance.exchange.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is a response message for the QueryParams endpoint.






<a name="provenance.exchange.v1.QuerySettlementFeeCalcRequest"></a>

### QuerySettlementFeeCalcRequest
QuerySettlementFeeCalcRequest is a request message for the QuerySettlementFeeCalc endpoint.






<a name="provenance.exchange.v1.QuerySettlementFeeCalcResponse"></a>

### QuerySettlementFeeCalcResponse
QuerySettlementFeeCalcResponse is a response message for the QuerySettlementFeeCalc endpoint.






<a name="provenance.exchange.v1.QueryValidateCreateMarketRequest"></a>

### QueryValidateCreateMarketRequest
QueryValidateCreateMarketRequest is a request message for the QueryValidateCreateMarket endpoint.






<a name="provenance.exchange.v1.QueryValidateCreateMarketResponse"></a>

### QueryValidateCreateMarketResponse
QueryValidateCreateMarketResponse is a response message for the QueryValidateCreateMarket endpoint.






<a name="provenance.exchange.v1.QueryValidateManageFeesRequest"></a>

### QueryValidateManageFeesRequest
QueryValidateManageFeesRequest is a request message for the QueryValidateManageFees endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `market_id` | [uint32](#uint32) |  | TODO[1658]: QueryValidateManageFeesRequest |






<a name="provenance.exchange.v1.QueryValidateManageFeesResponse"></a>

### QueryValidateManageFeesResponse
QueryValidateManageFeesResponse is a response message for the QueryValidateManageFees endpoint.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.exchange.v1.Query"></a>

### Query
Query is the service for exchange module's query endpoints.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `QueryOrderFeeCalc` | [QueryOrderFeeCalcRequest](#provenance.exchange.v1.QueryOrderFeeCalcRequest) | [QueryOrderFeeCalcResponse](#provenance.exchange.v1.QueryOrderFeeCalcResponse) | QueryOrderFeeCalc calculates the fees that will be associated with the provided order. | GET|/provenance/exchange/v1/fees/order|
| `QuerySettlementFeeCalc` | [QuerySettlementFeeCalcRequest](#provenance.exchange.v1.QuerySettlementFeeCalcRequest) | [QuerySettlementFeeCalcResponse](#provenance.exchange.v1.QuerySettlementFeeCalcResponse) | QuerySettlementFeeCalc calculates the fees that will be associated with the provided settlement. | GET|/provenance/exchange/v1/fees/settlement|
| `QueryGetOrder` | [QueryGetOrderRequest](#provenance.exchange.v1.QueryGetOrderRequest) | [QueryGetOrderResponse](#provenance.exchange.v1.QueryGetOrderResponse) | QueryGetOrder looks up an order by id. | GET|/provenance/exchange/v1/order/{order_id}|
| `QueryGetMarketOrders` | [QueryGetMarketOrdersRequest](#provenance.exchange.v1.QueryGetMarketOrdersRequest) | [QueryGetMarketOrdersResponse](#provenance.exchange.v1.QueryGetMarketOrdersResponse) | QueryGetMarketOrders looks up the orders in a market. | GET|/provenance/exchange/v1/market/{market_id}/orders|
| `QueryGetAddressOrders` | [QueryGetAddressOrdersRequest](#provenance.exchange.v1.QueryGetAddressOrdersRequest) | [QueryGetAddressOrdersResponse](#provenance.exchange.v1.QueryGetAddressOrdersResponse) | QueryGetAddressOrders looks up the orders from the provided address. | GET|/provenance/exchange/v1/orders/{address}|
| `QueryGetAllOrders` | [QueryGetAllOrdersRequest](#provenance.exchange.v1.QueryGetAllOrdersRequest) | [QueryGetAllOrdersResponse](#provenance.exchange.v1.QueryGetAllOrdersResponse) | QueryGetAllOrders gets all orders in the exchange module. | GET|/provenance/exchange/v1/orders|
| `QueryMarketInfo` | [QueryMarketInfoRequest](#provenance.exchange.v1.QueryMarketInfoRequest) | [QueryMarketInfoResponse](#provenance.exchange.v1.QueryMarketInfoResponse) | QueryMarketInfo returns the information/details about a market. | GET|/provenance/exchange/v1/market/{market_id}|
| `QueryParams` | [QueryParamsRequest](#provenance.exchange.v1.QueryParamsRequest) | [QueryParamsResponse](#provenance.exchange.v1.QueryParamsResponse) | QueryParams returns the exchange module parameters. | GET|/provenance/exchange/v1/params|
| `QueryValidateCreateMarket` | [QueryValidateCreateMarketRequest](#provenance.exchange.v1.QueryValidateCreateMarketRequest) | [QueryValidateCreateMarketResponse](#provenance.exchange.v1.QueryValidateCreateMarketResponse) | QueryValidateCreateMarket checks the provided MsgGovCreateMarketResponse and returns any errors it might have. | GET|/provenance/exchange/v1/validate/create_market|
| `QueryValidateManageFees` | [QueryValidateManageFeesRequest](#provenance.exchange.v1.QueryValidateManageFeesRequest) | [QueryValidateManageFeesResponse](#provenance.exchange.v1.QueryValidateManageFeesResponse) | QueryValidateManageFees checks the provided MsgGovManageFeesRequest and returns any errors that it might have. | GET|/provenance/exchange/v1/market/{market_id}/validate/manage_fees|

 <!-- end services -->



<a name="provenance/exchange/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/exchange/v1/tx.proto



<a name="provenance.exchange.v1.MsgCancelOrderRequest"></a>

### MsgCancelOrderRequest
MsgCancelOrderRequest is a request message for the CancelOrder endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signer` | [string](#string) |  | signer is the account requesting the order cancelation. It must be either the order owner (e.g. the buyer or seller), the governance module account address, or an account with cancel permission with the market that the order is in. |
| `order_id` | [uint64](#uint64) |  | order_id is the id of the order to cancel. |






<a name="provenance.exchange.v1.MsgCancelOrderResponse"></a>

### MsgCancelOrderResponse
MsgCancelOrderResponse is a response message for the CancelOrder endpoint.






<a name="provenance.exchange.v1.MsgCreateAskRequest"></a>

### MsgCreateAskRequest
MsgCreateAskRequest is a request message for the CreateAsk endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ask_order` | [AskOrder](#provenance.exchange.v1.AskOrder) |  | ask_order is the details of the order being created. |
| `order_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | order_creation_fee is the fee that is being paid to create this order. |






<a name="provenance.exchange.v1.MsgCreateAskResponse"></a>

### MsgCreateAskResponse
MsgCreateAskResponse is a response message for the CreateAsk endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the id of the order created. |






<a name="provenance.exchange.v1.MsgCreateBidRequest"></a>

### MsgCreateBidRequest
MsgCreateBidRequest is a request message for the CreateBid endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `bid_order` | [BidOrder](#provenance.exchange.v1.BidOrder) |  | bid_order is the details of the order being created. |
| `order_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | order_creation_fee is the fee that is being paid to create this order. |






<a name="provenance.exchange.v1.MsgCreateBidResponse"></a>

### MsgCreateBidResponse
MsgCreateBidResponse is a response message for the CreateBid endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `order_id` | [uint64](#uint64) |  | order_id is the id of the order created. |






<a name="provenance.exchange.v1.MsgFillAsksRequest"></a>

### MsgFillAsksRequest
MsgFillAsksRequest is a request message for the FillAsks endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `buyer` | [string](#string) |  | buyer is the address of the account attempting to buy some assets. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market with the asks to fill. All ask orders being filled must be in this market. |
| `total_price` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | total_price is the total amount being spent on some assets. It must be the sum of all ask order prices. |
| `ask_order_ids` | [uint64](#uint64) | repeated | ask_order_ids are the ids of the ask orders that you are trying to fill. All ids must be for ask orders, and must be in the same market as the market_id. |
| `buyer_settlement_fees` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | buyer_settlement_fees are the fees (both flat and proportional) that the buyer will pay (in addition to the price) for this settlement. |
| `bid_order_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | bid_order_creation_fee is the fee that is being paid to create this order (which is immediately then settled). |






<a name="provenance.exchange.v1.MsgFillAsksResponse"></a>

### MsgFillAsksResponse
MsgFillAsksResponse is a response message for the FillAsks endpoint.






<a name="provenance.exchange.v1.MsgFillBidsRequest"></a>

### MsgFillBidsRequest
MsgFillBidsRequest is a request message for the FillBids endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `seller` | [string](#string) |  | seller is the address of the account with the assets to sell. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market with the bids to fill. All bid orders being filled must be in this market. |
| `total_assets` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | total_assets are the things that the seller wishes to sell. It must be the sum of all bid order assets. |
| `bid_order_ids` | [uint64](#uint64) | repeated | bid_order_ids are the ids of the bid orders that you are trying to fill. All ids must be for bid orders, and must be in the same market as the market_id. |
| `seller_settlement_flat_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | seller_settlement_flat_fee is the flat fee for sellers that will be charged for this settlement. |
| `ask_order_creation_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | ask_order_creation_fee is the fee that is being paid to create this order (which is immediately then settled). |






<a name="provenance.exchange.v1.MsgFillBidsResponse"></a>

### MsgFillBidsResponse
MsgFillBidsResponse is a response message for the FillBids endpoint.






<a name="provenance.exchange.v1.MsgGovCreateMarketRequest"></a>

### MsgGovCreateMarketRequest
MsgGovCreateMarketRequest is a request message for the GovCreateMarket endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `market` | [Market](#provenance.exchange.v1.Market) |  | market is the initial market configuration. If the market_id is 0, the next available market_id will be used (once voting ends). If it is not zero, it must not yet be in use when the voting period ends. |






<a name="provenance.exchange.v1.MsgGovCreateMarketResponse"></a>

### MsgGovCreateMarketResponse
MsgGovCreateMarketResponse is a response message for the GovCreateMarket endpoint.






<a name="provenance.exchange.v1.MsgGovManageFeesRequest"></a>

### MsgGovManageFeesRequest
MsgGovManageFeesRequest is a request message for the GovManageFees endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `market_id` | [uint32](#uint32) |  | market_id is the market id that will get these fee updates. |
| `add_fee_create_ask_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | add_fee_create_ask_flat are the create-ask flat fee options to add. |
| `remove_fee_create_ask_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | remove_fee_create_ask_flat are the create-ask flat fee options to remove. |
| `add_fee_create_bid_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | add_fee_create_bid_flat are the create-bid flat fee options to add. |
| `remove_fee_create_bid_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | remove_fee_create_bid_flat are the create-bid flat fee options to remove. |
| `add_fee_seller_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | add_fee_seller_settlement_flat are the seller settlement flat fee options to add. |
| `remove_fee_seller_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | remove_fee_seller_settlement_flat are the seller settlement flat fee options to remove. |
| `add_fee_seller_settlement_ratios` | [FeeRatio](#provenance.exchange.v1.FeeRatio) | repeated | add_fee_seller_settlement_ratios are the seller settlement fee ratios to add. |
| `remove_fee_seller_settlement_ratios` | [FeeRatio](#provenance.exchange.v1.FeeRatio) | repeated | remove_fee_seller_settlement_ratios are the seller settlement fee ratios to remove. |
| `add_fee_buyer_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | add_fee_buyer_settlement_flat are the buyer settlement flat fee options to add. |
| `remove_fee_buyer_settlement_flat` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | remove_fee_buyer_settlement_flat are the buyer settlement flat fee options to remove. |
| `add_fee_buyer_settlement_ratios` | [FeeRatio](#provenance.exchange.v1.FeeRatio) | repeated | add_fee_buyer_settlement_ratios are the buyer settlement fee ratios to add. |
| `remove_fee_buyer_settlement_ratios` | [FeeRatio](#provenance.exchange.v1.FeeRatio) | repeated | remove_fee_buyer_settlement_ratios are the buyer settlement fee ratios to remove. |






<a name="provenance.exchange.v1.MsgGovManageFeesResponse"></a>

### MsgGovManageFeesResponse
MsgGovManageFeesResponse is a response message for the GovManageFees endpoint.






<a name="provenance.exchange.v1.MsgGovUpdateParamsRequest"></a>

### MsgGovUpdateParamsRequest
MsgGovUpdateParamsRequest is a request message for the GovUpdateParams endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority should be the governance module account address. |
| `params` | [Params](#provenance.exchange.v1.Params) |  | params are the new param values to set |






<a name="provenance.exchange.v1.MsgGovUpdateParamsResponse"></a>

### MsgGovUpdateParamsResponse
MsgGovUpdateParamsResponse is a response message for the GovUpdateParams endpoint.






<a name="provenance.exchange.v1.MsgMarketManagePermissionsRequest"></a>

### MsgMarketManagePermissionsRequest
MsgMarketManagePermissionsRequest is a request message for the MarketManagePermissions endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "permissions" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to update required attributes for. |
| `revoke_all` | [string](#string) | repeated | revoke_all are addresses that should have all their permissions revoked. |
| `to_revoke` | [AccessGrant](#provenance.exchange.v1.AccessGrant) | repeated | to_revoke are the specific permissions to remove for addresses. |
| `to_grant` | [AccessGrant](#provenance.exchange.v1.AccessGrant) | repeated | to_grant are the permissions to grant to addresses. |






<a name="provenance.exchange.v1.MsgMarketManagePermissionsResponse"></a>

### MsgMarketManagePermissionsResponse
MsgMarketManagePermissionsResponse is a response message for the MarketManagePermissions endpoint.






<a name="provenance.exchange.v1.MsgMarketManageReqAttrsRequest"></a>

### MsgMarketManageReqAttrsRequest
MsgMarketManageReqAttrsRequest is a request message for the MarketManageReqAttrs endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "attributes" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to update required attributes for. |
| `create_ask_to_add` | [string](#string) | repeated | create_ask_to_add are the attributes that should now also be required to create an ask order. |
| `create_ask_to_remove` | [string](#string) | repeated | create_ask_to_add are the attributes that should no longer be required to create an ask order. |
| `create_bid_to_add` | [string](#string) | repeated | create_ask_to_add are the attributes that should now also be required to create a bid order. |
| `create_bid_to_remove` | [string](#string) | repeated | create_ask_to_add are the attributes that should no longer be required to create a bid order. |






<a name="provenance.exchange.v1.MsgMarketManageReqAttrsResponse"></a>

### MsgMarketManageReqAttrsResponse
MsgMarketManageReqAttrsResponse is a response message for the MarketManageReqAttrs endpoint.






<a name="provenance.exchange.v1.MsgMarketSettleRequest"></a>

### MsgMarketSettleRequest
MsgMarketSettleRequest is a request message for the MarketSettle endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "settle" permission requesting this settlement. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to update required attributes for.

TODO[1658]: MsgMarketSettleRequest |






<a name="provenance.exchange.v1.MsgMarketSettleResponse"></a>

### MsgMarketSettleResponse
MsgMarketSettleResponse is a response message for the MarketSettle endpoint.






<a name="provenance.exchange.v1.MsgMarketUpdateDetailsRequest"></a>

### MsgMarketUpdateDetailsRequest
MsgMarketUpdateDetailsRequest is a request message for the MarketUpdateDetails endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "update" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to update required attributes for. |
| `market_details` | [MarketDetails](#provenance.exchange.v1.MarketDetails) |  | market_details is some information about this market. |






<a name="provenance.exchange.v1.MsgMarketUpdateDetailsResponse"></a>

### MsgMarketUpdateDetailsResponse
MsgMarketUpdateDetailsResponse is a response message for the MarketUpdateDetails endpoint.






<a name="provenance.exchange.v1.MsgMarketUpdateEnabledRequest"></a>

### MsgMarketUpdateEnabledRequest
MsgMarketUpdateEnabledRequest is a request message for the MarketUpdateEnabled endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "update" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to update required attributes for. |
| `accepting_orders` | [bool](#bool) |  | accepting_orders is whether this market is allowing orders to be created for it. |






<a name="provenance.exchange.v1.MsgMarketUpdateEnabledResponse"></a>

### MsgMarketUpdateEnabledResponse
MsgMarketUpdateEnabledResponse is a response message for the MarketUpdateEnabled endpoint.






<a name="provenance.exchange.v1.MsgMarketUpdateUserSettleRequest"></a>

### MsgMarketUpdateUserSettleRequest
MsgMarketUpdateUserSettleRequest is a request message for the MarketUpdateUserSettle endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with "update" permission requesting this change. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to update required attributes for. |
| `allow_user_settlement` | [bool](#bool) |  | allow_user_settlement is whether this market allows users to initiate their own settlements. For example, the FillBids and FillAsks endpoints are available if and only if this is true. The MarketSettle endpoint is only available to market actors regardless of the value of this field. |






<a name="provenance.exchange.v1.MsgMarketUpdateUserSettleResponse"></a>

### MsgMarketUpdateUserSettleResponse
MsgMarketUpdateUserSettleResponse is a response message for the MarketUpdateUserSettle endpoint.






<a name="provenance.exchange.v1.MsgMarketWithdrawRequest"></a>

### MsgMarketWithdrawRequest
MsgMarketWithdrawRequest is a request message for the MarketWithdraw endpoint.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `admin` | [string](#string) |  | admin is the account with withdraw permission requesting the withdrawal. |
| `market_id` | [uint32](#uint32) |  | market_id is the numerical identifier of the market to withdraw from. |
| `to_address` | [string](#string) |  | to_address is the address that will receive the funds. |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | amount is the funds to withdraw. |






<a name="provenance.exchange.v1.MsgMarketWithdrawResponse"></a>

### MsgMarketWithdrawResponse
MsgMarketWithdrawResponse is a response message for the MarketWithdraw endpoint.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.exchange.v1.Msg"></a>

### Msg
Msg is the service for exchange module's tx endpoints.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateAsk` | [MsgCreateAskRequest](#provenance.exchange.v1.MsgCreateAskRequest) | [MsgCreateAskResponse](#provenance.exchange.v1.MsgCreateAskResponse) | CreateAsk creates an ask order (to sell something you own). | |
| `CreateBid` | [MsgCreateBidRequest](#provenance.exchange.v1.MsgCreateBidRequest) | [MsgCreateBidResponse](#provenance.exchange.v1.MsgCreateBidResponse) | CreateBid creates a bid order (to buy something you want). | |
| `CancelOrder` | [MsgCancelOrderRequest](#provenance.exchange.v1.MsgCancelOrderRequest) | [MsgCancelOrderResponse](#provenance.exchange.v1.MsgCancelOrderResponse) | CancelOrder cancels an order. | |
| `FillBids` | [MsgFillBidsRequest](#provenance.exchange.v1.MsgFillBidsRequest) | [MsgFillBidsResponse](#provenance.exchange.v1.MsgFillBidsResponse) | FillBids uses the assets in your account to fulfill one or more bids (similar to a fill-or-cancel ask). | |
| `FillAsks` | [MsgFillAsksRequest](#provenance.exchange.v1.MsgFillAsksRequest) | [MsgFillAsksResponse](#provenance.exchange.v1.MsgFillAsksResponse) | FillAsks uses the funds in your account to fulfill one or more asks (similar to a fill-or-cancel bid). | |
| `MarketSettle` | [MsgMarketSettleRequest](#provenance.exchange.v1.MsgMarketSettleRequest) | [MsgMarketSettleResponse](#provenance.exchange.v1.MsgMarketSettleResponse) | MarketSettle is a market endpoint to trigger the settlement of orders. | |
| `MarketWithdraw` | [MsgMarketWithdrawRequest](#provenance.exchange.v1.MsgMarketWithdrawRequest) | [MsgMarketWithdrawResponse](#provenance.exchange.v1.MsgMarketWithdrawResponse) | MarketWithdraw is a market endpoint to withdraw fees that have been collected. | |
| `MarketUpdateDetails` | [MsgMarketUpdateDetailsRequest](#provenance.exchange.v1.MsgMarketUpdateDetailsRequest) | [MsgMarketUpdateDetailsResponse](#provenance.exchange.v1.MsgMarketUpdateDetailsResponse) | MarketUpdateDetails is a market endpoint to update its details. | |
| `MarketUpdateEnabled` | [MsgMarketUpdateEnabledRequest](#provenance.exchange.v1.MsgMarketUpdateEnabledRequest) | [MsgMarketUpdateEnabledResponse](#provenance.exchange.v1.MsgMarketUpdateEnabledResponse) | MarketUpdateEnabled is a market endpoint to update whether its accepting orders. | |
| `MarketUpdateUserSettle` | [MsgMarketUpdateUserSettleRequest](#provenance.exchange.v1.MsgMarketUpdateUserSettleRequest) | [MsgMarketUpdateUserSettleResponse](#provenance.exchange.v1.MsgMarketUpdateUserSettleResponse) | MarketUpdateUserSettle is a market endpoint to update whether it allows user-initiated settlement. | |
| `MarketManagePermissions` | [MsgMarketManagePermissionsRequest](#provenance.exchange.v1.MsgMarketManagePermissionsRequest) | [MsgMarketManagePermissionsResponse](#provenance.exchange.v1.MsgMarketManagePermissionsResponse) | MarketManagePermissions is a market endpoint to manage a market's user permissions. | |
| `MarketManageReqAttrs` | [MsgMarketManageReqAttrsRequest](#provenance.exchange.v1.MsgMarketManageReqAttrsRequest) | [MsgMarketManageReqAttrsResponse](#provenance.exchange.v1.MsgMarketManageReqAttrsResponse) | MarketManageReqAttrs is a market endpoint to manage the attributes required to interact with it. | |
| `GovCreateMarket` | [MsgGovCreateMarketRequest](#provenance.exchange.v1.MsgGovCreateMarketRequest) | [MsgGovCreateMarketResponse](#provenance.exchange.v1.MsgGovCreateMarketResponse) | GovCreateMarket is a governance proposal endpoint for creating a market. | |
| `GovManageFees` | [MsgGovManageFeesRequest](#provenance.exchange.v1.MsgGovManageFeesRequest) | [MsgGovManageFeesResponse](#provenance.exchange.v1.MsgGovManageFeesResponse) | GovManageFees is a governance proposal endpoint for updating a market's fees. | |
| `GovUpdateParams` | [MsgGovUpdateParamsRequest](#provenance.exchange.v1.MsgGovUpdateParamsRequest) | [MsgGovUpdateParamsResponse](#provenance.exchange.v1.MsgGovUpdateParamsResponse) | GovUpdateParams is a governance proposal endpoint for updating the exchange module's params. | |

 <!-- end services -->



<a name="provenance/hold/v1/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/hold/v1/events.proto



<a name="provenance.hold.v1.EventHoldAdded"></a>

### EventHoldAdded
EventHoldAdded is an event indicating that some funds were placed on hold in an account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the bech32 address string of the account with the funds. |
| `amount` | [string](#string) |  | amount is a Coins string of the funds placed on hold. |
| `reason` | [string](#string) |  | reason is a human-readable indicator of why this hold was added. |






<a name="provenance.hold.v1.EventHoldReleased"></a>

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



<a name="provenance/hold/v1/hold.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/hold/v1/hold.proto



<a name="provenance.hold.v1.AccountHold"></a>

### AccountHold
AccountHold associates an address with an amount on hold for that address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the account address that holds the funds on hold. |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | amount is the balances that are on hold for the address. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/hold/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/hold/v1/genesis.proto



<a name="provenance.hold.v1.GenesisState"></a>

### GenesisState
GenesisState defines the attribute module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `holds` | [AccountHold](#provenance.hold.v1.AccountHold) | repeated | holds defines the funds on hold at genesis. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/hold/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/hold/v1/query.proto



<a name="provenance.hold.v1.GetAllHoldsRequest"></a>

### GetAllHoldsRequest
GetAllHoldsRequest is the request type for the Query/GetAllHolds query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.hold.v1.GetAllHoldsResponse"></a>

### GetAllHoldsResponse
GetAllHoldsResponse is the response type for the Query/GetAllHolds query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `holds` | [AccountHold](#provenance.hold.v1.AccountHold) | repeated | holds is a list of addresses with funds on hold and the amounts being held. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance.hold.v1.GetHoldsRequest"></a>

### GetHoldsRequest
GetHoldsRequest is the request type for the Query/GetHolds query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the account address to get on-hold balances for. |






<a name="provenance.hold.v1.GetHoldsResponse"></a>

### GetHoldsResponse
GetHoldsResponse is the response type for the Query/GetHolds query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | amount is the total on hold for the requested address. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.hold.v1.Query"></a>

### Query
Query defines the gRPC querier service for attribute module.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `GetHolds` | [GetHoldsRequest](#provenance.hold.v1.GetHoldsRequest) | [GetHoldsResponse](#provenance.hold.v1.GetHoldsResponse) | GetHolds looks up the funds that are on hold for an address. | GET|/provenance/hold/v1/funds/{address}|
| `GetAllHolds` | [GetAllHoldsRequest](#provenance.hold.v1.GetAllHoldsRequest) | [GetAllHoldsResponse](#provenance.hold.v1.GetAllHoldsResponse) | GetAllHolds returns all addresses with funds on hold, and the amount held. | GET|/provenance/hold/v1/funds|

 <!-- end services -->



<a name="provenance/ibchooks/v1/params.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibchooks/v1/params.proto



<a name="provenance.ibchooks.v1.Params"></a>

### Params
Params defines the allowed async ack contracts


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `allowed_async_ack_contracts` | [string](#string) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/ibchooks/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibchooks/v1/genesis.proto



<a name="provenance.ibchooks.v1.GenesisState"></a>

### GenesisState
GenesisState is the IBC Hooks genesis state (params)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.ibchooks.v1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/ibchooks/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/ibchooks/v1/tx.proto



<a name="provenance.ibchooks.v1.MsgEmitIBCAck"></a>

### MsgEmitIBCAck
MsgEmitIBCAck is the IBC Acknowledgement


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |
| `packet_sequence` | [uint64](#uint64) |  |  |
| `channel` | [string](#string) |  |  |






<a name="provenance.ibchooks.v1.MsgEmitIBCAckResponse"></a>

### MsgEmitIBCAckResponse
MsgEmitIBCAckResponse is the IBC Acknowledgement response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_result` | [string](#string) |  |  |
| `ibc_ack` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.ibchooks.v1.Msg"></a>

### Msg
Msg defines the Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `EmitIBCAck` | [MsgEmitIBCAck](#provenance.ibchooks.v1.MsgEmitIBCAck) | [MsgEmitIBCAckResponse](#provenance.ibchooks.v1.MsgEmitIBCAckResponse) | EmitIBCAck checks the sender can emit the ack and writes the IBC acknowledgement | |

 <!-- end services -->



<a name="provenance/marker/v1/accessgrant.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/accessgrant.proto



<a name="provenance.marker.v1.AccessGrant"></a>

### AccessGrant
AccessGrant associates a collection of permissions with an address for delegated marker account control.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `permissions` | [Access](#provenance.marker.v1.Access) | repeated |  |





 <!-- end messages -->


<a name="provenance.marker.v1.Access"></a>

### Access
Access defines the different types of permissions that a marker supports granting to an address.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ACCESS_UNSPECIFIED | 0 | ACCESS_UNSPECIFIED defines a no-op vote option. |
| ACCESS_MINT | 1 | ACCESS_MINT is the ability to increase the supply of a marker |
| ACCESS_BURN | 2 | ACCESS_BURN is the ability to decrease the supply of the marker using coin held by the marker. |
| ACCESS_DEPOSIT | 3 | ACCESS_DEPOSIT is the ability to set a marker reference to this marker in the metadata/scopes module |
| ACCESS_WITHDRAW | 4 | ACCESS_WITHDRAW is the ability to remove marker references to this marker in from metadata/scopes or transfer coin from this marker account to another account. |
| ACCESS_DELETE | 5 | ACCESS_DELETE is the ability to move a proposed, finalized or active marker into the cancelled state. This access also allows cancelled markers to be marked for deletion |
| ACCESS_ADMIN | 6 | ACCESS_ADMIN is the ability to add access grants for accounts to the list of marker permissions. |
| ACCESS_TRANSFER | 7 | ACCESS_TRANSFER is the ability to invoke a send operation using the marker module to facilitate exchange. This access right is only supported on RESTRICTED markers. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/marker/v1/authz.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/authz.proto



<a name="provenance.marker.v1.MarkerTransferAuthorization"></a>

### MarkerTransferAuthorization
MarkerTransferAuthorization gives the grantee permissions to execute
a marker transfer on behalf of the granter's account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfer_limit` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | transfer_limit is the total amount the grantee can transfer |
| `allow_list` | [string](#string) | repeated | allow_list specifies an optional list of addresses to whom the grantee can send restricted coins on behalf of the granter. If omitted, any recipient is allowed. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/marker/v1/marker.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/marker.proto



<a name="provenance.marker.v1.EventDenomUnit"></a>

### EventDenomUnit
EventDenomUnit denom units for set denom metadata event


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `exponent` | [string](#string) |  |  |
| `aliases` | [string](#string) | repeated |  |






<a name="provenance.marker.v1.EventMarkerAccess"></a>

### EventMarkerAccess
EventMarkerAccess event access permissions for address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `permissions` | [string](#string) | repeated |  |






<a name="provenance.marker.v1.EventMarkerActivate"></a>

### EventMarkerActivate
EventMarkerActivate event emitted when marker is activated


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerAdd"></a>

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






<a name="provenance.marker.v1.EventMarkerAddAccess"></a>

### EventMarkerAddAccess
EventMarkerAddAccess event emitted when marker access is added


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `access` | [EventMarkerAccess](#provenance.marker.v1.EventMarkerAccess) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerBurn"></a>

### EventMarkerBurn
EventMarkerBurn event emitted when coin is burned from marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerCancel"></a>

### EventMarkerCancel
EventMarkerCancel event emitted when marker is cancelled


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerDelete"></a>

### EventMarkerDelete
EventMarkerDelete event emitted when marker is deleted


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerDeleteAccess"></a>

### EventMarkerDeleteAccess
EventMarkerDeleteAccess event emitted when marker access is revoked


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `remove_address` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerFinalize"></a>

### EventMarkerFinalize
EventMarkerFinalize event emitted when marker is finalized


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerMint"></a>

### EventMarkerMint
EventMarkerMint event emitted when additional marker supply is minted


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerSetDenomMetadata"></a>

### EventMarkerSetDenomMetadata
EventMarkerSetDenomMetadata event emitted when metadata is set on marker with denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata_base` | [string](#string) |  |  |
| `metadata_description` | [string](#string) |  |  |
| `metadata_display` | [string](#string) |  |  |
| `metadata_denom_units` | [EventDenomUnit](#provenance.marker.v1.EventDenomUnit) | repeated |  |
| `administrator` | [string](#string) |  |  |
| `metadata_name` | [string](#string) |  |  |
| `metadata_symbol` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerTransfer"></a>

### EventMarkerTransfer
EventMarkerTransfer event emitted when coins are transfered to from account to another


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `to_address` | [string](#string) |  |  |
| `from_address` | [string](#string) |  |  |






<a name="provenance.marker.v1.EventMarkerWithdraw"></a>

### EventMarkerWithdraw
EventMarkerWithdraw event emitted when coins are withdrew from marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `to_address` | [string](#string) |  |  |






<a name="provenance.marker.v1.MarkerAccount"></a>

### MarkerAccount
MarkerAccount holds the marker configuration information in addition to a base account structure.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_account` | [cosmos.auth.v1beta1.BaseAccount](#cosmos.auth.v1beta1.BaseAccount) |  | base cosmos account information including address and coin holdings. |
| `manager` | [string](#string) |  | Address that owns the marker configuration. This account must sign any requests to change marker config (only valid for statuses prior to finalization) |
| `access_control` | [AccessGrant](#provenance.marker.v1.AccessGrant) | repeated | Access control lists |
| `status` | [MarkerStatus](#provenance.marker.v1.MarkerStatus) |  | Indicates the current status of this marker record. |
| `denom` | [string](#string) |  | value denomination and total supply for the token. |
| `supply` | [string](#string) |  | the total supply expected for a marker. This is the amount that is minted when a marker is created. |
| `marker_type` | [MarkerType](#provenance.marker.v1.MarkerType) |  | Marker type information |
| `supply_fixed` | [bool](#bool) |  | A fixed supply will mint additional coin automatically if the total supply decreases below a set value. This may occur if the coin is burned or an account holding the coin is slashed. (default: true) |
| `allow_governance_control` | [bool](#bool) |  | indicates that governance based control is allowed for this marker |
| `allow_forced_transfer` | [bool](#bool) |  | Whether an admin can transfer restricted coins from a 3rd-party account without their signature. |
| `required_attributes` | [string](#string) | repeated | list of required attributes on restricted marker in order to send and receive transfers if sender does not have transfer authority |






<a name="provenance.marker.v1.Params"></a>

### Params
Params defines the set of params for the account module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_total_supply` | [uint64](#uint64) |  | maximum amount of supply to allow a marker to be created with |
| `enable_governance` | [bool](#bool) |  | indicates if governance based controls of markers is allowed. |
| `unrestricted_denom_regex` | [string](#string) |  | a regular expression used to validate marker denom values from normal create requests (governance requests are only subject to platform coin validation denom expression) |





 <!-- end messages -->


<a name="provenance.marker.v1.MarkerStatus"></a>

### MarkerStatus
MarkerStatus defines the various states a marker account can be in.

| Name | Number | Description |
| ---- | ------ | ----------- |
| MARKER_STATUS_UNSPECIFIED | 0 | MARKER_STATUS_UNSPECIFIED - Unknown/Invalid Marker Status |
| MARKER_STATUS_PROPOSED | 1 | MARKER_STATUS_PROPOSED - Initial configuration period, updates allowed, token supply not created. |
| MARKER_STATUS_FINALIZED | 2 | MARKER_STATUS_FINALIZED - Configuration finalized, ready for supply creation |
| MARKER_STATUS_ACTIVE | 3 | MARKER_STATUS_ACTIVE - Supply is created, rules are in force. |
| MARKER_STATUS_CANCELLED | 4 | MARKER_STATUS_CANCELLED - Marker has been cancelled, pending destroy |
| MARKER_STATUS_DESTROYED | 5 | MARKER_STATUS_DESTROYED - Marker supply has all been recalled, marker is considered destroyed and no further actions allowed. |



<a name="provenance.marker.v1.MarkerType"></a>

### MarkerType
MarkerType defines the types of marker

| Name | Number | Description |
| ---- | ------ | ----------- |
| MARKER_TYPE_UNSPECIFIED | 0 | MARKER_TYPE_UNSPECIFIED is an invalid/unknown marker type. |
| MARKER_TYPE_COIN | 1 | MARKER_TYPE_COIN is a marker that represents a standard fungible coin (default). |
| MARKER_TYPE_RESTRICTED | 2 | MARKER_TYPE_RESTRICTED is a marker that represents a denom with send_enabled = false. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/marker/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/genesis.proto



<a name="provenance.marker.v1.GenesisState"></a>

### GenesisState
GenesisState defines the account module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.marker.v1.Params) |  | params defines all the parameters of the module. |
| `markers` | [MarkerAccount](#provenance.marker.v1.MarkerAccount) | repeated | A collection of marker accounts to create on start |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/marker/v1/proposals.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/proposals.proto



<a name="provenance.marker.v1.AddMarkerProposal"></a>

### AddMarkerProposal
AddMarkerProposal is deprecated and can no longer be used.
Deprecated: This message is no longer usable. It is only still included for
backwards compatibility (e.g. looking up old governance proposals).
It is replaced by providing a MsgAddMarkerRequest in a governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `manager` | [string](#string) |  |  |
| `status` | [MarkerStatus](#provenance.marker.v1.MarkerStatus) |  |  |
| `marker_type` | [MarkerType](#provenance.marker.v1.MarkerType) |  |  |
| `access_list` | [AccessGrant](#provenance.marker.v1.AccessGrant) | repeated |  |
| `supply_fixed` | [bool](#bool) |  |  |
| `allow_governance_control` | [bool](#bool) |  |  |






<a name="provenance.marker.v1.ChangeStatusProposal"></a>

### ChangeStatusProposal
ChangeStatusProposal defines a governance proposal to administer a marker to change its status


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `new_status` | [MarkerStatus](#provenance.marker.v1.MarkerStatus) |  |  |






<a name="provenance.marker.v1.RemoveAdministratorProposal"></a>

### RemoveAdministratorProposal
RemoveAdministratorProposal defines a governance proposal to administer a marker and remove all permissions for a
given address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `removed_address` | [string](#string) | repeated |  |






<a name="provenance.marker.v1.SetAdministratorProposal"></a>

### SetAdministratorProposal
SetAdministratorProposal defines a governance proposal to administer a marker and set administrators with specific
access on the marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `access` | [AccessGrant](#provenance.marker.v1.AccessGrant) | repeated |  |






<a name="provenance.marker.v1.SetDenomMetadataProposal"></a>

### SetDenomMetadataProposal
SetDenomMetadataProposal defines a governance proposal to set the metadata for a denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `metadata` | [cosmos.bank.v1beta1.Metadata](#cosmos.bank.v1beta1.Metadata) |  |  |






<a name="provenance.marker.v1.SupplyDecreaseProposal"></a>

### SupplyDecreaseProposal
SupplyDecreaseProposal defines a governance proposal to administer a marker and decrease the total supply through
burning coin held within the marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="provenance.marker.v1.SupplyIncreaseProposal"></a>

### SupplyIncreaseProposal
SupplyIncreaseProposal defines a governance proposal to administer a marker and increase total supply of the marker
through minting coin and placing it within the marker or assigning it directly to an account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `target_address` | [string](#string) |  | an optional target address for the minted coin from this request |






<a name="provenance.marker.v1.WithdrawEscrowProposal"></a>

### WithdrawEscrowProposal
WithdrawEscrowProposal defines a governance proposal to withdraw escrow coins from a marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `denom` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |
| `target_address` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/marker/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/query.proto



<a name="provenance.marker.v1.Balance"></a>

### Balance
Balance defines an account address and balance pair used in queries for accounts holding a marker


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address is the address of the balance holder. |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | coins defines the different coins this balance holds. |






<a name="provenance.marker.v1.QueryAccessRequest"></a>

### QueryAccessRequest
QueryAccessRequest is the request type for the Query/MarkerAccess method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | address or denom for the marker |






<a name="provenance.marker.v1.QueryAccessResponse"></a>

### QueryAccessResponse
QueryAccessResponse is the response type for the Query/MarkerAccess method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `accounts` | [AccessGrant](#provenance.marker.v1.AccessGrant) | repeated |  |






<a name="provenance.marker.v1.QueryAccountDataRequest"></a>

### QueryAccountDataRequest
QueryAccountDataRequest is the request type for the Query/AccountData


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | The denomination to look up. |






<a name="provenance.marker.v1.QueryAccountDataResponse"></a>

### QueryAccountDataResponse
QueryAccountDataResponse is the response type for the Query/AccountData


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  | The accountdata for the requested denom. |






<a name="provenance.marker.v1.QueryAllMarkersRequest"></a>

### QueryAllMarkersRequest
QueryAllMarkersRequest is the request type for the Query/AllMarkers method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `status` | [MarkerStatus](#provenance.marker.v1.MarkerStatus) |  | Optional status to filter request |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.marker.v1.QueryAllMarkersResponse"></a>

### QueryAllMarkersResponse
QueryAllMarkersResponse is the response type for the Query/AllMarkers method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `markers` | [google.protobuf.Any](#google.protobuf.Any) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance.marker.v1.QueryDenomMetadataRequest"></a>

### QueryDenomMetadataRequest
QueryDenomMetadataRequest is the request type for Query/DenomMetadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |






<a name="provenance.marker.v1.QueryDenomMetadataResponse"></a>

### QueryDenomMetadataResponse
QueryDenomMetadataResponse is the response type for the Query/DenomMetadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata` | [cosmos.bank.v1beta1.Metadata](#cosmos.bank.v1beta1.Metadata) |  |  |






<a name="provenance.marker.v1.QueryEscrowRequest"></a>

### QueryEscrowRequest
QueryEscrowRequest is the request type for the Query/MarkerEscrow method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | address or denom for the marker |






<a name="provenance.marker.v1.QueryEscrowResponse"></a>

### QueryEscrowResponse
QueryEscrowResponse is the response type for the Query/MarkerEscrow method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `escrow` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="provenance.marker.v1.QueryHoldingRequest"></a>

### QueryHoldingRequest
QueryHoldingRequest is the request type for the Query/MarkerHolders method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | the address or denom of the marker |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.marker.v1.QueryHoldingResponse"></a>

### QueryHoldingResponse
QueryHoldingResponse is the response type for the Query/MarkerHolders method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balances` | [Balance](#provenance.marker.v1.Balance) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance.marker.v1.QueryMarkerRequest"></a>

### QueryMarkerRequest
QueryMarkerRequest is the request type for the Query/Marker method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | the address or denom of the marker |






<a name="provenance.marker.v1.QueryMarkerResponse"></a>

### QueryMarkerResponse
QueryMarkerResponse is the response type for the Query/Marker method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `marker` | [google.protobuf.Any](#google.protobuf.Any) |  |  |






<a name="provenance.marker.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance.marker.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.marker.v1.Params) |  | params defines the parameters of the module. |






<a name="provenance.marker.v1.QuerySupplyRequest"></a>

### QuerySupplyRequest
QuerySupplyRequest is the request type for the Query/MarkerSupply method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [string](#string) |  | address or denom for the marker |






<a name="provenance.marker.v1.QuerySupplyResponse"></a>

### QuerySupplyResponse
QuerySupplyResponse is the response type for the Query/MarkerSupply method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | amount is the supply of the marker. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.marker.v1.Query"></a>

### Query
Query defines the gRPC querier service for marker module.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#provenance.marker.v1.QueryParamsRequest) | [QueryParamsResponse](#provenance.marker.v1.QueryParamsResponse) | Params queries the parameters of x/bank module. | GET|/provenance/marker/v1/params|
| `AllMarkers` | [QueryAllMarkersRequest](#provenance.marker.v1.QueryAllMarkersRequest) | [QueryAllMarkersResponse](#provenance.marker.v1.QueryAllMarkersResponse) | Returns a list of all markers on the blockchain | GET|/provenance/marker/v1/all|
| `Marker` | [QueryMarkerRequest](#provenance.marker.v1.QueryMarkerRequest) | [QueryMarkerResponse](#provenance.marker.v1.QueryMarkerResponse) | query for a single marker by denom or address | GET|/provenance/marker/v1/detail/{id}|
| `Holding` | [QueryHoldingRequest](#provenance.marker.v1.QueryHoldingRequest) | [QueryHoldingResponse](#provenance.marker.v1.QueryHoldingResponse) | query for all accounts holding the given marker coins | GET|/provenance/marker/v1/holding/{id}|
| `Supply` | [QuerySupplyRequest](#provenance.marker.v1.QuerySupplyRequest) | [QuerySupplyResponse](#provenance.marker.v1.QuerySupplyResponse) | query for supply of coin on a marker account | GET|/provenance/marker/v1/supply/{id}|
| `Escrow` | [QueryEscrowRequest](#provenance.marker.v1.QueryEscrowRequest) | [QueryEscrowResponse](#provenance.marker.v1.QueryEscrowResponse) | query for coins on a marker account | GET|/provenance/marker/v1/escrow/{id}|
| `Access` | [QueryAccessRequest](#provenance.marker.v1.QueryAccessRequest) | [QueryAccessResponse](#provenance.marker.v1.QueryAccessResponse) | query for access records on an account | GET|/provenance/marker/v1/accesscontrol/{id}|
| `DenomMetadata` | [QueryDenomMetadataRequest](#provenance.marker.v1.QueryDenomMetadataRequest) | [QueryDenomMetadataResponse](#provenance.marker.v1.QueryDenomMetadataResponse) | query for access records on an account | GET|/provenance/marker/v1/getdenommetadata/{denom}|
| `AccountData` | [QueryAccountDataRequest](#provenance.marker.v1.QueryAccountDataRequest) | [QueryAccountDataResponse](#provenance.marker.v1.QueryAccountDataResponse) | query for account data associated with a denom | GET|/provenance/marker/v1/accountdata/{denom}|

 <!-- end services -->



<a name="provenance/marker/v1/si.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/si.proto


 <!-- end messages -->


<a name="provenance.marker.v1.SIPrefix"></a>

### SIPrefix
SIPrefix represents an International System of Units (SI) Prefix.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SI_PREFIX_NONE | 0 | 10^0 (none) |
| SI_PREFIX_DEKA | 1 | 10^1 deka da |
| SI_PREFIX_HECTO | 2 | 10^2 hecto h |
| SI_PREFIX_KILO | 3 | 10^3 kilo k |
| SI_PREFIX_MEGA | 6 | 10^6 mega M |
| SI_PREFIX_GIGA | 9 | 10^9 giga G |
| SI_PREFIX_TERA | 12 | 10^12 tera T |
| SI_PREFIX_PETA | 15 | 10^15 peta P |
| SI_PREFIX_EXA | 18 | 10^18 exa E |
| SI_PREFIX_ZETTA | 21 | 10^21 zetta Z |
| SI_PREFIX_YOTTA | 24 | 10^24 yotta Y |
| SI_PREFIX_DECI | -1 | 10^-1 deci d |
| SI_PREFIX_CENTI | -2 | 10^-2 centi c |
| SI_PREFIX_MILLI | -3 | 10^-3 milli m |
| SI_PREFIX_MICRO | -6 | 10^-6 micro µ |
| SI_PREFIX_NANO | -9 | 10^-9 nano n |
| SI_PREFIX_PICO | -12 | 10^-12 pico p |
| SI_PREFIX_FEMTO | -15 | 10^-15 femto f |
| SI_PREFIX_ATTO | -18 | 10^-18 atto a |
| SI_PREFIX_ZEPTO | -21 | 10^-21 zepto z |
| SI_PREFIX_YOCTO | -24 | 10^-24 yocto y |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/marker/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/tx.proto



<a name="provenance.marker.v1.MsgActivateRequest"></a>

### MsgActivateRequest
MsgActivateRequest defines the Msg/Activate request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgActivateResponse"></a>

### MsgActivateResponse
MsgActivateResponse defines the Msg/Activate response type






<a name="provenance.marker.v1.MsgAddAccessRequest"></a>

### MsgAddAccessRequest
MsgAddAccessRequest defines the Msg/AddAccess request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `access` | [AccessGrant](#provenance.marker.v1.AccessGrant) | repeated |  |






<a name="provenance.marker.v1.MsgAddAccessResponse"></a>

### MsgAddAccessResponse
MsgAddAccessResponse defines the Msg/AddAccess response type






<a name="provenance.marker.v1.MsgAddFinalizeActivateMarkerRequest"></a>

### MsgAddFinalizeActivateMarkerRequest
MsgAddFinalizeActivateMarkerRequest defines the Msg/AddFinalizeActivateMarker request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `manager` | [string](#string) |  |  |
| `from_address` | [string](#string) |  |  |
| `marker_type` | [MarkerType](#provenance.marker.v1.MarkerType) |  |  |
| `access_list` | [AccessGrant](#provenance.marker.v1.AccessGrant) | repeated |  |
| `supply_fixed` | [bool](#bool) |  |  |
| `allow_governance_control` | [bool](#bool) |  |  |
| `allow_forced_transfer` | [bool](#bool) |  |  |
| `required_attributes` | [string](#string) | repeated |  |






<a name="provenance.marker.v1.MsgAddFinalizeActivateMarkerResponse"></a>

### MsgAddFinalizeActivateMarkerResponse
MsgAddFinalizeActivateMarkerResponse defines the Msg/AddFinalizeActivateMarker response type






<a name="provenance.marker.v1.MsgAddMarkerRequest"></a>

### MsgAddMarkerRequest
MsgAddMarkerRequest defines the Msg/AddMarker request type.
If being provided as a governance proposal, set the from_address to the gov module's account address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `manager` | [string](#string) |  |  |
| `from_address` | [string](#string) |  |  |
| `status` | [MarkerStatus](#provenance.marker.v1.MarkerStatus) |  |  |
| `marker_type` | [MarkerType](#provenance.marker.v1.MarkerType) |  |  |
| `access_list` | [AccessGrant](#provenance.marker.v1.AccessGrant) | repeated |  |
| `supply_fixed` | [bool](#bool) |  |  |
| `allow_governance_control` | [bool](#bool) |  |  |
| `allow_forced_transfer` | [bool](#bool) |  |  |
| `required_attributes` | [string](#string) | repeated |  |






<a name="provenance.marker.v1.MsgAddMarkerResponse"></a>

### MsgAddMarkerResponse
MsgAddMarkerResponse defines the Msg/AddMarker response type






<a name="provenance.marker.v1.MsgBurnRequest"></a>

### MsgBurnRequest
MsgBurnRequest defines the Msg/Burn request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgBurnResponse"></a>

### MsgBurnResponse
MsgBurnResponse defines the Msg/Burn response type






<a name="provenance.marker.v1.MsgCancelRequest"></a>

### MsgCancelRequest
MsgCancelRequest defines the Msg/Cancel request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgCancelResponse"></a>

### MsgCancelResponse
MsgCancelResponse defines the Msg/Cancel response type






<a name="provenance.marker.v1.MsgDeleteAccessRequest"></a>

### MsgDeleteAccessRequest
MsgDeleteAccessRequest defines the Msg/DeleteAccess request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `removed_address` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgDeleteAccessResponse"></a>

### MsgDeleteAccessResponse
MsgDeleteAccessResponse defines the Msg/DeleteAccess response type






<a name="provenance.marker.v1.MsgDeleteRequest"></a>

### MsgDeleteRequest
MsgDeleteRequest defines the Msg/Delete request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgDeleteResponse"></a>

### MsgDeleteResponse
MsgDeleteResponse defines the Msg/Delete response type






<a name="provenance.marker.v1.MsgFinalizeRequest"></a>

### MsgFinalizeRequest
MsgFinalizeRequest defines the Msg/Finalize request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgFinalizeResponse"></a>

### MsgFinalizeResponse
MsgFinalizeResponse defines the Msg/Finalize response type






<a name="provenance.marker.v1.MsgGrantAllowanceRequest"></a>

### MsgGrantAllowanceRequest
MsgGrantAllowanceRequest validates permission to create a fee grant based on marker admin access. If
successful a feegrant is recorded where the marker account itself is the grantor


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `grantee` | [string](#string) |  | grantee is the address of the user being granted an allowance of another user's funds. |
| `allowance` | [google.protobuf.Any](#google.protobuf.Any) |  | allowance can be any of basic and filtered fee allowance (fee FeeGrant module). |






<a name="provenance.marker.v1.MsgGrantAllowanceResponse"></a>

### MsgGrantAllowanceResponse
MsgGrantAllowanceResponse defines the Msg/GrantAllowanceResponse response type.






<a name="provenance.marker.v1.MsgIbcTransferRequest"></a>

### MsgIbcTransferRequest
MsgIbcTransferRequest defines the Msg/IbcTransfer request type for markers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `transfer` | [ibc.applications.transfer.v1.MsgTransfer](#ibc.applications.transfer.v1.MsgTransfer) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgIbcTransferResponse"></a>

### MsgIbcTransferResponse
MsgIbcTransferResponse defines the Msg/IbcTransfer response type






<a name="provenance.marker.v1.MsgMintRequest"></a>

### MsgMintRequest
MsgMintRequest defines the Msg/Mint request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgMintResponse"></a>

### MsgMintResponse
MsgMintResponse defines the Msg/Mint response type






<a name="provenance.marker.v1.MsgSetAccountDataRequest"></a>

### MsgSetAccountDataRequest
MsgSetAccountDataRequest defines a msg to set/update/delete the account data for a marker.
Signer must have deposit authority or be a gov proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | The denomination of the marker to update. |
| `value` | [string](#string) |  | The desired accountdata value. |
| `signer` | [string](#string) |  | The signer of this message. Must have deposit authority or be the governance module account address. |






<a name="provenance.marker.v1.MsgSetAccountDataResponse"></a>

### MsgSetAccountDataResponse
MsgSetAccountDataResponse defines the Msg/SetAccountData response type






<a name="provenance.marker.v1.MsgSetDenomMetadataRequest"></a>

### MsgSetDenomMetadataRequest
MsgSetDenomMetadataRequest defines the Msg/SetDenomMetadata request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata` | [cosmos.bank.v1beta1.Metadata](#cosmos.bank.v1beta1.Metadata) |  |  |
| `administrator` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgSetDenomMetadataResponse"></a>

### MsgSetDenomMetadataResponse
MsgSetDenomMetadataResponse defines the Msg/SetDenomMetadata response type






<a name="provenance.marker.v1.MsgSupplyIncreaseProposalRequest"></a>

### MsgSupplyIncreaseProposalRequest
MsgSupplyIncreaseProposalRequest defines a governance proposal to administer a marker and increase total supply of
the marker through minting coin and placing it within the marker or assigning it directly to an account


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `target_address` | [string](#string) |  | an optional target address for the minted coin from this request |
| `authority` | [string](#string) |  | signer of the proposal |






<a name="provenance.marker.v1.MsgSupplyIncreaseProposalResponse"></a>

### MsgSupplyIncreaseProposalResponse
MsgSupplyIncreaseProposalResponse defines the Msg/SupplyIncreaseProposal response type






<a name="provenance.marker.v1.MsgTransferRequest"></a>

### MsgTransferRequest
MsgTransferRequest defines the Msg/Transfer request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `administrator` | [string](#string) |  |  |
| `from_address` | [string](#string) |  |  |
| `to_address` | [string](#string) |  |  |






<a name="provenance.marker.v1.MsgTransferResponse"></a>

### MsgTransferResponse
MsgTransferResponse defines the Msg/Transfer response type






<a name="provenance.marker.v1.MsgUpdateForcedTransferRequest"></a>

### MsgUpdateForcedTransferRequest
MsgUpdateForcedTransferRequest defines a msg to update the allow_forced_transfer field of a marker.
It is only usable via governance proposal.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | The denomination of the marker to update. |
| `allow_forced_transfer` | [bool](#bool) |  | Whether an admin can transfer restricted coins from a 3rd-party account without their signature. |
| `authority` | [string](#string) |  | The signer of this message. Must be the governance module account address. |






<a name="provenance.marker.v1.MsgUpdateForcedTransferResponse"></a>

### MsgUpdateForcedTransferResponse
MsgUpdateForcedTransferResponse defines the Msg/UpdateForcedTransfer response type






<a name="provenance.marker.v1.MsgUpdateRequiredAttributesRequest"></a>

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






<a name="provenance.marker.v1.MsgUpdateRequiredAttributesResponse"></a>

### MsgUpdateRequiredAttributesResponse
MsgUpdateRequiredAttributesResponse defines the Msg/UpdateRequiredAttributes response type






<a name="provenance.marker.v1.MsgUpdateSendDenyListRequest"></a>

### MsgUpdateSendDenyListRequest
MsgUpdateSendDenyListRequest defines a msg to add/remove addresses to send deny list for a resticted marker
signer must have transfer authority


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | The denomination of the marker to update. |
| `remove_denied_addresses` | [string](#string) | repeated | List of bech32 addresses to remove from the deny send list. |
| `add_denied_addresses` | [string](#string) | repeated | List of bech32 addresses to add to the deny send list. |
| `authority` | [string](#string) |  | The signer of the message. Must have admin authority to marker or be governance module account address. |






<a name="provenance.marker.v1.MsgUpdateSendDenyListResponse"></a>

### MsgUpdateSendDenyListResponse
MsgUpdateSendDenyListResponse defines the Msg/UpdateSendDenyList response type






<a name="provenance.marker.v1.MsgWithdrawRequest"></a>

### MsgWithdrawRequest
MsgWithdrawRequest defines the Msg/Withdraw request type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  |  |
| `administrator` | [string](#string) |  |  |
| `to_address` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated |  |






<a name="provenance.marker.v1.MsgWithdrawResponse"></a>

### MsgWithdrawResponse
MsgWithdrawResponse defines the Msg/Withdraw response type





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.marker.v1.Msg"></a>

### Msg
Msg defines the Marker Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Finalize` | [MsgFinalizeRequest](#provenance.marker.v1.MsgFinalizeRequest) | [MsgFinalizeResponse](#provenance.marker.v1.MsgFinalizeResponse) | Finalize | |
| `Activate` | [MsgActivateRequest](#provenance.marker.v1.MsgActivateRequest) | [MsgActivateResponse](#provenance.marker.v1.MsgActivateResponse) | Activate | |
| `Cancel` | [MsgCancelRequest](#provenance.marker.v1.MsgCancelRequest) | [MsgCancelResponse](#provenance.marker.v1.MsgCancelResponse) | Cancel | |
| `Delete` | [MsgDeleteRequest](#provenance.marker.v1.MsgDeleteRequest) | [MsgDeleteResponse](#provenance.marker.v1.MsgDeleteResponse) | Delete | |
| `Mint` | [MsgMintRequest](#provenance.marker.v1.MsgMintRequest) | [MsgMintResponse](#provenance.marker.v1.MsgMintResponse) | Mint | |
| `Burn` | [MsgBurnRequest](#provenance.marker.v1.MsgBurnRequest) | [MsgBurnResponse](#provenance.marker.v1.MsgBurnResponse) | Burn | |
| `AddAccess` | [MsgAddAccessRequest](#provenance.marker.v1.MsgAddAccessRequest) | [MsgAddAccessResponse](#provenance.marker.v1.MsgAddAccessResponse) | AddAccess | |
| `DeleteAccess` | [MsgDeleteAccessRequest](#provenance.marker.v1.MsgDeleteAccessRequest) | [MsgDeleteAccessResponse](#provenance.marker.v1.MsgDeleteAccessResponse) | DeleteAccess | |
| `Withdraw` | [MsgWithdrawRequest](#provenance.marker.v1.MsgWithdrawRequest) | [MsgWithdrawResponse](#provenance.marker.v1.MsgWithdrawResponse) | Withdraw | |
| `AddMarker` | [MsgAddMarkerRequest](#provenance.marker.v1.MsgAddMarkerRequest) | [MsgAddMarkerResponse](#provenance.marker.v1.MsgAddMarkerResponse) | AddMarker | |
| `Transfer` | [MsgTransferRequest](#provenance.marker.v1.MsgTransferRequest) | [MsgTransferResponse](#provenance.marker.v1.MsgTransferResponse) | Transfer marker denominated coin between accounts | |
| `IbcTransfer` | [MsgIbcTransferRequest](#provenance.marker.v1.MsgIbcTransferRequest) | [MsgIbcTransferResponse](#provenance.marker.v1.MsgIbcTransferResponse) | Transfer over ibc any marker(including restricted markers) between ibc accounts. The relayer is still needed to accomplish ibc middleware relays. | |
| `SetDenomMetadata` | [MsgSetDenomMetadataRequest](#provenance.marker.v1.MsgSetDenomMetadataRequest) | [MsgSetDenomMetadataResponse](#provenance.marker.v1.MsgSetDenomMetadataResponse) | Allows Denom Metadata (see bank module) to be set for the Marker's Denom | |
| `GrantAllowance` | [MsgGrantAllowanceRequest](#provenance.marker.v1.MsgGrantAllowanceRequest) | [MsgGrantAllowanceResponse](#provenance.marker.v1.MsgGrantAllowanceResponse) | GrantAllowance grants fee allowance to the grantee on the granter's account with the provided expiration time. | |
| `AddFinalizeActivateMarker` | [MsgAddFinalizeActivateMarkerRequest](#provenance.marker.v1.MsgAddFinalizeActivateMarkerRequest) | [MsgAddFinalizeActivateMarkerResponse](#provenance.marker.v1.MsgAddFinalizeActivateMarkerResponse) | AddFinalizeActivateMarker | |
| `SupplyIncreaseProposal` | [MsgSupplyIncreaseProposalRequest](#provenance.marker.v1.MsgSupplyIncreaseProposalRequest) | [MsgSupplyIncreaseProposalResponse](#provenance.marker.v1.MsgSupplyIncreaseProposalResponse) | SupplyIncreaseProposal can only be called via gov proposal | |
| `UpdateRequiredAttributes` | [MsgUpdateRequiredAttributesRequest](#provenance.marker.v1.MsgUpdateRequiredAttributesRequest) | [MsgUpdateRequiredAttributesResponse](#provenance.marker.v1.MsgUpdateRequiredAttributesResponse) | UpdateRequiredAttributes will only succeed if signer has transfer authority | |
| `UpdateForcedTransfer` | [MsgUpdateForcedTransferRequest](#provenance.marker.v1.MsgUpdateForcedTransferRequest) | [MsgUpdateForcedTransferResponse](#provenance.marker.v1.MsgUpdateForcedTransferResponse) | UpdateForcedTransfer updates the allow_forced_transfer field of a marker via governance proposal. | |
| `SetAccountData` | [MsgSetAccountDataRequest](#provenance.marker.v1.MsgSetAccountDataRequest) | [MsgSetAccountDataResponse](#provenance.marker.v1.MsgSetAccountDataResponse) | SetAccountData sets the accountdata for a denom. Signer must have deposit authority. | |
| `UpdateSendDenyList` | [MsgUpdateSendDenyListRequest](#provenance.marker.v1.MsgUpdateSendDenyListRequest) | [MsgUpdateSendDenyListResponse](#provenance.marker.v1.MsgUpdateSendDenyListResponse) | UpdateSendDenyList will only succeed if signer has admin authority | |

 <!-- end services -->



<a name="provenance/metadata/v1/events.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/events.proto



<a name="provenance.metadata.v1.EventContractSpecificationCreated"></a>

### EventContractSpecificationCreated
EventContractSpecificationCreated is an event message indicating a contract specification has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the specification id of the contract specification that was created. |






<a name="provenance.metadata.v1.EventContractSpecificationDeleted"></a>

### EventContractSpecificationDeleted
EventContractSpecificationDeleted is an event message indicating a contract specification has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the specification id of the contract specification that was deleted. |






<a name="provenance.metadata.v1.EventContractSpecificationUpdated"></a>

### EventContractSpecificationUpdated
EventContractSpecificationUpdated is an event message indicating a contract specification has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the specification id of the contract specification that was updated. |






<a name="provenance.metadata.v1.EventOSLocatorCreated"></a>

### EventOSLocatorCreated
EventOSLocatorCreated is an event message indicating an object store locator has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | owner is the owner in the object store locator that was created. |






<a name="provenance.metadata.v1.EventOSLocatorDeleted"></a>

### EventOSLocatorDeleted
EventOSLocatorDeleted is an event message indicating an object store locator has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | owner is the owner in the object store locator that was deleted. |






<a name="provenance.metadata.v1.EventOSLocatorUpdated"></a>

### EventOSLocatorUpdated
EventOSLocatorUpdated is an event message indicating an object store locator has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  | owner is the owner in the object store locator that was updated. |






<a name="provenance.metadata.v1.EventRecordCreated"></a>

### EventRecordCreated
EventRecordCreated is an event message indicating a record has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_addr` | [string](#string) |  | record_addr is the bech32 address string of the record id that was created. |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id this record belongs to. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this record belongs to. |






<a name="provenance.metadata.v1.EventRecordDeleted"></a>

### EventRecordDeleted
EventRecordDeleted is an event message indicating a record has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_addr` | [string](#string) |  | record is the bech32 address string of the record id that was deleted. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this record belonged to. |






<a name="provenance.metadata.v1.EventRecordSpecificationCreated"></a>

### EventRecordSpecificationCreated
EventRecordSpecificationCreated is an event message indicating a record specification has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specification_addr` | [string](#string) |  | record_specification_addr is the bech32 address string of the specification id of the record specification that was created. |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the contract specification id this record specification belongs to. |






<a name="provenance.metadata.v1.EventRecordSpecificationDeleted"></a>

### EventRecordSpecificationDeleted
EventRecordSpecificationDeleted is an event message indicating a record specification has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specification_addr` | [string](#string) |  | record_specification_addr is the bech32 address string of the specification id of the record specification that was deleted. |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the contract specification id this record specification belongs to. |






<a name="provenance.metadata.v1.EventRecordSpecificationUpdated"></a>

### EventRecordSpecificationUpdated
EventRecordSpecificationUpdated is an event message indicating a record specification has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specification_addr` | [string](#string) |  | record_specification_addr is the bech32 address string of the specification id of the record specification that was updated. |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the bech32 address string of the contract specification id this record specification belongs to. |






<a name="provenance.metadata.v1.EventRecordUpdated"></a>

### EventRecordUpdated
EventRecordUpdated is an event message indicating a record has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_addr` | [string](#string) |  | record_addr is the bech32 address string of the record id that was updated. |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id this record belongs to. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this record belongs to. |






<a name="provenance.metadata.v1.EventScopeCreated"></a>

### EventScopeCreated
EventScopeCreated is an event message indicating a scope has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id that was created. |






<a name="provenance.metadata.v1.EventScopeDeleted"></a>

### EventScopeDeleted
EventScopeDeleted is an event message indicating a scope has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id that was deleted. |






<a name="provenance.metadata.v1.EventScopeSpecificationCreated"></a>

### EventScopeSpecificationCreated
EventScopeSpecificationCreated is an event message indicating a scope specification has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specification_addr` | [string](#string) |  | scope_specification_addr is the bech32 address string of the specification id of the scope specification that was created. |






<a name="provenance.metadata.v1.EventScopeSpecificationDeleted"></a>

### EventScopeSpecificationDeleted
EventScopeSpecificationDeleted is an event message indicating a scope specification has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specification_addr` | [string](#string) |  | scope_specification_addr is the bech32 address string of the specification id of the scope specification that was deleted. |






<a name="provenance.metadata.v1.EventScopeSpecificationUpdated"></a>

### EventScopeSpecificationUpdated
EventScopeSpecificationUpdated is an event message indicating a scope specification has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specification_addr` | [string](#string) |  | scope_specification_addr is the bech32 address string of the specification id of the scope specification that was updated. |






<a name="provenance.metadata.v1.EventScopeUpdated"></a>

### EventScopeUpdated
EventScopeUpdated is an event message indicating a scope has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id that was updated. |






<a name="provenance.metadata.v1.EventSessionCreated"></a>

### EventSessionCreated
EventSessionCreated is an event message indicating a session has been created.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id that was created. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this session belongs to. |






<a name="provenance.metadata.v1.EventSessionDeleted"></a>

### EventSessionDeleted
EventSessionDeleted is an event message indicating a session has been deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id that was deleted. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this session belongs to. |






<a name="provenance.metadata.v1.EventSessionUpdated"></a>

### EventSessionUpdated
EventSessionUpdated is an event message indicating a session has been updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_addr` | [string](#string) |  | session_addr is the bech32 address string of the session id that was updated. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 address string of the scope id this session belongs to. |






<a name="provenance.metadata.v1.EventTxCompleted"></a>

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



<a name="provenance/metadata/v1/metadata.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/metadata.proto



<a name="provenance.metadata.v1.ContractSpecIdInfo"></a>

### ContractSpecIdInfo
ContractSpecIdInfo contains various info regarding a contract specification id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_spec_id` | [bytes](#bytes) |  | contract_spec_id is the raw bytes of the contract specification address. |
| `contract_spec_id_prefix` | [bytes](#bytes) |  | contract_spec_id_prefix is the prefix portion of the contract_spec_id. |
| `contract_spec_id_contract_spec_uuid` | [bytes](#bytes) |  | contract_spec_id_contract_spec_uuid is the contract_spec_uuid portion of the contract_spec_id. |
| `contract_spec_addr` | [string](#string) |  | contract_spec_addr is the bech32 string version of the contract_spec_id. |
| `contract_spec_uuid` | [string](#string) |  | contract_spec_uuid is the uuid hex string of the contract_spec_id_contract_spec_uuid. |






<a name="provenance.metadata.v1.Params"></a>

### Params
Params defines the set of params for the metadata module.






<a name="provenance.metadata.v1.RecordIdInfo"></a>

### RecordIdInfo
RecordIdInfo contains various info regarding a record id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_id` | [bytes](#bytes) |  | record_id is the raw bytes of the record address. |
| `record_id_prefix` | [bytes](#bytes) |  | record_id_prefix is the prefix portion of the record_id. |
| `record_id_scope_uuid` | [bytes](#bytes) |  | record_id_scope_uuid is the scope_uuid portion of the record_id. |
| `record_id_hashed_name` | [bytes](#bytes) |  | record_id_hashed_name is the hashed name portion of the record_id. |
| `record_addr` | [string](#string) |  | record_addr is the bech32 string version of the record_id. |
| `scope_id_info` | [ScopeIdInfo](#provenance.metadata.v1.ScopeIdInfo) |  | scope_id_info is information about the scope id referenced in the record_id. |






<a name="provenance.metadata.v1.RecordSpecIdInfo"></a>

### RecordSpecIdInfo
RecordSpecIdInfo contains various info regarding a record specification id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_spec_id` | [bytes](#bytes) |  | record_spec_id is the raw bytes of the record specification address. |
| `record_spec_id_prefix` | [bytes](#bytes) |  | record_spec_id_prefix is the prefix portion of the record_spec_id. |
| `record_spec_id_contract_spec_uuid` | [bytes](#bytes) |  | record_spec_id_contract_spec_uuid is the contract_spec_uuid portion of the record_spec_id. |
| `record_spec_id_hashed_name` | [bytes](#bytes) |  | record_spec_id_hashed_name is the hashed name portion of the record_spec_id. |
| `record_spec_addr` | [string](#string) |  | record_spec_addr is the bech32 string version of the record_spec_id. |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance.metadata.v1.ContractSpecIdInfo) |  | contract_spec_id_info is information about the contract spec id referenced in the record_spec_id. |






<a name="provenance.metadata.v1.ScopeIdInfo"></a>

### ScopeIdInfo
ScopeIdInfo contains various info regarding a scope id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope_id is the raw bytes of the scope address. |
| `scope_id_prefix` | [bytes](#bytes) |  | scope_id_prefix is the prefix portion of the scope_id. |
| `scope_id_scope_uuid` | [bytes](#bytes) |  | scope_id_scope_uuid is the scope_uuid portion of the scope_id. |
| `scope_addr` | [string](#string) |  | scope_addr is the bech32 string version of the scope_id. |
| `scope_uuid` | [string](#string) |  | scope_uuid is the uuid hex string of the scope_id_scope_uuid. |






<a name="provenance.metadata.v1.ScopeSpecIdInfo"></a>

### ScopeSpecIdInfo
ScopeSpecIdInfo contains various info regarding a scope specification id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_spec_id` | [bytes](#bytes) |  | scope_spec_id is the raw bytes of the scope specification address. |
| `scope_spec_id_prefix` | [bytes](#bytes) |  | scope_spec_id_prefix is the prefix portion of the scope_spec_id. |
| `scope_spec_id_scope_spec_uuid` | [bytes](#bytes) |  | scope_spec_id_scope_spec_uuid is the scope_spec_uuid portion of the scope_spec_id. |
| `scope_spec_addr` | [string](#string) |  | scope_spec_addr is the bech32 string version of the scope_spec_id. |
| `scope_spec_uuid` | [string](#string) |  | scope_spec_uuid is the uuid hex string of the scope_spec_id_scope_spec_uuid. |






<a name="provenance.metadata.v1.SessionIdInfo"></a>

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
| `scope_id_info` | [ScopeIdInfo](#provenance.metadata.v1.ScopeIdInfo) |  | scope_id_info is information about the scope id referenced in the session_id. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/metadata/v1/specification.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/specification.proto



<a name="provenance.metadata.v1.ContractSpecification"></a>

### ContractSpecification
ContractSpecification defines the required parties, resources, conditions, and consideration outputs for a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | unique identifier for this specification on chain |
| `description` | [Description](#provenance.metadata.v1.Description) |  | Description information for this contract specification |
| `owner_addresses` | [string](#string) | repeated | Address of the account that owns this specificaiton |
| `parties_involved` | [PartyType](#provenance.metadata.v1.PartyType) | repeated | a list of party roles that must be fullfilled when signing a transaction for this contract specification |
| `resource_id` | [bytes](#bytes) |  | the address of a record on chain that represents this contract |
| `hash` | [string](#string) |  | the hash of contract binary (off-chain instance) |
| `class_name` | [string](#string) |  | name of the class/type of this contract executable |






<a name="provenance.metadata.v1.Description"></a>

### Description
Description holds general information that is handy to associate with a structure.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | A Name for this thing. |
| `description` | [string](#string) |  | A description of this thing. |
| `website_url` | [string](#string) |  | URL to find even more info. |
| `icon_url` | [string](#string) |  | URL of an icon. |






<a name="provenance.metadata.v1.InputSpecification"></a>

### InputSpecification
InputSpecification defines a name, type_name, and source reference (either on or off chain) to define an input
parameter


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | name for this input |
| `type_name` | [string](#string) |  | a type_name (typically a proto name or class_name) |
| `record_id` | [bytes](#bytes) |  | the address of a record on chain (For Established Records) |
| `hash` | [string](#string) |  | the hash of an off-chain piece of information (For Proposed Records) |






<a name="provenance.metadata.v1.RecordSpecification"></a>

### RecordSpecification
RecordSpecification defines the specification for a Record including allowed/required inputs/outputs


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | unique identifier for this specification on chain |
| `name` | [string](#string) |  | Name of Record that will be created when this specification is used |
| `inputs` | [InputSpecification](#provenance.metadata.v1.InputSpecification) | repeated | A set of inputs that must be satisified to apply this RecordSpecification and create a Record |
| `type_name` | [string](#string) |  | A type name for data associated with this record (typically a class or proto name) |
| `result_type` | [DefinitionType](#provenance.metadata.v1.DefinitionType) |  | Type of result for this record specification (must be RECORD or RECORD_LIST) |
| `responsible_parties` | [PartyType](#provenance.metadata.v1.PartyType) | repeated | Type of party responsible for this record |






<a name="provenance.metadata.v1.ScopeSpecification"></a>

### ScopeSpecification
ScopeSpecification defines the required parties, resources, conditions, and consideration outputs for a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | unique identifier for this specification on chain |
| `description` | [Description](#provenance.metadata.v1.Description) |  | General information about this scope specification. |
| `owner_addresses` | [string](#string) | repeated | Addresses of the owners of this scope specification. |
| `parties_involved` | [PartyType](#provenance.metadata.v1.PartyType) | repeated | A list of parties that must be present on a scope (and their associated roles) |
| `contract_spec_ids` | [bytes](#bytes) | repeated | A list of contract specification ids allowed for a scope based on this specification. |





 <!-- end messages -->


<a name="provenance.metadata.v1.DefinitionType"></a>

### DefinitionType
DefinitionType indicates the required definition type for this value

| Name | Number | Description |
| ---- | ------ | ----------- |
| DEFINITION_TYPE_UNSPECIFIED | 0 | DEFINITION_TYPE_UNSPECIFIED indicates an unknown/invalid value |
| DEFINITION_TYPE_PROPOSED | 1 | DEFINITION_TYPE_PROPOSED indicates a proposed value is used here (a record that is not on-chain) |
| DEFINITION_TYPE_RECORD | 2 | DEFINITION_TYPE_RECORD indicates the value must be a reference to a record on chain |
| DEFINITION_TYPE_RECORD_LIST | 3 | DEFINITION_TYPE_RECORD_LIST indicates the value maybe a reference to a collection of values on chain having the same name |



<a name="provenance.metadata.v1.PartyType"></a>

### PartyType
PartyType are the different roles parties on a contract may use

| Name | Number | Description |
| ---- | ------ | ----------- |
| PARTY_TYPE_UNSPECIFIED | 0 | PARTY_TYPE_UNSPECIFIED is an error condition |
| PARTY_TYPE_ORIGINATOR | 1 | PARTY_TYPE_ORIGINATOR is an asset originator |
| PARTY_TYPE_SERVICER | 2 | PARTY_TYPE_SERVICER provides debt servicing functions |
| PARTY_TYPE_INVESTOR | 3 | PARTY_TYPE_INVESTOR is a generic investor |
| PARTY_TYPE_CUSTODIAN | 4 | PARTY_TYPE_CUSTODIAN is an entity that provides custodian services for assets |
| PARTY_TYPE_OWNER | 5 | PARTY_TYPE_OWNER indicates this party is an owner of the item |
| PARTY_TYPE_AFFILIATE | 6 | PARTY_TYPE_AFFILIATE is a party with an affiliate agreement |
| PARTY_TYPE_OMNIBUS | 7 | PARTY_TYPE_OMNIBUS is a special type of party that controls an omnibus bank account |
| PARTY_TYPE_PROVENANCE | 8 | PARTY_TYPE_PROVENANCE is used to indicate this party represents the blockchain or a smart contract action |
| PARTY_TYPE_CONTROLLER | 10 | PARTY_TYPE_CONTROLLER is an entity which controls a specific asset on chain (ie enote) |
| PARTY_TYPE_VALIDATOR | 11 | PARTY_TYPE_VALIDATOR is an entity which validates given assets on chain |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/metadata/v1/scope.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/scope.proto



<a name="provenance.metadata.v1.AuditFields"></a>

### AuditFields
AuditFields capture information about the last account to make modifications and when they were made


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `created_date` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | the date/time when this entry was created |
| `created_by` | [string](#string) |  | the address of the account that created this record |
| `updated_date` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | the date/time when this entry was last updated |
| `updated_by` | [string](#string) |  | the address of the account that modified this record |
| `version` | [uint32](#uint32) |  | an optional version number that is incremented with each update |
| `message` | [string](#string) |  | an optional message associated with the creation/update event |






<a name="provenance.metadata.v1.Party"></a>

### Party
A Party is an address with/in a given role associated with a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address of the account (on chain) |
| `role` | [PartyType](#provenance.metadata.v1.PartyType) |  | a role for this account within the context of the processes used |
| `optional` | [bool](#bool) |  | whether this party's signature is optional |






<a name="provenance.metadata.v1.Process"></a>

### Process
Process contains information used to uniquely identify what was used to generate this record


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | the address of a smart contract used for this process |
| `hash` | [string](#string) |  | the hash of an off-chain process used |
| `name` | [string](#string) |  | a name associated with the process (type_name, classname or smart contract common name) |
| `method` | [string](#string) |  | method is a name or reference to a specific operation (method) within a class/contract that was invoked |






<a name="provenance.metadata.v1.Record"></a>

### Record
A record (of fact) is attached to a session or each consideration output from a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | name/identifier for this record. Value must be unique within the scope. Also known as a Fact name |
| `session_id` | [bytes](#bytes) |  | id of the session context that was used to create this record (use with filtered kvprefix iterator) |
| `process` | [Process](#provenance.metadata.v1.Process) |  | process contain information used to uniquely identify an execution on or off chain that generated this record |
| `inputs` | [RecordInput](#provenance.metadata.v1.RecordInput) | repeated | inputs used with the process to achieve the output on this record |
| `outputs` | [RecordOutput](#provenance.metadata.v1.RecordOutput) | repeated | output(s) is the results of executing the process on the given process indicated in this record |
| `specification_id` | [bytes](#bytes) |  | specification_id is the id of the record specification that was used to create this record. |






<a name="provenance.metadata.v1.RecordInput"></a>

### RecordInput
Tracks the inputs used to establish this record


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | Name value included to link back to the definition spec. |
| `record_id` | [bytes](#bytes) |  | the address of a record on chain (For Established Records) |
| `hash` | [string](#string) |  | the hash of an off-chain piece of information (For Proposed Records) |
| `type_name` | [string](#string) |  | from proposed fact structure to unmarshal |
| `status` | [RecordInputStatus](#provenance.metadata.v1.RecordInputStatus) |  | Indicates if this input was a recorded fact on chain or just a given hashed input |






<a name="provenance.metadata.v1.RecordOutput"></a>

### RecordOutput
RecordOutput encapsulates the output of a process recorded on chain


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `hash` | [string](#string) |  | Hash of the data output that was output/generated for this record |
| `status` | [ResultStatus](#provenance.metadata.v1.ResultStatus) |  | Status of the process execution associated with this output indicating success,failure, or pending |






<a name="provenance.metadata.v1.Scope"></a>

### Scope
Scope defines a root reference for a collection of records owned by one or more parties.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | Unique ID for this scope. Implements sdk.Address interface for use where addresses are required in Cosmos |
| `specification_id` | [bytes](#bytes) |  | the scope specification that contains the specifications for data elements allowed within this scope |
| `owners` | [Party](#provenance.metadata.v1.Party) | repeated | These parties represent top level owners of the records within. These parties must sign any requests that modify the data within the scope. These addresses are in union with parties listed on the sessions. |
| `data_access` | [string](#string) | repeated | Addresses in this list are authorized to receive off-chain data associated with this scope. |
| `value_owner_address` | [string](#string) |  | An address that controls the value associated with this scope. Standard blockchain accounts and marker accounts are supported for this value. This attribute may only be changed by the entity indicated once it is set. |
| `require_party_rollup` | [bool](#bool) |  | Whether all parties in this scope and its sessions must be present in this scope's owners field. This also enables use of optional=true scope owners and session parties. |






<a name="provenance.metadata.v1.Session"></a>

### Session
Session defines an execution context against a specific specification instance.
The context will have a specification and set of parties involved.

NOTE: When there are no more Records within a Scope that reference a Session, the Session is removed.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_id` | [bytes](#bytes) |  |  |
| `specification_id` | [bytes](#bytes) |  | unique id of the contract specification that was used to create this session. |
| `parties` | [Party](#provenance.metadata.v1.Party) | repeated | parties is the set of identities that signed this contract |
| `name` | [string](#string) |  | name to associate with this session execution context, typically classname |
| `context` | [bytes](#bytes) |  | context is a field for storing client specific data associated with a session. |
| `audit` | [AuditFields](#provenance.metadata.v1.AuditFields) |  | Created by, updated by, timestamps, version number, and related info. |





 <!-- end messages -->


<a name="provenance.metadata.v1.RecordInputStatus"></a>

### RecordInputStatus
A set of types for inputs on a record (of fact)

| Name | Number | Description |
| ---- | ------ | ----------- |
| RECORD_INPUT_STATUS_UNSPECIFIED | 0 | RECORD_INPUT_STATUS_UNSPECIFIED indicates an invalid/unknown input type |
| RECORD_INPUT_STATUS_PROPOSED | 1 | RECORD_INPUT_STATUS_PROPOSED indicates this input was an arbitrary piece of data that was hashed |
| RECORD_INPUT_STATUS_RECORD | 2 | RECORD_INPUT_STATUS_RECORD indicates this input is a reference to a previously recorded fact on blockchain |



<a name="provenance.metadata.v1.ResultStatus"></a>

### ResultStatus
ResultStatus indicates the various states of execution of a record

| Name | Number | Description |
| ---- | ------ | ----------- |
| RESULT_STATUS_UNSPECIFIED | 0 | RESULT_STATUS_UNSPECIFIED indicates an unset condition |
| RESULT_STATUS_PASS | 1 | RESULT_STATUS_PASS indicates the execution was successful |
| RESULT_STATUS_SKIP | 2 | RESULT_STATUS_SKIP indicates condition/consideration was skipped due to missing inputs or delayed execution |
| RESULT_STATUS_FAIL | 3 | RESULT_STATUS_FAIL indicates the execution of the condition/consideration failed. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/metadata/v1/objectstore.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/objectstore.proto



<a name="provenance.metadata.v1.OSLocatorParams"></a>

### OSLocatorParams
Params defines the parameters for the metadata-locator module methods.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_uri_length` | [uint32](#uint32) |  |  |






<a name="provenance.metadata.v1.ObjectStoreLocator"></a>

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



<a name="provenance/metadata/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/genesis.proto



<a name="provenance.metadata.v1.GenesisState"></a>

### GenesisState
GenesisState defines the account module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.metadata.v1.Params) |  | params defines all the parameters of the module. |
| `scopes` | [Scope](#provenance.metadata.v1.Scope) | repeated | A collection of metadata scopes and specs to create on start |
| `sessions` | [Session](#provenance.metadata.v1.Session) | repeated |  |
| `records` | [Record](#provenance.metadata.v1.Record) | repeated |  |
| `scope_specifications` | [ScopeSpecification](#provenance.metadata.v1.ScopeSpecification) | repeated |  |
| `contract_specifications` | [ContractSpecification](#provenance.metadata.v1.ContractSpecification) | repeated |  |
| `record_specifications` | [RecordSpecification](#provenance.metadata.v1.RecordSpecification) | repeated |  |
| `o_s_locator_params` | [OSLocatorParams](#provenance.metadata.v1.OSLocatorParams) |  |  |
| `object_store_locators` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/metadata/v1/p8e/p8e.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/p8e/p8e.proto



<a name="provenance.metadata.v1.p8e.Condition"></a>

### Condition
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `condition_name` | [string](#string) |  |  |
| `result` | [ExecutionResult](#provenance.metadata.v1.p8e.ExecutionResult) |  |  |






<a name="provenance.metadata.v1.p8e.ConditionSpec"></a>

### ConditionSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `func_name` | [string](#string) |  |  |
| `input_specs` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) | repeated |  |
| `output_spec` | [OutputSpec](#provenance.metadata.v1.p8e.OutputSpec) |  |  |






<a name="provenance.metadata.v1.p8e.Consideration"></a>

### Consideration
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `consideration_name` | [string](#string) |  |  |
| `inputs` | [ProposedFact](#provenance.metadata.v1.p8e.ProposedFact) | repeated |  |
| `result` | [ExecutionResult](#provenance.metadata.v1.p8e.ExecutionResult) |  |  |






<a name="provenance.metadata.v1.p8e.ConsiderationSpec"></a>

### ConsiderationSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `func_name` | [string](#string) |  |  |
| `responsible_party` | [PartyType](#provenance.metadata.v1.p8e.PartyType) |  |  |
| `input_specs` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) | repeated |  |
| `output_spec` | [OutputSpec](#provenance.metadata.v1.p8e.OutputSpec) |  |  |






<a name="provenance.metadata.v1.p8e.Contract"></a>

### Contract
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `definition` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) |  |  |
| `spec` | [Fact](#provenance.metadata.v1.p8e.Fact) |  |  |
| `invoker` | [SigningAndEncryptionPublicKeys](#provenance.metadata.v1.p8e.SigningAndEncryptionPublicKeys) |  |  |
| `inputs` | [Fact](#provenance.metadata.v1.p8e.Fact) | repeated |  |
| `conditions` | [Condition](#provenance.metadata.v1.p8e.Condition) | repeated | **Deprecated.**  |
| `considerations` | [Consideration](#provenance.metadata.v1.p8e.Consideration) | repeated |  |
| `recitals` | [Recital](#provenance.metadata.v1.p8e.Recital) | repeated |  |
| `times_executed` | [int32](#int32) |  |  |
| `start_time` | [Timestamp](#provenance.metadata.v1.p8e.Timestamp) |  |  |
| `context` | [bytes](#bytes) |  |  |






<a name="provenance.metadata.v1.p8e.ContractSpec"></a>

### ContractSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `definition` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) |  |  |
| `input_specs` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) | repeated |  |
| `parties_involved` | [PartyType](#provenance.metadata.v1.p8e.PartyType) | repeated |  |
| `condition_specs` | [ConditionSpec](#provenance.metadata.v1.p8e.ConditionSpec) | repeated |  |
| `consideration_specs` | [ConsiderationSpec](#provenance.metadata.v1.p8e.ConsiderationSpec) | repeated |  |






<a name="provenance.metadata.v1.p8e.DefinitionSpec"></a>

### DefinitionSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `resource_location` | [Location](#provenance.metadata.v1.p8e.Location) |  |  |
| `signature` | [Signature](#provenance.metadata.v1.p8e.Signature) |  |  |
| `type` | [DefinitionSpecType](#provenance.metadata.v1.p8e.DefinitionSpecType) |  |  |






<a name="provenance.metadata.v1.p8e.ExecutionResult"></a>

### ExecutionResult
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `output` | [ProposedFact](#provenance.metadata.v1.p8e.ProposedFact) |  |  |
| `result` | [ExecutionResultType](#provenance.metadata.v1.p8e.ExecutionResultType) |  |  |
| `recorded_at` | [Timestamp](#provenance.metadata.v1.p8e.Timestamp) |  |  |
| `error_message` | [string](#string) |  |  |






<a name="provenance.metadata.v1.p8e.Fact"></a>

### Fact
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `data_location` | [Location](#provenance.metadata.v1.p8e.Location) |  |  |






<a name="provenance.metadata.v1.p8e.Location"></a>

### Location
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ref` | [ProvenanceReference](#provenance.metadata.v1.p8e.ProvenanceReference) |  |  |
| `classname` | [string](#string) |  |  |






<a name="provenance.metadata.v1.p8e.OutputSpec"></a>

### OutputSpec
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `spec` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) |  |  |






<a name="provenance.metadata.v1.p8e.ProposedFact"></a>

### ProposedFact
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `hash` | [string](#string) |  |  |
| `classname` | [string](#string) |  |  |
| `ancestor` | [ProvenanceReference](#provenance.metadata.v1.p8e.ProvenanceReference) |  |  |






<a name="provenance.metadata.v1.p8e.ProvenanceReference"></a>

### ProvenanceReference
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_uuid` | [UUID](#provenance.metadata.v1.p8e.UUID) |  |  |
| `group_uuid` | [UUID](#provenance.metadata.v1.p8e.UUID) |  |  |
| `hash` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |






<a name="provenance.metadata.v1.p8e.PublicKey"></a>

### PublicKey
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `public_key_bytes` | [bytes](#bytes) |  |  |
| `type` | [PublicKeyType](#provenance.metadata.v1.p8e.PublicKeyType) |  |  |
| `curve` | [PublicKeyCurve](#provenance.metadata.v1.p8e.PublicKeyCurve) |  |  |






<a name="provenance.metadata.v1.p8e.Recital"></a>

### Recital
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signer_role` | [PartyType](#provenance.metadata.v1.p8e.PartyType) |  |  |
| `signer` | [SigningAndEncryptionPublicKeys](#provenance.metadata.v1.p8e.SigningAndEncryptionPublicKeys) |  |  |
| `address` | [bytes](#bytes) |  |  |






<a name="provenance.metadata.v1.p8e.Recitals"></a>

### Recitals
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `parties` | [Recital](#provenance.metadata.v1.p8e.Recital) | repeated |  |






<a name="provenance.metadata.v1.p8e.Signature"></a>

### Signature
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `algo` | [string](#string) |  |  |
| `provider` | [string](#string) |  |  |
| `signature` | [string](#string) |  |  |
| `signer` | [SigningAndEncryptionPublicKeys](#provenance.metadata.v1.p8e.SigningAndEncryptionPublicKeys) |  |  |






<a name="provenance.metadata.v1.p8e.SignatureSet"></a>

### SignatureSet
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signatures` | [Signature](#provenance.metadata.v1.p8e.Signature) | repeated |  |






<a name="provenance.metadata.v1.p8e.SigningAndEncryptionPublicKeys"></a>

### SigningAndEncryptionPublicKeys
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signing_public_key` | [PublicKey](#provenance.metadata.v1.p8e.PublicKey) |  |  |
| `encryption_public_key` | [PublicKey](#provenance.metadata.v1.p8e.PublicKey) |  |  |






<a name="provenance.metadata.v1.p8e.Timestamp"></a>

### Timestamp
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `seconds` | [int64](#int64) |  |  |
| `nanos` | [int32](#int32) |  |  |






<a name="provenance.metadata.v1.p8e.UUID"></a>

### UUID
Deprecated: Do not use.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  |  |





 <!-- end messages -->


<a name="provenance.metadata.v1.p8e.DefinitionSpecType"></a>

### DefinitionSpecType
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| DEFINITION_SPEC_TYPE_UNKNOWN | 0 | Deprecated: Do not use. |
| DEFINITION_SPEC_TYPE_PROPOSED | 1 | Deprecated: Do not use. |
| DEFINITION_SPEC_TYPE_FACT | 2 | Deprecated: Do not use. |
| DEFINITION_SPEC_TYPE_FACT_LIST | 3 | Deprecated: Do not use. |



<a name="provenance.metadata.v1.p8e.ExecutionResultType"></a>

### ExecutionResultType
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| RESULT_TYPE_UNKNOWN | 0 | Deprecated: Do not use. |
| RESULT_TYPE_PASS | 1 | Deprecated: Do not use. |
| RESULT_TYPE_SKIP | 2 | Deprecated: Do not use. |
| RESULT_TYPE_FAIL | 3 | Deprecated: Do not use. |



<a name="provenance.metadata.v1.p8e.PartyType"></a>

### PartyType
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PARTY_TYPE_UNKNOWN | 0 | Deprecated: Do not use. |
| PARTY_TYPE_ORIGINATOR | 1 | Deprecated: Do not use. |
| PARTY_TYPE_SERVICER | 2 | Deprecated: Do not use. |
| PARTY_TYPE_INVESTOR | 3 | Deprecated: Do not use. |
| PARTY_TYPE_CUSTODIAN | 4 | Deprecated: Do not use. |
| PARTY_TYPE_OWNER | 5 | Deprecated: Do not use. |
| PARTY_TYPE_AFFILIATE | 6 | Deprecated: Do not use. |
| PARTY_TYPE_OMNIBUS | 7 | Deprecated: Do not use. |
| PARTY_TYPE_PROVENANCE | 8 | Deprecated: Do not use. |
| PARTY_TYPE_MARKER | 9 | Deprecated: Do not use. |
| PARTY_TYPE_CONTROLLER | 10 | Deprecated: Do not use. |
| PARTY_TYPE_VALIDATOR | 11 | Deprecated: Do not use. |



<a name="provenance.metadata.v1.p8e.PublicKeyCurve"></a>

### PublicKeyCurve
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SECP256K1 | 0 | Deprecated: Do not use. |
| P256 | 1 | Deprecated: Do not use. |



<a name="provenance.metadata.v1.p8e.PublicKeyType"></a>

### PublicKeyType
Deprecated: Do not use.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ELLIPTIC | 0 | Deprecated: Do not use. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/metadata/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/query.proto



<a name="provenance.metadata.v1.AccountDataRequest"></a>

### AccountDataRequest
AccountDataRequest is the request type for the Query/AccountData RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata_addr` | [bytes](#bytes) |  | The metadata address to look up. Currently, only scope ids are supported. |






<a name="provenance.metadata.v1.AccountDataResponse"></a>

### AccountDataResponse
AccountDataResponse is the response type for the Query/AccountData RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  | The accountdata for the requested metadata address. |






<a name="provenance.metadata.v1.ContractSpecificationRequest"></a>

### ContractSpecificationRequest
ContractSpecificationRequest is the request type for the Query/ContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [string](#string) |  | specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84 or a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn. It can also be a record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. |
| `include_record_specs` | [bool](#bool) |  | include_record_specs is a flag for whether to include the the record specifications of this contract specification in the response. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance.metadata.v1.ContractSpecificationResponse"></a>

### ContractSpecificationResponse
ContractSpecificationResponse is the response type for the Query/ContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification` | [ContractSpecificationWrapper](#provenance.metadata.v1.ContractSpecificationWrapper) |  | contract_specification is the wrapped contract specification. |
| `record_specifications` | [RecordSpecificationWrapper](#provenance.metadata.v1.RecordSpecificationWrapper) | repeated | record_specifications is any number or wrapped record specifications associated with this contract_specification (if requested). |
| `request` | [ContractSpecificationRequest](#provenance.metadata.v1.ContractSpecificationRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.ContractSpecificationWrapper"></a>

### ContractSpecificationWrapper
ContractSpecificationWrapper contains a single contract specification and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [ContractSpecification](#provenance.metadata.v1.ContractSpecification) |  | specification is the on-chain contract specification message. |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance.metadata.v1.ContractSpecIdInfo) |  | contract_spec_id_info contains information about the id/address of the contract specification. |






<a name="provenance.metadata.v1.ContractSpecificationsAllRequest"></a>

### ContractSpecificationsAllRequest
ContractSpecificationsAllRequest is the request type for the Query/ContractSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.ContractSpecificationsAllResponse"></a>

### ContractSpecificationsAllResponse
ContractSpecificationsAllResponse is the response type for the Query/ContractSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specifications` | [ContractSpecificationWrapper](#provenance.metadata.v1.ContractSpecificationWrapper) | repeated | contract_specifications are the wrapped contract specifications. |
| `request` | [ContractSpecificationsAllRequest](#provenance.metadata.v1.ContractSpecificationsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.GetByAddrRequest"></a>

### GetByAddrRequest
GetByAddrRequest is the request type for the Query/GetByAddr RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addrs` | [string](#string) | repeated | ids are the metadata addresses of the things to look up. |






<a name="provenance.metadata.v1.GetByAddrResponse"></a>

### GetByAddrResponse
GetByAddrResponse is the response type for the Query/GetByAddr RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scopes` | [Scope](#provenance.metadata.v1.Scope) | repeated | scopes contains any scopes that were requested and found. |
| `sessions` | [Session](#provenance.metadata.v1.Session) | repeated | sessions contains any sessions that were requested and found. |
| `records` | [Record](#provenance.metadata.v1.Record) | repeated | records contains any records that were requested and found. |
| `scope_specs` | [ScopeSpecification](#provenance.metadata.v1.ScopeSpecification) | repeated | scope_specs contains any scope specifications that were requested and found. |
| `contract_specs` | [ContractSpecification](#provenance.metadata.v1.ContractSpecification) | repeated | contract_specs contains any contract specifications that were requested and found. |
| `record_specs` | [RecordSpecification](#provenance.metadata.v1.RecordSpecification) | repeated | record_specs contains any record specifications that were requested and found. |
| `not_found` | [string](#string) | repeated | not_found contains any addrs requested but not found. |






<a name="provenance.metadata.v1.OSAllLocatorsRequest"></a>

### OSAllLocatorsRequest
OSAllLocatorsRequest is the request type for the Query/OSAllLocators RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.OSAllLocatorsResponse"></a>

### OSAllLocatorsResponse
OSAllLocatorsResponse is the response type for the Query/OSAllLocators RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locators` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) | repeated |  |
| `request` | [OSAllLocatorsRequest](#provenance.metadata.v1.OSAllLocatorsRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.OSLocatorParamsRequest"></a>

### OSLocatorParamsRequest
OSLocatorParamsRequest is the request type for the Query/OSLocatorParams RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance.metadata.v1.OSLocatorParamsResponse"></a>

### OSLocatorParamsResponse
OSLocatorParamsResponse is the response type for the Query/OSLocatorParams RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [OSLocatorParams](#provenance.metadata.v1.OSLocatorParams) |  | params defines the parameters of the module. |
| `request` | [OSLocatorParamsRequest](#provenance.metadata.v1.OSLocatorParamsRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.OSLocatorRequest"></a>

### OSLocatorRequest
OSLocatorRequest is the request type for the Query/OSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `owner` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance.metadata.v1.OSLocatorResponse"></a>

### OSLocatorResponse
OSLocatorResponse is the response type for the Query/OSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) |  |  |
| `request` | [OSLocatorRequest](#provenance.metadata.v1.OSLocatorRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.OSLocatorsByScopeRequest"></a>

### OSLocatorsByScopeRequest
OSLocatorsByScopeRequest is the request type for the Query/OSLocatorsByScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance.metadata.v1.OSLocatorsByScopeResponse"></a>

### OSLocatorsByScopeResponse
OSLocatorsByScopeResponse is the response type for the Query/OSLocatorsByScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locators` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) | repeated |  |
| `request` | [OSLocatorsByScopeRequest](#provenance.metadata.v1.OSLocatorsByScopeRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.OSLocatorsByURIRequest"></a>

### OSLocatorsByURIRequest
OSLocatorsByURIRequest is the request type for the Query/OSLocatorsByURI RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `uri` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.OSLocatorsByURIResponse"></a>

### OSLocatorsByURIResponse
OSLocatorsByURIResponse is the response type for the Query/OSLocatorsByURI RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locators` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) | repeated |  |
| `request` | [OSLocatorsByURIRequest](#provenance.metadata.v1.OSLocatorsByURIRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.OwnershipRequest"></a>

### OwnershipRequest
OwnershipRequest is the request type for the Query/Ownership RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.OwnershipResponse"></a>

### OwnershipResponse
OwnershipResponse is the response type for the Query/Ownership RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_uuids` | [string](#string) | repeated | A list of scope ids (uuid) associated with the given address. |
| `request` | [OwnershipRequest](#provenance.metadata.v1.OwnershipRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance.metadata.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.metadata.v1.Params) |  | params defines the parameters of the module. |
| `request` | [QueryParamsRequest](#provenance.metadata.v1.QueryParamsRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.RecordSpecificationRequest"></a>

### RecordSpecificationRequest
RecordSpecificationRequest is the request type for the Query/RecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [string](#string) |  | specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84 or a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn. It can also be a record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. |
| `name` | [string](#string) |  | name is the name of the record to look up. It is required if the specification_id is a uuid or contract specification address. It is ignored if the specification_id is a record specification address. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance.metadata.v1.RecordSpecificationResponse"></a>

### RecordSpecificationResponse
RecordSpecificationResponse is the response type for the Query/RecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specification` | [RecordSpecificationWrapper](#provenance.metadata.v1.RecordSpecificationWrapper) |  | record_specification is the wrapped record specification. |
| `request` | [RecordSpecificationRequest](#provenance.metadata.v1.RecordSpecificationRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.RecordSpecificationWrapper"></a>

### RecordSpecificationWrapper
RecordSpecificationWrapper contains a single record specification and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [RecordSpecification](#provenance.metadata.v1.RecordSpecification) |  | specification is the on-chain record specification message. |
| `record_spec_id_info` | [RecordSpecIdInfo](#provenance.metadata.v1.RecordSpecIdInfo) |  | record_spec_id_info contains information about the id/address of the record specification. |






<a name="provenance.metadata.v1.RecordSpecificationsAllRequest"></a>

### RecordSpecificationsAllRequest
RecordSpecificationsAllRequest is the request type for the Query/RecordSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.RecordSpecificationsAllResponse"></a>

### RecordSpecificationsAllResponse
RecordSpecificationsAllResponse is the response type for the Query/RecordSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specifications` | [RecordSpecificationWrapper](#provenance.metadata.v1.RecordSpecificationWrapper) | repeated | record_specifications are the wrapped record specifications. |
| `request` | [RecordSpecificationsAllRequest](#provenance.metadata.v1.RecordSpecificationsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.RecordSpecificationsForContractSpecificationRequest"></a>

### RecordSpecificationsForContractSpecificationRequest
RecordSpecificationsForContractSpecificationRequest is the request type for the
Query/RecordSpecificationsForContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [string](#string) |  | specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84 or a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn. It can also be a record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance.metadata.v1.RecordSpecificationsForContractSpecificationResponse"></a>

### RecordSpecificationsForContractSpecificationResponse
RecordSpecificationsForContractSpecificationResponse is the response type for the
Query/RecordSpecificationsForContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_specifications` | [RecordSpecificationWrapper](#provenance.metadata.v1.RecordSpecificationWrapper) | repeated | record_specifications is any number of wrapped record specifications associated with this contract_specification. |
| `contract_specification_uuid` | [string](#string) |  | contract_specification_uuid is the uuid of this contract specification. |
| `contract_specification_addr` | [string](#string) |  | contract_specification_addr is the contract specification address as a bech32 encoded string. |
| `request` | [RecordSpecificationsForContractSpecificationRequest](#provenance.metadata.v1.RecordSpecificationsForContractSpecificationRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.RecordWrapper"></a>

### RecordWrapper
RecordWrapper contains a single record and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record` | [Record](#provenance.metadata.v1.Record) |  | record is the on-chain record message. |
| `record_id_info` | [RecordIdInfo](#provenance.metadata.v1.RecordIdInfo) |  | record_id_info contains information about the id/address of the record. |
| `record_spec_id_info` | [RecordSpecIdInfo](#provenance.metadata.v1.RecordSpecIdInfo) |  | record_spec_id_info contains information about the id/address of the record specification. |






<a name="provenance.metadata.v1.RecordsAllRequest"></a>

### RecordsAllRequest
RecordsAllRequest is the request type for the Query/RecordsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.RecordsAllResponse"></a>

### RecordsAllResponse
RecordsAllResponse is the response type for the Query/RecordsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `records` | [RecordWrapper](#provenance.metadata.v1.RecordWrapper) | repeated | records are the wrapped records. |
| `request` | [RecordsAllRequest](#provenance.metadata.v1.RecordsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.RecordsRequest"></a>

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






<a name="provenance.metadata.v1.RecordsResponse"></a>

### RecordsResponse
RecordsResponse is the response type for the Query/Records RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [ScopeWrapper](#provenance.metadata.v1.ScopeWrapper) |  | scope is the wrapped scope that holds these records (if requested). |
| `sessions` | [SessionWrapper](#provenance.metadata.v1.SessionWrapper) | repeated | sessions is any number of wrapped sessions that hold these records (if requested). |
| `records` | [RecordWrapper](#provenance.metadata.v1.RecordWrapper) | repeated | records is any number of wrapped record results. |
| `request` | [RecordsRequest](#provenance.metadata.v1.RecordsRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.ScopeRequest"></a>

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






<a name="provenance.metadata.v1.ScopeResponse"></a>

### ScopeResponse
ScopeResponse is the response type for the Query/Scope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [ScopeWrapper](#provenance.metadata.v1.ScopeWrapper) |  | scope is the wrapped scope result. |
| `sessions` | [SessionWrapper](#provenance.metadata.v1.SessionWrapper) | repeated | sessions is any number of wrapped sessions in this scope (if requested). |
| `records` | [RecordWrapper](#provenance.metadata.v1.RecordWrapper) | repeated | records is any number of wrapped records in this scope (if requested). |
| `request` | [ScopeRequest](#provenance.metadata.v1.ScopeRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.ScopeSpecificationRequest"></a>

### ScopeSpecificationRequest
ScopeSpecificationRequest is the request type for the Query/ScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [string](#string) |  | specification_id can either be a uuid, e.g. dc83ea70-eacd-40fe-9adf-1cf6148bf8a2 or a bech32 scope specification address, e.g. scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m. |
| `include_contract_specs` | [bool](#bool) |  | include_contract_specs is a flag for whether to include the contract specifications of the scope specification in the response. |
| `include_record_specs` | [bool](#bool) |  | include_record_specs is a flag for whether to include the record specifications of the scope specification in the response. |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |






<a name="provenance.metadata.v1.ScopeSpecificationResponse"></a>

### ScopeSpecificationResponse
ScopeSpecificationResponse is the response type for the Query/ScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specification` | [ScopeSpecificationWrapper](#provenance.metadata.v1.ScopeSpecificationWrapper) |  | scope_specification is the wrapped scope specification. |
| `contract_specs` | [ContractSpecificationWrapper](#provenance.metadata.v1.ContractSpecificationWrapper) | repeated | contract_specs is any number of wrapped contract specifications in this scope specification (if requested). |
| `record_specs` | [RecordSpecificationWrapper](#provenance.metadata.v1.RecordSpecificationWrapper) | repeated | record_specs is any number of wrapped record specifications in this scope specification (if requested). |
| `request` | [ScopeSpecificationRequest](#provenance.metadata.v1.ScopeSpecificationRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.ScopeSpecificationWrapper"></a>

### ScopeSpecificationWrapper
ScopeSpecificationWrapper contains a single scope specification and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [ScopeSpecification](#provenance.metadata.v1.ScopeSpecification) |  | specification is the on-chain scope specification message. |
| `scope_spec_id_info` | [ScopeSpecIdInfo](#provenance.metadata.v1.ScopeSpecIdInfo) |  | scope_spec_id_info contains information about the id/address of the scope specification. |






<a name="provenance.metadata.v1.ScopeSpecificationsAllRequest"></a>

### ScopeSpecificationsAllRequest
ScopeSpecificationsAllRequest is the request type for the Query/ScopeSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.ScopeSpecificationsAllResponse"></a>

### ScopeSpecificationsAllResponse
ScopeSpecificationsAllResponse is the response type for the Query/ScopeSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specifications` | [ScopeSpecificationWrapper](#provenance.metadata.v1.ScopeSpecificationWrapper) | repeated | scope_specifications are the wrapped scope specifications. |
| `request` | [ScopeSpecificationsAllRequest](#provenance.metadata.v1.ScopeSpecificationsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.ScopeWrapper"></a>

### ScopeWrapper
SessionWrapper contains a single scope and its uuid.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [Scope](#provenance.metadata.v1.Scope) |  | scope is the on-chain scope message. |
| `scope_id_info` | [ScopeIdInfo](#provenance.metadata.v1.ScopeIdInfo) |  | scope_id_info contains information about the id/address of the scope. |
| `scope_spec_id_info` | [ScopeSpecIdInfo](#provenance.metadata.v1.ScopeSpecIdInfo) |  | scope_spec_id_info contains information about the id/address of the scope specification. |






<a name="provenance.metadata.v1.ScopesAllRequest"></a>

### ScopesAllRequest
ScopesAllRequest is the request type for the Query/ScopesAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.ScopesAllResponse"></a>

### ScopesAllResponse
ScopesAllResponse is the response type for the Query/ScopesAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scopes` | [ScopeWrapper](#provenance.metadata.v1.ScopeWrapper) | repeated | scopes are the wrapped scopes. |
| `request` | [ScopesAllRequest](#provenance.metadata.v1.ScopesAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.SessionWrapper"></a>

### SessionWrapper
SessionWrapper contains a single session and some extra identifiers for it.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session` | [Session](#provenance.metadata.v1.Session) |  | session is the on-chain session message. |
| `session_id_info` | [SessionIdInfo](#provenance.metadata.v1.SessionIdInfo) |  | session_id_info contains information about the id/address of the session. |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance.metadata.v1.ContractSpecIdInfo) |  | contract_spec_id_info contains information about the id/address of the contract specification. |






<a name="provenance.metadata.v1.SessionsAllRequest"></a>

### SessionsAllRequest
SessionsAllRequest is the request type for the Query/SessionsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `exclude_id_info` | [bool](#bool) |  | exclude_id_info is a flag for whether to exclude the id info from the response. |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.SessionsAllResponse"></a>

### SessionsAllResponse
SessionsAllResponse is the response type for the Query/SessionsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sessions` | [SessionWrapper](#provenance.metadata.v1.SessionWrapper) | repeated | sessions are the wrapped sessions. |
| `request` | [SessionsAllRequest](#provenance.metadata.v1.SessionsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.SessionsRequest"></a>

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






<a name="provenance.metadata.v1.SessionsResponse"></a>

### SessionsResponse
SessionsResponse is the response type for the Query/Sessions RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [ScopeWrapper](#provenance.metadata.v1.ScopeWrapper) |  | scope is the wrapped scope that holds these sessions (if requested). |
| `sessions` | [SessionWrapper](#provenance.metadata.v1.SessionWrapper) | repeated | sessions is any number of wrapped session results. |
| `records` | [RecordWrapper](#provenance.metadata.v1.RecordWrapper) | repeated | records is any number of wrapped records contained in these sessions (if requested). |
| `request` | [SessionsRequest](#provenance.metadata.v1.SessionsRequest) |  | request is a copy of the request that generated these results. |






<a name="provenance.metadata.v1.ValueOwnershipRequest"></a>

### ValueOwnershipRequest
ValueOwnershipRequest is the request type for the Query/ValueOwnership RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `include_request` | [bool](#bool) |  | include_request is a flag for whether to include this request in your result. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.ValueOwnershipResponse"></a>

### ValueOwnershipResponse
ValueOwnershipResponse is the response type for the Query/ValueOwnership RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_uuids` | [string](#string) | repeated | A list of scope ids (uuid) associated with the given address. |
| `request` | [ValueOwnershipRequest](#provenance.metadata.v1.ValueOwnershipRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.metadata.v1.Query"></a>

### Query
Query defines the Metadata Query service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#provenance.metadata.v1.QueryParamsRequest) | [QueryParamsResponse](#provenance.metadata.v1.QueryParamsResponse) | Params queries the parameters of x/metadata module. | GET|/provenance/metadata/v1/params|
| `Scope` | [ScopeRequest](#provenance.metadata.v1.ScopeRequest) | [ScopeResponse](#provenance.metadata.v1.ScopeResponse) | Scope searches for a scope.

The scope id, if provided, must either be scope uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. The session addr, if provided, must be a bech32 session address, e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. The record_addr, if provided, must be a bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3.

* If only a scope_id is provided, that scope is returned. * If only a session_addr is provided, the scope containing that session is returned. * If only a record_addr is provided, the scope containing that record is returned. * If more than one of scope_id, session_addr, and record_addr are provided, and they don't refer to the same scope, a bad request is returned.

Providing a session addr or record addr does not limit the sessions and records returned (if requested). Those parameters are only used to find the scope.

By default, sessions and records are not included. Set include_sessions and/or include_records to true to include sessions and/or records. | GET|/provenance/metadata/v1/scope/{scope_id}GET|/provenance/metadata/v1/session/{session_addr}/scopeGET|/provenance/metadata/v1/record/{record_addr}/scope|
| `ScopesAll` | [ScopesAllRequest](#provenance.metadata.v1.ScopesAllRequest) | [ScopesAllResponse](#provenance.metadata.v1.ScopesAllResponse) | ScopesAll retrieves all scopes. | GET|/provenance/metadata/v1/scopes/all|
| `Sessions` | [SessionsRequest](#provenance.metadata.v1.SessionsRequest) | [SessionsResponse](#provenance.metadata.v1.SessionsResponse) | Sessions searches for sessions.

The scope_id can either be scope uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. Similarly, the session_id can either be a uuid or session address, e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. The record_addr, if provided, must be a bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3.

* If only a scope_id is provided, all sessions in that scope are returned. * If only a session_id is provided, it must be an address, and that single session is returned. * If the session_id is a uuid, then either a scope_id or record_addr must also be provided, and that single session is returned. * If only a record_addr is provided, the session containing that record will be returned. * If a record_name is provided then either a scope_id, session_id as an address, or record_addr must also be provided, and the session containing that record will be returned.

A bad request is returned if: * The session_id is a uuid and is provided without a scope_id or record_addr. * A record_name is provided without any way to identify the scope (e.g. a scope_id, a session_id as an address, or a record_addr). * Two or more of scope_id, session_id as an address, and record_addr are provided and don't all refer to the same scope. * A record_addr (or scope_id and record_name) is provided with a session_id and that session does not contain such a record. * A record_addr and record_name are both provided, but reference different records.

By default, the scope and records are not included. Set include_scope and/or include_records to true to include the scope and/or records. | GET|/provenance/metadata/v1/session/{session_id}GET|/provenance/metadata/v1/scope/{scope_id}/sessionsGET|/provenance/metadata/v1/scope/{scope_id}/session/{session_id}GET|/provenance/metadata/v1/record/{record_addr}/sessionGET|/provenance/metadata/v1/scope/{scope_id}/record/{record_name}/session|
| `SessionsAll` | [SessionsAllRequest](#provenance.metadata.v1.SessionsAllRequest) | [SessionsAllResponse](#provenance.metadata.v1.SessionsAllResponse) | SessionsAll retrieves all sessions. | GET|/provenance/metadata/v1/sessions/all|
| `Records` | [RecordsRequest](#provenance.metadata.v1.RecordsRequest) | [RecordsResponse](#provenance.metadata.v1.RecordsResponse) | Records searches for records.

The record_addr, if provided, must be a bech32 record address, e.g. record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3. The scope-id can either be scope uuid, e.g. 91978ba2-5f35-459a-86a7-feca1b0512e0 or a scope address, e.g. scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel. Similarly, the session_id can either be a uuid or session address, e.g. session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr. The name is the name of the record you're interested in.

* If only a record_addr is provided, that single record will be returned. * If only a scope_id is provided, all records in that scope will be returned. * If only a session_id (or scope_id/session_id), all records in that session will be returned. * If a name is provided with a scope_id and/or session_id, that single record will be returned.

A bad request is returned if: * The session_id is a uuid and no scope_id is provided. * There are two or more of record_addr, session_id, and scope_id, and they don't all refer to the same scope. * A name is provided, but not a scope_id and/or a session_id. * A name and record_addr are provided and the name doesn't match the record_addr.

By default, the scope and sessions are not included. Set include_scope and/or include_sessions to true to include the scope and/or sessions. | GET|/provenance/metadata/v1/record/{record_addr}GET|/provenance/metadata/v1/scope/{scope_id}/recordsGET|/provenance/metadata/v1/scope/{scope_id}/record/{name}GET|/provenance/metadata/v1/scope/{scope_id}/session/{session_id}/recordsGET|/provenance/metadata/v1/scope/{scope_id}/session/{session_id}/record/{name}GET|/provenance/metadata/v1/session/{session_id}/recordsGET|/provenance/metadata/v1/session/{session_id}/record/{name}|
| `RecordsAll` | [RecordsAllRequest](#provenance.metadata.v1.RecordsAllRequest) | [RecordsAllResponse](#provenance.metadata.v1.RecordsAllResponse) | RecordsAll retrieves all records. | GET|/provenance/metadata/v1/records/all|
| `Ownership` | [OwnershipRequest](#provenance.metadata.v1.OwnershipRequest) | [OwnershipResponse](#provenance.metadata.v1.OwnershipResponse) | Ownership returns the scope identifiers that list the given address as either a data or value owner. | GET|/provenance/metadata/v1/ownership/{address}|
| `ValueOwnership` | [ValueOwnershipRequest](#provenance.metadata.v1.ValueOwnershipRequest) | [ValueOwnershipResponse](#provenance.metadata.v1.ValueOwnershipResponse) | ValueOwnership returns the scope identifiers that list the given address as the value owner. | GET|/provenance/metadata/v1/valueownership/{address}|
| `ScopeSpecification` | [ScopeSpecificationRequest](#provenance.metadata.v1.ScopeSpecificationRequest) | [ScopeSpecificationResponse](#provenance.metadata.v1.ScopeSpecificationResponse) | ScopeSpecification returns a scope specification for the given specification id.

The specification_id can either be a uuid, e.g. dc83ea70-eacd-40fe-9adf-1cf6148bf8a2 or a bech32 scope specification address, e.g. scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m.

By default, the contract and record specifications are not included. Set include_contract_specs and/or include_record_specs to true to include contract and/or record specifications. | GET|/provenance/metadata/v1/scopespec/{specification_id}|
| `ScopeSpecificationsAll` | [ScopeSpecificationsAllRequest](#provenance.metadata.v1.ScopeSpecificationsAllRequest) | [ScopeSpecificationsAllResponse](#provenance.metadata.v1.ScopeSpecificationsAllResponse) | ScopeSpecificationsAll retrieves all scope specifications. | GET|/provenance/metadata/v1/scopespecs/all|
| `ContractSpecification` | [ContractSpecificationRequest](#provenance.metadata.v1.ContractSpecificationRequest) | [ContractSpecificationResponse](#provenance.metadata.v1.ContractSpecificationResponse) | ContractSpecification returns a contract specification for the given specification id.

The specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84, a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn, or a bech32 record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. If it is a record specification address, then the contract specification that contains that record specification is looked up.

By default, the record specifications for this contract specification are not included. Set include_record_specs to true to include them in the result. | GET|/provenance/metadata/v1/contractspec/{specification_id}|
| `ContractSpecificationsAll` | [ContractSpecificationsAllRequest](#provenance.metadata.v1.ContractSpecificationsAllRequest) | [ContractSpecificationsAllResponse](#provenance.metadata.v1.ContractSpecificationsAllResponse) | ContractSpecificationsAll retrieves all contract specifications. | GET|/provenance/metadata/v1/contractspecs/all|
| `RecordSpecificationsForContractSpecification` | [RecordSpecificationsForContractSpecificationRequest](#provenance.metadata.v1.RecordSpecificationsForContractSpecificationRequest) | [RecordSpecificationsForContractSpecificationResponse](#provenance.metadata.v1.RecordSpecificationsForContractSpecificationResponse) | RecordSpecificationsForContractSpecification returns the record specifications for the given input.

The specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84, a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn, or a bech32 record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. If it is a record specification address, then the contract specification that contains that record specification is used. | GET|/provenance/metadata/v1/contractspec/{specification_id}/recordspecs|
| `RecordSpecification` | [RecordSpecificationRequest](#provenance.metadata.v1.RecordSpecificationRequest) | [RecordSpecificationResponse](#provenance.metadata.v1.RecordSpecificationResponse) | RecordSpecification returns a record specification for the given input. | GET|/provenance/metadata/v1/recordspec/{specification_id}GET|/provenance/metadata/v1/contractspec/{specification_id}/recordspec/{name}|
| `RecordSpecificationsAll` | [RecordSpecificationsAllRequest](#provenance.metadata.v1.RecordSpecificationsAllRequest) | [RecordSpecificationsAllResponse](#provenance.metadata.v1.RecordSpecificationsAllResponse) | RecordSpecificationsAll retrieves all record specifications. | GET|/provenance/metadata/v1/recordspecs/all|
| `GetByAddr` | [GetByAddrRequest](#provenance.metadata.v1.GetByAddrRequest) | [GetByAddrResponse](#provenance.metadata.v1.GetByAddrResponse) | GetByAddr retrieves metadata given any address(es). | GET|/provenance/metadata/v1/addr/{addrs}|
| `OSLocatorParams` | [OSLocatorParamsRequest](#provenance.metadata.v1.OSLocatorParamsRequest) | [OSLocatorParamsResponse](#provenance.metadata.v1.OSLocatorParamsResponse) | OSLocatorParams returns all parameters for the object store locator sub module. | GET|/provenance/metadata/v1/locator/params|
| `OSLocator` | [OSLocatorRequest](#provenance.metadata.v1.OSLocatorRequest) | [OSLocatorResponse](#provenance.metadata.v1.OSLocatorResponse) | OSLocator returns an ObjectStoreLocator by its owner's address. | GET|/provenance/metadata/v1/locator/{owner}|
| `OSLocatorsByURI` | [OSLocatorsByURIRequest](#provenance.metadata.v1.OSLocatorsByURIRequest) | [OSLocatorsByURIResponse](#provenance.metadata.v1.OSLocatorsByURIResponse) | OSLocatorsByURI returns all ObjectStoreLocator entries for a locator uri. | GET|/provenance/metadata/v1/locator/uri/{uri}|
| `OSLocatorsByScope` | [OSLocatorsByScopeRequest](#provenance.metadata.v1.OSLocatorsByScopeRequest) | [OSLocatorsByScopeResponse](#provenance.metadata.v1.OSLocatorsByScopeResponse) | OSLocatorsByScope returns all ObjectStoreLocator entries for a for all signer's present in the specified scope. | GET|/provenance/metadata/v1/locator/scope/{scope_id}|
| `OSAllLocators` | [OSAllLocatorsRequest](#provenance.metadata.v1.OSAllLocatorsRequest) | [OSAllLocatorsResponse](#provenance.metadata.v1.OSAllLocatorsResponse) | OSAllLocators returns all ObjectStoreLocator entries. | GET|/provenance/metadata/v1/locators/all|
| `AccountData` | [AccountDataRequest](#provenance.metadata.v1.AccountDataRequest) | [AccountDataResponse](#provenance.metadata.v1.AccountDataResponse) | AccountData gets the account data associated with a metadata address. Currently, only scope ids are supported. | GET|/provenance/metadata/v1/accountdata/{metadata_addr}|

 <!-- end services -->



<a name="provenance/metadata/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/tx.proto



<a name="provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest"></a>

### MsgAddContractSpecToScopeSpecRequest
MsgAddContractSpecToScopeSpecRequest is the request type for the Msg/AddContractSpecToScopeSpec RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_id` | [bytes](#bytes) |  | MetadataAddress for the contract specification to add. |
| `scope_specification_id` | [bytes](#bytes) |  | MetadataAddress for the scope specification to add contract specification to. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance.metadata.v1.MsgAddContractSpecToScopeSpecResponse"></a>

### MsgAddContractSpecToScopeSpecResponse
MsgAddContractSpecToScopeSpecResponse is the response type for the Msg/AddContractSpecToScopeSpec RPC method.






<a name="provenance.metadata.v1.MsgAddScopeDataAccessRequest"></a>

### MsgAddScopeDataAccessRequest
MsgAddScopeDataAccessRequest is the request to add data access AccAddress to scope


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope MetadataAddress for updating data access |
| `data_access` | [string](#string) | repeated | AccAddress addresses to be added to scope |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |






<a name="provenance.metadata.v1.MsgAddScopeDataAccessResponse"></a>

### MsgAddScopeDataAccessResponse
MsgAddScopeDataAccessResponse is the response for adding data access AccAddress to scope






<a name="provenance.metadata.v1.MsgAddScopeOwnerRequest"></a>

### MsgAddScopeOwnerRequest
MsgAddScopeOwnerRequest is the request to add owner AccAddress to scope


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope MetadataAddress for updating data access |
| `owners` | [Party](#provenance.metadata.v1.Party) | repeated | owner parties to add to the scope |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |






<a name="provenance.metadata.v1.MsgAddScopeOwnerResponse"></a>

### MsgAddScopeOwnerResponse
MsgAddScopeOwnerResponse is the response for adding owner AccAddresses to scope






<a name="provenance.metadata.v1.MsgBindOSLocatorRequest"></a>

### MsgBindOSLocatorRequest
MsgBindOSLocatorRequest is the request type for the Msg/BindOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) |  | The object locator to bind the address to bind to the URI. |






<a name="provenance.metadata.v1.MsgBindOSLocatorResponse"></a>

### MsgBindOSLocatorResponse
MsgBindOSLocatorResponse is the response type for the Msg/BindOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) |  |  |






<a name="provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecRequest"></a>

### MsgDeleteContractSpecFromScopeSpecRequest
MsgDeleteContractSpecFromScopeSpecRequest is the request type for the Msg/DeleteContractSpecFromScopeSpec RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specification_id` | [bytes](#bytes) |  | MetadataAddress for the contract specification to add. |
| `scope_specification_id` | [bytes](#bytes) |  | MetadataAddress for the scope specification to add contract specification to. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecResponse"></a>

### MsgDeleteContractSpecFromScopeSpecResponse
MsgDeleteContractSpecFromScopeSpecResponse is the response type for the Msg/DeleteContractSpecFromScopeSpec RPC
method.






<a name="provenance.metadata.v1.MsgDeleteContractSpecificationRequest"></a>

### MsgDeleteContractSpecificationRequest
MsgDeleteContractSpecificationRequest is the request type for the Msg/DeleteContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | MetadataAddress for the contract specification to delete. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance.metadata.v1.MsgDeleteContractSpecificationResponse"></a>

### MsgDeleteContractSpecificationResponse
MsgDeleteContractSpecificationResponse is the response type for the Msg/DeleteContractSpecification RPC method.






<a name="provenance.metadata.v1.MsgDeleteOSLocatorRequest"></a>

### MsgDeleteOSLocatorRequest
MsgDeleteOSLocatorRequest is the request type for the Msg/DeleteOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) |  | The record being removed |






<a name="provenance.metadata.v1.MsgDeleteOSLocatorResponse"></a>

### MsgDeleteOSLocatorResponse
MsgDeleteOSLocatorResponse is the response type for the Msg/DeleteOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) |  |  |






<a name="provenance.metadata.v1.MsgDeleteRecordRequest"></a>

### MsgDeleteRecordRequest
MsgDeleteRecordRequest is the request type for the Msg/DeleteRecord RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_id` | [bytes](#bytes) |  |  |
| `signers` | [string](#string) | repeated |  |






<a name="provenance.metadata.v1.MsgDeleteRecordResponse"></a>

### MsgDeleteRecordResponse
MsgDeleteRecordResponse is the response type for the Msg/DeleteRecord RPC method.






<a name="provenance.metadata.v1.MsgDeleteRecordSpecificationRequest"></a>

### MsgDeleteRecordSpecificationRequest
MsgDeleteRecordSpecificationRequest is the request type for the Msg/DeleteRecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | MetadataAddress for the record specification to delete. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance.metadata.v1.MsgDeleteRecordSpecificationResponse"></a>

### MsgDeleteRecordSpecificationResponse
MsgDeleteRecordSpecificationResponse is the response type for the Msg/DeleteRecordSpecification RPC method.






<a name="provenance.metadata.v1.MsgDeleteScopeDataAccessRequest"></a>

### MsgDeleteScopeDataAccessRequest
MsgDeleteScopeDataAccessRequest is the request to remove data access AccAddress to scope


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope MetadataAddress for removing data access |
| `data_access` | [string](#string) | repeated | AccAddress address to be removed from scope |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |






<a name="provenance.metadata.v1.MsgDeleteScopeDataAccessResponse"></a>

### MsgDeleteScopeDataAccessResponse
MsgDeleteScopeDataAccessResponse is the response from removing data access AccAddress to scope






<a name="provenance.metadata.v1.MsgDeleteScopeOwnerRequest"></a>

### MsgDeleteScopeOwnerRequest
MsgDeleteScopeOwnerRequest is the request to remove owner AccAddresses to scope


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | scope MetadataAddress for removing data access |
| `owners` | [string](#string) | repeated | AccAddress owner addresses to be removed from scope |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |






<a name="provenance.metadata.v1.MsgDeleteScopeOwnerResponse"></a>

### MsgDeleteScopeOwnerResponse
MsgDeleteScopeOwnerResponse is the response from removing owner AccAddress to scope






<a name="provenance.metadata.v1.MsgDeleteScopeRequest"></a>

### MsgDeleteScopeRequest
MsgDeleteScopeRequest is the request type for the Msg/DeleteScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [bytes](#bytes) |  | Unique ID for the scope to delete |
| `signers` | [string](#string) | repeated |  |






<a name="provenance.metadata.v1.MsgDeleteScopeResponse"></a>

### MsgDeleteScopeResponse
MsgDeleteScopeResponse is the response type for the Msg/DeleteScope RPC method.






<a name="provenance.metadata.v1.MsgDeleteScopeSpecificationRequest"></a>

### MsgDeleteScopeSpecificationRequest
MsgDeleteScopeSpecificationRequest is the request type for the Msg/DeleteScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [bytes](#bytes) |  | MetadataAddress for the scope specification to delete. |
| `signers` | [string](#string) | repeated |  |






<a name="provenance.metadata.v1.MsgDeleteScopeSpecificationResponse"></a>

### MsgDeleteScopeSpecificationResponse
MsgDeleteScopeSpecificationResponse is the response type for the Msg/DeleteScopeSpecification RPC method.






<a name="provenance.metadata.v1.MsgMigrateValueOwnerRequest"></a>

### MsgMigrateValueOwnerRequest
MsgMigrateValueOwnerRequest is the request to migrate all scopes with one value owner to another value owner.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `existing` | [string](#string) |  | existing is the value owner address that is being migrated. |
| `proposed` | [string](#string) |  | proposed is the new value owner address for all of existing's scopes. |
| `signers` | [string](#string) | repeated | signers is the list of addresses of those signing this request. |






<a name="provenance.metadata.v1.MsgMigrateValueOwnerResponse"></a>

### MsgMigrateValueOwnerResponse
MsgMigrateValueOwnerResponse is the response from migrating a value owner address.






<a name="provenance.metadata.v1.MsgModifyOSLocatorRequest"></a>

### MsgModifyOSLocatorRequest
MsgModifyOSLocatorRequest is the request type for the Msg/ModifyOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) |  | The object locator to bind the address to bind to the URI. |






<a name="provenance.metadata.v1.MsgModifyOSLocatorResponse"></a>

### MsgModifyOSLocatorResponse
MsgModifyOSLocatorResponse is the response type for the Msg/ModifyOSLocator RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locator` | [ObjectStoreLocator](#provenance.metadata.v1.ObjectStoreLocator) |  |  |






<a name="provenance.metadata.v1.MsgP8eMemorializeContractRequest"></a>

### MsgP8eMemorializeContractRequest
MsgP8eMemorializeContractRequest  has been deprecated and is no longer usable.
Deprecated: This message is no longer part of any endpoint and cannot be used for anything.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [string](#string) |  |  |
| `group_id` | [string](#string) |  |  |
| `scope_specification_id` | [string](#string) |  |  |
| `recitals` | [p8e.Recitals](#provenance.metadata.v1.p8e.Recitals) |  |  |
| `contract` | [p8e.Contract](#provenance.metadata.v1.p8e.Contract) |  |  |
| `signatures` | [p8e.SignatureSet](#provenance.metadata.v1.p8e.SignatureSet) |  |  |
| `invoker` | [string](#string) |  |  |






<a name="provenance.metadata.v1.MsgP8eMemorializeContractResponse"></a>

### MsgP8eMemorializeContractResponse
MsgP8eMemorializeContractResponse  has been deprecated and is no longer usable.
Deprecated: This message is no longer part of any endpoint and cannot be used for anything.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id_info` | [ScopeIdInfo](#provenance.metadata.v1.ScopeIdInfo) |  |  |
| `session_id_info` | [SessionIdInfo](#provenance.metadata.v1.SessionIdInfo) |  |  |
| `record_id_infos` | [RecordIdInfo](#provenance.metadata.v1.RecordIdInfo) | repeated |  |






<a name="provenance.metadata.v1.MsgSetAccountDataRequest"></a>

### MsgSetAccountDataRequest
MsgSetAccountDataRequest is the request to set/update/delete a scope's account data.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `metadata_addr` | [bytes](#bytes) |  | The identifier to associate the data with. Currently, only scope ids are supported. |
| `value` | [string](#string) |  | The desired accountdata value. |
| `signers` | [string](#string) | repeated | The signers of this message. Must fulfill owner requirements of the scope. |






<a name="provenance.metadata.v1.MsgSetAccountDataResponse"></a>

### MsgSetAccountDataResponse
MsgSetAccountDataResponse is the response from setting/updating/deleting a scope's account data.






<a name="provenance.metadata.v1.MsgUpdateValueOwnersRequest"></a>

### MsgUpdateValueOwnersRequest
MsgUpdateValueOwnersRequest is the request to update the value owner addresses in one or more scopes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_ids` | [bytes](#bytes) | repeated | scope_ids are the scope metadata addresses of all scopes to be updated. |
| `value_owner_address` | [string](#string) |  | value_owner_address is the address of the new value owner for the provided scopes. |
| `signers` | [string](#string) | repeated | signers is the list of addresses of those signing this request. |






<a name="provenance.metadata.v1.MsgUpdateValueOwnersResponse"></a>

### MsgUpdateValueOwnersResponse
MsgUpdateValueOwnersResponse is the response from updating value owner addresses in one or more scopes.






<a name="provenance.metadata.v1.MsgWriteContractSpecificationRequest"></a>

### MsgWriteContractSpecificationRequest
MsgWriteContractSpecificationRequest is the request type for the Msg/WriteContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [ContractSpecification](#provenance.metadata.v1.ContractSpecification) |  | specification is the ContractSpecification you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `spec_uuid` | [string](#string) |  | spec_uuid is an optional contract specification uuid string, e.g. "def6bc0a-c9dd-4874-948f-5206e6060a84" If provided, it will be used to generate the MetadataAddress for the contract specification which will override the specification_id in the provided specification. If not provided (or it is an empty string), nothing special happens. If there is a value in specification.specification_id that is different from the one created from this uuid, an error is returned. |






<a name="provenance.metadata.v1.MsgWriteContractSpecificationResponse"></a>

### MsgWriteContractSpecificationResponse
MsgWriteContractSpecificationResponse is the response type for the Msg/WriteContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance.metadata.v1.ContractSpecIdInfo) |  | contract_spec_id_info contains information about the id/address of the contract specification that was added or updated. |






<a name="provenance.metadata.v1.MsgWriteP8eContractSpecRequest"></a>

### MsgWriteP8eContractSpecRequest
MsgWriteP8eContractSpecRequest has been deprecated and is no longer usable.
Deprecated: This message is no longer part of any endpoint and cannot be used for anything.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contractspec` | [p8e.ContractSpec](#provenance.metadata.v1.p8e.ContractSpec) |  |  |
| `signers` | [string](#string) | repeated |  |






<a name="provenance.metadata.v1.MsgWriteP8eContractSpecResponse"></a>

### MsgWriteP8eContractSpecResponse
MsgWriteP8eContractSpecResponse  has been deprecated and is no longer usable.
Deprecated: This message is no longer part of any endpoint and cannot be used for anything.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance.metadata.v1.ContractSpecIdInfo) |  |  |
| `record_spec_id_infos` | [RecordSpecIdInfo](#provenance.metadata.v1.RecordSpecIdInfo) | repeated |  |






<a name="provenance.metadata.v1.MsgWriteRecordRequest"></a>

### MsgWriteRecordRequest
MsgWriteRecordRequest is the request type for the Msg/WriteRecord RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record` | [Record](#provenance.metadata.v1.Record) |  | record is the Record you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `session_id_components` | [SessionIdComponents](#provenance.metadata.v1.SessionIdComponents) |  | SessionIDComponents is an optional (alternate) way of defining what the session_id should be in the provided record. If provided, it must have both a scope and session_uuid. Those components will be used to create the MetadataAddress for the session which will override the session_id in the provided record. If not provided (or all empty), nothing special happens. If there is a value in record.session_id that is different from the one created from these components, an error is returned. |
| `contract_spec_uuid` | [string](#string) |  | contract_spec_uuid is an optional contract specification uuid string, e.g. "def6bc0a-c9dd-4874-948f-5206e6060a84" If provided, it will be combined with the record name to generate the MetadataAddress for the record specification which will override the specification_id in the provided record. If not provided (or it is an empty string), nothing special happens. If there is a value in record.specification_id that is different from the one created from this uuid and record.name, an error is returned. |
| `parties` | [Party](#provenance.metadata.v1.Party) | repeated | parties is the list of parties involved with this record. Deprecated: This field is ignored. The parties are identified in the session and as signers. |






<a name="provenance.metadata.v1.MsgWriteRecordResponse"></a>

### MsgWriteRecordResponse
MsgWriteRecordResponse is the response type for the Msg/WriteRecord RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_id_info` | [RecordIdInfo](#provenance.metadata.v1.RecordIdInfo) |  | record_id_info contains information about the id/address of the record that was added or updated. |






<a name="provenance.metadata.v1.MsgWriteRecordSpecificationRequest"></a>

### MsgWriteRecordSpecificationRequest
MsgWriteRecordSpecificationRequest is the request type for the Msg/WriteRecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [RecordSpecification](#provenance.metadata.v1.RecordSpecification) |  | specification is the RecordSpecification you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `contract_spec_uuid` | [string](#string) |  | contract_spec_uuid is an optional contract specification uuid string, e.g. "def6bc0a-c9dd-4874-948f-5206e6060a84" If provided, it will be combined with the record specification name to generate the MetadataAddress for the record specification which will override the specification_id in the provided specification. If not provided (or it is an empty string), nothing special happens. If there is a value in specification.specification_id that is different from the one created from this uuid and specification.name, an error is returned. |






<a name="provenance.metadata.v1.MsgWriteRecordSpecificationResponse"></a>

### MsgWriteRecordSpecificationResponse
MsgWriteRecordSpecificationResponse is the response type for the Msg/WriteRecordSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record_spec_id_info` | [RecordSpecIdInfo](#provenance.metadata.v1.RecordSpecIdInfo) |  | record_spec_id_info contains information about the id/address of the record specification that was added or updated. |






<a name="provenance.metadata.v1.MsgWriteScopeRequest"></a>

### MsgWriteScopeRequest
MsgWriteScopeRequest is the request type for the Msg/WriteScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope` | [Scope](#provenance.metadata.v1.Scope) |  | scope is the Scope you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `scope_uuid` | [string](#string) |  | scope_uuid is an optional uuid string, e.g. "91978ba2-5f35-459a-86a7-feca1b0512e0" If provided, it will be used to generate the MetadataAddress for the scope which will override the scope_id in the provided scope. If not provided (or it is an empty string), nothing special happens. If there is a value in scope.scope_id that is different from the one created from this uuid, an error is returned. |
| `spec_uuid` | [string](#string) |  | spec_uuid is an optional scope specification uuid string, e.g. "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2" If provided, it will be used to generate the MetadataAddress for the scope specification which will override the specification_id in the provided scope. If not provided (or it is an empty string), nothing special happens. If there is a value in scope.specification_id that is different from the one created from this uuid, an error is returned. |






<a name="provenance.metadata.v1.MsgWriteScopeResponse"></a>

### MsgWriteScopeResponse
MsgWriteScopeResponse is the response type for the Msg/WriteScope RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id_info` | [ScopeIdInfo](#provenance.metadata.v1.ScopeIdInfo) |  | scope_id_info contains information about the id/address of the scope that was added or updated. |






<a name="provenance.metadata.v1.MsgWriteScopeSpecificationRequest"></a>

### MsgWriteScopeSpecificationRequest
MsgWriteScopeSpecificationRequest is the request type for the Msg/WriteScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification` | [ScopeSpecification](#provenance.metadata.v1.ScopeSpecification) |  | specification is the ScopeSpecification you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `spec_uuid` | [string](#string) |  | spec_uuid is an optional scope specification uuid string, e.g. "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2" If provided, it will be used to generate the MetadataAddress for the scope specification which will override the specification_id in the provided specification. If not provided (or it is an empty string), nothing special happens. If there is a value in specification.specification_id that is different from the one created from this uuid, an error is returned. |






<a name="provenance.metadata.v1.MsgWriteScopeSpecificationResponse"></a>

### MsgWriteScopeSpecificationResponse
MsgWriteScopeSpecificationResponse is the response type for the Msg/WriteScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_spec_id_info` | [ScopeSpecIdInfo](#provenance.metadata.v1.ScopeSpecIdInfo) |  | scope_spec_id_info contains information about the id/address of the scope specification that was added or updated. |






<a name="provenance.metadata.v1.MsgWriteSessionRequest"></a>

### MsgWriteSessionRequest
MsgWriteSessionRequest is the request type for the Msg/WriteSession RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session` | [Session](#provenance.metadata.v1.Session) |  | session is the Session you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `session_id_components` | [SessionIdComponents](#provenance.metadata.v1.SessionIdComponents) |  | SessionIDComponents is an optional (alternate) way of defining what the session_id should be in the provided session. If provided, it must have both a scope and session_uuid. Those components will be used to create the MetadataAddress for the session which will override the session_id in the provided session. If not provided (or all empty), nothing special happens. If there is a value in session.session_id that is different from the one created from these components, an error is returned. |
| `spec_uuid` | [string](#string) |  | spec_uuid is an optional contract specification uuid string, e.g. "def6bc0a-c9dd-4874-948f-5206e6060a84" If provided, it will be used to generate the MetadataAddress for the contract specification which will override the specification_id in the provided session. If not provided (or it is an empty string), nothing special happens. If there is a value in session.specification_id that is different from the one created from this uuid, an error is returned. |






<a name="provenance.metadata.v1.MsgWriteSessionResponse"></a>

### MsgWriteSessionResponse
MsgWriteSessionResponse is the response type for the Msg/WriteSession RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `session_id_info` | [SessionIdInfo](#provenance.metadata.v1.SessionIdInfo) |  | session_id_info contains information about the id/address of the session that was added or updated. |






<a name="provenance.metadata.v1.SessionIdComponents"></a>

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


<a name="provenance.metadata.v1.Msg"></a>

### Msg
Msg defines the Metadata Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `WriteScope` | [MsgWriteScopeRequest](#provenance.metadata.v1.MsgWriteScopeRequest) | [MsgWriteScopeResponse](#provenance.metadata.v1.MsgWriteScopeResponse) | WriteScope adds or updates a scope. | |
| `DeleteScope` | [MsgDeleteScopeRequest](#provenance.metadata.v1.MsgDeleteScopeRequest) | [MsgDeleteScopeResponse](#provenance.metadata.v1.MsgDeleteScopeResponse) | DeleteScope deletes a scope and all associated Records, Sessions. | |
| `AddScopeDataAccess` | [MsgAddScopeDataAccessRequest](#provenance.metadata.v1.MsgAddScopeDataAccessRequest) | [MsgAddScopeDataAccessResponse](#provenance.metadata.v1.MsgAddScopeDataAccessResponse) | AddScopeDataAccess adds data access AccAddress to scope | |
| `DeleteScopeDataAccess` | [MsgDeleteScopeDataAccessRequest](#provenance.metadata.v1.MsgDeleteScopeDataAccessRequest) | [MsgDeleteScopeDataAccessResponse](#provenance.metadata.v1.MsgDeleteScopeDataAccessResponse) | DeleteScopeDataAccess removes data access AccAddress from scope | |
| `AddScopeOwner` | [MsgAddScopeOwnerRequest](#provenance.metadata.v1.MsgAddScopeOwnerRequest) | [MsgAddScopeOwnerResponse](#provenance.metadata.v1.MsgAddScopeOwnerResponse) | AddScopeOwner adds new owner parties to a scope | |
| `DeleteScopeOwner` | [MsgDeleteScopeOwnerRequest](#provenance.metadata.v1.MsgDeleteScopeOwnerRequest) | [MsgDeleteScopeOwnerResponse](#provenance.metadata.v1.MsgDeleteScopeOwnerResponse) | DeleteScopeOwner removes owner parties (by addresses) from a scope | |
| `UpdateValueOwners` | [MsgUpdateValueOwnersRequest](#provenance.metadata.v1.MsgUpdateValueOwnersRequest) | [MsgUpdateValueOwnersResponse](#provenance.metadata.v1.MsgUpdateValueOwnersResponse) | UpdateValueOwners sets the value owner of one or more scopes. | |
| `MigrateValueOwner` | [MsgMigrateValueOwnerRequest](#provenance.metadata.v1.MsgMigrateValueOwnerRequest) | [MsgMigrateValueOwnerResponse](#provenance.metadata.v1.MsgMigrateValueOwnerResponse) | MigrateValueOwner updates all scopes that have one value owner to have a another value owner. | |
| `WriteSession` | [MsgWriteSessionRequest](#provenance.metadata.v1.MsgWriteSessionRequest) | [MsgWriteSessionResponse](#provenance.metadata.v1.MsgWriteSessionResponse) | WriteSession adds or updates a session context. | |
| `WriteRecord` | [MsgWriteRecordRequest](#provenance.metadata.v1.MsgWriteRecordRequest) | [MsgWriteRecordResponse](#provenance.metadata.v1.MsgWriteRecordResponse) | WriteRecord adds or updates a record. | |
| `DeleteRecord` | [MsgDeleteRecordRequest](#provenance.metadata.v1.MsgDeleteRecordRequest) | [MsgDeleteRecordResponse](#provenance.metadata.v1.MsgDeleteRecordResponse) | DeleteRecord deletes a record. | |
| `WriteScopeSpecification` | [MsgWriteScopeSpecificationRequest](#provenance.metadata.v1.MsgWriteScopeSpecificationRequest) | [MsgWriteScopeSpecificationResponse](#provenance.metadata.v1.MsgWriteScopeSpecificationResponse) | WriteScopeSpecification adds or updates a scope specification. | |
| `DeleteScopeSpecification` | [MsgDeleteScopeSpecificationRequest](#provenance.metadata.v1.MsgDeleteScopeSpecificationRequest) | [MsgDeleteScopeSpecificationResponse](#provenance.metadata.v1.MsgDeleteScopeSpecificationResponse) | DeleteScopeSpecification deletes a scope specification. | |
| `WriteContractSpecification` | [MsgWriteContractSpecificationRequest](#provenance.metadata.v1.MsgWriteContractSpecificationRequest) | [MsgWriteContractSpecificationResponse](#provenance.metadata.v1.MsgWriteContractSpecificationResponse) | WriteContractSpecification adds or updates a contract specification. | |
| `DeleteContractSpecification` | [MsgDeleteContractSpecificationRequest](#provenance.metadata.v1.MsgDeleteContractSpecificationRequest) | [MsgDeleteContractSpecificationResponse](#provenance.metadata.v1.MsgDeleteContractSpecificationResponse) | DeleteContractSpecification deletes a contract specification. | |
| `AddContractSpecToScopeSpec` | [MsgAddContractSpecToScopeSpecRequest](#provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest) | [MsgAddContractSpecToScopeSpecResponse](#provenance.metadata.v1.MsgAddContractSpecToScopeSpecResponse) | AddContractSpecToScopeSpec adds contract specification to a scope specification. | |
| `DeleteContractSpecFromScopeSpec` | [MsgDeleteContractSpecFromScopeSpecRequest](#provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecRequest) | [MsgDeleteContractSpecFromScopeSpecResponse](#provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecResponse) | DeleteContractSpecFromScopeSpec deletes a contract specification from a scope specification. | |
| `WriteRecordSpecification` | [MsgWriteRecordSpecificationRequest](#provenance.metadata.v1.MsgWriteRecordSpecificationRequest) | [MsgWriteRecordSpecificationResponse](#provenance.metadata.v1.MsgWriteRecordSpecificationResponse) | WriteRecordSpecification adds or updates a record specification. | |
| `DeleteRecordSpecification` | [MsgDeleteRecordSpecificationRequest](#provenance.metadata.v1.MsgDeleteRecordSpecificationRequest) | [MsgDeleteRecordSpecificationResponse](#provenance.metadata.v1.MsgDeleteRecordSpecificationResponse) | DeleteRecordSpecification deletes a record specification. | |
| `BindOSLocator` | [MsgBindOSLocatorRequest](#provenance.metadata.v1.MsgBindOSLocatorRequest) | [MsgBindOSLocatorResponse](#provenance.metadata.v1.MsgBindOSLocatorResponse) | BindOSLocator binds an owner address to a uri. | |
| `DeleteOSLocator` | [MsgDeleteOSLocatorRequest](#provenance.metadata.v1.MsgDeleteOSLocatorRequest) | [MsgDeleteOSLocatorResponse](#provenance.metadata.v1.MsgDeleteOSLocatorResponse) | DeleteOSLocator deletes an existing ObjectStoreLocator record. | |
| `ModifyOSLocator` | [MsgModifyOSLocatorRequest](#provenance.metadata.v1.MsgModifyOSLocatorRequest) | [MsgModifyOSLocatorResponse](#provenance.metadata.v1.MsgModifyOSLocatorResponse) | ModifyOSLocator updates an ObjectStoreLocator record by the current owner. | |
| `SetAccountData` | [MsgSetAccountDataRequest](#provenance.metadata.v1.MsgSetAccountDataRequest) | [MsgSetAccountDataResponse](#provenance.metadata.v1.MsgSetAccountDataResponse) | SetAccountData associates some basic data with a metadata address. Currently, only scope ids are supported. | |

 <!-- end services -->



<a name="provenance/msgfees/v1/msgfees.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/msgfees.proto



<a name="provenance.msgfees.v1.EventMsgFee"></a>

### EventMsgFee
EventMsgFee final event property for msg fee on type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type` | [string](#string) |  |  |
| `count` | [string](#string) |  |  |
| `total` | [string](#string) |  |  |
| `recipient` | [string](#string) |  |  |






<a name="provenance.msgfees.v1.EventMsgFees"></a>

### EventMsgFees
EventMsgFees event emitted with summary of msg fees


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_fees` | [EventMsgFee](#provenance.msgfees.v1.EventMsgFee) | repeated |  |






<a name="provenance.msgfees.v1.MsgFee"></a>

### MsgFee
MsgFee is the core of what gets stored on the blockchain
it consists of four parts
1. the msg type url, i.e. /cosmos.bank.v1beta1.MsgSend
2. minimum additional fees(can be of any denom)
3. optional recipient of fee based on `recipient_basis_points`
4. if recipient is declared they will recieve the basis points of the fee (0-10,000)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type_url` | [string](#string) |  |  |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | additional_fee can pay in any Coin( basically a Denom and Amount, Amount can be zero) |
| `recipient` | [string](#string) |  | optional recipient address, the amount is split between recipient and fee module |
| `recipient_basis_points` | [uint32](#uint32) |  | optional split of funds between the recipient and fee module defaults to 50:50, |






<a name="provenance.msgfees.v1.Params"></a>

### Params
Params defines the set of params for the msgfees module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `floor_gas_price` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | constant used to calculate fees when gas fees shares denom with msg fee |
| `nhash_per_usd_mil` | [uint64](#uint64) |  | total nhash per usd mil for converting usd to nhash |
| `conversion_fee_denom` | [string](#string) |  | conversion fee denom is the denom usd is converted to |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/msgfees/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/genesis.proto



<a name="provenance.msgfees.v1.GenesisState"></a>

### GenesisState
GenesisState contains a set of msg fees, persisted from the store


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.msgfees.v1.Params) |  | params defines all the parameters of the module. |
| `msg_fees` | [MsgFee](#provenance.msgfees.v1.MsgFee) | repeated | msg_based_fees are the additional fees on specific tx msgs |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/msgfees/v1/proposals.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/proposals.proto



<a name="provenance.msgfees.v1.AddMsgFeeProposal"></a>

### AddMsgFeeProposal
AddMsgFeeProposal defines a governance proposal to add additional msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | propsal title |
| `description` | [string](#string) |  | propsal description |
| `msg_type_url` | [string](#string) |  | type url of msg to add fee |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | additional fee for msg type |
| `recipient` | [string](#string) |  | optional recipient to recieve basis points |
| `recipient_basis_points` | [string](#string) |  | basis points to use when recipient is present (1 - 10,000) |






<a name="provenance.msgfees.v1.RemoveMsgFeeProposal"></a>

### RemoveMsgFeeProposal
RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | propsal title |
| `description` | [string](#string) |  | propsal description |
| `msg_type_url` | [string](#string) |  | type url of msg fee to remove |






<a name="provenance.msgfees.v1.UpdateConversionFeeDenomProposal"></a>

### UpdateConversionFeeDenomProposal
UpdateConversionFeeDenomProposal defines a governance proposal to update the msg fee conversion denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | proposal title |
| `description` | [string](#string) |  | proposal description |
| `conversion_fee_denom` | [string](#string) |  | conversion_fee_denom is the denom that usd will be converted to |






<a name="provenance.msgfees.v1.UpdateMsgFeeProposal"></a>

### UpdateMsgFeeProposal
UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | propsal title |
| `description` | [string](#string) |  | propsal description |
| `msg_type_url` | [string](#string) |  | type url of msg to update fee |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | additional fee for msg type |
| `recipient` | [string](#string) |  | optional recipient to recieve basis points |
| `recipient_basis_points` | [string](#string) |  | basis points to use when recipient is present (1 - 10,000) |






<a name="provenance.msgfees.v1.UpdateNhashPerUsdMilProposal"></a>

### UpdateNhashPerUsdMilProposal
UpdateNhashPerUsdMilProposal defines a governance proposal to update the nhash per usd mil param


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | proposal title |
| `description` | [string](#string) |  | proposal description |
| `nhash_per_usd_mil` | [uint64](#uint64) |  | nhash_per_usd_mil is number of nhash per usd mil |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/msgfees/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/query.proto



<a name="provenance.msgfees.v1.CalculateTxFeesRequest"></a>

### CalculateTxFeesRequest
CalculateTxFeesRequest is the request type for the Query RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `tx_bytes` | [bytes](#bytes) |  | tx_bytes is the transaction to simulate. |
| `default_base_denom` | [string](#string) |  | default_base_denom is used to set the denom used for gas fees if not set it will default to nhash. |
| `gas_adjustment` | [float](#float) |  | gas_adjustment is the adjustment factor to be multiplied against the estimate returned by the tx simulation |






<a name="provenance.msgfees.v1.CalculateTxFeesResponse"></a>

### CalculateTxFeesResponse
CalculateTxFeesResponse is the response type for the Query RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `additional_fees` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | additional_fees are the amount of coins to be for addition msg fees |
| `total_fees` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | total_fees are the total amount of fees needed for the transactions (msg fees + gas fee) note: the gas fee is calculated with the floor gas price module param. |
| `estimated_gas` | [uint64](#uint64) |  | estimated_gas is the amount of gas needed for the transaction |






<a name="provenance.msgfees.v1.QueryAllMsgFeesRequest"></a>

### QueryAllMsgFeesRequest
QueryAllMsgFeesRequest queries all Msg which have fees associated with them.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.msgfees.v1.QueryAllMsgFeesResponse"></a>

### QueryAllMsgFeesResponse
response for querying all msg's with fees associated with them


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_fees` | [MsgFee](#provenance.msgfees.v1.MsgFee) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the request. |






<a name="provenance.msgfees.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance.msgfees.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.msgfees.v1.Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.msgfees.v1.Query"></a>

### Query
Query defines the gRPC querier service for marker module.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#provenance.msgfees.v1.QueryParamsRequest) | [QueryParamsResponse](#provenance.msgfees.v1.QueryParamsResponse) | Params queries the parameters for x/msgfees | GET|/provenance/msgfees/v1/params|
| `QueryAllMsgFees` | [QueryAllMsgFeesRequest](#provenance.msgfees.v1.QueryAllMsgFeesRequest) | [QueryAllMsgFeesResponse](#provenance.msgfees.v1.QueryAllMsgFeesResponse) | Query all Msgs which have fees associated with them. | GET|/provenance/msgfees/v1/all|
| `CalculateTxFees` | [CalculateTxFeesRequest](#provenance.msgfees.v1.CalculateTxFeesRequest) | [CalculateTxFeesResponse](#provenance.msgfees.v1.CalculateTxFeesResponse) | CalculateTxFees simulates executing a transaction for estimating gas usage and additional fees. | POST|/provenance/tx/v1/calculate_msg_based_fee|

 <!-- end services -->



<a name="provenance/msgfees/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/msgfees/v1/tx.proto



<a name="provenance.msgfees.v1.MsgAddMsgFeeProposalRequest"></a>

### MsgAddMsgFeeProposalRequest
AddMsgFeeProposal defines a governance proposal to add additional msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type_url` | [string](#string) |  | type url of msg to add fee |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | additional fee for msg type |
| `recipient` | [string](#string) |  | optional recipient to receive basis points |
| `recipient_basis_points` | [string](#string) |  | basis points to use when recipient is present (1 - 10,000) |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance.msgfees.v1.MsgAddMsgFeeProposalResponse"></a>

### MsgAddMsgFeeProposalResponse
MsgAddMsgFeeProposalResponse defines the Msg/AddMsgFeeProposal response type






<a name="provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest"></a>

### MsgAssessCustomMsgFeeRequest
MsgAssessCustomMsgFeeRequest defines an sdk.Msg type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | optional short name for custom msg fee, this will be emitted as a property of the event |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | amount of additional fee that must be paid |
| `recipient` | [string](#string) |  | optional recipient address, the basis points amount is sent to the recipient |
| `from` | [string](#string) |  | the signer of the msg |
| `recipient_basis_points` | [string](#string) |  | optional basis points 0 - 10,000 for recipient defaults to 10,000 |






<a name="provenance.msgfees.v1.MsgAssessCustomMsgFeeResponse"></a>

### MsgAssessCustomMsgFeeResponse
MsgAssessCustomMsgFeeResponse defines the Msg/AssessCustomMsgFeee response type.






<a name="provenance.msgfees.v1.MsgRemoveMsgFeeProposalRequest"></a>

### MsgRemoveMsgFeeProposalRequest
RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type_url` | [string](#string) |  | type url of msg fee to remove |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance.msgfees.v1.MsgRemoveMsgFeeProposalResponse"></a>

### MsgRemoveMsgFeeProposalResponse
MsgRemoveMsgFeeProposalResponse defines the Msg/RemoveMsgFeeProposal response type






<a name="provenance.msgfees.v1.MsgUpdateConversionFeeDenomProposalRequest"></a>

### MsgUpdateConversionFeeDenomProposalRequest
UpdateConversionFeeDenomProposal defines a governance proposal to update the msg fee conversion denom


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `conversion_fee_denom` | [string](#string) |  | conversion_fee_denom is the denom that usd will be converted to |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance.msgfees.v1.MsgUpdateConversionFeeDenomProposalResponse"></a>

### MsgUpdateConversionFeeDenomProposalResponse
MsgUpdateConversionFeeDenomProposalResponse defines the Msg/UpdateConversionFeeDenomProposal response type






<a name="provenance.msgfees.v1.MsgUpdateMsgFeeProposalRequest"></a>

### MsgUpdateMsgFeeProposalRequest
UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type_url` | [string](#string) |  | type url of msg to update fee |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | additional fee for msg type |
| `recipient` | [string](#string) |  | optional recipient to recieve basis points |
| `recipient_basis_points` | [string](#string) |  | basis points to use when recipient is present (1 - 10,000) |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance.msgfees.v1.MsgUpdateMsgFeeProposalResponse"></a>

### MsgUpdateMsgFeeProposalResponse
MsgUpdateMsgFeeProposalResponse defines the Msg/RemoveMsgFeeProposal response type






<a name="provenance.msgfees.v1.MsgUpdateNhashPerUsdMilProposalRequest"></a>

### MsgUpdateNhashPerUsdMilProposalRequest
UpdateNhashPerUsdMilProposal defines a governance proposal to update the nhash per usd mil param


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `nhash_per_usd_mil` | [uint64](#uint64) |  | nhash_per_usd_mil is number of nhash per usd mil |
| `authority` | [string](#string) |  | the signing authority for the proposal |






<a name="provenance.msgfees.v1.MsgUpdateNhashPerUsdMilProposalResponse"></a>

### MsgUpdateNhashPerUsdMilProposalResponse
MsgUpdateNhashPerUsdMilProposalResponse defines the Msg/UpdateNhashPerUsdMilProposal response type





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.msgfees.v1.Msg"></a>

### Msg
Msg defines the msgfees Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `AssessCustomMsgFee` | [MsgAssessCustomMsgFeeRequest](#provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest) | [MsgAssessCustomMsgFeeResponse](#provenance.msgfees.v1.MsgAssessCustomMsgFeeResponse) | AssessCustomMsgFee endpoint executes the additional fee charges. This will only emit the event and not persist it to the keeper. Fees are handled with the custom msg fee handlers Use Case: smart contracts will be able to charge additional fees and direct partial funds to specified recipient for executing contracts | |
| `AddMsgFeeProposal` | [MsgAddMsgFeeProposalRequest](#provenance.msgfees.v1.MsgAddMsgFeeProposalRequest) | [MsgAddMsgFeeProposalResponse](#provenance.msgfees.v1.MsgAddMsgFeeProposalResponse) | AddMsgFeeProposal defines a governance proposal to add additional msg based fee | |
| `UpdateMsgFeeProposal` | [MsgUpdateMsgFeeProposalRequest](#provenance.msgfees.v1.MsgUpdateMsgFeeProposalRequest) | [MsgUpdateMsgFeeProposalResponse](#provenance.msgfees.v1.MsgUpdateMsgFeeProposalResponse) | UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee | |
| `RemoveMsgFeeProposal` | [MsgRemoveMsgFeeProposalRequest](#provenance.msgfees.v1.MsgRemoveMsgFeeProposalRequest) | [MsgRemoveMsgFeeProposalResponse](#provenance.msgfees.v1.MsgRemoveMsgFeeProposalResponse) | RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee | |
| `UpdateNhashPerUsdMilProposal` | [MsgUpdateNhashPerUsdMilProposalRequest](#provenance.msgfees.v1.MsgUpdateNhashPerUsdMilProposalRequest) | [MsgUpdateNhashPerUsdMilProposalResponse](#provenance.msgfees.v1.MsgUpdateNhashPerUsdMilProposalResponse) | UpdateNhashPerUsdMilProposal defines a governance proposal to update the nhash per usd mil param | |
| `UpdateConversionFeeDenomProposal` | [MsgUpdateConversionFeeDenomProposalRequest](#provenance.msgfees.v1.MsgUpdateConversionFeeDenomProposalRequest) | [MsgUpdateConversionFeeDenomProposalResponse](#provenance.msgfees.v1.MsgUpdateConversionFeeDenomProposalResponse) | UpdateConversionFeeDenomProposal defines a governance proposal to update the msg fee conversion denom | |

 <!-- end services -->



<a name="provenance/name/v1/name.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/name/v1/name.proto



<a name="provenance.name.v1.CreateRootNameProposal"></a>

### CreateRootNameProposal
CreateRootNameProposal details a proposal to create a new root name
that is controlled by a given owner and optionally restricted to the owner
for the sole creation of sub names.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | proposal title |
| `description` | [string](#string) |  | proposal description |
| `name` | [string](#string) |  | the bound name |
| `owner` | [string](#string) |  | the address the name will resolve to |
| `restricted` | [bool](#bool) |  | a flag that indicates if an owner signature is required to add sub-names |






<a name="provenance.name.v1.EventNameBound"></a>

### EventNameBound
Event emitted when name is bound.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `restricted` | [bool](#bool) |  |  |






<a name="provenance.name.v1.EventNameUnbound"></a>

### EventNameUnbound
Event emitted when name is unbound.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `restricted` | [bool](#bool) |  |  |






<a name="provenance.name.v1.EventNameUpdate"></a>

### EventNameUpdate
Event emitted when name is updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `restricted` | [bool](#bool) |  |  |






<a name="provenance.name.v1.NameRecord"></a>

### NameRecord
NameRecord is a structure used to bind ownership of a name hierarchy to a collection of addresses


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | the bound name |
| `address` | [string](#string) |  | the address the name resolved to |
| `restricted` | [bool](#bool) |  | whether owner signature is required to add sub-names |






<a name="provenance.name.v1.Params"></a>

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



<a name="provenance/name/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/name/v1/genesis.proto



<a name="provenance.name.v1.GenesisState"></a>

### GenesisState
GenesisState defines the name module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.name.v1.Params) |  | params defines all the parameters of the module. |
| `bindings` | [NameRecord](#provenance.name.v1.NameRecord) | repeated | bindings defines all the name records present at genesis |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/name/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/name/v1/query.proto



<a name="provenance.name.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="provenance.name.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#provenance.name.v1.Params) |  | params defines the parameters of the module. |






<a name="provenance.name.v1.QueryResolveRequest"></a>

### QueryResolveRequest
QueryResolveRequest is the request type for the Query/Resolve method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | name to resolve the address for |






<a name="provenance.name.v1.QueryResolveResponse"></a>

### QueryResolveResponse
QueryResolveResponse is the response type for the Query/Resolve method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | a string containing the address the name resolves to |
| `restricted` | [bool](#bool) |  | Whether owner signature is required to add sub-names. |






<a name="provenance.name.v1.QueryReverseLookupRequest"></a>

### QueryReverseLookupRequest
QueryReverseLookupRequest is the request type for the Query/ReverseLookup method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address to find name records for |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.name.v1.QueryReverseLookupResponse"></a>

### QueryReverseLookupResponse
QueryReverseLookupResponse is the response type for the Query/Resolve method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) | repeated | an array of names bound against a given address |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the request. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.name.v1.Query"></a>

### Query
Query defines the gRPC querier service for distribution module.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#provenance.name.v1.QueryParamsRequest) | [QueryParamsResponse](#provenance.name.v1.QueryParamsResponse) | Params queries params of the name module. | GET|/provenance/name/v1/params|
| `Resolve` | [QueryResolveRequest](#provenance.name.v1.QueryResolveRequest) | [QueryResolveResponse](#provenance.name.v1.QueryResolveResponse) | Resolve queries for the address associated with a given name | GET|/provenance/name/v1/resolve/{name}|
| `ReverseLookup` | [QueryReverseLookupRequest](#provenance.name.v1.QueryReverseLookupRequest) | [QueryReverseLookupResponse](#provenance.name.v1.QueryReverseLookupResponse) | ReverseLookup queries for all names bound against a given address | GET|/provenance/name/v1/lookup/{address}|

 <!-- end services -->



<a name="provenance/name/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/name/v1/tx.proto



<a name="provenance.name.v1.MsgBindNameRequest"></a>

### MsgBindNameRequest
MsgBindNameRequest defines an sdk.Msg type that is used to add an address/name binding under an optional parent name.
The record may optionally be restricted to prevent additional names from being added under this one without the
owner signing the request.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `parent` | [NameRecord](#provenance.name.v1.NameRecord) |  | The parent record to bind this name under. |
| `record` | [NameRecord](#provenance.name.v1.NameRecord) |  | The name record to bind under the parent |






<a name="provenance.name.v1.MsgBindNameResponse"></a>

### MsgBindNameResponse
MsgBindNameResponse defines the Msg/BindName response type.






<a name="provenance.name.v1.MsgCreateRootNameRequest"></a>

### MsgCreateRootNameRequest
MsgCreateRootNameRequest defines an sdk.Msg type to create a new root name
that is controlled by a given owner and optionally restricted to the owner
for the sole creation of sub names.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | The signing authority for the request |
| `record` | [NameRecord](#provenance.name.v1.NameRecord) |  | NameRecord is a structure used to bind ownership of a name hierarchy to a collection of addresses |






<a name="provenance.name.v1.MsgCreateRootNameResponse"></a>

### MsgCreateRootNameResponse
MsgCreateRootNameResponse defines Msg/CreateRootName response type.






<a name="provenance.name.v1.MsgDeleteNameRequest"></a>

### MsgDeleteNameRequest
MsgDeleteNameRequest defines an sdk.Msg type that is used to remove an existing address/name binding.  The binding
may not have any child names currently bound for this request to be successful. All associated attributes on account
addresses will be deleted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record` | [NameRecord](#provenance.name.v1.NameRecord) |  | The record being removed |






<a name="provenance.name.v1.MsgDeleteNameResponse"></a>

### MsgDeleteNameResponse
MsgDeleteNameResponse defines the Msg/DeleteName response type.






<a name="provenance.name.v1.MsgModifyNameRequest"></a>

### MsgModifyNameRequest
MsgModifyNameRequest defines a governance method that is used to update an existing address/name binding.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | The address signing the message |
| `record` | [NameRecord](#provenance.name.v1.NameRecord) |  | The record being updated |






<a name="provenance.name.v1.MsgModifyNameResponse"></a>

### MsgModifyNameResponse
MsgModifyNameResponse defines the Msg/ModifyName response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.name.v1.Msg"></a>

### Msg
Msg defines the bank Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `BindName` | [MsgBindNameRequest](#provenance.name.v1.MsgBindNameRequest) | [MsgBindNameResponse](#provenance.name.v1.MsgBindNameResponse) | BindName binds a name to an address under a root name. | |
| `DeleteName` | [MsgDeleteNameRequest](#provenance.name.v1.MsgDeleteNameRequest) | [MsgDeleteNameResponse](#provenance.name.v1.MsgDeleteNameResponse) | DeleteName defines a method to verify a particular invariance. | |
| `ModifyName` | [MsgModifyNameRequest](#provenance.name.v1.MsgModifyNameRequest) | [MsgModifyNameResponse](#provenance.name.v1.MsgModifyNameResponse) | ModifyName defines a method to modify the attributes of an existing name. | |
| `CreateRootName` | [MsgCreateRootNameRequest](#provenance.name.v1.MsgCreateRootNameRequest) | [MsgCreateRootNameResponse](#provenance.name.v1.MsgCreateRootNameResponse) | CreateRootName defines a governance method for creating a root name. | |

 <!-- end services -->



<a name="provenance/reward/v1/reward.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/reward/v1/reward.proto



<a name="provenance.reward.v1.ActionCounter"></a>

### ActionCounter
ActionCounter is a key-value pair that maps action type to the number of times it was performed.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `action_type` | [string](#string) |  | The type of action performed. |
| `number_of_actions` | [uint64](#uint64) |  | The number of times this action has been performed |






<a name="provenance.reward.v1.ActionDelegate"></a>

### ActionDelegate
ActionDelegate represents the delegate action and its required eligibility criteria.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minimum_actions` | [uint64](#uint64) |  | Minimum number of successful delegates. |
| `maximum_actions` | [uint64](#uint64) |  | Maximum number of successful delegates. |
| `minimum_delegation_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Minimum amount that the user must have currently delegated on the validator. |
| `maximum_delegation_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Maximum amount that the user must have currently delegated on the validator. |
| `minimum_active_stake_percentile` | [string](#string) |  | Minimum percentile that can be below the validator's power ranking. |
| `maximum_active_stake_percentile` | [string](#string) |  | Maximum percentile that can be below the validator's power ranking. |






<a name="provenance.reward.v1.ActionTransfer"></a>

### ActionTransfer
ActionTransfer represents the transfer action and its required eligibility criteria.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minimum_actions` | [uint64](#uint64) |  | Minimum number of successful transfers. |
| `maximum_actions` | [uint64](#uint64) |  | Maximum number of successful transfers. |
| `minimum_delegation_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Minimum delegation amount the account must have across all validators, for the transfer action to be counted. |






<a name="provenance.reward.v1.ActionVote"></a>

### ActionVote
ActionVote represents the voting action and its required eligibility criteria.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minimum_actions` | [uint64](#uint64) |  | Minimum number of successful votes. |
| `maximum_actions` | [uint64](#uint64) |  | Maximum number of successful votes. |
| `minimum_delegation_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Minimum delegation amount the account must have across all validators, for the vote action to be counted. |
| `validator_multiplier` | [uint64](#uint64) |  | Positive multiplier that is applied to the shares awarded by the vote action when conditions are met(for now the only condition is the current vote is a validator vote). A value of zero will behave the same as one |






<a name="provenance.reward.v1.ClaimPeriodRewardDistribution"></a>

### ClaimPeriodRewardDistribution
ClaimPeriodRewardDistribution, this is updated at the end of every claim period.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `claim_period_id` | [uint64](#uint64) |  | The claim period id. |
| `reward_program_id` | [uint64](#uint64) |  | The id of the reward program that this reward belongs to. |
| `total_rewards_pool_for_claim_period` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | The sum of all the granted rewards for this claim period. |
| `rewards_pool` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | The final allocated rewards for this claim period. |
| `total_shares` | [int64](#int64) |  | The total number of granted shares for this claim period. |
| `claim_period_ended` | [bool](#bool) |  | A flag representing if the claim period for this reward has ended. |






<a name="provenance.reward.v1.QualifyingAction"></a>

### QualifyingAction
QualifyingAction can be one of many action types.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delegate` | [ActionDelegate](#provenance.reward.v1.ActionDelegate) |  |  |
| `transfer` | [ActionTransfer](#provenance.reward.v1.ActionTransfer) |  |  |
| `vote` | [ActionVote](#provenance.reward.v1.ActionVote) |  |  |






<a name="provenance.reward.v1.QualifyingActions"></a>

### QualifyingActions
QualifyingActions contains a list of QualifyingActions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `qualifying_actions` | [QualifyingAction](#provenance.reward.v1.QualifyingAction) | repeated | The actions that count towards the reward. |






<a name="provenance.reward.v1.RewardAccountState"></a>

### RewardAccountState
RewardAccountState contains state at the claim period level for a specific address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_program_id` | [uint64](#uint64) |  | The id of the reward program that this share belongs to. |
| `claim_period_id` | [uint64](#uint64) |  | The id of the claim period that the share belongs to. |
| `address` | [string](#string) |  | Owner of the reward account state. |
| `action_counter` | [ActionCounter](#provenance.reward.v1.ActionCounter) | repeated | The number of actions performed by this account, mapped by action type. |
| `shares_earned` | [uint64](#uint64) |  | The amount of granted shares for the address in the reward program's claim period. |
| `claim_status` | [RewardAccountState.ClaimStatus](#provenance.reward.v1.RewardAccountState.ClaimStatus) |  | The status of the claim. |






<a name="provenance.reward.v1.RewardProgram"></a>

### RewardProgram
RewardProgram


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | An integer to uniquely identify the reward program. |
| `title` | [string](#string) |  | Name to help identify the Reward Program.(MaxTitleLength=140) |
| `description` | [string](#string) |  | Short summary describing the Reward Program.(MaxDescriptionLength=10000) |
| `distribute_from_address` | [string](#string) |  | address that provides funds for the total reward pool. |
| `total_reward_pool` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | The total amount of funding given to the RewardProgram. |
| `remaining_pool_balance` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | The remaining funds available to distribute after n claim periods have passed. |
| `claimed_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | The total amount of all funds claimed by participants for all past claim periods. |
| `max_reward_by_address` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Maximum reward per claim period per address. |
| `minimum_rollover_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Minimum amount of coins for a program to rollover. |
| `claim_period_seconds` | [uint64](#uint64) |  | Number of seconds that a claim period lasts. |
| `program_start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time that a RewardProgram should start and switch to STARTED state. |
| `expected_program_end_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time that a RewardProgram is expected to end, based on data when it was setup. |
| `program_end_time_max` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time that a RewardProgram MUST end. |
| `claim_period_end_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Used internally to calculate and track the current claim period's ending time. |
| `actual_program_end_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | Time the RewardProgram switched to FINISHED state. Initially set as empty. |
| `claim_periods` | [uint64](#uint64) |  | Number of claim periods this program will run for. |
| `current_claim_period` | [uint64](#uint64) |  | Current claim period of the RewardProgram. Uses 1-based indexing. |
| `max_rollover_claim_periods` | [uint64](#uint64) |  | maximum number of claim periods a reward program can rollover. |
| `state` | [RewardProgram.State](#provenance.reward.v1.RewardProgram.State) |  | Current state of the RewardProgram. |
| `expiration_offset` | [uint64](#uint64) |  | Grace period after a RewardProgram FINISHED. It is the number of seconds until a RewardProgram enters the EXPIRED state. |
| `qualifying_actions` | [QualifyingAction](#provenance.reward.v1.QualifyingAction) | repeated | Actions that count towards the reward. |





 <!-- end messages -->


<a name="provenance.reward.v1.RewardAccountState.ClaimStatus"></a>

### RewardAccountState.ClaimStatus
ClaimStatus is the state a claim is in

| Name | Number | Description |
| ---- | ------ | ----------- |
| CLAIM_STATUS_UNSPECIFIED | 0 | undefined state |
| CLAIM_STATUS_UNCLAIMABLE | 1 | unclaimable status |
| CLAIM_STATUS_CLAIMABLE | 2 | unclaimable claimable |
| CLAIM_STATUS_CLAIMED | 3 | unclaimable claimed |
| CLAIM_STATUS_EXPIRED | 4 | unclaimable expired |



<a name="provenance.reward.v1.RewardProgram.State"></a>

### RewardProgram.State
State is the state of the reward program

| Name | Number | Description |
| ---- | ------ | ----------- |
| STATE_UNSPECIFIED | 0 | undefined program state |
| STATE_PENDING | 1 | pending state of reward program |
| STATE_STARTED | 2 | started state of reward program |
| STATE_FINISHED | 3 | finished state of reward program |
| STATE_EXPIRED | 4 | expired state of reward program |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/reward/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/reward/v1/genesis.proto



<a name="provenance.reward.v1.GenesisState"></a>

### GenesisState
GenesisState defines the reward module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_program_id` | [uint64](#uint64) |  | Reward program id is the next auto incremented id to be assigned to the next created reward program |
| `reward_programs` | [RewardProgram](#provenance.reward.v1.RewardProgram) | repeated | Reward programs to initially start with. |
| `claim_period_reward_distributions` | [ClaimPeriodRewardDistribution](#provenance.reward.v1.ClaimPeriodRewardDistribution) | repeated | Claim period reward distributions to initially start with. |
| `reward_account_states` | [RewardAccountState](#provenance.reward.v1.RewardAccountState) | repeated | Reward account states to initially start with. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/reward/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/reward/v1/query.proto



<a name="provenance.reward.v1.QueryClaimPeriodRewardDistributionsByIDRequest"></a>

### QueryClaimPeriodRewardDistributionsByIDRequest
QueryClaimPeriodRewardDistributionsByIDRequest queries for a single ClaimPeriodRewardDistribution


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_id` | [uint64](#uint64) |  | The reward program that the claim period reward distribution belongs to. |
| `claim_period_id` | [uint64](#uint64) |  | The claim period that the claim period reward distribution was created for. |






<a name="provenance.reward.v1.QueryClaimPeriodRewardDistributionsByIDResponse"></a>

### QueryClaimPeriodRewardDistributionsByIDResponse
QueryClaimPeriodRewardDistributionsByIDResponse returns the requested ClaimPeriodRewardDistribution


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `claim_period_reward_distribution` | [ClaimPeriodRewardDistribution](#provenance.reward.v1.ClaimPeriodRewardDistribution) |  | The ClaimPeriodRewardDistribution object that was queried for. |






<a name="provenance.reward.v1.QueryClaimPeriodRewardDistributionsRequest"></a>

### QueryClaimPeriodRewardDistributionsRequest
QueryClaimPeriodRewardDistributionsRequest queries for all the ClaimPeriodRewardDistributions with pagination.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.reward.v1.QueryClaimPeriodRewardDistributionsResponse"></a>

### QueryClaimPeriodRewardDistributionsResponse
QueryClaimPeriodRewardDistributionsResponse returns the list of paginated ClaimPeriodRewardDistributions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `claim_period_reward_distributions` | [ClaimPeriodRewardDistribution](#provenance.reward.v1.ClaimPeriodRewardDistribution) | repeated | List of all ClaimPeriodRewardDistribution objects queried for. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the response. |






<a name="provenance.reward.v1.QueryRewardDistributionsByAddressRequest"></a>

### QueryRewardDistributionsByAddressRequest
QueryRewardDistributionsByAddressRequest queries for reward claims by address that match the claim_status.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | The address that the claim belongs to. |
| `claim_status` | [RewardAccountState.ClaimStatus](#provenance.reward.v1.RewardAccountState.ClaimStatus) |  | The status that the reward account must have. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.reward.v1.QueryRewardDistributionsByAddressResponse"></a>

### QueryRewardDistributionsByAddressResponse
QueryRewardDistributionsByAddressResponse returns the reward claims for an address that match the claim_status.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | The address that the reward account belongs to. |
| `reward_account_state` | [RewardAccountResponse](#provenance.reward.v1.RewardAccountResponse) | repeated | List of RewardAccounts queried for. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the response. |






<a name="provenance.reward.v1.QueryRewardProgramByIDRequest"></a>

### QueryRewardProgramByIDRequest
QueryRewardProgramByIDRequest queries for the Reward Program with an identifier of id


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | The id of the reward program to query. |






<a name="provenance.reward.v1.QueryRewardProgramByIDResponse"></a>

### QueryRewardProgramByIDResponse
QueryRewardProgramByIDResponse contains the requested RewardProgram


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_program` | [RewardProgram](#provenance.reward.v1.RewardProgram) |  | The reward program object that was queried for. |






<a name="provenance.reward.v1.QueryRewardProgramsRequest"></a>

### QueryRewardProgramsRequest
QueryRewardProgramsRequest queries for all reward programs matching the query_type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `query_type` | [QueryRewardProgramsRequest.QueryType](#provenance.reward.v1.QueryRewardProgramsRequest.QueryType) |  | A filter on the types of reward programs. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.reward.v1.QueryRewardProgramsResponse"></a>

### QueryRewardProgramsResponse
QueryRewardProgramsResponse contains the list of RewardPrograms matching the query


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_programs` | [RewardProgram](#provenance.reward.v1.RewardProgram) | repeated | List of RewardProgram objects matching the query_type. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the response. |






<a name="provenance.reward.v1.RewardAccountResponse"></a>

### RewardAccountResponse
RewardAccountResponse is an address' reward claim for a reward program's claim period.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_program_id` | [uint64](#uint64) |  | The id of the reward program that this claim belongs to. |
| `total_reward_claim` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | total rewards claimed for all eligible claim periods in program. |
| `claim_status` | [RewardAccountState.ClaimStatus](#provenance.reward.v1.RewardAccountState.ClaimStatus) |  | The status of the claim. |
| `claim_id` | [uint64](#uint64) |  | The claim period that the claim belongs to. |





 <!-- end messages -->


<a name="provenance.reward.v1.QueryRewardProgramsRequest.QueryType"></a>

### QueryRewardProgramsRequest.QueryType
QueryType is the state of reward program to query

| Name | Number | Description |
| ---- | ------ | ----------- |
| QUERY_TYPE_UNSPECIFIED | 0 | unspecified type |
| QUERY_TYPE_ALL | 1 | all reward programs states |
| QUERY_TYPE_PENDING | 2 | pending reward program state= |
| QUERY_TYPE_ACTIVE | 3 | active reward program state |
| QUERY_TYPE_OUTSTANDING | 4 | pending and active reward program states |
| QUERY_TYPE_FINISHED | 5 | finished reward program state |


 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.reward.v1.Query"></a>

### Query
Query defines the gRPC querier service for reward module.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RewardProgramByID` | [QueryRewardProgramByIDRequest](#provenance.reward.v1.QueryRewardProgramByIDRequest) | [QueryRewardProgramByIDResponse](#provenance.reward.v1.QueryRewardProgramByIDResponse) | RewardProgramByID returns a reward program matching the ID. | GET|/provenance/rewards/v1/reward_programs/{id}|
| `RewardPrograms` | [QueryRewardProgramsRequest](#provenance.reward.v1.QueryRewardProgramsRequest) | [QueryRewardProgramsResponse](#provenance.reward.v1.QueryRewardProgramsResponse) | RewardPrograms returns a list of reward programs matching the query type. | GET|/provenance/rewards/v1/reward_programs|
| `ClaimPeriodRewardDistributions` | [QueryClaimPeriodRewardDistributionsRequest](#provenance.reward.v1.QueryClaimPeriodRewardDistributionsRequest) | [QueryClaimPeriodRewardDistributionsResponse](#provenance.reward.v1.QueryClaimPeriodRewardDistributionsResponse) | ClaimPeriodRewardDistributions returns a list of claim period reward distributions matching the claim_status. | GET|/provenance/rewards/v1/claim_period_reward_distributions|
| `ClaimPeriodRewardDistributionsByID` | [QueryClaimPeriodRewardDistributionsByIDRequest](#provenance.reward.v1.QueryClaimPeriodRewardDistributionsByIDRequest) | [QueryClaimPeriodRewardDistributionsByIDResponse](#provenance.reward.v1.QueryClaimPeriodRewardDistributionsByIDResponse) | ClaimPeriodRewardDistributionsByID returns a claim period reward distribution matching the ID. | GET|/provenance/rewards/v1/claim_period_reward_distributions/{reward_id}/claim_periods/{claim_period_id}|
| `RewardDistributionsByAddress` | [QueryRewardDistributionsByAddressRequest](#provenance.reward.v1.QueryRewardDistributionsByAddressRequest) | [QueryRewardDistributionsByAddressResponse](#provenance.reward.v1.QueryRewardDistributionsByAddressResponse) | RewardDistributionsByAddress returns a list of reward claims belonging to the account and matching the claim status. | GET|/provenance/rewards/v1/reward_claims/{address}|

 <!-- end services -->



<a name="provenance/reward/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/reward/v1/tx.proto



<a name="provenance.reward.v1.ClaimedRewardPeriodDetail"></a>

### ClaimedRewardPeriodDetail
ClaimedRewardPeriodDetail is information regarding an addresses' shares and reward for a claim period.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `claim_period_id` | [uint64](#uint64) |  | claim period id |
| `total_shares` | [uint64](#uint64) |  | total shares accumulated for claim period |
| `claim_period_reward` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | total rewards for claim period |






<a name="provenance.reward.v1.MsgClaimAllRewardsRequest"></a>

### MsgClaimAllRewardsRequest
MsgClaimRewardsResponse is the request type for claiming rewards from all reward programs RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_address` | [string](#string) |  | reward address and signer of msg to send claimed rewards to. |






<a name="provenance.reward.v1.MsgClaimAllRewardsResponse"></a>

### MsgClaimAllRewardsResponse
MsgClaimRewardsResponse is the response type for claiming rewards from all reward programs RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `total_reward_claim` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | total rewards claimed for all eligible claim periods in all programs. |
| `claim_details` | [RewardProgramClaimDetail](#provenance.reward.v1.RewardProgramClaimDetail) | repeated | details about acquired rewards from a reward program. |






<a name="provenance.reward.v1.MsgClaimRewardsRequest"></a>

### MsgClaimRewardsRequest
MsgClaimRewardsRequest is the request type for claiming reward from reward program RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_program_id` | [uint64](#uint64) |  | reward program id to claim rewards. |
| `reward_address` | [string](#string) |  | reward address and signer of msg to send claimed rewards to. |






<a name="provenance.reward.v1.MsgClaimRewardsResponse"></a>

### MsgClaimRewardsResponse
MsgClaimRewardsResponse is the response type for claiming reward from reward program RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `claim_details` | [RewardProgramClaimDetail](#provenance.reward.v1.RewardProgramClaimDetail) |  | details about acquired rewards from reward program. |






<a name="provenance.reward.v1.MsgCreateRewardProgramRequest"></a>

### MsgCreateRewardProgramRequest
MsgCreateRewardProgramRequest is the request type for creating a reward program RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title for the reward program. |
| `description` | [string](#string) |  | description for the reward program. |
| `distribute_from_address` | [string](#string) |  | provider address for the reward program funds and signer of message. |
| `total_reward_pool` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | total reward pool for the reward program. |
| `max_reward_per_claim_address` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | maximum amount of funds an address can be rewarded per claim period. |
| `program_start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | start time of the reward program. |
| `claim_periods` | [uint64](#uint64) |  | number of claim periods the reward program runs for. |
| `claim_period_days` | [uint64](#uint64) |  | number of days a claim period will exist. |
| `max_rollover_claim_periods` | [uint64](#uint64) |  | maximum number of claim periods a reward program can rollover. |
| `expire_days` | [uint64](#uint64) |  | number of days before a reward program will expire after it has ended. |
| `qualifying_actions` | [QualifyingAction](#provenance.reward.v1.QualifyingAction) | repeated | actions that count towards the reward. |






<a name="provenance.reward.v1.MsgCreateRewardProgramResponse"></a>

### MsgCreateRewardProgramResponse
MsgCreateRewardProgramResponse is the response type for creating a reward program RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | reward program id that is generated on creation. |






<a name="provenance.reward.v1.MsgEndRewardProgramRequest"></a>

### MsgEndRewardProgramRequest
MsgEndRewardProgramRequest is the request type for ending a reward program RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_program_id` | [uint64](#uint64) |  | reward program id to end. |
| `program_owner_address` | [string](#string) |  | owner of the reward program that funds were distributed from. |






<a name="provenance.reward.v1.MsgEndRewardProgramResponse"></a>

### MsgEndRewardProgramResponse
MsgEndRewardProgramResponse is the response type for ending a reward program RPC






<a name="provenance.reward.v1.RewardProgramClaimDetail"></a>

### RewardProgramClaimDetail
RewardProgramClaimDetail is the response object regarding an address's shares and reward for a reward program.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward_program_id` | [uint64](#uint64) |  | reward program id. |
| `total_reward_claim` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | total rewards claimed for all eligible claim periods in program. |
| `claimed_reward_period_details` | [ClaimedRewardPeriodDetail](#provenance.reward.v1.ClaimedRewardPeriodDetail) | repeated | claim period details. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.reward.v1.Msg"></a>

### Msg
Msg

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateRewardProgram` | [MsgCreateRewardProgramRequest](#provenance.reward.v1.MsgCreateRewardProgramRequest) | [MsgCreateRewardProgramResponse](#provenance.reward.v1.MsgCreateRewardProgramResponse) | CreateRewardProgram is the RPC endpoint for creating a rewards program | |
| `EndRewardProgram` | [MsgEndRewardProgramRequest](#provenance.reward.v1.MsgEndRewardProgramRequest) | [MsgEndRewardProgramResponse](#provenance.reward.v1.MsgEndRewardProgramResponse) | EndRewardProgram is the RPC endpoint for ending a rewards program | |
| `ClaimRewards` | [MsgClaimRewardsRequest](#provenance.reward.v1.MsgClaimRewardsRequest) | [MsgClaimRewardsResponse](#provenance.reward.v1.MsgClaimRewardsResponse) | ClaimRewards is the RPC endpoint for claiming rewards belonging to completed claim periods of a reward program | |
| `ClaimAllRewards` | [MsgClaimAllRewardsRequest](#provenance.reward.v1.MsgClaimAllRewardsRequest) | [MsgClaimAllRewardsResponse](#provenance.reward.v1.MsgClaimAllRewardsResponse) | ClaimAllRewards is the RPC endpoint for claiming rewards for completed claim periods of every reward program for the signer of the tx. | |

 <!-- end services -->



<a name="provenance/trigger/v1/event.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/event.proto



<a name="provenance.trigger.v1.EventTriggerCreated"></a>

### EventTriggerCreated
EventTriggerCreated is an event for when a trigger is created


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [string](#string) |  | trigger_id is a unique identifier of the trigger |






<a name="provenance.trigger.v1.EventTriggerDestroyed"></a>

### EventTriggerDestroyed
EventTriggerDestroyed is an event for when a trigger is destroyed


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [string](#string) |  | trigger_id is a unique identifier of the trigger |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/trigger/v1/trigger.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/trigger.proto



<a name="provenance.trigger.v1.Attribute"></a>

### Attribute
Attribute


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The name of the attribute that the event must have to be considered a match. |
| `value` | [string](#string) |  | The value of the attribute that the event must have to be considered a match. |






<a name="provenance.trigger.v1.BlockHeightEvent"></a>

### BlockHeightEvent
BlockHeightEvent


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `block_height` | [uint64](#uint64) |  | The height that the trigger should fire at. |






<a name="provenance.trigger.v1.BlockTimeEvent"></a>

### BlockTimeEvent
BlockTimeEvent


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | The time the trigger should fire at. |






<a name="provenance.trigger.v1.QueuedTrigger"></a>

### QueuedTrigger
QueuedTrigger


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `block_height` | [uint64](#uint64) |  | The block height the trigger was detected and queued. |
| `time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | The time the trigger was detected and queued. |
| `trigger` | [Trigger](#provenance.trigger.v1.Trigger) |  | The trigger that was detected. |






<a name="provenance.trigger.v1.TransactionEvent"></a>

### TransactionEvent
TransactionEvent


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The name of the event for a match. |
| `attributes` | [Attribute](#provenance.trigger.v1.Attribute) | repeated | The attributes that must be present for a match. |






<a name="provenance.trigger.v1.Trigger"></a>

### Trigger
Trigger


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | An integer to uniquely identify the trigger. |
| `owner` | [string](#string) |  | The owner of the trigger. |
| `event` | [google.protobuf.Any](#google.protobuf.Any) |  | The event that must be detected for the trigger to fire. |
| `actions` | [google.protobuf.Any](#google.protobuf.Any) | repeated | The messages to run when the trigger fires. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/trigger/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/genesis.proto



<a name="provenance.trigger.v1.GasLimit"></a>

### GasLimit
GasLimit defines the trigger module's grouping of a trigger and a gas limit


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [uint64](#uint64) |  | The identifier of the trigger this GasLimit belongs to. |
| `amount` | [uint64](#uint64) |  | The maximum amount of gas that the trigger can use. |






<a name="provenance.trigger.v1.GenesisState"></a>

### GenesisState
GenesisState defines the trigger module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger_id` | [uint64](#uint64) |  | Trigger id is the next auto incremented id to be assigned to the next created trigger |
| `queue_start` | [uint64](#uint64) |  | Queue start is the starting index of the queue. |
| `triggers` | [Trigger](#provenance.trigger.v1.Trigger) | repeated | Triggers to initially start with. |
| `gas_limits` | [GasLimit](#provenance.trigger.v1.GasLimit) | repeated | Maximum amount of gas that the triggers can use. |
| `queued_triggers` | [QueuedTrigger](#provenance.trigger.v1.QueuedTrigger) | repeated | Triggers to initially start with in the queue. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/trigger/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/query.proto



<a name="provenance.trigger.v1.QueryTriggerByIDRequest"></a>

### QueryTriggerByIDRequest
QueryTriggerByIDRequest queries for the Trigger with an identifier of id.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | The id of the trigger to query. |






<a name="provenance.trigger.v1.QueryTriggerByIDResponse"></a>

### QueryTriggerByIDResponse
QueryTriggerByIDResponse contains the requested Trigger.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `trigger` | [Trigger](#provenance.trigger.v1.Trigger) |  | The trigger object that was queried for. |






<a name="provenance.trigger.v1.QueryTriggersRequest"></a>

### QueryTriggersRequest
QueryTriggersRequest queries for all triggers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.trigger.v1.QueryTriggersResponse"></a>

### QueryTriggersResponse
QueryTriggersResponse contains the list of Triggers.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `triggers` | [Trigger](#provenance.trigger.v1.Trigger) | repeated | List of Trigger objects. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines an optional pagination for the response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.trigger.v1.Query"></a>

### Query
Query defines the gRPC querier service for trigger module.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `TriggerByID` | [QueryTriggerByIDRequest](#provenance.trigger.v1.QueryTriggerByIDRequest) | [QueryTriggerByIDResponse](#provenance.trigger.v1.QueryTriggerByIDResponse) | TriggerByID returns a trigger matching the ID. | GET|/provenance/trigger/v1/triggers/{id}|
| `Triggers` | [QueryTriggersRequest](#provenance.trigger.v1.QueryTriggersRequest) | [QueryTriggersResponse](#provenance.trigger.v1.QueryTriggersResponse) | Triggers returns the list of triggers. | GET|/provenance/trigger/v1/triggers|

 <!-- end services -->



<a name="provenance/trigger/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/trigger/v1/tx.proto



<a name="provenance.trigger.v1.MsgCreateTriggerRequest"></a>

### MsgCreateTriggerRequest
MsgCreateTriggerRequest is the request type for creating a trigger RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authorities` | [string](#string) | repeated | The signing authorities for the request |
| `event` | [google.protobuf.Any](#google.protobuf.Any) |  | The event that must be detected for the trigger to fire. |
| `actions` | [google.protobuf.Any](#google.protobuf.Any) | repeated | The messages to run when the trigger fires. |






<a name="provenance.trigger.v1.MsgCreateTriggerResponse"></a>

### MsgCreateTriggerResponse
MsgCreateTriggerResponse is the response type for creating a trigger RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | trigger id that is generated on creation. |






<a name="provenance.trigger.v1.MsgDestroyTriggerRequest"></a>

### MsgDestroyTriggerRequest
MsgDestroyTriggerRequest is the request type for creating a trigger RPC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `id` | [uint64](#uint64) |  | the id of the trigger to destroy. |
| `authority` | [string](#string) |  | The signing authority for the request |






<a name="provenance.trigger.v1.MsgDestroyTriggerResponse"></a>

### MsgDestroyTriggerResponse
MsgDestroyTriggerResponse is the response type for creating a trigger RPC





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.trigger.v1.Msg"></a>

### Msg
Msg

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateTrigger` | [MsgCreateTriggerRequest](#provenance.trigger.v1.MsgCreateTriggerRequest) | [MsgCreateTriggerResponse](#provenance.trigger.v1.MsgCreateTriggerResponse) | CreateTrigger is the RPC endpoint for creating a trigger | |
| `DestroyTrigger` | [MsgDestroyTriggerRequest](#provenance.trigger.v1.MsgDestroyTriggerRequest) | [MsgDestroyTriggerResponse](#provenance.trigger.v1.MsgDestroyTriggerResponse) | DestroyTrigger is the RPC endpoint for creating a trigger | |

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

