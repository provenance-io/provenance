package keeper_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/flatfees/keeper"
	"github.com/provenance-io/provenance/x/flatfees/types"
)

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

type GenesisTestSuite struct {
	suite.Suite

	app *simapp.App
	ctx sdk.Context
	kpr keeper.Keeper
}

func (s *GenesisTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.kpr = s.app.FlatFeesKeeper
	s.ctx = s.app.BaseApp.NewContext(false)
}

func (s *GenesisTestSuite) TestInitExportGenesis() {
	tests := []struct {
		name         string
		data         *types.GenesisState
		expInitPanic string
	}{
		{
			name:         "invalid genesis state",
			data:         nil,
			expInitPanic: "flatfees genesis state cannot be nil",
		},
		{
			name: "default genesis state",
			data: types.DefaultGenesisState(),
		},
		{
			name: "non-defaults",
			data: &types.GenesisState{
				Params: types.Params{
					DefaultCost: sdk.NewInt64Coin("apple", 14),
					ConversionFactor: types.ConversionFactor{
						DefinitionAmount: sdk.NewInt64Coin("apple", 7),
						ConvertedAmount:  sdk.NewInt64Coin("orange", 3),
					},
				},
				MsgFees: []*types.MsgFee{
					types.NewMsgFee("free"),
					types.NewMsgFee("same.denom", sdk.NewInt64Coin("apple", 7)),
					types.NewMsgFee("other.denom", sdk.NewInt64Coin("banana", 99)),
					types.NewMsgFee("two.coins", sdk.NewInt64Coin("apple", 21), sdk.NewInt64Coin("cherry", 4)),
					types.NewMsgFee("also.free"),
				},
			},
		},
		// Not sure how to make .Walk, .SetParams, or .SetMsgFee return an error, so those panics aren't tested.
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			testInit := func() {
				s.kpr.InitGenesis(s.ctx, tc.data)
			}
			assertions.RequirePanicEquals(s.T(), testInit, tc.expInitPanic, "InitGenesis")

			if len(tc.expInitPanic) != 0 {
				return
			}

			slices.SortFunc(tc.data.MsgFees, func(a, b *types.MsgFee) int {
				if a == b {
					return 0
				}
				if a == nil {
					return 1
				}
				if b == nil {
					return -1
				}
				return strings.Compare(a.MsgTypeUrl, b.MsgTypeUrl)
			})

			var actual *types.GenesisState
			testExport := func() {
				actual = s.kpr.ExportGenesis(s.ctx)
			}
			s.Require().NotPanics(testExport, "ExportGenesis")
			s.Require().NotNil(actual, "exported GenesisState")
			ok := assertEqualParams(s.T(), tc.data.Params, actual.Params)
			ok = assertEqualMsgFees(s.T(), tc.data.MsgFees, actual.MsgFees) && ok
			if ok {
				s.Assert().Equal(tc.data, actual, "exported GenesisState")
			}
		})
	}
}
