package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// scopeDenomPrefix is the string that will start every scope denom.
const scopeDenomPrefix = types.DenomPrefix + types.PrefixScope + "1"

func NewMDBankKeeper(bk bankkeeper.BaseKeeper) *MDBankKeeper {
	return &MDBankKeeper{BaseKeeper: bk}
}

// MDBankKeeper extends the SDK's bank keeper to add methods that act on fields so that we can mock all of it.
type MDBankKeeper struct {
	bankkeeper.BaseKeeper
}

// DenomOwner gets the singular owner of a denom.
// An error is returned if more than one account owns some of the denom.
// If no one owns the denom, this will return nil, nil.
func (k *MDBankKeeper) DenomOwner(ctx context.Context, denom string) (sdk.AccAddress, error) {
	var rv sdk.AccAddress
	ranger := collections.NewPrefixedPairRange[string, sdk.AccAddress](denom)
	err := k.Balances.Indexes.Denom.Walk(ctx, ranger, func(_ string, addr sdk.AccAddress) (bool, error) {
		if len(rv) > 0 {
			return true, fmt.Errorf("denom %q has more than one owner", denom)
		}
		rv = addr
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return rv, nil
}

// GetScopesForValueOwner will get the scopes owned by a specific value owner.
// If the pageReq is nil, this will get all their scopes and the resulting PageResponse will be nil.
// If a pageReq is provided, this will get just the requested page and it will return a PageResponse.
func (k *MDBankKeeper) GetScopesForValueOwner(ctx context.Context, valueOwner sdk.AccAddress, pageReq *query.PageRequest) (types.AccMDLinks, *query.PageResponse, error) {
	pfx := collections.Join(valueOwner, scopeDenomPrefix)

	if pageReq != nil {
		return query.CollectionPaginate(ctx, k.Balances, pageReq,
			func(key collections.Pair[sdk.AccAddress, string], _ sdkmath.Int) (*types.AccMDLink, error) {
				return k.balanceValueOwnerTransformer(key), nil
			},
			func(o *query.CollectionsPaginateOptions[collections.Pair[sdk.AccAddress, string]]) {
				o.Prefix = &pfx
			},
		)
	}

	ranger := &collections.Range[collections.Pair[sdk.AccAddress, string]]{}
	ranger.Prefix(pfx)
	var links types.AccMDLinks
	err := k.Balances.Walk(ctx, ranger, func(key collections.Pair[sdk.AccAddress, string], _ sdkmath.Int) (bool, error) {
		links = append(links, k.balanceValueOwnerTransformer(key))
		return false, nil
	})

	return links, nil, err
}

// balanceValueOwnerTransformer creates an AccMDLink from data in the key. If the denom in the key is not a
// metadata denom, an error is written to the logs and the resulting AccMDLink will not have an MDAddr.
func (k *MDBankKeeper) balanceValueOwnerTransformer(key collections.Pair[sdk.AccAddress, string]) *types.AccMDLink {
	accAddr := key.K1()
	denom := key.K2()
	mdAddr, err := types.MetadataAddressFromDenom(denom)
	if err != nil {
		// MetadataAddressFromDenom always includes the denom in the error message, so we don't need it again here.
		k.Logger().Error(fmt.Sprintf("invalid metadata balance entry for account %q: %v", accAddr.String(), err))
	}
	return types.NewAccMDLink(accAddr, mdAddr)
}
