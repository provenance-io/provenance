package exchange

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

// AccountKeeper defines the expected account keeper interface.
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
}

// AttributeKeeper defines the expected attribute keeper interface.
type AttributeKeeper interface {
	GetAllAttributesAddr(ctx sdk.Context, addr []byte) ([]attrtypes.Attribute, error)
}

// BankKeeper defines the expected bank keeper interface.
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	InputOutputCoinsProv(ctx context.Context, inputs []banktypes.Input, outputs []banktypes.Output) error
	BlockedAddr(addr sdk.AccAddress) bool
}

// HoldKeeper defines the expected hold keeper interface.
type HoldKeeper interface {
	AddHold(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins, reason string) error
	ReleaseHold(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error
	GetHoldCoin(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error)
}

// MarkerKeeper defines the expected marker keeper interface.
type MarkerKeeper interface {
	GetMarker(ctx sdk.Context, address sdk.AccAddress) (markertypes.MarkerAccountI, error)
	AddSetNetAssetValues(ctx sdk.Context, marker markertypes.MarkerAccountI, netAssetValues []markertypes.NetAssetValue, source string) error
	GetNetAssetValue(ctx sdk.Context, markerDenom, priceDenom string) (*markertypes.NetAssetValue, error)
}

// MetadataKeeper defines the expected metadata keeper interface
type MetadataKeeper interface {
	AddSetNetAssetValues(ctx sdk.Context, scopeID metadatatypes.MetadataAddress, netAssetValues []metadatatypes.NetAssetValue, source string) error
	GetNetAssetValue(ctx sdk.Context, metadataDenom, priceDenom string) (*metadatatypes.NetAssetValue, error)
}
