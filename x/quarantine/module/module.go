package module

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/x/quarantine"
	"github.com/provenance-io/provenance/x/quarantine/client/cli"
	"github.com/provenance-io/provenance/x/quarantine/keeper"
	"github.com/provenance-io/provenance/x/quarantine/simulation"
)

var (
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

type AppModuleBasic struct {
	cdc codec.Codec
}

func (AppModuleBasic) Name() string {
	return quarantine.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the quarantine module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(quarantine.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the quarantine module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data quarantine.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", quarantine.ModuleName, err)
	}
	return data.Validate()
}

// GetQueryCmd returns the cli query commands for the quarantine module
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.QueryCmd()
}

// GetTxCmd returns the transaction commands for the quarantine module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.TxCmd()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the quarantine module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	if err := quarantine.RegisterQueryHandlerClient(context.Background(), mux, quarantine.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers the quarantine module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	quarantine.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec registers the quarantine module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

type AppModule struct {
	AppModuleBasic
	keeper     keeper.Keeper
	accKeeper  quarantine.AccountKeeper
	bankKeeper quarantine.BankKeeper
	registry   cdctypes.InterfaceRegistry
}

func NewAppModule(cdc codec.Codec, quarantineKeeper keeper.Keeper, accKeeper quarantine.AccountKeeper, bankKeeper quarantine.BankKeeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         quarantineKeeper,
		accKeeper:      accKeeper,
		bankKeeper:     bankKeeper,
		registry:       registry,
	}
}

// IsOnePerModuleType is a dummy function that satisfies the OnePerModuleType interface (needed by AppModule).
func (AppModule) IsOnePerModuleType() {}

// IsAppModule is a dummy function that satisfies the AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the quarantine module's name.
func (AppModule) Name() string {
	return quarantine.ModuleName
}

// RegisterInvariants does nothing, there are no invariants to enforce for the quarantine module.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// InitGenesis performs genesis initialization for the quarantine module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState quarantine.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the quarantine module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// RegisterServices registers a gRPC query service to respond to the quarantine-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	quarantine.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	quarantine.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the quarantine module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState, am.keeper.GetFundsHolder())
}

// ProposalContents returns all the quarantine content functions used to
// simulate governance proposals.
func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized quarantine param changes for the simulator.
func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.LegacyParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for quarantine module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[quarantine.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the quarantine module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(simState, am.accKeeper, am.bankKeeper, am.keeper)
}
