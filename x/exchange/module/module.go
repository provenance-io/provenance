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

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
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

func NewAppModule(cdc codec.Codec, exchangeKeeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         exchangeKeeper,
	}
}

type AppModuleBasic struct {
	cdc codec.Codec
}

func (AppModuleBasic) Name() string {
	return exchange.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the exchange module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(exchange.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the exchange module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data exchange.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", exchange.ModuleName, err)
	}
	return data.Validate()
}

// GetQueryCmd returns the cli query commands for the exchange module.
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	// TODO[1658]: Create cli.QueryCmd()
	panic("not implemented")
	//return cli.QueryCmd()
}

// GetTxCmd returns the transaction commands for the exchange module.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	// TODO[1658]: Create cli.TxCmd()
	panic("not implemented")
	//return cli.TxCmd()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the exchange module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	if err := exchange.RegisterQueryHandlerClient(context.Background(), mux, exchange.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers the exchange module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	exchange.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec registers the exchange module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterInvariants registers the invariants for the exchange module.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Deprecated: Route returns the message routing key for the exchange module, empty.
func (am AppModule) Route() sdk.Route { return sdk.Route{} }

// Deprecated: QuerierRoute returns the route we respond to for abci queries, "".
func (AppModule) QuerierRoute() string { return "" }

// Deprecated: LegacyQuerierHandler returns the exchange module sdk.Querier (nil).
func (am AppModule) LegacyQuerierHandler(_ *codec.LegacyAmino) sdk.Querier { return nil }

// InitGenesis performs genesis initialization for the exchange module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState exchange.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	// TODO[1658]: Create keeper.InitGenesis
	panic("not implemented")
	//am.keeper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the exchange module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// TODO[1658]: Create keeper.ExportGenesis
	panic("not implemented")
	//gs := am.keeper.ExportGenesis(ctx)
	//return cdc.MustMarshalJSON(gs)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	exchange.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	// TODO[1658]: Uncomment this once the keeper implements the query server.
	// exchange.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the exchange module.
func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	// TODO[1658]: Create simulation.RandomizedGenState(simState)
	panic("not implemented")
	//simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the exchange content functions used to
// simulate governance proposals, of which there are none for the exchange module.
func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams returns randomized exchange param changes for the simulator,
// of which there are none since this module doesn't use the params module.
func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange { return nil }

// RegisterStoreDecoder registers a decoder for exchange module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	// TODO[1658]: Create simulation.NewDecodeStore(am.cdc)
	panic("not implemented")
	// sdr[exchange.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the exchange module operations with their respective weights,
// of which there are none.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	// TODO[1658]: Create the WeightedOperations.
	panic("not implemented")
}
