package module

import (
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/client/v2/autocli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/provenance-io/provenance/x/smartaccounts/client/cli"
	"github.com/provenance-io/provenance/x/smartaccounts/keeper"
	"github.com/provenance-io/provenance/x/smartaccounts/simulation"
	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

const (
	ConsensusVersion = 1
)

// ModuleAccountAddress defines the x/accounts module address.

var (
	_ module.AppModuleBasic         = AppModuleBasic{}
	_ module.AppModuleGenesis       = AppModule{}
	_ module.AppModule              = AppModule{}
	_ autocli.HasCustomQueryCommand = AppModule{}
	_ autocli.HasCustomTxCommand    = AppModule{}
)

type AppModule struct {
	AppModuleBasic
	smartaccountkeeper keeper.Keeper
	bankkeeper         bankkeeper.Keeper
	registry           codectypes.InterfaceRegistry
}

func (a AppModule) InitGenesis(ctx sdk.Context, marshaler codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	marshaler.MustUnmarshalJSON(message, &genesisState)
	err := a.smartaccountkeeper.ImportGenesis(ctx, &genesisState)
	if err != nil {
		panic(err)
	}
	return nil
}

func (a AppModule) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}
func (a AppModule) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// NewAppModule constructor
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	registry codectypes.InterfaceRegistry,
) AppModule {
	return AppModule{
		AppModuleBasic:     AppModuleBasic{cdc: cdc},
		smartaccountkeeper: keeper,
		bankkeeper:         bankKeeper,
		registry:           registry,
	}
}

// IsOnePerModuleType is a dummy function that satisfies the OnePerModuleType interface (needed by AppModule).
func (a AppModule) IsOnePerModuleType() {
}

func (a AppModule) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	return a.cdc.MustMarshalJSON(keeper.DefaultGenesisState())
}

func (a AppModule) ValidateGenesis(marshaler codec.JSONCodec, _ client.TxEncodingConfig, message json.RawMessage) error {
	var data types.GenesisState
	err := marshaler.UnmarshalJSON(message, &data)
	if err != nil {
		return err
	}
	return data.ValidateBasic()
}

func (a AppModule) ExportGenesis(ctx sdk.Context, marshaler codec.JSONCodec) json.RawMessage {
	gs, err := a.smartaccountkeeper.ExportGenesis(ctx)
	if err != nil {
		return nil
	}
	return marshaler.MustMarshalJSON(gs)
}

// RegisterServices registers module services.
func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(a.smartaccountkeeper))
	types.RegisterQueryServer(cfg.QueryServer(), a.smartaccountkeeper)
}

type AppModuleBasic struct {
	cdc codec.Codec
}

func (a AppModuleBasic) Name() string {
	return types.ModuleName
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {
}

func (a AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (AppModule) IsAppModule() {}

// Name returns the module's name.
// Deprecated: kept for legacy reasons.
func (a AppModule) Name() string { return types.ModuleName }

// ConsensusVersion is a sequence number for state-breaking change of the
// module. It should be incremented on each consensus-breaking change
// introduced by the module. To avoid wrong/empty versions, the initial version
// should be set to 1.
func (a AppModule) ConsensusVersion() uint64 {
	return ConsensusVersion
}

// WeightedOperations returns the all the marker module operations with their respective weights.
func (a AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState, a.smartaccountkeeper, a.bankkeeper, codec.NewProtoCodec(a.registry),
	)
}
