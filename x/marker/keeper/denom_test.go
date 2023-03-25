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
	s.app = simapp.Setup(s.T())
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
			name: "fails basic validation",
			proposed: banktypes.Metadata{
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
			existing: nil,
			markerStatus: types.StatusUndefined,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"invalid proposed metadata", "invalid metadata base denom"},
		},
		{
			name: "marker status undefined",
			proposed: banktypes.Metadata{
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
			existing: nil,
			markerStatus: types.StatusUndefined,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"cannot add or update denom metadata", "undefined"},
		},
		{
			name: "marker status destroyed",
			proposed: banktypes.Metadata{
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
			existing: nil,
			markerStatus: types.StatusDestroyed,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"cannot add or update denom metadata", "destroyed"},
		},
		{
			name: "marker status cancelled",
			proposed: banktypes.Metadata{
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
			existing: nil,
			markerStatus: types.StatusCancelled,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"cannot add or update denom metadata", "cancelled"},
		},
		{
			name: "denom fails extra regex",
			proposed: banktypes.Metadata{
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
			existing: nil,
			markerStatus: types.StatusProposed,
			denomValidationExpression: `[nu]hash`,
			wantInErr: []string{"fails unrestricted marker denom validation", "hash"},
		},
		{
			name: "alias fails extra regex",
			proposed: banktypes.Metadata{
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
			existing: nil,
			markerStatus: types.StatusProposed,
			denomValidationExpression: `[nu]?hash`,
			wantInErr: []string{"fails unrestricted marker denom validation", "nanohash"},
		},
		{
			name: "base changed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "uhash",
				Display: "hash",
			},
			markerStatus: types.StatusProposed,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"denom metadata base value cannot be changed"},
		},
		{
			name: "active denom unit removed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			markerStatus: types.StatusActive,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"cannot remove denom unit", "uhash"},
		},
		{
			name: "finalized denom unit removed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			markerStatus: types.StatusFinalized,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"cannot remove denom unit", "uhash"},
		},
		{
			name: "proposed denom unit removed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			markerStatus: types.StatusProposed,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{},
		},
		{
			name: "active denom unit denom changed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			markerStatus: types.StatusActive,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"denom unit Denom", "uhash", "microhash"},
		},
		{
			name: "finalized denom unit denom changed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			markerStatus: types.StatusFinalized,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"denom unit Denom", "uhash", "microhash"},
		},
		{
			name: "proposed denom unit denom changed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			markerStatus: types.StatusProposed,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{},
		},
		{
			name: "active denom unit alias removed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			markerStatus: types.StatusActive,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"cannot remove alias", "nanohash", "nhash"},
		},
		{
			name: "finalized denom unit alias removed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
				Description: "a description",
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: "nhash", Exponent: 0, Aliases: []string{"nanohash"}},
					{Denom: "uhash", Exponent: 3, Aliases: nil},
					{Denom: "hash", Exponent: 9, Aliases: nil},
				},
				Base:    "nhash",
				Display: "hash",
			},
			markerStatus: types.StatusFinalized,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{"cannot remove alias", "nanohash", "nhash"},
		},
		{
			name: "proposed denom unit alias removed",
			proposed: banktypes.Metadata{
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
			existing: &banktypes.Metadata{
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
			markerStatus: types.StatusProposed,
			denomValidationExpression: types.DefaultUnrestrictedDenomRegex,
			wantInErr: []string{},
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
