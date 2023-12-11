package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmclient "github.com/CosmWasm/wasmd/x/wasm/client"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/streaming"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	quarantinemodule "github.com/cosmos/cosmos-sdk/x/quarantine/module"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	sanctionmodule "github.com/cosmos/cosmos-sdk/x/sanction/module"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v6"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v6/types"
	ica "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v6/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v6/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v6/modules/core/02-client/client"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v6/testing/types"

	"github.com/provenance-io/provenance/app/keepers"
	appparams "github.com/provenance-io/provenance/app/params"
	appupgrades "github.com/provenance-io/provenance/app/upgrades"
	v_1_17_0 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.17.0"
	v_1_17_0_rc1 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.17.0/rc1"
	v_1_17_0_rc2 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.17.0/rc2"
	v_1_17_0_rc3 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.17.0/rc3"
	v_1_18_0 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.18.0"
	v_1_18_0_rc1 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.18.0/rc1"
	"github.com/provenance-io/provenance/internal/antewrapper"
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/internal/statesync"
	"github.com/provenance-io/provenance/x/attribute"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/exchange"
	exchangemodule "github.com/provenance-io/provenance/x/exchange/module"
	"github.com/provenance-io/provenance/x/hold"
	holdmodule "github.com/provenance-io/provenance/x/hold/module"
	"github.com/provenance-io/provenance/x/ibchooks"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
	ibcratelimit "github.com/provenance-io/provenance/x/ibcratelimit"
	ibcratelimitmodule "github.com/provenance-io/provenance/x/ibcratelimit/module"
	"github.com/provenance-io/provenance/x/marker"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeesmodule "github.com/provenance-io/provenance/x/msgfees/module"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	"github.com/provenance-io/provenance/x/name"
	nameclient "github.com/provenance-io/provenance/x/name/client"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	oraclemodule "github.com/provenance-io/provenance/x/oracle/module"
	oracletypes "github.com/provenance-io/provenance/x/oracle/types"
	rewardmodule "github.com/provenance-io/provenance/x/reward/module"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
	triggermodule "github.com/provenance-io/provenance/x/trigger/module"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"

	_ "github.com/provenance-io/provenance/client/docs/statik" // registers swagger-ui files with statik
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// DefaultPowerReduction pio specific value for power reduction for TokensFromConsensusPower
	DefaultPowerReduction = sdk.NewIntFromUint64(1000000000)

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(append(
			wasmclient.ProposalHandlers,
			paramsclient.ProposalHandler,
			distrclient.ProposalHandler,
			upgradeclient.LegacyProposalHandler,
			upgradeclient.LegacyCancelProposalHandler,
			ibcclientclient.UpdateClientProposalHandler,
			ibcclientclient.UpgradeProposalHandler,
			nameclient.RootNameProposalHandler,
		),
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		vesting.AppModuleBasic{},
		quarantinemodule.AppModuleBasic{},
		sanctionmodule.AppModuleBasic{},

		ibc.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		ica.AppModuleBasic{},
		icq.AppModuleBasic{},
		ibchooks.AppModuleBasic{},
		ibcratelimitmodule.AppModuleBasic{},

		marker.AppModuleBasic{},
		attribute.AppModuleBasic{},
		name.AppModuleBasic{},
		metadata.AppModuleBasic{},
		wasm.AppModuleBasic{},
		msgfeesmodule.AppModuleBasic{},
		rewardmodule.AppModuleBasic{},
		triggermodule.AppModuleBasic{},
		oraclemodule.AppModuleBasic{},
		holdmodule.AppModuleBasic{},
		exchangemodule.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},

		icatypes.ModuleName:         nil,
		ibctransfertypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		ibchookstypes.ModuleName:    nil,

		attributetypes.ModuleName: nil,
		markertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		wasm.ModuleName:           {authtypes.Burner},
		rewardtypes.ModuleName:    nil,
		triggertypes.ModuleName:   nil,
		oracletypes.ModuleName:    nil,
	}

	Upgrades = []appupgrades.Upgrade{
		v_1_17_0_rc1.Upgrade,
		v_1_17_0_rc2.Upgrade,
		v_1_17_0_rc3.Upgrade,
		v_1_17_0.Upgrade,
		v_1_18_0_rc1.Upgrade,
		v_1_18_0.Upgrade,
	}
)

var (
	_ CosmosApp               = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

// SdkCoinDenomRegex returns a new sdk base denom regex string
func SdkCoinDenomRegex() string {
	return pioconfig.DefaultReDnmString
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp
	keepers.AppKeepers

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

func init() {
	DefaultNodeHome = os.ExpandEnv("$PIO_HOME")

	if strings.TrimSpace(DefaultNodeHome) == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}
		DefaultNodeHome = filepath.Join(configDir, "Provenance")
	}

	// 614,400 = 600 * 1024 = our wasm params maxWasmCodeSize value before it was removed in wasmd v0.27.
	wasmtypes.MaxWasmSize = 614_400
}

// New returns a reference to an initialized Provenance Blockchain App.
func New(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp("provenanced", logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetMsgServiceRouter(piohandlers.NewPioMsgServiceRouter(encodingConfig.TxConfig.TxDecoder()))
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	sdk.SetCoinDenomRegex(SdkCoinDenomRegex)

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
	}

	// Register State listening services.
	app.RegisterStreamingServices(appOpts)

	// Register helpers for state-sync status.
	statesync.RegisterSyncStatus()

	moduleAccountAddresses := app.ModuleAccountAddrs()

	// Setup the APP
	app.AppKeepers = keepers.NewAppKeeper(
		app.appCodec,
		app.BaseApp,
		app.legacyAmino,
		maccPerms,
		moduleAccountAddresses,
		skipUpgradeHeights,
		homePath,
		app.invCheckPeriod,
		AccountAddressPrefix,
		logger,
		app.interfaceRegistry,
		encodingConfig,
	)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx, encodingConfig.TxConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		quarantinemodule.NewAppModule(appCodec, app.QuarantineKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		sanctionmodule.NewAppModule(appCodec, app.SanctionKeeper, app.AccountKeeper, app.BankKeeper, app.GovKeeper, app.interfaceRegistry),

		// PROVENANCE
		metadata.NewAppModule(appCodec, app.MetadataKeeper, app.AccountKeeper),
		marker.NewAppModule(appCodec, app.MarkerKeeper, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.GovKeeper, app.AttributeKeeper, app.interfaceRegistry),
		name.NewAppModule(appCodec, app.NameKeeper, app.AccountKeeper, app.BankKeeper),
		attribute.NewAppModule(appCodec, app.AttributeKeeper, app.AccountKeeper, app.BankKeeper, app.NameKeeper),
		msgfeesmodule.NewAppModule(appCodec, app.MsgFeesKeeper, app.interfaceRegistry),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		rewardmodule.NewAppModule(appCodec, app.RewardKeeper, app.AccountKeeper, app.BankKeeper),
		triggermodule.NewAppModule(appCodec, app.TriggerKeeper, app.AccountKeeper, app.BankKeeper),
		app.OracleModule,
		holdmodule.NewAppModule(appCodec, app.HoldKeeper),
		exchangemodule.NewAppModule(appCodec, app.ExchangeKeeper),

		// IBC
		ibc.NewAppModule(app.IBCKeeper),
		ibcratelimitmodule.NewAppModule(appCodec, *app.RateLimitingKeeper, app.AccountKeeper, app.BankKeeper),
		ibchooks.NewAppModule(app.AccountKeeper, *app.IBCHooksKeeper),
		ibctransfer.NewAppModule(*app.TransferKeeper),
		app.ICQModule,
		app.ICAModule,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibchost.ModuleName,
		markertypes.ModuleName,
		icatypes.ModuleName,
		attributetypes.ModuleName,
		rewardtypes.ModuleName,
		triggertypes.ModuleName,

		// no-ops
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		group.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		msgfeestypes.ModuleName,
		metadatatypes.ModuleName,
		oracletypes.ModuleName,
		wasm.ModuleName,
		ibcratelimit.ModuleName,
		ibchookstypes.ModuleName,
		ibctransfertypes.ModuleName,
		icqtypes.ModuleName,
		nametypes.ModuleName,
		vestingtypes.ModuleName,
		quarantine.ModuleName,
		sanction.ModuleName,
		hold.ModuleName,
		exchange.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		icatypes.ModuleName,
		group.ModuleName,
		rewardtypes.ModuleName,
		triggertypes.ModuleName,

		// no-ops
		vestingtypes.ModuleName,
		distrtypes.ModuleName,
		authz.ModuleName,
		metadatatypes.ModuleName,
		oracletypes.ModuleName,
		nametypes.ModuleName,
		genutiltypes.ModuleName,
		ibchost.ModuleName,
		ibcratelimit.ModuleName,
		ibchookstypes.ModuleName,
		ibctransfertypes.ModuleName,
		icqtypes.ModuleName,
		msgfeestypes.ModuleName,
		wasm.ModuleName,
		slashingtypes.ModuleName,
		upgradetypes.ModuleName,
		attributetypes.ModuleName,
		capabilitytypes.ModuleName,
		evidencetypes.ModuleName,
		banktypes.ModuleName,
		minttypes.ModuleName,
		markertypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		quarantine.ModuleName,
		sanction.ModuleName,
		hold.ModuleName,
		exchange.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		markertypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		group.ModuleName,
		feegrant.ModuleName,
		quarantine.ModuleName,
		sanction.ModuleName,

		nametypes.ModuleName,
		attributetypes.ModuleName,
		metadatatypes.ModuleName,
		msgfeestypes.ModuleName,
		hold.ModuleName,
		exchange.ModuleName, // must be after the hold module.

		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		icqtypes.ModuleName,
		icatypes.ModuleName,
		ibcratelimit.ModuleName,
		ibchookstypes.ModuleName,
		// wasm after ibc transfer
		wasm.ModuleName,
		rewardtypes.ModuleName,
		triggertypes.ModuleName,
		oracletypes.ModuleName,

		// no-ops
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		upgradetypes.ModuleName,
	)

	app.mm.SetOrderMigrations(
		banktypes.ModuleName,
		authz.ModuleName,
		group.ModuleName,
		capabilitytypes.ModuleName,
		crisistypes.ModuleName,
		distrtypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		genutiltypes.ModuleName,
		govtypes.ModuleName,
		ibchost.ModuleName,
		minttypes.ModuleName,
		paramstypes.ModuleName,
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		quarantine.ModuleName,
		sanction.ModuleName,
		hold.ModuleName,
		exchange.ModuleName,

		ibcratelimit.ModuleName,
		ibchookstypes.ModuleName,
		icatypes.ModuleName,
		icqtypes.ModuleName,
		wasm.ModuleName,

		attributetypes.ModuleName,
		markertypes.ModuleName,
		msgfeestypes.ModuleName,
		metadatatypes.ModuleName,
		nametypes.ModuleName,
		rewardtypes.ModuleName,
		triggertypes.ModuleName,
		oracletypes.ModuleName,

		// Last due to v0.44 issue: https://github.com/cosmos/cosmos-sdk/issues/10591
		authtypes.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
	app.configurator = module.NewConfigurator(app.appCodec, app.BaseApp.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		params.NewAppModule(app.ParamsKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		quarantinemodule.NewAppModule(appCodec, app.QuarantineKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		sanctionmodule.NewAppModule(appCodec, app.SanctionKeeper, app.AccountKeeper, app.BankKeeper, app.GovKeeper, app.interfaceRegistry),

		metadata.NewAppModule(appCodec, app.MetadataKeeper, app.AccountKeeper),
		marker.NewAppModule(appCodec, app.MarkerKeeper, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.GovKeeper, app.AttributeKeeper, app.interfaceRegistry),
		name.NewAppModule(appCodec, app.NameKeeper, app.AccountKeeper, app.BankKeeper),
		attribute.NewAppModule(appCodec, app.AttributeKeeper, app.AccountKeeper, app.BankKeeper, app.NameKeeper),
		msgfeesmodule.NewAppModule(appCodec, app.MsgFeesKeeper, app.interfaceRegistry),
		rewardmodule.NewAppModule(appCodec, app.RewardKeeper, app.AccountKeeper, app.BankKeeper),
		triggermodule.NewAppModule(appCodec, app.TriggerKeeper, app.AccountKeeper, app.BankKeeper),
		oraclemodule.NewAppModule(appCodec, app.OracleKeeper, app.AccountKeeper, app.BankKeeper, app.IBCKeeper.ChannelKeeper),
		holdmodule.NewAppModule(appCodec, app.HoldKeeper),
		exchangemodule.NewAppModule(appCodec, app.ExchangeKeeper),
		provwasm.NewWrapper(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.NameKeeper),

		// IBC
		ibc.NewAppModule(app.IBCKeeper),
		ibcratelimitmodule.NewAppModule(appCodec, *app.RateLimitingKeeper, app.AccountKeeper, app.BankKeeper),
		ibchooks.NewAppModule(app.AccountKeeper, *app.IBCHooksKeeper),
		ibctransfer.NewAppModule(*app.TransferKeeper),
		app.ICAModule,
	)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	anteHandler, err := antewrapper.NewAnteHandler(
		antewrapper.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			FeegrantKeeper:  app.FeeGrantKeeper,
			MsgFeesKeeper:   app.MsgFeesKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		})
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	msgfeehandler, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  app.AccountKeeper,
		BankKeeper:     app.BankKeeper,
		FeegrantKeeper: app.FeeGrantKeeper,
		MsgFeesKeeper:  app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})

	if err != nil {
		panic(err)
	}
	app.SetFeeHandler(msgfeehandler)

	app.SetEndBlocker(app.EndBlocker)

	app.SetAggregateEventsFunc(piohandlers.AggregateEvents)

	// Add upgrade plans for each release. This must be done before the baseapp seals via LoadLatestVersion() down below.
	appupgrades.InstallCustomUpgradeHandlers(app, Upgrades)
	appupgrades.AttemptUpgradeStoreLoaders(app, app.Keepers(), Upgrades)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

// GetBaseApp returns the base cosmos app
func (app *App) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetStakingKeeper returns the staking keeper (for ibc testing)
func (app *App) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetIBCKeeper returns the ibc keeper (for ibc testing)
func (app *App) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper // This is a *ibckeeper.Keeper
}

// GetScopedIBCKeeper returns the scoped ibc keeper (for ibc testing)
func (app *App) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig implements the TestingApp interface (for ibc testing).
func (app *App) GetTxConfig() client.TxConfig {
	return MakeEncodingConfig().TxConfig
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns Provenance's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Provenance's InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// RegisterNodeService registers the node query server.
func (app *App) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// RegisterStreamingServices registers types.ABCIListener State Listening services with the App.
func (app *App) RegisterStreamingServices(appOpts servertypes.AppOptions) {
	// register streaming services
	streamingCfg := cast.ToStringMap(appOpts.Get(baseapp.StreamingTomlKey))
	for service := range streamingCfg {
		pluginKey := fmt.Sprintf("%s.%s.%s", baseapp.StreamingTomlKey, service, baseapp.StreamingABCIPluginTomlKey)
		pluginName := strings.TrimSpace(cast.ToString(appOpts.Get(pluginKey)))
		if len(pluginName) > 0 {
			logLevel := cast.ToString(appOpts.Get(flags.FlagLogLevel))
			plugin, err := streaming.NewStreamingPlugin(pluginName, logLevel)
			if err != nil {
				app.Logger().Error("failed to load streaming plugin", "error", err)
				os.Exit(1)
			}
			if err := baseapp.RegisterStreamingPlugin(app.BaseApp, appOpts, app.GetKVStoreKey(), plugin); err != nil {
				app.Logger().Error("failed to register streaming plugin", "error", err)
				os.Exit(1)
			}
			app.Logger().Info("streaming service registered", "service", service, "plugin", pluginName)
		}
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// injectUpgrade causes the named upgrade to be run as the chain starts.
//
// To use this, add a call to it in New after the call to InstallCustomUpgradeHandlers
// but before the line that looks for an upgrade file.
//
// This function is for testing an upgrade against an existing chain's data (e.g. mainnet).
// Here's how:
//  1. Run a node for the chain you want to test using its normal release.
//  2. In this provenance repo, check out the branch/version with the upgrade you want to test.
//  3. Add a call to this function as described above, e.g. injectUpgrade("ochre").
//  4. Compile it with `make build`.
//  5. I suggest renaming build/provenanced to something like provenanced-ochre-force-upgrade
//     and moving it somewhere handier for your node.
//  6. Stop your node.
//  7. Back up your data directory because we're about to mess it up.
//  8. Seriously, your data directory will need to be thrown away after this.
//  9. Restart your node with `--log_level debug` using your new force-upgrade binary.
//
// As the node starts, it should think that an upgrade is needed and attempt to execute it.
// If the upgrade finishes successfully, your node will then try and fail to sync with the rest of the nodes.
// Your chain now has a different state than the rest of the network and will be generating different hashes.
// There's no reason to let it continue to run.
//
// Deprecated:  This function should never be called in anything that gets merged into main or any sort of release branch.
// It's marked as deprecated so that things can complain about its use (e.g. the linter).
func (app *App) injectUpgrade(name string) { //nolint:unused // This is designed to only be used in unofficial code.
	plan := upgradetypes.Plan{
		Name:   name,
		Height: app.LastBlockHeight() + 1,
	}
	// Write the plan to $home/data/upgrade-info.json
	if err := app.UpgradeKeeper.DumpUpgradeInfoToDisk(plan.Height, plan); err != nil {
		panic(err)
	}
	// Define a new BeginBlocker that will inject the upgrade.
	injected := false
	app.SetBeginBlocker(func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		if !injected {
			app.Logger().Info("Injecting upgrade plan", "plan", plan)
			// Ideally, we'd just call ScheduleUpgrade(ctx, plan) here (and panic on error).
			// But the upgrade keeper has a migration in v0.46 that changes some store key stuff.
			// ScheduleUpgrade tries to read some of that changed store stuff and fails if the migration hasn't
			// been applied yet. So we're doing things the hard way here.
			app.UpgradeKeeper.ClearUpgradePlan(ctx)
			store := ctx.KVStore(app.GetKey(upgradetypes.StoreKey))
			bz := app.appCodec.MustMarshal(&plan)
			store.Set(upgradetypes.PlanKey(), bz)
			injected = true
		}
		return app.BeginBlocker(ctx, req)
	})
}

func (app *App) ModuleManager() *module.Manager {
	return app.mm
}

func (app *App) Configurator() module.Configurator {
	return app.configurator
}

func (app *App) Keepers() *keepers.AppKeepers {
	return &app.AppKeepers
}
