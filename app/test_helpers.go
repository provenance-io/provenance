package app

// DONTCOVER

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal"
	"github.com/provenance-io/provenance/internal/pioconfig"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
)

// DefaultConsensusParams defines the default Tendermint consensus params used in
// SimApp testing.
var DefaultConsensusParams = &cmttmtypes.ConsensusParams{
	Block: &cmttmtypes.BlockParams{
		MaxBytes: 200000,
		MaxGas:   60_000_000,
	},
	Evidence: &cmttmtypes.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &cmttmtypes.ValidatorParams{
		PubKeyTypes: []string{
			cmttypes.ABCIPubKeyTypeEd25519,
		},
	},
}

// SetupOptions defines arguments that are passed into `Simapp` constructor.
type SetupOptions struct {
	Logger             log.Logger
	DB                 *dbm.MemDB
	InvCheckPeriod     uint
	HomePath           string
	SkipUpgradeHeights map[int64]bool
	EncConfig          params.EncodingConfig
	AppOpts            servertypes.AppOptions
	ChainID            string
}

func setup(t *testing.T, withGenesis bool, invCheckPeriod uint, chainID string) (*App, GenesisState) {
	db := dbm.NewMemDB()
	encCdc := MakeEncodingConfig()
	// set default config if not set by the flow
	if len(pioconfig.GetProvenanceConfig().FeeDenom) == 0 {
		pioconfig.SetProvenanceConfig("", 0)
	}

	app := New(loggerMaker(), db, nil, true, map[int64]bool{}, t.TempDir(), invCheckPeriod, encCdc, simtestutil.EmptyAppOptions{}, baseapp.SetChainID(chainID))
	if withGenesis {
		return app, NewDefaultGenesisState(encCdc.Marshaler)
	}
	return app, GenesisState{}
}

// A LoggerMakerFn is a function that makes a logger.
type LoggerMakerFn func() log.Logger

// loggerMaker is used during app setup for unit test.
// The default is a no-op logger.
// There's no way to update the logger after it's provided to New.
// Using SetLoggerMaker, though, we can enable logging for a unit test that needs help.
var loggerMaker LoggerMakerFn = log.NewNopLogger

// SetLoggerMaker sets the global loggerMaker variable (used for test setup) and returns what it was before.
//
// Example usage: defer SetLoggerMaker(SetLoggerMaker(NewDebugLogger))
//
// The inside SetLoggerMaker(NewDebugLogger) is called immediately at that line and it's result is
// defined as the argument to the outside SetLoggerMaker which is invoked via defer when the function returns.
// Basically, it temporarily changes the loggerMaker for the duration of the function in question.
// That line would be added before a call to one of the app setup functions, e.g. SetupWithGenesisRewardsProgram.
//
// This function should never be called in any committed code. It's only for test troubleshooting.
func SetLoggerMaker(newLoggerMaker LoggerMakerFn) LoggerMakerFn {
	orig := loggerMaker
	loggerMaker = newLoggerMaker
	return orig
}

// NewDebugLogger creates a new logger to stdout with level debug.
// Standard usage: defer SetLoggerMaker(SetLoggerMaker(NewDebugLogger))
func NewDebugLogger() log.Logger {
	lw := zerolog.ConsoleWriter{Out: os.Stdout}
	logger := zerolog.New(lw).Level(zerolog.DebugLevel).With().Timestamp().Logger()
	return log.NewCustomLogger(logger)
}

// NewInfoLogger creates a new logger to stdout with level info.
// Standard usage: defer SetLoggerMaker(SetLoggerMaker(NewInfoLogger))
func NewInfoLogger() log.Logger {
	lw := zerolog.ConsoleWriter{Out: os.Stdout}
	logger := zerolog.New(lw).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	return log.NewCustomLogger(logger)
}

// BufferedInfoLoggerMaker returns a logger maker function for a NewBufferedInfoLogger.
// Error log lines will start with "ERR ".
// Info log lines will start with "INF ".
func BufferedInfoLoggerMaker(buffer *bytes.Buffer) LoggerMakerFn {
	return func() log.Logger {
		return internal.NewBufferedInfoLogger(buffer)
	}
}

// NewAppWithCustomOptions initializes a new SimApp with custom options.
func NewAppWithCustomOptions(t *testing.T, isCheckTx bool, options SetupOptions) *App {
	t.Helper()
	pioconfig.SetProvenanceConfig("", 0)
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)
	// create validator set with single validator
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000_000_000)),
	}

	app := New(options.Logger, options.DB, nil, true, options.SkipUpgradeHeights, options.HomePath, options.InvCheckPeriod, options.EncConfig, options.AppOpts)
	genesisState := NewDefaultGenesisState(app.appCodec)
	genesisState = genesisStateWithValSet(t, app, genesisState, valSet, []authtypes.GenesisAccount{acc}, balance)

	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
		require.NoError(t, err)

		// Initialize the chain
		_, err = app.InitChain(
			&abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
				ChainId:         options.ChainID,
			},
		)
		require.NoError(t, err, "InitChain")
	}

	return app
}

// Setup initializes a new App. A Nop logger is set in App.
func Setup(t *testing.T) *App {
	t.Helper()
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000_000_000)),
	}

	app := SetupWithGenesisValSet(t, "", valSet, []authtypes.GenesisAccount{acc}, balance)

	return app
}

func genesisStateWithValSet(t *testing.T,
	app *App, genesisState GenesisState,
	valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) GenesisState {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		require.NoError(t, err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(t, err)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), sdk.ValAddress(val.Address).String(), sdkmath.LegacyOneDec()))
	}

	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	if len(delegations) > 0 {
		bondCoins := sdk.NewCoin(sdk.DefaultBondDenom, bondAmt.MulRaw(int64(len(delegations))))
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(bondCoins)
		// add bonded amount to bonded pool module account
		balances = append(balances, banktypes.Balance{
			Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
			Coins:   sdk.Coins{bondCoins},
		})
	}

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	return genesisState
}

// SetupQuerier initializes a new App without genesis and without calling InitChain.
func SetupQuerier(t *testing.T) *App {
	app, _ := setup(t, false, 0, "")
	return app
}

// SetupWithGenesisValSet initializes a new App with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit in the default token of the app from first genesis
// account. A Nop logger is set in App.
func SetupWithGenesisValSet(t *testing.T, chainID string, valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *App {
	t.Helper()

	app, genesisState := setup(t, true, 5, chainID)
	genesisState = genesisStateWithValSet(t, app, genesisState, valSet, genAccs, balances...)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err, "MarshalIndent genesis state")

	// init chain will set the validator set and initialize the genesis accounts
	_, err = app.InitChain(
		&abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
			ChainId:         chainID,
		},
	)
	require.NoError(t, err, "InitChain")

	// commit genesis changes
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Hash:               app.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
	})
	require.NoError(t, err, "FinalizeBlock")

	return app
}

// SetupWithGenesisAccounts initializes a new App with the provided genesis
// accounts and possible balances.
func SetupWithGenesisAccounts(t *testing.T, chainID string, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *App {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	return SetupWithGenesisValSet(t, chainID, valSet, genAccs, balances...)
}

// GenesisStateWithSingleValidator initializes GenesisState with a single validator and genesis accounts
// that also act as delegators.
func GenesisStateWithSingleValidator(t *testing.T, app *App) GenesisState {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balances := []banktypes.Balance{
		{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000_000_000)),
		},
	}

	genesisState := NewDefaultGenesisState(app.appCodec)
	genesisState = genesisStateWithValSet(t, app, genesisState, valSet, []authtypes.GenesisAccount{acc}, balances...)

	return genesisState
}

type GenerateAccountStrategy func(int) []sdk.AccAddress

// createRandomAccounts is a strategy used by addTestAddrs() in order to generated addresses in random order.
func createRandomAccounts(accNum int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, accNum)
	for i := 0; i < accNum; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

// createIncrementalAccounts is a strategy used by addTestAddrs() in order to generated addresses in ascending order.
func createIncrementalAccounts(accNum int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (accNum + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") // base address string

		buffer.WriteString(numString) // adding on final two digits to make addresses unique
		res, _ := sdk.AccAddressFromHexUnsafe(buffer.String())
		bech := res.String()
		addr, _ := TestAddr(buffer.String(), bech)

		addresses = append(addresses, addr)
		buffer.Reset()
	}

	return addresses
}

// createIncrementalAccountsLong is a strategy used by addTestAddrs() in order to generate 32-byte addresses in ascending order.
func createIncrementalAccountsLong(accNum int) []sdk.AccAddress {
	var addresses []sdk.AccAddress

	// There's nothing special about this base other than it's 30 bytes long (60 hex chars => 30 bytes).
	// That leaves 2 bytes for the incrementing number = 65536 addrs max.
	// It's the result of two calls to uuidgen with the last 4 chars removed.
	base := "9B4006D1F9794F07BEC52279C3C31480CCC9A1EB3FD64F628CC405E4E2E2"
	for i := 0; i < accNum; i++ {
		addrHex := fmt.Sprintf("%s%04X", base, i)
		addr, _ := sdk.AccAddressFromHexUnsafe(addrHex)
		addresses = append(addresses, addr)
	}

	return addresses
}

// AddTestAddrsFromPubKeys adds the addresses into the App providing only the public keys.
func AddTestAddrsFromPubKeys(app *App, ctx sdk.Context, pubKeys []cryptotypes.PubKey, accAmt sdkmath.Int) {
	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	if err != nil || bondDenom == "" {
		bondDenom = sdk.DefaultBondDenom
	}
	initCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, accAmt))

	for _, pk := range pubKeys {
		initAccountWithCoins(app, ctx, sdk.AccAddress(pk.Address()), initCoins)
	}
}

// AddTestAddrs constructs and returns accNum amount of accounts with an initial balance of accAmt in random order.
func AddTestAddrs(app *App, ctx sdk.Context, accNum int, accAmt sdkmath.Int) []sdk.AccAddress {
	return addTestAddrs(app, ctx, accNum, accAmt, createRandomAccounts)
}

// AddTestAddrsIncremental creates accNum accounts with 20-byte incrementing addresses and initialBondAmount of bond denom.
func AddTestAddrsIncremental(app *App, ctx sdk.Context, accNum int, accAmt sdkmath.Int) []sdk.AccAddress {
	return addTestAddrs(app, ctx, accNum, accAmt, createIncrementalAccounts)
}

// AddTestAddrsIncrementalLong creates accNum accounts with 32-byte incrementing addresses and initialBondAmount of bond denom.
func AddTestAddrsIncrementalLong(app *App, ctx sdk.Context, accNum int, initialBondAmount sdkmath.Int) []sdk.AccAddress {
	return addTestAddrs(app, ctx, accNum, initialBondAmount, createIncrementalAccountsLong)
}

func addTestAddrs(app *App, ctx sdk.Context, accNum int, accAmt sdkmath.Int, strategy GenerateAccountStrategy) []sdk.AccAddress {
	testAddrs := strategy(accNum)

	bondDenom, err := app.StakingKeeper.BondDenom(ctx)
	if err != nil || bondDenom == "" {
		bondDenom = sdk.DefaultBondDenom
	}
	initCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, accAmt))

	for _, addr := range testAddrs {
		initAccountWithCoins(app, ctx, addr, initCoins)
	}

	return testAddrs
}

func initAccountWithCoins(app *App, ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	if err != nil {
		panic(err)
	}

	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, coins)
	if err != nil {
		panic(err)
	}
}

func TestAddr(addr string, bech string) (sdk.AccAddress, error) {
	res, err := sdk.AccAddressFromHexUnsafe(addr)
	if err != nil {
		return nil, err
	}
	bechexpected := res.String()
	if bech != bechexpected {
		return nil, fmt.Errorf("bech encoding doesn't match reference")
	}

	bechres, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(bechres, res) {
		return nil, err
	}

	return res, nil
}

// CreateTestPubKeys returns a total of numPubKeys public keys in ascending order.
func CreateTestPubKeys(numPubKeys int) []cryptotypes.PubKey {
	var publicKeys []cryptotypes.PubKey
	var buffer bytes.Buffer

	// start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") // base pubkey string
		buffer.WriteString(numString)                                                       // adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, NewPubKeyFromHex(buffer.String()))
		buffer.Reset()
	}

	return publicKeys
}

// NewPubKeyFromHex returns a PubKey from a hex string.
func NewPubKeyFromHex(pk string) (res cryptotypes.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	if len(pkBytes) != ed25519.PubKeySize {
		panic(errors.ErrInvalidPubKey.Wrap("invalid pubkey size"))
	}
	return &ed25519.PubKey{Key: pkBytes}
}

// SetupWithGenesisRewardsProgram initializes a new SimApp with the provided
// rewards programs, genesis accounts, validators, and balances.
func SetupWithGenesisRewardsProgram(t *testing.T, nextRewardProgramID uint64, genesisRewards []rewardtypes.RewardProgram, genAccs []authtypes.GenesisAccount, valSet *cmttypes.ValidatorSet, balances ...banktypes.Balance) *App {
	t.Helper()

	// Make sure there's a validator set with at least one validator in it.
	if valSet == nil || len(valSet.Validators) == 0 {
		privVal := mock.NewPV()
		pubKey, err := privVal.GetPubKey()
		require.NoError(t, err)
		validator := cmttypes.NewValidator(pubKey, 1)
		if valSet == nil {
			valSet = cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})
		} else {
			require.NoError(t, valSet.UpdateWithChangeSet([]*cmttypes.Validator{validator}))
		}
	}

	app, genesisState := setup(t, true, 0, "")
	genesisState = genesisStateWithValSet(t, app, genesisState, valSet, genAccs, balances...)
	genesisState = genesisStateWithRewards(t, app, genesisState, nextRewardProgramID, genesisRewards)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err, "marshaling genesis state to json")

	_, err = app.InitChain(
		&abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	require.NoError(t, err, "InitChain")

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Hash:               app.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
		Time:               time.Now().UTC(),
	})
	require.NoError(t, err, "FinalizeBlock")

	return app
}

func genesisStateWithRewards(t *testing.T,
	app *App, genesisState GenesisState,
	nextRewardProgramID uint64, genesisRewards []rewardtypes.RewardProgram,
) GenesisState {
	rewardGenesisState := rewardtypes.NewGenesisState(
		nextRewardProgramID,
		genesisRewards,
		[]rewardtypes.ClaimPeriodRewardDistribution{},
		[]rewardtypes.RewardAccountState{},
	)
	var err error
	genesisState[rewardtypes.ModuleName], err = app.AppCodec().MarshalJSON(rewardGenesisState)
	require.NoError(t, err, "marshaling reward genesis state JSON")
	return genesisState
}
