package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/provenance-io/provenance/x/marker"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"
	attrTypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
)

func TestSendRestrictionFn(t *testing.T) {
	c := func(amt int64, denom string) sdk.Coin {
		return sdk.NewInt64Coin(denom, amt)
	}
	cz := func(coins ...sdk.Coin) sdk.Coins {
		return sdk.NewCoins(coins...)
	}

	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctxWithBypass := keeper.WithMarkerSendRestrictionBypass(ctx, true)
	owner := sdk.AccAddress("owner_address_______")
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, owner))
	require.NoError(t, app.NameKeeper.SetNameRecord(ctx, "kyc.provenance.io", owner, false), "SetNameRecord kyc.provenance.io")
	require.NoError(t, app.NameKeeper.SetNameRecord(ctx, "not-kyc.provenance.io", owner, false), "SetNameRecord not-kyc.provenance.io")

	addrWithAttrs := sdk.AccAddress("addr_with_attributes")
	addrWithAttrsStr := addrWithAttrs.String()
	require.NoError(t, app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "kyc.provenance.io",
			Value:         []byte("string value"),
			Address:       addrWithAttrsStr,
			AttributeType: attrTypes.AttributeType_String,
		},
		owner,
	), "SetAttribute kyc.provenance.io")
	require.NoError(t, app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "not-kyc.provenance.io",
			Value:         []byte("string value"),
			Address:       addrWithAttrsStr,
			AttributeType: attrTypes.AttributeType_String,
		},
		owner,
	), "SetAttribute not-kyc.provenance.io")

	addrWithoutAttrs := sdk.AccAddress("addr_without_attribs")
	addrWithTransfer := sdk.AccAddress("addr_with_transfer__")

	coin := types.MarkerType_Coin
	restricted := types.MarkerType_RestrictedCoin

	acctNum := uint64(0)
	newMarker := func(denom string, markerType types.MarkerType, reqAttrs []string) *types.MarkerAccount {
		baseAcct := authtypes.NewBaseAccount(types.MustGetMarkerAddress(denom), nil, acctNum, 0)
		acctNum++
		var access []types.AccessGrant
		if markerType == restricted {
			access = []types.AccessGrant{
				{Address: addrWithTransfer.String(), Permissions: types.AccessList{types.Access_Transfer}},
			}
		}
		rv := types.NewMarkerAccount(
			baseAcct,
			sdk.NewInt64Coin(denom, 1000),
			owner,
			access,
			types.StatusFinalized,
			markerType,
			true,  // supply fixed
			true,  // allow gov
			false, // no force transfer
			reqAttrs,
		)
		app.MarkerKeeper.SetMarker(ctx, rv)
		return rv
	}

	nrDenom := "nonrestrictedmarker"
	newMarker(nrDenom, coin, nil)

	rDenomNoAttr := "restrictedmarkernoreqattributes"
	newMarker(rDenomNoAttr, restricted, nil)

	rDenom1AttrNoOneHas := "restrictedmarkerreqattributes2"
	newMarker(rDenom1AttrNoOneHas, restricted, []string{"some.attribute.that.i.require"})

	rDenom1Attr := "restrictedmarkerreqattributes3"
	newMarker(rDenom1Attr, restricted, []string{"kyc.provenance.io"})

	rDenom2Attrs := "restrictedmarkerreqattributes4"
	newMarker(rDenom2Attrs, restricted, []string{"kyc.provenance.io", "not-kyc.provenance.io"})

	rDenom3Attrs := "restrictedmarkerreqattributes5"
	newMarker(rDenom3Attrs, restricted, []string{"kyc.provenance.io", "not-kyc.provenance.io", "foo.provenance.io"})

	testCases := []struct {
		name   string
		ctx    *sdk.Context
		from   sdk.AccAddress
		to     sdk.AccAddress
		amt    sdk.Coins
		expErr string
	}{
		{
			name:   "unknown denom",
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, "unknowncoin")),
			expErr: "",
		},
		{
			name:   "non restricted marker",
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, nrDenom)),
			expErr: "",
		},
		{
			name:   "addr has transfer, denom without attrs",
			from:   addrWithTransfer,
			to:     addrWithoutAttrs,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: "",
		},
		{
			name:   "addr has transfer, denom with 3 attrs, to has none",
			from:   addrWithTransfer,
			to:     addrWithoutAttrs,
			amt:    cz(c(1, rDenom3Attrs)),
			expErr: "",
		},
		// Untested: GetAllAttributesAddr returns an error. Only happens when store data can't be unmarshalled. Can't do that from here.
		{
			name:   "restricted marker with empty required attributes and no transfer rights",
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: fmt.Sprintf("%s does not have transfer permissions", owner.String()),
		},
		{
			name: "restricted marker with required attributes but none match",
			from: owner,
			to:   addrWithAttrs,
			amt:  cz(c(1, rDenom1AttrNoOneHas)),
			expErr: fmt.Sprintf("address %s does not contain the %q required attribute: \"some.attribute.that.i.require\"",
				addrWithAttrsStr, rDenom1AttrNoOneHas),
			// This should be the exact same test as the below one, but without a bypass context, so expect an error.
		},
		{
			// This should be the exact same test as the above one, but with a bypass context, so no error is expected.
			name:   "with bypass, restricted marker with required attributes but none match",
			ctx:    &ctxWithBypass,
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: "",
		},
		{
			name:   "account contains the needed attribute",
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenom1Attr)),
			expErr: "",
		},
		{
			name:   "account contains both needed attributes",
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenom2Attrs)),
			expErr: "",
		},
		{
			name: "account contains 2 of 3 needed attributes",
			from: owner,
			to:   addrWithAttrs,
			amt:  cz(c(1, rDenom3Attrs)),
			expErr: fmt.Sprintf("address %s does not contain the %q required attribute: \"foo.provenance.io\"",
				addrWithAttrsStr, rDenom3Attrs),
		},
		{
			name: "account has no attributes, needs 3",
			from: owner,
			to:   addrWithoutAttrs,
			amt:  cz(c(1, rDenom3Attrs)),
			expErr: fmt.Sprintf("address %s does not contain the %q required attributes: "+
				"\"kyc.provenance.io\", \"not-kyc.provenance.io\", \"foo.provenance.io\"",
				addrWithoutAttrs, rDenom3Attrs),
		},
		{
			name:   "account has no attributes, denom not restricted",
			from:   addrWithTransfer,
			to:     addrWithoutAttrs,
			amt:    cz(c(1, nrDenom)),
			expErr: "",
		},
		{
			name:   "two denoms, unrestricted and has needed attribute",
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, nrDenom), c(1, rDenom1Attr)),
			expErr: "",
		},
		{
			name:   "two denoms, has needed attribute and unrestricted",
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenom1Attr), c(1, nrDenom)),
			expErr: "",
		},
		{
			name: "two denoms, unrestricted and missing attribute",
			from: owner,
			to:   addrWithAttrs,
			amt:  cz(c(1, nrDenom), c(1, rDenom1AttrNoOneHas)),
			expErr: fmt.Sprintf("address %s does not contain the %q required attribute: \"some.attribute.that.i.require\"",
				addrWithAttrsStr, rDenom1AttrNoOneHas),
		},
		{
			name: "two denoms, missing attribute and unrestricted",
			from: owner,
			to:   addrWithAttrs,
			amt:  cz(c(1, rDenom1AttrNoOneHas), c(1, nrDenom)),
			expErr: fmt.Sprintf("address %s does not contain the %q required attribute: \"some.attribute.that.i.require\"",
				addrWithAttrsStr, rDenom1AttrNoOneHas),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tCtx := ctx
			if tc.ctx != nil {
				tCtx = *tc.ctx
			}
			newTo, err := app.MarkerKeeper.SendRestrictionFn(tCtx, tc.from, tc.to, tc.amt)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "SendRestrictionFn error")
			} else {
				assert.NoError(t, err, "SendRestrictionFn error")
				assert.Equal(t, tc.to, newTo, "SendRestrictionFn returned address")
			}
		})
	}
}

func TestBankSendUsesSendRestrictionFn(t *testing.T) {
	// This test only checks that the marker SendRestrictionFn is applied during a SendCoins.
	// Testing of the actual SendRestrictionFn is assumed to be done elsewhere more extensively.

	cz := func(amt int64, denom string) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(denom, amt))
	}

	markerDenom := "markercoin"

	addrNameOwner := sdk.AccAddress("name_owner__________")
	addrHasWithdraw := sdk.AccAddress("has_withdraw________")
	addrHasAttr := sdk.AccAddress("has_attribute_______")
	addrOther := sdk.AccAddress("other_address_______")

	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, addrNameOwner))
	err := app.NameKeeper.SetNameRecord(ctx, "kyc.provenance.io", addrNameOwner, false)
	require.NoError(t, err, "SetNameRecord kyc.provenance.io")
	require.NoError(t, app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "kyc.provenance.io",
			Value:         []byte("string value"),
			Address:       addrHasAttr.String(),
			AttributeType: attrTypes.AttributeType_String,
		},
		addrNameOwner,
	), "SetAttribute kyc.provenance.io")

	makeMarkerMsg := &types.MsgAddFinalizeActivateMarkerRequest{
		Amount:      sdk.NewInt64Coin(markerDenom, 1000),
		Manager:     addrHasWithdraw.String(),
		FromAddress: addrHasWithdraw.String(),
		MarkerType:  types.MarkerType_RestrictedCoin,
		AccessList: []types.AccessGrant{
			{Address: addrHasWithdraw.String(), Permissions: types.AccessList{types.Access_Withdraw}},
		},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    false,
		RequiredAttributes:     []string{"kyc.provenance.io"},
	}
	markerHandler := marker.NewHandler(app.MarkerKeeper)
	_, err = markerHandler(ctx, makeMarkerMsg)
	require.NoError(t, err, "makeMarkerMsg")
	err = app.MarkerKeeper.WithdrawCoins(ctx, addrHasWithdraw, addrHasAttr, markerDenom, cz(100, markerDenom))
	require.NoError(t, err, "WithdrawCoins to addrHasTransfer")
	err = app.MarkerKeeper.WithdrawCoins(ctx, addrHasWithdraw, addrOther, markerDenom, cz(100, markerDenom))
	require.NoError(t, err, "WithdrawCoins to addrOther")

	// Done with setup.
	// addrOther and addrHasAttr now each have 100 of the marker denom.
	// addrHasAttr has the attribute needed to receive the denom, and addrOther does not.

	t.Run("send to address without attributes", func(t *testing.T) {
		expErr := fmt.Sprintf("address %s does not contain the %q required attribute: \"kyc.provenance.io\"",
			addrOther, markerDenom)
		err = app.BankKeeper.SendCoins(ctx, addrHasAttr, addrOther, cz(5, markerDenom))
		assert.EqualError(t, err, expErr, "SendCoins")
		expBal := cz(100, markerDenom)
		hasAttrBal := app.BankKeeper.GetBalance(ctx, addrHasAttr, markerDenom)
		assert.Equal(t, expBal.String(), hasAttrBal.String(), "GetBalance addrHasAttr")
		otherBal := app.BankKeeper.GetBalance(ctx, addrOther, markerDenom)
		assert.Equal(t, expBal.String(), otherBal.String(), "GetBalance addrOther")
	})

	t.Run("send to address with attributes", func(t *testing.T) {
		err = app.BankKeeper.SendCoins(ctx, addrOther, addrHasAttr, cz(6, markerDenom))
		assert.NoError(t, err, "SendCoins")
		hasAttrExpBal := cz(106, markerDenom)
		hasAttrBal := app.BankKeeper.GetBalance(ctx, addrHasAttr, markerDenom)
		assert.Equal(t, hasAttrExpBal.String(), hasAttrBal.String(), "GetBalance addrHasAttr")
		otherExpBal := cz(94, markerDenom)
		otherBal := app.BankKeeper.GetBalance(ctx, addrOther, markerDenom)
		assert.Equal(t, otherExpBal.String(), otherBal.String(), "GetBalance addrOther")
	})
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
		t.Run(tc.name, func(t *testing.T) {
			result, err := app.MarkerKeeper.NormalizeRequiredAttributes(ctx, tc.requiredAttributes)
			if len(tc.expectedError) > 0 {
				require.EqualError(t, err, tc.expectedError, "NormalizeRequiredAttributes error")
			} else {
				require.NoError(t, err, "NormalizeRequiredAttributes error")
				require.Equal(t, tc.expectedNormalized, result, "NormalizeRequiredAttributes result")
			}
		})
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
		t.Run(tc.name, func(t *testing.T) {
			result := keeper.MatchAttribute(tc.reqAttr, tc.attr)
			require.Equal(t, tc.expectedResult, result, "MatchAttribute")
		})
	}
}
