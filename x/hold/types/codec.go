package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	// govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUnlockVestingAccountsRequest{},
	)
}
