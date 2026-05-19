package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestSnapshotRestorePreservesWasmBytecode(t *testing.T) {
	src, creator := newSnapshotApp(t, true)

	ctx := src.NewUncachedContext(false, tmproto.Header{
		ChainID: "testchain-snapshot",
		Height:  src.LastBlockHeight() + 1,
		Time:    time.Now(),
	})
	wasmCode, err := os.ReadFile(filepath.Join("sim_contracts", "tutorial.wasm"))
	require.NoError(t, err, "ReadFile: sim_contracts / tutorial.wasm")

	codeID, _, err := src.ContractKeeper.Create(ctx, creator, wasmCode, nil)
	require.NoError(t, err, "ContractKeeper.Create")

	src.CommitMultiStore().Commit()

	snapshotHeight := uint64(src.LastBlockHeight())
	snapshot, err := src.SnapshotManager().Create(snapshotHeight)
	require.NoError(t, err, "src.SnapshotManager().Create")

	dst, _ := newSnapshotApp(t, false)
	require.NoError(t, dst.SnapshotManager().Restore(*snapshot), "dst.SnapshotManager().Restore")
	for chunk := uint32(0); chunk < snapshot.Chunks; chunk++ {
		chunkBz, err := src.SnapshotManager().LoadChunk(snapshot.Height, snapshot.Format, chunk)
		require.NoError(t, err, "src.SnapshotManager.LoadChunk %d", chunk)
		done, err := dst.SnapshotManager().RestoreChunk(chunkBz)
		require.NoError(t, err, "dst.SnapshotManager.RestoreChunk %d", chunk)
		if done {
			break
		}
	}

	restoredCtx := dst.NewUncachedContext(false, tmproto.Header{
		ChainID: "testchain-snapshot",
		Height:  int64(snapshot.Height),
		Time:    time.Now(),
	})
	require.NotNil(t, dst.WasmKeeper.GetCodeInfo(restoredCtx, codeID), "dst.WasmKeeper.GetCodeInfo for original contract")

	restoredCode, err := dst.WasmKeeper.GetByteCode(restoredCtx, codeID)
	require.NoError(t, err, "dst.WasmKeeper.GetByteCode for original contract")
	require.Equal(t, wasmCode, restoredCode, "wasm byte code")
}

func newSnapshotApp(t *testing.T, initChain bool) (*App, sdk.AccAddress) {
	t.Helper()

	home := t.TempDir()
	snapshotDir := filepath.Join(home, "data", "snapshots")
	snapshotStore, err := snapshots.NewStore(dbm.NewMemDB(), snapshotDir)
	require.NoError(t, err, "snapshots.NewStore")

	appOpts := simtestutil.AppOptionsMap{
		flags.FlagHome:            home,
		server.FlagInvCheckPeriod: uint(0),
	}

	app := New(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		appOpts,
		baseapp.SetChainID("wf016"),
		baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(1, 2)),
	)

	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	if !initChain {
		return app, acc.GetAddress()
	}

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err, "privVal.GetPubKey()")
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000_000_000)),
	}

	genesisState := genesisStateWithValSet(t, app, app.DefaultGenesis(), valSet, []authtypes.GenesisAccount{acc}, balance)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err, "json.MarshalIndent genesisState")

	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: DefaultConsensusParams,
		AppStateBytes:   stateBytes,
		ChainId:         "wf016",
	})
	require.NoError(t, err, "app.InitChain")

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Hash:               app.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
	})
	require.NoError(t, err, "app.FinalizeBlock")
	_, err = app.Commit()
	require.NoError(t, err, "app.Commit")

	return app, acc.GetAddress()
}