package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/hold"
)

const balanceInvariant = "Hold-Account-Balances"

// RegisterInvariants registers all quarantine invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(hold.ModuleName, balanceInvariant, HoldAccountBalancesInvariant(keeper))
}

// HoldAccountBalancesInvariant checks that all funds on hold are also otherwise unlocked in the account.
func HoldAccountBalancesInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broken := holdAccountBalancesInvariantHelper(ctx, keeper)
		return sdk.FormatInvariant(hold.ModuleName, balanceInvariant, msg), broken
	}
}

// holdAccountBalancesInvariantHelper does all the heavy lifting for HoldAccountBalancesInvariant.
// It will look up all hold records and make sure that each address actually has the funds that are locked on hold.
func holdAccountBalancesInvariantHelper(ctx sdk.Context, keeper Keeper) (string, bool) {
	allHolds, err := keeper.GetAllAccountHolds(ctx)
	if err != nil {
		return fmt.Sprintf("Failed to get a record of all funds that are on hold: %v", err), true
	}

	var addr sdk.AccAddress
	var total sdk.Coins
	var errs []error
	ctx = hold.WithBypass(ctx)
	for _, ae := range allHolds {
		total = total.Add(ae.Amount...)
		addr, err = sdk.AccAddressFromBech32(ae.Address)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid address %q with funds on hold: %w", ae.Address, err))
		}
		if err = keeper.ValidateNewHold(ctx, addr, ae.Amount); err != nil {
			errs = append(errs, err)
		}
	}

	var msg strings.Builder

	allCount := len(allHolds)
	switch allCount {
	case 0:
		msg.WriteString("No accounts have funds on hold.")
	case 1:
		msg.WriteString(fmt.Sprintf("1 account has %s on hold.", total))
	default:
		msg.WriteString(fmt.Sprintf("%d accounts have %s on hold.", allCount, total))
	}

	msg.WriteByte(' ')
	broken := true
	errCount := len(errs)
	switch errCount {
	case 0:
		broken = false
		msg.WriteString("No problems detected.")
	case 1:
		msg.WriteString(fmt.Sprintf("1 problem detected: %v", errs[0]))
	case 2:
		msg.WriteString(fmt.Sprintf("%d problems detected:", errCount))
		for i, er := range errs {
			msg.WriteString(fmt.Sprintf("\n%d: %v", i+1, er))
		}
	}

	return msg.String(), broken
}
