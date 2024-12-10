package keeper_test

import (
	"github.com/provenance-io/provenance/x/nav"
)

func (s *TestSuite) TestKeeper_InitGenesis() {
	tests := []struct {
		name     string
		genState *nav.GenesisState
		expNAVs  nav.NAVRecords
		expPanic string
	}{
		{
			name:     "nil gen state",
			genState: nil,
			expNAVs:  nil,
		},
		{
			name:     "nil navs",
			genState: &nav.GenesisState{Navs: nil},
			expNAVs:  nil,
		},
		{
			name:     "empty navs",
			genState: &nav.GenesisState{Navs: nav.NAVRecords{}},
			expNAVs:  nil,
		},
		{
			name:     "invalid navs",
			genState: &nav.GenesisState{Navs: nav.NAVRecords{s.newNAVRec("10green", "12green", 14, "sixteen")}},
			expNAVs:  nil,
			expPanic: "0: nav assets \"10green\" and price \"12green\" must have different denoms",
		},
		{
			name:     "one nav",
			genState: &nav.GenesisState{Navs: nav.NAVRecords{s.newNAVRec("2gold", "16silver", 737, "airplane")}},
			expNAVs:  nav.NAVRecords{s.newNAVRec("2gold", "16silver", 737, "airplane")},
		},
		{
			name: "many navs",
			genState: &nav.GenesisState{Navs: nav.NAVRecords{
				s.newNAVRec("1white", "2purple", 1, "one"),
				s.newNAVRec("1white", "4blue", 1, "one"),
				s.newNAVRec("2white", "6green", 1, "one"),
				s.newNAVRec("2white", "8yellow", 2, "two"),
				s.newNAVRec("3white", "10orange", 3, "three"),
				s.newNAVRec("3white", "12red", 3, "three"),
				s.newNAVRec("8purple", "2gray", 4, "four"),
				s.newNAVRec("8blue", "4gray", 5, "five"),
				s.newNAVRec("8green", "6gray", 5, "five"),
				s.newNAVRec("8yellow", "8gray", 5, "five"),
				s.newNAVRec("8orange", "10gray", 5, "five"),
				s.newNAVRec("8red", "12gray", 5, "five"),
				s.newNAVRec("10brown", "13black", 6, "six"),
				s.newNAVRec("10black", "9brown", 6, "six"),
				s.newNAVRec("10black", "4green", 6, "six"),
				s.newNAVRec("12indigo", "5gray", 7, "seven"),
			}},
			expNAVs: nav.NAVRecords{
				s.newNAVRec("10black", "9brown", 6, "six"),
				s.newNAVRec("10black", "4green", 6, "six"),
				s.newNAVRec("8blue", "4gray", 5, "five"),
				s.newNAVRec("10brown", "13black", 6, "six"),
				s.newNAVRec("8green", "6gray", 5, "five"),
				s.newNAVRec("12indigo", "5gray", 7, "seven"),
				s.newNAVRec("8orange", "10gray", 5, "five"),
				s.newNAVRec("8purple", "2gray", 4, "four"),
				s.newNAVRec("8red", "12gray", 5, "five"),
				s.newNAVRec("1white", "4blue", 1, "one"),
				s.newNAVRec("2white", "6green", 1, "one"),
				s.newNAVRec("3white", "10orange", 3, "three"),
				s.newNAVRec("1white", "2purple", 1, "one"),
				s.newNAVRec("3white", "12red", 3, "three"),
				s.newNAVRec("2white", "8yellow", 2, "two"),
				s.newNAVRec("8yellow", "8gray", 5, "five"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx, _ := s.ctx.CacheContext()
			testFunc := func() {
				s.navKeeper.InitGenesis(ctx, tc.genState)
			}
			s.requirePanicEquals(testFunc, tc.expPanic, "InitGenesis")

			var actNAVs nav.NAVRecords
			testGetAll := func() {
				actNAVs = s.navKeeper.GetAllNAVRecords(ctx)
			}
			s.Require().NotPanics(testGetAll, "GetAllNAVRecords")
			s.assertEqualNAVRecords(tc.expNAVs, actNAVs, "navs now in state")
		})
	}
}

func (s *TestSuite) TestKeeper_ExportGenesis() {
	tests := []struct {
		name string
		ini  nav.NAVRecords
		exp  *nav.GenesisState
	}{
		{
			name: "empty state",
			ini:  nil,
			exp:  &nav.GenesisState{},
		},
		{
			name: "one nav",
			ini:  nav.NAVRecords{s.newNAVRec("15copper", "5gold", 1879, "pits")},
			exp:  &nav.GenesisState{Navs: nav.NAVRecords{s.newNAVRec("15copper", "5gold", 1879, "pits")}},
		},
		{
			name: "many navs",
			ini: nav.NAVRecords{
				s.newNAVRec("1white", "2purple", 1, "one"),
				s.newNAVRec("1white", "4blue", 1, "one"),
				s.newNAVRec("2white", "6green", 1, "one"),
				s.newNAVRec("2white", "8yellow", 2, "two"),
				s.newNAVRec("3white", "10orange", 3, "three"),
				s.newNAVRec("3white", "12red", 3, "three"),
				s.newNAVRec("8purple", "2gray", 4, "four"),
				s.newNAVRec("8blue", "4gray", 5, "five"),
				s.newNAVRec("8green", "6gray", 5, "five"),
				s.newNAVRec("8yellow", "8gray", 5, "five"),
				s.newNAVRec("8orange", "10gray", 5, "five"),
				s.newNAVRec("8red", "12gray", 5, "five"),
				s.newNAVRec("10brown", "13black", 6, "six"),
				s.newNAVRec("10black", "9brown", 6, "six"),
				s.newNAVRec("10black", "4green", 6, "six"),
				s.newNAVRec("12indigo", "5gray", 7, "seven"),
			},
			exp: &nav.GenesisState{Navs: nav.NAVRecords{
				s.newNAVRec("10black", "9brown", 6, "six"),
				s.newNAVRec("10black", "4green", 6, "six"),
				s.newNAVRec("8blue", "4gray", 5, "five"),
				s.newNAVRec("10brown", "13black", 6, "six"),
				s.newNAVRec("8green", "6gray", 5, "five"),
				s.newNAVRec("12indigo", "5gray", 7, "seven"),
				s.newNAVRec("8orange", "10gray", 5, "five"),
				s.newNAVRec("8purple", "2gray", 4, "four"),
				s.newNAVRec("8red", "12gray", 5, "five"),
				s.newNAVRec("1white", "4blue", 1, "one"),
				s.newNAVRec("2white", "6green", 1, "one"),
				s.newNAVRec("3white", "10orange", 3, "three"),
				s.newNAVRec("1white", "2purple", 1, "one"),
				s.newNAVRec("3white", "12red", 3, "three"),
				s.newNAVRec("2white", "8yellow", 2, "two"),
				s.newNAVRec("8yellow", "8gray", 5, "five"),
			}},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx, _ := s.ctx.CacheContext()
			s.storeNAVs(ctx, tc.ini)

			var act *nav.GenesisState
			testFunc := func() {
				act = s.navKeeper.ExportGenesis(ctx)
			}
			s.Require().NotPanics(testFunc, "ExportGenesis(ctx)")

			if tc.exp == nil {
				s.Assert().Nil(act, "ExportGenesis(ctx) result")
				return
			}
			s.Require().NotNil(act, "ExportGenesis(ctx) result")
			if s.assertEqualNAVRecords(tc.exp.Navs, act.Navs, "ExportGenesis(ctx) result Navs field") {
				s.Assert().Equal(tc.exp, act, "ExportGenesis(ctx) result")
			}
		})
	}
}
