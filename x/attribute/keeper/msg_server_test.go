package keeper_test

import (
	"fmt"
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
	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"
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
	s.msgServer = keeper.NewMsgServerImpl(s.app.AttributeKeeper)
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

func (s *MsgServerTestSuite) TestUpdateAttributeExpiration() {
	twoHoursInPast := time.Now().UTC().Add(-2 * time.Hour)
	twoHoursInFuture := time.Now().UTC().Add(2 * time.Hour)
	name := "jackthecat.io"
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, name, s.owner1Addr, false))
	attr := types.Attribute{
		Name:          name,
		Value:         []byte("my-value"),
		AttributeType: types.AttributeType_String,
		Address:       s.owner1,
	}
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.owner1Addr), "should save successfully")

	tests := []struct {
		name     string
		msg      types.MsgUpdateAttributeExpirationRequest
		errorMsg string
	}{
		{
			name: "Should fail to parse owner address",
			msg: *&types.MsgUpdateAttributeExpirationRequest{
				Name:           "",
				Value:          []byte{},
				Owner:          "wrong format",
				ExpirationDate: nil,
			},
			errorMsg: `decoding bech32 failed: invalid character in string: ' '`,
		},
		{
			name:     "Should fail to normalize attribute",
			msg:      *types.NewMsgUpdateAttributeExpirationRequest("", "", "", nil, s.owner1Addr),
			errorMsg: `unable to normalize attribute name "": segment of name is too short`,
		},
		{
			name:     "Should fail, block time is ahead of expiration time",
			msg:      *types.NewMsgUpdateAttributeExpirationRequest("", "", "", &twoHoursInPast, s.owner1Addr),
			errorMsg: fmt.Sprintf("attribute expiration date %v is before block time of %v", twoHoursInPast.UTC(), s.ctx.BlockTime().UTC()),
		},
		{
			name: "Should succeed",
			msg:  *types.NewMsgUpdateAttributeExpirationRequest(attr.Address, name, "my-value", &twoHoursInFuture, s.owner1Addr),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			response, err := s.msgServer.UpdateAttributeExpiration(s.ctx, &tt.msg)
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
