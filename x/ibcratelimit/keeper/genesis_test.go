package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

func (s *TestSuite) TestInitExportGenesis() {
	testAddress := sdk.AccAddress([]byte("addr1_______________")).String()
	k := s.app.RateLimitingKeeper

	initialGenesis := ibcratelimit.NewGenesisState(ibcratelimit.NewParams(testAddress))

	k.InitGenesis(s.ctx, initialGenesis)
	s.Assert().Equal(testAddress, k.GetContractAddress(s.ctx))
	exportedGenesis := k.ExportGenesis(s.ctx)
	s.Assert().Equal(initialGenesis, exportedGenesis)
}
