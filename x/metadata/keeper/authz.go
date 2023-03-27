package keeper

import (
	"bytes"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
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

// WrapPartyInDetails creates a PartyDetails from the provided Party.
func WrapPartyInDetails(party types.Party) *PartyDetails {
	return &PartyDetails{
		address:  party.Address,
		role:     party.Role,
		optional: party.Optional,
	}
}

// WrapPartyInUsableDetails creates a PartyDetails from the provided Party and marks it as usable.
func WrapPartyInUsableDetails(party types.Party) *PartyDetails {
	return &PartyDetails{
		address:         party.Address,
		role:            party.Role,
		optional:        party.Optional,
		canBeUsedBySpec: true,
	}
}

// BuildPartyDetails creates the list of PartyDetails to be used in party/signer/role validation.
func BuildPartyDetails(reqParties, availableParties []types.Party) []*PartyDetails {
	details := make([]*PartyDetails, len(availableParties), len(availableParties)+len(reqParties))

	// Start with creating details for each available party.
	for i, party := range availableParties {
		details[i] = WrapPartyInUsableDetails(party)
	}

	// Now update the details to include optional=false required parties.
	// If an equal party is already in the details, just update its optional flag
	// to false, otherwise, add it to the list.
reqPartiesLoop:
	for _, reqParty := range reqParties {
		if !reqParty.Optional {
			for _, party := range details {
				if party.EqualsParty(reqParty) {
					party.MakeRequired()
					continue reqPartiesLoop
				}
			}
			details = append(details, WrapPartyInDetails(reqParty))
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

func (p *PartyDetails) IsOptional() bool {
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

// IsStillUsableAs returns true if this party can be use, hasn't yet been used and has the provided role.
func (p *PartyDetails) IsStillUsableAs(role types.PartyType) bool {
	return p.CanBeUsed() && !p.IsUsed() && p.GetRole() == role
}

// EqualsParty is the same as the Party.Equals method.
func (p *PartyDetails) EqualsParty(p2 types.Party) bool {
	return p.GetAddress() == p2.Address && p.GetRole() == p2.Role
}

// SignersWrapper is a thing that holds the signers as strings and acc addresses.
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

// ValidateSignersWithParties ensures the following:
// * All optional=false reqParties and availableParties parties have signed.
// * All required roles are present in availableParties and are signers.
// * All parties with the PROVENANCE role are a smart contract account.
// * All parties with a smart contract account have the PROVENANCE role.
//
// The x/authz module is utilized to help facilitate signer checking.
//
// * reqParties are the parties that might be required to sign, but might not
// necessarily fulfill a required role. This usually comes from a parent entry
// and/or existing entry. They can only fulfill a required role if also provided
// in availableParties. Parties in reqParties with optional=true, are ignored.
// Parties in reqParties with optional=false are required to be in the msg signers.
// * availableParties are the parties available to fulfill required roles. These
// usually come from the proposed entry.
// * reqRoles are all the roles that are required. These usually come from a spec.
//
// If a party is in both reqParties and availableParties, they are only optional
// if both have optional=true.
// Only parties in availableParties that are in the msg signers list are able to
// fulfill an entry in reqRoles, and each such party can each only fulfill one entry.
func (k Keeper) ValidateSignersWithParties(
	ctx sdk.Context,
	reqParties, availableParties []types.Party,
	reqRoles []types.PartyType,
	msg types.MetadataMsg,
) error {
	parties := BuildPartyDetails(reqParties, availableParties)
	signers := NewSignersWrapper(msg.GetSignerStrs())

	// Make sure all required parties are signers.
	associateSigners(parties, signers)
	if err := k.associateAuthorizations(ctx, findMissingRequired(parties), signers, msg, nil); err != nil {
		return err
	}
	if missingReqParties := findMissingRequired(parties); len(missingReqParties) > 0 {
		missing := make([]string, len(missingReqParties))
		for i, party := range missingReqParties {
			missing[i] = party.GetAddress()
		}
		return fmt.Errorf("missing required signature%s: %s", pluralEnding(len(missing)), strings.Join(missing, ", "))
	}

	// Make sure all required roles are present as signers.
	missingRoles := associateRequiredRoles(parties, reqRoles)
	rolesAreMissing, err := k.associateAuthorizationsForRoles(ctx, missingRoles, parties, signers, msg)
	if err != nil {
		return err
	}
	if rolesAreMissing {
		return missingRolesError(parties, reqRoles)
	}

	// Make sure all smart contract accounts have the PROVENANCE role,
	// and all parties with the PROVENANCE role have smart contract accounts.
	if err = k.validateProvenanceRole(ctx, parties); err != nil {
		return err
	}

	return nil
}

// associateSigners updates each PartyDetails to indicate there's a signer if its
// address is in the signers list.
func associateSigners(parties []*PartyDetails, signers *SignersWrapper) {
	for _, party := range parties {
		partyAddress := party.GetAddress()
		for _, signer := range signers.Strings() {
			if partyAddress == signer {
				party.SetSigner(signer)
				break
			}
		}
	}
}

// findMissingRequired returns a list of parties that are required (optional=false)
// and don't have a signer.
func findMissingRequired(parties []*PartyDetails) []*PartyDetails {
	var rv []*PartyDetails
	for _, party := range parties {
		if party.IsRequired() && !party.HasSigner() {
			rv = append(rv, party)
		}
	}
	return rv
}

// associateRequiredRoles goes through the required roles, marking parties as used
// when possible. Returns a list of required role entries that haven't yet been fulfilled.
func associateRequiredRoles(parties []*PartyDetails, reqRoles []types.PartyType) []types.PartyType {
	var missingRoles []types.PartyType
reqRolesLoop:
	for _, role := range reqRoles {
		for _, party := range parties {
			if party.IsStillUsableAs(role) && party.HasSigner() {
				party.MarkAsUsed()
				continue reqRolesLoop
			}
		}
		missingRoles = append(missingRoles, role)
	}
	return missingRoles
}

// missingRolesError generates and returns an error message indicating that
// some required roles don't have signers.
func missingRolesError(parties []*PartyDetails, reqRoles []types.PartyType) error {
	reqCountByRole := make(map[types.PartyType]int)
	haveCountByRole := make(map[types.PartyType]int)
	for _, role := range reqRoles {
		reqCountByRole[role]++
	}
	for _, party := range parties {
		if party.IsUsed() {
			haveCountByRole[party.role]++
		}
	}
	var parts []string
	for _, role := range types.GetAllPartyTypes() {
		if reqCountByRole[role] > haveCountByRole[role] {
			parts = append(parts, fmt.Sprintf("%s need %d have %d",
				role.SimpleString(), reqCountByRole[role], haveCountByRole[role]))
		}
	}
	return fmt.Errorf("missing signers for roles required by spec: %s", strings.Join(parts, ", "))
}

// getAuthzMessageTypeURLs gets all msg type URLs that authz authorizations might
// be under for the provided msg type URL. It basically allows a single authorization
// to be usable from multiple related endpoints. E.g. a MsgWriteScopeRequest authorization
// is usable for the MsgAddScopeDataAccessRequest endpoint as well.
func (k Keeper) getAuthzMessageTypeURLs(msgTypeURL string) []string {
	urls := make([]string, 0, 2)
	if len(msgTypeURL) > 0 {
		urls = append(urls, msgTypeURL)
	}
	switch msgTypeURL {
	case types.TypeURLMsgAddScopeDataAccessRequest, types.TypeURLMsgDeleteScopeDataAccessRequest,
		types.TypeURLMsgAddScopeOwnerRequest, types.TypeURLMsgDeleteScopeOwnerRequest:
		urls = append(urls, types.TypeURLMsgWriteScopeRequest)
	case types.TypeURLMsgWriteRecordRequest:
		urls = append(urls, types.TypeURLMsgWriteSessionRequest)
	case types.TypeURLMsgAddContractSpecToScopeSpecRequest, types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest:
		urls = append(urls, types.TypeURLMsgWriteScopeSpecificationRequest)
	case types.TypeURLMsgWriteRecordSpecificationRequest:
		urls = append(urls, types.TypeURLMsgWriteContractSpecificationRequest)
	case types.TypeURLMsgDeleteRecordSpecificationRequest:
		urls = append(urls, types.TypeURLMsgDeleteContractSpecificationRequest)
	}
	return urls
}

// findAuthzGrantee checks authz for authorizations from the granter to each of the grantees.
// If one is found and accepted, the authorization is updated and that grantee is returned.
// An error is returned if there was a problem updating an authorization.
// If no authorization is found, and no error is encountered, then nil, nil is returned.
func (k Keeper) findAuthzGrantee(
	ctx sdk.Context,
	granter sdk.AccAddress,
	grantees []sdk.AccAddress,
	msg types.MetadataMsg,
) (sdk.AccAddress, error) {
	msgTypes := k.getAuthzMessageTypeURLs(sdk.MsgTypeURL(msg))
	for _, grantee := range grantees {
		for _, msgType := range msgTypes {
			authorization, exp := k.authzKeeper.GetAuthorization(ctx, grantee, granter, msgType)
			if authorization != nil {
				// If Accept returns an error, we just ignore this authorization
				// and look for another that'll work.
				resp, err := authorization.Accept(ctx, msg)
				if err == nil && resp.Accept {
					switch {
					case resp.Delete:
						err = k.authzKeeper.DeleteGrant(ctx, grantee, granter, msgType)
						if err != nil {
							return nil, err
						}
					case resp.Updated != nil:
						if err = k.authzKeeper.SaveGrant(ctx, grantee, granter, resp.Updated, exp); err != nil {
							return nil, err
						}
					}
					return grantee, nil
				}
			}
		}
	}
	return nil, nil
}

// associateAuthorizations checks authz for authorizations from each party (the granters) to
// each signer (the grantees). If found, updates the party details to indicate there's a signer.
// The onAssociation function is called when a grantee is found; it should return whether
// to stop checking (i.e. true => stop now, false => keep checking the rest of the parties).
func (k Keeper) associateAuthorizations(
	ctx sdk.Context,
	parties []*PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
	onAssociation func(party *PartyDetails) (stop bool),
) error {
	for _, party := range parties {
		if !party.HasSigner() {
			grantee, err := k.findAuthzGrantee(ctx, party.GetAcc(), signers.Accs(), msg)
			if err != nil {
				return err
			}
			if len(grantee) > 0 {
				party.SetSignerAcc(grantee)
				if onAssociation != nil && onAssociation(party) {
					break
				}
			}
		}
	}
	return nil
}

// associateAuthorizationsForRoles goes through each entry in roles and attempts
// to find an authorization from an appropriate party. When one is found, the
// party is marked as used.
// An error is returned if one is encountered while updating an authorization.
// True is returned if no usable party is found for one or more roles
// False is returned if all roles have been fulfilled.
func (k Keeper) associateAuthorizationsForRoles(
	ctx sdk.Context,
	roles []types.PartyType,
	parties []*PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
) (bool, error) {
	missingRoles := false
	for _, role := range roles {
		found := false
		var partiesToCheck []*PartyDetails
		for _, party := range parties {
			if party.IsStillUsableAs(role) && !party.HasSigner() {
				partiesToCheck = append(partiesToCheck, party)
			}
		}
		err := k.associateAuthorizations(ctx, partiesToCheck, signers, msg, func(party *PartyDetails) bool {
			party.MarkAsUsed()
			found = true
			return true
		})
		if err != nil {
			return true, err
		}
		if !found {
			missingRoles = true
			// We still want to process the rest so that the error message has the correct counts.
		}
	}

	return missingRoles, nil
}

// validateProvenanceRole makes sure that:
// * All parties with the address of a smart contract have the PROVENANCE role.
// * All parties with the PROVENANCE role have the address of a smart contract.
func (k Keeper) validateProvenanceRole(ctx sdk.Context, parties []*PartyDetails) error {
	for _, party := range parties {
		if party.CanBeUsed() {
			// Using the party address here (instead of the signer) because it's
			// that address that needs to be the smart contract.
			account := k.GetAccount(ctx, party.GetAcc())
			isWasmAcct := account != nil && account.GetSequence() == uint64(0) && account.GetPubKey() == nil
			isProvRole := party.role == types.PartyType_PARTY_TYPE_PROVENANCE
			if isWasmAcct && !isProvRole {
				return fmt.Errorf("account %q is a smart contract but does not have the PROVENANCE role", party.GetAddress())
			}
			if !isWasmAcct && isProvRole {
				return fmt.Errorf("account %q has role PROVENANCE but is not a smart contract", party.GetAddress())
			}
		}
	}
	return nil
}

// checkAuthzForMissing returns any of the provided addrs that have not been granted an authz authorization by one of the msg signers.
// An error is returned if there was a problem updating an authorization.
func (k Keeper) checkAuthzForMissing(
	ctx sdk.Context,
	addrs []string,
	msg types.MetadataMsg,
) ([]string, error) {
	stillMissing := []string{}
	if len(addrs) == 0 {
		return stillMissing, nil
	}

	signers := msg.GetSignerStrs()
	signerAddrs := make([]sdk.AccAddress, 0, len(signers))
	for _, signer := range signers {
		signerAddr, err := sdk.AccAddressFromBech32(signer)
		// If it's not an address, there's no way there's an authorization for them.
		// This is mostly allowed for unit tests.
		// In actual usage, there's very little chance of it not being an address here.
		if err == nil {
			signerAddrs = append(signerAddrs, signerAddr)
		}
	}

	// return as a list this message type and its parent
	// type if it is a message belonging to a hierarchy
	msgTypeURLs := k.getAuthzMessageTypeURLs(sdk.MsgTypeURL(msg))

	for _, addr := range addrs {
		granter, addrErr := sdk.AccAddressFromBech32(addr)
		found := false

		// if the addr wasn't an AccAddress, authz isn't going to help.
		// This is mostly allowed for unit tests.
		// In actual usage, there's very little chance of it not being an address here.
		if addrErr == nil {
			// loop through all the signers
			for _, grantee := range signerAddrs {
				for _, msgType := range msgTypeURLs {
					authorization, exp := k.authzKeeper.GetAuthorization(ctx, grantee, granter, msgType)
					if authorization != nil {
						resp, err := authorization.Accept(ctx, msg)
						if err == nil && resp.Accept {
							switch {
							case resp.Delete:
								err = k.authzKeeper.DeleteGrant(ctx, grantee, granter, msgType)
								if err != nil {
									return stillMissing, err
								}
							case resp.Updated != nil:
								if err = k.authzKeeper.SaveGrant(ctx, grantee, granter, resp.Updated, exp); err != nil {
									return stillMissing, err
								}
							}
							found = true
							break
						}
					}
				}
				if found {
					break
				}
			}
		}

		if !found {
			stillMissing = append(stillMissing, addr)
		}
	}

	return stillMissing, nil
}

// validateAllOwnersAreSigners makes sure that all entries in the existingOwners list
// are contained in the signers list and checks to see if any missing entries have an assigned grantee
func (k Keeper) validateAllOwnersAreSigners(
	ctx sdk.Context,
	existingOwners []string,
	msg types.MetadataMsg,
) error {
	signers := msg.GetSignerStrs()
	missing := findMissing(existingOwners, signers)

	// Check authz for any authorizations on missing signers.
	// If there isn't a message, skip this check (should only happen in unit tests).
	if len(missing) > 0 && msg != nil {
		var err error
		var granter, grantee sdk.AccAddress
		possibleGrantees := safeBech32ToAccAddresses(signers)
		stillMissing := make([]string, 0, len(missing))
		for _, addr := range missing {
			granter, err = sdk.AccAddressFromBech32(addr)
			if err == nil {
				grantee, err = k.findAuthzGrantee(ctx, granter, possibleGrantees, msg)
				if err != nil {
					return fmt.Errorf("error validating signers: %w", err)
				}
				if len(grantee) == 0 {
					stillMissing = append(stillMissing, addr)
				}
			}
		}
		missing = stillMissing
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing signature%s: %s", pluralEnding(len(missing)), strings.Join(missing, ", "))
	}

	return nil
}

// ValidateAllPartiesAreSignersWithAuthz validate all parties are signers with authz module
func (k Keeper) ValidateAllPartiesAreSignersWithAuthz(ctx sdk.Context, parties []types.Party, msg types.MetadataMsg) error {
	addresses := make([]string, len(parties))
	for i, party := range parties {
		addresses[i] = party.Address
	}
	signers := msg.GetSignerStrs()
	missing := findMissing(addresses, signers)
	stillMissing := missing
	var err error
	// Authz grants rights to address on specific message types.
	// If no message is provided, skip the Authz check.
	if msg != nil {
		stillMissing, err = k.checkAuthzForMissing(ctx, missing, msg)
		if err != nil {
			return fmt.Errorf("error validating signers: %w", err)
		}
	}

	if len(stillMissing) > 0 {
		missingWithRoles := make([]string, len(missing))
		for i, addr := range stillMissing {
			for _, party := range parties {
				if addr == party.Address {
					missingWithRoles[i] = fmt.Sprintf("%s (%s)", addr, party.Role.String())
					break
				}
			}
		}
		return fmt.Errorf("missing signature%s from %v", pluralEnding(len(missing)), missingWithRoles)
	}

	return nil
}

// ValidatePartiesInvolved validate that all required parties are involved
func (k Keeper) ValidatePartiesInvolved(parties []types.Party, requiredParties []types.PartyType) error {
	partyRoles := make([]string, len(parties))
	reqRoles := make([]string, len(requiredParties))
	for i, party := range parties {
		partyRoles[i] = party.Role.String()
	}
	for i, req := range requiredParties {
		reqRoles[i] = req.String()
	}
	missing := findMissing(reqRoles, partyRoles)
	if len(missing) > 0 {
		return fmt.Errorf("missing required party type%s %v from parties", pluralEnding(len(missing)), missing)
	}
	return nil
}

// ValidateScopeOwners is stateful validation for scope owners against a scope specification.
// This does NOT involve the Scope.ValidateOwnersBasic() function.
func (k Keeper) ValidateScopeOwners(owners []types.Party, spec types.ScopeSpecification) error {
	var missingPartyTypes []string
	for _, pt := range spec.PartiesInvolved {
		found := false
		for _, o := range owners {
			if o.Role == pt {
				found = true
				break
			}
		}
		if !found {
			// Get the party type without the "PARTY_TYPE_" prefix.
			missingPartyTypes = append(missingPartyTypes, pt.String()[11:])
		}
	}
	if len(missingPartyTypes) > 0 {
		return fmt.Errorf("missing party type%s required by spec: %v", pluralEnding(len(missingPartyTypes)), missingPartyTypes)
	}
	return nil
}

// findMissing returns all elements of the required list that are not found in the entries list
// It is exported only so that it can be unit tested.
func findMissing(required []string, entries []string) []string {
	retval := []string{}
	for _, req := range required {
		found := false
		for _, entry := range entries {
			if req == entry {
				found = true
				break
			}
		}
		if !found {
			retval = append(retval, req)
		}
	}
	return retval
}

// IsMarkerAndHasAuthority checks that the address is a marker addr and that one of the signers has the given role.
// First return boolean is whether the address is a marker.
// Second return boolean is whether one of the signers has the given role on that marker.
// If the first return boolean is false, they'll both be false.
func (k Keeper) IsMarkerAndHasAuthority(ctx sdk.Context, address string, signers []string, role markertypes.Access) (isMarker bool, hasAuth bool) {
	addr, err := sdk.AccAddressFromBech32(address)
	// if the address is invalid then it is not possible for it to be a marker.
	if err != nil {
		return false, false
	}

	acc := k.authKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return false, false
	}

	// Convert over to the actual underlying marker type, or not.
	marker, isMarker := acc.(*markertypes.MarkerAccount)
	if !isMarker {
		return false, false
	}

	// Check if any of the signers have the desired role.
	for _, signer := range signers {
		if marker.HasAccess(signer, role) {
			return true, true
		}
	}

	return true, false
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
