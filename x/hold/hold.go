package hold

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (e AccountHold) Validate() error {
	if _, err := sdk.AccAddressFromBech32(e.Address); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	if err := e.Amount.Validate(); err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}
	return nil
}
