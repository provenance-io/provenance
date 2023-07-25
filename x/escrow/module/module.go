package module

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/x/escrow"
	"github.com/provenance-io/provenance/x/escrow/client/cli"
	"github.com/provenance-io/provenance/x/escrow/keeper"
	"github.com/provenance-io/provenance/x/escrow/simulation"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, escrowKeeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         escrowKeeper,
	}
}

type AppModuleBasic struct {
	cdc codec.Codec
}

func (AppModuleBasic) Name() string {
	return escrow.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the escrow module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(escrow.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the escrow module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data escrow.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", escrow.ModuleName, err)
	}
	return data.Validate()
}

// GetQueryCmd returns the cli query commands for the escrow module.
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.QueryCmd()
}

// GetTxCmd returns the transaction commands for the escrow module.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the escrow module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	if err := escrow.RegisterQueryHandlerClient(context.Background(), mux, escrow.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers the escrow module's interface types
func (AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry) {}

// RegisterLegacyAminoCodec registers the escrow module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterInvariants registers the invariants for the escrow module.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// Deprecated: Route returns the message routing key for the escrow module, empty.
func (am AppModule) Route() sdk.Route { return sdk.Route{} }

// Deprecated: QuerierRoute returns the route we respond to for abci queries, "".
func (AppModule) QuerierRoute() string { return "" }

// Deprecated: LegacyQuerierHandler returns the escrow module sdk.Querier (nil).
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier { return nil }

// InitGenesis performs genesis initialization for the escrow module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState escrow.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the escrow module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// RegisterServices registers a gRPC query service to respond to the escrow-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	escrow.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the escrow module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the escrow content functions used to
// simulate governance proposals, of which there are none for the escrow module.
func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams returns randomized escrow param changes for the simulator,
// of which there are none since this module doesn't use the params module.
func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange { return nil }

// RegisterStoreDecoder registers a decoder for escrow module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[escrow.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the escrow module operations with their respective weights,
// of which there are none.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
