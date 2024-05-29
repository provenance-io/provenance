package module

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/provenance-io/provenance/x/ibcratelimit"
	ibcratelimitcli "github.com/provenance-io/provenance/x/ibcratelimit/client/cli"
	"github.com/provenance-io/provenance/x/ibcratelimit/keeper"
	"github.com/provenance-io/provenance/x/ibcratelimit/simulation"
)

var (
	_ module.AppModuleBasic      = (*AppModule)(nil)
	_ module.AppModuleSimulation = (*AppModule)(nil)
	_ module.HasProposalMsgs     = (*AppModule)(nil)

	_ appmodule.AppModule = (*AppModule)(nil)
)

// AppModuleBasic defines the basic application module used by the ibcratelimit module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the ibcratelimit module's name.
func (AppModuleBasic) Name() string { return ibcratelimit.ModuleName }

// RegisterLegacyAminoCodec registers the ibcratelimit module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {
}

// RegisterInterfaces registers interfaces and implementations of the ibcratelimit module.
func (AppModuleBasic) RegisterInterfaces(cdc codectypes.InterfaceRegistry) {
	ibcratelimit.RegisterInterfaces(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the ibcratelimit
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(ibcratelimit.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the ibcratelimit module.
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genState ibcratelimit.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", ibcratelimit.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterRESTRoutes registers the REST routes for the ibcratelimit module.
// Deprecated: RegisterRESTRoutes is deprecated.
func (b AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the ibcratelimit module.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := ibcratelimit.RegisterQueryHandlerClient(context.Background(), mux, ibcratelimit.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetQueryCmd returns the cli query commands for the ibcratelimit module
func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return ibcratelimitcli.GetQueryCmd()
}

// GetTxCmd returns the transaction commands for the ibcratelimit module
func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return ibcratelimitcli.NewTxCmd()
}

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsOnePerModuleType is a dummy function that satisfies the OnePerModuleType interface (needed by AppModule).
func (AppModule) IsOnePerModuleType() {}

// IsAppModule is a dummy function that satisfies the AppModule interface.
func (AppModule) IsAppModule() {}

// GenerateGenesisState creates a randomized GenState of the ibcratelimit module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RandomizedParams returns randomized module parameters for param change proposals.
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.LegacyParamChange {
	return nil
}

// RegisterStoreDecoder registers a func to decode each module's defined types from their corresponding store key
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[ibcratelimit.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns simulation operations (i.e msgs) with their respective weight
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}

// ProposalMsgs returns all the msgs to execute as governance proposals.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs(simState, &am.keeper)
}

// Name returns the ibcratelimit module's name.
func (AppModule) Name() string {
	return ibcratelimit.ModuleName
}

// RegisterInvariants does nothing, there are no invariants to enforce
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the txfees module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState ibcratelimit.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)
	am.keeper.InitGenesis(ctx, &genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the txfees module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	ibcratelimit.RegisterQueryServer(cfg.QueryServer(), am.keeper)
	ibcratelimit.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
}
