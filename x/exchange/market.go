package exchange

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	nametypes "github.com/provenance-io/provenance/x/name/types"
)

const (
	// MaxName is the maximum length of MarketDetails.Name
	MaxName = 250
	// MaxDescription is the maximum length of MarketDetails.Description
	MaxDescription = 2000
	// MaxWebsiteURL is the maximum length of MarketDetails.WebsiteUrl
	MaxWebsiteURL = 200
	// MaxIconURI is the maximum length of MarketDetails.IconUri
	MaxIconURI = 2000
)

var (
	_ authtypes.AccountI       = (*MarketAccount)(nil)
	_ authtypes.GenesisAccount = (*MarketAccount)(nil)
)

// Validate makes sure that everything in this Market is valid.
// The MarketId is allowed to be zero in here.
// Some uses might require it to have a value, but that check is left up to the caller.
func (m Market) Validate() error {
	return errors.Join(
		// Nothing to check on the MarketId. It's allowed to be zero to indicate to use the next one.
		m.MarketDetails.Validate(),
		ValidateFeeOptions("create-ask flat fee", m.FeeCreateAskFlat),
		ValidateFeeOptions("create-bid flat fee", m.FeeCreateBidFlat),
		ValidateFeeOptions("seller settlement flat fee", m.FeeSellerSettlementFlat),
		ValidateFeeOptions("buyer settlement flat fee", m.FeeBuyerSettlementFlat),
		ValidateFeeRatios(m.FeeSellerSettlementRatios, m.FeeBuyerSettlementRatios),
		ValidateAccessGrants(m.AccessGrants),
		// Nothing to check for with the AcceptingOrders and AllowUserSettlement booleans.
		ValidateReqAttrs("create-ask", m.ReqAttrCreateAsk),
		ValidateReqAttrs("create-bid", m.ReqAttrCreateBid),
	)
}

// ValidateFeeOptions returns an error if any of the provide coin values is not a valid fee option.
func ValidateFeeOptions(field string, options []sdk.Coin) error {
	var errs []error
	denoms := make(map[string]bool)
	dups := make(map[string]bool)
	for _, coin := range options {
		if denoms[coin.Denom] {
			if !dups[coin.Denom] {
				errs = append(errs, fmt.Errorf("invalid %s option %q: denom used in multiple entries", field, coin))
				dups[coin.Denom] = true
			}
			continue
		}
		denoms[coin.Denom] = true

		err := coin.Validate()
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid %s option %q: %w", field, coin, err))
			continue
		}
		if coin.IsZero() {
			errs = append(errs, fmt.Errorf("invalid %s option %q: amount cannot be zero", field, coin))
		}
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
	if len(d.WebsiteUrl) > MaxWebsiteURL {
		errs = append(errs, fmt.Errorf("website_url length %d exceeds maximum length of %d", len(d.WebsiteUrl), MaxWebsiteURL))
	}
	if len(d.IconUri) > MaxIconURI {
		errs = append(errs, fmt.Errorf("icon_uri length %d exceeds maximum length of %d", len(d.IconUri), MaxIconURI))
	}
	return errors.Join(errs...)
}

// ValidateFeeRatios makes sure that the provided fee ratios are valid and have the same price denoms.
func ValidateFeeRatios(sellerRatios, buyerRatios []FeeRatio) error {
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
		if !containsString(buyerPriceDenoms, denom) {
			errs = append(errs, fmt.Errorf("denom %q is defined in the seller settlement fee ratios but not buyer", denom))
		}
	}

	for _, denom := range buyerPriceDenoms {
		if !containsString(sellerPriceDenoms, denom) {
			errs = append(errs, fmt.Errorf("denom %q is defined in the buyer settlement fee ratios but not seller", denom))
		}
	}

	return errors.Join(errs...)
}

// ValidateSellerFeeRatios returns an error if the provided seller fee ratios contains an invalid entry.
func ValidateSellerFeeRatios(ratios []FeeRatio) error {
	if len(ratios) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	dups := make(map[string]bool)
	var errs []error
	for _, ratio := range ratios {
		key := ratio.Price.Denom
		if seen[key] {
			if !dups[key] {
				errs = append(errs, fmt.Errorf("seller fee ratio denom %q appears in multiple ratios", ratio.Price.Denom))
				dups[key] = true
			}
			continue
		}
		seen[key] = true

		if ratio.Price.Denom != ratio.Fee.Denom {
			errs = append(errs, fmt.Errorf("seller fee ratio price denom %q does not equal fee denom %q", ratio.Price.Denom, ratio.Fee.Denom))
			continue
		}

		if err := ratio.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("seller fee ratio %w", err))
		}
	}
	return errors.Join(errs...)
}

// ValidateBuyerFeeRatios returns an error if the provided buyer fee ratios contains an invalid entry.
func ValidateBuyerFeeRatios(ratios []FeeRatio) error {
	if len(ratios) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	dups := make(map[string]bool)
	var errs []error
	for _, ratio := range ratios {
		key := ratio.Price.Denom + ":" + ratio.Fee.Denom
		if seen[key] {
			if !dups[key] {
				errs = append(errs, fmt.Errorf("buyer fee ratio pair %q to %q appears in multiple ratios", ratio.Price.Denom, ratio.Fee.Denom))
				dups[key] = true
			}
			continue
		}
		seen[key] = true

		if err := ratio.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("buyer fee ratio %w", err))
		}
	}
	return errors.Join(errs...)
}

// containsString returns true if the string to find is in the vals slice.
func containsString(vals []string, toFind string) bool {
	return contains(vals, toFind, func(a, b string) bool {
		return a == b
	})
}

// String returns a string representation of this FeeRatio.
func (r FeeRatio) String() string {
	return fmt.Sprintf("%s:%s", r.Price, r.Fee)
}

// FeeRatiosString converts the provided ratios into a single string with format <ratio1>,<ratio2>,...
func FeeRatiosString(ratios []FeeRatio) string {
	entries := make([]string, len(ratios))
	for i, ratio := range ratios {
		entries[i] = ratio.String()
	}
	return strings.Join(entries, ",")
}

// Validate returns an error if this FeeRatio is invalid.
func (r FeeRatio) Validate() error {
	if !r.Price.Amount.IsPositive() {
		return fmt.Errorf("price amount %q must be positive", r.Price)
	}
	if r.Fee.Amount.IsNegative() {
		return fmt.Errorf("fee amount %q cannot be negative", r.Fee)
	}
	if r.Price.Denom == r.Fee.Denom && r.Fee.Amount.GT(r.Price.Amount) {
		return fmt.Errorf("fee amount %q cannot be greater than price amount %q", r.Fee, r.Price)
	}
	return nil
}

// Equals returns true if this FeeRatio has the same price and fee as the provided other FeeRatio.
func (r FeeRatio) Equals(other FeeRatio) bool {
	// Cannot use coin.IsEqual because it panics if the denoms are different, because that makes perfect sense.
	// The coin.Equal(interface{}) function behaves as expected, though, but with the extra casting costs.
	return r.Price.Equal(other.Price) && r.Fee.Equal(other.Fee)
}

// applyLooselyTo returns the amount that results from the application of this ratio to the given price.
// The second return value is whether rounding was needed. I.e. If price / ratio price * fee price is
// not a whole number, the returned amount is increased by one and the second return value will be true.
// If it is a whole number, the second return value is false.
// An error is returned if the price's denom does not equal the ratio's price denom, or if the ratio's price amount is zero.
func (r FeeRatio) applyLooselyTo(price sdk.Coin) (sdkmath.Int, bool, error) {
	if r.Price.Denom != price.Denom {
		return sdkmath.ZeroInt(), false, fmt.Errorf("cannot apply ratio %s to price %s: incorrect price denom", r, price)
	}
	if r.Price.Amount.IsZero() {
		return sdkmath.ZeroInt(), false, fmt.Errorf("cannot apply ratio %s to price %s: division by zero", r, price)
	}
	rv := price.Amount.Mul(r.Fee.Amount)
	mustRound := !rv.Mod(r.Price.Amount).IsZero()
	rv = rv.Quo(r.Price.Amount)
	if mustRound {
		rv = rv.Add(sdkmath.OneInt())
	}
	return rv, mustRound, nil
}

// ApplyTo attempts to calculate the fee that results from applying this fee ratio to the provided price.
func (r FeeRatio) ApplyTo(price sdk.Coin) (sdk.Coin, error) {
	rv := sdk.Coin{Denom: "", Amount: sdk.ZeroInt()}
	amt, wasRounded, err := r.applyLooselyTo(price)
	if err != nil {
		return rv, err
	}
	if wasRounded {
		return rv, fmt.Errorf("cannot apply ratio %s to price %s: price amount cannot be evenly divided by ratio price", r, price)
	}
	rv.Denom = r.Fee.Denom
	rv.Amount = amt
	return rv, nil
}

// ApplyToLoosely calculates the fee that results from applying this fee ratio to the provided price, allowing for the
// ratio to not evenly apply to the price.
func (r FeeRatio) ApplyToLoosely(price sdk.Coin) (sdk.Coin, error) {
	rv := sdk.Coin{Denom: "", Amount: sdk.ZeroInt()}
	amt, _, err := r.applyLooselyTo(price)
	if err != nil {
		return rv, err
	}
	rv.Denom = r.Fee.Denom
	rv.Amount = amt
	return rv, nil
}

// IntersectionOfFeeRatios returns each FeeRatio entry that is in both lists.
func IntersectionOfFeeRatios(list1, list2 []FeeRatio) []FeeRatio {
	return intersection(list1, list2, FeeRatio.Equals)
}

// ValidateDisjointFeeRatios returns an error if one or more entries appears in both lists.
func ValidateDisjointFeeRatios(field string, toAdd, toRemove []FeeRatio) error {
	shared := IntersectionOfFeeRatios(toAdd, toRemove)
	if len(shared) > 0 {
		return fmt.Errorf("cannot add and remove the same %s ratios %s", field, FeeRatiosString(shared))
	}
	return nil
}

// CoinEquals returns true if the two provided coin entries are equal.
// Designed for use with intersection.
//
// We can't just provide sdk.Coin.isEqual to intersection because that PANICS if the denoms are different.
// And we can't provide sdk.Coin.Equal to intersection because it takes in an interface{} (instead of sdk.Coin).
func CoinEquals(a, b sdk.Coin) bool {
	return a.Equal(b)
}

// IntersectionOfCoins returns each sdk.Coin entry that is in both lists.
func IntersectionOfCoins(list1, list2 []sdk.Coin) []sdk.Coin {
	return intersection(list1, list2, CoinEquals)
}

// ValidateAddRemoveFeeOptions returns an error if the toAdd list has an invalid
// entry or if the two lists have one or more common entries.
func ValidateAddRemoveFeeOptions(field string, toAdd, toRemove []sdk.Coin) error {
	var errs []error
	if err := ValidateFeeOptions(field+" to add", toAdd); err != nil {
		errs = append(errs, err)
	}
	shared := IntersectionOfCoins(toAdd, toRemove)
	if len(shared) > 0 {
		errs = append(errs, fmt.Errorf("cannot add and remove the same %s options %s", field, sdk.Coins(shared)))
	}
	return errors.Join(errs...)
}

// ValidateAccessGrants returns an error if any of the provided access grants are invalid.
func ValidateAccessGrants(accessGrants []AccessGrant) error {
	errs := make([]error, len(accessGrants))
	seen := make(map[string]bool)
	dups := make(map[string]bool)
	for i, ag := range accessGrants {
		if seen[ag.Address] && !dups[ag.Address] {
			errs[i] = fmt.Errorf("%s appears in multiple access grant entries", ag.Address)
			dups[ag.Address] = true
			continue
		}
		seen[ag.Address] = true
		errs[i] = ag.Validate()
	}
	return errors.Join(errs...)
}

// Validate returns an error if there is anything wrong with this AccessGrant.
func (a AccessGrant) Validate() error {
	_, err := sdk.AccAddressFromBech32(a.Address)
	if err != nil {
		return fmt.Errorf("invalid access grant: invalid address: %w", err)
	}
	if len(a.Permissions) == 0 {
		return fmt.Errorf("invalid access grant: no permissions provided for %s", a.Address)
	}
	seen := make(map[Permission]bool)
	for _, perm := range a.Permissions {
		if seen[perm] {
			return fmt.Errorf("invalid access grant: %s appears multiple times for %s", perm.SimpleString(), a.Address)
		}
		seen[perm] = true
		if err = perm.Validate(); err != nil {
			return fmt.Errorf("invalid access grant: %w for %s", err, a.Address)
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
	permUC := strings.ToUpper(strings.TrimSpace(permission))
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
	if len(permissions) == 0 {
		return nil, nil
	}
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
func ValidateReqAttrs(field string, attrs []string) error {
	var errs []error
	seen := make(map[string]bool)
	bad := make(map[string]bool)
	for _, attr := range attrs {
		normalized := nametypes.NormalizeName(attr)
		if seen[normalized] {
			if !bad[normalized] {
				errs = append(errs, fmt.Errorf("duplicate %s required attribute %q",
					field, attr))
				bad[normalized] = true
			}
			continue
		}
		seen[normalized] = true
		if !IsValidReqAttr(normalized) {
			errs = append(errs, fmt.Errorf("invalid %s required attribute %q", field, attr))
			bad[normalized] = true
		}
	}
	return errors.Join(errs...)
}

// IntersectionOfAttributes returns each attribute that is in both lists.
func IntersectionOfAttributes(list1, list2 []string) []string {
	return intersection(list1, list2, strings.EqualFold)
}

// ValidateAddRemoveReqAttrs returns an error if the toAdd list has an invalid
// entry or if the two lists have one or more common entries.
func ValidateAddRemoveReqAttrs(field string, toAdd, toRemove []string) error {
	var errs []error
	if err := ValidateReqAttrs(field+" to add", toAdd); err != nil {
		errs = append(errs, err)
	}
	// attributes are lower-cased during attribute, so we should use a case-insensitive matcher.
	shared := IntersectionOfAttributes(toAdd, toRemove)
	if len(shared) > 0 {
		errs = append(errs, fmt.Errorf("cannot add and remove the same %s required attributes \"%s\"", field, strings.Join(shared, "\",\"")))
	}
	return errors.Join(errs...)
}

// IsValidReqAttr returns true if the provided string is a valid required attribute entry.
// Assumes that the provided reqAttr has already been normalized.
func IsValidReqAttr(reqAttr string) bool {
	// Allow it to just be the wildcard character.
	if reqAttr == "*" {
		return true
	}

	// A leading wildcard segment is valid for us, but not the name module. So, remove it if it's there.
	reqAttr = strings.TrimPrefix(reqAttr, "*.")

	// IsValidName doesn't consider length, so an empty string is valid by it, but not valid in here.
	if len(reqAttr) == 0 {
		return false
	}

	return nametypes.IsValidName(reqAttr)
}

// FindUnmatchedReqAttrs returns all required attributes that don't have a match in the provided account attributes.
// This assumes that reqAttrs and accAttrs have all been normalized.
func FindUnmatchedReqAttrs(reqAttrs, accAttrs []string) []string {
	var rv []string
	for _, reqAttr := range reqAttrs {
		if !HasReqAttrMatch(reqAttr, accAttrs) {
			rv = append(rv, reqAttr)
		}
	}
	return rv
}

// HasReqAttrMatch returns true if one (or more) accAttrs is a match for the provided required attribute.
// This assumes that reqAttr and accAttrs have all been normalized.
func HasReqAttrMatch(reqAttr string, accAttrs []string) bool {
	for _, accAttr := range accAttrs {
		if IsReqAttrMatch(reqAttr, accAttr) {
			return true
		}
	}
	return false
}

// IsReqAttrMatch returns true if the provide account attribute is a match for the given required attribute.
// This assumes that reqAttr and accAttr have both been normalized.
func IsReqAttrMatch(reqAttr, accAttr string) bool {
	if len(reqAttr) == 0 || len(accAttr) == 0 {
		return false
	}
	if reqAttr == "*" {
		return true
	}
	if strings.HasPrefix(reqAttr, "*.") {
		// reqAttr[1:] is used here (instead of [2:]) because we need that . to be
		// part of the match. Otherwise "*.b.a" would match "c.b.a" as well as "c.evilb.a".
		return strings.HasSuffix(accAttr, reqAttr[1:])
	}
	return reqAttr == accAttr
}

// contains returns true if the provided toFind is present in the provided vals.
func contains[T any](vals []T, toFind T, equals func(T, T) bool) bool {
	for _, v := range vals {
		if equals(toFind, v) {
			return true
		}
	}
	return false
}

// intersection returns each entry that is in both lists.
func intersection[T any](list1, list2 []T, equals func(T, T) bool) []T {
	var rv []T
	for _, a := range list1 {
		if contains(list2, a, equals) && !contains(rv, a, equals) {
			rv = append(rv, a)
		}
	}
	return rv
}
