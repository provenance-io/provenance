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
			name: "base is not a valid coin denomination",
			md: banktypes.Metadata{
				Description: "a description",
				DenomUnits:  nil,
				Base:        "x",
				Display:     "hash",
				Name:        "Hash",
				Symbol:      "HASH",
			},
			wantInErr: []string{"denom metadata"},
		},
		{
			name: "display is not a valid coin denomination",
			md: banktypes.Metadata{
				Description: "a description",
				DenomUnits:  nil,
				Base:        "hash",
				Display:     "x",
				Name:        "Hash",
				Symbol:      "HASH",
			},
			wantInErr: []string{"denom metadata"},
		},
		{
			name: "first denom unit is not exponent 0",
			md: banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "hash", Exponent: 1, Aliases: nil},
				},
				Base:    "hash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			wantInErr: []string{"denom metadata"},
		},
		{
			name: "first denom unit is not base",
			md: banktypes.Metadata{
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
			wantInErr: []string{"denom metadata"},
		},
		{
			name: "denom units not ordered",
			md: banktypes.Metadata{
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
			wantInErr: []string{"denom metadata"},
		},
		{
			name: "description too long",
			md: banktypes.Metadata{
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
			wantInErr: []string{"description", fmt.Sprint(maxDenomMetadataDescriptionLength), fmt.Sprint(maxDenomMetadataDescriptionLength + 1)},
		},
		{
			name: "no root coin name",
			md: banktypes.Metadata{
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
			wantInErr: []string{"root coin name"},
		},
		{
			name: "base prefix not SI",
			md: banktypes.Metadata{
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
			wantInErr: []string{"root coin name", "is not a SI prefix"},
		},
		{
			name: "alias duplicates other name",
			md: banktypes.Metadata{
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
			wantInErr: []string{"denom or alias", "is not unique", "uhash"},
		},
		{
			name: "denom duplicates other name",
			md: banktypes.Metadata{
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
			wantInErr: []string{"denom or alias", "is not unique", "nanohash"},
		},
		{
			name: "denom unit denom is not valid a coin denomination",
			md: banktypes.Metadata{
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
			wantInErr: []string{"denom metadata"},
		},
		{
			name: "denom unit denom exponent is incorrect",
			md: banktypes.Metadata{
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
			wantInErr: []string{"exponent", "hash", "0", "-9", "= 9", "8"},
		},
		{
			name: "denom unit alias is not valid a coin denomination",
			md: banktypes.Metadata{
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
			wantInErr: []string{"invalid alias", "x"},
		},
		{
			name: "denom unit denom alias prefix mismatch",
			md: banktypes.Metadata{
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
			wantInErr: []string{"SI prefix", "mhash", "megahash"},
		},
		{
			name: "should successfully validate metadata",
			md: banktypes.Metadata{
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
			wantInErr: []string{},
		},
		{
			name: "base denom is not valid has a slash coin denomination",
			md: banktypes.Metadata{
				Description: "a description",
				DenomUnits:  nil,
				Base:        "my/hash",
				Display:     "hash",
				Name:        "Hash",
				Symbol:      "HASH",
			},
			wantInErr: []string{"denom metadata"},
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
			name: "empty metadata",
			md: banktypes.Metadata{},
			expected: "",
		},
		{
			name: "only one name",
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:   "onename",
						Aliases: nil,
					},
				},
			},
			expected: "",
		},
		{
			name: "no common root",
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:   "onename",
						Aliases: []string{"another"},
					},
				},
			},
			expected: "",
		},
		{
			name: "simple test",
			md: banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:   "onename",
						Aliases: []string{"twoname"},
					},
				},
			},
			expected: "name",
		},
		{
			name: "real-use test",
			md: banktypes.Metadata{
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
			expected: "hash",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := GetRootCoinName(tc.md)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
