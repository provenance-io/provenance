package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
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

// Helper function to find an event by type
func (s *MsgServerTestSuite) findEventByType(events sdk.Events, eventType string) *sdk.Event {
	for _, event := range events {
		if event.Type == eventType {
			return &event
		}
	}
	return nil
}

// Helper function to find an attribute by key
func (s *MsgServerTestSuite) findAttributeByKey(event *sdk.Event, key string) *abci.EventAttribute {
	for _, attr := range event.Attributes {
		if attr.Key == key {
			return &attr
		}
	}
	return nil
}

func (s *MsgServerTestSuite) TestCreateAssetClass() {
	msgServer := keeper.NewMsgServerImpl(s.app.AssetKeeper)
	
	// Clear events before test
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	
	ledgerClass := ledgertypes.LedgerClass{
		LedgerClassId: "asset-class-1",
		AssetClassId:  "asset-class-1",
		Denom:         "stake",
		MaintainerAddress: s.user1Addr.String(),
	}
	err := s.app.LedgerKeeper.CreateLedgerClass(s.ctx, s.user1Addr, ledgerClass)
	s.Require().NoError(err)
	
	msg := &types.MsgCreateAssetClass{
		AssetClass: &types.AssetClass{
			Id: "asset-class-1",
			Name: "AssetClass1",
			Symbol: "AC1",
		},
		LedgerClass: "asset-class-1",
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAssetClass(s.ctx, msg)
	s.Require().NoError(err)
	
	// Verify event emission
	events := s.ctx.EventManager().Events()
	event := s.findEventByType(events, types.EventTypeAssetClassCreated)
	s.Require().NotNil(event, "EventAssetClassCreated should be emitted")
	
	// Verify event attributes
	assetClassIdAttr := s.findAttributeByKey(event, types.AttributeKeyAssetClassId)
	s.Require().NotNil(assetClassIdAttr, "asset_class_id attribute should be present")
	s.Require().Equal("asset-class-1", assetClassIdAttr.Value)
	
	assetNameAttr := s.findAttributeByKey(event, types.AttributeKeyAssetName)
	s.Require().NotNil(assetNameAttr, "asset_name attribute should be present")
	s.Require().Equal("AssetClass1", assetNameAttr.Value)
	
	assetSymbolAttr := s.findAttributeByKey(event, types.AttributeKeyAssetSymbol)
	s.Require().NotNil(assetSymbolAttr, "asset_symbol attribute should be present")
	s.Require().Equal("AC1", assetSymbolAttr.Value)
	
	ledgerClassAttr := s.findAttributeByKey(event, types.AttributeKeyLedgerClass)
	s.Require().NotNil(ledgerClassAttr, "ledger_class attribute should be present")
	s.Require().Equal("asset-class-1", ledgerClassAttr.Value)
	
	ownerAttr := s.findAttributeByKey(event, types.AttributeKeyOwner)
	s.Require().NotNil(ownerAttr, "owner attribute should be present")
	s.Require().Equal(s.user1Addr.String(), ownerAttr.Value)
}

func (s *MsgServerTestSuite) TestCreateAsset() {
	msgServer := keeper.NewMsgServerImpl(s.app.AssetKeeper)
	
	// Clear events before test
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	
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
	
	assetClassMsg := &types.MsgCreateAssetClass{
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
	_, err = msgServer.CreateAssetClass(s.ctx, assetClassMsg)
	s.Require().NoError(err)
	
	// Clear events before creating asset
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	
	// Now create an asset in the class
	msg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-2",
			Id:      "asset-1",
			Uri:     "https://example.com/asset1",
			UriHash: "abc123",
			Data:    `{"name": "Test Asset", "description": "A test asset"}`,
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, msg)
	s.Require().NoError(err)
	
	// Verify event emission
	events := s.ctx.EventManager().Events()
	event := s.findEventByType(events, types.EventTypeAssetCreated)
	s.Require().NotNil(event, "EventAssetCreated should be emitted")
	
	// Verify event attributes
	assetClassIdAttr := s.findAttributeByKey(event, types.AttributeKeyAssetClassId)
	s.Require().NotNil(assetClassIdAttr, "asset_class_id attribute should be present")
	s.Require().Equal("asset-class-2", assetClassIdAttr.Value)
	
	assetIdAttr := s.findAttributeByKey(event, types.AttributeKeyAssetId)
	s.Require().NotNil(assetIdAttr, "asset_id attribute should be present")
	s.Require().Equal("asset-1", assetIdAttr.Value)
	
	ownerAttr := s.findAttributeByKey(event, types.AttributeKeyOwner)
	s.Require().NotNil(ownerAttr, "owner attribute should be present")
	s.Require().Equal(s.user1Addr.String(), ownerAttr.Value)
}

func (s *MsgServerTestSuite) TestCreatePool() {
	msgServer := keeper.NewMsgServerImpl(s.app.AssetKeeper)
	
	// Clear events before test
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	
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
	
	assetClassMsg := &types.MsgCreateAssetClass{
		AssetClass: &types.AssetClass{
			Id: "asset-class-3",
			Name: "AssetClass3",
		},
		LedgerClass: "asset-class-3",
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAssetClass(s.ctx, assetClassMsg)
	s.Require().NoError(err)
	
	// Create assets to add to the pool
	asset1Msg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-3",
			Id:      "asset-pool-1",
			Uri:     "https://example.com/asset1",
			UriHash: "abc123",
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, asset1Msg)
	s.Require().NoError(err)
	
	asset2Msg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-3",
			Id:      "asset-pool-2",
			Uri:     "https://example.com/asset2",
			UriHash: "def456",
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, asset2Msg)
	s.Require().NoError(err)
	
	// Clear events before creating pool
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	
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
	
	// Verify event emission
	events := s.ctx.EventManager().Events()
	event := s.findEventByType(events, types.EventTypePoolCreated)
	s.Require().NotNil(event, "EventPoolCreated should be emitted")
	
	// Verify event attributes
	poolDenomAttr := s.findAttributeByKey(event, types.AttributeKeyPoolDenom)
	s.Require().NotNil(poolDenomAttr, "pool_denom attribute should be present")
	s.Require().Equal("pooltoken", poolDenomAttr.Value)
	
	poolAmountAttr := s.findAttributeByKey(event, types.AttributeKeyPoolAmount)
	s.Require().NotNil(poolAmountAttr, "pool_amount attribute should be present")
	s.Require().Equal("1000", poolAmountAttr.Value)
	
	nftCountAttr := s.findAttributeByKey(event, types.AttributeKeyNftCount)
	s.Require().NotNil(nftCountAttr, "nft_count attribute should be present")
	s.Require().Equal("2", nftCountAttr.Value)
	
	ownerAttr := s.findAttributeByKey(event, types.AttributeKeyOwner)
	s.Require().NotNil(ownerAttr, "owner attribute should be present")
	s.Require().Equal(s.user1Addr.String(), ownerAttr.Value)
}

func (s *MsgServerTestSuite) TestCreateTokenization() {
	msgServer := keeper.NewMsgServerImpl(s.app.AssetKeeper)
	
	// Clear events before test
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	
	// First create an asset class and asset for the NFT
	ledgerClass := ledgertypes.LedgerClass{
		LedgerClassId: "asset-class-token",
		AssetClassId:  "asset-class-token",
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
	err = s.app.LedgerKeeper.AddClassStatusType(s.ctx, s.user1Addr, "asset-class-token", statusType)
	s.Require().NoError(err)
	
	assetClassMsg := &types.MsgCreateAssetClass{
		AssetClass: &types.AssetClass{
			Id: "asset-class-token",
			Name: "AssetClassToken",
		},
		LedgerClass: "asset-class-token",
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAssetClass(s.ctx, assetClassMsg)
	s.Require().NoError(err)
	
	// Create an asset for the NFT
	assetMsg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-token",
			Id:      "asset-token-1",
			Uri:     "https://example.com/asset-token",
			UriHash: "abc123",
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, assetMsg)
	s.Require().NoError(err)
	
	// Clear events before creating tokenization
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	
	msg := &types.MsgCreateTokenization{
		Denom: sdk.NewCoin("tokenization", sdkmath.NewInt(500)),
		Nft: &types.Nft{
			ClassId: "asset-class-token",
			Id:      "asset-token-1",
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateTokenization(s.ctx, msg)
	s.Require().NoError(err)
	
	// Verify event emission
	events := s.ctx.EventManager().Events()
	event := s.findEventByType(events, types.EventTypeTokenizationCreated)
	s.Require().NotNil(event, "EventTokenizationCreated should be emitted")
	
	// Verify event attributes
	tokenizationDenomAttr := s.findAttributeByKey(event, types.AttributeKeyTokenizationDenom)
	s.Require().NotNil(tokenizationDenomAttr, "tokenization_denom attribute should be present")
	s.Require().Equal("tokenization", tokenizationDenomAttr.Value)
	
	poolAmountAttr := s.findAttributeByKey(event, types.AttributeKeyPoolAmount)
	s.Require().NotNil(poolAmountAttr, "pool_amount attribute should be present")
	s.Require().Equal("500", poolAmountAttr.Value)
	
	nftClassIdAttr := s.findAttributeByKey(event, types.AttributeKeyNftClassId)
	s.Require().NotNil(nftClassIdAttr, "nft_class_id attribute should be present")
	s.Require().Equal("asset-class-token", nftClassIdAttr.Value)
	
	nftIdAttr := s.findAttributeByKey(event, types.AttributeKeyNftId)
	s.Require().NotNil(nftIdAttr, "nft_id attribute should be present")
	s.Require().Equal("asset-token-1", nftIdAttr.Value)
	
	ownerAttr := s.findAttributeByKey(event, types.AttributeKeyOwner)
	s.Require().NotNil(ownerAttr, "owner attribute should be present")
	s.Require().Equal(s.user1Addr.String(), ownerAttr.Value)
}

func (s *MsgServerTestSuite) TestCreateSecuritization() {
	msgServer := keeper.NewMsgServerImpl(s.app.AssetKeeper)
	
	// Clear events before test
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	
	// First create the pools that will be referenced in the securitization
	// Create an asset class for the pools
	ledgerClass := ledgertypes.LedgerClass{
		LedgerClassId: "asset-class-sec",
		AssetClassId:  "asset-class-sec",
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
	err = s.app.LedgerKeeper.AddClassStatusType(s.ctx, s.user1Addr, "asset-class-sec", statusType)
	s.Require().NoError(err)
	
	assetClassMsg := &types.MsgCreateAssetClass{
		AssetClass: &types.AssetClass{
			Id: "asset-class-sec",
			Name: "AssetClassSec",
		},
		LedgerClass: "asset-class-sec",
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAssetClass(s.ctx, assetClassMsg)
	s.Require().NoError(err)
	
	// Create assets for the pools
	asset1Msg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-sec",
			Id:      "asset-sec-1",
			Uri:     "https://example.com/asset1",
			UriHash: "abc123",
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, asset1Msg)
	s.Require().NoError(err)
	
	asset2Msg := &types.MsgCreateAsset{
		Asset: &types.Asset{
			ClassId: "asset-class-sec",
			Id:      "asset-sec-2",
			Uri:     "https://example.com/asset2",
			UriHash: "def456",
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateAsset(s.ctx, asset2Msg)
	s.Require().NoError(err)
	
	// Create the pools
	pool1Msg := &types.MsgCreatePool{
		Pool: &sdk.Coin{
			Denom:  "pool1",
			Amount: sdkmath.NewInt(1000),
		},
		Nfts: []*types.Nft{
			{
				ClassId: "asset-class-sec",
				Id:      "asset-sec-1",
			},
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreatePool(s.ctx, pool1Msg)
	s.Require().NoError(err)
	
	pool2Msg := &types.MsgCreatePool{
		Pool: &sdk.Coin{
			Denom:  "pool2",
			Amount: sdkmath.NewInt(2000),
		},
		Nfts: []*types.Nft{
			{
				ClassId: "asset-class-sec",
				Id:      "asset-sec-2",
			},
		},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreatePool(s.ctx, pool2Msg)
	s.Require().NoError(err)
	
	// Clear events before creating securitization
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	
	// Now create the securitization
	msg := &types.MsgCreateSecuritization{
		Id: "sec-1",
		Tranches: []*sdk.Coin{
			{Denom: "tranche-a", Amount: sdkmath.NewInt(100)},
			{Denom: "tranche-b", Amount: sdkmath.NewInt(200)},
		},
		Pools: []string{"pool1", "pool2"},
		FromAddress: s.user1Addr.String(),
	}
	_, err = msgServer.CreateSecuritization(s.ctx, msg)
	s.Require().NoError(err)
	
	// Verify event emission
	events := s.ctx.EventManager().Events()
	event := s.findEventByType(events, types.EventTypeSecuritizationCreated)
	s.Require().NotNil(event, "EventSecuritizationCreated should be emitted")
	
	// Verify event attributes
	securitizationIdAttr := s.findAttributeByKey(event, types.AttributeKeySecuritizationId)
	s.Require().NotNil(securitizationIdAttr, "securitization_id attribute should be present")
	s.Require().Equal("sec-1", securitizationIdAttr.Value)
	
	trancheCountAttr := s.findAttributeByKey(event, types.AttributeKeyTrancheCount)
	s.Require().NotNil(trancheCountAttr, "tranche_count attribute should be present")
	s.Require().Equal("2", trancheCountAttr.Value)
	
	poolCountAttr := s.findAttributeByKey(event, types.AttributeKeyPoolCount)
	s.Require().NotNil(poolCountAttr, "pool_count attribute should be present")
	s.Require().Equal("2", poolCountAttr.Value)
	
	ownerAttr := s.findAttributeByKey(event, types.AttributeKeyOwner)
	s.Require().NotNil(ownerAttr, "owner attribute should be present")
	s.Require().Equal(s.user1Addr.String(), ownerAttr.Value)
}
