package module

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/provenance-io/provenance/x/asset/client/cli"
	assetkeeper "github.com/provenance-io/provenance/x/asset/keeper"
	"github.com/provenance-io/provenance/x/asset/simulation"
	"github.com/provenance-io/provenance/x/asset/types"
)

var (
	_ module.AppModuleBasic      = (*AppModule)(nil)
	_ module.AppModuleSimulation = (*AppModule)(nil)

	_ appmodule.AppModule = (*AppModule)(nil)
)

type AppModuleBasic struct {
	cdc codec.Codec
}

type AppModule struct {
	AppModuleBasic
	keeper        assetkeeper.Keeper
	accountKeeper authkeeper.AccountKeeper
	nftKeeper     types.NFTKeeper
	router        baseapp.MessageRouter
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper assetkeeper.Keeper, accountKeeper authkeeper.AccountKeeper, nftKeeper types.NFTKeeper, router baseapp.MessageRouter) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  accountKeeper,
		nftKeeper:      nftKeeper,
		router:         router,
	}
}

func (am AppModule) IsOnePerModuleType() {}

func (am AppModule) IsAppModule() {}

// Name returns the module name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// GetTxCmd returns the root tx command for the asset module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the root query command for the asset module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {
}

func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), assetkeeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), assetkeeper.NewQueryServerImpl(am.keeper))
}

func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for asset module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[exchange.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations()
}
