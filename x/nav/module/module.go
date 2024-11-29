package module

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/provenance-io/provenance/x/nav"
	"github.com/provenance-io/provenance/x/nav/keeper"
)

var (
	_ module.AppModuleBasic = (*AppModuleBasic)(nil)
	_ module.AppModuleBasic = (*AppModule)(nil)
	_ appmodule.AppModule   = (*AppModule)(nil)
)

type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns "nav", the name of this module.
func (AppModuleBasic) Name() string {
	return nav.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the NAV module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	panic(nav.NotYetImplemented)
	// return cdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the NAV module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	panic(nav.NotYetImplemented)
	// var data GenesisState
	// if err := cdc.UnmarshalJSON(bz, &data); err != nil {
	// 	return fmt.Errorf("failed to unmarshal %s genesis state: %w", ModuleName, err)
	// }
	// return data.Validate()
}

// GetQueryCmd returns the cli query commands for the NAV module.
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	panic(nav.NotYetImplemented)
	// return cli.CmdQuery()
}

// GetTxCmd returns the transaction commands for the NAV module.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	panic(nav.NotYetImplemented)
	// return cli.CmdTx()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the NAV module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *runtime.ServeMux) {
	panic(nav.NotYetImplemented)
	// if err := RegisterQueryHandlerClient(context.Background(), mux, NewQueryClient(clientCtx)); err != nil {
	// 	panic(err)
	// }
}

// RegisterInterfaces registers the NAV module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	panic(nav.NotYetImplemented)
	// RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec registers the NAV module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, navKeeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         navKeeper,
	}
}

// IsOnePerModuleType is a dummy function that satisfies the OnePerModuleType interface (needed by AppModule).
func (AppModule) IsOnePerModuleType() {}

// IsAppModule is a dummy function that satisfies the AppModule interface.
func (AppModule) IsAppModule() {}

// RegisterInvariants registers the invariants for the NAV module.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs genesis initialization for the NAV module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	panic(nav.NotYetImplemented)
	// var genesisState GenesisState
	// cdc.MustUnmarshalJSON(data, &genesisState)
	// am.keeper.InitGenesis(ctx, &genesisState)
	// return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the NAV module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	panic(nav.NotYetImplemented)
	// gs := am.keeper.ExportGenesis(ctx)
	// return cdc.MustMarshalJSON(gs)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	panic(nav.NotYetImplemented)
	// RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	// RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
