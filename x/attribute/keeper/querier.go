package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// The module level router for state queries (using the Legacy Amino Codec)
func NewQuerier(keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, path[1:], req, keeper, legacyQuerierCdc)
		case types.QueryAttribute:
			return queryAttribute(ctx, path[1:], req, keeper, legacyQuerierCdc)
		case types.QueryAttributes:
			return queryAttributes(ctx, path[1:], req, keeper, legacyQuerierCdc)
		case types.QueryScanAttributes:
			return scanAttributes(ctx, path[1:], req, keeper, legacyQuerierCdc)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

func queryParams(ctx sdk.Context, _ []string, _ abci.RequestQuery, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// Query for account attributes by name.
func queryAttribute(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) < 2 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "missing account address and/or attribute name")
	}
	addr := strings.TrimSpace(path[0])
	err := types.ValidateAttributeAddress(addr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("invalid account address: %v", err))
	}
	name := strings.TrimSpace(path[1])
	if name == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "attribute name cannot be empty")
	}
	attrs, err := keeper.GetAttributes(ctx, addr, name)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	res := types.QueryAttributesResponse{Account: addr}
	res.Attributes = append(res.Attributes, attrs...)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// Query for all account attributes.
func queryAttributes(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) < 1 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "missing account address")
	}
	addr := strings.TrimSpace(path[0])
	err := types.ValidateAttributeAddress(addr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("invalid account address: %v", err))
	}
	attrs, err := keeper.GetAllAttributes(ctx, addr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	res := types.QueryAttributesResponse{Account: addr}
	res.Attributes = append(res.Attributes, attrs...)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

// Query for scanning account attributes by name suffix.
func scanAttributes(ctx sdk.Context, path []string, _ abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) < 2 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "missing account address and/or name suffix")
	}
	addr := strings.TrimSpace(path[0])
	if addr == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "account address cannot be empty")
	}
	account, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "account address must be a Bech32 string")
	}
	suffix := strings.TrimSpace(path[1])
	if suffix == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "attribute name suffix cannot be empty")
	}
	attrs, err := keeper.Scan(ctx.Context(), &types.QueryScanRequest{Account: addr, Suffix: suffix})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	res := types.QueryAttributesResponse{Account: account.String()}
	res.Attributes = append(res.Attributes, attrs.Attributes...)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
