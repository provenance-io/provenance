package exchange

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	nametypes "github.com/provenance-io/provenance/x/name/types"
)

const (
	// MaxName is the maximum length of MarketDetails.Name
	MaxName = 250
	// MaxDescription is the maximum length of MarketDetails.Description
	MaxDescription = 2000
	// MaxWebsiteUrl is the maximum length of MarketDetails.WebsiteUrl
	MaxWebsiteUrl = 200
	// MaxIconUri is the maximum length of MarketDetails.IconUri
	MaxIconUri = 2000
)

// Validate makes sure that everything in this Market is valid.
// The MarketId is allowed to be zero in here.
// Some uses might require it to have a value, but that check is left up to the caller.
func (m Market) Validate() error {
	// Nothing to check on the MarketId. It's allowed to be zero to indicate to use the next one.

	errs := []error{
		m.MarketDetails.Validate(),
		m.FeeCreateAskFlat.Validate(),
		m.FeeCreateBidFlat.Validate(),
		m.FeeSettlementSellerFlat.Validate(),
		m.FeeSettlementBuyerFlat.Validate(),
		ValidateFeeRatios(m.FeeSettlementSellerRatios, m.FeeSettlementBuyerRatios),
		ValidateAccessGrants(m.AccessGrants),
	}

	// Nothing to check for with the AcceptingOrders and AllowUserSettlement booleans.

	if err := ValidateReqAttrs(m.ReqAttrCreateAsk); err != nil {
		errs = append(errs, fmt.Errorf("invalid create ask required attributes: %w", err))
	}
	if err := ValidateReqAttrs(m.ReqAttrCreateBid); err != nil {
		errs = append(errs, fmt.Errorf("invalid create bid required attributes: %w", err))
	}

	return errors.Join(errs...)
}

// Validate returns an error if anything in this MarketDetails is invalid.
func (d MarketDetails) Validate() error {
	var errs []error
	if len(d.Name) > MaxName {
		errs = append(errs, fmt.Errorf("name length %d exceeds maximum length of %d", len(d.Name), MaxName))
	}
	if len(d.Description) > MaxDescription {
		errs = append(errs, fmt.Errorf("description length %d exceeds maximum length of %d", len(d.Description), MaxDescription))
	}
	if len(d.WebsiteUrl) > MaxWebsiteUrl {
		errs = append(errs, fmt.Errorf("website_url length %d exceeds maximum length of %d", len(d.WebsiteUrl), MaxWebsiteUrl))
	}
	if len(d.IconUri) > MaxIconUri {
		errs = append(errs, fmt.Errorf("icon_uri length %d exceeds maximum length of %d", len(d.IconUri), MaxIconUri))
	}
	return errors.Join(errs...)
}

// ValidateSellerFeeRatios returns an error if the provided seller fee ratios contains an invalid entry.
func ValidateSellerFeeRatios(ratios []*FeeRatio) error {
	seen := make(map[string]bool)
	dups := make(map[string]bool)
	var errs []error
	for _, ratio := range ratios {
		key := ratio.Price.Denom
		if seen[key] && !dups[key] {
			errs = append(errs, fmt.Errorf("seller fee ratio denom %q appears in multiple ratios", ratio.Price.Denom))
			dups[key] = true
		}
		seen[key] = true

		if ratio.Price.Denom != ratio.Fee.Denom {
			errs = append(errs, fmt.Errorf("seller fee ratio price denom %q does not equal fee denom %q", ratio.Price.Denom, ratio.Fee.Denom))
		} else {
			errs = append(errs, ratio.Validate())
		}
	}
	return errors.Join(errs...)
}

// ValidateBuyerFeeRatios returns an error if the provided buyer fee ratios contains an invalid entry.
func ValidateBuyerFeeRatios(ratios []*FeeRatio) error {
	seen := make(map[string]bool)
	dups := make(map[string]bool)
	var errs []error
	for _, ratio := range ratios {
		key := ratio.Price.Denom + ":" + ratio.Fee.Denom
		if seen[key] && !dups[key] {
			errs = append(errs, fmt.Errorf("buyer fee ratio pair %q to %q appears in multiple ratios", ratio.Price.Denom, ratio.Fee.Denom))
			dups[key] = true
		}
		seen[key] = true

		errs = append(errs, ratio.Validate())
	}
	return errors.Join(errs...)
}

// ValidateFeeRatios makes sure that the provided fee ratios are valid and have the same price denoms.
func ValidateFeeRatios(sellerRatios, buyerRatios []*FeeRatio) error {
	var errs []error
	if err := ValidateSellerFeeRatios(sellerRatios); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateBuyerFeeRatios(buyerRatios); err != nil {
		errs = append(errs, err)
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	// Make sure the denoms in the prices are the same in the two list.
	// For the seller ones, we know there's only one entry per denom.
	sellerPriceDenoms := make([]string, len(sellerRatios))
	for i, ratio := range sellerRatios {
		sellerPriceDenoms[i] = ratio.Price.Denom
	}

	// For the buyer ones, there can be multiple per price denom.
	buyerPriceDenomsMap := make(map[string]bool)
	buyerPriceDenoms := make([]string, 0)
	for _, ratio := range buyerRatios {
		if !buyerPriceDenomsMap[ratio.Price.Denom] {
			buyerPriceDenoms = append(buyerPriceDenoms, ratio.Price.Denom)
			buyerPriceDenomsMap[ratio.Price.Denom] = true
		}
	}

	for _, denom := range sellerPriceDenoms {
		if !contains(buyerPriceDenoms, denom) {
			errs = append(errs, fmt.Errorf("denom %s is defined in the seller settlement fee ratios but not buyer", denom))
		}
	}

	for _, denom := range buyerPriceDenoms {
		if !contains(sellerPriceDenoms, denom) {
			errs = append(errs, fmt.Errorf("denom %s is defined in the buyer settlement fee ratios but not seller", denom))
		}
	}

	return errors.Join(errs...)
}

// contains returns true if the provided toFind is present in the provided vals.
func contains[T comparable](vals []T, toFind T) bool {
	for _, v := range vals {
		if toFind == v {
			return true
		}
	}
	return false
}

// String returns a string representation of this FeeRatio.
func (r FeeRatio) String() string {
	return fmt.Sprintf("%s:%s", r.Price, r.Fee)
}

// Validate returns an error if this FeeRatio is invalid.
func (r FeeRatio) Validate() error {
	if r.Price.Denom == r.Fee.Denom && r.Fee.Amount.GT(r.Price.Amount) {
		return fmt.Errorf("fee amount %q cannot be greater than price amount %q", r.Fee, r.Price)
	}
	return nil
}

// ValidateAccessGrants returns an error if any of the provided access grants are invalid.
func ValidateAccessGrants(accessGrants []*AccessGrant) error {
	var errs []error
	seen := make(map[string]bool)
	dups := make(map[string]bool)
	for _, ag := range accessGrants {
		if seen[ag.Address] && !dups[ag.Address] {
			errs = append(errs, fmt.Errorf("%s appears in multiple access grant entries", ag.Address))
			dups[ag.Address] = true
		}
		seen[ag.Address] = true
		errs = append(errs, ag.Validate())
	}
	return errors.Join(errs...)
}

// Validate returns an error if there is anything wrong with this AccessGrant.
func (a AccessGrant) Validate() error {
	_, err := sdk.AccAddressFromBech32(a.Address)
	if err != nil {
		return err
	}
	if len(a.Permissions) == 0 {
		return fmt.Errorf("no permissions provided for %s", a.Address)
	}
	seen := make(map[Permission]bool)
	for _, perm := range a.Permissions {
		if seen[perm] {
			return fmt.Errorf("%s appears multiple times for %s", perm.String(), a.Address)
		}
		seen[perm] = true
		if err = perm.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// SimpleString returns a lower-cased version of the permission.String() without the leading "permission_"
// E.g. "settle", or "update".
func (p Permission) SimpleString() string {
	return strings.ToLower(strings.TrimPrefix(p.String(), "PERMISSION_"))
}

// Validate returns an error if this Permission is unspecified or an unknown value.
func (p Permission) Validate() error {
	if p == Permission_unspecified {
		return errors.New("permission is unspecified")
	}
	_, exists := Permission_name[int32(p)]
	if !exists {
		return fmt.Errorf("permission %d does not exist", p)
	}
	return nil
}

// AllPermissions returns all permission values except unspecified.
func AllPermissions() []Permission {
	rv := make([]Permission, 0, len(Permission_name)-1)
	for val := range Permission_name {
		if val != 0 {
			rv = append(rv, Permission(val))
		}
	}
	sort.Slice(rv, func(i, j int) bool {
		return rv[i] < rv[j]
	})
	return rv
}

// ParsePermission converts the provided permission string into a Permission value.
// An error is returned if unknown or Permission_unspecified.
// Example inputs: "settle", "CanCel", "WITHDRAW", "permission_update", "PermISSion_PErmissioNs", "PERMISSION_ATTRIBUTES"
func ParsePermission(permission string) (Permission, error) {
	permUC := strings.ToUpper(permission)
	val, found := Permission_value["PERMISSION_"+permUC]
	if found {
		if val != 0 {
			return Permission(val), nil
		}
	} else {
		val, found = Permission_value[permUC]
		if found && val != 0 {
			return Permission(val), nil
		}
	}
	return Permission_unspecified, fmt.Errorf("invalid permission: %q", permission)
}

// ParsePermissions converts the provided permissions strings into a []Permission.
// An error is returned if any are unknown or Permission_unspecified.
// See also: ParsePermission.
func ParsePermissions(permissions ...string) ([]Permission, error) {
	rv := make([]Permission, len(permissions))
	var errs []error
	for i, perm := range permissions {
		var err error
		rv[i], err = ParsePermission(perm)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return rv, errors.Join(errs...)
}

// ValidateReqAttrs makes sure that each provided attribute is valid and that no duplicate entries are provided.
func ValidateReqAttrs(attrLists ...[]string) error {
	var errs []error
	seen := make(map[string]bool)
	dups := make(map[string]bool)
	for _, attrs := range attrLists {
		for _, attr := range attrs {
			if seen[attr] {
				if !dups[attr] {
					errs = append(errs, fmt.Errorf("duplicate required attribute entry: %q", attr))
					dups[attr] = true
				}
				continue
			}
			seen[attr] = true
			if !IsValidReqAttr(attr) {
				errs = append(errs, fmt.Errorf("invalid required attribute %q", attr))
			}
		}
	}
	return errors.Join(errs...)
}

// IsValidReqAttr returns true if the provided string is a valid required attribute entry.
func IsValidReqAttr(reqAttr string) bool {
	// If it's already valid, we're all good.
	if nametypes.IsValidName(reqAttr) {
		return true
	}

	// If there isn't a wildcard in it, there's no saving it.
	if !strings.Contains(reqAttr, "*") {
		return false
	}

	// Get the normalized version so we can more accurately check things on it.
	normalized := nametypes.NormalizeName(reqAttr)

	// If it's just a wildcard, allow it.
	if normalized == "*" {
		return true
	}

	// The first segment can be a * wildcard.
	// If that's what we've got, make sure everything after it is valid.
	if normalized[:2] == "*." {
		return nametypes.IsValidName(normalized[2:])
	}

	// Nothing left that might save it.
	return false
}
