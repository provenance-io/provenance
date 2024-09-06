package keeper

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// scopeDenomPrefix is the string that will start every scope denom.
const scopeDenomPrefix = types.DenomPrefix + types.PrefixScope + "1"

var oneInt = sdkmath.OneInt()

// mintCoinsRestriction returns an error if any coin is not for a scope, or isn't for 1.
func mintCoinsRestriction(_ context.Context, coins sdk.Coins) error {
	for _, coin := range coins {
		if !strings.HasPrefix(coin.Denom, scopeDenomPrefix) {
			return fmt.Errorf("cannot mint %s: denom is not for a scope", coin)
		}
		if !coin.Amount.Equal(oneInt) {
			return fmt.Errorf("cannot mint %s: amount is not one", coin)
		}
	}
	return nil
}

func NewMDBankKeeper(bk bankkeeper.BaseKeeper) *MDBankKeeper {
	return &MDBankKeeper{BaseKeeper: bk.WithMintCoinsRestriction(mintCoinsRestriction)}
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

// TODO[2137]: Get rid of this GetBalancesCollection method and replace it with a method that returns
// an AccMDLinks with all the scopes owned by an address with optional pagination parameters.

// GetBalancesCollection gets the Balances collection from the underlying bank keeper.
func (k *MDBankKeeper) GetBalancesCollection() *collections.IndexedMap[collections.Pair[sdk.AccAddress, string], sdkmath.Int, bankkeeper.BalancesIndexes] {
	return k.Balances
}
