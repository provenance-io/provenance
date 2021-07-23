package types

import (
	"fmt"
	"strings"
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DenomTestSuite struct {
	suite.Suite
}

func (s *DenomTestSuite) SetupTest() {}

func TestDenomTestSuite(t *testing.T) {
	suite.Run(t, new(DenomTestSuite))
}

type denomMetadataTestCase struct {
	name      string
	md        banktypes.Metadata
	wantInErr []string
}

func getValidateDenomMetadataTestCases() []denomMetadataTestCase {
	return []denomMetadataTestCase{
		{
			"base is not a valid coin denomination",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits:  nil,
				Base:        "x",
				Display:     "hash",
				Name:        "Hash",
				Symbol:      "HASH",
			},
			[]string{"denom metadata"},
		},
		{
			"display is not a valid coin denomination",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits:  nil,
				Base:        "hash",
				Display:     "x",
				Name:        "Hash",
				Symbol:      "HASH",
			},
			[]string{"denom metadata"},
		},
		{
			"first denom unit is not exponent 0",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "hash", Exponent: 1, Aliases: nil},
				},
				Base:    "hash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"denom metadata"},
		},
		{
			"first denom unit is not base",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "hash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"denom metadata"},
		},
		{
			"denom units not ordered",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"denom metadata"},
		},
		{
			"description too long",
			banktypes.Metadata{
				Description: strings.Repeat("d", maxDenomMetadataDescriptionLength+1),
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"description", fmt.Sprint(maxDenomMetadataDescriptionLength), fmt.Sprint(maxDenomMetadataDescriptionLength + 1)},
		},
		{
			"no root coin name",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "hashx", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"root coin name"},
		},
		{
			"base prefix not SI",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "xhash", Exponent: 0, Aliases: nil},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "xhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"root coin name", "is not a SI prefix"},
		},
		{
			"alias duplicates other name",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: []string{"uhash"}},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"denom or alias", "is not unique", "uhash"},
		},
		{
			"denom duplicates other name",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
					{Denom: "nanohash", Exponent: 12, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"denom or alias", "is not unique", "nanohash"},
		},
		{
			"denom unit denom is not valid a coin denomination",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "x", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"denom metadata"},
		},
		{
			"denom unit denom exponent is incorrect",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 8, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"exponent", "hash", "0", "-9", "= 9", "8"},
		},
		{
			"denom unit alias is not valid a coin denomination",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: []string{strings.Repeat("x", 128) + "hash"}},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"invalid alias", "x"},
		},
		{
			"denom unit denom alias prefix mismatch",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
					{Denom: "megahash", Exponent: 15, Aliases: []string{"mhash"}},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{"SI prefix", "mhash", "megahash"},
		},
		{
			"should successfully validate metadata",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
					{Denom: "megahash", Exponent: 15, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			[]string{},
		},
		{
			"base denom is not valid has a slash coin denomination",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits:  nil,
				Base:        "my/hash",
				Display:     "hash",
				Name:        "Hash",
				Symbol:      "HASH",
			},
			[]string{"denom metadata"},
		},
	}
}

func (s *DenomTestSuite) TestValidateDenomMetadataBasic() {
	tests := getValidateDenomMetadataTestCases()

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			err := ValidateDenomMetadataBasic(tc.md)
			if len(tc.wantInErr) > 0 {
				require.Error(t, err, "ValidateDenomMetadataBasic expected error")
				for _, e := range tc.wantInErr {
					assert.Contains(t, err.Error(), e, "ValidateDenomMetadataBasic expected in error message")
				}
			} else {
				require.NoError(t, err, "ValidateDenomMetadataBasic unexpected error")
			}
		})
	}
}

func (s *DenomTestSuite) TestGetRootCoinName() {
	tests := []struct {
		name     string
		md       banktypes.Metadata
		expected string
	}{
		{
			"empty metadata",
			banktypes.Metadata{},
			"",
		},
		{
			"only one name",
			banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:   "onename",
						Aliases: nil,
					},
				},
			},
			"",
		},
		{
			"no common root",
			banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:   "onename",
						Aliases: []string{"another"},
					},
				},
			},
			"",
		},
		{
			"simple test",
			banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:   "onename",
						Aliases: []string{"twoname"},
					},
				},
			},
			"name",
		},
		{
			"real-use test",
			banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:   "nanohash",
						Aliases: []string{"nhash"},
					},
					{
						Denom:   "hash",
						Aliases: nil,
					},
					{
						Denom:   "kilohash",
						Aliases: []string{"khash"},
					},
				},
			},
			"hash",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := GetRootCoinName(tc.md)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
