package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/provenance-io/provenance/x/metadata/client/cli"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/simulation"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// type check to ensure the interface is properly implemented
var (
	_ module.AppModuleBasic      = (*AppModule)(nil)
	_ module.AppModuleSimulation = (*AppModule)(nil)

	_ appmodule.AppModule = (*AppModule)(nil)
)

// AppModuleBasic contains non-dependent elements for the metadata module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the metadata module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {
}

// DefaultGenesis returns the default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the metadata module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the metadata module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the metadata module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd returns the root query command for the metadata module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaces implements InterfaceModule
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// AppModule is the standard form metadata module.
type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper authkeeper.AccountKeeper
}

// NewAppModule creates a new AppModule Object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, accountKeeper authkeeper.AccountKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  accountKeeper,
	}
}

// IsOnePerModuleType is a dummy function that satisfies the OnePerModuleType interface (needed by AppModule).
func (AppModule) IsOnePerModuleType() {}

// IsAppModule is a dummy function that satisfies the AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the module name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants does not register any invariants for the metadata module
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// InitGenesis performs genesis initialization for the metadata module. It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the metadata module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ____________________________________________________________________________

// GenerateGenesisState creates a randomized GenState of the metadata module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RandomizedParams creates randomized metadata param changes for the simulator.
func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.LegacyParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for metadata module's types
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {
	// sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the metadata module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	// return simulation.WeightedOperations(
	//	simState.AppParams, simState.Cdc, am.keeper, simState.Contents,
	// )
	return []simtypes.WeightedOperation{}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 4 }
