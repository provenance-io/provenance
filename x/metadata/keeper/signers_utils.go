package keeper

import (
	"bytes"
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// PartyDetails is a struct used to help process party and signer validation.
// Even though all the fields are public, you should usually use the Get/Set methods
// which handle automatic bech32 conversion when needed and reduce duplicated efforts.
type PartyDetails struct {
	address  string
	role     types.PartyType
	optional bool

	acc       sdk.AccAddress
	signer    string
	signerAcc sdk.AccAddress

	canBeUsedBySpec bool
	usedBySpec      bool
}

// WrapRequiredParty creates a PartyDetails from the provided Party.
func WrapRequiredParty(party types.Party) *PartyDetails {
	return &PartyDetails{
		address:  party.Address,
		role:     party.Role,
		optional: party.Optional,
	}
}

// WrapAvailableParty creates a PartyDetails from the provided Party and marks it as optional and usable.
func WrapAvailableParty(party types.Party) *PartyDetails {
	return &PartyDetails{
		address:         party.Address,
		role:            party.Role,
		optional:        true, // An available party is optional unless something else says otherwise.
		canBeUsedBySpec: true,
	}
}

// BuildPartyDetails creates the list of PartyDetails to be used in party/signer/role validation.
func BuildPartyDetails(reqParties, availableParties []types.Party) []*PartyDetails {
	details := make([]*PartyDetails, 0, len(availableParties))

	// Start with creating details for each available party.
availablePartiesLoop:
	for _, party := range availableParties {
		for _, known := range details {
			if party.IsSameAs(known) {
				continue availablePartiesLoop
			}
		}
		details = append(details, WrapAvailableParty(party))
	}

	// Now update the details to include optional=false required parties.
	// If an equal party is already in the details, just update its optional flag
	// to false, otherwise, add it to the list.
reqPartiesLoop:
	for _, reqParty := range reqParties {
		if !reqParty.Optional {
			for _, party := range details {
				if reqParty.IsSameAs(party) {
					party.MakeRequired()
					continue reqPartiesLoop
				}
			}
			details = append(details, WrapRequiredParty(reqParty))
		}
	}

	return details
}

func (p *PartyDetails) SetAddress(address string) {
	if p.address != address {
		p.acc = nil
	}
	p.address = address
}

func (p *PartyDetails) GetAddress() string {
	if len(p.address) == 0 && len(p.acc) > 0 {
		p.address = p.acc.String()
	}
	return p.address
}

func (p *PartyDetails) SetAcc(addr sdk.AccAddress) {
	if !bytes.Equal(p.acc, addr) {
		p.address = ""
	}
	p.acc = addr
}

func (p *PartyDetails) GetAcc() sdk.AccAddress {
	if len(p.acc) == 0 && len(p.address) > 0 {
		p.acc, _ = sdk.AccAddressFromBech32(p.address)
	}
	return p.acc
}

func (p *PartyDetails) SetRole(role types.PartyType) {
	p.role = role
}

func (p *PartyDetails) GetRole() types.PartyType {
	return p.role
}

func (p *PartyDetails) SetOptional(optional bool) {
	p.optional = optional
}

func (p *PartyDetails) MakeRequired() {
	p.optional = false
}

func (p *PartyDetails) GetOptional() bool {
	return p.optional
}

func (p *PartyDetails) IsRequired() bool {
	return !p.optional
}

func (p *PartyDetails) SetSigner(signer string) {
	if p.signer != signer {
		p.signerAcc = nil
	}
	p.signer = signer
}

func (p *PartyDetails) GetSigner() string {
	if len(p.signer) == 0 && len(p.signerAcc) > 0 {
		p.signer = p.signerAcc.String()
	}
	return p.signer
}

func (p *PartyDetails) SetSignerAcc(signerAddr sdk.AccAddress) {
	if !bytes.Equal(p.signerAcc, signerAddr) {
		p.signer = ""
	}
	p.signerAcc = signerAddr
}

func (p *PartyDetails) GetSignerAcc() sdk.AccAddress {
	if len(p.signerAcc) == 0 && len(p.signer) > 0 {
		p.signerAcc, _ = sdk.AccAddressFromBech32(p.signer)
	}
	return p.signerAcc
}

func (p *PartyDetails) HasSigner() bool {
	return len(p.signer) > 0 || len(p.signerAcc) > 0
}

func (p *PartyDetails) CanBeUsed() bool {
	return p.canBeUsedBySpec
}

func (p *PartyDetails) MarkAsUsed() {
	p.usedBySpec = true
}

func (p *PartyDetails) IsUsed() bool {
	return p.usedBySpec
}

// IsStillUsableAs returns true if this party can be used, hasn't yet been used and has the provided role.
func (p *PartyDetails) IsStillUsableAs(role types.PartyType) bool {
	return p.CanBeUsed() && !p.IsUsed() && p.GetRole() == role
}

// IsSameAs returns true if this is the same as the provided Party or PartyDetails.
// Only the address and role are considered for this test.
func (p *PartyDetails) IsSameAs(p2 types.Partier) bool {
	return types.SamePartiers(p, p2)
}

// GetUsedSigners gets a map of bech32 strings to true with a key for each used signer.
func GetUsedSigners(parties []*PartyDetails) UsedSignersMap {
	rv := make(UsedSignersMap)
	for _, party := range parties {
		if party.HasSigner() {
			rv.Use(party.GetSigner())
		}
	}
	return rv
}

// SignersWrapper stores the signers as strings and acc addresses.
// One is created by providing the strings. They are then converted to acc addresses
// if they're needed that way.
type SignersWrapper struct {
	signers    []string
	signerAccs []sdk.AccAddress
	converted  bool
}

func NewSignersWrapper(signers []string) *SignersWrapper {
	return &SignersWrapper{signers: signers}
}

// Strings gets the string versions of the signers.
func (s *SignersWrapper) Strings() []string {
	return s.signers
}

// Accs gets the sdk.AccAddress versions of the signers.
// Conversion happens if it hasn't already been done yet.
// Any strings that fail to convert are simply ignored.
func (s *SignersWrapper) Accs() []sdk.AccAddress {
	if !s.converted {
		s.signerAccs = safeBech32ToAccAddresses(s.signers)
		s.converted = true
	}
	return s.signerAccs
}

// authzCacheAcceptableKey creates the key string used in the AuthzCache.acceptable map.
func authzCacheAcceptableKey(grantee, granter sdk.AccAddress, msgTypeURL string) string {
	return string(grantee) + "-" + string(granter) + "-" + msgTypeURL
}

// AuthzCache is a struct that houses a map of authz authorizations that are known to have a passed Accept (and been handled).
type AuthzCache struct {
	acceptable map[string]authz.Authorization
}

// NewAuthzCache creates a new AuthzCache.
func NewAuthzCache() *AuthzCache {
	return &AuthzCache{acceptable: make(map[string]authz.Authorization)}
}

// Clear deletes all entries from this AuthzCache.
func (c *AuthzCache) Clear() {
	for k := range c.acceptable {
		delete(c.acceptable, k)
	}
}

// SetAcceptable sets an authorization in this cache as acceptable.
func (c *AuthzCache) SetAcceptable(grantee, granter sdk.AccAddress, msgTypeURL string, authorization authz.Authorization) {
	c.acceptable[authzCacheAcceptableKey(grantee, granter, msgTypeURL)] = authorization
}

// GetAcceptable gets a previously set acceptable authorization.
// Returns nil if no such authorization exists.
func (c *AuthzCache) GetAcceptable(grantee, granter sdk.AccAddress, msgTypeURL string) authz.Authorization {
	return c.acceptable[authzCacheAcceptableKey(grantee, granter, msgTypeURL)]
}

// authzCacheContextKey is the key used in an sdk.Context to set/get the AuthzCache.
const authzCacheContextKey = "authzCacheContextKey"

// AddAuthzCacheToContext either returns a new sdk.Context with the addition of an AuthzCache,
// or clears out the AuthzCache if it already exists in the context.
// It panics if the AuthzCache key exists in the context but isn't an AuthzCache.
func AddAuthzCacheToContext(ctx sdk.Context) sdk.Context {
	// If it's already got one, leave it there but clear it out.
	// Otherwise, we'll add a new one.
	if cacheV := ctx.Value(authzCacheContextKey); cacheV != nil {
		if cache, ok := cacheV.(*AuthzCache); ok {
			cache.Clear()
			return ctx
		}
		// If the key was there, but not an AuthzCache, things are very wrong. Panic.
		panic(fmt.Errorf("context value %q is a %T, expected %T",
			authzCacheContextKey, cacheV, NewAuthzCache()))
	}
	return ctx.WithValue(authzCacheContextKey, NewAuthzCache())
}

// GetAuthzCache gets the AuthzCache from the context or panics.
func GetAuthzCache(ctx sdk.Context) *AuthzCache {
	cacheV := ctx.Value(authzCacheContextKey)
	if cacheV == nil {
		panic(fmt.Errorf("context does not contain a %q value", authzCacheContextKey))
	}
	cache, ok := cacheV.(*AuthzCache)
	if !ok {
		panic(fmt.Errorf("context value %q is a %T, expected %T",
			authzCacheContextKey, cacheV, NewAuthzCache()))
	}
	return cache
}

// UnwrapMetadataContext retrieves a Context from a context.Context instance attached with WrapSDKContext.
// It then adds an AuthzCache to it.
// It panics if a Context was not properly attached, or if the AuthzCache can't be added.
//
// This should be used for all Metadata msg server endpoints instead of sdk.UnwrapSDKContext.
// This should not be used outside of the Metadata module.
func UnwrapMetadataContext(goCtx context.Context) sdk.Context {
	return AddAuthzCacheToContext(sdk.UnwrapSDKContext(goCtx))
}

// UsedSignersMap is a type for recording that a signer has been used.
type UsedSignersMap map[string]bool

// NewUsedSignersMap creates a new UsedSignersMap
func NewUsedSignersMap() UsedSignersMap {
	return make(UsedSignersMap)
}

// Use notes that the provided addresses have been used.
func (m UsedSignersMap) Use(addrs ...string) UsedSignersMap {
	for _, addr := range addrs {
		m[addr] = true
	}
	return m
}

// IsUsed returns true if the provided address has been used.
func (m UsedSignersMap) IsUsed(addr string) bool {
	return m[addr]
}

// AlsoUse adds all the entries in the provided UsedSignersMap to this UsedSignersMap.
func (m UsedSignersMap) AlsoUse(m2 UsedSignersMap) UsedSignersMap {
	for k := range m2 {
		m[k] = true
	}
	return m
}

// findMissing returns all elements of the required list that are not found in the entries list.
//
// See also: findMissingComp
func findMissing(required, toCheck []string) []string {
	return findMissingComp(required, toCheck, func(r, c string) bool { return r == c })
}

// findMissingParties returns all parties from the required list that don't have a same party in the toCheck list.
//
// See also: findMissingComp
func findMissingParties(required, toCheck []types.Party) []types.Party {
	return findMissingComp(required, toCheck, func(r, c types.Party) bool { return types.SamePartiers(&r, &c) })
}

// findMissingComp returns all entries in required where an entry does not exist in toCheck
// such that the provided comp function returns true.
// Duplicate entries in required do not require duplicate entries in toCheck.
// E.g. findMissingComp([a, b, a], [a]) => [b], and findMissingComp([a, b, a], [b]) => [a, a].
func findMissingComp[R any, C any](required []R, toCheck []C, comp func(R, C) bool) []R {
	var rv []R
reqLoop:
	for _, req := range required {
		for _, entry := range toCheck {
			if comp(req, entry) {
				continue reqLoop
			}
		}
		rv = append(rv, req)
	}
	return rv
}

// pluralEnding returns "" if i == 1, or "s" otherwise.
func pluralEnding(i int) string {
	if i == 1 {
		return ""
	}
	return "s"
}

// safeBech32ToAccAddresses attempts to convert all provided strings to AccAddresses.
// Any that fail to convert are ignored.
func safeBech32ToAccAddresses(bech32s []string) []sdk.AccAddress {
	rv := make([]sdk.AccAddress, 0, len(bech32s))
	for _, bech32 := range bech32s {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err == nil {
			rv = append(rv, addr)
		}
	}
	return rv
}
