package keeper_test

import (
	"fmt"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

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
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
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

	cases := map[string]struct {
		attr      types.Attribute
		ownerAddr sdk.AccAddress
		wantErr   bool
		errorMsg  string
	}{
		"should successfully add attribute": {
			attr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			wantErr:   false,
			errorMsg:  "",
		},
		"should fail due to validate basic error": {
			attr: types.Attribute{
				Name:          "",
				Value:         []byte("01234567891"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  "invalid name: empty",
		},
		"should fail due to attribute length too long": {
			attr: types.Attribute{
				Name:          "name",
				Value:         []byte("01234567891"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  "attribute value length of 11 exceeds max length 10",
		},
		"should fail unable to find owner": {
			attr: types.Attribute{
				Name:          "example.attribute",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user2Addr,
			wantErr:   true,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2),
		},
		"should fail unable to normalize segment length too short": {
			attr: types.Attribute{
				Name:          "example.cant.normalize.me",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user2Addr,
			wantErr:   true,
			errorMsg:  "unable to normalize attribute name \"example.cant.normalize.me\": segment of name is too short",
		},
		"should fail unable to resolve name to user": {
			attr: types.Attribute{
				Name:          "example.not.found",
				Value:         []byte("0123456789"),
				Address:       s.user1,
				AttributeType: types.AttributeType_String,
			},
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  fmt.Sprintf("\"example.not.found\" does not resolve to address \"%s\"", s.user1),
		},
	}
	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.AttributeKeeper.SetAttribute(s.ctx, tc.attr, tc.ownerAddr)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
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
	s.NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")

	cases := []struct {
		name       string
		origAttr   types.Attribute
		updateAttr types.Attribute
		ownerAddr  sdk.AccAddress
		wantErr    bool
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
			wantErr:   true,
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
			wantErr:   true,
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
			wantErr:   true,
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
			wantErr:   true,
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
			wantErr:   true,
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
			wantErr:   true,
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
			wantErr:   true,
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
			wantErr:   true,
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
			wantErr:   false,
			errorMsg:  "",
		},
	}
	for _, tc := range cases {
		tc := tc

		s.Run(tc.name, func() {
			err := s.app.AttributeKeeper.UpdateAttribute(s.ctx, tc.origAttr, tc.updateAttr, tc.ownerAddr)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
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
	s.NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")

	// Create a name, make an attribute under it, then remove the name leaving an orphan attribute.
	s.NoError(s.app.NameKeeper.SetNameRecord(s.ctx, "deleted", s.user1Addr, false), "name record should save successfully")
	s.NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, types.NewAttribute("deleted", s.user1, types.AttributeType_String, []byte("test")), s.user1Addr), "should save successfully")
	s.NoError(s.app.NameKeeper.DeleteRecord(s.ctx, "deleted"), "name record should be removed successfully")

	cases := map[string]struct {
		name      string
		accAddr   string
		ownerAddr sdk.AccAddress
		wantErr   bool
		errorMsg  string
	}{
		"should fail to delete, cant resolve name wrong owner": {
			name:      "example.attribute",
			ownerAddr: s.user2Addr,
			wantErr:   true,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2Addr),
		},
		"no keys will be deleted with unknown name": {
			name:      "dne",
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  "no keys deleted with name dne",
		},
		"attribute will be removed without error when name has been removed": {
			name:      "deleted",
			accAddr:   s.user1,
			ownerAddr: s.user1Addr,
			wantErr:   false,
			errorMsg:  "",
		},
		"should successfully delete attribute": {
			name:      "example.attribute",
			accAddr:   s.user1,
			ownerAddr: s.user1Addr,
			wantErr:   false,
			errorMsg:  "",
		},
	}
	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.AttributeKeeper.DeleteAttribute(s.ctx, tc.accAddr, tc.name, nil, tc.ownerAddr)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
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
		testName  string
		name      string
		value     []byte
		attrType  types.AttributeType
		accAddr   string
		ownerAddr sdk.AccAddress
		wantErr   bool
		errorMsg  string
	}{
		{
			testName:  "should fail to delete, cant resolve name wrong owner",
			name:      "example.attribute",
			value:     []byte("123456789"),
			ownerAddr: s.user2Addr,
			wantErr:   true,
			errorMsg:  fmt.Sprintf("no account found for owner address \"%s\"", s.user2Addr),
		},
		{
			testName:  "no keys will be deleted with unknown name",
			name:      "dne",
			value:     []byte("123456789"),
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  "no keys deleted with name dne value 123456789",
		},
		{
			testName:  "should successfully delete attribute",
			name:      "example.attribute",
			value:     []byte("123456789"),
			accAddr:   s.user1,
			ownerAddr: s.user1Addr,
			wantErr:   false,
			errorMsg:  "",
		},
		{
			testName:  "should fail to delete attribute, was already deleted",
			name:      "example.attribute",
			value:     []byte("123456789"),
			accAddr:   s.user1,
			ownerAddr: s.user1Addr,
			wantErr:   true,
			errorMsg:  "no keys deleted with name example.attribute value 123456789",
		},
		{
			testName:  "should successfully delete attribute, with same key but different value",
			name:      "example.attribute",
			value:     []byte("diff value"),
			accAddr:   s.user1,
			ownerAddr: s.user1Addr,
			wantErr:   false,
			errorMsg:  "",
		},
	}
	for _, tc := range cases {
		tc := tc

		s.Run(tc.testName, func() {
			err := s.app.AttributeKeeper.DeleteAttribute(s.ctx, tc.accAddr, tc.name, &tc.value, tc.ownerAddr)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetAllAttributes() {

	attributes, err := s.app.AttributeKeeper.GetAllAttributes(s.ctx, s.user1)
	s.NoError(err)
	s.Equal(0, len(attributes))

	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	s.NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
	attributes, err = s.app.AttributeKeeper.GetAllAttributes(s.ctx, s.user1)
	s.NoError(err)
	s.Equal(attr.Name, attributes[0].Name)
	s.Equal(attr.Address, attributes[0].Address)
	s.Equal(attr.Value, attributes[0].Value)
}

func (s *KeeperTestSuite) TestGetAttributesByName() {

	attr := types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	}
	s.NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
	_, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1, "blah")
	s.Error(err)
	s.Equal("no address bound to name", err.Error())
	attributes, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.user1, "example.attribute")
	s.NoError(err)
	s.Equal(1, len(attributes))
	s.Equal(attr.Name, attributes[0].Name)
	s.Equal(attr.Address, attributes[0].Address)
	s.Equal(attr.Value, attributes[0].Value)
}

func (s *KeeperTestSuite) TestInitGenesisAddingAttributes() {
	var attributeData types.GenesisState
	attributeData.Attributes = append(attributeData.Attributes, types.Attribute{
		Name:          "example.attribute",
		Value:         []byte("0123456789"),
		Address:       s.user1,
		AttributeType: types.AttributeType_String,
	})
	s.Assert().NotPanics(func() { s.app.AttributeKeeper.InitGenesis(s.ctx, &attributeData) })
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
		s.NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.user1Addr), "should save successfully")
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
