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

	// MaxBips is the maximum bips value. 10,000 basis points = 100%.
	MaxBips = uint32(10_000)
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
		ValidateAccessGrantsField("", m.AccessGrants),
		// Nothing to check for with the AcceptingOrders and AllowUserSettlement booleans.
		ValidateReqAttrs("create-ask", m.ReqAttrCreateAsk),
		ValidateReqAttrs("create-bid", m.ReqAttrCreateBid),
		// Nothing to check for the AcceptingCommitments boolean.
		ValidateFeeOptions("create-commitment flat fee", m.FeeCreateCommitmentFlat),
		ValidateBips("commitment settlement", m.CommitmentSettlementBips),
		ValidateIntermediaryDenom(m.IntermediaryDenom),
		ValidateReqAttrs("create-commitment", m.ReqAttrCreateCommitment),
	)
}

// ValidateFeeOptions returns an error if any of the provide coin values is not a valid fee option.
func ValidateFeeOptions(field string, options []sdk.Coin) error {
	var errs []error
	denoms := make(map[string]bool, len(options))
	dups := make(map[string]bool, len(options))
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
	buyerPriceDenomsMap := make(map[string]bool, len(buyerRatios))
	buyerPriceDenoms := make([]string, 0)
	for _, ratio := range buyerRatios {
		if !buyerPriceDenomsMap[ratio.Price.Denom] {
			buyerPriceDenoms = append(buyerPriceDenoms, ratio.Price.Denom)
			buyerPriceDenomsMap[ratio.Price.Denom] = true
		}
	}

	for _, denom := range sellerPriceDenoms {
		if !ContainsString(buyerPriceDenoms, denom) {
			errs = append(errs, fmt.Errorf("denom %q is defined in the seller settlement fee ratios but not buyer", denom))
		}
	}

	for _, denom := range buyerPriceDenoms {
		if !ContainsString(sellerPriceDenoms, denom) {
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

	seen := make(map[string]bool, len(ratios))
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

	seen := make(map[string]bool, len(ratios))
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

// ParseCoin parses a string into an sdk.Coin
func ParseCoin(coinStr string) (sdk.Coin, error) {
	// The sdk.ParseCoinNormalized func allows for decimals and just truncates if there are some.
	// But I want an error if there's a decimal portion.
	// Its errors also always have "invalid decimal coin expression", and I don't want "decimal" in these errors.
	// I also like having the offending coin string quoted since it's safer and clearer when coinStr is "".
	decCoin, err := sdk.ParseDecCoin(coinStr)
	if err != nil || !decCoin.Amount.IsInteger() {
		return sdk.Coin{}, fmt.Errorf("invalid coin expression: %q", coinStr)
	}
	coin, _ := decCoin.TruncateDecimal()
	return coin, nil
}

// ParseFeeRatio parses a "<price>:<fee>" string into a FeeRatio.
func ParseFeeRatio(ratio string) (*FeeRatio, error) {
	parts := strings.Split(ratio, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("cannot create FeeRatio from %q: expected exactly one colon", ratio)
	}
	price, err := ParseCoin(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot create FeeRatio from %q: price: %w", ratio, err)
	}
	fee, err := ParseCoin(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cannot create FeeRatio from %q: fee: %w", ratio, err)
	}
	return &FeeRatio{Price: price, Fee: fee}, nil
}

// MustParseFeeRatio parses a "<price>:<fee>" string into a FeeRatio, panicking if there's a problem.
func MustParseFeeRatio(ratio string) FeeRatio {
	rv, err := ParseFeeRatio(ratio)
	if err != nil {
		panic(err)
	}
	return *rv
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
	rv, rem := QuoRemInt(price.Amount.Mul(r.Fee.Amount), r.Price.Amount)
	mustRound := !rem.IsZero()
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

// ContainsFeeRatio returns true if the fee ratio to find is in the vals slice.
func ContainsFeeRatio(vals []FeeRatio, toFind FeeRatio) bool {
	return contains(vals, toFind, FeeRatio.Equals)
}

// ContainsSameFeeRatioDenoms returns true if any ratio in vals has the same price and fee denoms as the ratio to find.
func ContainsSameFeeRatioDenoms(vals []FeeRatio, toFind FeeRatio) bool {
	return contains(vals, toFind, func(a, b FeeRatio) bool {
		return a.Price.Denom == b.Price.Denom && a.Fee.Denom == b.Fee.Denom
	})
}

// ValidateDisjointFeeRatios returns an error if one or more entries appears in both lists.
func ValidateDisjointFeeRatios(field string, toAdd, toRemove []FeeRatio) error {
	shared := IntersectionOfFeeRatios(toAdd, toRemove)
	if len(shared) > 0 {
		return fmt.Errorf("cannot add and remove the same %s ratios %s", field, FeeRatiosString(shared))
	}
	return nil
}

// ValidateAddRemoveFeeRatiosWithExisting returns errors for entries in toAdd that are
// already in existing, and entries in toRemove that are not in existing.
func ValidateAddRemoveFeeRatiosWithExisting(field string, existing, toAdd, toRemove []FeeRatio) []error {
	var errs []error
	for _, ratio := range toRemove {
		if !ContainsFeeRatio(existing, ratio) {
			errs = append(errs, fmt.Errorf("cannot remove %s ratio fee %q: no such ratio exists", field, ratio))
		}
	}
	newRatios := make([]FeeRatio, 0, len(existing))
	for _, ratio := range existing {
		if !ContainsFeeRatio(toRemove, ratio) {
			newRatios = append(newRatios, ratio)
		}
	}
	for _, ratio := range toAdd {
		if ContainsSameFeeRatioDenoms(newRatios, ratio) {
			errs = append(errs, fmt.Errorf("cannot add %s ratio fee %q: ratio with those denoms already exists",
				field, ratio))
		}
	}
	return errs
}

// ValidateRatioDenoms checks that the buyer and seller ratios have the same price denoms.
func ValidateRatioDenoms(sellerRatios, buyerRatios []FeeRatio) []error {
	var errs []error
	if len(sellerRatios) > 0 && len(buyerRatios) > 0 {
		// We only need to check the price denoms if *both* types have an entry.
		sellerPriceDenoms := make([]string, len(sellerRatios))
		sellerPriceDenomsKnown := make(map[string]bool, len(sellerRatios))
		for i, ratio := range sellerRatios {
			sellerPriceDenoms[i] = ratio.Price.Denom
			sellerPriceDenomsKnown[ratio.Price.Denom] = true
		}

		buyerPriceDenoms := make([]string, 0, len(sellerRatios))
		buyerPriceDenomsKnown := make(map[string]bool, len(sellerRatios))
		for _, ratio := range buyerRatios {
			if !buyerPriceDenomsKnown[ratio.Price.Denom] {
				buyerPriceDenoms = append(buyerPriceDenoms, ratio.Price.Denom)
				buyerPriceDenomsKnown[ratio.Price.Denom] = true
			}
		}

		for _, denom := range sellerPriceDenoms {
			if !buyerPriceDenomsKnown[denom] {
				errs = append(errs, fmt.Errorf("seller settlement fee ratios have price denom %q "+
					"but there are no buyer settlement fee ratios with that price denom", denom))
			}
		}

		for _, denom := range buyerPriceDenoms {
			if !sellerPriceDenomsKnown[denom] {
				errs = append(errs, fmt.Errorf("buyer settlement fee ratios have price denom %q "+
					"but there is not a seller settlement fee ratio with that price denom", denom))
			}
		}
	}

	return errs
}

// ValidateAddRemoveFeeOptions returns an error if the toAdd list has an invalid
// entry or if the two lists have one or more common entries.
func ValidateAddRemoveFeeOptions(field string, toAdd, toRemove []sdk.Coin) error {
	var errs []error
	if err := ValidateFeeOptions(field+" to add", toAdd); err != nil {
		errs = append(errs, err)
	}
	shared := IntersectionOfCoin(toAdd, toRemove)
	if len(shared) > 0 {
		errs = append(errs, fmt.Errorf("cannot add and remove the same %s options %s", field, sdk.Coins(shared)))
	}
	return errors.Join(errs...)
}

// ValidateAddRemoveFeeOptionsWithExisting returns errors for entries in toAdd that are
// already in existing, and entries in toRemove that are not in existing.
func ValidateAddRemoveFeeOptionsWithExisting(field string, existing, toAdd, toRemove []sdk.Coin) []error {
	var errs []error
	for _, coin := range toRemove {
		if !ContainsCoin(existing, coin) {
			errs = append(errs, fmt.Errorf("cannot remove %s flat fee %q: no such fee exists", field, coin))
		}
	}
	newOpts := make([]sdk.Coin, 0, len(existing))
	for _, coin := range existing {
		if !ContainsCoin(toRemove, coin) {
			newOpts = append(newOpts, coin)
		}
	}
	for _, coin := range toAdd {
		if ContainsCoinWithSameDenom(newOpts, coin) {
			errs = append(errs, fmt.Errorf("cannot add %s flat fee %q: fee with that denom already exists", field, coin))
		}
	}
	return errs
}

// ValidateAccessGrantsField returns an error if any of the provided access grants are invalid.
// The provided field is used in error messages.
func ValidateAccessGrantsField(field string, accessGrants []AccessGrant) error {
	if len(field) > 0 && !strings.HasSuffix(field, " ") {
		field += " "
	}
	errs := make([]error, len(accessGrants))
	seen := make(map[string]bool, len(accessGrants))
	dups := make(map[string]bool)
	for i, ag := range accessGrants {
		if seen[ag.Address] && !dups[ag.Address] {
			errs[i] = fmt.Errorf("%s appears in multiple %saccess grant entries", ag.Address, field)
			dups[ag.Address] = true
			continue
		}
		seen[ag.Address] = true
		errs[i] = ag.ValidateInField(field)
	}
	return errors.Join(errs...)
}

// Validate returns an error if there is anything wrong with this AccessGrant.
func (a AccessGrant) Validate() error {
	return a.ValidateInField("")
}

// ValidateInField returns an error if there is anything wrong with this AccessGrant.
// The provided field is included in any error message.
func (a AccessGrant) ValidateInField(field string) error {
	if len(field) > 0 && !strings.HasSuffix(field, " ") {
		field += " "
	}
	_, err := sdk.AccAddressFromBech32(a.Address)
	if err != nil {
		return fmt.Errorf("invalid %saccess grant: invalid address %q: %w", field, a.Address, err)
	}
	if len(a.Permissions) == 0 {
		return fmt.Errorf("invalid %saccess grant: no permissions provided for %s", field, a.Address)
	}
	seen := make(map[Permission]bool, len(a.Permissions))
	for _, perm := range a.Permissions {
		if seen[perm] {
			return fmt.Errorf("invalid %saccess grant: %s appears multiple times for %s", field, perm.SimpleString(), a.Address)
		}
		seen[perm] = true
		if err = perm.Validate(); err != nil {
			return fmt.Errorf("invalid %saccess grant: %w for %s", field, err, a.Address)
		}
	}
	return nil
}

// Contains returns true if this access grant contains the provided permission.
func (a AccessGrant) Contains(perm Permission) bool {
	for _, p := range a.Permissions {
		if p == perm {
			return true
		}
	}
	return false
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
	if !strings.HasPrefix(permUC, "PERMISSION_") {
		permUC = "PERMISSION_" + permUC
	}
	if val, found := Permission_value[permUC]; found && val != int32(Permission_unspecified) {
		return Permission(val), nil
	}
	// special case to allow the underscore to be optional in "set_ids".
	if permUC == "PERMISSION_SETIDS" {
		return Permission_set_ids, nil
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

// NormalizeReqAttrs normalizes/validates each of the provided require attributes.
// The normalized versions of the attributes are returned regardless of whether an error is also returned.
func NormalizeReqAttrs(reqAttrs []string) ([]string, error) {
	rv := make([]string, len(reqAttrs))
	var errs []error
	for i, attr := range reqAttrs {
		rv[i] = nametypes.NormalizeName(attr)
		if !IsValidReqAttr(rv[i]) {
			errs = append(errs, fmt.Errorf("invalid attribute %q", attr))
		}
	}
	return rv, errors.Join(errs...)
}

// ValidateReqAttrsAreNormalized checks that each of the provided attrs is equal to its normalized version.
func ValidateReqAttrsAreNormalized(field string, attrs []string) error {
	var errs []error
	for _, attr := range attrs {
		norm := nametypes.NormalizeName(attr)
		if attr != norm {
			errs = append(errs, fmt.Errorf("%s required attribute %q is not normalized, expected %q", field, attr, norm))
		}
	}
	return errors.Join(errs...)
}

// ValidateReqAttrs makes sure that each provided attribute is valid and that no duplicate entries are provided.
func ValidateReqAttrs(field string, attrs []string) error {
	var errs []error
	seen := make(map[string]bool, len(attrs))
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
// Casing is ignored. Returned values will have the casing that the entry has in list1.
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
	if strings.HasPrefix(reqAttr, "*.") {
		// reqAttr[1:] is used here (instead of [2:]) because we need that . to be
		// part of the match. Otherwise "*.b.a" would match "c.b.a" as well as "c.evilb.a".
		return strings.HasSuffix(accAttr, reqAttr[1:])
	}
	return reqAttr == accAttr
}

// ValidateBips returns an error if the provided bips value is bad. The name is part of the error message.
func ValidateBips(name string, bips uint32) error {
	if bips > MaxBips {
		return fmt.Errorf("invalid %s bips %d: exceeds max of %d", name, bips, MaxBips)
	}
	return nil
}

// ValidateIntermediaryDenom returns an error if a non-empty denom is provided that is not a valid denom.
func ValidateIntermediaryDenom(denom string) error {
	if len(denom) == 0 {
		return nil
	}
	if err := sdk.ValidateDenom(denom); err != nil {
		return fmt.Errorf("invalid intermediary denom: %w", err)
	}
	return nil
}
