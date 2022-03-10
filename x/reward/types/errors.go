package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrIncorrectModuleAccountBalance = sdkerrors.Register(ModuleName, 1100,
		"reward module account balance != sum of all reward record InitialRewardAmounts")
)
