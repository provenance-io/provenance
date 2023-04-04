package keeper_test

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	simapp "github.com/provenance-io/provenance/app"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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

func (s *MsgServerTestSuite) TestDeleteNameRemovingAttributeAccounts() {
	name := "test.io"
	s.Require().Error(s.app.NameKeeper.SetNameRecord(s.ctx, name, s.owner1Addr, false))
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

	result, err := s.msgServer.DeleteName(s.ctx, types.NewMsgDeleteNameRequest(types.NewNameRecord("test.io", s.owner1Addr, false)))
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
