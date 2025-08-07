package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

// RequireAuthority performs authorization checks to determine if the given address
// is allowed to interact with ledger transactions for a specific NFT.
// This function serves as the main entry point for authority validation.
// It uses a two-tier authorization system:
// 1. NFT Ownership: Direct ownership of the NFT grants authority
// 2. Registry Servicer Role: Registered servicers can act on behalf of NFT owners
//
// Parameters:
// - ctx: The SDK context
// - addr: The address requesting authorization
// - key: The registry key containing asset class ID and NFT ID
//
// Returns an error if the address is not authorized, nil otherwise.
func (k Keeper) RequireAuthority(ctx sdk.Context, addr string, key *registrytypes.RegistryKey) error {
	has, err := assertAuthority(ctx, k.RegistryKeeper, addr, key)
	if err != nil {
		return err
	}
	if !has {
		return types.NewErrCodeUnauthorized("authority is not the owner or servicer")
	}
	return nil
}

// assertOwner verifies that the given authority address is the direct owner of the NFT.
// This is the most basic level of authorization - direct NFT ownership.
//
// Parameters:
// - ctx: The SDK context
// - k: The registry keeper for NFT ownership queries
// - authorityAddr: The address to check for ownership
// - ledgerKey: The ledger key containing asset class ID and NFT ID
//
// Returns an error if the address is not the NFT owner, nil otherwise.
func assertOwner(ctx sdk.Context, k RegistryKeeper, authorityAddr string, ledgerKey *types.LedgerKey) error {
	// Check if the authority has ownership of the NFT
	nftOwner := k.GetNFTOwner(ctx, &ledgerKey.AssetClassId, &ledgerKey.NftId)
	if nftOwner == nil || nftOwner.String() != authorityAddr {
		return types.NewErrCodeUnauthorized("authority is not the nft owner")
	}

	return nil
}

// assertAuthority implements the complete authorization logic for ledger transactions.
// This function determines if an address has authority to perform ledger operations
// based on a hierarchical authorization system:
//
// Authorization Hierarchy:
// 1. Registry Servicer Role: If a servicer is registered for the NFT, only the servicer can act
// 2. NFT Ownership: If no servicer is registered, the NFT owner has authority
//
// The registry module acts as a gate that can override direct NFT ownership,
// allowing for delegated authority through registered servicers. This enables
// scenarios where NFT owners can delegate management of their assets to trusted
// third parties while maintaining security.
//
// Parameters:
// - ctx: The SDK context
// - k: The registry keeper for registry and NFT queries
// - authorityAddr: The address requesting authorization
// - rk: The registry key containing asset class ID and NFT ID
//
// Returns:
// - bool: true if authorized, false otherwise
// - error: error details if authorization fails
func assertAuthority(ctx sdk.Context, k RegistryKeeper, authorityAddr string, rk *registrytypes.RegistryKey) (bool, error) {
	// Get the registry entry for the NFT to determine if the authority has the servicer role.
	// The registry entry contains role assignments that can override direct NFT ownership.
	registryEntry, err := k.GetRegistry(ctx, rk)
	if err != nil {
		return false, err
	}

	lk := types.NewLedgerKey(rk.AssetClassId, rk.NftId)

	// If there is no registry entry, the authority is the owner.
	// This is the default case where no delegation has been set up.
	if registryEntry == nil {
		err = assertOwner(ctx, k, authorityAddr, lk)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	// Check if there is a registered servicer for this NFT.
	// If a servicer exists, only the servicer can perform ledger operations,
	// even if the NFT owner tries to act directly.
	for _, role := range registryEntry.Roles {
		if role.Role == registrytypes.RegistryRole_REGISTRY_ROLE_SERVICER {
			for _, address := range role.Addresses {
				// Check if the authority is the servicer
				if address == authorityAddr {
					return true, nil
				}
			}

			// Since there is a registered servicer, the owner is not authorized.
			// This enforces the delegation model where servicers have exclusive rights.
			return false, types.NewErrCodeUnauthorized("owner is not theregistered servicer")
		}
	}

	// Since there isn't a registered servicer, let's see if the authority is the owner.
	// This fallback ensures that NFT owners retain authority when no servicer is assigned.
	err = assertOwner(ctx, k, authorityAddr, lk)
	if err == nil {
		// The authority owns the asset, and there is no registered servicer
		return true, nil

	}

	return false, err
}
