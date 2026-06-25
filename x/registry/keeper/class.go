package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/registry/types"
)

// GetRegistryClass returns the registry class for the given id. If no class exists for the id, it
// returns nil, nil.
func (k Keeper) GetRegistryClass(ctx context.Context, registryClassID string) (*types.RegistryClass, error) {
	class, err := k.RegistryClasses.Get(ctx, registryClassID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &class, nil
}

// HasRegistryClass reports whether a registry class exists for the given id.
func (k Keeper) HasRegistryClass(ctx context.Context, registryClassID string) (bool, error) {
	return k.RegistryClasses.Has(ctx, registryClassID)
}

// validateRegistryClassForEntry verifies that an entry may reference the given registry class id.
// When the id is empty there is nothing to check. Otherwise the class must exist and its asset
// class id must match the entry's asset class id; a registry class only governs entries within its
// own asset class, so a mismatch would resolve authorization against the wrong policy tier.
func (k Keeper) validateRegistryClassForEntry(ctx context.Context, registryClassID, assetClassID string) error {
	if registryClassID == "" {
		return nil
	}
	class, err := k.GetRegistryClass(ctx, registryClassID)
	if err != nil {
		return err
	}
	if class == nil {
		return types.NewErrCodeRegistryClassNotFound(registryClassID)
	}
	if class.AssetClassId != assetClassID {
		return types.NewErrCodeRegistryClassAssetMismatch(registryClassID, class.AssetClassId, assetClassID)
	}
	return nil
}

// SetRegistryClass stores a registry class in state.
func (k Keeper) SetRegistryClass(ctx context.Context, class types.RegistryClass) error {
	return k.RegistryClasses.Set(ctx, class.RegistryClassId, class)
}

// CreateRegistryClass stores a new registry class. It returns an error if a class already exists
// with the same id.
func (k Keeper) CreateRegistryClass(ctx context.Context, class types.RegistryClass) error {
	if err := class.Validate(); err != nil {
		return err
	}

	has, err := k.RegistryClasses.Has(ctx, class.RegistryClassId)
	if err != nil {
		return fmt.Errorf("could not check if registry class exists: %w", err)
	}
	if has {
		return types.NewErrCodeRegistryClassExists(class.RegistryClassId)
	}

	if err = k.SetRegistryClass(ctx, class); err != nil {
		return fmt.Errorf("could not set registry class: %w", err)
	}

	k.EmitEvent(ctx, types.NewEventRegistryClassCreated(&class))
	return nil
}

// UpdateRegistryClassRoleAuthorization replaces the authorization rules for an existing registry
// class. Only the current maintainer (the provided signer) may update the rules. The maintainer
// and asset class id are immutable.
func (k Keeper) UpdateRegistryClassRoleAuthorization(ctx context.Context, signer, registryClassID string, roleAuths []types.RoleAuthorization) error {
	class, err := k.GetRegistryClass(ctx, registryClassID)
	if err != nil {
		return fmt.Errorf("could not get registry class: %w", err)
	}
	if class == nil {
		return types.NewErrCodeRegistryClassNotFound(registryClassID)
	}

	if signer != class.Maintainer {
		return types.NewErrCodeUnauthorized(
			fmt.Sprintf("signer %q is not the maintainer of registry class %q", signer, registryClassID),
		)
	}

	class.RoleAuthorizations = roleAuths
	if err = class.Validate(); err != nil {
		return err
	}
	if err = k.SetRegistryClass(ctx, *class); err != nil {
		return fmt.Errorf("could not set registry class: %w", err)
	}

	k.EmitEvent(ctx, types.NewEventRegistryClassUpdated(class))
	return nil
}

// GetRegistryClasses returns the registry classes paginated.
func (k Keeper) GetRegistryClasses(ctx context.Context, pagination *query.PageRequest) ([]types.RegistryClass, *query.PageResponse, error) {
	ptrs, pageRes, err := query.CollectionPaginate(ctx, k.RegistryClasses, pagination, func(_ string, class types.RegistryClass) (*types.RegistryClass, error) {
		return &class, nil
	})
	if err != nil {
		return nil, nil, err
	}

	classes := make([]types.RegistryClass, len(ptrs))
	for i, p := range ptrs {
		classes[i] = *p
	}
	return classes, pageRes, nil
}

// roleAuthorizationsForEntry resolves the effective role authorization policies that govern role
// updates for the given registry entry, implementing the two-tier resolution:
//
//  1. Registry class level (highest priority): if the entry references a registry class that
//     exists, use the authorization rules defined by that class.
//  2. Module params default (fallback): otherwise use the module's default policies (governance-
//     managed via MsgUpdateParams). Roles not covered by the returned map fall back to legacy
//     NFT-owner authorization at the call site.
func (k Keeper) roleAuthorizationsForEntry(ctx context.Context, entry *types.RegistryEntry) map[types.RegistryRole]types.RoleAuthorization {
	if entry != nil && entry.RegistryClassId != "" {
		class, err := k.GetRegistryClass(ctx, entry.RegistryClassId)
		if err != nil {
			// A store/read error must not silently downgrade to a weaker fallback policy. Fail fast
			// so the authorization decision is never made against the wrong tier.
			panic(fmt.Errorf("could not resolve registry class %q for authorization: %w", entry.RegistryClassId, err))
		}
		if class != nil {
			// Defense in depth: enforcement at write time (register, bulk update, genesis) keeps the
			// referenced class scoped to the entry's asset class. If a mismatched class is ever found
			// in state, fail fast rather than resolve authorization against the wrong policy tier.
			if entry.Key != nil && class.AssetClassId != entry.Key.AssetClassId {
				panic(types.NewErrCodeRegistryClassAssetMismatch(entry.RegistryClassId, class.AssetClassId, entry.Key.AssetClassId))
			}
			return types.RoleAuthorizationMapFrom(class.RoleAuthorizations)
		}
	}
	return k.GetParams(ctx).RoleAuthorizationMap()
}
