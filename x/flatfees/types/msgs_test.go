package types_test

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"

	. "github.com/provenance-io/provenance/x/flatfees/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgUpdateParamsRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateConversionFactorRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateMsgFeesRequest{Authority: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestMsgUpdateParamsRequest_ValidateBasic(t *testing.T) {
	pioconfig.SetProvConfig("cherry")

	tests := []struct {
		name   string
		msg    MsgUpdateParamsRequest
		expErr string
	}{
		{
			name: "no authority",
			msg: MsgUpdateParamsRequest{
				Authority: "",
				Params:    DefaultParams(),
			},
			expErr: "invalid authority: empty address string is not allowed",
		},
		{
			name: "bad authority",
			msg: MsgUpdateParamsRequest{
				Authority: "bad",
				Params:    DefaultParams(),
			},
			expErr: "invalid authority: decoding bech32 failed: invalid bech32 string length 3",
		},
		{
			name: "invalid params",
			msg: MsgUpdateParamsRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				Params:    Params{},
			},
			expErr: "invalid params: invalid default cost \"<nil>\": invalid denom: ",
		},
		{
			name: "valid",
			msg: MsgUpdateParamsRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				Params: Params{
					DefaultCost: sdk.NewInt64Coin("plum", 100),
					ConversionFactor: ConversionFactor{
						BaseAmount:      sdk.NewInt64Coin("plum", 1),
						ConvertedAmount: sdk.NewInt64Coin("acorn", 17),
					},
				},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "MsgUpdateParamsRequest.ValidateBasic()")
			assertions.AssertErrorValue(t, err, tc.expErr, "MsgUpdateParamsRequest.ValidateBasic() error")
		})
	}
}

func TestMsgUpdateConversionFactorRequest_ValidateBasic(t *testing.T) {
	authority := sdk.AccAddress("authority___________").String() // Not the actual authority.
	coin := func(str string) sdk.Coin {
		var amtStr, denom string
		for i, c := range str {
			if !unicode.IsDigit(c) && !(i == 0 && c == '-') {
				amtStr = str[:i]
				denom = str[i:]
				break
			}
		}
		require.NotEmpty(t, amtStr, "The amount extracted from %q", str)
		require.NotEmpty(t, denom, "The denom extracted from %q", str)
		amt, ok := sdkmath.NewIntFromString(amtStr)
		require.True(t, ok, "sdkmath.NewIntFromString(%q), the amount from %q", amtStr, str)
		return sdk.Coin{Denom: denom, Amount: amt}
	}
	cf := func(base, converted string) ConversionFactor {
		return ConversionFactor{BaseAmount: coin(base), ConvertedAmount: coin(converted)}
	}

	tests := []struct {
		name   string
		msg    MsgUpdateConversionFactorRequest
		expErr string
	}{
		{
			name: "all good",
			msg: MsgUpdateConversionFactorRequest{
				Authority:        authority,
				ConversionFactor: cf("4orange", "8apple"),
			},
			expErr: "",
		},
		{
			name: "empty authority",
			msg: MsgUpdateConversionFactorRequest{
				Authority:        "",
				ConversionFactor: cf("12banana", "3grape"),
			},
			expErr: "invalid authority: empty address string is not allowed",
		},
		{
			name: "invalid authority",
			msg: MsgUpdateConversionFactorRequest{
				Authority:        "nopenope",
				ConversionFactor: cf("12banana", "3grape"),
			},
			expErr: "invalid authority: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "invalid conversion factor",
			msg: MsgUpdateConversionFactorRequest{
				Authority:        authority,
				ConversionFactor: cf("-3apple", "4peach"),
			},
			expErr: "invalid conversion factor: invalid base amount \"-3apple\": negative coin amount: -3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "MsgUpdateConversionFactorRequest.ValidateBasic()")
			assertions.AssertErrorValue(t, err, tc.expErr, "MsgUpdateConversionFactorRequest.ValidateBasic() error")
		})
	}
}

func TestMsgUpdateMsgFeesRequest_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgUpdateMsgFeesRequest
		expErr string
	}{
		{
			name:   "no authority",
			msg:    MsgUpdateMsgFeesRequest{Authority: ""},
			expErr: "invalid authority: empty address string is not allowed",
		},
		{
			name:   "bad authority",
			msg:    MsgUpdateMsgFeesRequest{Authority: "bad"},
			expErr: "invalid authority: decoding bech32 failed: invalid bech32 string length 3",
		},
		{
			name:   "empty",
			msg:    MsgUpdateMsgFeesRequest{Authority: sdk.AccAddress("authority___________").String()},
			expErr: "at least one entry to set or unset must be provided: empty request",
		},
		{
			name: "invalid ToSet",
			msg: MsgUpdateMsgFeesRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				ToSet: []*MsgFee{
					NewMsgFee("/set.one"),
					NewMsgFee("/set.two", sdk.NewInt64Coin("banana", 10)),
					{MsgTypeUrl: "/set.three", Cost: sdk.Coins{{Denom: "x", Amount: sdkmath.NewInt(3)}}},
					NewMsgFee("/set.four", sdk.NewInt64Coin("banana", 7), sdk.NewInt64Coin("cherry", 15)),
				},
			},
			expErr: "invalid ToSet[2]: invalid /set.three cost \"3x\": invalid denom: x",
		},
		{
			name: "invalid ToUnset",
			msg: MsgUpdateMsgFeesRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				ToUnset:   []string{"/unset.one", "", "/unset.three", "/unset.four"},
			},
			expErr: "invalid ToUnset[1]: msg type url cannot be empty",
		},
		{
			name: "duplicate in ToSet",
			msg: MsgUpdateMsgFeesRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				ToSet: []*MsgFee{
					NewMsgFee("/set.one"),
					NewMsgFee("/set.two", sdk.NewInt64Coin("banana", 10)),
					NewMsgFee("/set.three", sdk.NewInt64Coin("banana", 44)),
					NewMsgFee("/set.two", sdk.NewInt64Coin("banana", 7), sdk.NewInt64Coin("cherry", 15)),
				},
			},
			expErr: "duplicate msg type url \"/set.two\" found in ToSet[1] and ToSet[3]",
		},
		{
			name: "duplicate in ToUnset",
			msg: MsgUpdateMsgFeesRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				ToUnset:   []string{"/unset.one", "/unset.two", "/unset.two", "/unset.four"},
			},
			expErr: "duplicate msg type url \"/unset.two\" found in ToUnset[1] and ToUnset[2]",
		},
		{
			name: "duplicate in both",
			msg: MsgUpdateMsgFeesRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				ToSet: []*MsgFee{
					NewMsgFee("/set.one"),
					NewMsgFee("/set.two", sdk.NewInt64Coin("banana", 10)),
					NewMsgFee("/both.set.unset.whatever", sdk.NewInt64Coin("banana", 1)),
					NewMsgFee("/set.three", sdk.NewInt64Coin("banana", 44)),
					NewMsgFee("/set.four", sdk.NewInt64Coin("banana", 7), sdk.NewInt64Coin("cherry", 15)),
				},
				ToUnset: []string{"/unset.one", "/unset.two", "/unset.three", "/unset.four", "/both.set.unset.whatever"},
			},
			expErr: "duplicate msg type url \"/both.set.unset.whatever\" found in ToSet[2] and ToUnset[4]",
		},
		{
			name: "valid: one ToSet",
			msg: MsgUpdateMsgFeesRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				ToSet:     []*MsgFee{NewMsgFee("/some.msg", sdk.NewInt64Coin("banana", 10))},
			},
		},
		{
			name: "valid: one ToUnset",
			msg: MsgUpdateMsgFeesRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				ToUnset:   []string{"/some.msg"},
			},
		},
		{
			name: "valid: multiple of both",
			msg: MsgUpdateMsgFeesRequest{
				Authority: sdk.AccAddress("authority___________").String(),
				ToSet: []*MsgFee{
					NewMsgFee("/set.one"),
					NewMsgFee("/set.two", sdk.NewInt64Coin("banana", 10)),
					NewMsgFee("/set.three", sdk.NewInt64Coin("banana", 44)),
					NewMsgFee("/set.four", sdk.NewInt64Coin("banana", 7), sdk.NewInt64Coin("cherry", 15)),
				},
				ToUnset: []string{"/unset.one", "/unset.two", "/unset.three", "/unset.four"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.msg.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "MsgUpdateMsgFeesRequest.ValidateBasic()")
			assertions.AssertErrorValue(t, err, tc.expErr, "MsgUpdateMsgFeesRequest.ValidateBasic() error")
		})
	}
}
