package keeper

import (
	"context"
	"fmt"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/provenance-io/provenance/x/marker/types"
)

// A WrappedBankKeeper is the bank module's BaseKeeper with extended functionality.
// The SDK does not have a DeleteDenomMetadata method, but one is needed.
// This allows us to continue to use an interface and mocks.
type WrappedBankKeeper struct {
	*bankkeeper.BaseKeeper
}

func NewWrappedBankKeeper(kpr *bankkeeper.BaseKeeper) WrappedBankKeeper {
	return WrappedBankKeeper{BaseKeeper: kpr}
}

var _ types.BankKeeper = (*WrappedBankKeeper)(nil)

// DeleteDenomMetadata deletes the provided denom's metadata from the state store.
func (k WrappedBankKeeper) DeleteDenomMetaData(ctx context.Context, denom string) error {
	err := k.BaseViewKeeper.DenomMetadata.Remove(ctx, denom)
	if err != nil {
		return fmt.Errorf("could not remove denom metadata for %q: %w", denom, err)
	}
	return nil
}
