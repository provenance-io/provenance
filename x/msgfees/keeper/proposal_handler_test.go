package keeper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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
	s.k = msgfeeskeeper.NewKeeper(s.app.AppCodec(), s.app.GetKey(msgfeestypes.ModuleName), s.app.GetSubspace(msgfeestypes.ModuleName), "", msgfeestypes.NhashDenom, nil, nil)
	s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestMarkerProposals() {
	writeRecordRequest := &metadatatypes.MsgWriteRecordRequest{}
	writeScopeRequest := &metadatatypes.MsgWriteScopeRequest{}

	testCases := []struct {
		name string
		prop govtypes.Content
		err  error
	}{
		{
			"add msgfees - valid",
			msgfeestypes.NewAddMsgFeeProposal("title add", "description", sdk.MsgTypeURL(writeRecordRequest), sdk.NewCoin("hotdog", sdk.NewInt(10))),
			nil,
		},
		{
			"add msgfees - invalid - cannot add when the same msgfee exists",
			msgfeestypes.NewAddMsgFeeProposal("title add", "description", sdk.MsgTypeURL(writeRecordRequest), sdk.NewCoin("hotdog", sdk.NewInt(10))),
			msgfeestypes.ErrMsgFeeAlreadyExists,
		},
		{
			"add msgfees - invalid - validate basic fail",
			msgfeestypes.NewAddMsgFeeProposal("title add", "description", sdk.MsgTypeURL(writeScopeRequest), sdk.NewCoin("hotdog", sdk.NewInt(0))),
			msgfeestypes.ErrInvalidFee,
		},
		{
			"update msgfees - valid",
			msgfeestypes.NewUpdateMsgFeeProposal("title update", "description", sdk.MsgTypeURL(writeRecordRequest), sdk.NewCoin("hotdog", sdk.NewInt(10))),
			nil,
		},
		{
			"update msgfees - invalid - cannot update a non-existing msgfee",
			msgfeestypes.NewUpdateMsgFeeProposal("title update", "description", sdk.MsgTypeURL(writeScopeRequest), sdk.NewCoin("hotdog", sdk.NewInt(10))),
			msgfeestypes.ErrMsgFeeDoesNotExist,
		},
		{
			"update msgfees - invalid - validate basic fail",
			msgfeestypes.NewUpdateMsgFeeProposal("title update", "description", sdk.MsgTypeURL(writeRecordRequest), sdk.NewCoin("hotdog", sdk.NewInt(0))),
			msgfeestypes.ErrInvalidFee,
		},
		{
			"remove msgfees - valid",
			msgfeestypes.NewRemoveMsgFeeProposal("title remove", "description", sdk.MsgTypeURL(writeRecordRequest)),
			nil,
		},
		{
			"remove msgfees - invalid - cannot remove a non-existing msgfee",
			msgfeestypes.NewRemoveMsgFeeProposal("title remove", "description", sdk.MsgTypeURL(writeRecordRequest)),
			msgfeestypes.ErrMsgFeeDoesNotExist,
		},
		{
			"remove msgfees - invalid - validate basic fail",
			msgfeestypes.NewRemoveMsgFeeProposal("title remove", "description", ""),
			msgfeestypes.ErrEmptyMsgType,
		},
		{
			"update nhash to usd mil - invalid - validate basic fail",
			msgfeestypes.NewUpdateNhashPerUsdMilProposal("title update conversion", "", 10),
			errors.New("proposal description cannot be blank: invalid proposal content"),
		},
		{
			"update nhash to usd mil - valid",
			msgfeestypes.NewUpdateNhashPerUsdMilProposal("title update conversion", "description", 1),
			nil,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.T().Run(tc.name, func(t *testing.T) {
			var err error
			switch c := tc.prop.(type) {
			case *msgfeestypes.AddMsgFeeProposal:
				err = msgfeeskeeper.HandleAddMsgFeeProposal(s.ctx, s.k, c, s.app.InterfaceRegistry())
			case *msgfeestypes.UpdateMsgFeeProposal:
				err = msgfeeskeeper.HandleUpdateMsgFeeProposal(s.ctx, s.k, c, s.app.InterfaceRegistry())
			case *msgfeestypes.RemoveMsgFeeProposal:
				err = msgfeeskeeper.HandleRemoveMsgFeeProposal(s.ctx, s.k, c, s.app.InterfaceRegistry())
			case *msgfeestypes.UpdateNhashPerUsdMilProposal:
				err = msgfeeskeeper.HandleUpdateNhashPerUsdMilProposal(s.ctx, s.k, c, s.app.InterfaceRegistry())
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
