package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtos "github.com/cometbft/cometbft/libs/os"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	// "cosmossdk.io/store/streaming" // TODO[1760]: streaming: See if we can use this directly or if we have needed modifications.
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	// icq "github.com/cosmos/ibc-apps/modules/async-icq/v7"              // TODO[1760]: async-icq
	// icqkeeper "github.com/cosmos/ibc-apps/modules/async-icq/v7/keeper" // TODO[1760]: async-icq
	// icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v7/types"   // TODO[1760]: async-icq
	// "github.com/cosmos/cosmos-sdk/x/quarantine" // TODO[1760]: quarantine
	// quarantinekeeper "github.com/cosmos/cosmos-sdk/x/quarantine/keeper" // TODO[1760]: quarantine
	// quarantinemodule "github.com/cosmos/cosmos-sdk/x/quarantine/module" // TODO[1760]: quarantine
	// "github.com/cosmos/cosmos-sdk/x/sanction" // TODO[1760]: sanction
	// sanctionkeeper "github.com/cosmos/cosmos-sdk/x/sanction/keeper" // TODO[1760]: sanction
	// sanctionmodule "github.com/cosmos/cosmos-sdk/x/sanction/module" // TODO[1760]: sanction

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"

	// icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host" // TODO[1760]: msg-service-router
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v8/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	appparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/antewrapper"
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/internal/statesync"
	"github.com/provenance-io/provenance/x/attribute"
	attributekeeper "github.com/provenance-io/provenance/x/attribute/keeper"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	attributewasm "github.com/provenance-io/provenance/x/attribute/wasm"
	"github.com/provenance-io/provenance/x/exchange"
	exchangekeeper "github.com/provenance-io/provenance/x/exchange/keeper"
	exchangemodule "github.com/provenance-io/provenance/x/exchange/module"
	"github.com/provenance-io/provenance/x/hold"
	holdkeeper "github.com/provenance-io/provenance/x/hold/keeper"
	holdmodule "github.com/provenance-io/provenance/x/hold/module"
	"github.com/provenance-io/provenance/x/ibchooks"
	ibchookskeeper "github.com/provenance-io/provenance/x/ibchooks/keeper"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	ibcratelimitkeeper "github.com/provenance-io/provenance/x/ibcratelimit/keeper"
	ibcratelimitmodule "github.com/provenance-io/provenance/x/ibcratelimit/module"
	"github.com/provenance-io/provenance/x/marker"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	markerwasm "github.com/provenance-io/provenance/x/marker/wasm"
	"github.com/provenance-io/provenance/x/metadata"
	metadatakeeper "github.com/provenance-io/provenance/x/metadata/keeper"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	metadatawasm "github.com/provenance-io/provenance/x/metadata/wasm"
	"github.com/provenance-io/provenance/x/msgfees"
	msgfeeskeeper "github.com/provenance-io/provenance/x/msgfees/keeper"
	msgfeesmodule "github.com/provenance-io/provenance/x/msgfees/module"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	msgfeeswasm "github.com/provenance-io/provenance/x/msgfees/wasm"
	"github.com/provenance-io/provenance/x/name"
	nameclient "github.com/provenance-io/provenance/x/name/client"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	namewasm "github.com/provenance-io/provenance/x/name/wasm"
	oraclekeeper "github.com/provenance-io/provenance/x/oracle/keeper"
	oraclemodule "github.com/provenance-io/provenance/x/oracle/module"
	oracletypes "github.com/provenance-io/provenance/x/oracle/types"
	rewardkeeper "github.com/provenance-io/provenance/x/reward/keeper"
	rewardmodule "github.com/provenance-io/provenance/x/reward/module"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
	triggerkeeper "github.com/provenance-io/provenance/x/trigger/keeper"
	triggermodule "github.com/provenance-io/provenance/x/trigger/module"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"

	_ "github.com/provenance-io/provenance/client/docs/statik" // registers swagger-ui files with statik
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// DefaultPowerReduction pio specific value for power reduction for TokensFromConsensusPower
	DefaultPowerReduction = sdkmath.NewIntFromUint64(1_000_000_000)

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
			[]govclient.ProposalHandler{},
			paramsclient.ProposalHandler,
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
		// quarantinemodule.AppModuleBasic{}, // TODO[1760]: quarantine
		// sanctionmodule.AppModuleBasic{}, // TODO[1760]: sanction

		ibc.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		ica.AppModuleBasic{},
		// icq.AppModuleBasic{}, // TODO[1760]: async-icq
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
)

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

// WasmWrapper allows us to use namespacing in the config file
// This is only used for parsing in the app, x/wasm expects WasmConfig
type WasmWrapper struct {
	Wasm wasm.Config `mapstructure:"wasm"`
}

// SdkCoinDenomRegex returns a new sdk base denom regex string
func SdkCoinDenomRegex() string {
	return pioconfig.DefaultReDnmString
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    *stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	AuthzKeeper      authzkeeper.Keeper
	GroupKeeper      groupkeeper.Keeper
	EvidenceKeeper   evidencekeeper.Keeper
	FeeGrantKeeper   feegrantkeeper.Keeper
	MsgFeesKeeper    msgfeeskeeper.Keeper
	RewardKeeper     rewardkeeper.Keeper
	// QuarantineKeeper quarantinekeeper.Keeper // TODO[1760]: quarantine
	// SanctionKeeper sanctionkeeper.Keeper // TODO[1760]: sanction
	TriggerKeeper triggerkeeper.Keeper
	OracleKeeper  oraclekeeper.Keeper

	IBCKeeper      *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCHooksKeeper *ibchookskeeper.Keeper
	ICAHostKeeper  *icahostkeeper.Keeper
	TransferKeeper *ibctransferkeeper.Keeper
	// ICQKeeper          icqkeeper.Keeper // TODO[1760]: async-icq
	RateLimitingKeeper *ibcratelimitkeeper.Keeper

	MarkerKeeper    markerkeeper.Keeper
	MetadataKeeper  metadatakeeper.Keeper
	AttributeKeeper attributekeeper.Keeper
	NameKeeper      namekeeper.Keeper
	HoldKeeper      holdkeeper.Keeper
	ExchangeKeeper  exchangekeeper.Keeper
	WasmKeeper      *wasm.Keeper
	ContractKeeper  *wasmkeeper.PermissionedKeeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper
	ScopedICQKeeper      capabilitykeeper.ScopedKeeper
	ScopedOracleKeeper   capabilitykeeper.ScopedKeeper

	TransferStack    *ibchooks.IBCMiddleware
	Ics20WasmHooks   *ibchooks.WasmHooks
	Ics20MarkerHooks *ibchooks.MarkerHooks
	IbcHooks         *ibchooks.IbcHooks
	HooksICS4Wrapper ibchooks.ICS4Middleware

	// the module manager
	mm                 *module.Manager
	BasicModuleManager module.BasicManager

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
	// TODO[1760]: msg-service-router: Switch back to the PioMsgServiceRouter.
	// bApp.SetMsgServiceRouter(piohandlers.NewPioMsgServiceRouter(encodingConfig.TxConfig.TxDecoder()))
	bApp.SetMsgServiceRouter(baseapp.NewMsgServiceRouter()) // TODO[1760]: msg-service-router: delete this line.
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	sdk.SetCoinDenomRegex(SdkCoinDenomRegex)

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, capabilitytypes.StoreKey,
		authzkeeper.StoreKey, group.StoreKey,

		// ibchost.StoreKey, // TODO[1760]: ibc-host
		ibctransfertypes.StoreKey,
		icahosttypes.StoreKey,
		// icqtypes.StoreKey, // TODO[1760]: async-icq
		ibchookstypes.StoreKey,
		ibcratelimit.StoreKey,

		metadatatypes.StoreKey,
		markertypes.StoreKey,
		attributetypes.StoreKey,
		nametypes.StoreKey,
		msgfeestypes.StoreKey,
		wasm.StoreKey,
		rewardtypes.StoreKey,
		// quarantine.StoreKey, // TODO[1760]: quarantine
		// sanction.StoreKey, // TODO[1760]: sanction
		triggertypes.StoreKey,
		oracletypes.StoreKey,
		hold.StoreKey,
		exchange.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	govAuthority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	// Register State listening services.
	app.RegisterStreamingServices(appOpts)

	// Register helpers for state-sync status.
	statesync.RegisterSyncStatus()

	app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	// bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())) // TODO[1760]: params

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule("ibc") // TODO[1760]: ibc-host: was ibchost.ModuleName
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasm.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	// scopedICQKeeper := app.CapabilityKeeper.ScopeToModule(icqtypes.ModuleName) // TODO[1760]: async-icq
	scopedOracleKeeper := app.CapabilityKeeper.ScopeToModule(oracletypes.ModuleName)

	// capability keeper must be sealed after scope to module registrations are completed.
	app.CapabilityKeeper.Seal()

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(appCodec, runtime.NewKVStoreService(keys[authtypes.StoreKey]), authtypes.ProtoBaseAccount, maccPerms, authcodec.NewBech32Codec(AccountAddressPrefix), AccountAddressPrefix, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		app.ModuleAccountAddrs(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]), app.AccountKeeper, app.BankKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(), authcodec.NewBech32Codec(ValidatorAddressPrefix), authcodec.NewBech32Codec(ConsNodeAddressPrefix),
	)

	app.MintKeeper = mintkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[minttypes.StoreKey]), app.StakingKeeper, app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	app.DistrKeeper = distrkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[distrtypes.StoreKey]), app.AccountKeeper, app.BankKeeper, app.StakingKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, legacyAmino, runtime.NewKVStoreService(keys[slashingtypes.StoreKey]), app.StakingKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.CrisisKeeper = crisiskeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[crisistypes.StoreKey]), app.invCheckPeriod,
		app.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String(), app.AccountKeeper.AddressCodec())

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[feegrant.StoreKey]), app.AccountKeeper)

	app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, runtime.NewKVStoreService(keys[upgradetypes.StoreKey]), appCodec, homePath, app.BaseApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	app.MsgFeesKeeper = msgfeeskeeper.NewKeeper(
		appCodec, keys[msgfeestypes.StoreKey], app.GetSubspace(msgfeestypes.ModuleName),
		authtypes.FeeCollectorName, pioconfig.GetProvenanceConfig().FeeDenom,
		app.Simulate, encodingConfig.TxConfig.TxDecoder(), interfaceRegistry,
	)

	// pioMsgFeesRouter := app.MsgServiceRouter().(*piohandlers.PioMsgServiceRouter) // TODO[1760]: msg-service-router
	// pioMsgFeesRouter.SetMsgFeesKeeper(app.MsgFeesKeeper)                          // TODO[1760]: msg-service-router

	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	restrictHooks := piohandlers.NewStakingRestrictionHooks(app.StakingKeeper, *piohandlers.DefaultRestrictionOptions)
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(restrictHooks, app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	app.RewardKeeper = rewardkeeper.NewKeeper(appCodec, keys[rewardtypes.StoreKey], app.StakingKeeper, &app.GovKeeper, app.BankKeeper, app.AccountKeeper)

	// TODO[1760]: authz: Put back this call to NewKeeper with fixed arguments.
	// app.AuthzKeeper = authzkeeper.NewKeeper(
	// 	keys[authzkeeper.StoreKey], appCodec, app.BaseApp.MsgServiceRouter(), app.AccountKeeper,
	// )

	app.GroupKeeper = groupkeeper.NewKeeper(keys[group.StoreKey], appCodec, app.BaseApp.MsgServiceRouter(), app.AccountKeeper, group.DefaultConfig())

	// Create IBC Keeper
	// TODO[1760]: ibc-host: Put back this call to NewKeeper with fixed arguments.
	// app.IBCKeeper = ibckeeper.NewKeeper(
	// 	appCodec, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName), app.StakingKeeper, app.UpgradeKeeper, scopedIBCKeeper,
	// )

	// Configure the hooks keeper
	hooksKeeper := ibchookskeeper.NewKeeper(
		keys[ibchookstypes.StoreKey],
		app.GetSubspace(ibchookstypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		nil,
	)
	app.IBCHooksKeeper = &hooksKeeper

	// Setup the ICS4Wrapper used by the hooks middleware
	addrPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()        // We use this approach so running tests which use "cosmos" will work while we use "pb"
	wasmHooks := ibchooks.NewWasmHooks(&hooksKeeper, nil, addrPrefix) // The contract keeper needs to be set later
	app.Ics20WasmHooks = &wasmHooks
	markerHooks := ibchooks.NewMarkerHooks(nil)
	app.Ics20MarkerHooks = &markerHooks
	ibcHooks := ibchooks.NewIbcHooks(appCodec, &hooksKeeper, app.IBCKeeper, app.Ics20WasmHooks, app.Ics20MarkerHooks, nil)
	app.IbcHooks = &ibcHooks

	app.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		app.IBCKeeper.ChannelKeeper,
		app.IbcHooks,
	)

	rateLimtingKeeper := ibcratelimitkeeper.NewKeeper(appCodec, keys[ibcratelimit.StoreKey], nil)
	app.RateLimitingKeeper = &rateLimtingKeeper

	// Create Transfer Keepers
	rateLimitingTransferModule := ibcratelimitmodule.NewIBCMiddleware(nil, app.HooksICS4Wrapper, app.RateLimitingKeeper)
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		&rateLimitingTransferModule,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		govAuthority,
	)
	app.TransferKeeper = &transferKeeper
	transferModule := ibctransfer.NewIBCModule(*app.TransferKeeper)
	rateLimitingTransferModule = *rateLimitingTransferModule.WithIBCModule(transferModule)
	hooksTransferModule := ibchooks.NewIBCMiddleware(&rateLimitingTransferModule, &app.HooksICS4Wrapper)
	app.TransferStack = &hooksTransferModule

	app.NameKeeper = namekeeper.NewKeeper(
		appCodec, keys[nametypes.StoreKey], app.GetSubspace(nametypes.ModuleName),
	)

	app.AttributeKeeper = attributekeeper.NewKeeper(
		appCodec, keys[attributetypes.StoreKey], app.GetSubspace(attributetypes.ModuleName), app.AccountKeeper, &app.NameKeeper,
	)

	app.MetadataKeeper = metadatakeeper.NewKeeper(
		appCodec, keys[metadatatypes.StoreKey], app.GetSubspace(metadatatypes.ModuleName), app.AccountKeeper, app.AuthzKeeper, app.AttributeKeeper,
	)

	markerReqAttrBypassAddrs := []sdk.AccAddress{
		authtypes.NewModuleAddress(authtypes.FeeCollectorName), // Allow collecting fees in restricted coins.
		authtypes.NewModuleAddress(rewardtypes.ModuleName),     // Allow rewards to hold onto restricted coins.
		// authtypes.NewModuleAddress(quarantine.ModuleName),          // Allow quarantine to hold onto restricted coins. // TODO[1760]: quarantine
		authtypes.NewModuleAddress(govtypes.ModuleName),            // Allow restricted coins in deposits.
		authtypes.NewModuleAddress(distrtypes.ModuleName),          // Allow fee denoms to be restricted coins.
		authtypes.NewModuleAddress(stakingtypes.BondedPoolName),    // Allow bond denom to be a restricted coin.
		authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName), // Allow bond denom to be a restricted coin.
	}
	app.MarkerKeeper = markerkeeper.NewKeeper(
		appCodec, keys[markertypes.StoreKey], app.GetSubspace(markertypes.ModuleName),
		app.AccountKeeper, app.BankKeeper, app.AuthzKeeper, app.FeeGrantKeeper,
		app.AttributeKeeper, app.NameKeeper, app.TransferKeeper, markerReqAttrBypassAddrs,
	)

	app.HoldKeeper = holdkeeper.NewKeeper(
		appCodec, keys[hold.StoreKey], app.BankKeeper,
	)

	app.ExchangeKeeper = exchangekeeper.NewKeeper(
		appCodec, keys[exchange.StoreKey], authtypes.FeeCollectorName,
		app.AccountKeeper, app.AttributeKeeper, app.BankKeeper, app.HoldKeeper, app.MarkerKeeper,
	)

	// TODO[1760]: msg-service-router: Put the pioMessageRouter back into use (and the ica host module back in).
	/*
		pioMessageRouter := MessageRouterFunc(func(msg sdk.Msg) baseapp.MsgServiceHandler {
			return pioMsgFeesRouter.Handler(msg)
		})
		app.TriggerKeeper = triggerkeeper.NewKeeper(appCodec, keys[triggertypes.StoreKey], app.MsgServiceRouter())
		icaHostKeeper := icahostkeeper.NewKeeper(
			appCodec, keys[icahosttypes.StoreKey], app.GetSubspace(icahosttypes.SubModuleName),
			app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
			app.AccountKeeper, scopedICAHostKeeper, pioMessageRouter,
		)
		app.ICAHostKeeper = &icaHostKeeper
		icaModule := ica.NewAppModule(nil, app.ICAHostKeeper)
		icaHostIBCModule := icahost.NewIBCModule(*app.ICAHostKeeper)
	*/

	// TODO[1760]: async-icq
	// app.ICQKeeper = icqkeeper.NewKeeper(
	// 	appCodec, keys[icqtypes.StoreKey], app.GetSubspace(icqtypes.ModuleName),
	// 	app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
	// 	scopedICQKeeper, app.BaseApp.GRPCQueryRouter(),
	//)
	// icqModule := icq.NewAppModule(app.ICQKeeper)
	// icqIBCModule := icq.NewIBCModule(app.ICQKeeper)

	// Init CosmWasm module
	wasmDir := filepath.Join(homePath, "data", "wasm")

	wasmWrap := WasmWrapper{Wasm: wasm.DefaultWasmConfig()}
	err := viper.Unmarshal(&wasmWrap)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}
	wasmConfig := wasmWrap.Wasm

	// Init CosmWasm encoder integrations
	encoderRegistry := provwasm.NewEncoderRegistry()
	encoderRegistry.RegisterEncoder(nametypes.RouterKey, namewasm.Encoder)
	encoderRegistry.RegisterEncoder(attributetypes.RouterKey, attributewasm.Encoder)
	encoderRegistry.RegisterEncoder(markertypes.RouterKey, markerwasm.Encoder)
	encoderRegistry.RegisterEncoder(metadatatypes.RouterKey, metadatawasm.Encoder)
	encoderRegistry.RegisterEncoder(msgfeestypes.RouterKey, msgfeeswasm.Encoder)

	// Init CosmWasm query integrations
	querierRegistry := provwasm.NewQuerierRegistry()
	querierRegistry.RegisterQuerier(nametypes.RouterKey, namewasm.Querier(app.NameKeeper))
	querierRegistry.RegisterQuerier(attributetypes.RouterKey, attributewasm.Querier(app.AttributeKeeper))
	querierRegistry.RegisterQuerier(markertypes.RouterKey, markerwasm.Querier(app.MarkerKeeper))
	querierRegistry.RegisterQuerier(metadatatypes.RouterKey, metadatawasm.Querier(app.MetadataKeeper))

	// Add the staking feature and indicate that provwasm contracts can be run on this chain.
	// Addition of cosmwasm_1_1 adds capability defined here: https://github.com/CosmWasm/cosmwasm/pull/1356
	supportedFeatures := "staking,provenance,stargate,iterator,cosmwasm_1_1"

	// The last arguments contain custom message handlers, and custom query handlers,
	// to allow smart contracts to use provenance modules.
	// TODO[1760]: wasm: Figure out the replacement for NewKeeper.
	_, _ = scopedWasmKeeper, supportedFeatures
	_, _ = wasmDir, wasmConfig
	/*
		wasmKeeperInstance := wasm.NewKeeper(
			appCodec,
			keys[wasm.StoreKey],
			app.GetSubspace(wasm.ModuleName),
			app.AccountKeeper,
			app.BankKeeper,
			app.StakingKeeper,
			app.DistrKeeper,
			app.IBCKeeper.ChannelKeeper,
			&app.IBCKeeper.PortKeeper,
			scopedWasmKeeper,
			app.TransferKeeper,
			// pioMessageRouter, // TODO[1760]: msg-service-router
			app.GRPCQueryRouter(),
			wasmDir,
			wasmConfig,
			supportedFeatures,
			authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			wasmkeeper.WithQueryPlugins(provwasm.QueryPlugins(querierRegistry, *app.GRPCQueryRouter(), appCodec)),
			wasmkeeper.WithMessageEncoders(provwasm.MessageEncoders(encoderRegistry, logger)),
		)
		app.WasmKeeper = &wasmKeeperInstance
	*/

	// Pass the wasm keeper to all the wrappers that need it
	app.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	app.Ics20WasmHooks.ContractKeeper = app.WasmKeeper // app.ContractKeeper -- this changes in the next version of wasm to a permissioned keeper
	app.IBCHooksKeeper.ContractKeeper = app.ContractKeeper
	app.Ics20MarkerHooks.MarkerKeeper = &app.MarkerKeeper
	app.RateLimitingKeeper.PermissionedKeeper = app.ContractKeeper

	app.IbcHooks.SendPacketPreProcessors = []ibchookstypes.PreSendPacketDataProcessingFn{app.Ics20MarkerHooks.SetupMarkerMemoFn, app.Ics20WasmHooks.GetWasmSendPacketPreProcessor}

	app.ScopedOracleKeeper = scopedOracleKeeper
	app.OracleKeeper = *oraclekeeper.NewKeeper(
		appCodec,
		keys[oracletypes.StoreKey],
		keys[oracletypes.MemStoreKey],
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		scopedOracleKeeper,
		wasmkeeper.Querier(app.WasmKeeper),
	)
	oracleModule := oraclemodule.NewAppModule(appCodec, app.OracleKeeper, app.AccountKeeper, app.BankKeeper, app.IBCKeeper.ChannelKeeper)

	// TODO[1760]: sanction
	// unsanctionableAddrs := make([]sdk.AccAddress, 0, len(maccPerms)+1)
	// for mName := range maccPerms {
	// 	unsanctionableAddrs = append(unsanctionableAddrs, authtypes.NewModuleAddress(mName))
	// }
	// unsanctionableAddrs = append(unsanctionableAddrs, authtypes.NewModuleAddress(quarantine.ModuleName))
	// app.SanctionKeeper = sanctionkeeper.NewKeeper(appCodec, keys[sanction.StoreKey],
	// 	app.BankKeeper, &app.GovKeeper,
	// 	authtypes.NewModuleAddress(govtypes.ModuleName).String(), unsanctionableAddrs)

	// register the proposal types
	govRouter := govtypesv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		// AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(app.WasmKeeper, wasm.EnableAllProposals)).  // TODO[1760]: wasm
		AddRoute(nametypes.ModuleName, name.NewProposalHandler(app.NameKeeper)).
		AddRoute(markertypes.ModuleName, marker.NewProposalHandler(app.MarkerKeeper)).
		AddRoute(msgfeestypes.ModuleName, msgfees.NewProposalHandler(app.MsgFeesKeeper, app.InterfaceRegistry()))
	// TODO[1760]: gov: Put back this call to NewKeeper with fixed arguments.
	// app.GovKeeper = govkeeper.NewKeeper(
	// 	appCodec, keys[govtypes.StoreKey], app.GetSubspace(govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
	// 	&stakingKeeper, govRouter, app.BaseApp.MsgServiceRouter(), govtypes.Config{MaxMetadataLen: 10000},
	// )
	// app.GovKeeper.SetHooks(govtypes.NewMultiGovHooks(app.SanctionKeeper)) // TODO[1760]: sanction

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.
		AddRoute(ibctransfertypes.ModuleName, app.TransferStack).
		AddRoute(wasm.ModuleName, wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper)).
		// AddRoute(icahosttypes.SubModuleName, icaHostIBCModule). // TODO[1760]: ica-host
		// AddRoute(icqtypes.ModuleName, icqIBCModule). // TODO[1760]: async-icq
		AddRoute(oracletypes.ModuleName, oracleModule)
	app.IBCKeeper.SetRouter(ibcRouter)

	// Create evidence Keeper for to register the IBC light client misbehavior evidence route
	// TODO[1760]: evidence: Put back this call to NewKeeper with fixed arguments.
	// evidenceKeeper := evidencekeeper.NewKeeper(
	// 	appCodec, keys[evidencetypes.StoreKey], &app.StakingKeeper, app.SlashingKeeper,
	// )
	// If evidence needs to be handled for the app, set routes in router here and seal
	// app.EvidenceKeeper = *evidenceKeeper // TODO[1760]: evidence

	// app.QuarantineKeeper = quarantinekeeper.NewKeeper(appCodec, keys[quarantine.StoreKey], app.BankKeeper, authtypes.NewModuleAddress(quarantine.ModuleName)) // TODO[1760]: quarantine

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))
	_ = skipGenesisInvariants // TODO[1760]: crisis

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app.BaseApp, encodingConfig.TxConfig),
		// auth.NewAppModule(appCodec, app.AccountKeeper, nil), // TODO[1760]: auth
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		// bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper), // TODO[1760]: bank
		// capability.NewAppModule(appCodec, *app.CapabilityKeeper), // TODO[1760]: capability
		// crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants), // TODO[1760]: crisis
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		// gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper), // TODO[1760]: gov
		// mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil), // TODO[1760]: mint
		// slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper), // TODO[1760]: slashing
		// distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper), // TODO[1760]: distr
		// staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper), // TODO[1760]: staking
		// upgrade.NewAppModule(app.UpgradeKeeper), // TODO[1760]: upgrade
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		// quarantinemodule.NewAppModule(appCodec, app.QuarantineKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry), // TODO[1760]: quarantine
		// sanctionmodule.NewAppModule(appCodec, app.SanctionKeeper, app.AccountKeeper, app.BankKeeper, app.GovKeeper, app.interfaceRegistry), // TODO[1760]: sanction

		// PROVENANCE
		metadata.NewAppModule(appCodec, app.MetadataKeeper, app.AccountKeeper),
		marker.NewAppModule(appCodec, app.MarkerKeeper, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.GovKeeper, app.AttributeKeeper, app.interfaceRegistry),
		name.NewAppModule(appCodec, app.NameKeeper, app.AccountKeeper, app.BankKeeper),
		attribute.NewAppModule(appCodec, app.AttributeKeeper, app.AccountKeeper, app.BankKeeper, app.NameKeeper),
		msgfeesmodule.NewAppModule(appCodec, app.MsgFeesKeeper, app.interfaceRegistry),
		// wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper), // TODO[1760]: wasm
		rewardmodule.NewAppModule(appCodec, app.RewardKeeper, app.AccountKeeper, app.BankKeeper),
		triggermodule.NewAppModule(appCodec, app.TriggerKeeper, app.AccountKeeper, app.BankKeeper),
		oracleModule,
		holdmodule.NewAppModule(appCodec, app.HoldKeeper),
		exchangemodule.NewAppModule(appCodec, app.ExchangeKeeper),

		// IBC
		ibc.NewAppModule(app.IBCKeeper),
		ibcratelimitmodule.NewAppModule(appCodec, *app.RateLimitingKeeper, app.AccountKeeper, app.BankKeeper),
		ibchooks.NewAppModule(app.AccountKeeper, *app.IBCHooksKeeper),
		ibctransfer.NewAppModule(*app.TransferKeeper),
		// icqModule, // TODO[1760]: async-icq
		// icaModule, // TODO[1760]: ica-host
	)

	// TODO[1760]: app-module: BasicModuleManager: Make sure that this setup has everything we need (it was just copied from the SDK).
	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.mm,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
				},
			),
		})
	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

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
		// ibchost.ModuleName, // TODO[1760]: ibc-host
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
		// icqtypes.ModuleName, // TODO[1760]: async-icq
		nametypes.ModuleName,
		vestingtypes.ModuleName,
		// quarantine.ModuleName, // TODO[1760]: quarantine
		// sanction.ModuleName, // TODO[1760]: sanction
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
		// ibchost.ModuleName, // TODO[1760]: ibc-host
		ibcratelimit.ModuleName,
		ibchookstypes.ModuleName,
		ibctransfertypes.ModuleName,
		// icqtypes.ModuleName, // TODO[1760]: async-icq
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
		// quarantine.ModuleName, // TODO[1760]: quarantine
		// sanction.ModuleName, // TODO[1760]: sanction
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
		// quarantine.ModuleName, // TODO[1760]: quarantine
		// sanction.ModuleName, // TODO[1760]: sanction

		nametypes.ModuleName,
		attributetypes.ModuleName,
		metadatatypes.ModuleName,
		msgfeestypes.ModuleName,
		hold.ModuleName,
		exchange.ModuleName, // must be after the hold module.

		// ibchost.ModuleName, // TODO[1760]: ibc-host
		ibctransfertypes.ModuleName,
		// icqtypes.ModuleName, // TODO[1760]: async-icq
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
		// ibchost.ModuleName, // TODO[1760]: ibc-host
		minttypes.ModuleName,
		paramstypes.ModuleName,
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		// quarantine.ModuleName, // TODO[1760]: quarantine
		// sanction.ModuleName, // TODO[1760]: sanction
		hold.ModuleName,
		exchange.ModuleName,

		ibcratelimit.ModuleName,
		ibchookstypes.ModuleName,
		icatypes.ModuleName,
		// icqtypes.ModuleName, // TODO[1760]: async-icq
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

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.BaseApp.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	app.sm = module.NewSimulationManager(
		// auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts), // TODO[1760]: auth
		// bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper), // TODO[1760]: bank
		// capability.NewAppModule(appCodec, *app.CapabilityKeeper), // TODO[1760]: capability
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		// gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper), // TODO[1760]: gov
		// mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil), // TODO[1760]: mint
		// staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper), // TODO[1760]: staking
		// distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper), // TODO[1760]: distr
		// slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper), // TODO[1760]: slashing
		params.NewAppModule(app.ParamsKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		// quarantinemodule.NewAppModule(appCodec, app.QuarantineKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry), // TODO[1760]: quarantine
		// sanctionmodule.NewAppModule(appCodec, app.SanctionKeeper, app.AccountKeeper, app.BankKeeper, app.GovKeeper, app.interfaceRegistry), // TODO[1760]: sanction

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
		// icaModule, // TODO[1760]: ica-host
	)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	anteHandler, err := antewrapper.NewAnteHandler(
		antewrapper.HandlerOptions{
			AccountKeeper:       app.AccountKeeper,
			BankKeeper:          app.BankKeeper,
			TxSigningHandlerMap: encodingConfig.TxConfig.SignModeHandler(),
			FeegrantKeeper:      app.FeeGrantKeeper,
			MsgFeesKeeper:       app.MsgFeesKeeper,
			SigGasConsumer:      ante.DefaultSigVerificationGasConsumer,
		})
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	// TODO[1760]: fee-handler: Add the msgfeehandler back to the app.
	/*
		msgFeeHandler, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
			AccountKeeper:  app.AccountKeeper,
			BankKeeper:     app.BankKeeper,
			FeegrantKeeper: app.FeeGrantKeeper,
			MsgFeesKeeper:  app.MsgFeesKeeper,
			Decoder:        encodingConfig.TxConfig.TxDecoder(),
		})

		if err != nil {
			panic(err)
		}
		app.SetFeeHandler(msgFeeHandler)
	*/

	app.SetEndBlocker(app.EndBlocker)

	// app.SetAggregateEventsFunc(piohandlers.AggregateEvents) // TODO[1760]: event-history

	// Add upgrade plans for each release. This must be done before the baseapp seals via LoadLatestVersion() down below.
	InstallCustomUpgradeHandlers(app)

	// Use the dump of $home/data/upgrade-info.json:{"name":"$plan","height":321654} to determine
	// if we load a store upgrade from the handlers. No file == no error from read func.
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	// Currently in an upgrade hold for this block.
	if upgradeInfo.Name != "" && upgradeInfo.Height == app.LastBlockHeight()+1 {
		if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			app.Logger().Info("Skipping upgrade based on height",
				"plan", upgradeInfo.Name,
				"upgradeHeight", upgradeInfo.Height,
				"lastHeight", app.LastBlockHeight(),
			)
		} else {
			app.Logger().Info("Managing upgrade",
				"plan", upgradeInfo.Name,
				"upgradeHeight", upgradeInfo.Height,
				"lastHeight", app.LastBlockHeight(),
			)
			// See if we have a custom store loader to use for upgrades.
			storeLoader := GetUpgradeStoreLoader(app, upgradeInfo)
			if storeLoader != nil {
				app.SetStoreLoader(storeLoader)
			}
		}
	}
	// --

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			cmtos.Exit(err.Error())
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	// app.ScopedICQKeeper = scopedICQKeeper // TODO[1760]: async-icq
	app.ScopedICAHostKeeper = scopedICAHostKeeper

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
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
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

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *App) DefaultGenesis() map[string]json.RawMessage {
	return a.BasicModuleManager.DefaultGenesis(a.appCodec)
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig serverconfig.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

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
	cmtservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// RegisterNodeService registers the node query server.
func (app *App) RegisterNodeService(clientCtx client.Context, cfg serverconfig.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// RegisterStreamingServices registers types.ABCIListener State Listening services with the App.
func (app *App) RegisterStreamingServices(appOpts servertypes.AppOptions) {
	// TODO[1760]: streaming: Ensure that this change is correct.
	if err := app.BaseApp.RegisterStreamingServices(appOpts, app.keys); err != nil {
		app.Logger().Error("failed to register streaming plugin", "error", err)
		os.Exit(1)
	}
	/*
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
				if err := app.BaseApp.RegisterStreamingServices(appOpts, app.keys, plugin); err != nil {
					app.Logger().Error("failed to register streaming plugin", "error", err)
					os.Exit(1)
				}
				app.Logger().Info("streaming service registered", "service", service, "plugin", pluginName)
			}
		}
	*/
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

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypesv1.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)

	paramsKeeper.Subspace(metadatatypes.ModuleName)  // TODO[1760]: params: Migrate metadata params.
	paramsKeeper.Subspace(markertypes.ModuleName)    // TODO[1760]: params: Migrate marker params.
	paramsKeeper.Subspace(nametypes.ModuleName)      // TODO[1760]: params: Migrate name params.
	paramsKeeper.Subspace(attributetypes.ModuleName) // TODO[1760]: params: Migrate attribute params.
	paramsKeeper.Subspace(msgfeestypes.ModuleName)   // TODO[1760]: params: Migrate msgFees params.
	paramsKeeper.Subspace(wasm.ModuleName)
	paramsKeeper.Subspace(rewardtypes.ModuleName)  // TODO[1760]: params: Migrate reward params.
	paramsKeeper.Subspace(triggertypes.ModuleName) // TODO[1760]: params: Migrate trigger params.

	paramsKeeper.Subspace(ibctransfertypes.ModuleName) // TODO[1760]: params: Migrate ibc-transfer params.
	// paramsKeeper.Subspace(ibchost.ModuleName)          // TODO[1760]: params: Migrate ibc-host params.
	paramsKeeper.Subspace(icahosttypes.SubModuleName) // TODO[1760]: params: Migrate ica-host params.
	// paramsKeeper.Subspace(icqtypes.ModuleName) // TODO[1760]: params: Migrate icq params.
	paramsKeeper.Subspace(ibchookstypes.ModuleName) // TODO[1760]: params: Migrate ibc-hooks params.

	return paramsKeeper
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
	app.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
		if !injected {
			app.Logger().Info("Injecting upgrade plan", "plan", plan)
			// Ideally, we'd just call ScheduleUpgrade(ctx, plan) here (and panic on error).
			// But the upgrade keeper often its own migration stuff that change some store key stuff.
			// ScheduleUpgrade tries to read some of that changed store stuff and fails if the migration hasn't
			// been applied yet. So we're doing things the hard way here.
			app.UpgradeKeeper.ClearUpgradePlan(ctx)
			store := ctx.KVStore(app.GetKey(upgradetypes.StoreKey))
			bz := app.appCodec.MustMarshal(&plan)
			store.Set(upgradetypes.PlanKey(), bz)
			injected = true
		}
		return app.BeginBlocker(ctx)
	})
}
