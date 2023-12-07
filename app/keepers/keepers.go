package keepers

import (
	"path/filepath"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	quarantinekeeper "github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	sanctionkeeper "github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icq "github.com/cosmos/ibc-apps/modules/async-icq/v6"
	icqkeeper "github.com/cosmos/ibc-apps/modules/async-icq/v6/keeper"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v6/types"
	ica "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts"
	icahost "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/host/types"
	ibctransfer "github.com/cosmos/ibc-go/v6/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v6/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"

	appparams "github.com/provenance-io/provenance/app/params"
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/internal/provwasm"
	attributekeeper "github.com/provenance-io/provenance/x/attribute/keeper"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	attributewasm "github.com/provenance-io/provenance/x/attribute/wasm"
	"github.com/provenance-io/provenance/x/exchange"
	exchangekeeper "github.com/provenance-io/provenance/x/exchange/keeper"
	"github.com/provenance-io/provenance/x/hold"
	holdkeeper "github.com/provenance-io/provenance/x/hold/keeper"
	"github.com/provenance-io/provenance/x/ibchooks"
	ibchookskeeper "github.com/provenance-io/provenance/x/ibchooks/keeper"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
	ibcratelimit "github.com/provenance-io/provenance/x/ibcratelimit"
	ibcratelimitkeeper "github.com/provenance-io/provenance/x/ibcratelimit/keeper"
	ibcratelimitmodule "github.com/provenance-io/provenance/x/ibcratelimit/module"
	"github.com/provenance-io/provenance/x/marker"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	markerwasm "github.com/provenance-io/provenance/x/marker/wasm"
	metadatakeeper "github.com/provenance-io/provenance/x/metadata/keeper"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	metadatawasm "github.com/provenance-io/provenance/x/metadata/wasm"
	"github.com/provenance-io/provenance/x/msgfees"
	msgfeeskeeper "github.com/provenance-io/provenance/x/msgfees/keeper"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	msgfeeswasm "github.com/provenance-io/provenance/x/msgfees/wasm"
	"github.com/provenance-io/provenance/x/name"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	namewasm "github.com/provenance-io/provenance/x/name/wasm"
	oraclekeeper "github.com/provenance-io/provenance/x/oracle/keeper"
	oracle "github.com/provenance-io/provenance/x/oracle/module"
	oraclemodule "github.com/provenance-io/provenance/x/oracle/module"
	oracletypes "github.com/provenance-io/provenance/x/oracle/types"
	rewardkeeper "github.com/provenance-io/provenance/x/reward/keeper"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
	triggerkeeper "github.com/provenance-io/provenance/x/trigger/keeper"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"

	_ "github.com/provenance-io/provenance/client/docs/statik" // registers swagger-ui files with statik
)

type AppKeepers struct {
	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     crisiskeeper.Keeper
	UpgradeKeeper    upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	AuthzKeeper      authzkeeper.Keeper
	GroupKeeper      groupkeeper.Keeper
	EvidenceKeeper   evidencekeeper.Keeper
	FeeGrantKeeper   feegrantkeeper.Keeper
	MsgFeesKeeper    msgfeeskeeper.Keeper
	RewardKeeper     rewardkeeper.Keeper
	QuarantineKeeper quarantinekeeper.Keeper
	SanctionKeeper   sanctionkeeper.Keeper
	TriggerKeeper    triggerkeeper.Keeper
	OracleKeeper     oraclekeeper.Keeper

	IBCKeeper          *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCHooksKeeper     *ibchookskeeper.Keeper
	ICAHostKeeper      *icahostkeeper.Keeper
	TransferKeeper     *ibctransferkeeper.Keeper
	ICQKeeper          icqkeeper.Keeper
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

	ICAModule    ica.AppModule
	ICQModule    icq.AppModule
	OracleModule oracle.AppModule
}

// WasmWrapper allows us to use namespacing in the config file
// This is only used for parsing in the app, x/wasm expects WasmConfig
type WasmWrapper struct {
	Wasm wasm.Config `mapstructure:"wasm"`
}

func NewAppKeeper(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	legacyAmino *codec.LegacyAmino,
	maccPerms map[string][]string,
	modAccAddrs map[string]bool,
	//blockedAddress map[string]bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	appOpts servertypes.AppOptions,
	prefix string,
	logger log.Logger,
	interfaceRegistry types.InterfaceRegistry,
	encodingConfig appparams.EncodingConfig,

) AppKeepers {
	appKeepers := AppKeepers{}

	appKeepers.GenerateKeys()

	/*
		configure state listening capabilities using AppOptions
		we are doing nothing with the returned streamingServices and waitGroup in this case
	*/

	// Not applicable
	//if _, _, err := streaming.LoadStreamingServices(bApp, appOpts, appCodec, appKeepers.keys); err != nil {
	//	tmos.Exit(err.Error())
	//}

	appKeepers.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		appKeepers.keys[paramstypes.StoreKey],
		appKeepers.tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	bApp.SetParamStore(appKeepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, appKeepers.keys[capabilitytypes.StoreKey], appKeepers.memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := appKeepers.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := appKeepers.CapabilityKeeper.ScopeToModule(wasm.ModuleName)
	scopedICAHostKeeper := appKeepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedICQKeeper := appKeepers.CapabilityKeeper.ScopeToModule(icqtypes.ModuleName)
	scopedOracleKeeper := appKeepers.CapabilityKeeper.ScopeToModule(oracletypes.ModuleName)

	// capability keeper must be sealed after scope to module registrations are completed.
	appKeepers.CapabilityKeeper.Seal()

	// add keepers
	appKeepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, appKeepers.keys[authtypes.StoreKey], appKeepers.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms, prefix,
	)

	appKeepers.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec, appKeepers.keys[banktypes.StoreKey], appKeepers.AccountKeeper, appKeepers.GetSubspace(banktypes.ModuleName), modAccAddrs,
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec, appKeepers.keys[stakingtypes.StoreKey], appKeepers.AccountKeeper, appKeepers.BankKeeper, appKeepers.GetSubspace(stakingtypes.ModuleName),
	)
	appKeepers.MintKeeper = mintkeeper.NewKeeper(
		appCodec, appKeepers.keys[minttypes.StoreKey], appKeepers.GetSubspace(minttypes.ModuleName), &stakingKeeper,
		appKeepers.AccountKeeper, appKeepers.BankKeeper, authtypes.FeeCollectorName,
	)
	appKeepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, appKeepers.keys[distrtypes.StoreKey], appKeepers.GetSubspace(distrtypes.ModuleName), appKeepers.AccountKeeper, appKeepers.BankKeeper,
		&stakingKeeper, authtypes.FeeCollectorName,
	)
	appKeepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, appKeepers.keys[slashingtypes.StoreKey], &stakingKeeper, appKeepers.GetSubspace(slashingtypes.ModuleName),
	)
	appKeepers.CrisisKeeper = crisiskeeper.NewKeeper(
		appKeepers.GetSubspace(crisistypes.ModuleName), invCheckPeriod, appKeepers.BankKeeper, authtypes.FeeCollectorName,
	)

	appKeepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, appKeepers.keys[feegrant.StoreKey], appKeepers.AccountKeeper)
	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights, appKeepers.keys[upgradetypes.StoreKey], appCodec, homePath, bApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.MsgFeesKeeper = msgfeeskeeper.NewKeeper(
		appCodec, appKeepers.keys[msgfeestypes.StoreKey], appKeepers.GetSubspace(msgfeestypes.ModuleName), authtypes.FeeCollectorName, pioconfig.GetProvenanceConfig().FeeDenom, bApp.Simulate, encodingConfig.TxConfig.TxDecoder(), interfaceRegistry)

	pioMsgFeesRouter := bApp.MsgServiceRouter().(*piohandlers.PioMsgServiceRouter)
	pioMsgFeesRouter.SetMsgFeesKeeper(appKeepers.MsgFeesKeeper)

	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	restrictHooks := piohandlers.NewStakingRestrictionHooks(&appKeepers.StakingKeeper, *piohandlers.DefaultRestrictionOptions)
	appKeepers.StakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(restrictHooks, appKeepers.DistrKeeper.Hooks(), appKeepers.SlashingKeeper.Hooks()),
	)

	appKeepers.RewardKeeper = rewardkeeper.NewKeeper(appCodec, appKeepers.keys[rewardtypes.StoreKey], appKeepers.StakingKeeper, &appKeepers.GovKeeper, appKeepers.BankKeeper, appKeepers.AccountKeeper)

	appKeepers.AuthzKeeper = authzkeeper.NewKeeper(
		appKeepers.keys[authzkeeper.StoreKey], appCodec, bApp.MsgServiceRouter(), appKeepers.AccountKeeper,
	)

	appKeepers.GroupKeeper = groupkeeper.NewKeeper(appKeepers.keys[group.StoreKey], appCodec, bApp.MsgServiceRouter(), appKeepers.AccountKeeper, group.DefaultConfig())

	// Create IBC Keeper
	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, appKeepers.keys[ibchost.StoreKey], appKeepers.GetSubspace(ibchost.ModuleName), appKeepers.StakingKeeper, appKeepers.UpgradeKeeper, scopedIBCKeeper,
	)

	// Configure the hooks keeper
	hooksKeeper := ibchookskeeper.NewKeeper(
		appKeepers.keys[ibchookstypes.StoreKey],
		appKeepers.GetSubspace(ibchookstypes.ModuleName),
		appKeepers.IBCKeeper.ChannelKeeper,
		nil,
	)
	appKeepers.IBCHooksKeeper = &hooksKeeper

	// Setup the ICS4Wrapper used by the hooks middleware
	addrPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()        // We use this approach so running tests which use "cosmos" will work while we use "pb"
	wasmHooks := ibchooks.NewWasmHooks(&hooksKeeper, nil, addrPrefix) // The contract keeper needs to be set later
	appKeepers.Ics20WasmHooks = &wasmHooks
	markerHooks := ibchooks.NewMarkerHooks(nil)
	appKeepers.Ics20MarkerHooks = &markerHooks
	ibcHooks := ibchooks.NewIbcHooks(appCodec, &hooksKeeper, appKeepers.IBCKeeper, appKeepers.Ics20WasmHooks, appKeepers.Ics20MarkerHooks, nil)
	appKeepers.IbcHooks = &ibcHooks

	appKeepers.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IbcHooks,
	)

	rateLimtingKeeper := ibcratelimitkeeper.NewKeeper(appCodec, appKeepers.keys[ibcratelimit.StoreKey], nil)
	appKeepers.RateLimitingKeeper = &rateLimtingKeeper

	// Create Transfer Keepers
	rateLimitingTransferModule := ibcratelimitmodule.NewIBCMiddleware(nil, appKeepers.HooksICS4Wrapper, appKeepers.RateLimitingKeeper)
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibctransfertypes.StoreKey],
		appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		&rateLimitingTransferModule,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		scopedTransferKeeper,
	)
	appKeepers.TransferKeeper = &transferKeeper
	transferModule := ibctransfer.NewIBCModule(*appKeepers.TransferKeeper)
	rateLimitingTransferModule = *rateLimitingTransferModule.WithIBCModule(transferModule)
	hooksTransferModule := ibchooks.NewIBCMiddleware(&rateLimitingTransferModule, &appKeepers.HooksICS4Wrapper)
	appKeepers.TransferStack = &hooksTransferModule

	appKeepers.NameKeeper = namekeeper.NewKeeper(
		appCodec, appKeepers.keys[nametypes.StoreKey], appKeepers.GetSubspace(nametypes.ModuleName),
	)

	appKeepers.AttributeKeeper = attributekeeper.NewKeeper(
		appCodec, appKeepers.keys[attributetypes.StoreKey], appKeepers.GetSubspace(attributetypes.ModuleName), appKeepers.AccountKeeper, &appKeepers.NameKeeper,
	)

	appKeepers.MetadataKeeper = metadatakeeper.NewKeeper(
		appCodec, appKeepers.keys[metadatatypes.StoreKey], appKeepers.GetSubspace(metadatatypes.ModuleName), appKeepers.AccountKeeper, appKeepers.AuthzKeeper, appKeepers.AttributeKeeper,
	)

	markerReqAttrBypassAddrs := []sdk.AccAddress{
		authtypes.NewModuleAddress(authtypes.FeeCollectorName),     // Allow collecting fees in restricted coins.
		authtypes.NewModuleAddress(rewardtypes.ModuleName),         // Allow rewards to hold onto restricted coins.
		authtypes.NewModuleAddress(quarantine.ModuleName),          // Allow quarantine to hold onto restricted coins.
		authtypes.NewModuleAddress(govtypes.ModuleName),            // Allow restricted coins in deposits.
		authtypes.NewModuleAddress(distrtypes.ModuleName),          // Allow fee denoms to be restricted coins.
		authtypes.NewModuleAddress(stakingtypes.BondedPoolName),    // Allow bond denom to be a restricted coin.
		authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName), // Allow bond denom to be a restricted coin.
	}
	appKeepers.MarkerKeeper = markerkeeper.NewKeeper(
		appCodec, appKeepers.keys[markertypes.StoreKey], appKeepers.GetSubspace(markertypes.ModuleName),
		appKeepers.AccountKeeper, appKeepers.BankKeeper, appKeepers.AuthzKeeper, appKeepers.FeeGrantKeeper,
		appKeepers.AttributeKeeper, appKeepers.NameKeeper, appKeepers.TransferKeeper, markerReqAttrBypassAddrs,
	)

	appKeepers.HoldKeeper = holdkeeper.NewKeeper(
		appCodec, appKeepers.keys[hold.StoreKey], appKeepers.BankKeeper,
	)

	appKeepers.ExchangeKeeper = exchangekeeper.NewKeeper(
		appCodec, appKeepers.keys[exchange.StoreKey], authtypes.FeeCollectorName,
		appKeepers.AccountKeeper, appKeepers.AttributeKeeper, appKeepers.BankKeeper, appKeepers.HoldKeeper, appKeepers.MarkerKeeper,
	)

	pioMessageRouter := MessageRouterFunc(func(msg sdk.Msg) baseapp.MsgServiceHandler {
		return pioMsgFeesRouter.Handler(msg)
	})
	appKeepers.TriggerKeeper = triggerkeeper.NewKeeper(appCodec, appKeepers.keys[triggertypes.StoreKey], bApp.MsgServiceRouter())
	icaHostKeeper := icahostkeeper.NewKeeper(
		appCodec, appKeepers.keys[icahosttypes.StoreKey], appKeepers.GetSubspace(icahosttypes.SubModuleName),
		appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCKeeper.ChannelKeeper, &appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper, scopedICAHostKeeper, pioMessageRouter,
	)
	appKeepers.ICAHostKeeper = &icaHostKeeper
	appKeepers.ICAModule = ica.NewAppModule(nil, appKeepers.ICAHostKeeper)
	icaHostIBCModule := icahost.NewIBCModule(*appKeepers.ICAHostKeeper)

	appKeepers.ICQKeeper = icqkeeper.NewKeeper(
		appCodec, appKeepers.keys[icqtypes.StoreKey], appKeepers.GetSubspace(icqtypes.ModuleName),
		appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCKeeper.ChannelKeeper, &appKeepers.IBCKeeper.PortKeeper,
		scopedICQKeeper, bApp.GRPCQueryRouter(),
	)
	appKeepers.ICQModule = icq.NewAppModule(appKeepers.ICQKeeper)
	icqIBCModule := icq.NewIBCModule(appKeepers.ICQKeeper)

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
	querierRegistry.RegisterQuerier(nametypes.RouterKey, namewasm.Querier(appKeepers.NameKeeper))
	querierRegistry.RegisterQuerier(attributetypes.RouterKey, attributewasm.Querier(appKeepers.AttributeKeeper))
	querierRegistry.RegisterQuerier(markertypes.RouterKey, markerwasm.Querier(appKeepers.MarkerKeeper))
	querierRegistry.RegisterQuerier(metadatatypes.RouterKey, metadatawasm.Querier(appKeepers.MetadataKeeper))

	// Add the staking feature and indicate that provwasm contracts can be run on this chain.
	// Addition of cosmwasm_1_1 adds capability defined here: https://github.com/CosmWasm/cosmwasm/pull/1356
	supportedFeatures := "staking,provenance,stargate,iterator,cosmwasm_1_1"

	// The last arguments contain custom message handlers, and custom query handlers,
	// to allow smart contracts to use provenance modules.
	wasmKeeperInstance := wasm.NewKeeper(
		appCodec,
		appKeepers.keys[wasm.StoreKey],
		appKeepers.GetSubspace(wasm.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		appKeepers.DistrKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		appKeepers.TransferKeeper,
		pioMessageRouter,
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmkeeper.WithQueryPlugins(provwasm.QueryPlugins(querierRegistry, *bApp.GRPCQueryRouter(), appCodec)),
		wasmkeeper.WithMessageEncoders(provwasm.MessageEncoders(encoderRegistry, logger)),
	)
	appKeepers.WasmKeeper = &wasmKeeperInstance

	// Pass the wasm keeper to all the wrappers that need it
	appKeepers.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(appKeepers.WasmKeeper)
	appKeepers.Ics20WasmHooks.ContractKeeper = appKeepers.WasmKeeper // app.ContractKeeper -- this changes in the next version of wasm to a permissioned keeper
	appKeepers.IBCHooksKeeper.ContractKeeper = appKeepers.ContractKeeper
	appKeepers.Ics20MarkerHooks.MarkerKeeper = &appKeepers.MarkerKeeper
	appKeepers.RateLimitingKeeper.PermissionedKeeper = appKeepers.ContractKeeper

	appKeepers.IbcHooks.SendPacketPreProcessors = []ibchookstypes.PreSendPacketDataProcessingFn{appKeepers.Ics20MarkerHooks.SetupMarkerMemoFn, appKeepers.Ics20WasmHooks.GetWasmSendPacketPreProcessor}

	appKeepers.ScopedOracleKeeper = scopedOracleKeeper
	appKeepers.OracleKeeper = *oraclekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[oracletypes.StoreKey],
		appKeepers.keys[oracletypes.MemStoreKey],
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		scopedOracleKeeper,
		wasmkeeper.Querier(appKeepers.WasmKeeper),
	)
	appKeepers.OracleModule = oraclemodule.NewAppModule(appCodec, appKeepers.OracleKeeper, appKeepers.AccountKeeper, appKeepers.BankKeeper, appKeepers.IBCKeeper.ChannelKeeper)

	unsanctionableAddrs := make([]sdk.AccAddress, 0, len(maccPerms)+1)
	for mName := range maccPerms {
		unsanctionableAddrs = append(unsanctionableAddrs, authtypes.NewModuleAddress(mName))
	}
	unsanctionableAddrs = append(unsanctionableAddrs, authtypes.NewModuleAddress(quarantine.ModuleName))
	appKeepers.SanctionKeeper = sanctionkeeper.NewKeeper(appCodec, appKeepers.keys[sanction.StoreKey],
		appKeepers.BankKeeper, &appKeepers.GovKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(), unsanctionableAddrs)

	// register the proposal types
	govRouter := govtypesv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(appKeepers.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(appKeepers.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(appKeepers.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(appKeepers.IBCKeeper.ClientKeeper)).
		AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(appKeepers.WasmKeeper, wasm.EnableAllProposals)).
		AddRoute(nametypes.ModuleName, name.NewProposalHandler(appKeepers.NameKeeper)).
		AddRoute(markertypes.ModuleName, marker.NewProposalHandler(appKeepers.MarkerKeeper)).
		AddRoute(msgfeestypes.ModuleName, msgfees.NewProposalHandler(appKeepers.MsgFeesKeeper, interfaceRegistry))
	appKeepers.GovKeeper = govkeeper.NewKeeper(
		appCodec, appKeepers.keys[govtypes.StoreKey], appKeepers.GetSubspace(govtypes.ModuleName), appKeepers.AccountKeeper, appKeepers.BankKeeper,
		&stakingKeeper, govRouter, bApp.MsgServiceRouter(), govtypes.Config{MaxMetadataLen: 10000},
	)
	appKeepers.GovKeeper.SetHooks(govtypes.NewMultiGovHooks(appKeepers.SanctionKeeper))

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.
		AddRoute(ibctransfertypes.ModuleName, appKeepers.TransferStack).
		AddRoute(wasm.ModuleName, wasm.NewIBCHandler(appKeepers.WasmKeeper, appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCKeeper.ChannelKeeper)).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(icqtypes.ModuleName, icqIBCModule).
		AddRoute(oracletypes.ModuleName, appKeepers.OracleModule)
	appKeepers.IBCKeeper.SetRouter(ibcRouter)

	// Create evidence Keeper for to register the IBC light client misbehavior evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, appKeepers.keys[evidencetypes.StoreKey], &appKeepers.StakingKeeper, appKeepers.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	appKeepers.EvidenceKeeper = *evidenceKeeper

	appKeepers.QuarantineKeeper = quarantinekeeper.NewKeeper(appCodec, appKeepers.keys[quarantine.StoreKey], appKeepers.BankKeeper, authtypes.NewModuleAddress(quarantine.ModuleName))

	appKeepers.ScopedIBCKeeper = scopedIBCKeeper
	appKeepers.ScopedTransferKeeper = scopedTransferKeeper
	appKeepers.ScopedICQKeeper = scopedICQKeeper
	appKeepers.ScopedICAHostKeeper = scopedICAHostKeeper

	return appKeepers
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

	paramsKeeper.Subspace(metadatatypes.ModuleName)
	paramsKeeper.Subspace(markertypes.ModuleName)
	paramsKeeper.Subspace(nametypes.ModuleName)
	paramsKeeper.Subspace(attributetypes.ModuleName)
	paramsKeeper.Subspace(msgfeestypes.ModuleName)
	paramsKeeper.Subspace(wasm.ModuleName)
	paramsKeeper.Subspace(rewardtypes.ModuleName)
	paramsKeeper.Subspace(triggertypes.ModuleName)

	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(icqtypes.ModuleName)
	paramsKeeper.Subspace(ibchookstypes.ModuleName)

	return paramsKeeper
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}
