package keeper

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/google/uuid"

	"github.com/provenance-io/provenance/x/metadata/types"
)

const defaultLimit = 100

var _ types.QueryServer = Keeper{}

// Params queries params of metadata module.
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "Params")
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.QueryParamsResponse{Params: params, Request: req}, nil
}

// Scope returns a specific scope by id.
func (k Keeper) Scope(c context.Context, req *types.ScopeRequest) (*types.ScopeResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "Scope")
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.ScopeResponse{Request: req}

	var scopeAddr, sessionAddr types.MetadataAddress
	if len(req.ScopeId) > 0 {
		var err error
		scopeAddr, err = ParseScopeID(req.ScopeId)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	if len(req.SessionAddr) > 0 {
		var err error
		sessionAddr, err = ParseSessionAddr(req.SessionAddr)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		// ParseSessionAddr if this would fail.
		scopeAddr2 := sessionAddr.MustGetAsScopeAddress()
		if scopeAddr.Empty() {
			scopeAddr = scopeAddr2
		} else if !scopeAddr.Equals(scopeAddr2) {
			return &retval, status.Errorf(codes.InvalidArgument, "session %s is not in scope %s", sessionAddr, scopeAddr)
		}
	}
	if len(req.RecordAddr) > 0 {
		recordAddr, err := ParseRecordAddr(req.RecordAddr)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		// ParseRecordAddr if this would fail.
		scopeAddr2 := recordAddr.MustGetAsScopeAddress()
		switch {
		case !sessionAddr.Empty():
			// This assumes that we have checked and set scopeAddr while processing the sessionAddr.
			scopeAddr3 := sessionAddr.MustGetAsScopeAddress()
			if !scopeAddr2.Equals(scopeAddr3) {
				return &retval, status.Errorf(codes.InvalidArgument, "session %s and record %s are not associated with the same scope", sessionAddr, recordAddr)
			}
		case scopeAddr.Empty():
			scopeAddr = scopeAddr2
		case !scopeAddr.Equals(scopeAddr2):
			return &retval, status.Errorf(codes.InvalidArgument, "session %s is not part of scope %s", recordAddr, scopeAddr)
		}
	}

	if scopeAddr.Empty() {
		return &retval, status.Error(codes.InvalidArgument, "empty request parameters")
	}

	ctx := sdk.UnwrapSDKContext(c)
	scope, found := k.GetScope(ctx, scopeAddr)
	if found {
		retval.Scope = types.WrapScope(&scope)
	} else {
		retval.Scope = types.WrapScopeNotFound(scopeAddr)
	}

	var sessErr, recErr error

	if req.IncludeSessions {
		err := k.IterateSessions(ctx, scopeAddr, func(session types.Session) (stop bool) {
			retval.Sessions = append(retval.Sessions, types.WrapSession(&session))
			return false
		})
		if err != nil {
			sessErr = fmt.Errorf("error iterating scope [%s] sessions: %w", scopeAddr, err)
		}
	}

	if req.IncludeRecords {
		err := k.IterateRecords(ctx, scopeAddr, func(record types.Record) (stop bool) {
			retval.Records = append(retval.Records, types.WrapRecord(&record))
			return false
		})
		if err != nil {
			recErr = fmt.Errorf("error iterating scope [%s] records: %w", scopeAddr, err)
		}
	}

	var err error
	switch {
	case sessErr != nil && recErr != nil:
		err = fmt.Errorf("errors getting sessions and records: %s, %s", sessErr.Error(), recErr.Error())
	case sessErr != nil:
		err = sessErr
	case recErr != nil:
		err = recErr
	}

	if err != nil {
		return &retval, status.Error(codes.Unavailable, err.Error())
	}

	return &retval, nil
}

// ScopesAll returns all scopes (limited by pagination).
func (k Keeper) ScopesAll(c context.Context, req *types.ScopesAllRequest) (*types.ScopesAllResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "ScopesAll")
	retval := types.ScopesAllResponse{Request: req}

	pageRequest := getPageRequest(req)

	ctx := sdk.UnwrapSDKContext(c)
	kvStore := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.ScopeKeyPrefix)

	pageRes, err := query.Paginate(prefixStore, pageRequest, func(key, value []byte) error {
		var scope types.Scope
		vErr := scope.Unmarshal(value)
		if vErr == nil {
			retval.Scopes = append(retval.Scopes, types.WrapScope(&scope))
			return nil
		}
		// Something's wrong. Let's do what we can to give indications of it.
		var addr types.MetadataAddress
		kErr := addr.Unmarshal(key)
		if kErr == nil {
			k.Logger(ctx).Error("failed to unmarshal scope", "address", addr, "error", vErr)
			retval.Scopes = append(retval.Scopes, types.WrapScopeNotFound(addr))
		} else {
			k64 := b64.StdEncoding.EncodeToString(key)
			k.Logger(ctx).Error("failed to unmarshal scope key and value",
				"key error", kErr, "value error", vErr, "key (base64)", k64)
			retval.Scopes = append(retval.Scopes, &types.ScopeWrapper{})
		}
		return nil // Still want to move on to the next.
	})
	if err != nil {
		return &retval, status.Error(codes.Unavailable, err.Error())
	}
	retval.Pagination = pageRes
	return &retval, nil
}

// Sessions returns sessions based on the provided request.
func (k Keeper) Sessions(c context.Context, req *types.SessionsRequest) (*types.SessionsResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "Sessions")
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.SessionsResponse{Request: req}

	ctx := sdk.UnwrapSDKContext(c)

	var scopeAddr, sessionAddr, recordAddr types.MetadataAddress

	if len(req.ScopeId) > 0 {
		var err error
		scopeAddr, err = ParseScopeID(req.ScopeId)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	if len(req.RecordAddr) > 0 {
		var err error
		recordAddr, err = ParseRecordAddr(req.RecordAddr)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		scopeAddr2 := recordAddr.MustGetAsScopeAddress()
		if scopeAddr.Empty() {
			scopeAddr = scopeAddr2
		} else if !scopeAddr.Equals(scopeAddr2) {
			return &retval, status.Errorf(codes.InvalidArgument, "record %s is not part of scope %s", recordAddr, scopeAddr)
		}
	}
	if len(req.SessionId) > 0 {
		var err error
		scopeIDForParsing := req.ScopeId
		if len(scopeIDForParsing) == 0 && !scopeAddr.Empty() {
			scopeIDForParsing = scopeAddr.String()
		}
		sessionAddr, err = ParseSessionID(scopeIDForParsing, req.SessionId)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		// ParseSessionID ensures that this will not return an error.
		scopeAddr2 := sessionAddr.MustGetAsScopeAddress()
		switch {
		case !recordAddr.Empty():
			// This assumes that we have checked and set scopeAddr while processing the recordAddr.
			scopeAddr3 := recordAddr.MustGetAsScopeAddress()
			if !scopeAddr2.Equals(scopeAddr3) {
				return &retval, status.Errorf(codes.InvalidArgument, "session %s and record %s are not associated with the same scope", sessionAddr, recordAddr)
			}
		case scopeAddr.Empty():
			scopeAddr = scopeAddr2
		case !scopeAddr.Equals(scopeAddr2):
			return &retval, status.Errorf(codes.InvalidArgument, "session %s is not part of scope %s", recordAddr, scopeAddr)
		}
	}
	if len(req.RecordName) > 0 {
		if scopeAddr.Empty() {
			// assumes scopeAddr is set previously while parsing other input.
			return &retval, status.Error(codes.InvalidArgument, "a scope is required to look up sessions by record name")
		}
		// We know that scopeAddr is legit, and that we have a name. So this won't give an error.
		recordAddr2 := scopeAddr.MustGetAsRecordAddress(req.RecordName)
		if recordAddr.Empty() {
			recordAddr = recordAddr2
		} else if !recordAddr.Equals(recordAddr2) {
			return &retval, status.Errorf(codes.InvalidArgument, "record %s does not have name %s", recordAddr, req.RecordName)
		}
	}

	// If a record was identified in the search, we need to get it and either use it to set the sessionAddr,
	// or make sure the provided sessionAddr matches what the record has.
	if !recordAddr.Empty() {
		record, found := k.GetRecord(ctx, recordAddr)
		switch {
		case !found:
			return &retval, status.Errorf(codes.InvalidArgument, "record %s does not exist", recordAddr)
		case !sessionAddr.Empty():
			if !sessionAddr.Equals(record.SessionId) {
				return &retval, status.Errorf(codes.InvalidArgument, "record %s belongs to session %s (not %s)",
					recordAddr, record.SessionId, sessionAddr)
			}
		default:
			sessionAddr = record.SessionId
		}
	}

	// Get all the sessions based on the input, and set things up for extra info.
	switch {
	case !sessionAddr.Empty():
		session, found := k.GetSession(ctx, sessionAddr)
		if found {
			retval.Sessions = append(retval.Sessions, types.WrapSession(&session))
		} else {
			retval.Sessions = append(retval.Sessions, types.WrapSessionNotFound(sessionAddr))
		}
	case !scopeAddr.Empty():
		itErr := k.IterateSessions(ctx, scopeAddr, func(s types.Session) (stop bool) {
			retval.Sessions = append(retval.Sessions, types.WrapSession(&s))
			return false
		})
		if itErr != nil {
			return &retval, status.Error(codes.Unavailable, fmt.Sprintf("error getting sessions for scope with address %s", scopeAddr))
		}
	default:
		return &retval, status.Error(codes.InvalidArgument, "empty request parameters")
	}

	if req.IncludeScope {
		scope, found := k.GetScope(ctx, scopeAddr)
		if found {
			retval.Scope = types.WrapScope(&scope)
		} else {
			retval.Scope = types.WrapScopeNotFound(scopeAddr)
		}
	}

	if req.IncludeRecords {
		// Get all the session ids
		sessionAddrs := []types.MetadataAddress{}
		for _, s := range retval.Sessions {
			if s.Session != nil {
				sessionAddrs = append(sessionAddrs, s.Session.SessionId)
			}
		}
		// Iterate the records for the whole scope, and just keep the ones for our sessions.
		err := k.IterateRecords(ctx, scopeAddr, func(r types.Record) (stop bool) {
			keep := false
			for _, a := range sessionAddrs {
				if r.SessionId.Equals(a) {
					keep = true
					break
				}
			}
			if keep {
				retval.Records = append(retval.Records, types.WrapRecord(&r))
			}
			return false
		})
		if err != nil {
			return &retval, status.Errorf(codes.Unavailable, "error iterating scope [%s] records: %s", scopeAddr, err.Error())
		}
	}

	return &retval, nil
}

// SessionsAll returns all sessions (limited by pagination).
func (k Keeper) SessionsAll(c context.Context, req *types.SessionsAllRequest) (*types.SessionsAllResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "SessionsAll")
	retval := types.SessionsAllResponse{Request: req}

	pageRequest := getPageRequest(req)

	ctx := sdk.UnwrapSDKContext(c)
	kvStore := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.SessionKeyPrefix)

	pageRes, err := query.Paginate(prefixStore, pageRequest, func(key, value []byte) error {
		var session types.Session
		vErr := session.Unmarshal(value)
		if vErr == nil {
			retval.Sessions = append(retval.Sessions, types.WrapSession(&session))
			return nil
		}
		// Something's wrong. Let's do what we can to give indications of it.
		var addr types.MetadataAddress
		kErr := addr.Unmarshal(key)
		if kErr == nil {
			k.Logger(ctx).Error("failed to unmarshal session", "address", addr, "error", vErr)
			retval.Sessions = append(retval.Sessions, types.WrapSessionNotFound(addr))
		} else {
			k64 := b64.StdEncoding.EncodeToString(key)
			k.Logger(ctx).Error("failed to unmarshal session key and value",
				"key error", kErr, "value error", vErr, "key (base64)", k64)
			retval.Sessions = append(retval.Sessions, &types.SessionWrapper{})
		}
		return nil // Still want to move on to the next.
	})
	if err != nil {
		return &retval, status.Error(codes.Unavailable, err.Error())
	}
	retval.Pagination = pageRes
	return &retval, nil
}

// Records returns records based on the provided request.
func (k Keeper) Records(c context.Context, req *types.RecordsRequest) (*types.RecordsResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "Records")
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.RecordsResponse{Request: req}
	ctx := sdk.UnwrapSDKContext(c)

	var scopeAddr, sessionAddr, recordAddr types.MetadataAddress

	if len(req.ScopeId) > 0 {
		var err error
		scopeAddr, err = ParseScopeID(req.ScopeId)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	if len(req.RecordAddr) > 0 {
		var err error
		recordAddr, err = ParseRecordAddr(req.RecordAddr)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		scopeAddr2 := recordAddr.MustGetAsScopeAddress()
		if scopeAddr.Empty() {
			scopeAddr = scopeAddr2
		} else if !scopeAddr.Equals(scopeAddr2) {
			return &retval, status.Errorf(codes.InvalidArgument, "record %s is not part of scope %s", recordAddr, scopeAddr)
		}
	}
	if len(req.SessionId) > 0 {
		var err error
		sessionAddr, err = ParseSessionID(req.ScopeId, req.SessionId)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		// ParseSessionID ensures that this will not return an error.
		scopeAddr2 := sessionAddr.MustGetAsScopeAddress()
		switch {
		case !recordAddr.Empty():
			// This assumes that we have checked and set scopeAddr while processing the recordAddr.
			scopeAddr3 := recordAddr.MustGetAsScopeAddress()
			if !scopeAddr2.Equals(scopeAddr3) {
				return &retval, status.Errorf(codes.InvalidArgument, "session %s and record %s are not associated with the same scope", sessionAddr, recordAddr)
			}
		case scopeAddr.Empty():
			scopeAddr = scopeAddr2
		case !scopeAddr.Equals(scopeAddr2):
			return &retval, status.Errorf(codes.InvalidArgument, "session %s is not part of scope %s", recordAddr, scopeAddr)
		}
	}
	if len(req.Name) > 0 {
		if scopeAddr.Empty() {
			// assumes scopeAddr is set previously while parsing other input.
			return &retval, status.Error(codes.InvalidArgument, "a scope or session is required to look up records by name")
		}
		// We know that scopeAddr is legit, and that we have a name. So this won't give an error.
		recordAddr2 := scopeAddr.MustGetAsRecordAddress(req.Name)
		if recordAddr.Empty() {
			recordAddr = recordAddr2
		} else if !recordAddr.Equals(recordAddr2) {
			return &retval, status.Errorf(codes.InvalidArgument, "record %s does not have name %s", recordAddr, req.Name)
		}
	}

	// Get all the records based on the input, and set things up for extra info.
	switch {
	case !recordAddr.Empty():
		record, found := k.GetRecord(ctx, recordAddr)
		if found {
			retval.Records = append(retval.Records, types.WrapRecord(&record))
		} else {
			retval.Records = append(retval.Records, types.WrapRecordNotFound(recordAddr))
		}
	case !scopeAddr.Empty():
		var records []*types.Record
		// Get all the records for the scope (and thin them out later if needed).
		var err error
		records, err = k.GetRecords(ctx, scopeAddr, req.Name)
		if err != nil {
			return &retval, status.Error(codes.InvalidArgument, err.Error())
		}
		// Wrap (and possibly filter) the records and add them to the return value.
		if len(records) > 0 {
			haveSessionAddr := !sessionAddr.Empty()
			for _, r := range records {
				if !haveSessionAddr || sessionAddr.Equals(r.SessionId) {
					retval.Records = append(retval.Records, types.WrapRecord(r))
				}
			}
		}
	default:
		return &retval, status.Error(codes.InvalidArgument, "empty request parameters")
	}

	if req.IncludeScope {
		scope, found := k.GetScope(ctx, scopeAddr)
		if found {
			retval.Scope = types.WrapScope(&scope)
		} else {
			retval.Scope = types.WrapScopeNotFound(scopeAddr)
		}
	}

	if req.IncludeSessions {
		// Get a list of unique session addresses from the records.
		sessionAddrs := []types.MetadataAddress{}
		for _, r := range retval.Records {
			if r.Record != nil {
				alreadyHave := false
				for _, a := range sessionAddrs {
					if r.Record.SessionId.Equals(a) {
						alreadyHave = true
						break
					}
				}
				if !alreadyHave {
					sessionAddrs = append(sessionAddrs, r.Record.SessionId)
				}
			}
		}
		// Get each session.
		for _, a := range sessionAddrs {
			session, found := k.GetSession(ctx, a)
			if found {
				retval.Sessions = append(retval.Sessions, types.WrapSession(&session))
			} else {
				retval.Sessions = append(retval.Sessions, types.WrapSessionNotFound(a))
			}
		}
	}

	return &retval, nil
}

// RecordsAll returns all records (limited by pagination).
func (k Keeper) RecordsAll(c context.Context, req *types.RecordsAllRequest) (*types.RecordsAllResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "RecordsAll")
	retval := types.RecordsAllResponse{Request: req}

	pageRequest := getPageRequest(req)

	ctx := sdk.UnwrapSDKContext(c)
	kvStore := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.RecordKeyPrefix)

	pageRes, err := query.Paginate(prefixStore, pageRequest, func(key, value []byte) error {
		var record types.Record
		vErr := record.Unmarshal(value)
		if vErr == nil {
			retval.Records = append(retval.Records, types.WrapRecord(&record))
			return nil
		}
		// Something's wrong. Let's do what we can to give indications of it.
		var addr types.MetadataAddress
		kErr := addr.Unmarshal(key)
		if kErr == nil {
			k.Logger(ctx).Error("failed to unmarshal record", "address", addr, "error", vErr)
			retval.Records = append(retval.Records, types.WrapRecordNotFound(addr))
		} else {
			k64 := b64.StdEncoding.EncodeToString(key)
			k.Logger(ctx).Error("failed to unmarshal record key and value",
				"key error", kErr, "value error", vErr, "key (base64)", k64)
			retval.Records = append(retval.Records, &types.RecordWrapper{})
		}
		return nil // Still want to move on to the next.
	})
	if err != nil {
		return &retval, status.Error(codes.Unavailable, err.Error())
	}
	retval.Pagination = pageRes
	return &retval, nil
}

// Ownership returns a list of scope identifiers that list the given address as a data or value owner.
func (k Keeper) Ownership(c context.Context, req *types.OwnershipRequest) (*types.OwnershipResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "Ownership")
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.OwnershipResponse{Request: req}

	if req.Address == "" {
		return &retval, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return &retval, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	scopeStore := prefix.NewStore(store, types.GetAddressScopeCacheIteratorPrefix(addr))

	pageRes, err := query.Paginate(scopeStore, req.Pagination, func(key, _ []byte) error {
		var ma types.MetadataAddress
		if mErr := ma.Unmarshal(key); mErr != nil {
			return mErr
		}
		scopeUUID, sErr := ma.ScopeUUID()
		if sErr != nil {
			return sErr
		}
		retval.ScopeUuids = append(retval.ScopeUuids, scopeUUID.String())
		return nil
	})
	if err != nil {
		return &retval, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}
	retval.Pagination = pageRes

	return &retval, nil
}

// ValueOwnership returns a list of scope identifiers that list the given address as a value owner.
func (k Keeper) ValueOwnership(c context.Context, req *types.ValueOwnershipRequest) (*types.ValueOwnershipResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "ValueOwnership")
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.ValueOwnershipResponse{Request: req}

	if req.Address == "" {
		return &retval, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return &retval, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	scopeStore := prefix.NewStore(store, types.GetValueOwnerScopeCacheIteratorPrefix(addr))

	pageRes, err := query.Paginate(scopeStore, req.Pagination, func(key, _ []byte) error {
		var ma types.MetadataAddress
		if mErr := ma.Unmarshal(key); mErr != nil {
			return mErr
		}
		scopeID, sErr := ma.ScopeUUID()
		if sErr != nil {
			return sErr
		}
		retval.ScopeUuids = append(retval.ScopeUuids, scopeID.String())
		return nil
	})
	if err != nil {
		return &retval, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}
	retval.Pagination = pageRes
	return &retval, nil
}

// ScopeSpecification returns a specific scope specification by id.
func (k Keeper) ScopeSpecification(c context.Context, req *types.ScopeSpecificationRequest) (*types.ScopeSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "ScopeSpecification")
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.ScopeSpecificationResponse{Request: req}

	if len(req.SpecificationId) == 0 {
		return &retval, status.Error(codes.InvalidArgument, "specification id cannot be empty")
	}

	specAddr, err := ParseScopeSpecID(req.SpecificationId)
	if err != nil {
		return &retval, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetScopeSpecification(ctx, specAddr)
	if found {
		retval.ScopeSpecification = types.WrapScopeSpec(&spec)
	} else {
		retval.ScopeSpecification = types.WrapScopeSpecNotFound(specAddr)
	}

	return &retval, nil
}

// ScopeSpecificationsAll returns all scope specifications (limited by pagination).
func (k Keeper) ScopeSpecificationsAll(c context.Context, req *types.ScopeSpecificationsAllRequest) (*types.ScopeSpecificationsAllResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "ScopeSpecificationsAll")
	retval := types.ScopeSpecificationsAllResponse{Request: req}

	pageRequest := getPageRequest(req)

	ctx := sdk.UnwrapSDKContext(c)
	kvStore := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.ScopeSpecificationKeyPrefix)

	pageRes, err := query.Paginate(prefixStore, pageRequest, func(key, value []byte) error {
		var scopeSpec types.ScopeSpecification
		vErr := scopeSpec.Unmarshal(value)
		if vErr == nil {
			retval.ScopeSpecifications = append(retval.ScopeSpecifications, types.WrapScopeSpec(&scopeSpec))
			return nil
		}
		// Something's wrong. Let's do what we can to give indications of it.
		var addr types.MetadataAddress
		kErr := addr.Unmarshal(key)
		if kErr == nil {
			k.Logger(ctx).Error("failed to unmarshal scope spec", "address", addr, "error", vErr)
			retval.ScopeSpecifications = append(retval.ScopeSpecifications, types.WrapScopeSpecNotFound(addr))
		} else {
			k64 := b64.StdEncoding.EncodeToString(key)
			k.Logger(ctx).Error("failed to unmarshal scope spec key and value",
				"key error", kErr, "value error", vErr, "key (base64)", k64)
			retval.ScopeSpecifications = append(retval.ScopeSpecifications, &types.ScopeSpecificationWrapper{})
		}
		return nil // Still want to move on to the next.
	})
	if err != nil {
		return &retval, status.Error(codes.Unavailable, err.Error())
	}
	retval.Pagination = pageRes
	return &retval, nil
}

// ContractSpecification returns a specific contract specification by id.
func (k Keeper) ContractSpecification(c context.Context, req *types.ContractSpecificationRequest) (*types.ContractSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "ContractSpecification")
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.ContractSpecificationResponse{Request: req}

	if len(req.SpecificationId) == 0 {
		return &retval, status.Error(codes.InvalidArgument, "specification id cannot be empty")
	}

	specAddr, addrErr := ParseContractSpecID(req.SpecificationId)
	if addrErr != nil {
		return &retval, status.Errorf(codes.InvalidArgument, "invalid specification id: %s", addrErr.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetContractSpecification(ctx, specAddr)
	if found {
		retval.ContractSpecification = types.WrapContractSpec(&spec)
	} else {
		retval.ContractSpecification = types.WrapContractSpecNotFound(specAddr)
	}

	if req.IncludeRecordSpecs {
		recSpecs, err := k.GetRecordSpecificationsForContractSpecificationID(ctx, specAddr)
		if err != nil {
			return &retval, status.Errorf(codes.Unavailable, "error getting record specifications for contract spec %s: %s",
				specAddr, err.Error())
		}
		retval.RecordSpecifications = types.WrapRecordSpecs(recSpecs)
	}

	return &retval, nil
}

// ContractSpecificationsAll returns all contract specifications (limited by pagination).
func (k Keeper) ContractSpecificationsAll(c context.Context, req *types.ContractSpecificationsAllRequest) (*types.ContractSpecificationsAllResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "ContractSpecificationsAll")
	retval := types.ContractSpecificationsAllResponse{Request: req}

	pageRequest := getPageRequest(req)

	ctx := sdk.UnwrapSDKContext(c)
	kvStore := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.ContractSpecificationKeyPrefix)

	pageRes, err := query.Paginate(prefixStore, pageRequest, func(key, value []byte) error {
		var contractSpec types.ContractSpecification
		vErr := contractSpec.Unmarshal(value)
		if vErr == nil {
			retval.ContractSpecifications = append(retval.ContractSpecifications, types.WrapContractSpec(&contractSpec))
			return nil
		}
		// Something's wrong. Let's do what we can to give indications of it.
		var addr types.MetadataAddress
		kErr := addr.Unmarshal(key)
		if kErr == nil {
			k.Logger(ctx).Error("failed to unmarshal contract spec", "address", addr, "error", vErr)
			retval.ContractSpecifications = append(retval.ContractSpecifications, types.WrapContractSpecNotFound(addr))
		} else {
			k64 := b64.StdEncoding.EncodeToString(key)
			k.Logger(ctx).Error("failed to unmarshal contract spec key and value",
				"key error", kErr, "value error", vErr, "key (base64)", k64)
			retval.ContractSpecifications = append(retval.ContractSpecifications, &types.ContractSpecificationWrapper{})
		}
		return nil // Still want to move on to the next.
	})
	if err != nil {
		return &retval, status.Error(codes.Unavailable, err.Error())
	}
	retval.Pagination = pageRes
	return &retval, nil
}

// RecordSpecificationsForContractSpecification returns the record specifications associated with a contract specification.
func (k Keeper) RecordSpecificationsForContractSpecification(
	c context.Context,
	req *types.RecordSpecificationsForContractSpecificationRequest,
) (*types.RecordSpecificationsForContractSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "RecordSpecificationsForContractSpecification")
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.RecordSpecificationsForContractSpecificationResponse{Request: req}

	if len(req.SpecificationId) == 0 {
		return &retval, status.Error(codes.InvalidArgument, "contract specification id cannot be empty")
	}
	contractSpecAddr, cSpecAddrErr := ParseContractSpecID(req.SpecificationId)
	if cSpecAddrErr != nil {
		return &retval, status.Errorf(codes.InvalidArgument, "invalid specification id: %s", cSpecAddrErr.Error())
	}
	contractSpecUUID, cSpecUUIDErr := contractSpecAddr.ContractSpecUUID()
	if cSpecUUIDErr != nil {
		return &retval, status.Errorf(codes.InvalidArgument, "could not extract contract spec uuid: %s", cSpecUUIDErr.Error())
	}

	retval.ContractSpecificationAddr = contractSpecAddr.String()
	retval.ContractSpecificationUuid = contractSpecUUID.String()

	ctx := sdk.UnwrapSDKContext(c)
	recSpecs, err := k.GetRecordSpecificationsForContractSpecificationID(ctx, contractSpecAddr)
	if err != nil {
		return &retval, status.Errorf(codes.Unavailable, "error getting record specifications for contract spec %s: %s",
			contractSpecAddr, err.Error())
	}

	retval.RecordSpecifications = types.WrapRecordSpecs(recSpecs)

	return &retval, err
}

// RecordSpecification returns a specific record specification.
func (k Keeper) RecordSpecification(c context.Context, req *types.RecordSpecificationRequest) (*types.RecordSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "RecordSpecification")
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.RecordSpecificationResponse{Request: req}

	if len(req.SpecificationId) == 0 {
		return &retval, status.Error(codes.InvalidArgument, "specification id cannot be empty")
	}

	recSpecAddr, recSpecAddrErr := ParseRecordSpecID(req.SpecificationId, req.Name)
	if recSpecAddrErr != nil {
		return &retval, status.Errorf(codes.InvalidArgument, "invalid input: %s", recSpecAddrErr.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	spec, found := k.GetRecordSpecification(ctx, recSpecAddr)
	if found {
		retval.RecordSpecification = types.WrapRecordSpec(&spec)
	} else {
		retval.RecordSpecification = types.WrapRecordSpecNotFound(recSpecAddr)
	}

	return &retval, nil
}

// RecordSpecificationsAll returns all record specifications (limited by pagination).
func (k Keeper) RecordSpecificationsAll(c context.Context, req *types.RecordSpecificationsAllRequest) (*types.RecordSpecificationsAllResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "RecordSpecificationsAll")
	retval := types.RecordSpecificationsAllResponse{Request: req}

	pageRequest := getPageRequest(req)

	ctx := sdk.UnwrapSDKContext(c)
	kvStore := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(kvStore, types.RecordSpecificationKeyPrefix)

	pageRes, err := query.Paginate(prefixStore, pageRequest, func(key, value []byte) error {
		var recordSpec types.RecordSpecification
		vErr := recordSpec.Unmarshal(value)
		if vErr == nil {
			retval.RecordSpecifications = append(retval.RecordSpecifications, types.WrapRecordSpec(&recordSpec))
			return nil
		}
		// Something's wrong. Let's do what we can to give indications of it.
		var addr types.MetadataAddress
		kErr := addr.Unmarshal(key)
		if kErr == nil {
			k.Logger(ctx).Error("failed to unmarshal record spec", "address", addr, "error", vErr)
			retval.RecordSpecifications = append(retval.RecordSpecifications, types.WrapRecordSpecNotFound(addr))
		} else {
			k64 := b64.StdEncoding.EncodeToString(key)
			k.Logger(ctx).Error("failed to unmarshal record spec key and value",
				"key error", kErr, "value error", vErr, "key (base64)", k64)
			retval.RecordSpecifications = append(retval.RecordSpecifications, &types.RecordSpecificationWrapper{})
		}
		return nil // Still want to move on to the next.
	})
	if err != nil {
		return &retval, status.Error(codes.Unavailable, err.Error())
	}
	retval.Pagination = pageRes
	return &retval, nil
}

func (k Keeper) OSLocatorParams(c context.Context, request *types.OSLocatorParamsRequest) (*types.OSLocatorParamsResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "OSLocatorParams")
	ctx := sdk.UnwrapSDKContext(c)
	var params types.OSLocatorParams
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.OSLocatorParamsResponse{Params: params, Request: request}, nil
}

func (k Keeper) OSLocator(c context.Context, request *types.OSLocatorRequest) (*types.OSLocatorResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "OSLocator")
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.OSLocatorResponse{Request: request}

	ctx := sdk.UnwrapSDKContext(c)
	accAddr, err := sdk.AccAddressFromBech32(request.Owner)
	if err != nil {
		return &retval, types.ErrInvalidAddress
	}

	record, exists := k.GetOsLocatorRecord(ctx, accAddr)
	if !exists {
		return &retval, types.ErrAddressNotBound
	}
	retval.Locator = &record

	return &retval, nil
}

func (k Keeper) OSLocatorsByURI(ctx context.Context, request *types.OSLocatorsByURIRequest) (*types.OSLocatorsByURIResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "OSLocatorsByURI")
	retval := types.OSLocatorsByURIResponse{Request: request}
	if request == nil {
		return &retval, status.Error(codes.InvalidArgument, "empty request")
	}

	var sDec []byte
	// rest request send in base64 encoded uri, using a URL-compatible base64 format.
	if IsBase64(request.Uri) {
		sDec, _ = b64.StdEncoding.DecodeString(request.Uri)
	} else {
		sDec = []byte(request.Uri)
	}
	uri, err := url.Parse(string(sDec))
	if err != nil {
		return &retval, err
	}
	uriStr := uri.String()

	osLocatorStore := prefix.NewStore(sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey), types.OSLocatorAddressKeyPrefix)
	retval.Pagination, err = query.FilteredPaginate(osLocatorStore, request.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		record := types.ObjectStoreLocator{}
		if rerr := k.cdc.Unmarshal(value, &record); rerr != nil {
			return false, rerr
		}
		if record.LocatorUri != uriStr {
			return false, nil
		}
		if accumulate {
			retval.Locators = append(retval.Locators, record)
		}
		return true, nil
	})
	if err != nil {
		return &retval, err
	}
	if len(retval.Locators) == 0 {
		return &retval, types.ErrNoRecordsFound
	}
	return &retval, nil
}

func (k Keeper) OSLocatorsByScope(ctx context.Context, request *types.OSLocatorsByScopeRequest) (*types.OSLocatorsByScopeResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "OSLocatorsByScope")
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	retval := types.OSLocatorsByScopeResponse{Request: request}

	ctxSDK := sdk.UnwrapSDKContext(ctx)
	if request.ScopeId == "" {
		return &retval, status.Error(codes.InvalidArgument, "scope id cannot be empty")
	}

	locators, err := k.GetOSLocatorByScope(ctxSDK, request.ScopeId)
	if err != nil {
		return &retval, status.Error(codes.InvalidArgument, err.Error())
	}
	retval.Locators = locators

	return &retval, nil
}

func (k Keeper) OSAllLocators(ctx context.Context, request *types.OSAllLocatorsRequest) (*types.OSAllLocatorsResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "query", "OSAllLocators")
	retval := types.OSAllLocatorsResponse{Request: request}
	if request == nil {
		return &retval, status.Error(codes.InvalidArgument, "empty request")
	}

	osLocatorStore := prefix.NewStore(sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey), types.OSLocatorAddressKeyPrefix)
	var err error
	retval.Pagination, err = query.Paginate(osLocatorStore, request.Pagination, func(key []byte, value []byte) error {
		record := types.ObjectStoreLocator{}
		if rerr := k.cdc.Unmarshal(value, &record); rerr != nil {
			return rerr
		}
		retval.Locators = append(retval.Locators, record)
		return nil
	})
	if err != nil {
		return &retval, err
	}

	return &retval, nil
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
	scopeAddr, scopeAddrErr := ParseScopeID(scopeID)
	sessionAddr, sessionAddrErr := types.MetadataAddressFromBech32(sessionID)
	if scopeAddrErr == nil && sessionAddrErr == nil {
		scopeAddr2, err := sessionAddr.AsScopeAddress()
		if err != nil {
			return types.MetadataAddress{}, fmt.Errorf("error extracting scope address: %w", err)
		}
		if !scopeAddr.Equals(scopeAddr2) {
			return types.MetadataAddress{}, fmt.Errorf("session %s is not in scope %s", sessionAddr, scopeAddr)
		}
	}
	if sessionAddrErr == nil {
		if sessionAddr.IsSessionAddress() {
			return sessionAddr, nil
		}
		return types.MetadataAddress{}, fmt.Errorf("address [%s] is not a session address", sessionID)
	} else if len(scopeID) == 0 {
		return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into a session address: %w", sessionID, sessionAddrErr)
	}
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

// ParseSessionAddr parses the provided input into a session MetadataAddress.
// The input must be a session address bech32 string.
func ParseSessionAddr(sessionAddr string) (types.MetadataAddress, error) {
	addr, addrErr := types.MetadataAddressFromBech32(sessionAddr)
	if addrErr != nil {
		return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into a session address: %w", sessionAddr, addrErr)
	}
	if !addr.IsSessionAddress() {
		return types.MetadataAddress{}, fmt.Errorf("address [%s] is not a session address", sessionAddr)
	}
	return addr, nil
}

// ParseRecordAddr parses the provided input into a record MetadataAddress.
// The input must be a record address bech32 string.
func ParseRecordAddr(recordAddr string) (types.MetadataAddress, error) {
	addr, addrErr := types.MetadataAddressFromBech32(recordAddr)
	if addrErr != nil {
		return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into a record address: %w", recordAddr, addrErr)
	}
	if !addr.IsRecordAddress() {
		return types.MetadataAddress{}, fmt.Errorf("address [%s] is not a record address", recordAddr)
	}
	return addr, nil
}

// ParseScopeSpecID parses the provided input into a scope spec MetadataAddress.
// The input can either be a uuid string or scope spec address bech32 string.
func ParseScopeSpecID(scopeSpecID string) (types.MetadataAddress, error) {
	addr, addrErr := types.MetadataAddressFromBech32(scopeSpecID)
	if addrErr == nil {
		if addr.IsScopeSpecificationAddress() {
			return addr, nil
		}
		return types.MetadataAddress{}, fmt.Errorf("address [%s] is not a scope spec address", scopeSpecID)
	}
	uid, uidErr := uuid.Parse(scopeSpecID)
	if uidErr == nil {
		return types.ScopeSpecMetadataAddress(uid), nil
	}
	return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into either a scope spec address (%s) or uuid (%s)",
		scopeSpecID, addrErr, uidErr)
}

// ParseContractSpecID parses the provided input into a contract spec MetadataAddress.
// The input can either be a uuid string, a contract spec address bech32 string, or a record spec address bech32 string.
func ParseContractSpecID(contractSpecID string) (types.MetadataAddress, error) {
	addr, addrErr := types.MetadataAddressFromBech32(contractSpecID)
	if addrErr == nil {
		if addr.IsContractSpecificationAddress() {
			return addr, nil
		}
		if addr.IsRecordSpecificationAddress() {
			return addr.MustGetAsContractSpecAddress(), nil
		}
		return types.MetadataAddress{}, fmt.Errorf("address [%s] is not a contract spec address", contractSpecID)
	}
	uid, uidErr := uuid.Parse(contractSpecID)
	if uidErr == nil {
		return types.ContractSpecMetadataAddress(uid), nil
	}
	return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into either a contract spec address (%s) or uuid (%s)",
		contractSpecID, addrErr, uidErr)
}

// ParseRecordSpecID parses the provided input into a record spec MetadataAddress.
// The recordSpecID can either be a uuid string, a record spec address bech32 string, or a contract spec address bech32 string.
// If it's a contract spec address or a uuid, then a name is required.
func ParseRecordSpecID(specID string, name string) (types.MetadataAddress, error) {
	addr, addrErr := types.MetadataAddressFromBech32(specID)
	if addrErr == nil {
		if addr.IsRecordSpecificationAddress() {
			return addr, nil
		}
		if addr.IsContractSpecificationAddress() {
			if len(name) == 0 {
				return types.MetadataAddress{}, errors.New("a name is required when providing a contract spec address")
			}
			return addr.AsRecordSpecAddress(name)
		}
		return types.MetadataAddress{}, fmt.Errorf("address [%s] is not a valid type", specID)
	}
	uid, uidErr := uuid.Parse(specID)
	if uidErr != nil {
		return types.MetadataAddress{}, fmt.Errorf("could not parse [%s] into either a record spec address (%s) or uuid (%s)",
			specID, addrErr, uidErr)
	}
	if len(name) == 0 {
		return types.MetadataAddress{}, errors.New("a name is required when providing a uuid")
	}
	return types.RecordSpecMetadataAddress(uid, name), nil
}

// hasPageRequest is just for use with the getPageRequest func below.
type hasPageRequest interface {
	GetPagination() *query.PageRequest
}

// Gets the query.PageRequest from the provided request if there is one.
// Also sets the default limit if it's not already set yet.
func getPageRequest(req hasPageRequest) *query.PageRequest {
	var pageRequest *query.PageRequest
	if req != nil {
		pageRequest = req.GetPagination()
	}
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}
	if pageRequest.Limit == 0 {
		pageRequest.Limit = defaultLimit
	}
	return pageRequest
}
