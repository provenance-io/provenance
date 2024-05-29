package keeper_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/ibchooks/keeper"
	"github.com/provenance-io/provenance/x/ibchooks/types"
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
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
	s.msgServer = keeper.NewMsgServerImpl(*s.app.IBCHooksKeeper)

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.owner1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.owner1 = s.owner1Addr.String()
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)

	var ibcHooksData types.GenesisState
	ibcHooksData.Params = types.DefaultParams()

	s.app.IBCHooksKeeper.InitGenesis(s.ctx, ibcHooksData)
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

func (s *MsgServerTestSuite) TestUpdateParams() {
	authority := s.app.IBCHooksKeeper.GetAuthority()

	tests := []struct {
		name          string
		expectedError string
		msg           *types.MsgUpdateParamsRequest
		expectedEvent proto.Message
	}{
		{
			name: "valid authority with valid params",
			msg: types.NewMsgUpdateParamsRequest(
				[]string{"cosmos1vh3htvc46rshps02w0p5hchdkrjvc4d8nxkw5t"},
				authority,
			),
			expectedEvent: &types.EventIBCHooksParamsUpdated{
				AllowedAsyncAckContracts: []string{"cosmos1vh3htvc46rshps02w0p5hchdkrjvc4d8nxkw5t"},
			},
		},
		{
			name: "invalid authority",
			msg: types.NewMsgUpdateParamsRequest(
				[]string{"cosmos1vh3htvc46rshps02w0p5hchdkrjvc4d8nxkw5t"},
				"invalid-authority",
			),
			expectedError: `expected "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" got "invalid-authority": expected gov account as only signer for proposal message`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.UpdateParams(s.ctx, tc.msg)
			if tc.expectedError != "" {
				s.Require().Error(err, "UpdateParams expected error but got none for case: %s", tc.name)
				s.Require().EqualError(err, tc.expectedError, "UpdateParams unexpected error message for case: %s", tc.name)
			} else {
				s.Require().NoError(err, "UpdateParams unexpected error for case: %s", tc.name)
			}
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Require().True(result, "UpdateParams expected typed event was not found: %v for case: %s", tc.expectedEvent, tc.name)
			}
		})
	}
}
