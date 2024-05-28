package keeper_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	simapp "github.com/provenance-io/provenance/app"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
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

	privkey2   cryptotypes.PrivKey
	pubkey2    cryptotypes.PubKey
	owner2     string
	owner2Addr sdk.AccAddress
	acct2      sdk.AccountI

	addresses []sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
	s.msgServer = keeper.NewMsgServerImpl(s.app.NameKeeper)
	s.app.AccountKeeper.Params.Set(s.ctx, authtypes.DefaultParams())
	s.app.BankKeeper.SetParams(s.ctx, banktypes.DefaultParams())

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.owner1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.owner1 = s.owner1Addr.String()
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)

	s.privkey2 = secp256k1.GenPrivKey()
	s.pubkey2 = s.privkey2.PubKey()
	s.owner2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.owner2 = s.owner2Addr.String()
	acc2 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner2Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc2)

	var nameData types.GenesisState
	nameData.Bindings = append(nameData.Bindings, types.NewNameRecord("name", s.owner1Addr, false))
	nameData.Bindings = append(nameData.Bindings, types.NewNameRecord("example.name", s.owner1Addr, false))
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

func (s *MsgServerTestSuite) TestDeleteNameRequest() {
	name := "jackthecat.io"
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, name, s.owner1Addr, false))
	tests := []struct {
		name     string
		msg      types.MsgDeleteNameRequest
		errorMsg string
	}{
		{
			name:     "Should fail to validatebasic on msg",
			msg:      *types.NewMsgDeleteNameRequest(types.NewNameRecord("", sdk.AccAddress{}, false)),
			errorMsg: "name cannot be empty: invalid request",
		},
		{
			name:     "Should fail to normalize name",
			msg:      *types.NewMsgDeleteNameRequest(types.NewNameRecord("i", s.owner1Addr, false)),
			errorMsg: "segment of name is too short: invalid request",
		},
		{
			name:     "Should fail to parse address",
			msg:      *types.NewMsgDeleteNameRequest(types.NameRecord{Name: "provenance.io", Address: "s.owner1Addr", Restricted: false}),
			errorMsg: "decoding bech32 failed: string not all lowercase or all uppercase: invalid request",
		},
		{
			name:     "Should fail to name does not exist",
			msg:      *types.NewMsgDeleteNameRequest(types.NewNameRecord("provenance.io", s.owner1Addr, false)),
			errorMsg: "name does not exist: invalid request",
		},
		{
			name:     "Should fail name does not resolve to owner",
			msg:      *types.NewMsgDeleteNameRequest(types.NewNameRecord(name, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()), false)),
			errorMsg: "msg sender cannot delete name: unauthorized",
		},
		{
			name:     "Should succeed to delete",
			msg:      *types.NewMsgDeleteNameRequest(types.NewNameRecord(name, s.owner1Addr, false)),
			errorMsg: "",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			response, err := s.msgServer.DeleteName(s.ctx, &tt.msg)
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

func (s *MsgServerTestSuite) TestDeleteNameRemovingAttributeAccounts() {
	name := "jackthecat.io"
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, name, s.owner1Addr, false))
	attrAccounts := make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		attrAccounts[i] = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attrtypes.NewAttribute(name, attrAccounts[i].String(), attrtypes.AttributeType_String, []byte(attrAccounts[i].String()), nil), s.owner1Addr))
		attrStore := s.ctx.KVStore(s.app.GetKey(attrtypes.StoreKey))
		key := attrtypes.AttributeNameAddrKeyPrefix(name, attrAccounts[i])
		address, _ := attrtypes.GetAddressFromKey(key)
		bz := attrStore.Get(key)
		s.Assert().Equal(attrAccounts[i], address)
		s.Assert().Equal(uint64(1), binary.BigEndian.Uint64(bz))

	}
	attrAddresses, err := s.app.AttributeKeeper.AccountsByAttribute(s.ctx, name)
	s.Assert().NoError(err)
	s.Assert().ElementsMatch(attrAccounts, attrAddresses)

	result, err := s.msgServer.DeleteName(s.ctx, types.NewMsgDeleteNameRequest(types.NewNameRecord(name, s.owner1Addr, false)))
	s.Assert().NotNil(result)
	s.Assert().NoError(err)

	attrAddresses, err = s.app.AttributeKeeper.AccountsByAttribute(s.ctx, name)
	s.Assert().NoError(err)
	s.Assert().Len(attrAddresses, 0)

	for i := 0; i < 10; i++ {
		attrStore := s.ctx.KVStore(s.app.GetKey(attrtypes.StoreKey))
		key := attrtypes.AttributeNameAddrKeyPrefix(name, attrAccounts[i])
		bz := attrStore.Get(key)
		s.Assert().Nil(bz)

	}
}

// create name record
func (s *MsgServerTestSuite) TestCreateName() {
	tests := []struct {
		name          string
		expectedError error
		msg           *types.MsgBindNameRequest
		expectedEvent proto.Message
	}{
		{
			name:          "create name record",
			msg:           types.NewMsgBindNameRequest(types.NewNameRecord("new", s.owner2Addr, false), types.NewNameRecord("example.name", s.owner1Addr, false)),
			expectedError: nil,
			expectedEvent: types.NewEventNameBound(s.owner2, "new.example.name", false),
		},
		{
			name:          "create bad name record",
			msg:           types.NewMsgBindNameRequest(types.NewNameRecord("new", s.owner2Addr, false), types.NewNameRecord("foo.name", s.owner1Addr, false)),
			expectedError: sdkerrors.ErrInvalidRequest.Wrap(types.ErrNameNotBound.Error()),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.BindName(s.ctx, tc.msg)
			if tc.expectedError != nil {
				s.Require().EqualError(err, tc.expectedError.Error())
			} else {
				s.Require().NoError(err)
			}
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Require().True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
			}
		})
	}
}

// delete name record
func (s *MsgServerTestSuite) TestDeleteName() {
	tests := []struct {
		name          string
		expectedError error
		msg           *types.MsgDeleteNameRequest
		expectedEvent proto.Message
	}{
		{
			name:          "delete name record",
			msg:           types.NewMsgDeleteNameRequest(types.NewNameRecord("example.name", s.owner1Addr, false)),
			expectedError: nil,
			expectedEvent: types.NewEventNameUnbound(s.owner1, "example.name", false),
		},
		{
			name:          "create bad name record",
			msg:           types.NewMsgDeleteNameRequest(types.NewNameRecord("example.name", s.owner1Addr, false)),
			expectedError: sdkerrors.ErrInvalidRequest.Wrap("name does not exist"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.DeleteName(s.ctx, tc.msg)
			if tc.expectedError != nil {
				s.Require().EqualError(err, tc.expectedError.Error())
			} else {
				s.Require().NoError(err)
			}
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Require().True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
			}
		})
	}
}

func (s *MsgServerTestSuite) TestModifyName() {
	authority := s.app.NameKeeper.GetAuthority()

	tests := []struct {
		name          string
		expectedError error
		msg           *types.MsgModifyNameRequest
		expectedEvent proto.Message
	}{
		{
			name:          "modify name record, via gov ",
			msg:           types.NewMsgModifyNameRequest(authority, "name", s.owner1Addr, true),
			expectedError: nil,
			expectedEvent: types.NewEventNameUpdate(s.owner1, "name", true),
		},
		{
			name:          "modify name record, via owner",
			msg:           types.NewMsgModifyNameRequest(s.owner1, "name", s.owner1Addr, true),
			expectedError: nil,
			expectedEvent: types.NewEventNameUpdate(s.owner1, "name", true),
		},
		{
			name:          "modify name record with multi level",
			msg:           types.NewMsgModifyNameRequest(authority, "example.name", s.owner1Addr, true),
			expectedError: nil,
			expectedEvent: types.NewEventNameUpdate(s.owner1, "example.name", true),
		},
		{
			name:          "modify name - fails with invalid address",
			msg:           types.NewMsgModifyNameRequest(authority, "name", sdk.AccAddress{}, true),
			expectedError: sdkerrors.ErrInvalidRequest.Wrap("empty address string is not allowed"),
			expectedEvent: nil,
		},
		{
			name:          "modify name - fails with non existent root record",
			msg:           types.NewMsgModifyNameRequest(authority, "jackthecat", s.owner1Addr, true),
			expectedError: sdkerrors.ErrInvalidRequest.Wrap(types.ErrNameNotBound.Error()),
			expectedEvent: nil,
		},
		{
			name:          "modify name - fails with non existent subdomain record",
			msg:           types.NewMsgModifyNameRequest(authority, "jackthecat.name", s.owner1Addr, true),
			expectedError: sdkerrors.ErrInvalidRequest.Wrap(types.ErrNameNotBound.Error()),
			expectedEvent: nil,
		},
		{
			name:          "modify name - fails with invalid authority",
			msg:           types.NewMsgModifyNameRequest("jackthecat", "name", s.owner1Addr, true),
			expectedError: sdkerrors.ErrUnauthorized.Wrapf("expected %s or %s got %s", authority, s.owner1, "jackthecat"),
			expectedEvent: nil,
		},
		{
			name:          "modify name - fails with empty authority",
			msg:           types.NewMsgModifyNameRequest("", "name", s.owner1Addr, true),
			expectedError: sdkerrors.ErrUnauthorized.Wrapf("expected %s or %s got %s", authority, s.owner1, ""),
			expectedEvent: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.ModifyName(s.ctx, tc.msg)
			if tc.expectedError != nil {
				s.Require().EqualError(err, tc.expectedError.Error())
			} else {
				s.Require().NoError(err)
			}
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Require().True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
			}
		})
	}
}

func (s *MsgServerTestSuite) TestCreateRootName() {
	tests := []struct {
		name          string
		expectedError error
		msg           *types.MsgCreateRootNameRequest
		expectedEvent proto.Message
	}{
		{
			name:          "invalid authority",
			msg:           types.NewMsgCreateRootNameRequest("invalid-authority", "example", s.owner1, false),
			expectedError: errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", s.app.NameKeeper.GetAuthority(), "invalid-authority"),
		},
		{
			name:          "valid authority with invalid root name",
			msg:           types.NewMsgCreateRootNameRequest(s.app.NameKeeper.GetAuthority(), "name", s.owner1, false),
			expectedError: nametypes.ErrNameAlreadyBound,
		},
		{
			name:          "valid authority with valid root name",
			msg:           types.NewMsgCreateRootNameRequest(s.app.NameKeeper.GetAuthority(), "example", s.owner1, false),
			expectedEvent: types.NewEventNameBound(s.owner1, "example", false),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.CreateRootName(s.ctx, tc.msg)
			if tc.expectedError != nil {
				s.Require().EqualError(err, tc.expectedError.Error())
			} else {
				s.Require().NoError(err)
			}
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Require().True(result, fmt.Sprintf("Expected typed event was not found: %v", tc.expectedEvent))
			}
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateParams() {
	authority := s.app.NameKeeper.GetAuthority()

	tests := []struct {
		name          string
		expErr        string
		msg           *types.MsgUpdateParamsRequest
		expectedEvent proto.Message
	}{
		{
			name: "valid authority with valid params",
			msg: types.NewMsgUpdateParamsRequest(
				100,
				3,
				10,
				true,
				authority,
			),
			expectedEvent: types.NewEventNameParamsUpdated(
				true,
				10,
				100,
				3,
			),
		},
		{
			name: "invalid authority",
			msg: types.NewMsgUpdateParamsRequest(
				100,
				3,
				10,
				true,
				"invalid-authority",
			),
			expErr: `expected "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" got "invalid-authority": expected gov account as only signer for proposal message`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			_, err := s.msgServer.UpdateParams(s.ctx, tc.msg)
			if len(tc.expErr) > 0 {
				s.Require().EqualError(err, tc.expErr, "Expected error message did not match")
			} else {
				s.Require().NoError(err, "Expected no error but got: %v", err)
			}
			if tc.expectedEvent != nil {
				result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), tc.expectedEvent)
				s.Require().True(result, "Expected typed event was not found: %v", tc.expectedEvent)
			}
		})
	}
}
