package module

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	abci "github.com/cometbft/cometbft/abci/types"
	client "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/provenance-io/provenance/x/ledger/client/cli"
	"github.com/provenance-io/provenance/x/ledger/keeper"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

var (
	_ module.AppModuleBasic = (*AppModuleBasic)(nil)
	_ appmodule.AppModule   = (*AppModule)(nil)
)

type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module name.
func (AppModuleBasic) Name() string {
	return ledger.ModuleName
}

// GetTxCmd returns the root tx command for the attribute module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.CmdTx()
}

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.CmdQuery()
}

func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// Register the protobuf message types and services with the sdk.
func (AppModuleBasic) RegisterInterfaces(registry types.InterfaceRegistry) {
	ledger.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&ledger.GenesisState{})
}

// ValidateGenesis validates the genesis state.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(ctx client.Context, mux *runtime.ServeMux) {
	err := ledger.RegisterQueryHandlerClient(context.Background(), mux, ledger.NewQueryClient(ctx))
	if err != nil {
		panic(err)
	}
}

// AppModule implements an application module for mymodule.
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
	var genState ledger.GenesisState
	// Initialize default genesis state
	if gs == nil {
		genState = ledger.GenesisState{}
	} else {
		cdc.MustUnmarshalJSON(gs, &genState)
	}

	am.keeper.InitGenesis(ctx, &genState)
	return nil
}

// ExportGenesis exports the module's genesis state.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// Satisfy the AppModule interface.
func (AppModule) IsAppModule()        {}
func (AppModule) IsOnePerModuleType() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	ledger.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	ledger.RegisterQueryServer(cfg.QueryServer(), keeper.NewLedgerQueryServer(am.keeper))
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
