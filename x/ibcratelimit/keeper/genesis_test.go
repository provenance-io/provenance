package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

func (s *TestSuite) TestInitExportGenesis() {
	testAddress := sdk.AccAddress([]byte("addr1_______________")).String()
	k := s.app.RateLimitingKeeper

	initialGenesis := ibcratelimit.GenesisState{
		Params: ibcratelimit.Params{
			ContractAddress: testAddress,
		},
	}

	k.InitGenesis(s.ctx, &initialGenesis)
	s.Require().Equal(testAddress, k.GetContractAddress(s.ctx))
	exportedGenesis := k.ExportGenesis(s.ctx)
	s.Require().Equal(initialGenesis, *exportedGenesis)
}
