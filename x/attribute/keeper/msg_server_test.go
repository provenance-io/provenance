package keeper_test

import (
	"fmt"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
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
	acct1      sdk.AccountI

	addresses []sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
	s.ctx = s.ctx.WithBlockHeight(1).WithBlockTime(time.Now())
	s.msgServer = keeper.NewMsgServerImpl(s.app.AttributeKeeper)
	s.app.AccountKeeper.Params.Set(s.ctx, authtypes.DefaultParams())
	s.app.BankKeeper.SetParams(s.ctx, banktypes.DefaultParams())

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.owner1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.owner1 = s.owner1Addr.String()
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)

	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("name", s.owner1Addr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.name", s.owner1Addr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 16
	nameData.Params.MinSegmentLength = 2
	nameData.Params.MaxSegmentLength = 16

	s.app.NameKeeper.InitGenesis(s.ctx, nameData)
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) containsMessage(events []abci.Event, msg proto.Message) bool {
	for _, event := range events {
		typeEvent, _ := sdk.ParseTypedEvent(event)
		if assert.ObjectsAreEqual(msg, typeEvent) {
			return true
		}
	}
	return false
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

func (s *MsgServerTestSuite) TestMsgAddAttributeRequest() {

	testcases := []struct {
		name          string
		msg           *types.MsgAddAttributeRequest
		signers       []string
		errorMsg      string
		expectedEvent proto.Message
	}{
		{
			name: "should successfully add new attribute",
			msg: types.NewMsgAddAttributeRequest(s.owner1,
				s.owner1Addr, "example.name", types.AttributeType_String, []byte("value")),
			signers: []string{s.owner1},
			expectedEvent: types.NewEventAttributeAdd(
				types.Attribute{
					Address:       s.owner1,
					Name:          "example.name",
					Value:         []byte("value"),
					AttributeType: types.AttributeType_String,
				},
				s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.AddAttribute(s.ctx, tc.msg)

			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg)
			} else {
				if tc.expectedEvent != nil {
					result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
					s.True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
				}

			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgUpdateAttributeRequest() {
	testAttr := types.Attribute{
		Address:       s.owner1,
		Name:          "example.name",
		Value:         []byte("value"),
		AttributeType: types.AttributeType_String,
	}
	var attrData types.GenesisState
	attrData.Attributes = append(attrData.Attributes, testAttr)
	attrData.Params.MaxValueLength = 100
	s.app.AttributeKeeper.InitGenesis(s.ctx, &attrData)

	testcases := []struct {
		name          string
		msg           *types.MsgUpdateAttributeRequest
		signers       []string
		errorMsg      string
		expectedEvent proto.Message
	}{
		{
			name: "should successfully update attribute",
			msg: types.NewMsgUpdateAttributeRequest(
				s.owner1,
				s.owner1Addr, "example.name",
				[]byte("value"), []byte("1"),
				types.AttributeType_String,
				types.AttributeType_Int),
			signers: []string{s.owner1},
			expectedEvent: types.NewEventAttributeUpdate(
				testAttr,
				types.Attribute{
					Address:       s.owner1,
					Name:          "example.name",
					Value:         []byte("1"),
					AttributeType: types.AttributeType_Int,
				},
				s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.UpdateAttribute(s.ctx, tc.msg)

			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg)
			} else {
				if tc.expectedEvent != nil {
					result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
					s.True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
				}

			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgDistinctDeleteAttributeRequest() {
	testAttr := types.Attribute{
		Address:       s.owner1,
		Name:          "example.name",
		Value:         []byte("value"),
		AttributeType: types.AttributeType_String,
	}
	var attrData types.GenesisState
	attrData.Attributes = append(attrData.Attributes, testAttr)
	attrData.Params.MaxValueLength = 100
	s.app.AttributeKeeper.InitGenesis(s.ctx, &attrData)

	testcases := []struct {
		name          string
		msg           *types.MsgDeleteDistinctAttributeRequest
		signers       []string
		errorMsg      string
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully delete attribute with value",
			msg:           types.NewMsgDeleteDistinctAttributeRequest(s.owner1, s.owner1Addr, "example.name", []byte("value")),
			signers:       []string{s.owner1},
			expectedEvent: types.NewEventDistinctAttributeDelete("example.name", "value", s.owner1, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.DeleteDistinctAttribute(s.ctx, tc.msg)

			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg)
			} else {
				if tc.expectedEvent != nil {
					result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
					s.True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
				}

			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgDeleteAttributeRequest() {
	testAttr := types.Attribute{
		Address:       s.owner1,
		Name:          "example.name",
		Value:         []byte("value"),
		AttributeType: types.AttributeType_String,
	}
	var attrData types.GenesisState
	attrData.Attributes = append(attrData.Attributes, testAttr)
	attrData.Params.MaxValueLength = 100
	s.app.AttributeKeeper.InitGenesis(s.ctx, &attrData)

	testcases := []struct {
		name          string
		msg           *types.MsgDeleteAttributeRequest
		signers       []string
		errorMsg      string
		expectedEvent proto.Message
	}{
		{
			name:          "should successfully add new attribute",
			msg:           types.NewMsgDeleteAttributeRequest(s.owner1, s.owner1Addr, "example.name"),
			signers:       []string{s.owner1},
			expectedEvent: types.NewEventAttributeDelete("example.name", s.owner1, s.owner1),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.DeleteAttribute(s.ctx, tc.msg)

			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg)
			} else {
				if tc.expectedEvent != nil {
					result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
					s.True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
				}

			}
		})
	}
}
