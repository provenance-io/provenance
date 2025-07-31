package keeper_test

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"

// 	sdkmath "cosmossdk.io/math"
// 	"cosmossdk.io/x/nft"
// 	"github.com/provenance-io/provenance/app"
// 	ledger "github.com/provenance-io/provenance/x/ledger/types"
// )

// func TestInitGenesis(t *testing.T) {
// 	testApp := app.Setup(t)
// 	ctx := testApp.BaseApp.NewContext(false)
// 	keeper := testApp.LedgerKeeper

// 	// Create test addresses
// 	testAddrs := app.AddTestAddrsIncremental(testApp, ctx, 1, sdkmath.NewInt(1000000))
// 	maintainerAddr := testAddrs[0]

// 	// Create a test NFT class and NFT first
// 	nftClass := nft.Class{
// 		Id: "test-asset-1",
// 	}
// 	testApp.NFTKeeper.SaveClass(ctx, nftClass)

// 	nft := nft.NFT{
// 		ClassId: nftClass.Id,
// 		Id:      "test-nft-1",
// 	}
// 	testApp.NFTKeeper.Mint(ctx, nft, maintainerAddr)

// 	// Create a test genesis state with simplified structure
// 	genState := &ledger.GenesisState{
// 		LedgerToEntries: []ledger.LedgerToEntries{
// 			{
// 				LedgerKey: &ledger.LedgerKey{
// 					AssetClassId: nftClass.Id,
// 					NftId:        nft.Id,
// 				},
// 				Ledger: &ledger.Ledger{
// 					Key: &ledger.LedgerKey{
// 						AssetClassId: nftClass.Id,
// 						NftId:        nft.Id,
// 					},
// 					LedgerClassId: "test-class-1",
// 					StatusTypeId:  1,
// 				},
// 				Entries: []*ledger.LedgerEntry{},
// 			},
// 		},
// 	}

// 	// Initialize genesis - this should not import any data for new chains
// 	keeper.InitGenesis(ctx, genState)

// 	// Verify that the ledger was NOT created (since InitGenesis is now minimal)
// 	ledgerKey := &ledger.LedgerKey{
// 		AssetClassId: nftClass.Id,
// 		NftId:        nft.Id,
// 	}
// 	ledger, err := keeper.GetLedger(ctx, ledgerKey)
// 	require.NoError(t, err)
// 	require.Nil(t, ledger)

// 	// Verify that ledger entries were NOT created
// 	entries, err := keeper.ListLedgerEntries(ctx, ledgerKey)
// 	require.NoError(t, err)
// 	require.Empty(t, entries)
// }

// func TestExportGenesis(t *testing.T) {
// 	testApp := app.Setup(t)
// 	ctx := testApp.BaseApp.NewContext(false)
// 	keeper := testApp.LedgerKeeper

// 	// Get the bond denom
// 	bondDenom, err := testApp.StakingKeeper.BondDenom(ctx)
// 	require.NoError(t, err)

// 	// Create some test data with a valid address
// 	// For testing purposes, we'll use a test address from the app setup
// 	testAddrs := app.AddTestAddrsIncremental(testApp, ctx, 1, sdkmath.NewInt(1000000))
// 	maintainerAddr := testAddrs[0]

// 	// Create a test NFT class and NFT first
// 	nftClass := nft.Class{
// 		Id: "test-asset-1",
// 	}
// 	testApp.NFTKeeper.SaveClass(ctx, nftClass)

// 	nft := nft.NFT{
// 		ClassId: nftClass.Id,
// 		Id:      "test-nft-1",
// 	}
// 	testApp.NFTKeeper.Mint(ctx, nft, maintainerAddr)

// 	ledgerClass := ledger.LedgerClass{
// 		LedgerClassId:     "test-class-1",
// 		AssetClassId:      nftClass.Id,
// 		Denom:             bondDenom,
// 		MaintainerAddress: maintainerAddr.String(),
// 	}

// 	// Create the ledger class
// 	err = keeper.AddLedgerClass(ctx, maintainerAddr, ledgerClass)
// 	require.NoError(t, err)

// 	// Add some status types
// 	statusType := ledger.LedgerClassStatusType{Id: 1, Code: "IN_REPAYMENT", Description: "In Repayment"}
// 	err = keeper.AddClassStatusType(ctx, maintainerAddr, ledgerClass.LedgerClassId, statusType)
// 	require.NoError(t, err)

// 	// Create a ledger
// 	ledgerKey := &ledger.LedgerKey{
// 		AssetClassId: nftClass.Id,
// 		NftId:        nft.Id,
// 	}
// 	ledgerObj := ledger.Ledger{
// 		Key:           ledgerKey,
// 		LedgerClassId: ledgerClass.LedgerClassId,
// 		StatusTypeId:  1,
// 	}
// 	err = keeper.AddLedger(ctx, maintainerAddr, ledgerObj)
// 	require.NoError(t, err)

// 	// Export genesis
// 	exportedState := keeper.ExportGenesis(ctx)
// 	require.NotNil(t, exportedState)
// 	require.Len(t, exportedState.LedgerToEntries, 1)
// 	require.Equal(t, ledgerKey.NftId, exportedState.LedgerToEntries[0].LedgerKey.NftId)
// 	require.Equal(t, ledgerKey.AssetClassId, exportedState.LedgerToEntries[0].LedgerKey.AssetClassId)
// }
