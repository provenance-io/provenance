package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/msgfees/keeper"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

type MsgServerTestSuite struct {
	suite.Suite

	app       *simapp.App
	ctx       sdk.Context
	msgServer types.MsgServer

	privkey1   cryptotypes.PrivKey
	pubkey1    cryptotypes.PubKey
	owner1     string
	owner1Addr sdk.AccAddress
	acct1      authtypes.AccountI

	addresses []sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(true, tmproto.Header{})
	s.ctx = s.ctx.WithBlockHeight(1).WithBlockTime(time.Now())
	s.msgServer = keeper.NewMsgServerImpl(s.app.MsgFeesKeeper)
	s.app.AccountKeeper.SetParams(s.ctx, authtypes.DefaultParams())
	s.app.BankKeeper.SetParams(s.ctx, banktypes.DefaultParams())

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.owner1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.owner1 = s.owner1Addr.String()
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) TestAddMsgFeeProposal() {
	typeUrl := sdk.MsgTypeURL(&types.MsgAddMsgFeeProposalRequest{})
	tests := []struct {
		name     string
		msg      types.MsgAddMsgFeeProposalRequest
		errorMsg string
	}{
		{
			name: "expected gov account for signer",
			msg: types.MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:    typeUrl,
				AdditionalFee: sdk.NewInt64Coin("nhash", 1),
				Authority:     "",
			},
			errorMsg: `expected cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn got : expected gov account as only signer for proposal message`,
		},
		{
			name: "msg type is empty",
			msg: types.MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:    "",
				AdditionalFee: sdk.NewInt64Coin("nhash", 1),
				Authority:     "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
			errorMsg: `msg type is empty`,
		},
		{
			name: "successful",
			msg: types.MsgAddMsgFeeProposalRequest{
				MsgTypeUrl:    typeUrl,
				AdditionalFee: sdk.NewInt64Coin("nhash", 1),
				Authority:     "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			response, err := s.msgServer.AddMsgFeeProposal(s.ctx, &tt.msg)
			if len(tt.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tt.errorMsg, err.Error())
				s.Assert().Nil(response)
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateMsgFeeProposal() {
	typeUrl := sdk.MsgTypeURL(&types.MsgAddMsgFeeProposalRequest{})
	msgFee := types.MsgFee{
		MsgTypeUrl:    typeUrl,
		AdditionalFee: sdk.NewInt64Coin("nhash", 1),
	}
	s.app.MsgFeesKeeper.SetMsgFee(s.ctx, msgFee)
	tests := []struct {
		name     string
		msg      types.MsgUpdateMsgFeeProposalRequest
		errorMsg string
	}{
		{
			name: "expected gov account for signer",
			msg: types.MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:    typeUrl,
				AdditionalFee: sdk.NewInt64Coin("nhash", 1),
				Authority:     "",
			},
			errorMsg: `expected cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn got : expected gov account as only signer for proposal message`,
		},
		{
			name: "msg type is empty",
			msg: types.MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:    "",
				AdditionalFee: sdk.NewInt64Coin("nhash", 1),
				Authority:     "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
			errorMsg: `msg type is empty`,
		},
		{
			name: "successful",
			msg: types.MsgUpdateMsgFeeProposalRequest{
				MsgTypeUrl:    typeUrl,
				AdditionalFee: sdk.NewInt64Coin("nhash", 2),
				Authority:     "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			response, err := s.msgServer.UpdateMsgFeeProposal(s.ctx, &tt.msg)
			if len(tt.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tt.errorMsg, err.Error())
				s.Assert().Nil(response)
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestRemoveMsgFeeProposal() {
	typeUrl := sdk.MsgTypeURL(&types.MsgAddMsgFeeProposalRequest{})
	msgFee := types.MsgFee{
		MsgTypeUrl:    typeUrl,
		AdditionalFee: sdk.NewInt64Coin("nhash", 1),
	}
	s.app.MsgFeesKeeper.SetMsgFee(s.ctx, msgFee)
	tests := []struct {
		name     string
		msg      types.MsgRemoveMsgFeeProposalRequest
		errorMsg string
	}{
		{
			name: "expected gov account for signer",
			msg: types.MsgRemoveMsgFeeProposalRequest{
				MsgTypeUrl: typeUrl,
				Authority:  "",
			},
			errorMsg: `expected cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn got : expected gov account as only signer for proposal message`,
		},
		{
			name: "msg type is empty",
			msg: types.MsgRemoveMsgFeeProposalRequest{
				MsgTypeUrl: "",
				Authority:  "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
			errorMsg: `msg type is empty`,
		},
		{
			name: "successful",
			msg: types.MsgRemoveMsgFeeProposalRequest{
				MsgTypeUrl: typeUrl,
				Authority:  "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			response, err := s.msgServer.RemoveMsgFeeProposal(s.ctx, &tt.msg)
			if len(tt.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tt.errorMsg, err.Error())
				s.Assert().Nil(response)
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateNhashPerUsdMilProposal() {
	tests := []struct {
		name     string
		msg      types.MsgUpdateNhashPerUsdMilProposalRequest
		errorMsg string
	}{
		{
			name: "expected gov account for signer",
			msg: types.MsgUpdateNhashPerUsdMilProposalRequest{
				Authority: "",
			},
			errorMsg: `expected cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn got : expected gov account as only signer for proposal message`,
		},
		{
			name: "invalid NhashPerUsdMil amount",
			msg: types.MsgUpdateNhashPerUsdMilProposalRequest{
				NhashPerUsdMil: 0,
				Authority:      "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
			errorMsg: `nhash per usd mil must be greater than 0`,
		},
		{
			name: "successful",
			msg: types.MsgUpdateNhashPerUsdMilProposalRequest{
				NhashPerUsdMil: 10,
				Authority:      "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			response, err := s.msgServer.UpdateNhashPerUsdMilProposal(s.ctx, &tt.msg)
			if len(tt.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tt.errorMsg, err.Error())
				s.Assert().Nil(response)
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(response)
			}
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateConversionFeeDenomProposal() {
	tests := []struct {
		name     string
		msg      types.MsgUpdateConversionFeeDenomProposalRequest
		errorMsg string
	}{
		{
			name: "expected gov account for signer",
			msg: types.MsgUpdateConversionFeeDenomProposalRequest{
				Authority: "",
			},
			errorMsg: `expected cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn got : expected gov account as only signer for proposal message`,
		},
		{
			name: "invalid denom",
			msg: types.MsgUpdateConversionFeeDenomProposalRequest{
				Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
			errorMsg: `invalid denom: `,
		},
		{
			name: "successful",
			msg: types.MsgUpdateConversionFeeDenomProposalRequest{
				ConversionFeeDenom: "nhash",
				Authority:          "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			response, err := s.msgServer.UpdateConversionFeeDenomProposal(s.ctx, &tt.msg)
			if len(tt.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tt.errorMsg, err.Error())
				s.Assert().Nil(response)
			} else {
				s.Assert().NoError(err)
				s.Assert().NotNil(response)
			}
		})
	}
}
