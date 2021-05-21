package keeper

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// NewQuerier is the module level router for state queries (using the Legacy Amino Codec)
func NewQuerier(keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryMarkers:
			return queryMarkers(ctx, path[1:], req, keeper, legacyQuerierCdc)
		case types.QueryHolders:
			return queryHolders(ctx, path[1:], req, keeper, legacyQuerierCdc)
		case types.QueryMarker:
			return queryMarker(ctx, path[1:], req, keeper, legacyQuerierCdc)
		case types.QueryMarkerAccess:
			return queryAccess(ctx, path[1:], req, keeper, legacyQuerierCdc)
		case types.QueryMarkerEscrow:
			return queryCoins(ctx, path[1:], req, keeper, legacyQuerierCdc)
		case types.QueryMarkerSupply:
			return querySupply(ctx, path[1:], req, keeper, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}

// query for a single marker by denom or address
func queryMarker(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	account, err := accountForDenomOrAddress(ctx, keeper, path[0])
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// query for access records on an account
func queryAccess(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	account, err := accountForDenomOrAddress(ctx, keeper, path[0])
	if err != nil {
		return nil, err
	}

	m := account.(*types.MarkerAccount)
	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, m.AccessControl)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// query for coins on a marker account
func queryCoins(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	account, err := accountForDenomOrAddress(ctx, keeper, path[0])
	if err != nil {
		return nil, err
	}

	m := account.(*types.MarkerAccount)
	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, keeper.GetEscrow(ctx, m))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// query for supply of coin on a marker account
func querySupply(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	account, err := accountForDenomOrAddress(ctx, keeper, path[0])
	if err != nil {
		return nil, err
	}

	m := account.(*types.MarkerAccount)
	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, m.Supply)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// query for all marker accounts
func queryMarkers(ctx sdk.Context, _ []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryMarkersParams

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	filteredMarkers := make([]types.MarkerAccount, 0)
	keeper.IterateMarkers(ctx, func(record types.MarkerAccountI) bool {
		if len(params.Status) < 1 || strings.EqualFold(record.GetStatus().String(), params.Status) {
			m := record.(*types.MarkerAccount)
			filteredMarkers = append(filteredMarkers, *m)
		}
		return false
	})

	start, end := client.Paginate(len(filteredMarkers), params.Page, params.Limit, len(filteredMarkers))
	if start < 0 || end < 0 {
		filteredMarkers = []types.MarkerAccount{}
	} else {
		filteredMarkers = filteredMarkers[start:end]
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, filteredMarkers)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// query for all accounts holding the given marker coins
func queryHolders(ctx sdk.Context, _ []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryMarkersParams

	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	holders := keeper.GetAllMarkerHolders(ctx, params.Denom)

	start, end := client.Paginate(len(holders), params.Page, params.Limit, len(holders))
	if start < 0 || end < 0 {
		holders = []types.Balance{}
	} else {
		holders = holders[start:end]
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, holders)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func accountForDenomOrAddress(ctx sdk.Context, keeper Keeper, lookup string) (types.MarkerAccountI, error) {
	var addrErr, err error
	var addr sdk.AccAddress
	var account types.MarkerAccountI

	// try to parse the argument as an address, if this fails try as a denom string.
	if addr, addrErr = sdk.AccAddressFromBech32(lookup); addrErr != nil {
		account, err = keeper.GetMarkerByDenom(ctx, lookup)
	} else {
		account, err = keeper.GetMarker(ctx, addr)
	}
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrMarkerNotFound, "invalid denom or address")
	}
	return account, nil
}
