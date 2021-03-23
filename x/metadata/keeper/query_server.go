package keeper

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/google/uuid"

	"github.com/provenance-io/provenance/x/metadata/types"
)

const defaultLimit = 100

var _ types.QueryServer = Keeper{}

// ObjectStoreLocators within the GenesisState
type ObjectStoreLocators []types.ObjectStoreLocator

// Params queries params of metadata module
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Scope returns a specific scope by id
func (k Keeper) Scope(c context.Context, req *types.ScopeRequest) (*types.ScopeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ScopeId == "" {
		return nil, status.Error(codes.InvalidArgument, "scope uuid cannot be empty")
	}

	scopeUUID, err := uuid.Parse(req.ScopeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope uuid: %s", err.Error())
	}
	scopeAddress := types.ScopeMetadataAddress(scopeUUID)

	retval := types.ScopeResponse{ScopeUuid: scopeUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	scope, found := k.GetScope(ctx, scopeAddress)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "scope uuid %s not found", scopeUUID)
	}
	retval.Scope = &scope

	err = k.IterateRecords(ctx, scopeAddress, func(record types.Record) (stop bool) {
		retval.Records = append(retval.Records, &record)
		return false
	})
	if err != nil {
		return &retval, status.Errorf(codes.Internal, "error iterating scope records: %v", err)
	}

	err = k.IterateSessions(ctx, scopeAddress, func(session types.Session) (stop bool) {
		retval.Sessions = append(retval.Sessions, &session)
		return false
	})
	if err != nil {
		return &retval, status.Errorf(codes.Internal, "error iterating scope sessions: %v", err)
	}

	return &retval, nil
}

// Sessions returns sessions based on the provided request.
func (k Keeper) Sessions(c context.Context, req *types.SessionsRequest) (*types.SessionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.SessionsResponse{Request: req}

	haveScopeID := len(req.ScopeId) > 0
	haveSessionID := len(req.SessionId) > 0
	if !haveScopeID && !haveSessionID {
		return &retval, status.Error(codes.InvalidArgument, "empty request parameters")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if haveSessionID {
		sessionAddr, err := ParseSessionID(req.ScopeId, req.SessionId)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		session, found := k.GetSession(ctx, sessionAddr)
		if found {
			retval.Sessions = append(retval.Sessions, types.WrapSession(&session))
		} else {
			retval.Sessions = append(retval.Sessions, types.WrapSessionNotFound(sessionAddr))
		}
		return &retval, nil
	}

	scopeID, scopeIDErr := ParseScopeID(req.ScopeId)
	if scopeIDErr != nil {
		return &retval, scopeIDErr
	}

	itErr := k.IterateSessions(ctx, scopeID, func(s types.Session) (stop bool) {
		retval.Sessions = append(retval.Sessions, types.WrapSession(&s))
		return false
	})
	if itErr != nil {
		return &retval, status.Error(codes.Unavailable, fmt.Sprintf("error getting sessions for scope with address %s", scopeID))
	}

	return &retval, nil
}

// Records returns records based ont eh provided request.
func (k Keeper) Records(c context.Context, req *types.RecordsRequest) (*types.RecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.RecordsResponse{Request: req}
	ctx := sdk.UnwrapSDKContext(c)

	haveRecordAddr := len(req.RecordAddr) > 0
	haveScopeID := len(req.ScopeId) > 0
	haveSessionID := len(req.SessionId) > 0

	if !haveRecordAddr && !haveScopeID && !haveSessionID {
		return &retval, status.Error(codes.InvalidArgument, "empty request parameters")
	}

	if haveRecordAddr {
		recordAddr, err := ParseRecordAddr(req.RecordAddr)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		record, found := k.GetRecord(ctx, recordAddr)
		if found {
			retval.Records = append(retval.Records, types.WrapRecord(&record))
		} else {
			retval.Records = append(retval.Records, types.WrapRecordNotFound(recordAddr))
		}
		return &retval, nil
	}

	var scopeAddr, sessionAddr types.MetadataAddress

	if haveScopeID {
		var err error
		scopeAddr, err = ParseScopeID(req.ScopeId)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	if haveSessionID {
		var err error
		sessionAddr, err = ParseSessionID(req.ScopeId, req.SessionId)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		if scopeAddr.Empty() {
			scopeAddr, err = sessionAddr.AsScopeAddress()
			if err != nil {
				return &retval, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	var records []*types.Record

	if len(req.Name) > 0 {
		recordAddr, err := scopeAddr.AsRecordAddress(req.Name)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		record, found := k.GetRecord(ctx, recordAddr)
		if found {
			records = append(records, &record)
		}
	} else {
		var err error
		records, err = k.GetRecords(ctx, scopeAddr, req.Name)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	if len(records) > 0 {
		haveSessionAddr := sessionAddr.Empty()
		for _, r := range records {
			if !haveSessionAddr || sessionAddr.Equals(r.SessionId) {
				retval.Records = append(retval.Records, types.WrapRecord(r))
			}
		}
	}

	return &retval, nil
}

// Ownership returns a list of scope identifiers that list the given address as a data or value owner
func (k Keeper) Ownership(c context.Context, req *types.OwnershipRequest) (*types.OwnershipResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	scopeStore := prefix.NewStore(store, types.GetAddressScopeCacheIteratorPrefix(addr))

	scopeUUIDs := make([]string, req.Pagination.Size())
	pageRes, err := query.Paginate(scopeStore, req.Pagination, func(key, _ []byte) error {
		var ma types.MetadataAddress
		if mErr := ma.Unmarshal(key); mErr != nil {
			return mErr
		}
		scopeUUID, sErr := ma.ScopeUUID()
		if sErr != nil {
			return sErr
		}
		scopeUUIDs = append(scopeUUIDs, scopeUUID.String())
		return nil
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}
	return &types.OwnershipResponse{ScopeUuids: scopeUUIDs, Pagination: pageRes}, nil
}

// ValueOwnership returns a list of scope identifiers that list the given address as a value owner
func (k Keeper) ValueOwnership(c context.Context, req *types.ValueOwnershipRequest) (*types.ValueOwnershipResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	scopeStore := prefix.NewStore(store, types.GetValueOwnerScopeCacheIteratorPrefix(addr))

	scopes := []string{}
	pageRes, err := query.Paginate(scopeStore, req.Pagination, func(key, _ []byte) error {
		var ma types.MetadataAddress
		if mErr := ma.Unmarshal(key); mErr != nil {
			return mErr
		}
		scopeID, sErr := ma.ScopeUUID()
		if sErr != nil {
			return sErr
		}
		scopes = append(scopes, scopeID.String())
		return nil
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}
	return &types.ValueOwnershipResponse{ScopeUuids: scopes, Pagination: pageRes}, nil
}

// ScopeSpecification returns a specific scope specification by id
func (k Keeper) ScopeSpecification(c context.Context, req *types.ScopeSpecificationRequest) (*types.ScopeSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.SpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "specification uuid cannot be empty")
	}

	specUUID, err := uuid.Parse(req.SpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid specification uuid: %s", err.Error())
	}
	specID := types.ScopeSpecMetadataAddress(specUUID)

	retval := types.ScopeSpecificationResponse{SpecificationUuid: specUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetScopeSpecification(ctx, specID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "scope specification uuid %s not found", req.SpecificationUuid)
	}
	retval.ScopeSpecification = &spec

	return &retval, nil
}

// ContractSpecification returns a specific contract specification by id
func (k Keeper) ContractSpecification(c context.Context, req *types.ContractSpecificationRequest) (*types.ContractSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.SpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "specification uuid cannot be empty")
	}

	specUUID, err := uuid.Parse(req.SpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid specification uuid: %s", err.Error())
	}
	specID := types.ContractSpecMetadataAddress(specUUID)

	retval := types.ContractSpecificationResponse{ContractSpecificationUuid: specUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetContractSpecification(ctx, specID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "contract specification with id %s (uuid %s) not found",
			specID, req.SpecificationUuid)
	}
	retval.ContractSpecification = &spec

	return &retval, nil
}

// ContractSpecificationExtended returns a specific contract specification and record specifications by contract specification id
func (k Keeper) ContractSpecificationExtended(c context.Context, req *types.ContractSpecificationExtendedRequest) (*types.ContractSpecificationExtendedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.SpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "specification uuid cannot be empty")
	}

	contractSpecUUID, err := uuid.Parse(req.SpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid specification uuid: %s", err.Error())
	}
	contractSpecID := types.ContractSpecMetadataAddress(contractSpecUUID)

	retval := types.ContractSpecificationExtendedResponse{ContractSpecificationUuid: contractSpecUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	contractSpec, found := k.GetContractSpecification(ctx, contractSpecID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "contract specification with id %s (uuid %s) not found",
			contractSpecID, req.SpecificationUuid)
	}
	retval.ContractSpecification = &contractSpec

	recSpecs, err := k.GetRecordSpecificationsForContractSpecificationID(ctx, contractSpecID)
	if err != nil {
		return &retval, status.Errorf(codes.Aborted, "error getting record specifications for contract spec uuid %s: %s",
			contractSpecUUID, err.Error())
	}
	retval.RecordSpecifications = recSpecs

	return &retval, nil
}

// RecordSpecificationsForContractSpecification returns the record specifications associated with a contract specification
func (k Keeper) RecordSpecificationsForContractSpecification(
	c context.Context,
	req *types.RecordSpecificationsForContractSpecificationRequest,
) (*types.RecordSpecificationsForContractSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.ContractSpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "contract specification uuid cannot be empty")
	}
	contractSpecUUID, err := uuid.Parse(req.ContractSpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid contract specification uuid: %s", err.Error())
	}
	contractSpecID := types.ContractSpecMetadataAddress(contractSpecUUID)

	retval := types.RecordSpecificationsForContractSpecificationResponse{ContractSpecificationUuid: contractSpecUUID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	recSpecs, err := k.GetRecordSpecificationsForContractSpecificationID(ctx, contractSpecID)
	if err != nil {
		return &retval, status.Errorf(codes.Aborted, "error getting record specifications for contract spec uuid %s: %s",
			contractSpecUUID, err.Error())
	}
	if len(recSpecs) == 0 {
		return &retval, status.Errorf(codes.NotFound, "no record specifications found for contract spec uuid %s", contractSpecUUID)
	}
	retval.RecordSpecifications = recSpecs

	return &retval, err
}

// RecordSpecification returns a specific record specification by contract spec uuid and record name
func (k Keeper) RecordSpecification(c context.Context, req *types.RecordSpecificationRequest) (*types.RecordSpecificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.ContractSpecificationUuid) == 0 {
		return nil, status.Error(codes.InvalidArgument, "contract specification uuid cannot be empty")
	}
	contractSpecUUID, err := uuid.Parse(req.ContractSpecificationUuid)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid contract specification uuid: %s", err.Error())
	}

	if len(strings.TrimSpace(req.Name)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}

	recSpecID := types.RecordSpecMetadataAddress(contractSpecUUID, req.Name)

	retval := types.RecordSpecificationResponse{
		ContractSpecificationUuid: contractSpecUUID.String(),
		Name:                      req.Name,
	}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetRecordSpecification(ctx, recSpecID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "record specification not found for id %s (contract spec uuid %s and name %s)",
			recSpecID, req.ContractSpecificationUuid, req.Name)
	}
	retval.RecordSpecification = &spec

	return &retval, nil
}

// RecordSpecification returns a specific record specification by contract spec uuid and record name
func (k Keeper) RecordSpecificationByID(c context.Context, req *types.RecordSpecificationByIDRequest) (*types.RecordSpecificationByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.RecordSpecificationId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "record specification id cannot be empty")
	}

	recSpecID, err := types.MetadataAddressFromBech32(req.RecordSpecificationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid record specification id: %s", err.Error())
	}
	if !recSpecID.IsRecordSpecificationAddress() {
		return nil, status.Errorf(codes.InvalidArgument, "metadata address %s is not a record specification id", recSpecID.String())
	}

	retval := types.RecordSpecificationByIDResponse{RecordSpecificationId: recSpecID.String()}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetRecordSpecification(ctx, recSpecID)
	if !found {
		return &retval, status.Errorf(codes.NotFound, "record specification not found for id %s", recSpecID)
	}
	retval.RecordSpecification = &spec

	return &retval, nil
}

func (k Keeper) OSLocatorParams(c context.Context, request *types.OSLocatorParamsRequest) (*types.OSLocatorParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.OSLocatorParams
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.OSLocatorParamsResponse{Params: params}, nil
}

func (k Keeper) OSLocator(c context.Context, request *types.OSLocatorRequest) (*types.OSLocatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	accAddr, err := sdk.AccAddressFromBech32(request.Owner)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	record, exists := k.GetOsLocatorRecord(ctx, accAddr)

	if !exists {
		return nil, types.ErrAddressNotBound
	}
	return &types.OSLocatorResponse{Locator: &record}, nil
}

func (k Keeper) OSLocatorByURI(ctx context.Context, request *types.OSLocatorByURIRequest) (*types.OSLocatorByURIResponse, error) {
	ctxSDK := sdk.UnwrapSDKContext(ctx)
	var sDec []byte
	// rest request send in base64 encoded uri, using a URL-compatible base64
	// format.
	if IsBase64(request.Uri) {
		sDec, _ = b64.StdEncoding.DecodeString(request.Uri)
	} else {
		sDec = []byte(request.Uri)
	}
	url, err := url.Parse(string(sDec))
	if err != nil {
		return nil, err
	}
	// Return value data structure.
	var records []types.ObjectStoreLocator
	// Handler that adds records if account address matches.
	appendToRecords := func(record types.ObjectStoreLocator) bool {
		if record.LocatorUri == url.String() {
			records = append(records, record)
			// have to get all the uri associated with an address..imo..check
		}
		return false
	}

	if err := k.IterateLocators(ctxSDK, appendToRecords); err != nil {
		return nil, err
	}
	if records == nil {
		return nil, types.ErrNoRecordsFound
	}
	uniqueRecords := uniqueRecords(records)

	pageRequest := request.Pagination
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}

	limit := pageRequest.Limit
	if limit == 0 {
		limit = defaultLimit
	}
	end := pageRequest.Offset + limit
	totalResults := uint64(len(uniqueRecords))

	if pageRequest.Offset > totalResults {
		return nil, fmt.Errorf("invalid offset")
	}

	if end > totalResults {
		end = totalResults
	}

	return &types.OSLocatorByURIResponse{
		Locator:    uniqueRecords[pageRequest.Offset:end],
		Pagination: &query.PageResponse{Total: totalResults},
	}, nil
}

func (k Keeper) OSLocatorByScopeUUID(ctx context.Context, request *types.OSLocatorByScopeUUIDRequest) (*types.OSLocatorByScopeUUIDResponse, error) {
	ctxSDK := sdk.UnwrapSDKContext(ctx)
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if request.ScopeUuid == "" {
		return nil, status.Error(codes.InvalidArgument, "scope uuid cannot be empty")
	}

	return k.GetOSLocatorByScopeUUID(ctxSDK, request.ScopeUuid)
}

func (k Keeper) OSAllLocators(ctx context.Context, request *types.OSAllLocatorsRequest) (*types.OSAllLocatorsResponse, error) {
	ctxSDK := sdk.UnwrapSDKContext(ctx)

	// Return value data structure.
	var records []types.ObjectStoreLocator
	// Handler that adds records if account address matches.
	appendToRecords := func(record types.ObjectStoreLocator) bool {
		records = append(records, record)
		// have to get all the uri associated with an address..imo..check
		return false
	}

	if err := k.IterateLocators(ctxSDK, appendToRecords); err != nil {
		return nil, err
	}
	if records == nil {
		return nil, types.ErrNoRecordsFound
	}
	uniqueRecords := uniqueRecords(records)

	pageRequest := request.Pagination
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}

	limit := pageRequest.Limit
	if limit == 0 {
		limit = defaultLimit
	}
	end := pageRequest.Offset + limit
	totalResults := uint64(len(uniqueRecords))

	if pageRequest.Offset > totalResults {
		return nil, fmt.Errorf("invalid offset")
	}

	if end > totalResults {
		end = totalResults
	}

	return &types.OSAllLocatorsResponse{
		Locator:    uniqueRecords[pageRequest.Offset:end],
		Pagination: &query.PageResponse{Total: totalResults},
	}, nil
}

func IsBase64(s string) bool {
	_, err := b64.StdEncoding.DecodeString(s)
	return err == nil
}

// ParseScopeID parses the provided input into a scope MetadataAddress.
// The input can either be a uuid string or scope address bech32 string.
func ParseScopeID(scopeID string) (types.MetadataAddress, error) {
	addr, addrErr := types.MetadataAddressFromBech32(scopeID)
	if addrErr == nil {
		if addr.IsScopeAddress() {
			return addr, nil
		}
		return types.MetadataAddress{}, fmt.Errorf("address [%s] is not a scope address", scopeID)
	}
	uid, uidErr := uuid.Parse(scopeID)
	if uidErr == nil {
		return types.ScopeMetadataAddress(uid), nil
	}
	return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into either a scope address (%s) or uuid (%s)",
		scopeID, addrErr, uidErr)
}

// ParseSessionID parses the provided input into a session MetadataAddress.
// The scopeID field can be either a uuid or scope address bech32 string.
// The sessionID field can be either a uuid or session address bech32 string.
// If the sessionID field is a bech32 address, the scopeID field is ignored.
// Otherwise, the scope id field is parsed using ParseScopeID and converted to a session MetadataAddress using the uuid in the sessionID field.
func ParseSessionID(scopeID string, sessionID string) (types.MetadataAddress, error) {
	sessionAddr, sessionAddrErr := types.MetadataAddressFromBech32(sessionID)
	if sessionAddrErr == nil {
		if sessionAddr.IsSessionAddress() {
			return sessionAddr, nil
		}
		return types.MetadataAddress{}, fmt.Errorf("address [%s] is not a session address", sessionID)
	} else if len(scopeID) == 0 {
		return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into a session address: %w", sessionID, sessionAddrErr)
	}
	scopeAddr, scopeAddrErr := ParseScopeID(scopeID)
	if scopeAddrErr != nil {
		return types.MetadataAddress{}, scopeAddrErr
	}
	sessionUUID, sessionUUIDErr := uuid.Parse(sessionID)
	if sessionUUIDErr == nil {
		return scopeAddr.AsSessionAddress(sessionUUID)
	}
	return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into either a session address (%s) or uuid (%s)",
		sessionID, sessionAddrErr, sessionUUIDErr)
}

// ParseRecordAddr parses the provided input into a scope MetadataAddress.
// The input must be a record address bech32 string.
func ParseRecordAddr(recordAddr string) (types.MetadataAddress, error) {
	addr, addrErr := types.MetadataAddressFromBech32(recordAddr)
	if addrErr == nil {
		if addr.IsRecordAddress() {
			return addr, nil
		}
		return types.MetadataAddress{}, fmt.Errorf("address [%s] is not a record address", recordAddr)
	}
	return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into a record address: %w", recordAddr, addrErr)
}
