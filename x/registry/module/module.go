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
	"github.com/provenance-io/provenance/x/registry"
	"github.com/provenance-io/provenance/x/registry/client/cli"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/spf13/cobra"
)

type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module name.
func (AppModuleBasic) Name() string {
	return registry.ModuleName
}

// GetTxCmd returns the root tx command for the registry module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.CmdTx()
}

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.CmdQuery()
}

// RegisterCodec registers the module's types.
func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {
	// Register any concrete types if needed.
}

// AppModule implements an application module for the registry module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.RegistryKeeper
}

// NewAppModule creates a new AppModule instance.
func NewAppModule(cdc codec.Codec, k keeper.RegistryKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

// InitGenesis initializes the genesis state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	ctx.Logger().Info("Genesising Registry Module")
	var state registry.GenesisState
	cdc.MustUnmarshalJSON(gs, &state)
	am.keeper.InitGenesis(ctx, &state)
	return nil
}

// DefaultGenesis returns default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&registry.GenesisState{})
}

// ValidateGenesis validates the genesis state.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// Satisfy the AppModule interface.
func (AppModule) IsAppModule()                                {}
func (AppModule) IsOnePerModuleType()                         {}
func (AppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	registry.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	registry.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// Register the protobuf message types and services with the sdk.
func (AppModule) RegisterInterfaces(r types.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(r, &registry.Msg_serviceDesc)
}

func (AppModule) RegisterGRPCGatewayRoutes(ctx client.Context, mux *runtime.ServeMux) {
	err := registry.RegisterQueryHandlerClient(context.Background(), mux, registry.NewQueryClient(ctx))
	if err != nil {
		panic(err)
	}
}
