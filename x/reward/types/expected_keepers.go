package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// DistributionKeeper expected distribution keeper (noalias)
type DistributionKeeper interface {
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
}

// StakingKeeper defines a subset of methods implemented by the cosmos-sdk staking keeper
type StakingKeeper interface {
	GetAllDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.Delegation
	GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool)
}

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type KeeperProvider interface {
	GetDistributionKeeper() DistributionKeeper
	GetStakingKeeper() StakingKeeper
	GetAccountKeeper() AccountKeeper
}
