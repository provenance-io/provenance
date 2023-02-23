package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"
	attrTypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/marker/keeper"
)

func TestNormalizeRequiredAttributes(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	testCases := []struct {
		name               string
		requiredAttributes []string
		expectedNormalized []string
		expectedError      string
	}{
		{
			name:               "should succeed - empty required attrs",
			requiredAttributes: []string{},
			expectedNormalized: []string{},
			expectedError:      "",
		},
		{
			name:               "should fail - segment name too short",
			requiredAttributes: []string{"."},
			expectedNormalized: []string{},
			expectedError:      "segment of name is too short",
		},
		{
			name:               "should fail - segment name too short2",
			requiredAttributes: []string{"provenance.io"},
			expectedNormalized: []string{"provenance.io"},
			expectedError:      "",
		},
		{
			name:               "should fail - invalid wild card value",
			requiredAttributes: []string{"*b.provenance.io"},
			expectedNormalized: []string{},
			expectedError:      "value provided for name is invalid",
		},
		{
			name:               "should succeed - valid wild card value",
			requiredAttributes: []string{"*.provenance.io"},
			expectedNormalized: []string{"*.provenance.io"},
			expectedError:      "",
		},
	}
	for _, tc := range testCases {
		result, err := app.MarkerKeeper.NormalizeRequiredAttributes(ctx, tc.requiredAttributes)
		if len(tc.expectedError) > 0 {
			require.NotNil(t, err)
			require.EqualError(t, err, tc.expectedError)

		} else {
			require.NoError(t, err)
			require.Equal(t, tc.expectedNormalized, result)
		}
	}
}

func TestContainsRequiredAttributes(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	owner, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	acct := app.AccountKeeper.NewAccountWithAddress(ctx, owner)
	app.AccountKeeper.SetAccount(ctx, acct)
	app.NameKeeper.SetNameRecord(ctx, "kyc.provenance.io", owner, false)
	app.NameKeeper.SetNameRecord(ctx, "not-kyc.provenance.io", owner, false)

	app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "kyc.provenance.io",
			Value:         []byte("string value"),
			Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			AttributeType: attrTypes.AttributeType_String,
		},
		owner,
	)
	app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "not-kyc.provenance.io",
			Value:         []byte("string value"),
			Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			AttributeType: attrTypes.AttributeType_String,
		},
		owner,
	)
	testCases := []struct {
		name               string
		requiredAttributes []string
		address            string
		expectedResult     bool
		expectedError      string
	}{
		{
			name:               "should succeed - empty required attrs",
			requiredAttributes: []string{},
			address:            "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			expectedResult:     true,
			expectedError:      "",
		},
		{
			name:               "should succeed - wildcard match",
			requiredAttributes: []string{"*.io"},
			address:            "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			expectedResult:     true,
			expectedError:      "",
		},
		{
			name:               "should succeed - wildcard match 2",
			requiredAttributes: []string{"*.provenance.io"},
			address:            "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			expectedResult:     true,
			expectedError:      "",
		},
		{
			name:               "should succeed - exact match",
			requiredAttributes: []string{"kyc.provenance.io"},
			address:            "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			expectedResult:     true,
			expectedError:      "",
		},
		{
			name:               "should succeed - exact match multiple",
			requiredAttributes: []string{"kyc.provenance.io", "kyc.provenance.io"},
			address:            "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			expectedResult:     true,
			expectedError:      "",
		},
		{
			name:               "should fail - no match for notfound.provenance.io",
			requiredAttributes: []string{"notfound.provenance.io", "kyc.provenance.io"},
			address:            "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			expectedResult:     false,
			expectedError:      "",
		},
		{
			name:               "should succeed - account has no attributes and required attributes empty",
			requiredAttributes: []string{},
			address:            "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
			expectedResult:     true,
			expectedError:      "",
		},
		{
			name:               "should fail - account has no attributes and required attributes populated",
			requiredAttributes: []string{"kyc.provenance.io"},
			address:            "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
			expectedResult:     false,
			expectedError:      "",
		},
	}
	for _, tc := range testCases {
		result, err := app.MarkerKeeper.ContainsRequiredAttributes(ctx, tc.requiredAttributes, tc.address)
		if len(tc.expectedError) > 0 {
			assert.NotNil(t, err, tc.name)
			assert.EqualError(t, err, tc.expectedError, tc.name)

		} else {
			assert.NoError(t, err, tc.name)
			assert.Equal(t, tc.expectedResult, result, tc.name)
		}
	}
}

func TestEnsureAllRequiredAttributesExist(t *testing.T) {
	testCases := []struct {
		name           string
		requiredAtts   []string
		attributes     []attrTypes.Attribute
		expectedResult bool
	}{
		{
			name:           "should succeed - empty required attrs and attributes",
			requiredAtts:   []string{},
			attributes:     []attrTypes.Attribute{},
			expectedResult: true,
		},
		{
			name:         "should succeed - required with wildcard is contained in attributes",
			requiredAtts: []string{"*.provenance.io"},
			attributes: []attrTypes.Attribute{
				{Name: "kyc.provenance.io"},
			},
			expectedResult: true,
		},
		{
			name:         "should succeed - required is contained in attributes",
			requiredAtts: []string{"kyc.provenance.io"},
			attributes: []attrTypes.Attribute{
				{Name: "kyc.provenance.io"},
			},
			expectedResult: true,
		},
		{
			name:         "should succeed - multiple attrs and required attrs",
			requiredAtts: []string{"kyc.provenance.io", "kyc.provenance.com", "kyc.provenance.net"},
			attributes: []attrTypes.Attribute{
				{Name: "kyc.provenance.io"},
				{Name: "kyc.provenance.com"},
				{Name: "kyc.provenance.net"},
				{Name: "kyc.provenance.de"},
			},
			expectedResult: true,
		},
		{
			name:           "should fail - missing required attr",
			requiredAtts:   []string{"kyc.provenance.io", "non-kyc.provenance.io"},
			attributes:     []attrTypes.Attribute{{Name: "kyc.provenance.io"}},
			expectedResult: false,
		},
		{
			name:           "should fail - missing required attr with wildcard",
			requiredAtts:   []string{"*.provenance.io", "non-kyc.provenance.io"},
			attributes:     []attrTypes.Attribute{{Name: "kyc.provenance.io"}},
			expectedResult: false,
		},
	}
	for _, tc := range testCases {
		result := keeper.EnsureAllRequiredAttributesExist(tc.requiredAtts, tc.attributes)
		require.Equal(t, tc.expectedResult, result, fmt.Sprintf("%s", tc.name))
	}
}

func TestMatchAttribute(t *testing.T) {
	testCases := []struct {
		name           string
		reqAttr        string
		attr           string
		expectedResult bool
	}{
		{
			name:           "should succeed - wildcard on single name",
			reqAttr:        "*.provenance.io",
			attr:           "test.provenance.io",
			expectedResult: true,
		},
		{
			name:           "should succeed - wildcard on multiple names",
			reqAttr:        "*.provenance.io",
			attr:           "test.test.test.provenance.io",
			expectedResult: true,
		},
		{
			name:           "should succeed - literal match",
			reqAttr:        "test.provenance.io",
			attr:           "test.provenance.io",
			expectedResult: true,
		},
		{
			name:           "should fail - wildcard match",
			reqAttr:        "*.provenance.io",
			attr:           "test.provenance.com",
			expectedResult: false,
		},
		{
			name:           "should fail - literal match",
			reqAttr:        "test.provenance.io",
			attr:           "test.provenance.com",
			expectedResult: false,
		},
		{
			name:           "should fail - empty required attr",
			reqAttr:        "",
			attr:           "test.provenance.com",
			expectedResult: false,
		},
		{
			name:           "should fail - empty required attr and attr ",
			reqAttr:        "",
			attr:           "",
			expectedResult: false,
		},
	}
	for _, tc := range testCases {
		result := keeper.MatchAttribute(tc.reqAttr, tc.attr)
		require.Equal(t, tc.expectedResult, result, fmt.Sprintf("%s", tc.name))
	}
}
