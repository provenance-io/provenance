package streaming

import (
	"github.com/cosmos/cosmos-sdk/codec"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// TOML configuration parameter keys
const (
	// TomlKey is the top-level TOML key for streaming configuration
	TomlKey = "streaming"

	// EnabledParam is the flag that enables ABCI request response streaming services
	EnabledParam = "enabled"
)

// StreamService interface used to hook into the ABCI message processing of the App
type StreamService interface {
	// StreamBeginBlocker updates the streaming service with the latest BeginBlock messages
	StreamBeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock)
	// StreamEndBlocker updates the steaming service with the latest EndBlock messages
	StreamEndBlocker(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock)
}

// StreamServiceInitializer interface initializes StreamService
// implementations which are then registered with the App
type StreamServiceInitializer interface {
	// Init configures and initializes the streaming service
	Init(opts serverTypes.AppOptions, marshaller codec.BinaryCodec) StreamService
}
