package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	attrTypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/marker/types"
)

const AddressHasAccessKey = "address_has_access"

func (k Keeper) AllowMarkerSend(ctx sdk.Context, from, to, denom string) error {
	markerAddr := types.MustGetMarkerAddress(denom)
	marker, err := k.GetMarker(ctx, markerAddr)
	if err != nil {
		return err
	}
	if marker == nil { // this should only occur in tests
		return nil
	}

	if marker.GetMarkerType() != types.MarkerType_RestrictedCoin {
		return nil
	}

	caller, err := sdk.AccAddressFromBech32(from)
	if err != nil {
		return err
	}

	hasAccess, err := GetAddressHasAccess(ctx)
	if err != nil {
		return err
	}

	// address used for adjusting circulation
	moduleAdrr := k.authKeeper.GetModuleAddress(types.CoinPoolName)

	// if the marker has authority it is allowed to send to receiver without checking of attributes
	if hasAccess ||
		marker.AddressHasAccess(caller, types.Access_Transfer) ||
		moduleAdrr.String() == from {
		return nil
	}

	if len(marker.GetRequiredAttributes()) == 0 {
		return fmt.Errorf("%s does not have transfer permissions", caller.String())
	}
	contains, err := k.ContainsRequiredAttributes(ctx, marker.GetRequiredAttributes(), to)
	if err != nil {
		return err
	}
	if !contains {
		return fmt.Errorf("address %s does not contain the required attributes %v", to, marker.GetRequiredAttributes())
	}
	return nil
}

// NormalizeRequiredAttributes normalizes the required attribute names using name module's Normalize method
func (k Keeper) NormalizeRequiredAttributes(ctx sdk.Context, requiredAttributes []string) ([]string, error) {
	maxLength := int(k.attrKeeper.GetMaxValueLength(ctx))
	result := make([]string, len(requiredAttributes))
	for i, attr := range requiredAttributes {
		if len(attr) > maxLength {
			return nil, fmt.Errorf("required attribute %v length is too long %v : %v ", attr, len(attr), maxLength)
		}

		// for now just check if required attribute starts with a *
		var prefix string
		if ContainsWildCard(attr) {
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

// ContainsRequiredAttributes retrieves the attributes from address and checks that all required attributes are present
func (k Keeper) ContainsRequiredAttributes(ctx sdk.Context, requiredAttributes []string, address string) (bool, error) {
	attributes, err := k.attrKeeper.GetAllAttributes(ctx, address)
	if err != nil {
		return false, err
	}
	return EnsureAllRequiredAttributesExist(requiredAttributes, attributes), nil
}

// EnsureAllRequiredAttributesExist checks that all requiredAttributes are in attributes list
func EnsureAllRequiredAttributesExist(requiredAttributes []string, attributes []attrTypes.Attribute) bool {
	for _, reqAttr := range requiredAttributes {
		var match bool
		for _, attr := range attributes {
			match = MatchAttribute(reqAttr, attr.Name)
			if match {
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}

func GetAddressHasAccess(ctx sdk.Context) (bool, error) {
	hasAccess := ctx.Value(AddressHasAccessKey)
	if hasAccess == nil {
		return false, nil
	}
	accessAllowed, success := hasAccess.(bool)
	if !success {
		return false, fmt.Errorf("incorrect type for context %s value", AddressHasAccessKey)
	}
	return accessAllowed, nil
}

func SetAddressHasAccess(ctx sdk.Context, hasAccess bool) sdk.Context {
	return ctx.WithValue(AddressHasAccessKey, hasAccess)
}

// MatchAttribute compares required attribute against attribute string
func MatchAttribute(reqAttr string, attr string) bool {
	if len(reqAttr) < 1 {
		return false
	}
	if strings.HasPrefix(reqAttr, "*.") {
		return strings.HasSuffix(attr, reqAttr[2:])
	}
	return reqAttr == attr
}

// ContainsWildCard checks if attribute starts with wildcard
func ContainsWildCard(attr string) bool {
	segs := strings.Split(attr, ".")
	return len(segs) > 1 && segs[0] == "*"
}
