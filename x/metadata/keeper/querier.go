package keeper

import (
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/x/metadata/types"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/google/uuid"
)

// NewQuerier creates a querier for auth REST endpoints
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryScope:
			return queryScope(ctx, path, req, k, legacyQuerierCdc)
		case types.QueryOwnership:
			return queryAddressScopes(ctx, path, req, k, legacyQuerierCdc)
		case types.QueryParams:
			return queryParams(ctx, k, legacyQuerierCdc)
		case types.QueryScopeSpec:
			return queryScopeSpecification(ctx, path, k, legacyQuerierCdc)
		// TODO: add contract spec stuff
		// TODO: add record spec stuff
		case types.QueryOSParams:
			return queryOSLocatorParams(ctx, k, legacyQuerierCdc)
		case types.QueryOSGet:
			return queryOSGet(ctx, path, k, legacyQuerierCdc)
		case types.QueryOSGetByURI:
			return queryOSGetByURI(ctx, path, k, legacyQuerierCdc)
		case types.QueryOSGetByScope:
			return queryOSGetByScope(ctx, path, k, legacyQuerierCdc)
		case types.QueryOSGetAll:
			return queryOSGetAll(ctx, k, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown query endpoint")
		}
	}
}

// Query for a scope by UUID.
func queryScope(ctx sdk.Context, path []string, _ abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	scopeUUID, err := uuid.Parse(strings.TrimSpace(path[1]))
	if err != nil {
		ctx.Logger().Error(err.Error())
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	scope, found := k.GetScope(ctx, types.ScopeMetadataAddress(scopeUUID))
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "scope does not exist")
	}
	scopeBytes, err := k.cdc.Marshal(&scope)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	res, err := legacyQuerierCdc.MarshalJSON(types.QueryResScope{Scope: scopeBytes})
	if err != nil {
		ctx.Logger().Error("unable to marshal scope to JSON", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

// Query for a scopes associated with an address.
func queryAddressScopes(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := types.QueryMetadataParams{Page: 0, Limit: 100}
	address, err := sdk.AccAddressFromBech32(strings.TrimSpace(path[1]))
	if err != nil || address.Empty() {
		errm := "invalid address to query scopes for"
		ctx.Logger().Error(errm)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, errm)
	}
	scopes := make([]string, 0)
	err = k.IterateScopesForAddress(ctx, address, func(scopeID types.MetadataAddress) (stop bool) {
		scopes = append(scopes, scopeID.String())
		return false
	})
	// check for parameters used for paging.
	if len(req.Data) > 0 {
		err = legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
		}
	}

	// TODO: consider a parameter configuration item for the limit here (1000)

	// BOOKMARK -- create a v1 to v0 migration function to re-assemble a scope from groups and records

	start, end := client.Paginate(len(scopes), params.Page, params.Limit, 1000)
	if start < 0 || end < 0 {
		scopes = []string{}
	} else {
		scopes = scopes[start:end]
	}
	if err != nil {
		ctx.Logger().Error("unable to get scope IDs for address", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	res, err := legacyQuerierCdc.MarshalJSON(types.QueryResOwnership{Address: address, ScopeID: scopes})
	if err != nil {
		ctx.Logger().Error("unable to marshal scope ids to JSON", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryParams(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// query for a scope specification by specification id
func queryScopeSpecification(ctx sdk.Context, path []string, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	specificationUUID, err := uuid.Parse(strings.TrimSpace(path[1]))
	if err != nil {
		ctx.Logger().Error(err.Error())
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	scopeSpec, found := k.GetScopeSpecification(ctx, types.ScopeMetadataAddress(specificationUUID))
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("scope specification uuid [%s] does not exist", specificationUUID))
	}
	res, err := legacyQuerierCdc.MarshalJSON(types.NewQueryResScopeSpec(scopeSpec))
	if err != nil {
		ctx.Logger().Error("unable to marshal scope spec to JSON", "specificationUUID", specificationUUID, "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return res, nil
}

func queryOSLocatorParams(ctx sdk.Context, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := keeper.GetOSLocatorParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryOSGet(ctx sdk.Context, path []string, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(path[1])
	if err != nil {
		return nil, types.ErrInvalidAddress
	}
	msgs, _ := keeper.GetOsLocatorRecord(ctx, accAddr)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, msgs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryOSGetByURI(ctx sdk.Context, path []string, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	// intentionally leaving out pagination for now for this one, not really anything legacy rest for this :shrug:
	var records []types.ObjectStoreLocator

	appendToRecords := func(record types.ObjectStoreLocator) bool {
		if record.LocatorUri == path[1] {
			records = append(records, record)
			// have to get all the uri associated with an address..imo..check
		}
		return false
	}
	if err := keeper.IterateOSLocators(ctx, appendToRecords); err != nil {
		return nil, err
	}
	uniqueRecords := uniqueRecords(records)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, &uniqueRecords)

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryOSGetAll(ctx sdk.Context, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	// intentionally leaving out pagination for now for this one, not really anything legacy rest for this :shrug:
	var records []types.ObjectStoreLocator

	appendToRecords := func(record types.ObjectStoreLocator) bool {
		records = append(records, record)
		// have to get all the uri associated with an address..imo..check
		return false
	}
	if err := keeper.IterateOSLocators(ctx, appendToRecords); err != nil {
		return nil, err
	}
	uniqueRecords := uniqueRecords(records)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, &uniqueRecords)

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryOSGetByScope(ctx sdk.Context, path []string, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	locators, lErr := keeper.GetOSLocatorByScope(ctx, path[1])
	if lErr != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, lErr.Error())
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, locators)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func uniqueRecords(records []types.ObjectStoreLocator) []types.ObjectStoreLocator {
	if len(records) == 0 {
		return records
	}
	seen := make([]types.ObjectStoreLocator, 0, len(records))

slice:
	for i, n := range records {
		if i == 0 {
			records = records[:0]
		}
		for _, t := range seen {
			if n.Owner == t.Owner && n.LocatorUri == t.LocatorUri {
				continue slice
			}
		}
		seen = append(seen, n)
		records = append(records, n)
	}
	return records
}
