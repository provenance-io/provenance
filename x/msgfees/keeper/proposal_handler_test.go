package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeeskeeper "github.com/provenance-io/provenance/x/msgfees/keeper"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"

	provenance "github.com/provenance-io/provenance/app"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *provenance.App
	ctx sdk.Context
	k   msgfeeskeeper.Keeper

	accountAddr sdk.AccAddress
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.app = provenance.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.k = msgfeeskeeper.NewKeeper(s.app.AppCodec(), s.app.GetKey(msgfeestypes.ModuleName), s.app.GetSubspace(msgfeestypes.ModuleName), "")
	s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestMarkerProposals() {
	writeRecordRequest, err := cdctypes.NewAnyWithValue(&metadatatypes.MsgWriteRecordRequest{})
	s.Require().NoError(err)
	writeScopeRequest, err := cdctypes.NewAnyWithValue(&metadatatypes.MsgWriteScopeRequest{})
	s.Require().NoError(err)

	testCases := []struct {
		name string
		prop govtypes.Content
		err  error
	}{
		{
			"add msgfees - valid",
			msgfeestypes.NewAddMsgBasedFeeProposal("title add", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), writeRecordRequest, sdk.NewCoin("hotdog", sdk.NewInt(10)), sdk.OneDec()),
			nil,
		},
		{
			"add msgfees - invalid - cannot add when the same msgbasedfee exists",
			msgfeestypes.NewAddMsgBasedFeeProposal("title add", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), writeRecordRequest, sdk.NewCoin("hotdog", sdk.NewInt(10)), sdk.OneDec()),
			msgfeestypes.ErrMsgFeeAlreadyExists,
		},
		{
			"update msgfees - valid",
			msgfeestypes.NewUpdateMsgBasedFeeProposal("title update", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), writeRecordRequest, sdk.NewCoin("hotdog", sdk.NewInt(10)), sdk.OneDec()),
			nil,
		},
		{
			"update msgfees - invalid - cannot update a non-existing msgbasedfee",
			msgfeestypes.NewUpdateMsgBasedFeeProposal("title update", "description", sdk.NewCoin("hotdog", sdk.NewInt(10)), writeScopeRequest, sdk.NewCoin("hotdog", sdk.NewInt(10)), sdk.OneDec()),
			msgfeestypes.ErrMsgFeeDoesNotExist,
		},
		{
			"remove msgfees - valid",
			msgfeestypes.NewRemoveMsgBasedFeeProposal("title remove", "description", writeRecordRequest),
			nil,
		},
		{
			"remove msgfees - invalid - cannot remove a non-existing msgbasedfee",
			msgfeestypes.NewRemoveMsgBasedFeeProposal("title remove", "description", writeRecordRequest),
			msgfeestypes.ErrMsgFeeDoesNotExist,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {

			var err error
			switch c := tc.prop.(type) {
			case *msgfeestypes.AddMsgBasedFeeProposal:
				err = msgfeeskeeper.HandleAddMsgBasedFeeProposal(s.ctx, s.k, c)
			case *msgfeestypes.UpdateMsgBasedFeeProposal:
				err = msgfeeskeeper.HandleUpdateMsgBasedFeeProposal(s.ctx, s.k, c)
			case *msgfeestypes.RemoveMsgBasedFeeProposal:
				err = msgfeeskeeper.HandleRemoveMsgBasedFeeProposal(s.ctx, s.k, c)
			default:
				panic("invalid proposal type")
			}

			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}

}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
