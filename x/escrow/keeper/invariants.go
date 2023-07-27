package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/escrow"
)

const balanceInvariant = "Escrow-Account-Balances"

// RegisterInvariants registers all quarantine invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(escrow.ModuleName, balanceInvariant, EscrowAccountBalancesInvariant(keeper))
}

// EscrowAccountBalancesInvariant checks that all funds in escrow are also otherwise unlocked in the account.
func EscrowAccountBalancesInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broken := escrowAccountBalancesInvariantHelper(ctx, keeper)
		return sdk.FormatInvariant(escrow.ModuleName, balanceInvariant, msg), broken
	}
}

// escrowAccountBalancesInvariantHelper does all the heavy lifting for EscrowAccountBalancesInvariant.
// It will look up all escrow records and make sure that each address actually has the funds that are locked in escrow.
func escrowAccountBalancesInvariantHelper(ctx sdk.Context, keeper Keeper) (string, bool) {
	allEscrows, err := keeper.GetAllAccountEscrows(ctx)
	if err != nil {
		return fmt.Sprintf("Failed to get a record of all funds that are in escrow: %v", err), true
	}

	var addr sdk.AccAddress
	var total sdk.Coins
	var errs []error
	ctx = escrow.WithBypass(ctx)
	for _, ae := range allEscrows {
		total = total.Add(ae.Amount...)
		addr, err = sdk.AccAddressFromBech32(ae.Address)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid address %q with funds in escrow: %w", ae.Address, err))
		}
		if err = keeper.ValidateNewEscrow(ctx, addr, ae.Amount); err != nil {
			errs = append(errs, err)
		}
	}

	var msg strings.Builder

	allCount := len(allEscrows)
	switch allCount {
	case 0:
		msg.WriteString("No accounts have funds in escrow.")
	case 1:
		msg.WriteString(fmt.Sprintf("1 account has %s in escrow.", total))
	default:
		msg.WriteString(fmt.Sprintf("%d accounts have %s in escrow.", allCount, total))
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
