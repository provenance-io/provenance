package keeper

import (
	"context"
	"slices"

	"github.com/provenance-io/provenance/x/ledger/types"
	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

// RequireAuthorization performs authorization checks to determine if the given address
// is allowed to interact with ledger transactions for a specific NFT.
// This function serves as the main entry point for authorization validation.
// It uses a two-tier authorization system:
// 1. NFT Ownership: Direct ownership of the NFT grants authorization
// 2. Registry Servicer Role: Registered servicers can act on behalf of NFT owners
//
// Parameters:
// - ctx: The SDK context
// - addr: The address requesting authorization
// - key: The registry key containing asset class ID and NFT ID
//
// Returns an error if the address is not authorized, nil otherwise.
func (k Keeper) RequireAuthorization(ctx context.Context, addr string, key *registrytypes.RegistryKey) error {
	return assertAuthorization(ctx, k.RegistryKeeper, addr, key)
}

// assertOwner verifies that the given address is the direct owner of the NFT.
// This is the most basic level of authorization - direct NFT ownership.
//
// Parameters:
// - ctx: The SDK context
// - k: The registry keeper for NFT ownership queries
// - signerAddr: The address to check for ownership
// - ledgerKey: The ledger key containing asset class ID and NFT ID
//
// Returns an error if the address is not the NFT owner, nil otherwise.
func assertOwner(ctx context.Context, k RegistryKeeper, signerAddr string, ledgerKey *types.LedgerKey) error {
	// Check if the address has ownership of the NFT
	nftOwner := k.GetNFTOwner(ctx, &ledgerKey.AssetClassId, &ledgerKey.NftId)
	if len(nftOwner) == 0 || nftOwner.String() != signerAddr {
		return types.NewErrCodeUnauthorized("signer is not the nft owner")
	}

	return nil
}

// assertAuthorization implements the complete authorization logic for ledger transactions.
// This function determines if an address has authorization to perform ledger operations
// based on a hierarchical authorization system:
//
// Authorization Logic:
//  1. No Registry Entry: If no registry entry exists, authorization is based on NFT ownership
//  2. Registry Entry with Servicer: If a servicer role is registered, only the servicer is authorized
//     (NFT owner is explicitly denied when a servicer exists)
//  3. Registry Entry without Servicer: If registry entry exists but no servicer role,
//     authorization falls back to NFT ownership
//
// The registry module acts as a gate that can override direct NFT ownership,
// allowing for delegated authorization through registered servicers. This enables
// scenarios where NFT owners can delegate management of their assets to trusted
// third parties while maintaining security.
//
// Parameters:
// - ctx: The SDK context
// - k: The registry keeper for registry and NFT queries
// - signerAddr: The address requesting authorization
// - rk: The registry key containing asset class ID and NFT ID
//
// Returns:
// - error: nil if authorized, error details if authorization fails
func assertAuthorization(ctx context.Context, k RegistryKeeper, signerAddr string, rk *registrytypes.RegistryKey) error {
	if k == nil {
		return types.NewErrCodeInternal("registry keeper is nil")
	}
	if rk == nil {
		return types.NewErrCodeInternal("registry key is nil")
	}

	// Get the registry entry for the NFT to determine if the address has the servicer role.
	// The registry entry contains role assignments that can override direct NFT ownership.
	registryEntry, err := k.GetRegistry(ctx, rk)
	if err != nil {
		return err
	}

	lk := types.NewLedgerKey(rk.AssetClassId, rk.NftId)

	// If there is no registry entry, authorization is based on ownership.
	// This is the default case where no delegation has been set up.
	if registryEntry == nil {
		err = assertOwner(ctx, k, signerAddr, lk)
		if err != nil {
			return err
		}

		return nil
	}

	// Check if there is a registered servicer for this NFT.
	// If a servicer exists, only the servicer can perform ledger operations,
	// even if the NFT owner tries to act directly.
	hasServicer := false
	for _, role := range registryEntry.Roles {
		if role.Role != registrytypes.RegistryRole_REGISTRY_ROLE_SERVICER {
			continue
		}
		hasServicer = true

		if slices.Contains(role.Addresses, signerAddr) {
			return nil
		}
	}
	if hasServicer {
		return types.NewErrCodeUnauthorized("owner is not the registered servicer")
	}

	// Since there isn't a registered servicer, let's see if the address is the owner.
	// This fallback ensures that NFT owners retain authorization when no servicer is assigned.
	err = assertOwner(ctx, k, signerAddr, lk)
	if err == nil {
		// The address owns the asset, and there is no registered servicer
		return nil
	}

	return err
}
