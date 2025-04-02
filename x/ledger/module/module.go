package mymodule

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	client "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/provenance-io/provenance/x/ledger"
	"github.com/provenance-io/provenance/x/ledger/keeper"
)

type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module name.
func (AppModuleBasic) Name() string {
	return ledger.ModuleName
}

// RegisterCodec registers the module's types. (In newer SDK versions, use RegisterLegacyAminoCodec.)
func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {
	// Register any concrete types if needed.
}

// DefaultGenesis returns default genesis state as raw bytes.
// func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
// 	// Return a default (empty) genesis state.
// 	return cdc.MustMarshalJSON()
// }

// ValidateGenesis validates the genesis state.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// AppModule implements an application module for mymodule.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule instance.
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

// RegisterInvariants registers the invariants of the module.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	ledger.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
}

// Route returns the message routing key for the module.
// func (am AppModule) Route() types.RouterKey {
// 	return ledger.RouterKey
// }

// NewHandler returns an sdk.Handler for the module.
// func (am AppModule) NewHandler() sdk.Handler {
// 	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
// 		// Handle messages. This is a placeholder.
// 		return nil, nil
// 	}
// }

// QuerierRoute returns the module's querier route name.
func (am AppModule) QuerierRoute() string {
	return ledger.ModuleName
}

// LegacyQuerierHandler returns the module sdk.Querier.
// func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
// 	// Implement query handlers if needed.
// 	return nil
// }

// InitGenesis initializes the genesis state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	// Initialize genesis state if needed.
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports the module's genesis state.
// func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
// 	// Export current state as genesis state.
// 	return cdc.MustMarshalJSON(struct{}{})
// }

// Satisfy the AppModule interface for now..
func (AppModule) IsAppModule()                                                {}
func (AppModule) IsOnePerModuleType()                                         {}
func (AppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino)                 {}
func (AppModule) RegisterInterfaces(types.InterfaceRegistry)                  {}
func (AppModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}
