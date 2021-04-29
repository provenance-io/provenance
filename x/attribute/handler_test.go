package attribute_test

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute"
	"github.com/provenance-io/provenance/x/attribute/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type HandlerTestSuite struct {
	suite.Suite

	app     *app.App
	ctx     sdk.Context
	handler sdk.Handler

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func (s *HandlerTestSuite) SetupTest() {
	s.app = app.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.handler = attribute.NewHandler(s.app.AttributeKeeper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("name", s.user1Addr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.name", s.user1Addr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 16
	nameData.Params.MinSegmentLength = 2
	nameData.Params.MaxSegmentLength = 16

	s.app.NameKeeper.InitGenesis(s.ctx, nameData)

}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

type CommonTest struct {
	name          string
	msg           sdk.Msg
	signers       []string
	errorMsg      string
	expectedEvent proto.Message
}

func (s HandlerTestSuite) containsMessage(msg proto.Message) bool {
	events := s.ctx.EventManager().Events().ToABCIEvents()
	for _, event := range events {
		typeEvent, _ := sdk.ParseTypedEvent(event)
		if assert.ObjectsAreEqual(msg, typeEvent) {
			return true
		}
	}
	return false
}

func (s HandlerTestSuite) runTests(cases []CommonTest) {
	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.handler(s.ctx, tc.msg)

			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				if tc.expectedEvent != nil {
					result := s.containsMessage(tc.expectedEvent)
					s.True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
				}

			}
		})
	}
}

func (s HandlerTestSuite) TestMsgAddAttributeRequest() {
	cases := []CommonTest{
		{
			"should successfully add new attribute",
			types.NewMsgAddAttributeRequest(s.user1Addr,
				s.user1Addr, "example.name", types.AttributeType_String, []byte("value")),
			[]string{s.user1},
			"",
			types.NewEventAttributeAdd(
				types.Attribute{
					Address:       s.user1,
					Name:          "example.name",
					Value:         []byte("value"),
					AttributeType: types.AttributeType_String,
				},
				s.user1),
		},
	}
	s.runTests(cases)
}

func (s HandlerTestSuite) TestMsgDeleteAttributeRequest() {
	testAttr := types.Attribute{
		Address:       s.user1,
		Name:          "example.name",
		Value:         []byte("value"),
		AttributeType: types.AttributeType_String,
	}
	var attrData types.GenesisState
	attrData.Attributes = append(attrData.Attributes, testAttr)
	attrData.Params.MaxValueLength = 100
	s.app.AttributeKeeper.InitGenesis(s.ctx, &attrData)

	cases := []CommonTest{
		{
			"should successfully add new attribute",
			types.NewMsgDeleteAttributeRequest(s.user1Addr, s.user1Addr, "example.name"),
			[]string{s.user1},
			"",
			types.NewEventAttributeDelete("example.name", s.user1, s.user1),
		},
	}
	s.runTests(cases)
}
