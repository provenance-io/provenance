package reward

import (
	"context"
	"encoding/json"
	rewardModule "github.com/provenance-io/provenance/x/reward"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	//cli "github.com/provenance-io/provenance/x/reward/client/cli"

	"github.com/provenance-io/provenance/x/reward/keeper"
	reward "github.com/provenance-io/provenance/x/reward/types"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the reward module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the reward module's name.
func (AppModuleBasic) Name() string {
	return reward.ModuleName
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// TODO
	reward.RegisterQueryServer(cfg.QueryServer(), &reward.UnimplementedQueryServer{})
}

// RegisterLegacyAminoCodec registers the reward module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the reward module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	reward.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the reward
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(reward.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the reward module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data reward.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return sdkerrors.Wrapf(err, "failed to unmarshal %q genesis state", reward.ModuleName)
	}

	return data.Validate()
}

// RegisterRESTRoutes registers the REST routes for the reward module.
// Deprecated: RegisterRESTRoutes is deprecated.
func (AppModuleBasic) RegisterRESTRoutes(_ sdkclient.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the reward module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	if err := reward.RegisterQueryHandlerClient(context.Background(), mux, reward.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetQueryCmd returns the cli query commands for the reward module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	// TODO
	// return cli.GetQueryCmd()
	return nil
}

// GetTxCmd returns the transaction commands for the reward module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	AppModuleBasic
	keeper   keeper.Keeper
	registry cdctypes.InterfaceRegistry
}

func (am AppModule) GenerateGenesisState(input *module.SimulationState) {
	panic("implement me")
}

func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	panic("implement me")
}

func (am AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	panic("implement me")
}

func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	// sdr[keeper.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	panic("implement me")
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		registry:       registry,
	}
}

// Name returns the reward module's name.
func (AppModule) Name() string {
	return reward.ModuleName
}

// RegisterInvariants does nothing, there are no invariants to enforce
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Deprecated: Route returns the message routing key for the reward module.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

func (am AppModule) NewHandler() sdk.Handler {
	return nil
}

// QuerierRoute returns the route we respond to for abci queries
func (AppModule) QuerierRoute() string { return "" }

// LegacyQuerierHandler returns the reward module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// InitGenesis performs genesis initialization for the reward module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState reward.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	// TODO
	// keeper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the reward
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// TODO
	// gs := keeper.ExportGenesis(ctx)
	// return cdc.MustMarshalJSON(gs)
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	// TODO
	// rewardmodule.BeginBlocker(ctx, am.keeper)
}

// EndBlock does nothing
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	rewardModule.EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}
}
