package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	attrTypes "github.com/provenance-io/provenance/x/attribute/types"
)

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

func (k Keeper) ContainsRequiredAttributes(ctx sdk.Context, requiredAttributes []string, address string) (bool, error) {
	attributes, err := k.attrKeeper.GetAllAttributes(ctx, address)
	if err != nil {
		return false, err
	}

	return EnsureAllRequiredAttributesExist(requiredAttributes, attributes), nil
}

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

func MatchAttribute(reqAttr string, attr string) bool {
	if len(reqAttr) < 1 {
		return false
	}
	if strings.HasPrefix(reqAttr, "*.") {

		return strings.HasSuffix(attr, reqAttr[2:])
	}
	return reqAttr == attr
}

func ContainsWildCard(attr string) bool {
	segs := strings.Split(attr, ".")
	return len(segs) > 1 && segs[0] == "*"
}
