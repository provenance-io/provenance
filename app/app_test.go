package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	dbm "github.com/tendermint/tm-db"

	sdksim "github.com/cosmos/cosmos-sdk/simapp"
)

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	opts := SetupOptions{
		Logger:             log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
		DB:                 dbm.NewMemDB(),
		InvCheckPeriod:     0,
		HomePath:           t.TempDir(),
		SkipUpgradeHeights: map[int64]bool{},
		EncConfig:          MakeEncodingConfig(),
		AppOpts:            sdksim.EmptyAppOptions{},
	}
	app := NewAppWithCustomOptions(t, false, opts)

	for acc := range maccPerms {
		require.True(
			t,
			app.BankKeeper.BlockedAddr(app.AccountKeeper.GetModuleAddress(acc)),
			"ensure that blocked addresses are properly set in bank keeper",
		)
	}

	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := New(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), opts.DB, nil, true,
		map[int64]bool{}, opts.HomePath, 0, opts.EncConfig, sdksim.EmptyAppOptions{})
	var err error
	require.NotPanics(t, func() {
		_, err = app2.ExportAppStateAndValidators(false, []string{})
	}, "exporting app state at current height")
	require.NoError(t, err, "ExportAppStateAndValidators at current height")

	require.NotPanics(t, func() {
		_, err = app2.ExportAppStateAndValidators(true, []string{})
	}, "exporting app state at zero height")
	require.NoError(t, err, "ExportAppStateAndValidators at zero height")
}

func TestGetMaccPerms(t *testing.T) {
	dup := GetMaccPerms()
	require.Equal(t, maccPerms, dup, "duplicated module account permissions differed from actual module account permissions")
}
