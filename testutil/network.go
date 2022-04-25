package testutil

import (
	"fmt"
	"testing"
	"time"

	tmrand "github.com/tendermint/tendermint/libs/rand"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	provenanceapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/app/params"
)

// NewAppConstructor returns a new provenanceapp AppConstructor
func NewAppConstructor(encodingCfg params.EncodingConfig) testnet.AppConstructor {
	return func(val testnet.Validator) servertypes.Application {
		return provenanceapp.New(
			val.Ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
			encodingCfg,
			sdksim.EmptyAppOptions{},
			baseapp.SetPruning(storetypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
}

// DefaultTestNetworkConfig creates a network configuration for inproc testing
func DefaultTestNetworkConfig() testnet.Config {
	encCfg := provenanceapp.MakeEncodingConfig()
	return testnet.Config{
		Codec:             encCfg.Marshaler,
		TxConfig:          encCfg.TxConfig,
		LegacyAmino:       encCfg.Amino,
		InterfaceRegistry: encCfg.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor:    NewAppConstructor(encCfg),
		GenesisState:      provenanceapp.ModuleBasics.DefaultGenesis(encCfg.Marshaler),
		TimeoutCommit:     2 * time.Second,
		ChainID:           "chain-" + tmrand.NewRand().Str(6),
		NumValidators:     4,
		BondDenom:         sdk.DefaultBondDenom, // we use the SDK bond denom here, at least until the entire genesis is rewritten to match bond denom
		MinGasPrices:      fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
		AccountTokens:     sdk.TokensFromConsensusPower(1000000, sdk.DefaultPowerReduction),
		StakingTokens:     sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:      sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		PruningStrategy:   storetypes.PruningOptionNothing,
		CleanupDir:        true,
		SigningAlgo:       string(hd.Secp256k1Type),
		KeyringOptions:    []keyring.Option{},
	}
}

func CleanUp(n *testnet.Network, t *testing.T) {
	t.Log("teardown waiting for next block")
	//nolint:errcheck // The test shouldn't fail because cleanup was a problem. So ignoring any error from this.
	n.WaitForNextBlock()
	t.Log("teardown cleaning up testnet")
	n.Cleanup()
	// Give things a chance to finish closing up. Hopefully will prevent things like address collisions. 100ms chosen randomly.
	time.Sleep(100 * time.Millisecond)
	t.Log("teardown done")
}
