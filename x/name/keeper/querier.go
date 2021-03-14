package keeper

import (
	"strings"

	"github.com/provenance-io/provenance/x/name/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier creates a new legacy amino query service
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, path[1:], req, k, legacyQuerierCdc)
		case types.QueryResolve:
			return queryResolveName(ctx, path[1:], req, k, legacyQuerierCdc)
		case types.QueryLookup:
			return queryLookupNames(ctx, path[1:], req, k, legacyQuerierCdc)

		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, _ []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := keeper.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// Query for the address a given name is bound to
func queryResolveName(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	name := strings.TrimSpace(path[0])
	if name == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be empty")
	}
	record, err := keeper.GetRecordByName(ctx, name)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if record == nil {
		return nil, types.ErrNameNotBound
	}
	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, queryResFromNameRecord(*record))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// Query for the names that point to a given address.
func queryLookupNames(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	addrs := strings.TrimSpace(path[0])
	if addrs == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address cannot be empty")
	}
	addr, err := sdk.AccAddressFromBech32(addrs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	records, err := keeper.GetRecordsByAddress(ctx, addr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	result := types.QueryNameResults{}
	for _, r := range records {
		result.Records = append(result.Records, queryResFromNameRecord(r))
	}
	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, result)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryResFromNameRecord(r types.NameRecord) types.QueryNameResult {
	return types.QueryNameResult(r)
}
