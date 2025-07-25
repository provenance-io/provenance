package keeper_test

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	startBlockTime time.Time

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.startBlockTime = time.Now()
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: s.startBlockTime})

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("attribute", s.user1Addr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.attribute", s.user1Addr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.empty", s.user1Addr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 3
	nameData.Params.MinSegmentLength = 3
	nameData.Params.MaxSegmentLength = 12

	s.app.NameKeeper.InitGenesis(s.ctx, nameData)

	params := s.app.AttributeKeeper.GetParams(s.ctx)
	params.MaxValueLength = 10
	s.app.AttributeKeeper.SetParams(s.ctx, params)
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
}

func (s *KeeperTestSuite) TestSetAttribute() {
	past := time.Now().Add(-2 * time.Hour)

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
			name: "should fail due to check expiration date",
			attr: types.Attribute{
				Name:           "example.attribute",
				Value:          []byte("01234567891"),
				Address:        s.user1,
				AttributeType:  types.AttributeType_String,
				ExpirationDate: &past,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  fmt.Sprintf("attribute expiration date %v is before block time of %v", past.UTC(), s.ctx.BlockTime().UTC()),
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

func (s *KeeperTestSuite) TestUpdateAttributeExpiration() {
	now := time.Now().UTC()
	past := now.Add(-2 * time.Hour)
	attr := types.Attribute{
		Name:           "example.attribute",
		Value:          []byte("my-value"),
		AttributeType:  types.AttributeType_String,
		ExpirationDate: &now,
		Address:        s.user1,
	}
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")

	cases := []struct {
		name       string
		updateAttr types.Attribute
		ownerAddr  sdk.AccAddress
		errorMsg   string
	}{
		{
			name: "should fail to update attribute expiration, validatebasic original attr",
			updateAttr: types.Attribute{
				Name:           "",
				Value:          []byte("my-value"),
				Address:        s.user1,
				AttributeType:  types.AttributeType_String,
				ExpirationDate: nil,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  `unable to normalize attribute name "": segment of name is too short`,
		},
		{
			name: "should fail to update attribute expiration, value not found",
			updateAttr: types.Attribute{
				Name:           "example.attribute",
				Value:          []byte("notfound"),
				Address:        s.user1,
				AttributeType:  types.AttributeType_String,
				ExpirationDate: nil,
			},
			ownerAddr: s.user1Addr,
			errorMsg:  `no attributes updated with name "example.attribute" : value "notfound" : type: ATTRIBUTE_TYPE_STRING`,
		},
		{
			name: "should fail to update attribute expiration, owner not correct",
			updateAttr: types.Attribute{
				Name:           "example.attribute",
				Value:          []byte("my-value"),
				Address:        s.user1,
				AttributeType:  types.AttributeType_String,
				ExpirationDate: nil,
			},
			ownerAddr: s.user2Addr,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2Addr),
		},
		{
			name: "should fail to update attribute expiration, owner not correct",
			updateAttr: types.Attribute{
				Name:           "example.attribute",
				Value:          []byte("my-value"),
				Address:        s.user1,
				AttributeType:  types.AttributeType_String,
				ExpirationDate: nil,
			},
			ownerAddr: s.user2Addr,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2Addr),
		},
		{
			name: "should fail to update attribute expiration, time in the past",
			updateAttr: types.Attribute{
				Name:           "example.attribute",
				Value:          []byte("my-value"),
				Address:        s.user1,
				AttributeType:  types.AttributeType_String,
				ExpirationDate: &past,
			},
			errorMsg:  fmt.Sprintf("attribute expiration date %v is before block time of %v", past.UTC(), s.ctx.BlockTime().UTC()),
			ownerAddr: s.user1Addr,
		},
		{
			name: "should succeed to update attribute expiration, with nil time",
			updateAttr: types.Attribute{
				Name:           "example.attribute",
				Value:          []byte("my-value"),
				Address:        s.user1,
				AttributeType:  types.AttributeType_String,
				ExpirationDate: nil,
			},
			ownerAddr: s.user1Addr,
		},
		{
			name: "should succeed to update attribute expiration, with non-nil time",
			updateAttr: types.Attribute{
				Name:           "example.attribute",
				Value:          []byte("my-value"),
				Address:        s.user1,
				AttributeType:  types.AttributeType_String,
				ExpirationDate: &now,
			},
			ownerAddr: s.user1Addr,
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			err := s.app.AttributeKeeper.UpdateAttributeExpiration(s.ctx, tc.updateAttr, tc.ownerAddr)
			if len(tc.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().EqualError(err, tc.errorMsg, "UpdateAttributeExpiration")
			} else {
				s.Assert().NoError(err, "UpdateAttributeExpiration")
				attrs, err := s.app.AttributeKeeper.GetAttributes(s.ctx, tc.updateAttr.Address, tc.updateAttr.Name)
				s.Assert().NoError(err, "GetAttributes(%q, %q)", tc.updateAttr.Address, tc.updateAttr.Name)
				s.Assert().Len(attrs, 1, "number of attributes returned by GetAttributes")
				s.Assert().Equal(tc.updateAttr.ExpirationDate, attrs[0].ExpirationDate, "expiration date of attribute returned by GetAttributes")
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
		ConcreteType:  "",
	}
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")

	deletedAttr := types.NewAttribute("deleted", s.user1, types.AttributeType_String, []byte("test"), nil, "")
	// Create a name, make an attribute under it, then remove the name leaving an orphan attribute.
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, "deleted", s.user1Addr, false), "name record should save successfully")
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, deletedAttr, s.user1Addr), "should save successfully")
	s.Require().NoError(s.app.NameKeeper.DeleteRecord(s.ctx, "deleted"), "name record should be removed successfully")

	// Create multiple attributes for a address with same name, to test the delete counter
	deleteCounterName := "deleted2"
	deletedCounterAttr1 := types.NewAttribute(deleteCounterName, s.user1, types.AttributeType_String, []byte("test1"), nil, "")
	deletedCounterAttr2 := types.NewAttribute(deleteCounterName, s.user1, types.AttributeType_String, []byte("test2"), nil, "")
	deletedCounterAttr3 := types.NewAttribute(deleteCounterName, s.user1, types.AttributeType_String, []byte("test3"), nil, "")
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, deleteCounterName, s.user1Addr, false), "name record should save successfully")
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, deletedCounterAttr1, s.user1Addr), "should save successfully")
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, deletedCounterAttr2, s.user1Addr), "should save successfully")
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, deletedCounterAttr3, s.user1Addr), "should save successfully")

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
			errorMsg:  `no keys deleted with name "dne"`,
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
			errorMsg:  `no keys deleted with name "dne" and value "123456789"`,
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
			errorMsg:  `no keys deleted with name "example.attribute" and value "123456789"`,
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
	s.runGetAllAttributesTests("GetAllAttributes", func() ([]types.Attribute, error) {
		return s.app.AttributeKeeper.GetAllAttributes(s.ctx, s.user1)
	})
}

func (s *KeeperTestSuite) TestGetAllAttributesAddr() {
	s.runGetAllAttributesTests("GetAllAttributesAddr", func() ([]types.Attribute, error) {
		return s.app.AttributeKeeper.GetAllAttributesAddr(s.ctx, s.user1Addr)
	})
}

func (s *KeeperTestSuite) runGetAllAttributesTests(funcName string, attrGetter func() ([]types.Attribute, error)) {
	attrs := make([]types.Attribute, 10)
	for i := range attrs {
		attrs[i] = types.Attribute{
			Name:          fmt.Sprintf("example%d.attribute", i),
			Value:         []byte(strings.Repeat(fmt.Sprintf("%d", i), 10)),
			Address:       s.user1,
			AttributeType: types.AttributeType_String,
		}
	}
	attrsSetUp := uint(0)
	// Attributes are keyed using a hash of the name. So they aren't in alphabetical order.
	// The order should never change, though, unless their name changes.
	attrStoreOrder := []uint{2, 7, 9, 6, 8, 1, 3, 5, 0, 4}
	getExpAttrs := func(count uint) []types.Attribute {
		// providing a count instead of just using attrsSetUp so that the size is dictated by the test.
		var rv []types.Attribute
		for _, i := range attrStoreOrder {
			if i < count {
				rv = append(rv, attrs[i])
			}
		}
		return rv
	}

	// These tests build on each other. A setup failure in one will all the rest to fail.
	// So, if a failure happens during setup for a test, fail that test, then skip the rest.
	// But if a non-setup failure happens in a test, we still want the rest to run.
	setupFailed := false
	reqNoErrorSetup := func(t *testing.T, err error, msgAndArgs ...interface{}) {
		if !s.Assert().NoError(err, msgAndArgs...) {
			setupFailed = true
			t.FailNow()
		}
	}

	tests := []struct {
		name  string
		count uint // The count cannot decrease from one test case to the next.
	}{
		{name: "no attributes set", count: 0},
		{name: "1 attribute", count: 1},
		{name: "2 attributes", count: 2},
		{name: "5 attributes", count: 5},
		{name: "5 attributes 2nd time", count: 5},
		{name: "10 attributes", count: 10},
		{name: "10 attributes 2nd time", count: 10},
		{name: "10 attributes 3rd time", count: 10},
		{name: "10 attributes 4th time", count: 10},
		{name: "10 attributes 5th time", count: 10},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if setupFailed {
				s.T().Skipf("skipping due to a setup failure in a previous test")
			}
			if tc.count > uint(len(attrs)) {
				s.FailNowf("Invalid test case.",
					"Not enough test attributes defined.\n\tcount: %3d\n\tmax:   %3d", tc.count, len(attrs))
			}
			for attrsSetUp < tc.count {
				err := s.app.NameKeeper.SetNameRecord(s.ctx, attrs[attrsSetUp].Name, s.user1Addr, false)
				reqNoErrorSetup(s.T(), err, "SetNameRecord attrs[%d]", attrsSetUp)
				err = s.app.AttributeKeeper.SetAttribute(s.ctx, attrs[attrsSetUp], s.user1Addr)
				reqNoErrorSetup(s.T(), err, "SetAttribute attrs[%d]", attrsSetUp)
				attrsSetUp++
			}
			if attrsSetUp > tc.count {
				// Not doing a setup failure here because this error might not cause tests after it to fail.
				s.FailNowf("Invalid test case ordering.",
					"Test case count cannot decrease.\n\tcount: %3d\n\tmin:   %3d", tc.count, attrsSetUp)
			}

			expAttrs := getExpAttrs(tc.count)
			attributes, err := attrGetter()
			s.Assert().NoError(err, "%s error", funcName)
			s.Assert().Equal(expAttrs, attributes, "%s attributes", funcName)
		})
	}
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
		s.app.AttributeKeeper.IncAttrNameAddressLookup(s.ctx, attr.Name, attr.GetAddressBytes())
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
		s.app.AttributeKeeper.DecAttrNameAddressLookup(s.ctx, attr.Name, attr.GetAddressBytes())
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

func (s *KeeperTestSuite) TestPurgeAttributes() {

	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
		ConcreteType:  "",
	}
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")

	deletedAttr := types.NewAttribute("deleted", s.user1, types.AttributeType_String, []byte("test"), nil, "")
	// Create a name, make an attribute under it, then remove the name leaving an orphan attribute.
	s.Assert().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, "deleted", s.user1Addr, false), "name record should save successfully")
	s.Assert().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, deletedAttr, s.user1Addr), "should save successfully")
	s.Assert().NoError(s.app.NameKeeper.DeleteRecord(s.ctx, "deleted"), "name record should be removed successfully")

	// Create multiple attributes for a address with same name, to test the delete counter
	deleteCounterName := "deleted2"
	deletedCounterAttr1 := types.NewAttribute(deleteCounterName, s.user1, types.AttributeType_String, []byte("test1"), nil, "")
	deletedCounterAttr2 := types.NewAttribute(deleteCounterName, s.user1, types.AttributeType_String, []byte("test2"), nil, "")
	deletedCounterAttr3 := types.NewAttribute(deleteCounterName, s.user1, types.AttributeType_String, []byte("test3"), nil, "")
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
			name:        "attribute does not exist should no-op",
			attrName:    "dne",
			beforeCount: 0,
			ownerAddr:   s.user1Addr,
			errorMsg:    "",
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
		s.Run(tc.name, func() {
			attrStore := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
			if len(tc.lookupKey) > 0 {
				count := binary.BigEndian.Uint64(attrStore.Get(tc.lookupKey))
				s.Assert().Equal(tc.beforeCount, count, "should have correct count in lookup table")
			}
			err := s.app.AttributeKeeper.PurgeAttribute(s.ctx, tc.attrName, tc.ownerAddr)
			if len(tc.errorMsg) > 0 {
				s.Assert().Error(err)
				s.Assert().Equal(tc.errorMsg, err.Error())
			} else {
				s.Assert().NoError(err)
				if len(tc.lookupKey) > 0 {
					s.Assert().False(attrStore.Has(tc.lookupKey), "should not have attribute key after deletion")
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestDeleteExpiredAttributes() {
	store := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
	past := s.startBlockTime.Add(-2 * time.Hour)
	future := s.startBlockTime.Add(time.Hour)

	s.ctx = s.ctx.WithBlockTime(past)
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, "one.expire.testing", s.user1Addr, false), "SetNameRecord one.expire.testing")
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, "two.expire.testing", s.user1Addr, false), "SetNameRecord two.expire.testing")
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, "three.expire.testing", s.user1Addr, false), "SetNameRecord three.expire.testing")
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, "four.expire.testing", s.user1Addr, false), "SetNameRecord four.expire.testing")
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, "five.expire.testing", s.user1Addr, false), "SetNameRecord five.expire.testing")

	attr1 := types.NewAttribute("one.expire.testing", s.user1, types.AttributeType_String, []byte("test1"), nil, "")
	attr1.ExpirationDate = &past
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr1, s.user1Addr), "SetAttribute attr1")
	s.Require().NotNil(store.Get(types.AttributeExpireKey(attr1)), "store.Get attr1 AttributeExpireKey")
	s.Require().NotNil(store.Get(types.AttributeNameAddrKeyPrefix(attr1.Name, attr1.GetAddressBytes())), "store.Get attr1 AttributeNameAddrKeyPrefix")

	attr2 := types.NewAttribute("two.expire.testing", s.user1, types.AttributeType_String, []byte("test2"), nil, "")
	attr2.ExpirationDate = &past
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr2, s.user1Addr), "SetAttribute attr2")
	s.Require().NotNil(store.Get(types.AttributeExpireKey(attr2)), "store.Get attr2 AttributeExpireKey")
	s.Require().NotNil(store.Get(types.AttributeNameAddrKeyPrefix(attr2.Name, attr2.GetAddressBytes())), "store.Get attr2 AttributeNameAddrKeyPrefix")

	attr3 := types.NewAttribute("three.expire.testing", s.user1, types.AttributeType_String, []byte("test3"), nil, "")
	attr3.ExpirationDate = &s.startBlockTime
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr3, s.user1Addr), "SetAttribute attr3")
	s.Require().NotNil(store.Get(types.AttributeExpireKey(attr3)), "store.Get attr3 AttributeExpireKey")
	s.Require().NotNil(store.Get(types.AttributeNameAddrKeyPrefix(attr3.Name, attr3.GetAddressBytes())), "store.Get attr3 AttributeNameAddrKeyPrefix")

	attr4 := types.NewAttribute("four.expire.testing", s.user1, types.AttributeType_String, []byte("test4"), nil, "")
	attr4.ExpirationDate = &future
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr4, s.user1Addr), "SetAttribute attr4")
	s.Require().NotNil(store.Get(types.AttributeExpireKey(attr4)), "store.Get attr4 AttributeExpireKey")
	s.Require().NotNil(store.Get(types.AttributeNameAddrKeyPrefix(attr4.Name, attr4.GetAddressBytes())), "store.Get attr4 AttributeNameAddrKeyPrefix")

	attr5 := types.NewAttribute("five.expire.testing", s.user1, types.AttributeType_String, []byte("test5"), nil, "")
	attr5.ExpirationDate = &future
	s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr5, s.user1Addr), "SetAttribute attr5")
	s.Require().NotNil(store.Get(types.AttributeExpireKey(attr5)), "store.Get attr5 AttributeExpireKey")
	s.Require().NotNil(store.Get(types.AttributeNameAddrKeyPrefix(attr5.Name, attr5.GetAddressBytes())), "store.Get attr5 AttributeNameAddrKeyPrefix")

	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager()).WithBlockTime(s.startBlockTime)
	s.app.AttributeKeeper.DeleteExpiredAttributes(s.ctx, 0)

	attr1Event := s.ctx.EventManager().Events()[0]
	attr2Event := s.ctx.EventManager().Events()[1]
	s.Assert().Equal("provenance.attribute.v1.EventAttributeExpired", attr1Event.Type)
	s.Assert().Equal("provenance.attribute.v1.EventAttributeExpired", attr2Event.Type)

	s.Assert().Nil(store.Get(types.AttributeExpireKey(attr1)), "store.Get attr1 AttributeExpireKey")
	s.Assert().Nil(store.Get(types.AttributeNameAddrKeyPrefix(attr1.Name, attr1.GetAddressBytes())), "store.Get attr1 AttributeNameAddrKeyPrefix")
	s.Assert().Nil(store.Get(types.AttributeExpireKey(attr2)), "store.Get attr2 AttributeExpireKey")
	s.Assert().Nil(store.Get(types.AttributeNameAddrKeyPrefix(attr2.Name, attr2.GetAddressBytes())), "store.Get attr2 AttributeNameAddrKeyPrefix")

	s.Assert().NotNil(store.Get(types.AttributeExpireKey(attr3)), "store.Get attr3 AttributeExpireKey")
	s.Assert().NotNil(store.Get(types.AttributeNameAddrKeyPrefix(attr3.Name, attr3.GetAddressBytes())), "store.Get attr3 AttributeNameAddrKeyPrefix")

	s.Assert().NotNil(store.Get(types.AttributeExpireKey(attr4)), "store.Get attr4 AttributeExpireKey")
	s.Assert().NotNil(store.Get(types.AttributeNameAddrKeyPrefix(attr4.Name, attr4.GetAddressBytes())), "store.Get attr4 AttributeNameAddrKeyPrefix")
	s.Assert().NotNil(store.Get(types.AttributeExpireKey(attr5)), "store.Get attr5 AttributeExpireKey")
	s.Assert().NotNil(store.Get(types.AttributeNameAddrKeyPrefix(attr5.Name, attr5.GetAddressBytes())), "store.Get attr5 AttributeNameAddrKeyPrefix")
}

func (s *KeeperTestSuite) TestGetAccountData() {
	params := s.app.AttributeKeeper.GetParams(s.ctx)
	if params.MaxValueLength < 100 {
		defer s.app.AttributeKeeper.SetParams(s.ctx, params)
		params.MaxValueLength = 100
		s.app.AttributeKeeper.SetParams(s.ctx, params)
	}

	withoutAddr := sdk.AccAddress("withoutAddr_________").String()
	withOneAddr := sdk.AccAddress("withOneAddr_________").String()
	withTwoAddr := sdk.AccAddress("withTwoAddr_________").String()

	withOneVal := "This is the value for the address with only one attribute."
	withOneAttr := types.Attribute{
		Name:          types.AccountDataName,
		Value:         []byte(withOneVal),
		AttributeType: types.AttributeType_String,
		Address:       withOneAddr,
	}
	withTwoVal1 := "This is the first of two entries."
	withTwoVal2 := "This is the second of two entries."
	withTwoAttr1 := types.Attribute{
		Name:          types.AccountDataName,
		Value:         []byte(withTwoVal1),
		AttributeType: types.AttributeType_String,
		Address:       withTwoAddr,
	}
	withTwoAttr2 := types.Attribute{
		Name:          types.AccountDataName,
		Value:         []byte(withTwoVal2),
		AttributeType: types.AttributeType_String,
		Address:       withTwoAddr,
	}

	// Use GetModuleAccount to ensure that the account exists.
	attrModAcc := s.app.AccountKeeper.GetModuleAccount(s.ctx, types.ModuleName)
	attrModAddr := attrModAcc.GetAddress()

	err := s.app.AttributeKeeper.SetAttribute(s.ctx, withOneAttr, attrModAddr)
	s.Require().NoError(err, "SetAttribute withOneAttr")
	err = s.app.AttributeKeeper.SetAttribute(s.ctx, withTwoAttr1, attrModAddr)
	s.Require().NoError(err, "SetAttribute withTwoAttr1")
	err = s.app.AttributeKeeper.SetAttribute(s.ctx, withTwoAttr2, attrModAddr)
	s.Require().NoError(err, "SetAttribute withTwoAttr2")

	tests := []struct {
		name   string
		addr   string
		nameK  *mockNameKeeper
		expVal string
		expErr string
	}{
		{name: "no data exists", addr: withoutAddr},
		{name: "one entry", addr: withOneAddr, expVal: withOneVal},
		// Which of these it is depends on how they hash, and withTwoVal1 hashes lower than 2.
		{name: "two entries", addr: withTwoAddr, expVal: withTwoVal1},
		{
			name:   "error getting account attributes",
			addr:   withOneAddr,
			nameK:  newMockNameKeeper(&s.app.NameKeeper).WithGetRecordByNameError("injected error"),
			expErr: `error finding ` + types.AccountDataName + ` for "` + withOneAddr + `": injected error`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			attrKeeper := s.app.AttributeKeeper
			if tc.nameK != nil {
				attrKeeper = attrKeeper.WithNameKeeper(tc.nameK)
			}
			val, err := attrKeeper.GetAccountData(s.ctx, tc.addr)
			if len(tc.expErr) > 0 {
				s.Assert().EqualErrorf(err, tc.expErr, "GetAccountData error")
			} else {
				s.Assert().NoError(err, "GetAccountData error")
			}
			s.Assert().Equal(tc.expVal, val, "GetAccountData value")
		})
	}
}

func (s *KeeperTestSuite) TestSetAccountData() {
	params := s.app.AttributeKeeper.GetParams(s.ctx)
	if params.MaxValueLength < 100 {
		defer s.app.AttributeKeeper.SetParams(s.ctx, params)
		params.MaxValueLength = 100
		s.app.AttributeKeeper.SetParams(s.ctx, params)
	}

	hasNoneAddr := sdk.AccAddress("hasNoneAddr_________").String()
	alreadyHasOneAddr := sdk.AccAddress("alreadyHasOneAddr___").String()
	alreadyHasTwoAddr := sdk.AccAddress("alreadyHasTwoAddr___").String()

	alreadyHasOneAttr := types.Attribute{
		Name:          types.AccountDataName,
		Value:         []byte("alreadyHasOneAttr original value"),
		AttributeType: types.AttributeType_String,
		Address:       alreadyHasOneAddr,
	}
	alreadyHasTwoAttr1 := types.Attribute{
		Name:          types.AccountDataName,
		Value:         []byte("alreadyHasTwoAddr first original value"),
		AttributeType: types.AttributeType_String,
		Address:       alreadyHasTwoAddr,
	}
	alreadyHasTwoAttr2 := types.Attribute{
		Name:          types.AccountDataName,
		Value:         []byte("alreadyHasTwoAddr second original value"),
		AttributeType: types.AttributeType_String,
		Address:       alreadyHasTwoAddr,
	}

	// Use GetModuleAccount to ensure that the account exists.
	attrModAcc := s.app.AccountKeeper.GetModuleAccount(s.ctx, types.ModuleName)
	attrModAddr := attrModAcc.GetAddress()

	err := s.app.AttributeKeeper.SetAttribute(s.ctx, alreadyHasOneAttr, attrModAddr)
	s.Require().NoError(err, "SetAttribute alreadyHasOneAttr")
	err = s.app.AttributeKeeper.SetAttribute(s.ctx, alreadyHasTwoAttr1, attrModAddr)
	s.Require().NoError(err, "SetAttribute alreadyHasTwoAttr1")
	err = s.app.AttributeKeeper.SetAttribute(s.ctx, alreadyHasTwoAttr2, attrModAddr)
	s.Require().NoError(err, "SetAttribute alreadyHasTwoAttr2")

	tests := []struct {
		name  string
		addr  string
		value string
		// I've no idea how to make any of GetAttributes, DeleteAttribute, or SetAttribute error.
		// Those are the only error conditions, so the expected error is omitted as a test param.
	}{
		{
			name:  "does not yet have data",
			addr:  hasNoneAddr,
			value: "This is a new value for hasNoneAddr.",
		},
		{
			name:  "overwrites existing entry",
			addr:  alreadyHasOneAddr,
			value: "This is a new value for alreadyHasOneAddr.",
		},
		{
			name:  "extra entries deleted",
			addr:  alreadyHasTwoAddr,
			value: "This is now the one and only entry for alreadyHasTwoAddr.",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			err = s.app.AttributeKeeper.SetAccountData(ctx, tc.addr, tc.value)
			s.Require().NoError(err, "SetAccountData")

			// Make sure there's exactly one attribute now and it has the expected value.
			attrs, err := s.app.AttributeKeeper.GetAttributes(s.ctx, tc.addr, types.AccountDataName)
			s.Require().NoError(err, "GetAttributes(%q) after SetAccountData", types.AccountDataName)
			if s.Assert().Len(attrs, 1, "attributes after SetAccountData") {
				s.Assert().Equal(types.AccountDataName, attrs[0].Name, "attribute Name")
				s.Assert().Equal(tc.addr, attrs[0].Address, "attribute Address")
				s.Assert().Equal(types.AttributeType_String, attrs[0].AttributeType, "attribute AttributeType")
				s.Assert().Equal(tc.value, string(attrs[0].Value), "attribute Value")
			}

			// Make sure the last event emitted is about updating account data.
			events := em.Events()
			if s.Assert().GreaterOrEqual(len(events), 1, "events emitted during SetAccountData") {
				event := events[len(events)-1]
				s.Assert().Contains(event.Type, "EventAccountDataUpdated", "event type")
				if s.Assert().Len(event.Attributes, 1, "event attributes") {
					s.Assert().Equal("account", event.Attributes[0].Key, "attribute key")
					s.Assert().Equal(`"`+tc.addr+`"`, event.Attributes[0].Value, "attribute value")
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestAttributeWithConcreteType() {
	testCases := []struct {
		name          string
		attributeName string
		attributeType types.AttributeType
		ownerAddr     sdk.Address
		value         []byte
		concreteType  string
		expectError   bool
	}{
		{
			name:          "should successfully JSON attribute with concrete type",
			attributeName: "example.attribute",
			attributeType: types.AttributeType_JSON,
			value:         []byte(`"123456"`),
			concreteType:  "provenance.attributes.v1.TestJson",
			ownerAddr:     s.user1Addr,
			expectError:   false,
		},
		{
			name:          "should successfully PROTO attribute with concrete type",
			attributeName: "attribute",
			attributeType: types.AttributeType_Proto,
			ownerAddr:     s.user1Addr,
			value:         []byte("2"), // Some proto encoded bytes
			concreteType:  "provenance.attributes.v1.TestProto",
			expectError:   false,
		},
		{
			name:          "should successfully STRING attribute with concrete type",
			attributeName: "example.attribute",
			attributeType: types.AttributeType_String,
			ownerAddr:     s.user1Addr,
			value:         []byte("teststring"),
			concreteType:  "provenance.attributes.v1.TestString",
			expectError:   false,
		},
		{
			name:          "should successfully Empty concrete type",
			attributeName: "example.empty",
			attributeType: types.AttributeType_JSON,
			ownerAddr:     s.user1Addr,
			value:         []byte("3"),
			concreteType:  "",
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {

			defer func() {
				if r := recover(); r != nil {
					s.Fail(fmt.Sprintf("panic occurred: %v", r))
				}
			}()
			attr := types.Attribute{
				Name:          tc.attributeName,
				AttributeType: tc.attributeType,
				Value:         tc.value,
				Address:       tc.ownerAddr.String(),
				ConcreteType:  tc.concreteType,
			}

			err := s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr)
			if tc.expectError {
				s.Assert().Error(err, "SetAttribute should return an error when expectError is true")
				return
			}
			s.Assert().NoError(err, "SetAttribute should not return an error when expectError is false")
			savedAttr, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1Addr.String(), tc.attributeName)
			if tc.expectError {
				s.Assert().Error(err, "GetAttributes should return an error when expectError is true")
				return
			}
			s.Assert().NoError(err, "s.app.AttributeKeeper.GetAttributes")
			s.Assert().Equal(tc.attributeName, savedAttr[0].Name, "Attribute Name mismatch")
			s.Assert().Equal(tc.attributeType, savedAttr[0].AttributeType, "Attribute Type mismatch")
			s.Assert().Equal(tc.value, savedAttr[0].Value, "Attribute Value mismatch")
			s.Assert().Equal(s.user1Addr.String(), savedAttr[0].Address, "Attribute Address mismatch")
			s.Assert().Equal(tc.concreteType, savedAttr[0].ConcreteType, "Attribute ConcreteType mismatch")
		})
	}
}

func (s *KeeperTestSuite) TestUpdateAttributeConcreteType() {
	// Set initial attribute
	initialAttr := types.Attribute{
		Name:          "example.attribute",
		AttributeType: types.AttributeType_JSON,
		Value:         []byte(`{"ver":1}`),
		Address:       s.user1,
		ConcreteType:  "/provenance.attributes.v1.VersionOne",
	}

	err := s.app.AttributeKeeper.SetAttribute(s.ctx, initialAttr, s.user1Addr)
	s.Assert().NoError(err)

	// Update just the concrete type
	updatedAttr := types.Attribute{
		Name:          "example.attribute",
		AttributeType: types.AttributeType_JSON,
		Value:         []byte(`{"ver":1}`),
		Address:       s.user1,
		ConcreteType:  "/provenance.attributes.v1.VersionTwo",
	}

	err = s.app.AttributeKeeper.SetAttribute(s.ctx, updatedAttr, s.user1Addr)
	s.Assert().NoError(err)

	// Verify update
	savedAttr, _ := s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1, updatedAttr.Name)
	s.Assert().Equal(updatedAttr.Value, savedAttr[0].Value)
	s.Assert().Equal(updatedAttr.ConcreteType, savedAttr[0].ConcreteType)

	// Update value and remove concrete type
	noTypeAttr := types.Attribute{
		Name:          "example.attribute",
		AttributeType: types.AttributeType_JSON,
		Value:         []byte(`{"ver":1}`),
		Address:       s.user1,
		ConcreteType:  "",
	}

	err = s.app.AttributeKeeper.SetAttribute(s.ctx, noTypeAttr, s.user1Addr)
	s.Assert().NoError(err)

	// Verify update removed concrete type
	savedAttr, _ = s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1, noTypeAttr.Name)
	s.Assert().Equal(noTypeAttr.Value, savedAttr[0].Value)

}

func (s *KeeperTestSuite) TestValidateAttributeConcreteType() {
	testCases := []struct {
		name         string
		concreteType string
		expErr       string
	}{
		{
			name:         "Valid type URL",
			concreteType: "provenance.attributes.v1.KnowYourCustomer",
			expErr:       "", // No error expected
		},
		{
			name:         "Empty concrete type",
			concreteType: "",
			expErr:       "", // No error expected
		},
		{
			name:         "Maximum length",
			concreteType: strings.Repeat("a", 200),
			expErr:       "", // No error expected
		},
		{
			name:         "Exceeds maximum length",
			concreteType: strings.Repeat("a", 201),
			expErr:       "concrete_type exceeds maximum length of 200 characters",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			attr := types.Attribute{
				Name:          "test",
				AttributeType: types.AttributeType_JSON,
				Value:         []byte(`{}`),
				Address:       s.user1,
				ConcreteType:  tc.concreteType,
			}

			err := attr.ValidateBasic()
			if tc.expErr == "" {
				s.Assert().NoError(err, "Expected no error for concreteType: %q", tc.concreteType)
			} else {
				s.Assert().ErrorContains(err, tc.expErr, "Expected error message mismatch for concreteType: %q", tc.concreteType)
			}
		})
	}
}
