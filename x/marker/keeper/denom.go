package keeper

import (
	"errors"
	"fmt"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/marker/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateDenomMetadata performs extended validation of the denom metadata fields.
// It checks that:
//  - The proposed metadata passes ValidateDenomMetadataBasic.
//  - The marker status is one that allows the denom metadata to be manipulated.
//  - All DenomUnit Denom and Aliases strings pass the unrestricted denom regex.
//    If there is an existing record:
//     - The Base doesn't change.
//       If marker status is active or finalized:
//        - No DenomUnit entries are removed.
//        - DenomUnit Denom fields aren't changed.
//        - No aliases are removed from a DenomUnit.
func (k Keeper) ValidateDenomMetadata(ctx sdk.Context, proposed banktypes.Metadata, existing *banktypes.Metadata, markerStatus types.MarkerStatus) error {
	// Run all of the basic validation.
	if err := types.ValidateDenomMetadataBasic(proposed); err != nil {
		return fmt.Errorf("invalid proposed metadata: %w", err)
	}

	// Make sure the marker is in a status to allow any denom metadata adding or updating.
	if !markerStatus.IsOneOf(types.StatusProposed, types.StatusActive, types.StatusFinalized) {
		return fmt.Errorf("cannot add or update denom metadata for a marker with status [%s]", markerStatus)
	}

	// Make sure all the DenomUnit Denom and alias strings pass the extra validation regex.
	for _, du := range proposed.DenomUnits {
		if err := k.ValidateUnrestictedDenom(ctx, du.Denom); err != nil {
			return fmt.Errorf("invalid denom unit denom: %w", err)
		}
		for _, a := range du.Aliases {
			if err := k.ValidateUnrestictedDenom(ctx, a); err != nil {
				return fmt.Errorf("invalid denom unit alias: %w", err)
			}
		}
	}

	if existing != nil {
		// No matter what, the base cannot change.
		if proposed.Base != existing.Base {
			return errors.New("denom metadata base value cannot be changed")
		}

		// Some further restrictions apply for active and finalized entries.
		// Note: If you add or remove a status here, you might also need to alter a similar call above.
		if markerStatus.IsOneOf(types.StatusActive, types.StatusFinalized) {
			for _, edu := range existing.DenomUnits {
				// Make sure the existing DenomUnit hasn't been removed.
				var pdu *banktypes.DenomUnit
				for _, du := range proposed.DenomUnits {
					if edu.Exponent == du.Exponent {
						pdu = du
						break
					}
				}
				if pdu == nil {
					return fmt.Errorf("cannot remove denom unit [%s] for a marker with status [%s]",
						edu.Denom, markerStatus)
				}
				// Make sure the Denom value hasn't changed.
				if edu.Denom != pdu.Denom {
					return fmt.Errorf("cannot change denom unit Denom from [%s] to [%s] for a marker with status [%s]",
						edu.Denom, pdu.Denom, markerStatus)
				}
				// Make sure none of the aliases have been removed.
				for _, ea := range edu.Aliases {
					found := false
					for _, pa := range pdu.Aliases {
						if ea == pa {
							found = true
							break
						}
					}
					if !found {
						return fmt.Errorf("cannot remove alias [%s] from denom unit [%s] for a marker with status [%s]",
							ea, edu.Denom, markerStatus)
					}
				}
			}
		}
	}

	return nil
}
