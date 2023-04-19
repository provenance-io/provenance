package keeper_test

import (
	"encoding/binary"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
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
	s.msgServer = keeper.NewMsgServerImpl(s.app.NameKeeper)
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
		s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attrtypes.NewAttribute(name, attrAccounts[i].String(), attrtypes.AttributeType_String, []byte(attrAccounts[i].String())), s.owner1Addr))
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
