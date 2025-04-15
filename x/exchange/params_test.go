package exchange

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestMaxSplit(t *testing.T) {
	// The MaxSplit should never be changed.
	// But if it is changed for some reason, it can never be more than 100%.
	absoluteMax := uint32(10_000)
	assert.LessOrEqual(t, MaxSplit, absoluteMax)
}

func TestDefaultParams(t *testing.T) {
	feeDenom := pioconfig.GetProvConfig().FeeDenom
	expCreate := fmt.Sprintf("%d%s", DefaultFeeCreatePaymentFlatAmount, feeDenom)
	expAccept := fmt.Sprintf("%d%s", DefaultFeeAcceptPaymentFlatAmount, feeDenom)

	actual := DefaultParams()
	assert.Equal(t, int(DefaultDefaultSplit), int(actual.DefaultSplit), "DefaultSplit")
	assert.Nil(t, actual.DenomSplits, "DenomSplits")
	if assert.Len(t, actual.FeeCreatePaymentFlat, 1, "FeeCreatePaymentFlat") {
		assert.Equal(t, expCreate, actual.FeeCreatePaymentFlat[0].String(), "FeeCreatePaymentFlat[0]")
	}
	if assert.Len(t, actual.FeeAcceptPaymentFlat, 1, "FeeAcceptPaymentFlat") {
		assert.Equal(t, expAccept, actual.FeeAcceptPaymentFlat[0].String(), "FeeAcceptPaymentFlat[0]")
	}
}

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name   string
		params Params
		expErr []string
	}{
		{
			name:   "zero values",
			params: Params{},
			expErr: nil,
		},
		{
			name:   "default values",
			params: *DefaultParams(),
			expErr: nil,
		},
		{
			name:   "bad default split",
			params: Params{DefaultSplit: 10_001},
			expErr: []string{"default split 10001 cannot be greater than 10000"},
		},
		{
			name:   "one denom split and it is bad",
			params: Params{DenomSplits: []DenomSplit{{Denom: "badcointype", Split: 10_001}}},
			expErr: []string{"badcointype split 10001 cannot be greater than 10000"},
		},
		{
			name:   "empty create payment flat fees",
			params: Params{FeeCreatePaymentFlat: []sdk.Coin{}},
			expErr: nil,
		},
		{
			name:   "create payment flat fee with zero amount",
			params: Params{FeeCreatePaymentFlat: []sdk.Coin{{Denom: "blueberry", Amount: sdkmath.ZeroInt()}}},
			expErr: []string{"invalid create payment flat fee \"0blueberry\": zero amount not allowed"},
		},
		{
			name:   "too many create payment flat fees",
			params: Params{FeeCreatePaymentFlat: []sdk.Coin{sdk.NewInt64Coin("orange", 3), sdk.NewInt64Coin("cherry", 5)}},
			expErr: []string{"invalid create payment flat fee \"3orange,5cherry\": max entries is 1"},
		},
		{
			name:   "invalid create payment flat fees",
			params: Params{FeeCreatePaymentFlat: []sdk.Coin{{Denom: "banana", Amount: sdkmath.NewInt(-1)}}},
			expErr: []string{"invalid create payment flat fee \"-1banana\": negative coin amount: -1"},
		},
		{
			name:   "empty accept payment flat fees",
			params: Params{FeeAcceptPaymentFlat: []sdk.Coin{}},
			expErr: nil,
		},
		{
			name:   "accept payment flat fee with zero amount",
			params: Params{FeeAcceptPaymentFlat: []sdk.Coin{{Denom: "blueberry", Amount: sdkmath.ZeroInt()}}},
			expErr: []string{"invalid accept payment flat fee \"0blueberry\": zero amount not allowed"},
		},
		{
			name:   "too many accept payment flat fees",
			params: Params{FeeAcceptPaymentFlat: []sdk.Coin{sdk.NewInt64Coin("orange", 3), sdk.NewInt64Coin("cherry", 5)}},
			expErr: []string{"invalid accept payment flat fee \"3orange,5cherry\": max entries is 1"},
		},
		{
			name:   "invalid accept payment flat fees",
			params: Params{FeeAcceptPaymentFlat: []sdk.Coin{{Denom: "banana", Amount: sdkmath.NewInt(-1)}}},
			expErr: []string{"invalid accept payment flat fee \"-1banana\": negative coin amount: -1"},
		},
		{
			name: "multiple errors",
			params: Params{
				DefaultSplit: 10_001,
				DenomSplits: []DenomSplit{
					{Denom: "badcointype", Split: 10_002},
					{Denom: "x", Split: 3},
					{Denom: "thisIsNotGood", Split: 10_003},
				},
				FeeCreatePaymentFlat: []sdk.Coin{sdk.NewInt64Coin("orange", 3), sdk.NewInt64Coin("cherry", 5)},
				FeeAcceptPaymentFlat: []sdk.Coin{{Denom: "banana", Amount: sdkmath.NewInt(-1)}},
			},
			expErr: []string{
				"default split 10001 cannot be greater than 10000",
				"badcointype split 10002 cannot be greater than 10000",
				"invalid denom: x",
				"thisIsNotGood split 10003 cannot be greater than 10000",
				"invalid create payment flat fee \"3orange,5cherry\": max entries is 1",
				"invalid accept payment flat fee \"-1banana\": negative coin amount: -1",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()

			assertions.AssertErrorContents(t, err, tc.expErr, "Validate")
		})
	}
}

func TestNewDenomSplit(t *testing.T) {
	tests := []struct {
		name  string
		denom string
		split uint32
		exp   *DenomSplit
	}{
		{
			name:  "zero values",
			denom: "",
			split: 0,
			exp:   &DenomSplit{},
		},
		{
			name:  "denom without split",
			denom: "somedenom",
			split: 0,
			exp:   &DenomSplit{Denom: "somedenom"},
		},
		{
			name:  "split without denom",
			denom: "",
			split: 541,
			exp:   &DenomSplit{Split: 541},
		},
		{
			name:  "both provided",
			denom: "anotherdenom",
			split: 565,
			exp:   &DenomSplit{Denom: "anotherdenom", Split: 565},
		},
		{
			name:  "neither value valid",
			denom: "z",
			split: 10_008,
			exp:   &DenomSplit{Denom: "z", Split: 10_008},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewDenomSplit(tc.denom, tc.split)
			assert.Equal(t, tc.exp, actual, "NewDenomSplit")
		})
	}
}

func TestDenomSplit_Validate(t *testing.T) {
	tests := []struct {
		name       string
		denomSplit DenomSplit
		expErr     string
	}{
		{
			name:       "no denom",
			denomSplit: DenomSplit{Denom: "", Split: 5},
			expErr:     "invalid denom: ",
		},
		{
			name:       "invalid denom",
			denomSplit: DenomSplit{Denom: "f", Split: 6},
			expErr:     "invalid denom: f",
		},
		{
			name:       "invalid split",
			denomSplit: DenomSplit{Denom: "thisDenomIsOkay", Split: 10_123},
			expErr:     "thisDenomIsOkay split 10123 cannot be greater than 10000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.denomSplit.Validate()
			assertions.AssertErrorValue(t, err, tc.expErr, "Validate")
		})
	}
}

func TestValidateSplit(t *testing.T) {
	name := "<name>"
	split := MaxSplit + 1
	splitStr := fmt.Sprintf("%d", split)
	maxSplitStr := fmt.Sprintf("%d", MaxSplit)

	err := validateSplit(name, split)
	if assert.Error(t, err, "validateSplit") {
		assert.ErrorContains(t, err, name, "provided name")
		assert.ErrorContains(t, err, splitStr, "provided split value")
		assert.ErrorContains(t, err, maxSplitStr, "maximum split value")
	}
}
