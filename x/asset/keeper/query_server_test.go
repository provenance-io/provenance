package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/asset/keeper"
	"github.com/provenance-io/provenance/x/asset/types"
	ledgertypes "github.com/provenance-io/provenance/x/ledger/types"
)

type QueryServerTestSuite struct {
	suite.Suite
	app *app.App
	ctx sdk.Context
	user1Addr sdk.AccAddress
	user2Addr sdk.AccAddress
}

func TestQueryServerTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (s *QueryServerTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
	
	// Create two test users
	priv1 := secp256k1.GenPrivKey()
	s.user1Addr = sdk.AccAddress(priv1.PubKey().Address())
	
	priv2 := secp256k1.GenPrivKey()
	s.user2Addr = sdk.AccAddress(priv2.PubKey().Address())
}

func (s *QueryServerTestSuite) setupAssetClassAndAssets() {
	msgServer := keeper.NewMsgServerImpl(s.app.AssetKeeper)
	
	// Create ledger class for asset class 1
	ledgerClass1 := ledgertypes.LedgerClass{
		LedgerClassId:     "asset-class-1",
		AssetClassId:      "asset-class-1",
		Denom:             "stake",
		MaintainerAddress: s.user1Addr.String(),
	}
	err := s.app.LedgerKeeper.CreateLedgerClass(s.ctx, s.user1Addr, ledgerClass1)
	s.Require().NoError(err)
	
	// Add status type to ledger class
	statusType := ledgertypes.LedgerClassStatusType{
		Id:          1,
		Code:        "ACTIVE",
		Description: "Active",
	}
	err = s.app.LedgerKeeper.AddClassStatusType(s.ctx, s.user1Addr, "asset-class-1", statusType)
	s.Require().NoError(err)
	
	// Create asset class 1
	assetClassMsg1 := &types.MsgAddAssetClass{
		AssetClass: &types.AssetClass{
			Id:          "asset-class-1",
			Name:        "AssetClass1",
			Symbol:      "AC1",
			Description: "First asset class",
			Uri:         "https://example.com/class1",
			UriHash:     "hash1",
			Data:        `{"schema": "test"}`,
		},
		LedgerClass: "asset-class-1",
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAssetClass(s.ctx, assetClassMsg1)
	s.Require().NoError(err)
	
	// Create ledger class for asset class 2
	ledgerClass2 := ledgertypes.LedgerClass{
		LedgerClassId:     "asset-class-2",
		AssetClassId:      "asset-class-2",
		Denom:             "stake",
		MaintainerAddress: s.user1Addr.String(),
	}
	err = s.app.LedgerKeeper.CreateLedgerClass(s.ctx, s.user1Addr, ledgerClass2)
	s.Require().NoError(err)
	
	err = s.app.LedgerKeeper.AddClassStatusType(s.ctx, s.user1Addr, "asset-class-2", statusType)
	s.Require().NoError(err)
	
	// Create asset class 2
	assetClassMsg2 := &types.MsgAddAssetClass{
		AssetClass: &types.AssetClass{
			Id:          "asset-class-2",
			Name:        "AssetClass2",
			Symbol:      "AC2",
			Description: "Second asset class",
			Uri:         "https://example.com/class2",
			UriHash:     "hash2",
		},
		LedgerClass: "asset-class-2",
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAssetClass(s.ctx, assetClassMsg2)
	s.Require().NoError(err)
	
	// Create assets for user1
	asset1Msg := &types.MsgAddAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-1",
			Id:      "asset-1",
			Uri:     "https://example.com/asset1",
			UriHash: "asset1hash",
			Data:    `{"name": "Asset 1", "value": 100}`,
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAsset(s.ctx, asset1Msg)
	s.Require().NoError(err)
	
	asset2Msg := &types.MsgAddAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-1",
			Id:      "asset-2",
			Uri:     "https://example.com/asset2",
			UriHash: "asset2hash",
			Data:    `{"name": "Asset 2", "value": 200}`,
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAsset(s.ctx, asset2Msg)
	s.Require().NoError(err)
	
	asset3Msg := &types.MsgAddAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-2",
			Id:      "asset-3",
			Uri:     "https://example.com/asset3",
			UriHash: "asset3hash",
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAsset(s.ctx, asset3Msg)
	s.Require().NoError(err)
	
	// Create assets for user2
	asset4Msg := &types.MsgAddAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-1",
			Id:      "asset-4",
			Uri:     "https://example.com/asset4",
			UriHash: "asset4hash",
		},
		FromAddress: s.user2Addr.String(),
	}
	_, err = msgServer.AddAsset(s.ctx, asset4Msg)
	s.Require().NoError(err)
}

func (s *QueryServerTestSuite) TestListAssets() {
	s.setupAssetClassAndAssets()
	
	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)
	
	tests := []struct {
		name          string
		address       string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "list assets for user1",
			address:       s.user1Addr.String(),
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "list assets for user2",
			address:       s.user2Addr.String(),
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "list assets for non-existent user",
			address:       "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:        "invalid address",
			address:     "invalid-address",
			expectError: true,
		},
	}
	
	for _, tc := range tests {
		s.Run(tc.name, func() {
			req := &types.QueryListAssets{
				Address: tc.address,
			}
			
			resp, err := queryServer.ListAssets(sdk.WrapSDKContext(s.ctx), req)
			
			if tc.expectError {
				s.Require().Error(err)
				return
			}
			
			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Len(resp.Assets, tc.expectedCount)
			
			// Verify asset details for user1
			if tc.address == s.user1Addr.String() {
				s.Require().Len(resp.Assets, 3)
				
				// Check that all assets belong to user1
				for _, asset := range resp.Assets {
					// Verify the asset exists in the NFT keeper and belongs to user1
					owner := s.app.NFTKeeper.GetOwner(s.ctx, asset.ClassId, asset.Id)
					s.Require().Equal(s.user1Addr.String(), owner.String())
				}
				
				// Verify specific asset details
				foundAsset1 := false
				foundAsset2 := false
				foundAsset3 := false
				
				for _, asset := range resp.Assets {
					if asset.Id == "asset-1" {
						foundAsset1 = true
						s.Require().Equal("asset-class-1", asset.ClassId)
						s.Require().Equal("https://example.com/asset1", asset.Uri)
						s.Require().Equal("asset1hash", asset.UriHash)
						s.Require().Equal(`{"name": "Asset 1", "value": 100}`, asset.Data)
					} else if asset.Id == "asset-2" {
						foundAsset2 = true
						s.Require().Equal("asset-class-1", asset.ClassId)
						s.Require().Equal("https://example.com/asset2", asset.Uri)
						s.Require().Equal("asset2hash", asset.UriHash)
						s.Require().Equal(`{"name": "Asset 2", "value": 200}`, asset.Data)
					} else if asset.Id == "asset-3" {
						foundAsset3 = true
						s.Require().Equal("asset-class-2", asset.ClassId)
						s.Require().Equal("https://example.com/asset3", asset.Uri)
						s.Require().Equal("asset3hash", asset.UriHash)
						s.Require().Equal("", asset.Data) // No data for this asset
					}
				}
				
				s.Require().True(foundAsset1, "Asset 1 not found")
				s.Require().True(foundAsset2, "Asset 2 not found")
				s.Require().True(foundAsset3, "Asset 3 not found")
			}
			
			// Verify asset details for user2
			if tc.address == s.user2Addr.String() {
				s.Require().Len(resp.Assets, 1)
				asset := resp.Assets[0]
				s.Require().Equal("asset-4", asset.Id)
				s.Require().Equal("asset-class-1", asset.ClassId)
				s.Require().Equal("https://example.com/asset4", asset.Uri)
				s.Require().Equal("asset4hash", asset.UriHash)
			}
		})
	}
}

func (s *QueryServerTestSuite) TestListAssetClasses() {
	s.setupAssetClassAndAssets()
	
	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)
	
	req := &types.QueryListAssetClasses{}
	resp, err := queryServer.ListAssetClasses(sdk.WrapSDKContext(s.ctx), req)
	
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.AssetClasses, 2)
	
	// Verify asset class details
	foundClass1 := false
	foundClass2 := false
	
	for _, assetClass := range resp.AssetClasses {
		if assetClass.Id == "asset-class-1" {
			foundClass1 = true
			s.Require().Equal("AssetClass1", assetClass.Name)
			s.Require().Equal("AC1", assetClass.Symbol)
			s.Require().Equal("First asset class", assetClass.Description)
			s.Require().Equal("https://example.com/class1", assetClass.Uri)
			s.Require().Equal("hash1", assetClass.UriHash)
			s.Require().Equal(`{"schema": "test"}`, assetClass.Data)
		} else if assetClass.Id == "asset-class-2" {
			foundClass2 = true
			s.Require().Equal("AssetClass2", assetClass.Name)
			s.Require().Equal("AC2", assetClass.Symbol)
			s.Require().Equal("Second asset class", assetClass.Description)
			s.Require().Equal("https://example.com/class2", assetClass.Uri)
			s.Require().Equal("hash2", assetClass.UriHash)
			s.Require().Equal("", assetClass.Data) // No data for this class
		}
	}
	
	s.Require().True(foundClass1, "Asset class 1 not found")
	s.Require().True(foundClass2, "Asset class 2 not found")
}

func (s *QueryServerTestSuite) TestGetClass() {
	s.setupAssetClassAndAssets()
	
	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)
	
	tests := []struct {
		name        string
		classId     string
		expectError bool
	}{
		{
			name:        "get existing asset class 1",
			classId:     "asset-class-1",
			expectError: false,
		},
		{
			name:        "get existing asset class 2",
			classId:     "asset-class-2",
			expectError: false,
		},
		{
			name:        "get non-existent asset class",
			classId:     "non-existent-class",
			expectError: true,
		},
	}
	
	for _, tc := range tests {
		s.Run(tc.name, func() {
			req := &types.QueryGetClass{
				Id: tc.classId,
			}
			
			resp, err := queryServer.GetClass(sdk.WrapSDKContext(s.ctx), req)
			
			if tc.expectError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "class not found")
				return
			}
			
			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().NotNil(resp.AssetClass)
			
			assetClass := resp.AssetClass
			s.Require().Equal(tc.classId, assetClass.Id)
			
			if tc.classId == "asset-class-1" {
				s.Require().Equal("AssetClass1", assetClass.Name)
				s.Require().Equal("AC1", assetClass.Symbol)
				s.Require().Equal("First asset class", assetClass.Description)
				s.Require().Equal("https://example.com/class1", assetClass.Uri)
				s.Require().Equal("hash1", assetClass.UriHash)
				s.Require().Equal(`{"schema": "test"}`, assetClass.Data)
			} else if tc.classId == "asset-class-2" {
				s.Require().Equal("AssetClass2", assetClass.Name)
				s.Require().Equal("AC2", assetClass.Symbol)
				s.Require().Equal("Second asset class", assetClass.Description)
				s.Require().Equal("https://example.com/class2", assetClass.Uri)
				s.Require().Equal("hash2", assetClass.UriHash)
				s.Require().Equal("", assetClass.Data)
			}
		})
	}
}

func (s *QueryServerTestSuite) TestListAssetsEmptyState() {
	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)
	
	req := &types.QueryListAssets{
		Address: s.user1Addr.String(),
	}
	
	resp, err := queryServer.ListAssets(sdk.WrapSDKContext(s.ctx), req)
	
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Assets, 0)
}

func (s *QueryServerTestSuite) TestListAssetClassesEmptyState() {
	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)
	
	req := &types.QueryListAssetClasses{}
	resp, err := queryServer.ListAssetClasses(sdk.WrapSDKContext(s.ctx), req)
	
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.AssetClasses, 0)
}
