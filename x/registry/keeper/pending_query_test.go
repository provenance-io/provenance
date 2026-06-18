package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/provenance-io/provenance/x/registry/types"
)

// newPendingChange builds and stores a pending role change for the given key and addresses.
func (s *KeeperTestSuite) newPendingChange(key *types.RegistryKey, addresses []string) types.PendingRoleChange {
	op := types.RoleChangeOperation_ROLE_CHANGE_OPERATION_GRANT
	role := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	change := types.PendingRoleChange{
		Id:        types.NewPendingRoleChangeID(key, role, op, addresses),
		Key:       key,
		Role:      role,
		Operation: op,
		Addresses: addresses,
		Proposer:  s.user1,
		Approvals: []string{s.user1},
	}
	s.Require().NoError(s.app.RegistryKeeper.SetPendingRoleChange(s.ctx, change))
	return change
}

func (s *KeeperTestSuite) TestQueryPendingRoleChange() {
	key := &types.RegistryKey{AssetClassId: s.validNFTClass.Id, NftId: s.validNFT.Id}
	change := s.newPendingChange(key, []string{s.user2})

	queryServer := keeper.NewQueryServer(s.app.RegistryKeeper)

	tests := []struct {
		name   string
		req    *types.QueryPendingRoleChangeRequest
		expErr string
		expID  string
	}{
		{
			name:   "nil request",
			req:    nil,
			expErr: "empty request",
		},
		{
			name:   "empty id",
			req:    &types.QueryPendingRoleChangeRequest{Id: ""},
			expErr: "id cannot be empty",
		},
		{
			name:   "not found",
			req:    &types.QueryPendingRoleChangeRequest{Id: "does-not-exist"},
			expErr: "pending role change",
		},
		{
			name:  "found",
			req:   &types.QueryPendingRoleChangeRequest{Id: change.Id},
			expID: change.Id,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := queryServer.PendingRoleChange(s.ctx, tc.req)
			if tc.expErr != "" {
				assertions.AssertErrorContents(s.T(), err, []string{tc.expErr}, "PendingRoleChange")
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Equal(tc.expID, resp.PendingRoleChange.Id)
		})
	}
}

func (s *KeeperTestSuite) TestQueryPendingRoleChanges() {
	key1 := &types.RegistryKey{AssetClassId: s.validNFTClass.Id, NftId: s.validNFT.Id}
	key2 := &types.RegistryKey{AssetClassId: "other-class", NftId: "other-nft"}

	c1 := s.newPendingChange(key1, []string{s.user2})
	c2 := s.newPendingChange(key2, []string{s.user2})

	queryServer := keeper.NewQueryServer(s.app.RegistryKeeper)

	s.Run("nil request", func() {
		_, err := queryServer.PendingRoleChanges(s.ctx, nil)
		assertions.AssertErrorContents(s.T(), err, []string{"empty request"}, "PendingRoleChanges")
	})

	s.Run("all changes", func() {
		resp, err := queryServer.PendingRoleChanges(s.ctx, &types.QueryPendingRoleChangesRequest{})
		s.Require().NoError(err)
		s.Require().Len(resp.PendingRoleChanges, 2)
		ids := []string{resp.PendingRoleChanges[0].Id, resp.PendingRoleChanges[1].Id}
		s.Require().ElementsMatch([]string{c1.Id, c2.Id}, ids)
	})

	s.Run("filter by key", func() {
		resp, err := queryServer.PendingRoleChanges(s.ctx, &types.QueryPendingRoleChangesRequest{Key: key1})
		s.Require().NoError(err)
		s.Require().Len(resp.PendingRoleChanges, 1)
		s.Require().Equal(c1.Id, resp.PendingRoleChanges[0].Id)
	})

	s.Run("filter by key with no matches", func() {
		empty := &types.RegistryKey{AssetClassId: "nope", NftId: "nope"}
		resp, err := queryServer.PendingRoleChanges(s.ctx, &types.QueryPendingRoleChangesRequest{Key: empty})
		s.Require().NoError(err)
		s.Require().Empty(resp.PendingRoleChanges)
	})

	s.Run("paginated", func() {
		resp, err := queryServer.PendingRoleChanges(s.ctx, &types.QueryPendingRoleChangesRequest{
			Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
		})
		s.Require().NoError(err)
		s.Require().Len(resp.PendingRoleChanges, 1)
		s.Require().NotNil(resp.Pagination)
		s.Require().Equal(uint64(2), resp.Pagination.Total)
	})
}

func (s *KeeperTestSuite) TestPendingRoleChangeGenesisRoundTrip() {
	key := &types.RegistryKey{AssetClassId: s.validNFTClass.Id, NftId: s.validNFT.Id}
	change := s.newPendingChange(key, []string{s.user2})

	genesis := s.app.RegistryKeeper.ExportGenesis(s.ctx)
	s.Require().Len(genesis.PendingRoleChanges, 1)
	s.Require().Equal(change.Id, genesis.PendingRoleChanges[0].Id)

	// Clear and re-import.
	s.Require().NoError(s.app.RegistryKeeper.RemovePendingRoleChange(s.ctx, change.Id))
	got, err := s.app.RegistryKeeper.GetPendingRoleChange(s.ctx, change.Id)
	s.Require().NoError(err)
	s.Require().Nil(got)

	s.Require().NotPanics(func() {
		s.app.RegistryKeeper.InitGenesis(s.ctx, genesis)
	})

	got, err = s.app.RegistryKeeper.GetPendingRoleChange(s.ctx, change.Id)
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().Equal(change.Approvals, got.Approvals)
	s.Require().Equal(change.Addresses, got.Addresses)
}
