package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// ValidateSetAccountData makes sure that the msg signers have proper authority to
// set the account data of the provided metadata address.
func (k Keeper) ValidateSetAccountData(ctx sdk.Context, msg *types.MsgSetAccountDataRequest) error {
	// Assume that the metadata address has already been validated (e.g. in ValidateBasic).
	prefix, _ := msg.MetadataAddr.Prefix()
	switch prefix {
	case types.PrefixScope:
		return k.ValidateSetScopeAccountData(ctx, msg)
	default:
		return fmt.Errorf("unsupported metadata address type: %s", prefix)
	}
}
