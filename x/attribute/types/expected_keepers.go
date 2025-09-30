package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	nametypes "github.com/provenance-io/provenance/x/name/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// NameKeeper defines the expected account keeper used for simulations (noalias)
type NameKeeper interface {
	ResolvesTo(ctx sdk.Context, name string, addr sdk.AccAddress) bool
	Normalize(ctx sdk.Context, name string) (string, error)
	GetRecordByName(ctx sdk.Context, name string) (record *nametypes.NameRecord, err error)
	NameExists(ctx sdk.Context, name string) bool
	SetAttributeKeeper(attrKeeper nametypes.AttributeKeeper)
	SetNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error
	UpdateNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error
	IterateRecords(ctx sdk.Context, prefix []byte, handle func(nametypes.NameRecord) error) error
}
