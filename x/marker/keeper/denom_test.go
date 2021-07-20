package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"

	"github.com/provenance-io/provenance/x/marker/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DenomTestSuite struct {
	app *simapp.App
	ctx sdk.Context

	suite.Suite
}

func (s *DenomTestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})

	s.app.MarkerKeeper.SetParams(s.ctx, types.DefaultParams())
}

func TestDenomTestSuite(t *testing.T) {
	suite.Run(t, new(DenomTestSuite))
}
func (s *DenomTestSuite) TestInvalidDenomExpression() {
	s.T().Run("invalid denom expression", func(t *testing.T) {
		assert.Panics(t,
			func() { s.app.MarkerKeeper.SetParams(s.ctx, types.Params{UnrestrictedDenomRegex: `(invalid`}) },
			"value from ParamSetPair is invalid: error parsing regexp: missing closing ): `^(invalid$`",
		)
	})
}

func (s *DenomTestSuite) TestValidateDenomMetadataExtended() {

	tests := []struct {
		name                      string
		proposed                  banktypes.Metadata
		existing                  *banktypes.Metadata
		markerStatus              types.MarkerStatus
		denomValidationExpression string
		wantInErr                 []string
	}{
		{
			"fails basic validation",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: nil},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "invalid", Exponent: 9, Aliases: nil},
				},
				Base:    "1234invalid",
				Display: "invalid",
				Name:    "invalid",
				Symbol:  "INV",
			},
			nil,
			types.StatusUndefined,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"invalid proposed metadata", "invalid metadata base denom"},
		},
		{
			"marker status undefined",
			banktypes.Metadata{
				Description: "a description",
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
			nil,
			types.StatusUndefined,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"cannot add or update denom metadata", "undefined"},
		},
		{
			"marker status destroyed",
			banktypes.Metadata{
				Description: "a description",
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
			nil,
			types.StatusDestroyed,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"cannot add or update denom metadata", "destroyed"},
		},
		{
			"marker status cancelled",
			banktypes.Metadata{
				Description: "a description",
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
			nil,
			types.StatusCancelled,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"cannot add or update denom metadata", "cancelled"},
		},
		{
			"denom fails extra regex",
			banktypes.Metadata{
				Description: "a description",
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
			nil,
			types.StatusProposed,
			`[nu]hash`,
			[]string{"fails unrestricted marker denom validation", "hash"},
		},
		{
			"alias fails extra regex",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			nil,
			types.StatusProposed,
			`[nu]?hash`,
			[]string{"fails unrestricted marker denom validation", "nanohash"},
		},
		{
			"base changed",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "uhash",
				Display: "hash",
			},
			types.StatusProposed,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"denom metadata base value cannot be changed"},
		},
		{
			"active denom unit removed",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			types.StatusActive,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"cannot remove denom unit", "uhash"},
		},
		{
			"finalized denom unit removed",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			types.StatusFinalized,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"cannot remove denom unit", "uhash"},
		},
		{
			"proposed denom unit removed",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			types.StatusProposed,
			types.DefaultUnrestrictedDenomRegex,
			[]string{},
		},
		{
			"active denom unit denom changed",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "microhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			types.StatusActive,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"denom unit Denom", "uhash", "microhash"},
		},
		{
			"finalized denom unit denom changed",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "microhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			types.StatusFinalized,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"denom unit Denom", "uhash", "microhash"},
		},
		{
			"proposed denom unit denom changed",
			banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "microhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			types.StatusProposed,
			types.DefaultUnrestrictedDenomRegex,
			[]string{},
		},
		{
			"active denom unit alias removed",
			banktypes.Metadata{
				Description: "a description",
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
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			types.StatusActive,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"cannot remove alias", "nanohash", "nhash"},
		},
		{
			"finalized denom unit alias removed",
			banktypes.Metadata{
				Description: "a description",
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
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			types.StatusFinalized,
			types.DefaultUnrestrictedDenomRegex,
			[]string{"cannot remove alias", "nanohash", "nhash"},
		},
		{
			"proposed denom unit alias removed",
			banktypes.Metadata{
				Description: "a description",
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
			&banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
				Name:    "Hash",
				Symbol:  "HASH",
			},
			types.StatusProposed,
			types.DefaultUnrestrictedDenomRegex,
			[]string{},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			s.app.MarkerKeeper.SetParams(s.ctx, types.Params{UnrestrictedDenomRegex: tc.denomValidationExpression})

			err := s.app.MarkerKeeper.ValidateDenomMetadata(s.ctx, tc.proposed, tc.existing, tc.markerStatus)
			if len(tc.wantInErr) > 0 {
				require.Error(t, err, "ValidateDenomMetadataExtended expected error")
				for _, e := range tc.wantInErr {
					assert.Contains(t, err.Error(), e, "ValidateDenomMetadataExtended expected in error message")
				}
			} else {
				require.NoError(t, err, "ValidateDenomMetadataExtended unexpected error")
			}
		})
	}
}
