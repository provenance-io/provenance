package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	cmtrand "github.com/cometbft/cometbft/libs/rand"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	provenanceapp "github.com/provenance-io/provenance/app"
)

// NewAppConstructor returns a new provenanceapp AppConstructor
func NewAppConstructor() testnet.AppConstructor {
	return func(val testnet.ValidatorI) servertypes.Application {
		ctx := val.GetCtx()
		appCfg := val.GetAppConfig()
		return provenanceapp.New(
			ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), ctx.Config.RootDir, 0,
			simtestutil.EmptyAppOptions{},
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(appCfg.Pruning)),
			baseapp.SetMinGasPrices(appCfg.MinGasPrices),
			baseapp.SetChainID(ctx.Viper.GetString(flags.FlagChainID)),
		)
	}
}

// DefaultTestNetworkConfig creates a network configuration for inproc testing
func DefaultTestNetworkConfig() testnet.Config {
	tempDir, err := os.MkdirTemp("", "tempprovapp")
	if err != nil {
		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
	}
	defer os.RemoveAll(tempDir)

	tempApp := provenanceapp.New(
		log.NewNopLogger(), dbm.NewMemDB(), nil, true, make(map[int64]bool), tempDir, 0,
		simtestutil.NewAppOptionsWithFlagHome(tempDir),
	)
	encCfg := provenanceapp.MakeTestEncodingConfig(nil)

	return testnet.Config{
		Codec:             encCfg.Marshaler,
		TxConfig:          encCfg.TxConfig,
		LegacyAmino:       encCfg.Amino,
		InterfaceRegistry: encCfg.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor:    NewAppConstructor(),
		GenesisState:      tempApp.DefaultGenesis(),
		TimeoutCommit:     500 * time.Millisecond,
		ChainID:           "chain-" + cmtrand.NewRand().Str(6),
		NumValidators:     4,
		BondDenom:         sdk.DefaultBondDenom, // we use the SDK bond denom here, at least until the entire genesis is rewritten to match bond denom
		MinGasPrices:      fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
		AccountTokens:     sdk.TokensFromConsensusPower(1000000, sdk.DefaultPowerReduction),
		StakingTokens:     sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:      sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		PruningStrategy:   pruningtypes.PruningOptionNothing,
		CleanupDir:        true,
		SigningAlgo:       string(hd.Secp256k1Type),
		KeyringOptions:    []keyring.Option{},
	}
}

func CleanUp(n *testnet.Network, t *testing.T) {
	if n == nil {
		t.Log("nothing to tear down")
		return
	}
	t.Log("teardown waiting for next block")
	//nolint:errcheck // The test shouldn't fail because cleanup was a problem. So ignoring any error from this.
	n.WaitForNextBlock()
	t.Log("teardown cleaning up testnet")
	n.Cleanup()
	// Give things a chance to finish closing up. Hopefully will prevent things like address collisions. 100ms chosen randomly.
	time.Sleep(100 * time.Millisecond)
	t.Log("teardown done")
}
