package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

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
