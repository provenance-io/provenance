package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	simapp "github.com/provenance-io/provenance/app"
	attrTypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/marker"
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
	ctx := app.BaseApp.NewContext(false)
	// ctxP returns a pointer to the provided context.
	ctxP := func(ctx sdk.Context) *sdk.Context {
		return &ctx
	}
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
	addrWithForceTransfer := sdk.AccAddress("addr_with_force_tran")
	addrWithDeposit := sdk.AccAddress("addrWithDeposit_____")
	addrWithWithdraw := sdk.AccAddress("addrWithWithdraw____")
	addrWithTranDep := sdk.AccAddress("addrWithTranDep_____")
	addrWithTranWithdraw := sdk.AccAddress("addrWithTranWithdraw")
	addrWithTranDepWithdraw := sdk.AccAddress("addrWithTranDepWithd")
	addrWithDepWithdraw := sdk.AccAddress("addrWithDepWithdraw_")
	addrWithDenySend := sdk.AccAddress("addrWithDenySend_____")
	addrOther := sdk.AccAddress("addrOther___________")

	addrFeeCollector := app.MarkerKeeper.GetFeeCollectorAddr()
	bypassAddrs := app.MarkerKeeper.GetReqAttrBypassAddrs()
	var addrWithBypass, addrWithBypassNoDep sdk.AccAddress
	for _, addr := range bypassAddrs {
		if addr.Equals(addrFeeCollector) {
			continue
		}
		if len(addrWithBypass) == 0 {
			addrWithBypass = addr
			continue
		}
		addrWithBypassNoDep = addr
		break
	}
	require.NotEmpty(t, addrWithBypass, "addrWithBypass (first bypass address that is not the fee collector)")
	require.NotEmpty(t, addrWithBypassNoDep, "addrWithBypassNoDep (second bypass address that is not the fee collector)")

	coin := types.MarkerType_Coin
	restricted := types.MarkerType_RestrictedCoin

	newMarkerAcc := func(denom string, markerType types.MarkerType, reqAttrs []string) *types.MarkerAccount {
		addr, err := types.MarkerAddress(denom)
		require.NoError(t, err, "MarkerAddress(%q)", denom)

		rv := &types.MarkerAccount{
			BaseAccount:            authtypes.NewBaseAccountWithAddress(addr),
			Manager:                owner.String(),
			Status:                 types.StatusProposed,
			Denom:                  denom,
			Supply:                 sdkmath.NewInt(1000),
			MarkerType:             markerType,
			SupplyFixed:            true,
			AllowGovernanceControl: true,
			AllowForcedTransfer:    false,
			RequiredAttributes:     reqAttrs,
		}

		if markerType == restricted {
			rv.AccessControl = []types.AccessGrant{
				{Address: addrWithTransfer.String(), Permissions: types.AccessList{types.Access_Transfer}},
				{Address: addrWithForceTransfer.String(), Permissions: types.AccessList{types.Access_ForceTransfer}},
				{Address: addrWithDeposit.String(), Permissions: types.AccessList{types.Access_Deposit}},
				{Address: addrWithWithdraw.String(), Permissions: types.AccessList{types.Access_Withdraw}},
				{Address: addrWithTranDep.String(), Permissions: types.AccessList{types.Access_Deposit, types.Access_Transfer}},
				{Address: addrWithTranWithdraw.String(), Permissions: types.AccessList{types.Access_Withdraw, types.Access_Transfer}},
				{Address: addrWithDepWithdraw.String(), Permissions: types.AccessList{types.Access_Deposit, types.Access_Withdraw}},
				{Address: addrWithTranDepWithdraw.String(), Permissions: types.AccessList{types.Access_Deposit, types.Access_Withdraw, types.Access_Transfer}},
				// It's silly to give any permissions to a bypass address, but I do so in here to hit some test cases.
				{Address: addrWithBypass.String(), Permissions: types.AccessList{types.Access_Deposit}},
			}
		}

		return rv
	}
	createActiveMarker := func(marker *types.MarkerAccount) *types.MarkerAccount {
		nav := []types.NetAssetValue{types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, int64(1)), 1)}
		err := app.MarkerKeeper.AddSetNetAssetValues(ctx, marker, nav, t.Name())
		require.NoError(t, err, "AddSetNetAssetValues(%v, %v, %v)", marker.Denom, nav, t.Name())
		err = app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, marker)
		require.NoError(t, err, "AddFinalizeAndActivateMarker(%s)", marker.Denom)
		return marker
	}
	newMarker := func(denom string, markerType types.MarkerType, reqAttrs []string) *types.MarkerAccount {
		return createActiveMarker(newMarkerAcc(denom, markerType, reqAttrs))
	}
	newProposedMarker := func(denom string, markerType types.MarkerType, reqAttrs []string) *types.MarkerAccount {
		rv := newMarkerAcc(denom, markerType, reqAttrs)
		err := app.MarkerKeeper.AddMarkerAccount(ctx, rv)
		require.NoError(t, err, "AddMarkerAccount(%s)", denom)
		return rv

	}

	nrDenom := "nonrestrictedmarker"
	nrMarker := newMarker(nrDenom, coin, nil)

	rDenomNoAttr := "restrictedmarkernoreqattributes"
	rMarkerNoAttr := newMarker(rDenomNoAttr, restricted, nil)
	app.MarkerKeeper.AddSendDeny(ctx, rMarkerNoAttr.GetAddress(), addrWithDenySend)

	rDenom1AttrNoOneHas := "restrictedmarkerreqattributes2"
	newMarker(rDenom1AttrNoOneHas, restricted, []string{"some.attribute.that.i.require"})

	rDenom1Attr := "restrictedmarkerreqattributes3"
	rMarker1Attr := newMarker(rDenom1Attr, restricted, []string{"kyc.provenance.io"})
	require.NoError(t, app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "kyc.provenance.io",
			Value:         []byte("string value"),
			Address:       rMarker1Attr.GetAddress().String(),
			AttributeType: attrTypes.AttributeType_String,
		},
		owner,
	), "SetAttribute kyc.provenance.io")

	rDenom2Attrs := "restrictedmarkerreqattributes4"
	rMarker2Attrs := newMarker(rDenom2Attrs, restricted, []string{"kyc.provenance.io", "not-kyc.provenance.io"})

	rDenom3Attrs := "restrictedmarkerreqattributes5"
	newMarker(rDenom3Attrs, restricted, []string{"kyc.provenance.io", "not-kyc.provenance.io", "foo.provenance.io"})

	rDenomProposed := "stillproposed"
	rMarkerProposed := newProposedMarker(rDenomProposed, restricted, nil)

	noAccessErr := func(addr sdk.AccAddress, role types.Access, denom string) string {
		mAddr, err := types.MarkerAddress(denom)
		require.NoError(t, err, "MarkerAddress(%q)", denom)
		return fmt.Sprintf("%s does not have %s on %s marker (%s)", addr, role, denom, mAddr)
	}

	testCases := []struct {
		name       string
		attrKeeper *WrappedAttrKeeper
		ctx        *sdk.Context
		from       sdk.AccAddress
		to         sdk.AccAddress
		amt        sdk.Coins
		expErr     string
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
			name: "restricted to fee collector from normal account",
			// include a transfer agent just to make sure that doesn't bypass anything.
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithTransfer)),
			from:   addrOther,
			to:     addrFeeCollector,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: "restricted denom " + rDenomNoAttr + " cannot be sent to the fee collector",
		},
		{
			name: "restricted to fee collector from marker module account",
			// include a transfer agent just to make sure that doesn't bypass anything.
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithTranWithdraw)),
			from:   app.MarkerKeeper.GetMarkerModuleAddr(),
			to:     addrFeeCollector,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: "cannot send restricted denom " + rDenomNoAttr + " to the fee collector",
		},
		{
			name: "restricted to fee collector from ibc transfer module account",
			// include a transfer agent just to make sure that doesn't bypass anything.
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithTransfer)),
			from:   app.MarkerKeeper.GetIbcTransferModuleAddr(),
			to:     addrFeeCollector,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: "cannot send restricted denom " + rDenomNoAttr + " to the fee collector",
		},
		{
			name:   "addr has transfer, denom without attrs",
			from:   addrWithTransfer,
			to:     addrWithoutAttrs,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: "",
		},
		{
			name:   "addr has force transfer, denom without attrs",
			from:   addrWithForceTransfer,
			to:     addrWithoutAttrs,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: fmt.Sprintf("%s does not have transfer permissions for %s", addrWithForceTransfer.String(), rDenomNoAttr),
		},
		{
			name:   "addr has transfer, denom with 3 attrs, to has none",
			from:   addrWithTransfer,
			to:     addrWithoutAttrs,
			amt:    cz(c(1, rDenom3Attrs)),
			expErr: "",
		},
		{
			name:       "error getting attrs",
			attrKeeper: NewWrappedAttrKeeper().WithGetAllAttributesAddrErrs("crazy injected attr error"),
			from:       addrOther,
			to:         addrWithAttrs,
			amt:        cz(c(1, rDenom3Attrs)),
			expErr:     "could not get attributes for " + addrWithAttrsStr + ": crazy injected attr error",
		},
		{
			name:   "restricted marker with empty required attributes and no transfer rights",
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: fmt.Sprintf("%s does not have transfer permissions for %s", owner.String(), rDenomNoAttr),
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
			ctx:    ctxP(types.WithBypass(ctx)),
			from:   owner,
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: "",
		},
		{
			name:   "from marker module account",
			from:   app.MarkerKeeper.GetMarkerModuleAddr(),
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: "",
		},
		{
			name:   "from ibc transfer module account",
			from:   app.MarkerKeeper.GetIbcTransferModuleAddr(),
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: "",
		},
		{
			name:   "send from an account on denied list",
			from:   addrWithDenySend,
			to:     addrWithAttrs,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: addrWithDenySend.String() + " is on deny list for sending restricted marker",
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
		{
			name:   "send to marker from account without deposit",
			from:   addrWithAttrs,
			to:     rMarkerNoAttr.GetAddress(),
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: noAccessErr(addrWithAttrs, types.Access_Deposit, rDenomNoAttr),
		},
		{
			name:   "send to marker from account with deposit but no transfer",
			from:   addrWithDeposit,
			to:     rMarkerNoAttr.GetAddress(),
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: noAccessErr(addrWithDeposit, types.Access_Transfer, rDenomNoAttr),
		},
		{
			name:   "send to another marker with transfer on denom but no deposit on to",
			from:   addrWithTransfer,
			to:     rMarker1Attr.GetAddress(),
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: noAccessErr(addrWithTransfer, types.Access_Deposit, rDenom1Attr),
		},
		{
			name:   "send to another marker without transfer on denom but with deposit on to",
			from:   addrWithDeposit,
			to:     rMarker1Attr.GetAddress(),
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: noAccessErr(addrWithDeposit, types.Access_Transfer, rDenomNoAttr),
		},
		{
			name:   "send to another marker with transfer on denom and deposit on to",
			from:   addrWithTranDep,
			to:     rMarker1Attr.GetAddress(),
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: "",
		},
		{
			name:   "send non-restricted coin to the marker",
			from:   addrWithoutAttrs,
			to:     nrMarker.GetAddress(),
			amt:    cz(c(1, nrDenom)),
			expErr: "",
		},
		{
			name:   "to a marker from addr with bypass but no deposit",
			from:   addrWithBypassNoDep,
			to:     rMarkerNoAttr.GetAddress(),
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: noAccessErr(addrWithBypassNoDep, types.Access_Deposit, rDenomNoAttr),
		},
		{
			name:   "to a marker with req attrs from an addr with bypass",
			from:   addrWithBypass,
			to:     rMarker1Attr.GetAddress(),
			amt:    cz(c(1, rDenom1Attr)),
			expErr: noAccessErr(addrWithBypass, types.Access_Transfer, rDenom1Attr),
		},
		{
			name:   "to marker without req attrs from addr with bypass",
			from:   addrWithBypass,
			to:     rMarkerNoAttr.GetAddress(),
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: noAccessErr(addrWithBypass, types.Access_Transfer, rDenomNoAttr),
		},
		{
			name:   "no req attrs from addr with bypass",
			from:   addrWithBypass,
			to:     addrOther,
			amt:    cz(c(3, rDenomNoAttr)),
			expErr: "",
		},
		{
			name: "with req attrs from addr with bypass",
			from: addrWithBypass,
			to:   addrOther,
			amt:  cz(c(1, rDenom1AttrNoOneHas)),
			expErr: fmt.Sprintf("address %s does not contain the %q required attribute: \"some.attribute.that.i.require\"",
				addrOther, rDenom1AttrNoOneHas),
		},
		{
			name:   "no req attrs to addr with bypass from without transfer",
			from:   addrOther,
			to:     addrWithBypass,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: addrOther.String() + " does not have transfer permissions for " + rDenomNoAttr,
		},
		{
			name:   "no req attrs to addr with bypass from with transfer",
			from:   addrWithTransfer,
			to:     addrWithBypass,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: "",
		},
		{
			name:   "with req attrs to addr with bypass from without transfer",
			from:   addrOther,
			to:     addrWithBypass,
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: "",
		},
		{
			name:   "with req attrs to addr with bypass from with transfer",
			from:   addrWithTransfer,
			to:     addrWithBypass,
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: "",
		},
		{
			name:   "with req attrs between bypass addrs",
			from:   addrWithBypass,
			to:     addrWithBypassNoDep,
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: "",
		},
		{
			name:   "without req attrs between bypass addrs",
			from:   addrWithBypass,
			to:     addrWithBypassNoDep,
			amt:    cz(c(1, rDenomNoAttr)),
			expErr: "",
		},
		{
			name:   "from marker: no admin",
			from:   rMarkerNoAttr.GetAddress(),
			to:     addrWithAttrs,
			amt:    cz(c(2, rDenomNoAttr)),
			expErr: "cannot withdraw from marker account " + rMarkerNoAttr.GetAddress().String() + " (" + rDenomNoAttr + ")",
		},
		{
			name:   "from marker: admin without withdraw permission",
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithTransfer)),
			from:   rMarkerNoAttr.GetAddress(),
			to:     addrWithAttrs,
			amt:    cz(c(2, rDenomNoAttr)),
			expErr: noAccessErr(addrWithTransfer, types.Access_Withdraw, rDenomNoAttr),
		},
		{
			name: "from marker: withdraw marker funds from inactive marker",
			ctx:  ctxP(types.WithTransferAgent(ctx, addrWithTranWithdraw)),
			from: rMarkerProposed.GetAddress(),
			to:   addrWithAttrs,
			amt:  cz(c(2, rDenomNoAttr), c(1, rDenomProposed), c(5, rDenom3Attrs)),
			expErr: fmt.Sprintf("cannot withdraw %s from %s marker (%s): marker status (proposed) is not active",
				c(1, rDenomProposed), rDenomProposed, rMarkerProposed.GetAddress(),
			),
		},
		{
			name: "from marker: withdraw non-marker funds from inactive marker",
			ctx:  ctxP(types.WithTransferAgent(ctx, addrWithTranWithdraw)),
			from: rMarkerProposed.GetAddress(),
			to:   addrWithAttrs,
			amt:  cz(c(2, rDenomNoAttr), c(5, rDenom3Attrs)),
		},
		{
			name: "from marker: withdraw from active marker",
			ctx:  ctxP(types.WithTransferAgent(ctx, addrWithTranWithdraw)),
			from: rMarkerNoAttr.GetAddress(),
			to:   addrWithAttrs,
			amt:  cz(c(3, rDenomNoAttr)),
		},
		{
			name: "with admin: does not have transfer: okay otherwise",
			ctx:  ctxP(types.WithTransferAgent(ctx, addrOther)),
			from: owner,
			to:   addrWithAttrs,
			amt:  cz(c(1, rDenom1Attr), c(1, nrDenom)),
		},
		{
			name: "with admin: has transfer: would otherwise fail",
			ctx:  ctxP(types.WithTransferAgent(ctx, addrWithTransfer)),
			from: addrWithDenySend,
			to:   addrWithAttrs,
			amt:  cz(c(1, rDenomNoAttr)),
		},
		{
			name:   "from marker to marker: admin only has transfer",
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithTransfer)),
			from:   rMarkerNoAttr.GetAddress(),
			to:     rMarker1Attr.GetAddress(),
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: noAccessErr(addrWithTransfer, types.Access_Withdraw, rDenomNoAttr),
		},
		{
			name:   "from marker to marker: admin only has deposit",
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithDeposit)),
			from:   rMarkerNoAttr.GetAddress(),
			to:     rMarker1Attr.GetAddress(),
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: noAccessErr(addrWithDeposit, types.Access_Withdraw, rDenomNoAttr),
		},
		{
			name:   "from marker to marker: admin only has withdraw",
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithWithdraw)),
			from:   rMarkerNoAttr.GetAddress(),
			to:     rMarker1Attr.GetAddress(),
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: noAccessErr(addrWithWithdraw, types.Access_Deposit, rDenom1Attr),
		},
		{
			name:   "from marker to marker: admin only has transfer and deposit",
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithTranDep)),
			from:   rMarkerNoAttr.GetAddress(),
			to:     rMarker1Attr.GetAddress(),
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: noAccessErr(addrWithTranDep, types.Access_Withdraw, rDenomNoAttr),
		},
		{
			name:   "from marker to marker: admin only has transfer and withdraw",
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithTranWithdraw)),
			from:   rMarkerNoAttr.GetAddress(),
			to:     rMarker1Attr.GetAddress(),
			amt:    cz(c(1, rDenom1AttrNoOneHas)),
			expErr: noAccessErr(addrWithTranWithdraw, types.Access_Deposit, rDenom1Attr),
		},
		{
			name:   "from marker to marker: admin only has deposit and withdraw",
			ctx:    ctxP(types.WithTransferAgent(ctx, addrWithDepWithdraw)),
			from:   rMarker1Attr.GetAddress(),
			to:     rMarker2Attrs.GetAddress(),
			amt:    cz(c(1, rDenom1Attr)),
			expErr: noAccessErr(addrWithDepWithdraw, types.Access_Transfer, rDenom1Attr),
		},
		{
			name: "from marker to marker: admin has transfer and deposit and withdraw",
			ctx:  ctxP(types.WithTransferAgent(ctx, addrWithTranDepWithdraw)),
			from: rMarker1Attr.GetAddress(),
			to:   rMarker2Attrs.GetAddress(),
			amt:  cz(c(1, rDenomNoAttr)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tCtx := ctx
			if tc.ctx != nil {
				tCtx = *tc.ctx
			}
			kpr := app.MarkerKeeper
			if tc.attrKeeper != nil {
				kpr = kpr.WithAttrKeeper(tc.attrKeeper.WithParent(app.AttributeKeeper))
			}
			newTo, err := kpr.SendRestrictionFn(tCtx, tc.from, tc.to, tc.amt)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "SendRestrictionFn error")
			} else {
				assert.NoError(t, err, "SendRestrictionFn error")
				assert.Equal(t, tc.to, newTo, "SendRestrictionFn returned address")
			}
		})
	}
}

func TestBankSendCoinsUsesSendRestrictionFn(t *testing.T) {
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
	ctx := app.BaseApp.NewContext(false)
	app.MarkerKeeper.AddMarkerAccount(ctx, types.NewEmptyMarkerAccount("navcoin", addrNameOwner.String(), []types.AccessGrant{}))
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

	// sendWithCache uses a cache context to call SendCoins, and only writes it if there wasn't an error.
	// This is needed because SendCoins subtracts the amount from the fromAddr before checking the send restriction.
	// That's usually okay because it's all being done in a single transaction, which gets rolled back on error.
	sendWithCache := func(fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
		cacheCtx, writeCache := ctx.CacheContext()
		err = app.BankKeeper.SendCoins(cacheCtx, fromAddr, toAddr, amt)
		if err == nil {
			writeCache()
		}
		return err
	}

	t.Run("send to address without attributes", func(t *testing.T) {
		expErr := fmt.Sprintf("address %s does not contain the %q required attribute: \"kyc.provenance.io\"",
			addrOther, markerDenom)
		err = sendWithCache(addrHasAttr, addrOther, cz(5, markerDenom))
		assert.EqualError(t, err, expErr, "SendCoins")
		expBal := cz(100, markerDenom)
		hasAttrBal := app.BankKeeper.GetBalance(ctx, addrHasAttr, markerDenom)
		assert.Equal(t, expBal.String(), hasAttrBal.String(), "GetBalance addrHasAttr")
		otherBal := app.BankKeeper.GetBalance(ctx, addrOther, markerDenom)
		assert.Equal(t, expBal.String(), otherBal.String(), "GetBalance addrOther")
	})

	t.Run("send to address with attributes", func(t *testing.T) {
		err = sendWithCache(addrOther, addrHasAttr, cz(6, markerDenom))
		assert.NoError(t, err, "SendCoins")
		hasAttrExpBal := cz(106, markerDenom)
		hasAttrBal := app.BankKeeper.GetBalance(ctx, addrHasAttr, markerDenom)
		assert.Equal(t, hasAttrExpBal.String(), hasAttrBal.String(), "GetBalance addrHasAttr")
		otherExpBal := cz(94, markerDenom)
		otherBal := app.BankKeeper.GetBalance(ctx, addrOther, markerDenom)
		assert.Equal(t, otherExpBal.String(), otherBal.String(), "GetBalance addrOther")
	})
}

func TestBankInputOutputCoinsUsesSendRestrictionFn(t *testing.T) {
	// This test only checks that the marker SendRestrictionFn is applied during a InputOutputCoins.
	// Testing of the actual SendRestrictionFn is assumed to be done elsewhere more extensively.

	markerDenom := "cowcoin"
	cz := func(amt int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(markerDenom, amt))
	}

	addrManager := sdk.AccAddress("addrManager_________")
	addrInput := sdk.AccAddress("addrInput___________")
	addrOutput1 := sdk.AccAddress("addrOutput1_________")
	addrOutput2 := sdk.AccAddress("addrOutput2_________")
	addrWithoutTransfer := sdk.AccAddress("addrWithoutTransfer_")
	addrWithAttr1 := sdk.AccAddress("addrWithAttr1_______")
	addrWithAttr2 := sdk.AccAddress("addrWithAttr2_______")

	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	app.MarkerKeeper.AddMarkerAccount(ctx, types.NewEmptyMarkerAccount("navcoin", addrManager.String(), []types.AccessGrant{}))
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, addrManager))
	err := app.NameKeeper.SetNameRecord(ctx, "rando.io", addrManager, false)
	require.NoError(t, err, "SetNameRecord rando.io")
	require.NoError(t, app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "rando.io",
			Value:         []byte("random value 1"),
			Address:       addrWithAttr1.String(),
			AttributeType: attrTypes.AttributeType_String,
		},
		addrManager,
	), "SetAttribute rando.io on addrWithAttr1")
	require.NoError(t, app.AttributeKeeper.SetAttribute(ctx,
		attrTypes.Attribute{
			Name:          "rando.io",
			Value:         []byte("random value 2"),
			Address:       addrWithAttr2.String(),
			AttributeType: attrTypes.AttributeType_String,
		},
		addrManager,
	), "SetAttribute rando.io on addrWithAttr2")

	makeMarkerMsg := &types.MsgAddFinalizeActivateMarkerRequest{
		Amount:      sdk.NewInt64Coin(markerDenom, 1000),
		Manager:     addrManager.String(),
		FromAddress: addrManager.String(),
		MarkerType:  types.MarkerType_RestrictedCoin,
		AccessList: []types.AccessGrant{
			{Address: addrManager.String(), Permissions: types.AccessList{
				types.Access_Mint, types.Access_Burn,
				types.Access_Deposit, types.Access_Withdraw,
				types.Access_Delete, types.Access_Admin, types.Access_Transfer,
			}},
		},
		SupplyFixed:            true,
		AllowGovernanceControl: true,
		AllowForcedTransfer:    false,
		RequiredAttributes:     []string{"rando.io"},
	}
	markerHandler := marker.NewHandler(app.MarkerKeeper)
	_, err = markerHandler(ctx, makeMarkerMsg)
	require.NoError(t, err, "MsgAddFinalizeActivateMarkerRequest")
	err = app.MarkerKeeper.WithdrawCoins(ctx, addrManager, addrManager, markerDenom, cz(100))
	require.NoError(t, err, "WithdrawCoins to addrInput")
	err = app.MarkerKeeper.WithdrawCoins(ctx, addrManager, addrInput, markerDenom, cz(100))
	require.NoError(t, err, "WithdrawCoins to addrInput")
	err = app.MarkerKeeper.WithdrawCoins(ctx, addrManager, addrWithoutTransfer, markerDenom, cz(100))
	require.NoError(t, err, "WithdrawCoins to addrWithoutTransfer")

	type expBal struct {
		name string
		addr sdk.AccAddress
		bal  sdk.Coins
	}
	newExpBal := func(name string, addr sdk.AccAddress, bal sdk.Coins) expBal {
		return expBal{
			name: name,
			addr: addr,
			bal:  bal,
		}
	}
	assertBalance := func(t *testing.T, exp expBal) bool {
		t.Helper()
		bal := app.BankKeeper.GetBalance(ctx, exp.addr, markerDenom)
		return assert.Equal(t, exp.bal.String(), bal.String(), "GetBalance %s", exp.name)
	}

	noAttrErr := func(addr sdk.AccAddress) string {
		return fmt.Sprintf("address %s does not contain the %q required attribute: %q",
			addr.String(), markerDenom, "rando.io")
	}

	tests := []struct {
		name    string
		input   banktypes.Input
		outputs []banktypes.Output
		expErr  string
		expBals []expBal
	}{
		{
			name:  "from address with transfer permission",
			input: banktypes.Input{Address: addrManager.String(), Coins: cz(99)},
			outputs: []banktypes.Output{
				{Address: addrOutput1.String(), Coins: cz(33)},
				{Address: addrOutput2.String(), Coins: cz(66)},
			},
			expErr: "",
			expBals: []expBal{
				newExpBal("addrManager", addrManager, cz(1)),
				newExpBal("addrOutput1", addrOutput1, cz(33)),
				newExpBal("addrOutput2", addrOutput2, cz(66)),
			},
		},
		{
			name:  "from address without transfer permission",
			input: banktypes.Input{Address: addrInput.String(), Coins: cz(100)},
			outputs: []banktypes.Output{
				{Address: addrOutput1.String(), Coins: cz(60)},
				{Address: addrOutput2.String(), Coins: cz(40)},
			},
			expErr: noAttrErr(addrOutput1),
			// Note: The input coins are subtracted before running the restriction function.
			//       Usually this is done in a transaction so the error would roll it back.
			//       Here, we just skip checking that balance.
			expBals: []expBal{
				newExpBal("addrOutput1", addrOutput1, cz(33)), // from previous test
				newExpBal("addrOutput2", addrOutput2, cz(66)), // from previous test
			},
		},
		{
			name:  "to addresses with attributes",
			input: banktypes.Input{Address: addrWithoutTransfer.String(), Coins: cz(77)},
			outputs: []banktypes.Output{
				{Address: addrWithAttr1.String(), Coins: cz(33)},
				{Address: addrWithAttr2.String(), Coins: cz(44)},
			},
			expErr: "",
			expBals: []expBal{
				newExpBal("addrWithoutTransfer", addrWithoutTransfer, cz(23)),
				newExpBal("addrWithAttr1", addrWithAttr1, cz(33)),
				newExpBal("addrWithAttr2", addrWithAttr2, cz(44)),
			},
		},
		{
			name:  "to one address with and one without",
			input: banktypes.Input{Address: addrWithoutTransfer.String(), Coins: cz(20)},
			outputs: []banktypes.Output{
				{Address: addrWithAttr1.String(), Coins: cz(3)},
				{Address: addrOutput2.String(), Coins: cz(17)},
			},
			expErr: noAttrErr(addrOutput2),
			// Note: Here too, the input should come out and the first output go through before getting the error.
			//       Normally, that'd get rolled back because of the error, but we're not in a Tx here.
			//       So all I'm going to do is check that the last output didn't go through.
			expBals: []expBal{newExpBal("addrOutput2", addrOutput2, cz(66))}, // from first test.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// err = app.BankKeeper.InputOutputCoins(ctx, []banktypes.Input{tc.input}, tc.outputs) // TODO[1760]: bank
			if len(tc.expErr) != 0 {
				assert.EqualError(t, err, tc.expErr, "InputOutputCoins")
			} else {
				assert.NoError(t, err, "InputOutputCoins")
			}

			for _, exp := range tc.expBals {
				assertBalance(t, exp)
			}
		})
	}
}

func TestNormalizeRequiredAttributes(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

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

func TestQuarantineOfRestrictedCoins(t *testing.T) {
	// Directly tests the bug described in https://github.com/provenance-io/provenance/issues/1626

	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	owner := sdk.AccAddress("owner_address_______")
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, owner))
	reqAttr := "quarantinetest.provenance.io"
	require.NoError(t, app.NameKeeper.SetNameRecord(ctx, reqAttr, owner, false), "SetNameRecord(%q)", reqAttr)

	// Two source addresses, one with transfer on both markers, one without on either.
	addrWithTransfer := sdk.AccAddress("addrWithTransfer____")
	addrWithWithdraw := sdk.AccAddress("addrWithWithdraw____")
	addrWithoutTransfer := sdk.AccAddress("addrWithoutTransfer_")
	nav := []types.NetAssetValue{types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, int64(1)), 1)}

	newMarker := func(denom string, reqAttrs []string) *types.MarkerAccount {
		rv := types.NewMarkerAccount(
			app.AccountKeeper.NewAccountWithAddress(ctx, types.MustGetMarkerAddress(denom)).(*authtypes.BaseAccount),
			sdk.NewInt64Coin(denom, 1000),
			owner,
			[]types.AccessGrant{
				{Address: addrWithTransfer.String(), Permissions: types.AccessList{types.Access_Transfer}},
				{Address: addrWithWithdraw.String(), Permissions: types.AccessList{types.Access_Withdraw}},
			},
			types.StatusProposed,
			types.MarkerType_RestrictedCoin,
			true,  // supply fixed
			true,  // allow gov
			false, // no force transfer
			reqAttrs,
		)
		err := app.MarkerKeeper.AddSetNetAssetValues(ctx, rv, nav, types.ModuleName)
		require.NoError(t, err, "AddSetNetAssetValues(%v, %v, %v)", rv, nav, types.ModuleName)
		err = app.MarkerKeeper.AddFinalizeAndActivateMarker(ctx, rv)
		require.NoError(t, err, "AddFinalizeAndActivateMarker(%s)", denom)
		return rv
	}

	// Two markers, one with a required attribute, one without any.
	denomNoReqAttr := "denomNoReqAttr"
	denom1ReqAttr := "denom1ReqAttr"

	coinsNoReqAttr := sdk.NewCoins(sdk.NewInt64Coin(denomNoReqAttr, 3))
	coins1ReqAttr := sdk.NewCoins(sdk.NewInt64Coin(denom1ReqAttr, 2))

	newMarker(denomNoReqAttr, nil)
	newMarker(denom1ReqAttr, []string{reqAttr})

	mustWithdraw := func(recipient sdk.AccAddress, denom string) {
		coins := sdk.NewCoins(sdk.NewInt64Coin(denom, 100))
		err := app.MarkerKeeper.WithdrawCoins(ctx, addrWithWithdraw, recipient, denom, coins)
		require.NoError(t, err, "WithdrawCoins(%q, %q)", string(recipient), coins)
	}
	mustWithdraw(addrWithTransfer, denomNoReqAttr)
	mustWithdraw(addrWithTransfer, denom1ReqAttr)
	mustWithdraw(addrWithoutTransfer, denomNoReqAttr)
	mustWithdraw(addrWithoutTransfer, denom1ReqAttr)

	// Create two quarantined address: one with the required attributes, one without.
	optIn := func(t *testing.T, addr sdk.AccAddress) {
		// require.NoError(t, app.QuarantineKeeper.SetOptIn(ctx, addr), "SetOptIn(%q)", string(addr)) // TODO[1760]: quarantine
	}
	addrQWithAttr := sdk.AccAddress("addrQWithReqAttrs____")
	addrQWithoutAttr := sdk.AccAddress("addrQWithoutReqAttrs____")
	optIn(t, addrQWithAttr)
	optIn(t, addrQWithoutAttr)

	attrVal := []byte("string value")
	setAttr := func(t *testing.T, addr sdk.AccAddress) {
		attr := attrTypes.Attribute{
			Name:          reqAttr,
			Value:         attrVal,
			Address:       addr.String(),
			AttributeType: attrTypes.AttributeType_String,
		}
		err := app.AttributeKeeper.SetAttribute(ctx, attr, owner)
		require.NoError(t, err, "SetAttribute(%q, %q)", string(addr), attr.Name)
	}
	setAttr(t, addrQWithAttr)

	noTransErr := addrWithoutTransfer.String() + " does not have transfer permissions for " + denomNoReqAttr
	noAttrErr := func(addr sdk.AccAddress) string {
		return fmt.Sprintf("address %s does not contain the %q required attribute: %q", addr, denom1ReqAttr, reqAttr)
	}

	quarantineModAddr := authtypes.NewModuleAddress("quarantine")

	tests := []struct {
		name         string
		fromAddr     sdk.AccAddress
		toAddr       sdk.AccAddress
		amt          sdk.Coins
		expSendErr   string
		expAcceptErr string
	}{
		{
			name:     "no req attrs from addr with transfer to quarantined",
			fromAddr: addrWithTransfer,
			toAddr:   addrQWithoutAttr,
			amt:      coinsNoReqAttr,
		},
		{
			name:       "no req attrs from addr without transfer to quarantined",
			fromAddr:   addrWithoutTransfer,
			toAddr:     addrQWithoutAttr,
			amt:        coinsNoReqAttr,
			expSendErr: noTransErr,
		},
		{
			name:         "with req attrs from addr with transfer to quarantined without attrs",
			fromAddr:     addrWithTransfer,
			toAddr:       addrQWithoutAttr,
			amt:          coins1ReqAttr,
			expSendErr:   "",
			expAcceptErr: noAttrErr(addrQWithoutAttr),
		},
		{
			name:     "with req attrs from addr with transfer to quarantined with attrs",
			fromAddr: addrWithTransfer,
			toAddr:   addrQWithAttr,
			amt:      coins1ReqAttr,
		},
		{
			name:       "with req attrs from addr without transfer to quarantined without attrs",
			fromAddr:   addrWithoutTransfer,
			toAddr:     addrQWithoutAttr,
			amt:        coins1ReqAttr,
			expSendErr: noAttrErr(addrQWithoutAttr),
		},
		{
			name:     "with req attrs from addr without transfer to quarantined with attrs",
			fromAddr: addrWithoutTransfer,
			toAddr:   addrQWithAttr,
			amt:      coins1ReqAttr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if t.Failed() {
					t.Logf("fromAddr: %s", tc.fromAddr)
					t.Logf("  toAddr: %s", tc.toAddr)
					t.Logf("quarantine module address: %s", quarantineModAddr)
				}
			}()
			sendErr := app.BankKeeper.SendCoins(ctx, tc.fromAddr, tc.toAddr, tc.amt)
			if len(tc.expSendErr) != 0 {
				require.EqualError(t, sendErr, tc.expSendErr, "SendCoins")
			} else {
				require.NoError(t, sendErr, "SendCoins")
			}
			if sendErr != nil {
				return
			}

			// TODO[1760]: quarantine: Uncomment once it's back.
			/*
				amt, acceptErr := app.QuarantineKeeper.AcceptQuarantinedFunds(ctx, tc.toAddr, tc.fromAddr)
				if len(tc.expAcceptErr) != 0 {
					require.EqualError(t, acceptErr, tc.expAcceptErr, "AcceptQuarantinedFunds")
				} else {
					require.NoError(t, acceptErr, "AcceptQuarantinedFunds")
					assert.Equal(t, tc.amt.String(), amt.String(), "accepted quarantined funds")
				}
			*/
		})
	}

	t.Run("attr deleted after funds quarantined", func(t *testing.T) {
		fromAddr := addrWithoutTransfer
		toAddr := sdk.AccAddress("addr_attr_del_______")
		amt := coins1ReqAttr
		optIn(t, toAddr)
		setAttr(t, toAddr)
		sendErr := app.BankKeeper.SendCoins(ctx, fromAddr, toAddr, amt)
		require.NoError(t, sendErr, "SendCoins")
		delErr := app.AttributeKeeper.DeleteAttribute(ctx, toAddr.String(), reqAttr, &attrVal, owner)
		require.NoError(t, delErr, "DeleteAttribute")
		// TODO[1760]: quarantine: Uncomment once it's back.
		// expAcceptErr := noAttrErr(toAddr)
		// _, acceptErr := app.QuarantineKeeper.AcceptQuarantinedFunds(ctx, toAddr, fromAddr)
		// require.EqualError(t, acceptErr, expAcceptErr, "AcceptQuarantinedFunds")
	})

	t.Run("attr added after funds quarantined", func(t *testing.T) {
		fromAddr := addrWithTransfer
		toAddr := sdk.AccAddress("addr_attr_add_______")
		amt := coins1ReqAttr
		optIn(t, toAddr)
		sendErr := app.BankKeeper.SendCoins(ctx, fromAddr, toAddr, amt)
		require.NoError(t, sendErr, "SendCoins")
		setAttr(t, toAddr)
		// TODO[1760]: quarantine: Uncomment once it's back.
		/*
			acceptedAmt, acceptErr := app.QuarantineKeeper.AcceptQuarantinedFunds(ctx, toAddr, fromAddr)
			require.NoError(t, acceptErr, "AcceptQuarantinedFunds error")
			assert.Equal(t, amt.String(), acceptedAmt.String(), "AcceptQuarantinedFunds amount")
		*/
	})

	t.Run("marker restriction applied before quarantine", func(t *testing.T) {
		// This test makes sure that the marker SendRestrictionFn is being applied before the quarantine one.
		// If the quarantine one is applied first, then the toAddr in the marker's restriction will be the
		// quarantine module, which will have a bypass.
		// So we attempt to send from the addr without transfer permission to the address without the required attribute.
		// If we get an error about the attribute not being there, we're good.
		// If we don't get an error, the toAddr was probably the quarantine module which bypasses that attribute check.
		fromAddr := addrWithoutTransfer
		toAddr := addrQWithoutAttr
		amt := coins1ReqAttr

		err := app.BankKeeper.SendCoins(ctx, fromAddr, toAddr, amt)
		require.Error(t, err, "SendCoins error\n"+
			"If this assertion fails, it's probably because the quarantine\n"+
			"SendRestrictionFn is being applied before the marker's")
		require.EqualError(t, err, noAttrErr(toAddr))
	})
}
