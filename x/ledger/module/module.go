package module

import (
	"context"
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	client "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/provenance-io/provenance/x/ledger"
	"github.com/provenance-io/provenance/x/ledger/client/cli"
	"github.com/provenance-io/provenance/x/ledger/keeper"
	"github.com/spf13/cobra"
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

// RegisterCodec registers the module's types. (In newer SDK versions, use RegisterLegacyAminoCodec.)
func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {
	// Register any concrete types if needed.
}

// AppModule implements an application module for mymodule.
type AppModule struct {
	AppModuleBasic
	keeper keeper.BaseKeeper
}

// NewAppModule creates a new AppModule instance.
func NewAppModule(cdc codec.Codec, k keeper.BaseKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

// InitGenesis initializes the genesis state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	// no-op: we start with a clean ledger state.
	return nil
}

// DefaultGenesis returns default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&ledger.GenesisState{})
}

// ValidateGenesis validates the genesis state.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// ExportGenesis exports the module's genesis state.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// TODO wire up the export genesis.
	return cdc.MustMarshalJSON(&ledger.GenesisState{})
}

// Satisfy the AppModule interface for now..
func (AppModule) IsAppModule()                                {}
func (AppModule) IsOnePerModuleType()                         {}
func (AppModule) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	ledger.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))

	ledger.RegisterQueryServer(cfg.QueryServer(), keeper.NewLedgerQueryServer(am.keeper))
}

// Register the protobuf message types and services with the sdk.
func (AppModule) RegisterInterfaces(registry types.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &ledger.Msg_serviceDesc)
}

func (AppModule) RegisterGRPCGatewayRoutes(ctx client.Context, mux *runtime.ServeMux) {
	err := ledger.RegisterQueryHandlerClient(context.Background(), mux, ledger.NewQueryClient(ctx))
	if err != nil {
		panic(err)
	}
}
