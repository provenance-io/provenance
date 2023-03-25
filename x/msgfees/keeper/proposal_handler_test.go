package keeper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/provenance-io/provenance/internal/pioconfig"
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
	s.app = provenance.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.k = msgfeeskeeper.NewKeeper(s.app.AppCodec(), s.app.GetKey(msgfeestypes.ModuleName), s.app.GetSubspace(msgfeestypes.ModuleName), "", pioconfig.GetProvenanceConfig().FeeDenom, nil, nil)
	s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestMsgFeeProposals() {
	writeRecordRequest := &metadatatypes.MsgWriteRecordRequest{}
	writeScopeRequest := &metadatatypes.MsgWriteScopeRequest{}

	testCases := []struct {
		name string
		prop govtypesv1beta1.Content
		err  error
	}{
		{
			name: "add msgfees - valid",
			prop: msgfeestypes.NewAddMsgFeeProposal("title add", "description", sdk.MsgTypeURL(writeRecordRequest), sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			err: nil,
		},
		{
			name: "add msgfees - invalid - cannot add when the same msgfee exists",
			prop: msgfeestypes.NewAddMsgFeeProposal("title add", "description", sdk.MsgTypeURL(writeRecordRequest), sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			err: msgfeestypes.ErrMsgFeeAlreadyExists,
		},
		{
			name: "add msgfees - invalid - validate basic fail",
			prop: msgfeestypes.NewAddMsgFeeProposal("title add", "description", sdk.MsgTypeURL(writeScopeRequest), sdk.NewCoin("hotdog", sdk.NewInt(0)), "", ""),
			err: msgfeestypes.ErrInvalidFee,
		},
		{
			name: "update msgfees - valid",
			prop: msgfeestypes.NewUpdateMsgFeeProposal("title update", "description", sdk.MsgTypeURL(writeRecordRequest), sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			err: nil,
		},
		{
			name: "update msgfees - invalid - cannot update a non-existing msgfee",
			prop: msgfeestypes.NewUpdateMsgFeeProposal("title update", "description", sdk.MsgTypeURL(writeScopeRequest), sdk.NewCoin("hotdog", sdk.NewInt(10)), "", ""),
			err: msgfeestypes.ErrMsgFeeDoesNotExist,
		},
		{
			name: "update msgfees - invalid - validate basic fail",
			prop: msgfeestypes.NewUpdateMsgFeeProposal("title update", "description", sdk.MsgTypeURL(writeRecordRequest), sdk.NewCoin("hotdog", sdk.NewInt(0)), "", ""),
			err: msgfeestypes.ErrInvalidFee,
		},
		{
			name: "remove msgfees - valid",
			prop: msgfeestypes.NewRemoveMsgFeeProposal("title remove", "description", sdk.MsgTypeURL(writeRecordRequest)),
			err: nil,
		},
		{
			name: "remove msgfees - invalid - cannot remove a non-existing msgfee",
			prop: msgfeestypes.NewRemoveMsgFeeProposal("title remove", "description", sdk.MsgTypeURL(writeRecordRequest)),
			err: msgfeestypes.ErrMsgFeeDoesNotExist,
		},
		{
			name: "remove msgfees - invalid - validate basic fail",
			prop: msgfeestypes.NewRemoveMsgFeeProposal("title remove", "description", ""),
			err: msgfeestypes.ErrEmptyMsgType,
		},
		{
			name: "update nhash to usd mil - invalid - validate basic fail",
			prop: msgfeestypes.NewUpdateNhashPerUsdMilProposal("title update conversion", "", 10),
			err: errors.New("proposal description cannot be blank: invalid proposal content"),
		},
		{
			name: "update nhash to usd mil - valid",
			prop: msgfeestypes.NewUpdateNhashPerUsdMilProposal("title update conversion", "description", 1),
			err: nil,
		},
		{
			name: "update conversion fee denom - invalid - validate basic fail",
			prop: msgfeestypes.NewUpdateConversionFeeDenomProposal("title update conversion fee denom", "description", ""),
			err: errors.New("invalid denom: "),
		},
		{
			name: "update conversion fee denom - invalid - validate basic fail regex failure on denom",
			prop: msgfeestypes.NewUpdateConversionFeeDenomProposal("title update conversion fee denom", "description", "??"),
			err: errors.New("invalid denom: ??"),
		},
		{
			name: "update conversion fee denom - valid",
			prop: msgfeestypes.NewUpdateConversionFeeDenomProposal("title update conversion", "description", "hotdog"),
			err: nil,
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
			case *msgfeestypes.UpdateConversionFeeDenomProposal:
				err = msgfeeskeeper.HandleUpdateConversionFeeDenomProposal(s.ctx, s.k, c, s.app.InterfaceRegistry())
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

func (s *IntegrationTestSuite) TestDetermineBipsProposals() {
	testCases := []struct {
		name           string
		recipient      string
		bips           string
		expectedBips   uint32
		expectedErrMsg string
	}{
		{
		name: 	"valid - has recipient empty bips string, should return default bips",
		recipient: 	"recipient",
		bips: 	"",
		expectedBips: 	msgfeestypes.DefaultMsgFeeBips,
		expectedErrMsg: 	"",
		},
		{
		    name: 	"valid - has recipient and bips string, should return bips as uint32",
		    recipient: 	"recipient",
		    bips: 	"100",
		    expectedBips: 	100,
		    expectedErrMsg: 	"",
		},
		{
			name: "valid - has no recipient and a bips string, should return 0 bips",
			recipient: "",
			bips: "10",
			expectedBips: 0,
			expectedErrMsg: "",
		},
		{
			name: "invalid - has recipient and bips string too high, should error",
			recipient: "recipient",
			bips: "10001",
			expectedBips: 0,
			expectedErrMsg: "recipient basis points can only be between 0 and 10,000 : 10001: invalid bips amount",
		},
		{
			name: "invalid - has recipient and bips string not a number, should error",
			recipient: "recipient",
			bips: "error",
			expectedBips: 0,
			expectedErrMsg: "strconv.ParseUint: parsing \"error\": invalid syntax: invalid bips amount",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			bips, err := msgfeeskeeper.DetermineBips(tc.recipient, tc.bips)
			if len(tc.expectedErrMsg) != 0 {
				assert.Equal(t, uint32(0), bips, "should return 0 bips on error")
				assert.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBips, bips, "expected bips should match")
			}
		})
	}

}

func TestIntegrationTestSuite(t *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	suite.Run(t, new(IntegrationTestSuite))
}
