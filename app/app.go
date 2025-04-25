package app

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/gorilla/mux"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtos "github.com/cometbft/cometbft/libs/os"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/circuit"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	circuittypes "cosmossdk.io/x/circuit/types"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	nftmodule "cosmossdk.io/x/nft/module"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	sigtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramprops "github.com/cosmos/cosmos-sdk/x/params/types/proposal" //nolint:depguard // Need this here to register old types.
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v8"
	icqkeeper "github.com/cosmos/ibc-apps/modules/async-icq/v8/keeper"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v8/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/client/docs"
	"github.com/provenance-io/provenance/internal/antewrapper"
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/internal/provwasm"
	assetkeeper "github.com/provenance-io/provenance/x/asset/keeper"
	assetmodule "github.com/provenance-io/provenance/x/asset/module"
	assettypes "github.com/provenance-io/provenance/x/asset/types"
	"github.com/provenance-io/provenance/x/attribute"
	attributekeeper "github.com/provenance-io/provenance/x/attribute/keeper"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
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
	"github.com/provenance-io/provenance/x/ledger"
	ledgerkeeper "github.com/provenance-io/provenance/x/ledger/keeper"
	ledgermodule "github.com/provenance-io/provenance/x/ledger/module"
	"github.com/provenance-io/provenance/x/marker"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata"
	metadatakeeper "github.com/provenance-io/provenance/x/metadata/keeper"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeeskeeper "github.com/provenance-io/provenance/x/msgfees/keeper"
	msgfeesmodule "github.com/provenance-io/provenance/x/msgfees/module"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	"github.com/provenance-io/provenance/x/name"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	oraclekeeper "github.com/provenance-io/provenance/x/oracle/keeper"
	oraclemodule "github.com/provenance-io/provenance/x/oracle/module"
	oracletypes "github.com/provenance-io/provenance/x/oracle/types"
	"github.com/provenance-io/provenance/x/registry"
	registrykeeper "github.com/provenance-io/provenance/x/registry/keeper"
	registrymodule "github.com/provenance-io/provenance/x/registry/module"
	triggerkeeper "github.com/provenance-io/provenance/x/trigger/keeper"
	triggermodule "github.com/provenance-io/provenance/x/trigger/module"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// DefaultPowerReduction pio specific value for power reduction for TokensFromConsensusPower
	DefaultPowerReduction = sdkmath.NewIntFromUint64(1_000_000_000)

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

		assettypes.ModuleName:     nil,
		attributetypes.ModuleName: nil,
		markertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		// TODO[wasm]: Delete or uncomment: wasmtypes.ModuleName:      {authtypes.Burner},
		triggertypes.ModuleName:  nil,
		oracletypes.ModuleName:   nil,
		metadatatypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		nft.ModuleName:           nil,
	}
)

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

// These are some values defined in the params module that we still need so that
// the params module can be deleted. But I don't want the imports, so they're copied here.
// TODO[viridian]: Delete these params constants after the upgrade.
const (
	paramsName = "params" // = paramstypes.ModuleName
)

// WasmWrapper allows us to use namespacing in the config file
// This is only used for parsing in the app, x/wasm expects WasmConfig
type WasmWrapper struct {
	Wasm wasmtypes.WasmConfig `mapstructure:"wasm"`
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
	txConfig          client.TxConfig

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	CircuitKeeper         circuitkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	NFTKeeper             nftkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	MsgFeesKeeper         msgfeeskeeper.Keeper
	TriggerKeeper         triggerkeeper.Keeper
	OracleKeeper          oraclekeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	IBCKeeper          *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCHooksKeeper     *ibchookskeeper.Keeper
	ICAHostKeeper      *icahostkeeper.Keeper
	TransferKeeper     *ibctransferkeeper.Keeper
	ICQKeeper          icqkeeper.Keeper
	RateLimitingKeeper *ibcratelimitkeeper.Keeper

	MarkerKeeper    markerkeeper.Keeper
	MetadataKeeper  metadatakeeper.Keeper
	AssetKeeper     assetkeeper.Keeper
	AttributeKeeper attributekeeper.Keeper
	NameKeeper      namekeeper.Keeper
	HoldKeeper      holdkeeper.Keeper
	RegistryKeeper  registrykeeper.RegistryKeeper
	LedgerKeeper    ledgerkeeper.BaseKeeper
	ExchangeKeeper  exchangekeeper.Keeper
	WasmKeeper      *wasmkeeper.Keeper
	ContractKeeper  *wasmkeeper.PermissionedKeeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper
	ScopedICQKeeper      capabilitykeeper.ScopedKeeper
	ScopedOracleKeeper   capabilitykeeper.ScopedKeeper

	TransferStack       *ibchooks.IBCMiddleware
	Ics20WasmHooks      *ibchooks.WasmHooks
	Ics20MarkerHooks    *ibchooks.MarkerHooks
	IbcHooks            *ibchooks.IbcHooks
	HooksICS4Wrapper    ibchooks.ICS4Middleware
	RateLimitMiddleware porttypes.Middleware

	// the module manager
	mm                 *module.Manager
	BasicModuleManager module.BasicManager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

func init() {
	clienthelpers.EnvPrefix = EnvPrefix
	var err error
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory("Provenance")
	if err != nil {
		panic(err)
	}

	// 614,400 = 600 * 1024 = our wasm params maxWasmCodeSize value before it was removed in wasmd v0.27.
	wasmtypes.MaxWasmSize = 614_400
}

// New returns a reference to an initialized Provenance Blockchain App.
func New(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	sdkConfig := sdk.GetConfig()
	addrPrefix := sdkConfig.GetBech32AccountAddrPrefix()
	valAddrPrefix := sdkConfig.GetBech32ValidatorAddrPrefix()
	consAddrPrefix := sdkConfig.GetBech32ConsensusAddrPrefix()

	signingOptions := signing.Options{
		AddressCodec:          address.Bech32Codec{Bech32Prefix: addrPrefix},
		ValidatorAddressCodec: address.Bech32Codec{Bech32Prefix: valAddrPrefix},
	}
	exchange.DefineCustomGetSigners(&signingOptions)
	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfigOpts := authtx.ConfigOptions{
		EnabledSignModes: authtx.DefaultSignModes,
		SigningOptions:   &signingOptions,
	}
	txConfig, err := authtx.NewTxConfigWithOptions(appCodec, txConfigOpts)
	if err != nil {
		panic(err)
	}

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	bApp := baseapp.NewBaseApp("provenanced", logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetMsgServiceRouter(piohandlers.NewPioMsgServiceRouter(txConfig.TxDecoder()))
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	sdk.SetCoinDenomRegex(SdkCoinDenomRegex)

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, consensusparamtypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, capabilitytypes.StoreKey, circuittypes.StoreKey,
		authzkeeper.StoreKey, group.StoreKey, crisistypes.StoreKey,

		ibcexported.StoreKey,
		ibctransfertypes.StoreKey,
		icahosttypes.StoreKey,
		icqtypes.StoreKey,
		ibchookstypes.StoreKey,
		ibcratelimit.StoreKey,

		metadatatypes.StoreKey,
		markertypes.StoreKey,
		assettypes.StoreKey,
		attributetypes.StoreKey,
		nametypes.StoreKey,
		msgfeestypes.StoreKey,
		wasmtypes.StoreKey,
		triggertypes.StoreKey,
		oracletypes.StoreKey,
		hold.StoreKey,
		ledger.StoreKey,
		exchange.StoreKey,
		registry.StoreKey,
		nft.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys()
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	govAuthority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	// Register State listening services.
	if err := app.BaseApp.RegisterStreamingServices(appOpts, app.keys); err != nil {
		app.Logger().Error("failed to register streaming plugin", "error", err)
		os.Exit(1)
	}

	// set the BaseApp's parameter store

	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		govAuthority,
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	app.ScopedICAHostKeeper = app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.ScopedICQKeeper = app.CapabilityKeeper.ScopeToModule(icqtypes.ModuleName)
	scopedOracleKeeper := app.CapabilityKeeper.ScopeToModule(oracletypes.ModuleName)

	// capability keeper must be sealed after scope to module registrations are completed.
	app.CapabilityKeeper.Seal()

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(appCodec, runtime.NewKVStoreService(keys[authtypes.StoreKey]), authtypes.ProtoBaseAccount, maccPerms, signingOptions.AddressCodec, addrPrefix, govAuthority)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		app.ModuleAccountAddrs(),
		govAuthority,
		logger,
	)

	// optional: enable sign mode textual by overwriting the default tx config (after setting the bank keeper)
	txConfigOpts.EnabledSignModes = append(txConfigOpts.EnabledSignModes, sigtypes.SignMode_SIGN_MODE_TEXTUAL)
	txConfigOpts.TextualCoinMetadataQueryFn = txmodule.NewBankKeeperCoinMetadataQueryFn(app.BankKeeper)
	txConfig, err = authtx.NewTxConfigWithOptions(appCodec, txConfigOpts)
	if err != nil {
		panic(err)
	}
	if err = txConfig.SigningContext().Validate(); err != nil {
		panic(err)
	}
	app.txConfig = txConfig

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]), app.AccountKeeper, app.BankKeeper, govAuthority, authcodec.NewBech32Codec(valAddrPrefix), authcodec.NewBech32Codec(consAddrPrefix),
	)

	app.CircuitKeeper = circuitkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[circuittypes.StoreKey]), govAuthority, app.AccountKeeper.AddressCodec())
	app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)

	app.MintKeeper = mintkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[minttypes.StoreKey]), app.StakingKeeper, app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName, govAuthority)

	app.DistrKeeper = distrkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[distrtypes.StoreKey]), app.AccountKeeper, app.BankKeeper, app.StakingKeeper, authtypes.FeeCollectorName, govAuthority)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, legacyAmino, runtime.NewKVStoreService(keys[slashingtypes.StoreKey]), app.StakingKeeper, govAuthority,
	)

	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	app.CrisisKeeper = crisiskeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[crisistypes.StoreKey]), invCheckPeriod,
		app.BankKeeper, authtypes.FeeCollectorName, govAuthority, app.AccountKeeper.AddressCodec())

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[feegrant.StoreKey]), app.AccountKeeper).SetBankKeeper(app.BankKeeper)

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	if len(homePath) == 0 {
		homePath = DefaultNodeHome
	}
	app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, runtime.NewKVStoreService(keys[upgradetypes.StoreKey]), appCodec, homePath, app.BaseApp, govAuthority)

	app.MsgFeesKeeper = msgfeeskeeper.NewKeeper(
		appCodec, keys[msgfeestypes.StoreKey], authtypes.FeeCollectorName,
		pioconfig.GetProvenanceConfig().FeeDenom, app.SimulateProv,
		app.txConfig.TxDecoder(), interfaceRegistry,
	)

	pioMsgFeesRouter := app.MsgServiceRouter().(*piohandlers.PioMsgServiceRouter)
	pioMsgFeesRouter.SetMsgFeesKeeper(app.MsgFeesKeeper)

	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	restrictHooks := piohandlers.NewStakingRestrictionHooks(app.StakingKeeper, *piohandlers.DefaultRestrictionOptions)
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(restrictHooks, app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(runtime.NewKVStoreService(keys[authzkeeper.StoreKey]), appCodec, app.BaseApp.MsgServiceRouter(), app.AccountKeeper).SetBankKeeper(app.BankKeeper)

	app.GroupKeeper = groupkeeper.NewKeeper(keys[group.StoreKey], appCodec, app.BaseApp.MsgServiceRouter(), app.AccountKeeper, group.DefaultConfig())

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, keys[ibcexported.StoreKey], nil, app.StakingKeeper, app.UpgradeKeeper, app.ScopedIBCKeeper, govAuthority,
	)

	// Configure the hooks keeper
	hooksKeeper := ibchookskeeper.NewKeeper(
		appCodec,
		keys[ibchookstypes.StoreKey],
		app.IBCKeeper.ChannelKeeper,
		nil,
	)
	app.IBCHooksKeeper = &hooksKeeper

	// Setup the ICS4Wrapper used by the hooks middleware
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

	rateLimitingKeeper := ibcratelimitkeeper.NewKeeper(appCodec, keys[ibcratelimit.StoreKey], nil)
	app.RateLimitingKeeper = &rateLimitingKeeper

	// Create Transfer Keepers
	rateLimitingTransferModule := ibcratelimitmodule.NewIBCMiddleware(nil, app.HooksICS4Wrapper, app.RateLimitingKeeper)
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		nil,
		&rateLimitingTransferModule,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.ScopedTransferKeeper,
		govAuthority,
	)
	app.TransferKeeper = &transferKeeper
	transferModule := ibctransfer.NewIBCModule(*app.TransferKeeper)
	app.RateLimitMiddleware = rateLimitingTransferModule.WithIBCModule(transferModule)
	hooksTransferModule := ibchooks.NewIBCMiddleware(app.RateLimitMiddleware, &app.HooksICS4Wrapper)
	app.TransferStack = &hooksTransferModule

	app.NameKeeper = namekeeper.NewKeeper(appCodec, keys[nametypes.StoreKey])

	app.AttributeKeeper = attributekeeper.NewKeeper(
		appCodec, keys[attributetypes.StoreKey], app.AccountKeeper, &app.NameKeeper,
	)

	markerReqAttrBypassAddrs := []sdk.AccAddress{
		authtypes.NewModuleAddress(authtypes.FeeCollectorName),     // Allow collecting fees in restricted coins.
		authtypes.NewModuleAddress(govtypes.ModuleName),            // Allow restricted coins in deposits.
		authtypes.NewModuleAddress(distrtypes.ModuleName),          // Allow fee denoms to be restricted coins.
		authtypes.NewModuleAddress(stakingtypes.BondedPoolName),    // Allow bond denom to be a restricted coin.
		authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName), // Allow bond denom to be a restricted coin.
	}

	app.MarkerKeeper = markerkeeper.NewKeeper(
		appCodec, keys[markertypes.StoreKey], app.AccountKeeper,
		app.BankKeeper, app.AuthzKeeper, app.FeeGrantKeeper,
		app.AttributeKeeper, app.NameKeeper, app.TransferKeeper,
		markerReqAttrBypassAddrs, NewGroupCheckerFunc(app.GroupKeeper),
	)

	app.MetadataKeeper = metadatakeeper.NewKeeper(
		appCodec, keys[metadatatypes.StoreKey], app.AccountKeeper, app.AuthzKeeper, app.AttributeKeeper, app.MarkerKeeper, app.BankKeeper,
	)

	app.NFTKeeper = nftkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[nft.StoreKey]),
		appCodec,
		app.AccountKeeper,
		app.BankKeeper,
	)

	app.AssetKeeper = assetkeeper.NewKeeper(
		appCodec, keys[assettypes.StoreKey], app.NFTKeeper, app.BaseApp.MsgServiceRouter(),
	)

	app.HoldKeeper = holdkeeper.NewKeeper(
		appCodec, keys[hold.StoreKey], app.BankKeeper,
	)

	app.LedgerKeeper = ledgerkeeper.NewKeeper(appCodec, keys[ledger.StoreKey], runtime.NewKVStoreService(keys[ledger.StoreKey]), app.BankKeeper, app.NFTKeeper)

	app.RegistryKeeper = registrykeeper.NewKeeper(appCodec, keys[registry.StoreKey], runtime.NewKVStoreService(keys[registry.StoreKey]))

	app.ExchangeKeeper = exchangekeeper.NewKeeper(
		appCodec, keys[exchange.StoreKey], authtypes.FeeCollectorName,
		app.AccountKeeper, app.AttributeKeeper, app.BankKeeper, app.HoldKeeper, app.MarkerKeeper,
		app.MetadataKeeper,
	)

	pioMessageRouter := MessageRouterFunc(func(msg sdk.Msg) baseapp.MsgServiceHandler {
		return pioMsgFeesRouter.Handler(msg)
	})
	app.TriggerKeeper = triggerkeeper.NewKeeper(appCodec, keys[triggertypes.StoreKey], app.MsgServiceRouter())
	icaHostKeeper := icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey], nil,
		app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.PortKeeper,
		app.AccountKeeper, app.ScopedICAHostKeeper, pioMessageRouter, govAuthority,
	)
	app.ICAHostKeeper = &icaHostKeeper
	app.ICAHostKeeper.WithQueryRouter(app.GRPCQueryRouter())
	icaModule := ica.NewAppModule(nil, app.ICAHostKeeper)
	icaHostIBCModule := icahost.NewIBCModule(*app.ICAHostKeeper)

	app.ICQKeeper = icqkeeper.NewKeeper(
		appCodec, keys[icqtypes.StoreKey],
		app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.PortKeeper,
		app.ScopedICQKeeper, app.BaseApp.GRPCQueryRouter(), govAuthority,
	)
	icqModule := icq.NewAppModule(app.ICQKeeper, nil)
	icqIBCModule := icq.NewIBCModule(app.ICQKeeper)

	// Init CosmWasm module
	wasmDir := filepath.Join(homePath, "data", "wasm")

	wasmWrap := WasmWrapper{Wasm: wasmtypes.DefaultWasmConfig()}
	err = viper.Unmarshal(&wasmWrap)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}
	wasmConfig := wasmWrap.Wasm

	// Add the capabilities and indicate that provwasm contracts can be run on this chain.
	// Capabilities defined here: https://github.com/CosmWasm/cosmwasm/blob/main/docs/CAPABILITIES-BUILT-IN.md
	supportedFeatures := []string{"staking", "provenance", "stargate", "iterator", "cosmwasm_1_1", "cosmwasm_1_2", "cosmwasm_1_3", "cosmwasm_1_4", "cosmwasm_2_0", "cosmwasm_2_1"}

	// The last arguments contain custom message handlers, and custom query handlers,
	// to allow smart contracts to use provenance modules.
	wasmKeeperInstance := wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		app.TransferKeeper,
		pioMessageRouter,
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		govAuthority,
		wasmkeeper.WithQueryPlugins(provwasm.QueryPlugins(*app.GRPCQueryRouter(), appCodec)),
	)
	app.WasmKeeper = &wasmKeeperInstance

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

	// register the proposal types
	govRouter := govtypesv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler)
	govKeeper := govkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(keys[govtypes.StoreKey]), app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, app.DistrKeeper, app.BaseApp.MsgServiceRouter(), govtypes.Config{MaxMetadataLen: 10000}, govAuthority,
	)

	// Set legacy router for backwards compatibility with gov v1beta1
	govKeeper.SetLegacyRouter(govRouter)
	app.GovKeeper = *govKeeper

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.
		AddRoute(ibctransfertypes.ModuleName, app.TransferStack).
		AddRoute(wasmtypes.ModuleName, wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper)).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(icqtypes.ModuleName, icqIBCModule).
		AddRoute(oracletypes.ModuleName, oracleModule)
	app.IBCKeeper.SetRouter(ibcRouter)

	// Create evidence Keeper for to register the IBC light client misbehavior evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		app.StakingKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app.BaseApp, app.txConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, nil),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, nil),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, nil),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, nil),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil, app.interfaceRegistry),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, nil),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		circuit.NewAppModule(appCodec, app.CircuitKeeper),
		nftmodule.NewAppModule(appCodec, app.NFTKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),

		// PROVENANCE
		metadata.NewAppModule(appCodec, app.MetadataKeeper, app.AccountKeeper),
		assetmodule.NewAppModule(appCodec, app.AssetKeeper, app.AccountKeeper, app.NFTKeeper, app.BaseApp.MsgServiceRouter()),
		marker.NewAppModule(appCodec, app.MarkerKeeper, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.GovKeeper, app.AttributeKeeper, app.interfaceRegistry),
		name.NewAppModule(appCodec, app.NameKeeper, app.AccountKeeper, app.BankKeeper),
		attribute.NewAppModule(appCodec, app.AttributeKeeper, app.AccountKeeper, app.BankKeeper, app.NameKeeper),
		msgfeesmodule.NewAppModule(appCodec, app.MsgFeesKeeper, app.interfaceRegistry),
		// TODO[wasm]: Delete or uncomment: wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), nil),
		triggermodule.NewAppModule(appCodec, app.TriggerKeeper, app.AccountKeeper, app.BankKeeper),
		oracleModule,
		holdmodule.NewAppModule(appCodec, app.HoldKeeper),
		ledgermodule.NewAppModule(appCodec, app.LedgerKeeper),
		registrymodule.NewAppModule(appCodec, app.RegistryKeeper),
		exchangemodule.NewAppModule(appCodec, app.ExchangeKeeper),

		// IBC
		ibc.NewAppModule(app.IBCKeeper),
		ibcratelimitmodule.NewAppModule(appCodec, *app.RateLimitingKeeper),
		ibchooks.NewAppModule(app.AccountKeeper, *app.IBCHooksKeeper),
		ibctransfer.NewAppModule(*app.TransferKeeper),
		icqModule,
		icaModule,
		ibctm.AppModule{},
	)

	// Set the upgrade-keeper's version map that it uses during init-genesis.
	// When using depinject, this is done by means of the PopulateVersionMap function.
	app.UpgradeKeeper.SetInitVersionMap(app.mm.GetVersionMap())

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.mm,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		})
	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	// We removed the params module, but have several gov props with a ParameterChangeProposal in them.
	paramprops.RegisterLegacyAminoCodec(legacyAmino)
	paramprops.RegisterInterfaces(interfaceRegistry)

	// NOTE: upgrade module is required to be prioritized
	app.mm.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibcexported.ModuleName,
		markertypes.ModuleName,
		attributetypes.ModuleName,
		authz.ModuleName,
		triggertypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		triggertypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	moduleGenesisOrder := []string{
		capabilitytypes.ModuleName, // Must be first.
		authtypes.ModuleName,
		banktypes.ModuleName,
		markertypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName, // Must be after both staking and auth.
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		circuittypes.ModuleName,
		nft.ModuleName,
		nametypes.ModuleName,
		attributetypes.ModuleName,
		metadatatypes.ModuleName,
		msgfeestypes.ModuleName,
		hold.ModuleName,
		ledger.ModuleName,
		registry.ModuleName,
		exchange.ModuleName, // must be after the hold module.
		assettypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		icqtypes.ModuleName,
		icatypes.ModuleName,
		ibcratelimit.ModuleName,
		ibchookstypes.ModuleName,
		// TODO[wasm]: Delete or uncomment: wasmtypes.ModuleName, // must be after ibctransfer.
		triggertypes.ModuleName,
		oracletypes.ModuleName,
	}
	app.mm.SetOrderInitGenesis(moduleGenesisOrder...)
	app.mm.SetOrderExportGenesis(moduleGenesisOrder...)

	moduleMigrationOrder := []string{
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
		ibcexported.ModuleName,
		minttypes.ModuleName,
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		ibctm.ModuleName,
		ibctransfertypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		hold.ModuleName,
		exchange.ModuleName,
		consensusparamtypes.ModuleName,
		circuittypes.ModuleName,
		nft.ModuleName,
		ledger.ModuleName,
		registry.ModuleName,
		ibcratelimit.ModuleName,
		ibchookstypes.ModuleName,
		icatypes.ModuleName,
		icqtypes.ModuleName,
		// TODO[wasm]: Delete or uncomment: wasmtypes.ModuleName,

		attributetypes.ModuleName,
		markertypes.ModuleName,
		msgfeestypes.ModuleName,
		metadatatypes.ModuleName,
		nametypes.ModuleName,
		triggertypes.ModuleName,
		oracletypes.ModuleName,
		assettypes.ModuleName,

		// Last due to v0.44 issue: https://github.com/cosmos/cosmos-sdk/issues/10591
		authtypes.ModuleName,
	}
	app.mm.SetOrderMigrations(moduleMigrationOrder...)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.BaseApp.MsgServiceRouter(), app.GRPCQueryRouter())
	if err := app.mm.RegisterServices(app.configurator); err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
		// TODO[wasm]: Delete or uncomment: wasmtypes.ModuleName: provwasm.NewWrapper(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.NameKeeper, pioMessageRouter),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.mm.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler()
	app.setPostHandler()
	app.setFeeHandler()
	app.SetAggregateEventsFunc(piohandlers.AggregateEvents)

	// Register upgrade handlers and set the store loader.
	// This must be done after the module manager, configurator, and pre-blocker are set,
	// but before the baseapp is sealed via LoadLatestVersion() below.
	app.registerUpgradeHandlers()

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			cmtos.Exit(err.Error())
		}
	}

	simappparams.AppEncodingConfig = app.GetEncodingConfig()
	return app
}

func (app *App) setAnteHandler() {
	anteHandler, err := antewrapper.NewAnteHandler(
		antewrapper.HandlerOptions{
			AccountKeeper:       app.AccountKeeper,
			BankKeeper:          app.BankKeeper,
			TxSigningHandlerMap: app.txConfig.SignModeHandler(),
			FeegrantKeeper:      app.FeeGrantKeeper,
			MsgFeesKeeper:       app.MsgFeesKeeper,
			CircuitKeeper:       &app.CircuitKeeper,
			SigGasConsumer:      ante.DefaultSigVerificationGasConsumer,
		})
	if err != nil {
		panic(err)
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
}

func (app *App) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

func (app *App) setFeeHandler() {
	msgFeeHandler, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  app.AccountKeeper,
		BankKeeper:     app.BankKeeper,
		FeegrantKeeper: app.FeeGrantKeeper,
		MsgFeesKeeper:  app.MsgFeesKeeper,
		Decoder:        app.txConfig.TxDecoder(),
	})
	if err != nil {
		panic(err)
	}

	app.SetFeeHandler(msgFeeHandler)
}

func (app *App) registerUpgradeHandlers() {
	// Add the upgrade handlers for each release.
	InstallCustomUpgradeHandlers(app)

	// Use the dump of $home/data/upgrade-info.json:{"name":"$plan","height":321654} to determine
	// if we load a store upgrade from the handlers. No file == no error from read func.
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	// Currently in an upgrade hold for this block.
	var storeLoader baseapp.StoreLoader
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
			storeLoader = GetUpgradeStoreLoader(app, upgradeInfo)
		}
	}
	if storeLoader == nil {
		storeLoader = baseapp.DefaultStoreLoader
	}
	app.SetStoreLoader(storeLoader)
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
	return app.txConfig
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.mm.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	resp, err := app.mm.BeginBlock(ctx)
	if err != nil {
		return resp, err
	}
	resp.Events = filterBeginBlockerEvents(resp)
	return resp, nil
}

// filterBeginBlockerEvents filters out events from a given abci.ResponseBeginBlock according to the criteria defined in shouldFilterEvent.
func filterBeginBlockerEvents(responseBeginBlock sdk.BeginBlock) []abci.Event {
	filteredEvents := make([]abci.Event, 0)
	for _, e := range responseBeginBlock.Events {
		if shouldFilterEvent(e) {
			continue
		}
		filteredEvents = append(filteredEvents, e)
	}
	return filteredEvents
}

// shouldFilterEvent checks if an abci.Event should be filtered based on its type and attributes.
func shouldFilterEvent(e abci.Event) bool {
	if e.Type == distrtypes.EventTypeCommission || e.Type == distrtypes.EventTypeRewards || e.Type == distrtypes.EventTypeProposerReward || e.Type == banktypes.EventTypeTransfer || e.Type == banktypes.EventTypeCoinSpent || e.Type == banktypes.EventTypeCoinReceived {
		for _, a := range e.Attributes {
			if a.Key == sdk.AttributeKeyAmount && len(a.Value) == 0 {
				return true
			}
		}
	}
	return false
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

// GetEncodingConfig returns the various encoding configurations used in this app.
func (app *App) GetEncodingConfig() simappparams.EncodingConfig {
	return simappparams.EncodingConfig{
		InterfaceRegistry: app.InterfaceRegistry(),
		Marshaler:         app.AppCodec(),
		TxConfig:          app.GetTxConfig(),
		Amino:             app.LegacyAmino(),
	}
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *App) DefaultGenesis() map[string]json.RawMessage {
	return app.BasicModuleManager.DefaultGenesis(app.appCodec)
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetStoreKeys returns all the stored store keys.
func (app *App) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.keys))
	for _, key := range app.keys {
		keys = append(keys, key)
	}

	return keys
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

	// Register new CometBFT queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register swagger API
	if err := RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
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

// AutoCliOpts returns the autocli options for the app.
func (app *App) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	cfg := sdk.GetConfig()
	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.mm.Modules),
		AddressCodec:          authcodec.NewBech32Codec(cfg.GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(cfg.GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(cfg.GetBech32ConsensusAddrPrefix()),
	}
}

// RegisterSwaggerAPI provides a common function which registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router, swaggerEnabled bool) error {
	if !swaggerEnabled {
		return nil
	}

	root, err := fs.Sub(docs.SwaggerUI, "swagger-ui")
	if err != nil {
		return err
	}

	staticServer := http.FileServer(http.FS(root))
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))

	return nil
}

// injectUpgrade causes the named upgrade to be run as the chain starts.
//
// To use this, add a call to it in registerUpgradeHandlers immediately after the
// call to InstallCustomUpgradeHandlers, but before the line that looks for an upgrade file.
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

	// Define a new PreBlocker that will inject the upgrade.
	injected := false
	app.SetPreBlocker(func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		if !injected {
			app.Logger().Info("Injecting upgrade plan", "plan", plan)
			// Ideally, we'd just call ScheduleUpgrade(ctx, plan) here (and panic on error).
			// But the upgrade keeper often has its own migration stuff that change some store stuff.
			// ScheduleUpgrade would try to read some of that old state using the update pattern,
			// causing a failure. So we're doing things the hard way here.
			if err := app.UpgradeKeeper.ClearUpgradePlan(ctx); err != nil {
				panic(err)
			}
			store := ctx.KVStore(app.GetKey(upgradetypes.StoreKey))
			bz := app.appCodec.MustMarshal(&plan)
			store.Set(upgradetypes.PlanKey(), bz)
			injected = true
		}
		return app.PreBlocker(ctx, req)
	})
}
