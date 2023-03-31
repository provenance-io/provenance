package keeper_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/attribute/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func TestKeeperTestSuite(t *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(s.T())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	s.app = app
	s.ctx = ctx

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("attribute", s.user1Addr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.attribute", s.user1Addr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 3
	nameData.Params.MinSegmentLength = 3
	nameData.Params.MaxSegmentLength = 12

	app.NameKeeper.InitGenesis(ctx, nameData)

	params := app.AttributeKeeper.GetParams(ctx)
	params.MaxValueLength = 10
	app.AttributeKeeper.SetParams(ctx, params)
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
}

func (s *KeeperTestSuite) TestSetAttribute() {

	cases := []struct {
		name        string
		attr        types.Attribute
		ownerAddr   sdk.AccAddress
		errorMsg    string
		lookupCount uint64
	}{
		{
			name: "should successfully add attribute",
			attr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("1"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr:   s.user1Addr,
			errorMsg:    "",
			lookupCount: 1,
		},
		{
			name: "should successfully add attribute with same name but different type",
			attr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("1"),
				Address:       s.user1,
				AttributeType: types.AttributeType_Int,
			},
			ownerAddr:   s.user1Addr,
			errorMsg:    "",
			lookupCount: 2,
		},
		{
			name: "should successfully add attribute with same name and type but diff value",
			attr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("2"),
				Address:       s.user1,
				AttributeType: types.AttributeType_Int,
			},
			ownerAddr:   s.user1Addr,
			errorMsg:    "",
			lookupCount: 3,
		},
		{
			name: "should fail due to validate basic error",
			attr: types.Attribute{
				Name:          "",
				Value:         []byte("01234567891"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "invalid name: empty",
		},
		{
			name: "should fail due to attribute length too long",
			attr: types.Attribute{
				Name:          "name",
				Value:         []byte("01234567891"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "attribute value length of 11 exceeds max length 10",
		},
		{
			name: "should fail unable to find owner",
			attr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user2Addr,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2),
		},
		{
			name: "should fail unable to normalize segment length too short",
			attr: types.Attribute{
				Name:          "example.cant.normalize.me",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user2Addr,
			errorMsg:  "unable to normalize attribute name \"example.cant.normalize.me\": segment of name is too short",
		},
		{
			name: "should fail unable to resolve name to user",
			attr: types.Attribute{
				Name:          "example.not.found",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  fmt.Sprintf("\"example.not.found\" does not resolve to address \"%s\"", s.user1),
		},
	}
	for _, tc := range cases {
		tc := tc

		s.Run(tc.name, func() {
			err := s.app.AttributeKeeper.SetAttribute(s.ctx, tc.attr, tc.ownerAddr)
			if len(tc.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tc.errorMsg, err.Error())
			} else {
				s.Assert().NoError(err)
				lookup, err := s.app.AttributeKeeper.AccountsByAttribute(s.ctx, tc.attr.Name)
				s.Assert().NoError(err, "should have attribute in lookup table")
				s.Assert().Len(lookup, 1)
				s.Assert().Equal(tc.attr.GetAddressBytes(), lookup[0].Bytes())
				attrStore := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
				value := attrStore.Get(types.AttributeNameAddrKeyPrefix(tc.attr.Name, tc.attr.GetAddressBytes()))
				s.Assert().Equal(tc.lookupCount, binary.BigEndian.Uint64(value))
			}
		})
	}

}

func (s *KeeperTestSuite) TestUpdateAttribute() {

	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("my-value"),
		AttributeType: types.AttributeType_String,
		Address:       s.user1,
	}
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")

	cases := []struct {
		name       string
		origAttr   types.Attribute
		updateAttr types.Attribute
		ownerAddr  sdk.AccAddress
		errorMsg   string
	}{
		{
			name: "should fail to update attribute, validatebasic original attr",
			origAttr: types.Attribute{
				Name:          "",
				Value:         []byte("my-value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			updateAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("10"),
				Address:       s.user1,
				AttributeType: types.AttributeType_Int,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "invalid name: empty",
		}, {
			name: "should fail to update attribute, validatebasic update attr",
			origAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("my-value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			updateAttr: types.Attribute{
				Name:          "",
				Value:         []byte("10"),
				Address:       s.user1,
				AttributeType: types.AttributeType_Int,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "invalid name: empty",
		},
		{
			name: "should fail to update attribute, names mismatch",
			origAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("my-value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			updateAttr: types.Attribute{
				Name:          "example.noteq",
				Value:         []byte("10"),
				Address:       s.user1,
				AttributeType: types.AttributeType_Int,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "update and original names must match example.noteq : example.attribute",
		},
		{
			name: "should fail to update attribute, length too long",
			origAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("my-value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			updateAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("0123456789123"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "update attribute value length of 13 exceeds max length 10",
		},
		{
			name: "should fail to update attribute, unable to find owner",
			origAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("my-value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			updateAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("new string"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user2Addr,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2Addr),
		},
		{
			name: "should fail to update attribute, unable to resolve name",
			origAttr: types.Attribute{
				Name:          "example.not.found",
				Value:         []byte("my-value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			updateAttr: types.Attribute{
				Name:          "example.not.found",
				Value:         []byte("new value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  fmt.Sprintf("\"example.not.found\" does not resolve to address \"%s\"", s.user1Addr),
		},
		{
			name: "should fail to update attribute, to find original, no original value match",
			origAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("not original value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			updateAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "no attributes updated with name \"example.attribute\" : value \"not original value\" : type: ATTRIBUTE_TYPE_STRING",
		},
		{
			name: "should fail to update attribute, to find original, no original attribute type match",
			origAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("my-value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_Bytes,
			},
			updateAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "no attributes updated with name \"example.attribute\" : value \"my-value\" : type: ATTRIBUTE_TYPE_BYTES",
		},
		{
			name: "should successfully update attribute",
			origAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("my-value"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			updateAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("10"),
				Address:       s.user1,
				AttributeType: types.AttributeType_Int,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "",
		},
		{
			name: "should successfully update attribute changing user address",
			origAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("10"),
				Address:       s.user1,
				AttributeType: types.AttributeType_Int,
			},
			updateAttr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("10"),
				Address:       s.user2,
				AttributeType: types.AttributeType_Int,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  "",
		},
	}
	for _, tc := range cases {
		tc := tc

		s.Run(tc.name, func() {
			err := s.app.AttributeKeeper.UpdateAttribute(s.ctx, tc.origAttr, tc.updateAttr, tc.ownerAddr)
			if len(tc.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tc.errorMsg, err.Error())
			} else {
				s.Assert().NoError(err)
				attrStore := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
				if tc.origAttr.Address != tc.updateAttr.Address {
					s.Assert().False(attrStore.Has(types.AttributeNameAddrKeyPrefix(tc.origAttr.Name, tc.origAttr.GetAddressBytes())), "original key should have been removed")
				}
				s.Assert().True(attrStore.Has(types.AttributeNameAddrKeyPrefix(tc.updateAttr.Name, tc.updateAttr.GetAddressBytes())), "updated key should have been added")
			}
		})
	}

}

func (s *KeeperTestSuite) TestDeleteAttribute() {

	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")

	deletedAttr := types.NewAttribute("deleted", s.user1, types.AttributeType_String, []byte("test"))
	// Create a name, make an attribute under it, then remove the name leaving an orphan attribute.
	s.Assert().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, "deleted", s.user1Addr, false), "name record should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, deletedAttr, s.user1Addr), "should save successfully")
	s.Assert().NoError(s.app.NameKeeper.DeleteRecord(s.ctx, "deleted"), "name record should be removed successfully")

	// Create multiple attributes for a address with same name, to test the delete counter
	deleteCounterName := "deleted2"
	deletedCounterAttr1 := types.NewAttribute(deleteCounterName, s.user1, types.AttributeType_String, []byte("test1"))
	deletedCounterAttr2 := types.NewAttribute(deleteCounterName, s.user1, types.AttributeType_String, []byte("test2"))
	deletedCounterAttr3 := types.NewAttribute(deleteCounterName, s.user1, types.AttributeType_String, []byte("test3"))
	s.Assert().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, deleteCounterName, s.user1Addr, false), "name record should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, deletedCounterAttr1, s.user1Addr), "should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, deletedCounterAttr2, s.user1Addr), "should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, deletedCounterAttr3, s.user1Addr), "should save successfully")

	cases := []struct {
		name        string
		attrName    string
		accAddr     string
		beforeCount uint64
		lookupKey   []byte
		ownerAddr   sdk.AccAddress
		errorMsg    string
	}{
		{
			name:      "should fail to delete, cant resolve name wrong owner",
			attrName:  "example.attribute",
			ownerAddr: s.user2Addr,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2Addr),
		},
		{
			name:      "no keys will be deleted with unknown name",
			attrName:  "dne",
			ownerAddr: s.user1Addr,
			errorMsg:  "no keys deleted with name dne",
		},
		{
			name:        "attribute will be removed without error when name has been removed",
			attrName:    deletedAttr.Name,
			beforeCount: 1,
			lookupKey:   types.AttributeNameAddrKeyPrefix(deletedAttr.Name, deletedAttr.GetAddressBytes()),
			accAddr:     s.user1,
			ownerAddr:   s.user1Addr,
			errorMsg:    "",
		},
		{
			name:        "should successfully delete attribute",
			attrName:    attr.Name,
			beforeCount: 1,
			lookupKey:   types.AttributeNameAddrKeyPrefix(attr.Name, attr.GetAddressBytes()),
			accAddr:     s.user1,
			ownerAddr:   s.user1Addr,
			errorMsg:    "",
		},
		{
			name:        "should successfully delete multiple attributes",
			attrName:    deleteCounterName,
			beforeCount: 3,
			lookupKey:   types.AttributeNameAddrKeyPrefix(deleteCounterName, deletedCounterAttr1.GetAddressBytes()),
			accAddr:     s.user1,
			ownerAddr:   s.user1Addr,
			errorMsg:    "",
		},
	}
	for _, tc := range cases {
		tc := tc

		s.Run(tc.name, func() {
			attrStore := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
			if len(tc.lookupKey) > 0 {
				count := binary.BigEndian.Uint64(attrStore.Get(tc.lookupKey))
				s.Assert().Equal(tc.beforeCount, count, "should have correct count in lookup table")
			}
			err := s.app.AttributeKeeper.DeleteAttribute(s.ctx, tc.accAddr, tc.attrName, nil, tc.ownerAddr)
			if len(tc.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tc.errorMsg, err.Error())
			} else {
				s.Assert().NoError(err)
				s.Assert().False(attrStore.Has(tc.lookupKey), "should not have attribute key after deletion")
			}
		})
	}
}

func (s *KeeperTestSuite) TestDeleteDistinctAttribute() {

	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	attrValue := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("diff value"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	s.NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
	s.NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attrValue, s.user1Addr), "should save successfully")

	cases := []struct {
		testName     string
		name         string
		value        []byte
		attrType     types.AttributeType
		accAddr      string
		ownerAddr    sdk.AccAddress
		expectlookup bool
		errorMsg     string
	}{
		{
			testName:  "should fail to delete, cant resolve name wrong owner",
			name:      "example.attribute",
			value:     []byte("123456789"),
			ownerAddr: s.user2Addr,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2Addr),
		},
		{
			testName:  "no keys will be deleted with unknown name",
			name:      "dne",
			value:     []byte("123456789"),
			ownerAddr: s.user1Addr,
			errorMsg:  "no keys deleted with name dne value 123456789",
		},
		{
			testName:     "should successfully delete attribute",
			name:         "example.attribute",
			value:        []byte("123456789"),
			accAddr:      s.user1,
			ownerAddr:    s.user1Addr,
			expectlookup: true,
			errorMsg:     "",
		},
		{
			testName:  "should fail to delete attribute, was already deleted",
			name:      "example.attribute",
			value:     []byte("123456789"),
			accAddr:   s.user1,
			ownerAddr: s.user1Addr,
			errorMsg:  "no keys deleted with name example.attribute value 123456789",
		},
		{
			testName:     "should successfully delete attribute, with same key but different value",
			name:         "example.attribute",
			value:        []byte("diff value"),
			accAddr:      s.user1,
			ownerAddr:    s.user1Addr,
			expectlookup: false,
			errorMsg:     "",
		},
	}
	for _, tc := range cases {
		tc := tc

		s.Run(tc.testName, func() {
			err := s.app.AttributeKeeper.DeleteAttribute(s.ctx, tc.accAddr, tc.name, &tc.value, tc.ownerAddr)
			if len(tc.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tc.errorMsg, err.Error())
			} else {
				s.Assert().NoError(err)
				lookupAccts, err := s.app.AttributeKeeper.AccountsByAttribute(s.ctx, tc.name)
				s.Assert().NoError(err)
				if tc.expectlookup {
					s.Assert().ElementsMatch([]sdk.AccAddress{tc.ownerAddr}, lookupAccts)
				} else {
					s.Assert().Len(lookupAccts, 0)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetAllAttributes() {
	attributes, err := s.app.AttributeKeeper.GetAllAttributes(s.ctx, s.user1)
	s.Assert().NoError(err)
	s.Assert().Equal(0, len(attributes))

	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
	attributes, err = s.app.AttributeKeeper.GetAllAttributes(s.ctx, s.user1)
	s.Assert().NoError(err)
	s.Assert().Equal(attr.Name, attributes[0].Name)
	s.Assert().Equal(attr.Address, attributes[0].Address)
	s.Assert().Equal(attr.Value, attributes[0].Value)
}

func (s *KeeperTestSuite) TestIncAndDecAddNameAddressLookup() {
	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	lookupKey := types.AttributeNameAddrKeyPrefix(attr.Name, attr.GetAddressBytes())

	store := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
	bz := store.Get(lookupKey)
	s.Assert().Nil(bz)

	for i := 1; i <= 100; i++ {
		s.app.AttributeKeeper.IncAddNameAddressLookup(s.ctx, attr)
		bz := store.Get(lookupKey)
		s.Assert().NotNil(bz)
		s.Assert().Equal(uint64(i), binary.BigEndian.Uint64(bz))
	}

	bz = store.Get(lookupKey)
	s.Assert().Equal(uint64(100), binary.BigEndian.Uint64(bz))

	for i := 100; i > 0; i-- {
		bz := store.Get(lookupKey)
		s.Assert().NotNil(bz)
		s.Assert().Equal(uint64(i), binary.BigEndian.Uint64(bz))
		s.app.AttributeKeeper.DecAddNameAddressLookup(s.ctx, attr)
	}

	store = s.ctx.KVStore(s.app.GetKey(types.StoreKey))
	bz = store.Get(lookupKey)
	s.Assert().Nil(bz)

}

func (s *KeeperTestSuite) TestGetAttributesByName() {

	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
	_, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1, "blah")
	s.Assert().Error(err)
	s.Assert().Equal("no address bound to name", err.Error())
	attributes, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1, "example.attribute")
	s.Assert().NoError(err)
	s.Assert().Equal(1, len(attributes))
	s.Assert().Equal(attr.Name, attributes[0].Name)
	s.Assert().Equal(attr.Address, attributes[0].Address)
	s.Assert().Equal(attr.Value, attributes[0].Value)
}

func (s *KeeperTestSuite) TestGetAddressesByName() {

	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
	addrs, err := s.app.AttributeKeeper.AccountsByAttribute(s.ctx, attr.Name)
	s.Assert().NoError(err)
	s.Assert().ElementsMatch([]sdk.AccAddress{s.user1Addr}, addrs)

	attr.Address = s.user2
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
	addrs, err = s.app.AttributeKeeper.AccountsByAttribute(s.ctx, attr.Name)
	s.Assert().NoError(err)
	s.Assert().ElementsMatch([]sdk.AccAddress{s.user1Addr, s.user2Addr}, addrs)

	attr.AttributeType = types.AttributeType_Int
	attr.Value = []byte("1")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
	addrs, err = s.app.AttributeKeeper.AccountsByAttribute(s.ctx, attr.Name)
	s.Assert().NoError(err)
	s.Assert().ElementsMatch([]sdk.AccAddress{s.user1Addr, s.user2Addr}, addrs)
}

func (s *KeeperTestSuite) TestInitGenesisAddingAttributes() {
	genAttr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	var attributeData types.GenesisState
	attributeData.Attributes = append(attributeData.Attributes, genAttr)
	s.Assert().NotPanics(func() { s.app.AttributeKeeper.InitGenesis(s.ctx, &attributeData) })
	attrs, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1, genAttr.Name)
	s.Assert().NoError(err)
	s.Assert().Contains(attrs, genAttr)
	accts, err := s.app.AttributeKeeper.AccountsByAttribute(s.ctx, genAttr.Name)
	s.Assert().NoError(err)
	s.Assert().ElementsMatch([]sdk.AccAddress{s.user1Addr}, accts)

	s.Assert().NotPanics(func() { s.app.AttributeKeeper.ExportGenesis(s.ctx) })
	attributeData.Attributes = append(attributeData.Attributes, types.Attribute{
		Name:          "",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	})

	s.Assert().Panics(func() { s.app.AttributeKeeper.InitGenesis(s.ctx, &attributeData) })
}

func (s *KeeperTestSuite) TestIterateRecord() {
	s.Run("iterate attribute's", func() {
		attr := types.Attribute{
			Name:          "example.attribute",
			Value:         []byte("0123456789"),
			Address:       s.user1,
			AttributeType: types.AttributeType_String,
		}
		s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
		records := []types.Attribute{}
		// Callback func that adds records to genesis state.
		appendToRecords := func(record types.Attribute) error {
			records = append(records, record)
			return nil
		}
		// Collect and return genesis state.
		err := s.app.AttributeKeeper.IterateRecords(s.ctx, types.AttributeKeyPrefix, appendToRecords)
		s.Require().NoError(err)
		s.Require().Equal(1, len(records))
	})

}

func (s *KeeperTestSuite) TestPopulateAddressAttributeNameTable() {
	store := s.ctx.KVStore(s.app.GetKey(types.StoreKey))

	example1Attr := "example.one"
	exampleAttr1 := types.NewAttribute(example1Attr, s.user1, types.AttributeType_String, []byte("test1"))
	exampleAttr2 := types.NewAttribute(example1Attr, s.user1, types.AttributeType_String, []byte("test2"))
	exampleAttr3 := types.NewAttribute(example1Attr, s.user1, types.AttributeType_String, []byte("test3"))
	exampleAttr4 := types.NewAttribute(example1Attr, s.user2, types.AttributeType_String, []byte("test4"))
	s.Assert().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, example1Attr, s.user1Addr, false), "name record should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, exampleAttr1, s.user1Addr), "should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, exampleAttr2, s.user1Addr), "should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, exampleAttr3, s.user1Addr), "should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, exampleAttr4, s.user1Addr), "should save successfully")

	example2Attr := "example.two"
	example2Attr1 := types.NewAttribute(example2Attr, s.user1, types.AttributeType_String, []byte("test1"))
	example2Attr2 := types.NewAttribute(example2Attr, s.user2, types.AttributeType_String, []byte("test2"))
	s.Assert().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, example2Attr, s.user1Addr, false), "name record should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, example2Attr1, s.user1Addr), "should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, example2Attr2, s.user1Addr), "should save successfully")

	// Clear the kv store of all address look up prefixes
	// This is because the SetAttribute call would have populated it in the test setup
	it := sdk.KVStorePrefixIterator(store, types.AttributeKeyPrefixAddrLookup)
	for ; it.Valid(); it.Next() {
		store.Delete(it.Key())
	}

	s.app.AttributeKeeper.PopulateAddressAttributeNameTable(s.ctx)

	lookupKey := types.AttributeNameAddrKeyPrefix(example1Attr, s.user1Addr)
	bz := store.Get(lookupKey)
	s.Assert().NotNil(bz)
	s.Assert().Equal(uint64(3), binary.BigEndian.Uint64(bz))

	lookupKey = types.AttributeNameAddrKeyPrefix(example1Attr, s.user2Addr)
	bz = store.Get(lookupKey)
	s.Assert().NotNil(bz)
	s.Assert().Equal(uint64(1), binary.BigEndian.Uint64(bz))

	lookupKey = types.AttributeNameAddrKeyPrefix(example2Attr, s.user1Addr)
	bz = store.Get(lookupKey)
	s.Assert().NotNil(bz)
	s.Assert().Equal(uint64(1), binary.BigEndian.Uint64(bz))

	lookupKey = types.AttributeNameAddrKeyPrefix(example2Attr, s.user2Addr)
	bz = store.Get(lookupKey)
	s.Assert().NotNil(bz)
	s.Assert().Equal(uint64(1), binary.BigEndian.Uint64(bz))
}
