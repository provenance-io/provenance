package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"
	attrTypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

func TestAllowMarkerSend(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	owner, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	acct := app.AccountKeeper.NewAccountWithAddress(ctx, owner)
	app.AccountKeeper.SetAccount(ctx, acct)
	app.NameKeeper.SetNameRecord(ctx, "kyc.provenance.io", owner, false)
	app.NameKeeper.SetNameRecord(ctx, "not-kyc.provenance.io", owner, false)

	acctWithAttrs := "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"
	app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "kyc.provenance.io",
			Value:         []byte("string value"),
			Address:       acctWithAttrs,
			AttributeType: attrTypes.AttributeType_String,
		},
		owner,
	)
	app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "not-kyc.provenance.io",
			Value:         []byte("string value"),
			Address:       acctWithAttrs,
			AttributeType: attrTypes.AttributeType_String,
		},
		owner,
	)
	acctWithoutDepositAccess := testUserAddress("acctWithoutDepositAccess")

	nrMarkerDenom := "nonrestrictedmarker"
	nrMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(nrMarkerDenom), nil, 0, 0)
	app.MarkerKeeper.SetMarker(ctx, types.NewMarkerAccount(nrMarkerAcct, sdk.NewInt64Coin(nrMarkerDenom, 1000), acct.GetAddress(), nil, types.StatusFinalized, types.MarkerType_Coin, true, true, false, []string{}))

	rMarkerDenom := "restrictedmarkernoreqattributes"
	rMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerDenom), nil, 1, 0)
	app.MarkerKeeper.SetMarker(ctx, types.NewMarkerAccount(rMarkerAcct, sdk.NewInt64Coin(rMarkerDenom, 1000), acct.GetAddress(), nil, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, true, false, []string{}))

	rMarkerDenom2 := "restrictedmarkerreqattributes2"
	rMarkerAcct2 := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerDenom2), nil, 2, 0)
	rMarker2 := types.NewMarkerAccount(rMarkerAcct2, sdk.NewInt64Coin(rMarkerDenom2, 1000), acct.GetAddress(), nil, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, true, false, []string{"some.attribute.that.i.require"})
	app.MarkerKeeper.SetMarker(ctx, rMarker2)

	rMarkerDenom3 := "restrictedmarkerreqattributes3"
	rMarkerAcct3 := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerDenom3), nil, 3, 0)
	rMarker3 := types.NewMarkerAccount(rMarkerAcct3, sdk.NewInt64Coin(rMarkerDenom3, 1000), acct.GetAddress(), nil, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, true, false, []string{"kyc.provenance.io"})
	app.MarkerKeeper.SetMarker(ctx, rMarker3)

	rMarkerDenom4 := "restrictedmarkerreqattributes4"
	rMarkerAcct4 := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerDenom4), nil, 4, 0)
	rMarker4 := types.NewMarkerAccount(rMarkerAcct4, sdk.NewInt64Coin(rMarkerDenom4, 1000), acct.GetAddress(), nil, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, true, false, []string{"kyc.provenance.io", "not-kyc.provenance.io"})
	app.MarkerKeeper.SetMarker(ctx, rMarker4)

	rMarkerDenom5 := "restrictedmarkerreqattributes5"
	rMarkerAcct5 := authtypes.NewBaseAccount(types.MustGetMarkerAddress(rMarkerDenom5), nil, 5, 0)
	rMarker5 := types.NewMarkerAccount(rMarkerAcct5, sdk.NewInt64Coin(rMarkerDenom5, 1000), acct.GetAddress(), nil, types.StatusFinalized, types.MarkerType_RestrictedCoin, true, true, false, []string{"kyc.provenance.io", "not-kyc.provenance.io", "foo.provenance.io"})
	app.MarkerKeeper.SetMarker(ctx, rMarker5)

	dMarkerDenom := "depositaccessdenom"
	dMarkerAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(dMarkerDenom), nil, 5, 0)
	dMarker := types.NewMarkerAccount(dMarkerAcct, sdk.NewInt64Coin(dMarkerDenom, 1000), acct.GetAddress(),
		[]types.AccessGrant{*types.NewAccessGrant(owner, []types.Access{types.Access_Deposit})}, types.StatusFinalized, types.MarkerType_Coin, true, true, false, []string{})
	app.MarkerKeeper.SetMarker(ctx, dMarker)

	testCases := []struct {
		name          string
		from          string
		to            string
		denom         string
		expectedError string
	}{
		{
			name:          "should succeed - non restricted marker",
			from:          acct.GetAddress().String(),
			to:            acctWithAttrs,
			denom:         nrMarkerDenom,
			expectedError: "",
		},
		{
			name:          "should succeed - sent from marker module",
			from:          authtypes.NewModuleAddress(types.CoinPoolName).String(),
			to:            acctWithAttrs,
			denom:         rMarkerDenom,
			expectedError: "",
		},
		{
			name:          "should succeed - sent from ibc transfer module",
			from:          authtypes.NewModuleAddress(ibctypes.ModuleName).String(),
			to:            acctWithAttrs,
			denom:         rMarkerDenom,
			expectedError: "",
		},
		{
			name:          "should fail - restricted marker with empty required attributes and no transfer rights",
			from:          acct.GetAddress().String(),
			to:            acctWithAttrs,
			denom:         rMarkerDenom,
			expectedError: fmt.Sprintf("%s does not have transfer permissions", acct.GetAddress().String()),
		},
		{
			name:          "should fail - restricted marker with required attributes but none match",
			from:          acct.GetAddress().String(),
			to:            acctWithAttrs,
			denom:         rMarkerDenom2,
			expectedError: fmt.Sprintf("address %s does not contain the required attributes %v", acctWithAttrs, rMarker2.GetRequiredAttributes()),
		},
		{
			name:          "should succeed - account contains the needed attribute",
			from:          acct.GetAddress().String(),
			to:            acctWithAttrs,
			denom:         rMarkerDenom3,
			expectedError: "",
		},
		{
			name:          "should succeed - account contains the both needed attribute",
			from:          acct.GetAddress().String(),
			to:            acctWithAttrs,
			denom:         rMarkerDenom4,
			expectedError: "",
		},
		{
			name:          "should fail - account does not contain needed attribute",
			from:          acct.GetAddress().String(),
			to:            acctWithAttrs,
			denom:         rMarkerDenom5,
			expectedError: fmt.Sprintf("address %s does not contain the required attributes %v", acctWithAttrs, rMarker5.GetRequiredAttributes()),
		},
		{
			name:          "should fail - to send marker to escrow without deposit access",
			from:          acctWithoutDepositAccess.String(),
			to:            types.MustGetMarkerAddress(dMarkerDenom).String(),
			denom:         dMarkerDenom,
			expectedError: fmt.Sprintf("%s does not have deposit access for %v", acctWithoutDepositAccess.String(), dMarkerDenom),
		},
		{
			name:          "should succeed - to send marker to escrow user had deposit rights",
			from:          acct.GetAddress().String(),
			to:            types.MustGetMarkerAddress(dMarkerDenom).String(),
			denom:         dMarkerDenom,
			expectedError: "",
		},
	}
	for _, tc := range testCases {
		err := app.MarkerKeeper.AllowMarkerSend(ctx, tc.from, tc.to, tc.denom)
		if len(tc.expectedError) > 0 {
			assert.NotNil(t, err, tc.name)
			assert.EqualError(t, err, tc.expectedError, tc.name)

		} else {
			assert.NoError(t, err, tc.name)
		}
	}
}

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
			name:           "should fail - empty required attr and attr",
			reqAttr:        "",
			attr:           "",
			expectedResult: false,
		},
		{
			name:           "should fail - extra ending",
			reqAttr:        "test.provenance.io",
			attr:           "test.provenance.iox",
			expectedResult: false,
		},
		{
			name:           "should fail - wildcard extra ending",
			reqAttr:        "*.provenance.io",
			attr:           "test.provenance.iox",
			expectedResult: false,
		},
		{
			name:           "should fail - wildcard extra beginning",
			reqAttr:        "*.provenance.io",
			attr:           "test.xprovenance.io",
			expectedResult: false,
		},
	}
	for _, tc := range testCases {
		result := keeper.MatchAttribute(tc.reqAttr, tc.attr)
		require.Equal(t, tc.expectedResult, result, fmt.Sprintf("%s", tc.name))
	}
}

func TestAddToRequiredAttributes(t *testing.T) {
	tests := []struct {
		name          string
		addList       []string
		reqAttrs      []string
		expectedAttrs []string
		expectedError string
	}{
		{
			name:          "should fail, duplicate value",
			addList:       []string{"foo", "bar"},
			reqAttrs:      []string{"foo", "baz"},
			expectedError: "cannot add duplicate entry to required attributes foo",
		},
		{
			name:          "should succeed, add elements to none empty list",
			addList:       []string{"qux", "fix"},
			reqAttrs:      []string{"foo", "bar", "baz"},
			expectedAttrs: []string{"foo", "bar", "baz", "qux", "fix"},
		},
		{
			name:          "should succeed, add elements to empty list",
			addList:       []string{"qux", "fix"},
			reqAttrs:      []string{},
			expectedAttrs: []string{"qux", "fix"},
		},
		{
			name:          "should succeed, nothing added",
			addList:       []string{},
			reqAttrs:      []string{"foo", "bar", "baz"},
			expectedAttrs: []string{"foo", "bar", "baz"},
		},
		{
			name:          "should succeed, two empty lists",
			addList:       []string{},
			reqAttrs:      []string{},
			expectedAttrs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualAttrs, err := keeper.AddToRequiredAttributes(tt.addList, tt.reqAttrs)
			if len(tt.expectedError) == 0 {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedAttrs, actualAttrs)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Nil(t, tt.expectedAttrs)
			}
		})
	}
}

func TestRemovesFromRequiredAttributes(t *testing.T) {
	tests := []struct {
		name          string
		currentAttrs  []string
		removeAttrs   []string
		expectedAttrs []string
		expectedError string
	}{
		{
			name:          "should succeed, removing a single element",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{"bar"},
			expectedAttrs: []string{"foo", "baz"},
		},
		{
			name:          "should fail, element doesn't exist",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{"qux"},
			expectedAttrs: nil,
			expectedError: "remove required attributes list had incorrect entries",
		},
		{
			name:          "should succeed, removing multiple elements",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{"foo", "baz"},
			expectedAttrs: []string{"bar"},
		},
		{
			name:          "should succeed, removing no elements",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{},
			expectedAttrs: []string{"foo", "bar", "baz"},
		},
		{
			name:          "should succeed, remove all elements",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{"baz", "foo", "bar"},
			expectedAttrs: []string{},
		},
		{
			name:          "should succeed, both empty lists",
			currentAttrs:  []string{},
			removeAttrs:   []string{},
			expectedAttrs: []string{},
		},
		{
			name:          "should fail, trying to remove elements from empty list",
			currentAttrs:  []string{},
			removeAttrs:   []string{"blah"},
			expectedAttrs: []string{},
			expectedError: "remove required attributes list had incorrect entries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualAttrs, err := keeper.RemovesFromRequiredAttributes(tt.currentAttrs, tt.removeAttrs)
			if len(tt.expectedError) == 0 {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedAttrs, actualAttrs)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			}
		})
	}
}
