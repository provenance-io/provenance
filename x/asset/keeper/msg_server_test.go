package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/asset/keeper"
	"github.com/provenance-io/provenance/x/asset/types"
	ledgertypes "github.com/provenance-io/provenance/x/ledger/types"
)

type MsgServerTestSuite struct {
	suite.Suite
	app *app.App
	ctx sdk.Context
	user1Addr sdk.AccAddress
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
	priv := secp256k1.GenPrivKey()
	s.user1Addr = sdk.AccAddress(priv.PubKey().Address())
}

func (s *MsgServerTestSuite) TestAddAssetClass() {
	msgServer := keeper.NewMsgServerImpl(s.app.AssetKeeper)
	ledgerClass := ledgertypes.LedgerClass{
		LedgerClassId: "asset-class-1",
		AssetClassId:  "asset-class-1",
		Denom:         "stake",
		MaintainerAddress: s.user1Addr.String(),
	}
	err := s.app.LedgerKeeper.CreateLedgerClass(s.ctx, s.user1Addr, ledgerClass)
	s.Require().NoError(err)
	msg := &types.MsgAddAssetClass{
		AssetClass: &types.AssetClass{
			Id: "asset-class-1",
			Name: "AssetClass1",
		},
		LedgerClass: "asset-class-1",
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAssetClass(s.ctx, msg)
	s.Require().NoError(err)
}

func (s *MsgServerTestSuite) TestAddAsset() {
	msgServer := keeper.NewMsgServerImpl(s.app.AssetKeeper)
	
	// First create an asset class
	ledgerClass := ledgertypes.LedgerClass{
		LedgerClassId: "asset-class-2",
		AssetClassId:  "asset-class-2",
		Denom:         "stake",
		MaintainerAddress: s.user1Addr.String(),
	}
	err := s.app.LedgerKeeper.CreateLedgerClass(s.ctx, s.user1Addr, ledgerClass)
	s.Require().NoError(err)
	
	// Add a default status type to the ledger class
	statusType := ledgertypes.LedgerClassStatusType{
		Id:          1,
		Code:        "ACTIVE",
		Description: "Active",
	}
	err = s.app.LedgerKeeper.AddClassStatusType(s.ctx, s.user1Addr, "asset-class-2", statusType)
	s.Require().NoError(err)
	
	assetClassMsg := &types.MsgAddAssetClass{
		AssetClass: &types.AssetClass{
			Id: "asset-class-2",
			Name: "AssetClass2",
			Data: `{
				"$schema": "http://json-schema.org/draft-07/schema#",
				"type": "object",
				"properties": {
					"name": {
						"type": "string",
						"description": "The name of the asset"
					},
					"description": {
						"type": "string",
						"description": "A description of the asset"
					}
				},
				"required": ["name", "description"]
			}`,
		},
		LedgerClass: "asset-class-2",
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAssetClass(s.ctx, assetClassMsg)
	s.Require().NoError(err)
	
	// Now add an asset to the class
	msg := &types.MsgAddAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-2",
			Id:      "asset-1",
			Uri:     "https://example.com/asset1",
			UriHash: "abc123",
			Data:    `{"name": "Test Asset", "description": "A test asset"}`,
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAsset(s.ctx, msg)
	s.Require().NoError(err)
}

func (s *MsgServerTestSuite) TestCreatePool() {
	msgServer := keeper.NewMsgServerImpl(s.app.AssetKeeper)
	
	// Create an asset class
	ledgerClass := ledgertypes.LedgerClass{
		LedgerClassId: "asset-class-3",
		AssetClassId:  "asset-class-3",
		Denom:         "stake",
		MaintainerAddress: s.user1Addr.String(),
	}
	err := s.app.LedgerKeeper.CreateLedgerClass(s.ctx, s.user1Addr, ledgerClass)
	s.Require().NoError(err)
	
	// Add a default status type to the ledger class
	statusType := ledgertypes.LedgerClassStatusType{
		Id:          1,
		Code:        "ACTIVE",
		Description: "Active",
	}
	err = s.app.LedgerKeeper.AddClassStatusType(s.ctx, s.user1Addr, "asset-class-3", statusType)
	s.Require().NoError(err)
	
	assetClassMsg := &types.MsgAddAssetClass{
		AssetClass: &types.AssetClass{
			Id: "asset-class-3",
			Name: "AssetClass3",
		},
		LedgerClass: "asset-class-3",
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAssetClass(s.ctx, assetClassMsg)
	s.Require().NoError(err)
	
	// Create assets to add to the pool
	asset1Msg := &types.MsgAddAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-3",
			Id:      "asset-pool-1",
			Uri:     "https://example.com/asset1",
			UriHash: "abc123",
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAsset(s.ctx, asset1Msg)
	s.Require().NoError(err)
	
	asset2Msg := &types.MsgAddAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-3",
			Id:      "asset-pool-2",
			Uri:     "https://example.com/asset2",
			UriHash: "def456",
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.AddAsset(s.ctx, asset2Msg)
	s.Require().NoError(err)
	
	// Create a pool with these assets
	msg := &types.MsgCreatePool{
		Pool: &sdk.Coin{
			Denom:  "pooltoken",
			Amount: sdkmath.NewInt(1000),
		},
		Nfts: []*types.Nft{
			{
				ClassId: "asset-class-3",
				Id:      "asset-pool-1",
			},
			{
				ClassId: "asset-class-3",
				Id:      "asset-pool-2",
			},
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreatePool(s.ctx, msg)
	s.Require().NoError(err)
}
