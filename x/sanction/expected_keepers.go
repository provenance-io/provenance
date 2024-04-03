package sanction

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// AccountKeeper defines the account/auth functionality needed from within the sanction module.
type AccountKeeper interface {
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)
}

// BankKeeper defines the bank functionality needed from within the sanction module.
type BankKeeper interface {
	AppendSendRestriction(restriction banktypes.SendRestrictionFn)
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// GovKeeper defines the gov functionality needed from within the sanction module.
type GovKeeper interface {
	GetProposal(ctx context.Context, proposalID uint64) (govv1.Proposal, bool)
}
