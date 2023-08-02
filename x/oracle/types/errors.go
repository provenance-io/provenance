package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrSample                      = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout        = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion              = sdkerrors.Register(ModuleName, 1501, "invalid version")
	ErrContractAddressDoesNotExist = sdkerrors.Register(ModuleName, 1502, "missing contract address")
)
