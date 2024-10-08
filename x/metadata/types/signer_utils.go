package types

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// PartyDetails is a struct used to help process party and signer validation.
type PartyDetails struct {
	// address is the bech32 account address string of this party.
	address string
	// role is the type of this party.
	role PartyType
	// optional indicates whether this party's signature is required (false => required, true => not required).
	optional bool

	// acc is the converted address field.
	acc sdk.AccAddress

	// signer is the bech32 string of the account signing for this party.
	// It might be different from the address field if authz is involved.
	signer string
	// signerAcc is the account that's signing for this party.
	// It might be different from the acc field if authz is involved.
	signerAcc sdk.AccAddress

	// canBeUsedBySpec indicates whether this party can be used to fulfill a role required by the spec.
	canBeUsedBySpec bool
	// usedBySpec indicates whether this party has been used to fulfill a role required by the spec.
	usedBySpec bool
}

// WrapRequiredParty creates a PartyDetails from the provided Party.
func WrapRequiredParty(party Party) *PartyDetails {
	return &PartyDetails{
		address:  party.Address,
		role:     party.Role,
		optional: party.Optional,
	}
}

// WrapAvailableParty creates a PartyDetails from the provided Party and marks it as optional and usable.
func WrapAvailableParty(party Party) *PartyDetails {
	return &PartyDetails{
		address:         party.Address,
		role:            party.Role,
		optional:        true, // An available party is optional unless something else says otherwise.
		canBeUsedBySpec: true,
	}
}

// BuildPartyDetails creates the list of PartyDetails to be used in party/signer/role validation.
func BuildPartyDetails(reqParties, availableParties []Party) []*PartyDetails {
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

// Copy returns a copy of this PartyDetails.
func (p *PartyDetails) Copy() *PartyDetails {
	if p == nil {
		return nil
	}
	rv := &PartyDetails{
		address:         p.address,
		role:            p.role,
		optional:        p.optional,
		acc:             nil,
		signer:          p.signer,
		signerAcc:       nil,
		canBeUsedBySpec: p.canBeUsedBySpec,
		usedBySpec:      p.usedBySpec,
	}
	if p.acc != nil {
		rv.acc = make(sdk.AccAddress, len(p.acc))
		copy(rv.acc, p.acc)
	}
	if p.signerAcc != nil {
		rv.signerAcc = make(sdk.AccAddress, len(p.signerAcc))
		copy(rv.signerAcc, p.signerAcc)
	}
	return rv
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

func (p *PartyDetails) SetRole(role PartyType) {
	p.role = role
}

func (p *PartyDetails) GetRole() PartyType {
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
func (p *PartyDetails) IsStillUsableAs(role PartyType) bool {
	return p.CanBeUsed() && !p.IsUsed() && p.GetRole() == role
}

// IsSameAs returns true if this is the same as the provided Party or PartyDetails.
// Only the address and role are considered for this test.
func (p *PartyDetails) IsSameAs(p2 Partier) bool {
	return SamePartiers(p, p2)
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

// TestablePartyDetails should only be used for testing. It's provides a way to customize the fields of a PartyDetails.
type TestablePartyDetails struct {
	Address         string
	Role            PartyType
	Optional        bool
	Acc             sdk.AccAddress
	Signer          string
	SignerAcc       sdk.AccAddress
	CanBeUsedBySpec bool
	UsedBySpec      bool
}

// NewTestablePartyDetails converts a PartyDetails into a TestablePartyDetails.
func NewTestablePartyDetails(pd *PartyDetails) TestablePartyDetails {
	orig := pd.Copy()
	return TestablePartyDetails{
		Address:         orig.address,
		Role:            orig.role,
		Optional:        orig.optional,
		Acc:             orig.acc,
		Signer:          orig.signer,
		SignerAcc:       orig.signerAcc,
		CanBeUsedBySpec: orig.canBeUsedBySpec,
		UsedBySpec:      orig.usedBySpec,
	}
}

// Real returns the PartyDetails version of this TestablePartyDetails.
func (p TestablePartyDetails) Real() *PartyDetails {
	return &PartyDetails{
		address:         p.Address,
		role:            p.Role,
		optional:        p.Optional,
		acc:             p.Acc,
		signer:          p.Signer,
		signerAcc:       p.SignerAcc,
		canBeUsedBySpec: p.CanBeUsedBySpec,
		usedBySpec:      p.UsedBySpec,
	}
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

// authzCacheAcceptableKey creates the key string used in the AuthzCache.acceptable map.
func authzCacheAcceptableKey(grantee, granter sdk.AccAddress, msgTypeURL string) string {
	return string(grantee) + "-" + string(granter) + "-" + msgTypeURL
}

// authzCacheIsWasmKey creates the key string used in the AuthzCache.known map.
func authzCacheIsWasmKey(addr sdk.AccAddress) string {
	return string(addr)
}

// AuthzCache is a struct that houses a map of authz authorizations that are known to have a passed Accept (and been handled).
type AuthzCache struct {
	acceptable map[string]authz.Authorization
	isWasm     map[string]bool
}

// NewAuthzCache creates a new AuthzCache.
func NewAuthzCache() *AuthzCache {
	return &AuthzCache{
		acceptable: make(map[string]authz.Authorization),
		isWasm:     make(map[string]bool),
	}
}

// Clear deletes all entries from this AuthzCache.
func (c *AuthzCache) Clear() {
	for k := range c.acceptable {
		delete(c.acceptable, k)
	}
	for k := range c.isWasm {
		delete(c.isWasm, k)
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

// SetIsWasm records whether an account is a wasm account.
func (c *AuthzCache) SetIsWasm(addr sdk.AccAddress, value bool) {
	c.isWasm[authzCacheIsWasmKey(addr)] = value
}

// HasIsWasm returns true if a cached IsWasm value has been recorded for the given address.
// Use GetIsWasm to get the previously recorded IsWasm value.
func (c *AuthzCache) HasIsWasm(addr sdk.AccAddress) bool {
	_, rv := c.isWasm[authzCacheIsWasmKey(addr)]
	return rv
}

// GetIsWasm returns true if the address was previously recorded as being a wasm account.
// Returns false if either:
//   - The address was previously recorded as NOT being a wasm account.
//   - The WASM status of the account hasn't yet been recorded.
//
// Use HasIsWasm to differentiate the false conditions.
func (c *AuthzCache) GetIsWasm(addr sdk.AccAddress) bool {
	return c.isWasm[authzCacheIsWasmKey(addr)]
}

// GetAcceptableMap returns a copy of the map of acceptable authorizations in this AuthzCache.
// It only exists for unit testing purposes.
func (c *AuthzCache) GetAcceptableMap() map[string]authz.Authorization {
	if c == nil || c.acceptable == nil {
		return nil
	}
	rv := make(map[string]authz.Authorization, len(c.acceptable))
	for k, v := range c.acceptable {
		rv[k] = v
	}
	return rv
}

// GetIsWasmMap returns a copy of the map of previously made IsWasm checks.
// It only exists for unit testing purposes.
func (c *AuthzCache) GetIsWasmMap() map[string]bool {
	if c == nil || c.isWasm == nil {
		return nil
	}
	rv := make(map[string]bool, len(c.isWasm))
	for k, v := range c.isWasm {
		rv[k] = v
	}
	return rv
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
		panic(fmt.Errorf("context value %q is a %T, expected %T", authzCacheContextKey, cacheV, NewAuthzCache()))
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
		panic(fmt.Errorf("context value %q is a %T, expected %T", authzCacheContextKey, cacheV, NewAuthzCache()))
	}
	return cache
}
