package module

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/x/asset/client/cli"
	"github.com/provenance-io/provenance/x/asset/keeper"
	"github.com/provenance-io/provenance/x/asset/types"
)

var (
	_ module.AppModuleBasic = (*AppModuleBasic)(nil)

	_ appmodule.AppModule = (*AppModule)(nil)
	_ module.HasServices  = (*AppModule)(nil)
)

type AppModuleBasic struct {
	cdc codec.Codec
}

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

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
	router baseapp.MessageRouter
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, router baseapp.MessageRouter) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		router:         router,
	}
}

func (am AppModule) IsOnePerModuleType() {}

func (am AppModule) IsAppModule() {}

func (AppModule) ConsensusVersion() uint64 { return 1 }

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}
