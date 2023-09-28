package ibcratelimitmodule

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"

	ibcratelimit "github.com/provenance-io/provenance/x/ibcratelimit"
	ibcratelimitclient "github.com/provenance-io/provenance/x/ibcratelimit/client"
	ibcratelimitcli "github.com/provenance-io/provenance/x/ibcratelimit/client/cli"
	"github.com/provenance-io/provenance/x/ibcratelimit/client/grpc"
	"github.com/provenance-io/provenance/x/ibcratelimit/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the ibcratelimit module.
type AppModuleBasic struct {
}

// Name returns the ibcratelimit module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the ibcratelimit module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
}

// RegisterInterfaces registers interfaces and implementations of the ibcratelimit module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// TODO Do we need to register interfaces?
}

// DefaultGenesis returns default genesis state as raw bytes for the ibcratelimit
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the ibcratelimit module.
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterRESTRoutes registers the REST routes for the ibcratelimit module.
// Deprecated: RegisterRESTRoutes is deprecated.
func (b AppModuleBasic) RegisterRESTRoutes(ctx client.Context, r *mux.Router) {
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the ibcratelimit module.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)) //nolint:errcheck
}

// GetQueryCmd returns the cli query commands for the ibcratelimit module
func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return ibcratelimitcli.GetQueryCmd()
}

// GetTxCmd returns the transaction commands for the ibcratelimit module
func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	AppModuleBasic
	ics4wrapper ibcratelimit.ICS4Wrapper
}

// NewAppModule creates a new AppModule object
func NewAppModule(ics4wrapper ibcratelimit.ICS4Wrapper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		ics4wrapper:    ics4wrapper,
	}
}

// GenerateGenesisState creates a randomized GenState of the ibcratelimit module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	// Todo When we finish simulation
	// simulation.RandomizedGenState(simState)
}

// ProposalContents returns content functions used to simulate governance proposals.
func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	// currently no gov proposals exist
	return nil
}

// RandomizedParams returns randomized module parameters for param change proposals.
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	// currently no module params exist
	return nil
}

// RegisterStoreDecoder registers a func to decode each module's defined types from their corresponding store key
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	// TODO When we finish sim tests
	// sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns simulation operations (i.e msgs) with their respective weight
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return []simtypes.WeightedOperation{}
	// TODO When we get sim tests
	/*return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, am.keeper, am.accountKeeper, am.bankKeeper,
	)*/
}

// Name returns the ibcratelimit module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants does nothing, there are no invariants to enforce
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Deprecated: Route returns the message routing key for the ibcratelimit module.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the route we respond to for abci queries
func (AppModule) QuerierRoute() string { return "" }

// LegacyQuerierHandler returns the ibcratelimit module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier {
	return nil
}

// InitGenesis performs the txfees module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)
	am.ics4wrapper.InitGenesis(ctx, genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the txfees module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.ics4wrapper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes all ABCI BeginBlock logic respective to the ibcratelimit module.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock executes all ABCI EndBlock logic respective to the ibcratelimit module. It
// returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), grpc.Querier{Q: ibcratelimitclient.Querier{K: am.ics4wrapper}})
}
