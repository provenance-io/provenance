package exchange

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestAllMsgsGetSigners(t *testing.T) {
	// getTypeName gets just the type name of the provided thing, e.g. "MsgGovCreateMarketRequest".
	getTypeName := func(thing interface{}) string {
		rv := fmt.Sprintf("%T", thing) // e.g. "*types.MsgGovCreateMarketRequest"
		lastDot := strings.LastIndex(rv, ".")
		if lastDot < 0 || lastDot+1 >= len(rv) {
			return rv
		}
		return rv[lastDot+1:]
	}

	testAddr := sdk.AccAddress("testAddr____________")
	badAddrStr := "badaddr"
	badAddrErr := "decoding bech32 failed: invalid bech32 string length 7"
	emptyAddrErr := "empty address string is not allowed"

	msgMakers := []func(signer string) sdk.Msg{
		func(signer string) sdk.Msg {
			return &MsgCreateAskRequest{AskOrder: AskOrder{Seller: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgCreateBidRequest{BidOrder: BidOrder{Buyer: signer}}
		},
		func(signer string) sdk.Msg {
			return &MsgCancelOrderRequest{Owner: signer}
		},
		// TODO[1658]: Add MsgFillBidsRequest once it's actually been defined.
		// TODO[1658]: Add MsgFillAsksRequest once it's actually been defined.
		func(signer string) sdk.Msg {
			return &MsgMarketSettleRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketWithdrawRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketUpdateDetailsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketUpdateEnabledRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketUpdateUserSettleRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketManagePermissionsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgMarketManageReqAttrsRequest{Admin: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovCreateMarketRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovManageFeesRequest{Authority: signer}
		},
		func(signer string) sdk.Msg {
			return &MsgGovUpdateParamsRequest{Authority: signer}
		},
	}

	signerCases := []struct {
		name       string
		msgSigner  string
		expSigners []sdk.AccAddress
		expPanic   string
	}{
		{
			name:      "no signer",
			msgSigner: "",
			expPanic:  emptyAddrErr,
		},
		{
			name:       "good signer",
			msgSigner:  testAddr.String(),
			expSigners: []sdk.AccAddress{testAddr},
		},
		{
			name:      "bad signer",
			msgSigner: badAddrStr,
			expPanic:  badAddrErr,
		},
	}

	type testCase struct {
		name       string
		msg        sdk.Msg
		expSigners []sdk.AccAddress
		expPanic   string
	}

	var tests []testCase
	hasMaker := make(map[string]bool)

	for _, msgMaker := range msgMakers {
		typeName := getTypeName(msgMaker(""))
		hasMaker[typeName] = true
		for _, tc := range signerCases {
			tests = append(tests, testCase{
				name:       typeName + " " + tc.name,
				msg:        msgMaker(tc.msgSigner),
				expSigners: tc.expSigners,
				expPanic:   tc.expPanic,
			})
		}
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var signers []sdk.AccAddress
			testFunc := func() {
				signers = tc.msg.GetSigners()
			}

			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetSigners")
			assert.Equal(t, tc.expSigners, signers, "GetSigners")
		})
	}

	// Make sure all of the GetSigners and GetSignerStrs funcs are tested.
	t.Run("all msgs have test case", func(t *testing.T) {
		for _, msg := range allRequestMsgs {
			typeName := getTypeName(msg)
			t.Run(typeName, func(t *testing.T) {
				// If this fails, a maker needs to be defined above for the missing msg type.
				assert.True(t, hasMaker[typeName], "There is not a GetSigners test case for %s", typeName)
			})
		}
	})
}

func testValidateBasic(t *testing.T, msg sdk.Msg, expErr []string) {
	t.Helper()
	var err error
	testFunc := func() {
		err = msg.ValidateBasic()
	}
	require.NotPanics(t, testFunc, "%T.ValidateBasic()", msg)

	assertions.AssertErrorContents(t, err, expErr, "%T.ValidateBasic() error", msg)
}

func TestMsgCreateAskRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreateAskRequest
		expErr []string
	}{
		{
			name: "empty ask order",
			msg:  MsgCreateAskRequest{},
			expErr: []string{
				"invalid market id: ",
				"invalid seller: ",
				"invalid price: ",
				"invalid assets: ",
			},
		},
		{
			name: "invalid fees",
			msg: MsgCreateAskRequest{
				AskOrder: AskOrder{
					MarketId: 1,
					Seller:   sdk.AccAddress("seller______________").String(),
					Assets:   sdk.NewCoins(sdk.NewInt64Coin("banana", 99)),
					Price:    sdk.NewInt64Coin("acorn", 12),
				},
				OrderCreationFee: &sdk.Coin{Denom: "cactus", Amount: sdkmath.NewInt(-3)},
			},
			expErr: []string{"invalid order creation fee: negative coin amount: -3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgCreateBidRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreateBidRequest
		expErr []string
	}{
		{
			name: "empty ask order",
			msg:  MsgCreateBidRequest{},
			expErr: []string{
				"invalid market id: ",
				"invalid buyer: ",
				"invalid price: ",
				"invalid assets: ",
			},
		},
		{
			name: "invalid fees",
			msg: MsgCreateBidRequest{
				BidOrder: BidOrder{
					MarketId: 1,
					Buyer:    sdk.AccAddress("buyer_______________").String(),
					Assets:   sdk.NewCoins(sdk.NewInt64Coin("banana", 99)),
					Price:    sdk.NewInt64Coin("acorn", 12),
				},
				OrderCreationFee: &sdk.Coin{Denom: "cactus", Amount: sdkmath.NewInt(-3)},
			},
			expErr: []string{"invalid order creation fee: negative coin amount: -3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgCancelOrderRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCancelOrderRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgCancelOrderRequest{
				Owner:   sdk.AccAddress("owner_______________").String(),
				OrderId: 1,
			},
			expErr: nil,
		},
		{
			name: "missing owner",
			msg: MsgCancelOrderRequest{
				Owner:   "",
				OrderId: 1,
			},
			expErr: []string{"invalid owner: ", "empty address string is not allowed"},
		},
		{
			name: "invalid owner",
			msg: MsgCancelOrderRequest{
				Owner:   "notgonnawork",
				OrderId: 1,
			},
			expErr: []string{"invalid owner: ", "decoding bech32 failed: invalid separator index -1"},
		},
		{
			name: "order 0",
			msg: MsgCancelOrderRequest{
				Owner:   sdk.AccAddress("valid_owner_________").String(),
				OrderId: 0,
			},
			expErr: []string{"invalid order id: cannot be zero"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

// TODO[1658]: func TestMsgFillBidsRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgFillAsksRequest_ValidateBasic(t *testing.T)

// TODO[1658]: func TestMsgMarketSettleRequest_ValidateBasic(t *testing.T)

func TestMsgMarketWithdrawRequest_ValidateBasic(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	goodAdminAddr := sdk.AccAddress("Administrator_______").String()
	goodToAddr := sdk.AccAddress("ToAddress___________").String()
	goodCoins := sdk.NewCoins(coin(3, "acorns"))

	tests := []struct {
		name   string
		msg    MsgMarketWithdrawRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    goodCoins,
			},
			expErr: nil,
		},
		{
			name: "no administrator",
			msg: MsgMarketWithdrawRequest{
				Admin:     "",
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    goodCoins,
			},
			expErr: []string{`invalid administrator ""`, "empty address string is not allowed"},
		},
		{
			name: "bad administrator",
			msg: MsgMarketWithdrawRequest{
				Admin:     "notright",
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    goodCoins,
			},
			expErr: []string{`invalid administrator "notright"`, "decoding bech32 failed"},
		},
		{
			name: "market id zero",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  0,
				ToAddress: goodToAddr,
				Amount:    goodCoins,
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "no to address",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: "",
				Amount:    goodCoins,
			},
			expErr: []string{`invalid to address ""`, "empty address string is not allowed"},
		},
		{
			name: "bad to address",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: "notright",
				Amount:    goodCoins,
			},
			expErr: []string{`invalid to address "notright"`, "decoding bech32 failed"},
		},
		{
			name: "invalid denom in amount",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    sdk.Coins{coin(3, "x")},
			},
			expErr: []string{`invalid amount "3x"`, "invalid denom: x"},
		},
		{
			name: "negative amount",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    sdk.Coins{coin(-1, "negcoin")},
			},
			expErr: []string{`invalid amount "-1negcoin"`, "coin -1negcoin amount is not positive"},
		},
		{
			name: "empty amount",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    sdk.Coins{},
			},
			expErr: []string{`invalid amount ""`, "cannot be zero"},
		},
		{
			name: "zero coin amount",
			msg: MsgMarketWithdrawRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				ToAddress: goodToAddr,
				Amount:    sdk.Coins{coin(0, "acorn"), coin(0, "banana")},
			},
			expErr: []string{`invalid amount "0acorn,0banana"`, "coin 0acorn amount is not positive"},
		},
		{
			name: "multiple errors",
			msg:  MsgMarketWithdrawRequest{},
			expErr: []string{
				"invalid administrator",
				"invalid market id",
				"invalid to address",
				"invalid amount",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketUpdateDetailsRequest_ValidateBasic(t *testing.T) {
	admin := sdk.AccAddress("admin_______________").String()
	tooLongErr := func(field string, max int) string {
		return fmt.Sprintf("%s length %d exceeds maximum length of %d", field, max+1, max)
	}

	tests := []struct {
		name   string
		msg    MsgMarketUpdateDetailsRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					Name:        "MyMarket",
					Description: "This is my own market only for me.",
					WebsiteUrl:  "https://example.com",
					IconUri:     "https://example.com/icon",
				},
			},
			expErr: nil,
		},
		{
			name: "empty admin",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:         "",
				MarketId:      1,
				MarketDetails: MarketDetails{},
			},
			expErr: []string{`invalid administrator ""`, "empty address string is not allowed"},
		},
		{
			name: "invalid admin",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:         "notvalidadmin",
				MarketId:      1,
				MarketDetails: MarketDetails{},
			},
			expErr: []string{`invalid administrator "notvalidadmin"`, "decoding bech32 failed"},
		},
		{
			name: "market id zero",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:         admin,
				MarketId:      0,
				MarketDetails: MarketDetails{},
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "name too long",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					Name: strings.Repeat("p", MaxName+1),
				},
			},
			expErr: []string{tooLongErr("name", MaxName)},
		},
		{
			name: "description too long",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					Description: strings.Repeat("d", MaxDescription+1),
				},
			},
			expErr: []string{tooLongErr("description", MaxDescription)},
		},
		{
			name: "website_url too long",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					WebsiteUrl: strings.Repeat("w", MaxWebsiteURL+1),
				},
			},
			expErr: []string{tooLongErr("website_url", MaxWebsiteURL)},
		},
		{
			name: "icon_uri too long",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    admin,
				MarketId: 1,
				MarketDetails: MarketDetails{
					IconUri: strings.Repeat("i", MaxIconURI+1),
				},
			},
			expErr: []string{tooLongErr("icon_uri", MaxIconURI)},
		},
		{
			name: "multiple errors",
			msg: MsgMarketUpdateDetailsRequest{
				Admin:    "",
				MarketId: 0,
				MarketDetails: MarketDetails{
					Name:        strings.Repeat("p", MaxName+1),
					Description: strings.Repeat("d", MaxDescription+1),
					WebsiteUrl:  strings.Repeat("w", MaxWebsiteURL+1),
					IconUri:     strings.Repeat("i", MaxIconURI+1),
				},
			},
			expErr: []string{
				"invalid administrator",
				"invalid market id",
				tooLongErr("name", MaxName),
				tooLongErr("description", MaxDescription),
				tooLongErr("website_url", MaxWebsiteURL),
				tooLongErr("icon_uri", MaxIconURI),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketUpdateEnabledRequest_ValidateBasic(t *testing.T) {
	admin := sdk.AccAddress("admin_______________").String()

	tests := []struct {
		name   string
		msg    MsgMarketUpdateEnabledRequest
		expErr []string
	}{
		{
			name: "control: true",
			msg: MsgMarketUpdateEnabledRequest{
				Admin:           admin,
				MarketId:        1,
				AcceptingOrders: true,
			},
			expErr: nil,
		},
		{
			name: "control: false",
			msg: MsgMarketUpdateEnabledRequest{
				Admin:           admin,
				MarketId:        1,
				AcceptingOrders: true,
			},
			expErr: nil,
		},
		{
			name: "empty admin",
			msg: MsgMarketUpdateEnabledRequest{
				Admin:    "",
				MarketId: 1,
			},
			expErr: []string{
				`invalid administrator ""`, "empty address string is not allowed",
			},
		},
		{
			name: "bad admin",
			msg: MsgMarketUpdateEnabledRequest{
				Admin:    "badadmin",
				MarketId: 1,
			},
			expErr: []string{
				`invalid administrator "badadmin"`, "decoding bech32 failed",
			},
		},
		{
			name: "market id zero",
			msg: MsgMarketUpdateEnabledRequest{
				Admin:    admin,
				MarketId: 0,
			},
			expErr: []string{
				"invalid market id", "cannot be zero",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketUpdateUserSettleRequest_ValidateBasic(t *testing.T) {
	admin := sdk.AccAddress("admin_______________").String()

	tests := []struct {
		name   string
		msg    MsgMarketUpdateUserSettleRequest
		expErr []string
	}{
		{
			name: "control: true",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:               admin,
				MarketId:            1,
				AllowUserSettlement: true,
			},
			expErr: nil,
		},
		{
			name: "control: false",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:               admin,
				MarketId:            1,
				AllowUserSettlement: true,
			},
			expErr: nil,
		},
		{
			name: "empty admin",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:    "",
				MarketId: 1,
			},
			expErr: []string{
				`invalid administrator ""`, "empty address string is not allowed",
			},
		},
		{
			name: "bad admin",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:    "badadmin",
				MarketId: 1,
			},
			expErr: []string{
				`invalid administrator "badadmin"`, "decoding bech32 failed",
			},
		},
		{
			name: "market id zero",
			msg: MsgMarketUpdateUserSettleRequest{
				Admin:    admin,
				MarketId: 0,
			},
			expErr: []string{
				"invalid market id", "cannot be zero",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketManagePermissionsRequest_ValidateBasic(t *testing.T) {
	goodAdminAddr := sdk.AccAddress("goodAdminAddr_______").String()
	goodAddr1 := sdk.AccAddress("goodAddr1___________").String()
	goodAddr2 := sdk.AccAddress("goodAddr2___________").String()
	goodAddr3 := sdk.AccAddress("goodAddr3___________").String()

	tests := []struct {
		name   string
		msg    MsgMarketManagePermissionsRequest
		expErr []string
	}{
		{
			name: "control",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				RevokeAll: []string{goodAddr1},
				ToRevoke:  []AccessGrant{{Address: goodAddr2, Permissions: []Permission{Permission_settle}}},
				ToGrant:   []AccessGrant{{Address: goodAddr3, Permissions: []Permission{Permission_cancel}}},
			},
			expErr: nil,
		},
		{
			name: "empty admin",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     "",
				MarketId:  1,
				RevokeAll: []string{goodAddr1},
			},
			expErr: []string{`invalid administrator ""`, "empty address string is not allowed"},
		},
		{
			name: "invalid admin",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     "bad1admin",
				MarketId:  1,
				RevokeAll: []string{goodAddr1},
			},
			expErr: []string{`invalid administrator "bad1admin"`, "decoding bech32 failed"},
		},
		{
			name: "market id zero",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  0,
				RevokeAll: []string{goodAddr1},
			},
			expErr: []string{"invalid market id", "cannot be zero"},
		},
		{
			name: "no updates",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
			},
			expErr: []string{"no updates"},
		},
		{
			name: "two invalid addresses in revoke all",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				RevokeAll: []string{"bad1addr", "bad2addr"},
			},
			expErr: []string{
				`invalid revoke-all address "bad1addr": decoding bech32 failed`,
				`invalid revoke-all address "bad2addr": decoding bech32 failed`,
			},
		},
		{
			name: "two invalid to-revoke entries",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
				ToRevoke: []AccessGrant{
					{Address: "badaddr", Permissions: []Permission{Permission_withdraw}},
					{Address: goodAddr1, Permissions: []Permission{Permission_unspecified}},
				},
			},
			expErr: []string{
				`invalid to-revoke access grant: invalid address "badaddr"`,
				"invalid to-revoke access grant: permission is unspecified for " + goodAddr1,
			},
		},
		{
			name: "two addrs in both revoke-all and to-revoke",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				RevokeAll: []string{goodAddr1, goodAddr2, goodAddr3},
				ToRevoke: []AccessGrant{
					{Address: goodAddr2, Permissions: []Permission{Permission_update}},
					{Address: goodAddr1, Permissions: []Permission{Permission_permissions}},
				},
			},
			expErr: []string{
				"address " + goodAddr2 + " appears in both the revoke-all and to-revoke fields",
				"address " + goodAddr1 + " appears in both the revoke-all and to-revoke fields",
			},
		},
		{
			name: "two invalid to-grant entries",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
				ToGrant: []AccessGrant{
					{Address: "badaddr", Permissions: []Permission{Permission_withdraw}},
					{Address: goodAddr1, Permissions: []Permission{Permission_unspecified}},
				},
			},
			expErr: []string{
				`invalid to-grant access grant: invalid address "badaddr"`,
				"invalid to-grant access grant: permission is unspecified for " + goodAddr1,
			},
		},
		{
			name: "revoke all for two addrs and some for one, then add some for all three",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     goodAdminAddr,
				MarketId:  1,
				RevokeAll: []string{goodAddr1, goodAddr2},
				ToRevoke:  []AccessGrant{{Address: goodAddr3, Permissions: []Permission{Permission_settle}}},
				ToGrant: []AccessGrant{
					{Address: goodAddr1, Permissions: []Permission{Permission_settle, Permission_update}},
					{Address: goodAddr2, Permissions: []Permission{Permission_cancel, Permission_permissions}},
					{Address: goodAddr3, Permissions: []Permission{Permission_withdraw, Permission_attributes}},
				},
			},
			expErr: nil,
		},
		{
			name: "revoke and grant the same permission for two addresses",
			msg: MsgMarketManagePermissionsRequest{
				Admin:    goodAdminAddr,
				MarketId: 1,
				ToRevoke: []AccessGrant{
					{Address: goodAddr2, Permissions: []Permission{Permission_permissions, Permission_cancel, Permission_attributes}},
					{Address: goodAddr1, Permissions: []Permission{Permission_settle, Permission_withdraw, Permission_update}},
				},
				ToGrant: []AccessGrant{
					{Address: goodAddr1, Permissions: []Permission{Permission_withdraw}},
					{Address: goodAddr2, Permissions: []Permission{Permission_attributes, Permission_permissions}},
				},
			},
			expErr: []string{
				"address " + goodAddr1 + " has both revoke and grant \"withdraw\"",
				"address " + goodAddr2 + " has both revoke and grant \"attributes\"",
				"address " + goodAddr2 + " has both revoke and grant \"permissions\"",
			},
		},
		{
			name: "multiple errs",
			msg: MsgMarketManagePermissionsRequest{
				Admin:     "",
				MarketId:  0,
				RevokeAll: []string{"bad-revoke-addr"},
				ToRevoke:  []AccessGrant{{Address: goodAddr1, Permissions: []Permission{Permission_unspecified}}},
				ToGrant:   []AccessGrant{{Address: "bad-grant-addr", Permissions: []Permission{Permission_withdraw}}},
			},
			expErr: []string{
				"invalid administrator \"\"",
				"invalid market id",
				"invalid revoke-all address \"bad-revoke-addr\"",
				"invalid to-revoke access grant: permission is unspecified for " + goodAddr1,
				`invalid to-grant access grant: invalid address "bad-grant-addr"`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgMarketManageReqAttrsRequest_ValidateBasic(t *testing.T) {
	goodAdmin := sdk.AccAddress("goodAdmin___________").String()

	tests := []struct {
		name   string
		msg    MsgMarketManageReqAttrsRequest
		expErr []string
	}{
		{
			name: "no admin",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          "",
				MarketId:       1,
				CreateAskToAdd: []string{"abc"},
			},
			expErr: []string{"invalid administrator", "empty address string is not allowed"},
		},
		{
			name: "bad admin",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          "not1valid",
				MarketId:       1,
				CreateAskToAdd: []string{"abc"},
			},
			expErr: []string{"invalid administrator", "decoding bech32 failed"},
		},
		{
			name: "market id zero",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          goodAdmin,
				CreateAskToAdd: []string{"abc"},
			},
			expErr: []string{"invalid market id: cannot be zero"},
		},
		{
			name: "no updates",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:    goodAdmin,
				MarketId: 1,
			},
			expErr: []string{"no updates"},
		},
		{
			name: "invalid create ask to add entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          goodAdmin,
				MarketId:       1,
				CreateAskToAdd: []string{"in-valid-attr"},
			},
			expErr: []string{"invalid create-ask to add required attribute \"in-valid-attr\""},
		},
		{
			name: "invalid create ask to remove entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateAskToRemove: []string{"in-valid-attr"},
			},
		},
		{
			name: "invalid create bid to add entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:          goodAdmin,
				MarketId:       1,
				CreateBidToAdd: []string{"in-valid-attr"},
			},
			expErr: []string{"invalid create-bid to add required attribute \"in-valid-attr\""},
		},
		{
			name: "invalid create bid to remove entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateBidToRemove: []string{"in-valid-attr"},
			},
		},
		{
			name: "add and remove same create ask entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateAskToAdd:    []string{"abc", "def", "ghi"},
				CreateAskToRemove: []string{"jkl", "abc"},
			},
			expErr: []string{"cannot add and remove the same create-ask required attributes \"abc\""},
		},
		{
			name: "add and remove same create bid entry",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateBidToAdd:    []string{"abc", "def", "ghi"},
				CreateBidToRemove: []string{"jkl", "abc"},
			},
			expErr: []string{"cannot add and remove the same create-bid required attributes \"abc\""},
		},
		{
			name: "add to create-ask the same as remove from create-bid",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateAskToAdd:    []string{"abc", "def", "ghi"},
				CreateBidToRemove: []string{"jkl", "abc"},
			},
		},
		{
			name: "add to create-bid the same as remove from create-ask",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateBidToAdd:    []string{"abc", "def", "ghi"},
				CreateAskToRemove: []string{"jkl", "abc"},
			},
		},
		{
			name: "add one to and remove one from each",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             goodAdmin,
				MarketId:          1,
				CreateAskToAdd:    []string{"to-add.ask"},
				CreateAskToRemove: []string{"to-remove.ask"},
				CreateBidToAdd:    []string{"to-add.bid"},
				CreateBidToRemove: []string{"to-remove.bid"},
			},
		},
		{
			name: "no admin and no market id and no updates",
			msg:  MsgMarketManageReqAttrsRequest{},
			expErr: []string{
				"invalid administrator",
				"invalid market id",
				"no updates",
			},
		},
		{
			name: "multiple errors",
			msg: MsgMarketManageReqAttrsRequest{
				Admin:             "not1valid",
				MarketId:          0,
				CreateAskToAdd:    []string{"bad-ask-attr", "dup-ask"},
				CreateAskToRemove: []string{"dup-ask"},
				CreateBidToAdd:    []string{"bad-bid-attr", "dup-bid"},
				CreateBidToRemove: []string{"dup-bid"},
			},
			expErr: []string{
				"invalid administrator",
				"invalid market id",
				"invalid create-ask to add required attribute \"bad-ask-attr\"",
				"cannot add and remove the same create-ask required attributes \"dup-ask\"",
				"invalid create-bid to add required attribute \"bad-bid-attr\"",
				"cannot add and remove the same create-bid required attributes \"dup-bid\"",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgGovCreateMarketRequest_ValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("authority___________").String()

	validMarket := Market{
		MarketId: 1,
		MarketDetails: MarketDetails{
			Name:        "Test Market One",
			Description: "This is the first test market",
		},
		FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("nhash", 10)},
		FeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("nhash", 20)},
		FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("nhash", 50)},
		FeeSellerSettlementRatios: []FeeRatio{
			{Price: sdk.NewInt64Coin("nhash", 100), Fee: sdk.NewInt64Coin("nhash", 1)},
		},
		FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("nhash", 100)},
		FeeBuyerSettlementRatios: []FeeRatio{
			{Price: sdk.NewInt64Coin("nhash", 100), Fee: sdk.NewInt64Coin("nhash", 1)},
		},
		AcceptingOrders:     true,
		AllowUserSettlement: true,
		AccessGrants: []AccessGrant{
			{
				Address:     sdk.AccAddress("just_some_addr______").String(),
				Permissions: AllPermissions(),
			},
		},
		ReqAttrCreateAsk: []string{"one.attr.pb"},
		ReqAttrCreateBid: []string{"*.attr.pb"},
	}

	tests := []struct {
		name   string
		msg    MsgGovCreateMarketRequest
		expErr []string
	}{
		{
			name:   "zero value",
			msg:    MsgGovCreateMarketRequest{},
			expErr: []string{"invalid authority"},
		},
		{
			name: "control",
			msg: MsgGovCreateMarketRequest{
				Authority: authority,
				Market:    validMarket,
			},
			expErr: nil,
		},
		{
			name: "no authority",
			msg: MsgGovCreateMarketRequest{
				Authority: "",
				Market:    validMarket,
			},
			expErr: []string{"invalid authority", "empty address string is not allowed"},
		},
		{
			name: "bad authority",
			msg: MsgGovCreateMarketRequest{
				Authority: "bad",
				Market:    validMarket,
			},
			expErr: []string{"invalid authority", "decoding bech32 failed"},
		},
		{
			name: "invalid market",
			msg: MsgGovCreateMarketRequest{
				Authority: authority,
				Market: Market{
					FeeCreateAskFlat: []sdk.Coin{{Denom: "badbad", Amount: sdkmath.NewInt(0)}},
				},
			},
			expErr: []string{`invalid create-ask flat fee option "0badbad": amount cannot be zero`},
		},
		{
			name: "multiple errors",
			msg: MsgGovCreateMarketRequest{
				Authority: "",
				Market: Market{
					FeeCreateBidFlat: []sdk.Coin{{Denom: "badbad", Amount: sdkmath.NewInt(0)}},
				},
			},
			expErr: []string{
				"invalid authority",
				`invalid create-bid flat fee option "0badbad": amount cannot be zero`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgGovManageFeesRequest_ValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("authority___________").String()
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	ratio := func(priceAmount int64, priceDenom string, feeAmount int64, feeDenom string) FeeRatio {
		return FeeRatio{Price: coin(priceAmount, priceDenom), Fee: coin(feeAmount, feeDenom)}
	}

	tests := []struct {
		name   string
		msg    MsgGovManageFeesRequest
		expErr []string
	}{
		{
			name:   "zero value",
			msg:    MsgGovManageFeesRequest{},
			expErr: []string{"invalid authority", "no updates"},
		},
		{
			name: "no authority",
			msg: MsgGovManageFeesRequest{
				Authority:           "",
				AddFeeCreateAskFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"invalid authority", "empty address string is not allowed"},
		},
		{
			name: "bad authority",
			msg: MsgGovManageFeesRequest{
				Authority:           "bad",
				AddFeeCreateAskFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"invalid authority", "decoding bech32 failed"},
		},
		{
			name: "invalid add create-ask flat",
			msg: MsgGovManageFeesRequest{
				Authority:           authority,
				AddFeeCreateAskFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid create-ask flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove create-ask flat",
			msg: MsgGovManageFeesRequest{
				Authority:              authority,
				AddFeeCreateAskFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeCreateAskFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same create-ask flat fee options 1nhash"},
		},
		{
			name: "invalid add create-bid flat",
			msg: MsgGovManageFeesRequest{
				Authority:           authority,
				AddFeeCreateBidFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid create-bid flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove create-bid flat",
			msg: MsgGovManageFeesRequest{
				Authority:              authority,
				AddFeeCreateBidFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeCreateBidFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same create-bid flat fee options 1nhash"},
		},
		{
			name: "invalid add seller settlement flat",
			msg: MsgGovManageFeesRequest{
				Authority:                  authority,
				AddFeeSellerSettlementFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid seller settlement flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove seller settlement flat",
			msg: MsgGovManageFeesRequest{
				Authority:                     authority,
				AddFeeSellerSettlementFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeSellerSettlementFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same seller settlement flat fee options 1nhash"},
		},
		{
			name: "invalid add seller settlement ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                    authority,
				AddFeeSellerSettlementRatios: []FeeRatio{ratio(1, "nhash", 2, "nhash")},
			},
			expErr: []string{`seller fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`},
		},
		{
			name: "same add and remove seller settlement ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                       authority,
				AddFeeSellerSettlementRatios:    []FeeRatio{ratio(2, "nhash", 1, "nhash")},
				RemoveFeeSellerSettlementRatios: []FeeRatio{ratio(2, "nhash", 1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same seller settlement fee ratios 2nhash:1nhash"},
		},
		{
			name: "invalid add buyer settlement flat",
			msg: MsgGovManageFeesRequest{
				Authority:                 authority,
				AddFeeBuyerSettlementFlat: []sdk.Coin{coin(0, "nhash")},
			},
			expErr: []string{`invalid buyer settlement flat fee to add option "0nhash": amount cannot be zero`},
		},
		{
			name: "same add and remove buyer settlement flat",
			msg: MsgGovManageFeesRequest{
				Authority:                    authority,
				AddFeeBuyerSettlementFlat:    []sdk.Coin{coin(1, "nhash")},
				RemoveFeeBuyerSettlementFlat: []sdk.Coin{coin(1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same buyer settlement flat fee options 1nhash"},
		},
		{
			name: "invalid add buyer settlement ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                   authority,
				AddFeeBuyerSettlementRatios: []FeeRatio{ratio(1, "nhash", 2, "nhash")},
			},
			expErr: []string{`buyer fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`},
		},
		{
			name: "same add and remove buyer settlement ratio",
			msg: MsgGovManageFeesRequest{
				Authority:                      authority,
				AddFeeBuyerSettlementRatios:    []FeeRatio{ratio(2, "nhash", 1, "nhash")},
				RemoveFeeBuyerSettlementRatios: []FeeRatio{ratio(2, "nhash", 1, "nhash")},
			},
			expErr: []string{"cannot add and remove the same buyer settlement fee ratios 2nhash:1nhash"},
		},
		{
			name: "multiple errors",
			msg: MsgGovManageFeesRequest{
				Authority:                       "",
				AddFeeCreateAskFlat:             []sdk.Coin{coin(0, "nhash")},
				RemoveFeeCreateAskFlat:          []sdk.Coin{coin(0, "nhash")},
				AddFeeCreateBidFlat:             []sdk.Coin{coin(0, "nhash")},
				RemoveFeeCreateBidFlat:          []sdk.Coin{coin(0, "nhash")},
				AddFeeSellerSettlementFlat:      []sdk.Coin{coin(0, "nhash")},
				RemoveFeeSellerSettlementFlat:   []sdk.Coin{coin(0, "nhash")},
				AddFeeSellerSettlementRatios:    []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				RemoveFeeSellerSettlementRatios: []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				AddFeeBuyerSettlementFlat:       []sdk.Coin{coin(0, "nhash")},
				RemoveFeeBuyerSettlementFlat:    []sdk.Coin{coin(0, "nhash")},
				AddFeeBuyerSettlementRatios:     []FeeRatio{ratio(1, "nhash", 2, "nhash")},
				RemoveFeeBuyerSettlementRatios:  []FeeRatio{ratio(1, "nhash", 2, "nhash")},
			},
			expErr: []string{
				"invalid authority", "empty address string is not allowed",
				`invalid create-ask flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same create-ask flat fee options 0nhash",
				`invalid create-bid flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same create-bid flat fee options 0nhash",
				`invalid seller settlement flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same seller settlement flat fee options 0nhash",
				`seller fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`,
				"cannot add and remove the same seller settlement fee ratios 1nhash:2nhash",
				`invalid buyer settlement flat fee to add option "0nhash": amount cannot be zero`,
				"cannot add and remove the same buyer settlement flat fee options 0nhash",
				`buyer fee ratio fee amount "2nhash" cannot be greater than price amount "1nhash"`,
				"cannot add and remove the same buyer settlement fee ratios 1nhash:2nhash",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}

func TestMsgGovManageFeesRequest_HasUpdates(t *testing.T) {
	oneCoin := []sdk.Coin{{}}
	oneRatio := []FeeRatio{{}}

	tests := []struct {
		name string
		msg  MsgGovManageFeesRequest
		exp  bool
	}{
		{
			name: "empty",
			msg:  MsgGovManageFeesRequest{},
			exp:  false,
		},
		{
			name: "empty except for authority",
			msg: MsgGovManageFeesRequest{
				Authority: "authority",
			},
			exp: false,
		},
		{
			name: "one add fee create-ask flat",
			msg:  MsgGovManageFeesRequest{AddFeeCreateAskFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee create-ask flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeCreateAskFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee create-bid flat",
			msg:  MsgGovManageFeesRequest{AddFeeCreateBidFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee create-bid flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeCreateBidFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee seller settlement flat",
			msg:  MsgGovManageFeesRequest{AddFeeSellerSettlementFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee seller settlement flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeSellerSettlementFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee seller settlement ratio",
			msg:  MsgGovManageFeesRequest{AddFeeSellerSettlementRatios: oneRatio},
			exp:  true,
		},
		{
			name: "one remove fee seller settlement ratio",
			msg:  MsgGovManageFeesRequest{RemoveFeeSellerSettlementRatios: oneRatio},
			exp:  true,
		},
		{
			name: "one add fee buyer settlement flat",
			msg:  MsgGovManageFeesRequest{AddFeeBuyerSettlementFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one remove fee buyer settlement flat",
			msg:  MsgGovManageFeesRequest{RemoveFeeBuyerSettlementFlat: oneCoin},
			exp:  true,
		},
		{
			name: "one add fee buyer settlement ratio",
			msg:  MsgGovManageFeesRequest{AddFeeBuyerSettlementRatios: oneRatio},
			exp:  true,
		},
		{
			name: "one remove fee buyer settlement ratio",
			msg:  MsgGovManageFeesRequest{RemoveFeeBuyerSettlementRatios: oneRatio},
			exp:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.msg.HasUpdates()
			}
			require.NotPanics(t, testFunc, "%T.HasUpdates()", tc.msg)
			assert.Equal(t, tc.exp, actual, "%T.HasUpdates() result", tc.msg)
		})
	}
}

func TestMsgGovUpdateParamsRequest_ValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("authority___________").String()

	tests := []struct {
		name   string
		msg    MsgGovUpdateParamsRequest
		expErr []string
	}{
		{
			name:   "zero value",
			msg:    MsgGovUpdateParamsRequest{},
			expErr: []string{"invalid authority", "empty address string is not allowed"},
		},
		{
			name: "default params",
			msg: MsgGovUpdateParamsRequest{
				Authority: authority,
				Params:    *DefaultParams(),
			},
			expErr: nil,
		},
		{
			name: "control",
			msg: MsgGovUpdateParamsRequest{
				Authority: authority,
				Params: Params{
					DefaultSplit: 543,
					DenomSplits: []DenomSplit{
						{Denom: "nhash", Split: 222},
						{Denom: "nusdf", Split: 123},
						{Denom: "musdm", Split: 8},
					},
				},
			},
			expErr: nil,
		},
		{
			name: "no authority",
			msg: MsgGovUpdateParamsRequest{
				Authority: "",
				Params:    *DefaultParams(),
			},
			expErr: []string{"invalid authority", "empty address string is not allowed"},
		},
		{
			name: "bad authority",
			msg: MsgGovUpdateParamsRequest{
				Authority: "bad",
				Params:    *DefaultParams(),
			},
			expErr: []string{"invalid authority", "decoding bech32 failed"},
		},
		{
			name: "bad params",
			msg: MsgGovUpdateParamsRequest{
				Authority: authority,
				Params: Params{
					DefaultSplit: 10_123,
					DenomSplits: []DenomSplit{
						{Denom: "x", Split: 500},
						{Denom: "nhash", Split: 20_000},
					},
				},
			},
			expErr: []string{
				"default split 10123 cannot be greater than 10000",
				"nhash split 20000 cannot be greater than 10000",
			},
		},
		{
			name: "multiple errors",
			msg: MsgGovUpdateParamsRequest{
				Authority: "",
				Params:    Params{DefaultSplit: 10_555},
			},
			expErr: []string{
				"invalid authority",
				"default split 10555 cannot be greater than 10000",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testValidateBasic(t, &tc.msg, tc.expErr)
		})
	}
}
