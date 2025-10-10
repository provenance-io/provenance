package module

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/x/registry/client/cli"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/provenance-io/provenance/x/registry/types"
)

var (
	_ module.AppModuleBasic      = (*AppModuleBasic)(nil)
	_ appmodule.AppModule        = (*AppModule)(nil)
	_ module.HasServices         = (*AppModule)(nil)
	_ module.HasGenesis          = (*AppModule)(nil)
	_ module.HasConsensusVersion = (*AppModule)(nil)
)

// AppModuleBasic contains non-dependent elements for the registry module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module name ("registry").
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// GetTxCmd returns the root tx command for the registry module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.CmdTx()
}

// GetQueryCmd returns the root query command for the registry module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.CmdQuery()
}

// RegisterLegacyAminoCodec is deprecated, not used anymore, and does nothing.
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterInterfaces registers the registry module's types for the given codec.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the registry module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(ctx client.Context, mux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(ctx))
	if err != nil {
		panic(err)
	}
}

// AppModule implements an application module for the registry module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule instance for the registry module.
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

func (AppModule) IsAppModule()        {}
func (AppModule) IsOnePerModuleType() {}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// DefaultGenesis returns default genesis state as raw bytes.
func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis validates the genesis state.
func (AppModule) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return data.Validate()
}

// InitGenesis initializes the genesis state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var state types.GenesisState
	cdc.MustUnmarshalJSON(gs, &state)
	am.keeper.InitGenesis(ctx, &state)
}

// ExportGenesis returns the exported genesis state as raw bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}
