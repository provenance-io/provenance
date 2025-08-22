package keeper_test

func (s *TestSuite) GenesisTest() {
	genesis1 := s.keeper.ExportGenesis(s.ctx)

	// Clear the state
	s.keeper.LedgerClasses.Clear(s.ctx, nil)
	s.keeper.LedgerClassEntryTypes.Clear(s.ctx, nil)
	s.keeper.LedgerClassStatusTypes.Clear(s.ctx, nil)
	s.keeper.LedgerClassBucketTypes.Clear(s.ctx, nil)
	s.keeper.Ledgers.Clear(s.ctx, nil)
	s.keeper.LedgerEntries.Clear(s.ctx, nil)
	s.keeper.FundTransfersWithSettlement.Clear(s.ctx, nil)

	// Import the genesis state
	s.keeper.InitGenesis(s.ctx, genesis1)

	// Export the genesis state
	genesis2 := s.keeper.ExportGenesis(s.ctx)

	// Compare the two genesis states
	s.Require().Equal(genesis1, genesis2)
}
