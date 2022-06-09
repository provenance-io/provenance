<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [provenance/attribute/v1/attribute.proto](#provenance/attribute/v1/attribute.proto)
    - [Attribute](#provenance.attribute.v1.Attribute)
    - [EventAttributeAdd](#provenance.attribute.v1.EventAttributeAdd)
    - [EventAttributeDelete](#provenance.attribute.v1.EventAttributeDelete)
    - [EventAttributeDistinctDelete](#provenance.attribute.v1.EventAttributeDistinctDelete)
    - [EventAttributeUpdate](#provenance.attribute.v1.EventAttributeUpdate)
    - [Params](#provenance.attribute.v1.Params)
  
    - [AttributeType](#provenance.attribute.v1.AttributeType)
  
- [provenance/attribute/v1/genesis.proto](#provenance/attribute/v1/genesis.proto)
    - [GenesisState](#provenance.attribute.v1.GenesisState)
  
- [provenance/attribute/v1/query.proto](#provenance/attribute/v1/query.proto)
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
    - [MsgUpdateAttributeRequest](#provenance.attribute.v1.MsgUpdateAttributeRequest)
    - [MsgUpdateAttributeResponse](#provenance.attribute.v1.MsgUpdateAttributeResponse)
  
    - [Msg](#provenance.attribute.v1.Msg)
  
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
    - [MsgMintRequest](#provenance.marker.v1.MsgMintRequest)
    - [MsgMintResponse](#provenance.marker.v1.MsgMintResponse)
    - [MsgSetDenomMetadataRequest](#provenance.marker.v1.MsgSetDenomMetadataRequest)
    - [MsgSetDenomMetadataResponse](#provenance.marker.v1.MsgSetDenomMetadataResponse)
    - [MsgTransferRequest](#provenance.marker.v1.MsgTransferRequest)
    - [MsgTransferResponse](#provenance.marker.v1.MsgTransferResponse)
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
    - [ContractSpecificationRequest](#provenance.metadata.v1.ContractSpecificationRequest)
    - [ContractSpecificationResponse](#provenance.metadata.v1.ContractSpecificationResponse)
    - [ContractSpecificationWrapper](#provenance.metadata.v1.ContractSpecificationWrapper)
    - [ContractSpecificationsAllRequest](#provenance.metadata.v1.ContractSpecificationsAllRequest)
    - [ContractSpecificationsAllResponse](#provenance.metadata.v1.ContractSpecificationsAllResponse)
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
    - [MsgModifyOSLocatorRequest](#provenance.metadata.v1.MsgModifyOSLocatorRequest)
    - [MsgModifyOSLocatorResponse](#provenance.metadata.v1.MsgModifyOSLocatorResponse)
    - [MsgP8eMemorializeContractRequest](#provenance.metadata.v1.MsgP8eMemorializeContractRequest)
    - [MsgP8eMemorializeContractResponse](#provenance.metadata.v1.MsgP8eMemorializeContractResponse)
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
    - [MsgAssessCustomMsgFeeRequest](#provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest)
    - [MsgAssessCustomMsgFeeResponse](#provenance.msgfees.v1.MsgAssessCustomMsgFeeResponse)
  
    - [Msg](#provenance.msgfees.v1.Msg)
  
- [provenance/name/v1/name.proto](#provenance/name/v1/name.proto)
    - [CreateRootNameProposal](#provenance.name.v1.CreateRootNameProposal)
    - [EventNameBound](#provenance.name.v1.EventNameBound)
    - [EventNameUnbound](#provenance.name.v1.EventNameUnbound)
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
    - [MsgDeleteNameRequest](#provenance.name.v1.MsgDeleteNameRequest)
    - [MsgDeleteNameResponse](#provenance.name.v1.MsgDeleteNameResponse)
  
    - [Msg](#provenance.name.v1.Msg)
  
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
QueryAttributesResponse is the response type for the Query/Attribute method.


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
QueryScanRequest is the request type for the Query/Scan account attributes method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `account` | [string](#string) |  | account defines the address to query for. |
| `suffix` | [string](#string) |  | name defines the partial attribute name to search for base on names being in RDNS format. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="provenance.attribute.v1.QueryScanResponse"></a>

### QueryScanResponse
QueryScanResponse is the response type for the Query/Attribute method.


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

 <!-- end services -->



<a name="provenance/attribute/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/attribute/v1/tx.proto



<a name="provenance.attribute.v1.MsgAddAttributeRequest"></a>

### MsgAddAttributeRequest
MsgAddAttributeRequest defines an sdk.Msg type that is used to add a new attribute to an account
Attributes may only be set in an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The attribute value. |
| `attribute_type` | [AttributeType](#provenance.attribute.v1.AttributeType) |  | The attribute value type. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance.attribute.v1.MsgAddAttributeResponse"></a>

### MsgAddAttributeResponse
MsgAddAttributeResponse defines the Msg/Vote response type.






<a name="provenance.attribute.v1.MsgDeleteAttributeRequest"></a>

### MsgDeleteAttributeRequest
MsgDeleteAttributeRequest defines a message to delete an attribute from an account
Attributes may only be remove from an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance.attribute.v1.MsgDeleteAttributeResponse"></a>

### MsgDeleteAttributeResponse
MsgDeleteAttributeResponse defines the Msg/Vote response type.






<a name="provenance.attribute.v1.MsgDeleteDistinctAttributeRequest"></a>

### MsgDeleteDistinctAttributeRequest
MsgDeleteDistinctAttributeRequest defines a message to delete an attribute with matching name, value, and type from
an account Attributes may only be remove from an account by the account that the attribute name resolves to.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The attribute name. |
| `value` | [bytes](#bytes) |  | The attribute value. |
| `account` | [string](#string) |  | The account to add the attribute to. |
| `owner` | [string](#string) |  | The address that the name must resolve to. |






<a name="provenance.attribute.v1.MsgDeleteDistinctAttributeResponse"></a>

### MsgDeleteDistinctAttributeResponse
MsgDeleteDistinctAttributeResponse defines the Msg/Vote response type.






<a name="provenance.attribute.v1.MsgUpdateAttributeRequest"></a>

### MsgUpdateAttributeRequest
MsgUpdateAttributeRequest defines an sdk.Msg type that is used to update an existing attribute to an account
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
MsgUpdateAttributeResponse defines the Msg/Vote response type.





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
| `DeleteAttribute` | [MsgDeleteAttributeRequest](#provenance.attribute.v1.MsgDeleteAttributeRequest) | [MsgDeleteAttributeResponse](#provenance.attribute.v1.MsgDeleteAttributeResponse) | DeleteAttribute defines a method to verify a particular invariance. | |
| `DeleteDistinctAttribute` | [MsgDeleteDistinctAttributeRequest](#provenance.attribute.v1.MsgDeleteDistinctAttributeRequest) | [MsgDeleteDistinctAttributeResponse](#provenance.attribute.v1.MsgDeleteDistinctAttributeResponse) | DeleteDistinctAttribute defines a method to verify a particular invariance. | |

 <!-- end services -->



<a name="provenance/marker/v1/accessgrant.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/marker/v1/accessgrant.proto



<a name="provenance.marker.v1.AccessGrant"></a>

### AccessGrant
AccessGrant associates a colelction of permisssions with an address for delegated marker account control.


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
AddMarkerProposal defines defines a governance proposal to create a new marker


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






<a name="provenance.marker.v1.MsgAddMarkerRequest"></a>

### MsgAddMarkerRequest
MsgAddMarkerRequest defines the Msg/AddMarker request type


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
| `SetDenomMetadata` | [MsgSetDenomMetadataRequest](#provenance.marker.v1.MsgSetDenomMetadataRequest) | [MsgSetDenomMetadataResponse](#provenance.marker.v1.MsgSetDenomMetadataResponse) | Allows Denom Metadata (see bank module) to be set for the Marker's Denom | |
| `GrantAllowance` | [MsgGrantAllowanceRequest](#provenance.marker.v1.MsgGrantAllowanceRequest) | [MsgGrantAllowanceResponse](#provenance.marker.v1.MsgGrantAllowanceResponse) | GrantAllowance grants fee allowance to the grantee on the granter's account with the provided expiration time. | |

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
| `data_access` | [string](#string) | repeated | Addessses in this list are authorized to recieve off-chain data associated with this scope. |
| `value_owner_address` | [string](#string) |  | An address that controls the value associated with this scope. Standard blockchain accounts and marker accounts are supported for this value. This attribute may only be changed by the entity indicated once it is set. |






<a name="provenance.metadata.v1.Session"></a>

### Session
A Session is created for an execution context against a specific specification instance

The context will have a specification and set of parties involved.  The Session may be updated several
times so long as the parties listed are signers on the transaction.  NOTE: When there are no Records within a Scope
that reference a Session it is removed.


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
| RESULT_STATUS_PASS | 1 | RESULT_STATUS_PASS indicates the execution was successfult |
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



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `condition_name` | [string](#string) |  |  |
| `result` | [ExecutionResult](#provenance.metadata.v1.p8e.ExecutionResult) |  |  |






<a name="provenance.metadata.v1.p8e.ConditionSpec"></a>

### ConditionSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `func_name` | [string](#string) |  |  |
| `input_specs` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) | repeated |  |
| `output_spec` | [OutputSpec](#provenance.metadata.v1.p8e.OutputSpec) |  |  |






<a name="provenance.metadata.v1.p8e.Consideration"></a>

### Consideration



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `consideration_name` | [string](#string) |  |  |
| `inputs` | [ProposedFact](#provenance.metadata.v1.p8e.ProposedFact) | repeated | Data pushed to a consideration that will ultimately match the output_spec of the consideration |
| `result` | [ExecutionResult](#provenance.metadata.v1.p8e.ExecutionResult) |  |  |






<a name="provenance.metadata.v1.p8e.ConsiderationSpec"></a>

### ConsiderationSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `func_name` | [string](#string) |  |  |
| `responsible_party` | [PartyType](#provenance.metadata.v1.p8e.PartyType) |  | Invoking party |
| `input_specs` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) | repeated |  |
| `output_spec` | [OutputSpec](#provenance.metadata.v1.p8e.OutputSpec) |  |  |






<a name="provenance.metadata.v1.p8e.Contract"></a>

### Contract



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `definition` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) |  |  |
| `spec` | [Fact](#provenance.metadata.v1.p8e.Fact) |  | Points to the proto for the contractSpec |
| `invoker` | [SigningAndEncryptionPublicKeys](#provenance.metadata.v1.p8e.SigningAndEncryptionPublicKeys) |  | Invoker of this contract |
| `inputs` | [Fact](#provenance.metadata.v1.p8e.Fact) | repeated | Constructor arguments. These are always the output of a previously recorded consideration. |
| `conditions` | [Condition](#provenance.metadata.v1.p8e.Condition) | repeated | **Deprecated.** conditions is a deprecated field that is not used at all anymore. |
| `considerations` | [Consideration](#provenance.metadata.v1.p8e.Consideration) | repeated |  |
| `recitals` | [Recital](#provenance.metadata.v1.p8e.Recital) | repeated |  |
| `times_executed` | [int32](#int32) |  |  |
| `start_time` | [Timestamp](#provenance.metadata.v1.p8e.Timestamp) |  | This is only set once when the contract is initially executed |
| `context` | [bytes](#bytes) |  |  |






<a name="provenance.metadata.v1.p8e.ContractSpec"></a>

### ContractSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `definition` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) |  |  |
| `input_specs` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) | repeated |  |
| `parties_involved` | [PartyType](#provenance.metadata.v1.p8e.PartyType) | repeated |  |
| `condition_specs` | [ConditionSpec](#provenance.metadata.v1.p8e.ConditionSpec) | repeated |  |
| `consideration_specs` | [ConsiderationSpec](#provenance.metadata.v1.p8e.ConsiderationSpec) | repeated |  |






<a name="provenance.metadata.v1.p8e.DefinitionSpec"></a>

### DefinitionSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `resource_location` | [Location](#provenance.metadata.v1.p8e.Location) |  |  |
| `signature` | [Signature](#provenance.metadata.v1.p8e.Signature) |  |  |
| `type` | [DefinitionSpecType](#provenance.metadata.v1.p8e.DefinitionSpecType) |  |  |






<a name="provenance.metadata.v1.p8e.ExecutionResult"></a>

### ExecutionResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `output` | [ProposedFact](#provenance.metadata.v1.p8e.ProposedFact) |  |  |
| `result` | [ExecutionResultType](#provenance.metadata.v1.p8e.ExecutionResultType) |  |  |
| `recorded_at` | [Timestamp](#provenance.metadata.v1.p8e.Timestamp) |  |  |
| `error_message` | [string](#string) |  |  |






<a name="provenance.metadata.v1.p8e.Fact"></a>

### Fact



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `data_location` | [Location](#provenance.metadata.v1.p8e.Location) |  |  |






<a name="provenance.metadata.v1.p8e.Location"></a>

### Location



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `ref` | [ProvenanceReference](#provenance.metadata.v1.p8e.ProvenanceReference) |  |  |
| `classname` | [string](#string) |  |  |






<a name="provenance.metadata.v1.p8e.OutputSpec"></a>

### OutputSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `spec` | [DefinitionSpec](#provenance.metadata.v1.p8e.DefinitionSpec) |  |  |






<a name="provenance.metadata.v1.p8e.ProposedFact"></a>

### ProposedFact



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  |  |
| `hash` | [string](#string) |  |  |
| `classname` | [string](#string) |  |  |
| `ancestor` | [ProvenanceReference](#provenance.metadata.v1.p8e.ProvenanceReference) |  |  |






<a name="provenance.metadata.v1.p8e.ProvenanceReference"></a>

### ProvenanceReference



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_uuid` | [UUID](#provenance.metadata.v1.p8e.UUID) |  | [Req] [Scope.uuid] Scope ID |
| `group_uuid` | [UUID](#provenance.metadata.v1.p8e.UUID) |  | [Opt] [RecordGroup.group_uuid] require record to be within a specific group |
| `hash` | [string](#string) |  | [Opt] [Record.result_hash] specify a specific record inside a scope (and group) by result-hash |
| `name` | [string](#string) |  | [Opt] [Record.result_name] specify a result-name of a record within a scope |






<a name="provenance.metadata.v1.p8e.PublicKey"></a>

### PublicKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `public_key_bytes` | [bytes](#bytes) |  |  |
| `type` | [PublicKeyType](#provenance.metadata.v1.p8e.PublicKeyType) |  |  |
| `curve` | [PublicKeyCurve](#provenance.metadata.v1.p8e.PublicKeyCurve) |  |  |






<a name="provenance.metadata.v1.p8e.Recital"></a>

### Recital



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signer_role` | [PartyType](#provenance.metadata.v1.p8e.PartyType) |  |  |
| `signer` | [SigningAndEncryptionPublicKeys](#provenance.metadata.v1.p8e.SigningAndEncryptionPublicKeys) |  |  |
| `address` | [bytes](#bytes) |  |  |






<a name="provenance.metadata.v1.p8e.Recitals"></a>

### Recitals



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `parties` | [Recital](#provenance.metadata.v1.p8e.Recital) | repeated |  |






<a name="provenance.metadata.v1.p8e.Signature"></a>

### Signature



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `algo` | [string](#string) |  | Signature Detail |
| `provider` | [string](#string) |  |  |
| `signature` | [string](#string) |  |  |
| `signer` | [SigningAndEncryptionPublicKeys](#provenance.metadata.v1.p8e.SigningAndEncryptionPublicKeys) |  | Identity of signer |






<a name="provenance.metadata.v1.p8e.SignatureSet"></a>

### SignatureSet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signatures` | [Signature](#provenance.metadata.v1.p8e.Signature) | repeated |  |






<a name="provenance.metadata.v1.p8e.SigningAndEncryptionPublicKeys"></a>

### SigningAndEncryptionPublicKeys



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `signing_public_key` | [PublicKey](#provenance.metadata.v1.p8e.PublicKey) |  |  |
| `encryption_public_key` | [PublicKey](#provenance.metadata.v1.p8e.PublicKey) |  |  |






<a name="provenance.metadata.v1.p8e.Timestamp"></a>

### Timestamp
A Timestamp represents a point in time using values relative to the epoch.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `seconds` | [int64](#int64) |  | Represents seconds of UTC time since Unix epoch |
| `nanos` | [int32](#int32) |  | Non-negative fractions of a second at nanosecond resolution. |






<a name="provenance.metadata.v1.p8e.UUID"></a>

### UUID



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `value` | [string](#string) |  |  |





 <!-- end messages -->


<a name="provenance.metadata.v1.p8e.DefinitionSpecType"></a>

### DefinitionSpecType


| Name | Number | Description |
| ---- | ------ | ----------- |
| DEFINITION_SPEC_TYPE_UNKNOWN | 0 |  |
| DEFINITION_SPEC_TYPE_PROPOSED | 1 |  |
| DEFINITION_SPEC_TYPE_FACT | 2 |  |
| DEFINITION_SPEC_TYPE_FACT_LIST | 3 |  |



<a name="provenance.metadata.v1.p8e.ExecutionResultType"></a>

### ExecutionResultType


| Name | Number | Description |
| ---- | ------ | ----------- |
| RESULT_TYPE_UNKNOWN | 0 |  |
| RESULT_TYPE_PASS | 1 |  |
| RESULT_TYPE_SKIP | 2 | Couldn't process the condition/consideration due to missing facts being generated by other considerations. |
| RESULT_TYPE_FAIL | 3 |  |



<a name="provenance.metadata.v1.p8e.PartyType"></a>

### PartyType


| Name | Number | Description |
| ---- | ------ | ----------- |
| PARTY_TYPE_UNKNOWN | 0 |  |
| PARTY_TYPE_ORIGINATOR | 1 |  |
| PARTY_TYPE_SERVICER | 2 |  |
| PARTY_TYPE_INVESTOR | 3 |  |
| PARTY_TYPE_CUSTODIAN | 4 |  |
| PARTY_TYPE_OWNER | 5 |  |
| PARTY_TYPE_AFFILIATE | 6 |  |
| PARTY_TYPE_OMNIBUS | 7 |  |
| PARTY_TYPE_PROVENANCE | 8 |  |
| PARTY_TYPE_MARKER | 9 |  |
| PARTY_TYPE_CONTROLLER | 10 |  |
| PARTY_TYPE_VALIDATOR | 11 |  |



<a name="provenance.metadata.v1.p8e.PublicKeyCurve"></a>

### PublicKeyCurve


| Name | Number | Description |
| ---- | ------ | ----------- |
| SECP256K1 | 0 |  |
| P256 | 1 |  |



<a name="provenance.metadata.v1.p8e.PublicKeyType"></a>

### PublicKeyType


| Name | Number | Description |
| ---- | ------ | ----------- |
| ELLIPTIC | 0 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="provenance/metadata/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## provenance/metadata/v1/query.proto



<a name="provenance.metadata.v1.ContractSpecificationRequest"></a>

### ContractSpecificationRequest
ContractSpecificationRequest is the request type for the Query/ContractSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `specification_id` | [string](#string) |  | specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84 or a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn. It can also be a record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. |
| `include_record_specs` | [bool](#bool) |  | include_record_specs is a flag for whether or not the record specifications in this contract specification should be included in the result. |






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
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines optional pagination parameters for the request. |






<a name="provenance.metadata.v1.ContractSpecificationsAllResponse"></a>

### ContractSpecificationsAllResponse
ContractSpecificationsAllResponse is the response type for the Query/ContractSpecificationsAll RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_specifications` | [ContractSpecificationWrapper](#provenance.metadata.v1.ContractSpecificationWrapper) | repeated | contract_specifications are the wrapped contract specifications. |
| `request` | [ContractSpecificationsAllRequest](#provenance.metadata.v1.ContractSpecificationsAllRequest) |  | request is a copy of the request that generated these results. |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination provides the pagination information of this response. |






<a name="provenance.metadata.v1.OSAllLocatorsRequest"></a>

### OSAllLocatorsRequest
OSAllLocatorsRequest is the request type for the Query/OSAllLocators RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
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
| `include_scope` | [bool](#bool) |  | include_scope is a flag for whether or not the scope containing these records should be included. |
| `include_sessions` | [bool](#bool) |  | include_sessions is a flag for whether or not the sessions containing these records should be included. |






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
| `include_sessions` | [bool](#bool) |  | include_sessions is a flag for whether or not the sessions in the scope should be included. |
| `include_records` | [bool](#bool) |  | include_records is a flag for whether or not the records in the scope should be included. |






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






<a name="provenance.metadata.v1.ScopeSpecificationResponse"></a>

### ScopeSpecificationResponse
ScopeSpecificationResponse is the response type for the Query/ScopeSpecification RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_specification` | [ScopeSpecificationWrapper](#provenance.metadata.v1.ScopeSpecificationWrapper) |  | scope_specification is the wrapped scope specification. |
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
| `include_scope` | [bool](#bool) |  | include_scope is a flag for whether or not the scope containing these sessions should be included. |
| `include_records` | [bool](#bool) |  | include_records is a flag for whether or not the records in these sessions should be included. |






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

The specification_id can either be a uuid, e.g. dc83ea70-eacd-40fe-9adf-1cf6148bf8a2 or a bech32 scope specification address, e.g. scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m. | GET|/provenance/metadata/v1/scopespec/{specification_id}|
| `ScopeSpecificationsAll` | [ScopeSpecificationsAllRequest](#provenance.metadata.v1.ScopeSpecificationsAllRequest) | [ScopeSpecificationsAllResponse](#provenance.metadata.v1.ScopeSpecificationsAllResponse) | ScopeSpecificationsAll retrieves all scope specifications. | GET|/provenance/metadata/v1/scopespecs/all|
| `ContractSpecification` | [ContractSpecificationRequest](#provenance.metadata.v1.ContractSpecificationRequest) | [ContractSpecificationResponse](#provenance.metadata.v1.ContractSpecificationResponse) | ContractSpecification returns a contract specification for the given specification id.

The specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84, a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn, or a bech32 record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. If it is a record specification address, then the contract specification that contains that record specification is looked up.

By default, the record specifications for this contract specification are not included. Set include_record_specs to true to include them in the result. | GET|/provenance/metadata/v1/contractspec/{specification_id}|
| `ContractSpecificationsAll` | [ContractSpecificationsAllRequest](#provenance.metadata.v1.ContractSpecificationsAllRequest) | [ContractSpecificationsAllResponse](#provenance.metadata.v1.ContractSpecificationsAllResponse) | ContractSpecificationsAll retrieves all contract specifications. | GET|/provenance/metadata/v1/contractspecs/all|
| `RecordSpecificationsForContractSpecification` | [RecordSpecificationsForContractSpecificationRequest](#provenance.metadata.v1.RecordSpecificationsForContractSpecificationRequest) | [RecordSpecificationsForContractSpecificationResponse](#provenance.metadata.v1.RecordSpecificationsForContractSpecificationResponse) | RecordSpecificationsForContractSpecification returns the record specifications for the given input.

The specification_id can either be a uuid, e.g. def6bc0a-c9dd-4874-948f-5206e6060a84, a bech32 contract specification address, e.g. contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn, or a bech32 record specification address, e.g. recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44. If it is a record specification address, then the contract specification that contains that record specification is used. | GET|/provenance/metadata/v1/contractspec/{specification_id}/recordspecs|
| `RecordSpecification` | [RecordSpecificationRequest](#provenance.metadata.v1.RecordSpecificationRequest) | [RecordSpecificationResponse](#provenance.metadata.v1.RecordSpecificationResponse) | RecordSpecification returns a record specification for the given input. | GET|/provenance/metadata/v1/recordspec/{specification_id}GET|/provenance/metadata/v1/contractspec/{specification_id}/recordspec/{name}|
| `RecordSpecificationsAll` | [RecordSpecificationsAllRequest](#provenance.metadata.v1.RecordSpecificationsAllRequest) | [RecordSpecificationsAllResponse](#provenance.metadata.v1.RecordSpecificationsAllResponse) | RecordSpecificationsAll retrieves all record specifications. | GET|/provenance/metadata/v1/recordspecs/all|
| `OSLocatorParams` | [OSLocatorParamsRequest](#provenance.metadata.v1.OSLocatorParamsRequest) | [OSLocatorParamsResponse](#provenance.metadata.v1.OSLocatorParamsResponse) | OSLocatorParams returns all parameters for the object store locator sub module. | GET|/provenance/metadata/v1/locator/params|
| `OSLocator` | [OSLocatorRequest](#provenance.metadata.v1.OSLocatorRequest) | [OSLocatorResponse](#provenance.metadata.v1.OSLocatorResponse) | OSLocator returns an ObjectStoreLocator by its owner's address. | GET|/provenance/metadata/v1/locator/{owner}|
| `OSLocatorsByURI` | [OSLocatorsByURIRequest](#provenance.metadata.v1.OSLocatorsByURIRequest) | [OSLocatorsByURIResponse](#provenance.metadata.v1.OSLocatorsByURIResponse) | OSLocatorsByURI returns all ObjectStoreLocator entries for a locator uri. | GET|/provenance/metadata/v1/locator/uri/{uri}|
| `OSLocatorsByScope` | [OSLocatorsByScopeRequest](#provenance.metadata.v1.OSLocatorsByScopeRequest) | [OSLocatorsByScopeResponse](#provenance.metadata.v1.OSLocatorsByScopeResponse) | OSLocatorsByScope returns all ObjectStoreLocator entries for a for all signer's present in the specified scope. | GET|/provenance/metadata/v1/locator/scope/{scope_id}|
| `OSAllLocators` | [OSAllLocatorsRequest](#provenance.metadata.v1.OSAllLocatorsRequest) | [OSAllLocatorsResponse](#provenance.metadata.v1.OSAllLocatorsResponse) | OSAllLocators returns all ObjectStoreLocator entries. | GET|/provenance/metadata/v1/locators/all|

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
| `owners` | [Party](#provenance.metadata.v1.Party) | repeated | AccAddress owner addresses to be added to scope |
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
MsgP8eMemorializeContractRequest is the request type for the Msg/P8eMemorializeContract RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id` | [string](#string) |  | The scope id of the object being add or modified on blockchain. |
| `group_id` | [string](#string) |  | The uuid of the contract execution. |
| `scope_specification_id` | [string](#string) |  | The scope specification id. |
| `recitals` | [p8e.Recitals](#provenance.metadata.v1.p8e.Recitals) |  | The new recitals for the scope. Used in leu of Contract for direct ownership changes. |
| `contract` | [p8e.Contract](#provenance.metadata.v1.p8e.Contract) |  | The executed contract. |
| `signatures` | [p8e.SignatureSet](#provenance.metadata.v1.p8e.SignatureSet) |  | The contract signatures |
| `invoker` | [string](#string) |  | The bech32 address of the notary (ie the broadcaster of this message). |






<a name="provenance.metadata.v1.MsgP8eMemorializeContractResponse"></a>

### MsgP8eMemorializeContractResponse
MsgP8eMemorializeContractResponse is the response type for the Msg/P8eMemorializeContract RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `scope_id_info` | [ScopeIdInfo](#provenance.metadata.v1.ScopeIdInfo) |  | scope_id_info contains information about the id/address of the scope that was added or updated. |
| `session_id_info` | [SessionIdInfo](#provenance.metadata.v1.SessionIdInfo) |  | session_id_info contains information about the id/address of the session that was added or updated. |
| `record_id_infos` | [RecordIdInfo](#provenance.metadata.v1.RecordIdInfo) | repeated | record_id_infos contains information about the ids/addresses of the records that were added or updated. |






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
MsgWriteP8eContractSpecRequest is the request type for the Msg/WriteP8eContractSpec RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contractspec` | [p8e.ContractSpec](#provenance.metadata.v1.p8e.ContractSpec) |  | ContractSpec v39 p8e ContractSpect to be converted into a v40 |
| `signers` | [string](#string) | repeated |  |






<a name="provenance.metadata.v1.MsgWriteP8eContractSpecResponse"></a>

### MsgWriteP8eContractSpecResponse
MsgWriteP8eContractSpecResponse is the response type for the Msg/WriteP8eContractSpec RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_spec_id_info` | [ContractSpecIdInfo](#provenance.metadata.v1.ContractSpecIdInfo) |  | contract_spec_id_info contains information about the id/address of the contract specification that was added or updated. |
| `record_spec_id_infos` | [RecordSpecIdInfo](#provenance.metadata.v1.RecordSpecIdInfo) | repeated | record_spec_id_infos contains information about the ids/addresses of the record specifications that were added or updated. |






<a name="provenance.metadata.v1.MsgWriteRecordRequest"></a>

### MsgWriteRecordRequest
MsgWriteRecordRequest is the request type for the Msg/WriteRecord RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record` | [Record](#provenance.metadata.v1.Record) |  | record is the Record you want added or updated. |
| `signers` | [string](#string) | repeated | signers is the list of address of those signing this request. |
| `session_id_components` | [SessionIdComponents](#provenance.metadata.v1.SessionIdComponents) |  | SessionIDComponents is an optional (alternate) way of defining what the session_id should be in the provided record. If provided, it must have both a scope and session_uuid. Those components will be used to create the MetadataAddress for the session which will override the session_id in the provided record. If not provided (or all empty), nothing special happens. If there is a value in record.session_id that is different from the one created from these components, an error is returned. |
| `contract_spec_uuid` | [string](#string) |  | contract_spec_uuid is an optional contract specification uuid string, e.g. "def6bc0a-c9dd-4874-948f-5206e6060a84" If provided, it will be combined with the record name to generate the MetadataAddress for the record specification which will override the specification_id in the provided record. If not provided (or it is an empty string), nothing special happens. If there is a value in record.specification_id that is different from the one created from this uuid and record.name, an error is returned. |
| `parties` | [Party](#provenance.metadata.v1.Party) | repeated | parties is the list of parties involved with this record. |






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
| `AddScopeOwner` | [MsgAddScopeOwnerRequest](#provenance.metadata.v1.MsgAddScopeOwnerRequest) | [MsgAddScopeOwnerResponse](#provenance.metadata.v1.MsgAddScopeOwnerResponse) | AddScopeOwner adds new owner AccAddress to scope | |
| `DeleteScopeOwner` | [MsgDeleteScopeOwnerRequest](#provenance.metadata.v1.MsgDeleteScopeOwnerRequest) | [MsgDeleteScopeOwnerResponse](#provenance.metadata.v1.MsgDeleteScopeOwnerResponse) | DeleteScopeOwner removes data access AccAddress from scope | |
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
| `WriteP8eContractSpec` | [MsgWriteP8eContractSpecRequest](#provenance.metadata.v1.MsgWriteP8eContractSpecRequest) | [MsgWriteP8eContractSpecResponse](#provenance.metadata.v1.MsgWriteP8eContractSpecResponse) | WriteP8eContractSpec adds a P8e v39 contract spec as a v40 ContractSpecification It only exists to help facilitate the transition. Users should transition to WriteContractSpecification. | |
| `P8eMemorializeContract` | [MsgP8eMemorializeContractRequest](#provenance.metadata.v1.MsgP8eMemorializeContractRequest) | [MsgP8eMemorializeContractResponse](#provenance.metadata.v1.MsgP8eMemorializeContractResponse) | P8EMemorializeContract records the results of a P8e contract execution as a session and set of records in a scope It only exists to help facilitate the transition. Users should transition to calling the individual Write methods. | |
| `BindOSLocator` | [MsgBindOSLocatorRequest](#provenance.metadata.v1.MsgBindOSLocatorRequest) | [MsgBindOSLocatorResponse](#provenance.metadata.v1.MsgBindOSLocatorResponse) | BindOSLocator binds an owner address to a uri. | |
| `DeleteOSLocator` | [MsgDeleteOSLocatorRequest](#provenance.metadata.v1.MsgDeleteOSLocatorRequest) | [MsgDeleteOSLocatorResponse](#provenance.metadata.v1.MsgDeleteOSLocatorResponse) | DeleteOSLocator deletes an existing ObjectStoreLocator record. | |
| `ModifyOSLocator` | [MsgModifyOSLocatorRequest](#provenance.metadata.v1.MsgModifyOSLocatorRequest) | [MsgModifyOSLocatorResponse](#provenance.metadata.v1.MsgModifyOSLocatorResponse) | ModifyOSLocator updates an ObjectStoreLocator record by the current owner. | |

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
it consists of two parts
1. the msg type url, i.e. /cosmos.bank.v1beta1.MsgSend
2. minimum additional fees(can be of any denom)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `msg_type_url` | [string](#string) |  |  |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | additional_fee can pay in any Coin( basically a Denom and Amount, Amount can be zero) |






<a name="provenance.msgfees.v1.Params"></a>

### Params
Params defines the set of params for the msgfees module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `floor_gas_price` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | constant used to calculate fees when gas fees shares denom with msg fee |
| `nhash_per_usd_mil` | [uint64](#uint64) |  | total nhash per usd mil for converting usd to nhash |





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
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `msg_type_url` | [string](#string) |  |  |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="provenance.msgfees.v1.RemoveMsgFeeProposal"></a>

### RemoveMsgFeeProposal
RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `msg_type_url` | [string](#string) |  |  |






<a name="provenance.msgfees.v1.UpdateMsgFeeProposal"></a>

### UpdateMsgFeeProposal
UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `msg_type_url` | [string](#string) |  |  |
| `additional_fee` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="provenance.msgfees.v1.UpdateNhashPerUsdMilProposal"></a>

### UpdateNhashPerUsdMilProposal
UpdateNhashPerUsdMilProposal defines a governance proposal to update the nhash per usd mil param


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
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



<a name="provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest"></a>

### MsgAssessCustomMsgFeeRequest
MsgAssessCustomMsgFeeRequest defines an sdk.Msg type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | optional short name for custom msg fee, this will be emitted as a property of the event |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | amount of additional fee that must be paid |
| `recipient` | [string](#string) |  | optional recipient address, the amount is split 50/50 between recipient and fee module. If |
| `from` | [string](#string) |  | empty, whole amount goes to fee module

the signer of the msg |






<a name="provenance.msgfees.v1.MsgAssessCustomMsgFeeResponse"></a>

### MsgAssessCustomMsgFeeResponse
MsgAssessCustomMsgFeeResponse defines the Msg/AssessCustomMsgFeee response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="provenance.msgfees.v1.Msg"></a>

### Msg
Msg defines the msgfees Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `AssessCustomMsgFee` | [MsgAssessCustomMsgFeeRequest](#provenance.msgfees.v1.MsgAssessCustomMsgFeeRequest) | [MsgAssessCustomMsgFeeResponse](#provenance.msgfees.v1.MsgAssessCustomMsgFeeResponse) | AssessCustomMsgFee endpoint executes the additional fee charges. This will only emit the event and not persist it to the keeper. Fees are handled with the custom msg fee handlers Use Case: smart contracts will be able to charge additional fees and direct partial funds to specified recipient for executing contracts | |

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
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `name` | [string](#string) |  |  |
| `owner` | [string](#string) |  |  |
| `restricted` | [bool](#bool) |  |  |






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






<a name="provenance.name.v1.NameRecord"></a>

### NameRecord
NameRecord is a structure used to bind ownership of a name hierarchy to a collection of addresses


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | The bound name |
| `address` | [string](#string) |  | The address the name resolved to. |
| `restricted` | [bool](#bool) |  | Whether owner signature is required to add sub-names. |






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






<a name="provenance.name.v1.MsgDeleteNameRequest"></a>

### MsgDeleteNameRequest
MsgDeleteNameRequest defines an sdk.Msg type that is used to remove an existing address/name binding.  The binding
may not have any child names currently bound for this request to be successful.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `record` | [NameRecord](#provenance.name.v1.NameRecord) |  | The record being removed |






<a name="provenance.name.v1.MsgDeleteNameResponse"></a>

### MsgDeleteNameResponse
MsgDeleteNameResponse defines the Msg/DeleteName response type.





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

