package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	maxDenomMetadataDescriptionLength = 200
)

// ValidateDenomMetadataBasic performs validation of the denom metadata fields.
// It checks that:
//  - Base and Display denominations are valid coin denominations
//  - Base and Display denominations are present in the DenomUnit slice
//  - The first denomination unit entry is the Base denomination and has Exponent 0
//  - Denomination units are sorted in ascending order by Exponent
//  - Description is no more than 200 characters.
//  - All Denomination unit Denom and Alias strings contain the same root name.
//  - That root name is a valid coin denomination.
//  - All Denomination unit Denom and Alias strings are valid coin denominations
//  - All Denomination unit Denom and Alias strings are a SI prefix + the root name (or just the root name).
//  - All Denomination unit Denom and Alias strings are unique.
//  - All Denomination unit Aliases have the same SI prefix as their Denom (but maybe different forms, e.g. name vs symbol)
//  - All Denomination unit Exponents are {SI prefix exponent of the Denom} - {SI prefix exponent of the base}.
func ValidateDenomMetadataBasic(md banktypes.Metadata) error {
	if err := md.Validate(); err != nil {
		return fmt.Errorf("denom metadata %w", err)
	}
	if len(md.Description) > maxDenomMetadataDescriptionLength {
		return fmt.Errorf("denom metadata description too long (expected <= %d, actual: %d)",
			maxDenomMetadataDescriptionLength, len(md.Description))
	}

	rootCoinName := GetRootCoinName(md)
	if len(rootCoinName) == 0 {
		return errors.New("denom metadata root coin name could not be found, invalid DenomUnit denom and alias values")
	}
	if _, err := validateDenom(rootCoinName, rootCoinName); err != nil {
		return fmt.Errorf("denom metadata root coin name %w", err)
	}

	basePrefix, ok := ParseSIPrefixedString(md.Base, rootCoinName)
	if !ok {
		return fmt.Errorf("denom metadata base [%s] is not a SI prefix + root coin name [%s]",
			md.Base, rootCoinName)
	}
	basePrefixExp := basePrefix.GetExponent()

	// Make sure all the DenomUnits are valid and that the denom and alias strings are unique.
	seenNames := make(map[string]bool)
	for _, du := range md.DenomUnits {
		if seenNames[du.Denom] {
			return fmt.Errorf("denom metadata denom or alias [%s] is not unique", du.Denom)
		}
		seenNames[du.Denom] = true
		for _, a := range du.Aliases {
			if seenNames[a] {
				return fmt.Errorf("denom metadata denom or alias [%s] is not unique", a)
			}
			seenNames[a] = true
		}
		if err := denomUnitValidateBasic(du, rootCoinName, basePrefixExp); err != nil {
			return fmt.Errorf("denom metadata invalid denom unit :%w", err)
		}
	}

	return nil
}

// GetRootCoinName gathers all the names (Denom or Alias) and tries to find a common root name for them all.
// An empty string indicates that there is no common root among all the names.
func GetRootCoinName(md banktypes.Metadata) string {
	// First get all the names (Denom and Alias strings) together in one place.
	allNames := []string{}
	for _, du := range md.DenomUnits {
		allNames = append(allNames, du.Denom)
		if len(du.Aliases) > 0 {
			allNames = append(allNames, du.Aliases...)
		}
	}
	// If there's zero or one names total, there's no way of telling the root name. Return that there isn't one.
	if len(allNames) < 2 {
		return ""
	}

	// Now find the shortest one.
	shortest := allNames[0]
	for _, n := range allNames {
		if len(n) < len(shortest) {
			shortest = n
		}
	}

	// The shortest one is either the root name, or contains the root name.
	// First, check if all other names end with it.
	// If not, remove the leftmost character and try again.
	// Repeat until it passes or there's nothing left.
	for i := 0; i < len(shortest); i++ {
		rootName := shortest[i:]
		allGood := true
		for _, n := range allNames {
			if rootName != n[len(n)-len(rootName):] {
				allGood = false
				break
			}
		}
		if allGood {
			return rootName
		}
	}

	// Nothing found. There isn't a common root name.
	return ""
}

// denomUnitValidateBasic performs validation of the denom unit fields.
//  - The Denom must pass validateDenom.
//  - The Exponenet must be {SI prefix exponent of this DenomUnit} - basePrefixExp
//  - The aliases must all pass validateDenom.
//  - The aliases must all have the same SI prefix as the Denom (but maybe different forms, e.g. name vs symbol)
func denomUnitValidateBasic(du *banktypes.DenomUnit, rootCoinName string, basePrefixExp int) error {
	// Make sure the Denom is valid.
	denomPrefix, denomError := validateDenom(du.Denom, rootCoinName)
	if denomError != nil {
		return fmt.Errorf("invalid denom: %w", denomError)
	}
	// Make sure the exponent is as expected: {SI prefix exponent of this DenomUnit} - basePrefixExp
	expectedExponent := denomPrefix.GetExponent() - basePrefixExp
	if int(du.Exponent) != expectedExponent {
		return fmt.Errorf("denom [%s] exponent is invalid (expected: %d - %d = %d, actual: %d)",
			du.Denom, denomPrefix.GetExponent(), basePrefixExp, expectedExponent, du.Exponent)
	}
	// Make sure the aliases are all valid and a SI prefix + root using the same prefix as the denom.
	for _, alias := range du.Aliases {
		ap, apErr := validateDenom(alias, rootCoinName)
		if apErr != nil {
			return fmt.Errorf("invalid alias: %w", apErr)
		}
		if denomPrefix != ap {
			return fmt.Errorf("denom [%s] SI prefix is not the same as alias [%s] SI prefix",
				du.Denom, alias)
		}
	}
	return nil
}

// validateDenom checks that:
//  - The denom passes sdk.validateDenom.
//  - The denom is a SI prefix + root coin name (or just the root coin name).
func validateDenom(denom string, rootCoinName string) (SIPrefix, error) {
	if err := sdk.ValidateDenom(denom); err != nil {
		return invalidSIPrefix, err
	}
	prefix, ok := ParseSIPrefixedString(denom, rootCoinName)
	if !ok {
		return invalidSIPrefix, fmt.Errorf("denom [%s] is not a SI prefix + the root coin name [%s]", denom, rootCoinName)
	}
	return prefix, nil
}
