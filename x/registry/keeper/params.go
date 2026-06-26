package keeper

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/registry/types"
)

// GetAuthority returns the module's governance authority address.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// IsAuthority returns true if the provided bech32 address is the module's authority.
func (k Keeper) IsAuthority(addr string) bool {
	return strings.EqualFold(k.authority, addr)
}

// ValidateAuthority returns an error if the provided address is not the module's authority.
func (k Keeper) ValidateAuthority(addr string) error {
	if !k.IsAuthority(addr) {
		return govtypes.ErrInvalidSigner.Wrapf("expected %q got %q", k.GetAuthority(), addr)
	}
	return nil
}

// GetParams returns the registry module params. If no params have been set, it returns the module
// defaults.
func (k Keeper) GetParams(ctx context.Context) types.Params {
	params, err := k.Params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.DefaultParams()
		}
		panic(fmt.Errorf("could not get registry params: %w", err))
	}
	return params
}

// SetParams stores the registry module params in state.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.Params.Set(ctx, params)
}
