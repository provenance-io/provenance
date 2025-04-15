package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/provenance-io/provenance/x/exchange"

	"github.com/provenance-io/provenance/app"
)

func Test_TestnetCmd(t *testing.T) {
	// We need to disable address caching in here because this test goes a couple
	// address conversions using a different HRP from most of the other tests in
	// this package. With address caching on, those tests can randomly fail if
	// they end up running after this test.
	origCache := sdk.IsAddrCacheEnabled()
	defer sdk.SetAddrCacheEnabled(origCache)
	sdk.SetAddrCacheEnabled(false)
	home := t.TempDir()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	tempApp := app.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(home))
	encodingConfig := tempApp.GetEncodingConfig()

	require.NoError(t, err)

	err = genutiltest.ExecInitCmd(tempApp.BasicModuleManager, home, encodingConfig.Marshaler)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithHomeDir(home).
		WithTxConfig(encodingConfig.TxConfig)

	ctx := context.Background()
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	cmd := testnetCmd(tempApp.BasicModuleManager, banktypes.GenesisBalancesIterator{})
	cmd.SetArgs([]string{fmt.Sprintf("--%s=test", flags.FlagKeyringBackend), fmt.Sprintf("--output-dir=%s", home)})
	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	genFile := cfg.GenesisFile()
	appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
	require.NoError(t, err)

	bankGenState := banktypes.GetGenesisStateFromAppState(encodingConfig.Marshaler, appState)
	require.NotEmpty(t, bankGenState.Supply.String())

	var exGenState exchange.GenesisState
	err = clientCtx.Codec.UnmarshalJSON(appState[exchange.ModuleName], &exGenState)
	if assert.NoError(t, err, "UnmarshalJSON exchange genesis state") {
		assert.Len(t, exGenState.Markets, 1, "markets in exchange genesis state")
	}
}
