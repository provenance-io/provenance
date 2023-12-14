package keeper

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	attrTypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/marker/types"
)

const AddressHasAccessKey = "address_has_access"

var _ banktypes.SendRestrictionFn = Keeper{}.SendRestrictionFn

func (k Keeper) SendRestrictionFn(goCtx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// In some cases, it might not be possible to add a bypass to the context.
	// If it's from either the Marker or IBC Transfer module accounts, assume proper validation has been done elsewhere.
	if types.HasBypass(ctx) || fromAddr.Equals(k.markerModuleAddr) || fromAddr.Equals(k.ibcTransferModuleAddr) {
		return toAddr, nil
	}

	for _, coin := range amt {
		if err := k.validateSendDenom(ctx, fromAddr, toAddr, coin.Denom); err != nil {
			return nil, err
		}
	}

	return toAddr, nil
}

// validateSendDenom makes sure a send of the given denom is allowed for the given addresses.
// This is NOT the validation that is needed for the marker Transfer endpoint.
func (k Keeper) validateSendDenom(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, denom string) error {
	// If it's going to a restricted marker, fromAddr must have deposit access on that marker.
	var toMarker types.MarkerAccountI
	if toAcct := k.authKeeper.GetAccount(ctx, toAddr); toAcct != nil {
		toAcctAsMarker, isMarker := toAcct.(types.MarkerAccountI)
		if isMarker {
			toMarker = toAcctAsMarker
		}
	}
	if toMarker != nil && toMarker.GetMarkerType() == types.MarkerType_RestrictedCoin && !toMarker.AddressHasAccess(fromAddr, types.Access_Deposit) {
		return fmt.Errorf("%s does not have deposit access for %s (%s)", fromAddr.String(), toAddr.String(), toMarker.GetDenom())
	}

	markerAddr := types.MustGetMarkerAddress(denom)
	marker, err := k.GetMarker(ctx, markerAddr)
	if err != nil {
		return err
	}

	// If there's no marker for the denom, or it's not a restricted marker, there's nothing more to do here.
	if marker == nil || marker.GetMarkerType() != types.MarkerType_RestrictedCoin {
		return nil
	}

	// If from address is in deny list, prevent sending of restricted marker
	if k.IsSendDeny(ctx, markerAddr, fromAddr) {
		return fmt.Errorf("%s is on deny list for sending restricted marker", fromAddr.String())
	}

	// If the fromAddr has transfer access, there's nothing left to check.
	if marker.AddressHasAccess(fromAddr, types.Access_Transfer) {
		return nil
	}

	reqAttr := marker.GetRequiredAttributes()

	// If going to a marker, transfer permission is required regardless of whether it's coming from a bypass.
	// If someone wants to deposit funds from a bypass account, they can either send the funds to a valid
	// intermediary account and deposit them from there, or give the bypass account deposit and transfer permissions.
	// It's assumed that a marker address cannot be in the bypass list.
	if toMarker != nil {
		return fmt.Errorf("%s does not have transfer permissions", fromAddr.String())
	}

	// If there aren't any required attributes, transfer permission is required unless coming from a bypass account.
	// It's assumed that the only way the restricted coins without required attributes can got into a bypass
	// account is by someone with transfer permission, which is then conveyed for this transfer too.
	if len(reqAttr) == 0 {
		if k.IsReqAttrBypassAddr(fromAddr) {
			return nil
		}
		return fmt.Errorf("%s does not have transfer permissions", fromAddr.String())
	}

	// At this point, we know there are required attributes and that fromAddr does not have transfer permission.
	// If the toAddress has a bypass, skip checking the attributes and allow the transfer.
	// When these funds are then being moved out of the bypass account, attributes are checked on that destination.
	if k.IsReqAttrBypassAddr(toAddr) {
		return nil
	}

	attributes, err := k.attrKeeper.GetAllAttributesAddr(ctx, toAddr)
	if err != nil {
		return fmt.Errorf("could not get attributes for %s: %w", toAddr.String(), err)
	}
	missing := findMissingAttributes(reqAttr, attributes)
	if len(missing) != 0 {
		pl := ""
		if len(missing) != 1 {
			pl = "s"
		}
		return fmt.Errorf("address %s does not contain the %q required attribute%s: \"%s\"", toAddr.String(), denom, pl, strings.Join(missing, `", "`))
	}

	return nil
}

// findMissingAttributes returns all entries in required that don't pass
// MatchAttribute on at least one of the provided attribute names.
func findMissingAttributes(required []string, attributes []attrTypes.Attribute) []string {
	var rv []string
reqLoop:
	for _, req := range required {
		for _, attr := range attributes {
			if MatchAttribute(req, attr.Name) {
				continue reqLoop
			}
		}
		rv = append(rv, req)
	}
	return rv
}

// NormalizeRequiredAttributes normalizes the required attribute names using name module's Normalize method
func (k Keeper) NormalizeRequiredAttributes(ctx sdk.Context, requiredAttributes []string) ([]string, error) {
	maxLength := int(k.attrKeeper.GetMaxValueLength(ctx))
	result := make([]string, len(requiredAttributes))
	for i, attr := range requiredAttributes {
		if len(attr) > maxLength {
			return nil, fmt.Errorf("required attribute %v length is too long %v : %v ", attr, len(attr), maxLength)
		}

		// for now just check if required attribute starts with a *.
		var prefix string
		if strings.HasPrefix(attr, "*.") {
			prefix = attr[:2]
			attr = attr[2:]
		}
		normalizedAttr, err := k.nameKeeper.Normalize(ctx, attr)
		if err != nil {
			return nil, err
		}
		result[i] = fmt.Sprintf("%s%s", prefix, normalizedAttr)
	}
	return result, nil
}

// MatchAttribute returns true if the provided attr satisfies the reqAttr.
func MatchAttribute(reqAttr string, attr string) bool {
	if len(reqAttr) < 1 {
		return false
	}
	if strings.HasPrefix(reqAttr, "*.") {
		// [1:] because we only want to ignore the '*'; the '.' needs to be part of the check.
		return strings.HasSuffix(attr, reqAttr[1:])
	}
	return reqAttr == attr
}
