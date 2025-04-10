package module

import (
	"context"
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	client "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/provenance-io/provenance/x/ledger"
	"github.com/provenance-io/provenance/x/ledger/client/cli"
	"github.com/provenance-io/provenance/x/ledger/keeper"
	"github.com/spf13/cobra"
)

type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module name.
func (AppModuleBasic) Name() string {
	return ledger.ModuleName
}

// GetTxCmd returns the root tx command for the attribute module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.CmdTx()
}

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.CmdQuery()
}

// RegisterCodec registers the module's types. (In newer SDK versions, use RegisterLegacyAminoCodec.)
func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {
	// Register any concrete types if needed.
}

// AppModule implements an application module for mymodule.
type AppModule struct {
	AppModuleBasic
	keeper keeper.BaseKeeper
}

// NewAppModule creates a new AppModule instance.
func NewAppModule(cdc codec.Codec, k keeper.BaseKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

// RegisterInvariants registers the invariants of the module.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

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
	ctx.Logger().Info("Genesising Ledger Module")
	var state ledger.GenesisState
	cdc.MustUnmarshalJSON(gs, &state)
	am.keeper.InitGenesis(ctx, &state)
	return nil
}

// DefaultGenesis returns default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&ledger.GenesisState{})
}

// ValidateGenesis validates the genesis state.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// ExportGenesis exports the module's genesis state.
// func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
// 	// Export current state as genesis state.
// 	return cdc.MustMarshalJSON(struct{}{})
// }

// Satisfy the AppModule interface for now..
func (AppModule) IsAppModule()                                {}
func (AppModule) IsOnePerModuleType()                         {}
func (AppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	ledger.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))

	ledger.RegisterQueryServer(cfg.QueryServer(), keeper.NewLedgerQueryServer(am.keeper))
}

// Register the protobuf message types and services with the sdk.
func (AppModule) RegisterInterfaces(registry types.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &ledger.Msg_serviceDesc)
}

func (AppModule) RegisterGRPCGatewayRoutes(ctx client.Context, mux *runtime.ServeMux) {
	err := ledger.RegisterQueryHandlerClient(context.Background(), mux, ledger.NewQueryClient(ctx))
	if err != nil {
		panic(err)
	}
}
