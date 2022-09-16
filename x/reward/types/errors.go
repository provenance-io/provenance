package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/rewards module errors
var (
	ErrIterateAllRewardAccountStates  = sdkerrors.Register(ModuleName, 2, "error iterating all reward account states")
	ErrRewardProgramNotFound          = sdkerrors.Register(ModuleName, 3, "reward program not found")
	ErrEndRewardProgramNotAuthorized  = sdkerrors.Register(ModuleName, 4, "not authorized to end the reward program")
	ErrEndrewardProgramIncorrectState = sdkerrors.Register(ModuleName, 5, "unable to end a reward program that is finished or expired")
)
