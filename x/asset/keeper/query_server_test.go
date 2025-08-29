package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/asset/keeper"
	"github.com/provenance-io/provenance/x/asset/types"
)

type QueryServerTestSuite struct {
	suite.Suite
	app       *app.App
	ctx       sdk.Context
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

	// Create asset class 1
	assetClassMsg1 := &types.MsgCreateAssetClass{
		AssetClass: &types.AssetClass{
			Id:          "asset-class-1",
			Name:        "AssetClass1",
			Symbol:      "AC1",
			Description: "First asset class",
			Uri:         "https://example.com/class1",
			UriHash:     "hash1",
			Data:        `{"schema": "test"}`,
		},
		Signer: s.user1Addr.String(),
	}
	_, err := msgServer.CreateAssetClass(s.ctx, assetClassMsg1)
	s.Require().NoError(err)

	// Create asset class 2
	assetClassMsg2 := &types.MsgCreateAssetClass{
		AssetClass: &types.AssetClass{
			Id:          "asset-class-2",
			Name:        "AssetClass2",
			Symbol:      "AC2",
			Description: "Second asset class",
			Uri:         "https://example.com/class2",
			UriHash:     "hash2",
		},
		Signer: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAssetClass(s.ctx, assetClassMsg2)
	s.Require().NoError(err)

	// Create assets for user1
	asset1Msg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-1",
			Id:      "asset-1",
			Uri:     "https://example.com/asset1",
			UriHash: "asset1hash",
			Data:    `{"name": "Asset 1", "value": 100}`,
		},
		Signer: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, asset1Msg)
	s.Require().NoError(err)

	asset2Msg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-1",
			Id:      "asset-2",
			Uri:     "https://example.com/asset2",
			UriHash: "asset2hash",
			Data:    `{"name": "Asset 2", "value": 200}`,
		},
		Signer: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, asset2Msg)
	s.Require().NoError(err)

	asset3Msg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-2",
			Id:      "asset-3",
			Uri:     "https://example.com/asset3",
			UriHash: "asset3hash",
		},
		Signer: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, asset3Msg)
	s.Require().NoError(err)

	// Create assets for user2
	asset4Msg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-1",
			Id:      "asset-4",
			Uri:     "https://example.com/asset4",
			UriHash: "asset4hash",
		},
		Signer: s.user2Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, asset4Msg)
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
			req := &types.QueryAssetsRequest{
				Owner: tc.address,
			}

			resp, err := queryServer.Assets(sdk.WrapSDKContext(s.ctx), req)

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

	req := &types.QueryAssetClassesRequest{}
	resp, err := queryServer.AssetClasses(sdk.WrapSDKContext(s.ctx), req)

	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Classes, 2)

	// Verify asset class details
	foundClass1 := false
	foundClass2 := false

	for _, assetClass := range resp.Classes {
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
			req := &types.QueryAssetClassRequest{
				Id: tc.classId,
			}

			resp, err := queryServer.AssetClass(sdk.WrapSDKContext(s.ctx), req)

			if tc.expectError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "not found class")
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().NotNil(resp.Class)

			assetClass := resp.Class
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

	req := &types.QueryAssetsRequest{
		Owner: s.user1Addr.String(),
	}

	resp, err := queryServer.Assets(sdk.WrapSDKContext(s.ctx), req)

	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Assets, 0)
}

func (s *QueryServerTestSuite) TestListAssetClassesEmptyState() {
	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)

	req := &types.QueryAssetClassesRequest{}
	resp, err := queryServer.AssetClasses(sdk.WrapSDKContext(s.ctx), req)

	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Classes, 0)
}

func (s *QueryServerTestSuite) TestListAssetsWithPagination() {
	s.setupAssetClassAndAssets()

	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)

	tests := []struct {
		name          string
		address       string
		pagination    *query.PageRequest
		expectedCount int
		expectedTotal uint64
		expectError   bool
	}{
		{
			name:          "list assets with limit 2",
			address:       s.user1Addr.String(),
			pagination:    &query.PageRequest{Limit: 2},
			expectedCount: 2,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "list assets with offset 1 and limit 2",
			address:       s.user1Addr.String(),
			pagination:    &query.PageRequest{Offset: 1, Limit: 2},
			expectedCount: 2,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "list assets with offset 2 and limit 5",
			address:       s.user1Addr.String(),
			pagination:    &query.PageRequest{Offset: 2, Limit: 5},
			expectedCount: 1,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "list assets with offset beyond total count",
			address:       s.user1Addr.String(),
			pagination:    &query.PageRequest{Offset: 10, Limit: 5},
			expectedCount: 0,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "list assets with large limit",
			address:       s.user1Addr.String(),
			pagination:    &query.PageRequest{Limit: 100},
			expectedCount: 3,
			expectedTotal: 3,
			expectError:   false,
		},
		{
			name:          "list assets for user2 with pagination",
			address:       s.user2Addr.String(),
			pagination:    &query.PageRequest{Limit: 5},
			expectedCount: 1,
			expectedTotal: 1,
			expectError:   false,
		},
		{
			name:          "list assets without pagination",
			address:       s.user1Addr.String(),
			pagination:    nil,
			expectedCount: 3,
			expectedTotal: 3,
			expectError:   false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			req := &types.QueryAssetsRequest{
				Owner:      tc.address,
				Pagination: tc.pagination,
			}
			if req.Pagination == nil {
				req.Pagination = &query.PageRequest{CountTotal: true}
			} else {
				req.Pagination.CountTotal = true
			}

			resp, err := queryServer.Assets(sdk.WrapSDKContext(s.ctx), req)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Len(resp.Assets, tc.expectedCount)
			s.Require().NotNil(resp.Pagination)
			s.Require().Equal(tc.expectedTotal, resp.Pagination.Total)

			// Verify that all returned assets belong to the requested address
			for _, asset := range resp.Assets {
				owner := s.app.NFTKeeper.GetOwner(s.ctx, asset.ClassId, asset.Id)
				s.Require().Equal(tc.address, owner.String())
			}
		})
	}
}

func (s *QueryServerTestSuite) TestListAssetClassesWithPagination() {
	s.setupAssetClassAndAssets()

	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)

	tests := []struct {
		name          string
		pagination    *query.PageRequest
		expectedCount int
		expectedTotal uint64
		expectError   bool
	}{
		{
			name:          "list asset classes with limit 1",
			pagination:    &query.PageRequest{Limit: 1},
			expectedCount: 1,
			expectedTotal: 2,
			expectError:   false,
		},
		{
			name:          "list asset classes with offset 1 and limit 1",
			pagination:    &query.PageRequest{Offset: 1, Limit: 1},
			expectedCount: 1,
			expectedTotal: 2,
			expectError:   false,
		},
		{
			name:          "list asset classes with offset 0 and limit 5",
			pagination:    &query.PageRequest{Limit: 5, CountTotal: true},
			expectedCount: 2,
			expectedTotal: 2,
			expectError:   false,
		},
		{
			name:          "list asset classes with offset beyond total count",
			pagination:    &query.PageRequest{Offset: 10, Limit: 5},
			expectedCount: 0,
			expectedTotal: 2,
			expectError:   false,
		},
		{
			name:          "list asset classes with large limit",
			pagination:    &query.PageRequest{Limit: 100},
			expectedCount: 2,
			expectedTotal: 2,
			expectError:   false,
		},
		{
			name:          "list asset classes without pagination",
			pagination:    nil,
			expectedCount: 2,
			expectedTotal: 2,
			expectError:   false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			req := &types.QueryAssetClassesRequest{
				Pagination: tc.pagination,
			}
			if req.Pagination == nil {
				req.Pagination = &query.PageRequest{CountTotal: true}
			} else {
				req.Pagination.CountTotal = true
			}

			resp, err := queryServer.AssetClasses(sdk.WrapSDKContext(s.ctx), req)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Len(resp.Classes, tc.expectedCount)
			s.Require().NotNil(resp.Pagination)
			s.Require().Equal(tc.expectedTotal, resp.Pagination.Total)

			// Verify that returned asset classes are valid
			for _, assetClass := range resp.Classes {
				s.Require().NotEmpty(assetClass.Id)
				s.Require().NotEmpty(assetClass.Name)
				s.Require().NotEmpty(assetClass.Symbol)
			}
		})
	}
}

func (s *QueryServerTestSuite) TestListAssetsPaginationEdgeCases() {
	s.setupAssetClassAndAssets()

	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)

	tests := []struct {
		name          string
		address       string
		pagination    *query.PageRequest
		expectedCount int
		expectError   bool
	}{
		{
			name:          "list assets with zero limit",
			address:       s.user1Addr.String(),
			pagination:    &query.PageRequest{Limit: 0, CountTotal: true},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "list assets with zero offset",
			address:       s.user1Addr.String(),
			pagination:    &query.PageRequest{Limit: 2, CountTotal: true},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "list assets with offset equal to total count",
			address:       s.user1Addr.String(),
			pagination:    &query.PageRequest{Offset: 3, Limit: 5, CountTotal: true},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "list assets for empty address with pagination",
			address:       "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			pagination:    &query.PageRequest{Limit: 5, CountTotal: true},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			req := &types.QueryAssetsRequest{
				Owner:      tc.address,
				Pagination: tc.pagination,
			}

			resp, err := queryServer.Assets(sdk.WrapSDKContext(s.ctx), req)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Len(resp.Assets, tc.expectedCount)
			s.Require().NotNil(resp.Pagination)
		})
	}
}

func (s *QueryServerTestSuite) TestListAssetClassesPaginationEdgeCases() {
	s.setupAssetClassAndAssets()

	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)

	tests := []struct {
		name          string
		pagination    *query.PageRequest
		expectedCount int
		expectError   bool
	}{
		{
			name:          "list asset classes with zero limit",
			pagination:    &query.PageRequest{Limit: 0, CountTotal: true},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "list asset classes with zero offset",
			pagination:    &query.PageRequest{Limit: 1, CountTotal: true},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "list asset classes with offset equal to total count",
			pagination:    &query.PageRequest{Offset: 2, Limit: 5, CountTotal: true},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			req := &types.QueryAssetClassesRequest{
				Pagination: tc.pagination,
			}

			resp, err := queryServer.AssetClasses(sdk.WrapSDKContext(s.ctx), req)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Len(resp.Classes, tc.expectedCount)
			s.Require().NotNil(resp.Pagination)
		})
	}
}

func (s *QueryServerTestSuite) TestListAssetsPaginationConsistency() {
	s.setupAssetClassAndAssets()

	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)

	// Test that pagination returns consistent results
	req1 := &types.QueryAssetsRequest{
		Owner:      s.user1Addr.String(),
		Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
	}

	req2 := &types.QueryAssetsRequest{
		Owner:      s.user1Addr.String(),
		Pagination: &query.PageRequest{Offset: 1, Limit: 1, CountTotal: true},
	}

	req3 := &types.QueryAssetsRequest{
		Owner:      s.user1Addr.String(),
		Pagination: &query.PageRequest{Offset: 2, Limit: 1, CountTotal: true},
	}

	resp1, err := queryServer.Assets(sdk.WrapSDKContext(s.ctx), req1)
	s.Require().NoError(err)
	s.Require().Len(resp1.Assets, 1)

	resp2, err := queryServer.Assets(sdk.WrapSDKContext(s.ctx), req2)
	s.Require().NoError(err)
	s.Require().Len(resp2.Assets, 1)

	resp3, err := queryServer.Assets(sdk.WrapSDKContext(s.ctx), req3)
	s.Require().NoError(err)
	s.Require().Len(resp3.Assets, 1)

	// Verify that all three responses have different assets
	assetIds := make(map[string]bool)
	assetIds[resp1.Assets[0].Id] = true
	assetIds[resp2.Assets[0].Id] = true
	assetIds[resp3.Assets[0].Id] = true

	s.Require().Len(assetIds, 3, "All three paginated responses should return different assets")

	// Verify total counts are consistent
	s.Require().Equal(uint64(3), resp1.Pagination.Total)
	s.Require().Equal(uint64(3), resp2.Pagination.Total)
	s.Require().Equal(uint64(3), resp3.Pagination.Total)
}

func (s *QueryServerTestSuite) TestListAssetClassesPaginationConsistency() {
	s.setupAssetClassAndAssets()

	queryServer := keeper.NewQueryServerImpl(s.app.AssetKeeper)

	// Test that pagination returns consistent results
	req1 := &types.QueryAssetClassesRequest{
		Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
	}

	req2 := &types.QueryAssetClassesRequest{
		Pagination: &query.PageRequest{Offset: 1, Limit: 1, CountTotal: true},
	}

	resp1, err := queryServer.AssetClasses(sdk.WrapSDKContext(s.ctx), req1)
	s.Require().NoError(err)
	s.Require().Len(resp1.Classes, 1)

	resp2, err := queryServer.AssetClasses(sdk.WrapSDKContext(s.ctx), req2)
	s.Require().NoError(err)
	s.Require().Len(resp2.Classes, 1)

	// Verify that the two responses have different asset classes
	s.Require().NotEqual(resp1.Classes[0].Id, resp2.Classes[0].Id, "Pagination should return different asset classes")

	// Verify total counts are consistent
	s.Require().Equal(uint64(2), resp1.Pagination.Total)
	s.Require().Equal(uint64(2), resp2.Pagination.Total)
}
