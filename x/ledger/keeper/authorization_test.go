package keeper_test

import (
	ledger "github.com/provenance-io/provenance/x/ledger/types"
	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

// TestRequireAuthorization_NoRegistry_Owner verifies owner is authorized when no registry exists.
func (s *TestSuite) TestRequireAuthorization_NoRegistry_Owner() {
	rk := &registrytypes.RegistryKey{AssetClassId: s.validNFTClass.Id, NftId: s.validNFT.Id}
	err := s.keeper.RequireAuthorization(s.ctx, s.addr1.String(), rk)
	s.Require().NoError(err)
}

// TestRequireAuthorization_NoRegistry_NonOwner verifies non-owner is unauthorized when no registry exists.
func (s *TestSuite) TestRequireAuthorization_NoRegistry_NonOwner() {
	rk := &registrytypes.RegistryKey{AssetClassId: s.validNFTClass.Id, NftId: s.validNFT.Id}
	err := s.keeper.RequireAuthorization(s.ctx, s.addr2.String(), rk)
	s.Require().ErrorIs(err, ledger.ErrUnauthorized)
}

// TestRequireAuthorization_WithServicer verifies servicer is authorized and owner is denied when a servicer role exists.
func (s *TestSuite) TestRequireAuthorization_WithServicer() {
	rk := &registrytypes.RegistryKey{AssetClassId: s.validNFTClass.Id, NftId: s.validNFT.Id}
	roles := []registrytypes.RolesEntry{{Role: registrytypes.RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{s.addr2.String()}}}
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, rk, roles))

	// Servicer allowed
	err := s.keeper.RequireAuthorization(s.ctx, s.addr2.String(), rk)
	s.Require().NoError(err)

	// Owner explicitly denied when servicer exists
	err = s.keeper.RequireAuthorization(s.ctx, s.addr1.String(), rk)
	s.Require().ErrorIs(err, ledger.ErrUnauthorized)
}

// TestRequireAuthorization_EmptyServicerAddresses falls back to owner auth when servicer role has no addresses.
func (s *TestSuite) TestRequireAuthorization_EmptyServicerAddresses() {
	rk := &registrytypes.RegistryKey{AssetClassId: s.validNFTClass.Id, NftId: s.validNFT.Id}
	roles := []registrytypes.RolesEntry{{Role: registrytypes.RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{}}}
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, rk, roles))

	// Owner allowed
	err := s.keeper.RequireAuthorization(s.ctx, s.addr1.String(), rk)
	s.Require().NoError(err)

	// Non-owner denied
	err = s.keeper.RequireAuthorization(s.ctx, s.addr3.String(), rk)
	s.Require().ErrorIs(err, ledger.ErrUnauthorized)
}
