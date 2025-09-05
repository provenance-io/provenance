package module

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	client "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/x/registry/client/cli"
	"github.com/provenance-io/provenance/x/registry/keeper"
	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

var _ module.AppModule = AppModule{}
var _ module.AppModuleBasic = AppModuleBasic{}

type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module name.
func (AppModuleBasic) Name() string {
	return registrytypes.ModuleName
}

// GetTxCmd returns the root tx command for the registry module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.CmdTx()
}

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.CmdQuery()
}

// RegisterCodec registers the module's types.
func (AppModuleBasic) RegisterCodec(_ *codec.LegacyAmino) {
	// Register any concrete types if needed.
}

func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// Register the protobuf message types and services with the sdk.
func (AppModuleBasic) RegisterInterfaces(registry types.InterfaceRegistry) {
	registrytypes.RegisterInterfaces(registry)
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(ctx client.Context, mux *runtime.ServeMux) {
	err := registrytypes.RegisterQueryHandlerClient(context.Background(), mux, registrytypes.NewQueryClient(ctx))
	if err != nil {
		panic(err)
	}
}

// DefaultGenesis returns default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&registrytypes.GenesisState{})
}

// ValidateGenesis validates the genesis state.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data registrytypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", registrytypes.ModuleName, err)
	}
	return data.Validate()
}

// AppModule implements an application module for the registry module.
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

// InitGenesis initializes the genesis state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	ctx.Logger().Info("Genesising Registry Module")
	var state registrytypes.GenesisState
	cdc.MustUnmarshalJSON(gs, &state)
	am.keeper.InitGenesis(ctx, &state)
	return nil
}

// Satisfy the AppModule interface.
func (AppModule) IsAppModule()                                {}
func (AppModule) IsOnePerModuleType()                         {}
func (AppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	registrytypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	registrytypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}
