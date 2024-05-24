package testutil

import (
	"context"
	"errors"
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
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
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
	encCfg := tempApp.GetEncodingConfig()

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

// Cleanup runs the standard cleanup for our test networks.
func Cleanup(n *testnet.Network, t *testing.T) {
	t.Log("Cleanup: Tearing down test network")
	if n == nil {
		t.Log("Cleanup: Nothing to tear down. Done.")
		return
	}
	t.Log("Cleanup: Waiting for next block.")
	err := WaitForNextBlock(n)
	if err != nil {
		t.Logf("Cleanup: Error (ignored) waiting for next block: %v", err)
	}
	t.Log("Cleanup: Cleaning up testnet.")
	n.Cleanup()
	// Give things a chance to finish closing up. Hopefully will prevent things like address collisions. 100ms chosen randomly.
	time.Sleep(100 * time.Millisecond)
	t.Log("Cleanup: Done.")
}

// queryCurrentHeight executes a query to get the current height in a separate process.
// Returns a channel that will receive the height.
func queryCurrentHeight(queryClient cmtservice.ServiceClient) <-chan int64 {
	rv := make(chan int64, 1)
	go func() {
		res, err := queryClient.GetLatestBlock(context.Background(), &cmtservice.GetLatestBlockRequest{})
		curHeight := int64(0)
		if err == nil && res != nil {
			curHeight = res.SdkBlock.Header.Height
		}
		rv <- curHeight
	}()
	return rv
}

// LatestHeight returns the latest height of the network or an error if the
// query fails or no validators exist.
//
// This is similar to Network.LatestHeight() except that this
// one doesn't wait for a first tick before trying to get the current height
// (so it can return earlier). It also checks every 500ms instead of every 1s.
func LatestHeight(n *testnet.Network) (int64, error) {
	return WaitForHeightWithTimeout(n, 1, 5*time.Second)
}

// WaitForHeight performs a blocking check where it waits for a block to be
// committed after a given block. If that height is not reached within a timeout,
// an error is returned. Regardless, the latest height queried is returned.
//
// This is similar to Network.WaitForHeight(h) except that this
// one doesn't wait for a first tick before trying to get the current height
// (so it can return earlier). It also checks every 500ms instead of every 1s.
func WaitForHeight(n *testnet.Network, h int64) (int64, error) {
	return WaitForHeightWithTimeout(n, h, 10*time.Second)
}

// WaitForHeightWithTimeout is the same as WaitForHeight except the caller can
// provide a custom timeout.
//
// This is similar to Network.WaitForHeightWithTimeout(h, t) except that this
// one doesn't wait for a first tick before trying to get the current height
// (so it can return earlier). It also checks every 500ms instead of every 1s.
func WaitForHeightWithTimeout(n *testnet.Network, h int64, t time.Duration) (int64, error) {
	if len(n.Validators) == 0 {
		return 0, errors.New("no validators available")
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(t)
	defer timeout.Stop()

	queryClient := cmtservice.NewServiceClient(n.Validators[0].ClientCtx)
	qch := queryCurrentHeight(queryClient)
	getting := true
	var latestHeight int64

	for {
		select {
		case latestHeight = <-qch:
			getting = false
			if latestHeight >= h {
				return latestHeight, nil
			}
		case <-ticker.C:
			if !getting {
				qch = queryCurrentHeight(queryClient)
				getting = true
			}
		case <-timeout.C:
			return latestHeight, errors.New("timeout exceeded waiting for block")
		}
	}
}

// WaitForNextBlock waits for the next block to be committed, returning an error
// upon failure.
//
// This is similar to Network.WaitForNextBlock(h, t) except that this
// one doesn't wait for a first tick before trying to get the current height
// (so it can return earlier). It also checks every 500ms instead of every 1s.
func WaitForNextBlock(n *testnet.Network) error {
	return WaitForNBlocks(n, 1)
}

// WaitForNBlocks waits for the next count blocks to be committed, returning an error upon failure.
func WaitForNBlocks(n *testnet.Network, count int) error {
	if len(n.Validators) == 0 {
		return errors.New("no validators available")
	}

	queryClient := cmtservice.NewServiceClient(n.Validators[0].ClientCtx)
	qch := queryCurrentHeight(queryClient)
	getting := true
	var endHeight int64

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case curHeight := <-qch:
			getting = false
			switch {
			case endHeight == 0:
				if curHeight > 0 {
					endHeight = curHeight + int64(count)
				}
			case curHeight >= endHeight:
				return nil
			}
		case <-ticker.C:
			if !getting {
				qch = queryCurrentHeight(queryClient)
			}
		case <-timeout.C:
			return errors.New("timeout exceeded waiting for next block")
		}
	}
}
