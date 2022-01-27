package app

import (
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *App
	ctx sdk.Context
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.app = Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

func (s *IntegrationTestSuite) TestMarkerProposals() {
	err := AddMsgBasedFees(s.app, s.ctx)
	s.Assert().NoError(err)

	msgfee, err := s.app.MsgFeesKeeper.GetMsgFee(s.ctx, sdk.MsgTypeURL(&nametypes.MsgBindNameRequest{}))
	s.Assert().NoError(err)
	s.Assert().NotNil(msgfee)
	s.Assert().Equal(sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)), msgfee.AdditionalFee)

	msgfee, err = s.app.MsgFeesKeeper.GetMsgFee(s.ctx, sdk.MsgTypeURL(&markertypes.MsgAddMarkerRequest{}))
	s.Assert().NoError(err)
	s.Assert().NotNil(msgfee)
	s.Assert().Equal(sdk.NewCoin("nhash", sdk.NewInt(100_000_000_000)), msgfee.AdditionalFee)

	msgfee, err = s.app.MsgFeesKeeper.GetMsgFee(s.ctx, sdk.MsgTypeURL(&attributetypes.MsgAddAttributeRequest{}))
	s.Assert().NoError(err)
	s.Assert().NotNil(msgfee)
	s.Assert().Equal(sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)), msgfee.AdditionalFee)

	msgfee, err = s.app.MsgFeesKeeper.GetMsgFee(s.ctx, sdk.MsgTypeURL(&metadatatypes.MsgWriteScopeRequest{}))
	s.Assert().NoError(err)
	s.Assert().NotNil(msgfee)
	s.Assert().Equal(sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)), msgfee.AdditionalFee)

	msgfee, err = s.app.MsgFeesKeeper.GetMsgFee(s.ctx, sdk.MsgTypeURL(&metadatatypes.MsgP8EMemorializeContractRequest{}))
	s.Assert().NoError(err)
	s.Assert().NotNil(msgfee)
	s.Assert().Equal(sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)), msgfee.AdditionalFee)
}
