package exchange

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxSplit(t *testing.T) {
	// The MaxSplit should never be changed.
	// But if it is changed for some reason, it can never be more than 100%.
	absoluteMax := uint32(10_000)
	assert.LessOrEqual(t, MaxSplit, absoluteMax)
}

func TestNewParams(t *testing.T) {
	tests := []struct {
		name         string
		defaultSplit uint32
		denomSplits  []DenomSplit
		expected     *Params
	}{
		{
			name:         "zero values",
			defaultSplit: 0,
			denomSplits:  nil,
			expected:     &Params{},
		},
		{
			name:         "100 nil",
			defaultSplit: 100,
			denomSplits:  nil,
			expected:     &Params{DefaultSplit: 100},
		},
		{
			name:         "123 with two denom splits",
			defaultSplit: 123,
			denomSplits: []DenomSplit{
				{Denom: "atom", Split: 234},
				{Denom: "nhash", Split: 56},
			},
			expected: &Params{
				DefaultSplit: 123,
				DenomSplits: []DenomSplit{
					{Denom: "atom", Split: 234},
					{Denom: "nhash", Split: 56},
				},
			},
		},
		{
			name:         "defaults",
			defaultSplit: DefaultDefaultSplit,
			denomSplits:  nil,
			expected:     DefaultParams(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewParams(tc.defaultSplit, tc.denomSplits)
			assert.Equal(t, tc.expected, actual, "NewParams")
		})
	}
}

func TestDefaultParams(t *testing.T) {
	actual := DefaultParams()
	assert.Equal(t, int(DefaultDefaultSplit), int(actual.DefaultSplit), "DefaultSplit")
	assert.Nil(t, actual.DenomSplits, "DenomSplits")
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
			name: "bad default and three bad denom splits",
			params: Params{
				DefaultSplit: 10_001,
				DenomSplits: []DenomSplit{
					{Denom: "badcointype", Split: 10_002},
					{Denom: "x", Split: 3},
					{Denom: "thisIsNotGood", Split: 10_003},
				},
			},
			expErr: []string{
				"default split 10001 cannot be greater than 10000",
				"badcointype split 10002 cannot be greater than 10000",
				"invalid denom: x",
				"thisIsNotGood split 10003 cannot be greater than 10000",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()

			// TODO[1658]: Replace this with testutils.AssertErrorContents(t, err, tc.expErr, "Validate")
			if len(tc.expErr) > 0 {
				if assert.Error(t, err, "Validate") {
					for _, exp := range tc.expErr {
						assert.ErrorContains(t, err, exp, "Validate\nExpecting: %q", exp)
					}
				}
			} else {
				assert.NoError(t, err, "Validate")
			}
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
			// TODO[1658]: Replace this with assertErrorValue(t, err, tc.expErr, "Validate")
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "Validate")
			} else {
				assert.NoError(t, err, "Validate")
			}
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
