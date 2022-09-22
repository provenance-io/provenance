package types

import (
	cerrs "cosmossdk.io/errors"
)

// x/rewards module errors
var (
	ErrIterateAllRewardAccountStates  = cerrs.Register(ModuleName, 2, "error iterating all reward account states")
	ErrRewardProgramNotFound          = cerrs.Register(ModuleName, 3, "reward program not found")
	ErrEndRewardProgramNotAuthorized  = cerrs.Register(ModuleName, 4, "not authorized to end the reward program")
	ErrEndrewardProgramIncorrectState = cerrs.Register(ModuleName, 5, "unable to end a reward program that is finished or expired")
)
