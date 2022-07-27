package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/rewards module errors
var (
	ErrIterateAllRewardAccountStates = sdkerrors.Register(ModuleName, 2, "error iterating all reward account states")
)
